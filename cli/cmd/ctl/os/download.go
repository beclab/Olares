package os

import (
	"bytetrade.io/web3os/installer/cmd/ctl/options"
	"bytetrade.io/web3os/installer/pkg/pipelines"
	"github.com/spf13/cobra"
	"log"
)

func NewCmdRootDownload() *cobra.Command {
	rootDownloadCmd := &cobra.Command{
		Use:   "download",
		Short: "Download the packages and components needed to install Olares",
	}

	rootDownloadCmd.AddCommand(NewCmdCheckDownload())
	rootDownloadCmd.AddCommand(NewCmdDownload())
	rootDownloadCmd.AddCommand(NewCmdDownloadWizard())

	return rootDownloadCmd
}

func NewCmdDownload() *cobra.Command {
	o := options.NewCliDownloadOptions()
	cmd := &cobra.Command{
		Use:   "component",
		Short: "Download the packages and components needed to install Olares",
		Run: func(cmd *cobra.Command, args []string) {

			if err := pipelines.DownloadInstallationPackage(o); err != nil {
				log.Fatalf("error: %v", err)
			}
		},
	}

	o.AddFlags(cmd)
	return cmd
}

func NewCmdDownloadWizard() *cobra.Command {
	o := options.NewCliDownloadWizardOptions()
	cmd := &cobra.Command{
		Use:   "wizard",
		Short: "Download the Olares installation wizard",
		Run: func(cmd *cobra.Command, args []string) {

			if err := pipelines.DownloadInstallationWizard(o); err != nil {
				log.Fatalf("error: %v", err)
			}
		},
	}

	o.AddFlags(cmd)
	return cmd
}

func NewCmdCheckDownload() *cobra.Command {
	o := options.NewCliDownloadOptions()
	cmd := &cobra.Command{
		Use:   "check",
		Short: "Check Downloaded Olares Installation Package",
		Run: func(cmd *cobra.Command, args []string) {

			if err := pipelines.CheckDownloadInstallationPackage(o); err != nil {
				log.Fatalf("error: %v", err)
			}
		},
	}

	o.AddFlags(cmd)
	return cmd
}
