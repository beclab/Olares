package gateway

import (
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
		ObjectMeta: metav1.ObjectMeta{Name: "demo"},
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
	spec, err := BuildSpecForEntrance(app, app.Spec.SharedEntrances[0], 0, svc, "olares.com")
	if err != nil {
		t.Fatalf("BuildSpecForEntrance: %v", err)
	}
	if spec.RouteMode != srrv1alpha1.RouteModeGateway {
		t.Errorf("routeMode = %q, want gateway", spec.RouteMode)
	}
	wantHost := appv1alpha1.SharedEntranceID(app.Spec.Appid, 0, len(app.Spec.SharedEntrances)) + ".shared.olares.com"
	if len(spec.HostPatterns) != 1 || spec.HostPatterns[0] != wantHost {
		t.Errorf("hostPatterns = %v, want %q", spec.HostPatterns, wantHost)
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
	_, err := BuildSpecForEntrance(app, app.Spec.SharedEntrances[0], 0, svc, "olares.com")
	if err == nil {
		t.Fatal("expected error when app.spec.appid is empty")
	}
}
