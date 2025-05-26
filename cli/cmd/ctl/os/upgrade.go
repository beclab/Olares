package os

import (
	"log"

	"bytetrade.io/web3os/installer/cmd/ctl/options"
	"bytetrade.io/web3os/installer/pkg/pipelines"
	"github.com/spf13/cobra"
)

type UpgradeOsOptions struct {
	UpgradeOptions *options.UpgradeOptions
}

func NewUpgradeOsOptions() *UpgradeOsOptions {
	return &UpgradeOsOptions{
		UpgradeOptions: options.NewUpgradeOptions(),
	}
}

func NewCmdUpgradeOs() *cobra.Command {
	o := NewUpgradeOsOptions()
	cmd := &cobra.Command{
		Use:   "upgrade",
		Short: "Upgrade Olares to a newer version",
		Run: func(cmd *cobra.Command, args []string) {
			if err := pipelines.UpgradeOlaresPipeline(o.UpgradeOptions); err != nil {
				log.Fatalf("error: %v", err)
			}
		},
	}
	o.UpgradeOptions.AddFlags(cmd)
	cmd.AddCommand(NewCmdUpgradePrecheck())
	return cmd
}

func NewCmdUpgradePrecheck() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "precheck",
		Short: "Precheck Olares for Upgrade",
		Run: func(cmd *cobra.Command, args []string) {
			if err := pipelines.UpgradePreCheckPipeline(); err != nil {
				log.Fatalf("error: %v", err)
			}
		},
	}
	return cmd
}
