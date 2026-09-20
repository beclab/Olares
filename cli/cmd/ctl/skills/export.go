package skills

import (
	"fmt"
	"path/filepath"

	"github.com/spf13/cobra"

	skillsuite "github.com/beclab/Olares/cli/skills"
)

type exportEnvelope struct {
	Directory string   `json:"directory"`
	Skills    []string `json:"skills"`
	Count     int      `json:"count"`
}

func newExportCommand() *cobra.Command {
	opts := &outputOptions{Output: "table"}
	cmd := &cobra.Command{
		Use:   "export <dir>",
		Short: "Write the embedded suite to a directory",
		Long: `Write the suite this binary carries into <dir>, one directory per skill.

Nothing else is involved: no network, no Node, no agent discovery, no state.
That makes it the way to place the suite somewhere a program already looks —
a container's skills directory at boot, a project checkout, an image build —
and it is what ` + "`skills install`" + ` runs before handing the result to the
skills CLI.

Each skill is replaced whole, so running it again both updates the copy and
takes away references that no longer exist. A skill that is already a
symbolic link is refused rather than followed, because for the local
development layout that link is the checkout this binary was built from.`,
		Example: `  olares-cli skills export ./skills
  olares-cli skills export "$LARES_DATA_DIR/skills"
  olares-cli skills export ~/.agents/skills`,
		Args: cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			return runExport(opts, args[0])
		},
	}
	addOutputFlags(cmd, opts)
	return cmd
}

func runExport(opts *outputOptions, dir string) error {
	absolute, err := filepath.Abs(dir)
	if err != nil {
		absolute = dir
	}
	written, err := skillsuite.Export(absolute)
	if err != nil {
		return err
	}
	if opts.isJSON() {
		return opts.printJSON(exportEnvelope{Directory: absolute, Skills: written, Count: len(written)})
	}
	if opts.Quiet {
		return nil
	}
	fmt.Printf("wrote %d skills to %s\n", len(written), absolute)
	return nil
}
