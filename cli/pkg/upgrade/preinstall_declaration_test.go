package upgrade

import (
	"testing"

	"github.com/Masterminds/semver/v3"
	"github.com/beclab/Olares/cli/pkg/core/task"
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

func TestOneTwelveSevenIsBreakingBoundaryForSuccessor(t *testing.T) {
	original := olversion.VERSION
	olversion.VERSION = "1.12.8"
	t.Cleanup(func() { olversion.VERSION = original })

	got := getLastBreakingVersion(mainUpgraders, semver.MustParse("1.12.8"))
	if got == nil || !got.Equal(version_1_12_7) {
		t.Fatalf("last breaking version before 1.12.8 = %v, want %s", got, version_1_12_7)
	}
}
