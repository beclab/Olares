package amd

import (
	"fmt"
	"path/filepath"
	"strings"

	"github.com/beclab/Olares/cli/pkg/common"
	"github.com/beclab/Olares/cli/pkg/core/connector"
	"github.com/beclab/Olares/cli/pkg/core/logger"
)

const (
	nsName        = "kube-system"                           // most AMD plugin examples live here; change if your YAML uses another ns
	dsName        = "amd-gpu-device-plugin"                 // must match the DaemonSet metadata.name in your YAML
	manifestOnHost = "/etc/kubernetes/addons/amd-gpu.yaml"  // destination on first CP node
)

// -------- Detect helpers (used by orchestrator) --------

// HasAmdKernelBits returns true if amdgpu module or /dev/kfd present.
func HasAmdKernelBits(r connector.Runtime) bool {
	checks := []string{
		"test -e /dev/kfd && echo yes || true",
		"lsmod | grep -q '^amdgpu' && echo yes || true",
	}
	for _, c := range checks {
		out, _ := r.GetRunner().SudoCmd(c, false, false)
		if strings.Contains(out, "yes") {
			return true
		}
	}
	return false
}

// HasAmdPci returns true if lspci sees an AMD/ATI VGA/3D controller.
func HasAmdPci(r connector.Runtime) bool {
	out, err := r.GetRunner().SudoCmd(`lspci -nn | egrep -i 'vga|3d|display' | egrep -i 'amd|ati' && echo yes || true`, false, false)
	return err == nil && strings.Contains(out, "yes")
}

// -------- Tasks --------

type WriteManifest struct{ common.KubeAction }

func (a *WriteManifest) Execute(runtime connector.Runtime) error {
	dir := filepath.Dir(manifestOnHost)
	if _, err := runtime.GetRunner().SudoCmd(fmt.Sprintf("mkdir -p %s", dir), false, false); err != nil {
		return err
	}
	payload := string(mustManifestYAML())
	cmd := fmt.Sprintf("cat > %s <<'EOF'\n%s\nEOF", manifestOnHost, payload)
	_, err := runtime.GetRunner().SudoCmd(cmd, false, false)
	return err
}

type ApplyManifest struct{ common.KubeAction }

func (a *ApplyManifest) Execute(runtime connector.Runtime) error {
	kubectl, _ := a.PipelineCache.GetMustString(common.CacheCommandKubectlPath)
	if kubectl == "" {
		kubectl = filepath.Join(common.BinDir, common.CommandKubectl)
	}
	cmd := fmt.Sprintf("%s apply -f %s", kubectl, manifestOnHost)
	_, err := runtime.GetRunner().SudoCmd(cmd, false, false)
	return err
}

type WaitReady struct{ common.KubeAction }

func (a *WaitReady) Execute(runtime connector.Runtime) error {
	kubectl, _ := a.PipelineCache.GetMustString(common.CacheCommandKubectlPath)
	if kubectl == "" {
		kubectl = filepath.Join(common.BinDir, common.CommandKubectl)
	}
	wait := fmt.Sprintf("%s -n %s rollout status ds/%s --timeout=300s", kubectl, nsName, dsName)
	out, err := runtime.GetRunner().SudoCmd(wait, false, false)
	if err != nil && !strings.Contains(out, "successfully rolled out") {
		logger.Errorf("waiting for AMD GPU DS failed: %v (%s)", err, out)
		return err
	}
	return nil
}
