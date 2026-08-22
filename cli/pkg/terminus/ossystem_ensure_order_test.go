package terminus

import (
	"testing"

	"github.com/beclab/Olares/cli/pkg/core/task"
	"github.com/beclab/Olares/cli/pkg/preinstall"
)

// InstallOsSystem reads EnsureAppsPublished to decide the ensureApps Helm value,
// so a declaration published after it would reach the file system but never the
// Market deployment. That is invisible until an app the OS requires goes
// missing, and it took an upgrade to repair.
func TestFreshInstallPublishesEnsureAppsBeforeInstallingOsSystem(t *testing.T) {
	module := &InstallOsSystemModule{}
	module.Init()

	publish, install := -1, -1
	for index, current := range module.Tasks {
		local, ok := current.(*task.LocalTask)
		if !ok {
			continue
		}
		switch local.Action.(type) {
		case *preinstall.PublishEnsureAppsAction:
			publish = index
		case *InstallOsSystem:
			install = index
		}
	}

	if publish < 0 {
		t.Fatal("fresh install does not publish the ensured apps declaration")
	}
	if install < 0 {
		t.Fatal("InstallOsSystem task is missing")
	}
	if publish >= install {
		t.Fatalf("declaration is published after the chart is installed: publish=%d install=%d",
			publish, install)
	}
}
