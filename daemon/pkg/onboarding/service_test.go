package onboarding

import (
	"context"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	sysv1 "github.com/beclab/api/api/sys.bytetrade.io/v1alpha1"
	iamv1 "github.com/beclab/api/iam/v1alpha2"
	appsv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	kubefake "k8s.io/client-go/kubernetes/fake"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/client/fake"
)

func newTestService(t *testing.T) (*Service, map[string]string) {
	t.Helper()
	scheme := runtime.NewScheme()
	if err := iamv1.AddToScheme(scheme); err != nil {
		t.Fatal(err)
	}
	if err := sysv1.AddToScheme(scheme); err != nil {
		t.Fatal(err)
	}
	resources := fake.NewClientBuilder().WithScheme(scheme).
		WithStatusSubresource(&iamv1.User{}).
		WithObjects(&sysv1.Terminus{ObjectMeta: metav1.ObjectMeta{Name: "terminus"}}).Build()
	replicas := int32(1)
	kube := kubefake.NewSimpleClientset(
		&appsv1.Deployment{
			ObjectMeta: metav1.ObjectMeta{Name: "bfl", Namespace: "os-frontend", Generation: 1},
			Spec:       appsv1.DeploymentSpec{Replicas: &replicas},
			Status:     appsv1.DeploymentStatus{ObservedGeneration: 1, ReadyReplicas: 1},
		},
	)
	release := make(map[string]string)
	return &Service{Kube: kube, Resources: resources, updateRelease: func(_ context.Context, entries map[string]string) error {
		for key, value := range entries {
			release[key] = value
		}
		return nil
	}}, release
}

func TestOnboardingWithDIDBoundAfterCreation(t *testing.T) {
	for _, tc := range []struct {
		name, initialDID, boundDID string
		retryCreate                bool
	}{
		{"no DID throughout", "", "", false},
		{"DID bound later without retry", "", "did:key:bound-later", false},
		{"DID changed during binding without retry", "did:key:initial", "did:key:bound-later", false},
		{"DID bound later with retry", "", "did:key:bound-later", true},
		{"DID changed during binding with retry", "did:key:initial", "did:key:bound-later", true},
	} {
		t.Run(tc.name, func(t *testing.T) {
			ctx := t.Context()
			s, release := newTestService(t)
			request := CreateRequest{Username: "owner", Password: "initial-password", DID: tc.initialDID}
			if status, err := s.Create(ctx, request); err != nil || status.State != Creating {
				t.Fatalf("create: status=%+v err=%v", status, err)
			}
			if release["OLARES_DID"] != tc.initialDID {
				t.Fatalf("initial release DID=%q, want %q", release["OLARES_DID"], tc.initialDID)
			}

			// Old daemons stored the creation-time DID in the reservation. It
			// remains stale when BFL binds a DID to the User later.
			records := s.Kube.CoreV1().Secrets(Namespace)
			record, err := records.Get(ctx, RecordName, metav1.GetOptions{})
			if err != nil {
				t.Fatal(err)
			}
			record.Data["did"] = []byte(tc.initialDID)
			if _, err := records.Update(ctx, record, metav1.UpdateOptions{}); err != nil {
				t.Fatal(err)
			}

			u := &iamv1.User{}
			if err := s.Resources.Get(ctx, client.ObjectKey{Name: request.Username}, u); err != nil {
				t.Fatal(err)
			}
			if tc.boundDID != "" {
				u.Annotations[UserDID] = tc.boundDID
			}
			if err := s.Resources.Update(ctx, u); err != nil {
				t.Fatal(err)
			}
			u.Status.State = "Created"
			if err := s.Resources.Status().Update(ctx, u); err != nil {
				t.Fatal(err)
			}
			if status, err := s.Status(ctx); err != nil || status.State != WaitActivateVault {
				t.Fatalf("status after DID binding: status=%+v err=%v", status, err)
			}

			// Retrying without a DID or with different optional metadata must
			// resume this user, and must preserve the DID already bound by BFL.
			if tc.retryCreate {
				for _, retryDID := range []string{"", "did:key:retry"} {
					request.DID = retryDID
					if status, err := s.Create(ctx, request); err != nil || status.State != WaitActivateVault {
						t.Fatalf("retry with DID %q: status=%+v err=%v", retryDID, status, err)
					}
					if release["OLARES_DID"] != tc.boundDID {
						t.Fatalf("retry overwrote the bound release DID: %q", release["OLARES_DID"])
					}
				}
			}

			if err := s.Resources.Get(ctx, client.ObjectKey{Name: request.Username}, u); err != nil {
				t.Fatal(err)
			}
			if u.Annotations[UserDID] != tc.boundDID {
				t.Fatal("retry overwrote the bound User DID")
			}
			u.Annotations[userWizardStatus] = Completed
			if err := s.Resources.Update(ctx, u); err != nil {
				t.Fatal(err)
			}
			handler := Handler(func(context.Context) (*Service, error) { return s, nil })
			response := httptest.NewRecorder()
			handler.ServeHTTP(response, httptest.NewRequest(http.MethodGet, "/onboarding", nil))
			if response.Code != http.StatusGone {
				t.Fatalf("completed onboarding status=%d body=%s", response.Code, response.Body.String())
			}
			if release["OLARES_DID"] != tc.boundDID {
				t.Fatalf("completion did not persist the bound DID: got %q, want %q", release["OLARES_DID"], tc.boundDID)
			}
			record, err = records.Get(ctx, RecordName, metav1.GetOptions{})
			if err != nil {
				t.Fatal(err)
			}
			if _, err := time.Parse(time.RFC3339, string(record.Data["completed_at"])); err != nil {
				t.Fatalf("completion was not persisted: %v", err)
			}
			if _, ok := record.Data["password_hash"]; ok {
				t.Fatal("completed onboarding retained the initial credentials")
			}
			if err := s.Resources.Delete(ctx, u); err != nil {
				t.Fatal(err)
			}
			restarted := &Service{Kube: s.Kube, Resources: s.Resources}
			if status, err := restarted.Status(ctx); err != nil || status.State != Completed {
				t.Fatalf("completion lost after restart and user deletion: status=%+v err=%v", status, err)
			}
			if _, err := restarted.Create(ctx, request); !errors.Is(err, ErrClosed) {
				t.Fatalf("completed onboarding reopened: %v", err)
			}
		})
	}
}

