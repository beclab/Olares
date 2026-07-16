package controllers

import (
	"context"
	"encoding/json"
	"errors"
	"path/filepath"
	"testing"
	"time"

	"github.com/beclab/Olares/framework/app-service/pkg/appcfg"
	"github.com/beclab/Olares/framework/app-service/pkg/appinstaller"
	"github.com/beclab/Olares/framework/app-service/pkg/appstate"
	"github.com/beclab/Olares/framework/app-service/pkg/compute/validation"
	"github.com/beclab/Olares/framework/app-service/pkg/images"
	appv1alpha1 "github.com/beclab/api/api/app.bytetrade.io/v1alpha1"

	kbopv1alpha1 "github.com/apecloud/kubeblocks/apis/operations/v1alpha1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"k8s.io/apimachinery/pkg/types"
	clientgoscheme "k8s.io/client-go/kubernetes/scheme"
	"k8s.io/client-go/rest"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/client/fake"
)

// These tests drive the real controller reconcile loop (Reconcile ->
// LoadStatefulApp -> *App.Exec -> async finally/poll) against a fake client
// with every external boundary replaced by a fake via appstate.NewDeps.
//
// Full state-machine edge coverage lives in state_flow_lineages_test.go,
// which drives every lineage from <none> end-to-end and pins every entry
// in appstate.ReconcileDrivenTransitions via TestStateFlow_Lineages_CoversReconcileEdges.
// The cases below are focused smoke tests for a few high-traffic paths and
// fake seam assertions.

// ---------------------------------------------------------------------------
// fakes implementing the appstate seams
// ---------------------------------------------------------------------------

type fakeHelmOps struct {
	installErr      error
	uninstallErr    error
	uninstallAllErr error
	scaleErr        error
	upgradeErr      error
	applyEnvErr     error

	// nil wait funcs default to "ready, no error".
	waitForStartUp func() (bool, error)
	waitForLaunch  func() (bool, error)

	installCalls      int
	uninstallCalls    int
	uninstallAllCalls int
	scaleCalls        []int32
}

var _ appinstaller.HelmOpsInterface = (*fakeHelmOps)(nil)

func (f *fakeHelmOps) Install() error      { f.installCalls++; return f.installErr }
func (f *fakeHelmOps) Uninstall() error    { f.uninstallCalls++; return f.uninstallErr }
func (f *fakeHelmOps) UninstallAll() error { f.uninstallAllCalls++; return f.uninstallAllErr }
func (f *fakeHelmOps) Upgrade() error      { return f.upgradeErr }
func (f *fakeHelmOps) ApplyEnv() error     { return f.applyEnvErr }
func (f *fakeHelmOps) RollBack() error     { return nil }
func (f *fakeHelmOps) Scale(replicas int32) error {
	f.scaleCalls = append(f.scaleCalls, replicas)
	return f.scaleErr
}
func (f *fakeHelmOps) WaitForStartUp() (bool, error) {
	if f.waitForStartUp != nil {
		return f.waitForStartUp()
	}
	return true, nil
}
func (f *fakeHelmOps) WaitForLaunch() (bool, error) {
	if f.waitForLaunch != nil {
		return f.waitForLaunch()
	}
	return true, nil
}

type fakeImageManager struct {
	createErr       error
	pollErr         error
	updateStatusErr error
	createCalls     int
	pollCalls       int
}

var _ images.ImageManager = (*fakeImageManager)(nil)

func (f *fakeImageManager) Create(ctx context.Context, am *appv1alpha1.ApplicationManager, refs []appv1alpha1.Ref) error {
	f.createCalls++
	return f.createErr
}
func (f *fakeImageManager) UpdateStatus(ctx context.Context, name, state, message string) error {
	return f.updateStatusErr
}
func (f *fakeImageManager) PollDownloadProgress(ctx context.Context, am *appv1alpha1.ApplicationManager) error {
	f.pollCalls++
	return f.pollErr
}

type fakeMiddlewareOp struct {
	startErr error
	stopErr  error
	started  bool
	stopped  bool
}

