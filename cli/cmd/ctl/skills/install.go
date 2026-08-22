package skills

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/spf13/cobra"

	skillsuite "github.com/beclab/Olares/cli/skills"
)

type installEnvelope struct {
	Skills  []string      `json:"skills"`
	Count   int           `json:"count"`
	Store   string        `json:"store"`
	Linked  []string      `json:"linked"`
	Copied  []string      `json:"copied"`
	Skipped []skippedPath `json:"skipped"`
}

// copyStagingPrefix names the directory copyInto exports into before moving
// the skills in. Dotted, like the one Export uses internally, so whatever
// reads an agent's skills directory mid-copy does not find a half-written
// skill sitting beside the finished ones.
const copyStagingPrefix = ".olares-cli-copy-"

type skippedPath struct {
	Path   string `json:"path"`
	Reason string `json:"reason"`
	// HandMade separates the skips --force would take over from the ones it
	// would not: somebody else's copy is not ours to replace at any flag.
	HandMade bool `json:"hand_made"`
}

func newInstallCommand() *cobra.Command {
	opts := &outputOptions{Output: "table"}
	var force bool
	cmd := &cobra.Command{
		Use:   "install",
		Short: "Install the embedded suite where this machine's agents look",
		Long: `Write the suite this binary carries to disk for the agents on this machine.

The bytes come from the binary, so an install cannot fetch a version that
disagrees with the verbs available here, and it needs neither GitHub nor Node.

It writes one copy under ~/.agents/skills — which is the shared location
agents read — and links it from each agent's own skills directory that
already exists. Directories are never created for an agent that is not
installed here.

Run it again after upgrading olares-cli to bring the skills along with it.`,
		Example: `  olares-cli skills install
  olares-cli skills install -o json`,
		Args: cobra.NoArgs,
		RunE: func(cmd *cobra.Command, args []string) error {
			return runInstall(opts, force)
		},
	}
	addOutputFlags(cmd, opts)
	cmd.Flags().BoolVar(&force, "force", false,
		"install over olares-* links that were made by hand, in the store or in an agent's directory")
	return cmd
}

func runInstall(opts *outputOptions, force bool) error {
	store := StorePath()
	if store == "" {
		return fmt.Errorf("cannot find your home directory, so there is nowhere to install; " +
			"name a directory instead: olares-cli skills export <dir>")
	}
	// The store holds the one copy every agent directory links to, so a
	// hand-made link here decides the whole install rather than one agent's
	// worth of it. Export refuses to write where a link is, which is what
	// keeps an export nobody asked for away from a checkout; --force is that
	// ask, so clear them here rather than teaching Export an exception.
	if links := handMadeLinksIn(store); len(links) > 0 {
		if !force {
			return refuseHandMadeLinks(links)
		}
		for _, link := range links {
			// Remove, not RemoveAll: this unlinks the link and never
			// descends into the checkout it points at.
			if err := os.Remove(link.Path); err != nil {
				return fmt.Errorf("remove %s: %w", link.Path, err)
			}
		}
	}

	written, err := skillsuite.Export(store)
	if err != nil {
		return err
	}

	result := installEnvelope{Skills: written, Count: len(written), Store: store}
	for _, dir := range agentSkillDirs() {
		if filepath.Clean(dir) == filepath.Clean(store) {
			continue
		}
		switch outcome := link(dir, store, written, force); outcome.kind {
		case linked:
			result.Linked = append(result.Linked, dir)
		case copied:
			result.Copied = append(result.Copied, dir)
		default:
			result.Skipped = append(result.Skipped, skippedPath{
				Path:     dir,
				Reason:   outcome.reason,
				HandMade: outcome.handMade,
			})
		}
	}

	if opts.isJSON() {
		return opts.printJSON(result)
	}
	if opts.Quiet {
		return nil
	}
	report(result)
	return nil
}

type outcomeKind int

const (
	skipped outcomeKind = iota
	linked
	copied
)

type outcome struct {
	kind     outcomeKind
	reason   string
	handMade bool
}

// link points one agent's skills directory at the store.
//
// The layout — a copy in the store, a relative link to it from each agent —
// is the one the skills CLI established, and following it means a machine
// that has used either tool has one copy of each skill rather than two
// competing ones. The link is relative so a home directory that moves, or is
// mounted somewhere else inside a container, does not leave every agent
// pointing at a path that no longer exists.
//
// Anything already there that is neither absent nor a link into the store is
// left alone and reported. It is somebody else's copy — a hand-placed
// directory, or an install made with copies instead of links — and replacing
// it would be the silent overwrite this command refuses to do elsewhere.
// Every entry is examined before any is touched, so a directory this cannot
// finish is a directory it never started on.
//
// The skip is one directory's, not the install's. An agent pointed at a
// checkout by hand is a decision about that agent, and failing the whole
// command over it would leave every other agent on the machine without the
// update — which is the same silence, arrived at from the other side.
func link(dir, store string, names []string, force bool) outcome {
	for _, name := range names {
		kind, reason := classify(filepath.Join(dir, name), store)
		if kind == entryForeign {
			return outcome{kind: skipped, reason: reason}
		}
		if kind == entryHandMade && !force {
			return outcome{kind: skipped, reason: reason, handMade: true}
		}
	}
	for _, name := range names {
		path := filepath.Join(dir, name)
		if err := os.RemoveAll(path); err != nil {
			return outcome{kind: skipped, reason: fmt.Sprintf("cannot replace %s: %v", path, err)}
		}
		// Rel fails when the two paths share no root — on Windows, a store
		// and an agent directory on different drives. The absolute target
		// works today and gives up only the property above: move the home
		// directory and this one link stops resolving.
		target, err := filepath.Rel(dir, filepath.Join(store, name))
		if err != nil {
			target = filepath.Join(store, name)
		}
		if err := os.Symlink(target, path); err != nil {
			// Windows refuses symlinks to an unprivileged process unless
			// developer mode is on, and a machine that cannot link still has
			// agents that need the files. Write them in instead — the
			// remaining entries are gone either way, since Export replaces
			// each skill whole.
			return copyInto(dir, names, err)
		}
	}
	return outcome{kind: linked}
}

