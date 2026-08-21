package upgrade

import (
	"context"
	"fmt"

	"github.com/Masterminds/semver/v3"
	"github.com/beclab/Olares/daemon/pkg/cluster/state"
	"github.com/beclab/Olares/daemon/pkg/commands"
	"k8s.io/klog/v2"
)

// nodePreparePhases bring one machine to a target version's binaries: the
// packages, the new olares-cli, the new olaresd, and the images.
//
// It is the shared definition of that work and nothing else — see
// ReleaseDownloadPhases and ReleaseAdoptPhases. What the control node's
// watcher adds around the same two sequences is not a node's business:
// NewPreCheck asks cluster-level questions the orchestrator has already
// answered for every node at once, and NewUpgrade is exactly what this is
// preparing the node to be told to do, one stage at a time, by the control
// node.
var nodePreparePhases = append(
	append([]func() commands.Interface{}, ReleaseDownloadPhases...),
	ReleaseAdoptPhases...,
)

// PrepareNode fetches and installs the target version's binaries on the
// machine it runs on. It is what the control node dispatches to each compute
// node before any stage of the upgrade plan.
//
// It very often does not return. NewInstallOlaresd restarts olaresd, and this
// runs inside olaresd, so the process ends in the middle of the sequence — the
// same way it already does on the control node, where the watcher picks the
// phases back up afterwards. Here the stage record left saying "running" is
// settled as interrupted by the daemon that replaces this one, the control
// node dispatches the stage again, and the second run finds the packages
// downloaded and both binaries already at the target version and completes.
// Every phase is idempotent for that reason; none of them may become one that
// is not.
func PrepareNode(ctx context.Context, version string) error {
	v, err := semver.NewVersion(version)
	if err != nil {
		return fmt.Errorf("invalid target version %q: %v", version, err)
	}
	target := state.UpgradeTarget{Version: *v}

	for _, newCMD := range nodePreparePhases {
		cmd := newCMD()
		name := string(cmd.OperationName())
		klog.Infof("preparing this node for %s: %s", version, name)

		if err := runPhaseToCompletion(ctx, cmd, target); err != nil {
			return fmt.Errorf("%s: %v", name, err)
		}
	}
	return nil
}

// runPhaseToCompletion executes one phase and waits for it, whether it reports
// completion straight away or through a progress channel.
func runPhaseToCompletion(ctx context.Context, cmd commands.Interface, target state.UpgradeTarget) error {
	res, err := cmd.Execute(ctx, target)
	if err != nil {
		return err
	}
	execution, ok := res.(ExecutionRes)
	if !ok {
		return fmt.Errorf("unexpected result type")
	}
	if execution.Finished() {
		return nil
	}

	var progress int
	for progress < commands.ProgressNumFinished {
		select {
		case <-ctx.Done():
			return ctx.Err()
		case p, ok := <-execution.Progress():
			if !ok {
				if progress != commands.ProgressNumFinished {
					return fmt.Errorf("did not run to completion")
				}
				return nil
			}
			if p > progress {
				progress = p
			}
		}
	}
	return nil
}
