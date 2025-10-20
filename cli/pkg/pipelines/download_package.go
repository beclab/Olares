package pipelines

import (
	"fmt"
	"path"

	"github.com/beclab/Olares/cli/cmd/ctl/options"
	"github.com/beclab/Olares/cli/pkg/common"
	"github.com/beclab/Olares/cli/pkg/core/logger"
	"github.com/beclab/Olares/cli/pkg/phase/download"
	"github.com/beclab/Olares/cli/pkg/utils"
)

func DownloadInstallationPackage(opts *options.CliDownloadOptions) error {
	arg := common.NewArgument()
	arg.SetBaseDir(opts.BaseDir)
	arg.SetKubeVersion(opts.KubeType)
	arg.SetOlaresVersion(opts.Version)
	arg.SetOlaresCDNService(opts.CDNService)

	runtime, err := common.NewKubeRuntime(common.AllInOne, *arg)
	if err != nil {
		return err
	}

	if ok := utils.CheckUrl(arg.OlaresCDNService); !ok {
		return fmt.Errorf("--cdn-service invalid")
	}

	manifest := opts.Manifest
	if manifest == "" {
		manifest = path.Join(runtime.GetInstallerDir(), "installation.manifest")
	}

	p := download.NewDownloadPackage(manifest, runtime)
	if err := p.Start(); err != nil {
		logger.Errorf("download package failed %v", err)
		return err
	}

	return nil
}
