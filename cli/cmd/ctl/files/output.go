package files

import (
	"fmt"

	"github.com/spf13/cobra"
)

// Every other tree in this CLI spells the output format `-o/--output
// table|json` — market, cluster, settings, router, search, knowledge,
// dashboard, skills. This one shipped a `--json` boolean, and an agent
// that learned the common spelling on any of the others gets an unknown
// flag here, on the tree it is most likely to reach for first.
//
// Renaming outright is not available: a script that passes `--json` has
// to keep working. So both are registered and resolved into the single
// bool each command already reads, and `--json` is hidden from help so
// that what an agent reads there names one spelling.
//
// Hidden rather than pflag-deprecated, which would otherwise be the
// obvious choice: pflag prints its own deprecation line to the flag
// set's writer, and that writer is the command's stdout — the stream
// the caller just asked to be JSON. The notice is written here instead,
// to stderr.
const outputFormatUsage = "output format: table, json"

func addOutputFormatFlag(cmd *cobra.Command, asJSON *bool, legacyUsage string) {
	registerOutputFormatFlag(cmd, asJSON, legacyUsage, "o")
}

// addOutputFormatFlagLongOnly is for `files archive entries`, whose
// sibling `files archive cat -o <path>` writes an entry's bytes to a
// local file. Within one subtree a short flag that means a path in one
// verb and a format in the next is worse than not having it, so this
// half of the pair takes `--output` only.
func addOutputFormatFlagLongOnly(cmd *cobra.Command, asJSON *bool, legacyUsage string) {
	registerOutputFormatFlag(cmd, asJSON, legacyUsage, "")
}

func registerOutputFormatFlag(cmd *cobra.Command, asJSON *bool, legacyUsage, shorthand string) {
	var format string
	cmd.Flags().StringVarP(&format, "output", shorthand, "table", outputFormatUsage)
	cmd.Flags().BoolVar(asJSON, "json", false, legacyUsage)
	if err := cmd.Flags().MarkHidden("json"); err != nil {
		panic(err)
	}

	previous := cmd.PreRunE
	cmd.PreRunE = func(c *cobra.Command, args []string) error {
		if previous != nil {
			if err := previous(c, args); err != nil {
				return err
			}
		}
		if c.Flags().Changed("json") {
			fmt.Fprintf(c.ErrOrStderr(), "--json is deprecated; use -o json\n")
		}
		switch format {
		case "table":
		case "json":
			*asJSON = true
		default:
			return fmt.Errorf("unknown output format %q for %q: use table or json", format, c.CommandPath())
		}
		return nil
	}
}
