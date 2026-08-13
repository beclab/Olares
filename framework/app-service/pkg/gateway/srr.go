package gateway

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"strings"

	"github.com/beclab/Olares/framework/app-service/pkg/appcfg"
	appv1alpha1 "github.com/beclab/api/api/app.bytetrade.io/v1alpha1"
	corev1 "k8s.io/api/core/v1"
	apierrors "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/types"
	"k8s.io/klog/v2"
	"sigs.k8s.io/controller-runtime/pkg/client"

	srrv1alpha1 "github.com/beclab/Olares/framework/app-service/pkg/gateway/v1alpha1"
)

// AnnotationRouteMode is the per-Application opt-in: app-service emits a SRR
// only when the value equals AnnotationRouteModeGateway.
const (
	AnnotationRouteMode        = "gateway.olares.io/route-mode"
	AnnotationRouteModeGateway = "gateway"
)

// ResourceName derives the legacy SRR name from the Application short name.
// Kept to delete legacy SRRs named "shared-<appName>".
func ResourceName(appName string) string {
	return fmt.Sprintf("shared-%s", appName)
}

// ResourceNameForEntrance names one SRR per sharedEntrance, identified by
// appid + entranceName so the name survives Application rename / Helm upgrades.
// Returns "" if either input is empty so the caller can fail closed.
func ResourceNameForEntrance(appid, entranceName string) string {
	appid = strings.ToLower(strings.TrimSpace(appid))
	entranceName = strings.ToLower(strings.TrimSpace(entranceName))
	if appid == "" || entranceName == "" {
		return ""
	}
	return fmt.Sprintf("shared-%s-%s", appid, entranceName)
}

// ResourceNameForEntranceApp names one SRR per application entrance.
// Returns "" if either input is empty so the caller can fail closed.
func ResourceNameForEntranceApp(appid, entranceName string) string {
	appid = strings.ToLower(strings.TrimSpace(appid))
	entranceName = strings.ToLower(strings.TrimSpace(entranceName))
	if appid == "" || entranceName == "" {
		return ""
	}
	return fmt.Sprintf("app-%s-%s", appid, entranceName)
}

// IsOptedIn reports whether the Application carries the gateway opt-in
// annotation. Callers must also check appcfg.IsGatewaySharedApp(app).
func IsOptedIn(app *appv1alpha1.Application) bool {
	if app == nil || app.Annotations == nil {
		return false
	}
	return app.Annotations[AnnotationRouteMode] == AnnotationRouteModeGateway
}