var _ appstate.MiddlewareOperator = (*fakeMiddlewareOp)(nil)

func (f *fakeMiddlewareOp) Start() error { f.started = true; return f.startErr }
func (f *fakeMiddlewareOp) Stop() error  { f.stopped = true; return f.stopErr }

type ctrlFakes struct {
	helm *fakeHelmOps
	img  *fakeImageManager
	mw   *fakeMiddlewareOp
}

// ---------------------------------------------------------------------------
// scheme / client / controller builders
// ---------------------------------------------------------------------------

func integrationScheme(t *testing.T) *runtime.Scheme {
	t.Helper()
	s := runtime.NewScheme()
	if err := clientgoscheme.AddToScheme(s); err != nil {
		t.Fatalf("add clientgo scheme: %v", err)
	}
	if err := appv1alpha1.AddToScheme(s); err != nil {
		t.Fatalf("add app.bytetrade.io scheme: %v", err)
	}
	gv := schema.GroupVersion{Group: "gpu.bytetrade.io", Version: "v1alpha1"}
	binding := &unstructured.Unstructured{}
	binding.SetGroupVersionKind(gv.WithKind("GPUBinding"))
	list := &unstructured.UnstructuredList{}
	list.SetGroupVersionKind(gv.WithKind("GPUBindingList"))
	s.AddKnownTypeWithName(gv.WithKind("GPUBinding"), binding)
	s.AddKnownTypeWithName(gv.WithKind("GPUBindingList"), list)
	return s
}

// newTestController builds a controller whose client and Deps share a single
// fake store, with all external boundaries faked.
func newTestController(t *testing.T, objs ...client.Object) (*ApplicationManagerController, *ctrlFakes) {
	t.Helper()
	// Point KUBECONFIG at a nonexistent file so any residual ctrl.GetConfig
	// reach (e.g. SuspendingApp's best-effort reason refresh) fails fast
	// instead of blocking on an inherited ~/.kube/config.
	t.Setenv("KUBECONFIG", filepath.Join(t.TempDir(), "nonexistent-kubeconfig"))

	c := fake.NewClientBuilder().
		WithScheme(integrationScheme(t)).
		WithObjects(objs...).
		Build()

	f := &ctrlFakes{
		helm: &fakeHelmOps{},
		img:  &fakeImageManager{},
		mw:   &fakeMiddlewareOp{},
	}

	deps := appstate.NewDeps(
		c,
		func() (*rest.Config, error) { return &rest.Config{}, nil },
		func(ctx context.Context, kubeConfig *rest.Config, app *appcfg.ApplicationConfig,
			token string, options appinstaller.Opt) (appinstaller.HelmOpsInterface, error) {
			return f.helm, nil
		},
		func(ctx context.Context, opsType kbopv1alpha1.OpsType,
			manager *appv1alpha1.ApplicationManager, cl client.Client) appstate.MiddlewareOperator {
			return f.mw
		},
		func(client.Client) images.ImageManager { return f.img },
		func(ctx context.Context, am *appv1alpha1.ApplicationManager,
			cfg *appcfg.ApplicationConfig) ([]appv1alpha1.Ref, error) {
			return nil, nil
		},
	)
	// Install-specific seams: no-op port assignment and accept-all validation
	// by default, so integration tests exercise the state machine rather than
	// real port/compute logic. Override r.Deps.* in a test to drive rejections.
	deps.SetExposePorts = func(ctx context.Context, cfg *appcfg.ApplicationConfig, prev map[string]int32) error {
		return nil
	}
	deps.RunInstallValidation = func(ctx context.Context, in validation.Input) (validation.Decision, error) {
		return validation.Decision{OK: true}, nil
	}
	// Count Downloading apps via the fake controller-runtime client. The
	// production seam reads a live clientset (utils.GetClient), which would
	// fail here without a real kubeconfig.
	deps.CountDownloading = func(ctx context.Context) (int, error) {
		return appstate.CountDownloadingViaClient(ctx, c)
	}
	// Default IsAdmin to a non-admin, non-error answer. The production seam
	// hits the kubesphere user API; without this override every flow case
	// that touches install-cancel / upgrade would error out on a real network
	// call to localhost. Tests that need the admin branch can override.
	deps.IsAdmin = func(ctx context.Context, kubeConfig *rest.Config, owner string) (bool, error) {
		return false, nil
	}

	r := &ApplicationManagerController{
		Client:      c,
		KubeConfig:  &rest.Config{},
		ImageClient: f.img,
		Deps:        deps,
	}
	return r, f
}

