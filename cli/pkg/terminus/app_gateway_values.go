package terminus

import agwconfig "github.com/beclab/Olares/framework/app-gateway/pkg/config"

// buildAppGatewayHelmValues maps framework/app-gateway/config/defaults.yaml into Helm values.
// No service mesh: only namespace, Gateway identity, TLS and EnvoyProxy overlay are surfaced.
func buildAppGatewayHelmValues(ns string, def agwconfig.Defaults) map[string]interface{} {
	gwName := def.Gateway.Name
	gwClass := def.Gateway.GatewayClassName
	if gwName == "" {
		gwName = "app-gateway"
	}
	if gwClass == "" {
		gwClass = "olares-app-gateway"
	}

	envoyProxyName := def.EnvoyProxy.Name
	if envoyProxyName == "" {
		envoyProxyName = "app-gateway-envoy-proxy"
	}

	return map[string]interface{}{
		"namespace":       ns,
		"namespaceCreate": false,
		"gateway": map[string]interface{}{
			"name":             gwName,
			"gatewayClassName": gwClass,
		},
		"tls": map[string]interface{}{
			"enabled": def.TLS.Enabled,
		},
		"envoyProxy": map[string]interface{}{
			"enabled": def.EnvoyProxy.Enabled,
			"name":    envoyProxyName,
			"accessLog": map[string]interface{}{
				"enabled": def.EnvoyProxy.AccessLog.Enabled,
			},
		},
	}
}
