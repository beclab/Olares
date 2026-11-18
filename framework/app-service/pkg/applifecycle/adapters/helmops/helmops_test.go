package helmops_test

import (
	"context"
	"encoding/json"
	"errors"
	"sync"
	"testing"
	"time"

	appv1 "github.com/beclab/api/api/app.bytetrade.io/v1alpha1"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/rest"

	"github.com/beclab/Olares/framework/app-service/pkg/appcfg"
	"github.com/beclab/Olares/framework/app-service/pkg/appinstaller"
	"github.com/beclab/Olares/framework/app-service/pkg/applifecycle/adapters/helmops"
	"github.com/beclab/Olares/framework/app-service/pkg/applifecycle/ports"
	"github.com/beclab/Olares/framework/app-service/pkg/testutil"
)

// -----------------------------------------------------------------
// Test fixtures shared across writes_test / reads_test /
// namespace_wait_test.
// -----------------------------------------------------------------

// fakeHelmOps is an appinstaller.HelmOpsInterface test double.
// Every method records itself and returns a canned error.
type fakeHelmOps struct {
	mu sync.Mutex

	installErr       error
	upgradeErr       error
	applyEnvErr      error
	uninstallErr     error
	uninstallAllErr  error
	scaleErr         error
	waitStartUpErr   error
	waitLaunchErr    error
	scaleReplicasLog []int32
	methodsCalled    []string
}

func (f *fakeHelmOps) record(m string) {
	f.mu.Lock()
	defer f.mu.Unlock()
	f.methodsCalled = append(f.methodsCalled, m)
}
func (f *fakeHelmOps) Install() error   { f.record("Install"); return f.installErr }
func (f *fakeHelmOps) Upgrade() error   { f.record("Upgrade"); return f.upgradeErr }
func (f *fakeHelmOps) ApplyEnv() error  { f.record("ApplyEnv"); return f.applyEnvErr }
func (f *fakeHelmOps) Uninstall() error { f.record("Uninstall"); return f.uninstallErr }
func (f *fakeHelmOps) UninstallAll() error {
	f.record("UninstallAll")
	return f.uninstallAllErr
}
func (f *fakeHelmOps) RollBack() error { f.record("RollBack"); return nil }
func (f *fakeHelmOps) WaitForLaunch() (bool, error) {
	f.record("WaitForLaunch")
	return f.waitLaunchErr == nil, f.waitLaunchErr
}
func (f *fakeHelmOps) WaitForStartUp() (bool, error) {
	f.record("WaitForStartUp")
	return f.waitStartUpErr == nil, f.waitStartUpErr
}
func (f *fakeHelmOps) Scale(r int32) error {
	f.record("Scale")
	f.mu.Lock()
	f.scaleReplicasLog = append(f.scaleReplicasLog, r)
	f.mu.Unlock()
	return f.scaleErr
}

func (f *fakeHelmOps) called() []string {
	f.mu.Lock()
	defer f.mu.Unlock()
	out := make([]string, len(f.methodsCalled))
	copy(out, f.methodsCalled)
	return out
}

// fakeFactory records every construction call and returns a
// pre-canned fakeHelmOps. Optional errFor overrides make it easy
// to test the "construction failed" path per method.
type fakeFactory struct {
	mu    sync.Mutex
	ops   *fakeHelmOps
	err   error
	calls []factoryCall
}

type factoryCall struct {
	AppName    string
	Token      string
	Source     string
	MarketSrc  string
	KubeConfig *rest.Config
}

func (ff *fakeFactory) fn() helmops.HelmOpsFactory {
	return func(_ context.Context, kc *rest.Config, app *appcfg.ApplicationConfig, token string, opt appinstaller.Opt) (appinstaller.HelmOpsInterface, error) {
		ff.mu.Lock()
		defer ff.mu.Unlock()
		ff.calls = append(ff.calls, factoryCall{
			AppName:    app.AppName,
			Token:      token,
			Source:     opt.Source,
			MarketSrc:  opt.MarketSource,
			KubeConfig: kc,
		})
		if ff.err != nil {
			return nil, ff.err
		}
		return ff.ops, nil
	}
}

func (ff *fakeFactory) recorded() []factoryCall {
	ff.mu.Lock()
	defer ff.mu.Unlock()
	out := make([]factoryCall, len(ff.calls))
	copy(out, ff.calls)
	return out
}

