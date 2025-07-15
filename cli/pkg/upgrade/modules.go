package upgrade

import (
	"github.com/beclab/Olares/cli/pkg/bootstrap/precheck"

	"github.com/Masterminds/semver/v3"
	"github.com/beclab/Olares/cli/pkg/common"
	"github.com/beclab/Olares/cli/pkg/core/task"
)

type Module struct {
	common.KubeModule
	TargetVersion *semver.Version
}

func (m *Module) Init() {
	m.Name = "UpgradeOlares"

	u := getUpgraderByVersion(m.TargetVersion)
	m.Tasks = append(m.Tasks, u.PrepareForUpgrade()...)
	m.Tasks = append(m.Tasks, u.ClearAppChartValues()...)
	m.Tasks = append(m.Tasks, u.ClearBFLChartValues()...)
	m.Tasks = append(m.Tasks, u.UpdateChartsInAppService()...)
	m.Tasks = append(m.Tasks, u.UpgradeUserComponents()...)
	m.Tasks = append(m.Tasks, u.UpdateReleaseFile()...)
	m.Tasks = append(m.Tasks, u.UpgradeSystemComponents()...)
	m.Tasks = append(m.Tasks, u.PostUpgrade()...)
}

type PrecheckModule struct {
	common.KubeModule
}

func (m *PrecheckModule) Init() {
	m.Name = "UpgradePrecheck"

	checkers := []precheck.Checker{
		new(precheck.MasterNodeReadyCheck),
		new(precheck.RootPartitionAvailableSpaceCheck),
	}
	runPreChecks := &task.LocalTask{
		Name: "UpgradePrecheck",
		Action: &precheck.RunChecks{
			Checkers: checkers,
		},
	}

	m.Tasks = []task.Interface{
		runPreChecks,
	}
}
