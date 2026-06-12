package appcfg

import (
	"testing"

	"github.com/beclab/Olares/framework/app-service/pkg/constants"
	appv1alpha1 "github.com/beclab/api/api/app.bytetrade.io/v1alpha1"
)

func TestIsGatewaySharedApp(t *testing.T) {
	t.Run("nil app", func(t *testing.T) {
		if IsGatewaySharedApp(nil) {
			t.Fatal("nil app must not be treated as gateway shared")
		}
	})

	t.Run("no shared entrances", func(t *testing.T) {
		app := &appv1alpha1.Application{}
		if IsGatewaySharedApp(app) {
			t.Fatal("app without shared entrances must not be treated as gateway shared")
		}
	})

	t.Run("shared app with shared entrance", func(t *testing.T) {
		app := &appv1alpha1.Application{
			Spec: appv1alpha1.ApplicationSpec{
				SharedEntrances: []appv1alpha1.Entrance{{Name: "web"}},
			},
		}
		app.Labels = map[string]string{constants.AppSharedLabel: constants.AppSharedTrue}

		if !IsGatewaySharedApp(app) {
			t.Fatal("shared app with shared entrance should be treated as gateway shared")
		}
	})

	t.Run("cluster scoped app with shared entrance", func(t *testing.T) {
		app := &appv1alpha1.Application{
			Spec: appv1alpha1.ApplicationSpec{
				Settings: map[string]string{
					"clusterScoped": "true",
				},
				SharedEntrances: []appv1alpha1.Entrance{{Name: "api"}},
			},
		}
		if !IsGatewaySharedApp(app) {
			t.Fatal("cluster scoped app with shared entrance should be treated as gateway shared")
		}
	})
}

func TestLogicalHostPattern(t *testing.T) {
	t.Run("single shared entrance without index suffix", func(t *testing.T) {
		got, err := LogicalHostPattern("demo1234", 0, 1, "olares.com")
		if err != nil {
			t.Fatalf("LogicalHostPattern() error = %v", err)
		}
		want := appv1alpha1.SharedEntranceID("demo1234", 0, 1) + ".shared.olares.com"
		if got != want {
			t.Fatalf("LogicalHostPattern() = %q, want %q", got, want)
		}
	})

	t.Run("multiple shared entrances use indexed id", func(t *testing.T) {
		got, err := LogicalHostPattern("demo1234", 1, 2, "olares.com.")
		if err != nil {
			t.Fatalf("LogicalHostPattern() error = %v", err)
		}
		want := appv1alpha1.SharedEntranceID("demo1234", 1, 2) + ".shared.olares.com"
		if got != want {
			t.Fatalf("LogicalHostPattern() = %q, want %q", got, want)
		}
	})

	t.Run("invalid inputs", func(t *testing.T) {
		cases := []struct {
			name           string
			appid          string
			entranceIndex  int
			entranceCount  int
			platformDomain string
		}{
			{
				name:           "empty appid",
				appid:          "",
				entranceIndex:  0,
				entranceCount:  1,
				platformDomain: "olares.com",
			},
			{
				name:           "invalid index",
				appid:          "demo1234",
				entranceIndex:  1,
				entranceCount:  1,
				platformDomain: "olares.com",
			},
			{
				name:           "empty platform domain",
				appid:          "demo1234",
				entranceIndex:  0,
				entranceCount:  1,
				platformDomain: "",
			},
		}
		for _, tc := range cases {
			t.Run(tc.name, func(t *testing.T) {
				if _, err := LogicalHostPattern(tc.appid, tc.entranceIndex, tc.entranceCount, tc.platformDomain); err == nil {
					t.Fatalf("LogicalHostPattern(%q, %d, %d, %q) expected error", tc.appid, tc.entranceIndex, tc.entranceCount,
						tc.platformDomain)
				}
			})
		}
	})
}
