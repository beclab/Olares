package routecontrol

import (
	"context"
	"testing"

	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/types"
	"sigs.k8s.io/controller-runtime/pkg/client/fake"
)

func TestSyncGatewayTLS(t *testing.T) {
	s := testScheme(t)
	cm := &corev1.ConfigMap{
		ObjectMeta: metav1.ObjectMeta{Name: zoneSSLConfigMapName, Namespace: "user-space-alice"},
		Data:       map[string]string{"cert": "CERT-A", "key": "KEY-A"},
	}
	c := fake.NewClientBuilder().WithScheme(s).WithObjects(cm).Build()

	if err := syncGatewayTLS(context.Background(), c, cm); err != nil {
		t.Fatalf("create sync: %v", err)
	}
	sec := &corev1.Secret{}
	key := types.NamespacedName{Namespace: defaultGatewayNS, Name: gatewayTLSSecretName}
	if err := c.Get(context.Background(), key, sec); err != nil {
		t.Fatalf("secret not created: %v", err)
	}
	if sec.Type != corev1.SecretTypeTLS {
		t.Errorf("secret type = %q, want kubernetes.io/tls", sec.Type)
	}
	firstHash := sec.Annotations[annotationTLSContentHash]
	if firstHash == "" {
		t.Fatal("content hash annotation missing")
	}

	// Idempotent: same content does not change the hash.
	if err := syncGatewayTLS(context.Background(), c, cm); err != nil {
		t.Fatalf("idempotent sync: %v", err)
	}

	// New cert content rotates the Secret.
	cm.Data["cert"] = "CERT-B"
	if err := syncGatewayTLS(context.Background(), c, cm); err != nil {
		t.Fatalf("rotate sync: %v", err)
	}
	if err := c.Get(context.Background(), key, sec); err != nil {
		t.Fatal(err)
	}
	if sec.Annotations[annotationTLSContentHash] == firstHash {
		t.Error("expected content hash to change after cert rotation")
	}
}

func TestSyncGatewayTLSSkipsIncomplete(t *testing.T) {
	s := testScheme(t)
	cm := &corev1.ConfigMap{
		ObjectMeta: metav1.ObjectMeta{Name: zoneSSLConfigMapName, Namespace: "user-space-bob"},
		Data:       map[string]string{"cert": "CERT-ONLY"},
	}
	c := fake.NewClientBuilder().WithScheme(s).WithObjects(cm).Build()
	if err := syncGatewayTLS(context.Background(), c, cm); err != nil {
		t.Fatalf("sync: %v", err)
	}
	sec := &corev1.Secret{}
	err := c.Get(context.Background(), types.NamespacedName{Namespace: defaultGatewayNS, Name: gatewayTLSSecretName}, sec)
	if err == nil {
		t.Error("incomplete CM should not create a Secret")
	}
}
