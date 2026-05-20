// Package controller hosts the app-service-routecontrol reconciler that turns
// SharedRouteRegistry objects into HTTPRoute + NetworkPolicy in the same
// namespace as the backend (F-2, F-4).
package controller

import (
	"context"
	"fmt"

	"reflect"

	"github.com/go-logr/logr"
	corev1 "k8s.io/api/core/v1"
	networkingv1 "k8s.io/api/networking/v1"
	apierrors "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"k8s.io/apimachinery/pkg/types"
	"k8s.io/apimachinery/pkg/util/intstr"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/log"

	gwapi "github.com/beclab/Olares/framework/app-gateway/pkg/api/v1alpha1"
)

// ManagedByLabel marks resources created by the app-service-routecontrol so they can
// be identified during debugging and during cleanup of orphaned objects.
const (
	ManagedByLabel       = "app.kubernetes.io/managed-by"
	ManagedByValue       = "app-service-routecontrol"
	InstanceLabel        = "app.kubernetes.io/instance"
	NetworkPolicyName    = "app-gateway-shared-ingress-np"
	GatewayParentName    = "app-gateway"
	GatewayParentSection = "http"
	ConditionReady       = "Ready"
	ReasonGatewayMode    = "GatewayMode"
	ReasonDirectMode     = "DirectMode"
	ReasonReconciled     = "Reconciled"
	ReasonBackendMissing = "BackendServiceMissing"
	ReasonInvalidSpec    = "InvalidSpec"
)

// SRRReconciler reconciles a SharedRouteRegistry object.
type SRRReconciler struct {
	client.Client
	Scheme            *runtime.Scheme
	GatewayNamespace  string
	GatewayName       string
	GatewaySectionRef string
}

// SetupWithManager wires the reconciler into the manager. The reconciler also
// owns its HTTPRoute (unstructured) and NetworkPolicy outputs so deletions
// trigger a re-reconcile.
func (r *SRRReconciler) SetupWithManager(mgr ctrl.Manager) error {
	if r.GatewayNamespace == "" {
		r.GatewayNamespace = "app-gateway"
	}
	if r.GatewayName == "" {
		r.GatewayName = GatewayParentName
	}
	if r.GatewaySectionRef == "" {
		r.GatewaySectionRef = GatewayParentSection
	}
	return ctrl.NewControllerManagedBy(mgr).
		Named("sharedrouteregistry").
		For(&gwapi.SharedRouteRegistry{}).
		Owns(&networkingv1.NetworkPolicy{}).
		Complete(r)
}

// Reconcile is the main controller loop.
func (r *SRRReconciler) Reconcile(ctx context.Context, req ctrl.Request) (ctrl.Result, error) {
	logger := log.FromContext(ctx).WithValues("srr", req.NamespacedName)

	srr := &gwapi.SharedRouteRegistry{}
	if err := r.Get(ctx, req.NamespacedName, srr); err != nil {
		if apierrors.IsNotFound(err) {
			return ctrl.Result{}, nil
		}
		return ctrl.Result{}, err
	}

	if !srr.DeletionTimestamp.IsZero() {
		return ctrl.Result{}, nil
	}

	switch srr.Spec.RouteMode {
	case "", gwapi.RouteModeGateway:
		return r.reconcileGatewayMode(ctx, logger, srr)
	case gwapi.RouteModeDirect:
		return r.reconcileDirectMode(ctx, logger, srr)
	default:
		msg := fmt.Sprintf("unknown routeMode %q", srr.Spec.RouteMode)
		logger.Info(msg)
		return ctrl.Result{}, r.setReady(ctx, srr, metav1.ConditionFalse, ReasonInvalidSpec, msg, "")
	}
}

