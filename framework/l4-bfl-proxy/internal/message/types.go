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
	BFLHost           string
	BFLPort           int
	AccessLevel       uint64
	AllowCIDRs        []string
	DenyAll           bool
	AllowedDomains    []string
	ServerNameDomains []string
	LocalDomainIP     string
	CreateTimestamp   int64
}

type AppInfo struct {
	Name      string
	Appid     string
	Owner     string
	Entrances []EntranceInfo
	Ports     []PortInfo
}

type EntranceInfo struct {
	Name      string
	AuthLevel string
}

type PortInfo struct {
	Name       string
	Host       string
	Port       int32
	ExposePort int32
	Protocol   string
}

type Resources struct {
	Users []*UserInfo
	Apps  []*AppInfo
}

func (r *Resources) DeepCopy() *Resources {
	if r == nil {
		return nil
	}
	out := &Resources{}
	for _, u := range r.Users {
		uc := *u
		uc.AllowCIDRs = append([]string(nil), u.AllowCIDRs...)
		uc.AllowedDomains = append([]string(nil), u.AllowedDomains...)
		uc.ServerNameDomains = append([]string(nil), u.ServerNameDomains...)
		out.Users = append(out.Users, &uc)
	}
	for _, a := range r.Apps {
		ac := *a
		ac.Entrances = append([]EntranceInfo(nil), a.Entrances...)
		ac.Ports = append([]PortInfo(nil), a.Ports...)
		out.Apps = append(out.Apps, &ac)
	}
	return out
}

// Sort sorts Users by Name and Apps by Name for deterministic comparison.
func (r *Resources) Sort() {
	sort.Slice(r.Users, func(i, j int) bool {
		return r.Users[i].Name < r.Users[j].Name
	})
	sort.Slice(r.Apps, func(i, j int) bool {
		return r.Apps[i].Name < r.Apps[j].Name
	})
}

func (r *Resources) Equal(other *Resources) bool {
	if r == nil && other == nil {
		return true
	}
	if r == nil || other == nil {
		return false
	}
	return reflect.DeepEqual(r.Users, other.Users) && reflect.DeepEqual(r.Apps, other.Apps)
}

func (x *XdsSnapshot) Equal(other *XdsSnapshot) bool {
	if x == nil && other == nil {
		return true
	}
	if x == nil || other == nil {
		return false
	}
	if len(x.Listeners) != len(other.Listeners) || len(x.Clusters) != len(other.Clusters) {
		return false
	}
	for i := range x.Listeners {
		if !proto.Equal(x.Listeners[i], other.Listeners[i]) {
			return false
		}
	}
	for i := range x.Clusters {
		if !proto.Equal(x.Clusters[i], other.Clusters[i]) {
			return false
		}
	}
	return true
}

type ProviderResources struct {
	watchable.Map[string, *Resources]
}

type XdsIR struct {
	watchable.Map[string, *ir.Xds]
}

type XdsSnapshot struct {
	Listeners []cachetypes.Resource
	Clusters  []cachetypes.Resource
}

func (x *XdsSnapshot) DeepCopy() *XdsSnapshot {
	if x == nil {
		return nil
	}
	out := &XdsSnapshot{}
	for _, l := range x.Listeners {
		out.Listeners = append(out.Listeners, proto.Clone(l))
	}
	for _, c := range x.Clusters {
		out.Clusters = append(out.Clusters, proto.Clone(c))
	}
	return out
}

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
