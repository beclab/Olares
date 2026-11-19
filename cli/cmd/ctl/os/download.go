package os

import (
	"fmt"
	"log"
	"os"
	"os/user"
	"path/filepath"
	"runtime"
	"strings"

	"github.com/beclab/Olares/cli/cmd/config"
	"github.com/beclab/Olares/cli/pkg/bootstrap/download"
	"github.com/beclab/Olares/cli/pkg/common"
	cc "github.com/beclab/Olares/cli/pkg/core/common"
	"github.com/beclab/Olares/cli/pkg/pipelines"
	"github.com/shirou/gopsutil/v4/host"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

func NewCmdRootDownload() *cobra.Command {
	rootDownloadCmd := &cobra.Command{
		Use:   "download",
		Short: "Download the packages and components needed to install Olares",
	}

	rootDownloadCmd.AddCommand(NewCmdCheckDownload())
	rootDownloadCmd.AddCommand(NewCmdDownload())
	rootDownloadCmd.AddCommand(NewCmdDownloadWizard())
	rootDownloadCmd.AddCommand(NewCmdCleanDownload())

	return rootDownloadCmd
}

func NewCmdCleanDownload() *cobra.Command {
	var dryRun bool
	cmd := &cobra.Command{
		Use:   "clean",
		Short: "Remove cached downloads no longer needed by the current Olares version",
		Long: "Remove cached packages and image archives absent from the selected version's installation manifest. " +
			"When no base directory is configured, use .olares under the invoking user's home, including when run with sudo.",
		Args: cobra.NoArgs,
		RunE: func(cmd *cobra.Command, args []string) error {
			arg := common.Argument{Environment: os.Environ()}
			arg.SetOlaresVersion(strings.TrimSpace(viper.GetString(common.FlagVersion)))
			baseDir := viper.GetString(common.FlagBaseDir)
			if baseDir == "" {
				homeDir, err := cleanDownloadHomeDir()
				if err != nil {
					return fmt.Errorf("resolve the default Olares base directory (use --base-dir to override): %w", err)
				}
				baseDir = filepath.Join(homeDir, cc.DefaultBaseDir)
			}

			cacheDir := baseDir
			if runtime.GOOS == cc.Linux && os.Getenv(common.ENV_CONTAINER_MODE) != "oic" {
				kernel, _ := host.KernelVersionWithContext(cmd.Context())
				if userPath := arg.GetWslUserPath(); userPath != "" && strings.Contains(kernel, "-WSL") {
					cacheDir = filepath.Join(userPath, cc.DefaultBaseDir)
				}
			}
			return download.CleanCache(cmd.Context(), baseDir, cacheDir, arg.OlaresVersion, dryRun, cmd.OutOrStdout())
		},
	}
	flagSetter := config.NewFlagSetterFor(cmd)
	config.AddVersionFlagBy(flagSetter)
	config.AddBaseDirFlagBy(flagSetter)
	cmd.Flags().BoolVar(&dryRun, "dry-run", false, "List the files that would be removed")
	return cmd
}

func cleanDownloadHomeDir() (string, error) {
	sudoUID := strings.TrimSpace(os.Getenv("SUDO_UID"))
	sudoUser := strings.TrimSpace(os.Getenv("SUDO_USER"))
	if sudoUID == "" && sudoUser == "" {
		return os.UserHomeDir()
	}

	var account *user.User
	var err error
	if sudoUID != "" {
		account, err = user.LookupId(sudoUID)
	} else {
		account, err = user.Lookup(sudoUser)
	}
	if err != nil {
		return "", fmt.Errorf("look up the user who invoked sudo: %w", err)
	}
	if !filepath.IsAbs(account.HomeDir) {
		return "", fmt.Errorf("user %q has no absolute home directory", account.Username)
	}
	return account.HomeDir, nil
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
