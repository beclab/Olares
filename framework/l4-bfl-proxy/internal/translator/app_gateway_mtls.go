package translator

import "github.com/beclab/l4-bfl-proxy/internal/ir"

const (
	defaultAppGatewayHost     = "app-gateway-l4-mtls.os-gateway.svc.cluster.local"
	defaultAppGatewayPort     = uint32(8443)
	appGatewayClientCertSDS   = "l4-app-gateway-client"
	appGatewayTrustedCASDS    = "l4-app-gateway-ca"
)

// clusterForAppGateway returns the upstream cluster when mTLS channel is enabled.
// Callers keep legacy app clusters when cfg.AppGatewayMTLS is false.
func (t *Translator) clusterForAppGateway(name string) *ir.ClusterIR {
	if t == nil || t.cfg == nil || !t.cfg.AppGatewayMTLS {
		return nil
	}
	host := t.cfg.AppGatewayHost
	if host == "" {
		host = defaultAppGatewayHost
	}
	port := t.cfg.AppGatewayPort
	if port == 0 {
		port = defaultAppGatewayPort
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
