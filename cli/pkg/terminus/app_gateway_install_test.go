package terminus

import (
	"context"
	"testing"

	agwconfig "github.com/beclab/Olares/framework/app-gateway/pkg/config"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/types"
	"sigs.k8s.io/controller-runtime/pkg/client/fake"
)

func TestResolveMeshProfileFromEnv(t *testing.T) {
	cases := []struct {
		name string
		env  string
		want string
	}{
		{name: "TC-LITE-2-1 default full", env: "", want: meshProfileFull},
		{name: "TC-LITE-2-2 explicit lite", env: "lite", want: meshProfileLite},
		{name: "TC-LITE-2-3 explicit full", env: "full", want: meshProfileFull},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			t.Setenv(envMeshProfileBootstrap, tc.env)
			if got := resolveMeshProfileFromEnv(); got != tc.want {
				t.Fatalf("resolveMeshProfileFromEnv() = %q, want %q", got, tc.want)
			}
		})
	}
}

func TestMeshProfileSkipsLinkerdInstall(t *testing.T) {
	if meshProfileSkipsLinkerdInstall(meshProfileFull) {
		t.Fatal("full must install Linkerd")
	}
	if !meshProfileSkipsLinkerdInstall(meshProfileLite) {
		t.Fatal("lite must skip Linkerd")
	}
}

func TestValidateMeshProfileHelmAlignment(t *testing.T) {
	fullDefaults := agwconfig.Defaults{}
	fullDefaults.Mesh.Linkerd.Enabled = true
	if err := validateMeshProfileHelmAlignment(meshProfileLite, fullDefaults); err == nil {
		t.Fatal("lite + mesh enabled defaults must fail fast")
	}
	liteDefaults := agwconfig.Defaults{}
	liteDefaults.Mesh.Linkerd.Enabled = false
	if err := validateMeshProfileHelmAlignment(meshProfileLite, liteDefaults); err != nil {
		t.Fatalf("aligned lite defaults should pass: %v", err)
	}
}

func TestBootstrapClusterConfigMeshProfile_Idempotent(t *testing.T) {
	scheme := runtime.NewScheme()
	c := fake.NewClientBuilder().WithScheme(scheme).Build()
	ctx := context.Background()

	if err := bootstrapClusterConfigMeshProfile(ctx, c, meshProfileLite); err != nil {
		t.Fatalf("create: %v", err)
	}
	if err := bootstrapClusterConfigMeshProfile(ctx, c, meshProfileLite); err != nil {
		t.Fatalf("repeat create: %v", err)
	}

	obj := &unstructured.Unstructured{}
	obj.SetGroupVersionKind(clusterConfigGVK)
	if err := c.Get(ctx, types.NamespacedName{Name: clusterConfigSingleton}, obj); err != nil {
		t.Fatalf("get ClusterConfig: %v", err)
	}
	got, _, err := unstructured.NestedString(obj.Object, "spec", "meshProfile")
	if err != nil || got != meshProfileLite {
		t.Fatalf("meshProfile = %q, want %q (err=%v)", got, meshProfileLite, err)
	}

	if err := bootstrapClusterConfigMeshProfile(ctx, c, meshProfileFull); err != nil {
		t.Fatalf("update to full: %v", err)
	}
	if err := c.Get(ctx, types.NamespacedName{Name: clusterConfigSingleton}, obj); err != nil {
		t.Fatalf("get after update: %v", err)
	}
	got, _, err = unstructured.NestedString(obj.Object, "spec", "meshProfile")
	if err != nil || got != meshProfileFull {
		t.Fatalf("meshProfile after update = %q, want %q (err=%v)", got, meshProfileFull, err)
	}
}

func TestApplyLiteMeshHelmOverrides(t *testing.T) {
	vals := map[string]interface{}{
		"mesh": map[string]interface{}{
			"linkerd": map[string]interface{}{
				"enabled": true,
			},
		},
	}
	applyLiteMeshHelmOverrides(vals)
	mesh := vals["mesh"].(map[string]interface{})
	linkerd := mesh["linkerd"].(map[string]interface{})
	if linkerd["enabled"] != false {
		t.Fatalf("mesh.linkerd.enabled = %v, want false", linkerd["enabled"])
	}
	guardian := vals["linkerdPkiGuardian"].(map[string]interface{})
	if guardian["enabled"] != false {
		t.Fatalf("linkerdPkiGuardian.enabled = %v, want false", guardian["enabled"])
	}
}

