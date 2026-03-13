package ir

import "reflect"

type ProtocolType string

const (
	ProtocolHTTP ProtocolType = "HTTP"
	ProtocolTLS  ProtocolType = "TLS"
	ProtocolTCP  ProtocolType = "TCP"
	ProtocolUDP  ProtocolType = "UDP"
)

type Xds struct {
	Listeners []*ListenerIR
}

type ListenerIR struct {
	Name          string
	Address       string
	Port          uint32
	Protocol      ProtocolType
	ProxyProtocol bool
	TLSInspector  bool
	Routes        []*RouteIR
	HTTPRedirect  *HTTPRedirectIR
}

type HTTPRedirectIR struct {
	Scheme string
	Code   int
}

type RouteIR struct {
	Name                  string
	SNIMatches            []string
	SourcePrefixRanges    []string
	Destination           *DestinationIR
	ProxyProtocolUpstream bool
}

type DestinationIR struct {
	Name string
	Host string
	Port uint32
}

func (x *Xds) Equal(other *Xds) bool {
	if x == nil && other == nil {
		return true
	}
	if x == nil || other == nil {
		return false
	}
	return reflect.DeepEqual(x.Listeners, other.Listeners)
}

func (x *Xds) DeepCopy() *Xds {
	if x == nil {
		return nil
	}
	out := &Xds{}
	for _, l := range x.Listeners {
		out.Listeners = append(out.Listeners, l.DeepCopy())
	}
	return out
}

func (l *ListenerIR) DeepCopy() *ListenerIR {
	if l == nil {
		return nil
	}
	out := *l
	out.Routes = nil
	for _, r := range l.Routes {
		out.Routes = append(out.Routes, r.DeepCopy())
	}
	if l.HTTPRedirect != nil {
		redir := *l.HTTPRedirect
		out.HTTPRedirect = &redir
	}
	return &out
}

func (r *RouteIR) DeepCopy() *RouteIR {
	if r == nil {
		return nil
	}
	out := *r
	out.SNIMatches = append([]string(nil), r.SNIMatches...)
	out.SourcePrefixRanges = append([]string(nil), r.SourcePrefixRanges...)
	if r.Destination != nil {
		dest := *r.Destination
		out.Destination = &dest
	}
	return &out
}
