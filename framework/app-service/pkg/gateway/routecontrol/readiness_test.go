package routecontrol

import (
	"context"
	"slices"
	"strings"

	"github.com/beclab/Olares/framework/app-service/pkg/gateway"
	srrv1alpha1 "github.com/beclab/Olares/framework/app-service/pkg/gateway/v1alpha1"
	"github.com/beclab/Olares/framework/app-service/pkg/mesh"
	corev1 "k8s.io/api/core/v1"
	discoveryv1 "k8s.io/api/discovery/v1"
	apierrors "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"k8s.io/apimachinery/pkg/types"
	"k8s.io/apimachinery/pkg/util/intstr"
	"k8s.io/utils/ptr"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/client/fake"
	"testing"

	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
)

func TestEntryRouteReadinessRequiresCurrentEntryListenerStatus(t *testing.T) {
	cond := func(kind string, generation int64) map[string]any {
		return map[string]any{"type": kind, "status": "True", "observedGeneration": generation}
	}
	route := &unstructured.Unstructured{Object: map[string]any{
		"status": map[string]any{"parents": []any{map[string]any{
			"controllerName": "gateway.envoyproxy.io/gatewayclass-controller",
			"parentRef":      map[string]any{"name": defaultGatewayName, "namespace": defaultGatewayNS, "sectionName": l4MTLSGatewaySectN},
			"conditions":     []any{cond("Accepted", 2), cond("ResolvedRefs", 2)},
		}}},
	}}
	route.SetNamespace("user-space-alice")
	route.SetGeneration(2)
	if !entryParentReady(route, GatewayRef{}) {
		t.Fatal("matching current-generation status must be ready")
	}
	route.SetGeneration(3)
	if entryParentReady(route, GatewayRef{}) {
		t.Fatal("status from an earlier generation must block cutover")
	}
	route.SetGeneration(2)
	parent := route.Object["status"].(map[string]any)["parents"].([]any)[0].(map[string]any)
	parent["parentRef"].(map[string]any)["sectionName"] = "https"
	if entryParentReady(route, GatewayRef{}) {
		t.Fatal("status for the regular HTTPS listener must not approve entry mTLS cutover")
	}
}

