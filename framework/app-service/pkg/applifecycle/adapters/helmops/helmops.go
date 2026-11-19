package helmops

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	appv1 "github.com/beclab/api/api/app.bytetrade.io/v1alpha1"

	"helm.sh/helm/v3/pkg/action"
	"k8s.io/client-go/rest"
	"sigs.k8s.io/controller-runtime/pkg/client"

	"github.com/beclab/Olares/framework/app-service/pkg/apiserver/api"
	"github.com/beclab/Olares/framework/app-service/pkg/appcfg"
	"github.com/beclab/Olares/framework/app-service/pkg/appinstaller"
	"github.com/beclab/Olares/framework/app-service/pkg/appinstaller/versioned"
	"github.com/beclab/Olares/framework/app-service/pkg/applifecycle/ports"
	"github.com/beclab/Olares/framework/app-service/pkg/helm"
	apputils "github.com/beclab/Olares/framework/app-service/pkg/utils/app"
)

// Default constants for the poll loops and helm-op timeouts.
// Values match the legacy pkg/appstate wall-time constants so a
// swap from legacy to Runner does not change observable timing.
const (
	// defaultPollInterval matches the 1-second tick in
	// pkg/appstate/types.go:267 (namespace-wait PollImmediate).
	defaultPollInterval = 1 * time.Second

	// defaultNamespaceWaitTimeout matches the 30-minute deadline
	// on the same PollImmediate. Any shorter and heavily loaded
	// clusters routinely time out on legitimate namespace
	// tear-downs (many finalizers).
	defaultNamespaceWaitTimeout = 30 * time.Minute
)

// HelmOpsFactory constructs a versioned appinstaller.HelmOpsInterface
// for the given AM/config triple. Production uses defaultHelmOpsFactory
// which wraps versioned.NewHelmOps; tests inject a factory returning
// a stub HelmOpsInterface via WithHelmOpsFactory.
type HelmOpsFactory func(ctx context.Context, kubeConfig *rest.Config, app *appcfg.ApplicationConfig, token string, opt appinstaller.Opt) (appinstaller.HelmOpsInterface, error)

// GetVersionFn is the seam GetDeployedVersion delegates to. Default
// wires helm.InitConfig + apputils.GetDeployedReleaseVersion (both
// unit-testable in isolation, but composing them requires an
// apiserver — the seam side-steps that).
type GetVersionFn func(ctx context.Context, kubeConfig *rest.Config, namespace, release string) (chartVersion string, revision int, err error)

// Clock is the minimum time surface WaitForNamespaceDeleted needs.
// Mirrors ports.Clock's shape but is declared here (rather than
// imported) so this package stays a leaf that can be constructed
// without pulling in the ports package for a fake tick source.
type Clock interface {
	Now() time.Time
	After(d time.Duration) <-chan time.Time
}

// realClock is the production Clock implementation. Fields-less so
// the zero value works everywhere.
type realClock struct{}

func (realClock) Now() time.Time                         { return time.Now() }
func (realClock) After(d time.Duration) <-chan time.Time { return time.After(d) }

// Ops is the production ports.WorkloadOps adapter.
type Ops struct {
	c             client.Client
	kubeConfig    *rest.Config
	newHelmOps    HelmOpsFactory
	getVersion    GetVersionFn
	clock         Clock
	pollInterval  time.Duration
	nsWaitTimeout time.Duration
}

// Option customises an Ops at construction time.
type Option func(*Ops)

// WithHelmOpsFactory overrides the versioned Helm-ops factory. nil
// is a no-op (keep default) so tests overriding only GetVersion
// don't have to re-assert the production factory.
func WithHelmOpsFactory(f HelmOpsFactory) Option {
	return func(o *Ops) {
		if f != nil {
			o.newHelmOps = f
		}
	}
}

// WithGetVersion overrides the deployed-version reader. nil is a
// no-op.
func WithGetVersion(f GetVersionFn) Option {
	return func(o *Ops) {
		if f != nil {
			o.getVersion = f
		}
	}
}

