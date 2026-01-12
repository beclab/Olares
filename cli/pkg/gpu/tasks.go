package gpu

import (
	"context"
	"fmt"
	"net/url"
	"os"
	"path"
	"path/filepath"
	"strings"
	"time"

	v1alpha1 "github.com/beclab/Olares/framework/app-service/api/sys.bytetrade.io/v1alpha1"
	apputils "github.com/beclab/Olares/framework/app-service/pkg/utils"
	ctrl "sigs.k8s.io/controller-runtime"
	ctrlclient "sigs.k8s.io/controller-runtime/pkg/client"

	"github.com/beclab/Olares/cli/pkg/clientset"
	"github.com/beclab/Olares/cli/pkg/common"
	cc "github.com/beclab/Olares/cli/pkg/core/common"
	"github.com/beclab/Olares/cli/pkg/core/connector"
	"github.com/beclab/Olares/cli/pkg/core/logger"
	"github.com/beclab/Olares/cli/pkg/core/util"
	k3sGpuTemplates "github.com/beclab/Olares/cli/pkg/gpu/templates"
	"github.com/beclab/Olares/cli/pkg/manifest"
	"github.com/beclab/Olares/cli/pkg/utils"
	criconfig "github.com/containerd/containerd/pkg/cri/config"
	cdsrvconfig "github.com/containerd/containerd/services/server/config"
	"github.com/pelletier/go-toml"

	"github.com/pkg/errors"
	apixclientset "k8s.io/apiextensions-apiserver/pkg/client/clientset/clientset"
	apierrors "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	kruntime "k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/types"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/util/retry"
)

type CheckWslGPU struct {
}

func (t *CheckWslGPU) CheckNvidiaSmiFileExists() bool {
	var nvidiaSmiFile = "/usr/lib/wsl/lib/nvidia-smi"
	return util.IsExist(nvidiaSmiFile)
}

func (t *CheckWslGPU) Execute(runtime *common.KubeRuntime) {
	if !runtime.GetSystemInfo().IsWsl() {
		return
	}
	exists := t.CheckNvidiaSmiFileExists()
	if !exists {
		return
	}

	stdout, _, err := util.Exec(context.Background(), "/usr/lib/wsl/lib/nvidia-smi -L|grep 'NVIDIA'|grep UUID", false, false)
	if err != nil {
		logger.Errorf("nvidia-smi not found")
		return
	}
	if stdout == "" {
		return
	}

	runtime.Arg.SetGPU(true)
}

type InstallCudaDriver struct {
	common.KubeAction
	manifest.ManifestAction
}

func (t *InstallCudaDriver) Execute(runtime connector.Runtime) error {
	_, _ = runtime.GetRunner().SudoCmd("apt-get update", false, true)
	// install build deps for dkms
	if _, err := runtime.GetRunner().SudoCmd("DEBIAN_FRONTEND=noninteractive apt-get install -y --no-install-recommends dkms build-essential linux-headers-$(uname -r)", false, true); err != nil {
		return errors.Wrap(errors.WithStack(err), "failed to install kernel build dependencies for nvidia runfile")
	}

	// fetch runfile from manifest
	item, err := t.Manifest.Get("cuda-driver")
	if err != nil {
		return err
	}
	runfile := item.FilePath(t.BaseDir)
	if !util.IsExist(runfile) {
		return fmt.Errorf("failed to find %s binary in %s", item.Filename, runfile)
	}
	if _, err := runtime.GetRunner().SudoCmd(fmt.Sprintf("chmod +x %s", runfile), false, true); err != nil {
		return errors.Wrap(errors.WithStack(err), "failed to chmod +x runfile")
	}
	// execute runfile with required flags
	cmd := fmt.Sprintf("sh %s -z --no-x-check --allow-installation-with-running-driver --no-check-for-alternate-installs --dkms --rebuild-initramfs -s", runfile)
	if _, err := runtime.GetRunner().SudoCmd(cmd, false, true); err != nil {
		return errors.Wrap(errors.WithStack(err), "failed to install nvidia driver via runfile")
	}

	// now that the nvidia driver is installed,
	// the nvidia-smi should work correctly,
	// if not, a manual reboot is needed by the user
	st, err := utils.GetNvidiaStatus(runtime)
	if err != nil || st == nil || !st.Installed || st.Mismatch {
		logger.Error("ERROR: nvidia driver has been installed, but is not running properly, please reboot the machine and try again")
		os.Exit(1)
	}

	return nil
}

