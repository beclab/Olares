package gpu

import (
	"context"
	"fmt"
	"io"
	"net/http"
	"os"
	"strings"
	"sync"
	"time"

	pkgdashboard "github.com/beclab/Olares/cli/pkg/dashboard"
)

// RunTaskByRef is the cmd-side entry point for `dashboard
// overview gpu tasks <ref>` — the SPA-aligned shorthand. `<ref>`
// matches the row's `name` OR its `podUid` (the two columns the
// SPA's TasksTable surfaces, and the two values the CLI prints in
// the bare `gpu` / `gpu tasks` listings — users naturally
// copy-paste either). The function reverse-resolves the task's
// pod-uid + sharemode by listing HAMI's /v1/containers (same path
// the SPA uses when the user clicks "View details"; that click
// passes podUid + deviceShareModes[0] into the route param). When
// the ref is unique we delegate to RunTaskDetail; ambiguity /
// not-found surface as typed errors with a copy-paste-friendly
// hint listing every candidate (so an agent doesn't have to round-
// trip the listing endpoint to recover). Empty container list →
// standard `no_gpu_detected` envelope (single fetch, no fan-out).
//
// Pod-uid match wins ties: if the same string somehow matches both
// a `name` and a `podUid` across two rows (impossible in practice
// because pod-uids are RFC 4122 UUIDs, but defensive nonetheless),
// the pod-uid match is used since it's globally unique. Equal-name
// rows still produce an ambiguity error pointing at the pod-uids.
func RunTaskByRef(ctx context.Context, c *pkgdashboard.Client, cf *pkgdashboard.CommonFlags, ref string) error {
	now := time.Now()
	advisoryNote, _ := pkgdashboard.GPUAdvisory(ctx, c, cf, os.Stderr)
	list, err := pkgdashboard.FetchTaskList(ctx, c, nil)
	if err != nil {
		if he, ok := pkgdashboard.IsHTTPError(err); ok && he.Status == http.StatusNotFound {
			env := pkgdashboard.Envelope{
				Kind: pkgdashboard.KindOverviewGPUTaskDetFull,
				Meta: pkgdashboard.NewMeta(now.In(cf.Timezone.Time()), c.OlaresID(), cf.User),
			}
			env.Meta.Empty = true
			env.Meta.EmptyReason = "no_vgpu_integration"
			env.Meta.HTTPStatus = he.Status
			if advisoryNote != "" {
				env.Meta.Note = advisoryNote
			}
			if cf.Output == pkgdashboard.OutputJSON {
				return pkgdashboard.WriteJSON(os.Stdout, env)
			}
			fmt.Fprintln(os.Stdout, "(task not found — HAMI integration absent or task ref invalid)")
			return nil
		}
		if unavail, ok := pkgdashboard.VgpuUnavailableFromError(c, cf, err, pkgdashboard.KindOverviewGPUTaskDetFull, now, os.Stderr); ok {
			if advisoryNote != "" {
				unavail.Meta.Note = advisoryNote + " | " + unavail.Meta.Note
			}
			if cf.Output == pkgdashboard.OutputJSON {
				return pkgdashboard.WriteJSON(os.Stdout, unavail)
			}
			return nil
		}
		return err
	}

	// First pass: exact pod-uid match (globally unique, wins over
	// name match if both happen to coincide).
	for _, t := range list {
		if fmt.Sprintf("%v", t["podUid"]) == ref {
			name := fmt.Sprintf("%v", t["name"])
			podUID := fmt.Sprintf("%v", t["podUid"])
			sharemode := fmt.Sprintf("%v", firstAnyInArray(t["deviceShareModes"]))
			return RunTaskDetail(ctx, c, cf, name, podUID, sharemode)
		}
	}
	// Second pass: name match — collect all to detect ambiguity.
	nameMatches := make([]map[string]any, 0, 2)
	for _, t := range list {
		if fmt.Sprintf("%v", t["name"]) == ref {
			nameMatches = append(nameMatches, t)
		}
	}
	switch len(nameMatches) {
	case 0:
		env := pkgdashboard.Envelope{
			Kind: pkgdashboard.KindOverviewGPUTaskDetFull,
			Meta: pkgdashboard.NewMeta(now.In(cf.Timezone.Time()), c.OlaresID(), cf.User),
		}
		env.Meta.Empty = true
		env.Meta.EmptyReason = "no_gpu_detected"
		if advisoryNote != "" {
			env.Meta.Note = advisoryNote
		}
		if cf.Output == pkgdashboard.OutputJSON {
			return pkgdashboard.WriteJSON(os.Stdout, env)
		}
		fmt.Fprintf(os.Stdout, "(no task matches %q — neither a name nor a pod-uid in HAMI's container list)\n", ref)
		// Nudge the user toward the listing command so they can
		// copy a real ref. Skip when the list itself is empty.
		if hint := candidateHintLine(list); hint != "" {
			fmt.Fprintln(os.Stdout, hint)
		}
		return nil
	case 1:
		t := nameMatches[0]
		name := fmt.Sprintf("%v", t["name"])
		podUID := fmt.Sprintf("%v", t["podUid"])
		sharemode := fmt.Sprintf("%v", firstAnyInArray(t["deviceShareModes"]))
		return RunTaskDetail(ctx, c, cf, name, podUID, sharemode)
	default:
		// Multiple tasks share the name. Mirror the SPA: it never
		// allows ambiguity because the user clicks the row directly,
		// but the CLI must produce a deterministic error AND the
		// disambiguating ref each pod-uid maps to so the user can
		// re-run without re-listing.
		uids := make([]string, 0, len(nameMatches))
		for _, m := range nameMatches {
			uids = append(uids, fmt.Sprintf("%v", m["podUid"]))
		}
		return fmt.Errorf("task name %q matches %d running pods (%s); rerun with one of the pod-uids: olares-cli dashboard overview gpu tasks <pod-uid>",
			ref, len(nameMatches), strings.Join(uids, ", "))
	}
}

