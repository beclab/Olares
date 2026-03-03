package amdgpu

import (
	"context"
	"fmt"
	"os/exec"
	"path"
	"path/filepath"

	"github.com/beclab/Olares/cli/pkg/clientset"
	"github.com/beclab/Olares/cli/pkg/common"
	cc "github.com/beclab/Olares/cli/pkg/core/common"
	"github.com/beclab/Olares/cli/pkg/core/connector"
	"github.com/beclab/Olares/cli/pkg/core/logger"
	"github.com/beclab/Olares/cli/pkg/core/task"
	"github.com/beclab/Olares/cli/pkg/core/util"
	"github.com/beclab/Olares/cli/pkg/gpu"

	"github.com/Masterminds/semver/v3"
	"github.com/pkg/errors"
)

// InstallAmdRocmModule installs AMD ROCm stack on supported Ubuntu if AMD GPU is present.
type InstallAmdRocmModule struct {
	common.KubeModule
}

func (m *InstallAmdRocmModule) Init() {
	m.Name = "InstallAMDGPU"

	installAmd := &task.RemoteTask{
		Name:   "InstallAmdRocm",
		Hosts:  m.Runtime.GetHostsByRole(common.Master),
		Action: &InstallAmdRocm{
			// no manifest needed
		},
		Parallel: false,
		Retry:    1,
	}

	m.Tasks = []task.Interface{
		installAmd,
	}
}

// InstallAmdRocm installs ROCm using amdgpu-install on Ubuntu 22.04/24.04 for AMD GPUs.
type InstallAmdRocm struct {
	common.KubeAction
}

func (t *InstallAmdRocm) Execute(runtime connector.Runtime) error {
	si := runtime.GetSystemInfo()
	if !si.IsLinux() || !si.IsUbuntu() || !(si.IsUbuntuVersionEqual(connector.Ubuntu2204) || si.IsUbuntuVersionEqual(connector.Ubuntu2404)) {
		return nil
	}

	amdGPUExists, err := connector.HasAmdAPUOrGPU(runtime)
	if err != nil {
		return err
	}
	// skip rocm install
	if !amdGPUExists {
		return nil
	}
	rocmV, _ := connector.RocmVersion()
	min := semver.MustParse("7.1.1")
	if rocmV != nil && rocmV.LessThan(min) {
		return fmt.Errorf("detected ROCm version %s, which is lower than required %s; please uninstall existing ROCm/AMDGPU components before installation with command: olares-cli amdgpu uninstall", rocmV.Original(), min.Original())
	}
	if rocmV != nil && rocmV.GreaterThan(min) {
		logger.Warnf("Warning: detected ROCm version %s great than maximum tested version %s")
		return nil
	}
	if rocmV != nil && rocmV.Equal(min) {
		logger.Infof("detected ROCm version %s, skip rocm install...", min.Original())
		return nil
	}

	// ensure python3-setuptools and python3-wheel
	_, _ = runtime.GetRunner().SudoCmd("apt-get update", false, true)
	checkPkgs := "dpkg -s python3-setuptools python3-wheel >/dev/null 2>&1 || DEBIAN_FRONTEND=noninteractive apt-get install -y python3-setuptools python3-wheel"
	if _, err := runtime.GetRunner().SudoCmd(checkPkgs, false, true); err != nil {
		return errors.Wrap(errors.WithStack(err), "failed to install python3-setuptools and python3-wheel")
	}
	// ensure amdgpu-install exists
	if _, err := exec.LookPath("amdgpu-install"); err != nil {
		var debURL string
		if si.IsUbuntuVersionEqual(connector.Ubuntu2404) {
			debURL = "https://repo.radeon.com/amdgpu-install/7.1.1/ubuntu/noble/amdgpu-install_7.1.1.70101-1_all.deb"
		} else {
			debURL = "https://repo.radeon.com/amdgpu-install/7.1.1/ubuntu/jammy/amdgpu-install_7.1.1.70101-1_all.deb"
		}
		tmpDeb := path.Join(runtime.GetBaseDir(), cc.PackageCacheDir, "gpu", "amdgpu-install_7.1.1.70101-1_all.deb")
		if _, err := runtime.GetRunner().SudoCmd(fmt.Sprintf("install -d -m 0755 %s", filepath.Dir(tmpDeb)), false, true); err != nil {
			return err
		}
		cmd := fmt.Sprintf("sh -c 'wget -O %s %s'", tmpDeb, debURL)
		if _, err := runtime.GetRunner().SudoCmd(cmd, false, true); err != nil {
			return errors.Wrap(errors.WithStack(err), "failed to download amdgpu-install deb")
		}
		if _, err := runtime.GetRunner().SudoCmd(fmt.Sprintf("DEBIAN_FRONTEND=noninteractive apt-get install -y %s", tmpDeb), false, true); err != nil {
			return errors.Wrap(errors.WithStack(err), "failed to install amdgpu-install deb")
		}
	}
	// run installer for rocm usecase
	if _, err := runtime.GetRunner().SudoCmd("amdgpu-install -y --usecase=rocm", false, true); err != nil {
		return errors.Wrap(errors.WithStack(err), "failed to install AMD ROCm via amdgpu-install")
	}
	fmt.Println()
	logger.Warn("Warning: To enable ROCm, please reboot your machine after installation.")
	return nil
}