type UpdateNvidiaContainerToolkitSource struct {
	common.KubeAction
	manifest.ManifestAction
}

func (t *UpdateNvidiaContainerToolkitSource) Execute(runtime connector.Runtime) error {
	var cmd string
	gpgkey, err := t.Manifest.Get("libnvidia-gpgkey")
	if err != nil {
		return err
	}

	keyPath := gpgkey.FilePath(t.BaseDir)

	if !util.IsExist(keyPath) {
		return fmt.Errorf("failed to find %s binary in %s", gpgkey.Filename, keyPath)
	}

	if _, err := runtime.GetRunner().SudoCmd("install -d -m 0755 /usr/share/keyrings", false, true); err != nil {
		return err
	}
	keyringPath := "/usr/share/keyrings/nvidia-container-toolkit-keyring.gpg"
	cmd = fmt.Sprintf("gpg --batch --yes --dearmor -o %s %s", keyringPath, keyPath)
	if _, err := runtime.GetRunner().SudoCmd(cmd, false, true); err != nil {
		return err
	}

	libnvidia, err := t.Manifest.Get("libnvidia-container.list")
	if err != nil {
		return err
	}

	libPath := libnvidia.FilePath(t.BaseDir)

	if !util.IsExist(libPath) {
		return fmt.Errorf("failed to find %s binary in %s", libnvidia.Filename, libPath)
	}

	// remove any conflicting libnvidia-container.list
	_, err = runtime.GetRunner().SudoCmd("rm -rf /etc/apt/sources.list.d/*nvidia-container*.list", false, false)
	if err != nil {
		return err
	}

	dstPath := filepath.Join("/etc/apt/sources.list.d", filepath.Base(libPath))
	cmd = fmt.Sprintf("cp %s %s", libPath, dstPath)
	if _, err := runtime.GetRunner().SudoCmd(cmd, false, true); err != nil {
		return err
	}

	if _, err := runtime.GetRunner().SudoCmd(fmt.Sprintf("sed -i 's#^deb https://#deb [signed-by=%s] https://#' %s", keyringPath, dstPath), false, true); err != nil {
		return err
	}

	// decide mirror based on OLARES_SYSTEM_CDN_SERVICE
	var mirrorHost string
	cdnService := os.Getenv(common.ENV_OLARES_CDN_SERVICE)
	if cdnService != "" {
		cdnRaw := cdnService
		if !strings.HasPrefix(cdnRaw, "http") {
			cdnRaw = "https://" + cdnRaw
		}
		if cdnURL, err := url.Parse(cdnRaw); err == nil {
			host := cdnURL.Host
			if host == "" {
				host = cdnService
			}
			if strings.HasSuffix(host, "olares.cn") {
				mirrorHost = "mirrors.ustc.edu.cn"
			}
		} else if strings.HasSuffix(cdnService, "olares.cn") {
			mirrorHost = "mirrors.ustc.edu.cn"
		}
	}
	if mirrorHost == "" {
		return nil
	}
	cmd = fmt.Sprintf("sed -i 's#nvidia.github.io#%s#g' %s", mirrorHost, dstPath)
	if _, err := runtime.GetRunner().SudoCmd(cmd, false, false); err != nil {
		return errors.Wrap(errors.WithStack(err), "failed to switch nvidia container repo to mirror site")
	}
	return nil
}

