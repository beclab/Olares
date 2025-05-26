package os

import (
	"bytetrade.io/web3os/installer/pkg/core/common"
	"bytetrade.io/web3os/installer/pkg/core/util"
	"bytetrade.io/web3os/installer/pkg/release/builder"
	"fmt"
	"os"
	"os/user"
	"path/filepath"
	"strings"
	"time"

	"github.com/spf13/cobra"
)

func NewCmdRelease() *cobra.Command {
	var (
		baseDir             string
		version             string
		cdn                 string
		ignoreMissingImages bool
		extract             bool
	)

	cmd := &cobra.Command{
		Use:   "release",
		Short: "Build release based on a local Olares repository",
		Run: func(cmd *cobra.Command, args []string) {
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
				baseDir = filepath.Join(usr.HomeDir, common.DefaultBaseDir)
				fmt.Printf("--base-dir unspecified, using: %s\n", baseDir)
				time.Sleep(1 * time.Second)
			}

			if version == "" {
				version = fmt.Sprintf("0.0.0-local-dev-%s", time.Now().Format("20060102150405"))
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

	cmd.Flags().StringVarP(&baseDir, "base-dir", "b", "", "base directory of Olares, where this release will be extracted to as a new version if --extract/-e is not disabled, defaults to $HOME/"+common.DefaultBaseDir)
	cmd.Flags().StringVarP(&version, "version", "v", "", "version of this release, defaults to 0.0.0-local-dev-{yyyymmddhhmmss}")
	cmd.Flags().StringVar(&cdn, "download-cdn-url", common.DownloadUrl, "CDN used for downloading checksums of dependencies and images")
	cmd.Flags().BoolVar(&ignoreMissingImages, "ignore-missing-images", true, "ignore missing images when downloading cheksums from CDN, only disable this if no new image is added, or the build may fail because the image is not uploaded to the CDN yet")
	cmd.Flags().BoolVarP(&extract, "extract", "e", true, "extract this release to --base-dir after build, this can be disabled if only the release file itself is needed")

	return cmd
}
