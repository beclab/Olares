package appstate

import (
	"context"
	"errors"
	"testing"
	"time"

	appv1alpha1 "github.com/beclab/api/api/app.bytetrade.io/v1alpha1"
)

// Happy-path uninstall: HelmOps.Uninstall succeeds, the (non-existent)
// namespace is treated as already deleted, and the finally hook records
// Uninstalled.
func TestUninstallingApp_Success(t *testing.T) {
	am := buildAM("demo", appv1alpha1.App, appv1alpha1.Uninstalling, appv1alpha1.UninstallOp,
		configJSON(t, "demo", false))
	c := newFakeClient(t, am.DeepCopy())
	deps, tf := newTestDeps(c)

	app, serr := NewUninstallingApp(deps, am, 0)
	if serr != nil {
		t.Fatalf("new uninstalling app: %v", serr)
	}
	if _, err := app.(OperationApp).Exec(context.TODO()); err != nil {
		t.Fatalf("exec: %v", err)
	}

	waitForState(t, c, "demo", appv1alpha1.Uninstalled, 5*time.Second)
	if tf.helm.uninstallCalls != 1 {
		t.Fatalf("expected Uninstall called once, got %d", tf.helm.uninstallCalls)
	}
}

// uninstall-all annotation routes to HelmOps.UninstallAll.
func TestUninstallingApp_UninstallAll(t *testing.T) {
	am := buildAM("demo", appv1alpha1.App, appv1alpha1.Uninstalling, appv1alpha1.UninstallOp,
		configJSON(t, "demo", false))
	am.Annotations = map[string]string{"bytetrade.io/uninstall-all": "true"}
	c := newFakeClient(t, am.DeepCopy())
	deps, tf := newTestDeps(c)

	app, serr := NewUninstallingApp(deps, am, 0)
	if serr != nil {
		t.Fatalf("new uninstalling app: %v", serr)
	}
	if _, err := app.(OperationApp).Exec(context.TODO()); err != nil {
		t.Fatalf("exec: %v", err)
	}

	waitForState(t, c, "demo", appv1alpha1.Uninstalled, 5*time.Second)
	if tf.helm.uninstallAllCalls != 1 {
		t.Fatalf("expected UninstallAll called once, got %d", tf.helm.uninstallAllCalls)
	}
}

// A failing helm uninstall drives the app to UninstallFailed.
func TestUninstallingApp_Failure(t *testing.T) {
	am := buildAM("demo", appv1alpha1.App, appv1alpha1.Uninstalling, appv1alpha1.UninstallOp,
		configJSON(t, "demo", false))
	c := newFakeClient(t, am.DeepCopy())
	deps, tf := newTestDeps(c)
	tf.helm.uninstallErr = errors.New("boom")

	app, serr := NewUninstallingApp(deps, am, 0)
	if serr != nil {
		t.Fatalf("new uninstalling app: %v", serr)
	}
	if _, err := app.(OperationApp).Exec(context.TODO()); err != nil {
		t.Fatalf("exec: %v", err)
	}

	waitForState(t, c, "demo", appv1alpha1.UninstallFailed, 5*time.Second)
}
