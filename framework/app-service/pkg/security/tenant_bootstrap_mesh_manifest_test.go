package security

import (
	"os"
	"testing"

	netv1 "k8s.io/api/networking/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/labels"
	"sigs.k8s.io/yaml"
)

func TestTenantBootstrapWorkloadsCanReachMeshIdentity(t *testing.T) {
	raw, err := os.ReadFile("../../../app-gateway/.olares/config/cluster/deploy/app-gateway-mesh-np-os-mesh.yaml")
	if err != nil {
		t.Fatal(err)
	}
	var policy netv1.NetworkPolicy
	if err := yaml.Unmarshal(raw, &policy); err != nil {
		t.Fatal(err)
	}

	canReachIdentity := func(namespace, pod labels.Set) bool {
		t.Helper()
		for _, rule := range policy.Spec.Ingress {
			for _, port := range rule.Ports {
				if port.Port == nil || port.Port.IntValue() != 8080 {
					continue
				}
				for _, from := range rule.From {
					ns, err := metav1.LabelSelectorAsSelector(from.NamespaceSelector)
					if err != nil {
						t.Fatal(err)
					}
					workload, err := metav1.LabelSelectorAsSelector(from.PodSelector)
					if err != nil {
						t.Fatal(err)
					}
					if ns.Matches(namespace) && (from.PodSelector == nil || workload.Matches(pod)) {
						return true
					}
				}
			}
		}
		return false
	}

	for _, tenant := range []string{"alice", "bob"} {
		t.Run(tenant, func(t *testing.T) {
			namespace := "user-space-" + tenant
			docs, err := renderMeshManifestsInNamespace(t, namespace, true, map[string]any{
				"bfl":       map[string]any{"username": tenant},
				"userspace": map[string]any{"appCache": "/cache", "appData": "/app", "userData": "/user"},
			}, "../../../../apps/.olares/config/user/helm-charts/wizard/templates/wizard_deploy.yaml",
				"../../../../apps/.olares/config/user/helm-charts/system-apps/templates/olares-app.yaml",
				"../../../authelia/.olares/config/user/helm-charts/auth/templates/auth_deploy.yaml")
			if err != nil {
				t.Fatal(err)
			}
			nsLabels := labels.Set{"kubernetes.io/metadata.name": namespace, NamespaceTypeLabel: UserSpace}
			for _, name := range []string{"wizard", "authelia-deployment", "olares-app-deployment"} {
				template := docs["Deployment/"+name]["spec"].(map[string]any)["template"].(map[string]any)
				metadata := template["metadata"].(map[string]any)
				podLabels := labels.Set{}
				for k, v := range metadata["labels"].(map[string]any) {
					podLabels[k] = v.(string)
				}
				annotations := metadata["annotations"].(map[string]any)
				if annotations["linkerd.io/inject"] != "enabled" || annotations["config.linkerd.io/proxy-enable-native-sidecar"] != "true" ||
					!canReachIdentity(nsLabels, podLabels) {
					t.Fatalf("%s cannot bootstrap its Linkerd proxy", name)
				}
				delete(podLabels, "traffic.olares.io/platform-mesh")
				if canReachIdentity(nsLabels, podLabels) {
					t.Fatalf("%s unexpectedly reaches the identity service without enrollment", name)
				}
				if name == "wizard" {
					containers := template["spec"].(map[string]any)["containers"].([]any)
					probe := containers[0].(map[string]any)["readinessProbe"].(map[string]any)["httpGet"].(map[string]any)
					if probe["port"] != 80 || probe["path"] != "/" {
						t.Fatalf("Wizard readiness does not exercise the web server: %v", probe)
					}
				}
			}
		})
	}
}
