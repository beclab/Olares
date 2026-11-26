package upgrade

import (
	"github.com/Masterminds/semver/v3"
	"github.com/beclab/Olares/daemon/pkg/cluster/state"
	"github.com/beclab/Olares/daemon/pkg/commands"
	"path/filepath"
	"testing"
)

func TestWatcherDistinguishesNewAuthorizationFromDownloadConfirmation(t *testing.T) {
	selected := state.UpgradeTarget{Version: *semver.MustParse("1.12.8"), RequestNonce: "old"}
	w := &upgradeWatcher{}
	w.startUpgrading(selected)
	confirmed := selected
	confirmed.Downloaded = true
	if w.runningAnotherTarget(confirmed) {
		t.Fatal("download state changed request identity")
	}
	newRequest := selected
	newRequest.RequestNonce = "new"
	if !w.runningAnotherTarget(newRequest) {
		t.Fatal("new request reused the in-flight attempt")
	}
	w.stopUpgrading()
	if w.runningAnotherTarget(newRequest) {
		t.Fatal("finished request still owns watcher")
	}
}

func TestOldDownloadCannotMarkNewRequestDownloaded(t *testing.T) {
	oldPath := commands.UPGRADE_TARGET_FILE
	commands.UPGRADE_TARGET_FILE = filepath.Join(t.TempDir(), "upgrade.target")
	t.Cleanup(func() { commands.UPGRADE_TARGET_FILE = oldPath })
	old := state.UpgradeTarget{Version: *semver.MustParse("1.12.8"), RequestNonce: "old"}
	selected := old
	selected.RequestNonce = "new"
	if err := selected.Save(); err != nil {
		t.Fatal(err)
	}
	if err := markDownloaded(old); err == nil {
		t.Fatal("old download changed a new request")
	}
	got, err := state.GetOlaresUpgradeTarget()
	if err != nil || got.Downloaded {
		t.Fatalf("new request = %+v, error=%v", got, err)
	}
}

// The control node only puts the target binaries in place and hands over; the
// checks and the image import are orchestrated stages that every node runs.
// Running them here as well would check and import twice on this machine and
// never on the others.
func TestControlNodeOnlyHandsTheUpgradeOver(t *testing.T) {
	want := []commands.Operations{
		commands.InstallCLI, commands.InstallOlaresd,
		commands.Upgrade, commands.RemoveUpgradeTarget,
	}
	if len(upgradePhases) != len(want) {
		t.Fatalf("%d upgrade phases, want %d", len(upgradePhases), len(want))
	}
	next := 0
	for i, phase := range upgradePhases {
		if got := phase.newCMD().OperationName(); got != want[i] {
			t.Errorf("phase %d = %s, want %s", i, got, want[i])
		}
		if phase.progressOffset != next {
			t.Errorf("phase %d starts at %d, want %d", i, phase.progressOffset, next)
		}
		next = phase.progressOffset + phase.progressSpan
	}
	if next != 100 {
		t.Errorf("upgrade phases fill %d%%, want 100", next)
	}
}
