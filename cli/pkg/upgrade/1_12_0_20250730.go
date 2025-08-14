package upgrade

import (
	"fmt"
	"github.com/Masterminds/semver/v3"
	"github.com/beclab/Olares/cli/pkg/bootstrap/precheck"
	"github.com/beclab/Olares/cli/pkg/common"
	"github.com/beclab/Olares/cli/pkg/core/connector"
	"github.com/beclab/Olares/cli/pkg/core/task"
	"github.com/beclab/Olares/cli/pkg/core/util"
	"github.com/beclab/Olares/cli/pkg/gpu"
	k3stemplates "github.com/beclab/Olares/cli/pkg/k3s/templates"
	"github.com/beclab/Olares/cli/pkg/manifest"
	"github.com/beclab/Olares/cli/pkg/terminus"
	"github.com/pkg/errors"
	"os"
	"path/filepath"
	"strings"
)

type upgrader_1_12_0_20250730 struct {
	breakingUpgraderBase
}

func (u upgrader_1_12_0_20250730) Version() *semver.Version {
	return semver.MustParse("1.12.0-20250730")
}

func (u upgrader_1_12_0_20250730) PrepareForUpgrade() []task.Interface {
	var preTasks []task.Interface
	if util.IsExist(filepath.Join("/etc/systemd/system/", k3stemplates.K3sService.Name())) {
		preTasks = append(preTasks,
			&task.LocalTask{
				Name:   "UpgradeK3sBinary",
				Action: new(upgradeK3sBinary),
			},
			&task.LocalTask{
				Name:   "UpdateK3sServiceEnv",
				Action: new(injectK3sCertExpireTime),
			},
			&task.LocalTask{
				Name: "RestartK3sService",
				Action: &terminus.SystemctlCommand{
					UnitNames:           []string{common.K3s},
					Command:             "restart",
					DaemonReloadPreExec: true,
				},
			},
			&task.LocalTask{
				Name:   "WaitForKubeAPIServerUp",
				Action: new(precheck.GetKubernetesNodesStatus),
				Retry:  10,
				Delay:  10,
			})
	}
	return append(preTasks, u.upgraderBase.PrepareForUpgrade()...)
}

func (u upgrader_1_12_0_20250730) UpgradeSystemComponents() []task.Interface {
	preTasks := []task.Interface{
		&task.LocalTask{
			Name:   "UpgradeGPUPlugin",
			Action: new(gpu.InstallPlugin),
		},
	}
	return append(preTasks, u.upgraderBase.UpgradeSystemComponents()...)
}

type upgradeK3sBinary struct {
	common.KubeAction
}

func (u *upgradeK3sBinary) Execute(runtime connector.Runtime) error {
	m, err := manifest.ReadAll(u.KubeConf.Arg.Manifest)
	if err != nil {
		return err
	}
	binary, err := m.Get(common.K3s)
	if err != nil {
		return fmt.Errorf("get k3s binary info failed: %v", err)
	}

	path := binary.FilePath(runtime.GetBaseDir())
	dst := filepath.Join(common.BinDir, common.K3s)
	// replacing the binary does not interrupt the running k3s server
	if err := runtime.GetRunner().SudoScp(path, dst); err != nil {
		return errors.Wrap(errors.WithStack(err), fmt.Sprintf("upgrade k3s binary failed"))
	}
	if _, err := runtime.GetRunner().SudoCmd(fmt.Sprintf("chmod +x %s", dst), false, false); err != nil {
		return err
	}
	return nil
}

type injectK3sCertExpireTime struct {
	common.KubeAction
}

func (u *injectK3sCertExpireTime) Execute(runtime connector.Runtime) error {
	expireTimeEnv := "CATTLE_NEW_SIGNED_CERT_EXPIRATION_DAYS"
	envFile := filepath.Join("/etc/systemd/system/", k3stemplates.K3sServiceEnv.Name())
	content, err := os.ReadFile(envFile)
	if err != nil && !os.IsNotExist(err) {
		return fmt.Errorf("failed to read k3s service env file: %v", err)
	}
	if strings.Contains(string(content), expireTimeEnv) {
		return nil
	}
	newContent := string(content) + fmt.Sprintf("\n%s=36500\n", expireTimeEnv)
	err = os.WriteFile(envFile, []byte(newContent), 0644)
	return err
}
