package routecontrol

import (
	"context"
	"fmt"
	"reflect"

	corev1 "k8s.io/api/core/v1"
	networkingv1 "k8s.io/api/networking/v1"
	apierrors "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"k8s.io/apimachinery/pkg/types"
	"sigs.k8s.io/controller-runtime/pkg/client"

	srrv1alpha1 "github.com/beclab/Olares/framework/app-service/pkg/gateway/v1alpha1"
)

const (
	ManagedByLabel         = "app.kubernetes.io/managed-by"
	ManagedByValue         = "app-service"
	InstanceLabel          = "app.kubernetes.io/instance"
	NetworkPolicyName      = "app-gateway-shared-ingress-np"
	defaultGatewayNS       = "os-gateway"
	defaultGatewayName     = "app-gateway"
	defaultGatewaySectN    = "https"
	sharedGatewaySectN     = "http"
	ConditionReady         = "Ready"
	ReasonGatewayMode      = "GatewayMode"
	ReasonDirectMode       = "DirectMode"
	ReasonReconciled       = "Reconciled"
	ReasonBackendMissing   = "BackendServiceMissing"
	ReasonInvalidSpec      = "InvalidSpec"
	ReasonRouteApplyFailed = "RouteApplyFailed"
)

// GatewayRef selects the parent Gateway for every HTTPRoute written by
// app-service route control. Zero value uses app-gateway/app-gateway/https.
type GatewayRef struct {
	GatewayNamespace string
	GatewayName      string
	GatewaySection   string
}

func (g GatewayRef) gatewayNamespace() string {
	if g.GatewayNamespace != "" {
		return g.GatewayNamespace
	}
	return defaultGatewayNS
}

func (g GatewayRef) gatewayName() string {
	if g.GatewayName != "" {
		return g.GatewayName
	}
	return defaultGatewayName
}

func (g GatewayRef) gatewaySection() string {
	if g.GatewaySection != "" {
		return g.GatewaySection
	}
	return defaultGatewaySectN
}

func gatewaySectionForSRR(gw GatewayRef, srr *srrv1alpha1.SharedRouteRegistry) string {
	if gw.GatewaySection != "" {
		return gw.GatewaySection
	}
	if srr == nil {
		return defaultGatewaySectN
	}
	switch srr.Spec.EntranceClass {
	case "", srrv1alpha1.EntranceClassShared:
		return sharedGatewaySectN
	default:
		return defaultGatewaySectN
	}
}

// ReconcileResult is written onto SRR.status after a reconcile pass.
type ReconcileResult struct {
	Status        metav1.ConditionStatus
	Reason        string
	Message       string
	HTTPRouteName string
}

// ReconcileSharedRoute applies the HTTPRoute (and companion NetworkPolicy +
// ReferenceGrant) for one SRR. No service mesh.
//
//   - routeMode=direct  -> remove route objects, Ready=True/DirectMode
//   - routeMode=gateway -> ensure HTTPRoute + NetworkPolicy, Ready per outcome
func ReconcileSharedRoute(ctx context.Context, c client.Client, gw GatewayRef, srr *srrv1alpha1.SharedRouteRegistry) (ReconcileResult, error) {
	if srr == nil {
		return ReconcileResult{}, fmt.Errorf("srr is nil")
	}
	switch srr.Spec.RouteMode {
	case srrv1alpha1.RouteModeDirect:
		if err := deleteHTTPRoute(ctx, c, srr); err != nil {
			return ReconcileResult{}, fmt.Errorf("delete HTTPRoute: %w", err)
		}
		if err := deleteNetworkPolicy(ctx, c, srr); err != nil {
			return ReconcileResult{}, fmt.Errorf("delete NetworkPolicy: %w", err)
		}
		if err := deleteReferenceGrant(ctx, c, srr); err != nil {
			return ReconcileResult{}, fmt.Errorf("delete ReferenceGrant: %w", err)
		}
		return ReconcileResult{
			Status:  metav1.ConditionTrue,
			Reason:  ReasonDirectMode,
			Message: "Direct mode: HTTPRoute and NetworkPolicy removed.",
		}, nil
	case "", srrv1alpha1.RouteModeGateway:
		return reconcileGatewayMode(ctx, c, gw, srr)
	default:
		return ReconcileResult{
			Status:  metav1.ConditionFalse,
			Reason:  ReasonInvalidSpec,
			Message: fmt.Sprintf("unknown routeMode %q", srr.Spec.RouteMode),
		}, nil
	}
}

