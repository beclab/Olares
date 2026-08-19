package ctl

import (
	"os"
	"regexp"
	"strings"
	"testing"
)

func TestSkillCommandPathsExist(t *testing.T) {
	paths := []string{
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
		"search gdrive",
		"search dropbox",
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
		"router status",
		"router whoami",
		"router list",
		"router capabilities",
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
		"router provider models",
		"router provider models import",
		"router provider models add",
		"router provider models update",
		"router provider models delete",
		"router default",
		"router default show",
		"router default enable",
		"router default disable",
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
		"router app",
		"router app catalog",
		"router app installed",
		"router app install",
		"router app upgrade",
		"router app uninstall",
		"router call",
		"router call chat",
		"router call embed",
		"router call transcribe",
		"router call speak",
		"router call ocr",
		"router key",
		"router key list",
		"router key issue",
		"router key update",
		"router key revoke",
		"router key local",
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
		"router user",
		"router user list",
		"router local",
		"router local status",
		"router local progress",
		"router local spec",
		"router local spec show",
		"router local spec file",
		"router local spec set",
		"router local config",
		"router local endpoints",
		"router local gpu",
		"router local perf",
		"router local retry",
		"router local restart",
	}

	root := NewDefaultCommand()
	for _, path := range paths {
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
