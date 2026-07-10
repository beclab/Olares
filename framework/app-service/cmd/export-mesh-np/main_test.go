package main

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/beclab/Olares/framework/app-service/pkg/security"
	"github.com/google/go-cmp/cmp"
	netv1 "k8s.io/api/networking/v1"
	"sigs.k8s.io/yaml"
)

// TC-EG6-004: exported YAML matches NewAppGatewayMeshNetworkPolicy construction.
func TestExportMeshNP(t *testing.T) {
	dir := t.TempDir()
	if err := run(dir); err != nil {
		t.Fatalf("run: %v", err)
	}
	cases := []struct {
		file   string
		ns     string
		peerNS string
	}{
		{"app-gateway-mesh-np-linkerd.yaml", "linkerd", "os-gateway"},
		{"app-gateway-mesh-np-os-gateway.yaml", "os-gateway", "linkerd"},
	}
	for _, tc := range cases {
		t.Run(tc.file, func(t *testing.T) {
			want := security.NewAppGatewayMeshNetworkPolicy(tc.ns, tc.peerNS)
			want.APIVersion = "networking.k8s.io/v1"
			want.Kind = "NetworkPolicy"
			raw, err := os.ReadFile(filepath.Join(dir, tc.file))
			if err != nil {
				t.Fatalf("read: %v", err)
			}
			var got netv1.NetworkPolicy
			if err := yaml.Unmarshal(raw, &got); err != nil {
				t.Fatalf("unmarshal: %v", err)
			}
			if got.APIVersion != "networking.k8s.io/v1" || got.Kind != "NetworkPolicy" {
				t.Fatalf("exported manifest must include apiVersion/kind; got apiVersion=%q kind=%q", got.APIVersion, got.Kind)
			}
			if diff := cmp.Diff(want, &got); diff != "" {
				t.Fatalf("snapshot mismatch (-want +got):\n%s", diff)
			}
		})
	}
}