func reconcileGatewayMode(ctx context.Context, c client.Client, gw GatewayRef, srr *srrv1alpha1.SharedRouteRegistry) (ReconcileResult, error) {
	if len(srr.Spec.HostPatterns) == 0 || srr.Spec.Upstream.ServiceName == "" {
		return ReconcileResult{
			Status:  metav1.ConditionFalse,
			Reason:  ReasonInvalidSpec,
			Message: "hostPatterns or upstream.serviceName missing",
		}, nil
	}

	upstreamNS := srr.Spec.Upstream.ServiceNamespace
	if upstreamNS == "" {
		upstreamNS = srr.Namespace
	}

	svc := &corev1.Service{}
	if err := c.Get(ctx, types.NamespacedName{Namespace: upstreamNS, Name: srr.Spec.Upstream.ServiceName}, svc); err != nil {
		if apierrors.IsNotFound(err) {
			return ReconcileResult{
				Status:  metav1.ConditionFalse,
				Reason:  ReasonBackendMissing,
				Message: fmt.Sprintf("backend Service %s/%s not found", upstreamNS, srr.Spec.Upstream.ServiceName),
			}, nil
		}
		return ReconcileResult{}, fmt.Errorf("get backend Service: %w", err)
	}

	port, err := resolveServicePort(svc, srr.Spec.Upstream)
	if err != nil {
		return ReconcileResult{
			Status:  metav1.ConditionFalse,
			Reason:  ReasonInvalidSpec,
			Message: err.Error(),
		}, nil
	}

	routeName, err := applyHTTPRoute(ctx, c, gw, srr, port)
	if err != nil {
		return ReconcileResult{}, fmt.Errorf("apply HTTPRoute: %w", err)
	}
	if err := applyReferenceGrant(ctx, c, srr, upstreamNS, srr.Spec.Upstream.ServiceName); err != nil {
		return ReconcileResult{}, fmt.Errorf("apply ReferenceGrant: %w", err)
	}
	if err := applyNetworkPolicy(ctx, c, gw, srr, svc, port); err != nil {
		return ReconcileResult{}, fmt.Errorf("apply NetworkPolicy: %w", err)
	}
	return ReconcileResult{
		Status:        metav1.ConditionTrue,
		Reason:        ReasonReconciled,
		Message:       "HTTPRoute and NetworkPolicy reconciled",
		HTTPRouteName: routeName,
	}, nil
}

func resolveServicePort(svc *corev1.Service, ref srrv1alpha1.UpstreamRef) (int32, error) {
	if ref.Port > 0 {
		for _, p := range svc.Spec.Ports {
			if p.Port == ref.Port {
				return p.Port, nil
			}
		}
		return 0, fmt.Errorf("service %s/%s has no port %d", svc.Namespace, svc.Name, ref.Port)
	}
	if ref.PortName != "" {
		for _, p := range svc.Spec.Ports {
			if p.Name == ref.PortName {
				return p.Port, nil
			}
		}
		return 0, fmt.Errorf("service %s/%s has no portName %q", svc.Namespace, svc.Name, ref.PortName)
	}
	for _, p := range svc.Spec.Ports {
		if p.Protocol == "" || p.Protocol == corev1.ProtocolTCP {
			return p.Port, nil
		}
	}
	return 0, fmt.Errorf("service %s/%s has no usable TCP port", svc.Namespace, svc.Name)
}

func applyHTTPRoute(ctx context.Context, c client.Client, gw GatewayRef, srr *srrv1alpha1.SharedRouteRegistry, port int32) (string, error) {
	name := httpRouteName(srr)
	hosts := HTTPRouteHostnames(srr.Spec.HostPatterns)
	headerMatches := HTTPRouteHostHeaderMatches(srr.Spec.HostPatterns)
	if len(hosts) == 0 {
		return "", fmt.Errorf("hostPatterns produced no usable hostnames: %v", srr.Spec.HostPatterns)
	}

	parentRef := map[string]any{
		"group":       "gateway.networking.k8s.io",
		"kind":        "Gateway",
		"namespace":   gw.gatewayNamespace(),
		"name":        gw.gatewayName(),
		"sectionName": gatewaySectionForSRR(gw, srr),
	}
	backendRef := map[string]any{
		"group":     "",
		"kind":      "Service",
		"name":      srr.Spec.Upstream.ServiceName,
		"namespace": srr.Namespace,
		"port":      int64(port),
		"weight":    int64(1),
	}
	if ns := srr.Spec.Upstream.ServiceNamespace; ns != "" && ns != srr.Namespace {
		backendRef["namespace"] = ns
	}

	matches := []any{}
	if len(headerMatches) == 0 {
		matches = append(matches, map[string]any{
			"path": map[string]any{"type": "PathPrefix", "value": "/"},
		})
	} else {
		for _, hm := range headerMatches {
			matches = append(matches, map[string]any{
				"path":    map[string]any{"type": "PathPrefix", "value": "/"},
				"headers": []any{hm},
			})
		}
	}
	rule := map[string]any{
		"matches":     matches,
		"backendRefs": []any{backendRef},
	}
	desired := &unstructured.Unstructured{Object: map[string]any{
		"apiVersion": "gateway.networking.k8s.io/v1",
		"kind":       "HTTPRoute",
		"metadata": map[string]any{
			"name":      name,
			"namespace": srr.Namespace,
			"labels": map[string]any{
				ManagedByLabel: ManagedByValue,
				InstanceLabel:  srr.Name,
			},
		},
		"spec": map[string]any{
			"parentRefs": []any{parentRef},
			"hostnames":  hosts,
			"rules":      []any{rule},
		},
	}}
	setOwnerSRR(desired, srr)

	current := &unstructured.Unstructured{}
	current.SetGroupVersionKind(schema.GroupVersionKind{Group: "gateway.networking.k8s.io", Version: "v1", Kind: "HTTPRoute"})
	err := c.Get(ctx, types.NamespacedName{Namespace: srr.Namespace, Name: name}, current)
	switch {
	case apierrors.IsNotFound(err):
		if err := c.Create(ctx, desired); err != nil {
			return "", err
		}
		return name, nil
	case err != nil:
		return "", err
	}
	if !reflect.DeepEqual(current.Object["spec"], desired.Object["spec"]) {
		current.Object["spec"] = desired.Object["spec"]
		labels := current.GetLabels()
		if labels == nil {
			labels = map[string]string{}
		}
		labels[ManagedByLabel] = ManagedByValue
		labels[InstanceLabel] = srr.Name
		current.SetLabels(labels)
		if err := c.Update(ctx, current); err != nil {
			return "", err
		}
	}
	return name, nil
}

