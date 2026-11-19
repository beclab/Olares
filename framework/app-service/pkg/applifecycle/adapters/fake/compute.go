package fake

import (
	"context"
	"sync"

	"github.com/beclab/Olares/framework/app-service/pkg/applifecycle/ports"
)

// ComputeOps is an in-memory ports.ComputeOps. Default: every
// mutating call succeeds no-op; ManagesSharedServer returns false;
// EnsureDeletedForTarget returns (true, nil) — the "cleaned" happy
// path. Override per method via ...Func hooks.
type ComputeOps struct {
	mu    sync.Mutex
	calls []ComputeCall

	DeleteForAppFunc           func(ctx context.Context, appName, owner string) error
	DeleteForTargetFunc        func(ctx context.Context, appCfg any, includeSharedServer bool) error
	EnsureDeletedForTargetFunc func(ctx context.Context, appCfg any, uninstallAll bool) (bool, error)
	ManagesSharedServerFunc    func(ctx context.Context, appCfg any) (bool, error)
}

// ComputeCall is one recorded invocation. Method is the port method
// name; App/Owner populated where relevant. AppCfg / bool flags are
// intentionally NOT recorded — inspect via the hook if precise
// arguments matter.
type ComputeCall struct {
	Method string
	App    string
	Owner  string
}

// Calls returns a copy of the recorded call log.
func (c *ComputeOps) Calls() []ComputeCall {
	c.mu.Lock()
	defer c.mu.Unlock()
	out := make([]ComputeCall, len(c.calls))
	copy(out, c.calls)
	return out
}

func (c *ComputeOps) recordCall(call ComputeCall) {
	c.mu.Lock()
	defer c.mu.Unlock()
	c.calls = append(c.calls, call)
}

func (c *ComputeOps) DeleteForApp(ctx context.Context, appName, owner string) error {
	c.recordCall(ComputeCall{Method: "DeleteForApp", App: appName, Owner: owner})
	if c.DeleteForAppFunc != nil {
		return c.DeleteForAppFunc(ctx, appName, owner)
	}
	return nil
}

func (c *ComputeOps) DeleteForTarget(ctx context.Context, appCfg any, includeSharedServer bool) error {
	c.recordCall(ComputeCall{Method: "DeleteForTarget"})
	if c.DeleteForTargetFunc != nil {
		return c.DeleteForTargetFunc(ctx, appCfg, includeSharedServer)
	}
	return nil
}

func (c *ComputeOps) EnsureDeletedForTarget(ctx context.Context, appCfg any, uninstallAll bool) (bool, error) {
	c.recordCall(ComputeCall{Method: "EnsureDeletedForTarget"})
	if c.EnsureDeletedForTargetFunc != nil {
		return c.EnsureDeletedForTargetFunc(ctx, appCfg, uninstallAll)
	}
	return true, nil
}

func (c *ComputeOps) ManagesSharedServer(ctx context.Context, appCfg any) (bool, error) {
	if c.ManagesSharedServerFunc != nil {
		return c.ManagesSharedServerFunc(ctx, appCfg)
	}
	return false, nil
}

// Interface conformance guard.
var _ ports.ComputeOps = (*ComputeOps)(nil)
