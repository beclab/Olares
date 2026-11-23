package translator

import (
	"slices"
	"strings"

	"github.com/beclab/l4-bfl-proxy/internal/ir"
	"github.com/beclab/l4-bfl-proxy/internal/message"
)

const edgeForwardClusterName = "app_gateway_edge_forward"

// buildEdgeForwardVirtualHosts builds Host-preserving reverse-proxy vhosts to AGW.
// L4 authorizes browser traffic, and AGW owns all upstream routing.
func (t *Translator) buildEdgeForwardVirtualHosts(user *message.UserInfo, zone string, clusterSet map[string]*ir.ClusterIR, extraNoAuthHosts ...string) []*ir.VirtualHostIR {
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
	// Login and activation must reach their own handlers without an ExtAuth
	// check. Shared routes keep the AGW caller-JWT contract.
	withoutExtAuth := map[string]bool{}
	for _, domain := range append([]string{zone}, user.ServerNameDomains...) {
		if domain != "" {
			withoutExtAuth["auth."+domain] = true
			withoutExtAuth["wizard."+domain] = true
		}
	}
	if local := toLocalDomain("auth." + zone); zone != "" {
		withoutExtAuth[local] = true
		withoutExtAuth[toLocalDomain("wizard."+zone)] = true
	}
	for _, app := range user.Apps {
		if app == nil || !app.IsShared {
			continue
		}
		// Reuse the same exact aliases emitted by existing app vhosts.
		unused := map[string]*ir.ClusterIR{}
		appHosts := t.buildAppVirtualHosts(user, app, zone, false, unused)
		appHosts = append(appHosts, t.buildCustomDomainVirtualHosts(user, app, unused)...)
		for _, vhost := range appHosts {
			for _, host := range vhost.Domains {
				withoutExtAuth[strings.ToLower(host)] = true
			}
		}
	}
	for _, host := range extraNoAuthHosts {
		withoutExtAuth[host] = true
	}
	makeVHost := func(suffix string, domains []string, check bool) *ir.VirtualHostIR {
		var extAuth *ir.ExtAuthConfigIR
		if check {
			extAuth = buildAppExtAuthConfig(user)
		}
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
				ExtAuth:          extAuth,
			}},
		}
	}
	allowed := make(map[string]bool, len(user.AllowedDomains))
	for _, host := range user.AllowedDomains {
		host = strings.ToLower(strings.TrimSpace(host))
		if host != "" && !strings.ContainsAny(host, "*:/") {
			allowed[host] = true
		}
	}
	selected := make(map[string]bool, len(domains))
	for _, domain := range domains {
		selected[domain] = true
		if suffix, wildcard := strings.CutPrefix(domain, "*."); wildcard {
			for host := range withoutExtAuth {
				if label, ok := strings.CutSuffix(host, "."+suffix); ok && label != "" && !strings.Contains(label, ".") {
					selected[host] = true
				}
			}
			for host := range allowed {
				if label, ok := strings.CutSuffix(host, "."+suffix); ok && label != "" && !strings.Contains(label, ".") {
					selected[host] = true
				}
			}
		}
	}
	var out []*ir.VirtualHostIR
	for _, group := range []struct {
		suffix        string
		check, public bool
	}{
		{"", true, false}, {"_public", true, true},
		{"_no_auth", false, false}, {"_no_auth_public", false, true},
	} {
		var hosts []string
		for host := range selected {
			if group.check == !withoutExtAuth[host] && group.public == (user.DenyAll && allowed[host]) {
				hosts = append(hosts, host)
			}
		}
		if len(hosts) > 0 {
			slices.Sort(hosts)
			out = append(out, makeVHost(group.suffix, hosts, group.check))
		}
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
