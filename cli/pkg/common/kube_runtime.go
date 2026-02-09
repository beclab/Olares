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
	"net"
	"os"
	"path/filepath"
	"strconv"
	"strings"

	"github.com/joho/godotenv"

	"github.com/pkg/errors"
	"github.com/spf13/viper"

	kubekeyapiv1alpha2 "github.com/beclab/Olares/cli/apis/kubekey/v1alpha2"
	"github.com/beclab/Olares/cli/pkg/core/common"
	"github.com/beclab/Olares/cli/pkg/core/connector"
	"github.com/beclab/Olares/cli/pkg/core/logger"
	"github.com/beclab/Olares/cli/pkg/core/util"
	kresource "k8s.io/apimachinery/pkg/api/resource"
)

type KubeRuntime struct {
	connector.BaseRuntime
	Cluster *kubekeyapiv1alpha2.ClusterSpec
	Arg     *Argument
}

type Argument struct {
	KubernetesVersion   string `json:"kubernetes_version"`
	OlaresVersion       string `json:"olares_version"`
	SecurityEnhancement bool   `json:"security_enhancement"`
	InCluster           bool   `json:"in_cluster"`
	ContainerManager    string `json:"container_manager"`
	Kubetype            string `json:"kube_type"`
	SystemInfo          connector.Systems

	RegistryMirrors  string `json:"registry_mirrors"`
	OlaresCDNService string `json:"olares_cdn_service"`

	// Swap config
	*SwapConfig

	// master node ssh config
	*MasterHostConfig

	// User
	User *User `json:"user"`
	// if juicefs is opted off, the local storage is used directly
	// only used in prepare phase
	// the existence of juicefs should be checked in other phases
	// to avoid wrong information given by user
	WithJuiceFS bool `json:"with_juicefs"`
	// the object storage service used as backend for JuiceFS
	Storage         *Storage         `json:"storage"`
	NetworkSettings *NetworkSettings `json:"network_settings"`
	GPU             *GPU             `json:"gpu"`

	IsCloudInstance    bool     `json:"is_cloud_instance"`
	MinikubeProfile    string   `json:"minikube_profile"`
	WSLDistribution    string   `json:"wsl_distribution"`
	Environment        []string `json:"environment"`
	BaseDir            string   `json:"base_dir"`
	Manifest           string   `json:"manifest"`
	ConsoleLogFileName string   `json:"console_log_file_name"`
	ConsoleLogTruncate bool     `json:"console_log_truncate"`
	HostIP             string   `json:"host_ip"`

	IsOlaresInContainer bool `json:"is_olares_in_container"`
}

