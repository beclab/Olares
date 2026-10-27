package appstate

import (
	"context"

	"github.com/beclab/Olares/framework/app-service/pkg/appcfg"
	"github.com/beclab/Olares/framework/app-service/pkg/appinstaller"
	"github.com/beclab/Olares/framework/app-service/pkg/appinstaller/versioned"
	"github.com/beclab/Olares/framework/app-service/pkg/compute/validation"
	"github.com/beclab/Olares/framework/app-service/pkg/images"
	"github.com/beclab/Olares/framework/app-service/pkg/kubeblocks"
	"github.com/beclab/Olares/framework/app-service/pkg/kubesphere"
	"github.com/beclab/Olares/framework/app-service/pkg/utils"
	apputils "github.com/beclab/Olares/framework/app-service/pkg/utils/app"
	appsv1 "github.com/beclab/api/api/app.bytetrade.io/v1alpha1"

	kbopv1alpha1 "github.com/apecloud/kubeblocks/apis/operations/v1alpha1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/rest"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

// HelmOpsFactory builds a HelmOpsInterface for an app. It mirrors
// versioned.NewHelmOps so the production wiring is a direct assignment
// while tests can supply a fake.
type HelmOpsFactory func(ctx context.Context, kubeConfig *rest.Config,
	app *appcfg.ApplicationConfig, token string, options appinstaller.Opt) (appinstaller.HelmOpsInterface, error)

// MiddlewareOperator is the subset of kubeblocks.OperationOptions used by
// the state machine (start/stop of KubeBlocks-backed middleware).
type MiddlewareOperator interface {
	Start() error
	Stop() error
}

// MiddlewareOpFactory builds a MiddlewareOperator. Mirrors
// kubeblocks.NewOperation.
type MiddlewareOpFactory func(ctx context.Context, opsType kbopv1alpha1.OpsType,
	manager *appsv1.ApplicationManager, c client.Client) MiddlewareOperator

// ImageRefResolver resolves the image references that need to be
// downloaded for an app. It wraps the kubeConfig + BuildBaseHelmValues +
// GetRefsForImageManager chain so the whole (network-touching) block can
// be replaced in tests.
type ImageRefResolver func(ctx context.Context, am *appsv1.ApplicationManager,
	cfg *appcfg.ApplicationConfig) ([]appsv1.Ref, error)

// ExposePortSetter assigns expose ports to an app config. The production
// implementation (apputils.SetExposePorts) reaches a real clientset to learn
// which ports are already taken, so it is a seam: install runs it before any
// helm work, which would otherwise pin a real cluster into every install test.
type ExposePortSetter func(ctx context.Context, cfg *appcfg.ApplicationConfig,
	prevPortsMap map[string]int32) error

// InstallValidator runs the install-time runtime-pressure / compute-allocation
// validators. It is a seam so tests can drive both the accept and reject
// branches deterministically without simulating cluster pressure.
type InstallValidator func(ctx context.Context, in validation.Input) (validation.Decision, error)

// IsAdminFunc reports whether owner is a cluster admin in KubeSphere. It is
// a seam (defaulting to kubesphere.IsAdmin) so tests can answer this query
// deterministically; the production implementation hits the kubesphere user
// API and would otherwise force every state-flow test that touches the
// admin-aware install/upgrade/cancel paths to spin up a real cluster.
type IsAdminFunc func(ctx context.Context, kubeConfig *rest.Config, owner string) (bool, error)

// DownloadingCounter reports how many ApplicationManagers are currently in the
// Downloading state cluster-wide. It gates the Pending -> Downloading
// transition (at most one concurrent download). It is a seam: production
// counts through a live, uncached clientset (utils.GetClient), while tests
// override it to count via the injected controller-runtime fake client so the
// gate can be exercised without a real cluster.
type DownloadingCounter func(ctx context.Context) (int, error)

// Deps is the dependency container threaded through every stateful app.
// All external boundaries (helm, kubeconfig, kubeblocks, image manager,
// ref resolution) are fields so production wires the real implementations
// and tests wire fakes. Client is the controller-runtime client and
// Factory tracks in-progress operations (a single long-lived instance per
// controller; a fresh instance per test for isolation).
type Deps struct {
	Client               client.Client
	Factory              *statefulAppFactory
	KubeConfig           func() (*rest.Config, error)
	NewHelmOps           HelmOpsFactory
	NewMiddlewareOp      MiddlewareOpFactory
	NewImageManager      func(client.Client) images.ImageManager
	ResolveImageRefs     ImageRefResolver
	SetExposePorts       ExposePortSetter
	RunInstallValidation InstallValidator
	IsAdmin              IsAdminFunc
	CountDownloading     DownloadingCounter
}

// NewDeps assembles a Deps from explicit seams, allocating a fresh
// in-progress factory. It is the single place the unexported
// statefulAppFactory is created, which lets callers outside this package
// (notably integration tests in the controllers package) build a Deps wired
// to fakes without needing access to the factory type. DefaultDeps layers the
// production seams on top.
func NewDeps(
	c client.Client,
	kubeConfig func() (*rest.Config, error),
	newHelmOps HelmOpsFactory,
	newMiddlewareOp MiddlewareOpFactory,
	newImageManager func(client.Client) images.ImageManager,
	resolveImageRefs ImageRefResolver,
) Deps {
	return Deps{
		Client:           c,
		Factory:          newStatefulAppFactory(),
		KubeConfig:       kubeConfig,
		NewHelmOps:       newHelmOps,
		NewMiddlewareOp:  newMiddlewareOp,
		NewImageManager:  newImageManager,
		ResolveImageRefs: resolveImageRefs,
		// Install-specific seams default to the production implementations so
		// a Deps from NewDeps is always valid; tests override the fields.
		SetExposePorts:       apputils.SetExposePorts,
		RunInstallValidation: defaultRunInstallValidation,
		IsAdmin:              kubesphere.IsAdmin,
		CountDownloading:     defaultCountDownloading,
	}
}

func defaultRunInstallValidation(ctx context.Context, in validation.Input) (validation.Decision, error) {
	return validation.Run(ctx, in, validation.InstallRuntimePressureValidators()...)
}

// defaultCountDownloading is the production DownloadingCounter. It reads
// through a live, uncached clientset (utils.GetClient) so the concurrency gate
// observes the authoritative cluster state rather than a possibly-stale
// controller-runtime cache.
func defaultCountDownloading(ctx context.Context) (int, error) {
	clientset, err := utils.GetClient()
	if err != nil {
		return 0, err
	}
	apps, err := clientset.AppV1alpha1().ApplicationManagers().List(ctx, metav1.ListOptions{})
	if err != nil {
		return 0, err
	}
	count := 0
	for _, app := range apps.Items {
		if app.Status.State == appsv1.Downloading {
			count++
		}
	}
	return count, nil
}

// CountDownloadingViaClient counts Downloading ApplicationManagers through a
// controller-runtime client. It is the client.Client-based counterpart to
// defaultCountDownloading and lets tests wire Deps.CountDownloading to the
// injected fake client (Deps.Client) instead of a live clientset.
func CountDownloadingViaClient(ctx context.Context, c client.Client) (int, error) {
	var apps appsv1.ApplicationManagerList
	if err := c.List(ctx, &apps); err != nil {
		return 0, err
	}
	count := 0
	for _, app := range apps.Items {
		if app.Status.State == appsv1.Downloading {
			count++
		}
	}
	return count, nil
}

// DefaultDeps returns the production dependency set bound to the given
// controller-runtime client.
func DefaultDeps(c client.Client) Deps {
	return NewDeps(
		c,
		ctrl.GetConfig,
		versioned.NewHelmOps,
		func(ctx context.Context, opsType kbopv1alpha1.OpsType,
			manager *appsv1.ApplicationManager, c client.Client) MiddlewareOperator {
			return kubeblocks.NewOperation(ctx, opsType, manager, c)
		},
		images.NewImageManager,
		defaultResolveImageRefs,
	)
}

func defaultResolveImageRefs(ctx context.Context, am *appsv1.ApplicationManager,
	cfg *appcfg.ApplicationConfig) ([]appsv1.Ref, error) {
	kubeConfig, err := ctrl.GetConfig()
	if err != nil {
		return nil, err
	}
	values, err := appinstaller.BuildBaseHelmValues(ctx, kubeConfig, cfg, am.Spec.AppOwner, true)
	if err != nil {
		return nil, err
	}
	return GetRefsForImageManager(cfg, values)
}
