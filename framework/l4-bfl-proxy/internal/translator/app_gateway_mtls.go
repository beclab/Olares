package translator

import (
	"github.com/beclab/l4-bfl-proxy/internal/ir"
	"github.com/beclab/l4-bfl-proxy/internal/message"
)

// DefaultAppGatewayHost is the cluster DNS of Service app-gateway-l4-mtls
// (see .olares/config/cluster/deploy/l4-mtls-channel.yaml). Keep in sync with that manifest.
const DefaultAppGatewayHost = "app-gateway-l4-mtls.os-gateway.svc.cluster.local"

// DefaultAppGatewayPort is the mTLS channel Service port.
const DefaultAppGatewayPort = uint32(8443)

// DefaultAppGatewayAuthPort is the dedicated authorization channel Service port.
const DefaultAppGatewayAuthPort = uint32(9444)

const (
	appGatewayClientCertSDS = "l4-app-gateway-client"
	appGatewayTrustedCASDS  = "l4-app-gateway-ca"
)

// clusterForAppGateway returns the upstream cluster when mTLS channel is enabled.
// Fail-closed: without Entry material, still returns the mTLS cluster so traffic
// does not fall back to cleartext; Envoy waits for SDS.
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

// appendEntryMTLSSDS adds client certificate and CA ValidationContext SDS resources.
func appendEntryMTLSSDS(xds *ir.Xds, mat *message.EntryMTLSMaterial) {
	if xds == nil || !mat.Ready() {
		return
	}
	xds.Secrets = append(xds.Secrets,
		&ir.SecretIR{
			Name:     appGatewayClientCertSDS,
			CertData: mat.ClientCert,
			KeyData:  mat.ClientKey,
		},
		&ir.SecretIR{
			Name:      appGatewayTrustedCASDS,
			TrustedCA: mat.CA,
		},
	)
}
