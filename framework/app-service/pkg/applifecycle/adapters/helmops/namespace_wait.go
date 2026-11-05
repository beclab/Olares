package helmops

import (
	"context"
	"fmt"

	corev1 "k8s.io/api/core/v1"
	apierrors "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/types"

	"github.com/beclab/Olares/framework/app-service/pkg/applifecycle/ports"
	apputils "github.com/beclab/Olares/framework/app-service/pkg/utils/app"
)

// WaitForNamespaceDeleted implements ports.WorkloadOps. Legacy
// call sites:
//   - pkg/appstate/types.go:267        (uninstall path)
//   - pkg/appstate/uninstalling_app.go:101
//   - pkg/appstate/install_failure_cleanup.go:142
//
// Contract:
//
//   - Issue Delete on every non-protected target first. Pending /
//     Downloading cancel never run helm Uninstall, so this is the
//     only place the app namespace is actually torn down. Delete
//     is idempotent: NotFound is success (already gone, or helm
//     Uninstall already deleted it).
//   - Then poll `Primary` every pollInterval; a NotFound from Get
//     is success. Any other Get error is fatal. Namespaces with
//     finalizers stay visible after Delete until GC finishes.
//   - Also poll every namespace in `SharedNS` — each must reach
//     NotFound for success.
//   - Skip protected namespaces (apputils.IsProtectedNamespace):
//     these are shared-with-system (user-space, user-system, etc)
//     and MUST NOT be deleted or block the state machine.
//   - Overall deadline is req.Timeout, falling back to the
//     adapter's default (nsWaitTimeout, 30m).
//
// Returns WorkloadResult{Err} on timeout / delete error /
// non-NotFound Get error; zero WorkloadResult on success.
func (o *Ops) WaitForNamespaceDeleted(ctx context.Context, req ports.NamespaceWaitRequest) ports.WorkloadResult {
	// Filter out protected namespaces up front so the loop body
	// stays a pure "still exists?" check.
	targets := make([]string, 0, 1+len(req.SharedNS))
	if req.Primary != "" && !apputils.IsProtectedNamespace(req.Primary) {
		targets = append(targets, req.Primary)
	}
	for _, ns := range req.SharedNS {
		if ns == "" || apputils.IsProtectedNamespace(ns) {
			continue
		}
		targets = append(targets, ns)
	}
	if len(targets) == 0 {
		return ports.WorkloadResult{}
	}

	if err := o.deleteNamespaces(ctx, targets); err != nil {
		return ports.WorkloadResult{Err: err}
	}

	timeout := req.Timeout
	if timeout <= 0 {
		timeout = o.nsWaitTimeout
	}
	deadline := o.clock.Now().Add(timeout)

	// Do one immediate probe before entering the tick loop so
	// namespaces that are ALREADY gone return fast (mirrors
	// util.PollImmediate semantics used by legacy).
	if gone, err := o.probeNamespaces(ctx, targets); err != nil {
		return ports.WorkloadResult{Err: err}
	} else if gone {
		return ports.WorkloadResult{}
	}

	for {
		// Deadline check before waiting for the next tick so a
		// zero-timeout call never blocks.
		if !o.clock.Now().Before(deadline) {
			return ports.WorkloadResult{Err: fmt.Errorf("helmops: WaitForNamespaceDeleted timed out after %s (targets=%v)", timeout, targets)}
		}

		select {
		case <-ctx.Done():
			return ports.WorkloadResult{Err: ctx.Err()}
		case <-o.clock.After(o.pollInterval):
			gone, err := o.probeNamespaces(ctx, targets)
			if err != nil {
				return ports.WorkloadResult{Err: err}
			}
			if gone {
				return ports.WorkloadResult{}
			}
		}
	}
}

// probeNamespaces returns (allGone, err). allGone is true iff every
// target namespace Get returns NotFound. Any non-NotFound error is
// fatal.
//
// Extracted so the immediate-probe pass and the tick loop share
// exactly one call site — the legacy PollImmediate condition was
// coded inline twice, which made drift possible.
func (o *Ops) probeNamespaces(ctx context.Context, targets []string) (bool, error) {
	for _, ns := range targets {
		var probe corev1.Namespace
		err := o.c.Get(ctx, types.NamespacedName{Name: ns}, &probe)
		if err == nil {
			// Namespace still exists — not done.
			return false, nil
		}
		if !apierrors.IsNotFound(err) {
			return false, fmt.Errorf("helmops: get namespace %q: %w", ns, err)
		}
		// NotFound: this target is gone. Continue to the next.
	}
	return true, nil
}

// deleteNamespaces issues Delete for each target. NotFound is
// treated as already-gone success so cancel-after-uninstall and
// cancel-when-never-created share one path.
func (o *Ops) deleteNamespaces(ctx context.Context, targets []string) error {
	for _, name := range targets {
		ns := &corev1.Namespace{ObjectMeta: metav1.ObjectMeta{Name: name}}
		if err := o.c.Delete(ctx, ns); err != nil && !apierrors.IsNotFound(err) {
			return fmt.Errorf("helmops: delete namespace %q: %w", name, err)
		}
	}
	return nil
}

// waitInterval + waitDeadline are kept small so tests can assert
// against a stable time.Duration without importing this package's
// private constants. Not currently referenced but kept as
// symmetric with WithPollInterval / WithNamespaceWaitTimeout for
// the option boundary.
var (
	_ = defaultPollInterval
	_ = defaultNamespaceWaitTimeout
)
