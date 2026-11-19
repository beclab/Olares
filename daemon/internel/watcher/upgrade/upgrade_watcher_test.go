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
