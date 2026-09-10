package files

import (
	"bytes"
	"strings"
	"testing"

	"github.com/spf13/cobra"
)

// The point of the pair is that an agent which learned `-o json` on any
// other tree gets the same answer here, and a script written against
// `--json` keeps working. Both halves have to hold at once.
func TestOutputFormatFlagAcceptsBothSpellings(t *testing.T) {
	for _, tc := range []struct {
		name string
		args []string
		want bool
	}{
		{name: "default is a table", args: nil, want: false},
		{name: "the spelling every other tree uses", args: []string{"-o", "json"}, want: true},
		{name: "the long form", args: []string{"--output", "json"}, want: true},
		{name: "the legacy boolean still runs", args: []string{"--json"}, want: true},
		{name: "an explicit table", args: []string{"-o", "table"}, want: false},
	} {
		t.Run(tc.name, func(t *testing.T) {
			var asJSON bool
			cmd := &cobra.Command{Use: "ls", RunE: func(*cobra.Command, []string) error { return nil }}
			addOutputFormatFlag(cmd, &asJSON, "legacy")
			cmd.SetArgs(tc.args)
			cmd.SetOut(&bytes.Buffer{})
			cmd.SetErr(&bytes.Buffer{})
			if err := cmd.Execute(); err != nil {
				t.Fatalf("execute %v: %v", tc.args, err)
			}
			if asJSON != tc.want {
				t.Fatalf("%v: asJSON = %v, want %v", tc.args, asJSON, tc.want)
			}
		})
	}
}

func TestUnknownOutputFormatIsRefusedBeforeTheRequest(t *testing.T) {
	var asJSON bool
	var ran bool
	cmd := &cobra.Command{Use: "ls", RunE: func(*cobra.Command, []string) error { ran = true; return nil }}
	addOutputFormatFlag(cmd, &asJSON, "legacy")
	cmd.SetArgs([]string{"-o", "yaml"})
	cmd.SetOut(&bytes.Buffer{})
	cmd.SetErr(&bytes.Buffer{})
	cmd.SilenceUsage = true
	cmd.SilenceErrors = true

	err := cmd.Execute()
	if err == nil {
		t.Fatal("an unsupported format was accepted")
	}
	if !strings.Contains(err.Error(), "use table or json") {
		t.Fatalf("error does not say what is accepted: %v", err)
	}
	if ran {
		t.Fatal("the command ran before the format was rejected")
	}
}

// pflag's own deprecation line goes to the flag set's writer, which is
// the command's stdout — the stream the caller just asked to be JSON.
// That is why `--json` is hidden rather than pflag-deprecated and the
// notice is written by hand, and this is the test that says so.
func TestTheLegacySpellingIsAnnouncedOnStderrOnly(t *testing.T) {
	var asJSON bool
	var out, errOut bytes.Buffer
	cmd := &cobra.Command{Use: "ls", RunE: func(*cobra.Command, []string) error { return nil }}
	addOutputFormatFlag(cmd, &asJSON, "legacy")
	cmd.SetArgs([]string{"--json"})
	cmd.SetOut(&out)
	cmd.SetErr(&errOut)
	if err := cmd.Execute(); err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(errOut.String(), "use -o json") {
		t.Fatalf("stderr does not name the replacement: %q", errOut.String())
	}
	if out.Len() != 0 {
		t.Fatalf("the notice reached stdout, which may be a JSON pipe: %q", out.String())
	}
}

// `files archive cat -o <path>` writes an entry to a local file. A short
// flag meaning a path in one verb of a subtree and a format in the next
// is worse than no short flag, so `archive entries` does without one.
func TestArchiveEntriesTakesTheLongFormatFlagOnly(t *testing.T) {
	var asJSON bool
	cmd := &cobra.Command{Use: "entries", RunE: func(*cobra.Command, []string) error { return nil }}
	addOutputFormatFlagLongOnly(cmd, &asJSON, "legacy")
	if shorthand := cmd.Flags().Lookup("output").Shorthand; shorthand != "" {
		t.Fatalf("archive entries claims -%s, which archive cat already means as a path", shorthand)
	}
}

// A deprecated flag keeps working and stops being advertised, so the
// help an agent reads names exactly one spelling.
func TestHelpAdvertisesOneSpelling(t *testing.T) {
	var asJSON bool
	cmd := &cobra.Command{Use: "ls"}
	addOutputFormatFlag(cmd, &asJSON, "legacy")
	usage := cmd.Flags().FlagUsages()
	if !strings.Contains(usage, "--output") {
		t.Fatalf("help does not offer --output: %s", usage)
	}
	if strings.Contains(usage, "--json") {
		t.Fatalf("help still offers the deprecated spelling: %s", usage)
	}
}
