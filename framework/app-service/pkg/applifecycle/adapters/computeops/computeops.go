// Package computeops is the real ports.ComputeOps, backed by the
// ConfigMap-based allocation store in pkg/compute.
//
// Every method is a mechanical pass-through to the legacy function
// listed in ports/compute.go's call-site table; the only work the
// adapter does is casting WorkloadRequest-style `any` app configs to
// the concrete *appcfg.ApplicationConfig the legacy signatures take.
// That cast is the price of keeping the ports package a leaf module
// (see the design note on ports.WorkloadRequest.AppCfg).
package computeops

import (
	"context"
	"fmt"

	"sigs.k8s.io/controller-runtime/pkg/client"

	"github.com/beclab/Olares/framework/app-service/pkg/appcfg"
	"github.com/beclab/Olares/framework/app-service/pkg/applifecycle/ports"
	"github.com/beclab/Olares/framework/app-service/pkg/compute"
)

// Ops implements ports.ComputeOps against a controller-runtime
// client. Stateless beyond the client, so a single instance is shared
// by the whole Runner.
type Ops struct {
	c client.Client
}

// New constructs an Ops over the given client. The client should be
// the manager's, so reads hit the shared informer cache.
func New(c client.Client) *Ops { return &Ops{c: c} }

// DeleteForApp implements ports.ComputeOps. Legacy:
// compute.DeleteAllocationsForApp (installing_app.go's helm-failure
// branch, install_failure_cleanup.go, installing_canceling_app.go).
func (o *Ops) DeleteForApp(ctx context.Context, appName, owner string) error {
	return compute.DeleteAllocationsForApp(ctx, o.c, appName, owner)
}

// DeleteForTarget implements ports.ComputeOps. Legacy:
// compute.DeleteAllocationsForComputeTarget (suspending_app.go,
// suspend_failed_app.go, uninstalling_app.go).
func (o *Ops) DeleteForTarget(ctx context.Context, appCfg any, includeSharedServer bool) error {
	cfg, err := castAppCfg(appCfg)
	if err != nil {
		return err
	}
	return compute.DeleteAllocationsForComputeTarget(ctx, o.c, cfg, includeSharedServer)
}

// EnsureDeletedForTarget implements ports.ComputeOps. Legacy:
// compute.EnsureAllocationsDeletedForComputeTarget (stopped_app.go,
// uninstalled_app.go, types.go's forceDeleteApp).
//
// The convergent (cleaned bool, err error) shape is the legacy
// function's own — callers poll until cleaned is true. The Runner
// owns that loop, bounded by the CallWorkload timeout.
func (o *Ops) EnsureDeletedForTarget(ctx context.Context, appCfg any, uninstallAll bool) (bool, error) {
	cfg, err := castAppCfg(appCfg)
	if err != nil {
		return false, err
	}
	return compute.EnsureAllocationsDeletedForComputeTarget(ctx, o.c, cfg, uninstallAll)
}

// ManagesSharedServer implements ports.ComputeOps. Legacy:
// compute.ManagesSharedServer (installing_app.go's post-helm
// rejection branch, where it decides whether stopping the app has to
// bring down the shared server with it).
func (o *Ops) ManagesSharedServer(ctx context.Context, appCfg any) (bool, error) {
	cfg, err := castAppCfg(appCfg)
	if err != nil {
		return false, err
	}
	return compute.ManagesSharedServer(ctx, o.c, cfg)
}

// castAppCfg unwraps the port's `any` app config. A nil config is
// legal and forwarded as a typed nil — every legacy function above
// nil-checks and no-ops, so the adapter does not add a second layer
// of guards that could disagree with them.
func castAppCfg(appCfg any) (*appcfg.ApplicationConfig, error) {
	if appCfg == nil {
		return nil, nil
	}
	cfg, ok := appCfg.(*appcfg.ApplicationConfig)
	if !ok {
		return nil, fmt.Errorf("computeops: expected *appcfg.ApplicationConfig, got %T", appCfg)
	}
	return cfg, nil
}

// Interface conformance guard.
var _ ports.ComputeOps = (*Ops)(nil)
