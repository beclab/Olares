package translator

import (
	"fmt"
	"sort"

	"github.com/beclab/l4-bfl-proxy/internal/ir"
	"github.com/beclab/l4-bfl-proxy/internal/message"
)

// machineVirtualHosts builds the shared front door once. Apps retain their
// owning user's route context; shared apps authorize the actual session user.
func (t *Translator) machineVirtualHosts(resources *message.Resources, clusters map[string]*ir.ClusterIR) ([]*ir.VirtualHostIR, map[string][]*ir.VirtualHostIR) {
	domain := resources.DefaultDomain
	localZone := resources.LocalZone
	if localZone == "" {
		localZone = "olares.local"
	}
	zone := domain
	if zone == "" {
		zone = localZone
	}
	vhosts := t.buildMachineSystemVirtualHosts(domain, localZone, clusters)
	custom := make(map[string][]*ir.VirtualHostIR)
	for _, user := range resources.Users {
		personal := *user
		personal.Zone = zone
		var privateApps []*ir.VirtualHostIR
		for _, app := range user.Apps {
			actor := &personal
			if app.IsShared {
				actor = &message.UserInfo{Zone: zone, Language: "en-US", FileserverNodes: user.FileserverNodes}
			}
			apps := t.buildAppVirtualHosts(actor, app, zone, clusters, localZone)
			if app.IsShared {
				vhosts = append(vhosts, apps...)
			} else {
				privateApps = append(privateApps, apps...)
			}
			custom[user.Name] = append(custom[user.Name], t.buildCustomDomainVirtualHosts(&personal, app, clusters)...)
		}
		t.applyDenyAllRestrictions(&personal, privateApps, nil, resources.LocalDomainIP)
		t.applyDenyAllRestrictions(&personal, custom[user.Name], nil, resources.LocalDomainIP)
		vhosts = append(vhosts, privateApps...)
	}
	return uniqueMachineVirtualHosts(vhosts), custom
}

func uniqueMachineVirtualHosts(vhosts []*ir.VirtualHostIR) []*ir.VirtualHostIR {
	sort.SliceStable(vhosts, func(i, j int) bool {
		if vhosts[i].Priority != vhosts[j].Priority {
			return vhosts[i].Priority > vhosts[j].Priority
		}
		return vhosts[i].Name < vhosts[j].Name
	})
	seen := map[string]bool{}
	var result []*ir.VirtualHostIR
	for _, vh := range vhosts {
		copy := vh.DeepCopy()
		copy.Domains = nil
		for _, host := range vh.Domains {
			if host == "" || seen[host] {
				continue
			}
			seen[host] = true
			copy.Domains = append(copy.Domains, host)
		}
		if len(copy.Domains) > 0 {
			result = append(result, copy)
		}
	}
	return result
}

func (t *Translator) buildMachineSystemVirtualHosts(domain, localZone string, clusters map[string]*ir.ClusterIR) []*ir.VirtualHostIR {
	zones := []string{localZone}
	if domain != "" && domain != localZone {
		zones = append(zones, domain)
	}
	clusters["profile_service"] = &ir.ClusterIR{Name: "profile_service", Host: "profile-service.os-frontend.svc.cluster.local", Port: 3000, UseDNS: true}
	result := []*ir.VirtualHostIR{{Name: "profile", Domains: zones, UserZone: domain, Language: "en-US", Priority: systemServicePriority, Routes: []*ir.HTTPRouteIR{{Name: "profile_root", PathPrefix: "/", Cluster: "profile_service", ExtAuth: buildAppExtAuthConfig()}}}}
	for _, def := range systemServices {
		name := "nonapp_" + def.Name
		namespace := def.Namespace
		if namespace == "" {
			namespace = "os-frontend"
		}
		clusters[name] = &ir.ClusterIR{Name: name, Host: fmt.Sprintf(def.SvcFormat, namespace), Port: 80, UseDNS: true}
		var domains []string
		for _, zone := range zones {
			domains = append(domains, def.Name+"."+zone)
		}
		route := &ir.HTTPRouteIR{Name: name + "_root", PathPrefix: "/", Cluster: name, WebSocketUpgrade: true}
		if def.Name != "auth" {
			route.ExtAuth = buildAppExtAuthConfig()
		}
		result = append(result, &ir.VirtualHostIR{Name: name, Domains: domains, UserZone: domain, Language: "en-US", Priority: systemServicePriority, Routes: []*ir.HTTPRouteIR{route}})
	}
	return result
}

func (t *Translator) machineTLSListener(name string, port uint32, proxyProtocol bool, vhosts []*ir.VirtualHostIR, certificate *ir.SecretIR, snis []string) *ir.HTTPListenerIR {
	return &ir.HTTPListenerIR{Name: name, Address: "0.0.0.0", Port: port, TLS: true, ProxyProtocol: proxyProtocol, VirtualHosts: scopeLocalExtAuth(vhosts), TLSCert: certificate, SNIMatches: snis, ExtAuthCluster: autheliaClusterPrefix, PingRoute: true, DefaultResponse: &ir.DirectResponseIR{Status: 421, Body: "<h1><a href='https://www.olares.com/'>Olares</a></h1>", ContentType: "text/html"}}
}
