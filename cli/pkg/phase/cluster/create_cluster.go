package cluster

import (
	cc "bytetrade.io/web3os/installer/pkg/core/common"
	"path"

	kubekeyapiv1alpha2 "bytetrade.io/web3os/installer/apis/kubekey/v1alpha2"

	"bytetrade.io/web3os/installer/pkg/addons"
	"bytetrade.io/web3os/installer/pkg/bootstrap/confirm"
	"bytetrade.io/web3os/installer/pkg/bootstrap/os"
	"bytetrade.io/web3os/installer/pkg/bootstrap/precheck"
	"bytetrade.io/web3os/installer/pkg/certs"
	"bytetrade.io/web3os/installer/pkg/common"
	"bytetrade.io/web3os/installer/pkg/core/module"
	"bytetrade.io/web3os/installer/pkg/etcd"
	"bytetrade.io/web3os/installer/pkg/filesystem"
	"bytetrade.io/web3os/installer/pkg/images"
	"bytetrade.io/web3os/installer/pkg/k3s"
	"bytetrade.io/web3os/installer/pkg/kubernetes"
	"bytetrade.io/web3os/installer/pkg/kubesphere"
	ksplugins "bytetrade.io/web3os/installer/pkg/kubesphere/plugins"
	"bytetrade.io/web3os/installer/pkg/manifest"
	"bytetrade.io/web3os/installer/pkg/plugins"
	"bytetrade.io/web3os/installer/pkg/plugins/dns"
	"bytetrade.io/web3os/installer/pkg/plugins/network"
	"bytetrade.io/web3os/installer/pkg/plugins/storage"
)

func NewDarwinClusterPhase(runtime *common.KubeRuntime, manifestMap manifest.InstallationManifest) []module.Module {
	m := []module.Module{
		&kubesphere.CheckMacOsCommandModule{},
		&images.PreloadImagesModule{
			ManifestModule: manifest.ManifestModule{
				Manifest: manifestMap,
				BaseDir:  runtime.GetBaseDir(),
			},
		},
		&kubesphere.DeployMiniKubeModule{},
		&kubesphere.DeployModule{Skip: !runtime.Cluster.KubeSphere.Enabled},
		&ksplugins.DeployKsPluginsModule{},
		//&ksplugins.DeploySnapshotControllerModule{},
		&ksplugins.DeployRedisModule{},
		&ksplugins.CreateKubeSphereSecretModule{},
		&ksplugins.DeployKsCoreConfigModule{}, // ks-core-config
		&ksplugins.CreateMonitorDashboardModule{},
		//&ksplugins.CreateNotificationModule{},
		&ksplugins.DeployPrometheusModule{},
		&ksplugins.DeployKsCoreModule{},
		&kubesphere.CheckResultModule{Skip: !runtime.Cluster.KubeSphere.Enabled},
	}

	return m
}

func NewK3sCreateClusterPhase(runtime *common.KubeRuntime, manifestMap manifest.InstallationManifest) []module.Module {
	systemInfo := runtime.GetSystemInfo()
	baseDir := runtime.GetBaseDir()
	if systemInfo.IsWsl() {
		baseDir = path.Join(runtime.Arg.GetWslUserPath(), cc.DefaultBaseDir)
	}

	skipLocalStorage := true
	if runtime.Arg.DeployLocalStorage != nil {
		skipLocalStorage = !*runtime.Arg.DeployLocalStorage
	} else if runtime.Cluster.KubeSphere.Enabled {
		skipLocalStorage = false
	}

	m := []module.Module{
		&k3s.StatusModule{},
		&os.ConfigureOSModule{},
		&etcd.PreCheckModule{Skip: runtime.Cluster.Etcd.Type != kubekeyapiv1alpha2.KubeKey},
		&etcd.CertsModule{},
		&etcd.InstallETCDBinaryModule{
			ManifestModule: manifest.ManifestModule{
				BaseDir:  baseDir,
				Manifest: manifestMap,
			},
			Skip: runtime.Cluster.Etcd.Type != kubekeyapiv1alpha2.KubeKey},
		&etcd.ConfigureModule{Skip: runtime.Cluster.Etcd.Type != kubekeyapiv1alpha2.KubeKey},
		&etcd.BackupModule{Skip: runtime.Cluster.Etcd.Type != kubekeyapiv1alpha2.KubeKey},
		&k3s.InstallKubeBinariesModule{
			ManifestModule: manifest.ManifestModule{
				BaseDir:  baseDir,
				Manifest: manifestMap,
			},
		},
		&k3s.InitClusterModule{},
		&dns.ClusterDNSModule{},
		&k3s.StatusModule{},
		&k3s.JoinNodesModule{},
		&network.DeployNetworkPluginModule{},
		&kubernetes.ConfigureKubernetesModule{},
		&filesystem.ChownModule{},
		&certs.AutoRenewCertsModule{Skip: !runtime.Cluster.Kubernetes.EnableAutoRenewCerts()},
		&k3s.SaveKubeConfigModule{},
		&addons.AddonsModule{}, // relative ks-installer
		&storage.DeployLocalVolumeModule{Skip: skipLocalStorage},
		&kubesphere.DeployModule{Skip: !runtime.Cluster.KubeSphere.Enabled}, //
		&ksplugins.DeployKsPluginsModule{},
		//&ksplugins.DeploySnapshotControllerModule{},
		&ksplugins.DeployRedisModule{},
		&ksplugins.CreateKubeSphereSecretModule{},
		&ksplugins.DeployKsCoreConfigModule{}, // ks-core-config
		&ksplugins.CreateMonitorDashboardModule{},
		//&ksplugins.CreateNotificationModule{},
		&ksplugins.DeployPrometheusModule{},
		&ksplugins.DeployKsCoreModule{},
		&kubesphere.CheckResultModule{Skip: !runtime.Cluster.KubeSphere.Enabled},
	}

	return m
}

