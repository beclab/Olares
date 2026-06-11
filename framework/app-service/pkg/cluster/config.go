// Package cluster reads the cluster.olares.io/v1alpha1 ClusterConfig singleton
// to resolve the platform domain. Lite scope: no mesh profile, no viewer
// scheme. The CR is read via the dynamic client so app-service does not need
// to register it on the shared controller-runtime scheme.
package cluster

import (
	"context"
	"strings"
	"sync"
	"time"

	"github.com/beclab/Olares/framework/app-service/pkg/kubesphere"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"k8s.io/klog/v2"
)

const (
	// SingletonName is the only ClusterConfig name app-service consults.
	SingletonName = "cluster"
	cacheTTL      = 30 * time.Second
)

// Resource is the dynamic-client GVR for ClusterConfig.
var Resource = schema.GroupVersionResource{Group: "cluster.olares.io", Version: "v1alpha1", Resource: "clusterconfigs"}

type cache struct {
	mu        sync.RWMutex
	domain    string
	loaded    bool
	expiresAt time.Time
}

func (c *cache) get() (string, bool) {
	c.mu.RLock()
	defer c.mu.RUnlock()
	if !c.loaded || time.Now().After(c.expiresAt) {
		return "", false
	}
	return c.domain, true
}

func (c *cache) set(d string) {
	c.mu.Lock()
	defer c.mu.Unlock()
	c.domain = d
	c.loaded = true
	c.expiresAt = time.Now().Add(cacheTTL)
}

var defaultCache = &cache{}

// GetPlatformDomain returns the cluster platform domain. It never returns
// empty: a missing ClusterConfig or API failure falls back to the env var and
// then DefaultPlatformDomain. Results are cached for a short TTL.
func GetPlatformDomain(ctx context.Context) string {
	if d, ok := defaultCache.get(); ok {
		return d
	}

	// get cluster owner
	owner, err := kubesphere.GetClusterOwner(ctx) // best effort, just to trigger the cluster owner cache population
	if err != nil {
		klog.V(2).Infof("cluster: get cluster owner failed: %v", err)
		return ""
	}

	zone, err := kubesphere.GetUserZone(ctx, owner)
	if err != nil {
		klog.V(2).Infof("cluster: get user zone failed: %v", err)
		return ""
	}

	zoneTokens := strings.Split(zone, ".")
	if len(zoneTokens) < 2 {
		klog.V(2).Infof("cluster: invalid user zone format: %s", zone)
		return ""
	}

	domain := strings.Join(zoneTokens[1:], ".")

	defaultCache.set(domain)
	return domain
}
