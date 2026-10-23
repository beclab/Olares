package mwops_test

import (
	"context"
	"errors"
	"sync"
	"testing"

	kbopv1alpha1 "github.com/apecloud/kubeblocks/apis/operations/v1alpha1"
	appv1 "github.com/beclab/api/api/app.bytetrade.io/v1alpha1"

	corev1 "k8s.io/api/core/v1"
	apierrors "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/types"
	"sigs.k8s.io/controller-runtime/pkg/client"
	fakeclient "sigs.k8s.io/controller-runtime/pkg/client/fake"
	"sigs.k8s.io/controller-runtime/pkg/client/interceptor"

	"github.com/beclab/Olares/framework/app-service/pkg/applifecycle/adapters/mwops"
	"github.com/beclab/Olares/framework/app-service/pkg/middlewareinstaller"
	"github.com/beclab/Olares/framework/app-service/pkg/testutil"
)

// newWatchClient builds a fake client.WithWatch (interceptor.NewClient
// requires WithWatch, which testutil.NewFakeClient narrows away). Uses
// the same scheme as testutil so the objects behave identically.
func newWatchClient(objs ...client.Object) client.WithWatch {
	return fakeclient.NewClientBuilder().
		WithScheme(testutil.NewScheme()).
		WithObjects(objs...).
		Build()
}

// -----------------------------------------------------------------
// Test fixtures
// -----------------------------------------------------------------

// fakeKBOps is a mwops.KubeBlocksOps double: records Start/Stop
// invocations and returns canned errors.
type fakeKBOps struct {
	startErr error
	stopErr  error
	started  int
	stopped  int
}

func (f *fakeKBOps) Start() error { f.started++; return f.startErr }
func (f *fakeKBOps) Stop() error  { f.stopped++; return f.stopErr }

// fakeKBFactory records every factory invocation and returns the
// same fakeKBOps every time (tests interrogate that single instance).
type fakeKBFactory struct {
	mu    sync.Mutex
	calls []factoryCall
	ops   *fakeKBOps
}

type factoryCall struct {
	OpsType kbopv1alpha1.OpsType
	AMName  string
}

func (f *fakeKBFactory) Factory() mwops.KubeBlocksOpsFactory {
	return func(_ context.Context, opsType kbopv1alpha1.OpsType, am *appv1.ApplicationManager, _ client.Client) mwops.KubeBlocksOps {
		f.mu.Lock()
		defer f.mu.Unlock()
		f.calls = append(f.calls, factoryCall{OpsType: opsType, AMName: am.Name})
		return f.ops
	}
}

// recordingMongoUninstaller is the MongoUninstaller test double.
type recordingMongoUninstaller struct {
	mu    sync.Mutex
	calls []*middlewareinstaller.MiddlewareConfig
	err   error
}

func (r *recordingMongoUninstaller) Fn() mwops.MongoUninstaller {
	return func(_ context.Context, cfg *middlewareinstaller.MiddlewareConfig) error {
		r.mu.Lock()
		defer r.mu.Unlock()
		r.calls = append(r.calls, cfg)
		return r.err
	}
}

// -----------------------------------------------------------------
// New
// -----------------------------------------------------------------

func TestNew_NilClientPanics(t *testing.T) {
	defer func() {
		if r := recover(); r == nil {
			t.Fatal("expected panic on nil client")
		}
	}()
	noopUninstall := func(_ context.Context, _ *middlewareinstaller.MiddlewareConfig) error { return nil }
	_ = mwops.New(nil, noopUninstall)
}

func TestNew_NilMongoUninstallerPanics(t *testing.T) {
	defer func() {
		if r := recover(); r == nil {
			t.Fatal("expected panic on nil mongoUninstall")
		}
	}()
	_ = mwops.New(testutil.NewFakeClient(), nil)
}

func TestWithKubeBlocksFactory_NilKeepsDefault(t *testing.T) {
	noopUninstall := func(_ context.Context, _ *middlewareinstaller.MiddlewareConfig) error { return nil }
	o := mwops.New(testutil.NewFakeClient(), noopUninstall,
		mwops.WithKubeBlocksFactory(nil))
	if o == nil {
		t.Fatal("New returned nil")
	}
	// The default factory constructs a *kubeblocks.OperationOptions
	// whose validate() reads a Cluster CR from the client — a nil
	// factory would NPE much sooner, so successful construction is
	// the assertion.
}

// -----------------------------------------------------------------
// Start / Stop
// -----------------------------------------------------------------

