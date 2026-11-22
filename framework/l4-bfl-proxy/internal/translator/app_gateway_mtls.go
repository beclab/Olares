package translator

import "github.com/beclab/l4-bfl-proxy/internal/ir"

// DefaultAppGatewayHost is the cluster DNS of Service app-gateway-l4-mtls
// (see .olares/config/cluster/deploy/l4-mtls-channel.yaml). Keep in sync with that manifest.
const DefaultAppGatewayHost = "app-gateway-l4-mtls.os-gateway.svc.cluster.local"

// DefaultAppGatewayPort is the mTLS channel Service port.
const DefaultAppGatewayPort = uint32(8443)

const (
	appGatewayClientCertSDS = "l4-app-gateway-client"
	appGatewayTrustedCASDS  = "l4-app-gateway-ca"
)

// clusterForAppGateway returns the upstream cluster when mTLS channel is enabled.
// Callers keep legacy app clusters when cfg.AppGatewayMTLS is false.
func (t *Translator) clusterForAppGateway(name string) *ir.ClusterIR {
	if t == nil || t.cfg == nil || !t.cfg.AppGatewayMTLS {
		return nil
	}
	host := t.cfg.AppGatewayHost
	if host == "" {
		host = DefaultAppGatewayHost
	}
	port := t.cfg.AppGatewayPort
	if port == 0 {
		port = DefaultAppGatewayPort
	}
	return &ir.ClusterIR{
		Name:   name,
		Host:   host,
		Port:   port,
		UseDNS: true,
		UpstreamMTLS: &ir.UpstreamMTLSIR{
			SNI:            host,
			ClientCertSDS:  appGatewayClientCertSDS,
			TrustedCACerts: appGatewayTrustedCASDS,
		},
	}
}
