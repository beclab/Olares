package translator

import (
	"testing"

	"github.com/beclab/l4-bfl-proxy/internal/ir"
	"github.com/beclab/l4-bfl-proxy/internal/message"
)

func TestBuildEdgeForwardVirtualHosts(t *testing.T) {
	tr := &Translator{cfg: &Config{AppGatewayMTLS: true}}
	user := &message.UserInfo{
		Name: "alice",
		Apps: []*message.AppInfo{{
			Name: "blog",
			Settings: map[string]string{
				settingsCustomDomain: `{"web":{"third_party_domain":"blog.example.com"}}`,
			},
			Entrances: []*message.EntranceInfo{{Name: "web", Host: "svc", Port: 80}},
		}},
	}
	cs := make(map[string]*ir.ClusterIR)
	vhosts := tr.buildEdgeForwardVirtualHosts(user, "alice.olares.com", cs)
	if len(vhosts) != 2 {
		t.Fatalf("vhosts=%d", len(vhosts))
	}
	if cs[edgeForwardClusterName] == nil || cs[edgeForwardClusterName].UpstreamMTLS == nil {
		t.Fatal("expected AGW mTLS cluster")
	}
	r := vhosts[0].Routes[0]
	if r.ExtAuth == nil {
		t.Fatal("edge business routes must authenticate through AGW")
	}
	if r.Cluster != edgeForwardClusterName {
		t.Fatalf("cluster=%s", r.Cluster)
	}
	foundApex, foundWild, foundLocal, foundCustom := false, false, false, false
	for _, d := range vhosts[0].Domains {
		switch d {
		case "alice.olares.com":
			foundApex = true
		case "*.alice.olares.com":
			foundWild = true
		case "alice.olares.local", "*.alice.olares.local":
			foundLocal = true
		case "blog.example.com":
			foundCustom = true
		}
	}
	if !foundApex || !foundWild || !foundLocal || !foundCustom {
		t.Fatalf("domains=%v", vhosts[0].Domains)
	}
}

func TestBuildUserVirtualHostsEdgeModeOnlyEdgeForward(t *testing.T) {
	tr := &Translator{cfg: &Config{AppGatewayMTLS: true, SSLServerPort: 443}}
	user := &message.UserInfo{
		Name:      "alice",
		Namespace: "user-space-alice",
		Apps: []*message.AppInfo{{
			Name:      "myapp",
			Namespace: "myapp-alice",
			Entrances: []*message.EntranceInfo{{Name: "web", Host: "myapp-svc", Port: 8080}},
		}},
	}
	cs := make(map[string]*ir.ClusterIR)
	vhosts := tr.buildUserVirtualHosts(user, "alice.olares.com", false, cs)
	if len(vhosts) != 2 || vhosts[0].Name != "edge_forward_alice" {
		t.Fatalf("want separate authenticated and login vhosts, got %#v", vhosts)
	}
	for name := range cs {
		if name != edgeForwardClusterName {
			t.Fatalf("unexpected cluster %s in edge mode", name)
		}
	}
	for i, vh := range vhosts {
		for _, r := range vh.Routes {
			if (r.ExtAuth != nil) != (i == 0) {
				t.Fatal("business vhost must authenticate and login vhost must avoid a loop")
			}
			if r.Cluster != edgeForwardClusterName {
				t.Fatalf("must not build local backend cluster %s", r.Cluster)
			}
			if i == 0 {
				if r.RequestHeaders["X-Original-URL"] != "%REQ(:SCHEME)%://%REQ(:AUTHORITY)%%REQ(:PATH)%" ||
					r.RequestHeaders["X-Original-Method"] != "%REQ(:METHOD)%" {
					t.Fatalf("authenticated traffic must preserve the original target: %v", r.RequestHeaders)
				}
				if _, found := r.RequestHeaders["X-BFL-USER"]; found {
					t.Fatal("edge must keep the identity returned by authorization")
				}
			} else if len(r.RequestHeaders) != 0 {
				t.Fatalf("login traffic must not receive application authorization metadata: %v", r.RequestHeaders)
			}
		}
	}
}

func TestEdgeForwardDenyAllUsesExactPublicHosts(t *testing.T) {
	tr := &Translator{cfg: &Config{AppGatewayMTLS: true}}
	user := &message.UserInfo{
		Name: "alice", DenyAll: true,
		AllowedDomains:    []string{"PUBLIC.alice.olares.com", "*.alice.olares.com", "other.example.org", "blog.example.com"},
		CustomDomainCerts: []*message.CertInfo{{Domain: "blog.example.com"}},
	}
	vhosts := tr.buildEdgeForwardVirtualHosts(user, "alice.olares.com", make(map[string]*ir.ClusterIR))
	tr.applyDenyAllRestrictions(user, vhosts, nil)
	got := make(map[string]bool)
	for _, vh := range vhosts {
		for _, host := range vh.Domains {
			got[host] = len(vh.SourceCIDRs) == 0
		}
	}
	for host, want := range map[string]bool{
		"public.alice.olares.com": true, "blog.example.com": true,
		"*.alice.olares.com": false, "alice.olares.com": false, "*.alice.olares.local": false,
	} {
		if allowed, exists := got[host]; !exists || allowed != want {
			t.Errorf("host %s: exists=%v public=%v, want public=%v", host, exists, allowed, want)
		}
	}
	if _, ok := got["other.example.org"]; ok {
		t.Fatal("foreign host was added to the route table")
	}
}

func TestProductionAuthorizationAndBusinessUseSeparateMTLSUpstreams(t *testing.T) {
	tr := &Translator{cfg: &Config{AppGatewayMTLS: true, SSLServerPort: 443}}
	user := &message.UserInfo{
		Name: "alice", ServerNameDomains: []string{"frontdoor.example.net"},
		SSL: &message.SSLConfig{Zone: "frontdoor.example.net", CertData: "test"},
	}
	snapshot := tr.Translate(&message.Resources{Users: []*message.UserInfo{user}})
	ports := map[string]uint32{"app_gateway_edge_forward": 8443, "authelia_backend_alice": 9444}
	for _, cluster := range snapshot.Clusters {
		port, expected := ports[cluster.Name]
		if !expected {
			t.Fatalf("unexpected upstream %s", cluster.Name)
		}
		if cluster.Port != port || cluster.Host != DefaultAppGatewayHost || cluster.UpstreamMTLS == nil ||
			cluster.UpstreamMTLS.SNI != DefaultAppGatewayHost || cluster.UpstreamMTLS.ClientCertSDS != appGatewayClientCertSDS {
			t.Fatalf("upstream %s could omit Entry mTLS: %+v", cluster.Name, cluster)
		}
		delete(ports, cluster.Name)
	}
	if len(ports) != 0 {
		t.Fatalf("missing upstreams %v", ports)
	}
	if len(snapshot.Secrets) != 1 {
		t.Fatal("missing Entry credentials must leave authorization waiting for SDS")
	}
}
