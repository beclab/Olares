package meshinagent

import (
	"testing"

	appv1alpha1 "github.com/beclab/api/api/app.bytetrade.io/v1alpha1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

func TestApplicationDeclaresSharedAccess(t *testing.T) {
	cases := []struct {
		name string
		app  *appv1alpha1.Application
		want bool
	}{
		{name: "nil app", app: nil, want: false},
		{
			name: "needsSharedAccess alone must not inject",
			app: &appv1alpha1.Application{
				Spec: appv1alpha1.ApplicationSpec{
					Settings: map[string]string{SettingNeedsSharedAccess: "true"},
				},
			},
			want: false,
		},
		{
			name: "sharedAppDeps",
			app: &appv1alpha1.Application{
				Spec: appv1alpha1.ApplicationSpec{
					Settings: map[string]string{SettingSharedAppDeps: "demo"},
				},
			},
			want: true,
		},
		{
			name: "clusterAppRef",
			app: &appv1alpha1.Application{
				Spec: appv1alpha1.ApplicationSpec{
					Settings: map[string]string{SettingClusterAppRef: "shared-demo"},
				},
			},
			want: true,
		},
		{
			name: "no deps",
			app: &appv1alpha1.Application{
				ObjectMeta: metav1.ObjectMeta{Name: "plain"},
				Spec:       appv1alpha1.ApplicationSpec{Settings: map[string]string{}},
			},
			want: false,
		},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			if got := ApplicationDeclaresSharedAccess(tc.app); got != tc.want {
				t.Fatalf("ApplicationDeclaresSharedAccess() = %v, want %v", got, tc.want)
			}
		})
	}
}

func TestShouldInjectMeshInAgent(t *testing.T) {
	intentOnly := &appv1alpha1.Application{
		Spec: appv1alpha1.ApplicationSpec{
			Settings: map[string]string{SettingNeedsSharedAccess: "true"},
		},
	}
	if ShouldInject(intentOnly, false) {
		t.Fatal("needsSharedAccess alone must not inject")
	}
	consumer := &appv1alpha1.Application{
		Spec: appv1alpha1.ApplicationSpec{
			Settings: map[string]string{SettingSharedAppDeps: "ollama"},
		},
	}
	if !ShouldInject(consumer, false) {
		t.Fatal("expected inject for named callee")
	}
	if ShouldInject(consumer, true) {
		t.Fatal("shared provider app must not receive mesh-in agent")
	}
	if ShouldInject(nil, false) {
		t.Fatal("nil app must not inject")
	}
}

func TestContainerSpecFailClosed(t *testing.T) {
	c := ContainerSpec()
	if c.Name != ContainerName {
		t.Fatalf("name = %q", c.Name)
	}
	foundFailClosed := false
	for _, env := range c.Env {
		if env.Name == FailClosedEnv && env.Value == "true" {
			foundFailClosed = true
		}
	}
	if !foundFailClosed {
		t.Fatalf("env missing %s=true: %#v", FailClosedEnv, c.Env)
	}
	foundJWT := false
	for _, m := range c.VolumeMounts {
		if m.MountPath == JWTSecretMountPath {
			foundJWT = true
			break
		}
	}
	if !foundJWT {
		t.Fatalf("missing JWT mount %s in %#v", JWTSecretMountPath, c.VolumeMounts)
	}
	for _, m := range c.VolumeMounts {
		if m.MountPath == ConfMountPath {
			t.Fatalf("must not mount empty ConfVolume over %s until seed/render is wired", ConfMountPath)
		}
	}
	v := JWTSecretVolume()
	if v.Secret == nil || v.Secret.SecretName != "caller-jwt" {
		t.Fatalf("JWT volume must mount caller-jwt secret, got %#v", v.Secret)
	}
}

func TestConfSeedInitContainerSpec(t *testing.T) {
	c := ConfSeedInitContainerSpec()
	if c.Name == "" || c.Image == "" {
		t.Fatalf("seed init incomplete: %#v", c)
	}
	found := false
	for _, m := range c.VolumeMounts {
		if m.Name == ConfVolumeName && m.MountPath == "/conf" {
			found = true
		}
	}
	if !found {
		t.Fatalf("seed init must mount %s at /conf: %#v", ConfVolumeName, c.VolumeMounts)
	}
}
