package router

import (
	"sort"
	"strings"
	"testing"

	"github.com/spf13/cobra"

	"github.com/beclab/Olares/cli/skills"
)

// The olares-router skill and this command tree ship in one binary, and nothing
// until now held them to each other. `skills/validate.py` checks structure,
// links and length — none of which notices a verb that was renamed, removed, or
// added and never written down.
//
// Both directions matter and they fail differently. A verb the documentation
// names and the tree does not have sends an agent to run a command that does not
// exist; a verb the tree has and the documentation never names is a capability
// nobody will reach for. The first produced a real report — `local endpoints`
// survived in the diagnosis reference for releases after the verb became `model
// diag endpoints`.
const routerSkill = "olares-router"

func TestEveryVerbTheSkillNamesExists(t *testing.T) {
	root := NewRouterCommand(nil)
	for _, doc := range skillDocs(t) {
		for _, span := range codeSpans(doc.text) {
			for _, path := range commandPathsIn(root, span) {
				written := strings.Join(path.words, " ")
				switch {
				case path.node == nil:
					t.Errorf("%s writes `router %s`, which no command answers to", doc.name, written)
				case path.hidden:
					// A hidden command still answers, so this is not a broken
					// example — it is the old spelling being taught, which is
					// how `local endpoints` outlived `model diag endpoints`.
					t.Errorf("%s writes `router %s`, which is the retired spelling of `%s`",
						doc.name, written, path.node.CommandPath())
				}
			}
		}
	}
}

func TestEveryVerbTheTreeOffersIsWrittenDown(t *testing.T) {
	root := NewRouterCommand(nil)
	named := map[string]bool{}
	for _, doc := range skillDocs(t) {
		for _, span := range codeSpans(doc.text) {
			for _, path := range commandPathsIn(root, span) {
				if path.node != nil {
					named[path.node.CommandPath()] = true
				}
			}
		}
	}
	for _, leaf := range leafVerbs(root) {
		if !named[leaf.CommandPath()] {
			t.Errorf("`%s` is callable and the olares-router skill never mentions it",
				leaf.CommandPath())
		}
	}
}

// Both tests above pass on their own if the walk finds nothing at all, so what
// it finds is pinned here. A parser that quietly stopped matching would report
// a documented tree and an undocumented one as equally fine.
func TestTheWalkReadsWhatTheDocumentationWrites(t *testing.T) {
	root := NewRouterCommand(nil)
	cases := map[string]struct {
		found   []string
		unknown []string
		retired []string
	}{
		"olares-cli router usage list --since 7d": {found: []string{"router usage list"}},
		"router provider list/get/create": {found: []string{
			"router provider list", "router provider get", "router provider create",
		}},
		"model retry <model>":            {found: []string{"router model retry"}},
		"router call music --model x/y":  {found: []string{"router call music"}},
		"router local endpoints":         {retired: []string{"router local endpoints"}},
		"router lcoal endpoints":         {unknown: []string{"lcoal"}},
		"the router is one per Olares":   {},
		"router model spec show <model>": {found: []string{"router model spec show"}},
	}
	for span, want := range cases {
		var found, unknown, retired []string
		for _, path := range commandPathsIn(root, span) {
			switch {
			case path.node == nil:
				unknown = append(unknown, strings.Join(path.words, " "))
			case path.hidden:
				retired = append(retired, path.node.CommandPath())
			default:
				found = append(found, path.node.CommandPath())
			}
		}
		for _, wanted := range want.found {
			if !containsString(found, wanted) {
				t.Errorf("%q: expected to read %q, read %v", span, wanted, found)
			}
		}
		if len(want.found) > 0 && len(found) == 0 {
			t.Errorf("%q: read no command at all", span)
		}
		if len(want.found) == 0 && len(found) > 0 {
			t.Errorf("%q: read %v, which is prose", span, found)
		}
		for _, wanted := range want.unknown {
			if !containsString(unknown, wanted) {
				t.Errorf("%q: expected %q to be reported as unknown, got %v", span, wanted, unknown)
			}
		}
		if len(want.unknown) == 0 && len(unknown) > 0 {
			t.Errorf("%q: reported %v as unknown verbs", span, unknown)
		}
		for _, wanted := range want.retired {
			if !containsString(retired, wanted) {
				t.Errorf("%q: expected %q to read as retired, got %v", span, wanted, retired)
			}
		}
		if len(want.retired) == 0 && len(retired) > 0 {
			t.Errorf("%q: reported %v as retired", span, retired)
		}
	}
}

type skillDoc struct {
	name string
	text string
}

// skillDocs reads the shipped bytes rather than the directory, so the test is
// about what an agent will be handed.
func skillDocs(t *testing.T) []skillDoc {
	t.Helper()
	files, err := skills.Files()
	if err != nil {
		t.Fatalf("skills.Files: %v", err)
	}
	var docs []skillDoc
	for _, file := range files {
		if !strings.HasPrefix(file, routerSkill+"/") || !strings.HasSuffix(file, ".md") {
			continue
		}
		name := strings.TrimPrefix(file, routerSkill+"/")
		source, err := skills.Read(routerSkill, name)
		if err != nil {
			t.Fatalf("skills.Read %s: %v", file, err)
		}
		docs = append(docs, skillDoc{name: name, text: string(source)})
	}
	if len(docs) < 2 {
		t.Fatalf("expected the skill and its references, found %d files", len(docs))
	}
	return docs
}

