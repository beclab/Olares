package linkerdpki

import (
	"context"
	"strings"
	"testing"

	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"sigs.k8s.io/controller-runtime/pkg/client/fake"
)

// TC-PKI-G04: when olares-linkerd-pki is absent, MaintainLinkerdPKI returns a
// transient error and never terminates the process.
func TestMaintainLinkerdPKISecretNotFound(t *testing.T) {
	scheme := runtime.NewScheme()
	if err := corev1.AddToScheme(scheme); err != nil {
		t.Fatal(err)
	}
	c := fake.NewClientBuilder().WithScheme(scheme).Build()

	err := MaintainLinkerdPKI(context.Background(), c, DefaultLinkerdNamespace)
	if err == nil {
		t.Fatal("expected error when olares-linkerd-pki secret is missing")
	}
	if !strings.Contains(err.Error(), "not found") {
		t.Fatalf("unexpected error: %v", err)
	}
}
