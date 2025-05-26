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
	"encoding/json"
	"fmt"
	"github.com/joho/godotenv"
	"net"
	"os"
	"path/filepath"
	"strconv"
	"strings"

	"github.com/pkg/errors"
	"github.com/spf13/pflag"

	kubekeyapiv1alpha2 "bytetrade.io/web3os/installer/apis/kubekey/v1alpha2"
	kubekeyclientset "bytetrade.io/web3os/installer/clients/clientset/versioned"
	"bytetrade.io/web3os/installer/pkg/core/common"
	"bytetrade.io/web3os/installer/pkg/core/connector"
	"bytetrade.io/web3os/installer/pkg/core/logger"
	"bytetrade.io/web3os/installer/pkg/core/storage"
	"bytetrade.io/web3os/installer/pkg/core/util"
	kresource "k8s.io/apimachinery/pkg/api/resource"
)

type KubeRuntime struct {
	connector.BaseRuntime
	ClusterName string
	Cluster     *kubekeyapiv1alpha2.ClusterSpec
	Kubeconfig  string
	ClientSet   *kubekeyclientset.Clientset
	Arg         *Argument
}

type Argument struct {
	NodeName            string `json:"node_name"`
	FilePath            string `json:"file_path"`
	KubernetesVersion   string `json:"kubernetes_version"`
	KsEnable            bool   `json:"ks_enable"`
	KsVersion           string `json:"ks_version"`
	OlaresVersion       string `json:"olares_version"`
	Debug               bool   `json:"debug"`
	IgnoreErr           bool   `json:"ignore_err"`
	SkipPullImages      bool   `json:"skip_pull_images"`
	SKipPushImages      bool   `json:"skip_push_images"`
	SecurityEnhancement bool   `json:"security_enhancement"`
	DeployLocalStorage  *bool  `json:"deploy_local_storage"`
	// DownloadCommand     func(path, url string) string
	SkipConfirmCheck bool   `json:"skip_confirm_check"`
	InCluster        bool   `json:"in_cluster"`
	ContainerManager string `json:"container_manager"`
	FromCluster      bool   `json:"from_cluster"`
	KubeConfig       string `json:"kube_config"`
	Artifact         string `json:"artifact"`
	InstallPackages  bool   `json:"install_packages"`
	ImagesDir        string `json:"images_dir"`
	Namespace        string `json:"namespace"`
	DeleteCRI        bool   `json:"delete_cri"`
	DeleteCache      bool   `json:"delete_cache"`
	Role             string `json:"role"`
	Type             string `json:"type"`
	Kubetype         string `json:"kube_type"`
	SystemInfo       connector.Systems

	// Extra args
	ExtraAddon      string `json:"extra_addon"` // addon yaml config
	RegistryMirrors string `json:"registry_mirrors"`
	DownloadCdnUrl  string `json:"download_cdn_url"`

	// Swap config
	*SwapConfig

	// master node ssh config
	*MasterHostConfig

	LocalSSHPort int `json:"-"`

	SkipMasterPullImages bool `json:"skip_master_pull_images"`

	// db
	Provider storage.Provider `json:"-"`
	// User
	User *User `json:"user"`
	// if juicefs is opted off, the local storage is used directly
	// only used in prepare phase
	// the existence of juicefs should be checked in other phases
	// to avoid wrong information given by user
	WithJuiceFS bool `json:"with_juicefs"`
	// the object storage service used as backend for JuiceFS
	Storage                *Storage           `json:"storage"`
	PublicNetworkInfo      *PublicNetworkInfo `json:"public_network_info"`
	GPU                    *GPU               `json:"gpu"`
	Cloudflare             *Cloudflare        `json:"cloudflare"`
	Frp                    *Frp               `json:"frp"`
	TokenMaxAge            int64              `json:"token_max_age"` // nanosecond
	MarketProvider         string             `json:"market_provider"`
	TerminusCertServiceAPI string             `json:"terminus_cert_service_api"`
	TerminusDNSServiceAPI  string             `json:"terminus_dns_service_api"`

	Request any `json:"-"`

	IsCloudInstance    bool     `json:"is_cloud_instance"`
	MinikubeProfile    string   `json:"minikube_profile"`
	WSLDistribution    string   `json:"wsl_distribution"`
	Environment        []string `json:"environment"`
	BaseDir            string   `json:"base_dir"`
	Manifest           string   `json:"manifest"`
	ConsoleLogFileName string   `json:"console_log_file_name"`
	ConsoleLogTruncate bool     `json:"console_log_truncate"`
	HostIP             string   `json:"host_ip"`

	CudaVersion string `json:"cuda_version"`

	IsOlaresInContainer bool `json:"is_olares_in_container"`
}

