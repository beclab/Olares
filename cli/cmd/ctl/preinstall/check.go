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
		olaresYAML     string
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
reads.

Pass --olares-yaml instead of --installation-manifest to run the same
chart-image check against output.containers in a source-tree Olares.yaml,
without packing an installer first. The two flags are mutually exclusive.
--images-dir is a packed-medium check and requires --installation-manifest.

All image checks are offline; none contact a registry or a CDN.

Examples:
  olares-cli preinstall check ./preinstall/market
  olares-cli preinstall check ./preinstall/market --full
  olares-cli preinstall check ./preinstall/market \
    --olares-yaml ./path/to/Olares.yaml
  olares-cli preinstall check ./preinstall/market \
    --installation-manifest ./installation.manifest --images-dir ./images`,
		Args: cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			if installationMF != "" && olaresYAML != "" {
				return fmt.Errorf("--installation-manifest and --olares-yaml are mutually exclusive")
			}
			if imagesDir != "" && installationMF == "" {
				return fmt.Errorf("--images-dir requires --installation-manifest")
			}
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
			}
			if olaresYAML != "" {
				images, err := pkgpreinstall.ReadOlaresYAMLContainers(olaresYAML)
				if err != nil {
					return fmt.Errorf("read olares.yaml %q: %w", olaresYAML, err)
				}
				opts.OlaresImages = images
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
	fs.StringVar(&installationMF, "installation-manifest", "", "path to installation.manifest; enables the bundled-chart image check against the packed medium")
	fs.StringVar(&olaresYAML, "olares-yaml", "", "path to a source-tree Olares.yaml; enables the bundled-chart image check against output.containers")
	fs.StringVar(&imagesDir, "images-dir", "", "directory holding preloaded image payloads (normally <baseDir>/images); requires --installation-manifest")
	fs.StringSliceVar(&gpuModes, "gpu-modes", nil, "GPU families to render charts under (default: no-override plus all known families)")
	return cmd
}
