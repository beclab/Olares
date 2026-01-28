package os

import (
	"log"

	"github.com/beclab/Olares/cli/cmd/config"
	"github.com/beclab/Olares/cli/pkg/common"
	"github.com/beclab/Olares/cli/pkg/pipelines"
	"github.com/spf13/cobra"
)

func NewCmdChangeIP() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "change-ip",
		Short: "change The IP address of Olares OS",
		Run: func(cmd *cobra.Command, args []string) {
			if err := pipelines.ChangeIPPipeline(); err != nil {
				log.Fatalf("error: %v", err)
			}
		},
	}

	// todo: merge master host config with release info
	// be backward compatible with old version and olaresd

	flagSetter := config.NewFlagSetterFor(cmd)
	config.AddVersionFlagBy(flagSetter)
	config.AddBaseDirFlagBy(flagSetter)
	config.AddMasterHostFlagsBy(flagSetter)
	flagSetter.Add(common.FlagWSLDistribution,
		"d",
		"",
		"Set WSL distribution name, only on Windows platform, defaults to "+common.WSLDefaultDistribution,
	).WithAlias(common.FlagLegacyWSLDistribution)
	config.AddMiniKubeProfileFlagBy(flagSetter)

	return cmd
}
