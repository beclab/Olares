package access

import (
	"context"
	"maps"
	"testing"

	"github.com/beclab/Olares/framework/app-service/pkg/security"
	corev1 "k8s.io/api/core/v1"
	discoveryv1 "k8s.io/api/discovery/v1"
	netv1 "k8s.io/api/networking/v1"
	apierrors "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"k8s.io/apimachinery/pkg/types"
	"k8s.io/apimachinery/pkg/util/intstr"
	"k8s.io/utils/ptr"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/client/fake"
)

func TestApprovedBindingProjectsAndWithdrawsOnIdentityDrift(t *testing.T) {
	c, u := bindingFixture(t, RouteModeDirect)
	ctx := context.Background()
	b, err := readBinding(u)
	if err != nil {
		t.Fatal(err)
	}
	// First pass only records the deletion hook.
	if err := ReconcileAccessPolicy(ctx, c, u, true); err != nil {
		t.Fatal(err)
	}
	plan, err := CompilePolicy(revoked(b.policy), true)
	if err != nil {
		t.Fatal(err)
	}
	mustAbsent(t, c, plan.AuthorizationPolicyName, b.policy.Namespace)
	// The route exists but does not grant access before Linkerd accepts it.
	if err := ReconcileAccessPolicy(ctx, c, u, true); err == nil {
		t.Fatal("unobserved route was considered ready")
	}
	mustAbsent(t, c, plan.AuthorizationPolicyName, b.policy.Namespace)
	observeRoute(t, c, b.policy.Namespace, plan.HTTPRouteName, "policy.linkerd.io", "Server", b.policy.Namespace, plan.ServerName, "")
	if err := ReconcileAccessPolicy(ctx, c, u, true); err != nil {
		t.Fatal(err)
	}
	mustObject(t, c, authzGVK(), b.policy.Namespace, plan.AuthorizationPolicyName)
	sa := mustObject(t, c, corev1.SchemeGroupVersion.WithKind("ServiceAccount"), "source-ns", "caller")
	if err := c.Delete(ctx, sa); err != nil {
		t.Fatal(err)
	}
	sa.SetResourceVersion("")
	sa.SetUID("different-account")
	if err := c.Create(ctx, sa); err != nil {
		t.Fatal(err)
	}
	if err := ReconcileAccessPolicy(ctx, c, u, true); err == nil {
		t.Fatal("recreated caller account was accepted")
	}
	mustAbsent(t, c, plan.AuthorizationPolicyName, b.policy.Namespace)
	mustObject(t, c, schema.GroupVersionKind{Group: "policy.linkerd.io", Version: "v1beta3", Kind: "Server"}, b.policy.Namespace, plan.ServerName)
}

func TestInternalGatewayGrantsSourceLastAndUsesSeparateIdentities(t *testing.T) {
	c, u := bindingFixture(t, RouteModeViaAppGateway)
	ctx := context.Background()
	b, err := readBinding(u)
	if err != nil {
		t.Fatal(err)
	}
	if err := b.resolve(ctx, c); err != nil {
		t.Fatal(err)
	}
	out, in := revoked(b.policy), revoked(b.gatewayPolicy())
	outPlan, _ := CompilePolicy(out, true)
	inPlan, _ := CompilePolicy(in, true)
	if err := ReconcileAccessPolicy(ctx, c, u, true); err != nil {
		t.Fatal(err)
	}
	if err := ReconcileAccessPolicy(ctx, c, u, true); err == nil {
		t.Fatal("unobserved gateway route was considered ready")
	}
	mustAbsent(t, c, inPlan.AuthorizationPolicyName, in.Namespace)
	mustAbsent(t, c, outPlan.AuthorizationPolicyName, out.Namespace)
	assertAccessTransport(t, c, b, false)
	route := b.gatewayRoute()
	observeRoute(t, c, route.Namespace, route.Name, policyRouteGroup, "Gateway", internalNamespace, internalGateway, internalListener)
	observeRoute(t, c, out.Namespace, outPlan.HTTPRouteName, "policy.linkerd.io", "Server", out.Namespace, outPlan.ServerName, "")
	observeRoute(t, c, in.Namespace, inPlan.HTTPRouteName, "policy.linkerd.io", "Server", in.Namespace, internalGateway+"-server", "")
	if err := ReconcileAccessPolicy(ctx, c, u, true); err != nil {
		t.Fatal(err)
	}
	for _, tc := range []struct{ ns, name, want string }{
		{in.Namespace, inPlan.MeshTLSAuthName, "caller.source-ns.serviceaccount.identity.linkerd.cluster.local"},
		{out.Namespace, outPlan.MeshTLSAuthName, internalAccount + ".os-gateway.serviceaccount.identity.linkerd.cluster.local"},
	} {
		obj := mustObject(t, c, schema.GroupVersionKind{Group: "policy.linkerd.io", Version: "v1alpha1", Kind: "MeshTLSAuthentication"}, tc.ns, tc.name)
		got, _, _ := unstructured.NestedStringSlice(obj.Object, "spec", "identities")
		if len(got) != 1 || got[0] != tc.want {
			t.Fatalf("bound mesh identity was not enforced: %v", got)
		}
	}
	mustObject(t, c, authzGVK(), in.Namespace, inPlan.AuthorizationPolicyName)
	assertAccessTransport(t, c, b, true)
	if err := unstructured.SetNestedField(u.Object, DesiredRevoked, "spec", "desiredState"); err != nil {
		t.Fatal(err)
	}
	if err := ReconcileAccessPolicy(ctx, c, u, true); err != nil {
		t.Fatal(err)
	}
	mustAbsent(t, c, inPlan.AuthorizationPolicyName, in.Namespace)
	mustAbsent(t, c, outPlan.AuthorizationPolicyName, out.Namespace)
	assertAccessTransport(t, c, b, false)
}

