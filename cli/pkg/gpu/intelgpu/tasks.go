package intelgpu

import (
	"context"
	"fmt"
	"os"
	"path"
	"strings"

	"github.com/beclab/Olares/cli/pkg/clientset"
	"github.com/beclab/Olares/cli/pkg/common"
	"github.com/beclab/Olares/cli/pkg/core/connector"
	"github.com/beclab/Olares/cli/pkg/core/logger"
	"github.com/beclab/Olares/cli/pkg/core/util"
	"github.com/beclab/Olares/cli/pkg/gpu"

	"github.com/pkg/errors"
)

// intelConfigDir is the installer-relative directory that holds the Intel GPU
// manifests (mirrors infrastructure/gpu/.olares/config/gpu/intel).
const intelConfigDir = "wizard/config/gpu/intel"

// UpdateNodeIntelGPUInfo labels the node as supporting the "intel" mode (Intel
// integrated GPU)
type UpdateNodeIntelGPUInfo struct {
	common.KubeAction
}

func (u *UpdateNodeIntelGPUInfo) Execute(runtime connector.Runtime) error {
	client, err := clientset.NewKubeClient()
	if err != nil {
		return errors.Wrap(errors.WithStack(err), "kubeclient create error")
	}

	if !connector.HasIntelGPU(runtime) {
		logger.Info("Intel GPU is not detected")
		return nil
	}

	return gpu.SetNodeGpuModeLabel(context.Background(), client.Kubernetes(), gpu.IntelType, nil, nil, nil)
}

// applyIntelManifest kubectl-applies a single manifest under intelConfigDir.
func applyIntelManifest(runtime connector.Runtime, fileName, desc string) error {
	manifestPath := path.Join(runtime.GetInstallerDir(), intelConfigDir, fileName)
	if _, err := runtime.GetRunner().SudoCmd(fmt.Sprintf("kubectl apply -f %s", manifestPath), false, true); err != nil {
		return errors.Wrap(errors.WithStack(err), fmt.Sprintf("failed to apply Intel %s (%s)", desc, fileName))
	}
	logger.Infof("Intel %s applied successfully", desc)
	return nil
}

// ApplyIntelNFD applies the node-feature-discovery manifests (nfd.yaml). This is
// kept in its own task because nfd.yaml installs CRDs (NodeFeature,
// NodeFeatureRule, NodeFeatureGroup): kubectl apply returns before the CRDs are
// fully established on the API server, so any manifest applied immediately after
// (e.g. the NodeFeatureRule CRs in node-feature-rules.yaml) may fail. Running it
// separately with a retry lets the CRDs settle before dependent applies.
type ApplyIntelNFD struct {
	common.KubeAction
}

func (t *ApplyIntelNFD) Execute(runtime connector.Runtime) error {
	return applyIntelManifest(runtime, "nfd.yaml", "node-feature-discovery")
}

// ApplyIntelNodeFeatureRules applies the NodeFeatureRule CRs
// (node-feature-rules.yaml). It depends on the CRDs created by ApplyIntelNFD, so
// it is retried until the CRDs are established.
type ApplyIntelNodeFeatureRules struct {
	common.KubeAction
}

func (t *ApplyIntelNodeFeatureRules) Execute(runtime connector.Runtime) error {
	return applyIntelManifest(runtime, "node-feature-rules.yaml", "node feature rules")
}

// ApplyIntelGPUPlugin applies the Intel GPU device plugin DaemonSet
// (gpu-plugin.yaml).
type ApplyIntelGPUPlugin struct {
	common.KubeAction
}

func (t *ApplyIntelGPUPlugin) Execute(runtime connector.Runtime) error {
	return applyIntelManifest(runtime, "gpu-plugin.yaml", "GPU device plugin")
}

// CheckIntelGpu waits until the whole Intel GPU stack is Running: the NFD pods
// (nfd-master, nfd-gc, nfd-worker) and the intel-gpu-plugin DaemonSet. Without
// this gate the install can finish while pods are still Pending (e.g. before NFD
// applies the intel.feature.node.kubernetes.io/gpu label that the plugin's
// nodeSelector requires), making the device-plugin setup look successful when it
// is not yet working.
type CheckIntelGpu struct {
	common.KubeAction
}

func (t *CheckIntelGpu) Execute(runtime connector.Runtime) error {
	kubectlpath, err := util.GetCommand(common.CommandKubectl)
	if err != nil {
		return fmt.Errorf("kubectl not found")
	}

	nodeName, err := os.Hostname()
	if err != nil {
		return errors.Wrap(errors.WithStack(err), "get hostname error")
	}
	nodeName = strings.ToLower(nodeName)

	checks := []struct {
		selector string
		withNode bool
	}{
		{selector: "app=nfd-master", withNode: false},
		{selector: "app=nfd-gc", withNode: false},
		{selector: "app=nfd-worker", withNode: true},
		{selector: "app=intel-gpu-plugin", withNode: true},
	}

	for _, c := range checks {
		cmd := fmt.Sprintf("%s get pod -n kube-system -l '%s' -o jsonpath='{.items[*].status.phase}'", kubectlpath, c.selector)
		if c.withNode {
			cmd = fmt.Sprintf("%s get pod -n kube-system -l '%s' --field-selector 'spec.nodeName=%s' -o jsonpath='{.items[*].status.phase}'", kubectlpath, c.selector, nodeName)
		}

		rphase, _ := runtime.GetRunner().SudoCmd(cmd, false, false)
		if !hasRunningPod(rphase) {
			return fmt.Errorf("pod for selector %q is not Running", c.selector)
		}
	}

	return nil
}

func hasRunningPod(phases string) bool {
	for _, phase := range strings.Fields(phases) {
		if phase == "Running" {
			return true
		}
	}
	return false
}