type SwapConfig struct {
	EnablePodSwap    bool   `json:"enable_pod_swap"`
	Swappiness       int    `json:"swappiness"`
	EnableZRAM       bool   `json:"enable_zram"`
	ZRAMSize         string `json:"zram_size"`
	ZRAMSwapPriority int    `json:"zram_swap_priority"`
}

func (cfg *SwapConfig) AddFlags(fs *pflag.FlagSet) {
	fs.BoolVar(&cfg.EnablePodSwap, "enable-pod-swap", false, "Enable pods on Kubernetes cluster to use swap, setting --enable-zram, --zram-size or --zram-swap-priority implicitly enables this option, regardless of the command line args, note that only pods of the BestEffort QOS group can use swap due to K8s design")
	fs.IntVar(&cfg.Swappiness, "swappiness", 0, "Configure the Linux swappiness value, if not set, the current configuration is remained")
	fs.BoolVar(&cfg.EnableZRAM, "enable-zram", false, "Set up a ZRAM device to be used for swap, setting --zram-size or --zram-swap-priority implicitly enables this option, regardless of the command line args")
	fs.StringVar(&cfg.ZRAMSize, "zram-size", "", "Set the size of the ZRAM device, takes a format of https://kubernetes.io/docs/reference/kubernetes-api/common-definitions/quantity, defaults to half of the total RAM")
	fs.IntVar(&cfg.ZRAMSwapPriority, "zram-swap-priority", 0, "Set the swap priority of the ZRAM device, between -1 and 32767, defaults to 100")
}

func (cfg *SwapConfig) Validate() error {
	if cfg.ZRAMSize == "" {
		return nil
	}
	processedZRAMSize := cfg.ZRAMSize
	if strings.HasSuffix(processedZRAMSize, "b") || strings.HasSuffix(processedZRAMSize, "B") {
		processedZRAMSize = strings.TrimSuffix(cfg.ZRAMSize, "b")
		processedZRAMSize = strings.TrimSuffix(cfg.ZRAMSize, "B")
	}
	processedZRAMSize = strings.ReplaceAll(processedZRAMSize, "g", "G")
	processedZRAMSize = strings.ReplaceAll(processedZRAMSize, "k", "K")
	processedZRAMSize = strings.ReplaceAll(processedZRAMSize, "m", "M")
	q, err := kresource.ParseQuantity(processedZRAMSize)
	if err != nil {
		return fmt.Errorf("invalid zram size %s: %w", cfg.ZRAMSize, err)
	}
	cfg.ZRAMSize = q.String() + "B"
	return nil
}

type MasterHostConfig struct {
	MasterHost              string `json:"master_host"`
	MasterNodeName          string `json:"master_node_name"`
	MasterSSHUser           string `json:"master_ssh_user"`
	MasterSSHPassword       string `json:"master_ssh_password"`
	MasterSSHPrivateKeyPath string `json:"master_ssh_private_key_path"`
	MasterSSHPort           int    `json:"master_ssh_port"`
}

func (cfg *MasterHostConfig) AddFlags(fs *pflag.FlagSet) {
	fs.StringVar(&cfg.MasterHost, "master-host", "", "IP address of the master node")
	fs.StringVar(&cfg.MasterNodeName, "master-node-name", "", "Name of the master node")
	fs.StringVar(&cfg.MasterSSHUser, "master-ssh-user", "", "Username of the master node, defaults to root")
	fs.StringVar(&cfg.MasterSSHPassword, "master-ssh-password", "", "Password of the master node")
	fs.StringVar(&cfg.MasterSSHPrivateKeyPath, "master-ssh-private-key-path", "", "Path to the SSH key to access the master node, defaults to ~/.ssh/id_rsa")
	fs.IntVar(&cfg.MasterSSHPort, "master-ssh-port", 0, "SSH Port of the master node, defaults to 22")
}

