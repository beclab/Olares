package routecontrol

import (
	"context"
	"testing"

	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"k8s.io/apimachinery/pkg/types"
	clientgoscheme "k8s.io/client-go/kubernetes/scheme"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/client/fake"
	"sigs.k8s.io/controller-runtime/pkg/reconcile"

	"github.com/beclab/Olares/framework/app-service/pkg/cluster"
)

// --- pure listener-building cases (TC-01/02/03/06/07/08/09/10) ---

func TestBuildDesiredListeners(t *testing.T) {
	cl := func(name, viewer string) currentListener {
		return currentListener{name: name, certViewer: viewer}
	}
	type want struct {
		name     string
		hostname string
		viewer   string
	}
	cases := []struct {
		name    string
		current []currentListener
		viewers []string
		want    []want
	}{
		{
			name:    "TC-01 first viewer adopts https as per-hostname",
			current: []currentListener{cl("http", ""), cl("https", "")},
			viewers: []string{"alice"},
			want:    []want{{"https", "*.alice.example.com", "alice"}},
		},
		{
			name:    "TC-02 second viewer appends dedicated listener, https sticky",
			current: []currentListener{cl("http", ""), cl("https", "alice")},
			viewers: []string{"alice", "bob"},
			want: []want{
				{"https", "*.alice.example.com", "alice"},
				{"https-bob", "*.bob.example.com", "bob"},
			},
		},
		{
			name:    "TC-03 rotation keeps both listeners structurally stable",
			current: []currentListener{cl("https", "alice"), cl("https-bob", "bob")},
			viewers: []string{"alice", "bob"},
			want: []want{
				{"https", "*.alice.example.com", "alice"},
				{"https-bob", "*.bob.example.com", "bob"},
			},
		},
		{
			name:    "TC-06 deleting bob removes its listener, alice retained",
			current: []currentListener{cl("https", "alice"), cl("https-bob", "bob")},
			viewers: []string{"alice"},
			want:    []want{{"https", "*.alice.example.com", "alice"}},
		},
		{
			name:    "TC-09 zero viewers empties https certRef (no placeholder)",
			current: []currentListener{cl("https", "alice")},
			viewers: []string{},
			want:    []want{{"https", "", ""}},
		},
		{
			name:    "TC-09b deleting https owner empties https, dedicated viewer kept",
			current: []currentListener{cl("https", "alice"), cl("https-bob", "bob")},
			viewers: []string{"bob"},
			want: []want{
				{"https", "", ""},
				{"https-bob", "*.bob.example.com", "bob"},
			},
		},
		{
			name:    "TC-07 free https adopts lexicographically-first unassigned",
			current: []currentListener{cl("http", ""), cl("https", "")},
			viewers: []string{"bob", "alice"},
			want: []want{
				{"https", "*.alice.example.com", "alice"},
				{"https-bob", "*.bob.example.com", "bob"},
			},
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			got := buildDesiredListeners(tc.current, tc.viewers, testPlatformDomain)
			if len(got) != len(tc.want) {
				t.Fatalf("listener count = %d, want %d (%+v)", len(got), len(tc.want), got)
			}
			for i, w := range tc.want {
				if got[i].Name != w.name || got[i].Hostname != w.hostname || got[i].Viewer != w.viewer {
					t.Fatalf("listener[%d] = %+v, want {%s %s %s}", i, got[i], w.name, w.hostname, w.viewer)
				}
			}
		})
	}
}

// TC-10: sanitisation yields valid, collision-free SectionNames and is stable.
func TestListenerSectionForViewer(t *testing.T) {
	if got := listenerSectionForViewer("bob"); got != "bob" {
		t.Fatalf("clean viewer section = %q, want bob", got)
	}
	// Two distinct viewers that sanitise to the same base must not collide.
	a := listenerSectionForViewer("a_b")
	b := listenerSectionForViewer("a.b")
	if a == b {
		t.Fatalf("sanitised collision: %q == %q", a, b)
	}
	if !dns1123SectionRE.MatchString(a) || !dns1123SectionRE.MatchString(b) {
		t.Fatalf("sanitised sections not DNS-1123: %q %q", a, b)
	}
	// Stability: same input → same output regardless of call site.
	if listenerSectionForViewer("a_b") != a {
		t.Fatalf("section not stable for a_b")
	}
}

// TC-08: each viewer maps to its own certRef Secret (SNI selection isolation).
func TestManagedListenerCertRefIsolation(t *testing.T) {
	got := buildDesiredListeners(
		[]currentListener{{name: "https", certViewer: "alice"}},
		[]string{"alice", "bob"}, testPlatformDomain)
	refs := map[string]string{}
	for _, l := range got {
		refs[l.Viewer] = l.toMap()["tls"].(map[string]any)["certificateRefs"].([]any)[0].(map[string]any)["name"].(string)
	}
	if refs["alice"] != "shared-entrance-tls-alice" || refs["bob"] != "shared-entrance-tls-bob" {
		t.Fatalf("certRef isolation broken: %+v", refs)
	}
}

