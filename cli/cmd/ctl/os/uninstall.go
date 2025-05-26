package os

import (
	"bytetrade.io/web3os/installer/cmd/ctl/options"
	"bytetrade.io/web3os/installer/pkg/pipelines"
	"github.com/spf13/cobra"
	"log"
)

type UninstallOsOptions struct {
	UninstallOptions *options.CliTerminusUninstallOptions
}

func NewUninstallOsOptions() *UninstallOsOptions {
	return &UninstallOsOptions{
		UninstallOptions: options.NewCliTerminusUninstallOptions(),
	}
}

func NewCmdUninstallOs() *cobra.Command {
	o := NewUninstallOsOptions()
	cmd := &cobra.Command{
		Use:   "uninstall",
		Short: "Uninstall Olares",
		Run: func(cmd *cobra.Command, args []string) {
			err := pipelines.UninstallTerminusPipeline(o.UninstallOptions)
			if err != nil {
				log.Fatalf("error: %v", err)
			}
		},
	}
	o.UninstallOptions.AddFlags(cmd)
	return cmd
}
