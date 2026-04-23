package oac

import (
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strconv"
	"strings"
	"testing"

	validation "github.com/go-ozzo/ozzo-validation/v4"
	"gopkg.in/yaml.v3"
	"helm.sh/helm/v3/pkg/kube"
	appsv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/client-go/kubernetes/scheme"

	"github.com/beclab/Olares/framework/oac/internal/manifest"
	"github.com/beclab/Olares/framework/oac/internal/resources"
)

// appResult captures the per-app outcome of the three OAC checks the fleet
// report exercises. It is shared between the per-fleet report and the
// cross-fleet comparison so both views agree on what "FAIL" means for a
// given app.
type appResult struct {
	name        string
	folderErr   error
	manifestErr error
	resourceErr error
}

// fleetReport bundles a fleet's identity (display name + on-disk root) with
// the per-app results scanned from it. The cross-fleet comparison consumes
// pairs of these to produce a diff section.
type fleetReport struct {
	name      string
	dir       string
	rows      []appResult
	skippedV2 []string
}

// TestAppsTestdataReport walks every immediate subdirectory of both
// testdata/apps/ and testdata/terminus-apps/ and runs the standard OAC
// checks against each one, then writes:
//
//  1. one markdown report per fleet (apps_report.md, terminus_apps_report.md)
//  2. a comparison report (apps_diff_report.md) highlighting membership
//     differences and per-check status drift between the two fleets.
//
// Each app is exercised through three independent checks:
//
//  1. CheckChartFolder  — folder layout / naming
//  2. ValidateManifestFile — OlaresManifest.yaml schema + cross-field rules
//  3. CheckResources — helm dry-run + container resource limit envelope
//
// Per-app failures are recorded in the report rather than failing the test;
// only filesystem / IO errors abort. Report paths default to filenames in
// the package directory and can be redirected with the
// OAC_APPS_REPORT, OAC_TERMINUS_APPS_REPORT, OAC_APPS_DIFF_REPORT
// environment variables.
//
// The test is skipped under -short because rendering ~600 charts (across
// both fleets) via Helm takes ~60s on a warm machine.
func TestAppsTestdataReport(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping apps fleet report in -short mode")
	}

	fleets := []struct {
		name   string
		dir    string
		envOut string
		defOut string
	}{
		{
			name:   "apps",
			dir:    filepath.Join("testdata", "apps"),
			envOut: "OAC_APPS_REPORT",
			defOut: "apps_report.md",
		},
		{
			name:   "terminus-apps",
			dir:    filepath.Join("testdata", "terminus-apps"),
			envOut: "OAC_TERMINUS_APPS_REPORT",
			defOut: "terminus_apps_report.md",
		},
	}

	reports := make([]fleetReport, 0, len(fleets))
	for _, f := range fleets {
		if _, err := os.Stat(f.dir); errors.Is(err, os.ErrNotExist) {
			t.Logf("skipping fleet %s: %s does not exist", f.name, f.dir)
			continue
		} else if err != nil {
			t.Fatalf("stat fleet dir %s: %v", f.dir, err)
		}

		rows, skippedV2, err := scanFleet(f.dir)
		if err != nil {
			t.Fatalf("scan fleet %s: %v", f.name, err)
		}

		out := os.Getenv(f.envOut)
		if out == "" {
			out = f.defOut
		}
		report := buildFleetReport(f.name, f.dir, rows, skippedV2)
		if err := os.WriteFile(out, []byte(report), 0o644); err != nil {
			t.Fatalf("write %s report %s: %v", f.name, out, err)
		}

		passed := countPassedApps(rows)
		t.Logf("%s report written to %s (total=%d, passed=%d, failed=%d, v2-skipped=%d)",
			f.name, out, len(rows), passed, len(rows)-passed, len(skippedV2))
		if len(skippedV2) > 0 {
			t.Logf("%s skipped %d apiVersion=v2 app(s): %s",
				f.name, len(skippedV2), strings.Join(skippedV2, ", "))
		}

		reports = append(reports, fleetReport{name: f.name, dir: f.dir, rows: rows, skippedV2: skippedV2})
	}

	if len(reports) >= 2 {
		diffOut := os.Getenv("OAC_APPS_DIFF_REPORT")
		if diffOut == "" {
			diffOut = "apps_diff_report.md"
		}
		diff := buildFleetDiffReport(reports[0], reports[1])
		if err := os.WriteFile(diffOut, []byte(diff), 0o644); err != nil {
			t.Fatalf("write diff report %s: %v", diffOut, err)
		}
		t.Logf("fleet diff report written to %s (%s vs %s)",
			diffOut, reports[0].name, reports[1].name)
	}
}

