package preinstall

import (
	"os"
	"reflect"
	"testing"
)

// contractConstantsFixture is the half of the preinstall contract that is a
// literal rather than a rule. This program writes the declaration and Market
// reads it, and the two ship separately, so every value below is duplicated in
// the other repository's copy of this file. Copying it into a fixture is what
// turns a silent disagreement into a failing test on both sides.
type contractConstantsFixture struct {
	SupportedSchemaVersion string `json:"supportedSchemaVersion"`
	OfficialSourceID       string `json:"officialSourceId"`
	DeclarationFilePrefix  string `json:"declarationFilePrefix"`
	DeclarationFileSuffix  string `json:"declarationFileSuffix"`
	Limits                 struct {
		MaxDeclarationBytes      int64 `json:"maxDeclarationBytes"`
		MaxDeclarationApps       int   `json:"maxDeclarationApps"`
		MaxChartBytes            int64 `json:"maxChartBytes"`
		MaxArtifactManifestBytes int64 `json:"maxArtifactManifestBytes"`
		MaxArtifactEntries       int   `json:"maxArtifactEntries"`
		MaxArtifactTotalSize     int64 `json:"maxArtifactTotalSize"`
	} `json:"limits"`
	ArtifactKindHFCache string   `json:"artifactKindHFCache"`
	EnvsJSONKey         string   `json:"envsJSONKey"`
	ChartSources        []string `json:"chartSources"`
	InstallScopes       []string `json:"installScopes"`
}

func readContractConstants(t *testing.T) contractConstantsFixture {
	t.Helper()
	data, err := os.ReadFile("testdata/contract-constants.json")
	if err != nil {
		t.Fatal(err)
	}
	var fixture contractConstantsFixture
	if err := strictDecode(data, &fixture); err != nil {
		t.Fatal(err)
	}
	return fixture
}

func TestContractConstantsMatchCrossRepositoryFixture(t *testing.T) {
	fixture := readContractConstants(t)

	for _, test := range []struct{ name, got, want string }{
		{"supportedSchemaVersion", DeclarationSchemaVersion, fixture.SupportedSchemaVersion},
		{"officialSourceId", OfficialSourceID, fixture.OfficialSourceID},
		{"declarationFilePrefix", declarationFilePrefix, fixture.DeclarationFilePrefix},
		{"declarationFileSuffix", declarationFileSuffix, fixture.DeclarationFileSuffix},
		{"artifactKindHFCache", ArtifactKindHFCache, fixture.ArtifactKindHFCache},
	} {
		if test.got != test.want {
			t.Errorf("%s = %q, fixture has %q", test.name, test.got, test.want)
		}
	}
	for _, test := range []struct {
		name      string
		got, want int64
	}{
		{"maxDeclarationBytes", MaxDeclarationBytes, fixture.Limits.MaxDeclarationBytes},
		{"maxDeclarationApps", MaxDeclarationApps, int64(fixture.Limits.MaxDeclarationApps)},
		{"maxChartBytes", MaxChartBytes, fixture.Limits.MaxChartBytes},
		{"maxArtifactManifestBytes", MaxArtifactManifestBytes, fixture.Limits.MaxArtifactManifestBytes},
		{"maxArtifactEntries", MaxArtifactEntries, int64(fixture.Limits.MaxArtifactEntries)},
		{"maxArtifactTotalSize", MaxArtifactTotalSize, fixture.Limits.MaxArtifactTotalSize},
	} {
		if test.got != test.want {
			t.Errorf("%s = %d, fixture has %d", test.name, test.got, test.want)
		}
	}
	if want := []string{string(ChartSourceCatalog), string(ChartSourceLocal)}; !reflect.DeepEqual(fixture.ChartSources, want) {
		t.Errorf("chartSources = %q, want %q", fixture.ChartSources, want)
	}
	if want := []string{string(InstallScopeShared), string(InstallScopePerUser)}; !reflect.DeepEqual(fixture.InstallScopes, want) {
		t.Errorf("installScopes = %q, want %q", fixture.InstallScopes, want)
	}
}

// The file name is a contract in itself: this program names the file and Market
// looks for that name and no other, so both halves are asserted against the same
// pair of literals rather than against each other.
func TestDeclarationFileNameMatchesTheFixture(t *testing.T) {
	fixture := readContractConstants(t)

	if got, want := DeclarationFileName("1.12.7-rc.1"),
		fixture.DeclarationFilePrefix+"1.12.7"+fixture.DeclarationFileSuffix; got != want {
		t.Fatalf("DeclarationFileName() = %q, want %q", got, want)
	}
}

func TestDeclarationJSONKeysMatchTheFixture(t *testing.T) {
	fixture := readContractConstants(t)

	appType := reflect.TypeOf(DeclarationAppV2{})
	for _, test := range []struct{ field, tag string }{
		{"Envs", fixture.EnvsJSONKey + ",omitempty"},
		{"InstallScope", "installScope"},
		{"ChartSource", "chartSource"},
	} {
		field, ok := appType.FieldByName(test.field)
		if !ok {
			t.Errorf("DeclarationAppV2.%s is missing", test.field)
			continue
		}
		if got := field.Tag.Get("json"); got != test.tag {
			t.Errorf("DeclarationAppV2.%s json tag = %q, want %q", test.field, got, test.tag)
		}
	}
}