// --- reconcile-level cases (TC-01 e2e / TC-04 / TC-05 / TC-11 / TC-12) ---

// primeSnapshot sets the process-global ClusterConfig snapshot for one test and
// restores the fallback-equivalent default afterwards so the shared cache does
// not leak meshProfile/platformDomain into sibling tests in this package.
func primeSnapshot(t *testing.T, snap cluster.Snapshot) {
	t.Helper()
	cluster.PrimeSnapshotForTest(snap)
	t.Cleanup(func() {
		cluster.PrimeSnapshotForTest(cluster.Snapshot{
			PlatformDomain:          cluster.DefaultPlatformDomain,
			InClusterGatewayEnabled: true,
			MeshProfile:             cluster.MeshProfileFull,
		})
	})
}

func TestEntranceTLSListenerReconcile_liteAppliesPerHostname(t *testing.T) {
	primeSnapshot(t, cluster.Snapshot{MeshProfile: cluster.MeshProfileLite, PlatformDomain: testPlatformDomain})
	c := listenerFakeClient(t, gatewayWithListeners(httpListener(), httpsPlaceholder()), entranceSecret("alice"))
	r := &EntranceTLSListenerReconciler{Client: c}

	if _, err := r.Reconcile(context.Background(), reconcile.Request{}); err != nil {
		t.Fatalf("reconcile: %v", err)
	}

	gw := getGateway(t, c)
	https := listenerByName(t, gw, "https")
	if got := certRefName(t, https); got != "shared-entrance-tls-alice" {
		t.Fatalf("https certRef = %q, want shared-entrance-tls-alice", got)
	}
	if got, _ := https["hostname"].(string); got != "*.alice.example.com" {
		t.Fatalf("https hostname = %q, want *.alice.example.com", got)
	}
}

func TestEntranceTLSListenerReconcile_fullProfileNoOp(t *testing.T) {
	primeSnapshot(t, cluster.Snapshot{MeshProfile: cluster.MeshProfileFull, PlatformDomain: testPlatformDomain})
	c := listenerFakeClient(t, gatewayWithListeners(httpListener(), httpsPlaceholder()), entranceSecret("alice"))
	r := &EntranceTLSListenerReconciler{Client: c}

	if _, err := r.Reconcile(context.Background(), reconcile.Request{}); err != nil {
		t.Fatalf("reconcile: %v", err)
	}
	gw := getGateway(t, c)
	if got := certRefName(t, listenerByName(t, gw, "https")); got != "app-gateway-tls" {
		t.Fatalf("full profile mutated https certRef to %q", got)
	}
}

func TestEntranceTLSListenerReconcile_absentMeshProfileNoOp(t *testing.T) {
	primeSnapshot(t, cluster.Snapshot{MeshProfile: "", PlatformDomain: testPlatformDomain})
	c := listenerFakeClient(t, gatewayWithListeners(httpListener(), httpsPlaceholder()), entranceSecret("alice"))
	r := &EntranceTLSListenerReconciler{Client: c}

	if _, err := r.Reconcile(context.Background(), reconcile.Request{}); err != nil {
		t.Fatalf("reconcile: %v", err)
	}
	gw := getGateway(t, c)
	if got := certRefName(t, listenerByName(t, gw, "https")); got != "app-gateway-tls" {
		t.Fatalf("absent profile mutated https certRef to %q", got)
	}
}

func TestEntranceTLSListenerReconcile_emptyPlatformDomainFailClosed(t *testing.T) {
	primeSnapshot(t, cluster.Snapshot{MeshProfile: cluster.MeshProfileLite, PlatformDomain: ""})
	c := listenerFakeClient(t, gatewayWithListeners(httpListener(), httpsPlaceholder()), entranceSecret("alice"))
	r := &EntranceTLSListenerReconciler{Client: c}

	res, err := r.Reconcile(context.Background(), reconcile.Request{})
	if err != nil {
		t.Fatalf("reconcile: %v", err)
	}
	if !res.Requeue {
		t.Fatal("expected requeue on empty platformDomain")
	}
	gw := getGateway(t, c)
	if got := certRefName(t, listenerByName(t, gw, "https")); got != "app-gateway-tls" {
		t.Fatalf("fail-closed still mutated https certRef to %q", got)
	}
}

