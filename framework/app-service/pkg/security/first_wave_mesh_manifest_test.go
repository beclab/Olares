package security

import "testing"

func TestFirstWaveWorkloadsRenderNativeLinkerdWithoutTrafficExclusions(t *testing.T) {
	for _, tc := range []struct {
		template    string
		workload    string
		account     string
		opaquePorts string
	}{
		{"../../../system-server/.olares/config/cluster/deploy/system_provider.yaml", "Deployment/system-provider", "provider-sa", ""},
		{"../../../files/.olares/config/cluster/deploy/files_deploy.yaml", "DaemonSet/files", "os-internal", "445"},
		{"../../../backup-server/.olares/config/cluster/deploy/backup_server.yaml", "Deployment/backup", "backup-cluster-view", ""},
	} {
		t.Run(tc.workload, func(t *testing.T) {
			docs, err := renderMeshManifests(t, map[string]any{
				"backup":  map[string]any{"is_cloud_version": false},
				"fs_type": "fuse",
			}, tc.template)
			if err != nil {
				t.Fatal(err)
			}
			pod := docs[tc.workload]["spec"].(map[string]any)["template"].(map[string]any)
			metadata := pod["metadata"].(map[string]any)
			annotations := metadata["annotations"].(map[string]any)
			if annotations["linkerd.io/inject"] != "enabled" || annotations["config.linkerd.io/proxy-enable-native-sidecar"] != "true" {
				t.Fatalf("workload must request a native Linkerd proxy")
			}
			if labels := metadata["labels"].(map[string]any); labels["traffic.olares.io/platform-mesh"] != "true" {
				t.Fatal("workload must match the Linkerd control-plane ingress selector")
			}
			opaquePorts, _ := annotations["config.linkerd.io/opaque-ports"].(string)
			if opaquePorts != tc.opaquePorts {
				t.Fatalf("opaque ports = %q, want %q", opaquePorts, tc.opaquePorts)
			}
			for _, key := range []string{"skip-inbound-ports", "skip-outbound-ports", "skip-subnets"} {
				if _, exists := annotations["config.linkerd.io/"+key]; exists {
					t.Fatalf("unexpected interception exclusion %s", key)
				}
			}
			spec := pod["spec"].(map[string]any)
			if spec["serviceAccountName"] != tc.account || spec["hostNetwork"] == true {
				t.Fatal("workload changed identity or uses host networking")
			}
		})
	}
}
