package access

import (
	"context"
	"crypto/sha256"
	"fmt"
	"maps"
	"reflect"
	"time"

	appsv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

type objectBinding struct {
	Name string
	UID  string
}

func bindingRef(m map[string]any, path ...string) objectBinding {
	return objectBinding{value(m, append(path, "name")...), value(m, append(path, "uid")...)}
}

func value(m map[string]any, path ...string) string {
	v, _, _ := unstructured.NestedString(m, path...)
	return v
}

func objectMap(m map[string]any, path ...string) map[string]any {
	v, _, _ := unstructured.NestedMap(m, path...)
	return v
}

func boundObject(ctx context.Context, c client.Client, ns string, gvk schema.GroupVersionKind, ref objectBinding) (*unstructured.Unstructured, error) {
	if ref.Name == "" || ref.UID == "" {
		return nil, fmt.Errorf("%s binding requires name and UID", gvk.Kind)
	}
	u := &unstructured.Unstructured{}
	u.SetGroupVersionKind(gvk)
	if err := c.Get(ctx, client.ObjectKey{Namespace: ns, Name: ref.Name}, u); err != nil {
		return nil, err
	}
	if string(u.GetUID()) != ref.UID || !u.GetDeletionTimestamp().IsZero() {
		return nil, fmt.Errorf("%s binding no longer identifies the live object", gvk.Kind)
	}
	return u, nil
}

func accessGVK(kind string) schema.GroupVersionKind {
	return schema.GroupVersionKind{Group: "access.olares.io", Version: "v1alpha1", Kind: kind}
}

type boundPolicy struct {
	uid             string
	resourceVersion string
	policy          PolicySpec
	binding         map[string]any
	source          map[string]any
	target          map[string]any
	scope           map[string]any
}

func readBinding(u *unstructured.Unstructured) (boundPolicy, error) {
	if u == nil || u.GetNamespace() != AdmissionTLSNamespace || u.GetUID() == "" || u.GroupVersionKind() != accessGVK("AppAccessPolicy") {
		return boundPolicy{}, fmt.Errorf("expected a persisted AppAccessPolicy in %s", AdmissionTLSNamespace)
	}
	b := objectMap(u.Object, "spec", "binding")
	source, target, scope := objectMap(b, "source"), objectMap(b, "target"), objectMap(b, "accessScope")
	p := PolicySpec{
		Name:              u.GetName(),
		ProjectionID:      fmt.Sprintf("grant-%x", sha256.Sum256([]byte(u.GetUID())))[:30],
		Namespace:         bindingRef(target, "namespaceRef").Name,
		DesiredState:      value(u.Object, "spec", "desiredState"),
		RouteMode:         value(scope, "delivery"),
		SourceSANamespace: bindingRef(source, "namespaceRef").Name,
		SourceSAName:      bindingRef(source, "serviceAccount").Name,
	}
	targetUID := bindingRef(target, "workload").UID
	targetPort, _, _ := unstructured.NestedInt64(target, "service", "containerPort")
	if targetUID == "" || targetPort < 1 || targetPort > 65535 {
		return boundPolicy{}, fmt.Errorf("binding requires workload UID and container port")
	}
	p.ServerID = fmt.Sprintf("workload-%x", sha256.Sum256([]byte(fmt.Sprintf("%s/%d", targetUID, targetPort))))[:34]
	if p.Namespace == "" || (p.DesiredState != DesiredActive && p.DesiredState != DesiredRevoked) {
		return boundPolicy{}, fmt.Errorf("binding target namespace and desiredState are required")
	}
	return boundPolicy{uid: string(u.GetUID()), resourceVersion: u.GetResourceVersion(), policy: p, binding: b, source: source, target: target, scope: scope}, nil
}

func (b boundPolicy) stillActive(ctx context.Context, c client.Client) error {
	u, err := boundObject(ctx, c, AdmissionTLSNamespace, accessGVK("AppAccessPolicy"), objectBinding{b.policy.Name, b.uid})
	if err != nil {
		return err
	}
	if value(u.Object, "spec", "desiredState") != DesiredActive || u.GetResourceVersion() != b.resourceVersion {
		return fmt.Errorf("grant changed during reconciliation")
	}
	return nil
}

func (b *boundPolicy) resolve(ctx context.Context, c client.Client) error {
	if b.policy.RouteMode != RouteModeDirect && b.policy.RouteMode != RouteModeViaAppGateway {
		return fmt.Errorf("unsupported binding delivery")
	}
	if value(b.scope, "userContext") != "WorkloadOnly" || objectMap(b.binding, "systemRuleRef") != nil {
		return fmt.Errorf("grant requires a verified-user or signed-system authority that is not configured")
	}
	req, err := boundObject(ctx, c, AdmissionTLSNamespace, accessGVK("AppAccessRequest"), bindingRef(b.binding, "requestRef"))
	if err != nil {
		return err
	}
	if err := b.checkApproval(req.Object); err != nil {
		return err
	}

	catalog, err := boundObject(ctx, c, AdmissionTLSNamespace, accessGVK("AppCapabilitySet"), bindingRef(b.target, "capabilitySetRef"))
	if err != nil {
		return err
	}
	if value(catalog.Object, "spec", "reviewRef") == "" || !validRelease(objectMap(catalog.Object, "spec", "release")) {
		return fmt.Errorf("catalog review or release identity is missing")
	}
	for _, side := range []map[string]any{b.source, b.target} {
		ns := bindingRef(side, "namespaceRef")
		if _, err := boundObject(ctx, c, "", corev1.SchemeGroupVersion.WithKind("Namespace"), ns); err != nil {
			return err
		}
		if _, err := boundObject(ctx, c, ns.Name, corev1.SchemeGroupVersion.WithKind("ServiceAccount"), bindingRef(side, "serviceAccount")); err != nil {
			return err
		}
	}
	sourceApp, err := boundApplication(ctx, c, b.source, objectMap(b.source, "release"), objectMap(b.scope, "sourceProfile"))
	if err != nil {
		return err
	}
	targetApp, err := boundApplication(ctx, c, b.target, objectMap(catalog.Object, "spec", "release"), objectMap(b.scope, "targetProfile"))
	if err != nil {
		return err
	}
	if sourceApp.GetUID() == targetApp.GetUID() {
		return fmt.Errorf("source and target must be distinct")
	}

	if _, err := boundWorkload(ctx, c, b.source); err != nil {
		return err
	}
	workload, err := boundWorkload(ctx, c, b.target)
	if err != nil {
		return err
	}
	selector, _, _ := unstructured.NestedStringMap(workload.Object, "spec", "selector", "matchLabels")
	expressions, _, _ := unstructured.NestedSlice(workload.Object, "spec", "selector", "matchExpressions")
	if len(selector) == 0 || len(expressions) != 0 {
		return fmt.Errorf("target workload must have an exact label selector")
	}
	ref := bindingRef(b.target, "service")
	serviceObj, err := boundObject(ctx, c, b.policy.Namespace, corev1.SchemeGroupVersion.WithKind("Service"), ref)
	if err != nil {
		return err
	}
	serviceSelector, _, _ := unstructured.NestedStringMap(serviceObj.Object, "spec", "selector")
	if !maps.Equal(selector, serviceSelector) || value(serviceObj.Object, "spec", "type") == "ExternalName" {
		return fmt.Errorf("Service does not select exactly the bound workload")
	}
	capability, err := selectCapability(catalog.Object, value(b.target, "capabilityId"))
	if err != nil {
		return err
	}
	if value(capability, "serviceKey") != ref.Name {
		return fmt.Errorf("capability service key does not resolve to the bound Service")
	}
	port, err := boundPort(b.target, serviceObj.Object, workload.Object, value(capability, "portName"))
	if err != nil {
		return err
	}
	b.policy.TargetSelector, b.policy.ServerPort = selector, port
	b.policy.Rules = readRules(capability, "rules")
	return ValidateHTTPRules(b.policy.Rules)
}

func (b boundPolicy) checkApproval(req map[string]any) error {
	spec, decision := objectMap(req, "spec"), objectMap(req, "status", "decision")
	consent := objectMap(b.binding, "consent")
	if value(decision, "outcome") != "Approved" ||
		!sameFields(decision, consent, "receiptId", "intentDigest", "intentDigestVersion") ||
		value(decision, "actorUserId") == "" || value(decision, "actorUserId") != value(consent, "userId") ||
		value(spec, "requesterUserId") != value(consent, "userId") ||
		value(spec, "requestDigest") == "" || value(spec, "requestDigest") != value(b.source, "requestDigest") ||
		!validRelease(objectMap(spec, "sourceRelease")) ||
		!reflect.DeepEqual(objectMap(spec, "sourceRelease"), objectMap(b.source, "release")) ||
		!reflect.DeepEqual(objectMap(spec, "accessScope"), b.scope) {
		return fmt.Errorf("binding does not match the approved request")
	}
	t, err := time.Parse(time.RFC3339Nano, value(decision, "decidedAt"))
	expires, expiresErr := time.Parse(time.RFC3339Nano, value(spec, "expiresAt"))
	confirmed, confirmedErr := time.Parse(time.RFC3339Nano, value(consent, "confirmedAt"))
	if err != nil || expiresErr != nil || confirmedErr != nil || !t.Before(expires) || !t.Equal(confirmed) {
		return fmt.Errorf("approval timestamps do not match the request")
	}
	target := objectMap(spec, "target")
	for _, field := range []string{"instanceRef", "namespaceRef", "capabilitySetRef"} {
		if bindingRef(target, field) != bindingRef(b.target, field) {
			return fmt.Errorf("approved target %s does not match", field)
		}
	}
	if value(target, "capabilityId") == "" || value(target, "capabilityId") != value(b.target, "capabilityId") {
		return fmt.Errorf("approved capability does not match")
	}
	if value(b.scope, "sourceProfile", "ownership") == "User" &&
		value(b.scope, "sourceProfile", "ownerUserId") != value(consent, "userId") {
		return fmt.Errorf("approval actor does not own the source instance")
	}
	switch value(b.scope, "relation") {
	case "SameUser":
		if value(b.scope, "sourceProfile", "ownership") != "User" || value(b.scope, "targetProfile", "ownership") != "User" ||
			!sameFields(objectMap(b.scope, "sourceProfile"), objectMap(b.scope, "targetProfile"), "ownerUserId") {
			return fmt.Errorf("same-user grants require the same bound owner")
		}
	case "SharedCapability":
		if value(b.scope, "targetProfile", "ownership") != "Shared" {
			return fmt.Errorf("target is not shared")
		}
	case "PlatformCapability":
		if value(b.scope, "targetProfile", "ownership") != "Platform" {
			return fmt.Errorf("target is not platform-owned")
		}
	default:
		return fmt.Errorf("unsupported approval relation")
	}
	return nil
}

func sameFields(a, b map[string]any, fields ...string) bool {
	for _, field := range fields {
		if value(a, field) == "" || value(a, field) != value(b, field) {
			return false
		}
	}
	return true
}
func validRelease(m map[string]any) bool {
	return sameFields(m, m, "applicationId", "publisherId", "version", "artifactDigest")
}

func boundApplication(ctx context.Context, c client.Client, side, release, profile map[string]any) (*unstructured.Unstructured, error) {
	obj, err := boundObject(ctx, c, "", schema.GroupVersionKind{Group: "app.bytetrade.io", Version: "v1alpha1", Kind: "Application"}, bindingRef(side, "instanceRef"))
	if err != nil {
		return nil, err
	}
	if value(obj.Object, "spec", "namespace") != bindingRef(side, "namespaceRef").Name ||
		value(obj.Object, "spec", "appid") != value(release, "applicationId") ||
		obj.GetAnnotations()["access.olares.io/release-version"] != value(release, "version") ||
		obj.GetAnnotations()["access.olares.io/artifact-digest"] != value(release, "artifactDigest") ||
		obj.GetAnnotations()["access.olares.io/publisher-id"] != value(release, "publisherId") {
		return nil, fmt.Errorf("bound Application does not match its namespace or reviewed release")
	}
	if value(profile, "ownership") == "User" && (value(profile, "ownerUserId") == "" || value(obj.Object, "spec", "owner") != value(profile, "ownerUserId")) {
		return nil, fmt.Errorf("bound Application does not match approved owner")
	}
	switch value(profile, "ownership") {
	case "Platform":
		system, _, _ := unstructured.NestedBool(obj.Object, "spec", "isSysApp")
		if !system {
			return nil, fmt.Errorf("bound Application is not platform-owned")
		}
	case "Shared":
		if obj.GetLabels()["app.bytetrade.io/scope"] != "shared" {
			return nil, fmt.Errorf("bound Application is not shared")
		}
	case "User":
	default:
		return nil, fmt.Errorf("unsupported Application ownership")
	}
	return obj, nil
}

func boundWorkload(ctx context.Context, c client.Client, side map[string]any) (*unstructured.Unstructured, error) {
	workload := objectMap(side, "workload")
	kind := value(workload, "kind")
	if kind != "Deployment" && kind != "StatefulSet" {
		return nil, fmt.Errorf("unsupported workload kind")
	}
	obj, err := boundObject(ctx, c, bindingRef(side, "namespaceRef").Name, appsv1.SchemeGroupVersion.WithKind(kind), bindingRef(side, "workload"))
	if err != nil {
		return nil, err
	}
	sa := value(obj.Object, "spec", "template", "spec", "serviceAccountName")
	hostnet, _, _ := unstructured.NestedBool(obj.Object, "spec", "template", "spec", "hostNetwork")
	if sa == "" || sa == "default" || sa != bindingRef(side, "serviceAccount").Name || hostnet {
		return nil, fmt.Errorf("workload must use its bound dedicated ServiceAccount and pod network")
	}
	return obj, nil
}

func selectCapability(catalog map[string]any, id string) (map[string]any, error) {
	items, _, _ := unstructured.NestedSlice(catalog, "spec", "capabilities")
	var got map[string]any
	for _, item := range items {
		capability, _ := item.(map[string]any)
		if id != "" && value(capability, "id") == id {
			if got != nil {
				return nil, fmt.Errorf("duplicate capability %q", id)
			}
			got = capability
		}
	}
	if got == nil || value(got, "protocol") != "HTTP" || (value(got, "enforcement") != "" && value(got, "enforcement") != "HTTPRoute") {
		return nil, fmt.Errorf("reviewed HTTPRoute capability was not found")
	}
	return got, nil
}

func boundPort(target, service, workload map[string]any, portName string) (int32, error) {
	expectedSP, _, _ := unstructured.NestedInt64(target, "service", "servicePort")
	expectedCP, _, _ := unstructured.NestedInt64(target, "service", "containerPort")
	ports, _, _ := unstructured.NestedSlice(service, "spec", "ports")
	containers, _, _ := unstructured.NestedSlice(workload, "spec", "template", "spec", "containers")
	if expectedSP < 1 || expectedSP > 65535 || expectedCP < 1 || expectedCP > 65535 || portName == "" {
		return 0, fmt.Errorf("invalid bound HTTP port")
	}
	for _, entry := range ports {
		port, _ := entry.(map[string]any)
		num, _, _ := unstructured.NestedInt64(port, "port")
		if value(port, "name") != portName || num != expectedSP || (value(port, "protocol") != "" && value(port, "protocol") != "TCP") {
			continue
		}
		for _, entry := range containers {
			container, _ := entry.(map[string]any)
			podPorts, _, _ := unstructured.NestedSlice(container, "ports")
			for _, entry := range podPorts {
				podPort, _ := entry.(map[string]any)
				num, _, _ := unstructured.NestedInt64(podPort, "containerPort")
				if num != expectedCP || (value(podPort, "protocol") != "" && value(podPort, "protocol") != "TCP") {
					continue
				}
				switch to := port["targetPort"].(type) {
				case nil:
					if expectedSP == expectedCP {
						return int32(expectedCP), nil
					}
				case int64:
					if to == expectedCP {
						return int32(expectedCP), nil
					}
				case string:
					if to != "" && to == value(podPort, "name") {
						return int32(expectedCP), nil
					}
				}
			}
		}
	}
	return 0, fmt.Errorf("approved Service and workload ports no longer match")
}

func readRules(obj map[string]any, path ...string) []HTTPRule {
	items, _, _ := unstructured.NestedSlice(obj, path...)
	out := make([]HTTPRule, 0, len(items))
	for _, item := range items {
		r, _ := item.(map[string]any)
		out = append(out, HTTPRule{Method: value(r, "method"), Path: value(r, "path"), PathType: value(r, "pathType")})
	}
	return out
}