// WithClock overrides the wall-clock. Production uses realClock{};
// tests use a fake clock to advance time deterministically in the
// WaitForNamespaceDeleted poll.
func WithClock(c Clock) Option {
	return func(o *Ops) {
		if c != nil {
			o.clock = c
		}
	}
}

// WithPollInterval tunes WaitForNamespaceDeleted's poll cadence.
// Values <= 0 are treated as "use default".
func WithPollInterval(d time.Duration) Option {
	return func(o *Ops) {
		if d > 0 {
			o.pollInterval = d
		}
	}
}

// WithNamespaceWaitTimeout tunes the WaitForNamespaceDeleted
// deadline. Values <= 0 are treated as "use default".
func WithNamespaceWaitTimeout(d time.Duration) Option {
	return func(o *Ops) {
		if d > 0 {
			o.nsWaitTimeout = d
		}
	}
}

// New constructs an Ops. `c` and `kubeConfig` are required — nil
// client would NPE inside every wait/read path; nil kubeConfig
// would NPE inside every helm write (the helm action pipeline
// derefs kubeConfig.Host). Both are unrecoverable at runtime.
func New(c client.Client, kubeConfig *rest.Config, opts ...Option) *Ops {
	if c == nil {
		panic("helmops.New: client is nil")
	}
	if kubeConfig == nil {
		panic("helmops.New: kubeConfig is nil")
	}
	o := &Ops{
		c:             c,
		kubeConfig:    kubeConfig,
		newHelmOps:    defaultHelmOpsFactory,
		getVersion:    defaultGetVersion,
		clock:         realClock{},
		pollInterval:  defaultPollInterval,
		nsWaitTimeout: defaultNamespaceWaitTimeout,
	}
	for _, opt := range opts {
		opt(o)
	}
	return o
}

// defaultHelmOpsFactory is the production HelmOpsFactory —
// versioned.NewHelmOps dispatches on appCfg.APIVersion internally
// (v1 / v2 / v3), so this adapter does NOT re-implement the
// dispatch; it just forwards.
func defaultHelmOpsFactory(ctx context.Context, kubeConfig *rest.Config, app *appcfg.ApplicationConfig, token string, opt appinstaller.Opt) (appinstaller.HelmOpsInterface, error) {
	return versioned.NewHelmOps(ctx, kubeConfig, app, token, opt)
}

// defaultGetVersion composes helm.InitConfig +
// apputils.GetDeployedReleaseVersion. Kept as a top-level func
// (rather than an anonymous closure in New) so any signature drift
// on either helper breaks the build here, not at the first
// GetDeployedVersion call site.
func defaultGetVersion(_ context.Context, kubeConfig *rest.Config, namespace, release string) (string, int, error) {
	actionConfig, _, err := helm.InitConfig(kubeConfig, namespace)
	if err != nil {
		return "", 0, fmt.Errorf("helmops: init helm action config: %w", err)
	}
	return apputils.GetDeployedReleaseVersion(actionConfig, release)
}

// buildHelmOps is the shared config-resolution + factory-invocation
// path for every WRITE method. See doc.go "Runner-side request
// today vs future" for the four-field fallback rationale.
//
// Returns a HelmOpsInterface ready to invoke, or an error the
// caller must surface via WorkloadResult{Err: err}.
func (o *Ops) buildHelmOps(ctx context.Context, req ports.WorkloadRequest) (appinstaller.HelmOpsInterface, error) {
	if req.AM == nil {
		return nil, fmt.Errorf("helmops: WorkloadRequest.AM is nil")
	}
	appCfg, err := resolveAppCfg(req)
	if err != nil {
		return nil, err
	}
	kubeConfig, err := resolveKubeConfig(req, o.kubeConfig)
	if err != nil {
		return nil, err
	}
	token := resolveToken(req)
	opt := resolveOpt(req)
	return o.newHelmOps(ctx, kubeConfig, appCfg, token, opt)
}