// scanFleet walks the immediate subdirectories of root, runs the three
// OAC checks against each app, and returns the result rows sorted by name
// alongside the names of any apps skipped because their manifest declares
// apiVersion: v2.
//
// v2 apps are intentionally excluded from the report: the v2 install path
// is a multi-chart layout that requires per-subchart helm context the
// fleet harness does not synthesize, so running the v1/v3 oriented
// CheckResources / ValidateManifestFile against them produces noise
// rather than signal. Hidden directories and any directory containing a
// .remove marker file are likewise silently ignored.
func scanFleet(root string) ([]appResult, []string, error) {
	entries, err := os.ReadDir(root)
	if err != nil {
		return nil, nil, fmt.Errorf("read fleet dir %s: %w", root, err)
	}
	var (
		rows      []appResult
		skippedV2 []string
	)
	for _, e := range entries {
		if !e.IsDir() {
			continue
		}
		if strings.HasPrefix(e.Name(), ".") {
			continue
		}
		dir := filepath.Join(root, e.Name())
		if hasSkipMarker(dir) {
			continue
		}
		if isV2ManifestDir(dir) {
			skippedV2 = append(skippedV2, e.Name())
			continue
		}
		c := New(WithOwner("alice"), WithAdmin("admin"))
		row := appResult{name: e.Name()}
		row.folderErr = runCheck(func() error { return c.CheckChartFolder(dir) })
		row.manifestErr = runCheck(func() error { return c.ValidateManifestFile(dir) })
		row.resourceErr = runCheck(func() error { return c.CheckResources(dir) })
		rows = append(rows, row)
	}
	sort.Slice(rows, func(i, j int) bool { return rows[i].name < rows[j].name })
	sort.Strings(skippedV2)
	return rows, skippedV2, nil
}

// isV2ManifestDir reports whether oacPath/OlaresManifest.yaml declares
// apiVersion: v2 (case-insensitive). Read or peek failures fall through
// as "not v2" so the regular check pipeline can surface those errors --
// scanFleet is meant to skip valid-but-unsupported v2 manifests, not
// hide structural problems behind a silent skip.
func isV2ManifestDir(oacPath string) bool {
	raw, err := readManifestFile(oacPath)
	if err != nil {
		return false
	}
	v, err := PeekManifestVersions(raw)
	if err != nil {
		return false
	}
	return strings.EqualFold(v.APIVersion, manifest.APIVersionV2)
}

// countPassedApps returns the number of rows where every check succeeded.
func countPassedApps(rows []appResult) int {
	var passed int
	for _, r := range rows {
		if r.folderErr == nil && r.manifestErr == nil && r.resourceErr == nil {
			passed++
		}
	}
	return passed
}

