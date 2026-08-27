package ctl

import (
	"os"
	"regexp"
	"strings"
	"testing"

	"github.com/spf13/cobra"
)

// skillCommandPaths is every command path the skill docs spell out, so
// renaming or removing one fails here rather than in a user's session.
func skillCommandPaths() []string {
	return []string{
		"profile",
		"profile login",
		"profile import",
		"profile list",
		"profile use",
		"profile remove",
		"profile whoami",
		"files",
		"files ls",
		"files cat",
		"files download",
		"files upload",
		"files edit",
		"files mkdir",
		"files rm",
		"files rename",
		"files cp",
		"files mv",
		"files chown",
		"files compress",
		"files extract",
		"files archive",
		"files nfs",
		"files task",
		"files share",
		"files smb",
		"files repos",
		"knowledge download",
		"knowledge download create",
		"knowledge download list",
		"knowledge download info",
		"knowledge download pause",
		"knowledge download resume",
		"knowledge download cancel",
		"knowledge download remove",
		"knowledge download inspect",
		"knowledge download prefs",
		"knowledge download unfinished",
		"knowledge download sync",
		"knowledge download torrent",
		"knowledge download file",
		"knowledge download settings",
		"search",
		"search drive",
		"search sync",
		"search knowledge",
		"search app",
		"market",
		"market list",
		"market get",
		"market categories",
		"market install",
		"market upgrade",
		"market uninstall",
		"market clone",
		"market stop",
		"market resume",
		"market cancel",
		"market restart",
		"market status",
		"market upload",
		"market download",
		"market delete",
		"settings",
		"settings me",
		"settings users",
		"settings apps",
		"settings vpn",
		"settings vpn routes enable",
		"settings vpn subroutes enable",
		"settings backup",
		"settings integration",
		"settings appearance",
		"settings appearance get",
		"settings appearance language set",
		"settings appearance theme set",
		"settings appearance widget set",
		"settings appearance wallpaper list",
		"settings appearance wallpaper set",
		"settings appearance wallpaper style set",
		"settings appearance wallpaper upload",
		"settings appearance wallpaper delete",
		"settings appearance layout reset",
		"settings network",
		"settings gpu",
		"settings compute",
		"settings video",
		"settings search",
		"settings restore",
		"settings advanced",
		"dashboard",
		"dashboard applications",
		"dashboard overview",
		"dashboard schema",
		"cluster",
		"cluster context",
		"cluster pod",
		"cluster container",
		"cluster workload",
		"cluster application",
		"cluster namespace",
		"cluster node",
		"cluster job",
		"cluster cronjob",
		"cluster middleware",
		"doctor",
		"doctor images",
		"doctor thirdleveldomain",
		"chart",
		"chart from-compose",
		"chart lint",
		"chart package",
		"router",
		"router provider",
		"router provider list",
		"router provider get",
		"router provider types",
		"router provider create",
		"router provider register",
		"router provider update",
		"router provider delete",
		"router provider validate",
		"router provider sync-models",
		"router provider credentials",
		"router provider history",
		"router provider rollback",
		"router model",
		"router model list",
		"router model get",
		"router model import",
		"router model add",
		"router model update",
		"router model remove",
		"router model status",
		"router model progress",
		"router model retry",
		"router model restart",
		"router model diag",
		"router model diag gpu",
		"router model diag config",
		"router model diag endpoints",
		"router route",
		"router route list",
		"router route get",
		"router route create",
		"router route rename",
		"router route enable",
		"router route disable",
		"router route delete",
		"router route add",
		"router route remove",
		"router model spec",
		"router model spec show",
		"router model spec edit",
		"router model spec file",
		"router model spec set",
		"router call",
		"router call models",
		"router call chat",
		"router call responses",
		"router call embed",
		"router call rerank",
		"router call search",
		"router call scrape",
		"router call translate",
		"router call image",
		"router call video",
		"router call transcribe",
		"router call speak",
		"router call clone",
		"router call dialogue",
		"router call listen",
		"router call vad",
		"router call diarize",
		"router call speaker-embed",
		"router call enhance",
		"router call align",
		"router call task",
		"router call task get",
		"router call task result",
		"router call task cancel",
		"router call task list",
		"router call ocr",
		"router key",
		"router key list",
		"router key issue",
		"router key update",
		"router key revoke",
		"router key current",
		"router quota",
		"router quota list",
		"router quota set",
		"router quota clear",
		"router usage",
		"router usage summary",
		"router usage list",
		"router usage export",
		"router usage retention",
		"router audit",
		"router audit list",
		"router audit get",
	}
}

func TestSkillCommandPathsExist(t *testing.T) {
	root := NewDefaultCommand()
	for _, path := range skillCommandPaths() {
		t.Run(path, func(t *testing.T) {
			cmd, args, err := root.Find(strings.Fields(path))
			if err != nil {
				t.Fatalf("find command: %v", err)
			}
			if len(args) != 0 {
				t.Fatalf("unresolved command tokens: %v", args)
			}
			if got, want := cmd.CommandPath(), "olares-cli "+path; got != want {
				t.Fatalf("resolved %q to %q", path, got)
			}
		})
	}
}

// TestSkillCommandPathsExist only proves the listed paths resolve, so a
// newly added verb can ship unlisted. This is the reverse check for the
// area under active change: every runnable verb under `settings
// appearance` must be listed above, and therefore documented.
func TestEveryAppearanceVerbIsListed(t *testing.T) {
	listed := map[string]bool{}
	for _, path := range skillCommandPaths() {
		listed[path] = true
	}

	root := NewDefaultCommand()
	appearance, _, err := root.Find(strings.Fields("settings appearance"))
	if err != nil {
		t.Fatalf("find settings appearance: %v", err)
	}

	var found int
	var walk func(cmd *cobra.Command)
	walk = func(cmd *cobra.Command) {
		if cmd.Runnable() {
			found++
			path := strings.TrimPrefix(cmd.CommandPath(), "olares-cli ")
			if !listed[path] {
				t.Errorf("verb %q is not in skillCommandPaths; add it there and document it in the skill reference", path)
			}
		}
		for _, child := range cmd.Commands() {
			walk(child)
		}
	}
	walk(appearance)

	// Guard against the walk silently finding nothing, which would make
	// the check above vacuous.
	if found < 10 {
		t.Fatalf("walked only %d runnable appearance verbs; the subtree has at least 10", found)
	}
}

func TestSkillRecoveryCommandPathsExist(t *testing.T) {
	source, err := os.ReadFile("download/common.go")
	if err != nil {
		t.Fatal(err)
	}
	matches := regexp.MustCompile("olares-cli knowledge download ([a-z-]+)").FindAllStringSubmatch(string(source), -1)
	if len(matches) == 0 {
		t.Fatal("no knowledge download recovery commands found")
	}

	root := NewDefaultCommand()
	for _, match := range matches {
		path := "knowledge download " + match[1]
		cmd, args, err := root.Find(strings.Fields(path))
		if err != nil {
			t.Fatalf("find recovery command %q: %v", path, err)
		}
		if len(args) != 0 || cmd.CommandPath() != "olares-cli "+path {
			t.Fatalf("recovery command %q does not resolve", path)
		}
	}
}
