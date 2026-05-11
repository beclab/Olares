package pipelines

import (
	"context"
	"fmt"

	"github.com/beclab/Olares/cli/pkg/common"
	"github.com/beclab/Olares/cli/pkg/core/logger"
	"github.com/beclab/Olares/cli/pkg/phase/download"
	"github.com/beclab/Olares/cli/pkg/utils"
	"github.com/spf13/viper"
)

func DownloadInstallationWizard() error {
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

	p := download.NewDownloadWizard(runtime, viper.GetString(common.FlagURLOverride), viper.GetString(common.FlagReleaseID))
	// TODO(ctx): plumb ctx in a follow-up; this entry point is not yet ctx-aware.
	if err := p.Start(context.Background()); err != nil {
		logger.Errorf("download wizard failed %v", err)
		return err
	}

	return nil
}
