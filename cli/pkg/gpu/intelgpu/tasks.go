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
func classifyIntelGPUs(runtime connector.Runtime) intelPlan {
	var plan intelPlan

	gpus := connector.IntelGPUs(runtime)
	if len(gpus) == 0 {
		return plan
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

	return plan
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

	plan := classifyIntelGPUs(runtime)
	if !plan.labelIntel && !plan.labelIntelGPU {
		logger.Info("No qualifying Intel GPU to label")
		return nil
	}

	if plan.labelIntel {
		if err := gpu.SetNodeGpuModeLabel(context.Background(), client.Kubernetes(), gpu.IntelType, nil, nil, nil); err != nil {
			return err
		}
	}
	if plan.labelIntelGPU {
		if err := gpu.SetNodeGpuModeLabel(context.Background(), client.Kubernetes(), gpu.IntelGpuType, nil, nil, nil); err != nil {
			return err
		}
	}
	return nil
}

// HasQualifyingIntelDGPU is a task Prepare that only lets the discrete-GPU driver
// install run when there is a discrete Intel GPU that is in-tree and whose
// running kernel meets the minimum requirement.
type HasQualifyingIntelDGPU struct {
	common.KubePrepare
}

func (p *HasQualifyingIntelDGPU) PreCheck(runtime connector.Runtime) (bool, error) {
	return classifyIntelGPUs(runtime).hasQualifyingDGPU, nil
}

// InstallIntelDGPUDrivers installs the Intel discrete-GPU host driver stack on
// supported Ubuntu by configuring the Intel GPU repository and installing
// intel-omix (the unified discrete-GPU driver stack). Only noble / 24.04 is
// supported. The kobuk-team/intel-graphics PPA is deliberately not used: it
// ships the same compute libraries at newer pinned versions, which conflict with
// intel-omix's exact-version dependencies.
type InstallIntelDGPUDrivers struct {
	common.KubeAction
}

func (t *InstallIntelDGPUDrivers) Execute(runtime connector.Runtime) error {
	si := runtime.GetSystemInfo()
	if !si.IsUbuntu() {
		logger.Warn("Intel discrete GPU host packages are only supported on Ubuntu; skipping package installation")
		return nil
	}
	// intel-omix is only published for noble / 24.04.
	if !si.IsUbuntuVersionEqual(connector.Ubuntu2404) {
		logger.Warnf("Ubuntu version %s is not supported for intel-omix (requires noble/24.04); skipping Intel discrete GPU driver installation", si.GetOsVersion())
		return nil
	}

	run := func(cmd string) error {
		if _, err := runtime.GetRunner().SudoCmd(cmd, false, true); err != nil {
			return errors.Wrap(errors.WithStack(err), fmt.Sprintf("failed to run %q", cmd))
		}
		return nil
	}

	// A previous install may have configured the kobuk-team/intel-graphics PPA and
	// installed its newer compute libraries (libze1, libze-intel-gpu1,
	// intel-opencl-icd, intel-ocloc). Those pinned versions conflict with
	// intel-omix's exact-version dependencies, so remove any leftover PPA source
	// before configuring the Intel repository. rm -f is a no-op when absent.
	if err := run("rm -f /etc/apt/sources.list.d/*kobuk*intel-graphics*.list /etc/apt/sources.list.d/*kobuk*intel-graphics*.sources"); err != nil {
		return err
	}

	// Prerequisites for fetching the repository key.
	if err := run("apt-get update"); err != nil {
		return err
	}
	if err := run("DEBIAN_FRONTEND=noninteractive apt-get install -y gnupg wget"); err != nil {
		return err
	}

	// Install the Intel GPU repository signing key referenced by the apt source.
	if err := run("install -d -m 0755 /usr/share/keyrings"); err != nil {
		return err
	}
	keyCmd := "wget -qO - https://repositories.intel.com/gpu/intel-graphics.key | gpg --yes --dearmor --output /usr/share/keyrings/intel-graphics.gpg"
	if err := run(keyCmd); err != nil {
		return err
	}

	// Configure the Intel GPU repository and install intel-omix (the unified
	// discrete-GPU driver stack). We intentionally do NOT add the
	// kobuk-team/intel-graphics PPA: it ships the same compute libraries at newer
	// pinned versions, which conflict with intel-omix's exact-version dependencies.
	const codename = "noble"
	srcLine := fmt.Sprintf("deb [arch=amd64 signed-by=/usr/share/keyrings/intel-graphics.gpg] https://repositories.intel.com/gpu/ubuntu %s/intel-omix/0.2 unified", codename)
	writeSrc := fmt.Sprintf("echo '%s' | tee /etc/apt/sources.list.d/intel-gpu-%s.list", srcLine, codename)
	if err := run(writeSrc); err != nil {
		return err
	}
	if err := run("apt-get update"); err != nil {
		return err
	}

	// --allow-downgrades lets apt replace any newer kobuk-installed compute libs
	// left over from a previous run with intel-omix's pinned versions.
	const installOmix = "DEBIAN_FRONTEND=noninteractive apt-get install -y --allow-downgrades intel-omix"
	if _, err := runtime.GetRunner().SudoCmd(installOmix, false, true); err != nil {
		// The first attempt can still fail on a machine where the leftover
		// kobuk-versioned compute libs cannot be downgraded in place. Purge the
		// conflicting packages (best-effort) and let intel-omix reinstall its own
		// pinned versions, then retry.
		logger.Warn("intel-omix install failed; purging leftover conflicting Intel compute packages and retrying")
		_, _ = runtime.GetRunner().SudoCmd("DEBIAN_FRONTEND=noninteractive apt-get purge -y libze1 libze-dev libze-intel-gpu1 intel-opencl-icd intel-ocloc || true", false, true)
		_, _ = runtime.GetRunner().SudoCmd("apt-get update", false, true)
		if err := run(installOmix); err != nil {
			return err
		}
	}

	logger.Info("Intel discrete GPU host packages (intel-omix) installed successfully")
	return nil
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