type InstallNvidiaContainerToolkit struct {
	common.KubeAction
}

func (t *InstallNvidiaContainerToolkit) Execute(runtime connector.Runtime) error {
	containerdDropInDir := "/etc/containerd/config.d"
	containerdConfigFile := "/etc/containerd/config.toml"
	if util.IsExist(containerdDropInDir) {
		if err := os.RemoveAll(containerdDropInDir); err != nil {
			return errors.Wrap(errors.WithStack(err), "Failed to remove containerd drop-in directory")
		}
	}
	if util.IsExist(containerdConfigFile) {
		if _, err := runtime.GetRunner().SudoCmd(fmt.Sprintf("sed -i '/^import/d' %s", containerdConfigFile), false, false); err != nil {
			return errors.Wrap(errors.WithStack(err), "Failed to remove import section from containerd config file")
		}
	}
	logger.Debugf("install nvidia-container-toolkit")
	if _, err := runtime.GetRunner().SudoCmd("apt-get update && sudo apt-get install -y --allow-downgrades nvidia-container-toolkit=1.17.9-1 nvidia-container-toolkit-base=1.17.9-1 jq", false, true); err != nil {
		return errors.Wrap(errors.WithStack(err), "failed to apt-get install nvidia-container-toolkit")
	}
	return nil
}

type PatchK3sDriver struct { // patch k3s on wsl
	common.KubeAction
}

func (t *PatchK3sDriver) Execute(runtime connector.Runtime) error {
	if !runtime.GetSystemInfo().IsWsl() {
		return nil
	}
	var cmd = "find /usr/lib/wsl/drivers/ -name libcuda.so.1.1|head -1"
	driverPath, err := runtime.GetRunner().SudoCmd(cmd, false, true)
	if err != nil {
		return err
	}

	if driverPath == "" {
		logger.Infof("cuda driver not found")
		return nil
	} else {
		logger.Infof("cuda driver found: %s", driverPath)
	}

	templateStr, err := util.Render(k3sGpuTemplates.K3sCudaFixValues, nil)
	if err != nil {
		return errors.Wrap(errors.WithStack(err), fmt.Sprintf("render template %s failed", k3sGpuTemplates.K3sCudaFixValues.Name()))
	}

	var fixName = "cuda_lib_fix.sh"
	var fixPath = path.Join(runtime.GetBaseDir(), cc.PackageCacheDir, "gpu", "cuda_lib_fix.sh")
	if err := util.WriteFile(fixPath, []byte(templateStr), cc.FileMode0755); err != nil {
		return errors.Wrap(errors.WithStack(err), fmt.Sprintf("write file %s failed", fixPath))
	}

	var dstName = path.Join(common.BinDir, fixName)
	if err := runtime.GetRunner().SudoScp(fixPath, dstName); err != nil {
		return errors.Wrap(errors.WithStack(err), fmt.Sprintf("scp file %s to remote %s failed", fixPath, dstName))
	}

	cmd = fmt.Sprintf("echo 'ExecStartPre=-/usr/local/bin/%s' >> /etc/systemd/system/k3s.service", fixName)
	if _, err := runtime.GetRunner().SudoCmd(cmd, false, false); err != nil {
		return err
	}

	if _, err := runtime.GetRunner().SudoCmd("systemctl daemon-reload", false, false); err != nil {
		return err
	}

	if _, err := runtime.GetRunner().SudoCmd("apt install -y strace", false, false); err != nil {
		return err
	}

	if _, err := runtime.GetRunner().SudoCmd(dstName, false, false); err != nil {
		return errors.Wrap(err, "failed to apply CUDA patch for WSL")
	}

	return nil
}

type ConfigureContainerdRuntime struct {
	common.KubeAction
}

func (t *ConfigureContainerdRuntime) Execute(runtime connector.Runtime) error {
	if _, err := runtime.GetRunner().SudoCmd("nvidia-ctk runtime configure --runtime=containerd --set-as-default --config-source=file", false, true); err != nil {
		return errors.Wrap(errors.WithStack(err), "Failed to nvidia-ctk runtime configure")
	}

	return nil
}