func deleteHTTPRoute(ctx context.Context, c client.Client, srr *srrv1alpha1.SharedRouteRegistry) error {
	obj := &unstructured.Unstructured{}
	obj.SetGroupVersionKind(schema.GroupVersionKind{Group: "gateway.networking.k8s.io", Version: "v1", Kind: "HTTPRoute"})
	obj.SetName(httpRouteName(srr))
	obj.SetNamespace(srr.Namespace)
	if err := c.Delete(ctx, obj); err != nil && !apierrors.IsNotFound(err) {
		return err
	}
	return nil
}

// networkPolicyNamespace is the namespace where the ingress NetworkPolicy must
// live. It follows the backend Service (upstream).
func networkPolicyNamespace(srr *srrv1alpha1.SharedRouteRegistry, svc *corev1.Service) string {
	if svc != nil && svc.Namespace != "" {
		return svc.Namespace
	}
	if ns := srr.Spec.Upstream.ServiceNamespace; ns != "" {
		return ns
	}
	return srr.Namespace
}

// networkPolicyNamespacesToClean returns namespaces that may hold a gateway NP
// for this SRR (upstream plus the SRR namespace after cross-ns placement).
func networkPolicyNamespacesToClean(srr *srrv1alpha1.SharedRouteRegistry) []string {
	upstream := srr.Spec.Upstream.ServiceNamespace
	if upstream == "" {
		upstream = srr.Namespace
	}
	if upstream == srr.Namespace {
		return []string{upstream}
	}
	return []string{upstream, srr.Namespace}
}

// applyNetworkPolicy reconciles a single ingress NetworkPolicy per upstream
// namespace that admits app-gateway traffic into the shared workload namespace.
// PodSelector and Ports are intentionally empty so any pod/port in the upstream
// namespace is reachable from the gateway namespace.
func applyNetworkPolicy(ctx context.Context, c client.Client, gw GatewayRef, srr *srrv1alpha1.SharedRouteRegistry, svc *corev1.Service, port int32) error {
	npNS := networkPolicyNamespace(srr, svc)
	desired := &networkingv1.NetworkPolicy{
		ObjectMeta: metav1.ObjectMeta{
			Name:      NetworkPolicyName,
			Namespace: npNS,
			Labels: map[string]string{
				ManagedByLabel: ManagedByValue,
			},
		},
		Spec: networkingv1.NetworkPolicySpec{
			PolicyTypes: []networkingv1.PolicyType{networkingv1.PolicyTypeIngress},
			PodSelector: metav1.LabelSelector{},
			Ingress: []networkingv1.NetworkPolicyIngressRule{
				{
					From: []networkingv1.NetworkPolicyPeer{
						{
							NamespaceSelector: &metav1.LabelSelector{
								MatchLabels: map[string]string{
									"kubernetes.io/metadata.name": gw.gatewayNamespace(),
								},
							},
						},
					},
				},
			},
		},
	}

	current := &networkingv1.NetworkPolicy{}
	err := c.Get(ctx, types.NamespacedName{Namespace: npNS, Name: NetworkPolicyName}, current)
	switch {
	case apierrors.IsNotFound(err):
		return c.Create(ctx, desired)
	case err != nil:
		return err
	}
	current.Spec = desired.Spec
	if current.Labels == nil {
		current.Labels = map[string]string{}
	}
	current.Labels[ManagedByLabel] = ManagedByValue
	delete(current.Labels, InstanceLabel)
	current.SetOwnerReferences(nil)
	return c.Update(ctx, current)
}

