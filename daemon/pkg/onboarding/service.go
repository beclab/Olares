package onboarding

import (
	"context"
	"errors"
	"fmt"
	"sync"
	"time"

	"golang.org/x/crypto/bcrypt"

	"github.com/beclab/Olares/cli/pkg/releasefile"
	sysv1 "github.com/beclab/api/api/sys.bytetrade.io/v1alpha1"
	iamv1 "github.com/beclab/api/iam/v1alpha2"
	corev1 "k8s.io/api/core/v1"
	apierrors "k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/api/resource"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/util/validation"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/rest"
	"k8s.io/client-go/util/retry"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

const (
	Namespace         = "os-framework"
	RecordName        = "olares-onboarding"
	UserMarker        = "bytetrade.io/onboarding"
	UserDID           = "bytetrade.io/user-did"
	userWizardStatus  = "bytetrade.io/wizard-status"
	waitResetPassword = "wait_reset_password"
	Waiting           = "waiting-for-user"
	Creating          = "creating"
	Created           = "created"
	WaitActivateVault = "wait_activate_vault"
	WaitingPassword   = "waiting-for-reset-password"
	Completed         = "completed"
	Failed            = "failed"
)

var (
	ErrConflict = errors.New("first user is already reserved or exists")
	ErrIdentity = errors.New("request identity does not match the first user")
	ErrClosed   = errors.New("onboarding is complete")
)

type CreateRequest struct {
	Username string `json:"username"`
	// Password is the salted MD5 value provided by the client.
	Password string `json:"password"`
	DID      string `json:"did"`
	CPU      string `json:"cpu_limit"`
	Memory   string `json:"memory_limit"`
}

type Status struct {
	State       string `json:"state"`
	Username    string `json:"username,omitempty"`
	Reason      string `json:"reason,omitempty"`
	CompletedAt string `json:"completed_at,omitempty"`
}

type Service struct {
	mu            sync.Mutex
	Kube          kubernetes.Interface
	Resources     client.Client
	updateRelease func(context.Context, map[string]string) error
}

func NewResourceClient(config *rest.Config) (client.Client, error) {
	scheme := runtime.NewScheme()
	if err := iamv1.AddToScheme(scheme); err != nil {
		return nil, err
	}
	if err := sysv1.AddToScheme(scheme); err != nil {
		return nil, err
	}
	return client.New(config, client.Options{Scheme: scheme})
}

func (r *CreateRequest) validate() error {
	if len(r.Username) > 32 || len(validation.IsDNS1123Label(r.Username)) != 0 {
		return errors.New("username must be a DNS label of at most 32 characters")
	}
	if r.Username == "olares" || r.Username == "admin" {
		return errors.New("username is reserved")
	}
	if len(r.Password) < 8 || len(r.Password) > 72 {
		return errors.New("password must contain 8 to 72 bytes")
	}
	if r.CPU == "" {
		r.CPU = "4"
	}
	if r.Memory == "" {
		r.Memory = "8Gi"
	}
	for _, value := range []string{r.CPU, r.Memory} {
		q, err := resource.ParseQuantity(value)
		if err != nil || q.Sign() <= 0 {
			return errors.New("cpu_limit and memory_limit must be positive quantities")
		}
	}
	return nil
}

// Create reserves one user in the cluster before creating the User. Retries
// by the same username and password resume this reservation. The optional DID
// is user metadata and may be set or changed later during binding.
// The Secret survives daemon restarts and is removed by a full uninstall.
func (s *Service) Create(ctx context.Context, request CreateRequest) (Status, error) {
	s.mu.Lock()
	defer s.mu.Unlock()
	if err := request.validate(); err != nil {
		return Status{}, err
	}
	records := s.Kube.CoreV1().Secrets(Namespace)
	record, err := records.Get(ctx, RecordName, metav1.GetOptions{})
	if apierrors.IsNotFound(err) {
		var users iamv1.UserList
		e := s.Resources.List(ctx, &users)
		if e != nil {
			return Status{}, e
		}
		if len(users.Items) != 0 {
			return Status{}, ErrConflict
		}
		hash, e := bcrypt.GenerateFromPassword([]byte(request.Password), bcrypt.DefaultCost)
		if e != nil {
			return Status{}, e
		}
		record, err = records.Create(ctx, &corev1.Secret{
			ObjectMeta: metav1.ObjectMeta{Name: RecordName, Namespace: Namespace},
			Data:       map[string][]byte{"username": []byte(request.Username), "password_hash": hash},
		}, metav1.CreateOptions{})
		if apierrors.IsAlreadyExists(err) {
			record, err = records.Get(ctx, RecordName, metav1.GetOptions{})
		}
	}
	if err != nil {
		return Status{}, err
	}
	if string(record.Data["username"]) != request.Username {
		return Status{}, ErrConflict
	}
	if string(record.Data["completed_at"]) != "" {
		return Status{}, ErrClosed
	}
	if bcrypt.CompareHashAndPassword(record.Data["password_hash"], []byte(request.Password)) != nil {
		return Status{}, ErrIdentity
	}

	u := &iamv1.User{}
	err = s.Resources.Get(ctx, client.ObjectKey{Name: request.Username}, u)
	createUser := apierrors.IsNotFound(err)
	if err != nil && !createUser {
		return Status{}, err
	}
	if !createUser && (u.GetDeletionTimestamp() != nil || u.GetAnnotations()[UserMarker] != "true") {
		return Status{}, ErrConflict
	}
	if createUser {
		// Recheck after acquiring the reservation, including after a crash.
		var list iamv1.UserList
		e := s.Resources.List(ctx, &list)
		if e != nil {
			return Status{}, e
		}
		if len(list.Items) != 0 {
			return Status{}, ErrConflict
		}
	}
	if err := s.initializeLocalName(ctx, request.Username); err != nil {
		return Status{}, err
	}
	if createUser {
		u = newUser(request)
		err = s.Resources.Create(ctx, u)
		if apierrors.IsAlreadyExists(err) {
			err = s.Resources.Get(ctx, client.ObjectKey{Name: request.Username}, u)
		}
	}
	if err != nil {
		return Status{}, err
	}
	if u.GetDeletionTimestamp() != nil || u.GetAnnotations()[UserMarker] != "true" {
		return Status{}, ErrConflict
	}
	if err := s.writeUserRelease(ctx, u); err != nil {
		return Status{}, err
	}
	return s.Status(ctx)
}

func (s *Service) writeUserRelease(ctx context.Context, u *iamv1.User) error {
	entries := map[string]string{"OLARES_NAME": u.Name + "@olares.local"}
	// An older Create request can still hold an empty DID after BFL binds it.
	// Leave the persisted DID intact until a non-empty value is available.
	if did := u.GetAnnotations()[UserDID]; did != "" {
		entries["OLARES_DID"] = did
	}
	updateRelease := s.updateRelease
	if updateRelease == nil {
		updateRelease = releasefile.Update
	}
	if err := updateRelease(ctx, entries); err != nil {
		return fmt.Errorf("write first user to release file: %w", err)
	}
	return nil
}

func (s *Service) initializeLocalName(ctx context.Context, username string) error {
	return retry.RetryOnConflict(retry.DefaultRetry, func() error {
		terminus := &sysv1.Terminus{}
		if err := s.Resources.Get(ctx, client.ObjectKey{Name: "terminus"}, terminus); err != nil {
			return err
		}
		if terminus.Spec.LocalName != "" {
			return nil
		}
		terminus.Spec.LocalName = username
		if err := s.Resources.Update(ctx, terminus); err != nil {
			return err
		}
		saved := &sysv1.Terminus{}
		if err := s.Resources.Get(ctx, client.ObjectKey{Name: "terminus"}, saved); err != nil {
			return err
		}
		if saved.Spec.LocalName != username {
			return errors.New("Terminus CR did not retain localName; check the CRD schema")
		}
		return nil
	})
}

func newUser(r CreateRequest) *iamv1.User {
	u := &iamv1.User{
		TypeMeta: metav1.TypeMeta{APIVersion: iamv1.SchemeGroupVersion.String(), Kind: "User"},
		ObjectMeta: metav1.ObjectMeta{Name: r.Username, Annotations: map[string]string{
			UserMarker:             "true",
			userWizardStatus:       WaitActivateVault,
			"bytetrade.io/creator": "cli", "bytetrade.io/owner-role": "admin",
			"bytetrade.io/terminus-name":  r.Username,
			"bytetrade.io/user-cpu-limit": r.CPU, "bytetrade.io/user-memory-limit": r.Memory,
			"bytetrade.io/launcher-auth-policy": "two_factor", "bytetrade.io/launcher-access-level": "1",
			"iam.kubesphere.io/uninitialized": "true", "iam.kubesphere.io/sync-to-lldap": "true",
			"iam.kubesphere.io/synced-to-lldap": "false", "iam.kubesphere.io/user-provider": "lldap",
		}},
		Spec: iamv1.UserSpec{
			DisplayName:     r.Username,
			Email:           r.Username + "@olares.local",
			InitialPassword: r.Password,
			Groups:          []string{"lldap_admin"},
		},
	}
	if r.DID != "" {
		u.Annotations[UserDID] = r.DID
	}
	return u
}

// Authorize authenticates this onboarding attempt with the initial salted MD5
// value provided by the client. The reservation stores its bcrypt hash.
func (s *Service) Authorize(ctx context.Context, username, password string) error {
	record, err := s.Kube.CoreV1().Secrets(Namespace).Get(ctx, RecordName, metav1.GetOptions{})
	if err != nil {
		return err
	}
	if string(record.Data["completed_at"]) != "" {
		return ErrClosed
	}
	if username != string(record.Data["username"]) || bcrypt.CompareHashAndPassword(record.Data["password_hash"], []byte(password)) != nil {
		return ErrIdentity
	}
	return nil
}

func (s *Service) Status(ctx context.Context) (Status, error) {
	record, err := s.Kube.CoreV1().Secrets(Namespace).Get(ctx, RecordName, metav1.GetOptions{})
	if apierrors.IsNotFound(err) {
		var list iamv1.UserList
		e := s.Resources.List(ctx, &list)
		if e != nil {
			return Status{}, e
		}
		if len(list.Items) != 0 {
			return Status{}, ErrConflict
		}
		return Status{State: Waiting}, nil
	}
	if err != nil {
		return Status{}, err
	}
	out := Status{State: Creating, Username: string(record.Data["username"])}
	if at := string(record.Data["completed_at"]); at != "" {
		out.State, out.CompletedAt = Completed, at
		return out, nil
	}
	u := &iamv1.User{}
	err = s.Resources.Get(ctx, client.ObjectKey{Name: out.Username}, u)
	if apierrors.IsNotFound(err) {
		out.Reason = "waiting for the first User CR"
		return out, nil
	}
	if err != nil {
		return Status{}, err
	}
	if u.GetDeletionTimestamp() != nil || u.GetAnnotations()[UserMarker] != "true" {
		return Status{}, ErrConflict
	}
	out.Reason = u.Status.Reason
	switch u.Status.State {
	case "Failed":
		out.State = Failed
	case "Created":
		activation := u.GetAnnotations()[userWizardStatus]
		if activation == Completed {
			// Binding may have supplied the DID after Create returned. Persist
			// it before closing onboarding, so a failed write can be retried.
			if err := s.writeUserRelease(ctx, u); err != nil {
				return Status{}, err
			}
			at, err := s.recordCompletion(ctx, record)
			if err != nil {
				return Status{}, err
			}
			out.State, out.CompletedAt, out.Reason = Completed, at, ""
			return out, nil
		}

		ready, err := s.userReady(ctx, out.Username)
		if err != nil {
			return Status{}, err
		}
		if !ready {
			out.Reason = "waiting for the first user services"
			return out, nil
		}
		out.Reason = ""
		switch activation {
		case "", WaitActivateVault:
			out.State = WaitActivateVault
		case waitResetPassword:
			out.State = WaitingPassword
		default:
			out.State = Failed
			out.Reason = fmt.Sprintf("unexpected user activation state: %s", activation)
		}
	}
	return out, nil
}

func (s *Service) userReady(ctx context.Context, name string) (bool, error) {
	bfl, err := s.Kube.AppsV1().StatefulSets("user-space-"+name).Get(ctx, "bfl", metav1.GetOptions{})
	if apierrors.IsNotFound(err) {
		return false, nil
	}
	if err != nil {
		return false, err
	}
	if bfl.DeletionTimestamp != nil || bfl.Spec.Replicas == nil || *bfl.Spec.Replicas < 1 || bfl.Status.ObservedGeneration < bfl.Generation || bfl.Status.ReadyReplicas < *bfl.Spec.Replicas {
		return false, nil
	}

	return true, nil
}

// Keep the machine's completion record after BFL finishes user activation,
// so restarting olaresd or later deleting the first user cannot reopen setup.
func (s *Service) recordCompletion(ctx context.Context, expected *corev1.Secret) (string, error) {
	var completedAt string
	err := retry.RetryOnConflict(retry.DefaultRetry, func() error {
		records := s.Kube.CoreV1().Secrets(Namespace)
		record, e := records.Get(ctx, RecordName, metav1.GetOptions{})
		if e != nil {
			return e
		}
		if record.UID != expected.UID || string(record.Data["username"]) != string(expected.Data["username"]) {
			return ErrConflict
		}
		completedAt = string(record.Data["completed_at"])
		if completedAt != "" {
			return nil
		}
		completedAt = time.Now().UTC().Format(time.RFC3339)
		record.Data["completed_at"] = []byte(completedAt)
		delete(record.Data, "password_hash")
		_, e = records.Update(ctx, record, metav1.UpdateOptions{})
		return e
	})
	return completedAt, err
}