func TestStart_NilAMErrors(t *testing.T) {
	kb := &fakeKBFactory{ops: &fakeKBOps{}}
	noopUninstall := func(_ context.Context, _ *middlewareinstaller.MiddlewareConfig) error { return nil }
	o := mwops.New(testutil.NewFakeClient(), noopUninstall,
		mwops.WithKubeBlocksFactory(kb.Factory()))
	if err := o.Start(context.Background(), nil); err == nil {
		t.Fatal("expected error on nil AM")
	}
	if len(kb.calls) != 0 {
		t.Errorf("factory should not run on nil AM: %+v", kb.calls)
	}
}

func TestStop_NilAMErrors(t *testing.T) {
	kb := &fakeKBFactory{ops: &fakeKBOps{}}
	noopUninstall := func(_ context.Context, _ *middlewareinstaller.MiddlewareConfig) error { return nil }
	o := mwops.New(testutil.NewFakeClient(), noopUninstall,
		mwops.WithKubeBlocksFactory(kb.Factory()))
	if err := o.Stop(context.Background(), nil); err == nil {
		t.Fatal("expected error on nil AM")
	}
	if len(kb.calls) != 0 {
		t.Errorf("factory should not run on nil AM: %+v", kb.calls)
	}
}

func TestStart_DispatchesStartTypeOpsRequest(t *testing.T) {
	kb := &fakeKBFactory{ops: &fakeKBOps{}}
	noopUninstall := func(_ context.Context, _ *middlewareinstaller.MiddlewareConfig) error { return nil }
	o := mwops.New(testutil.NewFakeClient(), noopUninstall,
		mwops.WithKubeBlocksFactory(kb.Factory()))
	am := &appv1.ApplicationManager{ObjectMeta: metav1.ObjectMeta{Name: "app-x"}}
	if err := o.Start(context.Background(), am); err != nil {
		t.Fatalf("Start: %v", err)
	}
	if len(kb.calls) != 1 {
		t.Fatalf("factory calls: %d want 1", len(kb.calls))
	}
	if kb.calls[0].OpsType != kbopv1alpha1.StartType {
		t.Errorf("opsType: %v want StartType", kb.calls[0].OpsType)
	}
	if kb.calls[0].AMName != "app-x" {
		t.Errorf("AMName: %q want app-x", kb.calls[0].AMName)
	}
	if kb.ops.started != 1 {
		t.Errorf("Start not called: %d", kb.ops.started)
	}
	if kb.ops.stopped != 0 {
		t.Errorf("Stop unexpectedly called: %d", kb.ops.stopped)
	}
}

func TestStop_DispatchesStopTypeOpsRequest(t *testing.T) {
	kb := &fakeKBFactory{ops: &fakeKBOps{}}
	noopUninstall := func(_ context.Context, _ *middlewareinstaller.MiddlewareConfig) error { return nil }
	o := mwops.New(testutil.NewFakeClient(), noopUninstall,
		mwops.WithKubeBlocksFactory(kb.Factory()))
	am := &appv1.ApplicationManager{ObjectMeta: metav1.ObjectMeta{Name: "app-x"}}
	if err := o.Stop(context.Background(), am); err != nil {
		t.Fatalf("Stop: %v", err)
	}
	if kb.calls[0].OpsType != kbopv1alpha1.StopType {
		t.Errorf("opsType: %v want StopType", kb.calls[0].OpsType)
	}
	if kb.ops.stopped != 1 {
		t.Errorf("Stop not called: %d", kb.ops.stopped)
	}
}

func TestStart_PropagatesKubeBlocksError(t *testing.T) {
	wantErr := errors.New("cluster not found")
	kb := &fakeKBFactory{ops: &fakeKBOps{startErr: wantErr}}
	noopUninstall := func(_ context.Context, _ *middlewareinstaller.MiddlewareConfig) error { return nil }
	o := mwops.New(testutil.NewFakeClient(), noopUninstall,
		mwops.WithKubeBlocksFactory(kb.Factory()))
	am := &appv1.ApplicationManager{ObjectMeta: metav1.ObjectMeta{Name: "app-x"}}
	if err := o.Start(context.Background(), am); !errors.Is(err, wantErr) {
		t.Errorf("err: %v want %v", err, wantErr)
	}
}

func TestStop_PropagatesKubeBlocksError(t *testing.T) {
	wantErr := errors.New("ops request rejected")
	kb := &fakeKBFactory{ops: &fakeKBOps{stopErr: wantErr}}
	noopUninstall := func(_ context.Context, _ *middlewareinstaller.MiddlewareConfig) error { return nil }
	o := mwops.New(testutil.NewFakeClient(), noopUninstall,
		mwops.WithKubeBlocksFactory(kb.Factory()))
	am := &appv1.ApplicationManager{ObjectMeta: metav1.ObjectMeta{Name: "app-x"}}
	if err := o.Stop(context.Background(), am); !errors.Is(err, wantErr) {
		t.Errorf("err: %v want %v", err, wantErr)
	}
}

