package os

import (
	"log"

	"github.com/beclab/Olares/cli/cmd/config"
	"github.com/beclab/Olares/cli/pkg/common"
	"github.com/beclab/Olares/cli/pkg/pipelines"
	"github.com/spf13/cobra"
)

func NewCmdPrepare() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "prepare [component1 component2 ...]",
		Short: "Prepare install",
		Run: func(cmd *cobra.Command, args []string) {
			if err := pipelines.PrepareSystemPipeline(args); err != nil {
				log.Fatalf("error: %v", err)
			}
		},
	}
	flagSetter := config.NewFlagSetterFor(cmd)

	flagSetter.Add(common.FlagRegistryMirrors,
		"r",
		"",
		"Extra Docker Container registry mirrors, multiple mirrors are separated by commas",
	)

	config.AddVersionFlagBy(flagSetter)
	config.AddBaseDirFlagBy(flagSetter)
	config.AddStorageFlagsBy(flagSetter)
	config.AddKubeTypeFlagBy(flagSetter)
	config.AddMiniKubeProfileFlagBy(flagSetter)
	return cmd
}