type InstallPlugin struct {
	common.KubeAction
}

func (t *InstallPlugin) Execute(runtime connector.Runtime) error {
	chartPath := path.Join(runtime.GetInstallerDir(), "wizard/config/gpu/hami")
	appName := "hami"
	ns := "kube-system"
	config, err := ctrl.GetConfig()
	if err != nil {
		return err
	}
	actionConfig, settings, err := utils.InitConfig(config, ns)
	if err != nil {
		return err
	}

	var ctx, cancel = context.WithTimeout(context.Background(), 5*time.Minute)
	defer cancel()

	vals := make(map[string]interface{})

	if err := utils.UpgradeCharts(ctx, actionConfig, settings, appName, chartPath, "", ns, vals, false); err != nil {
		return err
	}

	return nil
}

type CheckGpuStatus struct {
	common.KubeAction
}

func (t *CheckGpuStatus) Execute(runtime connector.Runtime) error {
	kubectlpath, err := util.GetCommand(common.CommandKubectl)
	if err != nil {
		return fmt.Errorf("kubectl not found")
	}

	cmd := fmt.Sprintf("%s get pod  -n kube-system -l 'app.kubernetes.io/component=hami-device-plugin' -o jsonpath='{.items[*].status.phase}'", kubectlpath)

	rphase, _ := runtime.GetRunner().SudoCmd(cmd, false, false)
	if rphase == "Running" {
		return nil
	}
	return fmt.Errorf("GPU Container State is Pending")
}

type UpdateNodeGPUInfo struct {
	common.KubeAction
}

func (u *UpdateNodeGPUInfo) Execute(runtime connector.Runtime) error {
	client, err := clientset.NewKubeClient()
	if err != nil {
		return errors.Wrap(errors.WithStack(err), "kubeclient create error")
	}

	st, err := utils.GetNvidiaStatus(runtime)
	if err != nil {
		return err
	}
	if st == nil || !st.Installed {
		logger.Info("NVIDIA driver is not installed")
		return nil
	}

	supported := "false"
	if st.Installed {
		supported = "true"
	}

	driverVersion := st.DriverVersion
	if st.Mismatch && st.LibraryVersion != "" {
		driverVersion = st.LibraryVersion
	}

	return UpdateNodeGpuLabel(context.Background(), client.Kubernetes(), &driverVersion, &st.CudaVersion, &supported)
}

type RemoveNodeLabels struct {
	common.KubeAction
}

func (u *RemoveNodeLabels) Execute(runtime connector.Runtime) error {
	client, err := clientset.NewKubeClient()
	if err != nil {
		return errors.Wrap(errors.WithStack(err), "kubeclient create error")
	}

	return UpdateNodeGpuLabel(context.Background(), client.Kubernetes(), nil, nil, nil)
}

