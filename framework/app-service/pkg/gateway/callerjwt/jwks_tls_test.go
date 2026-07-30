package callerjwt

import (
	"context"
	"os"
	"path/filepath"
	"testing"

	corev1 "k8s.io/api/core/v1"
	networkingv1 "k8s.io/api/networking/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/types"
	"sigs.k8s.io/controller-runtime/pkg/client/fake"
)

func TestIssuerURLAuthzHost(t *testing.T) {
	if IssuerHost != "authz.olares.system" {
		t.Fatalf("IssuerHost = %q", IssuerHost)
	}
	if IssuerURL != "https://authz.olares.system/" {
		t.Fatalf("IssuerURL = %q", IssuerURL)
	}
	if JWKSURI != "https://authz.olares.system/.well-known/jwks.json" {
		t.Fatalf("JWKSURI = %q", JWKSURI)
	}
}

func TestDesiredJWKSBackendTLSPolicy(t *testing.T) {
	obj := desiredJWKSBackendTLSPolicy()
	if obj.GetName() != JWKSBackendTLSPolicyName || obj.GetNamespace() != JWKSServiceNamespace {
		t.Fatalf("meta = %s/%s", obj.GetNamespace(), obj.GetName())
	}
	spec, ok := obj.Object["spec"].(map[string]any)
	if !ok {
		t.Fatalf("spec type %T", obj.Object["spec"])
	}
	validation, ok := spec["validation"].(map[string]any)
	if !ok {
		t.Fatalf("validation type %T", spec["validation"])
	}
	if validation["hostname"] != IssuerHost {
		t.Fatalf("hostname = %v", validation["hostname"])
	}
	refs, ok := validation["caCertificateRefs"].([]any)
	if !ok || len(refs) != 1 {
		t.Fatalf("caCertificateRefs = %#v", validation["caCertificateRefs"])
	}
	ref := refs[0].(map[string]any)
	if ref["name"] != JWKSCAConfigMapName || ref["kind"] != "ConfigMap" {
		t.Fatalf("ca ref = %#v", ref)
	}
	targets, ok := spec["targetRefs"].([]any)
	if !ok || len(targets) != 1 {
		t.Fatalf("targetRefs = %#v", spec["targetRefs"])
	}
	tr := targets[0].(map[string]any)
	if tr["name"] != JWKSServiceName || tr["sectionName"] != jwksServicePortName {
		t.Fatalf("targetRef = %#v", tr)
	}
}

func TestReconcileJWKSCAConfigMap(t *testing.T) {
	dir := t.TempDir()
	caPath := filepath.Join(dir, "ca.crt")
	if err := os.WriteFile(caPath, []byte("-----BEGIN CERTIFICATE-----\nMIIB\n-----END CERTIFICATE-----\n"), 0o600); err != nil {
		t.Fatalf("write ca: %v", err)
	}
	certPath := filepath.Join(dir, "server.crt")
	if err := os.WriteFile(certPath, []byte("unused"), 0o600); err != nil {
		t.Fatalf("write cert: %v", err)
	}
	t.Setenv("WEBHOOK_TLS_CERT", certPath)
	t.Setenv("WEBHOOK_TLS_KEY", filepath.Join(dir, "server.key"))

	scheme := runtime.NewScheme()
	if err := corev1.AddToScheme(scheme); err != nil {
		t.Fatalf("scheme: %v", err)
	}
	if err := networkingv1.AddToScheme(scheme); err != nil {
		t.Fatalf("net scheme: %v", err)
	}
	c := fake.NewClientBuilder().WithScheme(scheme).Build()
	r := &IssuerReconciler{Client: c, Scheme: scheme}
	if err := r.reconcileJWKSCAConfigMap(context.Background()); err != nil {
		t.Fatalf("reconcileJWKSCAConfigMap: %v", err)
	}
	cm := &corev1.ConfigMap{}
	if err := c.Get(context.Background(), types.NamespacedName{
		Namespace: JWKSServiceNamespace,
		Name:      JWKSCAConfigMapName,
	}, cm); err != nil {
		t.Fatalf("get cm: %v", err)
	}
	if cm.Data[JWKSCAConfigMapDataKey] == "" {
		t.Fatal("ca.crt data empty")
	}
}
