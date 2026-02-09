package os

import (
	"fmt"
	"os"
	"os/user"
	"path/filepath"
	"strings"
	"time"

	"github.com/beclab/Olares/cli/cmd/config"
	"github.com/beclab/Olares/cli/pkg/common"
	corecommon "github.com/beclab/Olares/cli/pkg/core/common"
	"github.com/beclab/Olares/cli/pkg/core/util"
	"github.com/beclab/Olares/cli/pkg/release/builder"

	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

func NewCmdRelease() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "release",
		Short: "Build release based on a local Olares repository",
		Run: func(cmd *cobra.Command, args []string) {
			baseDir := viper.GetString(common.FlagBaseDir)
			version := viper.GetString(common.FlagVersion)
			cdn := viper.GetString(common.FlagCDNService)
			ignoreMissingImages := viper.GetBool(common.FlagIgnoreMissingImages)
			extract := viper.GetBool(common.FlagExtract)

			cwd, err := os.Getwd()
			if err != nil {
				fmt.Printf("failed to get current working directory: %s\n", err)
				os.Exit(1)
			}
			if !strings.HasPrefix(strings.ToLower(filepath.Base(cwd)), "olares") {
				fmt.Println("error: please run release command under the root path of Olares repo")
				os.Exit(1)
			}
			if baseDir == "" {
				usr, err := user.Current()
				if err != nil {
					fmt.Printf("failed to get current user: %s\n", err)
					os.Exit(1)
				}
				baseDir = filepath.Join(usr.HomeDir, corecommon.DefaultBaseDir)
				fmt.Printf("--base-dir unspecified, using: %s\n", baseDir)
				time.Sleep(1 * time.Second)
			}

			if version == "" {
				version = fmt.Sprintf("1.12.6-%s", time.Now().Format("20060102150405"))
				fmt.Printf("--version unspecified, using: %s\n", version)
				time.Sleep(1 * time.Second)
			}

			wizardFile, err := builder.NewBuilder(cwd, version, cdn, ignoreMissingImages).Build()
			if err != nil {
				fmt.Printf("failed to build release: %s\n", err)
				os.Exit(1)
			}
			fmt.Printf("\nsuccessfully built release\nversion: %s\n package: %s\n", version, wizardFile)
			if extract {
				dest := filepath.Join(baseDir, "versions", "v"+version)
				if err := os.MkdirAll(dest, 0755); err != nil {
					fmt.Printf("Failed to create new version directory for this release: %s\n", err)
					os.Exit(1)
				}
				if err := util.Untar(wizardFile, dest); err != nil {
					fmt.Printf("failed to extract release package: %s\n", err)
					os.Exit(1)
				}
				fmt.Printf("\nrelease package is extracted to: %s\n", dest)
			}
		},
	}

	flagSetter := config.NewFlagSetterFor(cmd)
	config.AddBaseDirFlagBy(flagSetter)
	config.AddVersionFlagBy(flagSetter)
	config.AddCDNServiceFlagBy(flagSetter)
	flagSetter.Add(common.FlagIgnoreMissingImages,
		"",
		true,
		"ignore missing images when downloading checksums from CDN, only disable this if no new image is added, or the build may fail because the image is not uploaded to the CDN yet",
	)
	flagSetter.Add(common.FlagExtract,
		"e",
		true,
		"extract this release to --base-dir after build, this can be disabled if only the release file itself is needed",
	)

	return cmd
}
