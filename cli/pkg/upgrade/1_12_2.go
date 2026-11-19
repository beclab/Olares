package upgrade

import (
	"os/exec"
	"time"

	"github.com/Masterminds/semver/v3"
	"github.com/beclab/Olares/cli/pkg/core/task"
	"github.com/beclab/Olares/cli/pkg/gpu"
	"github.com/beclab/Olares/cli/version"
)

var version_1_12_2 = semver.MustParse("1.12.2")

type upgrader_1_12_2 struct {
	breakingUpgraderBase
}

func (u upgrader_1_12_2) Version() *semver.Version {
	cliVersion, err := semver.NewVersion(version.VERSION)
	// tolerate local dev version
	if err != nil {
		return version_1_12_2
	}
	if samePatchLevelVersion(version_1_12_2, cliVersion) && getReleaseLineOfVersion(cliVersion) == mainLine {
		return cliVersion
	}
	return version_1_12_2
}

func (u upgrader_1_12_2) AddedBreakingChange() bool {
	if u.Version().Equal(version_1_12_2) {
		// if this version introduced breaking change
		return true
	}
	return false
}

func nvidiactkNeedsMigration() (bool, error) {
	_, err := exec.LookPath("nvidia-ctk")
	if err != nil {
		return false, nil
	}
	// Both the old 1.17.x packages and full 1.19.x installations must migrate
	// to the base-only package set, even if their CLI version is already current.
	return true, nil
}

func (u upgrader_1_12_2) PrepareForUpgrade() []task.Interface {
	var preTasks []task.Interface
	needsMigration, err := nvidiactkNeedsMigration()
	if err != nil || needsMigration {
		preTasks = append(preTasks,
			&task.LocalTask{
				Name:    "InstallNvidiaContainerToolkit",
				Action:  new(gpu.InstallNvidiaContainerToolkit),
				Prepare: new(gpu.CudaInstalled),
				Retry:   5,
				Delay:   10 * time.Second,
			},
			&task.LocalTask{
				Name:    "ConfigureContainerdRuntime",
				Action:  new(gpu.ConfigureContainerdRuntime),
				Prepare: new(gpu.CudaInstalled),
				Retry:   5,
				Delay:   10 * time.Second,
			},
		)
	}
	preTasks = append(preTasks, u.upgraderBase.PrepareForUpgrade()...)
	return preTasks
}

func (u upgrader_1_12_2) UpgradeSystemComponents() []task.Interface {
	var preTasks []task.Interface
	preTasks = append(preTasks,
		&task.LocalTask{
			Name:   "UpgradeL4",
			Action: new(upgradeL4BFLProxy),
			Retry:  5,
			Delay:  10 * time.Second,
		})
	return append(preTasks, u.upgraderBase.UpgradeSystemComponents()...)
}

func init() {
	registerMainUpgrader(upgrader_1_12_2{})
}
