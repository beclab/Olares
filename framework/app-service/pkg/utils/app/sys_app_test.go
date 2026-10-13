package app

import (
	"context"
	"errors"
	"reflect"
	"testing"

	"github.com/beclab/Olares/framework/app-service/pkg/constants"
	"github.com/beclab/Olares/framework/app-service/pkg/utils"
	appv1 "github.com/beclab/api/api/app.bytetrade.io/v1alpha1"
	"github.com/beclab/api/pkg/generated/clientset/versioned/fake"
	apierrors "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	ktesting "k8s.io/client-go/testing"
)

func TestCreateSysAppMgrUsesWorkloadNamespace(t *testing.T) {
	for _, tt := range []struct {
		name, namespace, owner string
		shared                 bool
	}{
		{"shared before onboarding", "os-frontend", "olares", true},
		{"legacy per-user", "user-space-alice", "alice", false},
	} {
		t.Run(tt.name, func(t *testing.T) {
			ctx := context.Background()
			c := fake.NewSimpleClientset() // No User objects, including no olares User.
			app := &appv1.Application{Spec: appv1.ApplicationSpec{Name: "olares-app", Namespace: tt.namespace, Owner: tt.owner}}
			if tt.shared {
				app.Labels = map[string]string{constants.AppApiVersionLabel: "v3", constants.AppSharedLabel: "true"}
			}
			if err := CreateSysAppMgr(ctx, c, app); err != nil {
				t.Fatal(err)
			}
			name := tt.namespace + "-olares-app"
			am, err := c.AppV1alpha1().ApplicationManagers().Get(ctx, name, metav1.GetOptions{})
			if err != nil {
				t.Fatal(err)
			}
			if am.Spec.AppNamespace != tt.namespace || am.Spec.AppOwner != tt.owner || am.Spec.RawAppName != "olares-app" || am.Spec.Source != "system" || am.Status.State != appv1.Running {
				t.Fatalf("unexpected manager: %+v", am)
			}
			if !reflect.DeepEqual(am.Labels, app.Labels) {
				t.Fatalf("labels = %v, want %v", am.Labels, app.Labels)
			}
			if tt.shared {
				// Any viewer resolves the platform registration, not user-space-<viewer>.
				found, err := utils.FindAppManager(ctx, c, "olares-app", "bob")
				if err != nil || found.Name != name {
					t.Fatalf("lookup: %v, %v", found, err)
				}
			}
			actions := len(c.Actions())
			if err := CreateSysAppMgr(ctx, c, app); err != nil {
				t.Fatal(err)
			}
			for _, action := range c.Actions()[actions:] {
				if action.GetVerb() != "get" {
					t.Fatalf("repeated registration wrote: %s", action.GetVerb())
				}
			}
		})
	}
}

func TestCreateSysAppMgrRepairsLabelsWithoutResettingOperation(t *testing.T) {
	ctx := context.Background()
	am := &appv1.ApplicationManager{
		ObjectMeta: metav1.ObjectMeta{Name: "os-frontend-olares-app", Labels: map[string]string{"keep": "yes"}},
		Spec:       appv1.ApplicationManagerSpec{AppName: "olares-app", AppOwner: "olares", AppNamespace: "os-frontend", OpType: appv1.UpgradeOp},
		Status:     appv1.ApplicationManagerStatus{State: appv1.Running, OpGeneration: 7, Message: "preserve"},
	}
	c := fake.NewSimpleClientset(am)
	app := &appv1.Application{ObjectMeta: metav1.ObjectMeta{Labels: map[string]string{constants.AppSharedLabel: "true", constants.AppApiVersionLabel: "v3"}}, Spec: appv1.ApplicationSpec{Name: "olares-app", Owner: "olares", Namespace: "os-frontend"}}
	// Exercise optimistic-lock retries during reconciliation.
	conflict := true
	c.PrependReactor("update", "applicationmanagers", func(action ktesting.Action) (bool, runtime.Object, error) {
		if conflict {
			conflict = false
			return true, nil, apierrors.NewConflict(appv1.Resource("applicationmanagers"), am.Name, errors.New("concurrent update"))
		}
		return false, nil, nil
	})
	if err := CreateSysAppMgr(ctx, c, app); err != nil {
		t.Fatal(err)
	}
	got, err := c.AppV1alpha1().ApplicationManagers().Get(ctx, am.Name, metav1.GetOptions{})
	if err != nil {
		t.Fatal(err)
	}
	if got.Labels[constants.AppSharedLabel] != "true" || got.Labels["keep"] != "yes" {
		t.Fatalf("labels: %v", got.Labels)
	}
	if !reflect.DeepEqual(got.Status, am.Status) || got.Spec.OpType != am.Spec.OpType {
		t.Fatal("existing lifecycle state was reset")
	}
}

func TestCreateSysAppMgrReturnsAPIError(t *testing.T) {
	c := fake.NewSimpleClientset()
	want := errors.New("API unavailable")
	c.PrependReactor("create", "applicationmanagers", func(ktesting.Action) (bool, runtime.Object, error) { return true, nil, want })
	app := &appv1.Application{Spec: appv1.ApplicationSpec{Name: "olares-app", Owner: "olares", Namespace: "os-frontend"}}
	if err := CreateSysAppMgr(context.Background(), c, app); !errors.Is(err, want) {
		t.Fatalf("error = %v", err)
	}
}

func TestSharedFrontendNamespaceIsProtected(t *testing.T) {
	if !IsProtectedNamespace("os-frontend") {
		t.Fatal("shared frontend namespace must be protected from app cleanup")
	}
}
