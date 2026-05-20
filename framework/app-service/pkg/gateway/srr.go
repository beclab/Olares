package gateway

import (
	"context"
	"errors"
	"fmt"

	appv1alpha1 "github.com/beclab/api/api/app.bytetrade.io/v1alpha1"
	corev1 "k8s.io/api/core/v1"
	apierrors "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/types"
	"sigs.k8s.io/controller-runtime/pkg/client"

	srrv1alpha1 "github.com/beclab/Olares/framework/app-service/pkg/gateway/v1alpha1"
)

// AnnotationRouteMode is the per-Application opt-in (D-A10): only when the
// value equals AnnotationRouteModeGateway does app-service emit a SRR.
const (
	AnnotationRouteMode        = "gateway.olares.io/route-mode"
	AnnotationRouteModeGateway = "gateway"
)

// ResourceName derives the SRR name from the Application short name (D-A10).
// Using the same prefix as the rest of the v3 shared family makes the object
// easy to find with `kubectl get srr -A`.
func ResourceName(appName string) string {
	return fmt.Sprintf("shared-%s", appName)
}

// IsOptedIn reports whether the Application carries the gateway opt-in
// annotation. Returns false for non-v3 apps; the caller must additionally
// check appcfg.IsV3(app) before reconciling a SRR.
func IsOptedIn(app *appv1alpha1.Application) bool {
	if app == nil || app.Annotations == nil {
		return false
	}
	return app.Annotations[AnnotationRouteMode] == AnnotationRouteModeGateway
}

// BuildSpec projects a v3 Application + Service into a SharedRouteRegistrySpec.
// The caller resolves the backing Service (typically the first SharedEntrance's
// Host) so this helper stays free of cluster I/O for unit testing.
func BuildSpec(app *appv1alpha1.Application, svc *corev1.Service) (srrv1alpha1.SharedRouteRegistrySpec, error) {
	if app == nil {
		return srrv1alpha1.SharedRouteRegistrySpec{}, errors.New("application is nil")
	}
	if len(app.Spec.SharedEntrances) == 0 {
		return srrv1alpha1.SharedRouteRegistrySpec{}, errors.New("application has no shared entrances")
	}
	if svc == nil {
		return srrv1alpha1.SharedRouteRegistrySpec{}, errors.New("upstream service is nil")
	}

	hosts := make([]string, 0, len(app.Spec.SharedEntrances))
	for _, e := range app.Spec.SharedEntrances {
		switch {
		case e.URL != "":
			hosts = append(hosts, e.URL)
		case e.Host != "":
			hosts = append(hosts, e.Host)
		}
	}
	patterns, err := NormalizeHostPatterns(hosts)
	if err != nil {
		return srrv1alpha1.SharedRouteRegistrySpec{}, fmt.Errorf("normalize host patterns: %w", err)
	}
	if len(patterns) == 0 {
		return srrv1alpha1.SharedRouteRegistrySpec{}, errors.New("no usable host patterns")
	}

	upstream := srrv1alpha1.UpstreamRef{
		ServiceName:      svc.Name,
		ServiceNamespace: svc.Namespace,
	}
	if port := pickHTTPPort(svc, app.Spec.SharedEntrances[0].Port); port > 0 {
		upstream.Port = port
	} else {
		return srrv1alpha1.SharedRouteRegistrySpec{}, fmt.Errorf("service %s/%s has no usable TCP port", svc.Namespace, svc.Name)
	}

	return srrv1alpha1.SharedRouteRegistrySpec{
		RouteMode:    srrv1alpha1.RouteModeGateway,
		HostPatterns: patterns,
		Upstream:     upstream,
		AuthzRef: &srrv1alpha1.AuthzRef{
			DefaultAction: srrv1alpha1.AuthzDefaultAllow,
		},
	}, nil
}

// pickHTTPPort prefers the entrance-declared port; otherwise picks the first
// TCP port on the service. Returns 0 when nothing matches.
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

// Reconcile creates or updates a SharedRouteRegistry whose contents match
// spec, placing it in the same namespace as the v3 Application's workload
// (Application.spec.namespace = {app}-shared, F-2/F-4). The Application is
// recorded as the OwnerReference so garbage collection wipes the SRR on
// uninstall without an explicit finalizer.
func Reconcile(ctx context.Context, c client.Client, app *appv1alpha1.Application, spec srrv1alpha1.SharedRouteRegistrySpec) (*srrv1alpha1.SharedRouteRegistry, error) {
	if app == nil {
		return nil, errors.New("application is nil")
	}
	ns := app.Spec.Namespace
	if ns == "" {
		return nil, errors.New("application has empty spec.namespace")
	}
	name := ResourceName(app.Spec.Name)

	got := &srrv1alpha1.SharedRouteRegistry{}
	getErr := c.Get(ctx, types.NamespacedName{Namespace: ns, Name: name}, got)
	switch {
	case apierrors.IsNotFound(getErr):
		obj := &srrv1alpha1.SharedRouteRegistry{
			ObjectMeta: metav1.ObjectMeta{
				Name:      name,
				Namespace: ns,
				Labels: map[string]string{
					"app.kubernetes.io/managed-by": "app-service",
					"app.kubernetes.io/instance":   app.Spec.Name,
				},
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

	if err := c.Patch(ctx, patched, client.MergeFrom(got)); err != nil {
		return nil, fmt.Errorf("patch SRR %s/%s: %w", ns, name, err)
	}
	return patched, nil
}

// Delete removes the SRR associated with the Application, if any. Missing
// objects are not treated as an error.
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
