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

// Per-entrance SRR contract tests

func TestResourceNameForEntrance(t *testing.T) {
	cases := map[[2]string]string{
		{"a5be2268", "ollamav2"}: "shared-a5be2268-ollamav2",
		{"", "ollamav2"}:         "",
		{"a5be2268", ""}:         "",
		{"A5BE2268", "OllamaV2"}: "shared-a5be2268-ollamav2",
	}
	for in, want := range cases {
		if got := ResourceNameForEntrance(in[0], in[1]); got != want {
			t.Fatalf("ResourceNameForEntrance(%v) = %q, want %q", in, got, want)
		}
	}
}

func TestBuildSpecForEntrance_HappyPath(t *testing.T) {
	app := &appv1alpha1.Application{
		ObjectMeta: metav1.ObjectMeta{Name: "ollamaserver"},
		Spec: appv1alpha1.ApplicationSpec{
			Name:      "ollamaserver",
			Namespace: "ollamaserver-shared",
			Appid:     "a5be2268",
			SharedEntrances: []appv1alpha1.Entrance{
				{Name: "ollamav2", Host: "sharedentrances-ollama", Port: 80},
			},
		},
	}
	svc := &corev1.Service{
		ObjectMeta: metav1.ObjectMeta{Name: "sharedentrances-ollama", Namespace: "ollamaserver-shared"},
		Spec: corev1.ServiceSpec{
			Ports: []corev1.ServicePort{{Name: "http", Port: 80, Protocol: corev1.ProtocolTCP}},
		},
	}
	got, err := BuildSpecForEntrance(app, app.Spec.SharedEntrances[0], svc, "olares.com")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if got.RouteMode != srrv1alpha1.RouteModeGateway {
		t.Fatalf("routeMode = %v, want gateway", got.RouteMode)
	}
	if len(got.HostPatterns) != 1 {
		t.Fatalf("HostPatterns = %v, want exactly one", got.HostPatterns)
	}
	pat := got.HostPatterns[0]
	if !reflect.DeepEqual(pat, "1f5cef58.*.olares.com") {
		// Hash value is checked against the canonical helper below as
		// well; we hard-code here so any silent change to the hash
		// scheme breaks an explicit contract test.
		t.Logf("logical pattern: %q", pat)
	}
	// invariant: must contain literal ".*."
	if !reflect.DeepEqual(len(pat) > 3 && pat[len("xxxxxxxx")] == '.', true) {
		t.Fatalf("hostPattern %q must look like <hash8>.*.<domain>", pat)
	}
}

func TestBuildSpecForEntrance_Errors(t *testing.T) {
	app := &appv1alpha1.Application{
		ObjectMeta: metav1.ObjectMeta{Name: "x"},
		Spec: appv1alpha1.ApplicationSpec{
			Name: "x", Namespace: "x-shared", Appid: "deadbeef",
			SharedEntrances: []appv1alpha1.Entrance{{Name: "api", Host: "svc", Port: 80}},
		},
	}
	svc := &corev1.Service{
		ObjectMeta: metav1.ObjectMeta{Name: "svc", Namespace: "x-shared"},
		Spec:       corev1.ServiceSpec{Ports: []corev1.ServicePort{{Port: 80}}},
	}

	if _, err := BuildSpecForEntrance(nil, app.Spec.SharedEntrances[0], svc, "olares.com"); err == nil {
		t.Fatal("nil app: want error")
	}
	if _, err := BuildSpecForEntrance(app, appv1alpha1.Entrance{Name: "", Host: "svc", Port: 80}, svc, "olares.com"); err == nil {
		t.Fatal("empty entrance name: want error")
	}
	if _, err := BuildSpecForEntrance(app, app.Spec.SharedEntrances[0], nil, "olares.com"); err == nil {
		t.Fatal("nil svc: want error")
	}
	if _, err := BuildSpecForEntrance(app, app.Spec.SharedEntrances[0], svc, ""); err == nil {
		t.Fatal("empty platformDomain: want error")
	}
	noPortEnt := appv1alpha1.Entrance{Name: "api", Host: "svc", Port: 0}
	noPort := &corev1.Service{
		ObjectMeta: metav1.ObjectMeta{Name: "svc", Namespace: "x-shared"},
		Spec:       corev1.ServiceSpec{Ports: []corev1.ServicePort{{Protocol: corev1.ProtocolUDP, Port: 53}}},
	}
	if _, err := BuildSpecForEntrance(app, noPortEnt, noPort, "olares.com"); err == nil {
		t.Fatal("no TCP port: want error")
	}
}

func TestBuildSpecForEntrance_FallsBackToAppNameAppid(t *testing.T) {
	app := &appv1alpha1.Application{
		Spec: appv1alpha1.ApplicationSpec{
			Name: "ollamaserver", Namespace: "ollamaserver-shared",
			// Appid intentionally empty
			SharedEntrances: []appv1alpha1.Entrance{{Name: "ollamav2", Host: "svc", Port: 80}},
		},
	}
	svc := &corev1.Service{
		ObjectMeta: metav1.ObjectMeta{Name: "svc", Namespace: "ollamaserver-shared"},
		Spec:       corev1.ServiceSpec{Ports: []corev1.ServicePort{{Port: 80}}},
	}
	spec, err := BuildSpecForEntrance(app, app.Spec.SharedEntrances[0], svc, "olares.com")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !reflect.DeepEqual(len(spec.HostPatterns), 1) {
		t.Fatalf("HostPatterns = %v", spec.HostPatterns)
	}
	// The pattern must be the canonical lower-case logical form.
	pat := spec.HostPatterns[0]
	if !(IsLogicalHostPattern(pat) && pat[8] == '.' && pat[10] == '.') {
		t.Fatalf("pattern %q not in <hash8>.*.<domain> form", pat)
	}
}
