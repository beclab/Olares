package onboarding

import (
	"context"
	"errors"
	"fmt"
	"strings"
	"sync"
	"time"

	"golang.org/x/crypto/bcrypt"

	corev1 "k8s.io/api/core/v1"
	apierrors "k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/api/resource"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"k8s.io/apimachinery/pkg/util/validation"
	"k8s.io/client-go/dynamic"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/util/retry"
)

const (
	Namespace       = "os-framework"
	RecordName      = "olares-onboarding"
	UserMarker      = "bytetrade.io/onboarding"
	UserDID         = "bytetrade.io/user-did"
	Waiting         = "waiting-for-user"
	Creating        = "creating"
	Created         = "created"
	WaitingPassword = "waiting-for-reset-password"
	Completed       = "completed"
	Failed          = "failed"
)

var (
	Users                = schema.GroupVersionResource{Group: "iam.kubesphere.io", Version: "v1alpha2", Resource: "users"}
	ErrConflict          = errors.New("first user is already reserved or exists")
	ErrIdentity          = errors.New("request identity does not match the first user")
	ErrClosed            = errors.New("onboarding is complete")
	ErrNotReady          = errors.New("first user services are not ready")
	ErrPasswordUnchanged = errors.New("new password must differ from the initial password")
)

type CreateRequest struct {
	Username string `json:"username"`
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
	Dynamic       dynamic.Interface
	ResetPassword func(context.Context, string, string) error
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
	if !strings.HasPrefix(r.DID, "did:") || len(r.DID) > 1024 || strings.ContainsAny(r.DID, " \t\r\n") {
		return errors.New("invalid did")
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
// by the same DID and password resume this reservation; another DID cannot take it.
// The Secret survives daemon restarts and is removed by a full uninstall.
func (s *Service) Create(ctx context.Context, request CreateRequest) (Status, error) {
	s.mu.Lock()
	defer s.mu.Unlock()
	if err := request.validate(); err != nil {
		return Status{}, err
	}
	did := request.DID
	records := s.Kube.CoreV1().Secrets(Namespace)
	record, err := records.Get(ctx, RecordName, metav1.GetOptions{})
	if apierrors.IsNotFound(err) {
		users, e := s.Dynamic.Resource(Users).List(ctx, metav1.ListOptions{})
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
			Data:       map[string][]byte{"username": []byte(request.Username), "did": []byte(did), "password_hash": hash},
		}, metav1.CreateOptions{})
		if apierrors.IsAlreadyExists(err) {
			record, err = records.Get(ctx, RecordName, metav1.GetOptions{})
		}
	}
	if err != nil {
		return Status{}, err
	}
	if string(record.Data["did"]) != did || string(record.Data["username"]) != request.Username {
		return Status{}, ErrConflict
	}
	if string(record.Data["completed_at"]) != "" {
		return Status{}, ErrClosed
	}
	if bcrypt.CompareHashAndPassword(record.Data["password_hash"], []byte(request.Password)) != nil {
		return Status{}, ErrIdentity
	}

	users := s.Dynamic.Resource(Users)
	u, err := users.Get(ctx, request.Username, metav1.GetOptions{})
	if apierrors.IsNotFound(err) {
		// Recheck after acquiring the reservation, including after a crash.
		list, e := users.List(ctx, metav1.ListOptions{})
		if e != nil {
			return Status{}, e
		}
		if len(list.Items) != 0 {
			return Status{}, ErrConflict
		}
		u, err = users.Create(ctx, newUser(request), metav1.CreateOptions{})
		if apierrors.IsAlreadyExists(err) {
			u, err = users.Get(ctx, request.Username, metav1.GetOptions{})
		}
	}
	if err != nil {
		return Status{}, err
	}
	if u.GetDeletionTimestamp() != nil || u.GetAnnotations()[UserMarker] != "true" || u.GetAnnotations()[UserDID] != did {
		return Status{}, ErrConflict
	}
	return s.Status(ctx)
}