// candidateHintLine renders a short "available refs" suggestion
// for the not-found prose path. Lists at most 5 distinct names +
// the matching pod-uid so the user has a concrete copy target;
// returns "" when the input list is empty (the calling site
// already prints a plain "not found" line in that case).
func candidateHintLine(list []map[string]any) string {
	if len(list) == 0 {
		return ""
	}
	const maxShown = 5
	pairs := make([]string, 0, maxShown)
	for i, t := range list {
		if i >= maxShown {
			break
		}
		pairs = append(pairs, fmt.Sprintf("%v (%v)", t["name"], t["podUid"]))
	}
	more := ""
	if len(list) > maxShown {
		more = fmt.Sprintf(", … +%d more", len(list)-maxShown)
	}
	return fmt.Sprintf("hint: try one of: %s%s", strings.Join(pairs, ", "), more)
}

// RunTaskDetail is the cmd-side entry point for `dashboard
// overview gpu task-detail <name> <pod-uid>`. Watch-aware (Runner
// with 30s recommended cadence); sharemode is the SPA-supplied
// share-mode flag and toggles the allocation gauges.
func RunTaskDetail(ctx context.Context, c *pkgdashboard.Client, cf *pkgdashboard.CommonFlags, name, podUID, sharemode string) error {
	r := &pkgdashboard.Runner{
		Flags:       cf,
		Recommended: 30 * time.Second,
		Iter: func(ctx context.Context, iter int, now time.Time) (pkgdashboard.Envelope, error) {
			start, end, since := ResolveDetailWindow(cf, now, TaskDetailDefaultSince)
			env, err := BuildTaskDetailFullEnvelope(ctx, c, cf, name, podUID, sharemode, start, end, since)
			if err != nil {
				return env, err
			}
			if cf.Output == pkgdashboard.OutputJSON {
				return env, nil
			}
			return env, WriteDetailFullTable(os.Stdout, env, cf)
		},
	}
	return r.Run(ctx)
}

