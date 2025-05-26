package pipelines

import (
	"fmt"

	"bytetrade.io/web3os/installer/cmd/ctl/options"
	"bytetrade.io/web3os/installer/pkg/common"
	"bytetrade.io/web3os/installer/pkg/core/logger"
	"bytetrade.io/web3os/installer/pkg/phase/download"
	"bytetrade.io/web3os/installer/pkg/utils"
)

func DownloadInstallationWizard(opts *options.CliDownloadWizardOptions) error {
	arg := common.NewArgument()
	arg.SetKubeVersion(opts.KubeType)
	arg.SetOlaresVersion(opts.Version)
	arg.SetBaseDir(opts.BaseDir)
	arg.SetDownloadCdnUrl(opts.DownloadCdnUrl)

	runtime, err := common.NewKubeRuntime(common.AllInOne, *arg)
	if err != nil {
		return err
	}

	if ok := utils.CheckUrl(arg.DownloadCdnUrl); !ok {
		return fmt.Errorf("--download-cdn-url invalid")
	}

	p := download.NewDownloadWizard(runtime)
	if err := p.Start(); err != nil {
		logger.Errorf("download wizard failed %v", err)
		return err
	}

	return nil
}
