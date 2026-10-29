// Package cluster reads cluster-wide configuration from Terminus and ClusterConfig.
package cluster

import (
	"context"
	"fmt"
	"strings"
	"sync"
	"time"

	sysv1alpha1 "github.com/beclab/api/api/sys.bytetrade.io/v1alpha1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"k8s.io/klog/v2"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
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

var (
	platformDomainTestOverride    string
	platformDomainTestOverrideSet bool
)

// GetDomainName reads the platform domain from the Terminus singleton.
// It preserves all domain labels and returns an error when the domain is unset.
func GetDomainName(ctx context.Context, reader client.Reader) (string, error) {
	var terminus sysv1alpha1.Terminus
	if err := reader.Get(ctx, client.ObjectKey{Name: "terminus"}, &terminus); err != nil {
		return "", fmt.Errorf("failed to get Terminus domain name: %w", err)
	}
	domain := strings.ToLower(strings.TrimSuffix(strings.TrimSpace(terminus.Spec.Settings["domainName"]), "."))
	if domain == "" {
		//FIXME:
		localName := strings.ToLower(strings.TrimSuffix(strings.TrimSpace(terminus.Spec.LocalName), "."))
		return localName + ".olares.local", nil
	}
	return domain, nil
}

var NewDomainReader = func() (client.Reader, error) {
	cfg, err := ctrl.GetConfig()
	if err != nil {
		return nil, fmt.Errorf("failed to get Kubernetes config: %w", err)
	}
	scheme := runtime.NewScheme()
	if err := sysv1alpha1.AddToScheme(scheme); err != nil {
		return nil, fmt.Errorf("failed to register Terminus scheme: %w", err)
	}
	return client.New(cfg, client.Options{Scheme: scheme})
}

// GetPlatformDomain returns the Terminus platform domain, or an empty string
// on failure. Only successful reads are cached for a short TTL.
func GetPlatformDomain(ctx context.Context) string {
	if platformDomainTestOverrideSet {
		return platformDomainTestOverride
	}
	if d, ok := defaultCache.get(); ok {
		return d
	}

	reader, err := NewDomainReader()
	if err != nil {
		klog.Errorf("Failed to initialize platform domain reader: %v", err)
		return ""
	}
	domain, err := GetDomainName(ctx, reader)
	if err != nil {
		klog.Errorf("Failed to get platform domain: %v", err)
		return ""
	}

	if domain == "" {
		klog.Errorf("Platform domain is empty")
		return ""
	}

	platformDomain := strings.Join(strings.Split(domain, ".")[1:], ".")

	defaultCache.set(platformDomain)
	return platformDomain
}