// buildFleetReport produces the markdown body for a single fleet, including
// summary table, per-app status grid, and detailed failure diagnostics.
// skippedV2 lists apps that scanFleet skipped because their manifest
// declares apiVersion: v2; they are surfaced in the summary so the report
// reader can tell why the fleet directory entry count and the report row
// count disagree.
func buildFleetReport(name, dir string, rows []appResult, skippedV2 []string) string {
	totalApps := len(rows)
	var (
		passApps      int
		folderFails   int
		manifestFails int
		resourceFails int
	)
	for _, r := range rows {
		if r.folderErr == nil && r.manifestErr == nil && r.resourceErr == nil {
			passApps++
		}
		if r.folderErr != nil {
			folderFails++
		}
		if r.manifestErr != nil {
			manifestFails++
		}
		if r.resourceErr != nil {
			resourceFails++
		}
	}

	var b strings.Builder
	fmt.Fprintf(&b, "# OAC %s testdata report\n\n", name)
	fmt.Fprintf(&b, "_Generated by `TestAppsTestdataReport` against `framework/oac/%s/`._\n\n", dir)
	fmt.Fprintf(&b, "## Summary\n\n")
	fmt.Fprintf(&b, "| Total | Passed | Failed | Folder fails | Manifest fails | Resource fails | v2 skipped |\n")
	fmt.Fprintf(&b, "|---:|---:|---:|---:|---:|---:|---:|\n")
	fmt.Fprintf(&b, "| %d | %d | %d | %d | %d | %d | %d |\n\n",
		totalApps, passApps, totalApps-passApps,
		folderFails, manifestFails, resourceFails, len(skippedV2))

	if len(skippedV2) > 0 {
		fmt.Fprintf(&b, "## Skipped (apiVersion=v2)\n\n")
		fmt.Fprintf(&b, "v2 manifests follow the multi-chart install layout and are intentionally excluded from this report:\n\n")
		for _, name := range skippedV2 {
			fmt.Fprintf(&b, "- %s\n", name)
		}
		b.WriteByte('\n')
	}

	fmt.Fprintf(&b, "## Per-app results\n\n")
	fmt.Fprintf(&b, "| App | Folder | Manifest | Resources |\n")
	fmt.Fprintf(&b, "|---|:-:|:-:|:-:|\n")
	for _, r := range rows {
		fmt.Fprintf(&b, "| %s | %s | %s | %s |\n",
			r.name,
			cellStatus(r.folderErr),
			cellStatus(r.manifestErr),
			cellStatus(r.resourceErr),
		)
	}

	if totalApps != passApps {
		fmt.Fprintf(&b, "\n## Failure details\n\n")
		for _, r := range rows {
			if r.folderErr == nil && r.manifestErr == nil && r.resourceErr == nil {
				continue
			}
			fmt.Fprintf(&b, "### %s\n\n", r.name)
			if r.folderErr != nil {
				fmt.Fprintf(&b, "- **CheckChartFolder**: %s\n", flattenError(r.folderErr.Error()))
			}
			if r.manifestErr != nil {
				fmt.Fprintf(&b, "- **ValidateManifestFile**: %s\n", flattenError(r.manifestErr.Error()))
				if diag := diagnoseManifest(filepath.Join(dir, r.name), r.manifestErr); diag != "" {
					b.WriteString(diag)
				}
			}
			if r.resourceErr != nil {
				fmt.Fprintf(&b, "- **CheckResources**: %s\n", flattenError(r.resourceErr.Error()))
				if diag := diagnoseResources(filepath.Join(dir, r.name)); diag != "" {
					b.WriteString(diag)
				}
			}
			b.WriteByte('\n')
		}
	}
	return b.String()
}