func TestServingReadinessWaitsForPolicyAndMeshedEndpoints(t *testing.T) {
	ctx := context.Background()
	scheme := testScheme(t)
	_ = discoveryv1.AddToScheme(scheme)
	srr := &srrv1alpha1.SharedRouteRegistry{
		ObjectMeta: metav1.ObjectMeta{Namespace: "user-space-alice", Name: "srr"},
		Spec:       srrv1alpha1.SharedRouteRegistrySpec{Upstream: srrv1alpha1.UpstreamRef{ServiceName: "app"}},
	}
	obj := func(gvk schema.GroupVersionKind, ns, name string, status map[string]any) *unstructured.Unstructured {
		u := &unstructured.Unstructured{Object: map[string]any{"status": status}}
		u.SetGroupVersionKind(gvk)
		u.SetNamespace(ns)
		u.SetName(name)
		u.SetGeneration(1)
		return u
	}
	cond := func(kinds ...string) []any {
		var out []any
		for _, k := range kinds {
			out = append(out, map[string]any{"type": k, "status": "True", "observedGeneration": int64(1)})
		}
		return out
	}
	entryPolicy := func(gvk schema.GroupVersionKind, name, section string, status map[string]any) *unstructured.Unstructured {
		policy := obj(gvk, defaultGatewayNS, name, status)
		policy.Object["spec"] = map[string]any{
			"headers": map[string]any{"earlyRequestHeaders": entryIdentityHeaderSettings()},
			"targetRef": map[string]any{
				"group": "gateway.networking.k8s.io", "kind": "Gateway", "name": defaultGatewayName, "sectionName": section,
			},
			"tls": map[string]any{"clientValidation": map[string]any{
				"mode":              "RequireAndVerify",
				"caCertificateRefs": []any{map[string]any{"name": "olares-l4-agw-mtls", "kind": "Secret", "group": ""}},
				"subjectAltNames":   map[string]any{"dnsNames": []any{map[string]any{"type": "Exact", "value": "l4-bfl-proxy.entry.client"}}},
			}},
		}
		return policy
	}
	ref := map[string]any{"name": defaultGatewayName, "namespace": defaultGatewayNS, "sectionName": l4MTLSGatewaySectN}
	owner := "gateway.envoyproxy.io/gatewayclass-controller"
	gw := obj(gatewayGVK, defaultGatewayNS, defaultGatewayName, map[string]any{
		"conditions": cond("Programmed"), "listeners": []any{map[string]any{"name": l4MTLSGatewaySectN, "conditions": cond("Programmed", "ResolvedRefs")}},
	})
	route := obj(schema.GroupVersionKind{Group: "gateway.networking.k8s.io", Version: "v1", Kind: "HTTPRoute"}, srr.Namespace, "route", map[string]any{
		"parents": []any{map[string]any{"parentRef": ref, "controllerName": owner, "conditions": cond("Accepted", "ResolvedRefs")}},
	})
	policyStatus := map[string]any{"ancestors": []any{map[string]any{"ancestorRef": ref, "controllerName": owner, "conditions": cond("Accepted")}}}
	security := obj(securityPolicyGVK, srr.Namespace, securityPolicyName(srr), policyStatus)
	ctp := entryPolicy(schema.GroupVersionKind{Group: "gateway.envoyproxy.io", Version: "v1alpha1", Kind: "ClientTrafficPolicy"}, "app-gateway-l4-entry-mtls", l4MTLSGatewaySectN, policyStatus)
	objects := []client.Object{gw, route, security, ctp}
	for _, ep := range []struct{ ns, svc string }{{defaultGatewayNS, "app-gateway-l4-mtls"}, {srr.Namespace, "app"}} {
		uid := types.UID(ep.svc + "-uid")
		selector := map[string]string{"app": ep.svc}
		port := int32(8080)
		if ep.svc == entryMeshService.ServiceName {
			port = entryMeshService.Port
		}
		service := &corev1.Service{ObjectMeta: metav1.ObjectMeta{Namespace: ep.ns, Name: ep.svc, UID: uid + "-service"},
			Spec: corev1.ServiceSpec{Selector: selector, Ports: []corev1.ServicePort{{Port: port}}}}
		pod := &corev1.Pod{ObjectMeta: metav1.ObjectMeta{Namespace: ep.ns, Name: ep.svc, UID: uid, Labels: selector,
			Annotations: map[string]string{mesh.LinkerdDefaultInboundPolicyAnnotation: mesh.InboundPolicyClusterAuthenticated}},
			Spec: corev1.PodSpec{
				InitContainers: []corev1.Container{{Name: "linkerd-init", Args: []string{"--inbound-ports-to-ignore", "3389,4190,4191"}}},
				Containers: []corev1.Container{{Name: mesh.LinkerdProxyContainer,
					Env: []corev1.EnvVar{{Name: mesh.LinkerdProxyInboundPolicyEnv, Value: mesh.InboundPolicyClusterAuthenticated}}}}},
			Status: corev1.PodStatus{PodIP: "10.0.1.2",
				ContainerStatuses: []corev1.ContainerStatus{{Name: mesh.LinkerdProxyContainer, Ready: true}}}}
		slice := &discoveryv1.EndpointSlice{ObjectMeta: metav1.ObjectMeta{Namespace: ep.ns, Name: ep.svc,
			Labels:          map[string]string{discoveryv1.LabelServiceName: ep.svc},
			OwnerReferences: []metav1.OwnerReference{*metav1.NewControllerRef(service, corev1.SchemeGroupVersion.WithKind("Service"))}},
			Ports: []discoveryv1.EndpointPort{{Port: ptr.To(port)}},
			Endpoints: []discoveryv1.Endpoint{{TargetRef: &corev1.ObjectReference{Kind: "Pod", Namespace: ep.ns, Name: ep.svc, UID: uid},
				Addresses: []string{"10.0.1.2"}, Conditions: discoveryv1.EndpointConditions{Ready: ptr.To(true)}}}}
		objects = append(objects, service, pod, slice)
	}
	for _, index := range []int{0, 1, 2, 3, 4} {
		t.Run("missing "+objects[index].GetName(), func(t *testing.T) {
			var fixtures []client.Object
			for i, item := range objects {
				if i != index {
					fixtures = append(fixtures, item.DeepCopyObject().(client.Object))
				}
			}
			c := fake.NewClientBuilder().WithScheme(scheme).WithObjects(fixtures...).Build()
			if err := servingReadiness(ctx, c, GatewayRef{}, srr, "route"); err == nil {
				t.Fatal("missing dependency was accepted for cutover")
			}
		})
	}
	for _, field := range []string{"gateway", "listener", "route"} {
		t.Run("unobserved "+field, func(t *testing.T) {
			var fixtures []client.Object
			for _, item := range objects {
				fixtures = append(fixtures, item.DeepCopyObject().(client.Object))
			}
			switch field {
			case "gateway":
				fixtures[0].(*unstructured.Unstructured).Object["status"] = nil
			case "listener":
				unstructured.RemoveNestedField(fixtures[0].(*unstructured.Unstructured).Object, "status", "listeners")
			case "route":
				fixtures[1].(*unstructured.Unstructured).Object["status"] = nil
			}
			c := fake.NewClientBuilder().WithScheme(scheme).WithObjects(fixtures...).Build()
			if err := servingReadiness(ctx, c, GatewayRef{}, srr, "route"); err == nil {
				t.Fatal("unobserved dependency was accepted for cutover")
			}
		})
	}
	t.Run("protected entry requires the separate authorization channel", func(t *testing.T) {
		protected := srr.DeepCopy()
		protected.Spec.EntranceClass = srrv1alpha1.EntranceClassApplication
		protected.Annotations = map[string]string{gateway.AnnotationAuthMode: gateway.AuthModeExtAuth, gateway.AnnotationExtAuthUser: "alice"}
		fixtures := make([]client.Object, 0, len(objects))
		for _, item := range objects {
			fixtures = append(fixtures, item.DeepCopyObject().(client.Object))
		}
		protectedRoute := fixtures[1].(*unstructured.Unstructured)
		parents := parentRefsForSRR(GatewayRef{}, protected)
		if err := unstructured.SetNestedSlice(protectedRoute.Object, parents, "spec", "parentRefs"); err != nil {
			t.Fatal(err)
		}
		protectedPolicy := desiredApplicationExtAuthSecurityPolicy(protected)
		protectedPolicy.Object["status"] = security.DeepCopy().Object["status"]
		protectedPolicy.SetGeneration(security.GetGeneration())
		fixtures[2] = protectedPolicy
		check := func(message string) {
			t.Helper()
			c := fake.NewClientBuilder().WithScheme(scheme).WithObjects(fixtures...).Build()
			err := servingReadiness(ctx, c, GatewayRef{}, protected, "route")
			if err == nil || !strings.Contains(err.Error(), message) {
				t.Fatalf("readiness = %v, expected %q", err, message)
			}
		}
		if err := unstructured.SetNestedSlice(protectedRoute.Object, append(slices.Clone(parents), map[string]any{"sectionName": "https"}), "spec", "parentRefs"); err != nil {
			t.Fatal(err)
		}
		check("must only attach to Entry mTLS")
		if err := unstructured.SetNestedSlice(protectedRoute.Object, parents, "spec", "parentRefs"); err != nil {
			t.Fatal(err)
		}
		check("authorization listener is not programmed")
		newGateway := fixtures[0].(*unstructured.Unstructured)
		listeners, _, _ := unstructured.NestedSlice(newGateway.Object, "status", "listeners")
		listeners = append(listeners, map[string]any{"name": l4AuthzSection, "conditions": cond("Programmed", "ResolvedRefs")})
		if err := unstructured.SetNestedSlice(newGateway.Object, listeners, "status", "listeners"); err != nil {
			t.Fatal(err)
		}
		check("app-gateway-l4-authz-mtls")
		authRef := map[string]any{"name": defaultGatewayName, "namespace": defaultGatewayNS, "sectionName": l4AuthzSection}
		authPolicy := entryPolicy(ctp.GroupVersionKind(), "app-gateway-l4-authz-mtls", l4AuthzSection,
			map[string]any{"ancestors": []any{map[string]any{"ancestorRef": authRef, "controllerName": owner, "conditions": cond("Accepted")}}})
		fixtures = append(fixtures, authPolicy)
		modePath := []string{"spec", "tls", "clientValidation", "mode"}
		if err := unstructured.SetNestedField(authPolicy.Object, "AllowInsecure", modePath...); err != nil {
			t.Fatal(err)
		}
		check("does not enforce the expected Entry client identity")
		if err := unstructured.SetNestedField(authPolicy.Object, "RequireAndVerify", modePath...); err != nil {
			t.Fatal(err)
		}
		check(".l4-authz")
		authRoute := obj(route.GroupVersionKind(), protected.Namespace, l4AuthzRoute(protected).GetName(), map[string]any{})
		fixtures = append(fixtures, authRoute)
		check("authorization route is not accepted")
		if err := unstructured.SetNestedSlice(authRoute.Object,
			[]any{map[string]any{"parentRef": authRef, "controllerName": owner, "conditions": cond("Accepted", "ResolvedRefs")}},
			"status", "parents"); err != nil {
			t.Fatal(err)
		}
		check("9444")
	})
	t.Run("ready native proxy permits cutover", func(t *testing.T) {
		fixtures := make([]client.Object, 0, len(objects))
		for _, item := range objects {
			fixtures = append(fixtures, item.DeepCopyObject().(client.Object))
		}
		pod := fixtures[len(fixtures)-2].(*corev1.Pod)
		proxy := pod.Spec.Containers[0]
		proxy.RestartPolicy = ptr.To(corev1.ContainerRestartPolicyAlways)
		pod.Spec.InitContainers = append(pod.Spec.InitContainers, proxy)
		pod.Spec.Containers = nil
		pod.Status.InitContainerStatuses = pod.Status.ContainerStatuses
		pod.Status.ContainerStatuses = nil
		c := fake.NewClientBuilder().WithScheme(scheme).WithObjects(fixtures...).Build()
		if err := servingReadiness(ctx, c, GatewayRef{}, srr, "route"); err != nil {
			t.Fatalf("ready native sidecar rejected at cutover: %v", err)
		}
	})
	t.Run("Desktop requires the accepted fixed-path authorizer", func(t *testing.T) {
		desktop := srr.DeepCopy()
		desktop.SetName("sys-desktop")
		desktop.SetNamespace("user-space-alice")
		desktop.SetUID("desktop-entry")
		desktop.SetLabels(map[string]string{
			ManagedByLabel: ManagedByValue, InstanceLabel: gateway.SystemPlaneInstance,
			gateway.LabelSystemPlane: gateway.LabelSystemPlaneValue,
		})
		desktop.SetAnnotations(map[string]string{gateway.AnnotationExtAuthUser: "alice"})
		desktop.Spec.EntranceClass = srrv1alpha1.EntranceClassApplication
		desktopRoute := route.DeepCopy()
		desktopRoute.SetName(desktop.Name)
		desktopRoute.SetNamespace(desktop.Namespace)
		if err := unstructured.SetNestedSlice(desktopRoute.Object, parentRefsForSRR(GatewayRef{}, desktop), "spec", "parentRefs"); err != nil {
			t.Fatal(err)
		}
		policy := desiredApplicationExtAuthSecurityPolicy(desktop)
		policy.Object["status"] = security.DeepCopy().Object["status"]
		policy.SetGeneration(security.GetGeneration())
		backends, _, err := unstructured.NestedSlice(policy.Object, "spec", "extAuth", "http", "backendRefs")
		if err != nil || len(backends) != 1 {
			t.Fatalf("expected one Desktop authorization backend: %v, %v", backends, err)
		}
		setWeight := func(weight int64) {
			t.Helper()
			backends[0].(map[string]any)["weight"] = weight
			if err := unstructured.SetNestedSlice(policy.Object, backends, "spec", "extAuth", "http", "backendRefs"); err != nil {
				t.Fatal(err)
			}
		}
		// A fake client does not apply the SecurityPolicy CRD's backend weight default.
		setWeight(1)
		check := func(want string) {
			t.Helper()
			c := fake.NewClientBuilder().WithScheme(scheme).WithObjects(gw, desktopRoute, policy, ctp).Build()
			err := servingReadiness(ctx, c, GatewayRef{}, desktop, desktop.Name)
			if err == nil || !strings.Contains(err.Error(), want) {
				t.Fatalf("readiness = %v, want %q", err, want)
			}
		}
		check("authorization listener is not programmed")
		setWeight(0)
		check("does not enforce the expected external authorization")
		setWeight(1)
		unstructured.RemoveNestedField(policy.Object, "spec", "extAuth", "http", "pathOverride")
		if err := unstructured.SetNestedField(policy.Object, "/api/verify", "spec", "extAuth", "http", "path"); err != nil {
			t.Fatal(err)
		}
		check("does not enforce the expected external authorization")
		policy.SetGeneration(policy.GetGeneration() + 1)
		check("has no accepted status")

		desktop.Annotations[gateway.AnnotationAuthMode] = gateway.AuthModeNone
		setOwnerSRR(desktopRoute, desktop)
		authRoute := l4AuthzRoute(desktop)
		setOwnerSRR(authRoute, desktop)
		c := fake.NewClientBuilder().WithScheme(scheme).WithObjects(desktopRoute, authRoute).Build()
		res, err := reconcileGatewayMode(ctx, c, GatewayRef{}, desktop)
		if err != nil || res.Status != metav1.ConditionFalse || res.Reason != ReasonInvalidSpec {
			t.Fatalf("Desktop without external authorization was accepted: %+v, %v", res, err)
		}
		for _, obj := range []*unstructured.Unstructured{desktopRoute, authRoute} {
			if err := c.Get(ctx, client.ObjectKeyFromObject(obj), obj); !apierrors.IsNotFound(err) {
				t.Fatalf("route %s was not withdrawn: %v", obj.GetName(), err)
			}
		}
	})
	for _, tc := range []struct {
		name   string
		mutate func(pod *corev1.Pod, slice *discoveryv1.EndpointSlice)
	}{
		{name: "ready"},
		{name: "proxy not ready", mutate: func(p *corev1.Pod, _ *discoveryv1.EndpointSlice) {
			p.Status.ContainerStatuses[0].Ready = false
		}},
		{name: "annotation drift", mutate: func(p *corev1.Pod, _ *discoveryv1.EndpointSlice) {
			p.Annotations[mesh.LinkerdDefaultInboundPolicyAnnotation] = mesh.InboundPolicyAllUnauthenticated
		}},
		{name: "proxy configured permissive", mutate: func(p *corev1.Pod, _ *discoveryv1.EndpointSlice) {
			p.Spec.Containers[0].Env[0].Value = mesh.InboundPolicyAllUnauthenticated
		}},
		{name: "proxy interception bypass", mutate: func(p *corev1.Pod, _ *discoveryv1.EndpointSlice) {
			p.Annotations[mesh.LinkerdSkipSubnetsAnnotation] = "10.0.1.0/24"
		}},
		{name: "old interception exclusions", mutate: func(p *corev1.Pod, _ *discoveryv1.EndpointSlice) {
			p.Spec.InitContainers[0].Args[1] += ",4568"
		}},
		{name: "HTTP route to a reserved plaintext port", mutate: func(_ *corev1.Pod, sl *discoveryv1.EndpointSlice) {
			sl.Ports[0].Port = ptr.To(int32(3389))
		}},
		{name: "forged slice address", mutate: func(_ *corev1.Pod, sl *discoveryv1.EndpointSlice) {
			sl.Endpoints[0].Addresses = []string{"10.0.1.3"}
		}},
		{name: "unowned slice", mutate: func(_ *corev1.Pod, sl *discoveryv1.EndpointSlice) {
			sl.OwnerReferences = nil
		}},
	} {
		t.Run(tc.name, func(t *testing.T) {
			fixtures := make([]client.Object, 0, len(objects))
			for _, object := range objects {
				fixtures = append(fixtures, object.DeepCopyObject().(client.Object))
			}
			if tc.mutate != nil {
				tc.mutate(fixtures[len(fixtures)-2].(*corev1.Pod), fixtures[len(fixtures)-1].(*discoveryv1.EndpointSlice))
			}
			c := fake.NewClientBuilder().WithScheme(scheme).WithObjects(fixtures...).Build()
			if err := servingReadiness(ctx, c, GatewayRef{}, srr, "route"); (err == nil) != (tc.mutate == nil) {
				t.Fatalf("unexpected servingReadiness result: %v", err)
			}
		})
	}
}

