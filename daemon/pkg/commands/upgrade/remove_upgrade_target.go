package upgrade

import (
	"context"
	"os"

	"github.com/beclab/Olares/daemon/pkg/cluster/state"
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
	err = RemoveUpgradeFiles()
	if err != nil {
		return nil, err
	}

	state.CurrentState.UpgradingDownloadState = ""
	state.CurrentState.UpgradingDownloadStep = ""
	state.CurrentState.UpgradingDownloadProgress = ""
	state.CurrentState.UpgradingDownloadProgressNum = 0
	state.CurrentState.UpgradingDownloadError = ""

	state.StateTrigger <- struct{}{}

	return NewExecutionRes(true, nil), nil
}

func RemoveUpgradeFiles() error {
	// attempt to remove all files whether they exist or not (idempotent)
	files := []string{
		commands.UPGRADE_TARGET_FILE,
		commands.UPGRADE_DOWNLOADONLY_FILE,
		commands.UPGRADE_DOWNLOADED_FILE,
	}

	for _, file := range files {
		if err := os.Remove(file); err != nil && !os.IsNotExist(err) {
			return err
		}
	}

	return nil
}
