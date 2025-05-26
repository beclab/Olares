/*
 Copyright 2021 The KubeSphere Authors.

 Licensed under the Apache License, Version 2.0 (the "License");
 you may not use this file except in compliance with the License.
 You may obtain a copy of the License at

     http://www.apache.org/licenses/LICENSE-2.0

 Unless required by applicable law or agreed to in writing, software
 distributed under the License is distributed on an "AS IS" BASIS,
 WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
 See the License for the specific language governing permissions and
 limitations under the License.
*/

package common

import (
	"os"

	cc "bytetrade.io/web3os/installer/pkg/core/common"
)

const (
	DefaultK8sVersion          = "v1.32.2"
	DefaultK3sVersion          = "v1.32.2-k3s"
	DefaultKubernetesVersion   = ""
	DefaultKubeSphereVersion   = "v3.3.0"
	DefaultTokenMaxAge         = 31536000
	CurrentVerifiedCudaVersion = "12.9"
)

const (
	K3s        = "k3s"
	K8e        = "k8e"
	Kubernetes = "kubernetes"

	LocalHost = "localhost"

	AllInOne    = "allInOne"
	File        = "file"
	Operator    = "operator"
	CommandLine = "commandLine"

	Master        = "master"
	Worker        = "worker"
	ETCD          = "etcd"
	K8s           = "k8s"
	Registry      = "registry"
	KubeKey       = "kubekey"
	Harbor        = "harbor"
	DockerCompose = "compose"

	KubeBinaries      = "KubeBinaries"
	WslBinaries       = "WslBinaries"
	WslUbuntuBinaries = "WslUbuntuBinaries"

	RootDir                      = "/"
	TmpDir                       = "/tmp/kubekey"
	BinDir                       = "/usr/local/bin"
	KubeConfigDir                = "/etc/kubernetes"
	KubeAddonsDir                = "/etc/kubernetes/addons"
	KubeEtcdCertDir              = "/etc/kubernetes/etcd"
	KubeCertDir                  = "/etc/kubernetes/pki"
	KubeManifestDir              = "/etc/kubernetes/manifests"
	KubeScriptDir                = "/usr/local/bin/kube-scripts"
	KubeletFlexvolumesPluginsDir = "/usr/libexec/kubernetes/kubelet-plugins/volume/exec"
	K3sImageDir                  = "/var/lib/images"
	MinikubeDefaultProfile       = "olares-0"
	MinikubeEtcdCertDir          = "/var/lib/minikube/certs/etcd"
	WSLDefaultDistribution       = "Ubuntu"
	RunLockDir                   = "/var/run/lock"

	InstallerScriptsDir = "scripts"

	ETCDCertDir     = "/etc/ssl/etcd/ssl"
	RegistryCertDir = "/etc/ssl/registry/ssl"

	HaproxyDir = "/etc/kubekey/haproxy"

	IPv4Regexp = "[\\d]+\\.[\\d]+\\.[\\d]+\\.[\\d]+"
	IPv6Regexp = "[a-f0-9]{1,4}(:[a-f0-9]{1,4}){7}|[a-f0-9]{1,4}(:[a-f0-9]{1,4}){0,7}::[a-f0-9]{0,4}(:[a-f0-9]{1,4}){0,7}"

	Calico  = "calico"
	Flannel = "flannel"
	Cilium  = "cilium"
	Kubeovn = "kubeovn"

	Docker     = "docker"
	Crictl     = "crictl"
	Containerd = "containerd"
	Crio       = "crio"
	Isula      = "isula"
	Runc       = "runc"

	// global cache key
	// PreCheckModule
	NodePreCheck           = "nodePreCheck"
	K8sVersion             = "k8sVersion"        // current k8s version
	MaxK8sVersion          = "maxK8sVersion"     // max k8s version of nodes
	KubeSphereVersion      = "kubeSphereVersion" // current KubeSphere version
	ClusterNodeStatus      = "clusterNodeStatus"
	ClusterNodeCRIRuntimes = "ClusterNodeCRIRuntimes"
	DesiredK8sVersion      = "desiredK8sVersion"
	PlanK8sVersion         = "planK8sVersion"
	NodeK8sVersion         = "NodeK8sVersion"

	// ETCDModule
	ETCDCluster = "etcdCluster"
	ETCDName    = "etcdName"
	ETCDExist   = "etcdExist"

	// KubernetesModule
	ClusterStatus = "clusterStatus"
	ClusterExist  = "clusterExist"

	MasterInfo = "masterInfo"

	// CertsModule
	Certificate   = "certificate"
	CaCertificate = "caCertificate"

	// Artifact pipeline
	Artifact = "artifact"

	SkipMasterNodePullImages = "skipMasterNodePullImages"
)

