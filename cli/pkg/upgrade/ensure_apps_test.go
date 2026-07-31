package upgrade

import (
	"testing"

	"github.com/Masterminds/semver/v3"
	"github.com/beclab/Olares/cli/pkg/core/task"
)

func TestEnsureAppsTaskIsVersionSpecific(t *testing.T) {
	for _, test := range []struct {
		name    string
		tasks   []task.Interface
		present bool
	}{
		{name: "base", tasks: upgraderBase{}.PrepareForUpgrade()},
		{name: "stable", tasks: getUpgraderByVersion(semver.MustParse("1.12.7")).PrepareForUpgrade(), present: true},
		{name: "daily", tasks: getUpgraderByVersion(semver.MustParse("1.12.7-20260731")).PrepareForUpgrade(), present: true},
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
