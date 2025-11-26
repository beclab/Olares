package upgrade

import (
	"fmt"
	"path"
	"time"

	"github.com/beclab/Olares/cli/pkg/bootstrap/precheck"
	"github.com/beclab/Olares/cli/pkg/common"
	"github.com/beclab/Olares/cli/pkg/container"
	cc "github.com/beclab/Olares/cli/pkg/core/common"
	"github.com/beclab/Olares/cli/pkg/core/connector"
	"github.com/beclab/Olares/cli/pkg/core/logger"
	"github.com/beclab/Olares/cli/pkg/core/task"
	"github.com/beclab/Olares/cli/pkg/core/util"
	"github.com/beclab/Olares/cli/pkg/k3s"
	k3stemplates "github.com/beclab/Olares/cli/pkg/k3s/templates"
	"github.com/beclab/Olares/cli/pkg/kubernetes"
	"github.com/beclab/Olares/cli/pkg/kubesphere"
	"github.com/beclab/Olares/cli/pkg/kubesphere/plugins"
	"github.com/beclab/Olares/cli/pkg/manifest"
	"github.com/beclab/Olares/cli/pkg/phase"
	"github.com/beclab/Olares/cli/pkg/terminus"
	"github.com/pkg/errors"
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

func upgradePrometheusServiceMonitorKubelet() []task.Interface {
	return []task.Interface{
		// prometheus kubelet ServiceMonitor
		&task.LocalTask{
			Name:   "ApplyKubeletServiceMonitor",
			Action: new(applyKubeletServiceMonitorAction),
			Retry:  5,
			Delay:  5 * time.Second,
		},
	}
}

func upgradeKsConfig() []task.Interface {
	return []task.Interface{
		&task.LocalTask{
			Name:   "ApplyKsConfigManifests",
			Action: new(plugins.ApplyKsConfigManifests),
			Retry:  5,
			Delay:  5 * time.Second,
		},
	}
}

// applyKubeletServiceMonitorAction applies embedded prometheus kubelet ServiceMonitor
type applyKubeletServiceMonitorAction struct {
	common.KubeAction
}

func (a *applyKubeletServiceMonitorAction) Execute(runtime connector.Runtime) error {
	kubectlpath, err := util.GetCommand(common.CommandKubectl)
	if err != nil {
		return errors.Wrap(errors.WithStack(err), "kubectl not found")
	}
	manifest := path.Join(runtime.GetInstallerDir(), cc.BuildFilesCacheDir, cc.BuildDir, "prometheus", "kubernetes", "kubernetes-serviceMonitorKubelet.yaml")
	if _, err := runtime.GetRunner().SudoCmd(fmt.Sprintf("%s apply -f %s", kubectlpath, manifest), false, true); err != nil {
		return errors.Wrap(errors.WithStack(err), "apply kubelet ServiceMonitor failed")
	}
	return nil
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

type upgradeL4BFLProxy struct {
	common.KubeAction
	Tag string
}

func (u *upgradeL4BFLProxy) Execute(runtime connector.Runtime) error {
	if _, err := runtime.GetRunner().SudoCmd(fmt.Sprintf(
		"/usr/local/bin/kubectl set image deployment/l4-bfl-proxy proxy=beclab/l4-bfl-proxy:%s -n os-network", u.Tag), false, true); err != nil {
		return errors.Wrapf(errors.WithStack(err), "failed to upgrade L4 network proxy to version %s", u.Tag)
	}

	logger.Infof("L4 upgrade to version %s completed successfully", u.Tag)
	return nil
}
