package translator

import (
	"os"
	"path/filepath"
	"runtime"
	"strings"
	"testing"

	"github.com/beclab/l4-bfl-proxy/internal/ir"
	"github.com/beclab/l4-bfl-proxy/internal/message"
)

func TestClusterForAppGatewayDisabledByDefault(t *testing.T) {
	tr := &Translator{cfg: &Config{}}
	if tr.clusterForAppGateway("agw") != nil {
		t.Fatal("expected nil when AppGatewayMTLS is false")
	}
}

func TestClusterForAppGatewayEnabled(t *testing.T) {
	tr := &Translator{cfg: &Config{AppGatewayMTLS: true}}
	c := tr.clusterForAppGateway("agw_mtls")
	if c == nil || c.UpstreamMTLS == nil {
		t.Fatal("expected upstream mTLS cluster")
	}
	if c.Host != DefaultAppGatewayHost {
		t.Fatalf("host=%s want %s", c.Host, DefaultAppGatewayHost)
	}
	if c.UpstreamMTLS.SNI != DefaultAppGatewayHost {
		t.Fatalf("sni=%s", c.UpstreamMTLS.SNI)
	}
	if c.Port != DefaultAppGatewayPort {
		t.Fatalf("port=%d", c.Port)
	}
	if c.UpstreamMTLS.ClientCertSDS != appGatewayClientCertSDS {
		t.Fatalf("client SDS=%s", c.UpstreamMTLS.ClientCertSDS)
	}
}

func TestAppendEntryMTLSSDS(t *testing.T) {
	xds := &ir.Xds{}
	appendEntryMTLSSDS(xds, nil)
	if len(xds.Secrets) != 0 {
		t.Fatal("nil material")
	}
	appendEntryMTLSSDS(xds, &message.EntryMTLSMaterial{
		ClientCert: "cc", ClientKey: "ck", CA: "ca",
	})
	if len(xds.Secrets) != 2 {
		t.Fatalf("secrets=%d", len(xds.Secrets))
	}
	if xds.Secrets[0].Name != appGatewayClientCertSDS || xds.Secrets[0].CertData != "cc" {
		t.Fatal("client sds")
	}
	if xds.Secrets[1].Name != appGatewayTrustedCASDS || xds.Secrets[1].TrustedCA != "ca" {
		t.Fatal("ca sds")
	}
}

func TestBuildAppVirtualHostsUsesAppGatewayMTLS(t *testing.T) {
	tr := &Translator{cfg: &Config{AppGatewayMTLS: true, SSLServerPort: 443}}
	user := &message.UserInfo{Name: "alice"}
	app := &message.AppInfo{
		Name:      "myapp",
		Appid:     "myapp",
		Namespace: "myapp-alice",
		Entrances: []*message.EntranceInfo{
			{Name: "web", Host: "myapp-svc", Port: 8080},
		},
	}
	clusterSet := make(map[string]*ir.ClusterIR)
	_ = tr.buildAppVirtualHosts(user, app, "alice.example.com", false, clusterSet)
	c := clusterSet["app_alice_myapp_web"]
	if c == nil || c.UpstreamMTLS == nil {
		t.Fatal("expected app-gateway mTLS cluster")
	}
	if c.Host != DefaultAppGatewayHost {
		t.Fatalf("host=%s", c.Host)
	}
}

func TestBuildAppVirtualHostsDevSkipsAppGatewayMTLS(t *testing.T) {
	tr := &Translator{cfg: &Config{AppGatewayMTLS: true, SSLServerPort: 443}}
	user := &message.UserInfo{Name: "alice"}
	app := &message.AppInfo{
		Name:      "myapp",
		Appid:     "myapp",
		Namespace: "myapp-alice",
		Entrances: []*message.EntranceInfo{
			{Name: "web", Host: "10.0.0.9", Port: 8080, Type: entranceTypeDev},
		},
	}
	clusterSet := make(map[string]*ir.ClusterIR)
	_ = tr.buildAppVirtualHosts(user, app, "alice.example.com", false, clusterSet)
	c := clusterSet["app_alice_myapp_web"]
	if c == nil || c.UpstreamMTLS != nil {
		t.Fatal("dev entrance must stay on pod IP without AGW mTLS")
	}
	if c.Host != "10.0.0.9" || c.UseDNS {
		t.Fatalf("dev cluster=%+v", c)
	}
}

func TestDefaultAppGatewayHostMatchesChannelManifest(t *testing.T) {
	_, thisFile, _, ok := runtime.Caller(0)
	if !ok {
		t.Fatal("runtime.Caller failed")
	}
	// l4-bfl-proxy/internal/translator → repo framework/app-gateway deploy
	manifest := filepath.Clean(filepath.Join(filepath.Dir(thisFile),
		"..", "..", "..", "app-gateway", ".olares", "config", "cluster", "deploy", "l4-mtls-channel.yaml"))
	raw, err := os.ReadFile(manifest)
	if err != nil {
		t.Fatalf("read channel manifest %s: %v", manifest, err)
	}
	body := string(raw)
	if !strings.Contains(body, "name: app-gateway-l4-mtls") {
		t.Fatalf("manifest missing Service name app-gateway-l4-mtls: %s", manifest)
	}
	if !strings.Contains(body, `enabled: "true"`) {
		t.Fatal("channel CM must project enabled=true")
	}
	if !strings.Contains(body, "RequireAndVerify") {
		t.Fatal("ClientTrafficPolicy must RequireAndVerify")
	}
	if !strings.Contains(body, "owning-gateway-name: app-gateway") {
		t.Fatal("Service must select EG dataplane")
	}
	if !strings.Contains(body, "namespace: os-gateway") {
		t.Fatalf("manifest missing namespace os-gateway: %s", manifest)
	}
	if !strings.Contains(body, "sni: \""+DefaultAppGatewayHost+"\"") {
		t.Fatalf("manifest sni must equal DefaultAppGatewayHost %q", DefaultAppGatewayHost)
	}
	if DefaultAppGatewayHost != "app-gateway-l4-mtls.os-gateway.svc.cluster.local" {
		t.Fatalf("unexpected DefaultAppGatewayHost %q", DefaultAppGatewayHost)
	}
}
