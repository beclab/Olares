package pipelines

import (
	"fmt"

	"github.com/beclab/Olares/cli/cmd/ctl/options"
	"github.com/beclab/Olares/cli/pkg/common"
	"github.com/beclab/Olares/cli/pkg/core/logger"
	"github.com/beclab/Olares/cli/pkg/phase/download"
	"github.com/beclab/Olares/cli/pkg/utils"
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

	p := download.NewDownloadWizard(runtime, opts.UrlOverride, opts.ReleaseID)
	if err := p.Start(); err != nil {
		logger.Errorf("download wizard failed %v", err)
		return err
	}

	return nil
}