func assertAccessTransport(t *testing.T, c client.Client, b boundPolicy, present bool) {
	t.Helper()
	for _, leg := range b.transportLegs() {
		out := b.transportPolicy(leg.from.namespace, leg.name+"-egress")
		in := b.transportPolicy(leg.to.namespace, leg.name+"-ingress")
		for _, want := range []*netv1.NetworkPolicy{&out, &in} {
			var got netv1.NetworkPolicy
			err := c.Get(context.Background(), client.ObjectKeyFromObject(want), &got)
			if !present {
				if !apierrors.IsNotFound(err) {
					t.Fatalf("unexpected retained transport policy %s: %v", want.Name, err)
				}
				continue
			}
			if err != nil || !security.IsAppServiceManagedExternalNP(&got) {
				t.Fatalf("missing or unrecognizable transport policy %s: %v", want.Name, err)
			}
			var peers []netv1.NetworkPolicyPeer
			var namespace string
			var selector map[string]string
			if want.Name == out.Name {
				peers = got.Spec.Egress[0].To
				namespace, selector = leg.to.namespace, accessTransportSelector(leg.from.selector).MatchLabels
			} else {
				peers = got.Spec.Ingress[0].From
				namespace, selector = leg.from.namespace, accessTransportSelector(leg.to.selector).MatchLabels
			}
			if len(peers) != 1 || peers[0].PodSelector == nil || peers[0].NamespaceSelector == nil ||
				peers[0].NamespaceSelector.MatchLabels["kubernetes.io/metadata.name"] != namespace ||
				!maps.Equal(got.Spec.PodSelector.MatchLabels, selector) {
				t.Fatalf("transport policy %s is not restricted to the bound workloads: %+v", want.Name, got.Spec)
			}
		}
	}
}

