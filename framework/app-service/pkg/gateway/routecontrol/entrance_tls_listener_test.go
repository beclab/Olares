package routecontrol

import (
	"context"
	"strings"
	"testing"

	"github.com/beclab/Olares/framework/app-service/pkg/cluster"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"k8s.io/apimachinery/pkg/types"
	"sigs.k8s.io/controller-runtime/pkg/client/fake"
	"sigs.k8s.io/controller-runtime/pkg/reconcile"
)

func gatewayScheme(t *testing.T) *runtime.Scheme {
	t.Helper()
	s := testScheme(t)
	gw := schema.GroupVersion{Group: "gateway.networking.k8s.io", Version: "v1"}
	s.AddKnownTypeWithName(gw.WithKind("Gateway"), &unstructured.Unstructured{})
	s.AddKnownTypeWithName(gw.WithKind("GatewayList"), &unstructured.UnstructuredList{})
	return s
}

func baseGateway() *unstructured.Unstructured {
	gw := &unstructured.Unstructured{Object: map[string]interface{}{
		"apiVersion": "gateway.networking.k8s.io/v1",
		"kind":       "Gateway",
		"metadata": map[string]interface{}{
			"name":      defaultGatewayName,
			"namespace": defaultGatewayNS,
		},
		"spec": map[string]interface{}{
			"listeners": []interface{}{
				map[string]interface{}{
					"name":     "http",
					"protocol": "HTTP",
					"port":     int64(80),
					"allowedRoutes": map[string]interface{}{
						"namespaces": map[string]interface{}{"from": "All"},
					},
				},
				map[string]interface{}{
					"name":     "https",
					"protocol": "HTTPS",
					"port":     int64(443),
					"allowedRoutes": map[string]interface{}{
						"namespaces": map[string]interface{}{"from": "All"},
					},
					"tls": map[string]interface{}{
						"mode": "Terminate",
						"certificateRefs": []interface{}{
							map[string]interface{}{"kind": "Secret", "name": gatewayTLSSecretName},
						},
					},
				},
			},
		},
	}}
	gw.SetGroupVersionKind(gatewayGVK)
	return gw
}

func viewerSecret(viewer string) *corev1.Secret {
	return &corev1.Secret{
		ObjectMeta: metav1.ObjectMeta{
			Name:      sharedEntranceTLSName(viewer),
			Namespace: defaultGatewayNS,
			Labels: map[string]string{
				ManagedByLabel: ManagedByValue,
				labelTLSViewer: viewer,
			},
		},
		Type: corev1.SecretTypeTLS,
		StringData: map[string]string{
			corev1.TLSCertKey:       "cert",
			corev1.TLSPrivateKeyKey: "key",
		},
	}
}

func TestApplyGatewayHTTPSListeners_TC01_singleViewer(t *testing.T) {
	cluster.PrimePlatformDomainForTest("olares.com")
	t.Cleanup(cluster.ResetPlatformDomainForTest)
	s := gatewayScheme(t)
	gw := baseGateway()
	alice := viewerSecret("alice")
	c := fake.NewClientBuilder().WithScheme(s).WithObjects(gw, alice).Build()

	viewers, err := listViewerTLSSecrets(context.Background(), c)
	if err != nil || len(viewers) != 1 {
		t.Fatalf("list secrets: %v len=%d", err, len(viewers))
	}
	if err := applyGatewayHTTPSListeners(context.Background(), c, viewers, nil, "olares.com"); err != nil {
		t.Fatal(err)
	}
	updated := &unstructured.Unstructured{}
	updated.SetGroupVersionKind(gatewayGVK)
	if err := c.Get(context.Background(), types.NamespacedName{Namespace: defaultGatewayNS, Name: defaultGatewayName}, updated); err != nil {
		t.Fatal(err)
	}
	https := findListener(updated, "https")
	if https == nil {
		t.Fatal("https listener missing")
	}
	host, _, _ := unstructured.NestedString(https, "hostname")
	if host != "*.alice.olares.com" {
		t.Errorf("hostname = %q", host)
	}
	if got := listenerCertRef(https); got != "shared-entrance-tls-alice" {
		t.Errorf("certRef = %q", got)
	}
}