func NewCreateClusterPhase(runtime *common.KubeRuntime, manifestMap manifest.InstallationManifest) []module.Module {
	systemInfo := runtime.GetSystemInfo()
	baseDir := runtime.GetBaseDir()
	if systemInfo.IsWsl() {
		baseDir = path.Join(runtime.Arg.GetWslUserPath(), cc.DefaultBaseDir)
	}

	skipLocalStorage := true
	if runtime.Arg.DeployLocalStorage != nil {
		skipLocalStorage = !*runtime.Arg.DeployLocalStorage
	} else if runtime.Cluster.KubeSphere.Enabled {
		skipLocalStorage = false
	}

	m := []module.Module{
		&precheck.NodePreCheckModule{},
		&confirm.InstallConfirmModule{Skip: runtime.Arg.SkipConfirmCheck},
		&kubernetes.StatusModule{},
		&os.ConfigureOSModule{},
		&etcd.PreCheckModule{Skip: runtime.Cluster.Etcd.Type != kubekeyapiv1alpha2.KubeKey},
		&etcd.CertsModule{},
		&etcd.InstallETCDBinaryModule{
			ManifestModule: manifest.ManifestModule{
				BaseDir:  baseDir,
				Manifest: manifestMap,
			},
			Skip: runtime.Cluster.Etcd.Type != kubekeyapiv1alpha2.KubeKey,
		},
		&etcd.ConfigureModule{Skip: runtime.Cluster.Etcd.Type != kubekeyapiv1alpha2.KubeKey},
		&etcd.BackupModule{Skip: runtime.Cluster.Etcd.Type != kubekeyapiv1alpha2.KubeKey},
		&kubernetes.InstallKubeBinariesModule{
			ManifestModule: manifest.ManifestModule{
				BaseDir:  baseDir,
				Manifest: manifestMap,
			},
		},
		&kubernetes.InitKubernetesModule{},
		&dns.ClusterDNSModule{},
		&kubernetes.StatusModule{},
		&kubernetes.JoinNodesModule{},
		&network.DeployNetworkPluginModule{},
		&kubernetes.ConfigureKubernetesModule{},
		&filesystem.ChownModule{},
		&certs.AutoRenewCertsModule{Skip: !runtime.Cluster.Kubernetes.EnableAutoRenewCerts()},
		&kubernetes.SecurityEnhancementModule{Skip: !runtime.Arg.SecurityEnhancement},
		&kubernetes.SaveKubeConfigModule{},
		&plugins.DeployPluginsModule{},
		&addons.AddonsModule{},
		&storage.DeployLocalVolumeModule{Skip: skipLocalStorage},
		&kubesphere.DeployModule{Skip: !runtime.Cluster.KubeSphere.Enabled},
		&ksplugins.DeployKsPluginsModule{},
		//&ksplugins.DeploySnapshotControllerModule{},
		&ksplugins.DeployRedisModule{},
		&ksplugins.CreateKubeSphereSecretModule{},
		&ksplugins.DeployKsCoreConfigModule{}, // ! ks-core-config
		&ksplugins.CreateMonitorDashboardModule{},
		//&ksplugins.CreateNotificationModule{},
		&ksplugins.DeployPrometheusModule{},
		&ksplugins.DeployKsCoreModule{},
		&kubesphere.CheckResultModule{Skip: !runtime.Cluster.KubeSphere.Enabled}, // check ks-apiserver phase
	}

	return m
}
