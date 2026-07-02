package routecontrol

import (
	"context"
	"testing"

	corev1 "k8s.io/api/core/v1"
	apierrors "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/types"
	"sigs.k8s.io/controller-runtime/pkg/client/fake"
	"sigs.k8s.io/controller-runtime/pkg/reconcile"
)

func TestSyncPerViewerTLS_T4a1_singleViewer(t *testing.T) {
	s := testScheme(t)
	cm := &corev1.ConfigMap{
		ObjectMeta: metav1.ObjectMeta{Name: zoneSSLConfigMapName, Namespace: "user-space-alice"},
		Data:       map[string]string{"cert": "CERT-A", "key": "KEY-A"},
	}
	c := fake.NewClientBuilder().WithScheme(s).WithObjects(cm).Build()

	if err := syncPerViewerTLS(context.Background(), c, cm, "alice"); err != nil {
		t.Fatalf("sync: %v", err)
	}
	sec := &corev1.Secret{}
	if err := c.Get(context.Background(), types.NamespacedName{Namespace: defaultGatewayNS, Name: "shared-entrance-tls-alice"}, sec); err != nil {
		t.Fatalf("secret not created: %v", err)
	}
	if sec.Labels[labelTLSViewer] != "alice" {
		t.Errorf("tls-viewer label = %q", sec.Labels[labelTLSViewer])
	}
	err := c.Get(context.Background(), types.NamespacedName{Namespace: defaultGatewayNS, Name: gatewayTLSSecretName}, &corev1.Secret{})
	if err == nil {
		t.Error("app-gateway-tls should not be created")
	}
}

func TestSyncPerViewerTLS_T4a2_twoViewers(t *testing.T) {
	s := testScheme(t)
	aliceCM := &corev1.ConfigMap{
		ObjectMeta: metav1.ObjectMeta{Name: zoneSSLConfigMapName, Namespace: "user-space-alice"},
		Data:       map[string]string{"cert": "CERT-A", "key": "KEY-A"},
	}
	bobCM := &corev1.ConfigMap{
		ObjectMeta: metav1.ObjectMeta{Name: zoneSSLConfigMapName, Namespace: "user-space-bob"},
		Data:       map[string]string{"cert": "CERT-B", "key": "KEY-B"},
	}
	c := fake.NewClientBuilder().WithScheme(s).WithObjects(aliceCM, bobCM).Build()

	if err := syncPerViewerTLS(context.Background(), c, aliceCM, "alice"); err != nil {
		t.Fatal(err)
	}
	if err := syncPerViewerTLS(context.Background(), c, bobCM, "bob"); err != nil {
		t.Fatal(err)
	}
	alice := &corev1.Secret{}
	bob := &corev1.Secret{}
	if err := c.Get(context.Background(), types.NamespacedName{Namespace: defaultGatewayNS, Name: "shared-entrance-tls-alice"}, alice); err != nil {
		t.Fatal(err)
	}
	if err := c.Get(context.Background(), types.NamespacedName{Namespace: defaultGatewayNS, Name: "shared-entrance-tls-bob"}, bob); err != nil {
		t.Fatal(err)
	}
	if alice.Annotations[annotationTLSContentHash] == bob.Annotations[annotationTLSContentHash] {
		t.Error("viewer secrets should not share content hash")
	}
}

func TestSyncPerViewerTLS_T4a3_idempotent(t *testing.T) {
	s := testScheme(t)
	cm := &corev1.ConfigMap{
		ObjectMeta: metav1.ObjectMeta{Name: zoneSSLConfigMapName, Namespace: "user-space-alice"},
		Data:       map[string]string{"cert": "CERT-A", "key": "KEY-A"},
	}
	c := fake.NewClientBuilder().WithScheme(s).WithObjects(cm).Build()
	if err := syncPerViewerTLS(context.Background(), c, cm, "alice"); err != nil {
		t.Fatal(err)
	}
	sec := &corev1.Secret{}
	key := types.NamespacedName{Namespace: defaultGatewayNS, Name: "shared-entrance-tls-alice"}
	if err := c.Get(context.Background(), key, sec); err != nil {
		t.Fatal(err)
	}
	firstHash := sec.Annotations[annotationTLSContentHash]
	if err := syncPerViewerTLS(context.Background(), c, cm, "alice"); err != nil {
		t.Fatal(err)
	}
	if err := c.Get(context.Background(), key, sec); err != nil {
		t.Fatal(err)
	}
	if sec.Annotations[annotationTLSContentHash] != firstHash {
		t.Error("hash should be unchanged on idempotent sync")
	}
}

