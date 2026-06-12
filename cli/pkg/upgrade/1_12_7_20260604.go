package upgrade

import (
	"fmt"
	"path/filepath"

	"github.com/Masterminds/semver/v3"
	"github.com/beclab/Olares/cli/pkg/common"
	"github.com/beclab/Olares/cli/pkg/core/connector"
	"github.com/beclab/Olares/cli/pkg/core/logger"
	"github.com/beclab/Olares/cli/pkg/core/task"
	"github.com/beclab/Olares/cli/pkg/manifest"
	"github.com/pkg/errors"
)

type upgrader_1_12_7_20260604 struct {
	breakingUpgraderBase
}

func (u upgrader_1_12_7_20260604) Version() *semver.Version {
	return semver.MustParse("1.12.7-20260604")
}

func (u upgrader_1_12_7_20260604) PrepareForUpgrade() []task.Interface {
	tasks := make([]task.Interface, 0)
	tasks = append(tasks, &task.LocalTask{
		Name:   "UpgradeCniPluginsBinary",
		Action: new(upgradeCniPluginsBinary),
	},
	)
	tasks = append(tasks, upgradeNetworkManagerConfig()...)

	tasks = append(tasks, u.upgraderBase.PrepareForUpgrade()...)
	return tasks

}

func init() {
	registerDailyUpgrader(upgrader_1_12_7_20260604{})
}

// only for daily build to fix cni-plugins binary issue, will be removed in future
type upgradeCniPluginsBinary struct {
	common.KubeAction
}

func (u *upgradeCniPluginsBinary) Execute(runtime connector.Runtime) error {
	m, err := manifest.ReadAll(u.KubeConf.Arg.Manifest)

	binary, err := m.Get("cni-plugins")
	if err != nil {
		return fmt.Errorf("get cni-plugins binary info failed: %w", err)
	}

	path := binary.FilePath(runtime.GetBaseDir())

	fileName := binary.Filename
	dst := filepath.Join(common.TmpDir, fileName)
	logger.Debugf("SyncKubeBinary cp cni-plugins from %s to %s", path, dst)
	if err := runtime.GetRunner().Scp(path, dst); err != nil {
		return errors.Wrap(errors.WithStack(err), fmt.Sprintf("sync kube binaries failed"))
	}
	if _, err := runtime.GetRunner().SudoCmd(fmt.Sprintf("tar -zxf %s -C /opt/cni/bin", dst), false, false); err != nil {
		return err
	}

	if _, err := runtime.GetRunner().SudoCmd("systemctl restart cni-dhcp", false, false); err != nil {
		return err
	}
	return nil
}