// BuildTaskDetailFullEnvelope is the task-flavoured twin of
// BuildDetailFullEnvelope. The main difference is the placeholder
// substitution: `$container` / `$pod` / `$namespace` are pulled
// from the task detail itself (the SPA does the same — it can't
// fan out the monitor queries until /v1/container resolves).
func BuildTaskDetailFullEnvelope(ctx context.Context, c *pkgdashboard.Client, cf *pkgdashboard.CommonFlags, name, podUID, sharemode string, start, end time.Time, since time.Duration) (pkgdashboard.Envelope, error) {
	now := end
	advisoryNote, _ := pkgdashboard.GPUAdvisory(ctx, c, cf, os.Stderr)
	env := pkgdashboard.Envelope{
		Kind: pkgdashboard.KindOverviewGPUTaskDetFull,
		Meta: pkgdashboard.NewMeta(now.In(cf.Timezone.Time()), c.OlaresID(), cf.User),
	}
	if advisoryNote != "" {
		env.Meta.Note = advisoryNote
	}
	env.Meta.Window = &pkgdashboard.TimeWindow{
		Since: humanizeSince(since),
		Start: pkgdashboard.GPUTrendTimestampISO(start, cf.Timezone.Time()),
		End:   pkgdashboard.GPUTrendTimestampISO(end, cf.Timezone.Time()),
		Step:  pkgdashboard.GPUTrendStep(start, end),
	}
	// See BuildDetailFullEnvelope for the wire-vs-render TZ split
	// rationale. Same contract here.
	wireStart := pkgdashboard.GPUTrendTimestampWire(start)
	wireEnd := pkgdashboard.GPUTrendTimestampWire(end)

	detail, err := pkgdashboard.FetchTaskDetail(ctx, c, name, podUID, sharemode)
	if err != nil {
		if he, ok := pkgdashboard.IsHTTPError(err); ok && he.Status == http.StatusNotFound {
			env.Meta.Empty = true
			env.Meta.EmptyReason = "no_vgpu_integration"
			env.Meta.HTTPStatus = he.Status
			return env, nil
		}
		if unavail, ok := pkgdashboard.VgpuUnavailableFromError(c, cf, err, pkgdashboard.KindOverviewGPUTaskDetFull, now, os.Stderr); ok {
			if env.Meta.Note != "" {
				unavail.Meta.Note = env.Meta.Note + " | " + unavail.Meta.Note
			}
			unavail.Meta.Window = env.Meta.Window
			return unavail, nil
		}
		return env, err
	}
	if len(detail) == 0 {
		env.Meta.Empty = true
		env.Meta.EmptyReason = "no_gpu_detected"
		return env, nil
	}

	// SPA: const displayAllocation = sharemode !== TimeSlicing
	allocationGauges := sharemode != "2"
	gaugeSpecs := taskDetailGaugeSpecs(allocationGauges)
	trendSpecs := taskDetailTrendSpecs()
	repl := strings.NewReplacer(
		"$container", fmt.Sprintf("%v", detail["name"]),
		"$pod", fmt.Sprintf("%v", detail["appName"]),
		"$namespace", fmt.Sprintf("%v", detail["namespace"]),
	)

	var (
		warnMu   sync.Mutex
		warnings []string
	)
	addWarning := func(msg string) {
		warnMu.Lock()
		warnings = append(warnings, msg)
		warnMu.Unlock()
	}
	gaugeItems, trendItems, _ := fanoutGaugeAndTrend(
		ctx, c,
		gaugeSpecs, trendSpecs,
		repl, wireStart, wireEnd, env.Meta.Window.Step,
		cf.Timezone.Time(),
		addWarning,
		"", // task-detail page doesn't capture power_trend labels
	)

	detailEnv := pkgdashboard.Envelope{
		Kind:  pkgdashboard.KindOverviewGPUTaskDet,
		Meta:  pkgdashboard.NewMeta(now.In(cf.Timezone.Time()), c.OlaresID(), cf.User),
		Items: []pkgdashboard.Item{{Raw: detail, Display: gpuTaskDetailDisplayCopy(detail)}},
	}
	gaugesEnv := pkgdashboard.Envelope{
		Kind:  pkgdashboard.KindOverviewGPUGauges,
		Meta:  pkgdashboard.NewMeta(now.In(cf.Timezone.Time()), c.OlaresID(), cf.User),
		Items: gaugeItems,
	}
	trendsEnv := pkgdashboard.Envelope{
		Kind:  pkgdashboard.KindOverviewGPUTrends,
		Meta:  pkgdashboard.NewMeta(now.In(cf.Timezone.Time()), c.OlaresID(), cf.User),
		Items: trendItems,
	}
	env.Sections = map[string]pkgdashboard.Envelope{
		"detail": detailEnv,
		"gauges": gaugesEnv,
		"trends": trendsEnv,
	}
	if len(warnings) > 0 {
		env.Meta.Warnings = warnings
	}
	return env, nil
}

// ----------------------------------------------------------------------------
// Table renderer (shared by gpu detail / task-detail)
// ----------------------------------------------------------------------------