func doReconcile(t *testing.T, r *ApplicationManagerController, name string) {
	t.Helper()
	if _, err := r.Reconcile(context.TODO(), ctrl.Request{
		NamespacedName: types.NamespacedName{Name: name},
	}); err != nil {
		// Some flows (e.g. resume scale failure) legitimately return an
		// error from Reconcile while still recording the failed state, so
		// this is logged, not fatal.
		t.Logf("reconcile %s returned: %v", name, err)
	}
}

func waitForState(t *testing.T, c client.Client, name string,
	want appv1alpha1.ApplicationManagerState, timeout time.Duration) {
	t.Helper()
	var am appv1alpha1.ApplicationManager
	deadline := time.Now().Add(timeout)
	for time.Now().Before(deadline) {
		if err := c.Get(context.TODO(), types.NamespacedName{Name: name}, &am); err == nil {
			if am.Status.State == want {
				return
			}
		}
		time.Sleep(5 * time.Millisecond)
	}
	_ = c.Get(context.TODO(), types.NamespacedName{Name: name}, &am)
	t.Fatalf("timed out waiting for %s to reach %q, last state %q", name, want, am.Status.State)
}

func buildIntegrationAM(name string, typ appv1alpha1.Type, state appv1alpha1.ApplicationManagerState,
	opType appv1alpha1.OpType, withWorkloadReplicas bool) *appv1alpha1.ApplicationManager {
	return buildIntegrationAMAnnotated(name, typ, state, opType, withWorkloadReplicas, nil)
}

func buildIntegrationAMAnnotated(name string, typ appv1alpha1.Type, state appv1alpha1.ApplicationManagerState,
	opType appv1alpha1.OpType, withWorkloadReplicas bool, annotations map[string]string) *appv1alpha1.ApplicationManager {
	cfg := appcfg.ApplicationConfig{
		AppName:   name,
		OwnerName: "owner",
		Namespace: name + "-ns",
	}
	if withWorkloadReplicas {
		wr := appcfg.WorkloadReplicas{name: 1}
		cfg.WorkloadReplicas = &wr
	}
	raw, _ := json.Marshal(cfg)
	now := metav1.Now()
	return &appv1alpha1.ApplicationManager{
		ObjectMeta: metav1.ObjectMeta{Name: name, Annotations: annotations},
		Spec: appv1alpha1.ApplicationManagerSpec{
			AppName:      name,
			AppOwner:     "owner",
			AppNamespace: name + "-ns",
			Source:       "market",
			Type:         typ,
			OpType:       opType,
			Config:       string(raw),
		},
		Status: appv1alpha1.ApplicationManagerStatus{
			State:      state,
			OpType:     opType,
			StatusTime: &now,
			UpdateTime: &now,
		},
	}
}

// ---------------------------------------------------------------------------
// tests
// ---------------------------------------------------------------------------

// Reconciling a Stopping app runs the suspend flow to completion (Stopped).
func TestReconcile_Stop_Success(t *testing.T) {
	am := buildIntegrationAM("demo", appv1alpha1.App, appv1alpha1.Stopping, appv1alpha1.StopOp, false)
	r, _ := newTestController(t, am)

	doReconcile(t, r, "demo")

	waitForState(t, r.Client, "demo", appv1alpha1.Stopped, 5*time.Second)
}

