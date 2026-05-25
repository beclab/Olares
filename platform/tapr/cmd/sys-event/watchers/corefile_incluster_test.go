package watchers

import (
	"net"
	"strings"
	"testing"

	"github.com/coredns/corefile-migration/migration/corefile"
)

func TestBuildSharedInclusterHosts_empty(t *testing.T) {
	if got := buildSharedInclusterHosts(nil, "10.0.0.5"); got != nil {
		t.Fatalf("expected nil, got %#v", got)
	}
	if got := buildSharedInclusterHosts([]SharedInclusterEntrance{}, "10.0.0.5"); got != nil {
		t.Fatalf("expected nil, got %#v", got)
	}
	if got := buildSharedInclusterHosts([]SharedInclusterEntrance{
		{AppID: "a5be2268", EntranceName: "ollama", Viewer: "alice", PlatformDomain: "olares.com"},
	}, "not-an-ip"); got != nil {
		t.Fatalf("expected nil for invalid gateway IP, got %#v", got)
	}
}

func TestBuildSharedInclusterHosts_orderingAndDedup(t *testing.T) {
	entrances := []SharedInclusterEntrance{
		{AppID: "a5be2268", EntranceName: "shared", Viewer: "bob", PlatformDomain: "olares.com"},
		{AppID: "a5be2268", EntranceName: "shared", Viewer: "alice", PlatformDomain: "olares.com"},
		{AppID: "a5be2268", EntranceName: "shared", Viewer: "bob", PlatformDomain: "olares.com"},
	}
	plugin := buildSharedInclusterHosts(entrances, "10.0.0.8")
	if plugin == nil {
		t.Fatal("expected hosts plugin")
	}
	hosts := hostsPluginNames(plugin)
	if len(hosts) != 2 {
		t.Fatalf("expected 2 hosts, got %v", hosts)
	}
	if hosts[0] != sharedEntranceHostPrefix("a5be2268", "shared")+".alice.olares.com" {
		t.Fatalf("unexpected first host: %s", hosts[0])
	}
	if hosts[1] != sharedEntranceHostPrefix("a5be2268", "shared")+".bob.olares.com" {
		t.Fatalf("unexpected second host: %s", hosts[1])
	}
}

func TestBuildSharedInclusterHosts_perUserEntranceExcluded(t *testing.T) {
	// Per-user single-entrance host uses appid.owner.zone — not in SRR-expanded list.
	perUserHost := "a5be2268.bob.olares.com"
	entrances := []SharedInclusterEntrance{
		{AppID: "a5be2268", EntranceName: "ollama", Viewer: "alice", PlatformDomain: "olares.com"},
	}
	plugin := buildSharedInclusterHosts(entrances, "172.16.0.4")
	if plugin == nil {
		t.Fatal("expected hosts plugin")
	}
	sharedHost := sharedEntranceHostPrefix("a5be2268", "ollama") + ".alice.olares.com"
	body := plugin.ToString()
	if !strings.Contains(body, sharedHost) {
		t.Fatalf("expected shared host %q in %q", sharedHost, body)
	}
	if strings.Contains(body, perUserHost) {
		t.Fatalf("per-user host %q must not appear in allowlist output: %q", perUserHost, body)
	}
}

func TestBuildSharedInclusterHosts_clusterIPRotation(t *testing.T) {
	ent := SharedInclusterEntrance{
		AppID: "a5be2268", EntranceName: "api", Viewer: "alice", PlatformDomain: "olares.com",
	}
	p1 := buildSharedInclusterHosts([]SharedInclusterEntrance{ent}, "10.0.0.1")
	p2 := buildSharedInclusterHosts([]SharedInclusterEntrance{ent}, "10.0.0.2")
	if p1 == nil || p2 == nil {
		t.Fatal("expected plugins")
	}
	if gatewayIPFromHostsPlugin(p1) != "10.0.0.1" {
		t.Fatalf("ip1=%s", gatewayIPFromHostsPlugin(p1))
	}
	if gatewayIPFromHostsPlugin(p2) != "10.0.0.2" {
		t.Fatalf("ip2=%s", gatewayIPFromHostsPlugin(p2))
	}
}

func TestBuildSharedInclusterHosts_roundTripCorefile(t *testing.T) {
	plugin := buildSharedInclusterHosts([]SharedInclusterEntrance{
		{AppID: "bc2bd381", EntranceName: "litellm", Viewer: "alice", PlatformDomain: "olares.com"},
	}, "192.168.1.10")
	if plugin == nil {
		t.Fatal("expected plugin")
	}
	server := &corefile.Server{
		DomPorts: []string{".:53"},
		Plugins:  []*corefile.Plugin{plugin},
	}
	parsed, err := corefile.New(server.ToString())
	if err != nil {
		t.Fatalf("parse generated corefile: %v", err)
	}
	if len(parsed.Servers) != 1 || len(parsed.Servers[0].Plugins) != 1 {
		t.Fatalf("unexpected parsed structure: %+v", parsed)
	}
	got := parsed.Servers[0].Plugins[0]
	if got.Name != "hosts" {
		t.Fatalf("plugin name=%s", got.Name)
	}
	host := sharedEntranceHostPrefix("bc2bd381", "litellm") + ".alice.olares.com"
	if !strings.Contains(got.ToString(), host) {
		t.Fatalf("missing host %q in %q", host, got.ToString())
	}
}

func hostsPluginNames(plugin *corefile.Plugin) []string {
	if plugin == nil || plugin.Name != "hosts" {
		return nil
	}
	var out []string
	for _, opt := range plugin.Options {
		if net.ParseIP(opt.Name) != nil {
			out = append(out, opt.Args...)
		}
	}
	return out
}

func gatewayIPFromHostsPlugin(plugin *corefile.Plugin) string {
	if plugin == nil {
		return ""
	}
	for _, opt := range plugin.Options {
		if ip := net.ParseIP(opt.Name); ip != nil {
			return ip.String()
		}
	}
	return ""
}
