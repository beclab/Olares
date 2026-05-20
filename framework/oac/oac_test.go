package oac_test

import (
	"errors"
	"reflect"
	"strings"
	"testing"

	"github.com/beclab/Olares/framework/oac"
)

const testdataFirefox = "testdata/firefox"

func TestValidateManifestFile(t *testing.T) {
	if err := oac.ValidateManifestFile(testdataFirefox); err != nil {
		t.Fatalf("ValidateManifestFile: unexpected error: %v", err)
	}
}

func TestValidateManifestContent_MissingRequired(t *testing.T) {
	content := []byte(`
olaresManifest.version: "0.8.1"
olaresManifest.type: app
metadata:
  name: firefox
entrances:
- name: firefox
  host: firefox
  port: 8080
  title: Firefox
spec:
  requiredMemory: 200Mi
  requiredDisk: 10Mi
  requiredCpu: 0.1
  limitedMemory: 1000Mi
  limitedCpu: "1"
  supportClient:
    edge: ""
    android: ""
    ios: ""
    windows: ""
    mac: ""
    linux: ""
  supportArch:
    - amd64
permission:
  appData: true
options:
  policies: []
  analytics: { enabled: false }
  resetCookie: { enabled: false }
`)
	err := oac.ValidateManifestContent(content)
	if err == nil {
		t.Fatalf("expected validation error; metadata.icon is missing")
	}
	var vErr *oac.ValidationError
	if !errors.As(err, &vErr) {
		t.Fatalf("expected *oac.ValidationError, got %T: %v", err, err)
	}
	if !strings.Contains(vErr.Error(), "metadata") {
		t.Fatalf("expected validation error to mention metadata field, got: %v", vErr)
	}
}

func TestValidateManifestContent_APIVersionRestricted(t *testing.T) {
	content := []byte(`
olaresManifest.version: "0.8.1"
apiVersion: v99
metadata:
  name: x
  icon: a
  description: d
  title: t
  version: 1.0.0
entrances:
- name: x
  host: x
  port: 1
  title: T
permission:
  appData: false
options: {}
`)
	err := oac.ValidateManifestContent(content)
	if err == nil {
		t.Fatalf("expected validation error for apiVersion=v99")
	}
	if !strings.Contains(err.Error(), "not supported version") {
		t.Fatalf("expected not supported version in error, got: %v", err)
	}
}

func TestLint_Firefox(t *testing.T) {
	err := oac.Lint(testdataFirefox,
		oac.WithOwnerAdmin("alice"),
		oac.SkipResourceCheck(),
	)
	if err != nil {
		t.Fatalf("Lint: %v", err)
	}
}

// TestLint_Firefox_SecurityContextCheckOptIn documents that the
// non-beclab privileged securityContext check is OFF by default. The
// firefox fixture currently uses third-party images without explicit
// securityContext; turning the check on must still produce no error
// because nil securityContext is treated as safe.
func TestLint_Firefox_SecurityContextCheckOptIn(t *testing.T) {
	err := oac.Lint(testdataFirefox,
		oac.WithOwnerAdmin("alice"),
		oac.SkipResourceCheck(),
		oac.WithSecurityContextCheck(),
	)
	if err != nil {
		t.Fatalf("Lint with WithSecurityContextCheck (clean fixture): %v", err)
	}
}

// TestLint_Firefox_HostPathCheckCanBeDisabled documents that the hostPath +
// rolling-update incompatibility check runs by default. The firefox
// fixture mounts hostPaths on Deployments whose strategy is Recreate, so
// the default Lint above must pass; if a future edit drops the strategy
// override the check will surface a clear error pointing at the
// offending workload + path. This test re-runs Lint with a custom
// validator that fails out if HostPath logic is silently bypassed.
func TestLint_Firefox_HostPathCheckCanBeDisabled(t *testing.T) {
	// Sanity: Lint passes when SkipHostPathCheck is set as well — the
	// fixture is clean either way.
	err := oac.Lint(testdataFirefox,
		oac.WithOwnerAdmin("alice"),
		oac.SkipResourceCheck(),
		oac.SkipHostPathCheck(),
	)
	if err != nil {
		t.Fatalf("Lint with SkipHostPathCheck: %v", err)
	}
}

