package upgrade

import (
	"github.com/Masterminds/semver/v3"
	"github.com/beclab/Olares/cli/pkg/certs"
	"github.com/beclab/Olares/cli/pkg/common"
	"github.com/beclab/Olares/cli/pkg/core/task"
)

// upgrader_1_12_7_20260729 removes the kubeadm-only k8s-certs-renew timer and
// service from k3s nodes, which earlier installs enabled there even though no
// kubeadm binary exists to run.
type upgrader_1_12_7_20260729 struct {
	breakingUpgraderBase
}

func (u upgrader_1_12_7_20260729) Version() *semver.Version {
	return semver.MustParse("1.12.7-20260729")
}

func (u upgrader_1_12_7_20260729) PrepareForUpgrade() []task.Interface {
	tasks := []task.Interface{
		&task.LocalTask{
			Name:    "CleanupK3sCertsRenewService",
			Prepare: new(common.OnlyK3s),
			Action:  new(certs.UninstallAutoRenewCerts),
		},
	}
	return append(tasks, u.upgraderBase.PrepareForUpgrade()...)
}

func init() {
	registerDailyUpgrader(upgrader_1_12_7_20260729{})
}
