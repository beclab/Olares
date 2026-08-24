package upgrade

import (
	"testing"

	"github.com/Masterminds/semver/v3"
	"github.com/beclab/Olares/cli/pkg/core/task"
	"github.com/beclab/Olares/cli/pkg/preinstall"
	"github.com/beclab/Olares/cli/pkg/storage"
	olversion "github.com/beclab/Olares/cli/version"
)

// Each release declares the apps it expects, so the publish belongs to every
// upgrade rather than to the one that introduced the mechanism: a later release
// that skipped it would leave Market maintaining the previous release's list.
func TestEveryUpgradePublishesThePreinstallDeclaration(t *testing.T) {
	for _, test := range []struct {
		name  string
		tasks []task.Interface
	}{
		{name: "base", tasks: upgraderBase{}.PrepareForUpgrade()},
		{name: "mainline", tasks: getUpgraderByVersion(upgrader_1_12_7{}.Version()).PrepareForUpgrade()},
		{name: "daily", tasks: getUpgraderByVersion(semver.MustParse("1.12.7-20260731")).PrepareForUpgrade()},
		{name: "successor", tasks: getUpgraderByVersion(semver.MustParse("1.12.8")).PrepareForUpgrade()},
	} {
		t.Run(test.name, func(t *testing.T) {
			for _, candidate := range test.tasks {
				if candidate.GetName() == "PublishMarketPreinstallDeclaration" {
					return
				}
			}
			t.Fatal("no upgrade task publishes the preinstall declaration")
		})
	}
}

// An upgrade brings no medium, and that is the only thing about it the action
// may leave to a default: the version it declares and the root it writes under
// are stated here, because an upgrade that published under the wrong trunk, or
// somewhere Market does not mount, fails silently -- the task succeeds and the
// device is left with the previous release's list.
func TestTheUpgradePublishBringsNoMediumAndSaysWhichVersionItDeclares(t *testing.T) {
	var published *preinstall.PublishDeclarationAction
	for _, candidate := range (upgraderBase{}).PrepareForUpgrade() {
		local, ok := candidate.(*task.LocalTask)
		if !ok || local.GetName() != "PublishMarketPreinstallDeclaration" {
			continue
		}
		published, ok = local.Action.(*preinstall.PublishDeclarationAction)
		if !ok {
			t.Fatalf("the publish task runs %T, not a declaration publish", local.Action)
		}
	}
	if published == nil {
		t.Fatal("no upgrade task publishes the preinstall declaration")
	}
	if published.InstallerDir != "" {
		t.Errorf("an upgrade brought a medium at %q", published.InstallerDir)
	}
	if published.OSVersion != olversion.VERSION {
		t.Errorf("declares version %q, want this binary's %q",
			published.OSVersion, olversion.VERSION)
	}
	if published.RootDir != storage.OlaresRootDir {
		t.Errorf("writes under %q, want %q", published.RootDir, storage.OlaresRootDir)
	}
}

func TestOneTwelveSevenIsBreakingBoundaryForSuccessor(t *testing.T) {
	original := olversion.VERSION
	olversion.VERSION = "1.12.8"
	t.Cleanup(func() { olversion.VERSION = original })

	got := getLastBreakingVersion(mainUpgraders, semver.MustParse("1.12.8"))
	if got == nil || !got.Equal(version_1_12_7) {
		t.Fatalf("last breaking version before 1.12.8 = %v, want %s", got, version_1_12_7)
	}
}
