package gateway

import (
	"context"
	"testing"

	appv1alpha1 "github.com/beclab/api/api/app.bytetrade.io/v1alpha1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/runtime/schema"
	clientgoscheme "k8s.io/client-go/kubernetes/scheme"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/client/fake"

	"github.com/beclab/Olares/framework/app-service/pkg/cluster"
	"github.com/beclab/Olares/framework/app-service/pkg/constants"
)

func newSharedApp() *appv1alpha1.Application {
	return &appv1alpha1.Application{
		ObjectMeta: metav1.ObjectMeta{
			Name: "ollama",
			Labels: map[string]string{
				constants.AppApiVersionLabel: constants.AppVersionV3,
				constants.AppSharedLabel:     constants.AppSharedTrue,
			},
		},
		Spec: appv1alpha1.ApplicationSpec{
			Name:            "ollama",
			Namespace:       "ollama",
			SharedEntrances: []appv1alpha1.Entrance{{Name: "api", Host: "svc"}},
		},
	}
}

func newAcceptedGateway() *unstructured.Unstructured {
	gw := &unstructured.Unstructured{Object: map[string]interface{}{
		"apiVersion": "gateway.networking.k8s.io/v1",
		"kind":       "Gateway",
		"metadata": map[string]interface{}{
			"name":      defaultGatewayName,
			"namespace": defaultGatewayNS,
		},
		"status": map[string]interface{}{
			"conditions": []interface{}{
				map[string]interface{}{"type": "Accepted", "status": "True"},
			},
		},
	}}
	gw.SetGroupVersionKind(schema.GroupVersionKind{Group: "gateway.networking.k8s.io", Version: "v1", Kind: "Gateway"})
	return gw
}

func newFakeClient(t *testing.T, objs ...client.Object) client.Client {
	t.Helper()
	scheme := runtime.NewScheme()
	if err := clientgoscheme.AddToScheme(scheme); err != nil {
		t.Fatal(err)
	}
	return fake.NewClientBuilder().WithScheme(scheme).WithObjects(objs...).Build()
}

func TestComputeRouteModePatch_respectsExplicitAnnotation(t *testing.T) {
	ctx := context.Background()
	for _, pinned := range []string{AnnotationRouteModeGateway, AnnotationRouteModeDirect} {
		app := newSharedApp()
		app.Annotations = map[string]string{AnnotationRouteMode: pinned}
		need, mode, err := ComputeRouteModePatch(ctx, nil, app)
		if err != nil {
			t.Fatal(err)
		}
		if need || mode != pinned {
			t.Fatalf("pinned %q: got need=%v mode=%q, want need=false mode=%q", pinned, need, mode, pinned)
		}
	}
}

func TestComputeRouteModePatch_settingsOverride(t *testing.T) {
	ctx := context.Background()
	app := newSharedApp()
	app.Spec.Settings = map[string]string{SettingGatewayRouteMode: AnnotationRouteModeGateway}
	need, mode, err := ComputeRouteModePatch(ctx, nil, app)
	if err != nil || !need || mode != AnnotationRouteModeGateway {
		t.Fatalf("got need=%v mode=%q err=%v", need, mode, err)
	}

	app = newSharedApp()
	app.Spec.Settings = map[string]string{SettingGatewayRouteMode: AnnotationRouteModeDirect}
	need, mode, err = ComputeRouteModePatch(ctx, nil, app)
	if err != nil || !need || mode != AnnotationRouteModeDirect {
		t.Fatalf("got need=%v mode=%q err=%v", need, mode, err)
	}
}

func TestComputeRouteModePatch_autoGatewayForSharedApp(t *testing.T) {
	resetGatewayReadyCacheForTest()
	cluster.PrimeInClusterGatewayEnabledForTest(true)
	defer cluster.ResetInClusterGatewayEnabledForTest()

	ctx := context.Background()
	c := newFakeClient(t, newAcceptedGateway())

	need, mode, err := ComputeRouteModePatch(ctx, c, newSharedApp())
	if err != nil || !need || mode != AnnotationRouteModeGateway {
		t.Fatalf("got need=%v mode=%q err=%v", need, mode, err)
	}
}

func TestComputeRouteModePatch_skipsNonSharedApp(t *testing.T) {
	resetGatewayReadyCacheForTest()
	cluster.PrimeInClusterGatewayEnabledForTest(true)
	defer cluster.ResetInClusterGatewayEnabledForTest()

	ctx := context.Background()
	c := newFakeClient(t, newAcceptedGateway())

	// Per-user app: no options.shared label, no clusterScoped setting.
	app := newSharedApp()
	app.Labels = map[string]string{constants.AppApiVersionLabel: constants.AppVersionV3}
	need, mode, err := ComputeRouteModePatch(ctx, c, app)
	if err != nil || need || mode != "" {
		t.Fatalf("non-shared app: got need=%v mode=%q err=%v, want no patch", need, mode, err)
	}

	// Shared label without sharedEntrances still qualifies (entrances-only shared singleton).
	app = newSharedApp()
	app.Spec.SharedEntrances = nil
	need, mode, err = ComputeRouteModePatch(ctx, c, app)
	if err != nil || !need || mode != AnnotationRouteModeGateway {
		t.Fatalf("shared without sharedEntrances: got need=%v mode=%q err=%v, want gateway patch", need, mode, err)
	}
}

func TestComputeRouteModePatch_gateDisabled(t *testing.T) {
	resetGatewayReadyCacheForTest()
	cluster.PrimeInClusterGatewayEnabledForTest(false)
	defer cluster.ResetInClusterGatewayEnabledForTest()

	ctx := context.Background()
	c := newFakeClient(t, newAcceptedGateway())

	need, mode, err := ComputeRouteModePatch(ctx, c, newSharedApp())
	if err != nil || need || mode != "" {
		t.Fatalf("gate=false: got need=%v mode=%q err=%v, want no patch", need, mode, err)
	}
}

func TestComputeRouteModePatch_gatewayNotReady(t *testing.T) {
	resetGatewayReadyCacheForTest()
	cluster.PrimeInClusterGatewayEnabledForTest(true)
	defer cluster.ResetInClusterGatewayEnabledForTest()

	ctx := context.Background()
	c := newFakeClient(t) // no Gateway object

	need, mode, err := ComputeRouteModePatch(ctx, c, newSharedApp())
	if err != nil || need || mode != "" {
		t.Fatalf("gateway absent: got need=%v mode=%q err=%v, want no patch", need, mode, err)
	}
}

func TestApplyRouteModeAnnotation_mutatesApp(t *testing.T) {
	resetGatewayReadyCacheForTest()
	cluster.PrimeInClusterGatewayEnabledForTest(true)
	defer cluster.ResetInClusterGatewayEnabledForTest()

	ctx := context.Background()
	c := newFakeClient(t, newAcceptedGateway())

	app := newSharedApp()
	if err := ApplyRouteModeAnnotation(ctx, c, app); err != nil {
		t.Fatal(err)
	}
	if got := app.Annotations[AnnotationRouteMode]; got != AnnotationRouteModeGateway {
		t.Fatalf("annotation = %q, want gateway", got)
	}
}