func bindingFixture(t *testing.T, mode string) (client.Client, *unstructured.Unstructured) {
	t.Helper()
	s := runtime.NewScheme()
	_ = corev1.AddToScheme(s)
	_ = netv1.AddToScheme(s)
	_ = discoveryv1.AddToScheme(s)
	makeObj := func(gvk schema.GroupVersionKind, ns, name, uid string, spec map[string]any) *unstructured.Unstructured {
		u := &unstructured.Unstructured{Object: map[string]any{"spec": spec}}
		u.SetGroupVersionKind(gvk)
		u.SetNamespace(ns)
		u.SetName(name)
		u.SetUID(types.UID(uid))
		u.SetGeneration(1)
		return u
	}
	ref := func(name string) map[string]any { return map[string]any{"name": name, "uid": name + "-uid"} }
	release := map[string]any{"applicationId": "source-id", "publisherId": "publisher", "version": "1.0.0", "artifactDigest": "sha256:source"}
	targetRelease := map[string]any{"applicationId": "target-id", "publisherId": "publisher", "version": "1.0.0", "artifactDigest": "sha256:target"}
	profile := map[string]any{"ownership": "User", "workloadClass": "Application", "ownerUserId": "alice"}
	scope := map[string]any{"sourceProfile": profile, "targetProfile": profile, "relation": "SameUser", "userContext": "WorkloadOnly", "delivery": mode}
	if mode == RouteModeViaAppGateway {
		scope["gatewayRef"] = map[string]any{"name": internalGateway, "uid": "internal-uid", "listenerName": internalListener}
	}
	target := map[string]any{
		"instanceRef": ref("target"), "namespaceRef": ref("target-ns"), "capabilitySetRef": ref("catalog"), "capabilityId": "read",
		"workload": map[string]any{"name": "target", "uid": "target-dep", "kind": "Deployment"}, "serviceAccount": ref("provider"),
		"service": map[string]any{"name": "backend", "uid": "backend-uid", "servicePort": int64(80), "containerPort": int64(8080)},
	}
	source := map[string]any{
		"instanceRef": ref("source"), "namespaceRef": ref("source-ns"), "workload": map[string]any{"name": "source", "uid": "source-dep", "kind": "Deployment"},
		"serviceAccount": ref("caller"), "release": release, "requestDigest": "sha256:intent",
	}
	approval := "2026-11-01T10:00:00Z"
	req := makeObj(accessGVK("AppAccessRequest"), PolicyNamespace, "request", "request-uid", map[string]any{
		"requesterUserId": "alice", "requestDigest": "sha256:intent", "sourceRelease": release, "accessScope": scope,
		"target": target, "expiresAt": "2026-11-01T11:00:00Z",
	})
	req.Object["status"] = map[string]any{"decision": map[string]any{"outcome": "Approved", "receiptId": "receipt", "actorUserId": "alice", "decidedAt": approval, "intentDigest": "sha256:approved", "intentDigestVersion": "v4"}}
	catalog := makeObj(accessGVK("AppCapabilitySet"), PolicyNamespace, "catalog", "catalog-uid", map[string]any{
		"release": targetRelease, "reviewRef": "review", "capabilities": []any{map[string]any{"id": "read", "protocol": "HTTP", "portName": "http", "serviceKey": "backend",
			"rules": []any{map[string]any{"method": "GET", "path": "/read", "pathType": "Exact"}}}},
	})
	policy := makeObj(accessGVK("AppAccessPolicy"), PolicyNamespace, "request-request-uid", "binding-uid", map[string]any{
		"desiredState": "Active", "binding": map[string]any{"accessScope": scope, "source": source, "target": target, "requestRef": ref("request"),
			"consent": map[string]any{"receiptId": "receipt", "userId": "alice", "confirmedAt": approval, "intentDigest": "sha256:approved", "intentDigestVersion": "v4"}},
	})
	objs := []client.Object{req, catalog, policy}
	for _, v := range []struct {
		name, ns, sa string
		release      map[string]any
	}{
		{"source", "source-ns", "caller", release}, {"target", "target-ns", "provider", targetRelease},
	} {
		app := makeObj(schema.GroupVersionKind{Group: "app.bytetrade.io", Version: "v1alpha1", Kind: "Application"}, "", v.name, v.name+"-uid",
			map[string]any{"namespace": v.ns, "appid": v.release["applicationId"], "owner": "alice"})
		app.SetAnnotations(map[string]string{"access.olares.io/release-version": "1.0.0", "access.olares.io/publisher-id": "publisher", "access.olares.io/artifact-digest": v.release["artifactDigest"].(string)})
		dep := makeObj(schema.GroupVersionKind{Group: "apps", Version: "v1", Kind: "Deployment"}, v.ns, v.name, v.name+"-dep", map[string]any{
			"selector": map[string]any{"matchLabels": map[string]any{"app": v.name}},
			"template": map[string]any{"spec": map[string]any{"serviceAccountName": v.sa, "containers": []any{
				map[string]any{"name": "web", "ports": []any{map[string]any{"name": "http", "containerPort": int64(8080)}}},
			}}},
		})
		objs = append(objs, app, dep, &corev1.Namespace{ObjectMeta: metav1.ObjectMeta{Name: v.ns, UID: types.UID(v.ns + "-uid")}},
			&corev1.ServiceAccount{ObjectMeta: metav1.ObjectMeta{Name: v.sa, Namespace: v.ns, UID: types.UID(v.sa + "-uid")}})
	}
	objs = append(objs, meshServiceFixture("target-ns", "backend", "provider", 80, 8080, map[string]string{"app": "target"})...)
	if mode == RouteModeViaAppGateway {
		objs = append(objs, meshServiceFixture(internalNamespace, internalGateway, internalAccount, 8080, internalPort, internalSelector())...)
		gw := makeObj(gatewayGVK, internalNamespace, internalGateway, "internal-uid", map[string]any{
			"infrastructure": map[string]any{"parametersRef": map[string]any{"name": internalGateway}},
			"listeners": []any{map[string]any{
				"name": internalListener, "protocol": "HTTP", "hostname": "*.access.olares.internal", "port": int64(8080),
				"allowedRoutes": map[string]any{"namespaces": map[string]any{"from": "Selector",
					"selector": map[string]any{"matchLabels": map[string]any{"kubernetes.io/metadata.name": PolicyNamespace}},
				}},
			}},
		})
		gw.Object["status"] = map[string]any{"conditions": observed(1, "Programmed"), "listeners": []any{
			map[string]any{"name": internalListener, "conditions": observed(1, "Programmed", "ResolvedRefs")},
		}}
		server := makeObj(schema.GroupVersionKind{Group: "policy.linkerd.io", Version: "v1beta3", Kind: "Server"}, internalNamespace, internalGateway+"-server", "server-uid", map[string]any{
			"port": int64(internalPort), "proxyProtocol": "HTTP/1", "accessPolicy": "deny", "podSelector": map[string]any{},
		})
		_ = unstructured.SetNestedStringMap(server.Object, internalSelector(), "spec", "podSelector", "matchLabels")
		objs = append(objs, gw, server)
	}
	c := fake.NewClientBuilder().WithScheme(s).WithObjects(objs...).Build()
	u := mustObject(t, c, accessGVK("AppAccessPolicy"), PolicyNamespace, "request-request-uid")
	return c, u
}