// baseAM returns an AM with the minimum shape resolveAppCfg /
// resolveToken / resolveOpt need to succeed against fakeFactory.
func baseAMWithReplicas(t *testing.T, name string) *appv1.ApplicationManager {
	t.Helper()
	am := baseAM(t, name)
	replicas := appcfg.WorkloadReplicas{"app-x": 1}
	cfg := &appcfg.ApplicationConfig{
		AppName:          name,
		Namespace:        "user-x-" + name,
		OwnerName:        "user-x",
		WorkloadReplicas: &replicas,
	}
	am.Spec.Config = mustMarshalCfg(t, cfg)
	return am
}

func baseAM(t *testing.T, name string) *appv1.ApplicationManager {
	t.Helper()
	cfg := &appcfg.ApplicationConfig{
		AppName:   name,
		Namespace: "user-x-" + name,
		OwnerName: "user-x",
	}
	configJSON := mustMarshalCfg(t, cfg)
	return &appv1.ApplicationManager{
		ObjectMeta: metav1.ObjectMeta{
			Name: name,
			Annotations: map[string]string{
				"bytetrade.io/token":         "tok-" + name,
				"bytetrade.io/market-source": "annotation-market",
			},
		},
		Spec: appv1.ApplicationManagerSpec{
			AppName:      name,
			AppNamespace: "user-x-" + name,
			AppOwner:     "user-x",
			Config:       configJSON,
		},
	}
}

func mustMarshalCfg(t *testing.T, cfg *appcfg.ApplicationConfig) string {
	t.Helper()
	b, err := json.Marshal(cfg)
	if err != nil {
		t.Fatalf("marshal appCfg: %v", err)
	}
	return string(b)
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
	_ = helmops.New(nil, &rest.Config{})
}

func TestNew_NilKubeConfigPanics(t *testing.T) {
	defer func() {
		if r := recover(); r == nil {
			t.Fatal("expected panic on nil kubeConfig")
		}
	}()
	_ = helmops.New(testutil.NewFakeClient(), nil)
}

func TestWith_NilOverrides_KeepDefaults(t *testing.T) {
	o := helmops.New(testutil.NewFakeClient(), &rest.Config{},
		helmops.WithHelmOpsFactory(nil),
		helmops.WithGetVersion(nil),
		helmops.WithClock(nil),
		helmops.WithPollInterval(0),
		helmops.WithNamespaceWaitTimeout(0),
	)
	if o == nil {
		t.Fatal("New returned nil")
	}
}

// -----------------------------------------------------------------
// Config resolution: exercised via a single write method (Install),
// since all writes share buildHelmOps.
// -----------------------------------------------------------------

func TestBuildHelmOps_NilAM_Errors(t *testing.T) {
	ff := &fakeFactory{ops: &fakeHelmOps{}}
	o := helmops.New(testutil.NewFakeClient(), &rest.Config{},
		helmops.WithHelmOpsFactory(ff.fn()))
	res := o.Install(context.Background(), ports.WorkloadRequest{})
	if res.Err == nil {
		t.Fatal("expected error on nil AM")
	}
	if len(ff.recorded()) != 0 {
		t.Errorf("factory should not run on nil AM: %+v", ff.recorded())
	}
}

func TestBuildHelmOps_EmptySpecConfig_Errors(t *testing.T) {
	ff := &fakeFactory{ops: &fakeHelmOps{}}
	o := helmops.New(testutil.NewFakeClient(), &rest.Config{},
		helmops.WithHelmOpsFactory(ff.fn()))
	am := &appv1.ApplicationManager{
		ObjectMeta: metav1.ObjectMeta{Name: "app-x"},
	}
	res := o.Install(context.Background(), ports.WorkloadRequest{AM: am})
	if res.Err == nil {
		t.Fatal("expected error on empty AM.Spec.Config with no req.AppCfg")
	}
	if len(ff.recorded()) != 0 {
		t.Errorf("factory should not run on empty spec: %+v", ff.recorded())
	}
}

func TestBuildHelmOps_MalformedSpecConfig_Errors(t *testing.T) {
	ff := &fakeFactory{ops: &fakeHelmOps{}}
	o := helmops.New(testutil.NewFakeClient(), &rest.Config{},
		helmops.WithHelmOpsFactory(ff.fn()))
	am := &appv1.ApplicationManager{
		ObjectMeta: metav1.ObjectMeta{Name: "app-x"},
		Spec: appv1.ApplicationManagerSpec{
			Config: "not-valid-json",
		},
	}
	res := o.Install(context.Background(), ports.WorkloadRequest{AM: am})
	if res.Err == nil {
		t.Fatal("expected error on malformed spec.Config")
	}
}

