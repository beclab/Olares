package upgrade

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"os/exec"
	"path"
	"path/filepath"
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
	appv1alpha1 "github.com/beclab/Olares/framework/app-service/api/app.bytetrade.io/v1alpha1"
	"github.com/beclab/Olares/framework/app-service/pkg/appcfg"

	"github.com/pkg/errors"
	appsv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/resource"
	kruntime "k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/types"
	"k8s.io/utils/ptr"
	ctrl "sigs.k8s.io/controller-runtime"
	ctrlclient "sigs.k8s.io/controller-runtime/pkg/client"
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
			Name:   "UpgradeKSCore",
			Action: new(plugins.CreateKsCore),
			Retry:  10,
			Delay:  10 * time.Second,
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

// applyNodeExporterAction applies embedded node-exporter
type applyNodeExporterAction struct {
	common.KubeAction
}

func (a *applyNodeExporterAction) Execute(runtime connector.Runtime) error {
	kubectlpath, err := util.GetCommand(common.CommandKubectl)
	if err != nil {
		return errors.Wrap(errors.WithStack(err), "kubectl not found")
	}
	manifest := path.Join(runtime.GetInstallerDir(), cc.BuildFilesCacheDir, cc.BuildDir, "prometheus", "node-exporter", "node-exporter-daemonset.yaml")
	if _, err := runtime.GetRunner().SudoCmd(fmt.Sprintf("%s apply -f %s", kubectlpath, manifest), false, true); err != nil {
		return errors.Wrap(errors.WithStack(err), "apply node-exporter failed")
	}
	return nil
}

