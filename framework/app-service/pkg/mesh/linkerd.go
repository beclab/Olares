package mesh

import (
	"context"
	"sync"
	"time"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/client-go/dynamic"
	"k8s.io/klog/v2"
	ctrl "sigs.k8s.io/controller-runtime"

	"github.com/beclab/Olares/framework/app-service/pkg/cluster"
)

type boolCache struct {
	mu        sync.RWMutex
	value     bool
	loaded    bool
	expiresAt time.Time
}

func (c *boolCache) get() (bool, bool) {
	c.mu.RLock()
	defer c.mu.RUnlock()
	if !c.loaded || time.Now().After(c.expiresAt) {
		return false, false
	}
	return c.value, true
}

func (c *boolCache) set(v bool, ttl time.Duration) {
	c.mu.Lock()
	defer c.mu.Unlock()
	c.value = v
	c.loaded = true
	c.expiresAt = time.Now().Add(ttl)
}

var linkerdMeshEnabledCache = &boolCache{}

// GetLinkerdMeshEnabled reports whether the cluster runs Linkerd as the steady-state
// data plane (ADR-PLAN-11). Defaults to false until P0-a install sets the field.
func GetLinkerdMeshEnabled(ctx context.Context) bool {
	if v, ok := linkerdMeshEnabledCache.get(); ok {
		return v
	}
	cfg, err := ctrl.GetConfig()
	if err != nil {
		klog.V(2).Infof("mesh: GetConfig failed, default linkerdMeshEnabled=false: %v", err)
		return false
	}
	dc, err := dynamic.NewForConfig(cfg)
	if err != nil {
		klog.V(2).Infof("mesh: dynamic client init failed, default linkerdMeshEnabled=false: %v", err)
		return false
	}
	u, err := dc.Resource(cluster.Resource).Get(ctx, cluster.SingletonName, metav1.GetOptions{})
	if err != nil {
		klog.V(2).Infof("mesh: ClusterConfig get failed, default linkerdMeshEnabled=false: %v", err)
		return false
	}
	enabled := false
	if v, found, err := unstructured.NestedBool(u.Object, "spec", "linkerdMeshEnabled"); err == nil && found {
		enabled = v
	}
	linkerdMeshEnabledCache.set(enabled, 30*time.Second)
	return enabled
}

// PrimeLinkerdMeshEnabledForTest seeds the mesh gate cache for unit tests.
func PrimeLinkerdMeshEnabledForTest(v bool) {
	linkerdMeshEnabledCache.set(v, time.Hour)
}

// ResetLinkerdMeshEnabledForTest clears the mesh gate cache.
func ResetLinkerdMeshEnabledForTest() {
	linkerdMeshEnabledCache.mu.Lock()
	defer linkerdMeshEnabledCache.mu.Unlock()
	linkerdMeshEnabledCache.loaded = false
}

// ShouldSkipEnvoySidecar is the webhook gate for retiring olares-envoy-sidecar.
func ShouldSkipEnvoySidecar(ctx context.Context) bool {
	return GetLinkerdMeshEnabled(ctx)
}