// BuildSpecForEntrance projects one sharedEntrance into a SharedRouteRegistrySpec
// carrying the shared hostPattern (<hash8>.shared.<platformDomain>). The caller
// resolves the backing Service so this helper stays I/O-free.
func BuildSpecForEntrance(app *appv1alpha1.Application, entrance appv1alpha1.Entrance,
	entranceIndex, entranceCount int, svc *corev1.Service, platformDomain string,
	entranceClass srrv1alpha1.EntranceClass, materializer CertMaterializer) (srrv1alpha1.SharedRouteRegistrySpec, error) {
	if app == nil {
		return srrv1alpha1.SharedRouteRegistrySpec{}, errors.New("application is nil")
	}
	if entrance.Name == "" {
		return srrv1alpha1.SharedRouteRegistrySpec{}, errors.New("entrance has empty name")
	}
	if svc == nil {
		return srrv1alpha1.SharedRouteRegistrySpec{}, errors.New("upstream service is nil")
	}
	if entranceIndex < 0 || entranceIndex >= entranceCount {
		return srrv1alpha1.SharedRouteRegistrySpec{}, fmt.Errorf("entrance index out of range: index=%d count=%d",
			entranceIndex, entranceCount)
	}
	appid := strings.ToLower(strings.TrimSpace(app.Spec.Appid))
	if appid == "" {
		return srrv1alpha1.SharedRouteRegistrySpec{}, errors.New("application spec.appid is empty")
	}
	platformDomain = strings.ToLower(strings.TrimSpace(strings.TrimSuffix(platformDomain, ".")))
	if platformDomain == "" {
		return srrv1alpha1.SharedRouteRegistrySpec{}, errors.New("platformDomain is empty")
	}

	class := entranceClass
	if class == "" {
		class = srrv1alpha1.EntranceClassShared
	}

	var patterns []string
	switch class {
	case srrv1alpha1.EntranceClassApplication:
		// Primary = Resolve(defaultThirdLevel|EntranceID); when default
		// overrides, also keep the hash EntranceID so both friendly and
		// canonical hosts hit app-gateway (CoreDNS + HTTPRoute).
		patterns = applicationEntranceHostPatterns(app, entranceIndex, platformDomain)
	default:
		var err error
		isShared := appv1alpha1.IsShared(app)
		pattern, err := appcfg.LogicalHostPattern(appid, entranceIndex, entranceCount, platformDomain, isShared)
		if err != nil {
			return srrv1alpha1.SharedRouteRegistrySpec{}, err
		}
		patterns = []string{pattern}
	}
	norms := make([]string, 0, len(patterns))
	seen := make(map[string]struct{}, len(patterns))
	for _, pattern := range patterns {
		norm, err := NormalizeHostOrLogicalPattern(pattern)
		if err != nil {
			return srrv1alpha1.SharedRouteRegistrySpec{}, fmt.Errorf("normalize host pattern %q: %w", pattern, err)
		}
		if _, dup := seen[norm]; dup {
			continue
		}
		seen[norm] = struct{}{}
		norms = append(norms, norm)
	}
	// Eligible third_party_domain + per-owner third_level exact Hosts (Scheme A).
	// Keep default/logical patterns above; append only gated / owned exact FQDNs.
	exactOwners := MergeExactHostOwners(
		CollectEligibleExactHosts(app, entrance.Name, platformDomain, materializer),
		CollectUserThirdLevelExactHosts(app, entrance.Name, platformDomain),
	)
	for _, eh := range exactOwners {
		norm, err := NormalizeHostOrLogicalPattern(eh.Host)
		if err != nil {
			klog.Warningf("BuildSpecForEntrance: skip exact host app=%s entrance=%s: %v",
				app.Spec.Name, entrance.Name, err)
			continue
		}
		if _, dup := seen[norm]; dup {
			continue
		}
		seen[norm] = struct{}{}
		norms = append(norms, norm)
	}
	if len(norms) == 0 {
		return srrv1alpha1.SharedRouteRegistrySpec{}, errors.New("no usable host patterns")
	}

	upstream := srrv1alpha1.UpstreamRef{
		ServiceName:      svc.Name,
		ServiceNamespace: svc.Namespace,
	}
	if port := pickHTTPPort(svc, entrance.Port); port > 0 {
		upstream.Port = port
	} else {
		return srrv1alpha1.SharedRouteRegistrySpec{}, fmt.Errorf("service %s/%s has no usable TCP port", svc.Namespace, svc.Name)
	}

	return srrv1alpha1.SharedRouteRegistrySpec{
		RouteMode:     srrv1alpha1.RouteModeGateway,
		EntranceClass: class,
		HostPatterns:  norms,
		Upstream:      upstream,
	}, nil
}

// applicationEntranceHostPatterns returns SRR hostPatterns for an application
// entrance. Order: resolved primary (defaultThirdLevel or EntranceID), then
// the bare EntranceID when an override is active so hash hosts stay on
// app-gateway alongside the friendly name.
func applicationEntranceHostPatterns(app *appv1alpha1.Application, entranceIndex int, platformDomain string) []string {
	appid := strings.ToLower(strings.TrimSpace(app.Spec.Appid))
	entrances := app.Spec.Entrances
	canonical := appv1alpha1.EntranceID(appid, entranceIndex, len(entrances))
	primary := applicationEntranceHostPrefix(app, entranceIndex)

	out := []string{fmt.Sprintf("%s.*.%s", primary, platformDomain)}
	if primary != canonical && canonical != "" {
		out = append(out, fmt.Sprintf("%s.*.%s", canonical, platformDomain))
	}
	return out
}

// applicationEntranceHostPrefix returns the Host label used for an application
// entrance SRR primary pattern. It mirrors l4 buildAppVirtualHosts: parse
// defaultThirdLevelDomainConfig and call the shared api helper; on missing or
// malformed settings it falls back to EntranceID.
func applicationEntranceHostPrefix(app *appv1alpha1.Application, entranceIndex int) string {
	appid := strings.ToLower(strings.TrimSpace(app.Spec.Appid))
	entrances := app.Spec.Entrances
	if entranceIndex < 0 || entranceIndex >= len(entrances) {
		return appv1alpha1.EntranceID(appid, entranceIndex, len(entrances))
	}

	var cfgs []appv1alpha1.DefaultThirdLevelDomainConfig
	if raw := app.Spec.Settings["defaultThirdLevelDomainConfig"]; raw != "" {
		if err := json.Unmarshal([]byte(raw), &cfgs); err != nil {
			klog.Warningf("applicationEntranceHostPrefix: defaultThirdLevelDomainConfig unmarshal error app=%s: %v",
				app.Spec.Name, err)
		}
	}

	ptrs := make([]*appv1alpha1.Entrance, len(entrances))
	for i := range entrances {
		ptrs[i] = &entrances[i]
	}
	return appv1alpha1.ResolveEntranceIDWithDefaultThirdLevelDomainOverride(ptrs, entranceIndex, appid, cfgs)
}

