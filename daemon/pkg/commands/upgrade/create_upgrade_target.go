package upgrade

import (
	"context"
	"errors"
	"fmt"
	"os"

	"github.com/beclab/Olares/daemon/pkg/cluster/state"
	"github.com/beclab/Olares/daemon/pkg/commands"
)

type UpgradeRequest struct {
	Version      string `json:"version"`
	DownloadOnly bool   `json:"downloadOnly"`
}

type createUpgradeTarget struct {
	commands.Operation
}

var _ commands.Interface = &createUpgradeTarget{}

func NewCreateUpgradeTarget() commands.Interface {
	return &createUpgradeTarget{
		Operation: commands.Operation{
			Name: commands.CreateUpgradeTarget,
		},
	}
}

func (i *createUpgradeTarget) Execute(ctx context.Context, p any) (res any, err error) {
	req, ok := p.(UpgradeRequest)
	if !ok {
		return nil, errors.New("invalid param")
	}

	if err := checkVersionConflicts(req.Version); err != nil {
		return nil, err
	}

	if err := createUpgradeTargetFile(req.Version); err != nil {
		return nil, fmt.Errorf("failed to create upgrade target: %v", err)
	}

	if req.DownloadOnly {
		if err := createUpgradeDownloadOnlyFile(); err != nil {
			return nil, fmt.Errorf("failed to create upgrade downloadonly file: %v", err)
		}
	} else {
		if err := removeUpgradeDownloadOnlyFile(); err != nil && !os.IsNotExist(err) {
			return nil, fmt.Errorf("failed to remove upgrade downloadonly file: %v", err)
		}
	}

	state.StateTrigger <- struct{}{}

	return NewExecutionRes(true, nil), nil
}

func checkVersionConflicts(version string) error {
	if state.CurrentState.UpgradingState == state.InProgress {
		return fmt.Errorf("system is currently upgrading")
	}

	upgradeTarget, err := state.GetOlaresUpgradeTarget()
	if err == nil && upgradeTarget != nil && upgradeTarget.Original() != version {
		return fmt.Errorf("different upgrade version %s already exists, please cancel it first", upgradeTarget.Original())
	}

	return nil
}

func createUpgradeTargetFile(version string) error {
	return os.WriteFile(commands.UPGRADE_TARGET_FILE, []byte(version), 0755)
}

func createUpgradeDownloadOnlyFile() error {
	return os.WriteFile(commands.UPGRADE_DOWNLOADONLY_FILE, []byte(""), 0755)
}

func removeUpgradeDownloadOnlyFile() error {
	return os.Remove(commands.UPGRADE_DOWNLOADONLY_FILE)
}