func meshServiceFixture(ns, name, sa string, sp, cp int32, selector map[string]string) []client.Object {
	uid := types.UID(name + "-uid")
	podUID := types.UID(name + "-pod")
	return []client.Object{
		&corev1.Service{ObjectMeta: metav1.ObjectMeta{Namespace: ns, Name: name, UID: uid},
			Spec: corev1.ServiceSpec{Selector: selector, Ports: []corev1.ServicePort{{Name: "http", Port: sp, TargetPort: intstr.FromInt32(cp)}}}},
		&corev1.Pod{ObjectMeta: metav1.ObjectMeta{Namespace: ns, Name: name, UID: podUID}, Spec: corev1.PodSpec{ServiceAccountName: sa},
			Status: corev1.PodStatus{ContainerStatuses: []corev1.ContainerStatus{{Name: "linkerd-proxy", Ready: true}}}},
		&discoveryv1.EndpointSlice{ObjectMeta: metav1.ObjectMeta{Namespace: ns, Name: name,
			Labels:          map[string]string{discoveryv1.LabelServiceName: name},
			OwnerReferences: []metav1.OwnerReference{{Kind: "Service", UID: uid}}},
			Ports: []discoveryv1.EndpointPort{{Port: &cp}}, Endpoints: []discoveryv1.Endpoint{{
				Conditions: discoveryv1.EndpointConditions{Ready: ptr.To(true)}, TargetRef: &corev1.ObjectReference{Kind: "Pod", Namespace: ns, Name: name, UID: podUID},
			}}},
	}
}

func observed(gen int64, conditions ...string) []any {
	var out []any
	for _, kind := range conditions {
		out = append(out, map[string]any{"type": kind, "status": "True", "observedGeneration": gen})
	}
	return out
}

func observeRoute(t *testing.T, c client.Client, ns, name, group, kind, parentNS, parent, section string) {
	t.Helper()
	u := mustObject(t, c, routeGVK, ns, name)
	u.Object["status"] = map[string]any{"parents": []any{map[string]any{
		"controllerName": "gateway.envoyproxy.io/gatewayclass-controller",
		"parentRef":      map[string]any{"group": group, "kind": kind, "namespace": parentNS, "name": parent, "sectionName": section},
		"conditions":     observed(u.GetGeneration(), "Accepted", "ResolvedRefs"),
	}}}
	if err := c.Update(context.Background(), u); err != nil {
		t.Fatal(err)
	}
}
func authzGVK() schema.GroupVersionKind {
	return schema.GroupVersionKind{Group: "policy.linkerd.io", Version: "v1alpha1", Kind: "AuthorizationPolicy"}
}
func mustAbsent(t *testing.T, c client.Client, name, ns string) {
	t.Helper()
	u := &unstructured.Unstructured{}
	u.SetGroupVersionKind(authzGVK())
	if err := c.Get(context.Background(), client.ObjectKey{Namespace: ns, Name: name}, u); !apierrors.IsNotFound(err) {
		t.Fatalf("authorization %s/%s was not absent: %v", ns, name, err)
	}
}
func mustObject(t *testing.T, c client.Client, gvk schema.GroupVersionKind, ns, name string) *unstructured.Unstructured {
	t.Helper()
	u := &unstructured.Unstructured{}
	u.SetGroupVersionKind(gvk)
	if err := c.Get(context.Background(), client.ObjectKey{Namespace: ns, Name: name}, u); err != nil {
		t.Fatal(err)
	}
	return u
}

