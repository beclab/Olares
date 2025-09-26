package gpu

import (
	"fmt"
	"path/filepath"
	"strings"
	"time"

	"github.com/beclab/Olares/cli/pkg/common"
	"github.com/beclab/Olares/cli/pkg/core/connector"
	"github.com/beclab/Olares/cli/pkg/core/prepare"
	"github.com/beclab/Olares/cli/pkg/core/task"
)

// Applies the OFFICIAL AMD ROCm device-plugin DS if an AMD GPU is detected.
// Ref: https://github.com/ROCm/k8s-device-plugin  (apply from web) 
//      kubectl apply -f https://raw.githubusercontent.com/ROCm/k8s-device-plugin/master/k8s-ds-amdgpu-dp.yaml
// The DS name we wait for: amdgpu-device-plugin-daemonset (ns kube-system).
const (
	amdDevicePluginURL   = "https://raw.githubusercontent.com/ROCm/k8s-device-plugin/master/k8s-ds-amdgpu-dp.yaml"
	amdDPDaemonSetName   = "amdgpu-device-plugin-daemonset"
	amdDPNamespace       = "kube-system"
)

type InstallAmdDevicePluginModule struct{ common.KubeModule }

func (m *InstallAmdDevicePluginModule) Init() {
	m.Name = "InstallAmdDevicePlugin"

	apply := &task.RemoteTask{
		Name:  "ApplyAmdDevicePlugin",
		Hosts: m.Runtime.GetHostsByRole(common.Master),
		Prepare: &prepare.PrepareCollection{
			new(common.OnlyFirstMaster),
			new(AmdGpuPresent),
		},
		Action:   new(ApplyAmdDevicePlugin),
		Parallel: false,
		Retry:    1,
		Timeout:  2 * time.Minute,
	}
	wait := &task.RemoteTask{
		Name:  "WaitAmdDevicePluginReady",
		Hosts: m.Runtime.GetHostsByRole(common.Master),
		Prepare: &prepare.PrepareCollection{
			new(common.OnlyFirstMaster),
			new(AmdGpuPresent),
		},
		Action:   new(WaitAmdDevicePluginReady),
		Parallel: false,
		Timeout:  5 * time.Minute,
	}
	m.Tasks = []task.Interface{apply, wait}
}

// ---- detection prepare ----

type AmdGpuPresent struct{ common.KubePrepare }

func (p *AmdGpuPresent) PreCheck(runtime connector.Runtime) (bool, error) {
	checks := []string{
		"test -e /dev/kfd && echo yes || true",
		"lsmod | grep -q '^amdgpu' && echo yes || true",
		`lspci -nn | egrep -i 'vga|3d|display' | egrep -qi 'amd|ati' && echo yes || true`,
	}
	for _, c := range checks {
		out, _ := runtime.GetRunner().SudoCmd(c, false, false)
		if strings.Contains(out, "yes") {
			return true, nil
		}
	}
	return false, nil
}

// ---- actions ----

type ApplyAmdDevicePlugin struct{ common.KubeAction }
func (a *ApplyAmdDevicePlugin) Execute(runtime connector.Runtime) error {
	kubectl, _ := a.PipelineCache.GetMustString(common.CacheCommandKubectlPath)
	if kubectl == "" {
		kubectl = filepath.Join(common.BinDir, "kubectl")
	}
	_, err := runtime.GetRunner().SudoCmd(fmt.Sprintf("%s apply -f %s", kubectl, amdDevicePluginURL), false, false)
	return err
}

type WaitAmdDevicePluginReady struct{ common.KubeAction }
func (a *WaitAmdDevicePluginReady) Execute(runtime connector.Runtime) error {
	kubectl, _ := a.PipelineCache.GetMustString(common.CacheCommandKubectlPath)
	if kubectl == "" {
		kubectl = filepath.Join(common.BinDir, "kubectl")
	}
	_, err := runtime.GetRunner().SudoCmd(
		fmt.Sprintf("%s -n %s rollout status ds/%s --timeout=300s", kubectl, amdDPNamespace, amdDPDaemonSetName),
		false, false,
	)
	return err
}
