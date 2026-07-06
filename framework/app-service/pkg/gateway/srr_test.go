package gateway

import (
	"github.com/beclab/Olares/framework/app-service/pkg/constants"
	"testing"

	appv1alpha1 "github.com/beclab/api/api/app.bytetrade.io/v1alpha1"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	srrv1alpha1 "github.com/beclab/Olares/framework/app-service/pkg/gateway/v1alpha1"
)

func TestResourceNameForEntrance(t *testing.T) {
	if got := ResourceNameForEntrance("Ab12", "Web"); got != "shared-ab12-web" {
		t.Errorf("ResourceNameForEntrance = %q, want shared-ab12-web", got)
	}
	if got := ResourceNameForEntrance("", "web"); got != "" {
		t.Errorf("empty appid should yield empty name, got %q", got)
	}
}

func TestResourceNameForEntranceApp(t *testing.T) {
	if got := ResourceNameForEntranceApp("Ab12", "Web"); got != "app-ab12-web" {
		t.Errorf("ResourceNameForEntranceApp = %q, want app-ab12-web", got)
	}
	if got := ResourceNameForEntranceApp("", "web"); got != "" {
		t.Errorf("empty appid should yield empty name, got %q", got)
	}
}

func TestIsOptedIn(t *testing.T) {
	app := &appv1alpha1.Application{}
	if IsOptedIn(app) {
		t.Error("no annotation should not be opted in")
	}
	app.Annotations = map[string]string{AnnotationRouteMode: AnnotationRouteModeGateway}
	if !IsOptedIn(app) {
		t.Error("route-mode=gateway should be opted in")
	}
}

func TestBuildSpecForEntrance(t *testing.T) {
	app := &appv1alpha1.Application{
		ObjectMeta: metav1.ObjectMeta{
			Name: "demo",
			Labels: map[string]string{
				constants.AppApiVersionLabel: "v3",
			},
		},

		Spec: appv1alpha1.ApplicationSpec{
			Appid:     "demo1234",
			Name:      "demo",
			Namespace: "demo-shared",
			SharedEntrances: []appv1alpha1.Entrance{
				{Name: "web", Host: "demo-svc", Port: 8080},
			},
		},
	}
	svc := &corev1.Service{
		ObjectMeta: metav1.ObjectMeta{Name: "demo-svc", Namespace: "demo-shared"},
		Spec:       corev1.ServiceSpec{Ports: []corev1.ServicePort{{Port: 8080, Protocol: corev1.ProtocolTCP}}},
	}
	spec, err := BuildSpecForEntrance(app, app.Spec.SharedEntrances[0], 0, len(app.Spec.SharedEntrances), svc, "olares.com",
		srrv1alpha1.EntranceClassShared)
	if err != nil {
		t.Fatalf("BuildSpecForEntrance: %v", err)
	}
	if spec.RouteMode != srrv1alpha1.RouteModeGateway {
		t.Errorf("routeMode = %q, want gateway", spec.RouteMode)
	}
	wantHost := appv1alpha1.SharedEntranceID(app.Spec.Appid, 0) + ".shared.olares.com"
	if len(spec.HostPatterns) != 1 || spec.HostPatterns[0] != wantHost {
		t.Errorf("hostPatterns = %v, want %q", spec.HostPatterns, wantHost)
	}
	if spec.EntranceClass != srrv1alpha1.EntranceClassShared {
		t.Errorf("entranceClass = %q, want %q", spec.EntranceClass, srrv1alpha1.EntranceClassShared)
	}
	if spec.Upstream.ServiceName != "demo-svc" || spec.Upstream.Port != 8080 {
		t.Errorf("upstream = %+v, want demo-svc:8080", spec.Upstream)
	}
}

func TestBuildSpecForEntranceRequiresAppID(t *testing.T) {
	app := &appv1alpha1.Application{
		Spec: appv1alpha1.ApplicationSpec{
			Name:            "demo",
			Namespace:       "demo-shared",
			SharedEntrances: []appv1alpha1.Entrance{{Name: "web"}},
		},
	}
	svc := &corev1.Service{
		ObjectMeta: metav1.ObjectMeta{Name: "demo-svc", Namespace: "demo-shared"},
		Spec:       corev1.ServiceSpec{Ports: []corev1.ServicePort{{Port: 8080, Protocol: corev1.ProtocolTCP}}},
	}
	_, err := BuildSpecForEntrance(app, app.Spec.SharedEntrances[0], 0, len(app.Spec.SharedEntrances), svc, "olares.com",
		srrv1alpha1.EntranceClassShared)
	if err == nil {
		t.Fatal("expected error when app.spec.appid is empty")
	}
}

func TestBuildSpecForEntranceApplication(t *testing.T) {
	app := &appv1alpha1.Application{
		ObjectMeta: metav1.ObjectMeta{Name: "demo"},
		Spec: appv1alpha1.ApplicationSpec{
			Appid:     "demo1234",
			Name:      "demo",
			Namespace: "demo-user",
			Entrances: []appv1alpha1.Entrance{
				{Name: "web", Host: "demo-svc", Port: 8080},
				{Name: "api", Host: "demo-api", Port: 9090},
			},
		},
	}
	svc := &corev1.Service{
		ObjectMeta: metav1.ObjectMeta{Name: "demo-svc", Namespace: "demo-user"},
		Spec:       corev1.ServiceSpec{Ports: []corev1.ServicePort{{Port: 8080, Protocol: corev1.ProtocolTCP}}},
	}

	spec, err := BuildSpecForEntrance(app, app.Spec.Entrances[0], 0, len(app.Spec.Entrances), svc, "olares.com",
		srrv1alpha1.EntranceClassApplication)
	if err != nil {
		t.Fatalf("BuildSpecForEntrance application: %v", err)
	}

	wantHost := appv1alpha1.EntranceID(app.Spec.Appid, 0, len(app.Spec.Entrances)) + ".*.olares.com"
	if len(spec.HostPatterns) != 1 || spec.HostPatterns[0] != wantHost {
		t.Errorf("hostPatterns = %v, want %q", spec.HostPatterns, wantHost)
	}
	if spec.EntranceClass != srrv1alpha1.EntranceClassApplication {
		t.Errorf("entranceClass = %q, want %q", spec.EntranceClass, srrv1alpha1.EntranceClassApplication)
	}
}
