// Package compute hosts `olares-cli settings compute` — the new-version
// "Accelerator" surface (Olares PR #3118, TermiPass compute-resources
// page). It mirrors the SPA's Settings -> Accelerator pages
// (AcceleratorPage.vue / ManageNodePage.vue) backed by user-service's
// compute.controller.ts, which relays app-service /compute-resources.
//
// This is the 1.12.6+ replacement for the legacy `settings gpu list`
// (HAMI /api/gpu/list, 1.12.5). The two coexist: `gpu` stays for old
// backends, `compute` is version-gated to >= 1.12.6.
//
// Display + behavior rules are deliberately kept identical to the SPA. The
// pure helpers below are Go re-implementations of the TypeScript helpers in
// TermiPass packages/app/src/constant/compute.ts so the CLI prints and
// behaves the same way the web UI does.
package compute

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"math"
	"os"
	"strings"

	"github.com/spf13/cobra"

	"github.com/beclab/Olares/cli/cmd/ctl/settings/internal/preflight"
	"github.com/beclab/Olares/cli/pkg/cmdutil"
	"github.com/beclab/Olares/cli/pkg/credential"
	"github.com/beclab/Olares/cli/pkg/whoami"
)

// computeMinOlaresVersion is the first Olares line that exposes the
// compute-resources wire surface (/api/compute-resources and friends).
// On 1.12.5 those routes do not exist, so we fail fast with an actionable
// upgrade message and point users at the legacy `settings gpu list`.
const computeMinOlaresVersion = "1.12.6"

type Format string

const (
	FormatTable Format = "table"
	FormatJSON  Format = "json"
)

func parseFormat(s string) (Format, error) {
	v := strings.ToLower(strings.TrimSpace(s))
	switch v {
	case "", string(FormatTable):
		return FormatTable, nil
	case string(FormatJSON):
		return FormatJSON, nil
	default:
		return "", fmt.Errorf("unsupported --output %q (allowed: table, json)", s)
	}
}

func addOutputFlag(cmd *cobra.Command, target *string) {
	cmd.Flags().StringVarP(target, "output", "o", "table", "output format: table, json")
}

type Doer interface {
	DoJSON(ctx context.Context, method, path string, body, out interface{}) error
}

type preparedClient struct {
	profile *credential.ResolvedProfile
	doer    Doer
}

func prepare(ctx context.Context, f *cmdutil.Factory) (*preparedClient, error) {
	if f == nil {
		return nil, fmt.Errorf("internal error: settings compute not wired with cmdutil.Factory")
	}
	rp, err := f.ResolveProfile(ctx)
	if err != nil {
		return nil, err
	}
	hc, err := f.HTTPClient(ctx)
	if err != nil {
		return nil, err
	}
	return &preparedClient{
		profile: rp,
		doer:    whoami.NewHTTPClient(hc, rp.DesktopURL, rp.OlaresID),
	}, nil
}

// requireComputeBackendVersion is the client-side version preflight for the
// compute-resources surface. It mirrors TermiPass's `isLargeVersion12_6`
// gate: compute-resources only exists on Olares >= 1.12.6, so on an older
// (or undetectable) backend we reject up front with an actionable error
// instead of letting the user hit an opaque 404 from /api/compute-resources.
//
// It is a thin wrapper that supplies this area's gate copy to the shared
// preflight.RequireMinVersion helper.
func requireComputeBackendVersion(ctx context.Context, f *cmdutil.Factory) error {
	return preflight.RequireMinVersion(ctx, f, preflight.MinVersionGate{
		Verb:       "settings compute",
		MinVersion: computeMinOlaresVersion,
		Reason:     "compute-resources APIs",
		Fallback:   "upgrade the Olares system, or use the legacy `olares-cli settings gpu list` on 1.12.5",
	})
}

type bflEnvelope struct {
	Code    int             `json:"code"`
	Message string          `json:"message"`
	Data    json.RawMessage `json:"data"`
}

func doGetEnvelope(ctx context.Context, d Doer, path string, out interface{}) error {
	var env bflEnvelope
	if err := d.DoJSON(ctx, "GET", path, nil, &env); err != nil {
		return err
	}
	switch env.Code {
	case 0, 200:
	default:
		msg := strings.TrimSpace(env.Message)
		if msg == "" {
			return fmt.Errorf("GET %s: upstream returned code %d", path, env.Code)
		}
		return fmt.Errorf("GET %s: upstream returned code %d: %s", path, env.Code, msg)
	}
	if out == nil || len(env.Data) == 0 {
		return nil
	}
	if err := json.Unmarshal(env.Data, out); err != nil {
		return fmt.Errorf("GET %s: decode data: %w", path, err)
	}
	return nil
}

func doMutateEnvelope(ctx context.Context, d Doer, method, path string, body, out interface{}) error {
	var env bflEnvelope
	if err := d.DoJSON(ctx, method, path, body, &env); err != nil {
		return err
	}
	switch env.Code {
	case 0, 200:
	default:
		msg := strings.TrimSpace(env.Message)
		if msg == "" {
			return fmt.Errorf("%s %s: upstream returned code %d", method, path, env.Code)
		}
		return fmt.Errorf("%s %s: upstream returned code %d: %s", method, path, env.Code, msg)
	}
	if out == nil || len(env.Data) == 0 {
		return nil
	}
	if err := json.Unmarshal(env.Data, out); err != nil {
		return fmt.Errorf("%s %s: decode data: %w", method, path, err)
	}
	return nil
}

