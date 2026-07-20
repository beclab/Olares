package appstate

import (
	"context"
	"time"

	appsv1 "github.com/beclab/api/api/app.bytetrade.io/v1alpha1"

	"k8s.io/klog/v2"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

var _ OperationApp = &InstallFailedApp{}

type InstallFailedApp struct {
	*baseOperationApp
}

func NewInstallFailedApp(c client.Client,
	manager *appsv1.ApplicationManager) (StatefulApp, StateError) {

	return appFactory.New(c, manager, 0,
		func(c client.Client, manager *appsv1.ApplicationManager, ttl time.Duration) StatefulApp {
			return &InstallFailedApp{
				&baseOperationApp{
					ttl: ttl,
					baseStatefulApp: &baseStatefulApp{
						manager: manager,
						client:  c,
					},
				},
			}
		})
}

// Exec is the defensive retry loop that complements installing_app.go's
// synchronous cleanupAfterInstallFailure call. In the happy case the
// synchronous helper already confirmed Spec.AppNamespace is IsNotFound before
// the AM transitioned to InstallFailed, and the helper invocation here is
// effectively a no-op (helm release / NS / perms / Provider all gone, NS poll
// returns immediately on first IsNotFound). In the rare case where the
// synchronous wait timed out (NS finalizer stuck > 5min), every subsequent
// reconcile lands here and keeps retrying the same idempotent cleanup until
// the namespace truly disappears.
//
// This also covers the upgrade-time scenario where an AM was already sitting
// in InstallFailed under the old controller version (no synchronous cleanup):
// the first reconcile after the upgrade will run the full helper and re-align
// the AM with the new invariant.
//
// Cleanup failure or NS-still-present is reported as a returned error so the
// controller keeps re-driving Exec on backoff; once the helper returns nil the
// AM is left untouched in InstallFailed and the GC controller (§4.3) will pick
// it up after the retention period.
func (p *InstallFailedApp) Exec(ctx context.Context) (StatefulInProgressApp, error) {
	if err := cleanupAfterInstallFailure(ctx, p.client, p.manager); err != nil {
		klog.Warningf("InstallFailedApp.Exec cleanup for %s pending: %v", p.manager.Name, err)
		return nil, err
	}
	return nil, nil
}

func (p *InstallFailedApp) Cancel(ctx context.Context) error {
	return nil
}
