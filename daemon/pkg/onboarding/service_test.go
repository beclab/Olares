package onboarding

import (
	"context"
	"errors"
	"sync"
	"testing"

	appsv1 "k8s.io/api/apps/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/runtime/schema"
	dynamicfake "k8s.io/client-go/dynamic/fake"
	"k8s.io/client-go/kubernetes/fake"
	ktesting "k8s.io/client-go/testing"
	"k8s.io/utils/ptr"
)

func testService() *Service {
	return &Service{Kube: fake.NewSimpleClientset(), Dynamic: dynamicfake.NewSimpleDynamicClientWithCustomListKinds(
		runtime.NewScheme(), map[schema.GroupVersionResource]string{Users: "UserList"}),
		ResetPassword: func(context.Context, string, string) error { return nil },
	}
}

func request() CreateRequest {
	return CreateRequest{Username: "alice", DID: "did:key:alice", Password: "initial-password"}
}

func create(t *testing.T, s *Service) {
	t.Helper()
	status, err := s.Create(context.Background(), request())
	if err != nil || status.State != Creating {
		t.Fatalf("create: %+v, %v", status, err)
	}
}

func ready(t *testing.T, s *Service) {
	t.Helper()
	ctx := context.Background()
	u, err := s.Dynamic.Resource(Users).Get(ctx, "alice", metav1.GetOptions{})
	if err != nil {
		t.Fatal(err)
	}
	_ = unstructured.SetNestedField(u.Object, "Created", "status", "state")
	if _, err = s.Dynamic.Resource(Users).Update(ctx, u, metav1.UpdateOptions{}); err != nil {
		t.Fatal(err)
	}
	_, err = s.Kube.AppsV1().StatefulSets("user-space-alice").Create(ctx, &appsv1.StatefulSet{
		ObjectMeta: metav1.ObjectMeta{Name: "bfl", Generation: 1}, Spec: appsv1.StatefulSetSpec{Replicas: ptr.To(int32(1))},
		Status: appsv1.StatefulSetStatus{ObservedGeneration: 1, ReadyReplicas: 1},
	}, metav1.CreateOptions{})
	if err != nil {
		t.Fatal(err)
	}
	_, err = s.Kube.AppsV1().Deployments("user-system-alice").Create(ctx, &appsv1.Deployment{
		ObjectMeta: metav1.ObjectMeta{Name: "system-server", Generation: 1}, Spec: appsv1.DeploymentSpec{Replicas: ptr.To(int32(1))},
		Status: appsv1.DeploymentStatus{ObservedGeneration: 1, ReadyReplicas: 1},
	}, metav1.CreateOptions{})
	if err != nil {
		t.Fatal(err)
	}
}

func TestFirstAdminWithoutDomain(t *testing.T) {
	s := testService()
	create(t, s)
	u, err := s.Dynamic.Resource(Users).Get(context.Background(), "alice", metav1.GetOptions{})
	if err != nil {
		t.Fatal(err)
	}
	a := u.GetAnnotations()
	if a["bytetrade.io/owner-role"] != "admin" || a["bytetrade.io/creator"] != "cli" || a[UserDID] != request().DID {
		t.Fatalf("annotations: %v", a)
	}
	if a["bytetrade.io/terminus-name"] != request().Username {
		t.Fatal("terminus-name must be the local username without a domain")
	}
	for _, key := range []string{"bytetrade.io/zone", "bytetrade.io/wizard-status", "bytetrade.io/is-ephemeral"} {
		if _, ok := a[key]; ok {
			t.Fatalf("unexpected annotation %s", key)
		}
	}
	if _, exists, _ := unstructured.NestedFieldNoCopy(u.Object, "status"); exists {
		t.Fatal("controller must own initial status")
	}
	if email, exists, _ := unstructured.NestedString(u.Object, "spec", "email"); !exists || email == "" {
		t.Fatal("User CR requires an email for the local account")
	}
	password, _, _ := unstructured.NestedString(u.Object, "spec", "initialPassword")
	if password != passwordWire(request().Password) || password == request().Password {
		t.Fatal("User CR password does not match the login wire format")
	}
	groups, _, _ := unstructured.NestedStringSlice(u.Object, "spec", "groups")
	if len(groups) != 1 || groups[0] != "lldap_admin" {
		t.Fatal("admin is missing its LLDAP group")
	}
}

func TestConcurrentFirstUsersHaveOneWinner(t *testing.T) {
	s := testService()
	other := &Service{Kube: s.Kube, Dynamic: s.Dynamic}
	requests := []CreateRequest{request(), {Username: "bob", Password: "initial-password", DID: "did:key:bob"}}
	services := []*Service{s, other}
	errs := make([]error, 2)
	var wg sync.WaitGroup
	for i := range requests {
		wg.Add(1)
		go func(i int) { defer wg.Done(); _, errs[i] = services[i].Create(context.Background(), requests[i]) }(i)
	}
	wg.Wait()
	if (errs[0] == nil) == (errs[1] == nil) {
		t.Fatalf("expected exactly one winner: %v", errs)
	}
	list, _ := s.Dynamic.Resource(Users).List(context.Background(), metav1.ListOptions{})
	if len(list.Items) != 1 {
		t.Fatalf("created %d users", len(list.Items))
	}
}

