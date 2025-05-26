package kubesphere

import (
	"encoding/json"
	"fmt"
	"github.com/containerd/containerd/plugin"
	"github.com/pelletier/go-toml"
	"os"
	"os/exec"
	"path"
	"path/filepath"
	"strings"

	kubekeyapiv1alpha2 "bytetrade.io/web3os/installer/apis/kubekey/v1alpha2"
	"bytetrade.io/web3os/installer/pkg/common"
	"bytetrade.io/web3os/installer/pkg/core/connector"
	"bytetrade.io/web3os/installer/pkg/core/logger"
	"bytetrade.io/web3os/installer/pkg/core/task"
	"bytetrade.io/web3os/installer/pkg/core/util"
	"bytetrade.io/web3os/installer/pkg/files"
	mk "bytetrade.io/web3os/installer/pkg/version/minikube"
	criconfig "github.com/containerd/containerd/pkg/cri/config"
	cdsrvconfig "github.com/containerd/containerd/services/server/config"
	"github.com/pkg/errors"
)

var minikubeContainerdConfigFilePath = "/etc/containerd/config.toml"

type CreateMiniKubeCluster struct {
	common.KubeAction
}

func (t *CreateMiniKubeCluster) Execute(runtime connector.Runtime) error {
	minikube, err := util.GetCommand(common.CommandMinikube)
	if err != nil {
		return fmt.Errorf("Please install minikube on your machine")
	}

	cmd := fmt.Sprintf("%s profile %s", minikube, t.KubeConf.Arg.MinikubeProfile)
	stdout, err := runtime.GetRunner().Cmd(cmd, false, false)
	if err != nil {
		return errors.Wrap(err, "failed to check minikube profile")
	} else if !strings.Contains(stdout, "not found") {
		logger.Infof("found old minikube cluster %s, deleting...", t.KubeConf.Arg.MinikubeProfile)
		cmd = fmt.Sprintf("%s delete -p %s", minikube, t.KubeConf.Arg.MinikubeProfile)
		stdout, err = runtime.GetRunner().Cmd(cmd, false, true)
		if err != nil {
			return errors.Wrap(err, "failed to delete old minikube cluster")
		}
	}
	logger.Infof("creating minikube cluster %s ...", t.KubeConf.Arg.MinikubeProfile)
	cmd = fmt.Sprintf("%s start -p '%s' --kubernetes-version=v1.22.10 --container-runtime=containerd --network-plugin=cni --cni=calico --cpus='4' --memory='8g' --ports=30180:30180,443:443,80:80", minikube, t.KubeConf.Arg.MinikubeProfile)
	if _, err := runtime.GetRunner().Cmd(cmd, false, true); err != nil {
		return errors.Wrap(err, "failed to create minikube cluster")
	}

	return nil
}

type RetagMinikubeKubeImages struct {
	common.KubeAction
}

func (t *RetagMinikubeKubeImages) Execute(runtime connector.Runtime) error {
	legacyKubeImageRepo := "k8s.gcr.io"
	newKubeImageRepo := "registry.k8s.io"
	minikube, err := util.GetCommand(common.CommandMinikube)
	if err != nil {
		return fmt.Errorf("Please install minikube on your machine")
	}

	cmd := fmt.Sprintf("%s image ls -p %s", minikube, t.KubeConf.Arg.MinikubeProfile)
	stdout, err := runtime.GetRunner().Cmd(cmd, false, false)
	if err != nil {
		return errors.Wrap(err, "failed to check list minikube images")
	}
	images := strings.Split(stdout, "\n")
	for _, image := range images {
		if strings.HasPrefix(image, legacyKubeImageRepo) {
			newTag := strings.ReplaceAll(image, legacyKubeImageRepo, newKubeImageRepo)
			cmd = fmt.Sprintf("%s image tag %s %s -p %s", minikube, image, newTag, t.KubeConf.Arg.MinikubeProfile)
			_, err = runtime.GetRunner().Cmd(cmd, false, false)
			if err != nil {
				return errors.Wrap(err, "failed to retag minikube images")
			}
		}
	}
	return nil
}

type GetMiniKubeContainerdConfig struct {
	common.KubeAction
}

