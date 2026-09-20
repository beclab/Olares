package cmdutil

import (
	"strings"

	"github.com/spf13/cobra"
)

// Every tree renders `-o json` on a case-insensitive, space-trimmed
// comparison -- market's IsJSON, cluster's IsJSON, router's parseFormat
// all spell it the same way, arrived at independently. The entrypoint
// then answered a failure by comparing the raw string, so `-o JSON`
// succeeded as JSON and failed as prose. That is the kind of split a
// caller finds only in production, because it needs a failure to show.
//
// So the rule lives here once, and the entrypoint asks rather than
// re-deriving.
func OutputIsJSON(value string) bool {
	return strings.EqualFold(strings.TrimSpace(value), "json")
}

// AskedForJSON reports whether the invocation wants to be read by a
// program rather than a person.
//
// Two spellings, because the files tree kept a `--json` boolean for the
// scripts written against it. A caller who passes it has said the same
// thing as `-o json` and gets the same treatment; anything that reads
// only `--output` misses them, which is how the legacy spelling ended
// up excluded from the error envelope it should have had.
func AskedForJSON(cmd *cobra.Command) bool {
	if cmd == nil {
		return false
	}
	if flag := cmd.Flags().Lookup("output"); flag != nil && OutputIsJSON(flag.Value.String()) {
		return true
	}
	flag := cmd.Flags().Lookup("json")
	return flag != nil && flag.Changed && flag.Value.String() == "true"
}