// resolveAppCfg pulls *appcfg.ApplicationConfig from the request,
// falling back to json-unmarshaling AM.Spec.Config when the request
// carries none (the current Runner state). Legacy call sites do the
// same unmarshal inline; consolidating it here means the adapter
// works before AND after the Runner learns to populate req.AppCfg.
func resolveAppCfg(req ports.WorkloadRequest) (*appcfg.ApplicationConfig, error) {
	if req.AppCfg != nil {
		cfg, ok := req.AppCfg.(*appcfg.ApplicationConfig)
		if !ok || cfg == nil {
			return nil, fmt.Errorf("helmops: WorkloadRequest.AppCfg is %T, want *appcfg.ApplicationConfig", req.AppCfg)
		}
		return cfg, nil
	}
	if req.AM == nil || req.AM.Spec.Config == "" {
		return nil, fmt.Errorf("helmops: cannot resolve AppCfg (empty AM.Spec.Config and nil req.AppCfg)")
	}
	var cfg appcfg.ApplicationConfig
	if err := json.Unmarshal([]byte(req.AM.Spec.Config), &cfg); err != nil {
		return nil, fmt.Errorf("helmops: unmarshal AM.Spec.Config: %w", err)
	}
	return &cfg, nil
}

// resolveKubeConfig prefers the per-request KubeConfig (future
// Runner behaviour) and falls back to the adapter-injected one.
func resolveKubeConfig(req ports.WorkloadRequest, fallback *rest.Config) (*rest.Config, error) {
	if req.KubeConfig != nil {
		kc, ok := req.KubeConfig.(*rest.Config)
		if !ok || kc == nil {
			return nil, fmt.Errorf("helmops: WorkloadRequest.KubeConfig is %T, want *rest.Config", req.KubeConfig)
		}
		return kc, nil
	}
	if fallback == nil {
		return nil, fmt.Errorf("helmops: no KubeConfig on request and no fallback")
	}
	return fallback, nil
}

// resolveToken uses req.Token if non-empty, else falls back to the
// AM's AppTokenKey annotation. Matches every legacy call site
// verbatim (they read the annotation directly).
func resolveToken(req ports.WorkloadRequest) string {
	if req.Token != "" {
		return req.Token
	}
	if req.AM == nil || req.AM.Annotations == nil {
		return ""
	}
	return req.AM.Annotations[api.AppTokenKey]
}

// resolveOpt maps req.Opt → appinstaller.Opt. MarketSource falls
// back to the AM annotation via appcfg.GetMarketSource, mirroring
// every legacy call site (they do the same fallback inline).
func resolveOpt(req ports.WorkloadRequest) appinstaller.Opt {
	opt := appinstaller.Opt{
		Source:             req.Opt.Source,
		MarketSource:       req.Opt.MarketSource,
		SkipWaitForStartUp: req.Opt.SkipWaitForStartUp,
	}
	if opt.MarketSource == "" {
		opt.MarketSource = appcfg.GetMarketSource(req.AM)
	}
	return opt
}

// Compile-time interface guards. If the port signature drifts, or
// if versioned.NewHelmOps starts returning something that no
// longer satisfies HelmOpsInterface, the build breaks HERE rather
// than at first runtime dispatch.
var (
	_ ports.WorkloadOps = (*Ops)(nil)
	_ HelmOpsFactory    = defaultHelmOpsFactory
	_ GetVersionFn      = defaultGetVersion
	_ Clock             = realClock{}
)

// keep the appv1 import bound so a future refactor that drops the
// req.AM reference in this file does not accidentally break the
// var _ ports.WorkloadOps guard.
var _ = (*appv1.ApplicationManager)(nil)

// keep action import bound (used in default_get_version through
// apputils.GetDeployedReleaseVersion's exported signature) so a
// vendor bump that renames helm/v3/pkg/action fails here.
var _ = (*action.Configuration)(nil)
