package utils

import (
	"context"
	"testing"

	"github.com/beclab/Olares/framework/app-service/pkg/constants"
	appv1 "github.com/beclab/api/api/app.bytetrade.io/v1alpha1"
	"github.com/beclab/api/pkg/generated/clientset/versioned/fake"
	apierrors "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

func TestFindAppManagerPreservesSystemAppNameAndPrefersShared(t *testing.T) {
	ctx := context.Background()
	legacy := &appv1.ApplicationManager{ObjectMeta: metav1.ObjectMeta{Name: "a-legacy"}, Spec: appv1.ApplicationManagerSpec{AppName: "olares-app", AppNamespace: "user-space-alice", AppOwner: "alice"}}
	shared := &appv1.ApplicationManager{ObjectMeta: metav1.ObjectMeta{Name: "os-frontend-olares-app", Labels: map[string]string{constants.AppSharedLabel: "true"}}, Spec: appv1.ApplicationManagerSpec{AppName: "olares-app", AppNamespace: "os-frontend", AppOwner: "olares"}}
	c := fake.NewSimpleClientset(legacy, shared)
	for _, viewer := range []string{"alice", "bob", "olares"} {
		got, err := FindAppManager(ctx, c, "olares-app", viewer)
		if err != nil || got.Name != shared.Name || got.Spec.AppNamespace != "os-frontend" {
			t.Fatalf("viewer %s: %v, %v", viewer, got, err)
		}
	}
	if _, err := FindAppManager(ctx, c, "another-app", "alice"); !apierrors.IsNotFound(err) {
		t.Fatalf("unexpected error: %v", err)
	}
	c = fake.NewSimpleClientset(legacy)
	got, err := FindAppManager(ctx, c, "olares-app", "alice")
	if err != nil || got.Name != legacy.Name {
		t.Fatalf("legacy lookup: %v, %v", got, err)
	}
	if _, err := FindAppManager(ctx, c, "olares-app", "bob"); !apierrors.IsNotFound(err) {
		t.Fatalf("another user's app was exposed: %v", err)
	}
}

func TestAppNamespaceKeepsExplicitNamespace(t *testing.T) {
	t.Setenv("SYS_APPS", "olares-app")
	for _, namespace := range []string{"os-frontend", "user-space-alice", "user-system-alice"} {
		got, err := AppNamespace("olares-app", "alice", namespace)
		if err != nil || got != namespace {
			t.Fatalf("namespace %s: %s, %v", namespace, got, err)
		}
	}
}