// pickHTTPPort prefers the entrance-declared port; otherwise the first TCP port
// on the service. Returns 0 when nothing matches.
func pickHTTPPort(svc *corev1.Service, preferred int32) int32 {
	if svc == nil {
		return 0
	}
	for _, p := range svc.Spec.Ports {
		if p.Protocol != "" && p.Protocol != corev1.ProtocolTCP {
			continue
		}
		if preferred > 0 && p.Port == preferred {
			return p.Port
		}
	}
	for _, p := range svc.Spec.Ports {
		if p.Protocol == "" || p.Protocol == corev1.ProtocolTCP {
			return p.Port
		}
	}
	if preferred > 0 {
		return preferred
	}
	return 0
}

// ReconcileForEntrance creates or updates the per-entrance SRR in the
// Application's workload namespace, owned by the Application for GC.
// platformDomain is required to recompute exact-host ownership for the
// gateway.olares.io/exact-host-owners annotation.
func ReconcileForEntrance(ctx context.Context, c client.Client, app *appv1alpha1.Application,
	entrance appv1alpha1.Entrance, spec srrv1alpha1.SharedRouteRegistrySpec,
	platformDomain string, materializer CertMaterializer) (*srrv1alpha1.SharedRouteRegistry, error) {
	owners := MergeExactHostOwners(
		CollectEligibleExactHosts(app, entrance.Name, platformDomain, materializer),
		CollectUserThirdLevelExactHosts(app, entrance.Name, platformDomain),
	)
	return reconcileForEntranceWithOwners(ctx, c, app, entrance, spec, owners)
}

func reconcileForEntranceWithOwners(ctx context.Context, c client.Client, app *appv1alpha1.Application,
	entrance appv1alpha1.Entrance, spec srrv1alpha1.SharedRouteRegistrySpec,
	owners []ExactHostOwner) (*srrv1alpha1.SharedRouteRegistry, error) {
	if app == nil {
		return nil, errors.New("application is nil")
	}
	ns := app.Spec.Namespace
	if ns == "" {
		return nil, errors.New("application has empty spec.namespace")
	}
	appid := strings.ToLower(strings.TrimSpace(app.Spec.Appid))
	name := resourceNameForClass(appid, entrance.Name, spec.EntranceClass)
	if name == "" {
		return nil, fmt.Errorf("compute SRR name for entrance %q on app %s", entrance.Name, app.Spec.Name)
	}
	ownersJSON := EncodeExactHostOwnersJSON(owners)

	got := &srrv1alpha1.SharedRouteRegistry{}
	getErr := c.Get(ctx, types.NamespacedName{Namespace: ns, Name: name}, got)
	switch {
	case apierrors.IsNotFound(getErr):
		anns := map[string]string{}
		if ownersJSON != "" {
			anns[AnnotationExactHostOwners] = ownersJSON
		}
		obj := &srrv1alpha1.SharedRouteRegistry{
			ObjectMeta: metav1.ObjectMeta{
				Name:      name,
				Namespace: ns,
				Labels: map[string]string{
					"app.kubernetes.io/managed-by": "app-service",
					"app.kubernetes.io/instance":   app.Spec.Name,
					"gateway.olares.io/appid":      appid,
					"gateway.olares.io/entrance":   entrance.Name,
				},
				Annotations:     anns,
				OwnerReferences: ownerRefs(app),
			},
			Spec: spec,
		}
		if err := c.Create(ctx, obj); err != nil {
			return nil, fmt.Errorf("create SRR %s/%s: %w", ns, name, err)
		}
		return obj, nil
	case getErr != nil:
		return nil, fmt.Errorf("get SRR %s/%s: %w", ns, name, getErr)
	}

	patched := got.DeepCopy()
	patched.Spec = spec
	if !ownerRefAlreadyPresent(patched.OwnerReferences, app.UID) {
		patched.OwnerReferences = append(patched.OwnerReferences, ownerRefs(app)...)
	}
	if patched.Labels == nil {
		patched.Labels = map[string]string{}
	}
	patched.Labels["app.kubernetes.io/managed-by"] = "app-service"
	patched.Labels["app.kubernetes.io/instance"] = app.Spec.Name
	patched.Labels["gateway.olares.io/appid"] = appid
	patched.Labels["gateway.olares.io/entrance"] = entrance.Name
	if patched.Annotations == nil {
		patched.Annotations = map[string]string{}
	}
	if ownersJSON == "" {
		delete(patched.Annotations, AnnotationExactHostOwners)
	} else {
		patched.Annotations[AnnotationExactHostOwners] = ownersJSON
	}

	if err := c.Patch(ctx, patched, client.MergeFrom(got)); err != nil {
		return nil, fmt.Errorf("patch SRR %s/%s: %w", ns, name, err)
	}
	return patched, nil
}