// update k8s node labels gpu.bytetrade.io/driver and gpu.bytetrade.io/cuda.
// if labels are not exists, create it.
func UpdateNodeGpuLabel(ctx context.Context, client kubernetes.Interface, driver, cuda *string, supported *string) error {
	// get node name from hostname
	nodeName, err := os.Hostname()
	if err != nil {
		logger.Error("get hostname error, ", err)
		return err
	}

	node, err := client.CoreV1().Nodes().Get(ctx, nodeName, metav1.GetOptions{})
	if err != nil {
		logger.Error("get node error, ", err)
		return err
	}

	labels := node.GetLabels()
	if labels == nil {
		labels = make(map[string]string)
	}

	update := false
	for _, label := range []struct {
		key   string
		value *string
	}{
		{GpuDriverLabel, driver},
		{GpuCudaLabel, cuda},
		{GpuCudaSupportedLabel, supported},
	} {
		old, ok := labels[label.key]
		switch {
		case ok && label.value == nil: // delete label
			delete(labels, label.key)
			update = true

		case ok && *label.value != "" && old != *label.value: // update label
			labels[label.key] = *label.value
			update = true

		case !ok && label.value != nil && *label.value != "": // create label
			labels[label.key] = *label.value
			update = true
		}
	}

	if update {
		node.SetLabels(labels)
		safeString := func(s *string) string {
			if s == nil {
				return "nil"
			}
			return *s
		}
		err = retry.RetryOnConflict(retry.DefaultRetry, func() error {
			logger.Infof("updating node gpu labels, %s, %s", safeString(driver), safeString(cuda))
			_, err := client.CoreV1().Nodes().Update(ctx, node, metav1.UpdateOptions{})
			return err
		})

		if err != nil {
			logger.Error("update node error, ", err)
			return err
		}
	}

	if cuda != nil && *cuda != "" {
		if err := updateCudaVersionSystemEnv(ctx, *cuda); err != nil {
			logger.Errorf("failed to update SystemEnv for CUDA version: %v", err)
			return err
		}
	}

	return nil
}

func updateCudaVersionSystemEnv(ctx context.Context, cudaVersion string) error {
	envName := "OLARES_SYSTEM_CUDA_VERSION"
	common.SetSystemEnv(envName, cudaVersion)
	config, err := ctrl.GetConfig()
	if err != nil {
		return fmt.Errorf("failed to get rest config: %w", err)
	}

	apix, err := apixclientset.NewForConfig(config)
	if err != nil {
		return fmt.Errorf("failed to create crd client: %w", err)
	}

	_, err = apix.ApiextensionsV1().CustomResourceDefinitions().Get(ctx, "systemenvs.sys.bytetrade.io", metav1.GetOptions{})
	if err != nil {
		if apierrors.IsNotFound(err) {
			logger.Debugf("SystemEnv CRD not found, skipping CUDA version update")
			return nil
		}
		return fmt.Errorf("failed to get SystemEnv CRD: %w", err)
	}

	scheme := kruntime.NewScheme()
	if err := v1alpha1.AddToScheme(scheme); err != nil {
		return fmt.Errorf("failed to add systemenv scheme: %w", err)
	}

	c, err := ctrlclient.New(config, ctrlclient.Options{Scheme: scheme})
	if err != nil {
		return fmt.Errorf("failed to create client: %w", err)
	}

	resourceName, err := apputils.EnvNameToResourceName(envName)
	if err != nil {
		return fmt.Errorf("invalid system env name: %s", envName)
	}

	var existingSystemEnv v1alpha1.SystemEnv
	err = c.Get(ctx, types.NamespacedName{Name: resourceName}, &existingSystemEnv)
	if err == nil {
		if existingSystemEnv.Default != cudaVersion {
			existingSystemEnv.Default = cudaVersion
			if err := c.Update(ctx, &existingSystemEnv); err != nil {
				return fmt.Errorf("failed to update SystemEnv %s: %w", resourceName, err)
			}
			logger.Infof("Updated SystemEnv %s default to %s", resourceName, cudaVersion)
		}
		return nil
	}

	if !apierrors.IsNotFound(err) {
		return fmt.Errorf("failed to get SystemEnv %s: %w", resourceName, err)
	}

	systemEnv := &v1alpha1.SystemEnv{
		ObjectMeta: metav1.ObjectMeta{
			Name: resourceName,
		},
		EnvVarSpec: v1alpha1.EnvVarSpec{
			EnvName: envName,
			Default: cudaVersion,
		},
	}

	if err := c.Create(ctx, systemEnv); err != nil && !apierrors.IsAlreadyExists(err) {
		return fmt.Errorf("failed to create SystemEnv %s: %w", resourceName, err)
	}

	logger.Infof("Created SystemEnv: %s with default %s", envName, cudaVersion)
	return nil
}

type RemoveContainerRuntimeConfig struct {
	common.KubeAction
}