func printJSON(w io.Writer, v interface{}) error {
	if w == nil {
		w = os.Stdout
	}
	enc := json.NewEncoder(w)
	enc.SetIndent("", "  ")
	return enc.Encode(v)
}

func nonEmpty(s string) string {
	if s == "" {
		return "-"
	}
	return s
}

// ---------------------------------------------------------------------------
// Pure helpers re-implemented from TermiPass src/constant/compute.ts so the
// CLI's display + behavior match the SPA exactly.
// ---------------------------------------------------------------------------

const giRoundEpsilon = 1e-6

// computeMemoryValue mirrors formatComputeMemoryValue: bytes -> Gi, floored to
// 2 decimals so it never overstates capacity. Returns the bare number string.
func computeMemoryValue(bytes int64) string {
	if bytes <= 0 {
		return "0"
	}
	g := float64(bytes) / math.Pow(1024, 3)
	floored := math.Floor(g*100+giRoundEpsilon) / 100
	if floored == math.Trunc(floored) {
		return fmt.Sprintf("%d", int64(floored))
	}
	return fmt.Sprintf("%.2f", floored)
}

// formatComputeMemory mirrors the TS formatComputeMemory: "<value> Gi".
func formatComputeMemory(bytes int64) string {
	return computeMemoryValue(bytes) + " Gi"
}

// formatComputeCpuCores mirrors the TS helper: millicores -> "<cores> Core".
func formatComputeCpuCores(milli int64) string {
	if milli <= 0 {
		return "0 Core"
	}
	cores := float64(milli) / 1000
	rounded := math.Round(cores*100) / 100
	if rounded == math.Trunc(rounded) {
		return fmt.Sprintf("%d Core", int64(rounded))
	}
	return fmt.Sprintf("%.2f Core", rounded)
}

// normalizeComputeMode mirrors the TS helper: lowercase, trim, `_` -> `-`.
func normalizeComputeMode(mode string) string {
	return strings.ReplaceAll(strings.ToLower(strings.TrimSpace(mode)), "_", "-")
}

// computeModeTitle mirrors COMPUTE_MODE_TITLE — brand display names, shown
// identically in every locale. Unknown values fall back to the raw string.
var computeModeTitleMap = map[string]string{
	"cpu":         "CPU",
	"nvidia":      "NVIDIA GPU",
	"nvidia-gb10": "NVIDIA GB10",
	"apple-m":     "Apple Silicon-M",
	"intel":       "Intel",
	"amd":         "AMD",
	"intel-gpu":   "Intel GPU",
	"amd-gpu":     "AMD GPU",
	"moore-soc":   "Moore Threads",
}

func computeModeTitle(mode string) string {
	if t, ok := computeModeTitleMap[normalizeComputeMode(mode)]; ok {
		return t
	}
	return strings.TrimSpace(mode)
}

// vramComputeModes mirrors VRAM_COMPUTE_MODES: modes whose device memory is
// dedicated VRAM (discrete GPUs). Integrated accelerators share system RAM.
var vramComputeModes = map[string]struct{}{
	"nvidia":    {},
	"intel-gpu": {},
	"amd-gpu":   {},
}

func isVramComputeMode(mode string) bool {
	_, ok := vramComputeModes[normalizeComputeMode(mode)]
	return ok
}

// acceleratorSupportTypeLabel mirrors ACCELERATOR_SUPPORT_TYPE_LABEL_KEY —
// the wording shown on the Accelerator page.
var acceleratorSupportTypeLabelMap = map[string]string{
	"Exclusive":    "Exclusive",
	"MemorySlice":  "Memory Slicing",
	"TimeSlice":    "Time Slicing",
	"MemoryShared": "Memory Shared",
}

func acceleratorSupportTypeLabel(t string) string {
	if l, ok := acceleratorSupportTypeLabelMap[t]; ok {
		return l
	}
	return t
}

// normalizeSupportTypeKey folds a support-type string for matching:
// trimmed, lowercased, spaces removed. So "Memory Slicing", "MemorySlice"
// and "  memoryslice " all compare equal-ish to their canonical key.
func normalizeSupportTypeKey(s string) string {
	return strings.ToLower(strings.ReplaceAll(strings.TrimSpace(s), " ", ""))
}

// canonicalSupportType resolves a user-supplied --type to its canonical
// DeviceSupportType enum. It accepts BOTH the enum value (e.g. "MemorySlice")
// and the human label shown by `list` (e.g. "Memory Slicing"), case- and
// space-insensitively, so users can paste whichever form they see.
func canonicalSupportType(input string) (string, bool) {
	key := normalizeSupportTypeKey(input)
	if key == "" {
		return "", false
	}
	for _, enum := range validSupportTypes {
		if normalizeSupportTypeKey(enum) == key {
			return enum, true
		}
		if normalizeSupportTypeKey(acceleratorSupportTypeLabelMap[enum]) == key {
			return enum, true
		}
	}
	return "", false
}

func isExclusiveSupportType(t string) bool {
	return t == "Exclusive"
}
