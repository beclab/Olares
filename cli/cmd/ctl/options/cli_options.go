package options

import (
	"bytetrade.io/web3os/installer/pkg/common"
	cc "bytetrade.io/web3os/installer/pkg/core/common"
	"bytetrade.io/web3os/installer/pkg/phase/cluster"
	"github.com/spf13/cobra"
)

type CliTerminusUninstallOptions struct {
	Version string
	BaseDir string
	All     bool
	Phase   string
	Quiet   bool
}

func NewCliTerminusUninstallOptions() *CliTerminusUninstallOptions {
	return &CliTerminusUninstallOptions{}
}

func (o *CliTerminusUninstallOptions) AddFlags(cmd *cobra.Command) {
	cmd.Flags().StringVarP(&o.Version, "version", "v", "", "Set Olares version, e.g., 1.10.0, 1.10.0-20241109")
	cmd.Flags().StringVarP(&o.BaseDir, "base-dir", "b", "", "Set Olares package base dir, defaults to $HOME/"+cc.DefaultBaseDir)
	cmd.Flags().BoolVar(&o.All, "all", false, "Uninstall Olares completely, including prepared dependencies")
	cmd.Flags().StringVar(&o.Phase, "phase", cluster.PhaseInstall.String(), "Uninstall from a specified phase and revert to the previous one. For example, using --phase install will remove the tasks performed in the 'install' phase, effectively returning the system to the 'prepare' state.")
	cmd.Flags().BoolVar(&o.Quiet, "quiet", false, "Quiet mode, default: false")
}

type CliTerminusInstallOptions struct {
	Version         string
	KubeType        string
	WithJuiceFS     bool
	MiniKubeProfile string
	BaseDir         string
	common.SwapConfig
}

func NewCliTerminusInstallOptions() *CliTerminusInstallOptions {
	return &CliTerminusInstallOptions{}
}

func (o *CliTerminusInstallOptions) AddFlags(cmd *cobra.Command) {
	cmd.Flags().StringVarP(&o.Version, "version", "v", "", "Set Olares version, e.g., 1.10.0, 1.10.0-20241109")
	cmd.Flags().StringVar(&o.KubeType, "kube", "k3s", "Set kube type, e.g., k3s or k8s")
	cmd.Flags().BoolVar(&o.WithJuiceFS, "with-juicefs", false, "Use JuiceFS as the rootfs for Olares workloads, rather than the local disk.")
	cmd.Flags().StringVarP(&o.MiniKubeProfile, "profile", "p", "", "Set Minikube profile name, only in MacOS platform, defaults to "+common.MinikubeDefaultProfile)
	cmd.Flags().StringVarP(&o.BaseDir, "base-dir", "b", "", "Set Olares package base dir, defaults to $HOME/"+cc.DefaultBaseDir)
	(&o.SwapConfig).AddFlags(cmd.Flags())
}

type CliPrepareSystemOptions struct {
	Version         string
	KubeType        string
	RegistryMirrors string
	BaseDir         string
	MinikubeProfile string
}

func NewCliPrepareSystemOptions() *CliPrepareSystemOptions {
	return &CliPrepareSystemOptions{}
}

func (o *CliPrepareSystemOptions) AddFlags(cmd *cobra.Command) {
	cmd.Flags().StringVarP(&o.Version, "version", "v", "", "Set Olares version, e.g., 1.10.0, 1.10.0-20241109")
	cmd.Flags().StringVar(&o.KubeType, "kube", "k3s", "Set kube type, e.g., k3s or k8s")
	cmd.Flags().StringVarP(&o.RegistryMirrors, "registry-mirrors", "r", "", "Docker Container registry mirrors, multiple mirrors are separated by commas")
	cmd.Flags().StringVarP(&o.BaseDir, "base-dir", "b", "", "Set Olares package base dir, defaults to $HOME/"+cc.DefaultBaseDir)
	cmd.Flags().StringVarP(&o.MinikubeProfile, "profile", "p", "", "Set Minikube profile name, only in MacOS platform, defaults to "+common.MinikubeDefaultProfile)
}