func TestSyncPerViewerTLS_T4a4_deleteCM(t *testing.T) {
	s := testScheme(t)
	sec := desiredPerViewerTLSSecret("alice", "user-space-alice", "CERT", "KEY", "hash")
	c := fake.NewClientBuilder().WithScheme(s).WithObjects(sec).Build()
	if err := deletePerViewerTLSSecret(context.Background(), c, "alice"); err != nil {
		t.Fatal(err)
	}
	err := c.Get(context.Background(), types.NamespacedName{Namespace: defaultGatewayNS, Name: "shared-entrance-tls-alice"}, &corev1.Secret{})
	if !apierrors.IsNotFound(err) {
		t.Fatalf("secret should be deleted: %v", err)
	}
}

func TestSyncPerViewerTLS_T4a6_ephemeralSkip(t *testing.T) {
	s := testScheme(t)
	cm := &corev1.ConfigMap{
		ObjectMeta: metav1.ObjectMeta{Name: zoneSSLConfigMapName, Namespace: "user-space-guest"},
		Data:       map[string]string{"cert": "CERT", "key": "KEY", "ephemeral": "true"},
	}
	c := fake.NewClientBuilder().WithScheme(s).WithObjects(cm).Build()
	if err := syncPerViewerTLS(context.Background(), c, cm, "guest"); err != nil {
		t.Fatal(err)
	}
	err := c.Get(context.Background(), types.NamespacedName{Namespace: defaultGatewayNS, Name: "shared-entrance-tls-guest"}, &corev1.Secret{})
	if !apierrors.IsNotFound(err) {
		t.Fatalf("ephemeral CM should not create secret: %v", err)
	}

	sec := desiredPerViewerTLSSecret("guest", "user-space-guest", "OLD", "KEY", "old")
	c = fake.NewClientBuilder().WithScheme(s).WithObjects(cm, sec).Build()
	if err := syncPerViewerTLS(context.Background(), c, cm, "guest"); err != nil {
		t.Fatal(err)
	}
	err = c.Get(context.Background(), types.NamespacedName{Namespace: defaultGatewayNS, Name: "shared-entrance-tls-guest"}, &corev1.Secret{})
	if !apierrors.IsNotFound(err) {
		t.Fatalf("ephemeral CM should gc existing secret: %v", err)
	}
}

func TestSyncPerViewerTLS_T4a7_certRotation(t *testing.T) {
	s := testScheme(t)
	cm := &corev1.ConfigMap{
		ObjectMeta: metav1.ObjectMeta{Name: zoneSSLConfigMapName, Namespace: "user-space-alice"},
		Data:       map[string]string{"cert": "CERT-A", "key": "KEY-A"},
	}
	c := fake.NewClientBuilder().WithScheme(s).WithObjects(cm).Build()
	if err := syncPerViewerTLS(context.Background(), c, cm, "alice"); err != nil {
		t.Fatal(err)
	}
	sec := &corev1.Secret{}
	key := types.NamespacedName{Namespace: defaultGatewayNS, Name: "shared-entrance-tls-alice"}
	if err := c.Get(context.Background(), key, sec); err != nil {
		t.Fatal(err)
	}
	firstHash := sec.Annotations[annotationTLSContentHash]

	cm.Data["cert"] = "CERT-B"
	if err := syncPerViewerTLS(context.Background(), c, cm, "alice"); err != nil {
		t.Fatal(err)
	}
	if err := c.Get(context.Background(), key, sec); err != nil {
		t.Fatal(err)
	}
	if sec.Annotations[annotationTLSContentHash] == firstHash {
		t.Error("expected content hash to change after cert rotation")
	}
}