func TestEntranceTLSListenerReconcile_reconvergeAfterHelmOverwrite(t *testing.T) {
	primeSnapshot(t, cluster.Snapshot{MeshProfile: cluster.MeshProfileLite, PlatformDomain: testPlatformDomain})
	// Helm upgrade reverted the Gateway: https back to placeholder, https-bob gone.
	c := listenerFakeClient(t,
		gatewayWithListeners(httpListener(), httpsPlaceholder()),
		entranceSecret("alice"), entranceSecret("bob"))
	r := &EntranceTLSListenerReconciler{Client: c}

	if _, err := r.Reconcile(context.Background(), reconcile.Request{}); err != nil {
		t.Fatalf("reconcile: %v", err)
	}

	gw := getGateway(t, c)
	if got := certRefName(t, listenerByName(t, gw, "https")); got != "shared-entrance-tls-alice" {
		t.Fatalf("https certRef = %q, want shared-entrance-tls-alice", got)
	}
	bob := listenerByName(t, gw, "https-bob")
	if got := certRefName(t, bob); got != "shared-entrance-tls-bob" {
		t.Fatalf("https-bob certRef = %q, want shared-entrance-tls-bob", got)
	}
	if got, _ := bob["hostname"].(string); got != "*.bob.example.com" {
		t.Fatalf("https-bob hostname = %q", got)
	}
}

// --- test fixtures ---

func listenerFakeClient(t *testing.T, objs ...client.Object) client.Client {
	t.Helper()
	scheme := runtime.NewScheme()
	if err := clientgoscheme.AddToScheme(scheme); err != nil {
		t.Fatalf("clientgo scheme: %v", err)
	}
	gwGVK := gatewayGVK()
	scheme.AddKnownTypeWithName(gwGVK, &unstructured.Unstructured{})
	scheme.AddKnownTypeWithName(schema.GroupVersionKind{
		Group: gwGVK.Group, Version: gwGVK.Version, Kind: gwGVK.Kind + "List",
	}, &unstructured.UnstructuredList{})
	return fake.NewClientBuilder().WithScheme(scheme).WithObjects(objs...).Build()
}

func gatewayWithListeners(listeners ...map[string]any) *unstructured.Unstructured {
	items := make([]any, 0, len(listeners))
	for _, l := range listeners {
		items = append(items, l)
	}
	u := &unstructured.Unstructured{Object: map[string]any{
		"apiVersion": gatewayAPIGroup + "/" + gatewayAPIVersion,
		"kind":       gatewayAPIKind,
		"metadata": map[string]any{
			"name":      defaultGatewayName,
			"namespace": defaultGatewayNS,
		},
		"spec": map[string]any{
			"gatewayClassName": "olares-app-gateway",
			"listeners":        items,
		},
	}}
	u.SetGroupVersionKind(gatewayGVK())
	return u
}

func httpListener() map[string]any {
	return map[string]any{
		"name": "http", "protocol": "HTTP", "port": int64(80),
		"allowedRoutes": map[string]any{"namespaces": map[string]any{"from": "All"}},
	}
}

func httpsPlaceholder() map[string]any {
	return map[string]any{
		"name": "https", "protocol": "HTTPS", "port": int64(443),
		"allowedRoutes": map[string]any{"namespaces": map[string]any{"from": "All"}},
		"tls": map[string]any{
			"mode": "Terminate",
			"certificateRefs": []any{
				map[string]any{"kind": "Secret", "group": "", "name": "app-gateway-tls"},
			},
		},
	}
}

func entranceSecret(viewer string) *corev1.Secret {
	return &corev1.Secret{
		ObjectMeta: metav1.ObjectMeta{
			Name:      entranceTLSSecretName(viewer),
			Namespace: defaultGatewayNS,
			Labels:    map[string]string{labelTLSViewer: viewer},
		},
		Type: corev1.SecretTypeTLS,
	}
}

func getGateway(t *testing.T, c client.Client) *unstructured.Unstructured {
	t.Helper()
	gw := newGatewayUnstructured()
	if err := c.Get(context.Background(), types.NamespacedName{Namespace: defaultGatewayNS, Name: defaultGatewayName}, gw); err != nil {
		t.Fatalf("get gateway: %v", err)
	}
	return gw
}

func listenerByName(t *testing.T, gw *unstructured.Unstructured, name string) map[string]any {
	t.Helper()
	listeners, _, _ := unstructured.NestedSlice(gw.Object, "spec", "listeners")
	for _, item := range listeners {
		l, ok := item.(map[string]any)
		if !ok {
			continue
		}
		if n, _ := l["name"].(string); n == name {
			return l
		}
	}
	t.Fatalf("listener %q not found in %+v", name, listeners)
	return nil
}

func certRefName(t *testing.T, listener map[string]any) string {
	t.Helper()
	tls, ok := listener["tls"].(map[string]any)
	if !ok {
		return ""
	}
	refs, ok := tls["certificateRefs"].([]any)
	if !ok || len(refs) == 0 {
		return ""
	}
	ref, _ := refs[0].(map[string]any)
	name, _ := ref["name"].(string)
	return name
}
