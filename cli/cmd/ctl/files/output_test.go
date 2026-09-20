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
		// Every renderer in the CLI compares case-insensitively; the
		// entrypoint used to not, so `-o JSON` succeeded as JSON and
		// failed as prose.
		{name: "the shouted form", args: []string{"-o", "JSON"}, want: true},
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

// The legacy spelling says nothing on either stream. pflag's own
// deprecation line goes to the flag set's writer, which is the command's
// stdout — the stream the caller just asked to be JSON — which is why
// `--json` is hidden rather than pflag-deprecated. A hand-written notice
// on stderr is no better: it fires exactly when the caller wants machine
// -readable output, and stderr is where a failing `-o json` answers.
func TestTheLegacySpellingSaysNothing(t *testing.T) {
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
	if out.Len() != 0 || errOut.Len() != 0 {
		t.Fatalf("a JSON invocation was given prose to parse: stdout=%q stderr=%q", out.String(), errOut.String())
	}
}

// Two flags naming one setting either agree or stop the command. The
// version this replaced let `--json` write the bool directly, so
// `--json -o table` rendered JSON while every later reader of `--output`
// — including the entrypoint deciding whether a failure should be an
// error envelope — was told it was a table.
func TestTheTwoSpellingsMustAgree(t *testing.T) {
	for _, tc := range []struct {
		name     string
		args     []string
		wantErr  bool
		wantJSON bool
	}{
		{name: "both say json", args: []string{"--json", "-o", "json"}, wantJSON: true},
		{name: "both say table", args: []string{"--json=false", "-o", "table"}, wantJSON: false},
		{name: "they disagree", args: []string{"--json", "-o", "table"}, wantErr: true},
		{name: "they disagree the other way", args: []string{"--json=false", "-o", "json"}, wantErr: true},
	} {
		t.Run(tc.name, func(t *testing.T) {
			var asJSON, ran bool
			cmd := &cobra.Command{Use: "ls", RunE: func(*cobra.Command, []string) error { ran = true; return nil }}
			addOutputFormatFlag(cmd, &asJSON, "legacy")
			cmd.SetArgs(tc.args)
			cmd.SetOut(&bytes.Buffer{})
			cmd.SetErr(&bytes.Buffer{})
			cmd.SilenceUsage = true
			cmd.SilenceErrors = true

			err := cmd.Execute()
			if tc.wantErr {
				if err == nil {
					t.Fatal("contradictory flags were accepted")
				}
				if !strings.Contains(err.Error(), "disagree") {
					t.Fatalf("error does not say what is wrong: %v", err)
				}
				if ran {
					t.Fatal("the command ran on a setting it could not resolve")
				}
				return
			}
			if err != nil {
				t.Fatalf("execute %v: %v", tc.args, err)
			}
			if asJSON != tc.wantJSON {
				t.Fatalf("%v: asJSON = %v, want %v", tc.args, asJSON, tc.wantJSON)
			}
		})
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