func (cfg *MasterHostConfig) Validate() error {
	if cfg.MasterHost == "" {
		return errors.New("--master-host is not provided")
	}
	if cfg.MasterSSHUser != "" && cfg.MasterSSHUser != "root" && cfg.MasterSSHPassword == "" {
		return errors.New("--master-ssh-password must be provided for non-root user in order to execute sudo command")
	}
	return nil
}

type PublicNetworkInfo struct {
	// OSPublicIPs contains a list of public ip(s)
	// by looking at local network interfaces
	// if any
	OSPublicIPs []net.IP `json:"os_public_ips"`

	// AWS contains the info retrieved from the AWS instance metadata service
	// if any
	AWSPublicIP net.IP `json:"aws"`

	// ExternalPublicIP is the IP address seen by others on the internet
	// it may not be an IP address
	// that's directly bound to a local network interface, e.g. on an AWS EC2 instance
	// or may not be an IP address
	// that can be used to access the machine at all, e.g. a machine behind multiple NAT gateways
	// this is used as a fallback method to determine the machine's public IP address
	// if none can be found from the OS or AWS IMDS service
	// but the user explicitly specifies that the machine is publicly accessible
	ExternalPublicIP net.IP `json:"external_public_ip"`

	PubliclyAccessible bool `json:"publicly_accessible"`
}

type User struct {
	UserName          string `json:"user_name"`
	Password          string `json:"user_password"`
	EncryptedPassword string `json:"-"`
	Email             string `json:"user_email"`
	DomainName        string `json:"user_domain_name"`
}

type Storage struct {
	StorageVendor    string `json:"storage_vendor"`
	StorageType      string `json:"storage_type"`
	StorageBucket    string `json:"storage_bucket"`
	StoragePrefix    string `json:"storage_prefix"`
	StorageAccessKey string `json:"storage_access_key"`
	StorageSecretKey string `json:"storage_secret_key"`

	StorageToken        string `json:"storage_token"`       // juicefs  --> from env
	StorageClusterId    string `json:"storage_cluster_id"`  // use only on the Terminus cloud, juicefs  --> from env
	StorageSyncSecret   string `json:"storage_sync_secret"` // use only on the Terminus cloud  --> from env
	BackupClusterBucket string `json:"backup_cluster_bucket"`
}

type GPU struct {
	Enable bool `json:"gpu_enable"`
	Share  bool `json:"gpu_share"`
}

type Cloudflare struct {
	Enable string `json:"cloudflare_enable"`
}

type Frp struct {
	Enable     string `json:"frp_enable"`
	Server     string `json:"frp_server"`
	Port       string `json:"frp_port"`
	AuthMethod string `json:"frp_auth_method"`
	AuthToken  string `json:"frp_auth_token"`
}

func NewArgument() *Argument {
	arg := &Argument{
		KsEnable:         true,
		KsVersion:        DefaultKubeSphereVersion,
		InstallPackages:  false,
		SKipPushImages:   false,
		ContainerManager: Containerd,
		SystemInfo:       connector.GetSystemInfo(),
		Storage: &Storage{
			StorageType: ManagedMinIO,
		},
		GPU: &GPU{
			Enable: !strings.EqualFold(os.Getenv(ENV_LOCAL_GPU_ENABLE), "0"), // default enable GPU, not set or 1 means enable
			Share:  !strings.EqualFold(os.Getenv(ENV_LOCAL_GPU_ENABLE), "0"), // default share GPU
		},
		Cloudflare:             &Cloudflare{},
		Frp:                    &Frp{},
		User:                   &User{},
		PublicNetworkInfo:      &PublicNetworkInfo{},
		RegistryMirrors:        os.Getenv(ENV_REGISTRY_MIRRORS),
		DownloadCdnUrl:         os.Getenv(ENV_DOWNLOAD_CDN_URL),
		MarketProvider:         os.Getenv(ENV_MARKET_PROVIDER),
		TerminusCertServiceAPI: os.Getenv(ENV_TERMINUS_CERT_SERVICE_API),
		TerminusDNSServiceAPI:  os.Getenv(ENV_TERMINUS_DNS_SERVICE_API),
		HostIP:                 os.Getenv(ENV_HOST_IP),
		Environment:            os.Environ(),
		MasterHostConfig:       &MasterHostConfig{},
		SwapConfig:             &SwapConfig{},
	}
	arg.IsCloudInstance, _ = strconv.ParseBool(os.Getenv(ENV_TERMINUS_IS_CLOUD_VERSION))
	arg.PublicNetworkInfo.PubliclyAccessible, _ = strconv.ParseBool(os.Getenv(ENV_PUBLICLY_ACCESSIBLE))
	arg.IsOlaresInContainer = os.Getenv("CONTAINER_MODE") == "oic"

	if err := arg.LoadReleaseInfo(); err != nil {
		fmt.Printf("error loading release info: %v", err)
		os.Exit(1)
	}
	return arg
}

