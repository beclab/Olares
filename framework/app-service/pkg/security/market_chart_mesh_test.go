package security

import (
	"reflect"
	"regexp"
	"slices"
	"testing"

	corev1 "k8s.io/api/core/v1"
	netv1 "k8s.io/api/networking/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/labels"
	"k8s.io/apimachinery/pkg/types"
	"k8s.io/apimachinery/pkg/util/intstr"
	"sigs.k8s.io/yaml"
)

func TestMarketChartMeshAuthorizesOnlyInstallerReads(t *testing.T) {
	resources, err := renderAppServiceMeshManifests(t, map[string]any{
		"marketBackendMTLS": map[string]any{"prepare": true, "tenants": []string{"alice", "bob"}},
	})
	if err != nil {
		t.Fatal(err)
	}
	server := resources["Server/market-chart-repository"]["spec"].(map[string]any)
	marketSelector := map[string]any{"matchLabels": map[string]any{
		"app": "appstore", "traffic.olares.io/market-backend-mtls": "required",
	}}
	if server["port"] != 82 || server["accessPolicy"] != "deny" || server["proxyProtocol"] != "HTTP/1" ||
		!reflect.DeepEqual(server["podSelector"], marketSelector) {
		t.Fatalf("invalid Chart Server: %v", server)
	}
	service := resources["Service/appstore-service"]["spec"].(map[string]any)
	if !slices.ContainsFunc(service["ports"].([]any), func(port any) bool {
		return reflect.DeepEqual(port, map[string]any{"name": "chart-repo", "port": 82, "targetPort": 82, "protocol": "TCP"})
	}) {
		t.Fatalf("Chart Service must keep port 82 mapped to the real listener: %v", service)
	}
	route := resources["HTTPRoute/market-chart-read"]["spec"].(map[string]any)
	match := func(pathType, path string) map[string]any {
		return map[string]any{"method": "GET", "path": map[string]any{"type": pathType, "value": path}}
	}
	archivePattern := `^/charts/[A-Za-z0-9][A-Za-z0-9._+-]*\.tgz$`
	if !reflect.DeepEqual(route["parentRefs"], []any{map[string]any{
		"group": "policy.linkerd.io", "kind": "Server", "name": "market-chart-repository",
	}}) || !reflect.DeepEqual(route["rules"], []any{map[string]any{
		"matches": []any{match("Exact", "/static-index.yaml"), match("RegularExpression", archivePattern)},
	}}) {
		t.Fatalf("Chart route expanded beyond installer GETs: %v", route)
	}
	pattern := regexp.MustCompile(archivePattern)
	for _, path := range []string{"/charts/notes-1.0.0.tgz", "/charts/app-1.0.0-rc.1+build.tgz"} {
		if !pattern.MatchString(path) {
			t.Errorf("Chart download did not match %s", path)
		}
	}
	for _, path := range []string{"/charts/", "/charts/a.tgz/extra", "/charts/../a.tgz", "/charts/a%2Fb.tgz", "/api/v1/charts/a.tgz"} {
		if pattern.MatchString(path) {
			t.Errorf("Chart route admitted %s", path)
		}
	}
	authn := resources["MeshTLSAuthentication/market-chart-read"]["spec"].(map[string]any)
	authz := resources["AuthorizationPolicy/market-chart-read"]["spec"].(map[string]any)
	if !reflect.DeepEqual(authn, map[string]any{"identityRefs": []any{map[string]any{
		"kind": "ServiceAccount", "name": "os-internal", "namespace": "os-framework",
	}}}) || !reflect.DeepEqual(authz, map[string]any{
		"targetRef": map[string]any{"group": "gateway.networking.k8s.io", "kind": "HTTPRoute", "name": "market-chart-read"},
		"requiredAuthenticationRefs": []any{map[string]any{
			"group": "policy.linkerd.io", "kind": "MeshTLSAuthentication", "name": "market-chart-read",
		}},
	}) {
		t.Fatalf("expected only the app-service mesh identity: %v %v", authn, authz)
	}
	if resources["AuthorizationPolicy/market-legacy-provider"] != nil {
		t.Fatal("legacy provider must not receive global Market authorization")
	}

	var policy netv1.NetworkPolicy
	encoded, err := yaml.Marshal(resources["NetworkPolicy/market-backend-mtls-ingress"])
	if err != nil {
		t.Fatal(err)
	}
	// Kubernetes structs declare JSON names for their fields.
	if err := yaml.UnmarshalStrict(encoded, &policy); err != nil {
		t.Fatal(err)
	}
	if !IsAppServiceManagedExternalNP(&policy) || !reflect.DeepEqual(policy.Spec.PolicyTypes, []netv1.PolicyType{netv1.PolicyTypeIngress}) ||
		!reflect.DeepEqual(policy.Spec.PodSelector.MatchLabels, map[string]string{
			"app": "appstore", "traffic.olares.io/market-backend-mtls": "required",
		}) || !reflect.DeepEqual(policy.Spec.Ingress, []netv1.NetworkPolicyIngressRule{
		{From: []netv1.NetworkPolicyPeer{platformPeer("os-framework", appServiceChartRepository.from.selector)}, Ports: []netv1.NetworkPolicyPort{tcpPort(82), tcpPort(4143)}},
		{From: []netv1.NetworkPolicyPeer{platformPeer("user-system-alice", map[string]string{"app": "systemserver", platformMeshControlPlaneLabel: "os-mesh"})}, Ports: []netv1.NetworkPolicyPort{tcpPort(8080), tcpPort(4143)}},
		{From: []netv1.NetworkPolicyPeer{platformPeer("user-system-bob", map[string]string{"app": "systemserver", platformMeshControlPlaneLabel: "os-mesh"})}, Ports: []netv1.NetworkPolicyPort{tcpPort(8080), tcpPort(4143)}},
	}) {
		t.Fatalf("invalid Market CNI guard or Helm ownership: %+v", policy)
	}
	policy.Namespace = "user-system-alice"
	if IsAppServiceManagedExternalNP(&policy) {
		t.Fatal("another namespace cannot claim the Market policy's ownership exemption")
	}
}