func TestApplyGatewayHTTPSListeners_TC02_twoViewers(t *testing.T) {
	cluster.PrimePlatformDomainForTest("olares.com")
	t.Cleanup(cluster.ResetPlatformDomainForTest)
	s := gatewayScheme(t)
	gw := baseGateway()
	c := fake.NewClientBuilder().WithScheme(s).WithObjects(gw, viewerSecret("alice"), viewerSecret("bob")).Build()
	viewers, err := listViewerTLSSecrets(context.Background(), c)
	if err != nil {
		t.Fatal(err)
	}
	if err := applyGatewayHTTPSListeners(context.Background(), c, viewers, nil, "olares.com"); err != nil {
		t.Fatal(err)
	}
	updated := &unstructured.Unstructured{}
	updated.SetGroupVersionKind(gatewayGVK)
	if err := c.Get(context.Background(), types.NamespacedName{Namespace: defaultGatewayNS, Name: defaultGatewayName}, updated); err != nil {
		t.Fatal(err)
	}
	listeners, _, _ := unstructured.NestedSlice(updated.Object, "spec", "listeners")
	if len(listeners) != 3 {
		t.Fatalf("listener count = %d, want 3 (http + https + https-bob)", len(listeners))
	}
	alice := findListener(updated, "https")
	bob := findListener(updated, "https-bob")
	if alice == nil || bob == nil {
		t.Fatal("expected https and https-bob listeners")
	}
	aliceHost, _, _ := unstructured.NestedString(alice, "hostname")
	bobHost, _, _ := unstructured.NestedString(bob, "hostname")
	bobName, _, _ := unstructured.NestedString(bob, "name")
	if aliceHost != "*.alice.olares.com" || bobHost != "*.bob.olares.com" || bobName != "https-bob" {
		t.Errorf("alice host=%q bob host=%q bob listener=%q", aliceHost, bobHost, bobName)
	}
	if got := listenerCertRef(alice); got != "shared-entrance-tls-alice" {
		t.Errorf("alice certRef = %q", got)
	}
	if got := listenerCertRef(bob); got != "shared-entrance-tls-bob" {
		t.Errorf("bob certRef = %q", got)
	}
}

func TestApplyGatewayHTTPSListeners_TC06_deleteBob(t *testing.T) {
	s := gatewayScheme(t)
	gw := baseGateway()
	alice := viewerSecret("alice")
	c := fake.NewClientBuilder().WithScheme(s).WithObjects(gw, alice).Build()
	viewers, _ := listViewerTLSSecrets(context.Background(), c)
	if err := applyGatewayHTTPSListeners(context.Background(), c, viewers, nil, "olares.com"); err != nil {
		t.Fatal(err)
	}
	updated := &unstructured.Unstructured{}
	updated.SetGroupVersionKind(gatewayGVK)
	_ = c.Get(context.Background(), types.NamespacedName{Namespace: defaultGatewayNS, Name: defaultGatewayName}, updated)
	listeners, _, _ := unstructured.NestedSlice(updated.Object, "spec", "listeners")
	if len(listeners) != 2 {
		t.Fatalf("expected http+https only, got %d listeners", len(listeners))
	}
}

func TestApplyGatewayHTTPSListeners_TC09_zeroViewers(t *testing.T) {
	s := gatewayScheme(t)
	c := fake.NewClientBuilder().WithScheme(s).WithObjects(baseGateway()).Build()
	if err := applyGatewayHTTPSListeners(context.Background(), c, nil, nil, "olares.com"); err != nil {
		t.Fatal(err)
	}
	updated := &unstructured.Unstructured{}
	updated.SetGroupVersionKind(gatewayGVK)
	if err := c.Get(context.Background(), types.NamespacedName{Namespace: defaultGatewayNS, Name: defaultGatewayName}, updated); err != nil {
		t.Fatal(err)
	}
	https := findListener(updated, "https")
	if https == nil {
		t.Fatal("https listener missing")
	}
	refs, _, _ := unstructured.NestedSlice(https, "tls", "certificateRefs")
	if len(refs) != 0 {
		t.Fatalf("zero viewers should clear certRefs, got %v", refs)
	}
}

func findListener(gw *unstructured.Unstructured, name string) map[string]interface{} {
	listeners, _, _ := unstructured.NestedSlice(gw.Object, "spec", "listeners")
	for _, raw := range listeners {
		m, ok := raw.(map[string]interface{})
		if !ok {
			continue
		}
		if n, _, _ := unstructured.NestedString(m, "name"); n == name {
			return m
		}
	}
	return nil
}

func TestListViewerTLSSecrets_excludesCustomDomain(t *testing.T) {
	s := testScheme(t)
	customSec := desiredCustomDomainTLSSecret("shared-entrance-tls-custom-shop", "shop.example.com", "user-space-alice", "C", "K", "h")
	c := fake.NewClientBuilder().WithScheme(s).WithObjects(customSec, viewerSecret("alice")).Build()
	viewers, err := listViewerTLSSecrets(context.Background(), c)
	if err != nil {
		t.Fatal(err)
	}
	if len(viewers) != 1 || viewers[0].Viewer != "alice" {
		t.Fatalf("viewers = %v, want only alice", viewers)
	}
}

