package gpu

import (
	"context"
	"net/http"
	"os"
	"strings"
	"sync"
	"time"

	pkgdashboard "github.com/beclab/Olares/cli/pkg/dashboard"
)

// RunDetail is the cmd-side entry point for `dashboard overview
// gpu detail <uuid>`. Watch-aware (Runner with 30s recommended
// cadence); each tick re-resolves the time window so a long
// `--watch` stream slides the trend chart with wall-clock.
func RunDetail(ctx context.Context, c *pkgdashboard.Client, cf *pkgdashboard.CommonFlags, uuid string) error {
	r := &pkgdashboard.Runner{
		Flags:       cf,
		Recommended: 30 * time.Second,
		Iter: func(ctx context.Context, iter int, now time.Time) (pkgdashboard.Envelope, error) {
			start, end, since := ResolveDetailWindow(cf, now, GPUDetailDefaultSince)
			env, err := BuildDetailFullEnvelope(ctx, c, cf, uuid, start, end, since)
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

// BuildDetailFullEnvelope produces the full GPU detail sections
// envelope for `dashboard overview gpu detail <uuid>`.
//
// Flow:
//
//  1. Sequentially call HAMI /v1/gpu (basic info). 404/5xx
//     short-circuit to the standard `no_vgpu_integration` /
//     `vgpu_unavailable` empty envelopes via VgpuUnavailableFromError.
//  2. Concurrently fan out the gauge + trend queries with
//     errgroup. A per-query failure is captured into the gauge /
//     trend item itself (raw.error / display.error) and added to
//     env.Meta.Warnings; it does NOT abort sibling queries.
//  3. Extract device_no / driver_version from the *power* range
//     query's metric labels (the SPA does the same) and merge them
//     into the detail item so the renderer can show the SPA's
//     "Driver version: 590.44.01" cell.
func BuildDetailFullEnvelope(ctx context.Context, c *pkgdashboard.Client, cf *pkgdashboard.CommonFlags, uuid string, start, end time.Time, since time.Duration) (pkgdashboard.Envelope, error) {
	now := end
	advisoryNote, _ := pkgdashboard.GPUAdvisory(ctx, c, cf, os.Stderr)
	env := pkgdashboard.Envelope{
		Kind: pkgdashboard.KindOverviewGPUDetailFull,
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

	// Step 1 — flat detail.
	detail, err := pkgdashboard.FetchGraphicsDetail(ctx, c, uuid)
	if err != nil {
		if he, ok := pkgdashboard.IsHTTPError(err); ok && he.Status == http.StatusNotFound {
			env.Meta.Empty = true
			env.Meta.EmptyReason = "no_vgpu_integration"
			env.Meta.HTTPStatus = he.Status
			return env, nil
		}
		if unavail, ok := pkgdashboard.VgpuUnavailableFromError(c, cf, err, pkgdashboard.KindOverviewGPUDetailFull, now, os.Stderr); ok {
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

	// Step 2/3 — fan-out queries.
	gaugeSpecs := gpuDetailGaugeSpecs()
	trendSpecs := gpuDetailTrendSpecs()
	deviceuuidReplacer := strings.NewReplacer("$deviceuuid", uuid)

	var (
		warnMu   sync.Mutex
		warnings []string
	)
	addWarning := func(msg string) {
		warnMu.Lock()
		warnings = append(warnings, msg)
		warnMu.Unlock()
	}
	gaugeItems, trendItems, labelsFromPw := fanoutGaugeAndTrend(
		ctx, c,
		gaugeSpecs, trendSpecs,
		deviceuuidReplacer, env.Meta.Window.Start, env.Meta.Window.End, env.Meta.Window.Step,
		addWarning,
		"power_trend",
	)

	// Merge device_no / driver_version into detail (SPA's `detail2`).
	for k, v := range labelsFromPw {
		if k == "device_no" || k == "driver_version" {
			if _, present := detail[k]; !present {
				detail[k] = v
			}
		}
	}

	// Compose sections.
	detailEnv := pkgdashboard.Envelope{
		Kind:  pkgdashboard.KindOverviewGPUDetail,
		Meta:  pkgdashboard.NewMeta(now.In(cf.Timezone.Time()), c.OlaresID(), cf.User),
		Items: []pkgdashboard.Item{{Raw: detail, Display: gpuDetailDisplayCopy(detail)}},
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