// buildFleetDiffReport renders a markdown comparison between two fleet
// scans. It surfaces (1) summary deltas, (2) apps unique to either side,
// and (3) common apps whose per-check pass/fail status diverges. Apps that
// behave identically in both fleets are excluded so the report stays
// focused on actionable drift.
func buildFleetDiffReport(a, b fleetReport) string {
	indexA := indexResults(a.rows)
	indexB := indexResults(b.rows)

	var (
		onlyA   []string
		onlyB   []string
		common  []string
		drifted []string
		// counters for status drift summary
		folderDrift   int
		manifestDrift int
		resourceDrift int
	)
	allNames := unionKeys(indexA, indexB)
	for _, name := range allNames {
		_, inA := indexA[name]
		_, inB := indexB[name]
		switch {
		case inA && !inB:
			onlyA = append(onlyA, name)
		case !inA && inB:
			onlyB = append(onlyB, name)
		default:
			common = append(common, name)
			ra, rb := indexA[name], indexB[name]
			folder := statusChanged(ra.folderErr, rb.folderErr)
			manifest := statusChanged(ra.manifestErr, rb.manifestErr)
			resource := statusChanged(ra.resourceErr, rb.resourceErr)
			if folder {
				folderDrift++
			}
			if manifest {
				manifestDrift++
			}
			if resource {
				resourceDrift++
			}
			if folder || manifest || resource {
				drifted = append(drifted, name)
			}
		}
	}

	var out strings.Builder
	fmt.Fprintf(&out, "# OAC fleet comparison: %s vs %s\n\n", a.name, b.name)
	fmt.Fprintf(&out, "_Generated by `TestAppsTestdataReport` comparing `framework/oac/%s/` and `framework/oac/%s/`._\n\n", a.dir, b.dir)

	out.WriteString("## Summary\n\n")
	out.WriteString("| Fleet | Total | Passed | Failed | Folder fails | Manifest fails | Resource fails |\n")
	out.WriteString("|---|---:|---:|---:|---:|---:|---:|\n")
	for _, fr := range []fleetReport{a, b} {
		var folderF, manifestF, resourceF int
		for _, r := range fr.rows {
			if r.folderErr != nil {
				folderF++
			}
			if r.manifestErr != nil {
				manifestF++
			}
			if r.resourceErr != nil {
				resourceF++
			}
		}
		passed := countPassedApps(fr.rows)
		fmt.Fprintf(&out, "| %s | %d | %d | %d | %d | %d | %d |\n",
			fr.name, len(fr.rows), passed, len(fr.rows)-passed,
			folderF, manifestF, resourceF)
	}
	out.WriteString("\n")

	fmt.Fprintf(&out, "## Membership\n\n")
	fmt.Fprintf(&out, "- Common apps: **%d**\n", len(common))
	fmt.Fprintf(&out, "- Only in `%s`: **%d**\n", a.name, len(onlyA))
	fmt.Fprintf(&out, "- Only in `%s`: **%d**\n", b.name, len(onlyB))
	out.WriteString("\n")
	if len(onlyA) > 0 {
		fmt.Fprintf(&out, "<details><summary>Apps only in <code>%s</code> (%d)</summary>\n\n", a.name, len(onlyA))
		for _, n := range onlyA {
			fmt.Fprintf(&out, "- %s\n", n)
		}
		out.WriteString("\n</details>\n\n")
	}
	if len(onlyB) > 0 {
		fmt.Fprintf(&out, "<details><summary>Apps only in <code>%s</code> (%d)</summary>\n\n", b.name, len(onlyB))
		for _, n := range onlyB {
			fmt.Fprintf(&out, "- %s\n", n)
		}
		out.WriteString("\n</details>\n\n")
	}

	fmt.Fprintf(&out, "## Status drift on common apps\n\n")
	fmt.Fprintf(&out, "Apps whose per-check status differs between fleets: **%d** "+
		"(folder: %d, manifest: %d, resources: %d).\n\n",
		len(drifted), folderDrift, manifestDrift, resourceDrift)
	if len(drifted) > 0 {
		fmt.Fprintf(&out, "| App | Folder (%s → %s) | Manifest (%s → %s) | Resources (%s → %s) |\n",
			a.name, b.name, a.name, b.name, a.name, b.name)
		out.WriteString("|---|:-:|:-:|:-:|\n")
		for _, name := range drifted {
			ra, rb := indexA[name], indexB[name]
			fmt.Fprintf(&out, "| %s | %s | %s | %s |\n",
				name,
				diffCell(ra.folderErr, rb.folderErr),
				diffCell(ra.manifestErr, rb.manifestErr),
				diffCell(ra.resourceErr, rb.resourceErr),
			)
		}
		out.WriteString("\n")

		out.WriteString("### Drift details\n\n")
		for _, name := range drifted {
			ra, rb := indexA[name], indexB[name]
			fmt.Fprintf(&out, "#### %s\n\n", name)
			renderDriftDetail(&out, "CheckChartFolder", a.name, b.name, ra.folderErr, rb.folderErr)
			renderDriftDetail(&out, "ValidateManifestFile", a.name, b.name, ra.manifestErr, rb.manifestErr)
			renderDriftDetail(&out, "CheckResources", a.name, b.name, ra.resourceErr, rb.resourceErr)
			out.WriteString("\n")
		}
	}
	return out.String()
}

