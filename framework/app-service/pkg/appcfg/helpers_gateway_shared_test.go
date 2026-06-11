package appcfg

import (
	"testing"

	appv1alpha1 "github.com/beclab/api/api/app.bytetrade.io/v1alpha1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	"github.com/beclab/Olares/framework/app-service/pkg/constants"
)

func TestIsGatewaySharedApp(t *testing.T) {
	shared := []appv1alpha1.Entrance{{Name: "api", Host: "svc"}}
	cases := []struct {
		name string
		app  *appv1alpha1.Application
		want bool
	}{
		{
			name: "shared v3",
			app: &appv1alpha1.Application{
				ObjectMeta: metav1.ObjectMeta{Labels: map[string]string{
					constants.AppApiVersionLabel: constants.AppVersionV3,
					constants.AppSharedLabel:     constants.AppSharedTrue,
				}},
				Spec: appv1alpha1.ApplicationSpec{SharedEntrances: shared},
			},
			want: true,
		},
		{
			name: "v3 per-user not gateway shared",
			app: &appv1alpha1.Application{
				ObjectMeta: metav1.ObjectMeta{Labels: map[string]string{constants.AppApiVersionLabel: constants.AppVersionV3}},
				Spec:       appv1alpha1.ApplicationSpec{SharedEntrances: shared},
			},
			want: false,
		},
		{
			name: "v2 cluster scoped",
			app: &appv1alpha1.Application{
				Spec: appv1alpha1.ApplicationSpec{
					SharedEntrances: shared,
					Settings:        map[string]string{"clusterScoped": "true"},
				},
			},
			want: true,
		},
		{
			name: "v2 not cluster scoped",
			app: &appv1alpha1.Application{
				Spec: appv1alpha1.ApplicationSpec{SharedEntrances: shared},
			},
			want: false,
		},
		{
			name: "no shared entrances",
			app: &appv1alpha1.Application{
				ObjectMeta: metav1.ObjectMeta{Labels: map[string]string{constants.AppApiVersionLabel: constants.AppVersionV3}},
			},
			want: false,
		},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			if got := IsGatewaySharedApp(tc.app); got != tc.want {
				t.Fatalf("IsGatewaySharedApp() = %v, want %v", got, tc.want)
			}
		})
	}
}

func TestIsSharedServerApp(t *testing.T) {
	shared := []appv1alpha1.Entrance{{Name: "api", Host: "svc"}}
	sharedLabels := map[string]string{
		constants.AppApiVersionLabel: constants.AppVersionV3,
		constants.AppSharedLabel:     constants.AppSharedTrue,
	}
	cases := []struct {
		name string
		app  *appv1alpha1.Application
		want bool
	}{
		{
			name: "shared v3 with shared entrances qualifies without clusterScoped",
			app: &appv1alpha1.Application{
				ObjectMeta: metav1.ObjectMeta{Labels: sharedLabels},
				Spec:       appv1alpha1.ApplicationSpec{SharedEntrances: shared},
			},
			want: true,
		},
		{
			name: "v3 per-user with shared entrances does not qualify",
			app: &appv1alpha1.Application{
				ObjectMeta: metav1.ObjectMeta{Labels: map[string]string{constants.AppApiVersionLabel: constants.AppVersionV3}},
				Spec:       appv1alpha1.ApplicationSpec{SharedEntrances: shared},
			},
			want: false,
		},
		{
			name: "shared v3 without shared entrances does not qualify",
			app: &appv1alpha1.Application{
				ObjectMeta: metav1.ObjectMeta{Labels: sharedLabels},
			},
			want: false,
		},
		{
			name: "legacy v2 cluster scoped still qualifies",
			app: &appv1alpha1.Application{
				Spec: appv1alpha1.ApplicationSpec{Settings: map[string]string{"clusterScoped": "true"}},
			},
			want: true,
		},
		{
			name: "v2 plain app does not qualify",
			app: &appv1alpha1.Application{
				Spec: appv1alpha1.ApplicationSpec{SharedEntrances: shared},
			},
			want: false,
		},
		{
			name: "nil app",
			app:  nil,
			want: false,
		},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			if got := IsSharedServerApp(tc.app); got != tc.want {
				t.Fatalf("IsSharedServerApp() = %v, want %v", got, tc.want)
			}
		})
	}
}
