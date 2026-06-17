package appstate

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"time"

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

// installFailureNSDeletionTimeout caps how long the synchronous install-failure
// cleanup blocks waiting for the app namespace to disappear. Beyond this point
// the AM still moves to InstallFailed but InstallFailedApp.Exec keeps retrying
// the same helper on every reconcile until the namespace is truly gone.
// Aligned with cancel poll timeouts (which use install TTL via opCtx); 10 min is
// plenty for normal NS finalizers and short enough to keep the controller
// responsive.
const installFailureNSDeletionTimeout = 10 * time.Minute

// cleanupAfterInstallFailure performs the same cluster-side teardown as
// HelmOps.Uninstall (=UninstallAll: PVCs + helm UninstallCharts + perm/Provider
// unregister + ClearMiddlewareRequests + ClearCache/Data + DeleteNamespace)
// plus compute allocation cleanup, AND THEN BLOCKS until the app namespace is
// actually gone (IsNotFound) — only then is the caller allowed to mark the AM
// InstallFailed.
//
// This mirrors installing_canceling_app.go's poll() which gates the
// InstallingCanceled transition on the same namespace deletion event.
//
// It is idempotent:
//   - releases that don't exist: UninstallCharts swallows ErrReleaseNotFound
//   - namespace that doesn't exist: DeleteNamespace + poll return immediately
//   - permissions/providers already unregistered: existing klog.Warning path
//
// Callers:
//   - installing_app.go at each transition into InstallFailed (synchronous;
//     fills the D/E post-helm validation / scale failure gaps and confirms
//     NS gone on C paths);
//   - InstallFailedApp.Exec on every reconcile (defensive retry; covers the
//     rare case where the synchronous wait above timed out).
//
// Returns nil iff NS is confirmed gone. On cleanup timeout returns
// context.DeadlineExceeded — callers in installing_app.go still proceed to
// InstallFailed but with an explicit warning in Status.Message;
// InstallFailedApp.Exec will keep trying.
//
// manager.Spec.Config may be empty (e.g. failure happened before unmarshal):
// in that case only the compute-allocation cleanup runs and the NS poll
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

	pollCtx, cancel := context.WithTimeout(ctx, installFailureNSDeletionTimeout)
	defer cancel()
	return waitForNamespaceGone(pollCtx, c, manager.Spec.AppNamespace)
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

// waitForNamespaceGone polls every second until the named namespace returns
// IsNotFound, ctx is canceled, or ctx deadline is exceeded. Mirrors
// installingCancelInProgressApp.poll in spirit; kept as a package-private
// helper so failure-cleanup is not coupled to the cancel state machine.
func waitForNamespaceGone(ctx context.Context, c client.Client, name string) error {
	// First-shot check before starting the ticker: if the namespace is already
	// gone we can return immediately without burning a 1-second tick.
	var ns corev1.Namespace
	if err := c.Get(ctx, types.NamespacedName{Name: name}, &ns); apierrors.IsNotFound(err) {
		return nil
	}

	timer := time.NewTicker(time.Second)
	defer timer.Stop()
	for {
		select {
		case <-timer.C:
			err := c.Get(ctx, types.NamespacedName{Name: name}, &ns)
			if apierrors.IsNotFound(err) {
				return nil
			}
			if err != nil {
				klog.V(4).Infof("waitForNamespaceGone %s: transient get error: %v", name, err)
			}
		case <-ctx.Done():
			return ctx.Err()
		}
	}
}
