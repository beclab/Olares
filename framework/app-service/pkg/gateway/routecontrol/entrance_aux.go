package routecontrol

import (
	"context"
	"fmt"
	"reflect"
	"strings"

	corev1 "k8s.io/api/core/v1"
	apierrors "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/apimachinery/pkg/labels"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"k8s.io/apimachinery/pkg/types"
	"k8s.io/apimachinery/pkg/util/intstr"
	"k8s.io/klog/v2"
	"sigs.k8s.io/controller-runtime/pkg/client"

	"github.com/beclab/Olares/framework/app-service/pkg/constants"
	srrv1alpha1 "github.com/beclab/Olares/framework/app-service/pkg/gateway/v1alpha1"
	"github.com/beclab/Olares/framework/app-service/pkg/mesh"
)

const (
	auxHelperWSPort     = int32(40010)
	auxHelperUploadPort = int32(40030)
	auxImagesPort       = int32(8080)
	auxImagesSvcName    = "tapr-images-svc"
	auxWSPath           = "/ws"
	auxUploadPath       = "/upload"
	auxImagesPath       = "/images/upload"
)

// entranceAuxNeeds describes Aux capabilities discovered for an entrance.
type entranceAuxNeeds struct {
	WS     bool // olares-ws-sidecar present
	Upload bool // olares-upload-sidecar present
	Images bool // application entrances always need /images/upload
}

func needsEntranceAux(srr *srrv1alpha1.SharedRouteRegistry) bool {
	return needsEntranceExtAuth(srr)
}

func entranceAuxHelperServiceName(upstreamSvc string) string {
	return upstreamSvc + "-deenvy-aux"
}

func entranceImagesRouteName(srr *srrv1alpha1.SharedRouteRegistry) string {
	return mesh.EntranceImagesRouteName(srr.Name)
}

// detectEntranceAuxNeeds inspects upstream Pods for ws/upload sidecars.
// Images is true for all application-class entrances (legacy edge always exposed /images/upload).
func detectEntranceAuxNeeds(ctx context.Context, c client.Client, srr *srrv1alpha1.SharedRouteRegistry, svc *corev1.Service) (entranceAuxNeeds, error) {
	var n entranceAuxNeeds
	if !needsEntranceAux(srr) {
		return n, nil
	}
	n.Images = true
	if svc == nil || len(svc.Spec.Selector) == 0 {
		return n, nil
	}
	var pods corev1.PodList
	if err := c.List(ctx, &pods, &client.ListOptions{
		Namespace:     svc.Namespace,
		LabelSelector: labels.SelectorFromSet(svc.Spec.Selector),
	}); err != nil {
		return n, fmt.Errorf("list pods for aux: %w", err)
	}
	for i := range pods.Items {
		for _, ctn := range pods.Items[i].Spec.Containers {
			switch ctn.Name {
			case constants.WsContainerName:
				n.WS = true
			case constants.UploadContainerName:
				n.Upload = true
			}
		}
	}
	return n, nil
}

func applyOrDeleteEntranceAux(ctx context.Context, c client.Client, gw GatewayRef, srr *srrv1alpha1.SharedRouteRegistry, svc *corev1.Service, port int32) error {
	if !needsEntranceAux(srr) {
		return deleteEntranceAux(ctx, c, srr, svc)
	}
	needs, err := detectEntranceAuxNeeds(ctx, c, srr, svc)
	if err != nil {
		klog.Errorf("deenvy: detect aux for %s/%s failed: %v", srr.Namespace, srr.Name, err)
		return err
	}
	if err := ensureAuxHelperService(ctx, c, srr, svc, needs); err != nil {
		return err
	}
	if err := patchMainHTTPRouteAuxRules(ctx, c, srr, svc, port, needs); err != nil {
		return err
	}
	if needs.Images {
		owner := ownerFromSRRNamespace(srr.Namespace)
		if srr.Labels != nil {
			if o := strings.TrimSpace(srr.Labels["applications.app.bytetrade.io/owner"]); o != "" {
				owner = o
			}
		}
		if owner == "" {
			owner = "unknown"
		}
		imagesNS := "user-system-" + owner
		if err := applyImagesReferenceGrant(ctx, c, srr, imagesNS); err != nil {
			return fmt.Errorf("apply images ReferenceGrant: %w", err)
		}
		if err := applyUnstructured(ctx, c, desiredEntranceImagesHTTPRoute(gw, srr, imagesNS), schema.GroupVersionKind{
			Group: "gateway.networking.k8s.io", Version: "v1", Kind: "HTTPRoute",
		}); err != nil {
			return fmt.Errorf("apply images HTTPRoute: %w", err)
		}
	} else if err := deleteEntranceImagesRoute(ctx, c, srr); err != nil {
		return err
	}
	return nil
}

