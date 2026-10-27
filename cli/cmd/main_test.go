package main

import (
	"encoding/json"
	"os"
	"os/exec"
	"strings"
	"testing"
)

// The unit tests below cover one decision each. This one runs the real
// binary, because the bug it guards against only exists between the
// pieces: every part was tested and a `-o json` failure still answered
// in prose, since the entrypoint asked a different question about the
// output flag than the renderers did.
//
// The subprocess is this test binary re-entered through TestMain, so
// there is nothing to build first.
const subprocessEnv = "OLARES_CLI_MAIN_TEST_SUBPROCESS"

func TestMain(m *testing.M) {
	if os.Getenv(subprocessEnv) == "1" {
		main()
		return
	}
	os.Exit(m.Run())
}

func TestAFailedJSONInvocationAnswersInJSON(t *testing.T) {
	for _, tc := range []struct {
		name string
		args []string
		// A failure is either one JSON document or a prose line; the
		// point of the test is that the choice tracks what was asked.
		wantJSON bool
	}{
		{name: "the common spelling", args: []string{"files", "ls", "-o", "json"}, wantJSON: true},
		{name: "the long form", args: []string{"files", "ls", "--output", "json"}, wantJSON: true},
		// Every renderer compares case-insensitively. The entrypoint
		// did not, so this used to render JSON and fail in prose.
		{name: "shouted", args: []string{"files", "ls", "-o", "JSON"}, wantJSON: true},
		// The files tree keeps `--json` for scripts written against it.
		// A caller who passes it asked the same question.
		{name: "the legacy spelling", args: []string{"files", "ls", "--json"}, wantJSON: true},
		{name: "a person at a terminal", args: []string{"files", "ls"}, wantJSON: false},
		{name: "an explicit table", args: []string{"files", "ls", "-o", "table"}, wantJSON: false},
	} {
		t.Run(tc.name, func(t *testing.T) {
			// `files ls` with no path fails on arity, before any
			// profile or network is consulted, so this says the same
			// thing on a developer's machine and in CI.
			_, stderr, err := runCLI(t, tc.args...)
			if err == nil {
				t.Fatal("the invocation was expected to fail")
			}

			var envelope struct {
				Error struct {
					Code    string `json:"code"`
					Message string `json:"message"`
				} `json:"error"`
			}
			decodeErr := json.Unmarshal([]byte(stderr), &envelope)
			if !tc.wantJSON {
				if decodeErr == nil {
					t.Fatalf("a caller who did not ask for JSON got one: %q", stderr)
				}
				if !strings.HasPrefix(stderr, "Error: ") {
					t.Fatalf("stderr is neither the plain line nor JSON: %q", stderr)
				}
				return
			}
			// Unmarshal of the whole stream, not of a line found in it:
			// anything the CLI prepends -- the skill staleness notice
			// was the live risk -- makes this fail, which is the point.
			if decodeErr != nil {
				t.Fatalf("stderr does not parse as one JSON document (%v): %q", decodeErr, stderr)
			}
			if envelope.Error.Code == "" || envelope.Error.Message == "" {
				t.Fatalf("envelope is missing the fields a caller branches on: %q", stderr)
			}
		})
	}
}

func TestContradictoryOutputFlagsAreRefused(t *testing.T) {
	_, stderr, err := runCLI(t, "files", "ls", "drive/Home/", "--json", "-o", "table")
	if err == nil {
		t.Fatal("contradictory flags were accepted")
	}
	if !strings.Contains(stderr, "disagree") {
		t.Fatalf("the refusal does not say what is wrong: %q", stderr)
	}
}

// The notice cannot be provoked from a test build -- version.VERSION reads
// as a development build, which Notice deliberately says nothing about -- so
// the gate is tested where the decision is made rather than through its
// effect.
func TestSkillDriftIsAnnouncedOnlyWhereProseBelongs(t *testing.T) {
	for _, tc := range []struct {
		name            string
		machineReadable bool
		args            []string
		want            bool
	}{
		{name: "an ordinary verb", args: []string{"olares-cli", "market", "list"}, want: true},
		{name: "no verb at all", args: []string{"olares-cli"}, want: true},
		{name: "asked for JSON", machineReadable: true, args: []string{"olares-cli", "market", "list"}, want: false},
		{name: "the tree that installs them", args: []string{"olares-cli", "skills", "install"}, want: false},
	} {
		t.Run(tc.name, func(t *testing.T) {
			if got := shouldAnnounceSkillDrift(tc.machineReadable, tc.args); got != tc.want {
				t.Fatalf("shouldAnnounceSkillDrift = %v, want %v", got, tc.want)
			}
		})
	}
}

func runCLI(t *testing.T, args ...string) (stdout, stderr string, err error) {
	t.Helper()
	cmd := exec.Command(os.Args[0], args...)
	cmd.Env = append(os.Environ(), subprocessEnv+"=1")
	var out, errOut strings.Builder
	cmd.Stdout = &out
	cmd.Stderr = &errOut
	err = cmd.Run()
	return out.String(), errOut.String(), err
}
