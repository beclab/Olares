package k8sstore_test

import (
	"context"
	"testing"

	appv1 "github.com/beclab/api/api/app.bytetrade.io/v1alpha1"

	apierrors "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	"github.com/beclab/Olares/framework/app-service/pkg/applifecycle/adapters/k8sstore"
	"github.com/beclab/Olares/framework/app-service/pkg/testutil"
)

func TestGet_Found(t *testing.T) {
	ctx := context.Background()
	am := &appv1.ApplicationManager{
		ObjectMeta: metav1.ObjectMeta{Name: "user-app-a"},
		Spec:       appv1.ApplicationManagerSpec{AppName: "app-a", AppOwner: "user"},
	}
	c := testutil.NewFakeClient(am)
	s := k8sstore.New(c, k8sstore.WithShadowMode(true))

	got, err := s.Get(ctx, "user-app-a")
	if err != nil {
		t.Fatalf("Get: %v", err)
	}
	if got == nil || got.Name != "user-app-a" {
		t.Fatalf("Get returned %+v", got)
	}
	if got.Spec.AppName != "app-a" || got.Spec.AppOwner != "user" {
		t.Fatalf("Get spec mismatch: %+v", got.Spec)
	}
	// Fake client assigns a non-empty ResourceVersion on Create.
	// Assert it flows through so callers depending on the RV
	// pre-condition path (PatchStatus / PatchSpecAndAnnotations)
	// are not fed an empty precondition.
	if got.ResourceVersion == "" {
		t.Fatalf("Get returned empty ResourceVersion")
	}
}

func TestGet_NotFound(t *testing.T) {
	ctx := context.Background()
	c := testutil.NewFakeClient()
	s := k8sstore.New(c)

	_, err := s.Get(ctx, "missing")
	if err == nil {
		t.Fatal("Get: want error, got nil")
	}
	if !apierrors.IsNotFound(err) {
		t.Fatalf("Get: want NotFound, got %T %v", err, err)
	}
}

func TestList(t *testing.T) {
	ctx := context.Background()
	c := testutil.NewFakeClient(
		&appv1.ApplicationManager{ObjectMeta: metav1.ObjectMeta{Name: "a"}},
		&appv1.ApplicationManager{ObjectMeta: metav1.ObjectMeta{Name: "b"}},
		&appv1.ApplicationManager{ObjectMeta: metav1.ObjectMeta{Name: "c"}},
	)
	s := k8sstore.New(c)

	got, err := s.List(ctx)
	if err != nil {
		t.Fatalf("List: %v", err)
	}
	if len(got) != 3 {
		t.Fatalf("List: want 3, got %d", len(got))
	}
	// Order isn't guaranteed by the API contract, so build a set.
	names := map[string]bool{}
	for _, am := range got {
		names[am.Name] = true
	}
	for _, want := range []string{"a", "b", "c"} {
		if !names[want] {
			t.Errorf("List missing %q, got %v", want, names)
		}
	}
}

func TestList_Empty(t *testing.T) {
	ctx := context.Background()
	c := testutil.NewFakeClient()
	s := k8sstore.New(c)

	got, err := s.List(ctx)
	if err != nil {
		t.Fatalf("List: %v", err)
	}
	if len(got) != 0 {
		t.Fatalf("List: want 0, got %d", len(got))
	}
}

func TestGetApplication_Found(t *testing.T) {
	ctx := context.Background()
	app := &appv1.Application{
		ObjectMeta: metav1.ObjectMeta{Name: "user-app-a"},
		Spec:       appv1.ApplicationSpec{Name: "app-a", Owner: "user"},
	}
	c := testutil.NewFakeClient(app)
	s := k8sstore.New(c)

	got, err := s.GetApplication(ctx, "user-app-a")
	if err != nil {
		t.Fatalf("GetApplication: %v", err)
	}
	if got == nil || got.Name != "user-app-a" {
		t.Fatalf("GetApplication: %+v", got)
	}
}

func TestGetApplication_NotFound(t *testing.T) {
	ctx := context.Background()
	c := testutil.NewFakeClient()
	s := k8sstore.New(c)

	_, err := s.GetApplication(ctx, "missing")
	if err == nil {
		t.Fatal("GetApplication: want error, got nil")
	}
	if !apierrors.IsNotFound(err) {
		t.Fatalf("GetApplication: want NotFound, got %T %v", err, err)
	}
}

func TestNew_NilClientPanics(t *testing.T) {
	defer func() {
		if r := recover(); r == nil {
			t.Fatal("New(nil): want panic, got nil")
		}
	}()
	_ = k8sstore.New(nil)
}
