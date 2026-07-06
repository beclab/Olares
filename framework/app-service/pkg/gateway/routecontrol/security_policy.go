package routecontrol

import (
	"context"
	"fmt"
	"reflect"

	apierrors "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"k8s.io/apimachinery/pkg/types"
	"sigs.k8s.io/controller-runtime/pkg/client"

	srrv1alpha1 "github.com/beclab/Olares/framework/app-service/pkg/gateway/v1alpha1"
)

const (
	SecurityPolicySuffix          = "-jwt-authn"
	CallerJWTIssuer                 = "https://caller-jwt.olares.system/"
	CallerJWTAudience               = "app-gateway-data"
	CallerJWTProviderName           = "caller-jwt"
	CallerJWTViewerClaim            = "olares.viewer"
	CallerJWTViewerHeader           = "X-BFL-USER"
	CallerJWTJWKSServiceName        = "caller-jwt-jwks"
	CallerJWTJWKSServiceNamespace   = "os-app-service"
	CallerJWTJWKSServicePort        = int32(443)
	CallerJWTJWKSURI                  = "https://caller-jwt.olares.system/.well-known/jwks.json"
	AuthorizationHeaderName         = "Authorization"
	AuthorizationBearerValuePrefix  = "Bearer "
)

var securityPolicyGVK = schema.GroupVersionKind{
	Group:   "gateway.envoyproxy.io",
	Version: "v1alpha1",
	Kind:    "SecurityPolicy",
}

// securityPolicyName returns the SecurityPolicy object name for an SRR.
func securityPolicyName(srr *srrv1alpha1.SharedRouteRegistry) string {
	return srr.Name + SecurityPolicySuffix
}

// desiredSharedRouteSecurityPolicy builds the JWT authn SecurityPolicy for a
// gateway-mode Shared HTTPRoute (WI-OC-C2-01 L1-b).
func desiredSharedRouteSecurityPolicy(srr *srrv1alpha1.SharedRouteRegistry) *unstructured.Unstructured {
	routeName := httpRouteName(srr)
	return &unstructured.Unstructured{Object: map[string]any{
		"apiVersion": "gateway.envoyproxy.io/v1alpha1",
		"kind":       "SecurityPolicy",
		"metadata": map[string]any{
			"name":      securityPolicyName(srr),
			"namespace": srr.Namespace,
			"labels": map[string]any{
				ManagedByLabel: ManagedByValue,
				InstanceLabel:  srr.Name,
			},
		},
		"spec": map[string]any{
			"targetRef": map[string]any{
				"group": "gateway.networking.k8s.io",
				"kind":  "HTTPRoute",
				"name":  routeName,
			},
			"jwt": map[string]any{
				"providers": []any{
					map[string]any{
						"name":      CallerJWTProviderName,
						"issuer":    CallerJWTIssuer,
						"audiences": []any{CallerJWTAudience},
						"extractFrom": map[string]any{
							"headers": []any{
								map[string]any{
									"name":        AuthorizationHeaderName,
									"valuePrefix": AuthorizationBearerValuePrefix,
								},
							},
						},
						"claimToHeaders": []any{
							map[string]any{
								"claim":  CallerJWTViewerClaim,
								"header": CallerJWTViewerHeader,
							},
						},
						"remoteJWKS": map[string]any{
							"uri": CallerJWTJWKSURI,
							"backendRefs": []any{
								map[string]any{
									"group":     "",
									"kind":      "Service",
									"name":      CallerJWTJWKSServiceName,
									"namespace": CallerJWTJWKSServiceNamespace,
									"port":      int64(CallerJWTJWKSServicePort),
								},
							},
						},
					},
				},
			},
		},
	}}
}

func applySecurityPolicy(ctx context.Context, c client.Client, srr *srrv1alpha1.SharedRouteRegistry) error {
	if srr == nil {
		return fmt.Errorf("srr is nil")
	}
	desired := desiredSharedRouteSecurityPolicy(srr)
	current := &unstructured.Unstructured{}
	current.SetGroupVersionKind(securityPolicyGVK)
	key := types.NamespacedName{Namespace: srr.Namespace, Name: securityPolicyName(srr)}
	getErr := c.Get(ctx, key, current)
	setOwnerSRR(desired, srr)

	switch {
	case apierrors.IsNotFound(getErr):
		return c.Create(ctx, desired)
	case getErr != nil:
		return getErr
	}
	specChanged := !reflect.DeepEqual(current.Object["spec"], desired.Object["spec"])
	metaChanged := mergeSecurityPolicyMetadata(current, desired)
	if specChanged {
		current.Object["spec"] = desired.Object["spec"]
	}
	if specChanged || metaChanged {
		return c.Update(ctx, current)
	}
	return nil
}

func mergeSecurityPolicyMetadata(current, desired *unstructured.Unstructured) bool {
	changed := false
	labels := current.GetLabels()
	if labels == nil {
		labels = map[string]string{}
	}
	for _, k := range []string{ManagedByLabel, InstanceLabel} {
		if want := desired.GetLabels()[k]; want != "" && labels[k] != want {
			labels[k] = want
			changed = true
		}
	}
	current.SetLabels(labels)
	return changed
}

func deleteSecurityPolicy(ctx context.Context, c client.Client, srr *srrv1alpha1.SharedRouteRegistry) error {
	obj := &unstructured.Unstructured{}
	obj.SetGroupVersionKind(securityPolicyGVK)
	obj.SetName(securityPolicyName(srr))
	obj.SetNamespace(srr.Namespace)
	if err := c.Delete(ctx, obj); err != nil && !apierrors.IsNotFound(err) {
		return err
	}
	return nil
}

// SecurityPolicySpecForTest exposes the built spec for unit tests.
func SecurityPolicySpecForTest(srr *srrv1alpha1.SharedRouteRegistry) map[string]any {
	obj := desiredSharedRouteSecurityPolicy(srr)
	spec, _ := obj.Object["spec"].(map[string]any)
	return spec
}

// SecurityPolicyObjectMetaForTest exposes metadata for unit tests.
func SecurityPolicyObjectMetaForTest(srr *srrv1alpha1.SharedRouteRegistry) metav1.ObjectMeta {
	obj := desiredSharedRouteSecurityPolicy(srr)
	return metav1.ObjectMeta{
		Name:      obj.GetName(),
		Namespace: obj.GetNamespace(),
		Labels:    obj.GetLabels(),
	}
}
