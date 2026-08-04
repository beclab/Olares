package appstate

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	"github.com/beclab/Olares/framework/app-service/pkg/apiserver/api"
	"github.com/beclab/Olares/framework/app-service/pkg/appcfg"
	"github.com/beclab/Olares/framework/app-service/pkg/appinstaller"
	"github.com/beclab/Olares/framework/app-service/pkg/compute"
	apputils "github.com/beclab/Olares/framework/app-service/pkg/utils/app"
	appsv1 "github.com/beclab/api/api/app.bytetrade.io/v1alpha1"

	"helm.sh/helm/v3/pkg/storage/driver"
	corev1 "k8s.io/api/core/v1"
	apierrors "k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/types"
	"k8s.io/klog/v2"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

// cleanupAfterInstallFailure performs the same cluster-side teardown as
// HelmOps.Uninstall (=UninstallAll: PVCs + helm UninstallCharts + perm/Provider
// unregister + ClearMiddlewareRequests + ClearCache/Data + DeleteNamespace)
// plus compute allocation cleanup, AND THEN single-shot CHECKS whether the
// app namespace is gone (IsNotFound). If the namespace is still present it
// returns an appstate.RequeueError so the controller re-enqueues the request
// after a short delay instead of the calling goroutine blocking on a long
// poll — with MaxConcurrentReconciles=1 a multi-minute synchronous wait
// would starve every other ApplicationManager.
//
// It is idempotent (safe to re-run on every requeue iteration):
//   - releases that don't exist: UninstallCharts swallows ErrReleaseNotFound
//   - namespace that doesn't exist: DeleteNamespace + check return immediately
//   - permissions/providers already unregistered: existing klog.Warning path
//
// Callers:
//   - installing_app.go at each transition into InstallFailed (runs inside
//     the InstallingApp.Exec goroutine, so a RequeueError there is just
//     logged and the transition to InstallFailed still proceeds; the
//     subsequent InstallFailedApp.Exec re-runs the helper);
//   - InstallFailedApp.Exec on every reconcile, which propagates the
//     RequeueError up so the reconciler honors the short backoff.
//
// Returns nil iff NS is confirmed gone (or no NS was created).
// manager.Spec.Config may be empty (e.g. failure happened before unmarshal):
// in that case only the compute-allocation cleanup runs and the NS check
// short-circuits because no namespace was created either.
func cleanupAfterInstallFailure(ctx context.Context, c client.Client, manager *appsv1.ApplicationManager) error {
	appCfg, cfgErr := parseAppConfig(manager)
	if cfgErr != nil {
		klog.Warningf("install-failure cleanup %s: parse app config: %v; skipping helm uninstall", manager.Name, cfgErr)
	}

	if appCfg != nil {
		if err := runHelmUninstallForFailure(ctx, manager, appCfg); err != nil {
			klog.Warningf("install-failure cleanup %s: helm uninstall: %v", manager.Name, err)
		}

		if err := compute.DeleteAllocationsForApp(ctx, c, appCfg.AppName, appCfg.OwnerName); err != nil {
			klog.Warningf("install-failure cleanup %s: compute alloc: %v", manager.Name, err)
		}
	}

	if apputils.IsProtectedNamespace(manager.Spec.AppNamespace) {
		return nil
	}
	if manager.Spec.AppNamespace == "" {
		return nil
	}

	// Issue the namespace delete ourselves so that even when helm
	// UninstallAll bailed earlier (e.g. release parse failure, missing
	// kubeconfig) we still kick off termination here. IsNotFound is
	// success (already gone); any other failure is surfaced immediately
	// so the caller doesn't wait the full poll timeout on a delete
	// that never went through.
	ns := &corev1.Namespace{
		ObjectMeta: metav1.ObjectMeta{Name: manager.Spec.AppNamespace},
	}
	if err := c.Delete(ctx, ns); err != nil && !apierrors.IsNotFound(err) {
		return fmt.Errorf("delete namespace %s: %w", manager.Spec.AppNamespace, err)
	}

	return waitForNamespaceGone(ctx, c, manager.Spec.AppNamespace)
}

// parseAppConfig unmarshals the JSON config snapshot persisted on the AM. It
// returns (nil, nil) when Spec.Config is empty so the caller can short-circuit
// helm cleanup without treating it as an error.
func parseAppConfig(manager *appsv1.ApplicationManager) (*appcfg.ApplicationConfig, error) {
	if manager.Spec.Config == "" {
		return nil, nil
	}
	var cfg appcfg.ApplicationConfig
	if err := json.Unmarshal([]byte(manager.Spec.Config), &cfg); err != nil {
		return nil, fmt.Errorf("unmarshal app config: %w", err)
	}
	return &cfg, nil
}

// runHelmUninstallForFailure builds a HelmOps and runs Uninstall (=UninstallAll
// per helm_ops_uninstall.go: PVC deletion → helm UninstallCharts → perm/Provider
// unregister → middleware requests → cache/data → namespace deletion). The call
// is idempotent: ErrReleaseNotFound is swallowed; an already-deleted NS turns
// the inner DeleteNamespace into a no-op.
func runHelmUninstallForFailure(ctx context.Context, manager *appsv1.ApplicationManager, appCfg *appcfg.ApplicationConfig) error {
	kubeConfig, err := getKubeConfig()
	if err != nil {
		return fmt.Errorf("get kube config: %w", err)
	}
	token := manager.Annotations[api.AppTokenKey]
	ops, err := newHelmOps(ctx, kubeConfig, appCfg, token, appinstaller.Opt{
		Source:       manager.Spec.Source,
		MarketSource: appcfg.GetMarketSource(manager),
	})
	if err != nil {
		return fmt.Errorf("build helm ops: %w", err)
	}
	if err := ops.Uninstall(); err != nil && !errors.Is(err, driver.ErrReleaseNotFound) {
		return err
	}
	return nil
}

// waitForNamespaceGone performs a single-shot check: it returns nil when the
// named namespace is gone (IsNotFound), an appstate.RequeueError when the
// namespace is still present (asking the reconciler to retry after a short
// delay), or a transient client error otherwise.
//
// Single-shot semantics keep the reconcile worker free: with
// MaxConcurrentReconciles=1, blocking here for minutes would freeze every
// other ApplicationManager. The caller (typically cleanupAfterInstallFailure)
// is expected to propagate the error so controller-runtime re-enqueues.
func waitForNamespaceGone(ctx context.Context, c client.Client, name string) error {
	var ns corev1.Namespace
	err := c.Get(ctx, types.NamespacedName{Name: name}, &ns)
	if apierrors.IsNotFound(err) {
		return nil
	}
	if err != nil {
		klog.V(4).Infof("waitForNamespaceGone %s: transient get error: %v", name, err)
		return err
	}
	klog.Infof("namespace %s still exists, requeueing in 5s", name)
	return NewWaitingInLine(5)
}
