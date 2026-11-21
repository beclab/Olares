package upgrade

import (
	"context"
	"errors"
	"path/filepath"
	"testing"

	"github.com/Masterminds/semver/v3"
	"github.com/beclab/Olares/daemon/pkg/cluster/clusterop"
	"github.com/beclab/Olares/daemon/pkg/cluster/state"
	"github.com/beclab/Olares/daemon/pkg/commands"
)

func TestOperatorCannotCancelACommittedUpgrade(t *testing.T) {
	oldPath, oldCommitted := commands.UPGRADE_TARGET_FILE, upgradeCommitted
	commands.UPGRADE_TARGET_FILE = filepath.Join(t.TempDir(), "upgrade.target")
	upgradeCommitted = func(state.UpgradeTarget) bool { return true }
	t.Cleanup(func() { commands.UPGRADE_TARGET_FILE, upgradeCommitted = oldPath, oldCommitted })

	target := state.UpgradeTarget{Version: *semver.MustParse("1.12.8"), RequestNonce: "n"}
	if err := target.Save(); err != nil {
		t.Fatal(err)
	}
	if _, err := NewRemoveUpgradeTarget().Execute(context.Background(), OperatorCancel{}); !errors.Is(err, ErrUpgradeCommitted) {
		t.Fatalf("cancel error = %v, want %v", err, ErrUpgradeCommitted)
	}
	if got, err := state.GetOlaresUpgradeTarget(); err != nil || got == nil {
		t.Fatalf("target after a refused cancel = %+v, %v", got, err)
	}
}

func TestRunningStepIsTheStageInProgress(t *testing.T) {
	op := clusterop.Operation{Steps: []clusterop.Step{
		{Name: clusterop.StepPrecheck, Status: clusterop.StepSucceeded},
		{Name: "bootstrap", Status: clusterop.StepSucceeded},
		{Name: "download", Status: clusterop.StepRunning},
		{Name: "precheck-node", Status: clusterop.StepPending},
	}}
	if got := runningStep(op); got != "download" {
		t.Errorf("runningStep = %q, want download", got)
	}
	op.Steps[2].Status = clusterop.StepSucceeded
	if got := runningStep(op); got != "" {
		t.Errorf("runningStep between stages = %q, want empty", got)
	}
}
