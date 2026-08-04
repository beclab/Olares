package appstate

import (
	"context"
	"errors"
	"testing"
	"time"

	appv1alpha1 "github.com/beclab/api/api/app.bytetrade.io/v1alpha1"
)

// Happy-path initialize: WaitForLaunch reports the app launched and the
// finally hook records Running.
func TestInitializingApp_LaunchSuccess_Running(t *testing.T) {
	am := buildAM("demo", appv1alpha1.App, appv1alpha1.Initializing, appv1alpha1.InstallOp,
		configJSON(t, "demo", false))
	c := newFakeClient(t, am.DeepCopy())
	deps, tf := newTestDeps(c)
	tf.helm.waitForLaunch = func() (bool, error) { return true, nil }

	app, serr := NewInitializingApp(deps, am, 0)
	if serr != nil {
		t.Fatalf("new initializing app: %v", serr)
	}
	if _, err := app.(OperationApp).Exec(context.TODO()); err != nil {
		t.Fatalf("exec: %v", err)
	}

	waitForState(t, c, "demo", appv1alpha1.Running, 5*time.Second)
}

// WaitForLaunch failure drives the app to InitializingCanceling.
func TestInitializingApp_LaunchFailure_Canceling(t *testing.T) {
	am := buildAM("demo", appv1alpha1.App, appv1alpha1.Initializing, appv1alpha1.InstallOp,
		configJSON(t, "demo", false))
	c := newFakeClient(t, am.DeepCopy())
	deps, tf := newTestDeps(c)
	tf.helm.waitForLaunch = func() (bool, error) { return false, errors.New("launch failed") }

	app, serr := NewInitializingApp(deps, am, 0)
	if serr != nil {
		t.Fatalf("new initializing app: %v", serr)
	}
	if _, err := app.(OperationApp).Exec(context.TODO()); err != nil {
		t.Fatalf("exec: %v", err)
	}

	waitForState(t, c, "demo", appv1alpha1.InitializingCanceling, 5*time.Second)
}
