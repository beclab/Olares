package translator

import "testing"

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
	if c.Port != defaultAppGatewayPort {
		t.Fatalf("port=%d", c.Port)
	}
	if c.UpstreamMTLS.ClientCertSDS != appGatewayClientCertSDS {
		t.Fatalf("client SDS=%s", c.UpstreamMTLS.ClientCertSDS)
	}
}
