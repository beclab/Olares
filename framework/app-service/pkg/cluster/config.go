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

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"k8s.io/client-go/dynamic"
	"k8s.io/klog/v2"
	ctrl "sigs.k8s.io/controller-runtime"
	"os"
)

const (
	// SingletonName is the only ClusterConfig name app-service consults.
	SingletonName = "cluster"
	// DefaultPlatformDomain is the fallback when no ClusterConfig exists and
	// OLARES_PLATFORM_DOMAIN is unset.
	DefaultPlatformDomain = "olares.com"
	// envPlatformDomain lets dev environments inject a domain without the CR.
	envPlatformDomain = "OLARES_PLATFORM_DOMAIN"
	cacheTTL          = 30 * time.Second
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

func envOrDefaultDomain() string {
	if d := strings.ToLower(strings.TrimSpace(os.Getenv(envPlatformDomain))); d != "" {
		return d
	}
	return DefaultPlatformDomain
}

// GetPlatformDomain returns the cluster platform domain. It never returns
// empty: a missing ClusterConfig or API failure falls back to the env var and
// then DefaultPlatformDomain. Results are cached for a short TTL.
func GetPlatformDomain(ctx context.Context) string {
	if d, ok := defaultCache.get(); ok {
		return d
	}
	domain := envOrDefaultDomain()
	cfg, err := ctrl.GetConfig()
	if err != nil {
		klog.V(2).Infof("cluster: GetConfig failed, using fallback domain %q: %v", domain, err)
		return domain
	}
	dc, err := dynamic.NewForConfig(cfg)
	if err != nil {
		klog.V(2).Infof("cluster: dynamic client init failed, using fallback domain %q: %v", domain, err)
		return domain
	}
	u, err := dc.Resource(Resource).Get(ctx, SingletonName, metav1.GetOptions{})
	if err != nil {
		klog.V(2).Infof("cluster: ClusterConfig get failed, using fallback domain %q: %v", domain, err)
		return domain
	}
	if pd, found, _ := unstructured.NestedString(u.Object, "spec", "platformDomain"); found {
		if pd = strings.ToLower(strings.TrimSpace(strings.TrimSuffix(pd, "."))); pd != "" {
			domain = pd
		}
	}
	defaultCache.set(domain)
	return domain
}
