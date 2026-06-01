package pipelines

import (
	"context"
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
	// TODO(ctx): plumb ctx in a follow-up; this entry point is not yet ctx-aware.
	if err := p.Start(context.Background()); err != nil {
		logger.Errorf("check download package failed %v", err)
		return err
	}

	return nil
}
