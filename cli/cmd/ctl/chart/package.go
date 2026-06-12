package chart

import (
	"fmt"
	"os"

	chartpkg "github.com/beclab/Olares/cli/pkg/chart"
	"github.com/spf13/cobra"
)

func NewCmdChartPackage() *cobra.Command {
	var output string
	cmd := &cobra.Command{
		Use:   "package <chart-dir>",
		Short: "Package a chart directory into a .tgz for upload",
		Long: `Package an Olares chart directory into a <name>-<version>.tgz archive.

The archive name and version are read from the chart's Chart.yaml, and the
layout matches what 'helm package' produces — so the result is accepted as-is
by both 'olares-cli chart lint' and 'olares-cli market upload'. Non-standard
files such as OlaresManifest.yaml are preserved. Local-only; no Olares login
required.

Examples:
  olares-cli chart package ./myapp
  olares-cli chart package ./myapp -o ./dist`,
		Args: cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			out, err := chartpkg.Package(args[0], output)
			if err != nil {
				return err
			}
			fmt.Fprintf(os.Stdout, "packaged chart: %s\n", out)
			return nil
		},
	}
	cmd.Flags().StringVarP(&output, "output", "o", ".", "directory to write the .tgz into")
	return cmd
}
