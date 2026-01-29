package nfd

import (
	"fmt"
	"path/filepath"
	"strings"

	"github.com/beclab/Olares/cli/pkg/common"
	"github.com/beclab/Olares/cli/pkg/core/connector"
	"github.com/beclab/Olares/cli/pkg/core/logger"
)

// WriteNfdFile writes the embedded NFD manifest to /etc/kubernetes/addons/nfd.yaml.
type WriteNfdFile struct {
	common.KubeAction
}

func (a *WriteNfdFile) Execute(runtime connector.Runtime) error {
	dst := ManifestPath()

	// ensure addons dir exists
	if _, err := runtime.GetRunner().SudoCmd(fmt.Sprintf("mkdir -p %s", filepath.Dir(dst)), false, false); err != nil {
		return err
	}

	payload := string(mustNfdYAML())
	cmd := fmt.Sprintf("cat > %s <<'EOF'\n%s\nEOF", dst, payload)
	_, err := runtime.GetRunner().SudoCmd(cmd, false, false)
	return err
}

// ApplyNfd runs `kubectl apply -f /etc/kubernetes/addons/nfd.yaml`.
type ApplyNfd struct {
	common.KubeAction
}

func (a *ApplyNfd) Execute(runtime connector.Runtime) error {
	kubectl, _ := a.PipelineCache.GetMustString(common.CacheCommandKubectlPath)
	if kubectl == "" {
		kubectl = filepath.Join(common.BinDir, common.CommandKubectl)
	}
	cmd := fmt.Sprintf("%s apply -f %s", kubectl, ManifestPath())
	_, err := runtime.GetRunner().SudoCmd(cmd, false, false)
	return err
}

// WaitNfdReady waits for the main NFD workloads to be ready.
type WaitNfdReady struct {
	common.KubeAction
}

func (a *WaitNfdReady) Execute(runtime connector.Runtime) error {
	kubectl, _ := a.PipelineCache.GetMustString(common.CacheCommandKubectlPath)
	if kubectl == "" {
		kubectl = filepath.Join(common.BinDir, common.CommandKubectl)
	}

	// In the provided manifest, these objects exist:
	// - Deployment nfd-master (ns node-feature-discovery)
	// - Deployment nfd-gc (ns node-feature-discovery)
	// - DaemonSet nfd-worker (ns node-feature-discovery)
	waitCmds := []string{
		fmt.Sprintf("%s -n node-feature-discovery rollout status deploy/nfd-master --timeout=300s", kubectl),
		fmt.Sprintf("%s -n node-feature-discovery rollout status deploy/nfd-gc --timeout=300s", kubectl),
		fmt.Sprintf("%s -n node-feature-discovery rollout status ds/nfd-worker --timeout=300s", kubectl),
	}

	for _, c := range waitCmds {
		stdout, err := runtime.GetRunner().SudoCmd(c, false, false)
		// treat a successful completion string as success even if exit code handling differs across distros
		if err != nil && !strings.Contains(stdout, "successfully rolled out") {
			logger.Errorf("waiting for NFD with %q failed: %v (%s)", c, err, stdout)
			return err
		}
	}
	return nil
}
