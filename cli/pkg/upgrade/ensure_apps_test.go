package upgrade

import (
	"testing"

	"github.com/Masterminds/semver/v3"
	"github.com/beclab/Olares/cli/pkg/core/task"
	olversion "github.com/beclab/Olares/cli/version"
)

func TestEnsureAppsTaskIsVersionSpecific(t *testing.T) {
	for _, test := range []struct {
		name    string
		tasks   []task.Interface
		present bool
	}{
		{name: "base", tasks: upgraderBase{}.PrepareForUpgrade()},
		{name: "mainline", tasks: getUpgraderByVersion(upgrader_1_12_7{}.Version()).PrepareForUpgrade(), present: true},
		{name: "daily", tasks: getUpgraderByVersion(semver.MustParse("1.12.7-20260731")).PrepareForUpgrade(), present: true},
		{name: "successor", tasks: getUpgraderByVersion(semver.MustParse("1.12.8")).PrepareForUpgrade()},
	} {
		t.Run(test.name, func(t *testing.T) {
			found := false
			for _, candidate := range test.tasks {
				if candidate.GetName() == "PublishMarketEnsureApps" {
					found = true
				}
			}
			if found != test.present {
				t.Fatalf("PublishMarketEnsureApps present=%t, want %t", found, test.present)
			}
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
