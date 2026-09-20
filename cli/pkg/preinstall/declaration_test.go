package preinstall

import (
	"strings"
	"testing"
)

// Everything after the first hyphen is a build of the same release, and all of
// them read one declaration.
func TestTrunkVersionKeepsEveryBuildOfAReleaseTogether(t *testing.T) {
	for version, want := range map[string]string{
		"1.12.7":          "1.12.7",
		"1.12.7-rc.1":     "1.12.7",
		"1.12.7-20260731": "1.12.7",
		" 1.12.7-alpha ":  "1.12.7",
		"":                "",
	} {
		if got := TrunkVersion(version); got != want {
			t.Errorf("TrunkVersion(%q) = %q, want %q", version, got, want)
		}
	}
}

func TestBuildDeclarationMergesDefaultsWithoutMutatingInputs(t *testing.T) {
	bundle := decodeBundle(t, validBundleJSON)
	bundle.Apps[0].AllowedEnvs = []string{"DEFAULT_ONLY", "MODEL_PATH", "WORKER_COUNT"}
	bundle.Apps[0].DefaultEnvs = map[string]string{
		"DEFAULT_ONLY": "enabled",
		"MODEL_PATH":   "/models/default",
	}
	defaultsBefore := map[string]string{
		"DEFAULT_ONLY": "enabled",
		"MODEL_PATH":   "/models/default",
	}
	runtimeEnvs := map[string]string{
		"MODEL_PATH":   "/models/runtime",
		"WORKER_COUNT": "2",
	}
	runtimeBefore := map[string]string{
		"MODEL_PATH":   "/models/runtime",
		"WORKER_COUNT": "2",
	}
	selections := ProfileSelections{Apps: map[string]AppSelection{
		testBundledAppID: {SelectedGPUType: "nvidia", Envs: runtimeEnvs},
	}}

	declaration, err := BuildDeclaration(bundle, selections, nil)
	if err != nil {
		t.Fatalf("BuildDeclaration() error = %v", err)
	}

	app := declarationApp(t, declaration, testBundledAppID)
	if got := app.Envs; got["DEFAULT_ONLY"] != "enabled" || got["MODEL_PATH"] != "/models/runtime" || got["WORKER_COUNT"] != "2" {
		t.Fatalf("declared envs = %#v", got)
	}
	if app.SelectedGPUType != "nvidia" {
		t.Fatalf("declared selectedGpuType = %q", app.SelectedGPUType)
	}
	if !mapsEqual(bundle.Apps[0].DefaultEnvs, defaultsBefore) {
		t.Fatalf("bundle defaults mutated: got %#v want %#v", bundle.Apps[0].DefaultEnvs, defaultsBefore)
	}
	if !mapsEqual(runtimeEnvs, runtimeBefore) {
		t.Fatalf("runtime envs mutated: got %#v want %#v", runtimeEnvs, runtimeBefore)
	}
	app.Envs["MODEL_PATH"] = "changed"
	if bundle.Apps[0].DefaultEnvs["MODEL_PATH"] != "/models/default" || runtimeEnvs["MODEL_PATH"] != "/models/runtime" {
		t.Fatalf("declared envs alias the input maps")
	}
}

// The hardware the installer found only reaches apps that said they can use it.
func TestBuildDeclarationAppliesDetectedGPUOnlyToAllowedApps(t *testing.T) {
	bundle := decodeBundle(t, validBundleJSON)
	bundle.Apps[0].AllowedGPUTypes = []string{"nvidia", "cpu"}
	unsupported := bundle.Apps[0]
	unsupported.AppID = "app-b"
	unsupported.AppName = "app-b"
	unsupported.Chart = "charts/app-b-1.0.0.tgz"
	unsupported.AllowedGPUTypes = []string{"cpu"}
	bundle.Apps = append(bundle.Apps, unsupported)

	declaration, err := BuildDeclaration(bundle, ProfileSelections{
		HardwareProfile: "nvidia",
		DetectedGPUType: "nvidia",
	}, nil)
	if err != nil {
		t.Fatalf("BuildDeclaration() error = %v", err)
	}

	if got := declarationApp(t, declaration, testBundledAppID).SelectedGPUType; got != "nvidia" {
		t.Fatalf("allowed app selectedGpuType = %q", got)
	}
	if got := declarationApp(t, declaration, "app-b").SelectedGPUType; got != "" {
		t.Fatalf("unsupported app selectedGpuType = %q", got)
	}
}