// Delete removes the legacy "shared-<appName>" SRR. Missing objects are ignored.
func Delete(ctx context.Context, c client.Client, app *appv1alpha1.Application) error {
	if app == nil || app.Spec.Namespace == "" {
		return nil
	}
	obj := &srrv1alpha1.SharedRouteRegistry{
		ObjectMeta: metav1.ObjectMeta{
			Namespace: app.Spec.Namespace,
			Name:      ResourceName(app.Spec.Name),
		},
	}
	if err := c.Delete(ctx, obj); err != nil && !apierrors.IsNotFound(err) {
		return fmt.Errorf("delete SRR %s/%s: %w", obj.Namespace, obj.Name, err)
	}
	return nil
}

// PruneEntranceSRRs removes any SRR labeled app.kubernetes.io/instance=<app>
// whose name is not in keep (e.g. after sharedEntrances was trimmed).
func PruneEntranceSRRs(ctx context.Context, c client.Client, app *appv1alpha1.Application, keep map[string]struct{}) error {
	if app == nil || app.Spec.Namespace == "" {
		return nil
	}
	list := &srrv1alpha1.SharedRouteRegistryList{}
	if err := c.List(ctx, list,
		client.InNamespace(app.Spec.Namespace),
		client.MatchingLabels{"app.kubernetes.io/instance": app.Spec.Name},
	); err != nil {
		return fmt.Errorf("list SRRs for prune: %w", err)
	}
	legacyName := ResourceName(app.Spec.Name)
	for i := range list.Items {
		obj := &list.Items[i]
		if obj.Name == legacyName {
			continue
		}
		if _, ok := keep[obj.Name]; ok {
			continue
		}
		if err := c.Delete(ctx, obj); err != nil && !apierrors.IsNotFound(err) {
			return fmt.Errorf("prune SRR %s/%s: %w", obj.Namespace, obj.Name, err)
		}
	}
	return nil
}

// DeleteAllForApp removes every SRR owned by the Application (legacy and
// per-entrance forms). Used when the app is uninstalled or opts out.
func DeleteAllForApp(ctx context.Context, c client.Client, app *appv1alpha1.Application) error {
	if app == nil || app.Spec.Namespace == "" {
		return nil
	}
	if err := Delete(ctx, c, app); err != nil {
		return err
	}
	list := &srrv1alpha1.SharedRouteRegistryList{}
	if err := c.List(ctx, list, client.InNamespace(app.Spec.Namespace), client.MatchingLabels{
		"app.kubernetes.io/instance": app.Spec.Name,
	}); err != nil {
		return fmt.Errorf("list SRRs for app %s: %w", app.Spec.Name, err)
	}
	for i := range list.Items {
		obj := &list.Items[i]
		if err := c.Delete(ctx, obj); err != nil && !apierrors.IsNotFound(err) {
			return fmt.Errorf("delete SRR %s/%s: %w", obj.Namespace, obj.Name, err)
		}
	}
	return nil
}

func ownerRefs(app *appv1alpha1.Application) []metav1.OwnerReference {
	if app == nil || app.UID == "" {
		return nil
	}
	gvk := app.GroupVersionKind()
	if gvk.Kind == "" {
		gvk.Group = "app.bytetrade.io"
		gvk.Version = "v1alpha1"
		gvk.Kind = "Application"
	}
	t := true
	return []metav1.OwnerReference{{
		APIVersion:         gvk.GroupVersion().String(),
		Kind:               gvk.Kind,
		Name:               app.Name,
		UID:                app.UID,
		BlockOwnerDeletion: &t,
		Controller:         &t,
	}}
}

func ownerRefAlreadyPresent(refs []metav1.OwnerReference, uid types.UID) bool {
	for _, r := range refs {
		if r.UID == uid {
			return true
		}
	}
	return false
}

func resourceNameForClass(appid, entranceName string, entranceClass srrv1alpha1.EntranceClass) string {
	if entranceClass == srrv1alpha1.EntranceClassApplication {
		return ResourceNameForEntranceApp(appid, entranceName)
	}
	return ResourceNameForEntrance(appid, entranceName)
}