// Reconciling an Uninstalling app drives the async uninstall to Uninstalled.
func TestReconcile_Uninstall_Success(t *testing.T) {
	am := buildIntegrationAM("demo", appv1alpha1.App, appv1alpha1.Uninstalling, appv1alpha1.UninstallOp, false)
	r, f := newTestController(t, am)

	doReconcile(t, r, "demo")

	waitForState(t, r.Client, "demo", appv1alpha1.Uninstalled, 5*time.Second)
	if f.helm.uninstallCalls != 1 {
		t.Fatalf("expected Uninstall called once, got %d", f.helm.uninstallCalls)
	}
}

// A failing helm uninstall lands the app in UninstallFailed.
func TestReconcile_Uninstall_Failure(t *testing.T) {
	am := buildIntegrationAM("demo", appv1alpha1.App, appv1alpha1.Uninstalling, appv1alpha1.UninstallOp, false)
	r, f := newTestController(t, am)
	f.helm.uninstallErr = errors.New("boom")

	doReconcile(t, r, "demo")

	waitForState(t, r.Client, "demo", appv1alpha1.UninstallFailed, 5*time.Second)
}

// Reconciling a Downloading app exercises the controller's own poll kickoff:
// Exec returns a pollable in-progress app, Reconcile starts WaitAsync, the
// fake reports completion and the finally hook advances to Installing.
func TestReconcile_Download_To_Installing(t *testing.T) {
	am := buildIntegrationAM("demo", appv1alpha1.App, appv1alpha1.Downloading, appv1alpha1.InstallOp, false)
	r, f := newTestController(t, am)

	doReconcile(t, r, "demo")

	waitForState(t, r.Client, "demo", appv1alpha1.Installing, 5*time.Second)
	if f.img.createCalls != 1 {
		t.Fatalf("expected ImageManager.Create called once, got %d", f.img.createCalls)
	}
	if f.img.pollCalls < 1 {
		t.Fatalf("expected PollDownloadProgress to be called at least once, got %d", f.img.pollCalls)
	}
}

// A failing download poll lands the app in DownloadFailed.
func TestReconcile_Download_PollFailure(t *testing.T) {
	am := buildIntegrationAM("demo", appv1alpha1.App, appv1alpha1.Downloading, appv1alpha1.InstallOp, false)
	r, f := newTestController(t, am)
	f.img.pollErr = errors.New("pull failed")

	doReconcile(t, r, "demo")

	waitForState(t, r.Client, "demo", appv1alpha1.DownloadFailed, 5*time.Second)
}

// Reconciling an Installing app runs the full two-leg install (validate ->
// helm install -> scale-up -> startup) and advances to Initializing.
func TestReconcile_Install_To_Initializing(t *testing.T) {
	am := buildIntegrationAM("demo", appv1alpha1.App, appv1alpha1.Installing, appv1alpha1.InstallOp, false)
	r, f := newTestController(t, am)

	doReconcile(t, r, "demo")

	waitForState(t, r.Client, "demo", appv1alpha1.Initializing, 5*time.Second)
	if f.helm.scaleCalls == nil {
		t.Fatalf("expected Scale to be called during install")
	}
}

// Install validation rejection (via overriding the seam) lands InstallFailed.
func TestReconcile_Install_ValidationRejected(t *testing.T) {
	am := buildIntegrationAM("demo", appv1alpha1.App, appv1alpha1.Installing, appv1alpha1.InstallOp, false)
	r, _ := newTestController(t, am)
	r.Deps.RunInstallValidation = func(ctx context.Context, in validation.Input) (validation.Decision, error) {
		return validation.Decision{OK: false, Validator: "k8s-request", Message: "insufficient memory"}, nil
	}

	doReconcile(t, r, "demo")

	waitForState(t, r.Client, "demo", appv1alpha1.InstallFailed, 5*time.Second)
}

// A failing Scale during resume lands the app in ResumeFailed.
func TestReconcile_Resume_ScaleFailure(t *testing.T) {
	am := buildIntegrationAM("demo", appv1alpha1.App, appv1alpha1.Resuming, appv1alpha1.ResumeOp, true)
	r, f := newTestController(t, am)
	f.helm.scaleErr = errors.New("boom")

	doReconcile(t, r, "demo")

	waitForState(t, r.Client, "demo", appv1alpha1.ResumeFailed, 5*time.Second)
}
