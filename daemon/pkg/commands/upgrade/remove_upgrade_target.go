package upgrade

import (
	"context"
	"fmt"
	"github.com/beclab/Olares/daemon/pkg/cluster/state"
	"os"

	"github.com/beclab/Olares/daemon/pkg/commands"
)

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
