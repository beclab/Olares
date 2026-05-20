package terminus

import agwconfig "github.com/beclab/Olares/framework/app-gateway/pkg/config"

// buildAppGatewayHelmValues maps framework/app-gateway/config/defaults.yaml into Helm values.
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
	envoyProxyEnabled := def.EnvoyProxy.Enabled
	accessLogEnabled := def.EnvoyProxy.AccessLog.Enabled

	meshEnabled := def.Mesh.Linkerd.Enabled
	opaquePorts := def.Mesh.Linkerd.OpaquePorts
	if opaquePorts == "" {
		opaquePorts = "10080,19001"
	}

	return map[string]interface{}{
		"namespace":       ns,
		"namespaceCreate": false,
		"gateway": map[string]interface{}{
			"name":             gwName,
			"gatewayClassName": gwClass,
		},
		"envoyProxy": map[string]interface{}{
			"enabled": envoyProxyEnabled,
			"name":    envoyProxyName,
			"accessLog": map[string]interface{}{
				"enabled": accessLogEnabled,
			},
		},
		"mesh": map[string]interface{}{
			"linkerd": map[string]interface{}{
				"enabled":     meshEnabled,
				"opaquePorts": opaquePorts,
			},
		},
		"demo": map[string]interface{}{
			"enabled": def.Demo.Enabled,
			"host":    def.Demo.Host,
		},
	}
}