type mentionedPath struct {
	words []string
	// node is the command the words reached, or nil when the deepest word
	// names nothing — which is the failure the first test reports.
	node *cobra.Command
	// hidden is set when the path, or anything it descended through, is
	// hidden: an old spelling that still runs.
	hidden bool
}

// codeSpans returns what the document typeset as a command: every line of a
// fenced block, and every inline backtick span elsewhere. Restricting the walk
// to these is what lets the first test call an unknown word a broken command
// rather than a sentence about routers.
func codeSpans(text string) []string {
	var (
		spans  []string
		fenced bool
	)
	for _, line := range strings.Split(text, "\n") {
		if strings.HasPrefix(strings.TrimSpace(line), "```") {
			fenced = !fenced
			continue
		}
		if fenced {
			spans = append(spans, line)
			continue
		}
		parts := strings.Split(line, "`")
		for i := 1; i < len(parts); i += 2 {
			spans = append(spans, parts[i])
		}
	}
	return spans
}

// commandPathsIn walks one code span for command paths.
//
// A walk starts at the word `router` or at one of its own children, because the
// documentation spells both — a reference already inside a section about models
// writes `model retry <model>`. It ends at the first word that is not a
// subcommand: a flag, a placeholder, or an argument. A word carrying slashes is
// the documentation's own shorthand for siblings (`list/get/create`), so each
// alternative is looked up at the same level.
//
// A word after `router` that reads like a verb and names nothing is returned
// with a nil node. Reading that as a broken command is only safe where a
// command would be written, so the `router` branch is taken at the start of a
// span or after `olares-cli` and nowhere else — otherwise the `router` Market
// listing, mentioned in prose inside backticks, would have every word that
// follows it read as a verb.
func commandPathsIn(root *cobra.Command, span string) []mentionedPath {
	var found []mentionedPath
	words := strings.FieldsFunc(span, func(r rune) bool {
		switch r {
		case ' ', '\t', '"', ',', ';', '(', ')', '[', ']', '|', '*', '\'':
			return true
		}
		return false
	})
	for i := 0; i < len(words); i++ {
		start := words[i]
		if start == "router" && (i == 0 || words[i-1] == "olares-cli") {
			paths, used := walkFrom(root, words[i+1:])
			if used == 0 {
				if i+1 < len(words) && looksLikeVerb(words[i+1]) {
					found = append(found, mentionedPath{words: []string{words[i+1]}})
				}
				continue
			}
			found = append(found, paths...)
			i += used
			continue
		}
		// A span that opens with a verb of its own — `model retry <model>`,
		// written inside a section already about models. Only at the start:
		// mid-span, the word after a flag is that flag's value, and `--kind
		// default` would otherwise read as the retired `router default`.
		//
		// A hidden name does not open one. `spec` and `list` are both a
		// retired top-level group and an ordinary English word the references
		// use for the live `model spec` and `usage list` families, and the
		// retired reading is only unambiguous when `router` was written out.
		if child := childNamed(root, start); i == 0 && child != nil && !child.Hidden {
			if paths, used := walkFrom(root, words); used > 0 {
				found = append(found, paths...)
				i += used - 1
			}
		}
	}
	return found
}

// looksLikeVerb keeps flags, placeholders, versions and paths out of the
// stale-spelling report.
func looksLikeVerb(word string) bool {
	if word == "" {
		return false
	}
	for _, r := range word {
		if (r < 'a' || r > 'z') && r != '-' {
			return false
		}
	}
	return true
}

// walkFrom descends as far as the words go, returning one entry per alternative
// the documentation listed and how many words were consumed.
func walkFrom(root *cobra.Command, words []string) ([]mentionedPath, int) {
	var (
		out     []mentionedPath
		current = root
		prefix  []string
		used    int
		retired bool
	)
	for _, word := range words {
		if word == "" || strings.HasPrefix(word, "-") || strings.HasPrefix(word, "<") ||
			strings.HasPrefix(word, "{") || strings.Contains(word, "=") {
			break
		}
		alternatives := strings.Split(word, "/")
		first := childNamed(current, alternatives[0])
		if first == nil {
			break
		}
		used++
		retired = retired || first.Hidden
		for _, alt := range alternatives {
			node := childNamed(current, alt)
			out = append(out, mentionedPath{
				words:  append(append([]string{}, prefix...), alt),
				node:   node,
				hidden: retired || (node != nil && node.Hidden),
			})
		}
		prefix = append(prefix, alternatives[0])
		current = first
	}
	return out, used
}

func childNamed(parent *cobra.Command, name string) *cobra.Command {
	if parent == nil || name == "" {
		return nil
	}
	for _, child := range parent.Commands() {
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

// leafVerbs are the commands a person can actually run: the ones with no
// subcommands of their own, minus the hidden ones. A hidden command is either an
// old spelling kept working on purpose or something not offered yet, and neither
// belongs in a document an agent reads.
func leafVerbs(root *cobra.Command) []*cobra.Command {
	var out []*cobra.Command
	var walk func(cmd *cobra.Command)
	walk = func(cmd *cobra.Command) {
		children := cmd.Commands()
		runnable := make([]*cobra.Command, 0, len(children))
		for _, child := range children {
			if child.Hidden || child.Name() == "help" || child.Name() == "completion" {
				continue
			}
			runnable = append(runnable, child)
		}
		if len(runnable) == 0 && cmd != root {
			out = append(out, cmd)
			return
		}
		for _, child := range runnable {
			walk(child)
		}
	}
	walk(root)
	sort.Slice(out, func(i, j int) bool { return out[i].CommandPath() < out[j].CommandPath() })
	return out
}