func TestBuildDeclarationExplicitGPUOverridesDetectedGPU(t *testing.T) {
	bundle := decodeBundle(t, validBundleJSON)
	bundle.Apps[0].AllowedGPUTypes = []string{"nvidia", "cpu"}

	declaration, err := BuildDeclaration(bundle, ProfileSelections{
		DetectedGPUType: "nvidia",
		Apps:            map[string]AppSelection{testBundledAppID: {SelectedGPUType: "cpu"}},
	}, nil)
	if err != nil {
		t.Fatalf("BuildDeclaration() error = %v", err)
	}

	if got := declarationApp(t, declaration, testBundledAppID).SelectedGPUType; got != "cpu" {
		t.Fatalf("declared selectedGpuType = %q", got)
	}
}

// A choice the bundle never offered fails the install on the machine that can
// still explain it, rather than reaching Market as a declaration it refuses.
func TestBuildDeclarationRefusesSelectionsTheBundleDoesNotAllow(t *testing.T) {
	for name, test := range map[string]struct {
		selection AppSelection
		wantErr   string
	}{
		"env outside the allowlist": {
			selection: AppSelection{Envs: map[string]string{"OTHER": "x"}},
			wantErr:   "not allowed",
		},
		"gpu outside the allowlist": {
			selection: AppSelection{SelectedGPUType: "amd"},
			wantErr:   "selectedGpuType",
		},
	} {
		t.Run(name, func(t *testing.T) {
			bundle := decodeBundle(t, validBundleJSON)

			_, err := BuildDeclaration(bundle, ProfileSelections{
				Apps: map[string]AppSelection{testBundledAppID: test.selection},
			}, nil)

			if err == nil || !strings.Contains(err.Error(), test.wantErr) {
				t.Fatalf("BuildDeclaration() error = %v, want containing %q", err, test.wantErr)
			}
		})
	}
}

// A sensitive value has no business travelling in a file the whole device can
// read, and the bundle's own allowlist is checked for it too.
func TestBuildDeclarationRefusesSensitiveEnvs(t *testing.T) {
	bundle := decodeBundle(t, validBundleJSON)
	bundle.Apps[0].AllowedEnvs = append(bundle.Apps[0].AllowedEnvs, "API_TOKEN")

	_, err := BuildDeclaration(bundle, ProfileSelections{
		Apps: map[string]AppSelection{testBundledAppID: {Envs: map[string]string{"API_TOKEN": "x"}}},
	}, nil)

	if err == nil || !strings.Contains(err.Error(), "sensitive") {
		t.Fatalf("BuildDeclaration() error = %v, want a sensitive env rejection", err)
	}
}

// An air-gapped install has no catalog to fetch from, so the entry that carries
// its own chart is the one that survives the merge.
func TestBuildDeclarationPrefersABundledAppOverTheCatalogEntryForIt(t *testing.T) {
	bundle := decodeBundle(t, validBundleJSON)
	catalog := []DeclarationAppV2{
		{
			AppID: testBundledAppID, AppName: bundle.Apps[0].AppName,
			InstallScope: InstallScopeShared, InstallOrder: 20,
		},
		{
			AppID: "f3395cd5", AppName: "router",
			InstallScope: InstallScopeShared, InstallOrder: 20,
		},
	}

	declaration, err := BuildDeclaration(bundle, ProfileSelections{}, catalog)
	if err != nil {
		t.Fatalf("BuildDeclaration() error = %v", err)
	}

	if len(declaration.Apps) != 2 {
		t.Fatalf("declared apps = %#v", declaration.Apps)
	}
	if got := declarationApp(t, declaration, testBundledAppID); got.ChartSource != ChartSourceLocal {
		t.Fatalf("bundled app chartSource = %q", got.ChartSource)
	}
	if got := declarationApp(t, declaration, "f3395cd5"); got.ChartSource != ChartSourceCatalog {
		t.Fatalf("catalog app chartSource = %q", got.ChartSource)
	}
}

// The same app arriving under a second id would install twice, so a name
// collision is refused on the same footing as an id collision.
func TestBuildDeclarationTreatsANameCollisionAsTheSameApp(t *testing.T) {
	bundle := decodeBundle(t, validBundleJSON)
	catalog := []DeclarationAppV2{{
		AppID: "other-id", AppName: bundle.Apps[0].AppName,
		InstallScope: InstallScopeShared, InstallOrder: 20,
	}}

	declaration, err := BuildDeclaration(bundle, ProfileSelections{}, catalog)
	if err != nil {
		t.Fatalf("BuildDeclaration() error = %v", err)
	}

	if len(declaration.Apps) != 1 {
		t.Fatalf("declared apps = %#v, want the bundled app alone", declaration.Apps)
	}
}