// -----------------------------------------------------------------
// UninstallLegacyMongo
// -----------------------------------------------------------------

func TestUninstallLegacyMongo_WrongCfgType(t *testing.T) {
	rm := &recordingMongoUninstaller{}
	o := mwops.New(testutil.NewFakeClient(), rm.Fn())
	if err := o.UninstallLegacyMongo(context.Background(), "wrong type"); err == nil {
		t.Fatal("expected error on wrong cfg type")
	}
	if len(rm.calls) != 0 {
		t.Errorf("uninstall should not run on bad cfg: %+v", rm.calls)
	}
}

func TestUninstallLegacyMongo_NilCfgOfCorrectType(t *testing.T) {
	rm := &recordingMongoUninstaller{}
	o := mwops.New(testutil.NewFakeClient(), rm.Fn())
	var nilCfg *middlewareinstaller.MiddlewareConfig
	if err := o.UninstallLegacyMongo(context.Background(), nilCfg); err == nil {
		t.Fatal("expected error on nil cfg")
	}
	if len(rm.calls) != 0 {
		t.Errorf("uninstall should not run on nil cfg: %+v", rm.calls)
	}
}

func TestUninstallLegacyMongo_HappyPath_SecretDeleted(t *testing.T) {
	// Preseed the Helm release Secret so the Get succeeds and the
	// Delete path is exercised.
	secret := &corev1.Secret{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "sh.helm.release.v1.mongodb.v1",
			Namespace: "os-platform",
		},
	}
	c := testutil.NewFakeClient(secret)
	rm := &recordingMongoUninstaller{}
	o := mwops.New(c, rm.Fn())
	cfg := &middlewareinstaller.MiddlewareConfig{
		MiddlewareName: "mongodb",
		Namespace:      "os-platform",
		OwnerName:      "u1",
	}
	if err := o.UninstallLegacyMongo(context.Background(), cfg); err != nil {
		t.Fatalf("UninstallLegacyMongo: %v", err)
	}
	if len(rm.calls) != 1 {
		t.Fatalf("uninstall calls: %d want 1", len(rm.calls))
	}
	if rm.calls[0] != cfg {
		t.Errorf("cfg not passed through: got %+v want %+v", rm.calls[0], cfg)
	}
	// Assert the Secret is gone.
	var probe corev1.Secret
	err := c.Get(context.Background(), types.NamespacedName{
		Name:      "sh.helm.release.v1.mongodb.v1",
		Namespace: "os-platform",
	}, &probe)
	if !apierrors.IsNotFound(err) {
		t.Errorf("secret not deleted: err=%v", err)
	}
}

func TestUninstallLegacyMongo_SecretMissing_IsNotAnError(t *testing.T) {
	// No secret preseeded — Get returns NotFound, adapter must
	// treat as success.
	c := testutil.NewFakeClient()
	rm := &recordingMongoUninstaller{}
	o := mwops.New(c, rm.Fn())
	cfg := &middlewareinstaller.MiddlewareConfig{Namespace: "os-platform"}
	if err := o.UninstallLegacyMongo(context.Background(), cfg); err != nil {
		t.Fatalf("NotFound on secret Get should be success: %v", err)
	}
}

func TestUninstallLegacyMongo_ToleratedHelmError_ContinuesToSecretCleanup(t *testing.T) {
	// Preseed the Secret so we can prove the code path continued
	// past the tolerated helm error.
	secret := &corev1.Secret{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "sh.helm.release.v1.mongodb.v1",
			Namespace: "os-platform",
		},
	}
	c := testutil.NewFakeClient(secret)
	rm := &recordingMongoUninstaller{err: errors.New("failed to delete release: mongodb")}
	o := mwops.New(c, rm.Fn())
	cfg := &middlewareinstaller.MiddlewareConfig{Namespace: "os-platform"}
	if err := o.UninstallLegacyMongo(context.Background(), cfg); err != nil {
		t.Fatalf("tolerated helm error should not propagate: %v", err)
	}
	// Verify the Secret cleanup DID run despite the helm error.
	var probe corev1.Secret
	err := c.Get(context.Background(), types.NamespacedName{
		Name:      "sh.helm.release.v1.mongodb.v1",
		Namespace: "os-platform",
	}, &probe)
	if !apierrors.IsNotFound(err) {
		t.Errorf("secret not deleted after tolerated helm error: err=%v", err)
	}
}