func (t *GetMiniKubeContainerdConfig) Execute(runtime connector.Runtime) error {
	minikube, err := util.GetCommand(common.CommandMinikube)
	if err != nil {
		return fmt.Errorf("failed to get minikube command: %w", err)
	}
	tmpConfigFile, err := os.CreateTemp("", "minikube-containerd-config-*.toml")
	if err != nil {
		return fmt.Errorf("failed to create temp config file: %w", err)
	}
	t.ModuleCache.Set(common.CacheMinikubeTmpContainerdConfigFile, tmpConfigFile.Name())
	cmd := fmt.Sprintf("%s ssh cat %s > %s -p %s", minikube, minikubeContainerdConfigFilePath, tmpConfigFile.Name(), t.KubeConf.Arg.MinikubeProfile)
	_, err = runtime.GetRunner().Cmd(cmd, false, false)
	if err != nil {
		return fmt.Errorf("failed to get minikube containerd config: %w", err)
	}
	return nil
}

type SetMirrorsToMinikubeContainerdConfig struct {
	common.KubeAction
}

func (t *SetMirrorsToMinikubeContainerdConfig) Execute(runtime connector.Runtime) error {
	if len(t.KubeConf.Cluster.Registry.RegistryMirrors) == 0 {
		return nil
	}
	tmpConfigFilePath, ok := t.ModuleCache.GetMustString(common.CacheMinikubeTmpContainerdConfigFile)
	if !ok || tmpConfigFilePath == "" {
		return errors.New("failed to get minikube containerd config temp file path")
	}
	config := &cdsrvconfig.Config{}
	err := cdsrvconfig.LoadConfig(tmpConfigFilePath, config)
	if err != nil {
		return fmt.Errorf("failed to load minikube containerd config: %w", err)
	}
	var filteredImports []string
	for _, imp := range config.Imports {
		if strings.EqualFold(imp, tmpConfigFilePath) {
			continue
		}
		filteredImports = append(filteredImports, imp)
	}
	config.Imports = filteredImports

	if config.Plugins == nil {
		config.Plugins = make(map[string]toml.Tree)
	}
	criDefaultPluginConfig := criconfig.DefaultConfig()
	criPlugin := &plugin.Registration{
		Type:   plugin.GRPCPlugin,
		ID:     "cri",
		Config: &criDefaultPluginConfig,
	}
	criPluginConfigInterface, err := config.Decode(criPlugin)
	if err != nil {
		return fmt.Errorf("failed to load minikube containerd cri plugin config: %w", err)
	}
	criPluginConfig, ok := criPluginConfigInterface.(*criconfig.PluginConfig)
	if !ok {
		return fmt.Errorf("failed to load minikube containerd cri plugin config: decoded type mismatch")
	}
	if criPluginConfig.Registry.ConfigPath != "" {
		// reset config path as it will mask the other options
		// we do not set mirrors in the config path
		// because image-service expects an explicit inline config in the Mirrors field
		criPluginConfig.Registry.ConfigPath = ""
	}
	if criPluginConfig.Registry.Mirrors == nil {
		criPluginConfig.Registry.Mirrors = make(map[string]criconfig.Mirror)
	}
	registryHost := "docker.io"
	registryMirrors := criPluginConfig.Registry.Mirrors[registryHost]
	registryMirrors.Endpoints = append(registryMirrors.Endpoints, t.KubeConf.Cluster.Registry.RegistryMirrors...)
	criPluginConfig.Registry.Mirrors[registryHost] = registryMirrors
	criPluginConfigBytes, err := toml.Marshal(criPluginConfig)
	if err != nil {
		return fmt.Errorf("failed to marshal minikube containerd cri plugin config: %w", err)
	}
	criPluginConfigTree, err := toml.LoadBytes(criPluginConfigBytes)
	if err != nil {
		return fmt.Errorf("failed to load minikube containerd cri plugin config: %w", err)
	}
	config.Plugins[criPlugin.URI()] = *criPluginConfigTree

	tmpConfigFile, err := os.OpenFile(tmpConfigFilePath, os.O_WRONLY|os.O_CREATE|os.O_TRUNC, 0644)
	if err != nil {
		return fmt.Errorf("failed to open minikube containerd config temp file for writing: %w", err)
	}
	defer tmpConfigFile.Close()
	if err := toml.NewEncoder(tmpConfigFile).Encode(config); err != nil {
		return fmt.Errorf("failed to write minikube containerd config temp file: %w", err)
	}

	return nil
}

