package translator

import (
	"fmt"
	"testing"

	"github.com/beclab/l4-bfl-proxy/internal/ir"
	"github.com/beclab/l4-bfl-proxy/internal/message"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestSharedAutheliaBackendRetainsUserContexts(t *testing.T) {
	tr := &Translator{cfg: &Config{SSLServerPort: 443, LocalHTTPProxyPort: 82}}
	resources := &message.Resources{LocalZone: "lab.olares.local"}
	for _, name := range []string{"alice", "bob"} {
		resources.Users = append(resources.Users, &message.UserInfo{
			Name: name, Namespace: "user-space-" + name,
			Apps: []*message.AppInfo{{
				Name: name + "-app", Appid: name + "-app", Namespace: name + "-app",
				Entrances: []*message.EntranceInfo{{Name: "web", Host: "web", Port: 8080}},
			}},
		})
	}
	xds := tr.Translate(resources)
	clusters := make(map[string]*ir.ClusterIR)
	for _, c := range xds.Clusters {
		clusters[c.Name] = c
	}
	for _, name := range []string{"alice", "bob"} {
		cluster := clusters["authelia_backend_"+name]
		require.NotNil(t, cluster)
		assert.Equal(t, "authelia-backend.os-framework.svc.cluster.local", cluster.Host)
		assert.Equal(t, uint32(9091), cluster.Port)
	}
	contexts := make(map[string]string)
	for _, listener := range xds.HTTPListeners {
		for _, vh := range listener.VirtualHosts {
			for _, route := range vh.Routes {
				if route.ExtAuth != nil && route.ExtAuth.ServiceOverride {
					contexts[route.ExtAuth.User] = route.ExtAuth.Cluster
					assert.Empty(t, route.RequestHeaders["X-BFL-USER"], "outbound identity comes from Authelia")
				}
			}
		}
	}
	for _, name := range []string{"alice", "bob"} {
		assert.Equal(t, fmt.Sprintf("authelia_backend_%s", name), contexts[name])
	}
}
