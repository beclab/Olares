package pipelines

import (
	"bytetrade.io/web3os/installer/cmd/ctl/options"
	"bytetrade.io/web3os/installer/pkg/common"
	"bytetrade.io/web3os/installer/pkg/core/logger"
	"bytetrade.io/web3os/installer/pkg/phase/download"
	"path"
)

func CheckDownloadInstallationPackage(opts *options.CliDownloadOptions) error {
	arg := common.NewArgument()
	arg.SetOlaresVersion(opts.Version)
	arg.SetBaseDir(opts.BaseDir)

	runtime, err := common.NewKubeRuntime(common.AllInOne, *arg)
	if err != nil {
		return err
	}

	manifest := opts.Manifest
	if manifest == "" {
		manifest = path.Join(runtime.GetInstallerDir(), "installation.manifest")
	}

	p := download.NewCheckDownload(manifest, runtime)
	if err := p.Start(); err != nil {
		logger.Errorf("check download package failed %v", err)
		return err
	}

	return nil
}
