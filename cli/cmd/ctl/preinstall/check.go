package preinstall

import (
	"fmt"

	"github.com/beclab/Olares/cli/pkg/manifest"
	pkgpreinstall "github.com/beclab/Olares/cli/pkg/preinstall"
	"github.com/spf13/cobra"
)

func NewCmdPreinstallCheck() *cobra.Command {
	var (
		full           bool
		installationMF string
		imagesDir      string
		gpuModes       []string
	)
	cmd := &cobra.Command{
		Use:   "check <path>",
		Short: "Validate a static Market preinstall bundle directory",
		Long: `Validate that <path> is a well-formed static Market preinstall
bundle directory (the directory that contains bundle.json, typically
preinstall/market).

By default the command runs contract-level checks that mirror the
installer prepare path:

  - decode and semantically validate bundle.json
  - verify each chart exists, is a regular file within size limits,
    and matches chartSha256
  - validate unique Hugging Face cache targets
  - load and validate each artifact manifest

Pass --full to also verify artifact payload trees under each artifact's
source directory against the manifest (entry types and digests). This can
be slow for large model payloads.

Pass --installation-manifest to additionally render every bundled chart
and require the medium to preload the images they need. This catches the
case where a bundled chart was updated but the image list was not, which
on an offline device means the app cannot start at all. Add --images-dir
to also require the image payload to be present, which is what prepare
reads. Both checks are offline; neither contacts a registry or a CDN.

Examples:
  olares-cli preinstall check ./preinstall/market
  olares-cli preinstall check ./preinstall/market --full
  olares-cli preinstall check ./preinstall/market \
    --installation-manifest ./installation.manifest --images-dir ./images`,
		Args: cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			opts := pkgpreinstall.CheckOptions{
				Full:      full,
				ImagesDir: imagesDir,
				GPUModes:  gpuModes,
			}
			if installationMF != "" {
				installation, err := manifest.ReadAll(installationMF)
				if err != nil {
					return fmt.Errorf("read installation manifest %q: %w", installationMF, err)
				}
				opts.InstallationManifest = installation
			} else if imagesDir != "" {
				return fmt.Errorf("--images-dir requires --installation-manifest")
			}
			if err := pkgpreinstall.CheckStaticBundle(args[0], opts); err != nil {
				return err
			}
			fmt.Fprintln(cmd.OutOrStdout(), "ok")
			return nil
		},
	}
	fs := cmd.Flags()
	fs.BoolVar(&full, "full", false, "also verify artifact payload trees against manifests")
	fs.StringVar(&installationMF, "installation-manifest", "", "path to installation.manifest; enables the bundled-chart image check")
	fs.StringVar(&imagesDir, "images-dir", "", "directory holding preloaded image payloads (normally <baseDir>/images); requires --installation-manifest")
	fs.StringSliceVar(&gpuModes, "gpu-modes", nil, "GPU families to render charts under (default: no-override plus all known families)")
	return cmd
}