// copyInto writes the suite into an agent's own directory, for the machine
// that cannot make a symbolic link.
//
// It exports to a staging directory first and moves each skill in afterwards,
// rather than exporting straight into dir. Two reasons, and the second is the
// one that matters: Export refuses to write where a symlink is, and by the
// time this is reached some of dir's entries are the links link() just made;
// and clearing them to make room would leave the directory empty for as long
// as the export took, or forever if the export failed. Nothing is removed
// here until the bytes that replace it are already on this filesystem.
func copyInto(dir string, names []string, linkErr error) outcome {
	staging, err := os.MkdirTemp(dir, copyStagingPrefix+"*")
	if err != nil {
		return outcome{kind: skipped, reason: fmt.Sprintf("cannot link (%v) and cannot stage a copy (%v)", linkErr, err)}
	}
	defer os.RemoveAll(staging)

	if _, err := skillsuite.Export(staging); err != nil {
		return outcome{kind: skipped, reason: fmt.Sprintf("cannot link (%v) and cannot copy (%v)", linkErr, err)}
	}
	for _, name := range names {
		target := filepath.Join(dir, name)
		if err := os.RemoveAll(target); err != nil {
			return outcome{kind: skipped, reason: fmt.Sprintf("cannot link (%v) and cannot replace %s (%v)", linkErr, target, err)}
		}
		if err := os.Rename(filepath.Join(staging, name), target); err != nil {
			return outcome{kind: skipped, reason: fmt.Sprintf("cannot link (%v) and cannot move %s into place (%v)", linkErr, name, err)}
		}
	}
	return outcome{kind: copied}
}

type entryKind int

const (
	entryAbsent entryKind = iota
	entryStoreLink
	entryHandMade
	entryForeign
)

func classify(path, store string) (entryKind, string) {
	info, err := os.Lstat(path)
	if err != nil {
		if os.IsNotExist(err) {
			return entryAbsent, ""
		}
		return entryForeign, fmt.Sprintf("cannot inspect %s: %v", path, err)
	}
	if info.Mode()&os.ModeSymlink != 0 {
		target, err := os.Readlink(path)
		if err != nil {
			return entryForeign, fmt.Sprintf("cannot read the link at %s: %v", path, err)
		}
		if pointsInto(path, target, store) {
			return entryStoreLink, ""
		}
		return entryHandMade, fmt.Sprintf("%s is linked to %s by hand", path, target)
	}
	return entryForeign, fmt.Sprintf("%s already exists and is not a link into %s", path, store)
}

func report(result installEnvelope) {
	fmt.Printf("wrote %d skills to %s\n", result.Count, result.Store)
	for _, dir := range result.Linked {
		fmt.Printf("  linked  %s\n", dir)
	}
	for _, dir := range result.Copied {
		fmt.Printf("  copied  %s\n", dir)
	}
	handMade := false
	for _, skip := range result.Skipped {
		fmt.Printf("  skipped %s: %s\n", skip.Path, skip.Reason)
		handMade = handMade || skip.HandMade
	}
	if handMade {
		// The link is how a checkout gets read live, so it is left where it
		// is. Saying why matters more here than in the store's refusal: this
		// one is a line in the middle of a report that otherwise succeeded.
		fmt.Println("\nThe directories linked by hand were left alone. An agent reading one sees")
		fmt.Println("the directory the link names — a git checkout, for the local development")
		fmt.Println("setup in cli/README.md — and installing over it would end that without")
		fmt.Println("saying so. Remove the link, or pass --force, to get the installed copies.")
	}
	if len(result.Linked) == 0 && len(result.Copied) == 0 {
		// Not a failure: several agents read the shared directory directly,
		// and an agent that has never run has no directory to link. Saying so
		// is better than a silent success that looks like nothing happened.
		fmt.Println("\nNo agent directory on this machine needed a link. Agents that read")
		fmt.Println("~/.agents/skills find the suite there; for one that keeps its own")
		fmt.Println("directory, run it once and then run this again.")
	}
}

// refuseHandMadeLinks reports links in the store, which is the one place a
// hand-made link stops the whole install: every agent directory links into
// it, so until these are dealt with there is nothing to link to.
func refuseHandMadeLinks(links []handMadeLink) error {
	var message strings.Builder
	message.WriteString("refusing to install over skills in the store that were linked by hand:\n")
	for _, link := range links {
		fmt.Fprintf(&message, "  %s -> %s\n", link.Path, link.Target)
	}
	message.WriteString("\nEvery agent directory links into the store, so installing would replace these\n" +
		"with the copies from this binary and what agents read would stop being the\n" +
		"directories they name — without saying so. That layout is the local\n" +
		"development setup in cli/README.md.\n\n" +
		"To keep editing in place, leave them: an agent reading them already sees your\n" +
		"working tree. To install anyway, remove them, or pass --force.")
	return fmt.Errorf("%s", message.String())
}