type ReloadMinikubeContainerdConfig struct {
	common.KubeAction
}

func (t *ReloadMinikubeContainerdConfig) Execute(runtime connector.Runtime) error {
	minikube, err := util.GetCommand(common.CommandMinikube)
	if err != nil {
		return fmt.Errorf("failed to get minikube command: %w", err)
	}
	tmpConfigFilePath, ok := t.ModuleCache.GetMustString(common.CacheMinikubeTmpContainerdConfigFile)
	if !ok || tmpConfigFilePath == "" {
		return errors.New("failed to get minikube containerd config temp file path")
	}
	cmd := fmt.Sprintf("%s cp %s %s -p %s", minikube, tmpConfigFilePath, minikubeContainerdConfigFilePath, t.KubeConf.Arg.MinikubeProfile)
	_, err = runtime.GetRunner().Cmd(cmd, false, false)
	if err != nil {
		return fmt.Errorf("failed to cp back minikube containerd config: %w", err)
	}

	cmd = fmt.Sprintf("%s ssh sudo systemctl restart containerd -p %s", minikube, t.KubeConf.Arg.MinikubeProfile)
	_, err = runtime.GetRunner().Cmd(cmd, false, false)
	if err != nil {
		return fmt.Errorf("failed to restart containerd in minikube: %w", err)
	}

	if err := os.Remove(tmpConfigFilePath); err != nil {
		logger.Warnf("failed to remove temp minikube containerd config temp file %s: %v", tmpConfigFilePath, err)
	}

	return nil
}

type CreateMinikubeClusterModule struct {
	common.KubeModule
}

func (m *CreateMinikubeClusterModule) Init() {
	m.Name = "CreateMinikubeCluster"

	createCluster := &task.LocalTask{
		Name:   "CreateMinikubeCluster",
		Action: new(CreateMiniKubeCluster),
	}

	retagMinikubeKubeImages := &task.LocalTask{
		Name:   "RetagMinikubeKubeImages",
		Action: new(RetagMinikubeKubeImages),
	}

	getMiniKubeContainerdConfig := &task.LocalTask{
		Name:   "GetMiniKubeContainerdConfig",
		Action: new(GetMiniKubeContainerdConfig),
	}

	setMirrorsToMinikubeContainerdConfig := &task.LocalTask{
		Name:   "SetMirrorsToMinikubeContainerdConfig",
		Action: new(SetMirrorsToMinikubeContainerdConfig),
	}

	reloadMinikubeContainerdConfig := &task.LocalTask{
		Name:   "ReloadMinikubeContainerdConfig",
		Action: new(ReloadMinikubeContainerdConfig),
	}

	m.Tasks = []task.Interface{
		createCluster,
		retagMinikubeKubeImages,
		getMiniKubeContainerdConfig,
		setMirrorsToMinikubeContainerdConfig,
		reloadMinikubeContainerdConfig,
	}
}

type UninstallMinikube struct {
	common.KubeAction
}

func (t *UninstallMinikube) Execute(runtime connector.Runtime) error {
	var minikubepath string
	var err error
	if minikubepath, err = util.GetCommand(common.CommandMinikube); err != nil || minikubepath == "" {
		return fmt.Errorf("minikube not found")
	}

	if _, err := runtime.GetRunner().Cmd(fmt.Sprintf("%s stop --all && %s delete --all", minikubepath, minikubepath), false, true); err != nil {
		return err
	}

	var phaseStateFiles = []string{common.TerminusStateFileInstalled, common.TerminusStateFilePrepared}
	for _, ps := range phaseStateFiles {
		if util.IsExist(path.Join(runtime.GetBaseDir(), ps)) {
			util.RemoveFile(path.Join(runtime.GetBaseDir(), ps))
		}
	}
	return nil
}

type DeleteMinikubeModule struct {
	common.KubeModule
}

