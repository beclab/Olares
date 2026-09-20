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

// Unknown verbs are reported once by cmd/main.go, so the dashboard wrapper
// must leave the shared error untouched and print nothing itself.
func TestUnknownVerbIsLeftForTheEntrypoint(t *testing.T) {
	cmd := NewDashboardCommand(nil)
	var stderr bytes.Buffer
	cmd.SetErr(&stderr)
	cmd.SetOut(io.Discard)
	cmd.SetArgs([]string{"overview", "podz"}) // typo: pods

	if err := cmd.Execute(); err == nil {
		t.Fatal("Execute should return non-nil error on typo; got nil")
	} else if !strings.Contains(err.Error(), `unknown verb "podz"`) {
		t.Errorf("unexpected typo error: %v", err)
	}
	if stderr.Len() != 0 {
		t.Errorf("dashboard printed the unknown verb before cmd/main.go: %q", stderr.String())
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