func TestEntranceTLSListenerReconciler_viewerAndCustomCoexist(t *testing.T) {
	cluster.PrimePlatformDomainForTest("olares.com")
	t.Cleanup(cluster.ResetPlatformDomainForTest)
	s := gatewayScheme(t)
	customSec := desiredCustomDomainTLSSecret("shared-entrance-tls-custom-shop", "shop.example.com", "user-space-alice", "C", "K", "h")
	c := fake.NewClientBuilder().WithScheme(s).WithObjects(baseGateway(), viewerSecret("alice"), customSec).Build()
	r := &EntranceTLSListenerReconciler{Client: c}
	if _, err := r.Reconcile(context.Background(), reconcile.Request{
		NamespacedName: types.NamespacedName{Namespace: defaultGatewayNS, Name: "shared-entrance-tls-alice"},
	}); err != nil {
		t.Fatal(err)
	}
	updated := &unstructured.Unstructured{}
	updated.SetGroupVersionKind(gatewayGVK)
	if err := c.Get(context.Background(), types.NamespacedName{Namespace: defaultGatewayNS, Name: defaultGatewayName}, updated); err != nil {
		t.Fatal(err)
	}
	if findListener(updated, "https-custom-shop") != nil {
		t.Fatal("custom secret must not produce viewer-style listener https-custom-shop")
	}
	alice := findListener(updated, "https")
	custom := findListener(updated, "https-custom-shop-example-com")
	if alice == nil || custom == nil {
		t.Fatal("expected https and custom listeners")
	}
	host, _, _ := unstructured.NestedString(custom, "hostname")
	if host != "shop.example.com" {
		t.Errorf("custom hostname = %q", host)
	}
}

func TestApplyGatewayHTTPSListeners_TC03_bobRotation(t *testing.T) {
	cluster.PrimePlatformDomainForTest("olares.com")
	t.Cleanup(cluster.ResetPlatformDomainForTest)
	s := gatewayScheme(t)
	gw := baseGateway()
	alice := viewerSecret("alice")
	bob := viewerSecret("bob")
	c := fake.NewClientBuilder().WithScheme(s).WithObjects(gw, alice, bob).Build()
	viewers, _ := listViewerTLSSecrets(context.Background(), c)
	if err := applyGatewayHTTPSListeners(context.Background(), c, viewers, nil, "olares.com"); err != nil {
		t.Fatal(err)
	}
	updated := &unstructured.Unstructured{}
	updated.SetGroupVersionKind(gatewayGVK)
	_ = c.Get(context.Background(), types.NamespacedName{Namespace: defaultGatewayNS, Name: defaultGatewayName}, updated)
	aliceBefore := findListener(updated, "https")
	bobBefore := findListener(updated, "https-bob")

	bob.StringData[corev1.TLSCertKey] = "rotated-cert"
	if err := c.Update(context.Background(), bob); err != nil {
		t.Fatal(err)
	}
	viewers, _ = listViewerTLSSecrets(context.Background(), c)
	if err := applyGatewayHTTPSListeners(context.Background(), c, viewers, nil, "olares.com"); err != nil {
		t.Fatal(err)
	}
	_ = c.Get(context.Background(), types.NamespacedName{Namespace: defaultGatewayNS, Name: defaultGatewayName}, updated)
	aliceAfter := findListener(updated, "https")
	bobAfter := findListener(updated, "https-bob")
	if listenerCertRef(aliceBefore) != listenerCertRef(aliceAfter) {
		t.Errorf("alice certRef changed: %q -> %q", listenerCertRef(aliceBefore), listenerCertRef(aliceAfter))
	}
	if listenerCertRef(bobBefore) != listenerCertRef(bobAfter) || listenerCertRef(bobAfter) != "shared-entrance-tls-bob" {
		t.Errorf("bob certRef = %q", listenerCertRef(bobAfter))
	}
}

func TestApplyGatewayHTTPSListeners_TC08_bobSNICertRef(t *testing.T) {
	cluster.PrimePlatformDomainForTest("olares.com")
	t.Cleanup(cluster.ResetPlatformDomainForTest)
	s := gatewayScheme(t)
	c := fake.NewClientBuilder().WithScheme(s).WithObjects(baseGateway(), viewerSecret("alice"), viewerSecret("bob")).Build()
	viewers, _ := listViewerTLSSecrets(context.Background(), c)
	if err := applyGatewayHTTPSListeners(context.Background(), c, viewers, nil, "olares.com"); err != nil {
		t.Fatal(err)
	}
	updated := &unstructured.Unstructured{}
	updated.SetGroupVersionKind(gatewayGVK)
	_ = c.Get(context.Background(), types.NamespacedName{Namespace: defaultGatewayNS, Name: defaultGatewayName}, updated)
	bob := findListener(updated, "https-bob")
	if bob == nil {
		t.Fatal("bob listener missing")
	}
	host, _, _ := unstructured.NestedString(bob, "hostname")
	if host != "*.bob.olares.com" {
		t.Errorf("bob hostname = %q", host)
	}
	if got := listenerCertRef(bob); got != "shared-entrance-tls-bob" {
		t.Errorf("bob certRef = %q, want shared-entrance-tls-bob for SNI *.bob.olares.com", got)
	}
}