func (m *DeleteMinikubeModule) Init() {
	m.Name = "Uninstall"

	uninstallMinikube := &task.LocalTask{
		Name:   "Uninstall",
		Action: new(UninstallMinikube),
	}

	m.Tasks = []task.Interface{
		uninstallMinikube,
	}
}

type Download struct {
	common.KubeAction
}

func (t *Download) Execute(runtime connector.Runtime) error {
	var arch = runtime.RemoteHost().GetArch()

	var systemInfo = runtime.GetSystemInfo()
	var osType = systemInfo.GetOsType()
	var osVersion = systemInfo.GetOsVersion()
	var osPlatformFamily = systemInfo.GetOsPlatformFamily()
	helm := files.NewKubeBinary("helm", arch, osType, osVersion, osPlatformFamily, kubekeyapiv1alpha2.DefaultHelmVersion, runtime.GetWorkDir(), "")

	if err := helm.CreateBaseDir(); err != nil {
		return errors.Wrapf(errors.WithStack(err), "create file %s base dir failed", helm.FileName)
	}

	var exists = util.IsExist(helm.Path())
	if exists {
		p := helm.Path()
		if err := helm.SHA256Check(); err != nil {
			_ = exec.Command("/bin/sh", "-c", fmt.Sprintf("rm -f %s", p)).Run()
		}
	}

	if !exists || helm.OverWrite {
		logger.Infof("%s downloading %s %s %s ...", common.LocalHost, arch, helm.ID, helm.Version)
		if err := helm.Download(); err != nil {
			return fmt.Errorf("Failed to download %s binary: %s error: %w ", helm.ID, helm.Url, err)
		}
	}

	return nil
}

type DownloadMinikubeBinaries struct {
	common.KubeModule
}

func (m *DownloadMinikubeBinaries) Init() {
	m.Name = "DownloadMinikubeBinaries"

	downloadBinaries := &task.RemoteTask{
		Name:     "DownloadHelm",
		Hosts:    m.Runtime.GetHostsByRole(common.Master),
		Action:   new(Download),
		Parallel: false,
		Retry:    1,
	}

	m.Tasks = []task.Interface{
		downloadBinaries,
	}
}

type GetMinikubeProfile struct {
	common.KubeAction
}

func (t *GetMinikubeProfile) Execute(runtime connector.Runtime) error {
	var minikubecmd, ok = t.PipelineCache.GetMustString(common.CacheCommandMinikubePath)
	if !ok || minikubecmd == "" {
		minikubecmd = path.Join(common.BinDir, "minikube")
	}
	var cmd = fmt.Sprintf("%s -p %s profile list -o json --light=false", minikubecmd, runtime.RemoteHost().GetMinikubeProfile())
	stdout, err := runtime.GetRunner().Cmd(cmd, false, false)
	if err != nil {
		return err
	}

	var p mk.Minikube
	if err := json.Unmarshal([]byte(stdout), &p); err != nil {
		return err
	}

	if p.Valid == nil || len(p.Valid) == 0 {
		return fmt.Errorf("minikube profile not found")
	}

	var nodeIp string
	for _, v := range p.Valid {
		if v.Name != runtime.RemoteHost().GetMinikubeProfile() {
			continue
		}
		if v.Config.Nodes == nil || len(v.Config.Nodes) == 0 {
			return fmt.Errorf("minikube node not found")
		}
		nodeIp = v.Config.Nodes[0].IP
	}

	if nodeIp == "" {
		return fmt.Errorf("minikube node ip is empty")
	}

	if !util.IsExist(common.KubeAddonsDir) {
		if _, err := runtime.GetRunner().SudoCmd(fmt.Sprintf("mkdir -p %s", common.KubeAddonsDir), false, true); err != nil {
			return errors.Wrap(errors.WithStack(err), fmt.Sprintf("create dir %s failed", common.KubeAddonsDir))
		}
	}

	t.PipelineCache.Set(common.CacheMinikubeNodeIp, nodeIp)

	return nil

}

type PatchCoreDNSSVC struct {
	common.KubeAction
}

