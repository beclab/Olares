package pipelines

import (
	"fmt"
	"path"

	"github.com/beclab/Olares/cli/pkg/common"
	"github.com/beclab/Olares/cli/pkg/core/logger"
	"github.com/beclab/Olares/cli/pkg/phase/download"
	"github.com/beclab/Olares/cli/pkg/utils"
	"github.com/spf13/viper"
)

func DownloadInstallationPackage() error {
	arg := common.NewArgument()
	arg.SetOlaresVersion(viper.GetString(common.FlagVersion))
	arg.SetOlaresCDNService(viper.GetString(common.FlagCDNService))

	runtime, err := common.NewKubeRuntime(*arg)
	if err != nil {
		return err
	}

	if ok := utils.CheckUrl(arg.OlaresCDNService); !ok {
		return fmt.Errorf("invalid cdn service")
	}

	manifest := viper.GetString(common.FlagManifest)
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
