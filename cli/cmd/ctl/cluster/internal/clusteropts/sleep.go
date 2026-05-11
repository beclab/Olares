package clusteropts

import (
	"context"
	"time"
)

// SleepContext blocks for `d` or until `ctx` is cancelled, whichever
// comes first. Returns ctx.Err() on cancellation, nil on a clean
// timer fire.
//
// Used by every cluster polling loop (`cluster pod logs --follow`,
// `cluster pod get --watch`, `cluster workload rollout-status
// --watch`, `cluster application status --watch`, ...) so they share
// one definition instead of three identical copies. Same shape as
// cli/cmd/ctl/market/watch.go::sleepOrCancel — duplicated rather
// than cross-package-imported because the market helper is private
// to its package and the cluster tree shouldn't reach across.
func SleepContext(ctx context.Context, d time.Duration) error {
	t := time.NewTimer(d)
	defer t.Stop()
	select {
	case <-ctx.Done():
		return ctx.Err()
	case <-t.C:
		return nil
	}
}