func ensureAuxHelperService(ctx context.Context, c client.Client, srr *srrv1alpha1.SharedRouteRegistry, svc *corev1.Service, needs entranceAuxNeeds) error {
	if svc == nil || (!needs.WS && !needs.Upload) {
		if svc != nil {
			return deleteAuxHelperService(ctx, c, srr, svc)
		}
		return nil
	}
	name := entranceAuxHelperServiceName(svc.Name)
	var ports []corev1.ServicePort
	if needs.WS {
		ports = append(ports, corev1.ServicePort{Name: "ws", Port: auxHelperWSPort, TargetPort: intstr.FromInt32(auxHelperWSPort), Protocol: corev1.ProtocolTCP})
	}
	if needs.Upload {
		ports = append(ports, corev1.ServicePort{Name: "upload", Port: auxHelperUploadPort, TargetPort: intstr.FromInt32(auxHelperUploadPort), Protocol: corev1.ProtocolTCP})
	}
	desired := &corev1.Service{
		ObjectMeta: metav1.ObjectMeta{
			Name:      name,
			Namespace: svc.Namespace,
			Labels: map[string]string{
				ManagedByLabel:                ManagedByValue,
				InstanceLabel:                 srr.Name,
				"gateway.olares.io/auth-kind": mesh.AuthKindEntranceAux,
			},
		},
		Spec: corev1.ServiceSpec{
			Selector: svc.Spec.Selector,
			Ports:    ports,
		},
	}
	current := &corev1.Service{}
	err := c.Get(ctx, types.NamespacedName{Namespace: svc.Namespace, Name: name}, current)
	if apierrors.IsNotFound(err) {
		if err := c.Create(ctx, desired); err != nil {
			klog.Errorf("deenvy: create aux helper Service %s/%s failed: %v", svc.Namespace, name, err)
			return err
		}
		klog.Infof("deenvy: created aux helper Service %s/%s", svc.Namespace, name)
		return nil
	}
	if err != nil {
		return err
	}
	current.Spec.Selector = desired.Spec.Selector
	current.Spec.Ports = desired.Spec.Ports
	labels := current.Labels
	if labels == nil {
		labels = map[string]string{}
	}
	for k, v := range desired.Labels {
		labels[k] = v
	}
	current.Labels = labels
	if err := c.Update(ctx, current); err != nil {
		klog.Errorf("deenvy: update aux helper Service %s/%s failed: %v", svc.Namespace, name, err)
		return err
	}
	return nil
}

func deleteAuxHelperService(ctx context.Context, c client.Client, srr *srrv1alpha1.SharedRouteRegistry, svc *corev1.Service) error {
	if svc == nil {
		return nil
	}
	obj := &corev1.Service{ObjectMeta: metav1.ObjectMeta{Name: entranceAuxHelperServiceName(svc.Name), Namespace: svc.Namespace}}
	if err := c.Delete(ctx, obj); err != nil && !apierrors.IsNotFound(err) {
		return err
	}
	return nil
}