func defaultDenyAuthProxyFixture(objects []client.Object, namespace string) *unstructured.Unstructured {
	for _, object := range objects {
		if pod, ok := object.(*corev1.Pod); ok && pod.Namespace == namespace {
			pod.Labels["app"] = "systemserver"
			pod.Annotations[mesh.LinkerdDefaultInboundPolicyAnnotation] = "deny"
			pod.Spec.Containers[0].Env[0].Value = "deny"
			pod.Spec.Containers = append(pod.Spec.Containers, corev1.Container{
				Name: "proxy", Ports: []corev1.ContainerPort{{Name: "auth-proxy", ContainerPort: 9091, Protocol: corev1.ProtocolTCP}},
			})
		}
	}
	return &unstructured.Unstructured{Object: map[string]any{
		"apiVersion": "policy.linkerd.io/v1beta3",
		"kind":       "Server",
		"metadata":   map[string]any{"name": "systemserver-existing-9091", "namespace": namespace},
		"spec": map[string]any{
			"port": int64(9091), "accessPolicy": "cluster-authenticated",
			"podSelector": map[string]any{"matchLabels": map[string]any{"app": "systemserver"}},
		},
	}}
}

func TestAuthProxyMeshReadinessChecksEndpointPort(t *testing.T) {
	const namespace = "user-system-alice"
	for _, tc := range []struct {
		name   string
		change func(*corev1.Pod, *discoveryv1.EndpointSlice, *unstructured.Unstructured)
		want   string
	}{
		{name: "Service port differs from protected target"},
		{name: "wrong target", change: func(_ *corev1.Pod, ep *discoveryv1.EndpointSlice, _ *unstructured.Unstructured) {
			ep.Ports[0].Port = ptr.To(int32(80))
		}, want: "unsupported inbound policy"},
		{name: "invalid endpoint port", change: func(_ *corev1.Pod, ep *discoveryv1.EndpointSlice, _ *unstructured.Unstructured) {
			ep.Ports[0].Port = ptr.To(int32(65536))
		}, want: "without an intercepted HTTP target port"},
		{name: "wrong server port", change: func(_ *corev1.Pod, _ *discoveryv1.EndpointSlice, server *unstructured.Unstructured) {
			server.Object["spec"].(map[string]any)["port"] = int64(80)
		}, want: "does not enforce"},
		{name: "plaintext allowed", change: func(_ *corev1.Pod, _ *discoveryv1.EndpointSlice, server *unstructured.Unstructured) {
			server.Object["spec"].(map[string]any)["accessPolicy"] = "all-unauthenticated"
		}, want: "does not enforce"},
		{name: "different pods", change: func(_ *corev1.Pod, _ *discoveryv1.EndpointSlice, server *unstructured.Unstructured) {
			server.Object["spec"].(map[string]any)["podSelector"] = map[string]any{"matchLabels": map[string]any{"app": "other"}}
		}, want: "does not enforce"},
		{name: "proxy default drift", change: func(pod *corev1.Pod, _ *discoveryv1.EndpointSlice, _ *unstructured.Unstructured) {
			pod.Spec.Containers[0].Env[0].Value = mesh.InboundPolicyClusterAuthenticated
		}, want: "proxy policy does not match"},
		{name: "undeclared target", change: func(pod *corev1.Pod, _ *discoveryv1.EndpointSlice, _ *unstructured.Unstructured) {
			pod.Spec.Containers[1].Ports = nil
		}, want: "undeclared TCP port"},
	} {
		t.Run(tc.name, func(t *testing.T) {
			objects := meshedServiceObjects(&corev1.Service{
				ObjectMeta: metav1.ObjectMeta{Namespace: namespace, Name: AutheliaBackendServiceName},
				Spec: corev1.ServiceSpec{Selector: map[string]string{"app": "systemserver"},
					Ports: []corev1.ServicePort{{Name: "http", Port: 8080, TargetPort: intstr.FromInt32(9091)}}},
			})
			server := defaultDenyAuthProxyFixture(objects, namespace)
			endpoint := objects[2].(*discoveryv1.EndpointSlice)
			endpoint.Ports[0].Port = ptr.To(int32(9091))
			if tc.change != nil {
				tc.change(objects[1].(*corev1.Pod), endpoint, server)
			}
			c := fake.NewClientBuilder().WithScheme(testScheme(t)).WithObjects(append(objects, server)...).Build()
			err := meshedReadyEndpoints(t.Context(), c, namespace, srrv1alpha1.UpstreamRef{ServiceName: AutheliaBackendServiceName, Port: 8080})
			if tc.want == "" && err != nil || tc.want != "" && (err == nil || !strings.Contains(err.Error(), tc.want)) {
				t.Fatalf("readiness = %v, want %q", err, tc.want)
			}
		})
	}
}
func TestWizardReadinessRequiresOwnMeshedAuthProxy(t *testing.T) {
	ctx := t.Context()
	c := fake.NewClientBuilder().WithScheme(testScheme(t)).Build()
	for _, tenant := range []string{"alice", "bob"} {
		if err := gateway.EnsureSystemPlaneRoutes(ctx, c, tenant, "owner.olares.com", "user-space-"+tenant, true); err != nil {
			t.Fatal(err)
		}
	}
	create := func(ns, name, app string, port int32, contextVersion bool) {
		t.Helper()
		svc := &corev1.Service{ObjectMeta: metav1.ObjectMeta{Namespace: ns, Name: name},
			Spec: corev1.ServiceSpec{Selector: map[string]string{"app": app}, Ports: []corev1.ServicePort{{Port: port}}}}
		for _, obj := range meshedServiceObjects(svc) {
			if pod, ok := obj.(*corev1.Pod); ok && contextVersion {
				pod.Annotations["traffic.olares.io/auth-proxy-context-version"] = "1"
			}
			if err := c.Create(ctx, obj); err != nil {
				t.Fatal(err)
			}
		}
	}
	create("os-framework", "authelia-backend", "authelia-backend", 9091, false)
	create("user-system-alice", "authelia-backend", "systemserver", 9091, true)
	create("user-space-alice", "wizard", "wizard", 80, false)
	create("user-space-bob", "wizard", "wizard", 80, false)
	check := func(tenant string, wantReady bool) {
		t.Helper()
		var route srrv1alpha1.SharedRouteRegistry
		if err := c.Get(ctx, client.ObjectKey{Namespace: "user-space-" + tenant, Name: "sys-wizard"}, &route); err != nil {
			t.Fatal(err)
		}
		if err := httpRouteMeshDependenciesReady(ctx, c, route.Namespace, &route); (err == nil) != wantReady {
			t.Fatalf("wizard readiness for %s = %v; want ready=%t", tenant, err, wantReady)
		}
	}
	check("alice", true)
	check("bob", false)
	create("user-system-bob", "authelia-backend", "systemserver", 9091, false)
	check("bob", false)
	var pod corev1.Pod
	if err := c.Get(ctx, client.ObjectKey{Namespace: "user-system-bob", Name: "authelia-backend"}, &pod); err != nil {
		t.Fatal(err)
	}
	pod.Annotations["traffic.olares.io/auth-proxy-context-version"] = "1"
	init := pod.Spec.InitContainers
	pod.Spec.InitContainers = nil
	if err := c.Update(ctx, &pod); err != nil {
		t.Fatal(err)
	}
	check("bob", false)
	pod.Spec.InitContainers = init
	if err := c.Update(ctx, &pod); err != nil {
		t.Fatal(err)
	}
	check("bob", true)
}
