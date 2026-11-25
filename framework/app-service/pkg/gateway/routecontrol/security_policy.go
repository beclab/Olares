package routecontrol

import (
	"context"
	"fmt"
	"reflect"
	"strings"

	apierrors "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"k8s.io/apimachinery/pkg/types"
	"sigs.k8s.io/controller-runtime/pkg/client"

	"github.com/beclab/Olares/framework/app-service/pkg/gateway"
	"github.com/beclab/Olares/framework/app-service/pkg/gateway/callerjwt"
	srrv1alpha1 "github.com/beclab/Olares/framework/app-service/pkg/gateway/v1alpha1"
)

const (
	SecurityPolicySuffix       = "-jwt-authn"
	ApplicationExtAuthSuffix   = "-extauth"
	AutheliaBackendServiceName = "authelia-backend"
	AutheliaVerifyPath         = "/api/verify"
	AutheliaBackendPort        = int32(9091)
	CallerJWTIssuer            = callerjwt.IssuerURL
	CallerJWTAudience          = "app-gateway-data"
	CallerJWTProviderName      = "caller-jwt"
	// Claim paths must match callerjwt nested JSON (Envoy treats '.' as path sep).
	CallerJWTViewerClaim      = callerjwt.ClaimViewer
	CallerJWTAppidClaim       = callerjwt.ClaimAppid
	CallerJWTClientAppidClaim = callerjwt.ClaimClientAppid
	// CallerJWTViewerHeader is filled from the verified JWT viewer claim after
	// gateway authn. A client-supplied X-CALLER-USER is stripped before authn
	// and must not be trusted as identity on its own; a caller's own
	// X-BFL-USER is passed through untouched.
	CallerJWTViewerHeader = "X-CALLER-USER"
	// CallerJWTAppidHeader and CallerJWTClientAppidHeader both map to
	// X-Caller-Appid: Shared tokens carry olares.caller.appid, ordinary
	// callers carry olares.caller.clientAppid; the claims stay mutually
	// exclusive so only one value is written per request.
	CallerJWTAppidHeader          = "X-Caller-Appid"
	CallerJWTClientAppidHeader    = "X-Caller-Appid"
	CallerJWTJWKSServiceName      = "caller-jwt-jwks"
	CallerJWTJWKSServiceNamespace = "os-framework"
	CallerJWTJWKSServicePort      = int32(443)
	CallerJWTJWKSURI              = callerjwt.JWKSURI
)

var securityPolicyGVK = schema.GroupVersionKind{
	Group:   "gateway.envoyproxy.io",
	Version: "v1alpha1",
	Kind:    "SecurityPolicy",
}

// securityPolicyName returns the SecurityPolicy object name for an SRR.
func securityPolicyName(srr *srrv1alpha1.SharedRouteRegistry) string {
	if isApplicationSRR(srr) {
		return srr.Name + ApplicationExtAuthSuffix
	}
	return srr.Name + SecurityPolicySuffix
}

func isApplicationSRR(srr *srrv1alpha1.SharedRouteRegistry) bool {
	return srr != nil && srr.Spec.EntranceClass == srrv1alpha1.EntranceClassApplication
}

