package oac_test

import (
	"reflect"
	"testing"

	appv1 "github.com/beclab/api/api/app.bytetrade.io/v1alpha1"
	"github.com/beclab/api/manifest"

	"github.com/beclab/Olares/framework/oac"
	olm "github.com/beclab/Olares/framework/oac/internal/manifest"
)

func TestAppConfiguration_SameTypeAsInternal(t *testing.T) {
	var a *oac.AppConfiguration
	var b *olm.AppConfiguration
	if reflect.TypeOf(a) != reflect.TypeOf(b) {
		t.Fatalf("oac.AppConfiguration (%v) and internal manifest AppConfiguration (%v) must be the same type",
			reflect.TypeOf(a), reflect.TypeOf(b))
	}
}

func TestAppConfiguration_SameTypeAsAPIManifest(t *testing.T) {
	var a *oac.AppConfiguration
	var m *manifest.AppConfiguration
	if reflect.TypeOf(a) != reflect.TypeOf(m) {
		t.Fatalf("oac.AppConfiguration (%v) and manifest.AppConfiguration (%v) must be the same type",
			reflect.TypeOf(a), reflect.TypeOf(m))
	}
}

func TestAppConfiguration_Constructible(t *testing.T) {
	cfg := &oac.AppConfiguration{
		ConfigVersion: "0.12.0",
		ConfigType:    "app",
		APIVersion:    olm.APIVersionV1,
		Metadata: manifest.AppMetaData{
			Name:    "demo",
			Title:   "Demo",
			Version: "1.0.0",
		},
		Entrances: []appv1.Entrance{
			{Name: "web", Host: "demo", Port: 8080},
		},
		Spec: manifest.AppSpec{
			SubCharts: []manifest.Chart{
				{Name: "extra", Shared: true},
			},
			Resources: []manifest.ResourceMode{
				{
					Mode: "cpu",
					ResourceRequirement: manifest.ResourceRequirement{
						RequiredCPU:    "100m",
						LimitedCPU:     "200m",
						RequiredMemory: "128Mi",
						LimitedMemory:  "256Mi",
					},
				},
			},
		},
		Options: manifest.Options{
			Upload: manifest.Upload{Dest: "/tmp", LimitedSize: 1024},
			Dependencies: []manifest.Dependency{
				{Name: "olares", Version: ">=0.12.0", Type: "system"},
			},
		},
	}

	if cfg.Metadata.Name != "demo" {
		t.Fatalf("metadata round-trip broken: %+v", cfg.Metadata)
	}
	if cfg.Spec.Resources[0].LimitedCPU != "200m" {
		t.Fatalf("inline ResourceRequirement round-trip broken: %+v", cfg.Spec.Resources[0])
	}
	if cfg.Options.Upload.Dest != "/tmp" {
		t.Fatalf("options.upload round-trip broken: %+v", cfg.Options.Upload)
	}
}

func TestConstants_Mirror(t *testing.T) {
	cases := []struct {
		got, want string
		name      string
	}{
		{olm.APIVersionV1, "v1", "APIVersionV1"},
		{olm.APIVersionV2, "v2", "APIVersionV2"},
		{olm.APIVersionV3, "v3", "APIVersionV3"},
	}
	for _, c := range cases {
		if c.got != c.want {
			t.Errorf("%s = %q, want %q", c.name, c.got, c.want)
		}
	}
}
