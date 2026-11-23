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
	foundWild, foundCustom := false, false
	for _, d := range vhosts[0].Domains {
		if d == "*.alice.olares.com" {
			foundWild = true
		}
		if d == "blog.example.com" {
			foundCustom = true
		}
	}
	if !foundWild || !foundCustom {
		t.Fatalf("domains=%v", vhosts[0].Domains)
	}
}

func TestBuildUserVirtualHostsEdgeModeSkipsAppBackends(t *testing.T) {
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
	for _, vh := range vhosts {
		for _, r := range vh.Routes {
			if r.Cluster == "app_alice_myapp_web" {
				t.Fatal("must not build direct app cluster in edge mode")
			}
		}
	}
	if cs[edgeForwardClusterName] == nil {
		t.Fatal("missing edge forward cluster")
	}
	_ = vhosts
}
