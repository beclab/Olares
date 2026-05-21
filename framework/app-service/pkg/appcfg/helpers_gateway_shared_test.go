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
			name: "v3",
			app: &appv1alpha1.Application{
				ObjectMeta: metav1.ObjectMeta{Labels: map[string]string{constants.AppApiVersionLabel: constants.AppVersionV3}},
				Spec:       appv1alpha1.ApplicationSpec{SharedEntrances: shared},
			},
			want: true,
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
