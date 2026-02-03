package pipelines

import (
	"path"

	"github.com/beclab/Olares/cli/pkg/common"
	"github.com/beclab/Olares/cli/pkg/core/logger"
	"github.com/beclab/Olares/cli/pkg/phase/download"
	"github.com/spf13/viper"
)

func CheckDownloadInstallationPackage() error {
	arg := common.NewArgument()
	arg.SetOlaresVersion(viper.GetString(common.FlagVersion))

	runtime, err := common.NewKubeRuntime(*arg)
	if err != nil {
		return err
	}

	manifest := viper.GetString(common.FlagManifest)
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
