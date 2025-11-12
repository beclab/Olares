package upgrade

import (
	"time"

	"github.com/beclab/Olares/cli/pkg/bootstrap/precheck"
	"github.com/beclab/Olares/cli/pkg/common"
	"github.com/beclab/Olares/cli/pkg/container"
	"github.com/beclab/Olares/cli/pkg/core/connector"
	"github.com/beclab/Olares/cli/pkg/core/task"
	"github.com/beclab/Olares/cli/pkg/k3s"
	k3stemplates "github.com/beclab/Olares/cli/pkg/k3s/templates"
	"github.com/beclab/Olares/cli/pkg/kubernetes"
	"github.com/beclab/Olares/cli/pkg/kubesphere"
	"github.com/beclab/Olares/cli/pkg/kubesphere/plugins"
	"github.com/beclab/Olares/cli/pkg/manifest"
	"github.com/beclab/Olares/cli/pkg/phase"
	"github.com/beclab/Olares/cli/pkg/terminus"
)

type upgradeContainerdAction struct {
	common.KubeAction
}

func (u *upgradeContainerdAction) Execute(runtime connector.Runtime) error {
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

func upgradeContainerd() []task.Interface {
	return []task.Interface{
		&task.LocalTask{
			Name:   "UpgradeContainerd",
			Action: new(upgradeContainerdAction),
		},
		&task.LocalTask{
			Name:   "RestartContainerd",
			Action: new(container.RestartContainerd),
		},
	}
}

func upgradeKSCore() []task.Interface {
	return []task.Interface{
		&task.LocalTask{
			Name:   "CopyEmbeddedKSManifests",
			Action: new(plugins.CopyEmbedFiles),
		},
		&task.LocalTask{
			Name:    "UpgradeKSCore",
			Prepare: new(common.GetMasterNum),
			Action:  new(plugins.CreateKsCore),
			Retry:   10,
			Delay:   10 * time.Second,
		},
		&task.LocalTask{
			Name:   "CheckKSCoreRunning",
			Action: new(kubesphere.Check),
			Retry:  20,
			Delay:  10 * time.Second,
		},
	}
}

func regenerateKubeFiles() []task.Interface {
	var tasks []task.Interface
	kubeType := phase.GetKubeType()
	if kubeType == common.K3s {
		tasks = append(tasks,
			&task.LocalTask{
				Name:   "RegenerateK3sService",
				Action: new(k3s.GenerateK3sService),
			},
			&task.LocalTask{
				Name: "RestartK3sService",
				Action: &terminus.SystemctlCommand{
					Command:             "restart",
					UnitNames:           []string{k3stemplates.K3sService.Name()},
					DaemonReloadPreExec: true,
				},
			},
		)
	} else {
		tasks = append(tasks,
			&task.LocalTask{
				Name: "RegenerateKubeadmConfig",
				Action: &kubernetes.GenerateKubeadmConfig{
					IsInitConfiguration: true,
				},
			},
			&task.LocalTask{
				Name:   "RegenerateK8sFilesWithKubeadm",
				Action: new(terminus.RegenerateFilesForK8s),
			},
		)
	}

	tasks = append(tasks,
		&task.LocalTask{
			Name:   "WaitForKubeAPIServerUp",
			Action: new(precheck.GetKubernetesNodesStatus),
			Retry:  10,
			Delay:  10,
		},
	)
	return tasks
}
