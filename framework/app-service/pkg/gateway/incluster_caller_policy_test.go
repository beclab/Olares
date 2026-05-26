package gateway

import (
	"testing"

	appv1alpha1 "github.com/beclab/api/api/app.bytetrade.io/v1alpha1"
)

func TestComputeCallerInClusterPatch(t *testing.T) {
	app := &appv1alpha1.Application{
		Spec: appv1alpha1.ApplicationSpec{
			Settings: map[string]string{"clusterAppRef": "ollamav2"},
		},
	}
	need, v := ComputeCallerInClusterPatch(app)
	if !need || v != InClusterGateway {
		t.Fatalf("need=%v v=%q", need, v)
	}
	app.Annotations = map[string]string{AnnotationInCluster: InClusterGateway}
	if need, _ := ComputeCallerInClusterPatch(app); need {
		t.Fatal("already gateway")
	}
	app.Annotations[AnnotationInCluster] = InClusterDirect
	if need, _ := ComputeCallerInClusterPatch(app); need {
		t.Fatal("respect explicit direct")
	}
}

func TestComputeCallerInClusterPatch_fromSettings(t *testing.T) {
	app := &appv1alpha1.Application{
		Spec: appv1alpha1.ApplicationSpec{
			Settings: map[string]string{
				"clusterAppRef": "ollamav2",
				SettingInClusterMode: InClusterDirect,
			},
		},
	}
	need, v := ComputeCallerInClusterPatch(app)
	if !need || v != InClusterDirect {
		t.Fatalf("need=%v v=%q", need, v)
	}
}

func TestApplyCallerInClusterAnnotation(t *testing.T) {
	app := &appv1alpha1.Application{
		Spec: appv1alpha1.ApplicationSpec{
			Settings: map[string]string{"clusterAppRef": "ollamav2"},
		},
	}
	ApplyCallerInClusterAnnotation(app)
	if app.Annotations[AnnotationInCluster] != InClusterGateway {
		t.Fatalf("got %q", app.Annotations[AnnotationInCluster])
	}
}
