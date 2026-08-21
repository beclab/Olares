package router

import (
	"testing"

	"github.com/spf13/cobra"
	"github.com/spf13/pflag"

	"github.com/beclab/Olares/cli/pkg/cmdutil"
)

// --yes had three wordings and two shapes: the delete verbs said "skip the
// confirmation prompt", the spec and restart ones said "do not ask", and only
// three of them carried the -y shorthand. None of that was a decision — the
// flag means one thing everywhere it appears.
func TestEveryConfirmationFlagIsSpelledTheSameWay(t *testing.T) {
	var found int
	for _, cmd := range allRouterCommands(t) {
		flag := cmd.Flags().Lookup("yes")
		if flag == nil {
			continue
		}
		found++
		if flag.Usage != confirmFlagHelp {
			t.Errorf("%s: --yes help is %q, not the shared wording", cmd.CommandPath(), flag.Usage)
		}
		if flag.Shorthand != "y" {
			t.Errorf("%s: --yes has shorthand %q, not -y", cmd.CommandPath(), flag.Shorthand)
		}
	}
	if found == 0 {
		t.Fatal("no --yes flag anywhere; the walk is not reaching the verbs")
	}
}

// Cobra panics on a duplicate shorthand when the command runs rather than when
// it is built, so putting -y on every confirming verb could have gone unnoticed
// until somebody typed the one it collided on.
func TestNoCommandDeclaresTheSameShorthandTwice(t *testing.T) {
	for _, cmd := range allRouterCommands(t) {
		seen := map[string]string{}
		cmd.Flags().VisitAll(func(f *pflag.Flag) {
			if f.Shorthand == "" {
				return
			}
			if prev, dup := seen[f.Shorthand]; dup {
				t.Errorf("%s: -%s is both --%s and --%s", cmd.CommandPath(), f.Shorthand, prev, f.Name)
				return
			}
			seen[f.Shorthand] = f.Name
		})
	}
}

func allRouterCommands(t *testing.T) []*cobra.Command {
	t.Helper()
	var out []*cobra.Command
	var walk func(*cobra.Command)
	walk = func(c *cobra.Command) {
		out = append(out, c)
		for _, sub := range c.Commands() {
			walk(sub)
		}
	}
	walk(NewRouterCommand(&cmdutil.Factory{}))
	return out
}
