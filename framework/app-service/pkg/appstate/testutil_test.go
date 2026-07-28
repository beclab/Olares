package appstate

import (
	"context"
	"encoding/json"
	"path/filepath"
	"testing"
	"time"

	"github.com/beclab/Olares/framework/app-service/pkg/appcfg"
	"github.com/beclab/Olares/framework/app-service/pkg/appinstaller"
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
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/client/fake"
)

// ---------------------------------------------------------------------------
// fakeHelmOps is a programmable stand-in for appinstaller.HelmOpsInterface.
// Every method records that it was called and returns the configured error /
// readiness value so a test can drive each branch of the state machine
// without a real helm/cluster.
// ---------------------------------------------------------------------------

type fakeHelmOps struct {
	installErr      error
	uninstallErr    error
	uninstallAllErr error
	upgradeErr      error
	applyEnvErr     error
	rollbackErr     error
	scaleErr        error

	// waitForStartUp / waitForLaunch let a test override the default
	// "ready, no error" answer used by polling/launch flows.
	waitForStartUp func() (bool, error)
	waitForLaunch  func() (bool, error)

	installCalls      int
	uninstallCalls    int
	uninstallAllCalls int
	upgradeCalls      int
	applyEnvCalls     int
	scaleCalls        []int32
}

var _ appinstaller.HelmOpsInterface = (*fakeHelmOps)(nil)

func (f *fakeHelmOps) Install() error      { f.installCalls++; return f.installErr }
func (f *fakeHelmOps) Uninstall() error    { f.uninstallCalls++; return f.uninstallErr }
func (f *fakeHelmOps) UninstallAll() error { f.uninstallAllCalls++; return f.uninstallAllErr }
func (f *fakeHelmOps) Upgrade() error      { f.upgradeCalls++; return f.upgradeErr }
func (f *fakeHelmOps) ApplyEnv() error     { f.applyEnvCalls++; return f.applyEnvErr }
func (f *fakeHelmOps) RollBack() error     { return f.rollbackErr }

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

// ---------------------------------------------------------------------------
// fakeImageManager stands in for images.ImageManager.
// ---------------------------------------------------------------------------

type fakeImageManager struct {
	createErr error
	updateErr error
	pollErr   error

	createCalls int
	pollCalls   int
	lastRefs    []appv1alpha1.Ref
}

var _ images.ImageManager = (*fakeImageManager)(nil)

func (f *fakeImageManager) Create(ctx context.Context, am *appv1alpha1.ApplicationManager, refs []appv1alpha1.Ref) error {
	f.createCalls++
	f.lastRefs = refs
	return f.createErr
}

func (f *fakeImageManager) UpdateStatus(ctx context.Context, name, state, message string) error {
	return f.updateErr
}

func (f *fakeImageManager) PollDownloadProgress(ctx context.Context, am *appv1alpha1.ApplicationManager) error {
	f.pollCalls++
	return f.pollErr
}

// ---------------------------------------------------------------------------
// fakeMiddlewareOp stands in for MiddlewareOperator (kubeblocks start/stop).
// ---------------------------------------------------------------------------

type fakeMiddlewareOp struct {
	startErr error
	stopErr  error
	started  bool
	stopped  bool
}

var _ MiddlewareOperator = (*fakeMiddlewareOp)(nil)

func (f *fakeMiddlewareOp) Start() error { f.started = true; return f.startErr }
func (f *fakeMiddlewareOp) Stop() error  { f.stopped = true; return f.stopErr }

// ---------------------------------------------------------------------------
// testFakes bundles the configurable seams handed to a Deps so a test can
// both program inputs (errors / readiness) and assert on recorded calls.
// ---------------------------------------------------------------------------

type testFakes struct {
	helm *fakeHelmOps
	img  *fakeImageManager
	mw   *fakeMiddlewareOp

	// kubeConfigErr / newHelmOpsErr force the corresponding seam to fail.
	kubeConfigErr error
	newHelmOpsErr error

	// resolveRefs overrides ResolveImageRefs; nil means "return no refs".
	resolveRefs func(ctx context.Context, am *appv1alpha1.ApplicationManager, cfg *appcfg.ApplicationConfig) ([]appv1alpha1.Ref, error)

	// setExposePortsErr forces the SetExposePorts seam to fail.
	setExposePortsErr error
	// validation overrides RunInstallValidation; nil means "accept".
	validation func(ctx context.Context, in validation.Input) (validation.Decision, error)
}

// newTestDeps builds a Deps wired entirely to fakes, plus a fresh factory so
// in-progress state never leaks across tests.
func newTestDeps(c client.Client) (Deps, *testFakes) {
	tf := &testFakes{
		helm: &fakeHelmOps{},
		img:  &fakeImageManager{},
		mw:   &fakeMiddlewareOp{},
	}

	deps := Deps{
		Client:  c,
		Factory: newStatefulAppFactory(),
		KubeConfig: func() (*rest.Config, error) {
			if tf.kubeConfigErr != nil {
				return nil, tf.kubeConfigErr
			}
			return &rest.Config{}, nil
		},
		NewHelmOps: func(ctx context.Context, kubeConfig *rest.Config, app *appcfg.ApplicationConfig,
			token string, options appinstaller.Opt) (appinstaller.HelmOpsInterface, error) {
			if tf.newHelmOpsErr != nil {
				return nil, tf.newHelmOpsErr
			}
			return tf.helm, nil
		},
		NewMiddlewareOp: func(ctx context.Context, opsType kbopv1alpha1.OpsType,
			manager *appv1alpha1.ApplicationManager, cl client.Client) MiddlewareOperator {
			return tf.mw
		},
		NewImageManager: func(client.Client) images.ImageManager {
			return tf.img
		},
		ResolveImageRefs: func(ctx context.Context, am *appv1alpha1.ApplicationManager,
			cfg *appcfg.ApplicationConfig) ([]appv1alpha1.Ref, error) {
			if tf.resolveRefs != nil {
				return tf.resolveRefs(ctx, am, cfg)
			}
			return nil, nil
		},
		SetExposePorts: func(ctx context.Context, cfg *appcfg.ApplicationConfig,
			prevPortsMap map[string]int32) error {
			return tf.setExposePortsErr
		},
		RunInstallValidation: func(ctx context.Context, in validation.Input) (validation.Decision, error) {
			if tf.validation != nil {
				return tf.validation(ctx, in)
			}
			return validation.Decision{OK: true}, nil
		},
	}
	return deps, tf
}