func (r *SRRReconciler) reconcileDirectMode(ctx context.Context, logger logr.Logger, srr *gwapi.SharedRouteRegistry) (ctrl.Result, error) {
	if err := r.deleteHTTPRoute(ctx, srr); err != nil {
		return ctrl.Result{}, err
	}
	if err := r.deleteNetworkPolicy(ctx, srr); err != nil {
		return ctrl.Result{}, err
	}
	logger.Info("SRR routeMode=direct: cleaned HTTPRoute + NetworkPolicy")
	return ctrl.Result{}, r.setReady(ctx, srr, metav1.ConditionTrue, ReasonDirectMode, "Direct mode: HTTPRoute and NetworkPolicy removed.", "")
}

func (r *SRRReconciler) reconcileGatewayMode(ctx context.Context, logger logr.Logger, srr *gwapi.SharedRouteRegistry) (ctrl.Result, error) {
	if len(srr.Spec.HostPatterns) == 0 || srr.Spec.Upstream.ServiceName == "" {
		return ctrl.Result{}, r.setReady(ctx, srr, metav1.ConditionFalse, ReasonInvalidSpec, "hostPatterns or upstream.serviceName missing", "")
	}

	upstreamNS := srr.Spec.Upstream.ServiceNamespace
	if upstreamNS == "" {
		upstreamNS = srr.Namespace
	}

	svc := &corev1.Service{}
	if err := r.Get(ctx, types.NamespacedName{Namespace: upstreamNS, Name: srr.Spec.Upstream.ServiceName}, svc); err != nil {
		if apierrors.IsNotFound(err) {
			msg := fmt.Sprintf("backend Service %s/%s not found", upstreamNS, srr.Spec.Upstream.ServiceName)
			return ctrl.Result{Requeue: true}, r.setReady(ctx, srr, metav1.ConditionFalse, ReasonBackendMissing, msg, "")
		}
		return ctrl.Result{}, fmt.Errorf("get backend Service: %w", err)
	}

	port, err := resolveServicePort(svc, srr.Spec.Upstream)
	if err != nil {
		return ctrl.Result{}, r.setReady(ctx, srr, metav1.ConditionFalse, ReasonInvalidSpec, err.Error(), "")
	}

	routeName, err := r.applyHTTPRoute(ctx, srr, port)
	if err != nil {
		return ctrl.Result{}, fmt.Errorf("apply HTTPRoute: %w", err)
	}

	if err := r.applyNetworkPolicy(ctx, srr, svc, port); err != nil {
		return ctrl.Result{}, fmt.Errorf("apply NetworkPolicy: %w", err)
	}

	logger.Info("SRR reconciled", "httpRoute", routeName, "upstreamPort", port)
	return ctrl.Result{}, r.setReady(ctx, srr, metav1.ConditionTrue, ReasonReconciled, "HTTPRoute and NetworkPolicy reconciled", routeName)
}