// deleteNetworkPolicy removes the per-NS shared ingress NP only when no other
// gateway-mode SRR still targets the same upstream namespace.
func deleteNetworkPolicy(ctx context.Context, c client.Client, srr *srrv1alpha1.SharedRouteRegistry) error {
	for _, ns := range networkPolicyNamespacesToClean(srr) {
		stillNeeded, err := hasOtherGatewayModeSRRForUpstream(ctx, c, srr, ns)
		if err != nil {
			return err
		}
		if stillNeeded {
			continue
		}
		obj := &networkingv1.NetworkPolicy{
			ObjectMeta: metav1.ObjectMeta{
				Name:      NetworkPolicyName,
				Namespace: ns,
			},
		}
		if err := c.Delete(ctx, obj); err != nil && !apierrors.IsNotFound(err) {
			return err
		}
	}
	return nil
}

func hasOtherGatewayModeSRRForUpstream(ctx context.Context, c client.Client, self *srrv1alpha1.SharedRouteRegistry, upstreamNS string) (bool, error) {
	var list srrv1alpha1.SharedRouteRegistryList
	if err := c.List(ctx, &list); err != nil {
		return false, err
	}
	for i := range list.Items {
		other := &list.Items[i]
		if other.UID == self.UID {
			continue
		}
		switch other.Spec.RouteMode {
		case "", srrv1alpha1.RouteModeGateway:
		default:
			continue
		}
		otherNS := other.Spec.Upstream.ServiceNamespace
		if otherNS == "" {
			otherNS = other.Namespace
		}
		if otherNS == upstreamNS {
			return true, nil
		}
	}
	return false, nil
}

// UpdateSRRStatus writes ReconcileResult onto srr.status. Idempotent when
// status is unchanged.
func UpdateSRRStatus(ctx context.Context, c client.Client, srr *srrv1alpha1.SharedRouteRegistry, res ReconcileResult) error {
	if srr == nil {
		return fmt.Errorf("srr is nil")
	}
	if res.Status == "" {
		return nil
	}
	cond := metav1.Condition{
		Type:               ConditionReady,
		Status:             res.Status,
		Reason:             res.Reason,
		Message:            res.Message,
		ObservedGeneration: srr.Generation,
		LastTransitionTime: metav1.Now(),
	}
	patched := srr.DeepCopy()
	patched.Status.ObservedGeneration = srr.Generation
	patched.Status.HTTPRouteName = res.HTTPRouteName
	patched.Status.Conditions = upsertCondition(patched.Status.Conditions, cond)
	if statusEqual(srr.Status, patched.Status) {
		return nil
	}
	return c.Status().Patch(ctx, patched, client.MergeFrom(srr))
}

func upsertCondition(in []metav1.Condition, c metav1.Condition) []metav1.Condition {
	for i := range in {
		if in[i].Type == c.Type {
			if in[i].Status == c.Status && in[i].Reason == c.Reason && in[i].Message == c.Message {
				in[i].ObservedGeneration = c.ObservedGeneration
				return in
			}
			in[i] = c
			return in
		}
	}
	return append(in, c)
}

func statusEqual(a, b srrv1alpha1.SharedRouteRegistryStatus) bool {
	if a.HTTPRouteName != b.HTTPRouteName || a.ObservedGeneration != b.ObservedGeneration {
		return false
	}
	if len(a.Conditions) != len(b.Conditions) {
		return false
	}
	for i := range a.Conditions {
		ac, bc := a.Conditions[i], b.Conditions[i]
		if ac.Type != bc.Type || ac.Status != bc.Status || ac.Reason != bc.Reason || ac.Message != bc.Message {
			return false
		}
	}
	return true
}

func httpRouteName(srr *srrv1alpha1.SharedRouteRegistry) string {
	return srr.Name
}

func setOwnerSRR(obj client.Object, srr *srrv1alpha1.SharedRouteRegistry) {
	if obj.GetNamespace() != srr.GetNamespace() {
		return
	}
	t := true
	ref := metav1.OwnerReference{
		APIVersion:         srrv1alpha1.GroupVersion.String(),
		Kind:               "SharedRouteRegistry",
		Name:               srr.Name,
		UID:                srr.UID,
		Controller:         &t,
		BlockOwnerDeletion: &t,
	}
	refs := obj.GetOwnerReferences()
	for i, r := range refs {
		if r.UID == srr.UID {
			refs[i] = ref
			obj.SetOwnerReferences(refs)
			return
		}
	}
	obj.SetOwnerReferences(append(refs, ref))
}
