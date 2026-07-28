package preflight

import (
	"context"

	"github.com/beclab/Olares/cli/pkg/cmdutil"
)

// Re-exports of cmdutil version gates so existing settings callers keep
// compiling without a mass import rewrite. New command trees (e.g. download)
// should import cmdutil directly.

// MinVersionGate is an alias of cmdutil.MinVersionGate.
type MinVersionGate = cmdutil.MinVersionGate

// RemovedGate is an alias of cmdutil.RemovedGate.
type RemovedGate = cmdutil.RemovedGate

// RequireMinVersion delegates to cmdutil.RequireMinVersion.
func RequireMinVersion(ctx context.Context, f *cmdutil.Factory, gate MinVersionGate) error {
	return cmdutil.RequireMinVersion(ctx, f, gate)
}

// RejectIfRemoved delegates to cmdutil.RejectIfRemoved.
func RejectIfRemoved(ctx context.Context, f *cmdutil.Factory, gate RemovedGate) error {
	return cmdutil.RejectIfRemoved(ctx, f, gate)
}