// LoadReleaseInfo loads base directory and version settings
// from /etc/olares/release and environment variables,
// with the latter takes precedence.
// Note that the command line options --base-dir and --version
// still have the highest priority and will override any values loaded here
func (a *Argument) LoadReleaseInfo() error {
	// load envs from the release file
	// already existing envs are not overridden so
	err := godotenv.Load(OlaresReleaseFile)

	// silently ignore the non-existence of a release file
	// otherwise, return the error
	if err != nil && !os.IsNotExist(err) {
		return err
	}
	a.BaseDir = os.Getenv(ENV_OLARES_BASE_DIR)
	a.OlaresVersion = os.Getenv(ENV_OLARES_VERSION)
	return nil
}

func (a *Argument) SaveReleaseInfo() error {
	if a.BaseDir == "" {
		return errors.New("invalid: empty base directory")
	}
	if a.OlaresVersion == "" {
		return errors.New("invalid: empty olares version")
	}
	releaseInfoMap := map[string]string{
		ENV_OLARES_BASE_DIR: a.BaseDir,
		ENV_OLARES_VERSION:  a.OlaresVersion,
	}
	if !util.IsExist(filepath.Dir(OlaresReleaseFile)) {
		if err := os.MkdirAll(filepath.Dir(OlaresReleaseFile), 0755); err != nil {
			return fmt.Errorf("failed to create directory %s: %v", filepath.Dir(OlaresReleaseFile), err)
		}
	}
	return godotenv.Write(releaseInfoMap, OlaresReleaseFile)
}

func (a *Argument) GetWslUserPath() string {
	if a.Environment == nil || len(a.Environment) == 0 {
		return ""
	}

	var res string
	var wslSuffix = "/AppData/Local/Microsoft/WindowsApps"
	for _, v := range a.Environment {
		if strings.HasPrefix(v, "PATH=") {
			p := strings.ReplaceAll(v, "PATH=", "")
			s := strings.Split(p, ":")
			for _, s1 := range s {
				if strings.Contains(s1, wslSuffix) {
					res = strings.ReplaceAll(s1, wslSuffix, "")
					break
				}
			}
		}
	}
	return res
}

func (a *Argument) SetDownloadCdnUrl(downloadCdnUrl string) {
	u := strings.TrimSuffix(downloadCdnUrl, "/")
	if u == "" {
		u = common.DownloadUrl
	}
	a.DownloadCdnUrl = u
}

func (a *Argument) SetTokenMaxAge() {
	s := os.Getenv(ENV_TOKEN_MAX_AGE)
	age, err := strconv.ParseInt(s, 10, 64)
	if err != nil || age == 0 {
		age = DefaultTokenMaxAge
	}
	a.TokenMaxAge = age
}

func (a *Argument) SetGPU(enable bool, share bool) {
	if a.GPU == nil {
		a.GPU = new(GPU)
	}
	a.GPU.Enable = enable
	a.GPU.Share = share
}

func (a *Argument) SetOlaresVersion(version string) {
	if version == "" || len(version) <= 2 {
		return
	}

	if version[0] == 'v' {
		version = version[1:]
	}
	a.OlaresVersion = version
}