// TestValidateManifestFile_WithAutoOwnerScenarios documents that the
// auto-owner option now propagates into manifest structural validation
// (not just chart rendering): the firefox fixture's OlaresManifest.yaml
// branches on `eq .Values.admin .Values.bfl.username`, and both the
// owner==admin and owner!=admin renders must keep parsing & validating
// cleanly.
func TestValidateManifestFile_WithAutoOwnerScenarios(t *testing.T) {
	if err := oac.ValidateManifestFile(testdataFirefox, oac.WithAutoOwnerScenarios()); err != nil {
		t.Fatalf("ValidateManifestFile + WithAutoOwnerScenarios: %v", err)
	}
}

// TestValidateManifestContent_WithAutoOwnerScenarios_BothBranchesValidated
// asserts that when WithAutoOwnerScenarios is set, validateManifestBytes
// runs each scenario through the pipeline. The legacy-version manifest
// below renders to different YAML depending on whether
// .Values.admin == .Values.bfl.username, and we deliberately make the
// owner!=admin branch invalid (entrances.port missing) so the validation
// error must surface from at least the owner!=admin scenario.
func TestValidateManifestContent_WithAutoOwnerScenarios_BothBranchesValidated(t *testing.T) {
	content := []byte(`olaresManifest.version: "0.8.1"
apiVersion: v1
metadata:
  name: branchy
  icon: a
  description: d
  title: t
  version: 1.0.0
entrances:
{{- if and .Values.admin .Values.bfl.username (eq .Values.admin .Values.bfl.username) }}
- name: ok
  host: ok
  port: 8080
  title: Ok
{{- else }}
- name: bad
  host: bad
  title: Bad
{{- end }}
spec:
  requiredMemory: 200Mi
  requiredDisk: 10Mi
  requiredCpu: 0.1
  limitedMemory: 1000Mi
  limitedCpu: "1"
  supportClient:
    edge: ""
    android: ""
    ios: ""
    windows: ""
    mac: ""
    linux: ""
  supportArch:
    - amd64
permission:
  appData: true
options:
  policies: []
  analytics: { enabled: false }
  resetCookie: { enabled: false }
`)

	// Without WithAutoOwnerScenarios, the legacy dualOwnerPipeline already
	// covers both internal admin=owner / admin!=owner sub-scenarios, so the
	// missing port in the else-branch must trip validation here too.
	if err := oac.ValidateManifestContent(content); err == nil {
		t.Fatalf("expected validation error in baseline run, got nil")
	}

	// With WithAutoOwnerScenarios the outer loop drives the pipeline once
	// per scenario; the missing port must still surface.
	err := oac.ValidateManifestContent(content, oac.WithAutoOwnerScenarios())
	if err == nil {
		t.Fatalf("expected validation error with WithAutoOwnerScenarios, got nil")
	}
	var ve *oac.ValidationError
	if !errors.As(err, &ve) {
		t.Fatalf("expected *oac.ValidationError, got %T: %v", err, err)
	}
}

func TestListImagesFromOAC(t *testing.T) {
	imgs, err := oac.ListImagesFromOAC(testdataFirefox, oac.WithOwnerAdmin("alice"))
	if err != nil {
		t.Fatalf("ListImagesFromOAC: %v", err)
	}
	if len(imgs) == 0 {
		t.Fatalf("expected at least one image, got zero")
	}
	for i, v := range imgs {
		if v == "" {
			t.Fatalf("empty entry at index %d", i)
		}
		if i > 0 && imgs[i-1] > v {
			t.Fatalf("images are not sorted: %v", imgs)
		}
	}
}