func (t *PatchCoreDNSSVC) Execute(runtime connector.Runtime) error {
	var kubectlcmd, ok = t.PipelineCache.GetMustString(common.CacheCommandKubectlPath)
	if !ok || kubectlcmd == "" {
		kubectlcmd = path.Join(common.BinDir, "kubectl")
	}

	coreDNSSVCPatchFilePath := filepath.Join(runtime.GetInstallerDir(), "deploy/patch-k3s.yaml")
	_, err := runtime.GetRunner().Cmd(fmt.Sprintf("%s apply -f %s", kubectlcmd, coreDNSSVCPatchFilePath), false, true)
	if err != nil {
		return errors.Wrap(errors.WithStack(err), fmt.Sprintf("failed to patch coredns service", err))
	}
	return nil
}

type InitMinikubeNs struct {
	common.KubeAction
}

func (t *InitMinikubeNs) Execute(runtime connector.Runtime) error {
	var kubectlcmd, ok = t.PipelineCache.GetMustString(common.CacheCommandKubectlPath)
	if !ok || kubectlcmd == "" {
		kubectlcmd = path.Join(common.BinDir, "kubectl")
	}

	var allNs = []string{
		common.NamespaceKubekeySystem,
		common.NamespaceKubesphereSystem,
		common.NamespaceKubesphereMonitoringSystem,
	}

	for _, ns := range allNs {
		if stdout, err := runtime.GetRunner().Cmd(fmt.Sprintf("%s create ns %s", kubectlcmd, ns), false, true); err != nil {
			if !strings.Contains(stdout, "already exists") {
				logger.Errorf("create ns %s failed: %v", ns, err)
				return errors.Wrap(errors.WithStack(err), fmt.Sprintf("create namespace %s failed: %v", ns, err))
			}
		}
	}

	return nil
}

type CheckMacCommandExists struct {
	common.KubeAction
}

func (t *CheckMacCommandExists) Execute(runtime connector.Runtime) error {
	var err error
	var minikubepath string
	var kubectlpath string
	var dockerpath string

	if minikubepath, err = util.GetCommand(common.CommandMinikube); err != nil || minikubepath == "" {
		return fmt.Errorf("minikube not found")
	}

	if kubectlpath, err = util.GetCommand(common.CommandKubectl); err != nil || kubectlpath == "" {
		return fmt.Errorf("kubectl not found")
	}

	if dockerpath, err = util.GetCommand(common.CommandDocker); err != nil || dockerpath == "" {
		return fmt.Errorf("docker not found")
	}

	fmt.Println("kubectl path:", kubectlpath)
	fmt.Println("minikube path:", minikubepath)
	fmt.Println("docker path:", dockerpath)

	t.PipelineCache.Set(common.CacheCommandMinikubePath, minikubepath)
	t.PipelineCache.Set(common.CacheCommandKubectlPath, kubectlpath)
	t.PipelineCache.Set(common.CacheCommandDockerPath, dockerpath)

	return nil
}

type CheckMacOsCommandModule struct {
	common.KubeModule
}

func (m *CheckMacOsCommandModule) Init() {
	m.Name = "CheckCommandPath"

	checkMacCommandExists := &task.LocalTask{
		Name:   "CheckMiniKubeExists",
		Action: new(CheckMacCommandExists),
	}

	m.Tasks = []task.Interface{
		checkMacCommandExists,
	}
}

type DeployMiniKubeModule struct {
	common.KubeModule
}

func (m *DeployMiniKubeModule) Init() {
	m.Name = "DeployMiniKube"

	getMinikubeProfile := &task.RemoteTask{
		Name:     "GetMinikubeProfile",
		Hosts:    m.Runtime.GetHostsByRole(common.Master),
		Action:   new(GetMinikubeProfile),
		Parallel: false,
		Retry:    1,
	}

	patchCoreDNSSVC := &task.LocalTask{
		Name:   "PatchCoreDNSSVC",
		Action: new(PatchCoreDNSSVC),
		Retry:  1,
	}

	initMinikubeNs := &task.LocalTask{
		Name:   "InitMinikubeNs",
		Action: new(InitMinikubeNs),
		Retry:  1,
	}

	m.Tasks = []task.Interface{
		getMinikubeProfile,
		patchCoreDNSSVC,
		initMinikubeNs,
	}
}