func (a *Argument) SetRegistryMirrors(registryMirrors string) {
	a.RegistryMirrors = registryMirrors
}

func (a *Argument) SetDeleteCache(deleteCache bool) {
	a.DeleteCache = deleteCache
}

func (a *Argument) SetDeleteCRI(deleteCRI bool) {
	a.DeleteCRI = deleteCRI
}

func (a *Argument) SetStorage(storage *Storage) {
	a.Storage = storage
}

func (a *Argument) SetMinikubeProfile(profile string) {
	a.MinikubeProfile = profile
	if profile == "" && a.SystemInfo.IsDarwin() {
		fmt.Printf("\nNote: Minikube profile is not set, will try to use the default profile: \"%s\"\n", MinikubeDefaultProfile)
		fmt.Println("if this is not expected, please specify it explicitly by setting the --profile/-p option\n")
		a.MinikubeProfile = MinikubeDefaultProfile
	}
}

func (a *Argument) SetWSLDistribution(distribution string) {
	a.WSLDistribution = distribution
	if distribution == "" && a.SystemInfo.IsWindows() {
		fmt.Printf("\nNote: WSL distribution is not set, will try to use the default distribution: \"%s\"\n", WSLDefaultDistribution)
		fmt.Println("if this is not expected, please specify it explicitly by setting the --distribution/-d option\n")
		a.WSLDistribution = WSLDefaultDistribution
	}
}

func (a *Argument) SetReverseProxy() {
	var enableCloudflare = os.Getenv("CLOUDFLARE_ENABLE")
	var enableFrp = "0"
	var frpServer = ""
	var frpPort = "0"
	var frpAuthMethod = ""
	var frpAuthToken = ""

	if enableCloudflare == "" {
		enableCloudflare = "1"
	}
	if a.PublicNetworkInfo.PubliclyAccessible {
		enableCloudflare = "0"
	} else if os.Getenv("FRP_ENABLE") == "1" {
		enableCloudflare = "0"
		enableFrp = "1"
		frpServer = os.Getenv("FRP_SERVER")
		frpPort = os.Getenv("FRP_PORT")
		frpAuthMethod = os.Getenv("FRP_AUTH_METHOD")
		frpAuthToken = os.Getenv("FRP_AUTH_TOKEN")
	}

	a.Cloudflare.Enable = enableCloudflare
	a.Frp.Enable = enableFrp
	a.Frp.Server = util.RemoveHTTPPrefix(frpServer)
	a.Frp.Port = frpPort
	a.Frp.AuthMethod = frpAuthMethod
	a.Frp.AuthToken = frpAuthToken
}

func (a *Argument) SetKubeVersion(kubeType string) {
	var kubeVersion = DefaultK3sVersion
	if kubeType == K8s {
		kubeVersion = DefaultK8sVersion
	}
	a.KubernetesVersion = kubeVersion
	a.Kubetype = kubeType
}

func (a *Argument) SetKubernetesVersion(kubeType string, kubeVersion string) {
	a.KubernetesVersion = kubeVersion
	a.Kubetype = kubeType
}

func (a *Argument) SetBaseDir(dir string) {
	if dir != "" {
		a.BaseDir = dir
	}
	if a.BaseDir == "" {
		a.BaseDir = filepath.Join(a.SystemInfo.GetHomeDir(), common.DefaultBaseDir)
	}
	if !filepath.IsAbs(a.BaseDir) {
		var err error
		var absBaseDir string
		absBaseDir, err = filepath.Abs(a.BaseDir)
		if err != nil {
			panic(fmt.Errorf("failed to get absolute path for base directory %s: %v", a.BaseDir, err))
		}
		a.BaseDir = absBaseDir
	}
}

func (a *Argument) SetCudaVersion(cudaVersion string) {
	a.CudaVersion = CurrentVerifiedCudaVersion
	if cudaVersion != "" {
		a.CudaVersion = cudaVersion
	}
}

func (a *Argument) SetManifest(manifest string) {
	a.Manifest = manifest
}

func (a *Argument) SetConsoleLog(fileName string, truncate bool) {
	a.ConsoleLogFileName = fileName
	a.ConsoleLogTruncate = truncate
}