func newUser(r CreateRequest) *unstructured.Unstructured {
	return &unstructured.Unstructured{Object: map[string]interface{}{
		"apiVersion": "iam.kubesphere.io/v1alpha2", "kind": "User",
		"metadata": map[string]interface{}{"name": r.Username, "annotations": map[string]interface{}{
			UserMarker: "true", UserDID: r.DID,
			"bytetrade.io/creator": "cli", "bytetrade.io/owner-role": "admin",
			"bytetrade.io/terminus-name":  r.Username,
			"bytetrade.io/user-cpu-limit": r.CPU, "bytetrade.io/user-memory-limit": r.Memory,
			"bytetrade.io/launcher-auth-policy": "two_factor", "bytetrade.io/launcher-access-level": "1",
			"iam.kubesphere.io/uninitialized": "true", "iam.kubesphere.io/sync-to-lldap": "true",
			"iam.kubesphere.io/synced-to-lldap": "false", "iam.kubesphere.io/user-provider": "lldap",
		}},
		"spec": map[string]interface{}{
			"displayName":     r.Username,
			"email":           r.Username + "@olares.local",
			"initialPassword": passwordWire(r.Password),
			"groups":          []interface{}{"lldap_admin"},
		},
	}}
}

// Authorize authenticates this onboarding attempt with its initial password.
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
		list, e := s.Dynamic.Resource(Users).List(ctx, metav1.ListOptions{})
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
	u, err := s.Dynamic.Resource(Users).Get(ctx, out.Username, metav1.GetOptions{})
	if apierrors.IsNotFound(err) {
		out.Reason = "waiting for the first User CR"
		return out, nil
	}
	if err != nil {
		return Status{}, err
	}
	if u.GetDeletionTimestamp() != nil || u.GetAnnotations()[UserMarker] != "true" || u.GetAnnotations()[UserDID] != string(record.Data["did"]) {
		return Status{}, ErrConflict
	}
	phase, _, _ := unstructured.NestedString(u.Object, "status", "state")
	out.Reason, _, _ = unstructured.NestedString(u.Object, "status", "reason")
	switch phase {
	case "Failed":
		out.State = Failed
	case "Created":
		out.State = Created
		ready, err := s.userReady(ctx, out.Username)
		if err != nil {
			return Status{}, err
		}
		if ready {
			out.State, out.Reason = WaitingPassword, ""
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
	server, err := s.Kube.AppsV1().Deployments("user-system-"+name).Get(ctx, "system-server", metav1.GetOptions{})
	if apierrors.IsNotFound(err) {
		return false, nil
	}
	if err != nil {
		return false, err
	}
	return server.DeletionTimestamp == nil && server.Spec.Replicas != nil && *server.Spec.Replicas > 0 && server.Status.ObservedGeneration >= server.Generation && server.Status.ReadyReplicas >= *server.Spec.Replicas, nil
}

func (s *Service) Reset(ctx context.Context, username, initialPassword, password string) (Status, error) {
	s.mu.Lock()
	defer s.mu.Unlock()
	if err := s.Authorize(ctx, username, initialPassword); err != nil {
		return Status{}, err
	}
	if len(password) < 8 || len(password) > 72 {
		return Status{}, errors.New("password must contain 8 to 72 bytes")
	}
	status, err := s.Status(ctx)
	if err != nil {
		return Status{}, err
	}
	if status.State == Completed {
		return status, nil
	}
	if status.State != WaitingPassword {
		return Status{}, ErrNotReady
	}
	if password == initialPassword {
		return Status{}, ErrPasswordUnchanged
	}
	if s.ResetPassword == nil {
		return Status{}, errors.New("password reset backend is unavailable")
	}
	if err := s.ResetPassword(ctx, status.Username, password); err != nil {
		return Status{}, fmt.Errorf("password reset failed: %w", err)
	}
	// A backend failure must never mark onboarding complete. Repeating a reset
	// after a lost response sets the same password and repairs this marker.
	err = retry.RetryOnConflict(retry.DefaultRetry, func() error {
		records := s.Kube.CoreV1().Secrets(Namespace)
		record, e := records.Get(ctx, RecordName, metav1.GetOptions{})
		if e != nil {
			return e
		}
		if string(record.Data["username"]) != status.Username {
			return ErrIdentity
		}
		record.Data["completed_at"] = []byte(time.Now().UTC().Format(time.RFC3339))
		delete(record.Data, "password_hash")
		_, e = records.Update(ctx, record, metav1.UpdateOptions{})
		return e
	})
	if err != nil {
		return Status{}, err
	}
	return s.Status(ctx)
}