func TestBindingRequiresApprovalForSameScopeAndCatalogService(t *testing.T) {
	for _, tc := range []struct {
		name, resource, field, replacement string
	}{
		{"receipt drift", "request", "receiptId", "other-receipt"},
		{"different method scope", "request", "intentDigest", "sha256:other"},
		{"unapproved gateway", "request", "outcome", "Denied"},
		{"different catalog service", "catalog", "serviceKey", "other-service"},
	} {
		t.Run(tc.name, func(t *testing.T) {
			ctx := context.Background()
			c, u := bindingFixture(t, RouteModeDirect)
			b, err := readBinding(u)
			if err != nil {
				t.Fatal(err)
			}
			if tc.resource == "request" {
				req := mustObject(t, c, accessGVK("AppAccessRequest"), PolicyNamespace, tc.resource)
				if err := unstructured.SetNestedField(req.Object, tc.replacement, "status", "decision", tc.field); err != nil {
					t.Fatal(err)
				}
				if err := c.Update(ctx, req); err != nil {
					t.Fatal(err)
				}
			} else {
				catalog := mustObject(t, c, accessGVK("AppCapabilitySet"), PolicyNamespace, tc.resource)
				items, _, _ := unstructured.NestedSlice(catalog.Object, "spec", "capabilities")
				items[0].(map[string]any)[tc.field] = tc.replacement
				if err := unstructured.SetNestedSlice(catalog.Object, items, "spec", "capabilities"); err != nil {
					t.Fatal(err)
				}
				if err := c.Update(ctx, catalog); err != nil {
					t.Fatal(err)
				}
			}
			if err := ReconcileAccessPolicy(ctx, c, u, true); err == nil {
				t.Fatal("a grant with drifted approval or capability admitted")
			}
			plan, _ := CompilePolicy(revoked(b.policy), true)
			mustAbsent(t, c, plan.AuthorizationPolicyName, b.policy.Namespace)
		})
	}
}

func TestGrantCannotBeRenamedToReuseItsApproval(t *testing.T) {
	_, u := bindingFixture(t, RouteModeDirect)
	u.SetName("second-grant-for-the-same-request")
	if _, err := readBinding(u); err == nil {
		t.Fatal("a renamed grant could replay a revoked approval")
	}
}

func TestAdmissionOutageWithdrawsExistingAllowAndPreservesRevocation(t *testing.T) {
	ctx := context.Background()
	c, u := bindingFixture(t, RouteModeDirect)
	b, err := readBinding(u)
	if err != nil {
		t.Fatal(err)
	}
	if err := ReconcileAccessPolicy(ctx, c, u, true); err != nil {
		t.Fatal(err)
	}
	plan, err := CompilePolicy(revoked(b.policy), true)
	if err != nil {
		t.Fatal(err)
	}
	if err := ReconcileAccessPolicy(ctx, c, u, true); err == nil {
		t.Fatal("unobserved Linkerd route approved")
	}
	observeRoute(t, c, b.policy.Namespace, plan.HTTPRouteName, "policy.linkerd.io", "Server", b.policy.Namespace, plan.ServerName, "")
	if err := ReconcileAccessPolicy(ctx, c, u, true); err != nil {
		t.Fatal(err)
	}
	mustObject(t, c, authzGVK(), b.policy.Namespace, plan.AuthorizationPolicyName)

	if err := reconcilePolicyPass(ctx, c, false); err != nil {
		t.Fatal(err)
	}
	mustAbsent(t, c, plan.AuthorizationPolicyName, b.policy.Namespace)
	grant := mustObject(t, c, accessGVK("AppAccessPolicy"), PolicyNamespace, u.GetName())
	if value(grant.Object, "spec", "desiredState") != DesiredActive {
		t.Fatal("an admission outage changed the approved desired state")
	}
	if err := unstructured.SetNestedField(grant.Object, DesiredRevoked, "spec", "desiredState"); err != nil {
		t.Fatal(err)
	}
	if err := c.Update(ctx, grant); err != nil {
		t.Fatal(err)
	}
	if err := reconcilePolicyPass(ctx, c, true); err != nil {
		t.Fatal(err)
	}
	mustAbsent(t, c, plan.AuthorizationPolicyName, b.policy.Namespace)
	mustObject(t, c, schema.GroupVersionKind{Group: "policy.linkerd.io", Version: "v1beta3", Kind: "Server"}, b.policy.Namespace, plan.ServerName)
}