func upgradeNodeExporter() []task.Interface {
	return []task.Interface{
		&task.LocalTask{
			Name:   "CopyEmbeddedKSManifests",
			Action: new(plugins.CopyEmbedFiles),
		},
		&task.LocalTask{
			Name:   "applyNodeExporterManifests",
			Action: new(applyNodeExporterAction),
		},
	}
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

// fixProcModprobePath fixes the /proc/sys/kernel/modprobe path issue that can cause
// nvidia-installer to fail with error:
// "The path to the `modprobe` utility reported by '/proc/sys/kernel/modprobe', ”, differs from
// the path determined by `nvidia-installer`, `/bin/kmod`, and does not appear to point to a
// valid `modprobe` binary."
//
// This function checks if /proc/sys/kernel/modprobe is empty or invalid, and if so,
// writes a valid modprobe path to it.
func fixProcModprobePath() {
	const procModprobePath = "/proc/sys/kernel/modprobe"

	modprobePaths := []string{
		"/sbin/modprobe",
		"/usr/sbin/modprobe",
		"/bin/modprobe",
		"/usr/bin/modprobe",
	}

	data, err := os.ReadFile(procModprobePath)
	if err != nil {
		logger.Warnf("failed to read %s: %v", procModprobePath, err)
	}
	currentPath := strings.TrimSpace(string(data))

	// Check if current path is valid (non-empty and executable)
	if currentPath != "" {
		if util.IsExecutable(currentPath) {
			logger.Debugf("%s already contains valid path: %s", procModprobePath, currentPath)
			return
		}
		// in case it's a symlink that resolves to a valid executable
		if resolved, err := filepath.EvalSymlinks(currentPath); err == nil && resolved != "" {
			if util.IsExecutable(resolved) {
				logger.Debugf("%s contains symlink %s -> %s which is valid", procModprobePath, currentPath, resolved)
				return
			}
		}
		logger.Warnf("%s contains invalid path: '%s', attempting to fix", procModprobePath, currentPath)
	} else {
		logger.Warnf("%s is empty, attempting to fix", procModprobePath)
	}

	if lookPath, err := exec.LookPath("modprobe"); err == nil && lookPath != "" {
		modprobePaths = append([]string{lookPath}, modprobePaths...)
	}

	for _, modprobePath := range modprobePaths {
		if !util.IsExecutable(modprobePath) {
			continue
		}

		if err := os.WriteFile(procModprobePath, []byte(modprobePath), 0644); err != nil {
			logger.Warnf("failed to write %s to %s: %v", modprobePath, procModprobePath, err)
			continue
		}

		logger.Infof("successfully fixed %s: set to %s", procModprobePath, modprobePath)
		return
	}

	// If we get here, we couldn't fix it, but we log a warning and continue
	// The nvidia-installer might still work, or it might fail, but we don't want to block the upgrade
	logger.Warnf("could not fix %s, nvidia-installer may fail; continuing anyway", procModprobePath)
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

		fixProcModprobePath()

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
		err = gpu.UpdateNodeGpuLabel(context.Background(), client.Kubernetes(), &targetDriverVersionStr, ptr.To(common.CurrentVerifiedCudaVersion), ptr.To("true"), ptr.To(gpu.NvidiaCardType))
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

// applyNodeExporterServiceMonitorAction applies embedded prometheus node-exporter ServiceMonitor
type applyNodeExporterServiceMonitorAction struct {
	common.KubeAction
}

func (a *applyNodeExporterServiceMonitorAction) Execute(runtime connector.Runtime) error {
	kubectlpath, err := util.GetCommand(common.CommandKubectl)
	if err != nil {
		return errors.Wrap(errors.WithStack(err), "kubectl not found")
	}
	manifest := path.Join(runtime.GetInstallerDir(), cc.BuildFilesCacheDir, cc.BuildDir, "prometheus", "node-exporter", "node-exporter-serviceMonitor.yaml")
	if _, err := runtime.GetRunner().SudoCmd(fmt.Sprintf("%s apply -f %s", kubectlpath, manifest), false, true); err != nil {
		return errors.Wrap(errors.WithStack(err), "apply node-exporter ServiceMonitor failed")
	}
	return nil
}

// applyKubernetesPrometheusRuleAction applies embedded prometheus kubernetes prometheusRule
type applyKubernetesPrometheusRuleAction struct {
	common.KubeAction
}

func (a *applyKubernetesPrometheusRuleAction) Execute(runtime connector.Runtime) error {
	kubectlpath, err := util.GetCommand(common.CommandKubectl)
	if err != nil {
		return errors.Wrap(errors.WithStack(err), "kubectl not found")
	}
	manifest := path.Join(runtime.GetInstallerDir(), cc.BuildFilesCacheDir, cc.BuildDir, "prometheus", "kubernetes", "kubernetes-prometheusRule.yaml")
	if _, err := runtime.GetRunner().SudoCmd(fmt.Sprintf("%s apply -f %s", kubectlpath, manifest), false, true); err != nil {
		return errors.Wrap(errors.WithStack(err), "apply kubernetes prometheusRule failed")
	}
	return nil
}

func upgradeNodeExporterServiceMonitor() []task.Interface {
	return []task.Interface{
		// prometheus node-exporter ServiceMonitor
		&task.LocalTask{
			Name:   "ApplyNodeExporterServiceMonitor",
			Action: new(applyNodeExporterServiceMonitorAction),
			Retry:  5,
			Delay:  5 * time.Second,
		},
	}
}

func upgradeKubernetesPrometheusRule() []task.Interface {
	return []task.Interface{
		// prometheus kubernetes prometheusRule
		&task.LocalTask{
			Name:   "ApplyKubernetesPrometheusRule",
			Action: new(applyKubernetesPrometheusRuleAction),
			Retry:  5,
			Delay:  5 * time.Second,
		},
	}
}

type waitForStatefulSetReady struct {
	common.KubeAction
	Namespace string
	Name      string
	InitDelay time.Duration
}

func (w *waitForStatefulSetReady) Execute(_ connector.Runtime) error {
	if w.InitDelay > 0 {
		logger.Infof("waiting %s before checking statefulset %s/%s", w.InitDelay, w.Namespace, w.Name)
		time.Sleep(w.InitDelay)
	}

	config, err := ctrl.GetConfig()
	if err != nil {
		return errors.Wrap(errors.WithStack(err), "failed to get kubernetes config")
	}

	scheme := kruntime.NewScheme()
	if err := appsv1.AddToScheme(scheme); err != nil {
		return errors.Wrap(errors.WithStack(err), "failed to add apps/v1 scheme")
	}

	c, err := ctrlclient.New(config, ctrlclient.Options{Scheme: scheme})
	if err != nil {
		return errors.Wrap(errors.WithStack(err), "failed to create client")
	}

	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	var sts appsv1.StatefulSet
	key := ctrlclient.ObjectKey{Namespace: w.Namespace, Name: w.Name}
	if err := c.Get(ctx, key, &sts); err != nil {
		return errors.Wrapf(errors.WithStack(err), "failed to get statefulset %s/%s", w.Namespace, w.Name)
	}

	if sts.Status.ObservedGeneration < sts.Generation {
		return fmt.Errorf("statefulset %s/%s rollout not observed yet (generation %d, observed %d)",
			w.Namespace, w.Name, sts.Generation, sts.Status.ObservedGeneration)
	}

	replicas := int32(1)
	if sts.Spec.Replicas != nil {
		replicas = *sts.Spec.Replicas
	}

	if sts.Status.UpdatedReplicas < replicas {
		return fmt.Errorf("statefulset %s/%s not fully updated: %d/%d updated",
			w.Namespace, w.Name, sts.Status.UpdatedReplicas, replicas)
	}

	if sts.Status.ReadyReplicas < replicas {
		return fmt.Errorf("statefulset %s/%s not ready: %d/%d ready",
			w.Namespace, w.Name, sts.Status.ReadyReplicas, replicas)
	}

	if sts.Status.CurrentRevision != sts.Status.UpdateRevision {
		return fmt.Errorf("statefulset %s/%s revision mismatch: current=%s update=%s",
			w.Namespace, w.Name, sts.Status.CurrentRevision, sts.Status.UpdateRevision)
	}

	logger.Infof("statefulset %s/%s is ready", w.Namespace, w.Name)
	return nil
}

type backfillAppGPUConfig struct {
	common.KubeAction
}

var gpuMemoryByRawAppName = map[string]string{
	"ollamallama318bv2":  "8Gi",
	"ollamaminicpmv8bv2": "8Gi",
	"vllmhymt1518bv2":    "8Gi",

	"ollamacogito14bv2":     "12Gi",
	"ollamadeepseekr114bv2": "12Gi",
	"ollamallava1613bv2":    "12Gi",
	"ollamaphi414bv2":       "12Gi",
	"ollamaqwen314bv2":      "12Gi",
	"ollamaqwen359bv2":      "12Gi",
	"deepseekocrwebuiv2":    "12Gi",
	"vllmhymt157bv2":        "12Gi",

	"ollamagptoss20bv2": "19Gi",
	"indexttsv2":        "19Gi",
	"vllmgemma312bitv2": "19Gi",

	"ollamagemma327bv2":             "23Gi",
	"ollamaglm47flashv2":            "23Gi",
	"ollamaqwen330ba3bv2":           "23Gi",
	"ollamaqwen3527bq4kmv2":         "23Gi",
	"llamacppgptoss120bggufv2":      "23Gi",
	"vllmqwen330ba3binstruct4bitv2": "23Gi",
	"vllmgptoss20bv2":               "23Gi",
	"vllmgemma327bqatv2":            "23Gi",
}

const defaultGPUMemory = "2Gi"

func (a *backfillAppGPUConfig) Execute(_ connector.Runtime) error {
	config, err := ctrl.GetConfig()
	if err != nil {
		return errors.Wrap(errors.WithStack(err), "failed to get kubernetes config")
	}

	scheme := kruntime.NewScheme()
	if err := appv1alpha1.AddToScheme(scheme); err != nil {
		return errors.Wrap(errors.WithStack(err), "failed to add app-service scheme")
	}
	if err := appsv1.AddToScheme(scheme); err != nil {
		return errors.Wrap(errors.WithStack(err), "failed to add apps/v1 scheme")
	}
	if err := corev1.AddToScheme(scheme); err != nil {
		return errors.Wrap(errors.WithStack(err), "failed to add corev1 scheme")
	}

	c, err := ctrlclient.New(config, ctrlclient.Options{Scheme: scheme})
	if err != nil {
		return errors.Wrap(errors.WithStack(err), "failed to create controller-runtime client")
	}

	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Minute)
	defer cancel()

	var amList appv1alpha1.ApplicationManagerList
	if err := c.List(ctx, &amList); err != nil {
		return errors.Wrap(errors.WithStack(err), "failed to list applicationmanagers")
	}
	var gpuType = "nvidia"
	var nodeList corev1.NodeList
	if err := c.List(ctx, &nodeList); err != nil {
		return errors.Wrap(errors.WithStack(err), "failed to list nodes")
	}
	for _, node := range nodeList.Items {
		annoGpuType := node.Annotations["gpu.bytetrade.io/type"]
		if annoGpuType != "" {
			gpuType = annoGpuType
			break
		}
	}

	patchedCount := 0
	for i := range amList.Items {
		am := &amList.Items[i]
		if am.Spec.Config == "" {
			continue
		}

		var appCfg appcfg.ApplicationConfig
		if err := json.Unmarshal([]byte(am.Spec.Config), &appCfg); err != nil {
			return errors.Wrapf(errors.WithStack(err), "failed to unmarshal config for applicationmanager %s", am.Name)
		}

		if appCfg.RequiredGPU == "" {
			continue
		}

		modified := false

		gpuMem, ok := gpuMemoryByRawAppName[am.Spec.RawAppName]
		if !ok {
			gpuMem = defaultGPUMemory
		}
		q := resource.MustParse(gpuMem)

		if appCfg.RequiredGPU != q.String() {
			appCfg.RequiredGPU = q.String()
			modified = true
		}

		if appCfg.Requirement.GPU == nil || !appCfg.Requirement.GPU.Equal(q) {
			appCfg.Requirement.GPU = &q
			modified = true
		}

		if appCfg.SelectedGpuType == "" {
			appCfg.SelectedGpuType = gpuType
			modified = true
		}

		if !modified {
			continue
		}

		updatedConfig, err := json.Marshal(&appCfg)
		if err != nil {
			return errors.Wrapf(errors.WithStack(err), "failed to marshal updated config for %s", am.Name)
		}

		patchObj := map[string]interface{}{
			"spec": map[string]interface{}{
				"config": string(updatedConfig),
			},
		}
		patchContent, err := json.Marshal(patchObj)
		if err != nil {
			return errors.Wrapf(errors.WithStack(err), "failed to build patch for %s", am.Name)
		}

		if err := c.Patch(ctx, am, ctrlclient.RawPatch(types.MergePatchType, patchContent)); err != nil {
			return errors.Wrapf(errors.WithStack(err), "failed to patch applicationmanager %s", am.Name)
		}

		logger.Infof("backfilled GPU config for applicationmanager %s", am.Name)
		patchedCount++

		if err := annotateWorkloadsForGPUBackfill(ctx, c, am, &appCfg); err != nil {
			return errors.Wrapf(errors.WithStack(err), "failed to annotate workloads for %s", am.Name)
		}
	}

	logger.Infof("backfilled GPU config for %d applicationmanagers", patchedCount)
	return nil
}

func annotateWorkloadsForGPUBackfill(ctx context.Context, c ctrlclient.Client, am *appv1alpha1.ApplicationManager, appCfg *appcfg.ApplicationConfig) error {
	var namespaces []string
	if appCfg.IsMultiCharts() {
		for _, chart := range appCfg.SubCharts {
			namespaces = append(namespaces, chart.Namespace(am.Spec.AppOwner))
		}
	} else {
		if am.Spec.AppNamespace == "" {
			return nil
		}
		namespaces = []string{am.Spec.AppNamespace}
	}

	timestamp := time.Now().Format(time.RFC3339)
	annotationPatch := ctrlclient.RawPatch(types.MergePatchType,
		[]byte(fmt.Sprintf(`{"metadata":{"annotations":{"bytetrade.io/upgrade-gpu-backfill":"%s"}}}`, timestamp)))
	for _, ns := range namespaces {
		var deployList appsv1.DeploymentList
		if err := c.List(ctx, &deployList, ctrlclient.InNamespace(ns)); err != nil {
			return errors.Wrapf(errors.WithStack(err), "failed to list deployments in %s", ns)
		}
		for i := range deployList.Items {
			d := &deployList.Items[i]
			if err := c.Patch(ctx, d, annotationPatch); err != nil {
				return errors.Wrapf(errors.WithStack(err), "failed to annotate deployment %s/%s", ns, d.Name)
			}
			logger.Infof("annotated deployment %s/%s for GPU config backfill", ns, d.Name)
		}

		var stsList appsv1.StatefulSetList
		if err := c.List(ctx, &stsList, ctrlclient.InNamespace(ns)); err != nil {
			return errors.Wrapf(errors.WithStack(err), "failed to list statefulsets in %s", ns)
		}
		for i := range stsList.Items {
			s := &stsList.Items[i]
			if err := c.Patch(ctx, s, annotationPatch); err != nil {
				return errors.Wrapf(errors.WithStack(err), "failed to annotate statefulset %s/%s", ns, s.Name)
			}
			logger.Infof("annotated statefulset %s/%s for GPU config backfill", ns, s.Name)
		}
	}

	return nil
}
