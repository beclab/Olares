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

// The walk above stops at a leaf and calls that command's arguments none of
// its business, which leaves the half of an example an agent is most likely
// to copy wrong unchecked: the flags. A misspelled or renamed flag fails the
// same way an invented verb does, except later — the command exists, the
// agent runs it, and the CLI refuses an argument the docs promised.
func TestEveryFlagTheSkillsDocumentExists(t *testing.T) {
	root := NewDefaultCommand()
	var checked int
	for _, example := range documentedFlagUsages(t) {
		cmd := root
		for _, token := range example.tokens {
			next := childNamed(cmd, token)
			if next == nil {
				break
			}
			cmd = next
		}
		// A line that never left a group names flags belonging to a
		// subcommand this walk could not identify; guessing an owner would
		// report a spelling as wrong because the wrong command was asked.
		if cmd == root || hasSubcommands(cmd) {
			continue
		}
		for _, flag := range example.flags {
			checked++
			if cmd.Flags().Lookup(flag) != nil || cmd.InheritedFlags().Lookup(flag) != nil ||
				root.PersistentFlags().Lookup(flag) != nil || flag == "help" {
				continue
			}
			t.Errorf("%s:%d documents `--%s` on %q, which has no such flag",
				example.file, example.line, flag, cmd.CommandPath())
		}
	}
	if checked < 100 {
		t.Fatalf("checked only %d documented flags; the docs or the scanner moved", checked)
	}
}

// The check above is only as good as what the scanner hands it, and the
// scanner used to read physical lines. A long invocation is written
// across several with a trailing backslash — which is where the flags an
// author was least sure about sit — so everything past the first break
// was invisible. This says the scanner sees a whole command.
func TestAFlagOnAContinuationLineIsStillRead(t *testing.T) {
	const source = "```bash\n" +
		"olares-cli router call chat \\\n" +
		"  --model default-chat \\\n" +
		"  --prompt \"hello\" \\\n" +
		"  --invented-flag\n" +
		"```\n"

	usages := flagUsagesIn("synthetic.md", source)
	if len(usages) != 1 {
		t.Fatalf("a four-line invocation harvested as %d commands", len(usages))
	}
	usage := usages[0]
	if got := strings.Join(usage.tokens, " "); got != "router call chat" {
		t.Fatalf("command path = %q", got)
	}
	// The number is the line the reader has to go to, not the line the
	// offending flag happens to sit on.
	if usage.line != 2 {
		t.Fatalf("reported line %d, want the line the command starts on", usage.line)
	}
	want := []string{"model", "prompt", "invented-flag"}
	if got := strings.Join(usage.flags, ","); got != strings.Join(want, ",") {
		t.Fatalf("flags = %q, want %q", got, strings.Join(want, ","))
	}
}

// Quoted text is a prompt, a JSON body or a message, and a `--` inside one is
// not a flag. Everything after a pipe or a chained command belongs to another
// program entirely.
var (
	captureAssignmentPattern = regexp.MustCompile(`^[A-Za-z_][A-Za-z0-9_]*=\$?"?\$\(`)
	quotedSpanPattern        = regexp.MustCompile(`'[^']*'|"[^"]*"`)
	shellBreakPattern        = regexp.MustCompile(`\s(\||\|\||&&|;|>{1,2}|#)\s`)
	longFlagPattern          = regexp.MustCompile(`(?:^|\s)--([a-z][a-z0-9-]*)`)
)

type documentedFlagUsage struct {
	tokens []string
	flags  []string
	file   string
	line   int
}

func documentedFlagUsages(t *testing.T) []documentedFlagUsage {
	t.Helper()
	suite := skills.FS()
	var found []documentedFlagUsage

	err := fs.WalkDir(suite, ".", func(path string, entry fs.DirEntry, err error) error {
		if err != nil || entry.IsDir() || filepath.Ext(path) != ".md" {
			return err
		}
		source, err := fs.ReadFile(suite, path)
		if err != nil {
			return err
		}
		found = append(found, flagUsagesIn(path, string(source))...)
		return nil
	})
	if err != nil {
		t.Fatalf("walk the embedded suite: %v", err)
	}
	if len(found) < 50 {
		t.Fatalf("harvested only %d flagged examples; the docs or the scanner moved", len(found))
	}
	return found
}

