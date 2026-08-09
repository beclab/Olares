package dev

import (
	"fmt"
	"os"
	"text/tabwriter"

	"github.com/spf13/cobra"

	"github.com/beclab/Olares/cli/pkg/devdeploy"
)

// NewComponentsCommand: `olares-cli dev components [list|get|validate]`.
//
// This is the CLI half of the repo's `make dev-*` targets. Make cannot
// parse YAML without pulling in yq, and the repo already depends on this
// binary, so the map has exactly one parser and the Makefile stays a
// thin wrapper.
func NewComponentsCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "components",
		Short: "Inspect the component -> image build map (build/dev-components.yaml)",
		Long: `Read ` + devdeploy.ComponentsFile + `, the map from a component name to
its build context, Dockerfile, and released image repository.

Only useful inside an Olares checkout: the map is a repo file, not
something the CLI carries. The map mirrors the
.github/workflows/module_*_publish_*.yaml files, which remain the source
of truth for what actually gets published.
`,
		RunE: func(c *cobra.Command, args []string) error {
			if len(args) > 0 {
				return fmt.Errorf("unknown command %q for %q", args[0], c.CommandPath())
			}
			return c.Help()
		},
	}
	cmd.SilenceUsage = true
	cmd.AddCommand(newComponentsListCommand())
	cmd.AddCommand(newComponentsGetCommand())
	cmd.AddCommand(newComponentsValidateCommand())
	return cmd
}

// loadComponents resolves the checkout and parses the map.
func loadComponents(repoFlag string) (string, map[string]devdeploy.Component, error) {
	start := repoFlag
	if start == "" {
		cwd, err := os.Getwd()
		if err != nil {
			return "", nil, err
		}
		start = cwd
	}
	root, err := devdeploy.FindRepoRoot(start)
	if err != nil {
		return "", nil, err
	}
	components, err := devdeploy.LoadComponents(root)
	if err != nil {
		return "", nil, err
	}
	return root, components, nil
}

func addRepoFlag(cmd *cobra.Command, repo *string) {
	cmd.Flags().StringVar(repo, "repo", "",
		"path inside the Olares checkout (default: current directory)")
}

func newComponentsListCommand() *cobra.Command {
	var repo string
	cmd := &cobra.Command{
		Use:   "list",
		Short: "list every component that can be dev-built",
		RunE: func(c *cobra.Command, args []string) error {
			_, components, err := loadComponents(repo)
			if err != nil {
				return err
			}
			tw := tabwriter.NewWriter(os.Stdout, 0, 0, 2, ' ', 0)
			defer tw.Flush()
			fmt.Fprintln(tw, "COMPONENT\tIMAGE\tDOCKERFILE")
			for _, name := range devdeploy.ComponentNames(components) {
				c := components[name]
				fmt.Fprintf(tw, "%s\t%s\t%s\n", name, c.Image, c.Dockerfile)
			}
			return nil
		},
	}
	addRepoFlag(cmd, &repo)
	return cmd
}

func newComponentsGetCommand() *cobra.Command {
	var (
		repo   string
		format string
	)
	cmd := &cobra.Command{
		Use:   "get <component>",
		Short: "print one component's build coordinates",
		Long: `Print a component's image, context and Dockerfile.

--format shell emits KEY=value lines meant to be eval'd from a Make
recipe or shell script:

  eval "$(olares-cli dev components get app-service --format shell)"
  docker build -t "$IMAGE:dev" -f "$DOCKERFILE" "$CONTEXT"
`,
		Args: cobra.ExactArgs(1),
		RunE: func(c *cobra.Command, args []string) error {
			_, components, err := loadComponents(repo)
			if err != nil {
				return err
			}
			comp, err := devdeploy.Lookup(components, args[0])
			if err != nil {
				return err
			}
			switch format {
			case "shell":
				fmt.Printf("COMPONENT=%s\n", comp.Name)
				fmt.Printf("IMAGE=%s\n", comp.Image)
				fmt.Printf("CONTEXT=%s\n", comp.Context)
				fmt.Printf("DOCKERFILE=%s\n", comp.Dockerfile)
				return nil
			case "", "table":
				tw := tabwriter.NewWriter(os.Stdout, 0, 0, 2, ' ', 0)
				defer tw.Flush()
				fmt.Fprintf(tw, "Component:\t%s\n", comp.Name)
				fmt.Fprintf(tw, "Image:\t%s\n", comp.Image)
				fmt.Fprintf(tw, "Context:\t%s\n", comp.Context)
				fmt.Fprintf(tw, "Dockerfile:\t%s\n", comp.Dockerfile)
				fmt.Fprintf(tw, "Workflow:\t%s\n", comp.Workflow)
				return nil
			default:
				return fmt.Errorf("unknown --format %q (want table or shell)", format)
			}
		},
	}
	addRepoFlag(cmd, &repo)
	cmd.Flags().StringVar(&format, "format", "table", "output format: table | shell")
	return cmd
}

func newComponentsValidateCommand() *cobra.Command {
	var repo string
	cmd := &cobra.Command{
		Use:   "validate",
		Short: "check that every mapped context and Dockerfile still exists",
		RunE: func(c *cobra.Command, args []string) error {
			root, components, err := loadComponents(repo)
			if err != nil {
				return err
			}
			if err := devdeploy.Validate(root, components); err != nil {
				return err
			}
			fmt.Fprintf(os.Stdout, "%s: %d components, all paths resolve\n",
				devdeploy.ComponentsFile, len(components))
			return nil
		},
	}
	addRepoFlag(cmd, &repo)
	return cmd
}
