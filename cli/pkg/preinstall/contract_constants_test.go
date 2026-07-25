package preinstall

import (
	"bytes"
	"encoding/json"
	"errors"
	"io"
	"os"
	"reflect"
	"testing"
)

type contractConstantsFixture struct {
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

func TestContractConstantsFixtureMatchesGoContract(t *testing.T) {
	data, err := os.ReadFile("testdata/contract-constants.json")
	if err != nil {
		t.Fatal(err)
	}
	fixture, err := decodeContractConstantsFixture(data)
	if err != nil {
		t.Fatal(err)
	}

	if fixture.SupportedSchemaVersion != SupportedSchemaVersion ||
		fixture.OfficialSourceID != OfficialSourceID ||
		fixture.BundleFileName != BundleFileName ||
		fixture.ProfileFileName != ProfileFileName ||
		fixture.Limits.MaxBundleJSONBytes != MaxBundleJSONBytes ||
		fixture.Limits.MaxProfileJSONBytes != MaxProfileJSONBytes ||
		fixture.Limits.MaxBundleApps != MaxBundleApps ||
		fixture.Limits.MaxChartBytes != MaxChartBytes ||
		fixture.Limits.MaxArtifactManifestBytes != MaxArtifactManifestBytes ||
		fixture.Limits.MaxArtifactEntries != MaxArtifactEntries ||
		fixture.Limits.MaxArtifactTotalSize != MaxArtifactTotalSize ||
		fixture.ArtifactKindHFCache != ArtifactKindHFCache {
		t.Fatalf("fixture constants do not match Go contract: %#v", fixture)
	}

	field, ok := reflect.TypeOf(BundleAppV1{}).FieldByName("DefaultEnvs")
	if !ok {
		t.Fatal("BundleAppV1.DefaultEnvs field is missing")
	}
	if got, want := field.Tag.Get("json"), fixture.DefaultEnvsJSONKey+",omitempty"; got != want {
		t.Fatalf("BundleAppV1.DefaultEnvs JSON tag = %q, want %q", got, want)
	}
	if !reflect.DeepEqual(fixture.InstallScopes, []string{string(InstallScopeShared), string(InstallScopePerUser)}) {
		t.Fatalf("installScopes = %q", fixture.InstallScopes)
	}
	field, ok = reflect.TypeOf(BundleAppV1{}).FieldByName("InstallScope")
	if !ok {
		t.Fatal("BundleAppV1.InstallScope field is missing")
	}
	if got := field.Tag.Get("json"); got != "installScope" {
		t.Fatalf("BundleAppV1.InstallScope JSON tag = %q, want %q", got, "installScope")
	}
}

func TestDecodeContractConstantsFixtureStrictly(t *testing.T) {
	tests := map[string][]byte{
		"unknown field":        []byte(`{"unknown":true}`),
		"multiple JSON values": []byte(`{} {}`),
	}
	for name, data := range tests {
		t.Run(name, func(t *testing.T) {
			if _, err := decodeContractConstantsFixture(data); err == nil {
				t.Fatal("decodeContractConstantsFixture() error = nil")
			}
		})
	}
}

func decodeContractConstantsFixture(data []byte) (contractConstantsFixture, error) {
	var fixture contractConstantsFixture
	decoder := json.NewDecoder(bytes.NewReader(data))
	decoder.DisallowUnknownFields()
	if err := decoder.Decode(&fixture); err != nil {
		return fixture, err
	}
	var trailing any
	if err := decoder.Decode(&trailing); err != io.EOF {
		if err == nil {
			return fixture, errors.New("multiple JSON values")
		}
		return fixture, err
	}
	return fixture, nil
}
