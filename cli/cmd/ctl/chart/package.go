package chart

import (
	"errors"
	"fmt"
	"os"

	chartpkg "github.com/beclab/Olares/cli/pkg/chart"
	"github.com/spf13/cobra"
)

func NewCmdChartPackage() *cobra.Command {
	var (
		output string
		force  bool
	)
	cmd := &cobra.Command{
		Use:   "package <chart-dir>",
		Short: "Package a chart directory into a .tgz for upload",
		Long: `Package an Olares chart directory into a <name>-<version>.tgz archive.

The archive name and version are read from the chart's Chart.yaml, and the
layout matches what 'helm package' produces — so the result is accepted as-is
by both 'olares-cli chart lint' and 'olares-cli market upload'. Non-standard
files such as OlaresManifest.yaml are preserved. Local-only; no Olares login
required.

An existing archive with the same name is not overwritten unless --force is
passed. The name comes from Chart.yaml, so repackaging without bumping the
version targets the file the previous run produced.

Examples:
  olares-cli chart package ./myapp
  olares-cli chart package ./myapp -o ./dist
  olares-cli chart package ./myapp --force`,
		Args: cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			out, err := chartpkg.Package(args[0], output, force)
			if err != nil {
				if errors.Is(err, chartpkg.ErrArchiveExists) {
					return fmt.Errorf("%w\nbump the chart version, pick another -o directory, or pass --force to overwrite", err)
				}
				return err
			}
			fmt.Fprintf(os.Stdout, "packaged chart: %s\n", out)
			return nil
		},
	}
	cmd.Flags().StringVarP(&output, "output", "o", ".", "directory to write the .tgz into")
	cmd.Flags().BoolVar(&force, "force", false, "overwrite an existing archive with the same name")
	return cmd
}
