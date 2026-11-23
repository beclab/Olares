package translator

import (
	"slices"
	"strings"

	"github.com/beclab/l4-bfl-proxy/internal/ir"
	"github.com/beclab/l4-bfl-proxy/internal/message"
)

const edgeForwardClusterName = "app_gateway_edge_forward"

// buildEdgeForwardVirtualHosts builds Host-preserving reverse-proxy vhosts to AGW.
// Used when AppGatewayMTLS is on: L4 terminates user TLS only; AGW owns all HTTP routing.
func (t *Translator) buildEdgeForwardVirtualHosts(user *message.UserInfo, zone string, clusterSet map[string]*ir.ClusterIR) []*ir.VirtualHostIR {
	if user == nil || t == nil {
		return nil
	}
	agw := t.clusterForAppGateway(edgeForwardClusterName)
	if agw == nil {
		return nil
	}
	clusterSet[edgeForwardClusterName] = agw

	domains := edgeForwardDomains(user, zone)
	if len(domains) == 0 {
		return nil
	}
	makeVHost := func(suffix string, domains []string) *ir.VirtualHostIR {
		return &ir.VirtualHostIR{
			Name:     "edge_forward_" + user.Name + suffix,
			Domains:  domains,
			Language: user.Language,
			UserZone: zone,
			UserName: user.Name,
			Routes: []*ir.HTTPRouteIR{{
				Name:             "edge_forward_root_" + user.Name + suffix,
				PathPrefix:       "/",
				Cluster:          edgeForwardClusterName,
				WebSocketUpgrade: true,
			}},
		}
	}
	if !user.DenyAll {
		return []*ir.VirtualHostIR{makeVHost("", domains)}
	}

	// Exact public hosts must not share the virtual host containing the
	// private wildcard. Envoy prefers exact domains to wildcard domains.
	allowed := make(map[string]bool, len(user.AllowedDomains))
	for _, host := range user.AllowedDomains {
		host = strings.ToLower(strings.TrimSpace(host))
		if host != "" && !strings.ContainsAny(host, "*:/") {
			allowed[host] = true
		}
	}
	var public, restricted []string
	for _, domain := range domains {
		if allowed[domain] {
			public = append(public, domain)
			continue
		}
		restricted = append(restricted, domain)
		if suffix, ok := strings.CutPrefix(domain, "*."); ok {
			for host := range allowed {
				if label, ok := strings.CutSuffix(host, "."+suffix); ok && label != "" && !strings.Contains(label, ".") {
					public = append(public, host)
				}
			}
		}
	}
	slices.Sort(public)
	public = slices.Compact(public)
	var out []*ir.VirtualHostIR
	if len(restricted) != 0 {
		out = append(out, makeVHost("", restricted))
	}
	if len(public) != 0 {
		out = append(out, makeVHost("_public", public))
	}
	return out
}

func edgeForwardDomains(user *message.UserInfo, zone string) []string {
	seen := map[string]struct{}{}
	var out []string
	add := func(d string) {
		d = strings.TrimSpace(strings.ToLower(d))
		if d == "" {
			return
		}
		if _, ok := seen[d]; ok {
			return
		}
		seen[d] = struct{}{}
		out = append(out, d)
	}

	if zone != "" {
		// Zone apex serves profile; wildcards cover auth/desktop/wizard/apps.
		add(zone)
		add("*." + zone)
		tokens := strings.Split(zone, ".")
		if len(tokens) > 1 {
			localZone := tokens[0] + ".olares.local"
			add(localZone)
			add("*." + localZone)
		}
	}

	for _, app := range user.Apps {
		if app == nil {
			continue
		}
		customDomainMap := parseSettingsJSON(app.Settings, settingsCustomDomain)
		for _, entrance := range app.Entrances {
			if entrance == nil {
				continue
			}
			if entranceMap, ok := customDomainMap[entrance.Name]; ok {
				if cd := entranceMap[settingsCustomDomainThirdPartyDomain]; cd != "" {
					add(cd)
				}
			}
		}
	}
	for _, cert := range user.CustomDomainCerts {
		if cert != nil && cert.Domain != "" {
			add(cert.Domain)
		}
	}
	return out
}
