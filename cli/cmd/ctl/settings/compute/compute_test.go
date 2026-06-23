package compute

import (
	"encoding/json"
	"testing"
)

func TestComputeMemoryValue(t *testing.T) {
	const gib = int64(1) << 30
	cases := []struct {
		name  string
		bytes int64
		want  string
	}{
		{"zero", 0, "0"},
		{"negative", -5, "0"},
		{"exact 12Gi", 12 * gib, "12"},
		// Non-integer values keep 2 decimals (mirrors JS toFixed(2)).
		{"half", 12*gib + gib/2, "12.50"},
		// Floor (never overstate): 12.999 Gi -> 12.99, not 13.
		{"floor never overstates", 12*gib + gib*999/1000, "12.99"},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			if got := computeMemoryValue(tc.bytes); got != tc.want {
				t.Fatalf("computeMemoryValue(%d) = %q, want %q", tc.bytes, got, tc.want)
			}
		})
	}
	if got := formatComputeMemory(12 * gib); got != "12 Gi" {
		t.Fatalf("formatComputeMemory = %q, want %q", got, "12 Gi")
	}
}

func TestFormatComputeCpuCores(t *testing.T) {
	cases := []struct {
		milli int64
		want  string
	}{
		{0, "0 Core"},
		{-1, "0 Core"},
		{1000, "1 Core"},
		{150, "0.15 Core"},
		{2500, "2.50 Core"},
	}
	for _, tc := range cases {
		if got := formatComputeCpuCores(tc.milli); got != tc.want {
			t.Fatalf("formatComputeCpuCores(%d) = %q, want %q", tc.milli, got, tc.want)
		}
	}
}

func TestComputeModeTitleAndVram(t *testing.T) {
	titleCases := map[string]string{
		"nvidia":      "NVIDIA GPU",
		"NVIDIA":      "NVIDIA GPU",  // normalized (case)
		"nvidia_gb10": "NVIDIA GB10", // normalized (_ -> -)
		"apple-m":     "Apple Silicon-M",
		"weird-thing": "weird-thing", // unknown falls back to raw
	}
	for in, want := range titleCases {
		if got := computeModeTitle(in); got != want {
			t.Fatalf("computeModeTitle(%q) = %q, want %q", in, got, want)
		}
	}

	vram := []string{"nvidia", "intel-gpu", "amd-gpu", "INTEL_GPU"}
	for _, m := range vram {
		if !isVramComputeMode(m) {
			t.Fatalf("isVramComputeMode(%q) = false, want true", m)
		}
	}
	notVram := []string{"cpu", "intel", "apple-m", "nvidia-gb10", "moore-soc"}
	for _, m := range notVram {
		if isVramComputeMode(m) {
			t.Fatalf("isVramComputeMode(%q) = true, want false", m)
		}
	}
}

func TestAcceleratorSupportTypeLabel(t *testing.T) {
	want := map[string]string{
		"Exclusive":    "Exclusive",
		"MemorySlice":  "Memory Slicing",
		"TimeSlice":    "Time Slicing",
		"MemoryShared": "Memory Shared",
		"Mystery":      "Mystery", // unknown falls back to raw
	}
	for in, exp := range want {
		if got := acceleratorSupportTypeLabel(in); got != exp {
			t.Fatalf("acceleratorSupportTypeLabel(%q) = %q, want %q", in, got, exp)
		}
	}
}

func TestCanonicalSupportType(t *testing.T) {
	ok := map[string]string{
		// enum forms (case / space insensitive)
		"Exclusive":      "Exclusive",
		"exclusive":      "Exclusive",
		"MemorySlice":    "MemorySlice",
		"  memoryslice ": "MemorySlice",
		"TIMESLICE":      "TimeSlice",
		"MemoryShared":   "MemoryShared",
		// human labels shown by `list`
		"Memory Slicing": "MemorySlice",
		"memory slicing": "MemorySlice",
		"Time Slicing":   "TimeSlice",
		"Memory Shared":  "MemoryShared",
	}
	for in, want := range ok {
		got, valid := canonicalSupportType(in)
		if !valid || got != want {
			t.Fatalf("canonicalSupportType(%q) = (%q,%v), want (%q,true)", in, got, valid, want)
		}
	}
	for _, in := range []string{"", "   ", "bogus", "memory-slice", "exclusiveish"} {
		if got, valid := canonicalSupportType(in); valid {
			t.Fatalf("canonicalSupportType(%q) = (%q,true), want invalid", in, got)
		}
	}
}

func TestEffectiveUsedMemory(t *testing.T) {
	const gib = int64(1) << 30

	// Exclusive + bound app => full card capacity (mirrors the SPA).
	exclusive := computeDevice{
		Memory:      12 * gib,
		SupportType: "Exclusive",
		Bindings:    []computeBinding{{AppName: "comfyui", Memory: 3 * gib}},
	}
	if got := exclusive.effectiveUsedMemory(); got != 12*gib {
		t.Fatalf("exclusive effectiveUsedMemory = %d, want %d", got, 12*gib)
	}

	// Exclusive without binding => falls back to summed usage (0).
	exclusiveIdle := computeDevice{Memory: 12 * gib, SupportType: "Exclusive"}
	if got := exclusiveIdle.effectiveUsedMemory(); got != 0 {
		t.Fatalf("exclusive idle effectiveUsedMemory = %d, want 0", got)
	}

	// Non-exclusive => summed binding usage, capped at capacity.
	shared := computeDevice{
		Memory:      12 * gib,
		SupportType: "MemorySlice",
		Bindings: []computeBinding{
			{AppName: "a", Memory: 8 * gib},
			{AppName: "b", Memory: 8 * gib}, // sum 16Gi > 12Gi capacity => capped
		},
	}
	if got := shared.effectiveUsedMemory(); got != 12*gib {
		t.Fatalf("shared effectiveUsedMemory = %d, want %d (capped)", got, 12*gib)
	}
}

