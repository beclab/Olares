package upgrade

import (
	"context"
	"errors"

	"github.com/beclab/Olares/daemon/pkg/cluster/state"
	"github.com/beclab/Olares/daemon/pkg/commands"
)

type upgrade struct {
	commands.Operation
}

var _ commands.Interface = &upgrade{}

func NewUpgrade() commands.Interface {
	return &upgrade{
		Operation: commands.Operation{
			Name: commands.Upgrade,
		},
	}
}

// Execute hands the upgrade to the orchestrator on this machine, which runs
// the stages on every node, this one included.
//
// An error is not a reason to upgrade this machine alone: it means the
// cluster's shape is unknown, and upgrading one machine on that basis is how a
// cluster ends up split across two versions. The watcher retries.
func (i *upgrade) Execute(ctx context.Context, p any) (res any, err error) {
	target, ok := p.(state.UpgradeTarget)
	if !ok {
		return nil, errors.New("invalid param")
	}
	m, err := upgradeOrchestrator(ctx)
	if err != nil {
		return nil, err
	}
	return i.upgradeThroughOrchestrator(ctx, m, target)
}

type removeTarget struct {
	commands.Operation
}

var _ commands.Interface = &removeTarget{}

func NewRemoveTarget() commands.Interface {
	return &removeTarget{
		Operation: commands.Operation{
			Name: commands.RemoveUpgradeTarget,
		},
	}
}

func (i *removeTarget) Execute(ctx context.Context, p any) (res any, err error) {
	upgradeRemove := NewRemoveUpgradeTarget()
	return upgradeRemove.Execute(ctx, p)
}