// The rules below are Market's, mirrored so a declaration that would not import
// never leaves this machine.
func TestValidateDeclarationRefusesWhatMarketRefuses(t *testing.T) {
	valid := func() DeclarationV2 {
		return DeclarationV2{
			SchemaVersion: DeclarationSchemaVersion,
			Apps: []DeclarationAppV2{{
				AppID: "f3395cd5", AppName: "router", InstallScope: InstallScopeShared,
				InstallOrder: 20, ChartSource: ChartSourceCatalog,
			}},
		}
	}
	if err := ValidateDeclaration(&DeclarationV2{
		SchemaVersion: DeclarationSchemaVersion,
	}); err != nil {
		t.Fatalf("ValidateDeclaration() on an empty declaration = %v", err)
	}

	for name, test := range map[string]struct {
		mutate  func(*DeclarationV2)
		wantErr string
	}{
		"no schema": {
			func(d *DeclarationV2) { d.SchemaVersion = "1" }, "schemaVersion",
		},
		"unparsable timestamp": {
			func(d *DeclarationV2) { d.GeneratedAt = "yesterday" }, "RFC3339",
		},
		"no scope": {
			func(d *DeclarationV2) { d.Apps[0].InstallScope = "cluster" }, "installScope",
		},
		"no chart source": {
			func(d *DeclarationV2) { d.Apps[0].ChartSource = "" }, "chartSource",
		},
		"negative order": {
			func(d *DeclarationV2) { d.Apps[0].InstallOrder = -1 }, "installOrder",
		},
		"duplicate id": {
			func(d *DeclarationV2) { d.Apps = append(d.Apps, d.Apps[0]) }, "duplicate appId",
		},
		"catalog entry with a version": {
			func(d *DeclarationV2) { d.Apps[0].Version = "1.0.0" }, "must not carry version",
		},
		"catalog entry with a chart": {
			func(d *DeclarationV2) { d.Apps[0].Chart = "chart/router-1.0.0.tgz" }, "must not carry a chart",
		},
		"catalog entry with envs": {
			func(d *DeclarationV2) { d.Apps[0].Envs = map[string]string{"A": "b"} }, "must not carry envs",
		},
		"local entry without a version": {
			func(d *DeclarationV2) { d.Apps[0].ChartSource = ChartSourceLocal }, "requires version",
		},
	} {
		t.Run(name, func(t *testing.T) {
			declaration := valid()
			test.mutate(&declaration)

			err := ValidateDeclaration(&declaration)

			if err == nil || !strings.Contains(err.Error(), test.wantErr) {
				t.Fatalf("ValidateDeclaration() error = %v, want containing %q", err, test.wantErr)
			}
		})
	}
}

// Two entries whose payloads land on the same path would have one overwrite the
// other, and which one won would depend on the order they were staged in.
func TestValidateDeclarationRefusesOverlappingPayloadPaths(t *testing.T) {
	local := func(id, name, chart string) DeclarationAppV2 {
		return DeclarationAppV2{
			AppID: id, AppName: name, InstallScope: InstallScopeShared,
			ChartSource: ChartSourceLocal, Version: "1.0.0", Chart: chart,
			ChartSHA256: strings.Repeat("a", 64), AppEntry: []byte(`{}`),
		}
	}
	declaration := DeclarationV2{
		SchemaVersion: DeclarationSchemaVersion,
		Apps: []DeclarationAppV2{
			local("app-a", "app-a", "chart/app.tgz"),
			local("app-b", "app-b", "chart/app.tgz/nested.tgz"),
		},
	}

	err := ValidateDeclaration(&declaration)

	if err == nil || !strings.Contains(err.Error(), "overlaps") {
		t.Fatalf("ValidateDeclaration() error = %v, want an overlap rejection", err)
	}
}

// The compiled-in list is the one every device of this version gets, so it has
// to be a list Market would accept.
func TestEmbeddedCatalogAppsAreDeclarable(t *testing.T) {
	declaration := DeclarationV2{
		SchemaVersion: DeclarationSchemaVersion,
		Apps:          catalogDeclarationApps(),
	}

	if err := ValidateDeclaration(&declaration); err != nil {
		t.Fatalf("embedded catalog apps are not declarable: %v", err)
	}
	if len(declaration.Apps) == 0 {
		t.Fatal("no catalog apps are compiled in")
	}
}