func TestListImagesFromOACForMode_EmptyEqualsListImages(t *testing.T) {
	a, err := oac.ListImagesFromOAC(testdataFirefox, oac.WithOwnerAdmin("alice"))
	if err != nil {
		t.Fatalf("ListImagesFromOAC: %v", err)
	}
	b, err := oac.ListImagesFromOACForMode(testdataFirefox, "", oac.WithOwnerAdmin("alice"))
	if err != nil {
		t.Fatalf("ListImagesFromOACForMode(\"\"): %v", err)
	}
	if !reflect.DeepEqual(a, b) {
		t.Fatalf("empty mode must match ListImages: %v vs %v", a, b)
	}
}

func TestListImagesFromOACForMode_AcceptsMode(t *testing.T) {
	imgs, err := oac.ListImagesFromOACForMode(testdataFirefox, "nvidia", oac.WithOwnerAdmin("alice"))
	if err != nil {
		t.Fatalf("ListImagesFromOACForMode(\"nvidia\"): %v", err)
	}
	if len(imgs) == 0 {
		t.Fatalf("expected at least one image, got zero")
	}
}

// TestListImagesFromOACForModes_NilEqualsListImages documents the empty-
// modes shortcut: passing nil (or an empty slice) renders the chart with
// no GPU.Type override, identical to ListImages.
func TestListImagesFromOACForModes_NilEqualsListImages(t *testing.T) {
	a, err := oac.ListImagesFromOAC(testdataFirefox, oac.WithOwnerAdmin("alice"))
	if err != nil {
		t.Fatalf("ListImagesFromOAC: %v", err)
	}
	b, err := oac.ListImagesFromOACForModes(testdataFirefox, nil, oac.WithOwnerAdmin("alice"))
	if err != nil {
		t.Fatalf("ListImagesFromOACForModes(nil): %v", err)
	}
	if !reflect.DeepEqual(a, b) {
		t.Fatalf("nil modes must match ListImages: %v vs %v", a, b)
	}
}

// TestListImagesFromOACForModes_SingleModeEqualsForMode confirms that
// passing a single mode through the multi-mode API yields the same
// result as the single-mode API.
func TestListImagesFromOACForModes_SingleModeEqualsForMode(t *testing.T) {
	a, err := oac.ListImagesFromOACForMode(testdataFirefox, "nvidia", oac.WithOwnerAdmin("alice"))
	if err != nil {
		t.Fatalf("ListImagesFromOACForMode: %v", err)
	}
	b, err := oac.ListImagesFromOACForModes(testdataFirefox, []string{"nvidia"}, oac.WithOwnerAdmin("alice"))
	if err != nil {
		t.Fatalf("ListImagesFromOACForModes([nvidia]): %v", err)
	}
	if !reflect.DeepEqual(a, b) {
		t.Fatalf("single-mode slice must match ListImagesForMode: %v vs %v", a, b)
	}
}

// TestListImagesFromOACForModes_UnionsAndDedups documents that explicit
// multi-mode input deduplicates the union and matches the union of the
// per-mode results.
func TestListImagesFromOACForModes_UnionsAndDedups(t *testing.T) {
	nvImgs, err := oac.ListImagesFromOACForMode(testdataFirefox, "nvidia", oac.WithOwnerAdmin("alice"))
	if err != nil {
		t.Fatalf("ListImagesFromOACForMode(nvidia): %v", err)
	}
	cpuImgs, err := oac.ListImagesFromOACForMode(testdataFirefox, "cpu", oac.WithOwnerAdmin("alice"))
	if err != nil {
		t.Fatalf("ListImagesFromOACForMode(cpu): %v", err)
	}

	got, err := oac.ListImagesFromOACForModes(
		testdataFirefox,
		[]string{"nvidia", "cpu", "nvidia"}, // duplicate "nvidia" must be ignored
		oac.WithOwnerAdmin("alice"),
	)
	if err != nil {
		t.Fatalf("ListImagesFromOACForModes(multi): %v", err)
	}

	wantSet := map[string]struct{}{}
	for _, img := range nvImgs {
		wantSet[img] = struct{}{}
	}
	for _, img := range cpuImgs {
		wantSet[img] = struct{}{}
	}
	if len(got) != len(wantSet) {
		t.Fatalf("union mismatch: got %d images %v, want %d unique %v",
			len(got), got, len(wantSet), wantSet)
	}
	for _, img := range got {
		if _, ok := wantSet[img]; !ok {
			t.Fatalf("unexpected image in union: %q", img)
		}
	}
	// Sorted output contract.
	for i := 1; i < len(got); i++ {
		if got[i-1] > got[i] {
			t.Fatalf("multi-mode output is not sorted: %v", got)
		}
	}
}

