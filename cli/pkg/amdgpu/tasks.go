package amdgpu

import (
	"fmt"
	"os/exec"
	"path"
	"path/filepath"

	"github.com/beclab/Olares/cli/pkg/common"
	cc "github.com/beclab/Olares/cli/pkg/core/common"
	"github.com/beclab/Olares/cli/pkg/core/connector"
	"github.com/beclab/Olares/cli/pkg/core/logger"
	"github.com/beclab/Olares/cli/pkg/core/task"
	"github.com/beclab/Olares/cli/pkg/utils"

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
		Name:  "InstallAmdRocm",
		Hosts: m.Runtime.GetHostsByRole(common.Master),
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

	amdGPUExists, err := utils.HasAmdIGPU(runtime)
	if err != nil {
		return err
	}
	// skip rocm install
	if !amdGPUExists {
		return nil
	}
	rocmV, _ := utils.RocmVersion()
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