func TestApplyGatewayHTTPSListeners_TC10_listenerNameCollision(t *testing.T) {
	got := uniqueHTTPSListenerName("foo.bar", map[string]struct{}{"https-foo-bar": {}})
	if got == "https-foo-bar" {
		t.Fatalf("expected disambiguated listener name, got %q", got)
	}
	if !strings.HasPrefix(got, "https-foo-bar-") {
		t.Errorf("listener name = %q", got)
	}
}

func TestEntranceTLSListenerReconciler_TC12_helmDriftReconverge(t *testing.T) {
	cluster.PrimePlatformDomainForTest("olares.com")
	t.Cleanup(cluster.ResetPlatformDomainForTest)
	s := gatewayScheme(t)
	gw := baseGateway()
	c := fake.NewClientBuilder().WithScheme(s).WithObjects(gw, viewerSecret("alice"), viewerSecret("bob")).Build()
	r := &EntranceTLSListenerReconciler{Client: c}
	if _, err := r.Reconcile(context.Background(), reconcile.Request{
		NamespacedName: types.NamespacedName{Namespace: defaultGatewayNS, Name: "shared-entrance-tls-alice"},
	}); err != nil {
		t.Fatal(err)
	}

	drifted := &unstructured.Unstructured{}
	drifted.SetGroupVersionKind(gatewayGVK)
	if err := c.Get(context.Background(), types.NamespacedName{Namespace: defaultGatewayNS, Name: defaultGatewayName}, drifted); err != nil {
		t.Fatal(err)
	}
	driftedListeners := []interface{}{
		map[string]interface{}{
			"name":     "http",
			"protocol": "HTTP",
			"port":     int64(80),
			"allowedRoutes": map[string]interface{}{
				"namespaces": map[string]interface{}{"from": "All"},
			},
		},
		map[string]interface{}{
			"name":     "https",
			"protocol": "HTTPS",
			"port":     int64(443),
			"allowedRoutes": map[string]interface{}{
				"namespaces": map[string]interface{}{"from": "All"},
			},
			"tls": map[string]interface{}{
				"mode": "Terminate",
				"certificateRefs": []interface{}{
					map[string]interface{}{"kind": "Secret", "name": gatewayTLSSecretName},
				},
			},
		},
	}
	if err := unstructured.SetNestedSlice(drifted.Object, driftedListeners, "spec", "listeners"); err != nil {
		t.Fatal(err)
	}
	if err := c.Update(context.Background(), drifted); err != nil {
		t.Fatal(err)
	}

	if _, err := r.Reconcile(context.Background(), reconcile.Request{
		NamespacedName: types.NamespacedName{Namespace: defaultGatewayNS, Name: defaultGatewayName},
	}); err != nil {
		t.Fatal(err)
	}
	updated := &unstructured.Unstructured{}
	updated.SetGroupVersionKind(gatewayGVK)
	if err := c.Get(context.Background(), types.NamespacedName{Namespace: defaultGatewayNS, Name: defaultGatewayName}, updated); err != nil {
		t.Fatal(err)
	}
	alice := findListener(updated, "https")
	bob := findListener(updated, "https-bob")
	if alice == nil || bob == nil {
		t.Fatal("expected reconciler to restore per-viewer listeners after helm drift")
	}
	if got := listenerCertRef(alice); got != "shared-entrance-tls-alice" {
		t.Errorf("alice certRef = %q", got)
	}
}

func TestEntranceTLSListenerReconciler_requeueWithoutDomain(t *testing.T) {
	cluster.PrimePlatformDomainForTest("")
	t.Cleanup(cluster.ResetPlatformDomainForTest)
	s := gatewayScheme(t)
	c := fake.NewClientBuilder().WithScheme(s).WithObjects(baseGateway(), viewerSecret("alice")).Build()
	r := &EntranceTLSListenerReconciler{Client: c}
	res, err := r.Reconcile(context.Background(), reconcile.Request{
		NamespacedName: types.NamespacedName{Namespace: defaultGatewayNS, Name: "shared-entrance-tls-alice"},
	})
	if err != nil {
		t.Fatal(err)
	}
	if res.RequeueAfter == 0 {
		t.Error("expected requeue when platform domain empty")
	}
}
