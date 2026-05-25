package routecontrol

import (
	"context"
	"testing"

	corev1 "k8s.io/api/core/v1"
	apierrors "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/types"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

func TestReconcileEntranceTLS_create(t *testing.T) {
	ctx := context.Background()
	c := plainFixture(t, &corev1.Namespace{ObjectMeta: metav1.ObjectMeta{Name: defaultGatewayNS}})
	cm := zoneSSLConfigMap("user-space-alice", "cert-pem-data", "key-pem-data")

	wrote, err := reconcileEntranceTLS(ctx, c, cm)
	if err != nil {
		t.Fatalf("reconcile: %v", err)
	}
	if !wrote {
		t.Fatal("expected secret create")
	}

	secret := &corev1.Secret{}
	if err := c.Get(ctx, types.NamespacedName{
		Namespace: defaultGatewayNS,
		Name:      "shared-entrance-tls-alice",
	}, secret); err != nil {
		t.Fatalf("get secret: %v", err)
	}
	if secret.Type != corev1.SecretTypeTLS {
		t.Fatalf("type=%s", secret.Type)
	}
	if secret.Labels[labelTLSViewer] != "alice" {
		t.Fatalf("labels=%v", secret.Labels)
	}
	if secret.Labels[labelGatewayRouteControl] != gatewayRouteControlValue {
		t.Fatalf("managed-by label missing")
	}
	if got := string(secret.Data[corev1.TLSCertKey]); got == "" && secret.StringData[corev1.TLSCertKey] != "cert-pem-data" {
		t.Fatalf("cert not materialized: Data=%q StringData=%v", got, secret.StringData)
	}
}

func TestReconcileEntranceTLS_updateOnCertChange(t *testing.T) {
	ctx := context.Background()
	c := plainFixture(t, &corev1.Namespace{ObjectMeta: metav1.ObjectMeta{Name: defaultGatewayNS}})
	cm1 := zoneSSLConfigMap("user-space-alice", "cert-v1", "key-v1")
	if _, err := reconcileEntranceTLS(ctx, c, cm1); err != nil {
		t.Fatalf("first reconcile: %v", err)
	}

	wrote, err := reconcileEntranceTLS(ctx, c, cm1)
	if err != nil {
		t.Fatalf("noop reconcile: %v", err)
	}
	if wrote {
		t.Fatal("expected no-op when hash unchanged")
	}

	cm2 := zoneSSLConfigMap("user-space-alice", "cert-v2", "key-v2")
	wrote, err = reconcileEntranceTLS(ctx, c, cm2)
	if err != nil {
		t.Fatalf("update reconcile: %v", err)
	}
	if !wrote {
		t.Fatal("expected secret update")
	}
	secret := &corev1.Secret{}
	_ = c.Get(ctx, types.NamespacedName{Namespace: defaultGatewayNS, Name: "shared-entrance-tls-alice"}, secret)
	if secret.Annotations[annotationTLSContentHash] != tlsMaterialHash("cert-v2", "key-v2") {
		t.Fatalf("hash=%q", secret.Annotations[annotationTLSContentHash])
	}
}

func TestReconcileEntranceTLS_missingFields(t *testing.T) {
	ctx := context.Background()
	c := plainFixture(t, &corev1.Namespace{ObjectMeta: metav1.ObjectMeta{Name: defaultGatewayNS}})
	cm := zoneSSLConfigMap("user-space-bob", "", "key-only")

	wrote, err := reconcileEntranceTLS(ctx, c, cm)
	if err != nil {
		t.Fatalf("reconcile: %v", err)
	}
	if wrote {
		t.Fatal("expected skip")
	}
	secret := &corev1.Secret{}
	err = c.Get(ctx, types.NamespacedName{Namespace: defaultGatewayNS, Name: "shared-entrance-tls-bob"}, secret)
	if !apierrors.IsNotFound(err) {
		t.Fatalf("expected no secret, got %v", err)
	}
}

func TestDeleteEntranceTLSSecret(t *testing.T) {
	ctx := context.Background()
	c := plainFixture(t, &corev1.Namespace{ObjectMeta: metav1.ObjectMeta{Name: defaultGatewayNS}})
	cm := zoneSSLConfigMap("user-space-carol", "cert", "key")
	if _, err := reconcileEntranceTLS(ctx, c, cm); err != nil {
		t.Fatalf("create: %v", err)
	}
	if err := deleteEntranceTLSSecret(ctx, c, "carol"); err != nil {
		t.Fatalf("delete: %v", err)
	}
	secret := &corev1.Secret{}
	err := c.Get(ctx, types.NamespacedName{Namespace: defaultGatewayNS, Name: "shared-entrance-tls-carol"}, secret)
	if !apierrors.IsNotFound(err) {
		t.Fatalf("expected NotFound, got %v", err)
	}
}

func TestEntranceTLSReconciler_ReconcileConfigMap(t *testing.T) {
	ctx := context.Background()
	c := plainFixture(t, &corev1.Namespace{ObjectMeta: metav1.ObjectMeta{Name: defaultGatewayNS}})
	r := &EntranceTLSReconciler{Client: c}
	if err := r.ReconcileConfigMap(ctx, zoneSSLConfigMap("user-space-dave", "c", "k")); err != nil {
		t.Fatalf("reconcile: %v", err)
	}
	secret := &corev1.Secret{}
	if err := c.Get(ctx, client.ObjectKey{Namespace: defaultGatewayNS, Name: "shared-entrance-tls-dave"}, secret); err != nil {
		t.Fatalf("get: %v", err)
	}
}

func zoneSSLConfigMap(namespace, cert, key string) *corev1.ConfigMap {
	return &corev1.ConfigMap{
		ObjectMeta: metav1.ObjectMeta{
			Name:      zoneSSLConfigMapName,
			Namespace: namespace,
		},
		Data: map[string]string{
			"cert": cert,
			"key":  key,
			"zone": "alice.olares.com",
		},
	}
}