// desiredSharedRouteSecurityPolicy builds JWT authn for a Shared HTTPRoute.
// After verification, the gateway copies olares.viewer → X-CALLER-USER and
// either olares.caller.appid or olares.caller.clientAppid → X-Caller-Appid
// (Shared vs ordinary callers; claims are mutually exclusive). A header is
// written only when that claim exists in the token; missing JWT is rejected.
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
									"name": callerjwt.CallerJWTHeaderName,
								},
							},
						},
						"claimToHeaders": []any{
							map[string]any{
								"claim":  CallerJWTViewerClaim,
								"header": CallerJWTViewerHeader,
							},
							map[string]any{
								"claim":  CallerJWTAppidClaim,
								"header": CallerJWTAppidHeader,
							},
							map[string]any{
								"claim":  CallerJWTClientAppidClaim,
								"header": CallerJWTClientAppidHeader,
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

func srrAuthMode(srr *srrv1alpha1.SharedRouteRegistry) string {
	if srr == nil {
		return ""
	}
	if srr.Annotations != nil {
		if m := strings.TrimSpace(srr.Annotations[gateway.AnnotationAuthMode]); m != "" {
			return m
		}
	}
	if isApplicationSRR(srr) {
		return gateway.AuthModeExtAuth
	}
	return ""
}

func applySecurityPolicy(ctx context.Context, c client.Client, srr *srrv1alpha1.SharedRouteRegistry) error {
	if srr == nil {
		return fmt.Errorf("srr is nil")
	}
	// Drop the alternate policy name if entrance class changed.
	if isApplicationSRR(srr) {
		_ = deleteSecurityPolicyByName(ctx, c, srr.Namespace, srr.Name+SecurityPolicySuffix)
	} else {
		_ = deleteSecurityPolicyByName(ctx, c, srr.Namespace, srr.Name+ApplicationExtAuthSuffix)
	}

	if isApplicationSRR(srr) && srrAuthMode(srr) == gateway.AuthModeNone {
		return deleteSecurityPolicyByName(ctx, c, srr.Namespace, srr.Name+ApplicationExtAuthSuffix)
	}

	var desired *unstructured.Unstructured
	if isApplicationSRR(srr) {
		desired = desiredApplicationExtAuthSecurityPolicy(srr)
		if desired == nil {
			return fmt.Errorf("application SRR %s/%s missing %s for Authelia ExtAuth",
				srr.Namespace, srr.Name, gateway.AnnotationExtAuthUser)
		}
	} else {
		desired = desiredSharedRouteSecurityPolicy(srr)
	}
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

// desiredApplicationExtAuthSecurityPolicy checks the canonical Authelia endpoint.
func desiredApplicationExtAuthSecurityPolicy(srr *srrv1alpha1.SharedRouteRegistry) *unstructured.Unstructured {
	if srr == nil {
		return nil
	}
	user := ""
	if srr.Annotations != nil {
		user = strings.TrimSpace(srr.Annotations[gateway.AnnotationExtAuthUser])
	}
	if user == "" {
		return nil
	}
	routeName := httpRouteName(srr)
	autheliaNS := "user-system-" + user
	authHTTP := map[string]any{
		"backendRefs": []any{
			map[string]any{
				"group":     "",
				"kind":      "Service",
				"name":      AutheliaBackendServiceName,
				"namespace": autheliaNS,
				"port":      int64(AutheliaBackendPort),
				// Match API server defaulting so the live policy stays stable.
				"weight": int64(1),
			},
		},
	}
	headers := []any{"authorization", "cookie", "x-forwarded-for", "x-authorization", "x-bfl-user"}
	if isSystemDesktopRoute(srr) {
		// Unlike a path prefix, the override does not append the Desktop path.
		// This route only accepts Entry mTLS. L4 supplies the original request
		// metadata after authenticating the browser.
		authHTTP["pathOverride"] = AutheliaVerifyPath
		headers = append(headers, "accept", "x-original-url", "x-original-method")
	} else {
		authHTTP["path"] = AutheliaVerifyPath
	}
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
			"extAuth": map[string]any{
				"failOpen":         false,
				"headersToExtAuth": headers,
				"http":             authHTTP,
			},
		},
	}}
}

func deleteSecurityPolicyByName(ctx context.Context, c client.Client, ns, name string) error {
	obj := &unstructured.Unstructured{}
	obj.SetGroupVersionKind(securityPolicyGVK)
	obj.SetName(name)
	obj.SetNamespace(ns)
	if err := c.Delete(ctx, obj); err != nil && !apierrors.IsNotFound(err) {
		return err
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
	if srr == nil {
		return nil
	}
	if err := deleteSecurityPolicyByName(ctx, c, srr.Namespace, srr.Name+SecurityPolicySuffix); err != nil {
		return err
	}
	return deleteSecurityPolicyByName(ctx, c, srr.Namespace, srr.Name+ApplicationExtAuthSuffix)
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

// ApplicationExtAuthSecurityPolicySpecForTest exposes ExtAuth policy for tests.
func ApplicationExtAuthSecurityPolicySpecForTest(srr *srrv1alpha1.SharedRouteRegistry) map[string]any {
	obj := desiredApplicationExtAuthSecurityPolicy(srr)
	if obj == nil {
		return nil
	}
	spec, _ := obj.Object["spec"].(map[string]any)
	return spec
}

// ApplicationExtAuthSecurityPolicyMetaForTest exposes ExtAuth metadata for tests.
func ApplicationExtAuthSecurityPolicyMetaForTest(srr *srrv1alpha1.SharedRouteRegistry) metav1.ObjectMeta {
	obj := desiredApplicationExtAuthSecurityPolicy(srr)
	if obj == nil {
		return metav1.ObjectMeta{}
	}
	return metav1.ObjectMeta{
		Name:      obj.GetName(),
		Namespace: obj.GetNamespace(),
		Labels:    obj.GetLabels(),
	}
}
