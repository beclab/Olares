package upgrade

import (
	"github.com/Masterminds/semver/v3"
	"github.com/beclab/Olares/cli/apis/kubekey/v1alpha2"
	"github.com/beclab/Olares/cli/pkg/core/action"
	"github.com/beclab/Olares/cli/pkg/core/task"
	"github.com/beclab/Olares/cli/pkg/core/util"
	"github.com/beclab/Olares/cli/pkg/etcd/templates"
	"github.com/beclab/Olares/cli/pkg/terminus"
	"path/filepath"
)

type upgrader_1_12_1_20250826 struct {
	breakingUpgraderBase
}

func (u upgrader_1_12_1_20250826) Version() *semver.Version {
	return semver.MustParse("1.12.1-20250826")
}

func (u upgrader_1_12_1_20250826) PrepareForUpgrade() []task.Interface {
	preTasks := []task.Interface{
		&task.LocalTask{
			Name: "UpdateBackupETCDService",
			Action: &action.Template{
				Name:     "GenerateBackupETCDService",
				Template: templates.BackupETCDService,
				Dst:      filepath.Join("/etc/systemd/system/", templates.BackupETCDService.Name()),
				Data: util.Data{
					"ScriptPath": filepath.Join(v1alpha2.DefaultEtcdBackupScriptDir, templates.EtcdBackupScript.Name()),
				},
			},
		},
		&task.LocalTask{
			Name: "ReloadBackupETCDService",
			Action: &terminus.SystemctlCommand{
				DaemonReloadPreExec: true,
			},
		},
	}
	return append(preTasks, u.upgraderBase.PrepareForUpgrade()...)
}
