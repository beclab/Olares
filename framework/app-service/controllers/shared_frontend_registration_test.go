package controllers

import (
	"context"
	"encoding/json"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/beclab/Olares/framework/app-service/pkg/constants"
	appv1 "github.com/beclab/api/api/app.bytetrade.io/v1alpha1"
	"helm.sh/helm/v3/pkg/chart"
	"helm.sh/helm/v3/pkg/chartutil"
	"helm.sh/helm/v3/pkg/engine"
	appsv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"sigs.k8s.io/controller-runtime/pkg/client/fake"
	"sigs.k8s.io/yaml"
)

// Render the actual platform templates, then feed their metadata and Services
// into the controller. This catches lost registration annotations during moves
// between the per-user chart and the cluster chart.
func TestSharedFrontendRegistration(t *testing.T) {
	t.Setenv("SYS_APPS", "olares-app")
	chartDir := filepath.Join("..", "..", "..", "apps", ".olares", "config", "cluster", "deploy", "os-frontend")
	ch := &chart.Chart{Metadata: &chart.Metadata{Name: "test", Version: "0.1.0", APIVersion: "v2"}}
	for _, name := range []string{"olares-app.yaml", "olares-app-nginx.yaml"} {
		data, err := os.ReadFile(filepath.Join(chartDir, name))
		if err != nil {
			t.Fatal(err)
		}
		ch.Templates = append(ch.Templates, &chart.File{Name: "templates/" + name, Data: data})
	}
	values, err := chartutil.ToRenderValues(ch, map[string]interface{}{}, chartutil.ReleaseOptions{Name: "os-framework", Namespace: "os-framework", IsInstall: true}, nil)
	if err != nil {
		t.Fatal(err)
	}
	rendered, err := engine.Render(ch, values)
	if err != nil {
		t.Fatal(err)
	}
	scheme := runtime.NewScheme()
	if err := corev1.AddToScheme(scheme); err != nil {
		t.Fatal(err)
	}
	builder := fake.NewClientBuilder().WithScheme(scheme)
	var deployment appsv1.Deployment
	for _, doc := range strings.Split(rendered["test/templates/olares-app.yaml"], "\n---\n") {
		var meta struct {
			Kind string `json:"kind"`
		}
		if err := yaml.Unmarshal([]byte(doc), &meta); err != nil {
			t.Fatal(err)
		}
		switch meta.Kind {
		case "Service":
			var service corev1.Service
			if err := yaml.Unmarshal([]byte(doc), &service); err != nil {
				t.Fatal(err)
			}
			builder.WithObjects(&service)
		case "Deployment":
			if err := yaml.Unmarshal([]byte(doc), &deployment); err != nil {
				t.Fatal(err)
			}
		}
	}
	if !isApp(&deployment) || deployment.Namespace != "os-frontend" || deployment.Labels[constants.ApplicationOwnerLabel] != "olares" || !appv1.IsShared(&deployment) || !appv1.IsV3(&deployment) {
		t.Fatalf("invalid registration metadata: %+v", deployment.ObjectMeta)
	}
	r := &ApplicationReconciler{Client: builder.Build()} // Only Services; no User CR.
	entrances, err := r.getEntranceServiceAddress(context.Background(), &deployment, false)
	if err != nil {
		t.Fatal(err)
	}
	es := entrances["olares-app"]
	if len(es) != 9 {
		t.Fatalf("got %d entrances, want 9", len(es))
	}
	settings, _ := r.getAppSettings(context.Background(), "olares-app", "olares-app", "olares", &deployment, false, es)
	var domains []appv1.DefaultThirdLevelDomainConfig
	if err := json.Unmarshal([]byte(settings["defaultThirdLevelDomainConfig"]), &domains); err != nil {
		t.Fatal(err)
	}
	if len(domains) != len(es) {
		t.Fatalf("domains: %v", domains)
	}
	domainByEntrance := map[string]string{}
	for _, domain := range domains {
		domainByEntrance[domain.EntranceName] = domain.ThirdLevelDomain
	}
	var policies map[string]applicationSettingsPolicy
	if err := json.Unmarshal([]byte(settings[applicationSettingsPolicyKey]), &policies); err != nil {
		t.Fatal(err)
	}
	for _, entrance := range es {
		if domainByEntrance[entrance.Name] != entrance.Name {
			t.Fatalf("missing domain for %s", entrance.Name)
		}
		wantPolicy := "system"
		if entrance.Name == "share" {
			wantPolicy = "public"
			if !entrance.Invisible {
				t.Fatal("share must remain invisible")
			}
		}
		if policies[entrance.Name].DefaultPolicy != wantPolicy {
			t.Fatalf("unexpected policy for %s: %+v", entrance.Name, policies[entrance.Name])
		}
		if entrance.Name == "settings" && (entrance.Host != "settings-service" || entrance.Port != 80) {
			t.Fatalf("settings upstream: %+v", entrance)
		}
	}
	headscale := policies["headscale"].SubPolicies
	if len(headscale) != 1 || headscale[0].URI != "/api/refresh" || headscale[0].Policy != "public" {
		t.Fatalf("headscale refresh policy: %+v", headscale)
	}
	// Incomplete Service registration must surface an error for reconciliation to retry.
	broken := deployment.DeepCopy()
	broken.Namespace = "user-space-olares"
	if _, err := r.getEntranceServiceAddress(context.Background(), broken, false); err == nil {
		t.Fatal("accepted entrances in a namespace without Services")
	}
}