// indexResults flattens a slice of appResult into a name -> result lookup
// for the cross-fleet diff. It is intentionally small; the fleets are
// always sorted upstream so the order is preserved through the union.
func indexResults(rows []appResult) map[string]appResult {
	m := make(map[string]appResult, len(rows))
	for _, r := range rows {
		m[r.name] = r
	}
	return m
}

// unionKeys returns the sorted union of keys across the two maps so the
// diff report visits each app at most once and in a stable order.
func unionKeys(a, b map[string]appResult) []string {
	seen := make(map[string]struct{}, len(a)+len(b))
	for k := range a {
		seen[k] = struct{}{}
	}
	for k := range b {
		seen[k] = struct{}{}
	}
	out := make([]string, 0, len(seen))
	for k := range seen {
		out = append(out, k)
	}
	sort.Strings(out)
	return out
}

// statusChanged reports whether the pass/fail bit of a check flipped
// between two fleets. It deliberately ignores error message differences
// when both sides fail; the goal is to surface regressions, not noise.
func statusChanged(a, b error) bool {
	return (a == nil) != (b == nil)
}

// diffCell renders a one-cell summary of a status transition between
// fleets, e.g. "ok→FAIL" or "FAIL→ok". Identical statuses collapse to
// "ok" / "FAIL" so the column still encodes the current state.
func diffCell(a, b error) string {
	sa, sb := cellStatus(a), cellStatus(b)
	if sa == sb {
		return sa
	}
	return sa + "→" + sb
}

// renderDriftDetail prints a per-check diff block when statuses differ,
// including the offending error message from whichever side(s) failed.
// It emits nothing when both sides agree.
func renderDriftDetail(out *strings.Builder, check, aName, bName string, aErr, bErr error) {
	if !statusChanged(aErr, bErr) {
		return
	}
	fmt.Fprintf(out, "- **%s**: %s=%s, %s=%s\n",
		check,
		aName, errSummary(aErr),
		bName, errSummary(bErr),
	)
}

// errSummary returns a compact, single-line description of an error for
// embedding in markdown bullets. Nil errors render as "ok" so cells stay
// uniform.
func errSummary(err error) string {
	if err == nil {
		return "ok"
	}
	return flattenError(err.Error())
}

// hasSkipMarker reports whether a fleet entry should be excluded from the
// report. We honour two sibling marker files: `.remove` for charts that
// have been retired, and `.suspend` for charts that are temporarily
// disabled (e.g. while their helm rendering is broken upstream). Either
// marker causes the entry to be silently dropped from the scan.
func hasSkipMarker(dir string) bool {
	for _, marker := range []string{".remove", ".suspend"} {
		if _, err := os.Stat(filepath.Join(dir, marker)); err == nil {
			return true
		}
	}
	return false
}

// cellStatus renders a per-check status cell as "ok" / "FAIL".
func cellStatus(err error) string {
	if err == nil {
		return "ok"
	}
	return "FAIL"
}