// ---------------------------------------------------------------------------
// scheme / client helpers
// ---------------------------------------------------------------------------

// newTestScheme registers everything the appstate code touches with a fake
// client: core/apps (clientgoscheme), the app.bytetrade.io CRDs, and the
// HAMI GPUBinding GVK as Unstructured so compute cleanup's List call resolves
// to an empty list instead of an unregistered-kind error.
func newTestScheme(t *testing.T) *runtime.Scheme {
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

// newFakeClient builds a controller-runtime fake client seeded with objs.
// It deliberately does NOT register a status subresource because the state
// machine updates status with a plain Patch (see baseStatefulApp.updateStatus).
func newFakeClient(t *testing.T, objs ...client.Object) client.Client {
	t.Helper()
	return fake.NewClientBuilder().
		WithScheme(newTestScheme(t)).
		WithObjects(objs...).
		Build()
}

// ---------------------------------------------------------------------------
// ApplicationManager / config builders
// ---------------------------------------------------------------------------

// buildAM returns a minimal but valid ApplicationManager for tests.
func buildAM(name string, typ appv1alpha1.Type, state appv1alpha1.ApplicationManagerState,
	opType appv1alpha1.OpType, config string) *appv1alpha1.ApplicationManager {
	now := metav1.Now()
	return &appv1alpha1.ApplicationManager{
		ObjectMeta: metav1.ObjectMeta{
			Name: name,
		},
		Spec: appv1alpha1.ApplicationManagerSpec{
			AppName:      name,
			AppOwner:     "owner",
			AppNamespace: name + "-ns",
			Source:       "market",
			Type:         typ,
			OpType:       opType,
			Config:       config,
		},
		Status: appv1alpha1.ApplicationManagerStatus{
			State:      state,
			OpType:     opType,
			StatusTime: &now,
			UpdateTime: &now,
		},
	}
}

// configJSON marshals a minimal ApplicationConfig. When withWorkloadReplicas
// is true the config declares a single workload, routing suspend/resume down
// the HelmOps.Scale path instead of the legacy direct-patch path.
func configJSON(t *testing.T, appName string, withWorkloadReplicas bool) string {
	t.Helper()
	cfg := appcfg.ApplicationConfig{
		AppName:   appName,
		OwnerName: "owner",
		Namespace: appName + "-ns",
	}
	if withWorkloadReplicas {
		wr := appcfg.WorkloadReplicas{appName: 1}
		cfg.WorkloadReplicas = &wr
	}
	b, err := json.Marshal(cfg)
	if err != nil {
		t.Fatalf("marshal config: %v", err)
	}
	return string(b)
}

// ---------------------------------------------------------------------------
// drive / wait helpers
// ---------------------------------------------------------------------------

// drivePolling kicks off the async poll loop of a pollable in-progress app
// (downloading / resuming) the same way the controller would after Exec.
func drivePolling(ip StatefulInProgressApp) {
	p := ip.(PollableStatefulInProgressApp)
	ctx := p.CreatePollContext()
	p.WaitAsync(ctx)
}

// waitForState polls the fake client until the named ApplicationManager
// reaches want or the timeout elapses. It fails the test on timeout.
func waitForState(t *testing.T, c client.Client, name string,
	want appv1alpha1.ApplicationManagerState, timeout time.Duration) *appv1alpha1.ApplicationManager {
	t.Helper()
	var am appv1alpha1.ApplicationManager
	deadline := time.Now().Add(timeout)
	for time.Now().Before(deadline) {
		if err := c.Get(context.TODO(), types.NamespacedName{Name: name}, &am); err == nil {
			if am.Status.State == want {
				return &am
			}
		}
		time.Sleep(5 * time.Millisecond)
	}
	_ = c.Get(context.TODO(), types.NamespacedName{Name: name}, &am)
	t.Fatalf("timed out waiting for %s to reach %q, last state %q", name, want, am.Status.State)
	return nil
}

// isolateKubeconfig points KUBECONFIG at a nonexistent file so any code path
// that still reaches for a real client via ctrl.GetConfig (e.g. the
// best-effort reason refresh in SuspendingApp.Exec) fails fast instead of
// blocking ~30s on an unreachable API server inherited from the developer's
// ~/.kube/config. Tests assert on the injected fakes, not that read.
func isolateKubeconfig(t *testing.T) {
	t.Helper()
	t.Setenv("KUBECONFIG", filepath.Join(t.TempDir(), "nonexistent-kubeconfig"))
}

// getAM fetches the current ApplicationManager or fails the test.
func getAM(t *testing.T, c client.Client, name string) *appv1alpha1.ApplicationManager {
	t.Helper()
	var am appv1alpha1.ApplicationManager
	if err := c.Get(context.TODO(), types.NamespacedName{Name: name}, &am); err != nil {
		t.Fatalf("get app manager %s: %v", name, err)
	}
	return &am
}
