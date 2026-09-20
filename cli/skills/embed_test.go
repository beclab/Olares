package skills

import (
	"io/fs"
	"os"
	"path/filepath"
	"regexp"
	"strings"
	"testing"
)

// The embed directive names three patterns, which is a whitelist, and a
// whitelist quietly ships less than it claims the moment a skill grows a
// kind of file nobody listed. So this does not check that the patterns
// match something — it checks that what they match is the whole suite.
// Adding olares-<x>/assets/logo.png fails here rather than reaching a user
// as a reference link to a file that is not in their copy.
func TestEmbedCoversTheWholeSuite(t *testing.T) {
	embedded, err := Files()
	if err != nil {
		t.Fatalf("list embedded files: %v", err)
	}
	onDisk := suiteOnDisk(t)

	if diff := missing(onDisk, embedded); len(diff) > 0 {
		t.Errorf("on disk but not embedded (%d):\n  %s",
			len(diff), strings.Join(diff, "\n  "))
	}
	if diff := missing(embedded, onDisk); len(diff) > 0 {
		t.Errorf("embedded but not on disk (%d):\n  %s",
			len(diff), strings.Join(diff, "\n  "))
	}
}

// Every skill has to be readable as a skill, not merely present as bytes:
// an agent is handed the frontmatter's name and description, and publish.sh
// reads the version. A SKILL.md whose fence never closes parses as prose and
// would install as a skill with no name.
func TestEverySkillDeclaresItself(t *testing.T) {
	metas, err := List()
	if err != nil {
		t.Fatalf("list skills: %v", err)
	}
	names, err := Names()
	if err != nil {
		t.Fatalf("names: %v", err)
	}
	if len(metas) != len(names) {
		t.Fatalf("List returned %d metas for %d skills", len(metas), len(names))
	}
	for i, meta := range metas {
		dir := names[i]
		if meta.Name != dir {
			t.Errorf("%s: frontmatter name is %q; a skill's directory and name have to agree or an installer writes one and an agent asks for the other", dir, meta.Name)
		}
		if meta.Version == "" {
			t.Errorf("%s: no version in frontmatter", dir)
		}
		if meta.Description == "" {
			t.Errorf("%s: no description in frontmatter; it is what an agent matches a request against", dir)
		}
	}
}

// The suite is one artifact — compiled into one binary, written by one
// command — so it carries one version, and that version is the olares-cli
// release it ships in. Twelve independent numbers made "which release is this
// skill from" a question with twelve answers, and made the startup notice
// about mismatched skills impossible to phrase.
//
// This is stated in Go as well as in validate.py because this is the copy
// that ships: the Python validator runs in CI over the working tree, and a
// pattern that stopped matching a skill would take it out of the binary
// without taking it out of the validator's sight.
func TestTheSuiteShipsOneVersion(t *testing.T) {
	metas, err := List()
	if err != nil {
		t.Fatalf("list skills: %v", err)
	}
	declared := map[string][]string{}
	for _, meta := range metas {
		declared[meta.Version] = append(declared[meta.Version], meta.Name)
	}
	if len(declared) != 1 {
		for version, skills := range declared {
			t.Errorf("%s: %s", version, strings.Join(skills, ", "))
		}
		t.Fatalf("the suite declares %d versions", len(declared))
	}
	for version := range declared {
		if !releaseVersion.MatchString(version) {
			t.Errorf("version %q is not an olares-cli release (x.y.z-cli.n); "+
				"a number of the suite's own is a second thing to bump", version)
		}
	}
}

// The shape release-cli.yaml requires of an npm version, which is the only
// number the suite is allowed to carry.
var releaseVersion = regexp.MustCompile(`^\d+\.\d+\.\d+-cli\.\d+$`)

// What is committed is not a version, and saying so is the point: the release
// job stamps the release it is building over it (skills/stamp.py), so a
// number left in git would be a number nobody maintains and every reader has
// to decide whether to trust. A build from a checkout carries this, which is
// correct — it is not a release.
//
// This runs nowhere near the release job, which builds and does not test, so
// pinning the committed value here cannot fail a release that stamped
// properly. What it catches is somebody typing a version back in.
func TestTheCommittedSuiteCarriesThePlaceholder(t *testing.T) {
	metas, err := List()
	if err != nil {
		t.Fatalf("list skills: %v", err)
	}
	for _, meta := range metas {
		if meta.Version != placeholderVersion {
			t.Errorf("%s declares %s; the committed value is %s and the release stamps over it",
				meta.Name, meta.Version, placeholderVersion)
		}
	}
}

// Kept in step with PLACEHOLDER_VERSION in skills/stamp.py, which writes it,
// and with the check in publish.sh, which refuses it.
const placeholderVersion = "0.0.0-cli.0"

func TestReadDefaultsToTheEntryDocument(t *testing.T) {
	implicit, err := Read("olares-shared", "")
	if err != nil {
		t.Fatalf("read with empty name: %v", err)
	}
	explicit, err := Read("olares-shared", EntryFile)
	if err != nil {
		t.Fatalf("read %s: %v", EntryFile, err)
	}
	if string(implicit) != string(explicit) {
		t.Error("naming only a skill did not read its entry document")
	}
}

// A path an agent pastes is a path a user typed, and reading through it
// would hand out whatever the binary can see. The embedded FS refuses these
// on its own; the test states that we rely on that rather than on a caller
// sanitizing first.
func TestReadRefusesToEscapeTheSuite(t *testing.T) {
	for _, name := range []string{
		"../embed.go",
		"references/../../validate.py",
		"/etc/passwd",
	} {
		if _, err := Read("olares-shared", name); err == nil {
			t.Errorf("read %q succeeded; it should not resolve", name)
		}
	}
}

// suiteOnDisk walks the olares-* directories beside this file. Scoping to
// that prefix is the point: the same directory holds validate.py, stamp.py,
// publish.sh, requirements.txt, the suite README and this test, none of
// which belongs in the binary.
func suiteOnDisk(t *testing.T) []string {
	t.Helper()

	entries, err := os.ReadDir(".")
	if err != nil {
		t.Fatalf("read skills directory: %v", err)
	}
	var found []string
	for _, entry := range entries {
		if !entry.IsDir() || !strings.HasPrefix(entry.Name(), "olares-") {
			continue
		}
		err := filepath.WalkDir(entry.Name(), func(p string, e fs.DirEntry, err error) error {
			if err != nil {
				return err
			}
			if !e.IsDir() {
				found = append(found, filepath.ToSlash(p))
			}
			return nil
		})
		if err != nil {
			t.Fatalf("walk %s: %v", entry.Name(), err)
		}
	}
	if len(found) == 0 {
		t.Fatal("found no skills on disk; the test is not looking where the suite lives")
	}
	return found
}

func missing(want, have []string) []string {
	present := make(map[string]bool, len(have))
	for _, path := range have {
		present[path] = true
	}
	var absent []string
	for _, path := range want {
		if !present[path] {
			absent = append(absent, path)
		}
	}
	return absent
}
