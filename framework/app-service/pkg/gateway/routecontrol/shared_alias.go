package routecontrol

import (
	"context"
	"fmt"
	"reflect"
	"strings"

	"github.com/beclab/Olares/framework/app-service/pkg/gateway"
	"github.com/beclab/Olares/framework/app-service/pkg/gateway/callerjwt"
	gatewaytimeout "github.com/beclab/Olares/framework/app-service/pkg/gateway/timeout"
	srrv1alpha1 "github.com/beclab/Olares/framework/app-service/pkg/gateway/v1alpha1"
	apierrors "k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/api/meta"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"k8s.io/apimachinery/pkg/util/validation"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

const (
	sharedAliasSuffix    = ".workload"
	sharedAliasCondition = "WorkloadAliasesReady"
)

func aliasObject(srr *srrv1alpha1.SharedRouteRegistry, gvk schema.GroupVersionKind) *unstructured.Unstructured {
	obj := &unstructured.Unstructured{}
	obj.SetGroupVersionKind(gvk)
	obj.SetName(srr.Name + sharedAliasSuffix)
	obj.SetNamespace(srr.Namespace)
	return obj
}

func reconcileSharedAlias(ctx context.Context, c client.Client, gw GatewayRef, srr *srrv1alpha1.SharedRouteRegistry, enabled bool) error {
	if (!isApplicationSRR(srr) || srr.Labels[InstanceLabel] == "") && meta.FindStatusCondition(srr.Status.Conditions, sharedAliasCondition) == nil {
		return nil
	}
	var prefix string
	var err error
	if enabled && isApplicationSRR(srr) {
		prefix, err = sharedAliasPrefix(ctx, c, srr)
		if err != nil {
			return err
		}
	}
	if prefix == "" {
		if err := withdrawSharedAlias(ctx, c, srr); err != nil {
			return err
		}
		return setSharedAliasReady(ctx, c, srr, false)
	}

	route := desiredSharedAlias(srr, gw, prefix)
	policy := desiredSharedAliasPolicy(srr)
	if err := applyJWKSReferenceGrantOnly(ctx, c, srr); err != nil {
		return err
	}
	currentPolicy, err := upsertSharedAliasObject(ctx, c, policy)
	if err != nil {
		return err
	}
	// Start with no backend until Envoy Gateway has accepted the JWT policy.
	// DNS alone is never the access control for a known Host.
	_, policyErr := acceptedGatewayPolicy(ctx, c, securityPolicyGVK, srr.Namespace, policy.GetName(), gw, sharedGatewaySectN)
	rule := route.Object["spec"].(map[string]any)["rules"].([]any)[0].(map[string]any)
	if currentPolicy == nil || policyErr != nil || !sharedAliasIdentityStripReady(ctx, c, gw) {
		delete(rule, "backendRefs")
	}
	current, err := upsertSharedAliasObject(ctx, c, route)
	if err != nil {
		return err
	}
	ready := rule["backendRefs"] != nil && current != nil && entryParentReadyForSection(current, gw, sharedGatewaySectN)
	return setSharedAliasReady(ctx, c, srr, ready)
}

func desiredSharedAliasPolicy(srr *srrv1alpha1.SharedRouteRegistry) *unstructured.Unstructured {
	policy := desiredSharedRouteSecurityPolicy(srr)
	policy.SetName(srr.Name + sharedAliasSuffix)
	spec := policy.Object["spec"].(map[string]any)
	spec["targetRef"].(map[string]any)["name"] = policy.GetName()
	// Declare API defaults so policy readback can stabilize after admission.
	provider := spec["jwt"].(map[string]any)["providers"].([]any)[0].(map[string]any)
	jwks := provider["remoteJWKS"].(map[string]any)
	jwks["cacheDuration"] = "300s"
	jwks["backendRefs"].([]any)[0].(map[string]any)["weight"] = int64(1)
	spec["authorization"] = map[string]any{
		"defaultAction": "Deny",
		"rules": []any{map[string]any{
			"name":   "tenant-and-target",
			"action": "Allow",
			"principal": map[string]any{"jwt": map[string]any{
				"provider": CallerJWTProviderName,
				"claims": []any{map[string]any{
					"name": callerjwt.ClaimSharedTargets, "valueType": "StringArray",
					"values": []any{srr.Annotations[gateway.AnnotationExtAuthUser] + "/" + srr.Labels[InstanceLabel], srr.Annotations[gateway.AnnotationExtAuthUser] + "/@all"},
				}},
			}},
		}},
	}
	setOwnerSRR(policy, srr)
	return policy
}

func sharedAliasIdentityStripReady(ctx context.Context, c client.Client, gw GatewayRef) bool {
	policy := &unstructured.Unstructured{}
	policy.SetGroupVersionKind(schema.GroupVersionKind{Group: "gateway.envoyproxy.io", Version: "v1alpha1", Kind: "ClientTrafficPolicy"})
	key := client.ObjectKey{Namespace: gw.gatewayNamespace(), Name: "app-gateway-strip-identity-headers"}
	if err := c.Get(ctx, key, policy); err != nil || !policy.GetDeletionTimestamp().IsZero() {
		return false
	}
	ref, _, _ := unstructured.NestedMap(policy.Object, "spec", "targetRef")
	remove, _, _ := unstructured.NestedStringSlice(policy.Object, "spec", "headers", "earlyRequestHeaders", "remove")
	if ref["kind"] != "Gateway" || ref["group"] != "gateway.networking.k8s.io" || ref["name"] != gw.gatewayName() || ref["sectionName"] != nil {
		return false
	}
	var appID, callerUser bool
	for _, header := range remove {
		appID = appID || strings.EqualFold(header, CallerJWTAppidHeader)
		callerUser = callerUser || strings.EqualFold(header, CallerJWTViewerHeader)
	}
	return appID && callerUser
}

// Only route the ordinary API through workload auth when its backend is
// already published through an authenticated Shared entrance. Router's UI
// has a separate Service, so its platform contract is limited to /v1.
func sharedAliasPrefix(ctx context.Context, c client.Client, srr *srrv1alpha1.SharedRouteRegistry) (string, error) {
	tenant, name := srr.Annotations[gateway.AnnotationExtAuthUser], srr.Labels[InstanceLabel]
	if srr.Labels[ManagedByLabel] != ManagedByValue || srr.Labels["gateway.olares.io/appid"] == "" ||
		len(validation.IsDNS1123Label(tenant)) != 0 || len(validation.IsDNS1123Label(name)) != 0 ||
		len(srr.Spec.HostPatterns) == 0 || srr.Spec.Upstream.Port <= 0 ||
		(srr.Spec.Upstream.ServiceNamespace != "" && srr.Spec.Upstream.ServiceNamespace != srr.Namespace) {
		return "", nil
	}
	for _, host := range srr.Spec.HostPatterns {
		if len(strings.Split(host, ".")) < 4 || len(validation.IsDNS1123Subdomain(host)) != 0 {
			return "", nil
		}
	}
	var routes srrv1alpha1.SharedRouteRegistryList
	if err := c.List(ctx, &routes, client.InNamespace(srr.Namespace), client.MatchingLabels{InstanceLabel: name}); err != nil {
		return "", err
	}
	for i := range routes.Items {
		shared := &routes.Items[i]
		if shared.Labels[ManagedByLabel] != ManagedByValue || shared.Labels["gateway.olares.io/appid"] != srr.Labels["gateway.olares.io/appid"] ||
			shared.Spec.EntranceClass != srrv1alpha1.EntranceClassShared || !shared.DeletionTimestamp.IsZero() ||
			!isGatewayModeSRR(shared) || !sameApplicationOwner(shared, srr) {
			continue
		}
		ready := meta.FindStatusCondition(shared.Status.Conditions, ConditionReady)
		if ready == nil || ready.Status != metav1.ConditionTrue || ready.ObservedGeneration != shared.Generation {
			continue
		}
		sharedUpstream, ordinaryUpstream := shared.Spec.Upstream, srr.Spec.Upstream
		if sharedUpstream.ServiceNamespace == "" {
			sharedUpstream.ServiceNamespace = shared.Namespace
		}
		if ordinaryUpstream.ServiceNamespace == "" {
			ordinaryUpstream.ServiceNamespace = srr.Namespace
		}
		if sharedUpstream == ordinaryUpstream {
			return "/", nil
		}
		if name == "router" && srr.Labels["gateway.olares.io/entrance"] == "router-frontend" &&
			ordinaryUpstream.ServiceName == "router-svc" && ordinaryUpstream.Port == 80 &&
			sharedUpstream.ServiceName == "sharedentrances-router" && sharedUpstream.Port == 80 && sharedUpstream.ServiceNamespace == srr.Namespace {
			return "/v1", nil
		}
	}
	return "", nil
}

func sameApplicationOwner(a, b *srrv1alpha1.SharedRouteRegistry) bool {
	left, right := metav1.GetControllerOf(a), metav1.GetControllerOf(b)
	return left != nil && right != nil && left.Kind == "Application" && right.Kind == "Application" && left.UID != "" && left.UID == right.UID
}

func desiredSharedAlias(srr *srrv1alpha1.SharedRouteRegistry, gw GatewayRef, prefix string) *unstructured.Unstructured {
	route := aliasObject(srr, schema.GroupVersionKind{Group: "gateway.networking.k8s.io", Version: "v1", Kind: "HTTPRoute"})
	route.SetLabels(map[string]string{ManagedByLabel: ManagedByValue, InstanceLabel: srr.Name})
	setOwnerSRR(route, srr)
	hostnames := make([]any, 0, len(srr.Spec.HostPatterns))
	for _, host := range srr.Spec.HostPatterns {
		hostnames = append(hostnames, host)
	}
	timeout, _ := gatewaytimeout.FormatSeconds(gatewaytimeout.DefaultTimeout())
	rule := map[string]any{
		"matches": []any{map[string]any{"path": map[string]any{"type": "PathPrefix", "value": prefix}}},
		"backendRefs": []any{map[string]any{
			"group": "", "kind": "Service", "name": srr.Spec.Upstream.ServiceName, "port": int64(srr.Spec.Upstream.Port), "weight": int64(1),
		}},
		"timeouts": map[string]any{"request": timeout, "backendRequest": timeout},
		"filters": []any{map[string]any{
			"type": "RequestHeaderModifier",
			"requestHeaderModifier": map[string]any{"remove": []any{
				callerjwt.CallerJWTHeaderName, "X-BFL-USER", "Remote-User", "Remote-Groups", "Remote-Name", "Remote-Email", "Remote-Access-Token",
			}},
		}},
	}
	route.Object["spec"] = map[string]any{
		"parentRefs": []any{map[string]any{
			"group": "gateway.networking.k8s.io", "kind": "Gateway", "name": gw.gatewayName(),
			"namespace": gw.gatewayNamespace(), "sectionName": sharedGatewaySectN,
		}},
		"hostnames": hostnames,
		"rules":     []any{rule},
	}
	return route
}

func upsertSharedAliasObject(ctx context.Context, c client.Client, desired *unstructured.Unstructured) (*unstructured.Unstructured, error) {
	current := &unstructured.Unstructured{}
	current.SetGroupVersionKind(desired.GroupVersionKind())
	err := c.Get(ctx, client.ObjectKeyFromObject(desired), current)
	switch {
	case apierrors.IsNotFound(err):
		return nil, c.Create(ctx, desired)
	case err != nil:
		return nil, err
	case !reflect.DeepEqual(current.GetOwnerReferences(), desired.GetOwnerReferences()):
		return nil, fmt.Errorf("unexpected owner of %s %s", desired.GetKind(), client.ObjectKeyFromObject(desired))
	case !unstructuredSpecEqual(current.Object["spec"], desired.Object["spec"]):
		current.Object["spec"] = desired.Object["spec"]
		return nil, c.Update(ctx, current)
	default:
		return current, nil
	}
}

func withdrawSharedAlias(ctx context.Context, c client.Client, srr *srrv1alpha1.SharedRouteRegistry) error {
	if meta.FindStatusCondition(srr.Status.Conditions, sharedAliasCondition) == nil && srr.Labels[ManagedByLabel] != ManagedByValue {
		return nil
	}
	// Keep the attached policy until owner GC so asynchronous route removal
	// cannot briefly leave the old data-plane route without its policy.
	obj := aliasObject(srr, schema.GroupVersionKind{Group: "gateway.networking.k8s.io", Version: "v1", Kind: "HTTPRoute"})
	if err := c.Get(ctx, client.ObjectKeyFromObject(obj), obj); err != nil {
		return client.IgnoreNotFound(err)
	}
	if owner := metav1.GetControllerOf(obj); owner != nil && owner.UID == srr.UID {
		uid := obj.GetUID()
		return client.IgnoreNotFound(c.Delete(ctx, obj, client.Preconditions{UID: &uid}))
	}
	return nil
}

func setSharedAliasReady(ctx context.Context, c client.Client, srr *srrv1alpha1.SharedRouteRegistry, ready bool) error {
	existing := meta.FindStatusCondition(srr.Status.Conditions, sharedAliasCondition)
	if !ready && existing == nil {
		return nil
	}
	patched := srr.DeepCopy()
	cond := metav1.Condition{
		Type: sharedAliasCondition, Status: metav1.ConditionFalse,
		Reason: "AwaitingWorkloadRoute", Message: "The authenticated workload route is not ready", ObservedGeneration: srr.Generation,
	}
	if ready {
		cond.Status, cond.Reason, cond.Message = metav1.ConditionTrue, "WorkloadRouteReady", "The authenticated workload route is ready"
	}
	if !meta.SetStatusCondition(&patched.Status.Conditions, cond) {
		return nil
	}
	if err := c.Status().Patch(ctx, patched, client.MergeFrom(srr)); err != nil {
		return err
	}
	srr.Status = patched.Status
	return nil
}
