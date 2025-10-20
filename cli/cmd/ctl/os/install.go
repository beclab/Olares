package os

import (
	"log"

	"github.com/beclab/Olares/cli/cmd/ctl/options"
	"github.com/beclab/Olares/cli/pkg/pipelines"
	"github.com/spf13/cobra"
)

type InstallOsOptions struct {
	InstallOptions *options.CliTerminusInstallOptions
}

func NewInstallOsOptions() *InstallOsOptions {
	return &InstallOsOptions{
		InstallOptions: options.NewCliTerminusInstallOptions(),
	}
}

func NewCmdInstallOs() *cobra.Command {
	o := NewInstallOsOptions()
	cmd := &cobra.Command{
		Use:   "install",
		Short: "Install Olares",
		Run: func(cmd *cobra.Command, args []string) {
			if !cmd.Flags().Changed("enable-reverse-proxy") {
				o.InstallOptions.EnableReverseProxy = nil
			}
			if err := pipelines.CliInstallTerminusPipeline(o.InstallOptions); err != nil {
				log.Fatalf("error: %v", err)
			}
		},
	}
	o.InstallOptions.AddFlags(cmd)
	cmd.AddCommand(NewCmdInstallStorage())
	return cmd
}