func TestDeviceDisplayName(t *testing.T) {
	if got := (computeDevice{Name: "GPU0", CardModel: "RTX", Mode: "nvidia", ID: "x"}).deviceDisplayName(); got != "GPU0" {
		t.Fatalf("name priority = %q, want GPU0", got)
	}
	if got := (computeDevice{CardModel: "RTX 4090", Mode: "nvidia", ID: "x"}).deviceDisplayName(); got != "RTX 4090" {
		t.Fatalf("cardModel priority = %q, want RTX 4090", got)
	}
	if got := (computeDevice{Mode: "nvidia", ID: "x"}).deviceDisplayName(); got != "NVIDIA GPU" {
		t.Fatalf("mode-title priority = %q, want NVIDIA GPU", got)
	}
	if got := (computeDevice{ID: "dev-123"}).deviceDisplayName(); got != "dev-123" {
		t.Fatalf("id fallback = %q, want dev-123", got)
	}
}

func TestParseSupportTypeResult(t *testing.T) {
	// Direct shape: {status, stoppedApps}.
	direct := json.RawMessage(`{"status":"switched","stoppedApps":[{"appName":"a"}]}`)
	if got := parseSupportTypeResult(direct); got.Status != "switched" || len(got.StoppedApps) != 1 {
		t.Fatalf("direct parse = %+v", got)
	}

	// Discriminated blocked shape: {type:'computeDeviceSwitchBlocked', Data:{...}}.
	wrapped := json.RawMessage(`{"type":"computeDeviceSwitchBlocked","Data":{"status":"bound-apps-stop-blocked","blockedApps":[{"appName":"b","reason":"running job"}]}}`)
	got := parseSupportTypeResult(wrapped)
	if got.Status != "bound-apps-stop-blocked" || len(got.BlockedApps) != 1 || got.BlockedApps[0].Reason != "running job" {
		t.Fatalf("wrapped parse = %+v", got)
	}

	// Empty data => default to switched (success with no payload).
	if got := parseSupportTypeResult(nil); got.Status != "switched" {
		t.Fatalf("empty parse = %+v, want switched", got)
	}
}

func TestAppBindings(t *testing.T) {
	nodes := []computeNode{
		{
			NodeName: "node-a",
			Devices: []computeDevice{
				{ID: "d1", Bindings: []computeBinding{{AppName: "multi", Spec: &computeSpec{SupportMultiCards: true}}}},
				{ID: "d2", Bindings: []computeBinding{{AppName: "multi", Spec: &computeSpec{SupportMultiCards: true}}}},
				{ID: "d3", Bindings: []computeBinding{{AppName: "single"}}},
			},
		},
	}

	cards, multi := appBindings(nodes, "multi")
	if len(cards) != 2 || !multi {
		t.Fatalf("multi app bindings = %v multi=%v, want 2 cards multi=true", cards, multi)
	}
	cards, multi = appBindings(nodes, "single")
	if len(cards) != 1 || multi {
		t.Fatalf("single app bindings = %v multi=%v, want 1 card multi=false", cards, multi)
	}
	cards, _ = appBindings(nodes, "absent")
	if len(cards) != 0 {
		t.Fatalf("absent app bindings = %v, want empty", cards)
	}
}

func TestFindDevice(t *testing.T) {
	nodes := []computeNode{
		{NodeName: "n1", Devices: []computeDevice{{ID: "d1"}, {ID: "d2"}}},
		{NodeName: "n2", Devices: []computeDevice{{ID: "d3"}}},
	}
	if d := findDevice(nodes, "n1", "d2"); d == nil || d.ID != "d2" {
		t.Fatalf("findDevice n1/d2 = %+v", d)
	}
	if d := findDevice(nodes, "n2", "d1"); d != nil {
		t.Fatalf("findDevice n2/d1 = %+v, want nil (d1 is on n1)", d)
	}
	if d := findDevice(nodes, "nope", "d1"); d != nil {
		t.Fatalf("findDevice nope/d1 = %+v, want nil", d)
	}
}

func TestIsValidSupportType(t *testing.T) {
	for _, v := range []string{"Exclusive", "MemorySlice", "TimeSlice", "MemoryShared"} {
		if !isValidSupportType(v) {
			t.Fatalf("isValidSupportType(%q) = false, want true", v)
		}
	}
	for _, v := range []string{"", "exclusive", "memory-slice", "Bogus"} {
		if isValidSupportType(v) {
			t.Fatalf("isValidSupportType(%q) = true, want false", v)
		}
	}
}
