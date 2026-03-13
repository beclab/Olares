package ir

import "reflect"

type Xds struct {
	HTTPListeners   []*HTTPListenerIR
	StreamListeners []*StreamListenerIR
	Clusters        []*ClusterIR
	Secrets         []*SecretIR
}

type HTTPListenerIR struct {
	Name            string
	Address         string
	Port            uint32
	TLS             bool
	ProxyProtocol   bool
	IsRedirect      bool
	VirtualHosts    []*VirtualHostIR
	DefaultResponse *DirectResponseIR
	PingRoute       bool
}

type VirtualHostIR struct {
	Name                  string
	Domains               []string
	Routes                []*HTTPRouteIR
	EnableOIDC            bool
	EnableWindowPushState bool
	Language              string
	UserZone              string
	UserName              string
}

type HTTPRouteIR struct {
	Name             string
	PathPrefix       string
	PathExact        string
	PathRegex        string
	Cluster          string
	RequestHeaders   map[string]string
	DirectResponse   *DirectResponseIR
	ExtAuth          *ExtAuthConfigIR
	WebSocketUpgrade bool
}

type DirectResponseIR struct {
	Status      uint32
	Body        string
	ContentType string
}

type ExtAuthConfigIR struct {
	Cluster        string
	PathPrefix     string
	RequestHeaders []string
	Disabled       bool
}

type StreamListenerIR struct {
	Name     string
	Address  string
	Port     uint32
	Protocol string // "tcp" or "udp"
	Cluster  string
}

type ClusterIR struct {
	Name   string
	Host   string
	Port   uint32
	UseDNS bool
}

type SecretIR struct {
	Name     string
	CertData string
	KeyData  string
}

func (x *Xds) Equal(other *Xds) bool {
	if x == nil && other == nil {
		return true
	}
	if x == nil || other == nil {
		return false
	}
	return reflect.DeepEqual(x, other)
}

func (x *Xds) DeepCopy() *Xds {
	if x == nil {
		return nil
	}
	out := &Xds{}
	for _, l := range x.HTTPListeners {
		out.HTTPListeners = append(out.HTTPListeners, l.DeepCopy())
	}
	for _, l := range x.StreamListeners {
		cp := *l
		out.StreamListeners = append(out.StreamListeners, &cp)
	}
	for _, c := range x.Clusters {
		cp := *c
		out.Clusters = append(out.Clusters, &cp)
	}
	for _, s := range x.Secrets {
		cp := *s
		out.Secrets = append(out.Secrets, &cp)
	}
	return out
}

func (l *HTTPListenerIR) DeepCopy() *HTTPListenerIR {
	if l == nil {
		return nil
	}
	out := *l
	out.VirtualHosts = nil
	for _, vh := range l.VirtualHosts {
		out.VirtualHosts = append(out.VirtualHosts, vh.DeepCopy())
	}
	if l.DefaultResponse != nil {
		cp := *l.DefaultResponse
		out.DefaultResponse = &cp
	}
	return &out
}

func (vh *VirtualHostIR) DeepCopy() *VirtualHostIR {
	if vh == nil {
		return nil
	}
	out := *vh
	out.Domains = append([]string(nil), vh.Domains...)
	out.Routes = nil
	for _, r := range vh.Routes {
		out.Routes = append(out.Routes, r.DeepCopy())
	}
	return &out
}

func (r *HTTPRouteIR) DeepCopy() *HTTPRouteIR {
	if r == nil {
		return nil
	}
	out := *r
	if r.RequestHeaders != nil {
		out.RequestHeaders = make(map[string]string, len(r.RequestHeaders))
		for k, v := range r.RequestHeaders {
			out.RequestHeaders[k] = v
		}
	}
	if r.DirectResponse != nil {
		cp := *r.DirectResponse
		out.DirectResponse = &cp
	}
	if r.ExtAuth != nil {
		cp := *r.ExtAuth
		out.ExtAuth = &cp
	}
	return &out
}