const (
	Linux   = "linux"
	Darwin  = "darwin"
	Windows = "windows"

	Intel64 = "x86_64"
	Amd64   = "amd64"
	Arm     = "arm"
	Arm7    = "arm7"
	Armv7l  = "Armv7l"
	Armhf   = "armhf"
	Arm64   = "arm64"
	PPC64el = "ppc64el"
	PPC64le = "ppc64le"
	S390x   = "s390x"
	Riscv64 = "riscv64"

	Ubuntu   = "ubuntu"
	Debian   = "debian"
	CentOs   = "centos"
	Fedora   = "fedora"
	RHEl     = "rhel"
	Raspbian = "raspbian"
	PVE      = "pve"
	WSL      = "wsl"
)

const (
	TRUE  = "true"
	FALSE = "false"

	YES = "yes"
	NO  = "no"
)

const (
	OSS   = "oss"
	COS   = "cos"
	S3    = "s3"
	MinIO = "minio"

	//ManagedMinIO is MinIO instance that's managed by us
	ManagedMinIO = "managed-minio"
)

var (
	CloudVendor = os.Getenv("CLOUD_VENDOR")
	ResolvProxy = os.Getenv("PROXY")
)

const (
	OlaresRegistryMirrorHost       = "mirrors.joinolares.cn"
	OlaresRegistryMirrorHostLegacy = "mirrors.jointerminus.cn"
)

const (
	CloudVendorAliYun = "aliyun"
	CloudVendorAWS    = "aws"
)

const (
	RaspbianCmdlineFile  = "/boot/cmdline.txt"
	RaspbianFirmwareFile = "/boot/firmware/cmdline.txt"
)

const (
	ManifestImageList          = "images.mf"
	TerminusStateFilePrepared  = ".prepared"
	TerminusStateFileInstalled = ".installed"
	MasterHostConfigFile       = "master.conf"
	OlaresReleaseFile          = "/etc/olares/release"
)

const (
	CommandIpset        = "ipset"
	CommandIptables     = "iptables"
	CommandIp6tables    = "ip6tables"
	CommandGPG          = "gpg"
	CommandSudo         = "sudo"
	CommandSocat        = "socat"
	CommandConntrack    = "conntrack"
	CommandNtpdate      = "ntpdate"
	CommandHwclock      = "hwclock"
	CommandKubectl      = "kubectl"
	CommandDocker       = "docker"
	CommandMinikube     = "minikube"
	CommandUnzip        = "unzip"
	CommandVelero       = "velero"
	CommandUpdatePciids = "update-pciids"
	CommandNmcli        = "nmcli"
	CommandZRAMCtl      = "zramctl"

	CacheCommandKubectlPath  = "kubectl_bin_path"
	CacheCommandMinikubePath = "minikube_bin_path"
	CacheCommandDockerPath   = "docker_bin_path"
)

const (
	CacheKubeletVersion = "version_kubelet"

	CacheKubectlKey = "cmd_kubectl"

	CacheStorageVendor = "storage_vendor"
	CacheProxy         = "proxy"

	CacheEnableHA      = "enable_ha"
	CacheMasterNum     = "master_num"
	CacheNodeNum       = "node_num"
	CacheRedisPassword = "redis_password"
	CacheSecretsNum    = "secrets_num"
	CacheJwtSecret     = "jwt_secret"
	CacheCrdsNUm       = "users_iam_num"

	CacheMinioPath     = "minio_binary_path"
	CacheMinioDataPath = "minio_data_path"
	CacheMinioPassword = "minio_password"

	CacheMinioOperatorPath = "minio_operator_path"

	CacheHostRedisPassword = "hostredis_password"
	CacheHostRedisAddress  = "hostredis_address"
	CachePreparedState     = "prepare_state"
	CacheInstalledState    = "install_state"

	CacheJuiceFsPath     = "juicefs_binary_path"
	CacheJuiceFsFileName = "juicefs_binary_filename"

	CacheMinikubeNodeIp                  = "minikube_node_ip"
	CacheMinikubeTmpContainerdConfigFile = "minikube_tmp_containerd_config_file"

	CacheAccessKey = "storage_access_key"
	CacheSecretKey = "storage_secret_key"
	CacheToken     = "storage_token"
	CacheClusterId = "storage_cluster_id"

	CacheAppServicePod = "app_service_pod_name"
	CacheAppValues     = "app_built_in_values"

	CacheCountPodsUsingHostIP = "count_pods_using_host_ip"

	CacheUpgradeUsers     = "upgrade_users"
	CacheUpgradeAdminUser = "upgrade_admin_user"

	CacheWindowsDistroStoreLocation     = "windows_distro_store_location"
	CacheWindowsDistroStoreLocationNums = "windows_distro_store_location_nums"
)

const (
	CacheLaunchAppKey    = "launch_app_key"
	CacheLaunchAppSecret = "launch_app_secret"
)

