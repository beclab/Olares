package upgrade

import (
	"github.com/Masterminds/semver/v3"
	"github.com/beclab/Olares/cli/pkg/certs"
	"github.com/beclab/Olares/cli/pkg/common"
	"github.com/beclab/Olares/cli/pkg/core/task"
	"github.com/beclab/Olares/cli/pkg/etcd"
	"github.com/beclab/Olares/cli/pkg/kubesphere/plugins"
	"github.com/beclab/Olares/cli/version"
)

var version_1_12_8 = semver.MustParse("1.12.8")

type upgrader_1_12_8 struct {
	upgraderBase
}

func (u upgrader_1_12_8) Version() *semver.Version {
	cliVersion, err := semver.NewVersion(version.VERSION)
	if err != nil {
		return version_1_12_8
	}
	if samePatchLevelVersion(version_1_12_8, cliVersion) && getReleaseLineOfVersion(cliVersion) == mainLine {
		return cliVersion
	}
	return version_1_12_8
}

func (u upgrader_1_12_8) PrepareForUpgrade() []task.Interface {
	tasks := juiceFSXattrTasks()
	tasks = append(tasks, refreshBackupETCDScript()...)
	tasks = append(tasks, migrateContainerdConfigV3()...)
	tasks = append(tasks, &task.LocalTask{
		Name:   "RestartNvidiaApplicationPods",
		Action: new(restartNvidiaApplicationPods),
	})

	tasks = append(tasks, &task.LocalTask{
		Name:   "CopyEmbeddedKSManifests",
		Action: new(plugins.CopyEmbedFiles),
	})
	tasks = append(tasks, upgradeKubernetesPrometheusRule()...)
	tasks = append(tasks, upgradeUserReverseProxy()...)
	tasks = append(tasks, upgradeAmdDeviceMetricsExporter()...)
	tasks = append(tasks, upgradePrometheusOperator()...)
	tasks = append(tasks, upgradeAmdDevicePlugin()...)

	tasks = append(tasks, u.upgraderBase.PrepareForUpgrade()...)
	return tasks
}

func refreshBackupETCDScript() []task.Interface {
	return []task.Interface{
		&task.LocalTask{
			Name:   "RefreshBackupETCDScript",
			Action: new(etcd.BackupETCD),
		},
	}
}

// PreUpgradeNode rewrites the container runtime's configuration and restarts
// it. Every machine has its own /etc/containerd and its own GPU, so every
// machine migrates itself.
func (u upgrader_1_12_8) PreUpgradeNode() []task.Interface {
	tasks := append(migrateContainerdConfigV3(), &task.LocalTask{
		Name: "CleanupK3sCertsRenewService", Prepare: new(common.OnlyK3s), Action: new(certs.UninstallAutoRenewCerts),
	})
	return append(tasks, u.upgraderBase.PreUpgradeNode()...)
}

// PostUpgradeNode regenerates the kubelet configuration, for the reasons
// upgrader_1_12_7_20260824 gives at length: the reserve is computed from the
// RAM of the machine the task runs on and rewrites that machine's service
// file, so running it only on the control node would leave every other node on
// the old reserve.
func (u upgrader_1_12_8) PostUpgradeNode() []task.Interface {
	tasks := append(regenerateKubeFilesOnNode(), amdMetricsNodeChecks()...)
	return append(tasks, u.upgraderBase.PostUpgradeNode()...)
}

// PostUpgrade carries the kubeadm half, which only the control node may run.
// See regenerateKubeFilesOnControlNode.
func (u upgrader_1_12_8) PostUpgrade() []task.Interface {
	return append(regenerateKubeFilesOnControlNode(), u.upgraderBase.PostUpgrade()...)
}

func init() {
	registerMainUpgrader(upgrader_1_12_8{})
}
