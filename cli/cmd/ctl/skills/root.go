// Package skills implements the `olares-cli skills` subtree, which serves the
// skill suite compiled into this binary.
//
// Every verb here is local: no profile, no network, no cluster. That is what
// makes the tree the answer to "which instructions match this binary" —
// reading them cannot depend on anything that could be a different version.
package skills

import (
	"encoding/json"
	"fmt"
	"io/fs"
	"os"
	"path"
	"sort"
	"strings"

	"github.com/spf13/cobra"

	skillsuite "github.com/beclab/Olares/cli/skills"
)

// NewSkillsCommand assembles the `olares-cli skills` subtree.
func NewSkillsCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "skills",
		Short: "Agent skills that ship with this binary",
		Long: `The olares-* agent skills compiled into this olares-cli.

An agent reads these to learn which verbs exist and how to drive them, so a
skill from one release and a binary from another disagree silently. Reading
them from the binary removes that gap: what these verbs print is what this
build was made from.

  list / read   read the suite without writing anything
  export        write it to a directory
  install       write it where the agents on this machine look`,
	}
	cmd.SilenceErrors = true
	cmd.SilenceUsage = true
	cmd.PersistentPreRun = func(c *cobra.Command, args []string) {
		c.SilenceErrors = true
		c.SilenceUsage = true
	}
	cmd.AddCommand(newListCommand(), newReadCommand(), newExportCommand(), newInstallCommand())
	return cmd
}

// output flags follow the rest of the CLI: -o/--output, -q/--quiet. The
// skills document `-o json` on every read verb, so a tree that spelled it
// `--json` would be the one place the documented shape does not hold.
type outputOptions struct {
	Output string
	Quiet  bool
}

func (o *outputOptions) isJSON() bool {
	return strings.EqualFold(strings.TrimSpace(o.Output), "json")
}

func (o *outputOptions) printJSON(v interface{}) error {
	if o.Quiet {
		return nil
	}
	encoder := json.NewEncoder(os.Stdout)
	encoder.SetIndent("", "  ")
	return encoder.Encode(v)
}

type listEnvelope struct {
	Skills []skillListItem `json:"skills"`
	Count  int             `json:"count"`
}

type skillListItem struct {
	Name        string `json:"name"`
	Version     string `json:"version"`
	Description string `json:"description"`
}

type listPathEnvelope struct {
	Path    string      `json:"path"`
	Entries []pathEntry `json:"entries"`
	Count   int         `json:"count"`
}

type pathEntry struct {
	Name string `json:"name"`
	Dir  bool   `json:"dir"`
}

func newListCommand() *cobra.Command {
	opts := &outputOptions{Output: "table"}
	cmd := &cobra.Command{
		Use:   "list [<skill>[/<path>]]",
		Short: "List the embedded skills, or one layer inside one",
		Long: `With no argument, list every skill in this binary.

With a skill (or a path inside one), list what that directory holds — the
step between reading a SKILL.md and reading the references it links to,
which have no presence on disk to browse.`,
		Example: `  olares-cli skills list
  olares-cli skills list -o json
  olares-cli skills list olares-router
  olares-cli skills list olares-router/references`,
		Args: cobra.MaximumNArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			if len(args) == 1 {
				return runListPath(opts, args[0])
			}
			return runList(opts)
		},
	}
	addOutputFlags(cmd, opts)
	return cmd
}

func runList(opts *outputOptions) error {
	metas, err := skillsuite.List()
	if err != nil {
		return err
	}
	items := make([]skillListItem, 0, len(metas))
	for _, meta := range metas {
		items = append(items, skillListItem{
			Name:        meta.Name,
			Version:     meta.Version,
			Description: meta.Description,
		})
	}
	if opts.isJSON() {
		return opts.printJSON(listEnvelope{Skills: items, Count: len(items)})
	}
	if opts.Quiet {
		return nil
	}
	// Both columns are measured. A version is x.y.z-cli.n, which is eleven
	// or twelve characters, so a fixed width narrow enough to look tidy in
	// the source pushes every description out of line instead.
	names, versions := 0, 0
	for _, item := range items {
		if len(item.Name) > names {
			names = len(item.Name)
		}
		if len(item.Version) > versions {
			versions = len(item.Version)
		}
	}
	for _, item := range items {
		// The description is a paragraph written for matching a request
		// against, not for a column. One line of it is enough to choose
		// between twelve skills; `read` is where the whole thing lives.
		fmt.Printf("%-*s  %-*s  %s\n", names, item.Name, versions, item.Version, summarize(item.Description))
	}
	return nil
}

