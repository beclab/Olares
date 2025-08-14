package upgrade

import (
	"github.com/Masterminds/semver/v3"
	"github.com/beclab/Olares/cli/pkg/common"
	"github.com/beclab/Olares/cli/pkg/container"
	"github.com/beclab/Olares/cli/pkg/core/connector"
	"github.com/beclab/Olares/cli/pkg/core/task"
	"github.com/beclab/Olares/cli/pkg/manifest"
)

type upgrader_1_12_0_20250723 struct {
	breakingUpgraderBase
}

func (u upgrader_1_12_0_20250723) Version() *semver.Version {
	return semver.MustParse("1.12.0-20250723")
}

func (u upgrader_1_12_0_20250723) PrepareForUpgrade() []task.Interface {
	preTasks := []task.Interface{
		&task.LocalTask{
			Name:   "UpgradeContainerd",
			Action: new(upgradeContainerd),
		},
		&task.LocalTask{
			Name:   "RestartContainerd",
			Action: new(container.RestartContainerd),
		},
	}
	return append(preTasks, u.upgraderBase.PrepareForUpgrade()...)
}

type upgradeContainerd struct {
	common.KubeAction
}

func (u *upgradeContainerd) Execute(runtime connector.Runtime) error {
	m, err := manifest.ReadAll(u.KubeConf.Arg.Manifest)
	if err != nil {
		return err
	}
	action := &container.SyncContainerd{
		ManifestAction: manifest.ManifestAction{
			Manifest: m,
			BaseDir:  runtime.GetBaseDir(),
		},
	}
	return action.Execute(runtime)
}
