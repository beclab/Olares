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

// probe is the skill the check reads. Every install writes all twelve
// together, so one is enough to tell whether they came from this binary, and
// reading twelve files on every invocation of every verb would not be.
// olares-shared is the one an agent is always pointed at first.
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
	installed, ok := installedVersion(filepath.Join(store, probe, skillsuite.EntryFile))
	if !ok {
		return
	}
	embedded, err := skillsuite.Describe(probe)
	if err != nil || embedded.Version == "" || embedded.Version == installed {
		return
	}
	// The two numbers are both skill versions, and saying so matters: a
	// release stamps them with the CLI's own version, but a copy installed
	// some other way carries whatever it carried, and calling that "the
	// binary's version" would be wrong in exactly the case this fires.
	fmt.Fprintf(w, "olares-cli: the agent skills in %s were written by a different olares-cli "+
		"(they declare %s, this build carries %s). Run `olares-cli skills install` so they "+
		"describe the verbs you actually have (set %s=1 to stop saying this).\n",
		store, installed, embedded.Version, SilenceNoticeEnv)
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