func TestCreateWithoutDIDDoesNotClearPersistedDID(t *testing.T) {
	s, release := newTestService(t)
	release["OLARES_DID"] = "did:key:already-bound"
	if _, err := s.Create(t.Context(), CreateRequest{Username: "owner", Password: "initial-password"}); err != nil {
		t.Fatal(err)
	}
	if release["OLARES_DID"] != "did:key:already-bound" {
		t.Fatalf("empty DID overwrote persisted DID: %q", release["OLARES_DID"])
	}
}

func TestCompletionRetriesFailedReleaseWrite(t *testing.T) {
	ctx := t.Context()
	s, release := newTestService(t)
	if _, err := s.Create(ctx, CreateRequest{Username: "owner", Password: "initial-password"}); err != nil {
		t.Fatal(err)
	}
	u := &iamv1.User{}
	if err := s.Resources.Get(ctx, client.ObjectKey{Name: "owner"}, u); err != nil {
		t.Fatal(err)
	}
	u.Annotations[UserDID] = "did:key:bound-later"
	u.Annotations[userWizardStatus] = Completed
	if err := s.Resources.Update(ctx, u); err != nil {
		t.Fatal(err)
	}
	u.Status.State = "Created"
	if err := s.Resources.Status().Update(ctx, u); err != nil {
		t.Fatal(err)
	}
	writeRelease := s.updateRelease
	writeErr := errors.New("release write failed")
	s.updateRelease = func(context.Context, map[string]string) error { return writeErr }
	if _, err := s.Status(ctx); !errors.Is(err, writeErr) {
		t.Fatalf("status swallowed release write error: %v", err)
	}
	record, err := s.Kube.CoreV1().Secrets(Namespace).Get(ctx, RecordName, metav1.GetOptions{})
	if err != nil {
		t.Fatal(err)
	}
	if len(record.Data["completed_at"]) != 0 || len(record.Data["password_hash"]) == 0 {
		t.Fatal("onboarding closed before the bound DID was persisted")
	}
	s.updateRelease = writeRelease
	if status, err := s.Status(ctx); err != nil || status.State != Completed {
		t.Fatalf("completion retry: status=%+v err=%v", status, err)
	}
	if release["OLARES_DID"] != "did:key:bound-later" {
		t.Fatalf("completion retry did not persist DID: %q", release["OLARES_DID"])
	}
}