func TestBuildHelmOps_PrefersReqAppCfg_OverSpecConfig(t *testing.T) {
	ff := &fakeFactory{ops: &fakeHelmOps{}}
	o := helmops.New(testutil.NewFakeClient(), &rest.Config{},
		helmops.WithHelmOpsFactory(ff.fn()))
	am := baseAM(t, "app-x")
	// AM.Spec.Config unmarshals to AppName=app-x; req.AppCfg
	// overrides to app-y. Factory should see app-y.
	overrideCfg := &appcfg.ApplicationConfig{AppName: "app-y"}
	res := o.Install(context.Background(), ports.WorkloadRequest{
		AM:     am,
		AppCfg: overrideCfg,
	})
	if res.Err != nil {
		t.Fatalf("Install: %v", res.Err)
	}
	recs := ff.recorded()
	if len(recs) != 1 {
		t.Fatalf("factory calls: %d want 1", len(recs))
	}
	if recs[0].AppName != "app-y" {
		t.Errorf("appName: %q want app-y (req.AppCfg override lost)", recs[0].AppName)
	}
}

func TestBuildHelmOps_WrongReqAppCfgType_Errors(t *testing.T) {
	ff := &fakeFactory{ops: &fakeHelmOps{}}
	o := helmops.New(testutil.NewFakeClient(), &rest.Config{},
		helmops.WithHelmOpsFactory(ff.fn()))
	am := baseAM(t, "app-x")
	res := o.Install(context.Background(), ports.WorkloadRequest{
		AM:     am,
		AppCfg: "wrong-type",
	})
	if res.Err == nil {
		t.Fatal("expected error on wrong AppCfg type")
	}
}

func TestBuildHelmOps_PrefersReqKubeConfig(t *testing.T) {
	ff := &fakeFactory{ops: &fakeHelmOps{}}
	fallbackKC := &rest.Config{Host: "fallback"}
	overrideKC := &rest.Config{Host: "override"}
	o := helmops.New(testutil.NewFakeClient(), fallbackKC,
		helmops.WithHelmOpsFactory(ff.fn()))
	am := baseAM(t, "app-x")
	res := o.Install(context.Background(), ports.WorkloadRequest{
		AM:         am,
		KubeConfig: overrideKC,
	})
	if res.Err != nil {
		t.Fatalf("Install: %v", res.Err)
	}
	recs := ff.recorded()
	if recs[0].KubeConfig != overrideKC {
		t.Errorf("kubeConfig: got %p want %p (override)", recs[0].KubeConfig, overrideKC)
	}
}

func TestBuildHelmOps_FallsBackToAdapterKubeConfig(t *testing.T) {
	ff := &fakeFactory{ops: &fakeHelmOps{}}
	fallbackKC := &rest.Config{Host: "fallback"}
	o := helmops.New(testutil.NewFakeClient(), fallbackKC,
		helmops.WithHelmOpsFactory(ff.fn()))
	am := baseAM(t, "app-x")
	res := o.Install(context.Background(), ports.WorkloadRequest{AM: am})
	if res.Err != nil {
		t.Fatalf("Install: %v", res.Err)
	}
	recs := ff.recorded()
	if recs[0].KubeConfig != fallbackKC {
		t.Errorf("kubeConfig: got %p want %p (fallback)", recs[0].KubeConfig, fallbackKC)
	}
}

func TestBuildHelmOps_WrongReqKubeConfigType_Errors(t *testing.T) {
	ff := &fakeFactory{ops: &fakeHelmOps{}}
	o := helmops.New(testutil.NewFakeClient(), &rest.Config{},
		helmops.WithHelmOpsFactory(ff.fn()))
	am := baseAM(t, "app-x")
	res := o.Install(context.Background(), ports.WorkloadRequest{
		AM:         am,
		KubeConfig: "wrong-type",
	})
	if res.Err == nil {
		t.Fatal("expected error on wrong KubeConfig type")
	}
}

func TestBuildHelmOps_PrefersReqToken_OverAnnotation(t *testing.T) {
	ff := &fakeFactory{ops: &fakeHelmOps{}}
	o := helmops.New(testutil.NewFakeClient(), &rest.Config{},
		helmops.WithHelmOpsFactory(ff.fn()))
	am := baseAM(t, "app-x") // annotation token = "tok-app-x"
	res := o.Install(context.Background(), ports.WorkloadRequest{
		AM:    am,
		Token: "explicit-tok",
	})
	if res.Err != nil {
		t.Fatalf("Install: %v", res.Err)
	}
	if ff.recorded()[0].Token != "explicit-tok" {
		t.Errorf("token: %q want explicit-tok", ff.recorded()[0].Token)
	}
}

