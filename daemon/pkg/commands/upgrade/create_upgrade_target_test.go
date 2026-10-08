package upgrade

import (
	"context"
	"os"
	"path/filepath"
	"testing"

	"github.com/Masterminds/semver/v3"
	"github.com/beclab/Olares/daemon/pkg/cluster/state"
	"github.com/beclab/Olares/daemon/pkg/commands"
)

func TestSignedTargetReusesNonceUntilItIsRemoved(t *testing.T) {
	oldPath := commands.UPGRADE_TARGET_FILE
	oldTrigger := state.StateTrigger
	commands.UPGRADE_TARGET_FILE = filepath.Join(t.TempDir(), "upgrade.target")
	state.StateTrigger = make(chan struct{}, 3)
	t.Cleanup(func() {
		commands.UPGRADE_TARGET_FILE = oldPath
		state.StateTrigger = oldTrigger
	})

	version := semver.MustParse("1.12.8")
	req := state.UpgradeTarget{Version: *version, RequestNonce: "untrusted-caller-value"}
	create := NewCreateUpgradeTarget()
	if _, err := create.Execute(context.Background(), req); err != nil {
		t.Fatal(err)
	}
	first, err := state.GetOlaresUpgradeTarget()
	if err != nil || first == nil || first.RequestNonce == "" || first.RequestNonce == req.RequestNonce {
		t.Fatalf("first target = %+v, error = %v", first, err)
	}
	first.Downloaded = true
	if err := first.Save(); err != nil {
		t.Fatal(err)
	}
	if _, err := create.Execute(context.Background(), req); err != nil {
		t.Fatal(err)
	}
	repeated, err := state.GetOlaresUpgradeTarget()
	if err != nil || repeated.RequestNonce != first.RequestNonce || !repeated.Downloaded {
		t.Fatalf("repeat target = %+v, error = %v", repeated, err)
	}
	if err := os.Remove(commands.UPGRADE_TARGET_FILE); err != nil {
		t.Fatal(err)
	}
	if _, err := create.Execute(context.Background(), req); err != nil {
		t.Fatal(err)
	}
	newTarget, err := state.GetOlaresUpgradeTarget()
	if err != nil || newTarget.RequestNonce == first.RequestNonce {
		t.Fatalf("new authorization target = %+v, error = %v", newTarget, err)
	}
}
