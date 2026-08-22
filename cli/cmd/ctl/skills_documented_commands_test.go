package ctl

import (
	"io/fs"
	"path/filepath"
	"regexp"
	"strings"
	"testing"

	"github.com/spf13/cobra"

	"github.com/beclab/Olares/cli/skills"
)

// The skills are the instructions an agent follows, so a command spelled
// wrong there is not a typo in prose — it is a step the agent cannot take,
// discovered at the worst moment. TestSkillCommandPathsExist guards a list
// maintained by hand, which covered well under half of what the docs actually
// tell a reader to run. This one takes the list from the docs themselves, so
// a newly documented command is checked by having been written down.
//
// The invariant it enforces: a word following a group command has to be one
// of that group's subcommands. Group commands take no positional arguments,
// so there is nothing else the word could be, which is what makes this
// catch an invented verb rather than quietly stopping short of it. Leaf
// commands end the walk, and their arguments are none of this test's
// business.
func TestEveryCommandTheSkillsDocumentResolves(t *testing.T) {
	root := NewDefaultCommand()
	for _, invocation := range documentedInvocations(t) {
		t.Run(invocation.path(), func(t *testing.T) {
			cmd := root
			for i, token := range invocation.tokens {
				next := childNamed(cmd, token)
				if next == nil {
					if hasSubcommands(cmd) {
						t.Fatalf("%s:%d documents %q, but %q is not a subcommand of %q (its subcommands: %s)",
							invocation.file, invocation.line, invocation.path(),
							token, cmd.CommandPath(), strings.Join(subcommandNames(cmd), ", "))
					}
					// A leaf's arguments are not commands; stop reading.
					return
				}
				cmd = next
				_ = i
			}
		})
	}
}

type documentedInvocation struct {
	tokens []string
	file   string
	line   int
}

func (d documentedInvocation) path() string { return strings.Join(d.tokens, " ") }

// documentedInvocation harvesting stops at anything that cannot be a command
// name — a flag, a placeholder, an uppercase word — because the regexp only
// admits lowercase words, and at a second "olares-cli", which is how two
// invocations written on one line stay two invocations.
var invocationPattern = regexp.MustCompile(`olares-cli((?:\s+[a-z][a-z0-9-]*)+)`)

// The documents are harvested from the embedded suite rather than from the
// working tree, because the embedded copy is the one that reaches an agent.
// Reading ../../skills would check files that a release does not ship — and
// would keep passing if the embed patterns stopped matching some of them.
func documentedInvocations(t *testing.T) []documentedInvocation {
	t.Helper()

	suite := skills.FS()
	seen := map[string]bool{}
	var found []documentedInvocation

	err := fs.WalkDir(suite, ".", func(path string, entry fs.DirEntry, err error) error {
		if err != nil || entry.IsDir() || filepath.Ext(path) != ".md" {
			return err
		}
		source, err := fs.ReadFile(suite, path)
		if err != nil {
			return err
		}
		inFence := false
		for lineNumber, line := range strings.Split(string(source), "\n") {
			if strings.HasPrefix(strings.TrimSpace(line), "```") {
				inFence = !inFence
				continue
			}
			// Prose names commands it is telling the reader *not* to reach
			// for ("not an `olares-cli publish` lifecycle"), and English
			// sentences put ordinary words after the binary's name. A fenced
			// block is the one place every word is meant to be typed.
			if !inFence {
				continue
			}
			// A fenced block is not always shell — it also holds YAML, JSON
			// and output samples, any of which can mention the binary in
			// prose. A command to be typed starts its line.
			if !strings.HasPrefix(strings.TrimPrefix(strings.TrimSpace(line), "$ "), "olares-cli") {
				continue
			}
			for _, match := range invocationPattern.FindAllStringSubmatch(line, -1) {
				tokens := strings.Fields(match[1])
				// "olares-cli files ls ... olares-cli files cat ..." on one
				// line: the tail belongs to the next invocation, which the
				// regexp finds on its own.
				for i, token := range tokens {
					if token == "olares-cli" {
						tokens = tokens[:i]
						break
					}
				}
				if len(tokens) == 0 || seen[strings.Join(tokens, " ")] {
					continue
				}
				seen[strings.Join(tokens, " ")] = true
				found = append(found, documentedInvocation{
					tokens: tokens,
					file:   path,
					line:   lineNumber + 1,
				})
			}
		}
		return nil
	})
	if err != nil {
		t.Fatalf("walk the embedded suite: %v", err)
	}
	if len(found) < 100 {
		t.Fatalf("harvested only %d invocations; the docs or the pattern moved", len(found))
	}
	return found
}

func childNamed(cmd *cobra.Command, name string) *cobra.Command {
	for _, child := range cmd.Commands() {
		if child.Name() == name {
			return child
		}
		for _, alias := range child.Aliases {
			if alias == name {
				return child
			}
		}
	}
	return nil
}

func hasSubcommands(cmd *cobra.Command) bool { return len(cmd.Commands()) > 0 }

func subcommandNames(cmd *cobra.Command) []string {
	var names []string
	for _, child := range cmd.Commands() {
		names = append(names, child.Name())
	}
	return names
}