type SwapConfig struct {
	EnablePodSwap    bool   `json:"enable_pod_swap"`
	Swappiness       int    `json:"swappiness"`
	EnableZRAM       bool   `json:"enable_zram"`
	ZRAMSize         string `json:"zram_size"`
	ZRAMSwapPriority int    `json:"zram_swap_priority"`
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

func (cfg *MasterHostConfig) Validate() error {
	if cfg.MasterHost == "" {
		return errors.New("master host is not provided")
	}
	if cfg.MasterSSHUser != "" && cfg.MasterSSHUser != "root" && cfg.MasterSSHPassword == "" {
		return errors.New("master ssh password must be provided for non-root user in order to execute sudo command")
	}
	return nil
}

type NetworkSettings struct {
	// OSPublicIPs contains a list of public ip(s)
	// by looking at local network interfaces
	// if any
	OSPublicIPs []net.IP `json:"os_public_ips"`

	// CloudProviderPublicIP contains the info retrieved from the cloud provider instance metadata service
	// if any
	CloudProviderPublicIP net.IP `json:"cloud_provider_public_ip"`

	// ExternalPublicIP is the IP address seen by others on the internet
	// it may not be an IP address
	// that's directly bound to a local network interface, e.g. on an AWS EC2 instance
	// or may not be an IP address
	// that can be used to access the machine at all, e.g. a machine behind multiple NAT gateways
	// this is used as a fallback method to determine the machine's public IP address
	// if none can be found from the OS or AWS IMDS service
	// but the user explicitly specifies that the machine is publicly accessible
	ExternalPublicIP net.IP `json:"external_public_ip"`

	EnableReverseProxy *bool `json:"enable_reverse_proxy"`
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

	StorageToken        string `json:"storage_token"`       // juicefs
	StorageClusterId    string `json:"storage_cluster_id"`  // use only on the Terminus cloud, juicefs
	StorageSyncSecret   string `json:"storage_sync_secret"` // use only on the Terminus cloud
	BackupClusterBucket string `json:"backup_cluster_bucket"`
}

type GPU struct {
	Enable bool `json:"gpu_enable"`
}

func NewArgument() *Argument {
	si := connector.GetSystemInfo()
	arg := &Argument{
		ContainerManager: Containerd,
		SystemInfo:       si,
		Storage: &Storage{
			StorageType: ManagedMinIO,
		},
		GPU: &GPU{},
		User: &User{
			UserName:   strings.TrimSpace(viper.GetString(FlagOSUserName)),
			DomainName: strings.TrimSpace(viper.GetString(FlagOSDomainName)),
			Password:   strings.TrimSpace(viper.GetString(FlagOSPassword)),
		},
		NetworkSettings:  &NetworkSettings{},
		RegistryMirrors:  viper.GetString(FlagRegistryMirrors),
		OlaresCDNService: viper.GetString(FlagCDNService),
		HostIP:           viper.GetString(ENV_HOST_IP),
		Environment:      os.Environ(),
		MasterHostConfig: &MasterHostConfig{},
		SwapConfig:       &SwapConfig{},
	}
	// default enable GPU unless explicitly set to "0"
	arg.GPU.Enable = !strings.EqualFold(os.Getenv(ENV_LOCAL_GPU_ENABLE), "0")
	arg.IsCloudInstance, _ = strconv.ParseBool(os.Getenv(ENV_TERMINUS_IS_CLOUD_VERSION))
	arg.IsOlaresInContainer = os.Getenv(ENV_CONTAINER_MODE) == "oic"
	si.IsOIC = arg.IsOlaresInContainer

	// Ensure BaseDir is initialized before loading master.conf
	// so master host config can be loaded from ${base-dir}/master.conf reliably.
	arg.SetBaseDir(viper.GetString(FlagBaseDir))
	arg.loadMasterHostConfig()
	return arg
}

func (a *Argument) SaveReleaseInfo(withoutName bool) error {
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

	if !withoutName {
		if a.User != nil && a.User.UserName != "" && a.User.DomainName != "" {
			releaseInfoMap["OLARES_NAME"] = fmt.Sprintf("%s@%s", a.User.UserName, a.User.DomainName)
		} else {
			if util.IsExist(OlaresReleaseFile) {
				// if the user is not set, try to load the user name from the release file
				envs, err := godotenv.Read(OlaresReleaseFile)
				if err == nil {
					if userName, ok := envs["OLARES_NAME"]; ok {
						releaseInfoMap["OLARES_NAME"] = userName
					}
				}
			}
		}
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

func (a *Argument) SetOlaresCDNService(url string) {
	u := strings.TrimSuffix(url, "/")
	if u == "" {
		u = common.DefaultOlaresCDNService
	}
	a.OlaresCDNService = u
}

func (a *Argument) SetGPU(enable bool) {
	if a.GPU == nil {
		a.GPU = new(GPU)
	}
	a.GPU.Enable = enable
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

func (a *Argument) SetKubeVersion(kubeType string) {
	var kubeVersion = DefaultK3sVersion
	if kubeType == K8s {
		kubeVersion = DefaultK8sVersion
	}
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

// loadMasterHostConfig loads master host configuration from master.conf file (if exists)
// and then overrides with any values set via command line flags or environment variables.
func (a *Argument) loadMasterHostConfig() {
	// First, try to load from master.conf file
	configPath := filepath.Join(a.BaseDir, MasterHostConfigFile)
	if content, err := os.ReadFile(configPath); err == nil {
		json.Unmarshal(content, a.MasterHostConfig)
	}
	// Then override with viper values (from flags or env)
	if v := viper.GetString(FlagMasterHost); v != "" {
		a.MasterHost = v
	}
	if v := viper.GetString(FlagMasterNodeName); v != "" {
		a.MasterNodeName = v
	}
	if v := viper.GetString(FlagMasterSSHUser); v != "" {
		a.MasterSSHUser = v
	}
	if v := viper.GetString(FlagMasterSSHPassword); v != "" {
		a.MasterSSHPassword = v
	}
	if v := viper.GetString(FlagMasterSSHPrivateKeyPath); v != "" {
		a.MasterSSHPrivateKeyPath = v
	}
	if v := viper.GetInt(FlagMasterSSHPort); v != 0 {
		a.MasterSSHPort = v
	}
	// Set a dummy name to bypass validity checks if master host is set but node name is not
	if a.MasterHost != "" && a.MasterNodeName == "" {
		a.MasterNodeName = "master"
	}
}

func (a *Argument) ClearMasterHostConfig() {
	a.MasterHostConfig = &MasterHostConfig{}
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

func NewKubeRuntime(arg Argument) (*KubeRuntime, error) {
	loader := NewLoader(arg)
	cluster, err := loader.Load()
	if err != nil {
		return nil, err
	}

	base := connector.NewBaseRuntime(cluster.Name, connector.NewDialer(),
		arg.BaseDir, arg.OlaresVersion, arg.ConsoleLogFileName, arg.ConsoleLogTruncate, arg.SystemInfo)

	clusterSpec := &cluster.Spec
	defaultCluster, roleGroups := clusterSpec.SetDefaultClusterSpec(arg.InCluster, arg.SystemInfo.IsDarwin())
	hostSet := make(map[string]struct{})
	for _, role := range roleGroups {
		for _, host := range role {
			if host.IsRole(Master) || host.IsRole(Worker) {
				host.SetRole(K8s)
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

	r := &KubeRuntime{
		Cluster: defaultCluster,
		Arg:     &arg,
	}
	r.BaseRuntime = base

	return r, nil
}

// Copy is used to create a copy for Runtime.
func (k *KubeRuntime) Copy() connector.Runtime {
	runtime := *k
	return &runtime
}