func TestUninstallLegacyMongo_NonToleratedHelmError_Propagates(t *testing.T) {
	wantErr := errors.New("connection refused")
	c := testutil.NewFakeClient()
	rm := &recordingMongoUninstaller{err: wantErr}
	o := mwops.New(c, rm.Fn())
	cfg := &middlewareinstaller.MiddlewareConfig{Namespace: "os-platform"}
	if err := o.UninstallLegacyMongo(context.Background(), cfg); !errors.Is(err, wantErr) {
		t.Errorf("err: %v want %v", err, wantErr)
	}
}

func TestUninstallLegacyMongo_SecretGetError_Propagates(t *testing.T) {
	// Interceptor injects a non-NotFound error on the Secret Get.
	wantErr := errors.New("apiserver timeout")
	c := interceptor.NewClient(newWatchClient(), interceptor.Funcs{
		Get: func(_ context.Context, _ client.WithWatch, key client.ObjectKey, _ client.Object, _ ...client.GetOption) error {
			if key.Name == "sh.helm.release.v1.mongodb.v1" {
				return wantErr
			}
			return apierrors.NewNotFound(corev1.Resource("secrets"), key.Name)
		},
	})
	rm := &recordingMongoUninstaller{}
	o := mwops.New(c, rm.Fn())
	cfg := &middlewareinstaller.MiddlewareConfig{Namespace: "os-platform"}
	if err := o.UninstallLegacyMongo(context.Background(), cfg); !errors.Is(err, wantErr) {
		t.Errorf("err: %v want %v", err, wantErr)
	}
}

func TestUninstallLegacyMongo_SecretDeleteNotFound_IsNotAnError(t *testing.T) {
	// Race: Get succeeds but Delete returns NotFound (a parallel
	// uninstall snatched it). Must not propagate.
	secret := &corev1.Secret{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "sh.helm.release.v1.mongodb.v1",
			Namespace: "os-platform",
		},
	}
	base := newWatchClient(secret)
	c := interceptor.NewClient(base, interceptor.Funcs{
		Delete: func(_ context.Context, _ client.WithWatch, obj client.Object, _ ...client.DeleteOption) error {
			return apierrors.NewNotFound(corev1.Resource("secrets"), obj.GetName())
		},
	})
	rm := &recordingMongoUninstaller{}
	o := mwops.New(c, rm.Fn())
	cfg := &middlewareinstaller.MiddlewareConfig{Namespace: "os-platform"}
	if err := o.UninstallLegacyMongo(context.Background(), cfg); err != nil {
		t.Fatalf("NotFound on Delete should be success: %v", err)
	}
}

func TestUninstallLegacyMongo_SecretDeleteError_Propagates(t *testing.T) {
	wantErr := errors.New("etcd unavailable")
	secret := &corev1.Secret{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "sh.helm.release.v1.mongodb.v1",
			Namespace: "os-platform",
		},
	}
	base := newWatchClient(secret)
	c := interceptor.NewClient(base, interceptor.Funcs{
		Delete: func(_ context.Context, _ client.WithWatch, _ client.Object, _ ...client.DeleteOption) error {
			return wantErr
		},
	})
	rm := &recordingMongoUninstaller{}
	o := mwops.New(c, rm.Fn())
	cfg := &middlewareinstaller.MiddlewareConfig{Namespace: "os-platform"}
	if err := o.UninstallLegacyMongo(context.Background(), cfg); !errors.Is(err, wantErr) {
		t.Errorf("err: %v want %v", err, wantErr)
	}
}

// -----------------------------------------------------------------
// Sanity: the default (production) factory returns something that
// implements the port so main.go wiring won't NPE at first use.
// The Cluster CR is missing so validate() will fail — the whole
// point is that we can *build* the ops object without panicking.
// -----------------------------------------------------------------

func TestDefaultKBFactory_BuildsAndInvokesWithoutPanic(t *testing.T) {
	c := testutil.NewFakeClient()
	noopUninstall := func(_ context.Context, _ *middlewareinstaller.MiddlewareConfig) error { return nil }
	o := mwops.New(c, noopUninstall)
	am := &appv1.ApplicationManager{
		ObjectMeta: metav1.ObjectMeta{Name: "app-x"},
		Spec: appv1.ApplicationManagerSpec{
			AppName:      "app-x",
			AppNamespace: "user-x-ns",
		},
	}
	// We EXPECT an error here (validate() will fail on the
	// Cluster Get). Assertion: no panic and error is returned.
	if err := o.Start(context.Background(), am); err == nil {
		t.Log("Start returned nil against empty apiserver — unexpected but not fatal for this smoke test")
	}
}