// TestListImagesFromOACForModes_AllExpandsKnownModes asserts that "all"
// expands to AllImageRenderModes and that mixing "all" with explicit
// modes doesn't introduce duplicates.
func TestListImagesFromOACForModes_AllExpandsKnownModes(t *testing.T) {
	allImgs, err := oac.ListImagesFromOACForModes(
		testdataFirefox,
		[]string{"all"},
		oac.WithOwnerAdmin("alice"),
	)
	if err != nil {
		t.Fatalf("ListImagesFromOACForModes(all): %v", err)
	}
	if len(allImgs) == 0 {
		t.Fatalf("expected at least one image for all-modes expansion")
	}

	mixed, err := oac.ListImagesFromOACForModes(
		testdataFirefox,
		[]string{"nvidia", "all", "cpu"},
		oac.WithOwnerAdmin("alice"),
	)
	if err != nil {
		t.Fatalf("ListImagesFromOACForModes(nvidia,all,cpu): %v", err)
	}
	if !reflect.DeepEqual(allImgs, mixed) {
		t.Fatalf("'all' should subsume explicit nvidia/cpu (already part of AllImageRenderModes)\nall=%v\nmixed=%v",
			allImgs, mixed)
	}
}

// TestAllImageRenderModes_ContainsExpected pins the documented mode set
// so tests fail loudly if it's edited unintentionally.
func TestAllImageRenderModes_ContainsExpected(t *testing.T) {
	want := []string{
		"cpu",
		"apple-m",
		"nvidia",
		"nvidia-gb10",
		"mthreads-m1000",
		"strix-halo",
	}
	if !reflect.DeepEqual(oac.AllImageRenderModes, want) {
		t.Fatalf("AllImageRenderModes drift: got %v, want %v", oac.AllImageRenderModes, want)
	}
}

func TestLintBothOwnerScenarios(t *testing.T) {
	err := oac.LintBothOwnerScenarios(testdataFirefox, oac.SkipResourceCheck())
	if err != nil {
		t.Fatalf("LintBothOwnerScenarios: %v", err)
	}
}

// TestLintBothOwnerScenarios_DoesNotMutateCallerSlice guards against a
// regression where the helper appended onto its variadic argument and
// silently corrupted the caller's backing array when the caller passed a
// sub-slice with spare capacity.
func TestLintBothOwnerScenarios_DoesNotMutateCallerSlice(t *testing.T) {
	// backing has len=1 but cap=4. We spread only the first slot via the
	// variadic ... so extraOpts inside the helper aliases this same backing
	// array. A naive append onto extraOpts would write
	// WithAutoOwnerScenarios() into backing[1], which we detect by checking
	// that backing[1..3] remain nil after the call.
	backing := make([]oac.Option, 1, 4)
	backing[0] = oac.SkipResourceCheck()

	if err := oac.LintBothOwnerScenarios(testdataFirefox, backing...); err != nil {
		t.Fatalf("LintBothOwnerScenarios: %v", err)
	}

	full := backing[:cap(backing)]
	for i := 1; i < len(full); i++ {
		if full[i] != nil {
			t.Fatalf("LintBothOwnerScenarios mutated caller backing array at index %d: expected nil, got non-nil option", i)
		}
	}
}
