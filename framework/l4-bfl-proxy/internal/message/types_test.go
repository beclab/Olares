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
