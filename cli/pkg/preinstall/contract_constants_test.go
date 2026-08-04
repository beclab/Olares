package preinstall

import (
	"os"
	"reflect"
	"testing"
)

func TestContractConstantsMatchCrossRepositoryFixture(t *testing.T) {
	data, err := os.ReadFile("testdata/contract-constants.json")
	if err != nil {
		t.Fatal(err)
	}
	var raw struct {
		SupportedSchemaVersion string `json:"supportedSchemaVersion"`
		OfficialSourceID       string `json:"officialSourceId"`
		BundleFileName         string `json:"bundleFileName"`
		ProfileFileName        string `json:"profileFileName"`
		Limits                 struct {
			MaxBundleJSONBytes       int64 `json:"maxBundleJSONBytes"`
			MaxProfileJSONBytes      int64 `json:"maxProfileJSONBytes"`
			MaxBundleApps            int   `json:"maxBundleApps"`
			MaxChartBytes            int64 `json:"maxChartBytes"`
			MaxArtifactManifestBytes int64 `json:"maxArtifactManifestBytes"`
			MaxArtifactEntries       int   `json:"maxArtifactEntries"`
			MaxArtifactTotalSize     int64 `json:"maxArtifactTotalSize"`
		} `json:"limits"`
		ArtifactKindHFCache string   `json:"artifactKindHFCache"`
		DefaultEnvsJSONKey  string   `json:"defaultEnvsJSONKey"`
		InstallScopes       []string `json:"installScopes"`
	}
	if err := strictDecode(data, &raw); err != nil {
		t.Fatal(err)
	}
	if raw.SupportedSchemaVersion != SupportedSchemaVersion ||
		raw.OfficialSourceID != OfficialSourceID ||
		raw.BundleFileName != BundleFileName ||
		raw.ProfileFileName != ProfileFileName ||
		raw.Limits.MaxBundleJSONBytes != MaxBundleJSONBytes ||
		raw.Limits.MaxProfileJSONBytes != MaxProfileJSONBytes ||
		raw.Limits.MaxBundleApps != MaxBundleApps ||
		raw.Limits.MaxChartBytes != MaxChartBytes ||
		raw.Limits.MaxArtifactManifestBytes != MaxArtifactManifestBytes ||
		raw.Limits.MaxArtifactEntries != MaxArtifactEntries ||
		raw.Limits.MaxArtifactTotalSize != MaxArtifactTotalSize ||
		raw.ArtifactKindHFCache != ArtifactKindHFCache {
		t.Fatalf("fixture constants do not match Go contract: %#v", raw)
	}
	field, ok := reflect.TypeOf(BundleAppV1{}).FieldByName("DefaultEnvs")
	if !ok || field.Tag.Get("json") != raw.DefaultEnvsJSONKey+",omitempty" {
		t.Fatalf("DefaultEnvs contract does not match fixture")
	}
	if !reflect.DeepEqual(raw.InstallScopes, []string{string(InstallScopeShared), string(InstallScopePerUser)}) {
		t.Fatalf("installScopes = %q", raw.InstallScopes)
	}
	field, ok = reflect.TypeOf(BundleAppV1{}).FieldByName("InstallScope")
	if !ok || field.Tag.Get("json") != "installScope" {
		t.Fatalf("InstallScope JSON contract changed")
	}
}
