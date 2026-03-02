package message

import (
	"reflect"
	"sort"

	"github.com/beclab/l4-bfl-proxy/internal/ir"
	cachetypes "github.com/envoyproxy/go-control-plane/pkg/cache/types"
	"github.com/telepresenceio/watchable"
	"google.golang.org/protobuf/proto"
)

type UserInfo struct {
	Name              string
	Namespace         string
	Did               string
	Zone              string
	IsEphemeral       bool
	AccessLevel       uint64
	AllowCIDRs        []string
	DenyAll           bool
	AllowedDomains    []string
	ServerNameDomains []string
	LocalDomainIP     string
	CreateTimestamp   int64
	Language          string
	CustomDomainCerts []*CertInfo
	SSL               *SSLConfig
	Apps              []*AppInfo
	FileserverNodes   []*FileserverNodeInfo
}

type AppInfo struct {
	Name      string
	Appid     string
	IsSysApp  bool
	Namespace string
	Owner     string
	Entrances []*EntranceInfo
	Ports     []*PortInfo
	Settings  map[string]string
}

type EntranceInfo struct {
	Name            string
	Host            string
	Port            int32
	AuthLevel       string
	WindowPushState bool
}

type PortInfo struct {
	Name       string
	Host       string
	Port       int32
	ExposePort int32
	Protocol   string
}
type FileserverNodeInfo struct {
	NodeName string
	PodIP    string
	IsMaster bool
}

type CertInfo struct {
	Domain   string
	CertData string
	KeyData  string
}

type SSLConfig struct {
	Zone      string
	CertData  string
	KeyData   string
	Ephemeral bool
}

type Resources struct {
	Users []*UserInfo
}

func (r *Resources) DeepCopy() *Resources {
	if r == nil {
		return nil
	}
	out := &Resources{}
	if r.Users != nil {
		for _, u := range r.Users {
			uc := *u
			uc.AllowCIDRs = append([]string(nil), u.AllowCIDRs...)
			uc.AllowedDomains = append([]string(nil), u.AllowedDomains...)
			uc.ServerNameDomains = append([]string(nil), u.ServerNameDomains...)
			uc.CustomDomainCerts = append([]*CertInfo(nil), u.CustomDomainCerts...)
			if u.CustomDomainCerts != nil {
				for _, c := range u.CustomDomainCerts {
					cp := *c
					uc.CustomDomainCerts = append(uc.CustomDomainCerts, &cp)
				}
			}
			if u.FileserverNodes != nil {
				for _, fn := range u.FileserverNodes {
					cp := *fn
					uc.FileserverNodes = append(uc.FileserverNodes, &cp)
				}
			}
			if u.SSL != nil {
				cp := *u.SSL
				uc.SSL = &cp
			}
			if u.Apps != nil {
				uc.Apps = make([]*AppInfo, len(u.Apps))
				for i, app := range u.Apps {
					uc.Apps[i] = app.DeepCopy()
				}
			}
			if u.FileserverNodes != nil {
				uc.FileserverNodes = append([]*FileserverNodeInfo(nil), u.FileserverNodes...)

			}
			out.Users = append(out.Users, &uc)
		}
	}
	return out
}

// Sort sorts all list fields for deterministic comparison and stable xDS output.
func (r *Resources) Sort() {
	sort.Slice(r.Users, func(i, j int) bool {
		return r.Users[i].Name < r.Users[j].Name
	})
	for _, u := range r.Users {
		sort.Strings(u.AllowCIDRs)
		sort.Strings(u.AllowedDomains)
		sort.Strings(u.ServerNameDomains)
		sort.Slice(u.CustomDomainCerts, func(i, j int) bool {
			return u.CustomDomainCerts[i].Domain < u.CustomDomainCerts[j].Domain
		})
		sort.Slice(u.FileserverNodes, func(i, j int) bool {
			return u.FileserverNodes[i].NodeName < u.FileserverNodes[j].NodeName
		})
		sort.Slice(u.Apps, func(i, j int) bool {
			return u.Apps[i].Name < u.Apps[j].Name
		})
	}
}

func (r *Resources) Equal(other *Resources) bool {
	if r == nil && other == nil {
		return true
	}
	if r == nil || other == nil {
		return false
	}
	return reflect.DeepEqual(r, other)
}

func (a *AppInfo) DeepCopy() *AppInfo {
	if a == nil {
		return nil
	}
	out := *a
	if a.Entrances != nil {
		out.Entrances = make([]*EntranceInfo, len(a.Entrances))
		for i, e := range a.Entrances {
			cp := *e
			out.Entrances[i] = &cp
		}
	}
	if a.Ports != nil {
		out.Ports = make([]*PortInfo, len(a.Ports))
		for i, p := range a.Ports {
			cp := *p
			out.Ports[i] = &cp
		}
	}
	if a.Settings != nil {
		out.Settings = make(map[string]string, len(a.Settings))
		for k, v := range a.Settings {
			out.Settings[k] = v
		}
	}
	return &out
}

// ProviderResources is a watchable map connecting the provider to the translator.
type ProviderResources struct {
	watchable.Map[string, *Resources]
}

// XdsIR is a watchable map connecting the translator to the xDS translator.
type XdsIR struct {
	watchable.Map[string, *ir.Xds]
}

// XdsSnapshot holds the translated envoy protobuf resources.
type XdsSnapshot struct {
	Listeners []cachetypes.Resource
	Clusters  []cachetypes.Resource
	Routes    []cachetypes.Resource
	Secrets   []cachetypes.Resource
}

func (x *XdsSnapshot) DeepCopy() *XdsSnapshot {
	if x == nil {
		return nil
	}
	out := &XdsSnapshot{}
	out.Listeners = cloneResources(x.Listeners)
	out.Clusters = cloneResources(x.Clusters)
	out.Routes = cloneResources(x.Routes)
	out.Secrets = cloneResources(x.Secrets)
	return out
}

func (x *XdsSnapshot) Equal(other *XdsSnapshot) bool {
	if x == nil && other == nil {
		return true
	}
	if x == nil || other == nil {
		return false
	}
	return resourcesEqual(x.Listeners, other.Listeners) &&
		resourcesEqual(x.Clusters, other.Clusters) &&
		resourcesEqual(x.Routes, other.Routes) &&
		resourcesEqual(x.Secrets, other.Secrets)
}

// XdsResources is a watchable map connecting the xDS translator to the xDS server.
type XdsResources struct {
	watchable.Map[string, *XdsSnapshot]
}

func (p *ProviderResources) Close() {
	p.Map.Close()
}

func (x *XdsIR) Close() {
	x.Map.Close()
}

func (x *XdsResources) Close() {
	x.Map.Close()
}

func cloneResources(src []cachetypes.Resource) []cachetypes.Resource {
	if src == nil {
		return nil
	}
	out := make([]cachetypes.Resource, len(src))
	for i, r := range src {
		out[i] = proto.Clone(r)
	}
	return out
}

func resourcesEqual(a, b []cachetypes.Resource) bool {
	if len(a) != len(b) {
		return false
	}
	for i := range a {
		if !proto.Equal(a[i], b[i]) {
			return false
		}
	}
	return true
}