func flagUsagesIn(path, source string) []documentedFlagUsage {
	var found []documentedFlagUsage
	for _, shell := range fencedShellLines(source) {
		// An invocation whose output is being captured --
		// `SHARE_ID=$(olares-cli files share public … --json)` --
		// is still a command a reader will run, and it was
		// exactly where a flag that did not exist survived.
		line := captureAssignmentPattern.ReplaceAllString(shell.text, "")
		if !strings.HasPrefix(line, "olares-cli ") {
			continue
		}
		line = quotedSpanPattern.ReplaceAllString(line, " ")
		if cut := shellBreakPattern.FindStringIndex(line); cut != nil {
			line = line[:cut[0]]
		}
		var flags []string
		for _, match := range longFlagPattern.FindAllStringSubmatch(line, -1) {
			flags = append(flags, match[1])
		}
		if len(flags) == 0 {
			continue
		}
		var tokens []string
		for _, field := range strings.Fields(strings.TrimPrefix(line, "olares-cli ")) {
			if strings.HasPrefix(field, "-") {
				break
			}
			tokens = append(tokens, field)
		}
		found = append(found, documentedFlagUsage{
			tokens: tokens,
			flags:  flags,
			file:   path,
			line:   shell.number,
		})
	}
	return found
}

// fencedShellLines is every command a fenced block tells a reader to run,
// as one string per command.
//
// The joining is the point. A long invocation is written across several
// physical lines with a trailing backslash, which is exactly where the
// flags an author was least sure about end up; scanners that read
// physical lines see a first line starting with `olares-cli` and a
// second line starting with `--something`, and check only the first.
// Every flag past the first break went unchecked, in the examples most
// likely to be copied whole.
//
// The reported number is the line the command starts on, because that is
// where a reader sent to fix it should land.
type fencedLine struct {
	text   string
	number int
}