type ChangeIPOptions struct {
	Version         string
	BaseDir         string
	NewMasterHost   string
	WSLDistribution string
	MinikubeProfile string
}

func NewChangeIPOptions() *ChangeIPOptions {
	return &ChangeIPOptions{}
}

func (o *ChangeIPOptions) AddFlags(cmd *cobra.Command) {
	cmd.Flags().StringVarP(&o.Version, "version", "v", "", "Set Olares version, e.g., 1.10.0, 1.10.0-20241109")
	cmd.Flags().StringVarP(&o.BaseDir, "base-dir", "b", "", "Set Olares package base dir, defaults to $HOME/"+cc.DefaultBaseDir)
	cmd.Flags().StringVar(&o.NewMasterHost, "new-master-host", "", "Update the master node's IP if it's changed, only in Linux worker node")
	cmd.Flags().StringVarP(&o.WSLDistribution, "distribution", "d", "", "Set WSL distribution name, only in Windows platform, defaults to "+common.WSLDefaultDistribution)
	cmd.Flags().StringVarP(&o.MinikubeProfile, "profile", "p", "", "Set Minikube profile name, only in MacOS platform, defaults to "+common.MinikubeDefaultProfile)
}

type PreCheckOptions struct {
	Version string
	BaseDir string
}

func NewPreCheckOptions() *PreCheckOptions {
	return &PreCheckOptions{}
}

func (o *PreCheckOptions) AddFlags(cmd *cobra.Command) {
	cmd.Flags().StringVarP(&o.Version, "version", "v", "", "Set Olares version, e.g., 1.10.0, 1.10.0-20241109")
	cmd.Flags().StringVarP(&o.BaseDir, "base-dir", "b", "", "Set Olares package base dir, defaults to $HOME/"+cc.DefaultBaseDir)
}

type InstallStorageOptions struct {
	Version string
	BaseDir string
}

func NewInstallStorageOptions() *InstallStorageOptions {
	return &InstallStorageOptions{}
}

func (o *InstallStorageOptions) AddFlags(cmd *cobra.Command) {
	cmd.Flags().StringVarP(&o.Version, "version", "v", "", "Set Olares version, e.g., 1.10.0, 1.10.0-20241109")
	cmd.Flags().StringVarP(&o.BaseDir, "base-dir", "b", "", "Set Olares package base dir, defaults to $HOME/"+cc.DefaultBaseDir)
}

type AddNodeOptions struct {
	common.MasterHostConfig
	Version string
	BaseDir string
}

func NewAddNodeOptions() *AddNodeOptions {
	return &AddNodeOptions{}
}

func (o *AddNodeOptions) AddFlags(cmd *cobra.Command) {
	cmd.Flags().StringVarP(&o.Version, "version", "v", "", "Set Olares version, e.g., 1.10.0, 1.10.0-20241109")
	cmd.Flags().StringVarP(&o.BaseDir, "base-dir", "b", "", "Set Olares package base dir, defaults to $HOME/"+cc.DefaultBaseDir)
	(&o.MasterHostConfig).AddFlags(cmd.Flags())
}

type MasterInfoOptions struct {
	BaseDir string
	common.MasterHostConfig
}

func NewMasterInfoOptions() *MasterInfoOptions {
	return &MasterInfoOptions{}
}

func (o *MasterInfoOptions) AddFlags(cmd *cobra.Command) {
	cmd.Flags().StringVarP(&o.BaseDir, "base-dir", "b", "", "Set Olares package base dir, defaults to $HOME/"+cc.DefaultBaseDir)
	(&o.MasterHostConfig).AddFlags(cmd.Flags())
}

type UpgradeOptions struct {
	Version string
	BaseDir string
}

func NewUpgradeOptions() *UpgradeOptions {
	return &UpgradeOptions{}
}

func (o *UpgradeOptions) AddFlags(cmd *cobra.Command) {
	cmd.Flags().StringVarP(&o.Version, "version", "v", "", "Set target Olares version to upgrade to, e.g., 1.10.0, 1.10.0-20241109")
	cmd.Flags().StringVarP(&o.BaseDir, "base-dir", "b", "", "Set Olares package base dir, defaults to $HOME/"+cc.DefaultBaseDir)
}
