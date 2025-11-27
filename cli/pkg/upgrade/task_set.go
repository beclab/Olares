package upgrade

import (
	"context"
	"fmt"
	"path"
	"strings"
	"time"

	"github.com/Masterminds/semver/v3"
	"github.com/beclab/Olares/cli/pkg/bootstrap/precheck"
	"github.com/beclab/Olares/cli/pkg/clientset"
	"github.com/beclab/Olares/cli/pkg/common"
	"github.com/beclab/Olares/cli/pkg/container"
	cc "github.com/beclab/Olares/cli/pkg/core/common"
	"github.com/beclab/Olares/cli/pkg/core/connector"
	"github.com/beclab/Olares/cli/pkg/core/logger"
	"github.com/beclab/Olares/cli/pkg/core/task"
	"github.com/beclab/Olares/cli/pkg/core/util"
	"github.com/beclab/Olares/cli/pkg/gpu"
	"github.com/beclab/Olares/cli/pkg/k3s"
	k3stemplates "github.com/beclab/Olares/cli/pkg/k3s/templates"
	"github.com/beclab/Olares/cli/pkg/kubernetes"
	"github.com/beclab/Olares/cli/pkg/kubesphere"
	"github.com/beclab/Olares/cli/pkg/kubesphere/plugins"
	"github.com/beclab/Olares/cli/pkg/manifest"
	"github.com/beclab/Olares/cli/pkg/phase"
	"github.com/beclab/Olares/cli/pkg/terminus"
	"github.com/beclab/Olares/cli/pkg/utils"
	"github.com/pkg/errors"
	"k8s.io/utils/ptr"
)

const cacheRebootNeeded = "reboot.needed"

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
			Name:   "CopyEmbeddedKSManifests",
			Action: new(plugins.CopyEmbedFiles),
		},
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

type upgradeGPUDriverIfNeeded struct {
	common.KubeAction
}

func (a *upgradeGPUDriverIfNeeded) Execute(runtime connector.Runtime) error {
	sys := runtime.GetSystemInfo()
	if sys.IsWsl() {
		return nil
	}
	if !(sys.IsUbuntu() || sys.IsDebian()) {
		return nil
	}

	model, _, err := utils.DetectNvidiaModelAndArch(runtime)
	if err != nil {
		return err
	}
	if strings.TrimSpace(model) == "" {
		return nil
	}

	m, err := manifest.ReadAll(a.KubeConf.Arg.Manifest)
	if err != nil {
		return err
	}
	item, err := m.Get("cuda-driver")
	if err != nil {
		return err
	}
	var targetDriverVersionStr string
	if parts := strings.Split(item.Filename, "-"); len(parts) >= 3 {
		targetDriverVersionStr = strings.TrimSuffix(parts[len(parts)-1], ".run")
	}
	if targetDriverVersionStr == "" {
		return fmt.Errorf("failed to parse target CUDA driver version from %s", item.Filename)
	}
	targetVersion, err := semver.NewVersion(targetDriverVersionStr)
	if err != nil {
		return fmt.Errorf("invalid target driver version '%s': %v", targetDriverVersionStr, err)
	}

	var needUpgrade bool

	status, derr := utils.GetNvidiaStatus(runtime)
	// for now, consider it as not installed if error occurs
	// and continue to upgrade
	if derr != nil {
		logger.Warnf("failed to detect NVIDIA driver status, assuming upgrade is needed: %v", derr)
		needUpgrade = true
	}

	if status != nil && status.Installed {
		currentStr := status.DriverVersion
		if status.Mismatch && status.LibraryVersion != "" {
			currentStr = status.LibraryVersion
		}
		if v, perr := semver.NewVersion(currentStr); perr == nil {
			needUpgrade = targetVersion.GreaterThan(v)
		} else {
			// cannot parse current version, assume upgrade needed
			needUpgrade = true
		}
	} else {
		needUpgrade = true
	}

	changed := false
	if needUpgrade {
		// if apt-installed, uninstall apt nvidia packages but keep toolkit
		if status != nil && status.InstallMethod != utils.GPUDriverInstallMethodRunfile {
			if err := new(gpu.UninstallNvidiaDrivers).Execute(runtime); err != nil {
				return err
			}
		}
		_, _ = runtime.GetRunner().SudoCmd("apt-get update", false, true)
		if _, err := runtime.GetRunner().SudoCmd("DEBIAN_FRONTEND=noninteractive apt-get install -y --no-install-recommends dkms build-essential linux-headers-$(uname -r)", false, true); err != nil {
			return errors.Wrap(errors.WithStack(err), "failed to install kernel build dependencies for NVIDIA runfile")
		}
		// install runfile
		runfile := item.FilePath(runtime.GetBaseDir())
		if _, err := runtime.GetRunner().SudoCmd(fmt.Sprintf("chmod +x %s", runfile), false, true); err != nil {
			return errors.Wrap(errors.WithStack(err), "failed to chmod +x runfile")
		}
		cmd := fmt.Sprintf("sh %s -z --no-x-check --allow-installation-with-running-driver --no-check-for-alternate-installs --dkms --rebuild-initramfs -s", runfile)
		if _, err := runtime.GetRunner().SudoCmd(cmd, false, true); err != nil {
			return errors.Wrap(errors.WithStack(err), "failed to install NVIDIA driver via runfile")
		}
		client, err := clientset.NewKubeClient()
		if err != nil {
			return errors.Wrap(errors.WithStack(err), "kubeclient create error")
		}
		err = gpu.UpdateNodeGpuLabel(context.Background(), client.Kubernetes(), &targetDriverVersionStr, ptr.To(common.CurrentVerifiedCudaVersion), ptr.To("true"))
		if err != nil {
			return err
		}
		changed = true
	}

	needReboot := changed || (status != nil && status.Mismatch)
	a.PipelineCache.Set(cacheRebootNeeded, needReboot)
	return nil
}

type rebootIfNeeded struct {
	common.KubeAction
}

func (r *rebootIfNeeded) Execute(runtime connector.Runtime) error {
	val, ok := r.PipelineCache.GetMustBool(cacheRebootNeeded)
	if ok && val {
		_, _ = runtime.GetRunner().SudoCmd("reboot now", false, false)
	}
	return nil
}