func fencedShellLines(source string) []fencedLine {
	var (
		lines   []fencedLine
		inFence bool
		pending string
		start   int
	)
	flush := func() {
		if start != 0 {
			lines = append(lines, fencedLine{text: pending, number: start})
			pending, start = "", 0
		}
	}
	for index, raw := range strings.Split(source, "\n") {
		trimmed := strings.TrimSpace(raw)
		if strings.HasPrefix(trimmed, "```") {
			// A block that ends mid-continuation is malformed shell;
			// keep what there is rather than dropping the command.
			flush()
			inFence = !inFence
			continue
		}
		if !inFence {
			continue
		}
		text := strings.TrimPrefix(trimmed, "$ ")
		continues := strings.HasSuffix(text, `\`)
		if continues {
			text = strings.TrimSpace(strings.TrimSuffix(text, `\`))
		}
		if start == 0 {
			pending, start = text, index+1
		} else {
			pending = strings.TrimSpace(pending + " " + text)
		}
		if !continues {
			flush()
		}
	}
	flush()
	return lines
}

// The two tests below scan olares-router alone, which is where the
// inline verb index was first written. This one scans the other eleven
// skills' SKILL.md the same way.
//
// documentedCommandPaths walks these same spans to decide what counts as
// documented and drops the ones it cannot resolve without a word. So an
// inline command naming a verb that does not exist was not merely
// unchecked — it was quietly excluded from coverage, which is the
// arrangement where the docs and the CLI disagree and every test passes.
//
// Three things keep this from reporting prose as a broken command.
//
// It stops at the front door of each skill, as the router version does.
// References quote the CLI's own error text, and a backticked `cluster
// pod has multiple containers; please specify` opens with two real
// command names. Every one of those is in a reference; a SKILL.md is an
// index, and a backticked phrase in one is a command by construction.
//
// A phrase is only followed while it resolves. Once a token has matched
// a group command the next word has to be one of that group's
// subcommands — a group takes no positional arguments, so there is
// nothing else the word could be — while a phrase that resolves nothing
// at all is prose and is ignored.
//
// And a verb index is read relative to whatever tree its skill is about,
// which is not always that skill's own command: olares-knowledge indexes
// `settings get` under `knowledge download`, and names the unrelated
// top-level `download component` two paragraphs above it. So a phrase is
// tried against the root, the family, and the family's children, and is
// only reported when it resolves under none of them. That is looser than
// naming one base per skill, and it still answers the question worth
// asking, which is whether the verb exists anywhere.
func TestEveryInlineSkillCommandPathResolves(t *testing.T) {
	root := NewDefaultCommand()
	invocations := inlineSkillCommandPaths(t, root)
	if len(invocations) < 60 {
		t.Fatalf("harvested only %d inline commands outside olares-router; the docs or the scanner moved", len(invocations))
	}
	for _, invocation := range invocations {
		t.Run(invocation.file+":"+invocation.path(), func(t *testing.T) {
			var tried []string
			for _, base := range invocation.bases {
				if resolves(base, invocation.tokens) {
					return
				}
				tried = append(tried, base.CommandPath())
			}
			t.Errorf("%s:%d documents %q, which is not a command under any of %s",
				invocation.file, invocation.line, invocation.path(), strings.Join(tried, ", "))
		})
	}
}

// resolves reports whether tokens name a command under base. A walk that
// reaches a leaf stops there: a leaf's arguments are not commands.
func resolves(base *cobra.Command, tokens []string) bool {
	cmd := base
	for _, token := range tokens {
		next := childNamed(cmd, token)
		if next == nil {
			return !hasSubcommands(cmd)
		}
		cmd = next
	}
	return true
}

func inlineSkillCommandPaths(t *testing.T, root *cobra.Command) []documentedInvocation {
	t.Helper()
	suite := skills.FS()
	var found []documentedInvocation
	err := fs.WalkDir(suite, ".", func(path string, entry fs.DirEntry, err error) error {
		if err != nil || entry.IsDir() || filepath.Base(path) != "SKILL.md" {
			return err
		}
		if strings.HasPrefix(path, "olares-router/") {
			return nil
		}
		source, err := fs.ReadFile(suite, path)
		if err != nil {
			return err
		}
		family := childNamed(root, strings.TrimPrefix(strings.Split(path, "/")[0], "olares-"))
		bases := []*cobra.Command{root}
		if family != nil {
			bases = append(bases, family)
			bases = append(bases, family.Commands()...)
		}
		for lineNumber, line := range strings.Split(string(source), "\n") {
			for _, match := range inlineCodePattern.FindAllStringSubmatch(line, -1) {
				fields := commandNameFields(strings.Fields(strings.TrimSpace(match[1])))
				if len(fields) > 0 && fields[0] == "olares-cli" {
					fields = fields[1:]
				}
				fields = commandLikePrefix(fields)
				if len(fields) < 2 || !namesACommand(bases, fields[0]) {
					// Prose. Nothing here claims to be a command.
					continue
				}
				for _, tokens := range expandCommandAlternatives(fields) {
					found = append(found, documentedInvocation{
						tokens: tokens,
						bases:  bases,
						file:   path,
						line:   lineNumber + 1,
					})
				}
			}
		}
		return nil
	})
	if err != nil {
		t.Fatalf("walk the embedded suite: %v", err)
	}
	return found
}

// namesACommand is what admits a phrase to the check at all. A first
// word that is a command somewhere is the weakest claim that can still
// be a claim; everything else in backticks is a field, a state or a
// quoted message.
func namesACommand(bases []*cobra.Command, token string) bool {
	for _, base := range bases {
		if childNamed(base, token) != nil {
			return true
		}
	}
	return false
}

// This scans only the inline verb index in olares-router/SKILL.md. Reference
// prose contains backquoted error text and field names that are not commands;
// fenced executable examples across all embedded skills are covered above.
func TestEveryInlineRouterSkillCommandPathResolves(t *testing.T) {
	root := NewDefaultCommand()
	for _, invocation := range inlineRouterCommandPaths(t, root) {
		t.Run(invocation.path(), func(t *testing.T) {
			cmd := root
			for _, token := range invocation.tokens {
				next := childNamed(cmd, token)
				if next == nil {
					if !hasSubcommands(cmd) {
						return
					}
					t.Fatalf("%s:%d documents %q, but %q is not a subcommand of %q",
						invocation.file, invocation.line, invocation.path(), token, cmd.CommandPath())
				}
				cmd = next
			}
		})
	}
}

func TestEveryInlineRouterCLICommandPathResolves(t *testing.T) {
	root := NewDefaultCommand()
	for _, invocation := range inlineRouterCLICommandPaths(t) {
		t.Run(invocation.path(), func(t *testing.T) {
			cmd := root
			for _, token := range invocation.tokens {
				next := childNamed(cmd, token)
				if next == nil {
					if !hasSubcommands(cmd) {
						return
					}
					t.Fatalf("%s:%d documents %q, but %q is not a subcommand of %q",
						invocation.file, invocation.line, invocation.path(), token, cmd.CommandPath())
				}
				cmd = next
			}
		})
	}
}

func TestRouterSkillHasNoObsoleteCommandContracts(t *testing.T) {
	suite := skills.FS()
	oldSpeakerEmbed := regexp.MustCompile(`(?i)speaker[_-]embed[^\n]{0,80}router call embed`)
	retiredContracts := []string{
		"max_input_seconds",
		"MAX INPUT",
		"540-second input ceiling",
		"540-second limit",
		"exceeds its engine limit",
		"status read and a later `task result` can each carry the same measured duration",
		// Settlement moved onto a compare-and-swap on the task binding, and a
		// poll of a finished task closes the row. Both of these read the other
		// way round: a status read that costs nothing, and a collection that
		// must not be repeated.
		"may write a spend row, but it carries no duration",
		"`task result` is the exception",
		// Router waives the audio operation gate fifteen minutes after it last
		// observed the catalogue. Enforcing an aged one is the rule this skill
		// was written against before that.
		"Router enforces it regardless",
		// The three music submissions carry an idempotency key.
		"Router has no idempotency key",
	}
	var documented strings.Builder
	err := fs.WalkDir(suite, "olares-router", func(path string, entry fs.DirEntry, err error) error {
		if err != nil || entry.IsDir() || filepath.Ext(path) != ".md" {
			return err
		}
		source, err := fs.ReadFile(suite, path)
		if err != nil {
			return err
		}
		documented.Write(source)
		documented.WriteByte('\n')
		if strings.Contains(string(source), "diag perf") {
			t.Errorf("%s still documents removed `model diag perf`", path)
		}
		if oldSpeakerEmbed.Match(source) {
			t.Errorf("%s still maps speaker embedding to `router call embed`", path)
		}
		for _, contract := range retiredContracts {
			if strings.Contains(string(source), contract) {
				t.Errorf("%s still documents retired contract %q", path, contract)
			}
		}
		return nil
	})
	if err != nil {
		t.Fatalf("walk embedded router skill: %v", err)
	}
	for _, contract := range []string{
		"Model Console's `/api/endpoints`",
		"A poll now settles the call",
		"Fetching the same result twice is charged once",
		"audio_task_owner_unavailable",
		"no longer a reason to refuse",
		"supports_embedding_image_input",
		"`speak --sound-fx` resolves `default-sound-fx`",
		"Qwen accepts longer requests and internally splits",
		"`return_time_stamps=false`",
		"`terminus-apps` staging directories",
		"AAC, Opus",
		"task get <task-id> --model default-stt --wait -o json",
		"submit slices sequentially",
		"Router circuit open",
		"routing reference, not the engine's canonical model id",
	} {
		if !strings.Contains(documented.String(), contract) {
			t.Errorf("router skill lost required audio contract %q", contract)
		}
	}
}

var inlineCodePattern = regexp.MustCompile("`([^`\\n]+)`")

func inlineRouterCLICommandPaths(t *testing.T) []documentedInvocation {
	t.Helper()
	suite := skills.FS()
	var found []documentedInvocation
	err := fs.WalkDir(suite, "olares-router", func(path string, entry fs.DirEntry, err error) error {
		if err != nil || entry.IsDir() || filepath.Ext(path) != ".md" {
			return err
		}
		source, err := fs.ReadFile(suite, path)
		if err != nil {
			return err
		}
		for lineNumber, line := range strings.Split(string(source), "\n") {
			for _, match := range inlineCodePattern.FindAllStringSubmatch(line, -1) {
				fields := strings.Fields(strings.TrimSpace(match[1]))
				if len(fields) < 2 || fields[0] != "olares-cli" {
					continue
				}
				fields = commandLikePrefix(fields[1:])
				if len(fields) == 0 {
					continue
				}
				for _, tokens := range expandCommandAlternatives(fields) {
					found = append(found, documentedInvocation{
						tokens: tokens,
						file:   path,
						line:   lineNumber + 1,
					})
				}
			}
		}
		return nil
	})
	if err != nil {
		t.Fatalf("walk embedded router skill: %v", err)
	}
	if len(found) < 10 {
		t.Fatalf("harvested only %d inline router CLI commands; the docs or scanner moved", len(found))
	}
	return found
}

func inlineRouterCommandPaths(t *testing.T, root *cobra.Command) []documentedInvocation {
	t.Helper()
	suite := skills.FS()
	var found []documentedInvocation
	err := fs.WalkDir(suite, ".", func(path string, entry fs.DirEntry, err error) error {
		if err != nil || entry.IsDir() || filepath.Ext(path) != ".md" {
			return err
		}
		if path != "olares-router/SKILL.md" {
			return nil
		}
		source, err := fs.ReadFile(suite, path)
		if err != nil {
			return err
		}
		family := strings.TrimPrefix(strings.Split(path, "/")[0], "olares-")
		familyCommand := childNamed(root, family)
		for lineNumber, line := range strings.Split(string(source), "\n") {
			for _, match := range inlineCodePattern.FindAllStringSubmatch(line, -1) {
				fields := strings.Fields(strings.TrimSpace(match[1]))
				if len(fields) < 2 {
					continue
				}
				if fields[0] == "olares-cli" {
					fields = fields[1:]
				} else if childNamed(root, fields[0]) == nil {
					if familyCommand == nil || childNamed(familyCommand, fields[0]) == nil {
						continue
					}
					fields = append([]string{family}, fields...)
				}
				fields = commandLikePrefix(fields)
				if len(fields) == 0 {
					continue
				}
				for _, tokens := range expandCommandAlternatives(fields) {
					found = append(found, documentedInvocation{
						tokens: tokens,
						file:   path,
						line:   lineNumber + 1,
					})
				}
			}
		}
		return nil
	})
	if err != nil {
		t.Fatalf("walk the embedded suite: %v", err)
	}
	if len(found) < 40 {
		t.Fatalf("harvested only %d inline router commands; the index or scanner moved", len(found))
	}
	return found
}

// documentedCommandPaths is every command path the embedded suite names, in
// any of the three forms it uses: a fenced example, an inline backticked
// invocation, and a verb-index row that drops the tree's own name because
// the heading already carried it. Every prefix counts, so documenting
// `router local status` documents `router local` as well.
//
// The reverse coverage check reads this rather than the hand-maintained
// skillCommandPaths, because a list maintained by hand answers "was this
// remembered", and what the check needs to ask is "was this documented".
func documentedCommandPaths(t *testing.T, root *cobra.Command) map[string]bool {
	t.Helper()
	documented := map[string]bool{}
	record := func(cmd *cobra.Command, tokens []string) {
		for _, token := range tokens {
			next := childNamed(cmd, token)
			if next == nil {
				return
			}
			cmd = next
			documented[strings.TrimPrefix(cmd.CommandPath(), "olares-cli ")] = true
		}
	}

	for _, invocation := range documentedInvocations(t) {
		record(root, invocation.tokens)
	}

	suite := skills.FS()
	err := fs.WalkDir(suite, ".", func(path string, entry fs.DirEntry, err error) error {
		if err != nil || entry.IsDir() || filepath.Ext(path) != ".md" {
			return err
		}
		source, err := fs.ReadFile(suite, path)
		if err != nil {
			return err
		}
		family := childNamed(root, strings.TrimPrefix(strings.Split(path, "/")[0], "olares-"))
		for _, line := range strings.Split(string(source), "\n") {
			for _, match := range inlineCodePattern.FindAllStringSubmatch(line, -1) {
				fields := strings.Fields(strings.TrimSpace(match[1]))
				if len(fields) == 0 {
					continue
				}
				base := root
				if fields[0] == "olares-cli" {
					fields = fields[1:]
				} else if childNamed(root, fields[0]) == nil {
					// A verb index under `## olares-router` writes `local
					// status`, not the tree it is already inside.
					base = family
				}
				if base == nil {
					continue
				}
				for _, tokens := range expandCommandAlternatives(commandLikePrefix(fields)) {
					record(base, tokens)
				}
			}
		}
		return nil
	})
	if err != nil {
		t.Fatalf("walk the embedded suite: %v", err)
	}
	return documented
}

// commandNameFields truncates at the first word that cannot be a cobra
// command name. `commandLikePrefix` stops at the placeholder syntax an
// author writes deliberately; this stops at what a quoted sentence looks
// like — capitals, punctuation, quotes — which is what separates
// `job.Get` and `source 'upload'` from a command path.
var commandNamePattern = regexp.MustCompile(`^[a-z][a-z0-9-]*$`)

func commandNameFields(fields []string) []string {
	for i, field := range fields {
		if i == 0 && field == "olares-cli" {
			continue
		}
		if !commandNamePattern.MatchString(field) {
			return fields[:i]
		}
	}
	return fields
}

func commandLikePrefix(fields []string) []string {
	for i, field := range fields {
		if strings.HasPrefix(field, "-") || strings.ContainsAny(field, "<>[]") ||
			field == "..." || field == "→" || field == "." {
			return fields[:i]
		}
	}
	return fields
}

func expandCommandAlternatives(fields []string) [][]string {
	for i, field := range fields {
		parts := strings.FieldsFunc(field, func(r rune) bool { return r == '/' || r == '|' })
		if len(parts) < 2 {
			continue
		}
		var expanded [][]string
		for _, part := range parts {
			tokens := append([]string(nil), fields...)
			tokens[i] = part
			expanded = append(expanded, tokens[:i+1])
		}
		return expanded
	}
	return [][]string{fields}
}

type documentedInvocation struct {
	tokens []string
	// bases are the commands the tokens may be relative to, for the
	// scanners that read an index whose rows drop the tree's own name.
	// Empty means the root.
	bases []*cobra.Command
	file  string
	line  int
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
		// Prose names commands it is telling the reader *not* to reach
		// for ("not an `olares-cli publish` lifecycle"), and English
		// sentences put ordinary words after the binary's name. A fenced
		// block is the one place every word is meant to be typed.
		for _, shell := range fencedShellLines(string(source)) {
			line, lineNumber := shell.text, shell.number
			// A fenced block is not always shell — it also holds YAML, JSON
			// and output samples, any of which can mention the binary in
			// prose. A command to be typed starts its line.
			if !strings.HasPrefix(line, "olares-cli") {
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
					line:   lineNumber,
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
