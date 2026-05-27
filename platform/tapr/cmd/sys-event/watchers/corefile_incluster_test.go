package watchers

import (
	"context"
	"net"
	"sort"
	"strings"
	"testing"

	"github.com/coredns/corefile-migration/migration/corefile"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
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

func TestInClusterGatewayEnabled_defaultsTrue(t *testing.T) {
	if !inClusterGatewayEnabled(context.Background(), nil) {
		t.Fatal("nil client should default to enabled")
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

func TestSharedInclusterEntrancesFromSRRItems(t *testing.T) {
	prefix := sharedEntranceHostPrefix("a5be2268", "ollamav2")
	srr := unstructuredSRR("ollama-shared", "shared-a5be2268-ollamav2", map[string]string{
		labelSRRAppID:    "a5be2268",
		labelSRREntrance: "ollamav2",
	}, "gateway", []string{prefix + ".*.olares.com"})

	got := sharedInclusterEntrancesFromSRRItems(
		[]unstructured.Unstructured{*srr},
		nil,
		[]string{"alice", "bob"},
	)
	if len(got) != 2 {
		t.Fatalf("expected 2 entrances, got %d", len(got))
	}
	hosts := make([]string, 0, len(got))
	for _, e := range got {
		hosts = append(hosts, e.fqdn())
	}
	sort.Strings(hosts)
	want0 := prefix + ".alice.olares.com"
	want1 := prefix + ".bob.olares.com"
	if hosts[0] != want0 || hosts[1] != want1 {
		t.Fatalf("hosts=%v want %q and %q", hosts, want0, want1)
	}

	perUser := "a5be2268.bob.olares.com"
	plugin := buildSharedInclusterHosts(got, "10.0.0.5")
	if strings.Contains(plugin.ToString(), perUser) {
		t.Fatalf("per-user host must not appear: %s", plugin.ToString())
	}
}

func TestSharedInclusterEntrancesFromSRRItems_passthroughAndDirect(t *testing.T) {
	prefix := sharedEntranceHostPrefix("a5be2268", "api")
	srrGateway := unstructuredSRR("ollama-shared", "shared-a5be2268-api", map[string]string{
		labelSRRAppID: "a5be2268", labelSRREntrance: "api",
	}, "gateway", []string{prefix + ".*.olares.com"})
	srrDirect := unstructuredSRR("other-shared", "shared-bc2bd381-litellm", map[string]string{
		labelSRRAppID: "bc2bd381", labelSRREntrance: "litellm",
	}, "direct", []string{sharedEntranceHostPrefix("bc2bd381", "litellm") + ".*.olares.com"})
	passthrough := map[string]struct{}{"litellm-ns": {}}
	got := sharedInclusterEntrancesFromSRRItems(
		[]unstructured.Unstructured{*srrGateway, *srrDirect},
		passthrough,
		[]string{"alice"},
	)
	if len(got) != 1 {
		t.Fatalf("expected 1 entrance, got %d (%+v)", len(got), got)
	}
	if got[0].EntranceName != "api" {
		t.Fatalf("got %+v", got[0])
	}
}

func TestSharedInclusterEntrancesFromSRRItems_logicalPatternNotFirst(t *testing.T) {
	prefix := sharedEntranceHostPrefix("a5be2268", "api")
	srr := unstructuredSRR("ollama-shared", "shared-a5be2268-api", map[string]string{
		labelSRRAppID: "a5be2268", labelSRREntrance: "api",
	}, "gateway", []string{
		"api.shared.olares.com",
		prefix + ".*.olares.com",
	})
	got := sharedInclusterEntrancesFromSRRItems(
		[]unstructured.Unstructured{*srr},
		nil,
		[]string{"alice"},
	)
	if len(got) != 1 {
		t.Fatalf("expected 1 entrance, got %d (%+v)", len(got), got)
	}
	if got[0].fqdn() != prefix+".alice.olares.com" {
		t.Fatalf("unexpected fqdn %q", got[0].fqdn())
	}
}

func TestSharedInclusterEntrancesFromSRRItems_empty(t *testing.T) {
	if got := sharedInclusterEntrancesFromSRRItems(nil, nil, []string{"alice"}); got != nil {
		t.Fatalf("expected nil, got %v", got)
	}
}

func unstructuredSRR(ns, name string, labels map[string]string, routeMode string, hostPatterns []string) *unstructured.Unstructured {
	labelObj := make(map[string]interface{}, len(labels))
	for k, v := range labels {
		labelObj[k] = v
	}
	patterns := make([]interface{}, len(hostPatterns))
	for i, p := range hostPatterns {
		patterns[i] = p
	}
	obj := map[string]interface{}{
		"apiVersion": "gateway.olares.io/v1alpha1",
		"kind":       "SharedRouteRegistry",
		"metadata": map[string]interface{}{
			"name":      name,
			"namespace": ns,
			"labels":    labelObj,
		},
		"spec": map[string]interface{}{
			"routeMode":    routeMode,
			"hostPatterns": patterns,
		},
	}
	return &unstructured.Unstructured{Object: obj}
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