const (
	ENV_OLARES_BASE_DIR              = "OLARES_BASE_DIR"
	ENV_OLARES_VERSION               = "OLARES_VERSION"
	ENV_TERMINUS_IS_CLOUD_VERSION    = "TERMINUS_IS_CLOUD_VERSION"
	ENV_PUBLICLY_ACCESSIBLE          = "PUBLICLY_ACCESSIBLE"
	ENV_KUBE_TYPE                    = "KUBE_TYPE"
	ENV_REGISTRY_MIRRORS             = "REGISTRY_MIRRORS"
	ENV_NVIDIA_CONTAINER_REPO_MIRROR = "NVIDIA_CONTAINER_REPO_MIRROR"
	ENV_DOWNLOAD_CDN_URL             = "DOWNLOAD_CDN_URL"
	ENV_STORAGE                      = "STORAGE"
	ENV_S3_BUCKET                    = "S3_BUCKET"
	ENV_LOCAL_GPU_ENABLE             = "LOCAL_GPU_ENABLE"
	// ENV_LOCAL_GPU_SHARE             = "LOCAL_GPU_SHARE"
	ENV_CLOUDFLARE_ENABLE           = "CLOUDFLARE_ENABLE"
	ENV_FRP_ENABLE                  = "FRP_ENABLE"
	ENV_FRP_SERVER                  = "FRP_SERVER"
	ENV_FRP_PORT                    = "FRP_PORT"
	ENV_FRP_AUTH_METHOD             = "FRP_AUTH_METHOD"
	ENV_FRP_AUTH_TOKEN              = "FRP_AUTH_TOKEN"
	ENV_AWS_ACCESS_KEY_ID_SETUP     = "AWS_ACCESS_KEY_ID_SETUP"
	ENV_AWS_SECRET_ACCESS_KEY_SETUP = "AWS_SECRET_ACCESS_KEY_SETUP"
	ENV_AWS_SESSION_TOKEN_SETUP     = "AWS_SESSION_TOKEN_SETUP"
	ENV_BACKUP_KEY_PREFIX           = "BACKUP_KEY_PREFIX"
	ENV_BACKUP_SECRET               = "BACKUP_SECRET"
	ENV_CLUSTER_ID                  = "CLUSTER_ID"
	ENV_BACKUP_CLUSTER_BUCKET       = "BACKUP_CLUSTER_BUCKET"
	ENV_TOKEN_MAX_AGE               = "TOKEN_MAX_AGE"
	ENV_MARKET_PROVIDER             = "MARKET_PROVIDER"
	ENV_TERMINUS_CERT_SERVICE_API   = "TERMINUS_CERT_SERVICE_API"
	ENV_TERMINUS_DNS_SERVICE_API    = "TERMINUS_DNS_SERVICE_API"
	ENV_HOST_IP                     = "HOST_IP"
	ENV_PREINSTALL                  = "PREINSTALL"
	ENV_DISABLE_HOST_IP_PROMPT      = "DISABLE_HOST_IP_PROMPT"
	ENV_AUTO_ADD_FIREWALL_RULES     = "AUTO_ADD_FIREWALL_RULES"
	ENV_TERMINUS_OS_DOMAINNAME      = "TERMINUS_OS_DOMAINNAME"
	ENV_DEFAULT_WSL_DISTRO_LOCATION = "DEFAULT_WSL_DISTRO_LOCATION" // If set to 1, the default WSL distro storage will be used.

	ENV_CONTAINER      = "container"
	ENV_CONTAINER_MODE = "CONTAINER_MODE" // running in docker container
)

// TerminusGlobalEnvs holds a group of general environment variables
// which are used for many components
// along with their default values
// if they are set in the execution environment
// the default values are override
// note that we declare the key type as interface{} on purpose
// to avoid Helm bug when merging values
var TerminusGlobalEnvs = map[string]interface{}{
	"DID_GATE_URL":               "https://did-gate-v3.bttcdn.com/",
	"OLARES_SPACE_URL":           "https://cloud-api.bttcdn.com/",
	"FIREBASE_PUSH_URL":          "https://firebase-push-test.bttcdn.com/v1/api/push",
	"FRP_LIST_URL":               "https://terminus-frp.snowinning.com",
	"TAILSCALE_CONTROLPLANE_URL": "https://controlplane.snowinning.com",
	"OLARES_ROOT_DIR":            "/olares",
	ENV_DOWNLOAD_CDN_URL:         cc.DownloadUrl,
	ENV_MARKET_PROVIDER:          "appstore-server-prod.bttcdn.com",
}

const (
	HelmValuesKeyTerminusGlobalEnvs = "terminusGlobalEnvs"
	HelmValuesKeyOlaresRootFSPath   = "rootPath"
)

func init() {
	for envKey, _ := range TerminusGlobalEnvs {
		if val := os.Getenv(envKey); val != "" {
			TerminusGlobalEnvs[envKey] = val
		}
	}
}