// patchMainHTTPRouteAuxRules prepends PathPrefix /ws and /upload rules on the main HTTPRoute.
// Those paths keep the entrance ExtAuth SecurityPolicy (same as the primary backend).
func patchMainHTTPRouteAuxRules(ctx context.Context, c client.Client, srr *srrv1alpha1.SharedRouteRegistry, svc *corev1.Service, port int32, needs entranceAuxNeeds) error {
	route := &unstructured.Unstructured{}
	route.SetGroupVersionKind(schema.GroupVersionKind{Group: "gateway.networking.k8s.io", Version: "v1", Kind: "HTTPRoute"})
	key := types.NamespacedName{Namespace: srr.Namespace, Name: httpRouteName(srr)}
	if err := c.Get(ctx, key, route); err != nil {
		if apierrors.IsNotFound(err) && !needs.WS && !needs.Upload {
			return nil
		}
		return fmt.Errorf("get main HTTPRoute for aux: %w", err)
	}
	rules, _, _ := unstructured.NestedSlice(route.Object, "spec", "rules")
	// Drop previous aux helper rules we own (PathPrefix /ws|/upload to *-deenvy-aux).
	var kept []any
	for _, r := range rules {
		rm, ok := r.(map[string]any)
		if !ok {
			kept = append(kept, r)
			continue
		}
		if isAuxHelperRule(rm) {
			continue
		}
		kept = append(kept, r)
	}
	var auxRules []any
	if needs.WS || needs.Upload {
		helperName := entranceAuxHelperServiceName(svc.Name)
		helperNS := svc.Namespace
		if needs.WS {
			auxRules = append(auxRules, auxPathRule(auxWSPath, helperName, helperNS, auxHelperWSPort))
		}
		if needs.Upload {
			auxRules = append(auxRules, auxPathRule(auxUploadPath, helperName, helperNS, auxHelperUploadPort))
		}
	}
	newRules := append(auxRules, kept...)
	if reflect.DeepEqual(rules, newRules) {
		return nil
	}
	if err := unstructured.SetNestedSlice(route.Object, newRules, "spec", "rules"); err != nil {
		return err
	}
	ann := route.GetAnnotations()
	if ann == nil {
		ann = map[string]string{}
	}
	ann[mesh.AuxCapabilitiesAnnotation] = auxCapsAnnotation(needs)
	route.SetAnnotations(ann)
	if err := c.Update(ctx, route); err != nil {
		klog.Errorf("deenvy: patch main HTTPRoute aux rules %s/%s failed: %v", key.Namespace, key.Name, err)
		return err
	}
	klog.Infof("deenvy: patched main HTTPRoute aux rules %s/%s ws=%v upload=%v", key.Namespace, key.Name, needs.WS, needs.Upload)
	_ = port
	return nil
}

func auxCapsAnnotation(n entranceAuxNeeds) string {
	var parts []string
	if n.WS {
		parts = append(parts, "ws")
	}
	if n.Upload {
		parts = append(parts, "upload")
	}
	if n.Images {
		parts = append(parts, "images")
	}
	return strings.Join(parts, ",")
}

func isAuxHelperRule(rule map[string]any) bool {
	backends, _, _ := unstructured.NestedSlice(rule, "backendRefs")
	for _, b := range backends {
		bm, ok := b.(map[string]any)
		if !ok {
			continue
		}
		name, _ := bm["name"].(string)
		if strings.HasSuffix(name, "-deenvy-aux") {
			return true
		}
	}
	return false
}

func auxPathRule(path, svcName, svcNS string, port int32) map[string]any {
	backend := map[string]any{
		"group": "", "kind": "Service", "name": svcName, "port": int64(port), "weight": int64(1),
	}
	if svcNS != "" {
		backend["namespace"] = svcNS
	}
	return map[string]any{
		"matches": []any{
			map[string]any{"path": map[string]any{"type": "PathPrefix", "value": path}},
		},
		"backendRefs": []any{backend},
	}
}

