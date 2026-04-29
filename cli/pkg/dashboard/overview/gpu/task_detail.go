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
		Start: pkgdashboard.GPUTrendTimestampISO(start),
		End:   pkgdashboard.GPUTrendTimestampISO(end),
		Step:  pkgdashboard.GPUTrendStep(start, end),
	}

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
		repl, env.Meta.Window.Start, env.Meta.Window.End, env.Meta.Window.Step,
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
