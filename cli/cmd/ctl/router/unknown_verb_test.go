package router

import (
	"io"
	"strings"
	"testing"

	"github.com/spf13/cobra"
)

// A misspelled verb used to succeed. Cobra reaches a command's `Args` only if
// that command is runnable, so a group that only held others returned
// flag.ErrHelp for a word it could not resolve — and ExecuteC turns that into a
// help page and a zero exit status. `router call diar` was reported that way:
// the agent meant `diarize`, was handed the help for `call`, and read the exit
// status as the work having been done.
//
// The tests below go through Execute rather than reading fields, because the
// field that looked like it governed this — `Args` — is exactly the one that
// was never consulted.
func TestEveryGroupRefusesAnUnknownVerb(t *testing.T) {
	for _, group := range groupPaths(NewRouterCommand(nil)) {
		root := NewRouterCommand(nil)
		root.SetOut(io.Discard)
		root.SetErr(io.Discard)
		root.SetArgs(append(append([]string{}, group.path...), "definitely-not-a-verb"))
		if err := root.Execute(); err == nil {
			t.Errorf("`router %s definitely-not-a-verb` succeeded; a misspelling has to fail",
				strings.Join(group.path, " "))
		}
	}
}

// A bare group still prints its help and succeeds, which is the behaviour the
// refusal above must not have cost. A retired group lists nothing, because
// every verb under it is hidden — it still has to succeed.
func TestABareGroupStillPrintsHelp(t *testing.T) {
	for _, group := range groupPaths(NewRouterCommand(nil)) {
		root := NewRouterCommand(nil)
		var out strings.Builder
		root.SetOut(&out)
		root.SetErr(io.Discard)
		root.SetArgs(group.path)
		if err := root.Execute(); err != nil {
			t.Errorf("`router %s` failed: %v", strings.Join(group.path, " "), err)
			continue
		}
		if group.hidden {
			continue
		}
		if !strings.Contains(out.String(), "Available Commands:") {
			t.Errorf("`router %s` printed no command list", strings.Join(group.path, " "))
		}
	}
}

// The message names the near miss, because that is the case worth catching.
// `diar` is a prefix of the verb that was meant; `sumary` is a letter short of
// one, and only reads as a suggestion once the levenshtein threshold is set.
func TestTheRefusalSuggestsTheVerbThatWasMeant(t *testing.T) {
	for _, tc := range []struct{ typed, meant string }{
		{"diar", "diarize"},
		{"sumary", "summary"},
	} {
		root := NewRouterCommand(nil)
		root.SetOut(io.Discard)
		root.SetErr(io.Discard)
		parent := "call"
		if tc.typed == "sumary" {
			parent = "usage"
		}
		root.SetArgs([]string{parent, tc.typed})
		err := root.Execute()
		if err == nil {
			t.Fatalf("`router %s %s` succeeded", parent, tc.typed)
		}
		if !strings.Contains(err.Error(), tc.meant) {
			t.Errorf("the refusal for %q does not point at %q: %v", tc.typed, tc.meant, err)
		}
	}
}

// `help` is not a misspelling. Cobra mounts its help command on the binary's
// root, so a subtree never had one and the word reaches the group's own Run.
func TestAGroupAnswersHelp(t *testing.T) {
	for _, args := range [][]string{{"help"}, {"help", "call"}, {"call", "help"}} {
		root := NewRouterCommand(nil)
		var out strings.Builder
		root.SetOut(&out)
		root.SetErr(io.Discard)
		root.SetArgs(args)
		if err := root.Execute(); err != nil {
			t.Errorf("`router %s` failed: %v", strings.Join(args, " "), err)
			continue
		}
		if !strings.Contains(out.String(), "Available Commands:") {
			t.Errorf("`router %s` printed no command list", strings.Join(args, " "))
		}
	}
}

type group struct {
	path   []string
	hidden bool
}

// groupPaths returns the path to every command that exists only to hold others.
// A command with children and a positional of its own — `call music [prompt…]`,
// whose extra verbs hang beneath it — is not one of these, because there a bare
// word is legitimate input rather than a misspelling.
func groupPaths(root *cobra.Command) []group {
	var out []group
	var walk func(cmd *cobra.Command, path []string, hidden bool)
	walk = func(cmd *cobra.Command, path []string, hidden bool) {
		var children []*cobra.Command
		for _, child := range cmd.Commands() {
			if child.Name() == "help" || child.Name() == "completion" {
				continue
			}
			children = append(children, child)
		}
		if len(children) > 0 && !strings.ContainsAny(cmd.Use, "[<") {
			out = append(out, group{path: path, hidden: hidden})
		}
		for _, child := range children {
			walk(child, append(append([]string{}, path...), child.Name()), hidden || child.Hidden)
		}
	}
	walk(root, nil, root.Hidden)
	return out
}