func desiredEntranceImagesHTTPRoute(gw GatewayRef, srr *srrv1alpha1.SharedRouteRegistry, imagesNS string) *unstructured.Unstructured {
	hosts := HTTPRouteHostnames(srr.Spec.HostPatterns)
	parentRef := map[string]any{
		"group": "gateway.networking.k8s.io", "kind": "Gateway",
		"namespace": gw.gatewayNamespace(), "name": gw.gatewayName(),
	}
	if section, ok := gatewaySectionForSRR(gw, srr); ok {
		parentRef["sectionName"] = section
	}
	obj := &unstructured.Unstructured{Object: map[string]any{
		"apiVersion": "gateway.networking.k8s.io/v1",
		"kind":       "HTTPRoute",
		"metadata": map[string]any{
			"name":      entranceImagesRouteName(srr),
			"namespace": srr.Namespace,
			"labels": map[string]any{
				ManagedByLabel:                ManagedByValue,
				InstanceLabel:                 srr.Name,
				"gateway.olares.io/auth-kind": mesh.AuthKindEntranceImages,
			},
			"annotations": map[string]any{
				mesh.AuxExtAuthBypassAnnotation: "true",
			},
		},
		"spec": map[string]any{
			"parentRefs": []any{parentRef},
			"hostnames":  hosts,
			"rules": []any{
				auxPathRule(auxImagesPath, auxImagesSvcName, imagesNS, auxImagesPort),
			},
		},
	}}
	setOwnerSRR(obj, srr)
	return obj
}

func applyImagesReferenceGrant(ctx context.Context, c client.Client, srr *srrv1alpha1.SharedRouteRegistry, imagesNS string) error {
	if imagesNS == srr.Namespace {
		return nil
	}
	name := "allow-httproute-images-" + srr.Namespace
	if len(name) > 63 {
		name = name[:63]
		name = strings.TrimRight(name, "-")
	}
	desired := &unstructured.Unstructured{Object: map[string]any{
		"apiVersion": referenceGrantAPIVersion,
		"kind":       "ReferenceGrant",
		"metadata": map[string]any{
			"name":      name,
			"namespace": imagesNS,
			"labels": map[string]any{
				ManagedByLabel: ManagedByValue,
				InstanceLabel:  srr.Name,
			},
		},
		"spec": map[string]any{
			"from": []any{
				map[string]any{"group": "gateway.networking.k8s.io", "kind": "HTTPRoute", "namespace": srr.Namespace},
			},
			"to": []any{
				map[string]any{"group": "", "kind": "Service", "name": auxImagesSvcName},
			},
		},
	}}
	desired.SetGroupVersionKind(referenceGrantGVK)
	current := &unstructured.Unstructured{}
	current.SetGroupVersionKind(referenceGrantGVK)
	err := c.Get(ctx, types.NamespacedName{Namespace: imagesNS, Name: name}, current)
	if apierrors.IsNotFound(err) {
		return c.Create(ctx, desired)
	}
	if err != nil {
		return err
	}
	if !unstructuredSpecEqual(current.Object["spec"], desired.Object["spec"]) {
		current.Object["spec"] = desired.Object["spec"]
		return c.Update(ctx, current)
	}
	return nil
}

func deleteEntranceImagesRoute(ctx context.Context, c client.Client, srr *srrv1alpha1.SharedRouteRegistry) error {
	obj := &unstructured.Unstructured{}
	obj.SetGroupVersionKind(schema.GroupVersionKind{Group: "gateway.networking.k8s.io", Version: "v1", Kind: "HTTPRoute"})
	obj.SetNamespace(srr.Namespace)
	obj.SetName(entranceImagesRouteName(srr))
	if err := c.Delete(ctx, obj); err != nil && !apierrors.IsNotFound(err) {
		return err
	}
	return nil
}

func deleteEntranceAux(ctx context.Context, c client.Client, srr *srrv1alpha1.SharedRouteRegistry, svc *corev1.Service) error {
	if err := deleteEntranceImagesRoute(ctx, c, srr); err != nil {
		return err
	}
	if svc == nil && srr != nil && srr.Spec.Upstream.ServiceName != "" {
		ns := srr.Spec.Upstream.ServiceNamespace
		if ns == "" {
			ns = srr.Namespace
		}
		svc = &corev1.Service{ObjectMeta: metav1.ObjectMeta{Name: srr.Spec.Upstream.ServiceName, Namespace: ns}}
	}
	if err := deleteAuxHelperService(ctx, c, srr, svc); err != nil {
		return err
	}
	// Strip aux rules from main route if present.
	if srr != nil {
		_ = patchMainHTTPRouteAuxRules(ctx, c, srr, svc, 0, entranceAuxNeeds{})
	}
	return nil
}
