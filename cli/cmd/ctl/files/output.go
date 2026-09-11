package files

import (
	"fmt"
	"strings"

	"github.com/spf13/cobra"

	"github.com/beclab/Olares/cli/pkg/cmdutil"
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
	var (
		format string
		legacy bool
	)
	cmd.Flags().StringVarP(&format, "output", shorthand, "table", outputFormatUsage)
	cmd.Flags().BoolVar(&legacy, "json", false, legacyUsage)
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
		wantJSON, err := resolveOutputFormat(c, format, legacy)
		if err != nil {
			return err
		}
		*asJSON = wantJSON
		return nil
	}
}

// resolveOutputFormat folds the two spellings into the one answer the
// command runs on.
//
// Both being present is the case worth handling deliberately. The first
// version let `--json` set the bool directly, so `--json -o table`
// rendered JSON while everything downstream that consulted `--output` —
// the entrypoint's error envelope among them — believed it was a table.
// Two flags naming the same setting must either agree or stop the
// command; silently picking one is how a caller ends up debugging output
// it never asked for.
//
// Nothing is printed about the legacy spelling. It fires precisely when
// the caller asked for JSON, so any nudge would land in the stream they
// asked to be machine-readable, and `--help` already advertises the one
// spelling worth learning.
func resolveOutputFormat(cmd *cobra.Command, format string, legacy bool) (bool, error) {
	var wantJSON bool
	switch {
	case cmdutil.OutputIsJSON(format):
		wantJSON = true
	case strings.EqualFold(strings.TrimSpace(format), "table"):
		wantJSON = false
	default:
		return false, fmt.Errorf("unknown output format %q for %q: use table or json", format, cmd.CommandPath())
	}

	if !cmd.Flags().Changed("json") {
		return wantJSON, nil
	}
	if cmd.Flags().Changed("output") && wantJSON != legacy {
		return false, fmt.Errorf("%q got --json=%t and --output %q, which disagree: pass one of them",
			cmd.CommandPath(), legacy, format)
	}
	return legacy, nil
}