func runListPath(opts *outputOptions, target string) error {
	target = strings.Trim(path.Clean(strings.TrimSpace(target)), "/")
	if target == "" || target == "." || strings.HasPrefix(target, "..") {
		return fmt.Errorf("%q does not name a skill", target)
	}
	entries, err := fs.ReadDir(skillsuite.FS(), target)
	if err != nil {
		return fmt.Errorf("list %s: %w", target, err)
	}
	items := make([]pathEntry, 0, len(entries))
	for _, entry := range entries {
		items = append(items, pathEntry{Name: entry.Name(), Dir: entry.IsDir()})
	}
	sort.Slice(items, func(i, j int) bool { return items[i].Name < items[j].Name })

	if opts.isJSON() {
		return opts.printJSON(listPathEnvelope{Path: target, Entries: items, Count: len(items)})
	}
	if opts.Quiet {
		return nil
	}
	for _, item := range items {
		name := item.Name
		if item.Dir {
			name += "/"
		}
		fmt.Println(name)
	}
	return nil
}

type readEnvelope struct {
	Skill   string `json:"skill"`
	Path    string `json:"path"`
	Content string `json:"content"`
}

func newReadCommand() *cobra.Command {
	opts := &outputOptions{Output: "raw"}
	cmd := &cobra.Command{
		Use:   "read <skill>[/<path>] [path]",
		Short: "Print a skill's SKILL.md, or a file under it",
		Long: `Print embedded skill content. Naming only a skill prints its SKILL.md.

Default output is the file's bytes, unchanged, so it can be piped or diffed
against a checkout. -o json wraps them in an envelope instead.`,
		Example: `  olares-cli skills read olares-router
  olares-cli skills read olares-router references/olares-router-calling.md
  olares-cli skills read olares-router/references/olares-router-calling.md
  olares-cli skills read olares-router -o json`,
		Args: cobra.RangeArgs(1, 2),
		RunE: func(cmd *cobra.Command, args []string) error {
			skill, relative := splitTarget(args)
			return runRead(opts, skill, relative)
		},
	}
	cmd.Flags().StringVarP(&opts.Output, "output", "o", "raw", "output format: raw, json")
	cmd.Flags().BoolVarP(&opts.Quiet, "quiet", "q", false, "suppress output; exit code indicates success/failure")
	return cmd
}

func runRead(opts *outputOptions, skill, relative string) error {
	content, err := skillsuite.Read(skill, relative)
	if err != nil {
		return err
	}
	if relative == "" {
		relative = skillsuite.EntryFile
	}
	if opts.isJSON() {
		return opts.printJSON(readEnvelope{Skill: skill, Path: relative, Content: string(content)})
	}
	if opts.Quiet {
		return nil
	}
	if _, err := os.Stdout.Write(content); err != nil {
		return fmt.Errorf("write content: %w", err)
	}
	// A SKILL.md links its references with relative markdown paths, which
	// resolve to nothing for a reader who was handed the bytes rather than a
	// directory. Said once, on stderr, so stdout stays the file.
	if relative == skillsuite.EntryFile {
		fmt.Fprintf(os.Stderr,
			"\n> This skill's links are relative: read them with `olares-cli skills read %s <path>`, or `olares-cli skills install` to put the suite where your agents look.\n",
			skill)
	}
	return nil
}

// splitTarget maps one or two positional arguments onto a skill and a path
// inside it. One argument may carry both, split at the first slash, so a path
// copied out of a reference link works as written.
func splitTarget(args []string) (skill, relative string) {
	if len(args) == 2 {
		return args[0], args[1]
	}
	if index := strings.Index(args[0], "/"); index >= 0 {
		return args[0][:index], args[0][index+1:]
	}
	return args[0], ""
}

func addOutputFlags(cmd *cobra.Command, opts *outputOptions) {
	cmd.Flags().StringVarP(&opts.Output, "output", "o", "table", "output format: table, json")
	cmd.Flags().BoolVarP(&opts.Quiet, "quiet", "q", false, "suppress output; exit code indicates success/failure")
}

// summarize trims a description down to something that fits a column.
//
// These descriptions open with the product and the verb ("Olares Dashboard
// via olares-cli dashboard") and only then say what the skill covers, so a
// cut at the first punctuation leaves every row reading as its own name.
// Take the first sentence, and if that is still a paragraph, as much of it
// as fits.
const summaryWidth = 96

func summarize(description string) string {
	description = strings.Join(strings.Fields(description), " ")
	if index := strings.Index(description, ". "); index > 0 {
		description = description[:index+1]
	}
	if len(description) <= summaryWidth {
		return description
	}
	cut := strings.LastIndex(description[:summaryWidth], " ")
	if cut <= 0 {
		cut = summaryWidth
	}
	return description[:cut] + "…"
}
