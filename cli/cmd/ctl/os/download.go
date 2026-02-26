package os

import (
	"log"

	"github.com/beclab/Olares/cli/cmd/config"
	"github.com/beclab/Olares/cli/pkg/common"
	"github.com/beclab/Olares/cli/pkg/pipelines"
	"github.com/spf13/cobra"
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
	cmd := &cobra.Command{
		Use:   "component",
		Short: "Download the packages and components needed to install Olares",
		Run: func(cmd *cobra.Command, args []string) {

			if err := pipelines.DownloadInstallationPackage(); err != nil {
				log.Fatalf("error: %v", err)
			}
		},
	}
	flagSetter := config.NewFlagSetterFor(cmd)
	config.AddVersionFlagBy(flagSetter)
	config.AddBaseDirFlagBy(flagSetter)
	config.AddCDNServiceFlagBy(flagSetter)
	config.AddManifestFlagBy(flagSetter)

	return cmd
}

func NewCmdDownloadWizard() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "wizard",
		Short: "Download the Olares installation wizard",
		Run: func(cmd *cobra.Command, args []string) {

			if err := pipelines.DownloadInstallationWizard(); err != nil {
				log.Fatalf("error: %v", err)
			}
		},
	}
	flagSetter := config.NewFlagSetterFor(cmd)

	flagSetter.Add(common.FlagReleaseID, "", "", "Set the specific release id of the release version")
	flagSetter.Add(common.FlagURLOverride, "", "", "Set another URL for wizard download explicitly")

	config.AddVersionFlagBy(flagSetter)
	config.AddBaseDirFlagBy(flagSetter)
	config.AddCDNServiceFlagBy(flagSetter)
	return cmd
}

func NewCmdCheckDownload() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "check",
		Short: "Check Downloaded Olares Installation Package",
		Run: func(cmd *cobra.Command, args []string) {

			if err := pipelines.CheckDownloadInstallationPackage(); err != nil {
				log.Fatalf("error: %v", err)
			}
		},
	}
	flagSetter := config.NewFlagSetterFor(cmd)
	config.AddVersionFlagBy(flagSetter)
	config.AddBaseDirFlagBy(flagSetter)
	config.AddManifestFlagBy(flagSetter)
	return cmd
}
