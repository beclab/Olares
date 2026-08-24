package upgrade

import (
	"github.com/Masterminds/semver/v3"
	"github.com/beclab/Olares/cli/pkg/certs"
	"github.com/beclab/Olares/cli/pkg/common"
	"github.com/beclab/Olares/cli/pkg/core/task"
	"github.com/beclab/Olares/cli/pkg/kubesphere/plugins"
	"github.com/beclab/Olares/cli/version"
)

var version_1_12_7 = semver.MustParse("1.12.7")

type upgrader_1_12_7 struct {
	breakingUpgraderBase
}

func (u upgrader_1_12_7) Version() *semver.Version {
	cliVersion, err := semver.NewVersion(version.VERSION)
	if err != nil {
		return version_1_12_7
	}
	if samePatchLevelVersion(version_1_12_7, cliVersion) && getReleaseLineOfVersion(cliVersion) == mainLine {
		return cliVersion
	}
	return version_1_12_7
}

func (u upgrader_1_12_7) AddedBreakingChange() bool {
	if u.Version().Equal(version_1_12_7) {
		return true
	}
	return false
}

func (u upgrader_1_12_7) PrepareForUpgrade() []task.Interface {
	tasks := migrateContainerdConfigV3()
	tasks = append(tasks, &task.LocalTask{
		Name:    "CleanupK3sCertsRenewService",
		Prepare: new(common.OnlyK3s),
		Action:  new(certs.UninstallAutoRenewCerts),
	})

	tasks = append(tasks, &task.LocalTask{
		Name:   "CopyEmbeddedKSManifests",
		Action: new(plugins.CopyEmbedFiles),
	})
	tasks = append(tasks, upgradeKubernetesPrometheusRule()...)
	tasks = append(tasks, upgradeUserReverseProxy()...)
	tasks = append(tasks, upgradeAmdDeviceMetricsExporter()...)

	tasks = append(tasks, u.upgraderBase.PrepareForUpgrade()...)
	return tasks
}

func init() {
	registerMainUpgrader(upgrader_1_12_7{})
}