type AmdgpuInstallAction struct {
	common.KubeAction
}

func (t *AmdgpuInstallAction) Execute(runtime connector.Runtime) error {
	if _, err := runtime.GetRunner().SudoCmd("amdgpu-install -y --usecase=rocm", false, true); err != nil {
		return errors.Wrap(errors.WithStack(err), "failed to install AMD ROCm via amdgpu-install")
	}
	return nil
}

type AmdgpuUninstallAction struct {
	common.KubeAction
}

func (t *AmdgpuUninstallAction) Execute(runtime connector.Runtime) error {
	if _, err := runtime.GetRunner().SudoCmd("amdgpu-install --uninstall -y", false, true); err != nil {
		return errors.Wrap(errors.WithStack(err), "failed to uninstall AMD ROCm via amdgpu-install")
	}
	fmt.Println()
	logger.Warn("Warning: Please reboot your machine after uninstall to fully remove ROCm components.")
	return nil
}

// UpdateAmdContainerToolkitSource configures the AMD container toolkit APT repository.
type UpdateAmdContainerToolkitSource struct {
	common.KubeAction
}

func (t *UpdateAmdContainerToolkitSource) Execute(runtime connector.Runtime) error {
	// Install prerequisites
	if _, err := runtime.GetRunner().SudoCmd("apt update && apt install -y wget gnupg2", false, true); err != nil {
		return errors.Wrap(errors.WithStack(err), "failed to install prerequisites for AMD container toolkit")
	}

	if _, err := runtime.GetRunner().SudoCmd("install -d -m 0755 /etc/apt/keyrings", false, true); err != nil {
		return errors.Wrap(errors.WithStack(err), "failed to create /etc/apt/keyrings directory")
	}

	cmd := "wget https://repo.radeon.com/rocm/rocm.gpg.key -O - | gpg --dearmor | tee /etc/apt/keyrings/rocm.gpg > /dev/null"
	if _, err := runtime.GetRunner().SudoCmd(cmd, false, true); err != nil {
		return errors.Wrap(errors.WithStack(err), "failed to download and install AMD ROCm GPG key")
	}

	si := runtime.GetSystemInfo()
	var ubuntuCodename string
	if si.IsUbuntuVersionEqual(connector.Ubuntu2404) {
		ubuntuCodename = "noble"
	} else if si.IsUbuntuVersionEqual(connector.Ubuntu2204) {
		ubuntuCodename = "jammy"
	} else {
		return fmt.Errorf("unsupported Ubuntu version for AMD container toolkit")
	}

	aptSourceLine := fmt.Sprintf("deb [signed-by=/etc/apt/keyrings/rocm.gpg] https://repo.radeon.com/amd-container-toolkit/apt/ %s main", ubuntuCodename)
	cmd = fmt.Sprintf("echo '%s' > /etc/apt/sources.list.d/amd-container-toolkit.list", aptSourceLine)
	if _, err := runtime.GetRunner().SudoCmd(cmd, false, true); err != nil {
		return errors.Wrap(errors.WithStack(err), "failed to add AMD container toolkit APT source")
	}

	logger.Infof("AMD container toolkit repository configured successfully")
	return nil
}

// InstallAmdContainerToolkit installs the AMD container toolkit package.
type InstallAmdContainerToolkit struct {
	common.KubeAction
}

func (t *InstallAmdContainerToolkit) Execute(runtime connector.Runtime) error {
	logger.Infof("Installing AMD container toolkit...")
	if _, err := runtime.GetRunner().SudoCmd("apt update && apt install -y amd-container-toolkit", false, true); err != nil {
		return errors.Wrap(errors.WithStack(err), "failed to install AMD container toolkit")
	}
	logger.Infof("AMD container toolkit installed successfully")
	return nil
}

