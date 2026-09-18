package ctl

import (
	"io"
	"strings"
	"testing"

	"github.com/spf13/cobra"
)

type commandGroup struct {
	path  []string
	child string
}

func TestEveryCommandGroupRefusesAnUnknownVerb(t *testing.T) {
	root := NewDefaultCommand()
	root.SetOut(io.Discard)
	root.SetErr(io.Discard)

	for _, group := range commandGroups(root) {
		root.SetArgs(append(append([]string{}, group.path...), "definitely-not-a-verb"))
		err := root.Execute()
		if err == nil {
			t.Errorf("`olares-cli %s definitely-not-a-verb` succeeded", strings.Join(group.path, " "))
			continue
		}
		if !strings.Contains(err.Error(), `unknown verb "definitely-not-a-verb"`) {
			t.Errorf("`olares-cli %s`: unexpected error: %v", strings.Join(group.path, " "), err)
		}
	}
}

func TestEveryBareCommandGroupStillPrintsHelp(t *testing.T) {
	root := NewDefaultCommand()
	root.SetErr(io.Discard)
	var out strings.Builder
	root.SetOut(&out)

	for _, group := range commandGroups(root) {
		out.Reset()
		root.SetArgs(group.path)
		if err := root.Execute(); err != nil {
			t.Errorf("`olares-cli %s` failed: %v", strings.Join(group.path, " "), err)
			continue
		}
		if !strings.Contains(out.String(), "Usage:") {
			t.Errorf("`olares-cli %s` printed no help", strings.Join(group.path, " "))
		}
	}
}

func TestEveryCommandGroupAnswersHelpForAChild(t *testing.T) {
	root := NewDefaultCommand()
	root.SetErr(io.Discard)
	var out strings.Builder
	root.SetOut(&out)

	for _, group := range commandGroups(root) {
		out.Reset()
		args := append(append([]string{}, group.path...), "help", group.child)
		root.SetArgs(args)
		if err := root.Execute(); err != nil {
			t.Errorf("`olares-cli %s` failed: %v", strings.Join(args, " "), err)
			continue
		}
		if !strings.Contains(out.String(), "Usage:") {
			t.Errorf("`olares-cli %s` printed no help", strings.Join(args, " "))
		}
	}
}

func TestUnknownVerbSuggestsTheVerbThatWasMeant(t *testing.T) {
	root := NewDefaultCommand()
	root.SetOut(io.Discard)
	root.SetErr(io.Discard)

	for _, tc := range []struct {
		args []string
		want string
	}{
		{[]string{"router", "call", "diar"}, "diarize"},
		{[]string{"router", "usage", "sumary"}, "summary"},
		{[]string{"market", "instal"}, "install"},
	} {
		root.SetArgs(tc.args)
		err := root.Execute()
		if err == nil || !strings.Contains(err.Error(), tc.want) {
			t.Errorf("`olares-cli %s`: got %v, want suggestion %q", strings.Join(tc.args, " "), err, tc.want)
		}
	}
}

func commandGroups(root *cobra.Command) []commandGroup {
	var groups []commandGroup
	var walk func(*cobra.Command)
	walk = func(cmd *cobra.Command) {
		if cmd.Annotations[unknownVerbGroupAnnotation] == "true" {
			children := cmd.Commands()
			groups = append(groups, commandGroup{
				path:  strings.Fields(strings.TrimPrefix(cmd.CommandPath(), root.Name()+" ")),
				child: children[0].Name(),
			})
		}
		for _, child := range cmd.Commands() {
			walk(child)
		}
	}
	walk(root)
	return groups
}
