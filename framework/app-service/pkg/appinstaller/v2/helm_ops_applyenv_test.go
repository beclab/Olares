package v2

import (
	"context"
	"errors"
	"io"
	"testing"

	"github.com/beclab/Olares/framework/app-service/pkg/appcfg"
	"helm.sh/helm/v3/pkg/action"
	"helm.sh/helm/v3/pkg/chart"
	"helm.sh/helm/v3/pkg/chartutil"
	kubefake "helm.sh/helm/v3/pkg/kube/fake"
	"helm.sh/helm/v3/pkg/release"
	"helm.sh/helm/v3/pkg/storage"
	"helm.sh/helm/v3/pkg/storage/driver"
)

func memConfig(t *testing.T) *action.Configuration {
	t.Helper()
	return &action.Configuration{
		Releases:     storage.Init(driver.NewMemory()),
		KubeClient:   &kubefake.PrintingKubeClient{Out: io.Discard},
		Capabilities: chartutil.DefaultCapabilities,
		Log:          func(string, ...interface{}) {},
	}
}

func deployedChart(name, version string) *chart.Chart {
	return &chart.Chart{
		Metadata: &chart.Metadata{APIVersion: chart.APIVersionV2, Name: name, Version: version},
		Templates: []*chart.File{{
			Name: "templates/cm.yaml",
			Data: []byte("apiVersion: v1\nkind: ConfigMap\nmetadata:\n  name: " + name + "\ndata:\n  env: {{ .Values.env | quote }}\n  keep: {{ .Values.keep | quote }}\n"),
		}},
		Values: map[string]interface{}{"env": "default", "keep": "default"},
	}
}

func seedRelease(t *testing.T, cfg *action.Configuration, name, ns, version string) {
	t.Helper()
	rel := &release.Release{
		Name:      name,
		Namespace: ns,
		Version:   1,
		Chart:     deployedChart(name, version),
		Config:    map[string]interface{}{"keep": "user-set"},
		Info:      &release.Info{Status: release.StatusDeployed},
	}
	if err := cfg.Releases.Create(rel); err != nil {
		t.Fatal(err)
	}
}

func TestApplyEnvToSubReleasesReusesStoredChart(t *testing.T) {
	appCfg, sharedCfg := memConfig(t), memConfig(t)
	seedRelease(t, appCfg, "client", "demo-alice", "1.0.0")
	seedRelease(t, sharedCfg, "server", "server-shared", "1.0.0")

	charts := []appcfg.Chart{{Name: "client"}, {Name: "server", Shared: true}}
	configFor := func(c appcfg.Chart) (*action.Configuration, error) {
		if c.Shared {
			return sharedCfg, nil
		}
		return appCfg, nil
	}
	noSkip := func(appcfg.Chart) (bool, error) { return false, nil }

	values := map[string]interface{}{"env": "applied"}
	if err := applyEnvToSubReleases(context.Background(), charts, noSkip, configFor, values); err != nil {
		t.Fatalf("applyEnvToSubReleases: %v", err)
	}

	for _, tc := range []struct {
		cfg      *action.Configuration
		name, ns string
	}{{appCfg, "client", "demo-alice"}, {sharedCfg, "server", "server-shared"}} {
		rel, err := tc.cfg.Releases.Last(tc.name)
		if err != nil {
			t.Fatal(err)
		}
		if rel.Version != 2 {
			t.Errorf("%s: revision = %d, want 2", tc.name, rel.Version)
		}
		if rel.Namespace != tc.ns {
			t.Errorf("%s: namespace = %q, want %q (the release's own)", tc.name, rel.Namespace, tc.ns)
		}
		if got := rel.Chart.Metadata.Version; got != "1.0.0" {
			t.Errorf("%s: chart version = %q, want the stored 1.0.0", tc.name, got)
		}
		if rel.Config["env"] != "applied" || rel.Config["keep"] != "user-set" {
			t.Errorf("%s: config = %v, want env applied and prior user values kept", tc.name, rel.Config)
		}
	}
}

func TestApplyEnvToSubReleasesSkipsSharedForNonAdmin(t *testing.T) {
	appCfg, sharedCfg := memConfig(t), memConfig(t)
	seedRelease(t, appCfg, "client", "demo-bob", "1.0.0")
	seedRelease(t, sharedCfg, "server", "server-shared", "1.0.0")

	charts := []appcfg.Chart{{Name: "client"}, {Name: "server", Shared: true}}
	skipShared := func(c appcfg.Chart) (bool, error) { return c.Shared, nil }
	configFor := func(c appcfg.Chart) (*action.Configuration, error) {
		if c.Shared {
			return sharedCfg, nil
		}
		return appCfg, nil
	}

	if err := applyEnvToSubReleases(context.Background(), charts, skipShared, configFor,
		map[string]interface{}{"env": "applied"}); err != nil {
		t.Fatalf("applyEnvToSubReleases: %v", err)
	}
	if rel, _ := sharedCfg.Releases.Last("server"); rel.Version != 1 {
		t.Errorf("shared release revision = %d, want untouched 1", rel.Version)
	}
	if rel, _ := appCfg.Releases.Last("client"); rel.Version != 2 {
		t.Errorf("client release revision = %d, want 2", rel.Version)
	}
}

func TestApplyEnvToSubReleasesDoesNotInstallMissingRelease(t *testing.T) {
	cfg := memConfig(t)
	charts := []appcfg.Chart{{Name: "server", Shared: true}}
	err := applyEnvToSubReleases(context.Background(), charts,
		func(appcfg.Chart) (bool, error) { return false, nil },
		func(appcfg.Chart) (*action.Configuration, error) { return cfg, nil },
		map[string]interface{}{"env": "applied"})
	if !errors.Is(err, driver.ErrReleaseNotFound) {
		t.Fatalf("err = %v, want release not found", err)
	}
	if rels, _ := cfg.Releases.ListReleases(); len(rels) != 0 {
		t.Fatalf("applyEnv installed %d release(s); it must only change existing ones", len(rels))
	}
}