func TestBuildHelmOps_FallsBackToAnnotationToken(t *testing.T) {
	ff := &fakeFactory{ops: &fakeHelmOps{}}
	o := helmops.New(testutil.NewFakeClient(), &rest.Config{},
		helmops.WithHelmOpsFactory(ff.fn()))
	am := baseAM(t, "app-x")
	res := o.Install(context.Background(), ports.WorkloadRequest{AM: am})
	if res.Err != nil {
		t.Fatalf("Install: %v", res.Err)
	}
	if ff.recorded()[0].Token != "tok-app-x" {
		t.Errorf("token: %q want tok-app-x", ff.recorded()[0].Token)
	}
}

func TestBuildHelmOps_PrefersReqMarketSource_OverAnnotation(t *testing.T) {
	ff := &fakeFactory{ops: &fakeHelmOps{}}
	o := helmops.New(testutil.NewFakeClient(), &rest.Config{},
		helmops.WithHelmOpsFactory(ff.fn()))
	am := baseAM(t, "app-x") // annotation MarketSource = "annotation-market"
	res := o.Install(context.Background(), ports.WorkloadRequest{
		AM: am,
		Opt: ports.WorkloadOpt{
			MarketSource: "explicit-market",
			Source:       "market",
		},
	})
	if res.Err != nil {
		t.Fatalf("Install: %v", res.Err)
	}
	if ff.recorded()[0].MarketSrc != "explicit-market" {
		t.Errorf("MarketSource: %q want explicit-market", ff.recorded()[0].MarketSrc)
	}
	if ff.recorded()[0].Source != "market" {
		t.Errorf("Source: %q want market", ff.recorded()[0].Source)
	}
}

func TestBuildHelmOps_FallsBackToAnnotationMarketSource(t *testing.T) {
	ff := &fakeFactory{ops: &fakeHelmOps{}}
	o := helmops.New(testutil.NewFakeClient(), &rest.Config{},
		helmops.WithHelmOpsFactory(ff.fn()))
	am := baseAM(t, "app-x")
	res := o.Install(context.Background(), ports.WorkloadRequest{AM: am})
	if res.Err != nil {
		t.Fatalf("Install: %v", res.Err)
	}
	if ff.recorded()[0].MarketSrc != "annotation-market" {
		t.Errorf("MarketSource: %q want annotation-market", ff.recorded()[0].MarketSrc)
	}
}

func TestBuildHelmOps_FactoryError_Propagates(t *testing.T) {
	wantErr := errors.New("factory boom")
	ff := &fakeFactory{ops: &fakeHelmOps{}, err: wantErr}
	o := helmops.New(testutil.NewFakeClient(), &rest.Config{},
		helmops.WithHelmOpsFactory(ff.fn()))
	am := baseAM(t, "app-x")
	res := o.Install(context.Background(), ports.WorkloadRequest{AM: am})
	if !errors.Is(res.Err, wantErr) {
		t.Errorf("err: %v want %v", res.Err, wantErr)
	}
}

// -----------------------------------------------------------------
// Clock: a fake for the namespace-wait tests. Time only advances
// when tests call Advance.
// -----------------------------------------------------------------

type fakeClock struct {
	mu   sync.Mutex
	now  time.Time
	tick chan time.Time
}

func newFakeClock() *fakeClock {
	return &fakeClock{
		now:  time.Unix(0, 0),
		tick: make(chan time.Time, 1),
	}
}

func (fc *fakeClock) Now() time.Time {
	fc.mu.Lock()
	defer fc.mu.Unlock()
	return fc.now
}

// After returns the shared tick channel. Tests call Fire() to
// release the next After caller.
func (fc *fakeClock) After(_ time.Duration) <-chan time.Time {
	return fc.tick
}

// Fire advances time by `d` and releases whoever is waiting on After.
func (fc *fakeClock) Fire(d time.Duration) {
	fc.mu.Lock()
	fc.now = fc.now.Add(d)
	next := fc.now
	fc.mu.Unlock()
	fc.tick <- next
}

// Set overrides `now` without firing the tick channel — useful to
// jump past a deadline before triggering a probe.
func (fc *fakeClock) Set(t time.Time) {
	fc.mu.Lock()
	defer fc.mu.Unlock()
	fc.now = t
}