func resolveServicePort(svc *corev1.Service, ref gwapi.UpstreamRef) (int32, error) {
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

func (r *SRRReconciler) applyHTTPRoute(ctx context.Context, srr *gwapi.SharedRouteRegistry, port int32) (string, error) {
	name := httpRouteName(srr)
	// PR-7: logical patterns translate into one HTTPRoute hostname
	// (`*.<platformDomain>`) plus a per-pattern Host RegularExpression
	// header match. Exact-host patterns keep Phase-A behaviour
	// (hostnames carries the verbatim host, no header match).
	hosts := MaterializeHostnames(srr.Spec.HostPatterns)
	headerMatches := MaterializeHostHeaders(srr.Spec.HostPatterns)
	if len(hosts) == 0 {
		return "", fmt.Errorf("hostPatterns produced no usable hostnames: %v", srr.Spec.HostPatterns)
	}
	parentRef := map[string]any{
		"group":       "gateway.networking.k8s.io",
		"kind":        "Gateway",
		"namespace":   r.GatewayNamespace,
		"name":        r.GatewayName,
		"sectionName": r.GatewaySectionRef,
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
	if err := controllerutilSetOwner(desired, srr, r.Scheme); err != nil {
		return "", err
	}

	current := &unstructured.Unstructured{}
	current.SetGroupVersionKind(schema.GroupVersionKind{Group: "gateway.networking.k8s.io", Version: "v1", Kind: "HTTPRoute"})
	err := r.Get(ctx, types.NamespacedName{Namespace: srr.Namespace, Name: name}, current)
	switch {
	case apierrors.IsNotFound(err):
		if err := r.Create(ctx, desired); err != nil {
			return "", err
		}
		return name, nil
	case err != nil:
		return "", err
	}
	if specChanged(current.Object["spec"], desired.Object["spec"]) {
		current.Object["spec"] = desired.Object["spec"]
		if labels := current.GetLabels(); labels == nil {
			current.SetLabels(map[string]string{ManagedByLabel: ManagedByValue, InstanceLabel: srr.Name})
		} else {
			labels[ManagedByLabel] = ManagedByValue
			labels[InstanceLabel] = srr.Name
			current.SetLabels(labels)
		}
		if err := r.Update(ctx, current); err != nil {
			return "", err
		}
	}
	return name, nil
}

func (r *SRRReconciler) deleteHTTPRoute(ctx context.Context, srr *gwapi.SharedRouteRegistry) error {
	obj := &unstructured.Unstructured{}
	obj.SetGroupVersionKind(schema.GroupVersionKind{Group: "gateway.networking.k8s.io", Version: "v1", Kind: "HTTPRoute"})
	obj.SetName(httpRouteName(srr))
	obj.SetNamespace(srr.Namespace)
	if err := r.Delete(ctx, obj); err != nil && !apierrors.IsNotFound(err) {
		return err
	}
	return nil
}

func (r *SRRReconciler) applyNetworkPolicy(ctx context.Context, srr *gwapi.SharedRouteRegistry, svc *corev1.Service, port int32) error {
	protocol := corev1.ProtocolTCP
	intPort := intstr.FromInt32(port)
	desired := &networkingv1.NetworkPolicy{
		ObjectMeta: metav1.ObjectMeta{
			Name:      NetworkPolicyName,
			Namespace: srr.Namespace,
			Labels: map[string]string{
				ManagedByLabel: ManagedByValue,
				InstanceLabel:  srr.Name,
			},
		},
		Spec: networkingv1.NetworkPolicySpec{
			PolicyTypes: []networkingv1.PolicyType{networkingv1.PolicyTypeIngress},
			PodSelector: metav1.LabelSelector{MatchLabels: svc.Spec.Selector},
			Ingress: []networkingv1.NetworkPolicyIngressRule{
				{
					From: []networkingv1.NetworkPolicyPeer{
						{
							NamespaceSelector: &metav1.LabelSelector{
								MatchLabels: map[string]string{
									"kubernetes.io/metadata.name": r.GatewayNamespace,
								},
							},
						},
					},
					Ports: []networkingv1.NetworkPolicyPort{
						{Protocol: &protocol, Port: &intPort},
					},
				},
			},
		},
	}
	if err := controllerutilSetOwner(desired, srr, r.Scheme); err != nil {
		return err
	}

	current := &networkingv1.NetworkPolicy{}
	err := r.Get(ctx, types.NamespacedName{Namespace: srr.Namespace, Name: NetworkPolicyName}, current)
	switch {
	case apierrors.IsNotFound(err):
		return r.Create(ctx, desired)
	case err != nil:
		return err
	}
	current.Spec = desired.Spec
	if current.Labels == nil {
		current.Labels = map[string]string{}
	}
	current.Labels[ManagedByLabel] = ManagedByValue
	current.Labels[InstanceLabel] = srr.Name
	if err := mergeOwnerRefs(current, srr); err != nil {
		return err
	}
	return r.Update(ctx, current)
}

func (r *SRRReconciler) deleteNetworkPolicy(ctx context.Context, srr *gwapi.SharedRouteRegistry) error {
	obj := &networkingv1.NetworkPolicy{
		ObjectMeta: metav1.ObjectMeta{
			Name:      NetworkPolicyName,
			Namespace: srr.Namespace,
		},
	}
	if err := r.Delete(ctx, obj); err != nil && !apierrors.IsNotFound(err) {
		return err
	}
	return nil
}

func (r *SRRReconciler) setReady(ctx context.Context, srr *gwapi.SharedRouteRegistry, status metav1.ConditionStatus, reason, msg, httpRouteName string) error {
	cond := metav1.Condition{
		Type:               ConditionReady,
		Status:             status,
		Reason:             reason,
		Message:            msg,
		ObservedGeneration: srr.Generation,
		LastTransitionTime: metav1.Now(),
	}
	patched := srr.DeepCopy()
	patched.Status.ObservedGeneration = srr.Generation
	patched.Status.HTTPRouteName = httpRouteName
	patched.Status.Conditions = upsertCondition(patched.Status.Conditions, cond)
	if equalStatus(srr.Status, patched.Status) {
		return nil
	}
	return r.Status().Patch(ctx, patched, client.MergeFrom(srr))
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

func equalStatus(a, b gwapi.SharedRouteRegistryStatus) bool {
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

func httpRouteName(srr *gwapi.SharedRouteRegistry) string {
	return srr.Name
}

// controllerutilSetOwner is a tiny stand-in for
// sigs.k8s.io/controller-runtime/pkg/controller/controllerutil.SetControllerReference.
// We avoid the import to keep the dependency graph minimal: the constructor
// only needs the controller=true / blockOwnerDeletion=true semantics.
func controllerutilSetOwner(obj client.Object, owner *gwapi.SharedRouteRegistry, scheme *runtime.Scheme) error {
	if obj.GetNamespace() != owner.GetNamespace() {
		return fmt.Errorf("cannot own across namespaces: owner %s/%s vs object %s/%s", owner.Namespace, owner.Name, obj.GetNamespace(), obj.GetName())
	}
	gvk, err := apiutilGVKForObject(owner, scheme)
	if err != nil {
		return err
	}
	t := true
	ref := metav1.OwnerReference{
		APIVersion:         gvk.GroupVersion().String(),
		Kind:               gvk.Kind,
		Name:               owner.GetName(),
		UID:                owner.GetUID(),
		Controller:         &t,
		BlockOwnerDeletion: &t,
	}
	refs := obj.GetOwnerReferences()
	for i, r := range refs {
		if r.UID == owner.UID {
			refs[i] = ref
			obj.SetOwnerReferences(refs)
			return nil
		}
	}
	obj.SetOwnerReferences(append(refs, ref))
	return nil
}

func mergeOwnerRefs(obj client.Object, owner *gwapi.SharedRouteRegistry) error {
	for _, r := range obj.GetOwnerReferences() {
		if r.UID == owner.UID {
			return nil
		}
	}
	t := true
	refs := append(obj.GetOwnerReferences(), metav1.OwnerReference{
		APIVersion:         gwapi.GroupVersion.String(),
		Kind:               "SharedRouteRegistry",
		Name:               owner.GetName(),
		UID:                owner.GetUID(),
		Controller:         &t,
		BlockOwnerDeletion: &t,
	})
	obj.SetOwnerReferences(refs)
	return nil
}

func apiutilGVKForObject(obj runtime.Object, scheme *runtime.Scheme) (schema.GroupVersionKind, error) {
	gvks, _, err := scheme.ObjectKinds(obj)
	if err != nil {
		return schema.GroupVersionKind{}, err
	}
	if len(gvks) == 0 {
		return schema.GroupVersionKind{}, fmt.Errorf("no GVK registered for %T", obj)
	}
	return gvks[0], nil
}

// specChanged compares HTTPRoute specs through their unstructured representation.
func specChanged(a, b any) bool {
	return !reflect.DeepEqual(a, b)
}
