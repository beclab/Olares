package options

import (
	"bytetrade.io/web3os/installer/pkg/common"
	cc "bytetrade.io/web3os/installer/pkg/core/common"
	"fmt"
	"github.com/spf13/cobra"
)

type GpuOptions struct {
	Version string
	BaseDir string
}

func (o *GpuOptions) AddFlags(cmd *cobra.Command) {
	cmd.Flags().StringVarP(&o.Version, "version", "v", "", "Set Olares version, e.g., 1.10.0, 1.10.0-20241109")
	cmd.Flags().StringVarP(&o.BaseDir, "base-dir", "b", "", "Set Olares package base dir, defaults to $HOME/"+cc.DefaultBaseDir)
}

type InstallGpuOptions struct {
	GpuOptions
	Cuda string
}

func NewInstallGpuOptions() *InstallGpuOptions {
	return &InstallGpuOptions{}
}

func (o *InstallGpuOptions) AddFlags(cmd *cobra.Command) {
	o.GpuOptions.AddFlags(cmd)
	cmd.Flags().StringVar(&o.Cuda, "cuda", "", fmt.Sprintf("The version of the CUDA driver, current supported versions are %s", common.CurrentVerifiedCudaVersion))
}
