package appstate

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/beclab/Olares/framework/app-service/pkg/compute/validation"
	appv1alpha1 "github.com/beclab/api/api/app.bytetrade.io/v1alpha1"
)

// Happy path (single-phase, no workloadReplicas): validation accepts, helm
// installs, scale-up and startup succeed, app advances to Initializing.
func TestInstallingApp_Success_Initializing(t *testing.T) {
	am := buildAM("demo", appv1alpha1.App, appv1alpha1.Installing, appv1alpha1.InstallOp,
		configJSON(t, "demo", false))
	c := newFakeClient(t, am.DeepCopy())
	deps, tf := newTestDeps(c)

	app, serr := NewInstallingApp(deps, am, 0)
	if serr != nil {
		t.Fatalf("new installing app: %v", serr)
	}
	if _, err := app.(OperationApp).Exec(context.TODO()); err != nil {
		t.Fatalf("exec: %v", err)
	}

	waitForState(t, c, "demo", appv1alpha1.Initializing, 5*time.Second)
	if tf.helm.installCalls != 1 {
		t.Fatalf("expected Install called once, got %d", tf.helm.installCalls)
	}
	if len(tf.helm.scaleCalls) != 1 || tf.helm.scaleCalls[0] != -1 {
		t.Fatalf("expected Scale(-1) once, got %v", tf.helm.scaleCalls)
	}
}

// SetExposePorts failing aborts Exec synchronously with an error and no state
// change (it runs before any helm work).
func TestInstallingApp_SetExposePortsFailure(t *testing.T) {
	am := buildAM("demo", appv1alpha1.App, appv1alpha1.Installing, appv1alpha1.InstallOp,
		configJSON(t, "demo", false))
	c := newFakeClient(t, am.DeepCopy())
	deps, tf := newTestDeps(c)
	tf.setExposePortsErr = errors.New("port taken")

	app, serr := NewInstallingApp(deps, am, 0)
	if serr != nil {
		t.Fatalf("new installing app: %v", serr)
	}
	if _, err := app.(OperationApp).Exec(context.TODO()); err == nil {
		t.Fatalf("expected exec error on SetExposePorts failure")
	}

	if got := getAM(t, c, "demo").Status.State; got != appv1alpha1.Installing {
		t.Fatalf("state = %q, want unchanged %q", got, appv1alpha1.Installing)
	}
}

// Pre-helm validation rejection drives the app to InstallFailed.
func TestInstallingApp_ValidationRejected_InstallFailed(t *testing.T) {
	am := buildAM("demo", appv1alpha1.App, appv1alpha1.Installing, appv1alpha1.InstallOp,
		configJSON(t, "demo", false))
	c := newFakeClient(t, am.DeepCopy())
	deps, tf := newTestDeps(c)
	tf.validation = func(ctx context.Context, in validation.Input) (validation.Decision, error) {
		return validation.Decision{OK: false, Validator: "k8s-request", Message: "not enough memory"}, nil
	}

	app, serr := NewInstallingApp(deps, am, 0)
	if serr != nil {
		t.Fatalf("new installing app: %v", serr)
	}
	if _, err := app.(OperationApp).Exec(context.TODO()); err != nil {
		t.Fatalf("exec: %v", err)
	}

	waitForState(t, c, "demo", appv1alpha1.InstallFailed, 5*time.Second)
	if tf.helm.installCalls != 0 {
		t.Fatalf("helm Install should not run when validation rejects, got %d", tf.helm.installCalls)
	}
}

// A failing helm Install drives the app to InstallFailed.
func TestInstallingApp_HelmInstallFailure_InstallFailed(t *testing.T) {
	am := buildAM("demo", appv1alpha1.App, appv1alpha1.Installing, appv1alpha1.InstallOp,
		configJSON(t, "demo", false))
	c := newFakeClient(t, am.DeepCopy())
	deps, tf := newTestDeps(c)
	tf.helm.installErr = errors.New("boom")

	app, serr := NewInstallingApp(deps, am, 0)
	if serr != nil {
		t.Fatalf("new installing app: %v", serr)
	}
	if _, err := app.(OperationApp).Exec(context.TODO()); err != nil {
		t.Fatalf("exec: %v", err)
	}

	waitForState(t, c, "demo", appv1alpha1.InstallFailed, 5*time.Second)
}

// Workloads failing to come up after scale-up routes to Stopping.
func TestInstallingApp_StartupFailure_Stopping(t *testing.T) {
	am := buildAM("demo", appv1alpha1.App, appv1alpha1.Installing, appv1alpha1.InstallOp,
		configJSON(t, "demo", false))
	c := newFakeClient(t, am.DeepCopy())
	deps, tf := newTestDeps(c)
	tf.helm.waitForStartUp = func() (bool, error) { return false, errors.New("pods pending") }

	app, serr := NewInstallingApp(deps, am, 0)
	if serr != nil {
		t.Fatalf("new installing app: %v", serr)
	}
	if _, err := app.(OperationApp).Exec(context.TODO()); err != nil {
		t.Fatalf("exec: %v", err)
	}

	waitForState(t, c, "demo", appv1alpha1.Stopping, 5*time.Second)
}

// Middleware install waits for launch and lands Running.
func TestInstallingApp_Middleware_Running(t *testing.T) {
	am := buildAM("demo", appv1alpha1.Middleware, appv1alpha1.Installing, appv1alpha1.InstallOp,
		configJSON(t, "demo", false))
	c := newFakeClient(t, am.DeepCopy())
	deps, tf := newTestDeps(c)
	tf.helm.waitForLaunch = func() (bool, error) { return true, nil }

	app, serr := NewInstallingApp(deps, am, 0)
	if serr != nil {
		t.Fatalf("new installing app: %v", serr)
	}
	if _, err := app.(OperationApp).Exec(context.TODO()); err != nil {
		t.Fatalf("exec: %v", err)
	}

	waitForState(t, c, "demo", appv1alpha1.Running, 5*time.Second)
}
