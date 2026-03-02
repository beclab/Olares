package ir

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestXdsDeepCopy(t *testing.T) {
	original := &Xds{
		Listeners: []*ListenerIR{
			{
				Name:          "tls_443",
				Address:       "0.0.0.0",
				Port:          443,
				Protocol:      ProtocolTLS,
				ProxyProtocol: false,
				TLSInspector:  true,
				Routes: []*RouteIR{
					{
						Name:       "route_alice",
						SNIMatches: []string{"alice.example.com", "*.alice.example.com"},
						Destination: &DestinationIR{
							Name: "user_alice",
							Host: "bfl.user-space-alice",
							Port: 444,
						},
						ProxyProtocolUpstream: true,
					},
				},
			},
			{
				Name:     "http_redirect_81",
				Address:  "0.0.0.0",
				Port:     81,
				Protocol: ProtocolHTTP,
				HTTPRedirect: &HTTPRedirectIR{
					Scheme: "https",
					Code:   301,
				},
			},
		},
	}

	copied := original.DeepCopy()

	require.Equal(t, len(original.Listeners), len(copied.Listeners))
	assert.Equal(t, original.Listeners[0].Name, copied.Listeners[0].Name)
	assert.Equal(t, original.Listeners[0].Routes[0].SNIMatches, copied.Listeners[0].Routes[0].SNIMatches)
	assert.Equal(t, original.Listeners[1].HTTPRedirect.Scheme, copied.Listeners[1].HTTPRedirect.Scheme)

	// Mutate the copy and verify the original is unaffected.
	copied.Listeners[0].Name = "mutated"
	copied.Listeners[0].Routes[0].SNIMatches[0] = "mutated.example.com"
	copied.Listeners[0].Routes[0].Destination.Host = "mutated-host"
	copied.Listeners[1].HTTPRedirect.Scheme = "http"

	assert.Equal(t, "tls_443", original.Listeners[0].Name)
	assert.Equal(t, "alice.example.com", original.Listeners[0].Routes[0].SNIMatches[0])
	assert.Equal(t, "bfl.user-space-alice", original.Listeners[0].Routes[0].Destination.Host)
	assert.Equal(t, "https", original.Listeners[1].HTTPRedirect.Scheme)
}

func TestXdsEqual_Identical(t *testing.T) {
	a := &Xds{Listeners: []*ListenerIR{
		{Name: "tls_443", Port: 443, Protocol: ProtocolTLS},
		{Name: "http_81", Port: 81, Protocol: ProtocolHTTP},
	}}
	b := &Xds{Listeners: []*ListenerIR{
		{Name: "tls_443", Port: 443, Protocol: ProtocolTLS},
		{Name: "http_81", Port: 81, Protocol: ProtocolHTTP},
	}}
	assert.True(t, a.Equal(b))
}

func TestXdsEqual_Different(t *testing.T) {
	a := &Xds{Listeners: []*ListenerIR{
		{Name: "tls_443", Port: 443},
	}}
	b := &Xds{Listeners: []*ListenerIR{
		{Name: "tls_444", Port: 444},
	}}
	assert.False(t, a.Equal(b))
}

func TestXdsEqual_DifferentLength(t *testing.T) {
	a := &Xds{Listeners: []*ListenerIR{{Name: "a"}}}
	b := &Xds{Listeners: []*ListenerIR{{Name: "a"}, {Name: "b"}}}
	assert.False(t, a.Equal(b))
}

func TestXdsEqual_BothNil(t *testing.T) {
	var a, b *Xds
	assert.True(t, a.Equal(b))
}

func TestXdsEqual_OneNil(t *testing.T) {
	a := &Xds{}
	assert.False(t, a.Equal(nil))
	var b *Xds
	assert.False(t, b.Equal(a))
}

func TestXdsDeepCopyNil(t *testing.T) {
	var x *Xds
	assert.Nil(t, x.DeepCopy())
}

func TestListenerIRDeepCopyNil(t *testing.T) {
	var l *ListenerIR
	assert.Nil(t, l.DeepCopy())
}

func TestRouteIRDeepCopyNil(t *testing.T) {
	var r *RouteIR
	assert.Nil(t, r.DeepCopy())
}

func TestRouteIRDeepCopySourcePrefixRanges(t *testing.T) {
	original := &RouteIR{
		Name:               "route_restricted",
		SNIMatches:         []string{"a.example.com"},
		SourcePrefixRanges: []string{"10.0.0.0/8", "192.168.1.0/24"},
		Destination: &DestinationIR{
			Name: "cluster_a",
			Host: "10.0.0.1",
			Port: 8080,
		},
	}

	copied := original.DeepCopy()

	copied.SourcePrefixRanges[0] = "172.16.0.0/12"
	assert.Equal(t, "10.0.0.0/8", original.SourcePrefixRanges[0])
}
