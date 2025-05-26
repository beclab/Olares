package options

import (
	cc "bytetrade.io/web3os/installer/pkg/core/common"
	"github.com/spf13/cobra"
)

type CliDownloadWizardOptions struct {
	Version        string
	KubeType       string
	BaseDir        string
	DownloadCdnUrl string
}

func NewCliDownloadWizardOptions() *CliDownloadWizardOptions {
	return &CliDownloadWizardOptions{}
}

func (o *CliDownloadWizardOptions) AddFlags(cmd *cobra.Command) {
	cmd.Flags().StringVarP(&o.Version, "version", "v", "", "Set Olares version, e.g., 1.10.0, 1.10.0-20241109")
	cmd.Flags().StringVarP(&o.BaseDir, "base-dir", "b", "", "Set Olares package base dir, defaults to $HOME/"+cc.DefaultBaseDir)
	cmd.Flags().StringVar(&o.KubeType, "kube", "k3s", "Set kube type, e.g., k3s or k8s")
	cmd.Flags().StringVar(&o.DownloadCdnUrl, "download-cdn-url", "", "Set the CDN accelerated download address in the format https://example.cdn.com. If not set, the default download address will be used")
}

type CliDownloadOptions struct {
	Version        string
	KubeType       string
	Manifest       string
	BaseDir        string
	DownloadCdnUrl string
}

func NewCliDownloadOptions() *CliDownloadOptions {
	return &CliDownloadOptions{}
}

func (o *CliDownloadOptions) AddFlags(cmd *cobra.Command) {
	cmd.Flags().StringVarP(&o.Version, "version", "v", "", "Set Olares version, e.g., 1.10.0, 1.10.0-20241109")
	cmd.Flags().StringVarP(&o.BaseDir, "base-dir", "b", "", "Set Olares package base dir , defaults to $HOME/"+cc.DefaultBaseDir)
	cmd.Flags().StringVar(&o.Manifest, "manifest", "", "Set package manifest file , defaults to {base-dir}/versions/v{version}/installation.manifest")
	cmd.Flags().StringVar(&o.KubeType, "kube", "k3s", "Set kube type, e.g., k3s or k8s")
	cmd.Flags().StringVar(&o.DownloadCdnUrl, "download-cdn-url", "", "Set the CDN accelerated download address in the format https://example.cdn.com. If not set, the default download address will be used")
}