func (t *RemoveContainerRuntimeConfig) Execute(runtime connector.Runtime) error {
	var configFile = "/etc/containerd/config.toml"
	var nvidiaRuntime = "nvidia"
	var criPluginUri = "io.containerd.grpc.v1.cri"

	if !util.IsExist(configFile) {
		logger.Infof("containerd config file not found")
		return nil
	}

	config := &cdsrvconfig.Config{}
	err := cdsrvconfig.LoadConfig(configFile, config)
	if err != nil {
		return fmt.Errorf("failed to load containerd config: %w", err)
	}
	plugins := config.Plugins[criPluginUri]
	var criConfig criconfig.PluginConfig
	if err := plugins.Unmarshal(&criConfig); err != nil {
		logger.Error("unmarshal cri config error: ", err)
		return err
	}

	// found nvidia runtime, remove it
	if _, ok := criConfig.ContainerdConfig.Runtimes[nvidiaRuntime]; ok {
		delete(criConfig.ContainerdConfig.Runtimes, nvidiaRuntime)
		criConfig.DefaultRuntimeName = "runc"

		// save config
		criConfigData, err := toml.Marshal(criConfig)
		if err != nil {
			return fmt.Errorf("failed to marshal containerd cri plugin config: %w", err)
		}

		criPluginConfigTree, err := toml.LoadBytes(criConfigData)
		if err != nil {
			return fmt.Errorf("failed to load containerd cri plugin config: %w", err)
		}

		config.Plugins[criPluginUri] = *criPluginConfigTree

		// save config to file
		tmpConfigFile, err := os.OpenFile(configFile, os.O_WRONLY|os.O_CREATE|os.O_TRUNC, 0644)
		if err != nil {
			return fmt.Errorf("failed to open minikube containerd config temp file for writing: %w", err)
		}
		defer tmpConfigFile.Close()
		if err := toml.NewEncoder(tmpConfigFile).Encode(config); err != nil {
			return fmt.Errorf("failed to write minikube containerd config temp file: %w", err)
		}

	}

	return nil
}

type UninstallNvidiaDrivers struct {
	common.KubeAction
}

func (t *UninstallNvidiaDrivers) Execute(runtime connector.Runtime) error {
	_, _ = runtime.GetRunner().SudoCmd("DEBIAN_FRONTEND=noninteractive apt-get -y autoremove --purge", false, true)
	_, _ = runtime.GetRunner().SudoCmd("dpkg --configure -a || true", false, true)
	listCmd := "dpkg -l | awk '/^(ii|i[UuFHWt]|rc|..R)/ {print $2}' | grep nvidia | grep -v container"
	pkgs, _ := runtime.GetRunner().SudoCmd(listCmd, false, false)
	pkgs = strings.ReplaceAll(pkgs, "\n", " ")
	pkgs = strings.TrimSpace(pkgs)
	if pkgs != "" {
		removeCmd := fmt.Sprintf("DEBIAN_FRONTEND=noninteractive apt-get -y --auto-remove --purge remove %s", pkgs)
		if _, err := runtime.GetRunner().SudoCmd(removeCmd, false, true); err != nil {
			return errors.Wrap(errors.WithStack(err), "failed to remove nvidia packages via apt-get")
		}
		_, _ = runtime.GetRunner().SudoCmd("DEBIAN_FRONTEND=noninteractive apt-get -y autoremove --purge", false, true)
	}

	// also try to uninstall runfile-installed drivers if present
	if out, _ := runtime.GetRunner().SudoCmd("test -x /usr/bin/nvidia-uninstall && echo yes || true", false, false); strings.TrimSpace(out) == "yes" {
		if _, err := runtime.GetRunner().SudoCmd("/usr/bin/nvidia-uninstall -s", false, true); err != nil {
			return errors.Wrap(errors.WithStack(err), "failed to uninstall NVIDIA driver via nvidia-uninstall")
		}
	} else if out2, _ := runtime.GetRunner().SudoCmd("test -x /usr/bin/nvidia-installer && echo yes || true", false, false); strings.TrimSpace(out2) == "yes" {
		if _, err := runtime.GetRunner().SudoCmd("/usr/bin/nvidia-installer --uninstall -s", false, true); err != nil {
			return errors.Wrap(errors.WithStack(err), "failed to uninstall NVIDIA driver via nvidia-installer --uninstall")
		}
	}

	// clean up any leftover dkms-installed kernel modules for nvidia if present
	// only remove .ko files under updates/dkms to avoid removing other modules
	checkLeftoverCmd := "sh -c 'test -d /lib/modules/$(uname -r)/updates/dkms && find /lib/modules/$(uname -r)/updates/dkms -maxdepth 1 -type f -name \"nvidia*.ko\" -print -quit | grep -q . && echo yes || true'"
	if out, _ := runtime.GetRunner().SudoCmd(checkLeftoverCmd, false, false); strings.TrimSpace(out) == "yes" {
		if _, err := runtime.GetRunner().SudoCmd("find /lib/modules/$(uname -r)/updates/dkms -maxdepth 1 -type f -name 'nvidia*.ko' -print -delete", false, true); err != nil {
			return errors.Wrap(errors.WithStack(err), "Failed to remove leftover nvidia dkms kernel modules")
		}
		// refresh module dependency maps
		if _, err := runtime.GetRunner().SudoCmd("depmod -a $(uname -r)", false, true); err != nil {
			logger.Error("Failed to refresh module dependency maps: ", err)
		}
	}

	return nil
}