func TestEntranceTLSReconciler_Reconcile_cmDeleted(t *testing.T) {
	s := testScheme(t)
	sec := desiredPerViewerTLSSecret("alice", "user-space-alice", "CERT", "KEY", "hash")
	c := fake.NewClientBuilder().WithScheme(s).WithObjects(sec).Build()
	r := &EntranceTLSReconciler{Client: c}
	if _, err := r.Reconcile(context.Background(), reconcile.Request{
		NamespacedName: types.NamespacedName{Namespace: "user-space-alice", Name: zoneSSLConfigMapName},
	}); err != nil {
		t.Fatal(err)
	}
	err := c.Get(context.Background(), types.NamespacedName{Namespace: defaultGatewayNS, Name: "shared-entrance-tls-alice"}, &corev1.Secret{})
	if !apierrors.IsNotFound(err) {
		t.Fatalf("secret should be deleted when CM missing: %v", err)
	}
}

func TestEntranceTLSReconciler_Reconcile_nonUserSpaceNS(t *testing.T) {
	s := testScheme(t)
	cm := &corev1.ConfigMap{
		ObjectMeta: metav1.ObjectMeta{Name: zoneSSLConfigMapName, Namespace: "kube-system"},
		Data:       map[string]string{"cert": "CERT", "key": "KEY"},
	}
	c := fake.NewClientBuilder().WithScheme(s).WithObjects(cm).Build()
	r := &EntranceTLSReconciler{Client: c}
	if _, err := r.Reconcile(context.Background(), reconcile.Request{
		NamespacedName: types.NamespacedName{Namespace: "kube-system", Name: zoneSSLConfigMapName},
	}); err != nil {
		t.Fatal(err)
	}
	var secList corev1.SecretList
	if err := c.List(context.Background(), &secList); err != nil {
		t.Fatal(err)
	}
	if len(secList.Items) != 0 {
		t.Fatalf("expected no secrets, got %d", len(secList.Items))
	}
}

func TestViewerFromUserSpaceNS(t *testing.T) {
	tests := []struct {
		ns         string
		wantViewer string
		wantOK     bool
	}{
		{"user-space-alice", "alice", true},
		{"user-space-", "", false},
		{"user-space-Alice", "", false},
		{"kube-system", "", false},
	}
	for _, tt := range tests {
		viewer, ok := viewerFromUserSpaceNS(tt.ns)
		if ok != tt.wantOK || viewer != tt.wantViewer {
			t.Errorf("viewerFromUserSpaceNS(%q) = (%q, %v), want (%q, %v)", tt.ns, viewer, ok, tt.wantViewer, tt.wantOK)
		}
	}
}

func TestDeletePerViewerTLSSecret_skipsUnmanaged(t *testing.T) {
	s := testScheme(t)
	sec := &corev1.Secret{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "shared-entrance-tls-alice",
			Namespace: defaultGatewayNS,
		},
		Type: corev1.SecretTypeTLS,
	}
	c := fake.NewClientBuilder().WithScheme(s).WithObjects(sec).Build()
	if err := deletePerViewerTLSSecret(context.Background(), c, "alice"); err != nil {
		t.Fatal(err)
	}
	if err := c.Get(context.Background(), types.NamespacedName{Namespace: defaultGatewayNS, Name: "shared-entrance-tls-alice"}, &corev1.Secret{}); err != nil {
		t.Fatalf("unmanaged secret should not be deleted: %v", err)
	}
}

func TestSyncPerViewerTLS_T4a5_incompleteGC(t *testing.T) {
	s := testScheme(t)
	cm := &corev1.ConfigMap{
		ObjectMeta: metav1.ObjectMeta{Name: zoneSSLConfigMapName, Namespace: "user-space-bob"},
		Data:       map[string]string{"cert": "CERT-ONLY"},
	}
	sec := desiredPerViewerTLSSecret("bob", "user-space-bob", "OLD", "KEY", "old")
	c := fake.NewClientBuilder().WithScheme(s).WithObjects(cm, sec).Build()
	if err := syncPerViewerTLS(context.Background(), c, cm, "bob"); err != nil {
		t.Fatal(err)
	}
	err := c.Get(context.Background(), types.NamespacedName{Namespace: defaultGatewayNS, Name: "shared-entrance-tls-bob"}, &corev1.Secret{})
	if !apierrors.IsNotFound(err) {
		t.Fatalf("incomplete CM should gc secret: %v", err)
	}
}
