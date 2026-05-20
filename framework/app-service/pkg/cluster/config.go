// Package cluster exposes read helpers for the cluster.olares.io/v1alpha1
// ClusterConfig singleton introduced in Phase-A v2 (PR-6). The package only
// reads — write operations stay with the operator that owns the CRD lifecycle.
//
// ClusterConfig is intentionally consumed via the dynamic client so this
// package does not need to register the type with the in-tree scheme; that
// keeps the v2 increment additive and free of cross-cutting controller-runtime
// changes (Olares/framework/app-service has multiple controllers sharing the
// same scheme).
//
// References:
//   - archdoc/方案/shared应用/Shared外部访问主流程打通方案-2026-05-20-明确方案.md §4
//   - archdoc/方案/shared应用/Shared外部访问v2评审决议-2026-05-20.md  R-V2-1
package cluster

import (
	"context"
	"errors"
	"fmt"
	"os"
	"strings"
	"sync"
	"time"

	apierrors "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"k8s.io/client-go/dynamic"
	"k8s.io/client-go/rest"
	"k8s.io/klog/v2"
	ctrl "sigs.k8s.io/controller-runtime"
)

const (
	// SingletonName is the only ClusterConfig name app-service ever consults.
	SingletonName = "cluster"

	// DefaultPlatformDomain is the conservative fallback used when no
	// ClusterConfig exists yet and no environment override is supplied.
	// Aligned with the dev/test platform domain in archdoc.
	DefaultPlatformDomain = "olares.com"

	// SharedURLViewerSchemeEnabled signals that per-viewer Shared URLs and
	// logical hostPatterns should be emitted (R-V2-1).
	SharedURLViewerSchemeEnabled  = "enabled"
	SharedURLViewerSchemeDisabled = "disabled"

	// envPlatformDomain lets local/dev environments inject a platform domain
	// without first creating the ClusterConfig CR.
	envPlatformDomain = "OLARES_PLATFORM_DOMAIN"

	cacheTTL = 30 * time.Second
)

// GroupVersion of the ClusterConfig CRD.
var GroupVersion = schema.GroupVersion{Group: "cluster.olares.io", Version: "v1alpha1"}

// Resource is the dynamic-client GVR for ClusterConfig.
var Resource = schema.GroupVersionResource{Group: GroupVersion.Group, Version: GroupVersion.Version, Resource: "clusterconfigs"}

// Snapshot is a read-only projection of ClusterConfig.spec.
type Snapshot struct {
	PlatformDomain        string
	SharedURLViewerScheme string
}

// SharedURLViewerEnabled reports whether the v2 per-viewer Shared URL scheme
// should be used. Treats absent / "disabled" as false (Phase-A behaviour).
func (s Snapshot) SharedURLViewerEnabled() bool {
	return strings.EqualFold(s.SharedURLViewerScheme, SharedURLViewerSchemeEnabled)
}

// snapshotCache memoises the last ClusterConfig.Get for a short TTL so the
// per-Application reconcile loop does not hammer the API server. The cache is
// process-global because there is one ClusterConfig per cluster and every
// caller sees the same answer.
type snapshotCache struct {
	mu        sync.RWMutex
	snapshot  Snapshot
	loaded    bool
	expiresAt time.Time
}

func (c *snapshotCache) get() (Snapshot, bool) {
	c.mu.RLock()
	defer c.mu.RUnlock()
	if !c.loaded || time.Now().After(c.expiresAt) {
		return Snapshot{}, false
	}
	return c.snapshot, true
}

func (c *snapshotCache) set(s Snapshot) {
	c.mu.Lock()
	defer c.mu.Unlock()
	c.snapshot = s
	c.loaded = true
	c.expiresAt = time.Now().Add(cacheTTL)
}

var defaultCache = &snapshotCache{}

// Reader is the smallest interface this package needs from a Kubernetes client.
// It is satisfied by k8s.io/client-go/dynamic.NewForConfig(...).
// A custom Reader can be passed for tests.
type Reader interface {
	Get(ctx context.Context, name string, opts metav1.GetOptions) (*unstructured.Unstructured, error)
}

// dynamicReader adapts the dynamic client.
type dynamicReader struct{ ri dynamic.ResourceInterface }

func (d dynamicReader) Get(ctx context.Context, name string, opts metav1.GetOptions) (*unstructured.Unstructured, error) {
	return d.ri.Get(ctx, name, opts)
}

// GetSnapshot returns the current ClusterConfig.spec or a cached copy. It
// never returns an empty Snapshot: PlatformDomain is normalized to lowercase
// and falls back to the environment variable then DefaultPlatformDomain.
func GetSnapshot(ctx context.Context) (Snapshot, error) {
	if s, ok := defaultCache.get(); ok {
		return s, nil
	}
	cfg, err := ctrl.GetConfig()
	if err != nil {
		klog.V(2).Infof("cluster: GetConfig failed, using fallback platformDomain: %v", err)
		return fallbackSnapshot(), nil
	}
	return loadAndCache(ctx, cfg)
}