// GenerateAndValidateAmdCDI generates and validates the AMD CDI spec.
type GenerateAndValidateAmdCDI struct {
	common.KubeAction
}

func (t *GenerateAndValidateAmdCDI) Execute(runtime connector.Runtime) error {
	// Ensure /etc/cdi directory exists
	if _, err := runtime.GetRunner().SudoCmd("install -d -m 0755 /etc/cdi", false, true); err != nil {
		return errors.Wrap(errors.WithStack(err), "failed to create /etc/cdi directory")
	}

	// Generate CDI spec
	logger.Infof("Generating AMD CDI spec...")
	if _, err := runtime.GetRunner().SudoCmd("amd-ctk cdi generate --output=/etc/cdi/amd.json", false, true); err != nil {
		return errors.Wrap(errors.WithStack(err), "failed to generate AMD CDI spec")
	}

	// Validate CDI spec
	logger.Infof("Validating AMD CDI spec...")
	if _, err := runtime.GetRunner().SudoCmd("amd-ctk cdi validate --path=/etc/cdi/amd.json", false, true); err != nil {
		return errors.Wrap(errors.WithStack(err), "failed to validate AMD CDI spec")
	}

	logger.Infof("AMD CDI spec generated and validated successfully")
	return nil
}

// UpdateNodeAmdGPUInfo updates Kubernetes node labels with AMD GPU information.
type UpdateNodeAmdGPUInfo struct {
	common.KubeAction
}

func (u *UpdateNodeAmdGPUInfo) Execute(runtime connector.Runtime) error {
	client, err := clientset.NewKubeClient()
	if err != nil {
		return errors.Wrap(errors.WithStack(err), "kubeclient create error")
	}

	// Check if AMD GPU/APU exists
	amdGPUExists, err := connector.HasAmdAPUOrGPU(runtime)
	if err != nil {
		return err
	}
	if !amdGPUExists {
		logger.Info("AMD GPU/APU is not detected")
		return nil
	}

	// Get ROCm version
	rocmV, err := connector.RocmVersion()
	if err != nil || rocmV == nil {
		logger.Info("ROCm is not installed")
		return nil
	}

	rocmVersion := rocmV.Original()

	// Determine GPU type (APU vs discrete GPU)
	gpuType := gpu.AmdGpuCardType
	if runtime.GetSystemInfo().IsAmdApu() {
		gpuType = gpu.AmdApuCardType
	}

	// Use ROCm version as both driver and "cuda" version for AMD
	return gpu.UpdateNodeGpuLabel(context.Background(), client.Kubernetes(), &rocmVersion, nil, nil, &gpuType)
}

// InstallAmdPlugin installs the AMD GPU device plugin DaemonSet.
type InstallAmdPlugin struct {
	common.KubeAction
}

func (t *InstallAmdPlugin) Execute(runtime connector.Runtime) error {
	amdPluginPath := path.Join(runtime.GetInstallerDir(), "wizard/config/gpu/nvidia/amdgpu-device-plugin.yaml")
	_, err := runtime.GetRunner().SudoCmd(fmt.Sprintf("kubectl apply -f %s", amdPluginPath), false, true)
	if err != nil {
		return errors.Wrap(errors.WithStack(err), "failed to apply AMD GPU device plugin")
	}

	logger.Infof("AMD GPU device plugin installed successfully")
	return nil
}

// CheckAmdGpuStatus checks if the AMD GPU device plugin pod is running.
type CheckAmdGpuStatus struct {
	common.KubeAction
}

func (t *CheckAmdGpuStatus) Execute(runtime connector.Runtime) error {
	kubectlpath, err := util.GetCommand(common.CommandKubectl)
	if err != nil {
		return fmt.Errorf("kubectl not found")
	}

	// Check AMD device plugin pod status using the label from amdgpu-device-plugin.yaml
	selector := "name=amdgpu-dp-ds"
	cmd := fmt.Sprintf("%s get pod -n kube-system -l '%s' -o jsonpath='{.items[*].status.phase}'", kubectlpath, selector)

	rphase, _ := runtime.GetRunner().SudoCmd(cmd, false, false)
	if rphase == "Running" {
		logger.Infof("AMD GPU device plugin is running")
		return nil
	}
	return fmt.Errorf("AMD GPU device plugin state is not Running (current: %s)", rphase)
}
