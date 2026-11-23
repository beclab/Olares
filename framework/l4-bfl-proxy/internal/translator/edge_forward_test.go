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
	if len(vhosts) != 1 {
		t.Fatalf("vhosts=%d", len(vhosts))
	}
	if cs[edgeForwardClusterName] == nil || cs[edgeForwardClusterName].UpstreamMTLS == nil {
		t.Fatal("expected AGW mTLS cluster")
	}
	r := vhosts[0].Routes[0]
	if r.ExtAuth != nil {
		t.Fatal("edge forward must not attach ExtAuth")
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
	if len(vhosts) != 1 || vhosts[0].Name != "edge_forward_alice" {
		t.Fatalf("want single edge_forward vhost, got %#v", vhosts)
	}
	for name := range cs {
		if name != edgeForwardClusterName {
			t.Fatalf("unexpected cluster %s in edge mode", name)
		}
	}
	for _, vh := range vhosts {
		for _, r := range vh.Routes {
			if r.ExtAuth != nil {
				t.Fatal("edge mode must not attach ExtAuth")
			}
			if r.Cluster != edgeForwardClusterName {
				t.Fatalf("must not build local backend cluster %s", r.Cluster)
			}
		}
	}
}