func TestReservationSurvivesRestartAndRejectsAnotherIdentity(t *testing.T) {
	s := testService()
	create(t, s)
	s = &Service{Kube: s.Kube, Dynamic: s.Dynamic}
	if _, err := s.Create(context.Background(), request()); err != nil {
		t.Fatal(err)
	}
	for _, r := range []CreateRequest{
		{Username: "bob", DID: request().DID, Password: request().Password},
		{Username: "alice", DID: "did:key:bob", Password: request().Password},
		{Username: "alice", DID: request().DID, Password: "wrong-password"},
	} {
		if _, err := s.Create(context.Background(), r); err == nil {
			t.Fatal("accepted a different attempt")
		}
	}
	list, _ := s.Dynamic.Resource(Users).List(context.Background(), metav1.ListOptions{})
	if len(list.Items) != 1 {
		t.Fatalf("created %d users", len(list.Items))
	}
}

func TestCreateResumesAfterUserWriteFails(t *testing.T) {
	s := testService()
	d := s.Dynamic.(*dynamicfake.FakeDynamicClient)
	failed := false
	d.PrependReactor("create", "users", func(ktesting.Action) (bool, runtime.Object, error) {
		if !failed {
			failed = true
			return true, nil, errors.New("apiserver unavailable")
		}
		return false, nil, nil
	})
	if _, err := s.Create(context.Background(), request()); err == nil {
		t.Fatal("expected write failure")
	}
	s = &Service{Kube: s.Kube, Dynamic: s.Dynamic}
	create(t, s)
}

func TestExistingUserClosesFirstUserCreation(t *testing.T) {
	s := testService()
	_, err := s.Dynamic.Resource(Users).Create(context.Background(), newUser(request()), metav1.CreateOptions{})
	if err != nil {
		t.Fatal(err)
	}
	if _, err = s.Create(context.Background(), request()); !errors.Is(err, ErrConflict) {
		t.Fatalf("got %v", err)
	}
}

func TestCreatedDoesNotMeanReady(t *testing.T) {
	s := testService()
	create(t, s)
	ready(t, s)
	ctx := context.Background()
	bfl, _ := s.Kube.AppsV1().StatefulSets("user-space-alice").Get(ctx, "bfl", metav1.GetOptions{})
	bfl.Status.ReadyReplicas = 0
	_, _ = s.Kube.AppsV1().StatefulSets("user-space-alice").UpdateStatus(ctx, bfl, metav1.UpdateOptions{})
	status, err := s.Status(ctx)
	if err != nil || status.State != Created {
		t.Fatalf("status %+v, %v", status, err)
	}
	if _, err = s.Reset(ctx, "alice", request().Password, "new-password"); !errors.Is(err, ErrNotReady) {
		t.Fatalf("got %v", err)
	}
}

func TestPasswordResetCommitsOnlyAfterBackendSuccess(t *testing.T) {
	s := testService()
	create(t, s)
	ready(t, s)
	ctx := context.Background()
	s.ResetPassword = func(context.Context, string, string) error { return errors.New("backend unavailable") }
	if _, err := s.Reset(ctx, "alice", request().Password, "new-password"); err == nil {
		t.Fatal("expected failure")
	}
	status, _ := s.Status(ctx)
	if status.State != WaitingPassword {
		t.Fatalf("status %+v", status)
	}
	s.ResetPassword = func(_ context.Context, user, password string) error {
		if user != "alice" || password != "new-password" {
			t.Fatal("wrong reset target")
		}
		return nil
	}
	if _, err := s.Reset(ctx, "alice", "wrong-password", "new-password"); !errors.Is(err, ErrIdentity) {
		t.Fatalf("got %v", err)
	}
	status, err := s.Reset(ctx, "alice", request().Password, "new-password")
	if err != nil || status.State != Completed || status.CompletedAt == "" {
		t.Fatalf("reset %+v, %v", status, err)
	}
	record, _ := s.Kube.CoreV1().Secrets(Namespace).Get(ctx, RecordName, metav1.GetOptions{})
	if len(record.Data["password_hash"]) != 0 {
		t.Fatal("initial password hash retained")
	}
	if _, err = s.Create(ctx, request()); !errors.Is(err, ErrClosed) {
		t.Fatalf("got %v", err)
	}
	s = &Service{Kube: s.Kube, Dynamic: s.Dynamic}
	status, err = s.Status(ctx)
	if err != nil || status.State != Completed {
		t.Fatalf("restart %+v %v", status, err)
	}
}

func TestFailureReasonIsReported(t *testing.T) {
	s := testService()
	create(t, s)
	u, _ := s.Dynamic.Resource(Users).Get(context.Background(), "alice", metav1.GetOptions{})
	_ = unstructured.SetNestedField(u.Object, "Failed", "status", "state")
	_ = unstructured.SetNestedField(u.Object, "user with owner role not found", "status", "reason")
	_, _ = s.Dynamic.Resource(Users).Update(context.Background(), u, metav1.UpdateOptions{})
	status, err := s.Status(context.Background())
	if err != nil || status.State != Failed || status.Reason != "user with owner role not found" {
		t.Fatalf("status %+v %v", status, err)
	}
}

func TestClaimContainsNoPlaintextPassword(t *testing.T) {
	s := testService()
	create(t, s)
	record, _ := s.Kube.CoreV1().Secrets(Namespace).Get(context.Background(), RecordName, metav1.GetOptions{})
	for _, value := range record.Data {
		if string(value) == request().Password {
			t.Fatal("plaintext password persisted in onboarding record")
		}
	}
}
