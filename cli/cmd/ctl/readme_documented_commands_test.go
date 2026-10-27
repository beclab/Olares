package ctl

import (
	"os"
	"path/filepath"
	"regexp"
	"strings"
	"testing"

	"github.com/spf13/cobra"
)

// The skills have been guarded against inventing a verb since
// TestEveryCommandTheSkillsDocumentResolves; the two READMEs had nothing, and
// it showed. `olares-cli profile current` was the documented way to verify a
// login in both of them and in the install wizard's own output, and it has
// never been a command — Cobra answers an unknown verb by printing the parent
// group's help and exiting zero, so the mistake reads as success to anyone
// skimming and to any script checking an exit code.
//
// readmeSources are read from disk rather than from an embedded copy, which is
// the one place these tests differ from the skills ones deliberately. The
// skills are harvested from skills.FS() because the embedded copy is what
// reaches an agent and a file the embed pattern stopped matching should not be
// checked. A README is not embedded in anything: the bytes in the working tree
// are the bytes that reach GitHub and npmjs.com, so that is what to read.
var readmeSources = []string{
	filepath.Join("..", "..", "README.md"),
	filepath.Join("..", "..", "npm", "README.md"),
}

// Every command in these files is written out with the binary's name in front
// of it, which makes the scan far simpler than the skills one: there is no
// index whose rows drop the tree's own name, so there is no need to try a
// phrase against several bases, and no need to guess whether a backticked
// phrase is prose. A phrase that does not start with "olares-cli" is not a
// claim about a command and is not read as one.
const readmeBinary = "olares-cli"

// npxPattern matches the other way these files spell an invocation:
// `npx @olares/cli@latest <verb>`, which reaches the same command tree through
// the Node shim. Two of the nine documented `profile current` uses were
// written this way, so a scan that only knew the bare binary name would have
// left them behind while reporting the file fixed.
//
// `install` is the exception the pattern does not need to know about: the shim
// intercepts that verb for the setup wizard and it never reaches the Go binary
// (cli/npm/bin/olares-cli.js). It is also a real root command here, so it
// resolves either way.
var npxPattern = regexp.MustCompile(`^npx\s+@olares/cli(?:@[\w.-]+)?\s+(.*)$`)

// TestEveryCommandTheReadmesDocumentResolves walks the fenced examples.
//
// The invariant is the one the skills test established: a word following a
// group command has to be one of that group's subcommands, because a group
// takes no positional arguments and there is nothing else the word could be.
// A leaf ends the walk — its arguments are not this test's business.
func TestEveryCommandTheReadmesDocumentResolves(t *testing.T) {
	root := NewDefaultCommand()
	invocations := readmeFencedInvocations(t)
	// A parser that silently stops matching turns this file into a test that
	// passes by reading nothing. The READMEs carry dozens of examples between
	// them; a number this low can only mean the scan broke.
	if len(invocations) < 20 {
		t.Fatalf("harvested only %d fenced invocations from the READMEs; the docs or the scanner moved", len(invocations))
	}
	for _, invocation := range invocations {
		t.Run(invocation.file+":"+invocation.path(), func(t *testing.T) {
			assertResolves(t, root, invocation)
		})
	}
}

// TestEveryInlineReadmeCommandPathResolves covers the commands named in prose.
//
// Worth its own pass rather than folding into the one above: of the three
// places `profile current` appeared in cli/README.md, one was an inline
// mention inside a blockquote, and a scanner that only reads fenced blocks
// would have fixed two of the three and called the file clean.
func TestEveryInlineReadmeCommandPathResolves(t *testing.T) {
	root := NewDefaultCommand()
	invocations := readmeInlineInvocations(t)
	if len(invocations) < 10 {
		t.Fatalf("harvested only %d inline commands from the READMEs; the docs or the scanner moved", len(invocations))
	}
	for _, invocation := range invocations {
		t.Run(invocation.file+":"+invocation.path(), func(t *testing.T) {
			assertResolves(t, root, invocation)
		})
	}
}

func assertResolves(t *testing.T, root *cobra.Command, invocation documentedInvocation) {
	t.Helper()
	cmd := root
	for _, token := range invocation.tokens {
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
	}
}

func readmeFencedInvocations(t *testing.T) []documentedInvocation {
	t.Helper()
	var found []documentedInvocation
	for _, path := range readmeSources {
		source := readReadme(t, path)
		for _, shell := range fencedShellLines(source) {
			// A fenced block is not always shell — these files also fence JSON,
			// console output and directory trees, any of which can mention the
			// binary. A command to be typed starts its line.
			text := shell.text
			if npx := npxPattern.FindStringSubmatch(text); npx != nil {
				text = readmeBinary + " " + npx[1]
			}
			if !strings.HasPrefix(text, readmeBinary) {
				continue
			}
			for _, match := range invocationPattern.FindAllStringSubmatch(text, -1) {
				tokens := commandLikePrefix(strings.Fields(match[1]))
				if len(tokens) == 0 {
					continue
				}
				found = append(found, documentedInvocation{
					tokens: tokens,
					file:   path,
					line:   shell.number,
				})
			}
		}
	}
	return found
}

func readmeInlineInvocations(t *testing.T) []documentedInvocation {
	t.Helper()
	var found []documentedInvocation
	for _, path := range readmeSources {
		source := readReadme(t, path)
		for number, line := range strings.Split(source, "\n") {
			for _, match := range inlineCodePattern.FindAllStringSubmatch(line, -1) {
				phrase := strings.TrimSpace(match[1])
				if npx := npxPattern.FindStringSubmatch(phrase); npx != nil {
					phrase = readmeBinary + " " + npx[1]
				}
				fields := strings.Fields(phrase)
				// Prose backticks a great many things — flags, env vars, paths,
				// file names. Only a phrase that opens with the binary's name is
				// claiming to be a command.
				if len(fields) < 2 || fields[0] != readmeBinary {
					continue
				}
				tokens := commandLikePrefix(commandNameFields(fields)[1:])
				if len(tokens) == 0 {
					continue
				}
				found = append(found, documentedInvocation{
					tokens: tokens,
					file:   path,
					line:   number + 1,
				})
			}
		}
	}
	return found
}

func readReadme(t *testing.T, path string) string {
	t.Helper()
	source, err := os.ReadFile(path)
	if err != nil {
		// Not skipped: a README that moved without this list moving with it
		// leaves the file unguarded, which is the state this test exists to
		// end.
		t.Fatalf("read %s: %v", path, err)
	}
	return string(source)
}