type PrintGpuStatus struct {
	common.KubeAction
}

func (t *PrintGpuStatus) Execute(runtime connector.Runtime) error {
	st, err := utils.GetNvidiaStatus(runtime)
	if err != nil {
		return err
	}
	if st == nil {
		logger.Info("no NVIDIA GPU status available")
		return nil
	}
	// basic status
	logger.Infof("Installed: %t", st.Installed)
	if st.Installed {
		logger.Infof("Install method: %s", st.InstallMethod)
	}
	logger.Infof("Running: %t", st.Running)
	// running (kernel) driver version
	if st.Running && strings.TrimSpace(st.DriverVersion) != "" {
		logger.Infof("Running driver version (kernel): %s", st.DriverVersion)
	}
	// userland info from nvidia-smi (when available)
	if st.Installed {
		if st.Info != nil && strings.TrimSpace(st.Info.DriverVersion) != "" {
			logger.Infof("Installed driver version (nvidia-smi): %s", st.Info.DriverVersion)
		}
		if strings.TrimSpace(st.CudaVersion) != "" {
			logger.Infof("CUDA version (nvidia-smi): %s", st.CudaVersion)
		}
		if st.Mismatch {
			if strings.TrimSpace(st.LibraryVersion) != "" {
				logger.Warnf("Driver/library version mismatch, NVML library version: %s", st.LibraryVersion)
			} else {
				logger.Warn("Driver/library version mismatch detected")
			}
		}
	}
	if !st.Installed && !st.Running {
		logger.Info("no NVIDIA driver detected (neither installed nor running)")
	}
	return nil
}

type PrintPluginsStatus struct {
	common.KubeAction
}

