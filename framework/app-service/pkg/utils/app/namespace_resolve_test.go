package app

import (
	"testing"

	"github.com/beclab/Olares/framework/app-service/pkg/appcfg"
)

func TestResolveInstallNamespace(t *testing.T) {
	cases := []struct {
		name  string
		app   string
		owner string
		cfg   *appcfg.ApplicationConfig
		want  string
	}{
		{
			name:  "normal app ignores any pre-set Namespace",
			app:   "nginx",
			owner: "alice",
			cfg:   &appcfg.ApplicationConfig{AppName: "nginx", Namespace: "custom-or-manifest"},
			want:  "nginx-alice",
		},
		{
			name:  "shared v3",
			app:   "ollama",
			owner: "alice",
			cfg:   &appcfg.ApplicationConfig{AppName: "ollama", APIVersion: appcfg.V3, Shared: true, Namespace: "should-ignore"},
			want:  "ollama-shared",
		},
		{
			name:  "middleware overrides shared/default",
			app:   "pg",
			owner: "alice",
			cfg:   &appcfg.ApplicationConfig{AppName: "pg", Type: "middleware", Namespace: "should-ignore"},
			want:  "pg-middleware",
		},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			got := ResolveInstallNamespace(tc.app, tc.owner, tc.cfg)
			if got != tc.want {
				t.Fatalf("ResolveInstallNamespace()=%q, want %q", got, tc.want)
			}
		})
	}
}