// runCheck invokes fn and converts any panic into a regular error so a
// single misbehaving chart cannot take down the entire fleet report. Helm's
// install action is known to panic on certain malformed inputs.
func runCheck(fn func() error) (err error) {
	defer func() {
		if r := recover(); r != nil {
			err = fmt.Errorf("panic: %v", r)
		}
	}()
	return fn()
}

// flattenError collapses a multi-line error message into a single, escaped
// fragment so it fits inside a markdown bullet without breaking layout.
func flattenError(s string) string {
	s = strings.ReplaceAll(s, "\r\n", " ")
	s = strings.ReplaceAll(s, "\n", " ")
	s = strings.ReplaceAll(s, "|", "\\|")
	if len(s) > 800 {
		s = s[:800] + "…"
	}
	return s
}

// diagnoseManifest is invoked when ValidateManifestFile fails. It pulls
// every per-field validation error out of the wrapped chain, walks the raw
// OlaresManifest.yaml as a generic YAML tree, and prints the actual value
// observed at each offending field path. This lets the reader see "got
// openMethod=\"newtab\"" alongside the rule's "must be one of …" message
// without having to open the source manifest by hand.
//
// It tolerates any combination of missing inner errors / unreadable YAML
// and returns an empty string in that case so the report still renders.
func diagnoseManifest(oacPath string, err error) string {
	verrs, ok := extractValidationErrors(err)
	if !ok {
		return ""
	}
	flat := flattenValidationErrors("", verrs)
	if len(flat) == 0 {
		return ""
	}
	sort.Slice(flat, func(i, j int) bool { return flat[i].path < flat[j].path })

	var doc interface{}
	if raw, readErr := readManifestFile(oacPath); readErr == nil {
		_ = yaml.Unmarshal(raw, &doc)
	}

	var out strings.Builder
	out.WriteString("\n  <details><summary>Manifest field values</summary>\n\n")
	for _, f := range flat {
		actual, found := lookupYAMLPath(doc, f.path)
		display := "<not set>"
		if found {
			display = formatYAMLValue(actual)
		}
		fmt.Fprintf(&out, "  - `%s`: actual=%s — %s\n",
			f.path, display, flattenError(f.reason))
	}
	out.WriteString("  </details>\n")
	return out.String()
}

// extractValidationErrors walks the error chain looking for the underlying
// ozzo validation.Errors map produced by ValidateStruct. It first tries the
// standard errors.As walk; if that fails it falls back to peeking through a
// *ValidationError so callers don't have to know which wrapper produced the
// error.
func extractValidationErrors(err error) (validation.Errors, bool) {
	var verrs validation.Errors
	if errors.As(err, &verrs) {
		return verrs, true
	}
	var vErr *ValidationError
	if errors.As(err, &vErr) && vErr.Inner != nil {
		if errors.As(vErr.Inner, &verrs) {
			return verrs, true
		}
	}
	return nil, false
}

// lookupYAMLPath walks a dotted path (e.g. "entrances.0.openMethod") through
// the YAML tree produced by yaml.v3 and returns the resolved value plus a
// found flag. Numeric path segments index into sequences; everything else
// is treated as a mapping key.
func lookupYAMLPath(root interface{}, path string) (interface{}, bool) {
	cur := root
	for _, part := range strings.Split(path, ".") {
		if cur == nil {
			return nil, false
		}
		switch v := cur.(type) {
		case map[string]interface{}:
			n, ok := v[part]
			if !ok {
				return nil, false
			}
			cur = n
		case map[interface{}]interface{}:
			n, ok := v[part]
			if !ok {
				return nil, false
			}
			cur = n
		case []interface{}:
			i, perr := strconv.Atoi(part)
			if perr != nil || i < 0 || i >= len(v) {
				return nil, false
			}
			cur = v[i]
		default:
			return nil, false
		}
	}
	return cur, true
}

