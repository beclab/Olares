// Package gpu hosts `olares-cli settings gpu` — the profile-based Settings ->
// GPU / "Compute Acceleration" surface (NOT the kubeconfig-based top-level
// `olares-cli gpu` tree, which is a separate lower-level command for cluster
// admins).
//
// This area is the canonical "complete replacement" example for the
// version-compatibility framework (pkg/olaresclient): Olares 1.12.5 exposed a
// HAMI-backed GPU device/mode model under /api/gpu/*, while 1.12.6 replaced it
// wholesale with a node → device → supportType → binding model under
// /api/compute-resources* (the old GPUController was deleted upstream). The two
// share no wire format, so every verb dispatches through
// Factory.WithOlaresClient and the olaresclient.ComputeOps capability:
//
//   - list          works on both lines; renders the matching shape per the
//                   detected backend version (legacy GPUInfo vs ComputeResourceNode).
//   - bindings      1.12.6+ only (ErrUnsupportedVersion below).
//   - unbind        1.12.6+ only.
//   - support-type  1.12.6+ only.
//
// Subcommand VISIBILITY (root.go) is driven by the LOCALLY CACHED backend
// version so `--help` / completion stay offline; runtime CORRECTNESS is still
// enforced by WithOlaresClient + the capability gate, so a stale cache only
// affects what help advertises, never what actually runs.
package gpu

import (
	"encoding/json"
	"fmt"
	"io"
	"os"
	"strings"

	"github.com/Masterminds/semver/v3"
	"github.com/spf13/cobra"

	"github.com/beclab/Olares/cli/pkg/utils"
)

// computeResourcesFloor is the first Olares line that speaks the
// /api/compute-resources model. At or above this (core) version the new
// surface is used; below it the legacy /api/gpu/list shape is rendered.
var computeResourcesFloor = semver.MustParse("1.12.6")

// usesComputeResources reports whether version v should be treated as the
// new compute-resources line. A nil version (unknown / no cache) optimistically
// assumes the newest surface — the runtime capability gate corrects an actually
// older backend.
func usesComputeResources(v *semver.Version) bool {
	if v == nil {
		return true
	}
	return !utils.CoreVersion(v).LessThan(computeResourcesFloor)
}

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

func boolStr(b bool) string {
	if b {
		return "yes"
	}
	return "no"
}

// formatMem renders a byte count as a short "12G" / "1.5G" label, matching the
// SPA's formatComputeMemory. Zero / negative renders as "0G".
func formatMem(bytes int64) string {
	if bytes <= 0 {
		return "0G"
	}
	g := float64(bytes) / (1024 * 1024 * 1024)
	if g >= 10 {
		return fmt.Sprintf("%dG", int64(g+0.5))
	}
	return fmt.Sprintf("%.1fG", g)
}

// decodeData unmarshals an olaresclient op's returned `data` payload, tolerating
// an empty/absent body (treated as the zero value of T).
func decodeData[T any](raw json.RawMessage, out *T) error {
	if len(raw) == 0 || string(raw) == "null" {
		return nil
	}
	if err := json.Unmarshal(raw, out); err != nil {
		return fmt.Errorf("decode response: %w", err)
	}
	return nil
}