// WriteDetailFullTable emits a three-block table view:
//
//	== Detail ==        — flat key/value list (the basic info card).
//	== Gauges ==        — N rows with title / value / unit / percent / used-total.
//	== Trends ==        — for each trend, a label header + (timestamp, value)
//	                      rows truncated to --head 16 by default.
//
// The output is intentionally text-only (no ANSI), so the bash
// smoke script can grep / awk through it. Shared by RunDetail and
// RunTaskDetail.
func WriteDetailFullTable(w io.Writer, env pkgdashboard.Envelope, cf *pkgdashboard.CommonFlags) error {
	if env.Meta.Empty {
		fmt.Fprintf(w, "(empty: %s", env.Meta.EmptyReason)
		if env.Meta.Note != "" {
			fmt.Fprintf(w, "; note: %s", env.Meta.Note)
		}
		fmt.Fprintln(w, ")")
		return nil
	}
	if env.Meta.Note != "" {
		fmt.Fprintf(os.Stderr, "(advisory) %s\n", env.Meta.Note)
	}
	if env.Meta.Window != nil {
		fmt.Fprintf(w, "Window: start=%s end=%s step=%s",
			env.Meta.Window.Start, env.Meta.Window.End, env.Meta.Window.Step)
		if env.Meta.Window.Since != "" {
			fmt.Fprintf(w, " since=%s", env.Meta.Window.Since)
		}
		fmt.Fprintln(w)
	}

	// Detail section.
	fmt.Fprintln(w, "\n== Detail ==")
	if dEnv, ok := env.Sections["detail"]; ok && len(dEnv.Items) > 0 {
		writeKeyValueTable(w, dEnv.Items[0])
	} else {
		fmt.Fprintln(w, "-")
	}

	// Gauges section.
	fmt.Fprintln(w, "\n== Gauges ==")
	if gEnv, ok := env.Sections["gauges"]; ok {
		cols := []pkgdashboard.TableColumn{
			{Header: "KEY", Get: func(it pkgdashboard.Item) string { return pkgdashboard.DisplayString(it, "key") }},
			{Header: "TITLE", Get: func(it pkgdashboard.Item) string { return pkgdashboard.DisplayString(it, "title") }},
			{Header: "VALUE", Get: func(it pkgdashboard.Item) string { return pkgdashboard.DisplayString(it, "value") }},
			{Header: "UNIT", Get: func(it pkgdashboard.Item) string { return pkgdashboard.DisplayString(it, "unit") }},
			{Header: "PERCENT", Get: func(it pkgdashboard.Item) string { return pkgdashboard.DisplayString(it, "percent") }},
			{Header: "USED/TOTAL", Get: func(it pkgdashboard.Item) string { return pkgdashboard.DisplayString(it, "used_total") }},
		}
		_ = pkgdashboard.WriteTable(w, cols, gEnv.Items)
	}

	// Trends section.
	fmt.Fprintln(w, "\n== Trends ==")
	if tEnv, ok := env.Sections["trends"]; ok {
		head := cf.Head
		if head <= 0 {
			head = 16 // SPA renders ~16 buckets in the chart
		}
		for _, it := range tEnv.Items {
			title := pkgdashboard.DisplayString(it, "title")
			unit := pkgdashboard.DisplayString(it, "unit")
			fmt.Fprintf(w, "\n-- %s (%s) --\n", title, unit)
			if errStr, ok := it.Raw["error"].(string); ok && errStr != "" {
				fmt.Fprintf(w, "(error: %s)\n", errStr)
				continue
			}
			lines, _ := it.Raw["lines"].([]map[string]any)
			if len(lines) == 0 {
				fmt.Fprintln(w, "-")
				continue
			}
			for _, ln := range lines {
				label, _ := ln["label"].(string)
				fmt.Fprintf(w, "  %s:\n", label)
				points, _ := ln["points"].([]map[string]any)
				if len(points) == 0 {
					fmt.Fprintln(w, "    -")
					continue
				}
				rendered := points
				if head > 0 && head < len(points) {
					rendered = points[:head]
				}
				for _, p := range rendered {
					ts, _ := p["timestamp"].(string)
					v := p["value"]
					fmt.Fprintf(w, "    %s\t%v\n", ts, v)
				}
				if len(points) > len(rendered) {
					fmt.Fprintf(w, "    ... (%d more rows; pass --head 0 for full)\n", len(points)-len(rendered))
				}
			}
		}
	}

	if len(env.Meta.Warnings) > 0 {
		fmt.Fprintln(w, "\n== Warnings ==")
		for _, warn := range env.Meta.Warnings {
			fmt.Fprintf(w, "- %s\n", warn)
		}
	}
	return nil
}

// writeKeyValueTable renders one Item as a vertical key/value
// table. Sorted lexicographically so output is deterministic
// across runs.
func writeKeyValueTable(w io.Writer, it pkgdashboard.Item) {
	if it.Display == nil {
		fmt.Fprintln(w, "-")
		return
	}
	keys := make([]string, 0, len(it.Display))
	for k := range it.Display {
		keys = append(keys, k)
	}
	sortStrings(keys)
	for _, k := range keys {
		v := it.Display[k]
		fmt.Fprintf(w, "%s\t%v\n", k, v)
	}
}

// sortStrings is a tiny helper to avoid pulling sort into this
// file's imports just for one site.
func sortStrings(s []string) {
	for i := 1; i < len(s); i++ {
		for j := i; j > 0 && s[j-1] > s[j]; j-- {
			s[j-1], s[j] = s[j], s[j-1]
		}
	}
}
