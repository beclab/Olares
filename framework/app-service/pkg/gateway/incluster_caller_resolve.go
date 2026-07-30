package gateway

import (
	"context"
	"strings"

	"github.com/beclab/Olares/framework/app-service/pkg/appcfg"
	appv1alpha1 "github.com/beclab/api/api/app.bytetrade.io/v1alpha1"
	"k8s.io/klog/v2"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

// ClusterAppOwnerIndex maps cluster-scoped Application Spec.Name to owner viewer.
// Type alias keeps callers in pkg/gateway/routecontrol byte-equal compatible.
type ClusterAppOwnerIndex = map[string]string

// Test seams: WI-T1-2 behavior-preserving extraction preserves the original
// routecontrol-internal hooks as exported package-level vars so cross-package
// tests (TC-T1-2 characterization + TC-402d call-count regression) keep working.
var (
	TestBuildClusterAppOwnerIndexHook func()
	TestResolveClusterAppOwnerHook    func()
)

// NamespaceOptedIntoGateway reports whether ns hosts an Application that opts
// into cluster-internal gateway routing with a non-empty clusterAppRef.
//
// requirement: 详设 §2.2 抽取契约 (WI-T1-2). Extracted byte-equal from the
// former (*CallerReconciler).namespaceOptedIntoGateway in routecontrol.
func NamespaceOptedIntoGateway(ctx context.Context, c client.Client, ns string) (bool, error) {
	var list appv1alpha1.ApplicationList
	if err := c.List(ctx, &list); err != nil {
		return false, err
	}
	for i := range list.Items {
		app := &list.Items[i]
		if app.Spec.Namespace != ns {
			continue
		}
		if !strings.EqualFold(strings.TrimSpace(app.Annotations[AnnotationInCluster]), InClusterGateway) {
			continue
		}
		if strings.TrimSpace(app.Spec.Settings["clusterAppRef"]) != "" {
			return true, nil
		}
	}
	return false, nil
}

// BuildClusterAppOwnerIndex builds <calleeAppName> -> <ownerViewer> from the
// cluster-scoped Application list. Multiple owners for the same name are
// joined as comma-separated values (warn-only, never fail) so the demand
// reconciler can fan out replicas per owner.
//
// requirement: 详设 §2.2 抽取契约 (WI-T1-2). Extracted byte-equal from the
// former buildClusterAppOwnerIndex in pkg/gateway/routecontrol.
func BuildClusterAppOwnerIndex(apps []appv1alpha1.Application) ClusterAppOwnerIndex {
	if TestBuildClusterAppOwnerIndexHook != nil {
		TestBuildClusterAppOwnerIndexHook()
	}
	idx := make(ClusterAppOwnerIndex, len(apps))
	for i := range apps {
		app := apps[i]
		if !appcfg.IsSharedServerApp(&app) {
			continue
		}
		name := strings.TrimSpace(app.Spec.Name)
		owner := strings.TrimSpace(app.Spec.Owner)
		if name == "" || owner == "" {
			continue
		}
		if prev, ok := idx[name]; ok && prev != owner {
			klog.Warningf("replica.app_ref_multi_owner app=%s owner_old=%s owner_new=%s", name, prev, owner)
			owners := SplitClusterAppRefs(prev)
			already := false
			for _, o := range owners {
				if o == owner {
					already = true
					break
				}
			}
			if !already {
				owners = append(owners, owner)
				idx[name] = strings.Join(owners, ",")
			}
			continue
		}
		idx[name] = owner
	}
	return idx
}

// ResolveClusterAppOwner looks up an owner viewer (possibly comma-joined for
// multi-owner names) by application ref. Returns empty string when idx is nil
// or the ref is unknown.
//
// requirement: 详设 §2.2 抽取契约 (WI-T1-2). Extracted byte-equal from the
// former resolveClusterAppOwner in pkg/gateway/routecontrol.
func ResolveClusterAppOwner(idx ClusterAppOwnerIndex, appRef string) string {
	if TestResolveClusterAppOwnerHook != nil {
		TestResolveClusterAppOwnerHook()
	}
	if len(idx) == 0 {
		return ""
	}
	return strings.TrimSpace(idx[strings.TrimSpace(appRef)])
}

// SplitClusterAppRefs splits a comma-separated clusterAppRef setting into
// trimmed, non-empty refs. Order is preserved; no deduplication or sorting is
// applied here -- the "primary ref" determinism contract (详设 §2.3.2) is
// delivered by WI-T1-5 alongside its actual use site.
//
// requirement: 详设 §2.2 抽取契约 (WI-T1-2). Extracted byte-equal from the
// former splitClusterAppRefs in pkg/gateway/routecontrol.
func SplitClusterAppRefs(raw string) []string {
	if strings.TrimSpace(raw) == "" {
		return nil
	}
	parts := strings.Split(raw, ",")
	refs := make([]string, 0, len(parts))
	for _, part := range parts {
		ref := strings.TrimSpace(part)
		if ref == "" {
			continue
		}
		refs = append(refs, ref)
	}
	return refs
}
