package os

import (
	"log"

	"github.com/beclab/Olares/cli/cmd/ctl/options"
	"github.com/beclab/Olares/cli/pkg/pipelines"
	"github.com/spf13/cobra"
)

type PrepareSystemOptions struct {
	PrepareOptions *options.CliPrepareSystemOptions
}

func NewPrepareSystemOptions() *PrepareSystemOptions {
	return &PrepareSystemOptions{
		PrepareOptions: options.NewCliPrepareSystemOptions(),
	}
}

func NewCmdPrepare() *cobra.Command {
	o := NewPrepareSystemOptions()
	cmd := &cobra.Command{
		Use:   "prepare [component1 component2 ...]",
		Short: "Prepare install",
		Run: func(cmd *cobra.Command, args []string) {
			if err := pipelines.PrepareSystemPipeline(o.PrepareOptions, args); err != nil {
				log.Fatalf("error: %v", err)
			}
		},
	}
	o.PrepareOptions.AddFlags(cmd)
	return cmd
}
