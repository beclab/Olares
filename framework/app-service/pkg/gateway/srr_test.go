package gateway

import (
	"reflect"
	"testing"

	appv1alpha1 "github.com/beclab/api/api/app.bytetrade.io/v1alpha1"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	srrv1alpha1 "github.com/beclab/Olares/framework/app-service/pkg/gateway/v1alpha1"
)

func TestResourceName(t *testing.T) {
	if got, want := ResourceName("ollama"), "shared-ollama"; got != want {
		t.Fatalf("ResourceName: got %q want %q", got, want)
	}
}

func TestIsOptedIn(t *testing.T) {
	cases := []struct {
		name string
		app  *appv1alpha1.Application
		want bool
	}{
		{name: "nil", app: nil, want: false},
		{name: "no annotations", app: &appv1alpha1.Application{}, want: false},
		{name: "wrong value", app: &appv1alpha1.Application{ObjectMeta: metav1.ObjectMeta{Annotations: map[string]string{AnnotationRouteMode: "direct"}}}, want: false},
		{name: "gateway", app: &appv1alpha1.Application{ObjectMeta: metav1.ObjectMeta{Annotations: map[string]string{AnnotationRouteMode: AnnotationRouteModeGateway}}}, want: true},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			if got := IsOptedIn(tc.app); got != tc.want {
				t.Fatalf("IsOptedIn: got %v want %v", got, tc.want)
			}
		})
	}
}

func TestBuildSpec(t *testing.T) {
	app := &appv1alpha1.Application{
		ObjectMeta: metav1.ObjectMeta{Name: "ollama-shared-ollama"},
		Spec: appv1alpha1.ApplicationSpec{
			Name:      "ollama",
			Namespace: "ollama-shared",
			SharedEntrances: []appv1alpha1.Entrance{
				{Name: "api", Host: "ollama", Port: 11434, URL: "ABC.SHARED.example.com:11434"},
				{Name: "ui", Host: "ollama-ui", Port: 11435, URL: "abc.shared.example.com"},
			},
		},
	}
	svc := &corev1.Service{
		ObjectMeta: metav1.ObjectMeta{Name: "ollama", Namespace: "ollama-shared"},
		Spec: corev1.ServiceSpec{
			Ports: []corev1.ServicePort{
				{Name: "http", Port: 11434, Protocol: corev1.ProtocolTCP},
			},
		},
	}
	got, err := BuildSpec(app, svc)
	if err != nil {
		t.Fatalf("BuildSpec: %v", err)
	}
	wantHosts := []string{"abc.shared.example.com"}
	if !reflect.DeepEqual(got.HostPatterns, wantHosts) {
		t.Fatalf("hostPatterns = %v, want %v (lowercase + no port + dedup)", got.HostPatterns, wantHosts)
	}
	if got.RouteMode != srrv1alpha1.RouteModeGateway {
		t.Fatalf("routeMode = %v, want gateway", got.RouteMode)
	}
	if got.Upstream.ServiceName != "ollama" || got.Upstream.ServiceNamespace != "ollama-shared" {
		t.Fatalf("upstream service mismatch: %+v", got.Upstream)
	}
	if got.Upstream.Port != 11434 {
		t.Fatalf("upstream port = %d, want 11434", got.Upstream.Port)
	}
	if got.AuthzRef == nil || got.AuthzRef.DefaultAction != srrv1alpha1.AuthzDefaultAllow {
		t.Fatalf("authzRef = %+v, want defaultAction=allow", got.AuthzRef)
	}
}

func TestBuildSpec_Errors(t *testing.T) {
	good := &appv1alpha1.Application{
		Spec: appv1alpha1.ApplicationSpec{
			Name:      "ollama",
			Namespace: "ollama-shared",
			SharedEntrances: []appv1alpha1.Entrance{
				{Name: "api", Host: "ollama", Port: 11434, URL: "abc.shared.example.com"},
			},
		},
	}
	svc := &corev1.Service{
		ObjectMeta: metav1.ObjectMeta{Name: "ollama", Namespace: "ollama-shared"},
		Spec: corev1.ServiceSpec{Ports: []corev1.ServicePort{{Name: "http", Port: 80}}},
	}

	if _, err := BuildSpec(nil, svc); err == nil {
		t.Fatal("nil app: want error")
	}
	if _, err := BuildSpec(good, nil); err == nil {
		t.Fatal("nil svc: want error")
	}

	noShared := good.DeepCopy()
	noShared.Spec.SharedEntrances = nil
	if _, err := BuildSpec(noShared, svc); err == nil {
		t.Fatal("no shared entrances: want error")
	}

	badHost := good.DeepCopy()
	badHost.Spec.SharedEntrances[0].URL = "http://bad"
	badHost.Spec.SharedEntrances[0].Host = ""
	if _, err := BuildSpec(badHost, svc); err == nil {
		t.Fatal("invalid host: want error")
	}
}
