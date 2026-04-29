// Package dashboard's cmd-root test surface owns the tests that
// inherently bind to the cobra command-tree wiring assembled by
// NewDashboardCommand: typo handling on dispatch-only parents, and the
// "every cobra leaf must Silence{Errors,Usage}" regression net.
//
// All other tests — flag validation, fetcher wire shapes, aggregator
// math, format helpers, capability gates — live in
// cli/pkg/dashboard/*_test.go (P3c migration). Cmd subpackages
// (overview/, applications/, schema/) own area-local integration tests
// such as overview/gpu/detail_test.go.
package dashboard

import (
	"bytes"
	"io"
	"strings"
	"testing"

	"github.com/spf13/cobra"
)

// newTypoFixture builds a (root → parent w/ unknownSubcommandRunE) tree
// that mirrors the real olares-cli wiring (dashboard is a subcommand of
// the olares-cli root, never the root itself). The shape matters:
// cobra's default args validator (legacyArgs in cobra v1.9 args.go:24)
// takes a hard "unknown command" path when an unknown positional lands
// on a command that has subcommands AND no parent — testing on a bare
// root would bypass our RunE altogether. See cobra args.go:28-39.
func newTypoFixture() (*cobra.Command, *cobra.Command) {
	root := &cobra.Command{Use: "olares-cli"}
	parent := &cobra.Command{
		Use:           "dashboard",
		Short:         "the dashboard subtree",
		SilenceErrors: true,
		SilenceUsage:  true,
		RunE:          unknownSubcommandRunE,
	}
	parent.AddCommand(&cobra.Command{Use: "applications", RunE: func(*cobra.Command, []string) error { return nil }})
	parent.AddCommand(&cobra.Command{Use: "overview", RunE: func(*cobra.Command, []string) error { return nil }})
	root.AddCommand(parent)
	return root, parent
}

func TestUnknownSubcommandRunE_PrintsSuggestionAndFailsOnTypo(t *testing.T) {
	root, _ := newTypoFixture()

	var stderr bytes.Buffer
	root.SetErr(&stderr)
	root.SetOut(io.Discard)
	root.SetArgs([]string{"dashboard", "application"}) // typo, missing 's'

	err := root.Execute()
	if err == nil {
		t.Fatal("Execute should return non-nil error on typo; got nil")
	}
	out := stderr.String()
	if !strings.Contains(out, `unknown subcommand "application"`) {
		t.Errorf("stderr missing 'unknown subcommand' marker; got: %q", out)
	}
	if !strings.Contains(out, "Did you mean this?") || !strings.Contains(out, "applications") {
		t.Errorf("stderr missing suggestion 'applications'; got: %q", out)
	}
}

func TestUnknownSubcommandRunE_NoArgsPrintsHelp(t *testing.T) {
	root, _ := newTypoFixture()

	var stdout, stderr bytes.Buffer
	root.SetOut(&stdout)
	root.SetErr(&stderr)
	root.SetArgs([]string{"dashboard"}) // bare parent, no subcmd

	if err := root.Execute(); err != nil {
		t.Fatalf("Execute with no args should succeed (just print help); got %v", err)
	}
	if !strings.Contains(stdout.String(), "applications") {
		t.Errorf("help output should list 'applications'; got: %q", stdout.String())
	}
	if stderr.Len() != 0 {
		t.Errorf("stderr should be empty on no-arg help; got: %q", stderr.String())
	}
}

// TestLeafErrorsAreReported pins wrapLeafErrors: every leaf RunE that
// returns a non-sentinel error MUST surface the message on stderr so
// users / agents see WHY the process exits non-zero. Without this
// wrapper, dashboard's blanket SilenceErrors=true contract would
// silently swallow flag-validation errors (and friends) into a bare
// `exit 1` with no diagnostic.
//
// We exercise the path with `applications --output xyz` because
// CommonFlags.Validate fails fast on a bad --output, well before any
// leaf reaches into the (nil) factory — keeping the test hermetic.
func TestLeafErrorsAreReported(t *testing.T) {
	cmd := NewDashboardCommand(nil)
	var stderr bytes.Buffer
	cmd.SetErr(&stderr)
	cmd.SetOut(io.Discard)
	cmd.SetArgs([]string{"applications", "--output", "xyz"})

	err := cmd.Execute()
	if err == nil {
		t.Fatal("Execute should return non-nil error on bad --output; got nil")
	}
	out := stderr.String()
	if !strings.Contains(out, "unknown output format") {
		t.Errorf("stderr should contain leaf error; got: %q", out)
	}
}

// TestLeafErrorsSentinelNotDoublePrinted verifies wrapLeafErrors honours
// pkgdashboard.ErrAlreadyReported: a typo'd subcommand should produce
// the unknownSubcommandRunE-authored hint on stderr exactly ONCE, not
// twice. (Before the sentinel was introduced, the wrapper would happily
// fmt.Fprintln "unknown subcommand" on top of the suggestion text.)
func TestLeafErrorsSentinelNotDoublePrinted(t *testing.T) {
	cmd := NewDashboardCommand(nil)
	var stderr bytes.Buffer
	cmd.SetErr(&stderr)
	cmd.SetOut(io.Discard)
	cmd.SetArgs([]string{"overview", "podz"}) // typo: pods

	if err := cmd.Execute(); err == nil {
		t.Fatal("Execute should return non-nil error on typo; got nil")
	}
	out := stderr.String()
	if !strings.Contains(out, `unknown subcommand "podz"`) {
		t.Errorf("stderr missing typo hint; got: %q", out)
	}
	// The sentinel's own message must not be appended on top of the
	// suggestion block. Count only literal occurrences (not substring
	// matches inside the suggestion text).
	if c := strings.Count(out, "dashboard: error already reported"); c != 0 {
		t.Errorf("sentinel string leaked into stderr %d times: %q", c, out)
	}
}

// TestAllLeafCommandsSilenced is the regression net for the "Cobra
// printed usage when HAMI returned 5xx" bug. Every command in the
// dashboard subtree (root + leaves + intermediate sections-envelopes)
// MUST set both SilenceErrors and SilenceUsage so a runtime error
// does NOT produce help text on stderr — only the structured envelope
// (or the typed error) reaches the agent.
func TestAllLeafCommandsSilenced(t *testing.T) {
	root := NewDashboardCommand(nil)
	var visit func(c *cobra.Command)
	visit = func(c *cobra.Command) {
		// `dashboard` is the parent; help / completion are cobra
		// built-ins we don't own. Skip them.
		if c.Name() == "help" || c.Name() == "completion" {
			return
		}
		if !c.SilenceErrors {
			t.Errorf("cobra cmd %q lacks SilenceErrors=true", c.CommandPath())
		}
		if !c.SilenceUsage {
			t.Errorf("cobra cmd %q lacks SilenceUsage=true", c.CommandPath())
		}
		for _, sub := range c.Commands() {
			visit(sub)
		}
	}
	visit(root)
}