// formatYAMLValue renders a YAML value into a single-line, markdown-table-
// safe representation: scalars get %q / %v, structured values get marshalled
// back to YAML and flattened.
func formatYAMLValue(v interface{}) string {
	switch x := v.(type) {
	case nil:
		return "<nil>"
	case string:
		return fmt.Sprintf("%q", x)
	case bool, int, int32, int64, uint, uint32, uint64, float32, float64:
		return fmt.Sprintf("%v", x)
	}
	data, err := yaml.Marshal(v)
	if err != nil {
		return flattenError(fmt.Sprintf("%v", v))
	}
	return "`" + flattenError(strings.TrimSpace(string(data))) + "`"
}

// containerObservation captures one container's resource ask/cap as seen
// by a helm dry-run, tagged with the workload it lives in. The Subchart
// field is empty for non-multi-chart (apiVersion=v1 / legacy) renders.
type containerObservation struct {
	Subchart  string
	Kind      string
	Workload  string
	Container string
	ReqCPU    string
	ReqMem    string
	LimCPU    string
	LimMem    string
}

// diagnoseResources is invoked when CheckResources fails. It produces a
// markdown snippet that lists the manifest-declared resource envelope plus
// every Deployment/StatefulSet container's actual requests/limits, so the
// reader can immediately see why "sum of container resources.requests.cpu
// must be <= spec.requiredCpu"-style errors fired and which workload is
// over-budget. It tolerates panics and helm errors so it never aborts the
// report.
func diagnoseResources(oacPath string) string {
	c := New(WithOwner("alice"), WithAdmin("admin"))
	var m Manifest
	loadErr := runCheck(func() error {
		var err error
		m, err = c.LoadManifestFile(oacPath)
		return err
	})
	if loadErr != nil || m == nil {
		return ""
	}
	cfg, ok := m.Raw().(*manifest.AppConfiguration)
	if !ok {
		return ""
	}

	var out strings.Builder
	out.WriteString("\n  <details><summary>Resource diagnostic</summary>\n\n")

	limitLines := manifestLimitLines(cfg)
	if len(limitLines) > 0 {
		out.WriteString("  Manifest limits:\n\n")
		for _, l := range limitLines {
			fmt.Fprintf(&out, "  - %s\n", l)
		}
		out.WriteString("\n")
	}

	observations := collectContainerObservations(c, m, oacPath)
	if len(observations) > 0 {
		out.WriteString("  Container observations (helm dry-run):\n\n")
		out.WriteString("  | Subchart | Kind | Workload | Container | requests.cpu | requests.memory | limits.cpu | limits.memory |\n")
		out.WriteString("  |---|---|---|---|---|---|---|---|\n")
		for _, o := range observations {
			fmt.Fprintf(&out, "  | %s | %s | %s | %s | %s | %s | %s | %s |\n",
				dashIfEmpty(o.Subchart),
				o.Kind, o.Workload, o.Container,
				dashIfEmpty(o.ReqCPU), dashIfEmpty(o.ReqMem),
				dashIfEmpty(o.LimCPU), dashIfEmpty(o.LimMem),
			)
		}
		out.WriteString("\n")
	}

	out.WriteString("  </details>\n")
	return out.String()
}

// manifestLimitLines summarises every limit envelope the manifest declares,
// in the same shape CheckResourceLimits compares against: legacy flat
// fields or modern inline ResourceRequirement per mode.
func manifestLimitLines(cfg *manifest.AppConfiguration) []string {
	var lines []string
	if !manifest.IsModernResourcesManifest(cfg.ConfigVersion) {
		lines = append(lines, fmt.Sprintf(
			"legacy flat: requiredCpu=%s, limitedCpu=%s, requiredMemory=%s, limitedMemory=%s",
			displayQuantity(cfg.Spec.RequiredCPU), displayQuantity(cfg.Spec.LimitedCPU),
			displayQuantity(cfg.Spec.RequiredMemory), displayQuantity(cfg.Spec.LimitedMemory),
		))
		return lines
	}
	for _, rm := range cfg.Spec.Resources {
		lines = append(lines, fmt.Sprintf(
			"mode=%s: requiredCpu=%s, limitedCpu=%s, requiredMemory=%s, limitedMemory=%s",
			rm.Mode,
			displayQuantity(rm.RequiredCPU), displayQuantity(rm.LimitedCPU),
			displayQuantity(rm.RequiredMemory), displayQuantity(rm.LimitedMemory),
		))
	}
	return lines
}

