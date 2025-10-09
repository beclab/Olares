package upgrade

import (
	"fmt"

	"github.com/Masterminds/semver/v3"
	"github.com/beclab/Olares/cli/pkg/common"
	"github.com/beclab/Olares/cli/pkg/core/connector"
	"github.com/beclab/Olares/cli/pkg/core/task"
	"github.com/beclab/Olares/cli/pkg/core/util"
	"github.com/beclab/Olares/cli/version"
)

var version_12_1_1 = semver.MustParse("1.12.1")

type upgrader_1_12_1 struct {
	breakingUpgraderBase
}

func (u upgrader_1_12_1) Version() *semver.Version {
	cliVersion, err := semver.NewVersion(version.VERSION)
	// tolerate local dev version
	if err != nil {
		return version_12_1_1
	}
	if samePatchLevelVersion(version_12_1_1, cliVersion) && getReleaseLineOfVersion(cliVersion) == mainLine {
		return cliVersion
	}
	return version_12_1_1
}

func (u upgrader_1_12_1) AddedBreakingChange() bool {
	if u.Version().Equal(version_12_1_1) {
		// if this version introduced breaking change
		return true
	}
	return false
}

func (u upgrader_1_12_1) PrepareForUpgrade() []task.Interface {
	var preTasks []task.Interface
	preTasks = append(preTasks,
		&task.LocalTask{
			Name:   "RemoveLegacySystemFrontendManifest",
			Action: new(removeLegacySystemFrontendManifest),
			Retry:  5,
			Delay:  10,
		})
	preTasks = append(preTasks, upgrader_1_12_1_20250826{}.preToPrepareForUpgrade()...)
	return append(preTasks, u.upgraderBase.PrepareForUpgrade()...)
}

type removeLegacySystemFrontendManifest struct {
	common.KubeAction
}

func (c *removeLegacySystemFrontendManifest) Execute(runtime connector.Runtime) error {
	kubeclt, _ := util.GetCommand(common.CommandKubectl)
	var cmd = fmt.Sprintf("%s -n os-framework exec -it -c app-service app-service-0 -- rm -f /userapps/apps/system-apps/templates/system-frontend.yaml", kubeclt)
	if _, err := runtime.GetRunner().SudoCmd(cmd, false, true); err != nil {
		return fmt.Errorf("failed to remove legacy system frontend manifest from app-service")
	}

	return nil
}

func init() {
	registerMainUpgrader(upgrader_1_12_1{})
}
