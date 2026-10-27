package utils

import (
	"os"
	"path/filepath"
	"reflect"
	"regexp"
	"testing"
)

func TestHasDedicatedGPUMemory(t *testing.T) {
	tests := []struct {
		mode string
		want bool
	}{
		{NvidiaCardType, true},
		{AMDGPUType, true},
		{IntelGPUType, true},
		{GB10ChipType, false},
		{AppleMChipType, false},
		{IntelType, false},
		{AMDType, false},
		{MooreSocType, false},
		{CPUType, false},
		{"", false},
		{"strix-halo", false},
	}

	for _, tt := range tests {
		t.Run(tt.mode, func(t *testing.T) {
			if got := HasDedicatedGPUMemory(tt.mode); got != tt.want {
				t.Fatalf("HasDedicatedGPUMemory(%q) = %v, want %v", tt.mode, got, tt.want)
			}
		})
	}
}

// oacGPUMemoryModesPath locates the manifest validator's own list of modes that
// may declare a GPU-memory quota. It is the schema authority: a manifest is
// rejected for declaring requiredGPUMemory on a mode outside that list, and
// rejected for omitting it on a mode inside it. Our dedicatedGPUMemoryTypes has
// to name the same modes or app-service will size an app against a quantity the
// manifest was never allowed to carry.
const oacGPUMemoryModesPath = "../../../oac/internal/manifest/resources.go"

var (
	gpuMemoryModesBlock = regexp.MustCompile(`(?s)gpuMemoryModes = map\[string\]struct\{\}\{(.*?)\n\}`)
	resourceModeEntry   = regexp.MustCompile(`(ResourceMode\w+):\s*\{\}`)
	resourceModeConst   = regexp.MustCompile(`(ResourceMode\w+)\s*=\s*"([^"]+)"`)
)

// The two sets live in separately versioned modules and oac keeps its copy in
// an internal package, so app-service cannot import it and has to duplicate the
// membership. This test reads oac's source directly to catch the duplicate
// drifting — the failure mode it guards against is silent: a mode added to oac
// alone would have its declared VRAM parsed and then ignored at scheduling
// time, exactly the bug this list was introduced to fix.
//
// It skips rather than fails when oac is not on disk, since app-service also
// builds from a context that contains only its own module.
func TestDedicatedGPUMemoryTypesMatchOAC(t *testing.T) {
	path, err := filepath.Abs(oacGPUMemoryModesPath)
	if err != nil {
		t.Fatalf("resolve oac resources.go path: %v", err)
	}
	source, err := os.ReadFile(path)
	if os.IsNotExist(err) {
		t.Skipf("oac module not present at %s; nothing to compare against", path)
	}
	if err != nil {
		t.Fatalf("read %s: %v", path, err)
	}

	block := gpuMemoryModesBlock.FindSubmatch(source)
	if block == nil {
		t.Fatalf("could not find the gpuMemoryModes map in %s; if it was renamed or restructured, "+
			"update this test and re-check dedicatedGPUMemoryTypes against it", path)
	}

	modeValues := make(map[string]string)
	for _, m := range resourceModeConst.FindAllSubmatch(source, -1) {
		modeValues[string(m[1])] = string(m[2])
	}

	want := make(map[string]struct{})
	entries := resourceModeEntry.FindAllSubmatch(block[1], -1)
	if len(entries) == 0 {
		t.Fatalf("parsed no modes out of the gpuMemoryModes map in %s", path)
	}
	for _, entry := range entries {
		name := string(entry[1])
		value, ok := modeValues[name]
		if !ok {
			t.Fatalf("gpuMemoryModes references %s but no constant defines its value in %s", name, path)
		}
		want[value] = struct{}{}
	}

	if !reflect.DeepEqual(want, dedicatedGPUMemoryTypes) {
		t.Fatalf("dedicatedGPUMemoryTypes = %v, but oac's gpuMemoryModes = %v; "+
			"the two must name the same modes", keysOf(dedicatedGPUMemoryTypes), keysOf(want))
	}
}

func keysOf(set map[string]struct{}) []string {
	out := make([]string, 0, len(set))
	for k := range set {
		out = append(out, k)
	}
	return sortedCopy(out)
}
