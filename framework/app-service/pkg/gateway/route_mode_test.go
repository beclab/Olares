package gateway

import (
	"context"
	"testing"

	appv1alpha1 "github.com/beclab/api/api/app.bytetrade.io/v1alpha1"
	"github.com/beclab/Olares/framework/app-service/pkg/cluster"
	"github.com/beclab/Olares/framework/app-service/pkg/constants"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/runtime/schema"
	clientgoscheme "k8s.io/client-go/kubernetes/scheme"
	"sigs.k8s.io/controller-runtime/pkg/client/fake"
)

func TestComputeRouteModePatch_respectsExplicitAnnotation(t *testing.T) {
	ctx := context.Background()
	app := &appv1alpha1.Application{
		ObjectMeta: metav1.ObjectMeta{
			Annotations: map[string]string{AnnotationRouteMode: AnnotationRouteModeDirect},
		},
		Spec: appv1alpha1.ApplicationSpec{
			SharedEntrances: []appv1alpha1.Entrance{{Name: "api", Host: "svc"}},
			Settings:        map[string]string{"clusterScoped": "true"},
		},
	}
	need, mode, err := ComputeRouteModePatch(ctx, nil, app)
	if err != nil {
		t.Fatal(err)
	}
	if need || mode != AnnotationRouteModeDirect {
		t.Fatalf("got need=%v mode=%q, want need=false mode=direct", need, mode)
	}
}

func TestComputeRouteModePatch_settingsOverride(t *testing.T) {
	ctx := context.Background()
	app := &appv1alpha1.Application{
		Spec: appv1alpha1.ApplicationSpec{
			SharedEntrances: []appv1alpha1.Entrance{{Name: "api", Host: "svc"}},
			Settings: map[string]string{
				"clusterScoped":         "true",
				SettingGatewayRouteMode: AnnotationRouteModeGateway,
			},
		},
	}
	need, mode, err := ComputeRouteModePatch(ctx, nil, app)
	if err != nil || !need || mode != AnnotationRouteModeGateway {
		t.Fatalf("got need=%v mode=%q err=%v", need, mode, err)
	}
}

func TestComputeRouteModePatch_autoGatewayWhenClusterEnabled(t *testing.T) {
	resetGatewayReadyCacheForTest()
	cluster.PrimeSnapshotForTest(cluster.Snapshot{
		PlatformDomain:        "olares.com",
		SharedURLViewerScheme: cluster.SharedURLViewerSchemeEnabled,
	})

	ctx := context.Background()

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

	scheme := runtime.NewScheme()
	_ = clientgoscheme.AddToScheme(scheme)
	c := fake.NewClientBuilder().WithScheme(scheme).WithObjects(gw).Build()

	app := &appv1alpha1.Application{
		ObjectMeta: metav1.ObjectMeta{
			Name: "ollama",
			Labels: map[string]string{constants.AppApiVersionLabel: constants.AppVersionV3},
		},
		Spec: appv1alpha1.ApplicationSpec{
			SharedEntrances: []appv1alpha1.Entrance{{Name: "api", Host: "svc"}},
		},
	}
	need, mode, err := ComputeRouteModePatch(ctx, c, app)
	if err != nil || !need || mode != AnnotationRouteModeGateway {
		t.Fatalf("got need=%v mode=%q err=%v", need, mode, err)
	}
}

func TestApplyRouteModeAnnotation_mutatesApp(t *testing.T) {
	resetGatewayReadyCacheForTest()
	cluster.PrimeSnapshotForTest(cluster.Snapshot{
		PlatformDomain:        "olares.com",
		SharedURLViewerScheme: cluster.SharedURLViewerSchemeEnabled,
	})

	ctx := context.Background()

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

	scheme := runtime.NewScheme()
	_ = clientgoscheme.AddToScheme(scheme)
	c := fake.NewClientBuilder().WithScheme(scheme).WithObjects(gw).Build()

	app := &appv1alpha1.Application{
		Spec: appv1alpha1.ApplicationSpec{
			SharedEntrances: []appv1alpha1.Entrance{{Name: "api", Host: "svc"}},
			Settings:        map[string]string{"clusterScoped": "true"},
		},
	}
	if err := ApplyRouteModeAnnotation(ctx, c, app); err != nil {
		t.Fatal(err)
	}
	if got := app.Annotations[AnnotationRouteMode]; got != AnnotationRouteModeGateway {
		t.Fatalf("annotation = %q, want gateway", got)
	}
}
