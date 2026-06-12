package chart

import (
	"fmt"
	"os"
	"path/filepath"

	chartpkg "github.com/beclab/Olares/cli/pkg/chart"
	"github.com/spf13/cobra"
)

type fromComposeOpts struct {
	Files     []string
	Output    string
	Name      string
	Title     string
	Type      string
	NewSchema bool
}

func NewCmdChartFromCompose() *cobra.Command {
	o := &fromComposeOpts{}
	cmd := &cobra.Command{
		Use:     "from-compose --name <name> -f <docker-compose.yml>",
		Aliases: []string{"init"},
		Short:   "Scaffold an Olares chart from docker-compose file(s)",
		Long: `Convert one or more docker-compose files into an Olares chart skeleton.

Under the hood this runs the same kompose conversion the Olares Studio /
devbox uses, then writes an Olares chart layout into the output directory:

  <output>/
  ├── Chart.yaml
  ├── OlaresManifest.yaml
  ├── values.yaml
  └── templates/
      └── <kind>-<name>.yaml   # one file per converted resource

The result is a STARTING POINT, not a finished app. The generated manifest
deliberately leaves four areas for you to refine before publishing:

  1. Metadata   - title / icon / description / categories / developer info
  2. Storage    - map compose volumes onto .Values.userspace.appData/appCache
                  /userData and the matching permission block
  3. Middleware - replace bundled postgres/redis/mongo/... services with the
                  system middleware block + options.dependencies
  4. Entrances  - one entrance is auto-detected; add the rest and tune
                  host/port/authLevel, plus ports[] for non-HTTP services

Validate your edits at any time with:

  olares-cli chart lint <output>

Examples:
  olares-cli chart from-compose --name myapp -f docker-compose.yml
  olares-cli chart from-compose --name myapp -f compose.yml -o ./charts/myapp --title "My App"
  olares-cli chart from-compose --name myapp -f a.yml -f b.yml --new-schema`,
		Args: cobra.NoArgs,
		RunE: func(cmd *cobra.Command, args []string) error {
			return runFromCompose(o)
		},
	}
	fs := cmd.Flags()
	fs.StringArrayVarP(&o.Files, "file", "f", nil, "docker-compose file path (repeatable)")
	fs.StringVarP(&o.Output, "output", "o", "", "chart output directory (default ./<name>)")
	fs.StringVar(&o.Name, "name", "", "Olares app name (lowercase alphanumeric, required)")
	fs.StringVar(&o.Title, "title", "", "human-facing app title (default: name)")
	fs.StringVar(&o.Type, "type", "app", "OlaresManifest type: app | recommend | middleware")
	fs.BoolVar(&o.NewSchema, "new-schema", false, "emit the 0.12.0 schema (spec.accelerator) instead of legacy 0.8.0")
	_ = cmd.MarkFlagRequired("name")
	return cmd
}

func runFromCompose(o *fromComposeOpts) error {
	if len(o.Files) == 0 {
		return fmt.Errorf("at least one --file/-f docker-compose path is required")
	}
	output := o.Output
	if output == "" {
		output = "./" + o.Name
	}
	if err := chartpkg.FromCompose(chartpkg.Options{
		ComposeFiles: o.Files,
		OutputDir:    output,
		Name:         o.Name,
		Title:        o.Title,
		Type:         o.Type,
		NewSchema:    o.NewSchema,
	}); err != nil {
		return err
	}

	abs, _ := filepath.Abs(output)
	fmt.Fprintf(os.Stdout, "scaffolded Olares chart at %s\n", abs)
	fmt.Fprintf(os.Stdout, "next: refine metadata / storage / middleware / entrances, then run `olares-cli chart lint %s`\n", output)
	return nil
}