func TestStrictMarketIsExcludedFromAdditiveSystemIngressAllows(t *testing.T) {
	strict := labels.Set{"app": "appstore", "traffic.olares.io/market-backend-mtls": "required"}
	legacy := labels.Set{"app": "appstore"}
	policies := append([]netv1.NetworkPolicy{NPOSSystem}, BuildTrafficGroupIsolationBundle("os-framework")...)
	for _, policy := range policies {
		if len(policy.Spec.Ingress) == 0 {
			continue
		}
		selector, err := metav1.LabelSelectorAsSelector(&policy.Spec.PodSelector)
		if err != nil {
			t.Fatal(err)
		}
		if selector.Matches(strict) || !selector.Matches(legacy) {
			t.Errorf("policy %q must exclude strict Market and retain legacy Market", policy.Name)
		}
	}
}

func TestChartRepositoryEgressRequiresVerifiedService(t *testing.T) {
	key := appServiceChartRepository.service
	if !slices.Contains(PlatformTrafficServices(key.Namespace), key) ||
		!reflect.DeepEqual(PlatformTrafficNamespaces(key), []string{key.Namespace}) {
		t.Fatal("Chart Service changes must trigger egress reconciliation")
	}
	service := &corev1.Service{
		ObjectMeta: metav1.ObjectMeta{Name: key.Name, Namespace: key.Namespace},
		Spec: corev1.ServiceSpec{
			Type: corev1.ServiceTypeClusterIP, ClusterIP: "10.96.0.10",
			Selector: map[string]string{"app": "appstore"},
			Ports:    []corev1.ServicePort{{Port: 82, TargetPort: intstr.FromInt32(82), Protocol: corev1.ProtocolTCP}},
		},
	}
	services := map[types.NamespacedName]*corev1.Service{key: service}
	readPolicy := func() netv1.NetworkPolicySpec {
		t.Helper()
		for _, policy := range BuildPlatformTrafficPolicies("os-framework", services) {
			if policy.Name == "olares-platform-app-service-chart-repository-egress" {
				if !IsTrafficGroupManagedNP(&policy) {
					t.Fatal("namespace policy sweep must preserve Chart egress")
				}
				return policy.Spec
			}
		}
		t.Fatal("missing Chart egress policy")
		return netv1.NetworkPolicySpec{}
	}
	spec := readPolicy()
	if !reflect.DeepEqual(spec.PolicyTypes, []netv1.PolicyType{netv1.PolicyTypeEgress}) ||
		!reflect.DeepEqual(spec.PodSelector.MatchLabels, appServiceChartRepository.from.selector) ||
		!reflect.DeepEqual(spec.Egress, []netv1.NetworkPolicyEgressRule{
			{To: []netv1.NetworkPolicyPeer{platformPeer("os-framework", map[string]string{"app": "appstore", platformMeshControlPlaneLabel: "os-mesh"})}, Ports: []netv1.NetworkPolicyPort{tcpPort(82), tcpPort(4143)}},
			{To: []netv1.NetworkPolicyPeer{{IPBlock: &netv1.IPBlock{CIDR: "10.96.0.10/32"}}}, Ports: []netv1.NetworkPolicyPort{tcpPort(82)}},
		}) {
		t.Fatalf("invalid app-service egress: %+v", spec)
	}
	service.Spec.Ports[0].TargetPort = intstr.FromInt32(8080)
	if len(readPolicy().Egress) != 0 {
		t.Fatal("retargeted Chart service retained egress")
	}
	delete(services, key)
	if len(readPolicy().Egress) != 0 {
		t.Fatal("removed Chart service retained egress")
	}
}