func (a *Argument) SetSwapConfig(config SwapConfig) {
	a.SwapConfig = &SwapConfig{}
	if config.ZRAMSize != "" || config.ZRAMSwapPriority != 0 {
		a.EnableZRAM = true
	} else {
		a.EnableZRAM = config.EnableZRAM
	}
	if a.EnableZRAM {
		a.ZRAMSize = config.ZRAMSize
		a.ZRAMSwapPriority = config.ZRAMSwapPriority
		a.EnablePodSwap = true
	} else {
		a.EnablePodSwap = config.EnablePodSwap
	}
	a.Swappiness = config.Swappiness
}

func (a *Argument) SetMasterHostOverride(config MasterHostConfig) {
	if config.MasterHost != "" {
		a.MasterHost = config.MasterHost
	}
	if config.MasterNodeName != "" {
		a.MasterNodeName = config.MasterNodeName
	}

	// set a dummy name to bypass validity checks
	// as it will be overridden later when the node name is fetched
	if a.MasterNodeName == "" {
		a.MasterNodeName = "master"
	}
	if config.MasterSSHPassword != "" {
		a.MasterSSHPassword = config.MasterSSHPassword
	}
	if config.MasterSSHUser != "" {
		a.MasterSSHUser = config.MasterSSHUser
	}
	if config.MasterSSHPort != 0 {
		a.MasterSSHPort = config.MasterSSHPort
	}
	if config.MasterSSHPrivateKeyPath != "" {
		a.MasterSSHPrivateKeyPath = config.MasterSSHPrivateKeyPath
	}
}

func (a *Argument) LoadMasterHostConfigIfAny() error {
	if a.BaseDir == "" {
		return errors.New("basedir unset")
	}
	content, err := os.ReadFile(filepath.Join(a.BaseDir, MasterHostConfigFile))
	if os.IsNotExist(err) {
		return nil
	}
	if err != nil {
		return err
	}
	return json.Unmarshal(content, a.MasterHostConfig)
}

func NewKubeRuntime(flag string, arg Argument) (*KubeRuntime, error) {
	loader := NewLoader(flag, arg)
	cluster, err := loader.Load()
	if err != nil {
		return nil, err
	}

	if err = loadExtraAddons(cluster, arg.ExtraAddon); err != nil {
		return nil, err
	}

	base := connector.NewBaseRuntime(cluster.Name, connector.NewDialer(),
		arg.Debug, arg.IgnoreErr, arg.Provider, arg.BaseDir, arg.OlaresVersion, arg.ConsoleLogFileName, arg.ConsoleLogTruncate, arg.SystemInfo)

	clusterSpec := &cluster.Spec
	defaultCluster, roleGroups := clusterSpec.SetDefaultClusterSpec(arg.InCluster, arg.SystemInfo.IsDarwin())
	hostSet := make(map[string]struct{})
	for _, role := range roleGroups {
		for _, host := range role {
			if host.IsRole(Master) || host.IsRole(Worker) {
				host.SetRole(K8s)
			}
			if host.IsRole(Master) && arg.SkipMasterPullImages {
				host.GetCache().Set(SkipMasterNodePullImages, true)
			}
			if _, ok := hostSet[host.GetName()]; !ok {
				hostSet[host.GetName()] = struct{}{}
				base.AppendHost(host)
				base.AppendRoleMap(host)
			}
			host.SetOs(arg.SystemInfo.GetOsType())
			host.SetMinikubeProfile(arg.MinikubeProfile)
		}
	}

	args, _ := json.Marshal(arg)
	logger.Debugf("[runtime] arg: %s", string(args))

	arg.KsEnable = defaultCluster.KubeSphere.Enabled
	arg.KsVersion = defaultCluster.KubeSphere.Version
	r := &KubeRuntime{
		Cluster:     defaultCluster,
		ClusterName: cluster.Name,
		Arg:         &arg,
	}
	r.BaseRuntime = base

	return r, nil
}

// Copy is used to create a copy for Runtime.
func (k *KubeRuntime) Copy() connector.Runtime {
	runtime := *k
	return &runtime
}
