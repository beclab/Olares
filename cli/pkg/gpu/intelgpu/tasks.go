package intelgpu

import (
	"context"
	"fmt"
	"os"
	"path"
	"strings"
	"time"

	"github.com/beclab/Olares/cli/pkg/clientset"
	"github.com/beclab/Olares/cli/pkg/common"
	"github.com/beclab/Olares/cli/pkg/core/connector"
	"github.com/beclab/Olares/cli/pkg/core/logger"
	"github.com/beclab/Olares/cli/pkg/core/util"
	"github.com/beclab/Olares/cli/pkg/gpu"
	"github.com/beclab/Olares/cli/pkg/utils"

	"github.com/pkg/errors"
	ctrl "sigs.k8s.io/controller-runtime"
)

// intelConfigDir is the installer-relative directory that holds the Intel GPU
// manifests (mirrors infrastructure/gpu/.olares/config/gpu/intel).
const intelConfigDir = "wizard/config/gpu/intel"

const (
	xpumdChartRelDir = "wizard/config/gpu/intel/xpumd"
	xpumdReleaseName = "xpumd"
	xpumdNamespace   = "kube-system"
)

// intelPlan is the per-node decision derived from the detected Intel GPUs and
// the running kernel: which mode labels to apply and whether a discrete GPU
// qualifies for the host driver packages.
type intelPlan struct {
	labelIntel        bool // apply the "intel" (integrated) mode label
	labelIntelGPU     bool // apply the "intel-gpu" (discrete) mode label
	hasQualifyingDGPU bool // a discrete GPU is present, in-tree and kernel-supported
}

// classifyIntelGPUs enumerates the host's Intel GPUs. Labels are applied purely
// by device kind: an integrated GPU always yields the "intel" label and a
// discrete GPU always yields the "intel-gpu" label, regardless of the kernel
// requirement or Out-of-tree status. The kernel / Out-of-tree / table-presence
// checks only emit warnings and decide whether a discrete GPU qualifies for the
// host driver packages (in-tree and kernel-supported).
func classifyIntelGPUs(runtime connector.Runtime) (intelPlan, error) {
	var plan intelPlan

	gpus, err := connector.IntelGPUs(runtime)
	if err != nil {
		return plan, err
	}
	if len(gpus) == 0 {
		return plan, nil
	}

	kernelStr := runtime.GetSystemInfo().GetOsKernel()
	running, kerr := connector.ParseKernelVersion(kernelStr)

	for _, g := range gpus {
		kind := "integrated"
		if g.Discrete {
			kind = "discrete"
		}

		// Label by device kind, independent of kernel / Out-of-tree / table presence.
		if g.Discrete {
			plan.labelIntelGPU = true
		} else {
			plan.labelIntel = true
		}

		min, outOfTree, found := connector.IntelGPUMinKernel(g.ID)
		if !found {
			logger.Warnf("Intel %s GPU [%s] is not in the known support table; labeled but driver packages skipped", kind, g.ID)
			continue
		}
		if outOfTree {
			// Out-of-tree parts (Data Center GPU Max/Flex) need the intel-i915-dkms
			// module, which we do not install automatically.
			logger.Warnf("Intel %s GPU [%s] requires the out-of-tree intel-i915-dkms module; it is not installed automatically", kind, g.ID)
			continue
		}
		if kerr != nil {
			logger.Warnf("cannot parse running kernel version %q: %v; skipping driver packages for Intel %s GPU [%s]", kernelStr, kerr, kind, g.ID)
			continue
		}
		if running.Compare(min) < 0 {
			logger.Warnf("running kernel %s is below the minimum %s required for Intel %s GPU [%s]; skipping driver packages", running, min, kind, g.ID)
			continue
		}

		if g.Discrete {
			plan.hasQualifyingDGPU = true
		}
	}

	return plan, nil
}

// LabelIntelGPUs labels the node with the "intel" and/or "intel-gpu" modes based
// on the detected integrated/discrete Intel GPUs that meet the kernel
// requirement (Out-of-tree discrete parts are labeled too).
type LabelIntelGPUs struct {
	common.KubeAction
}

