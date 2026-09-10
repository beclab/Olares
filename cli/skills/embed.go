// Package skills carries the agent skill suite that ships inside the
// olares-cli binary, so the instructions an agent follows and the command
// tree it drives are one release rather than two.
//
// The suite used to reach a machine only over the network: `npx skills add
// beclab/Olares`, which resolves to GitHub's HEAD, or a maintainer running
// publish.sh against ClawHub. Both leave the version an agent reads free to
// disagree with the verbs its binary actually has, and the disagreement is
// silent — a skill declares `requires.bins: [olares-cli]`, which an old
// build satisfies. Embedding removes the degree of freedom: `skills export`
// writes these bytes and nothing else does.
package skills

import (
	"bytes"
	"fmt"
	"io/fs"
	"path"
	"sort"
	"strings"

	"embed"

	"gopkg.in/yaml.v3"
)

// Three patterns are the whole of a skill: SKILL.md, the references it
// links to, and the scripts those references tell a reader to run
// (olares-publish/references/olares-publish-icon.md invokes
// ../scripts/generate_icon.py, so leaving scripts/ out would ship an
// instruction that cannot be followed). Everything else in this directory
// — validate.py, stamp.py, publish.sh, requirements.txt, the suite README —
// is maintainer tooling and stays out of the binary.
//
// TestEmbedCoversTheWholeSuite asserts this set equals the olares-* subtree
// on disk, so a skill that grows a new kind of subdirectory fails a test
// rather than shipping half of itself.
//
//go:embed olares-*/SKILL.md olares-*/references olares-*/scripts
var content embed.FS

// EntryFile is the document an agent is pointed at first; everything else
// under a skill is reached from it.
const EntryFile = "SKILL.md"

// FS returns the embedded suite. Paths read "<skill>/SKILL.md" — the same
// shape they have in this directory and in ~/.agents/skills, so a path
// copied out of a reference link resolves without translation.
func FS() fs.FS { return content }

// Meta is the part of a skill's frontmatter that identifies it. The rest of
// the frontmatter (compatibility, metadata.openclaw.requires) is contract
// between the skill and whatever installed it, and is left as written.
type Meta struct {
	Name        string `yaml:"name"`
	Version     string `yaml:"version"`
	Description string `yaml:"description"`
}

// Names lists the embedded skills in sorted order.
func Names() ([]string, error) {
	entries, err := fs.ReadDir(content, ".")
	if err != nil {
		return nil, fmt.Errorf("read embedded suite: %w", err)
	}
	var names []string
	for _, entry := range entries {
		if entry.IsDir() {
			names = append(names, entry.Name())
		}
	}
	sort.Strings(names)
	return names, nil
}

// List returns every skill's Meta, in the order Names gives.
func List() ([]Meta, error) {
	names, err := Names()
	if err != nil {
		return nil, err
	}
	metas := make([]Meta, 0, len(names))
	for _, name := range names {
		meta, err := Describe(name)
		if err != nil {
			return nil, err
		}
		metas = append(metas, meta)
	}
	return metas, nil
}

// Describe parses one embedded skill's frontmatter.
func Describe(skill string) (Meta, error) {
	source, err := Read(skill, EntryFile)
	if err != nil {
		return Meta{}, err
	}
	meta, err := ParseMeta(source)
	if err != nil {
		return Meta{}, fmt.Errorf("%s/%s: %w", skill, EntryFile, err)
	}
	return meta, nil
}

// ParseMeta reads the frontmatter of a SKILL.md from anywhere, which is how a
// copy already on disk gets compared against the copy in here.
func ParseMeta(source []byte) (Meta, error) {
	block, err := frontmatter(source)
	if err != nil {
		return Meta{}, err
	}
	var meta Meta
	if err := yaml.Unmarshal(block, &meta); err != nil {
		return Meta{}, fmt.Errorf("parse frontmatter: %w", err)
	}
	return meta, nil
}

// Read returns one file's bytes. An empty name reads the skill's entry
// document, which is what a caller naming only a skill means.
//
// Path validation comes from the embedded FS: fs.ReadFile rejects anything
// fs.ValidPath does, so "..", an absolute path and a trailing slash are all
// errors here rather than reads of something else.
func Read(skill, name string) ([]byte, error) {
	if skill == "" {
		return nil, fmt.Errorf("no skill named")
	}
	if name == "" {
		name = EntryFile
	}
	target := path.Join(skill, name)
	source, err := fs.ReadFile(content, target)
	if err != nil {
		return nil, fmt.Errorf("read %s: %w", target, err)
	}
	return source, nil
}

// Files lists every embedded path, sorted. Used by the export writer and by
// the test that pins this set to the directory on disk.
func Files() ([]string, error) {
	var found []string
	err := fs.WalkDir(content, ".", func(p string, entry fs.DirEntry, err error) error {
		if err != nil {
			return err
		}
		if !entry.IsDir() {
			found = append(found, p)
		}
		return nil
	})
	if err != nil {
		return nil, fmt.Errorf("walk embedded suite: %w", err)
	}
	sort.Strings(found)
	return found, nil
}

// frontmatter returns the YAML block a SKILL.md opens with.
func frontmatter(source []byte) ([]byte, error) {
	const fence = "---"
	text := string(source)
	if !strings.HasPrefix(text, fence+"\n") {
		return nil, fmt.Errorf("does not open with a %q frontmatter fence", fence)
	}
	rest := text[len(fence)+1:]
	end := strings.Index(rest, "\n"+fence)
	if end < 0 {
		return nil, fmt.Errorf("frontmatter fence is never closed")
	}
	return bytes.TrimSpace([]byte(rest[:end])), nil
}
