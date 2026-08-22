package skills

import (
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"

	skillsuite "github.com/beclab/Olares/cli/skills"
)

// SilenceNoticeEnv turns the staleness notice off for callers that have
// decided not to install — a pinned image, a CI job, a machine where the
// skills are managed by something else.
const SilenceNoticeEnv = "OLARES_CLI_NO_SKILL_NOTICE"

// developmentVersion is what version.VERSION says when nothing was injected
// at build time. A local build's version has no relationship to a release,
// so comparing skills against it would report drift on every run.
const developmentVersion = "development"

// probe is the skill the fallback check reads. Every install writes all
// twelve together, so one is enough to tell whether they came from this
// binary, and reading twelve files on every invocation of every verb would
// not be. olares-shared is the one an agent is always pointed at first.
const probe = "olares-shared"

// Notice writes one line when the skills installed on this machine did not
// come from this binary.
//
// This is the failure embedding was meant to remove, arriving by the one
// route embedding leaves open: the skills are on disk, and upgrading
// olares-cli does not rewrite them. An agent then follows instructions for a
// version it is not running, and nothing in either the skill or the binary
// says so — the skill declares `requires.bins: [olares-cli]`, which is
// satisfied. So it is said out loud, with the command that fixes it.
//
// The comparison is over content, not over the version the skills declare.
// A version is bumped by a human every few weeks while a daily build ships
// whatever is on main, so within one release the label is stable and the
// skills are not: comparing labels is silent in exactly the case an agent is
// most likely to be reading something that has moved. Content also settles
// the reverse case, where two channels ship the same skills under different
// labels and a label comparison would have them accuse each other forever.
//
// Silence is the answer to every uncertainty: no skills installed, an
// unreadable copy, no home directory, a development build. A notice that
// fires when nothing is wrong is a notice people learn to skip.
func Notice(w io.Writer, binaryVersion string) {
	if os.Getenv(SilenceNoticeEnv) != "" {
		return
	}
	if strings.Contains(binaryVersion, developmentVersion) {
		return
	}
	store := StorePath()
	if store == "" {
		return
	}
	if complaint, ok := staleness(store); ok {
		fmt.Fprintf(w, "olares-cli: the agent skills in %s %s. Run `olares-cli skills install` "+
			"so they describe the verbs you actually have (set %s=1 to stop saying this).\n",
			store, complaint, SilenceNoticeEnv)
	}
}

// staleness describes what is wrong with the copy in store, or reports that
// there is nothing to say.
func staleness(store string) (string, bool) {
	carried, err := skillsuite.Digest()
	if err != nil {
		return "", false
	}
	installed, ok := skillsuite.ReadIdentity(store)
	if !ok {
		// No marker: written either by hand or by an olares-cli from before
		// there was one. Comparing the declared versions is what this check
		// used to do, and it is still worth doing — an upgrade across that
		// boundary is the common case, and the first install afterwards
		// leaves a marker.
		return legacyStaleness(store)
	}
	if installed.Digest == carried {
		return "", false
	}
	mine, err := skillsuite.SuiteVersion()
	if err != nil {
		return "", false
	}
	// Naming the same version twice reads as a bug rather than a diagnosis,
	// so when the labels agree the difference is stated as what it is. This
	// is the ordinary case on the daily channel.
	if installed.Version == mine {
		return fmt.Sprintf("are a different copy of %s than this build carries", mine), true
	}
	return fmt.Sprintf("were written by a different olares-cli (they declare %s, this build carries %s)",
		installed.Version, mine), true
}

// legacyStaleness compares the versions two SKILL.md files declare, for a
// store written before the marker existed.
func legacyStaleness(store string) (string, bool) {
	installed, ok := installedVersion(filepath.Join(store, probe, skillsuite.EntryFile))
	if !ok {
		return "", false
	}
	embedded, err := skillsuite.Describe(probe)
	if err != nil || embedded.Version == "" || embedded.Version == installed {
		return "", false
	}
	// Both numbers are skill versions, and saying so matters: a release
	// stamps them with the CLI's own version, but a copy installed some
	// other way carries whatever it carried, and calling that "the binary's
	// version" would be wrong in exactly the case this fires.
	return fmt.Sprintf("were written by a different olares-cli (they declare %s, this build carries %s)",
		installed, embedded.Version), true
}

func installedVersion(path string) (string, bool) {
	source, err := os.ReadFile(path)
	if err != nil {
		return "", false
	}
	meta, err := skillsuite.ParseMeta(source)
	if err != nil || meta.Version == "" {
		return "", false
	}
	return meta.Version, true
}
