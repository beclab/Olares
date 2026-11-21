package upgrade

import (
	"context"
	"errors"
	"fmt"
	"os"

	"github.com/beclab/Olares/daemon/pkg/cluster/state"
	"github.com/beclab/Olares/daemon/pkg/commands"
)

// OperatorCancel is passed by the operator's cancel. Removing the target is
// also how the watcher tidies up after an upgrade, which is never refused.
type OperatorCancel struct{}

var ErrUpgradeCommitted = errors.New("the upgrade has started changing the system and can no longer be cancelled")

type removeUpgradeTarget struct {
	commands.Operation
}

var _ commands.Interface = &removeUpgradeTarget{}

func NewRemoveUpgradeTarget() commands.Interface {
	return &removeUpgradeTarget{
		Operation: commands.Operation{
			Name: commands.RemoveUpgradeTarget,
		},
	}
}

func (i *removeUpgradeTarget) Execute(ctx context.Context, p any) (res any, err error) {
	if _, ok := p.(OperatorCancel); ok {
		if target, terr := state.GetOlaresUpgradeTarget(); terr == nil && target != nil && ClusterUpgradeCommitted(*target) {
			return nil, ErrUpgradeCommitted
		}
	}
	if expected, ok := p.(state.UpgradeTarget); ok {
		selected, err := state.GetOlaresUpgradeTarget()
		if err != nil {
			return nil, err
		}
		if selected == nil || !selected.SameRequest(expected) {
			return nil, fmt.Errorf("upgrade target changed before this attempt finished")
		}
	}
	// Read before deleting: the target names the version, and the version is
	// what identifies the orchestrated upgrade to stop. On a single node
	// there is nothing to stop and this does nothing.
	if target, terr := state.GetOlaresUpgradeTarget(); terr == nil && target != nil {
		StopClusterUpgrade(*target)
	}

	err = removeUpgradeTargetFile()
	if err != nil {
		return nil, err
	}

	err = state.CheckCurrentStatus(context.Background())
	if err != nil {
		return nil, fmt.Errorf("failed to refresh system state: %v", err)
	}

	return NewExecutionRes(true, nil), nil
}

func removeUpgradeTargetFile() error {
	if err := os.Remove(commands.UPGRADE_TARGET_FILE); err != nil && !os.IsNotExist(err) {
		return err
	}
	return nil
}