// collectContainerObservations renders the chart the same way CheckResources
// does (per-subchart for v2 multi-chart, oacPath alone otherwise) and
// flattens every Deployment/StatefulSet container into a stable row set.
// Render errors are silently swallowed so the diagnostic still emits the
// manifest-side limits when helm itself is the problem.
func collectContainerObservations(c *OAC, m Manifest, oacPath string) []containerObservation {
	sc := ownerScenario{owner: c.owner, admin: c.admin}
	if isV2Manifest(m) {
		cfg, ok := m.Raw().(*manifest.AppConfiguration)
		if !ok {
			return nil
		}
		var out []containerObservation
		for _, sub := range cfg.Spec.SubCharts {
			var subList kube.ResourceList
			_ = runCheck(func() error {
				var err error
				subList, err = c.renderAllSubCharts(oacPath, m, sc, []manifest.Chart{sub}, "")
				return err
			})
			out = append(out, observationsFromList(subList, sub.Name)...)
		}
		return out
	}
	var list kube.ResourceList
	_ = runCheck(func() error {
		var err error
		list, err = c.renderForStructuralChecks(oacPath, m, sc)
		return err
	})
	return observationsFromList(list, "")
}

// observationsFromList walks a single helm-rendered kube.ResourceList and
// emits one containerObservation per primary container on every Deployment
// and StatefulSet, mirroring the workloads CheckResourceLimits actually
// inspects.
func observationsFromList(list kube.ResourceList, subchart string) []containerObservation {
	var out []containerObservation
	for _, r := range list {
		kind := r.Object.GetObjectKind().GroupVersionKind().Kind
		switch kind {
		case resources.KindDeployment:
			var dep appsv1.Deployment
			if err := scheme.Scheme.Convert(r.Object, &dep, nil); err != nil {
				continue
			}
			for _, c := range dep.Spec.Template.Spec.Containers {
				out = append(out, observationFromContainer(subchart, kind, dep.Name, c))
			}
		case resources.KindStatefulSet:
			var sts appsv1.StatefulSet
			if err := scheme.Scheme.Convert(r.Object, &sts, nil); err != nil {
				continue
			}
			for _, c := range sts.Spec.Template.Spec.Containers {
				out = append(out, observationFromContainer(subchart, kind, sts.Name, c))
			}
		}
	}
	return out
}

func observationFromContainer(subchart, kind, wlName string, c corev1.Container) containerObservation {
	o := containerObservation{
		Subchart:  subchart,
		Kind:      kind,
		Workload:  wlName,
		Container: c.Name,
	}
	if q := c.Resources.Requests.Cpu(); q != nil && !q.IsZero() {
		o.ReqCPU = q.String()
	}
	if q := c.Resources.Requests.Memory(); q != nil && !q.IsZero() {
		o.ReqMem = q.String()
	}
	if q := c.Resources.Limits.Cpu(); q != nil && !q.IsZero() {
		o.LimCPU = q.String()
	}
	if q := c.Resources.Limits.Memory(); q != nil && !q.IsZero() {
		o.LimMem = q.String()
	}
	return o
}

func dashIfEmpty(s string) string {
	if s == "" {
		return "-"
	}
	return s
}

// displayQuantity renders an empty quantity string as "-" so the manifest
// limit table reads cleanly even when a field is omitted.
func displayQuantity(s string) string {
	if s == "" {
		return "-"
	}
	return s
}