func TestOptionalDIDDoesNotBypassReservationChecks(t *testing.T) {
	ctx := t.Context()
	s, _ := newTestService(t)
	request := CreateRequest{Username: "owner", Password: "initial-password"}
	if _, err := s.Create(ctx, request); err != nil {
		t.Fatal(err)
	}
	for _, tc := range []struct {
		name     string
		username string
		password string
		want     error
	}{
		{"another username", "another", request.Password, ErrConflict},
		{"another password", request.Username, "another-password", ErrIdentity},
	} {
		t.Run(tc.name, func(t *testing.T) {
			if _, err := s.Create(ctx, CreateRequest{Username: tc.username, Password: tc.password, DID: "did:key:changed"}); !errors.Is(err, tc.want) {
				t.Fatalf("create err=%v, want %v", err, tc.want)
			}
			if err := s.Authorize(ctx, tc.username, tc.password); !errors.Is(err, ErrIdentity) {
				t.Fatalf("authorize err=%v, want ErrIdentity", err)
			}
		})
	}
	for _, tc := range []struct {
		name     string
		modify   func(*iamv1.User)
		deleting bool
	}{
		{"unmarked user", func(u *iamv1.User) { delete(u.Annotations, UserMarker) }, false},
		{"deleting user", func(u *iamv1.User) {
			u.Finalizers = []string{"test/finalizer"}
		}, true},
	} {
		t.Run(tc.name, func(t *testing.T) {
			s, _ := newTestService(t)
			if _, err := s.Create(ctx, request); err != nil {
				t.Fatal(err)
			}
			u := &iamv1.User{}
			if err := s.Resources.Get(ctx, client.ObjectKey{Name: request.Username}, u); err != nil {
				t.Fatal(err)
			}
			tc.modify(u)
			if err := s.Resources.Update(ctx, u); err != nil {
				t.Fatal(err)
			}
			if tc.deleting {
				if err := s.Resources.Delete(ctx, u); err != nil {
					t.Fatal(err)
				}
			}
			if _, err := s.Create(ctx, request); !errors.Is(err, ErrConflict) {
				t.Fatalf("create err=%v, want ErrConflict", err)
			}
			if _, err := s.Status(ctx); !errors.Is(err, ErrConflict) {
				t.Fatalf("status err=%v, want ErrConflict", err)
			}
		})
	}
}

func TestRecordCompletionIgnoresDIDMetadataButChecksReservation(t *testing.T) {
	for _, tc := range []struct {
		name   string
		modify func(*corev1.Secret)
		want   error
	}{
		{"DID changed", func(s *corev1.Secret) { s.Data["did"] = []byte("did:key:bound-later") }, nil},
		{"record replaced", func(s *corev1.Secret) { s.UID = "another-record" }, ErrConflict},
		{"username changed", func(s *corev1.Secret) { s.Data["username"] = []byte("another") }, ErrConflict},
	} {
		t.Run(tc.name, func(t *testing.T) {
			expected := &corev1.Secret{
				ObjectMeta: metav1.ObjectMeta{Name: RecordName, Namespace: Namespace, UID: "reservation"},
				Data:       map[string][]byte{"username": []byte("owner"), "did": nil, "password_hash": []byte("hash")},
			}
			current := expected.DeepCopy()
			tc.modify(current)
			s := &Service{Kube: kubefake.NewSimpleClientset(current)}
			_, err := s.recordCompletion(t.Context(), expected)
			if !errors.Is(err, tc.want) {
				t.Fatalf("recordCompletion err=%v, want %v", err, tc.want)
			}
		})
	}
}