func (u *LabelIntelGPUs) Execute(runtime connector.Runtime) error {
	client, err := clientset.NewKubeClient()
	if err != nil {
		return errors.Wrap(errors.WithStack(err), "kubeclient create error")
	}

	plan, err := classifyIntelGPUs(runtime)
	if err != nil {
		return err
	}
	if !plan.labelIntel && !plan.labelIntelGPU {
		logger.Info("No qualifying Intel GPU to label")
		return nil
	}

	if plan.labelIntel {
		if err := gpu.SetNodeGpuModeLabel(context.Background(), client.CtrlRuntime(), gpu.IntelType, nil, nil, nil); err != nil {
			return err
		}
	}
	if plan.labelIntelGPU {
		if err := gpu.SetNodeGpuModeLabel(context.Background(), client.CtrlRuntime(), gpu.IntelGpuType, nil, nil, nil); err != nil {
			return err
		}
	}
	return nil
}

// HasAnyIntelGPU is a task Prepare that runs when the node has an Intel
// integrated or discrete GPU (matching InstallIntelPluginModule's Skip gate).
type HasAnyIntelGPU struct {
	common.KubePrepare
}

func (p *HasAnyIntelGPU) PreCheck(runtime connector.Runtime) (bool, error) {
	si := runtime.GetSystemInfo()
	return si.IsIntelGPU() || si.IsIntelDGPU(), nil
}

// HasQualifyingIntelDGPU is a task Prepare that only lets the discrete-GPU driver
// install run when there is a discrete Intel GPU that is in-tree and whose
// running kernel meets the minimum requirement.
type HasQualifyingIntelDGPU struct {
	common.KubePrepare
}

func (p *HasQualifyingIntelDGPU) PreCheck(runtime connector.Runtime) (bool, error) {
	plan, err := classifyIntelGPUs(runtime)
	if err != nil {
		return false, err
	}
	return plan.hasQualifyingDGPU, nil
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

// InstallXpumd installs the Intel XPUMD metrics daemon (Helm) for discrete GPUs.
type InstallXpumd struct {
	common.KubeAction
}

func (t *InstallXpumd) Execute(runtime connector.Runtime) error {
	chartPath := path.Join(runtime.GetInstallerDir(), xpumdChartRelDir)
	if !util.IsExist(chartPath) {
		return fmt.Errorf("xpumd chart not found at %s", chartPath)
	}

	config, err := ctrl.GetConfig()
	if err != nil {
		return err
	}
	actionConfig, settings, err := utils.InitConfig(config, xpumdNamespace)
	if err != nil {
		return err
	}

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Minute)
	defer cancel()

	// Chart values.yaml already contains Olares defaults (gpuAccess=plugin, prometheus, etc).
	vals := map[string]interface{}{}
	if err := utils.UpgradeCharts(ctx, actionConfig, settings, xpumdReleaseName, chartPath, "", xpumdNamespace, vals, false); err != nil {
		return errors.Wrap(err, "install/upgrade xpumd chart")
	}

	// Helm does not move resources across namespaces on upgrade; remove the old
	// ServiceMonitor left in kube-system from earlier chart revisions.
	kubectlpath, err := util.GetCommand(common.CommandKubectl)
	if err == nil {
		_, _ = runtime.GetRunner().SudoCmd(
			fmt.Sprintf("%s delete servicemonitor -n %s %s --ignore-not-found", kubectlpath, xpumdNamespace, xpumdReleaseName),
			false, false)
	}

	logger.Info("Intel xpumd chart installed/upgraded")
	return nil
}

// CheckXpumd waits until the xpumd pod on this node is Running.
type CheckXpumd struct {
	common.KubeAction
}

func (t *CheckXpumd) Execute(runtime connector.Runtime) error {
	kubectlpath, err := util.GetCommand(common.CommandKubectl)
	if err != nil {
		return fmt.Errorf("kubectl not found")
	}

	nodeName, err := os.Hostname()
	if err != nil {
		return errors.Wrap(errors.WithStack(err), "get hostname error")
	}
	nodeName = strings.ToLower(nodeName)

	selector := "app.kubernetes.io/name=xpumd,app.kubernetes.io/instance=xpumd"
	fieldSelector := fmt.Sprintf("spec.nodeName=%s", nodeName)
	cmd := fmt.Sprintf("%s get pod -n %s -l '%s' --field-selector '%s' -o jsonpath='{.items[*].status.phase}'",
		kubectlpath, xpumdNamespace, selector, fieldSelector)

	rphase, _ := runtime.GetRunner().SudoCmd(cmd, false, false)
	if hasRunningPod(rphase) {
		return nil
	}
	return fmt.Errorf("xpumd pod state is %q (want Running)", rphase)
}
