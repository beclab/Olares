package cluster

import (
	"context"
	"sync"
	"time"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/client-go/dynamic"
	"k8s.io/klog/v2"
	ctrl "sigs.k8s.io/controller-runtime"
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

var gatewayEnabledCache = &boolCache{}

// GetInClusterGatewayEnabled reads ClusterConfig.spec.inClusterGatewayEnabled.
// Absent CR, absent field, or API errors all default to true so the shared
// in-cluster gateway path is on by default (power-on ready); only an explicit
// false disables it. Successful reads are cached for a short TTL.
func GetInClusterGatewayEnabled(ctx context.Context) bool {
	if v, ok := gatewayEnabledCache.get(); ok {
		return v
	}
	cfg, err := ctrl.GetConfig()
	if err != nil {
		klog.V(2).Infof("cluster: GetConfig failed, default inClusterGatewayEnabled=true: %v", err)
		return true
	}
	dc, err := dynamic.NewForConfig(cfg)
	if err != nil {
		klog.V(2).Infof("cluster: dynamic client init failed, default inClusterGatewayEnabled=true: %v", err)
		return true
	}
	u, err := dc.Resource(Resource).Get(ctx, SingletonName, metav1.GetOptions{})
	if err != nil {
		klog.V(2).Infof("cluster: ClusterConfig get failed, default inClusterGatewayEnabled=true: %v", err)
		return true
	}
	enabled := true
	if v, found, err := unstructured.NestedBool(u.Object, "spec", "inClusterGatewayEnabled"); err == nil && found {
		enabled = v
	}
	gatewayEnabledCache.set(enabled, cacheTTL)
	return enabled
}

// PrimeInClusterGatewayEnabledForTest seeds the gate cache so unit tests can
// exercise both branches without an API server.
func PrimeInClusterGatewayEnabledForTest(v bool) {
	gatewayEnabledCache.set(v, time.Hour)
}

// ResetInClusterGatewayEnabledForTest clears the gate cache.
func ResetInClusterGatewayEnabledForTest() {
	gatewayEnabledCache.mu.Lock()
	defer gatewayEnabledCache.mu.Unlock()
	gatewayEnabledCache.loaded = false
}
