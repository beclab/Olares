package appstate

import (
	"context"
	"errors"
	"testing"
	"time"

	appv1alpha1 "github.com/beclab/api/api/app.bytetrade.io/v1alpha1"
)

// Happy-path download: refs resolve, ImageManager.Create succeeds, the poll
// loop reports completion and the app advances to Installing.
func TestDownloadingApp_Success_Installing(t *testing.T) {
	am := buildAM("demo", appv1alpha1.App, appv1alpha1.Downloading, appv1alpha1.InstallOp,
		configJSON(t, "demo", false))
	c := newFakeClient(t, am.DeepCopy())
	deps, tf := newTestDeps(c)

	app, serr := NewDownloadingApp(deps, am, 0)
	if serr != nil {
		t.Fatalf("new downloading app: %v", serr)
	}
	ip, err := app.(OperationApp).Exec(context.TODO())
	if err != nil {
		t.Fatalf("exec: %v", err)
	}
	if ip == nil {
		t.Fatalf("expected in-progress app from download exec")
	}
	drivePolling(ip)

	waitForState(t, c, "demo", appv1alpha1.Installing, 5*time.Second)
	if tf.img.createCalls != 1 {
		t.Fatalf("expected ImageManager.Create called once, got %d", tf.img.createCalls)
	}
}

// ImageManager.Create failure surfaces synchronously as DownloadFailed.
func TestDownloadingApp_CreateFailure_DownloadFailed(t *testing.T) {
	am := buildAM("demo", appv1alpha1.App, appv1alpha1.Downloading, appv1alpha1.InstallOp,
		configJSON(t, "demo", false))
	c := newFakeClient(t, am.DeepCopy())
	deps, tf := newTestDeps(c)
	tf.img.createErr = errors.New("boom")

	app, serr := NewDownloadingApp(deps, am, 0)
	if serr != nil {
		t.Fatalf("new downloading app: %v", serr)
	}
	if _, err := app.(OperationApp).Exec(context.TODO()); err == nil {
		t.Fatalf("expected exec error on create failure")
	}

	if got := getAM(t, c, "demo").Status.State; got != appv1alpha1.DownloadFailed {
		t.Fatalf("state = %q, want %q", got, appv1alpha1.DownloadFailed)
	}
}

// A failing download poll moves the app to DownloadFailed.
func TestDownloadingApp_PollFailure_DownloadFailed(t *testing.T) {
	am := buildAM("demo", appv1alpha1.App, appv1alpha1.Downloading, appv1alpha1.InstallOp,
		configJSON(t, "demo", false))
	c := newFakeClient(t, am.DeepCopy())
	deps, tf := newTestDeps(c)
	tf.img.pollErr = errors.New("pull failed")

	app, serr := NewDownloadingApp(deps, am, 0)
	if serr != nil {
		t.Fatalf("new downloading app: %v", serr)
	}
	ip, err := app.(OperationApp).Exec(context.TODO())
	if err != nil {
		t.Fatalf("exec: %v", err)
	}
	drivePolling(ip)

	waitForState(t, c, "demo", appv1alpha1.DownloadFailed, 5*time.Second)
}
