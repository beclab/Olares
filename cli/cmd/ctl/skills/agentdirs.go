package skills

import (
	"os"
	"path/filepath"
	"strings"
)

// storeDir is where the skills CLI keeps the one copy of each installed
// skill; every agent directory holds links into it.
const storeDir = ".agents/skills"

// skillPrefix is what this binary's skills are named, and the only thing
// scanning here is allowed to reason about. Somebody else's skills in the
// same directory are none of our business.
const skillPrefix = "olares-"

// agentSkillDirs finds the directories on this machine that hold skills.
//
// Each agent keeps its own — ~/.claude/skills, ~/.cursor/skills,
// ~/.codex/skills — and a few nest one level deeper
// (~/.codeium/windsurf/skills, ~/.config/crush/skills, ~/.pi/agent/skills).
// The skills CLI carries that list, which is over eighty entries long and
// grows; copying it here would mean a directory it learns about is one this
// scan silently stops covering. Looking for a directory named "skills" one
// or two levels inside a dotted directory in $HOME describes every shape on
// that list without restating any of it.
//
// The store is reported alongside the agent directories; callers that care
// about the difference compare against StorePath.
func agentSkillDirs() []string {
	home, err := os.UserHomeDir()
	if err != nil {
		return nil
	}
	var found []string
	consider := func(path string) {
		if info, err := os.Stat(path); err == nil && info.IsDir() {
			found = append(found, path)
		}
	}

	entries, err := os.ReadDir(home)
	if err != nil {
		return nil
	}
	for _, entry := range entries {
		if !strings.HasPrefix(entry.Name(), ".") || entry.Name() == "." || entry.Name() == ".." {
			continue
		}
		outer := filepath.Join(home, entry.Name())
		if info, err := os.Stat(outer); err != nil || !info.IsDir() {
			continue
		}
		consider(filepath.Join(outer, "skills"))

		inner, err := os.ReadDir(outer)
		if err != nil {
			continue
		}
		for _, child := range inner {
			if child.Name() == "skills" {
				continue
			}
			nested := filepath.Join(outer, child.Name())
			if info, err := os.Stat(nested); err != nil || !info.IsDir() {
				continue
			}
			consider(filepath.Join(nested, "skills"))
		}
	}
	return found
}

// StorePath is the canonical store, or "" if the home directory is unknown.
func StorePath() string {
	home, err := os.UserHomeDir()
	if err != nil {
		return ""
	}
	return filepath.Join(home, filepath.FromSlash(storeDir))
}

// handMadeLink is an olares-* entry that somebody linked by hand at
// something outside the store — which in practice means the local
// development layout in cli/README.md, pointing at a git checkout.
type handMadeLink struct {
	Path   string `json:"path"`
	Target string `json:"target"`
}

// handMadeLinksIn finds those links in one directory.
//
// An install replaces them with the standard store-and-link shape, so the
// checkout stops being what the agent reads and edits to it stop showing up.
// Nothing about that is visible while it happens, which is why it is worth
// looking for rather than mentioning in documentation.
//
// The caller is `skills install` looking at the store, which is the one copy
// every agent directory links to and so decides the whole install. An agent
// directory is examined by classify as that directory is installed to, which
// is what keeps one agent's development link from blocking the others.
func handMadeLinksIn(dir string) []handMadeLink {
	store := StorePath()
	entries, err := os.ReadDir(dir)
	if err != nil {
		return nil
	}
	var found []handMadeLink
	for _, entry := range entries {
		if !strings.HasPrefix(entry.Name(), skillPrefix) {
			continue
		}
		path := filepath.Join(dir, entry.Name())
		info, err := os.Lstat(path)
		if err != nil || info.Mode()&os.ModeSymlink == 0 {
			continue
		}
		target, err := os.Readlink(path)
		if err != nil {
			continue
		}
		if pointsInto(path, target, store) {
			continue
		}
		found = append(found, handMadeLink{Path: path, Target: target})
	}
	return found
}

// pointsInto reports whether a link resolves inside the store. The links the
// skills CLI writes are relative to the directory holding them, and on a
// system where the home directory is itself reached through a link the
// lexical answer and the resolved one differ, so both are tried.
func pointsInto(linkPath, target, store string) bool {
	if store == "" {
		return false
	}
	resolved := target
	if !filepath.IsAbs(resolved) {
		resolved = filepath.Join(filepath.Dir(linkPath), resolved)
	}
	if under(filepath.Clean(resolved), store) {
		return true
	}
	real, err := filepath.EvalSymlinks(resolved)
	if err != nil {
		return false
	}
	realStore, err := filepath.EvalSymlinks(store)
	if err != nil {
		return false
	}
	return under(real, realStore)
}

func under(path, root string) bool {
	relative, err := filepath.Rel(root, path)
	if err != nil {
		return false
	}
	return relative != ".." && !strings.HasPrefix(relative, ".."+string(filepath.Separator))
}
