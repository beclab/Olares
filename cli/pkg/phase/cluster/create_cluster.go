package cluster

import (
	"path"

	cc "github.com/beclab/Olares/cli/pkg/core/common"

	kubekeyapiv1alpha2 "github.com/beclab/Olares/cli/apis/kubekey/v1alpha2"

	"github.com/beclab/Olares/cli/pkg/bootstrap/confirm"
	"github.com/beclab/Olares/cli/pkg/bootstrap/os"
	"github.com/beclab/Olares/cli/pkg/bootstrap/precheck"
	"github.com/beclab/Olares/cli/pkg/certs"
	"github.com/beclab/Olares/cli/pkg/common"
	"github.com/beclab/Olares/cli/pkg/core/module"
	"github.com/beclab/Olares/cli/pkg/etcd"
	"github.com/beclab/Olares/cli/pkg/filesystem"
	"github.com/beclab/Olares/cli/pkg/images"
	"github.com/beclab/Olares/cli/pkg/k3s"
	"github.com/beclab/Olares/cli/pkg/kubernetes"
	"github.com/beclab/Olares/cli/pkg/kubesphere"
	ksplugins "github.com/beclab/Olares/cli/pkg/kubesphere/plugins"
	"github.com/beclab/Olares/cli/pkg/manifest"
	"github.com/beclab/Olares/cli/pkg/plugins/dns"
	"github.com/beclab/Olares/cli/pkg/plugins/network"
	"github.com/beclab/Olares/cli/pkg/plugins/storage"
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
		//&ksplugins.DeployRedisModule{},
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
		&storage.DeployLocalVolumeModule{Skip: skipLocalStorage},
		&kubesphere.DeployModule{Skip: !runtime.Cluster.KubeSphere.Enabled}, //
		&ksplugins.DeployKsPluginsModule{},
		//&ksplugins.DeployRedisModule{},
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
		&storage.DeployLocalVolumeModule{Skip: skipLocalStorage},
		&kubesphere.DeployModule{Skip: !runtime.Cluster.KubeSphere.Enabled},
		&ksplugins.DeployKsPluginsModule{},
		//&ksplugins.DeployRedisModule{},
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