func (t *PrintPluginsStatus) Execute(runtime connector.Runtime) error {
	client, err := clientset.NewKubeClient()
	if err != nil {
		return errors.Wrap(errors.WithStack(err), "kubeclient create error")
	}

	plugins, err := client.Kubernetes().CoreV1().Pods("kube-system").List(context.Background(), metav1.ListOptions{LabelSelector: "app.kubernetes.io/component=hami-device-plugin"})
	if err != nil {
		logger.Error("get plugin status error, ", err)
		return err
	}

	if len(plugins.Items) == 0 {
		logger.Info("hami-device-plugin not exists")

	} else {
		for _, plugin := range plugins.Items {
			logger.Infof("hami-device-plugin status: %s", plugin.Status.Phase)
			break
		}
	}

	gpuScheduler, err := client.Kubernetes().CoreV1().Pods("os-gpu").List(context.Background(), metav1.ListOptions{LabelSelector: "name=gpu-scheduler"})
	if err != nil {
		logger.Error("get gpu-scheduler status error, ", err)
	}

	if len(gpuScheduler.Items) == 0 {
		logger.Info("gpu-scheduler not exists")
	} else {
		for _, scheduler := range gpuScheduler.Items {
			logger.Infof("node: %s gpu-scheduler status: %s", scheduler.Spec.NodeName, scheduler.Status.Phase)
			break
		}
	}

	return nil
}

type RestartPlugin struct {
	common.KubeAction
}

func (t *RestartPlugin) Execute(runtime connector.Runtime) error {
	kubectlpath, err := util.GetCommand(common.CommandKubectl)
	if err != nil {
		return fmt.Errorf("kubectl not found")
	}

	if _, err := runtime.GetRunner().SudoCmd(fmt.Sprintf("%s rollout restart ds gpu-scheduler -n os-gpu", kubectlpath), false, true); err != nil {
		return errors.Wrap(errors.WithStack(err), "Failed to restart gpu-scheduler")
	}

	if _, err := runtime.GetRunner().SudoCmd(fmt.Sprintf("%s rollout restart ds hami-device-plugin -n kube-system", kubectlpath), false, true); err != nil {
		return errors.Wrap(errors.WithStack(err), "Failed to restart hami-device-plugin")
	}

	if _, err := runtime.GetRunner().SudoCmd(fmt.Sprintf("%s rollout restart deploy hami-scheduler -n kube-system", kubectlpath), false, true); err != nil {
		return errors.Wrap(errors.WithStack(err), "Failed to restart hami-scheduler")
	}

	return nil
}

type WriteNouveauBlacklist struct {
	common.KubeAction
}

func (t *WriteNouveauBlacklist) Execute(runtime connector.Runtime) error {
	if !runtime.GetSystemInfo().IsLinux() {
		return nil
	}
	const dir = "/usr/lib/modprobe.d"
	const dst = "/usr/lib/modprobe.d/olares-disable-nouveau.conf"
	const content = "blacklist nouveau\nblacklist lbm-nouveau\nalias nouveau off\nalias lbm-nouveau off\n"

	if _, err := runtime.GetRunner().SudoCmd("install -d -m 0755 "+dir, false, true); err != nil {
		return errors.Wrap(errors.WithStack(err), "failed to ensure /usr/lib/modprobe.d exists")
	}

	tmpPath := path.Join(runtime.GetBaseDir(), cc.PackageCacheDir, "gpu", "olares-disable-nouveau.conf")
	if err := os.MkdirAll(path.Dir(tmpPath), 0755); err != nil {
		return errors.Wrap(errors.WithStack(err), "failed to create temp dir for nouveau blacklist")
	}
	if err := util.WriteFile(tmpPath, []byte(content), 0644); err != nil {
		return errors.Wrap(errors.WithStack(err), "failed to write temp nouveau blacklist file")
	}
	if err := runtime.GetRunner().SudoScp(tmpPath, dst); err != nil {
		return errors.Wrap(errors.WithStack(err), "failed to install nouveau blacklist file")
	}

	if _, err := runtime.GetRunner().SudoCmd("update-initramfs -u", false, false); err != nil {
		return errors.Wrap(errors.WithStack(err), "failed to update initramfs")
	}

	if out, _ := runtime.GetRunner().SudoCmd("test -d /sys/module/nouveau && echo loaded || true", false, false); strings.TrimSpace(out) == "loaded" {
		logger.Infof("the disable file for nouveau kernel module has been written, but the nouveau kernel module is currently loaded. Please REBOOT your machine to make the disabling effective.")
		os.Exit(0)
	}
	return nil
}
