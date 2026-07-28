package message

import (
	"testing"

	clusterv3 "github.com/envoyproxy/go-control-plane/envoy/config/cluster/v3"
	listenerv3 "github.com/envoyproxy/go-control-plane/envoy/config/listener/v3"
	cachetypes "github.com/envoyproxy/go-control-plane/pkg/cache/types"
	"github.com/stretchr/testify/assert"
)

// ---------------------------------------------------------------------------
// Resources.Sort
// ---------------------------------------------------------------------------

func TestResourcesSort(t *testing.T) {
	r := &Resources{
		Users: []*UserInfo{
			{Name: "charlie"},
			{Name: "alice"},
			{Name: "bob"},
		},
	}

	r.Sort()

	assert.Equal(t, "alice", r.Users[0].Name)
	assert.Equal(t, "bob", r.Users[1].Name)
	assert.Equal(t, "charlie", r.Users[2].Name)
}

// ---------------------------------------------------------------------------
// Resources.DeepCopy
// ---------------------------------------------------------------------------

// DeepCopy must produce the same number of slice entries (no doubling) and must
// not share backing pointers with the original.
func TestResourcesDeepCopy_NoDoubling(t *testing.T) {
	orig := &Resources{
		Users: []*UserInfo{{
			Name:              "alice",
			CustomDomainCerts: []*CertInfo{{Domain: "a.example.com"}, {Domain: "b.example.com"}},
			FileserverNodes:   []*FileserverNodeInfo{{NodeName: "n1"}, {NodeName: "n2"}, {NodeName: "n3"}},
		}},
	}

	cp := orig.DeepCopy()

	u := cp.Users[0]
	assert.Len(t, u.CustomDomainCerts, len(orig.Users[0].CustomDomainCerts))
	assert.Len(t, u.FileserverNodes, len(orig.Users[0].FileserverNodes))

	// Mutating the copy must not affect the original (deep, not shallow).
	u.CustomDomainCerts[0].Domain = "mutated"
	u.FileserverNodes[0].NodeName = "mutated"
	assert.Equal(t, "a.example.com", orig.Users[0].CustomDomainCerts[0].Domain)
	assert.Equal(t, "n1", orig.Users[0].FileserverNodes[0].NodeName)
}

// ---------------------------------------------------------------------------
// Resources.Equal
// ---------------------------------------------------------------------------

func TestResourcesEqual_Identical(t *testing.T) {
	a := &Resources{
		Users: []*UserInfo{{Name: "alice", Zone: "example.com"}},
	}
	b := &Resources{
		Users: []*UserInfo{{Name: "alice", Zone: "example.com"}},
	}
	assert.True(t, a.Equal(b))
}

func TestResourcesEqual_DifferentUser(t *testing.T) {
	a := &Resources{Users: []*UserInfo{{Name: "alice"}}}
	b := &Resources{Users: []*UserInfo{{Name: "bob"}}}
	assert.False(t, a.Equal(b))
}

func TestResourcesEqual_DifferentApp(t *testing.T) {
	a := &Resources{Users: []*UserInfo{{Name: "alice", Apps: []*AppInfo{{Appid: "v1"}}}}}
	b := &Resources{Users: []*UserInfo{{Name: "alice", Apps: []*AppInfo{{Appid: "v2"}}}}}
	assert.False(t, a.Equal(b))
}

func TestResourcesEqual_DifferentLengths(t *testing.T) {
	a := &Resources{Users: []*UserInfo{{Name: "alice"}}}
	b := &Resources{Users: []*UserInfo{{Name: "alice"}, {Name: "bob"}}}
	assert.False(t, a.Equal(b))
}

func TestResourcesEqual_BothNil(t *testing.T) {
	var a, b *Resources
	assert.True(t, a.Equal(b))
}

func TestResourcesEqual_OneNil(t *testing.T) {
	a := &Resources{}
	assert.False(t, a.Equal(nil))
}

// ---------------------------------------------------------------------------
// XdsSnapshot.Equal
// ---------------------------------------------------------------------------

func TestXdsSnapshotEqual_Identical(t *testing.T) {
	a := &XdsSnapshot{
		Listeners: []cachetypes.Resource{&listenerv3.Listener{Name: "tls_443"}},
		Clusters:  []cachetypes.Resource{&clusterv3.Cluster{Name: "user_alice"}},
	}
	b := &XdsSnapshot{
		Listeners: []cachetypes.Resource{&listenerv3.Listener{Name: "tls_443"}},
		Clusters:  []cachetypes.Resource{&clusterv3.Cluster{Name: "user_alice"}},
	}
	assert.True(t, a.Equal(b))
}

func TestXdsSnapshotEqual_DifferentListener(t *testing.T) {
	a := &XdsSnapshot{
		Listeners: []cachetypes.Resource{&listenerv3.Listener{Name: "tls_443"}},
	}
	b := &XdsSnapshot{
		Listeners: []cachetypes.Resource{&listenerv3.Listener{Name: "tls_444"}},
	}
	assert.False(t, a.Equal(b))
}

func TestXdsSnapshotEqual_DifferentCluster(t *testing.T) {
	a := &XdsSnapshot{
		Clusters: []cachetypes.Resource{&clusterv3.Cluster{Name: "c1"}},
	}
	b := &XdsSnapshot{
		Clusters: []cachetypes.Resource{&clusterv3.Cluster{Name: "c2"}},
	}
	assert.False(t, a.Equal(b))
}

func TestXdsSnapshotEqual_DifferentLengths(t *testing.T) {
	a := &XdsSnapshot{
		Listeners: []cachetypes.Resource{&listenerv3.Listener{Name: "a"}},
	}
	b := &XdsSnapshot{
		Listeners: []cachetypes.Resource{
			&listenerv3.Listener{Name: "a"},
			&listenerv3.Listener{Name: "b"},
		},
	}
	assert.False(t, a.Equal(b))
}

func TestXdsSnapshotEqual_BothNil(t *testing.T) {
	var a, b *XdsSnapshot
	assert.True(t, a.Equal(b))
}

func TestXdsSnapshotEqual_OneNil(t *testing.T) {
	a := &XdsSnapshot{}
	assert.False(t, a.Equal(nil))
}

func TestXdsSnapshotEqual_BothEmpty(t *testing.T) {
	a := &XdsSnapshot{}
	b := &XdsSnapshot{}
	assert.True(t, a.Equal(b))
}