// GetPlatformDomain is the convenience accessor used by SRR writers and URL
// helpers. It is safe to call from per-reconcile code paths.
func GetPlatformDomain(ctx context.Context) string {
	s, _ := GetSnapshot(ctx)
	return s.PlatformDomain
}

// loadAndCache fetches the CR through the dynamic client, applies the
// fallback chain, and updates the package-level cache.
func loadAndCache(ctx context.Context, cfg *rest.Config) (Snapshot, error) {
	c, err := dynamic.NewForConfig(cfg)
	if err != nil {
		klog.V(2).Infof("cluster: dynamic client init failed, using fallback: %v", err)
		return fallbackSnapshot(), nil
	}
	return loadAndCacheWithReader(ctx, dynamicReader{ri: c.Resource(Resource)})
}

// loadAndCacheWithReader is the test-injectable form of loadAndCache.
func loadAndCacheWithReader(ctx context.Context, r Reader) (Snapshot, error) {
	u, err := r.Get(ctx, SingletonName, metav1.GetOptions{})
	if err != nil {
		if apierrors.IsNotFound(err) {
			s := fallbackSnapshot()
			defaultCache.set(s)
			return s, nil
		}
		// Transient API failure: serve the in-memory fallback but do not
		// poison the cache so the next call can succeed.
		klog.V(2).Infof("cluster: ClusterConfig get failed, using fallback: %v", err)
		return fallbackSnapshot(), nil
	}
	s, err := snapshotFromUnstructured(u)
	if err != nil {
		klog.Warningf("cluster: malformed ClusterConfig %q (%v); using fallback", SingletonName, err)
		s = fallbackSnapshot()
	}
	defaultCache.set(s)
	return s, nil
}

// snapshotFromUnstructured projects the dynamic-client object into Snapshot.
// Missing fields are tolerated and replaced by fallback values.
func snapshotFromUnstructured(u *unstructured.Unstructured) (Snapshot, error) {
	if u == nil {
		return Snapshot{}, errors.New("nil object")
	}
	spec, found, err := unstructured.NestedMap(u.Object, "spec")
	if err != nil {
		return Snapshot{}, fmt.Errorf("read spec: %w", err)
	}
	out := Snapshot{
		PlatformDomain:        envOrDefaultPlatformDomain(),
		SharedURLViewerScheme: SharedURLViewerSchemeDisabled,
	}
	if !found {
		return out, nil
	}
	if v, ok := spec["platformDomain"].(string); ok && v != "" {
		out.PlatformDomain = NormalizePlatformDomain(v)
	}
	if v, ok := spec["sharedURLViewerScheme"].(string); ok && v != "" {
		out.SharedURLViewerScheme = strings.ToLower(strings.TrimSpace(v))
	}
	if !validPlatformDomain(out.PlatformDomain) {
		return out, fmt.Errorf("invalid platformDomain %q", out.PlatformDomain)
	}
	return out, nil
}

// fallbackSnapshot is used whenever the CRD/API call cannot be served.
func fallbackSnapshot() Snapshot {
	return Snapshot{
		PlatformDomain:        envOrDefaultPlatformDomain(),
		SharedURLViewerScheme: SharedURLViewerSchemeDisabled,
	}
}

func envOrDefaultPlatformDomain() string {
	if v := strings.TrimSpace(os.Getenv(envPlatformDomain)); v != "" {
		n := NormalizePlatformDomain(v)
		if validPlatformDomain(n) {
			return n
		}
	}
	return DefaultPlatformDomain
}

// NormalizePlatformDomain trims, lowercases, and removes a trailing dot.
// It does not validate; use validPlatformDomain for that.
func NormalizePlatformDomain(s string) string {
	s = strings.TrimSpace(s)
	s = strings.ToLower(s)
	s = strings.TrimSuffix(s, ".")
	return s
}

func validPlatformDomain(s string) bool {
	if s == "" || len(s) > 253 {
		return false
	}
	if s[0] == '-' || s[0] == '.' || s[len(s)-1] == '-' || s[len(s)-1] == '.' {
		return false
	}
	for _, r := range s {
		switch {
		case r >= 'a' && r <= 'z':
		case r >= '0' && r <= '9':
		case r == '-' || r == '.':
		default:
			return false
		}
	}
	return true
}

// resetCacheForTest is exposed only to tests in the same package.
func resetCacheForTest() {
	defaultCache.mu.Lock()
	defaultCache.snapshot = Snapshot{}
	defaultCache.loaded = false
	defaultCache.expiresAt = time.Time{}
	defaultCache.mu.Unlock()
}
