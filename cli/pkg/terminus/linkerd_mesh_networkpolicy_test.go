package terminus

import (
	"context"
	"os"
	"path/filepath"
	"testing"

	agwconfig "github.com/beclab/Olares/framework/app-gateway/pkg/config"
	netv1 "k8s.io/api/networking/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"sigs.k8s.io/controller-runtime/pkg/client/fake"
)

func TestLinkerdMeshNPReconciledByAppService_missing(t *testing.T) {
	c := fake.NewClientBuilder().WithScheme(runtimeScheme()).Build()
	ok, err := linkerdMeshNPReconciledByAppService(context.Background(), c)
	if err != nil {
		t.Fatal(err)
	}
	if ok {
		t.Fatal("expected false when policies absent")
	}
}

func TestLinkerdMeshNPReconciledByAppService_present(t *testing.T) {
	c := fake.NewClientBuilder().WithScheme(runtimeScheme()).
		WithObjects(
			&netv1.NetworkPolicy{
				ObjectMeta: metav1.ObjectMeta{
					Name: appGatewayMeshNPName, Namespace: agwconfig.LinkerdNamespace(),
				},
			},
			&netv1.NetworkPolicy{
				ObjectMeta: metav1.ObjectMeta{
					Name: appGatewayMeshNPName, Namespace: agwconfig.Namespace(),
				},
			},
		).Build()
	ok, err := linkerdMeshNPReconciledByAppService(context.Background(), c)
	if err != nil {
		t.Fatal(err)
	}
	if !ok {
		t.Fatal("expected true when both namespaces have app-gateway-mesh-np")
	}
}

func TestLinkerdMeshNetworkPolicyManifest_vendorDir(t *testing.T) {
	dir := t.TempDir()
	npPath := filepath.Join(dir, "network-policies", "linkerd-mesh-ingress.yaml")
	if err := os.MkdirAll(filepath.Dir(npPath), 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(npPath, []byte("---\n"), 0o644); err != nil {
		t.Fatal(err)
	}
	got := linkerdMeshNetworkPolicyManifest(dir)
	if got != npPath {
		t.Fatalf("manifest path: got %q want %q", got, npPath)
	}
}

func runtimeScheme() *runtime.Scheme {
	s := runtime.NewScheme()
	_ = netv1.AddToScheme(s)
	return s
}
