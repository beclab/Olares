package message

import (
	"reflect"

	"bytetrade.io/web3os/bfl/internal/ingress/ir"
	cachetypes "github.com/envoyproxy/go-control-plane/pkg/cache/types"
	"github.com/telepresenceio/watchable"
	"google.golang.org/protobuf/proto"
)

type SSLConfig struct {
	Zone      string
	CertData  string
	KeyData   string
	Ephemeral bool
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

type Resources struct {
	SSL               *SSLConfig
	Apps              []*AppInfo
	Language          string
	UserZone          string
	UserName          string
	IsEphemeralUser   bool
	FileserverNodes   []*FileserverNodeInfo
	CustomDomainCerts []*CertInfo
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

func (r *Resources) DeepCopy() *Resources {
	if r == nil {
		return nil
	}
	out := *r
	if r.SSL != nil {
		cp := *r.SSL
		out.SSL = &cp
	}
	if r.Apps != nil {
		out.Apps = make([]*AppInfo, len(r.Apps))
		for i, app := range r.Apps {
			out.Apps[i] = app.DeepCopy()
		}
	}
	if r.FileserverNodes != nil {
		out.FileserverNodes = make([]*FileserverNodeInfo, len(r.FileserverNodes))
		for i, n := range r.FileserverNodes {
			cp := *n
			out.FileserverNodes[i] = &cp
		}
	}
	if r.CustomDomainCerts != nil {
		out.CustomDomainCerts = make([]*CertInfo, len(r.CustomDomainCerts))
		for i, c := range r.CustomDomainCerts {
			cp := *c
			out.CustomDomainCerts[i] = &cp
		}
	}
	return &out
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

// XdsResources is a watchable map connecting the xDS translator to the xDS server.
type XdsResources struct {
	watchable.Map[string, *XdsSnapshot]
}
