package translator

import (
	"os"
	"path/filepath"
	"runtime"
	"strings"
	"testing"
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
