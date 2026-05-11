package dashboard

// helpers_test.go is the in-package test trampoline surface. The
// historic cmd-root dashboard_test.go was written against a flat
// namespace (every pkg type / function was duplicated as a cmd-package
// alias) and against an unqualified `common` global; that aliases
// surface lived in cmd-root aliases_test.go.
//
// When the tests moved to cli/pkg/dashboard/ during P3c, the unqualified
// type / constant / function names resolve directly (same package).
// The only things still missing are:
//
//   1. a package-level `common *CommonFlags` so test bodies can mutate
//      common.Output / common.Timezone the same way they did under
//      cmd-root cobra binding;
//   2. lowercase wrappers around the exported helpers that take an
//      explicit *CommonFlags / io.Writer pair (GateOlaresOne,
//      GPUAdvisory, VgpuUnavailableFromError, FetchClusterMetrics,
//      FetchWorkloadsMetrics, MonitoringQuery), so test bodies keep
//      reading the same way they used to.
//
// The trampolines live behind a _test.go file so they never reach the
// production binary — production callers still go through the exported
// pkg API directly.

import (
	"context"
	"net/http"
	"net/url"
	"os"
	"time"

	"github.com/beclab/Olares/cli/pkg/dashboard/format"
)

// common is the shared *CommonFlags pointer test bodies mutate. Most
// fetcher tests only care about Output / Timezone; CommonFlags{}
// zero-value + LocalLocation() default mirrors the production cobra
// PersistentFlags binding ("output=table", "temp-unit=C", local TZ).
var common = &CommonFlags{Timezone: format.LocalLocation()}

func defaultClusterWindow() MonitoringWindow { return DefaultClusterWindow() }
func defaultDetailWindow() MonitoringWindow  { return DefaultDetailWindow() }

func monitoringQuery(metricsFilter []string, w MonitoringWindow, now time.Time, instant bool) url.Values {
	return MonitoringQuery(common, metricsFilter, w, now, instant)
}

func fetchClusterMetrics(ctx context.Context, c *Client, metrics []string, w MonitoringWindow, now time.Time, instant bool) (map[string]format.MonitoringResult, error) {
	return FetchClusterMetrics(ctx, c, common, metrics, w, now, instant)
}

func fetchWorkloadsMetrics(ctx context.Context, c *Client, req WorkloadRequest, w MonitoringWindow, now time.Time) ([]WorkloadAggregate, error) {
	return FetchWorkloadsMetrics(ctx, c, common, req, w, now)
}

func gateOlaresOne(ctx context.Context, c *Client, kind string, now time.Time) (Envelope, bool) {
	return GateOlaresOne(ctx, c, common, kind, now, os.Stderr)
}

func gpuAdvisory(ctx context.Context, c *Client) (note, reason string) {
	return GPUAdvisory(ctx, c, common, os.Stderr)
}

func vgpuUnavailableFromError(c *Client, err error, kind string, now time.Time) (Envelope, bool) {
	return VgpuUnavailableFromError(c, common, err, kind, now, os.Stderr)
}

// renderDiskTemperature was previously aliased here for the
// pkg-root TestRenderDiskTemperature test; in P7 the test moved to
// cli/pkg/dashboard/overview/disk/main_test.go (next to the actual
// production helper), so the shim no longer has any consumer.

// firstAnyInArray / toFloat / renderTemperature are exported under
// PascalCase in pkg; the test file references the lowercase shape that
// cmd-root aliases_test.go used. Define thin shims so the existing
// test bodies keep compiling. (No production caller goes through these.)
func firstAnyInArray(v any) any                { return FirstAnyInArray(v) }
func toFloat(v any) float64                    { return ToFloat(v) }
func percentString(r float64) string           { return PercentString(r) }
func percentDirect(p float64) string           { return PercentDirect(p) }
func gpuModeLabel(raw any) string              { return GPUModeLabel(raw) }
func gpuHealthLabel(raw any) string            { return GPUHealthLabel(raw) }
func gpuVRAMHuman(mibVal any) string           { return GPUVRAMHuman(mibVal) }
func gpuTrendStep(start, end time.Time) string { return GPUTrendStep(start, end) }
func renderTemperature(c float64, target format.TempUnit) string {
	return RenderTemperature(c, target)
}

func extractHAMIMessage(body string) string { return ExtractHAMIMessage(body) }

func fetchAppsList(ctx context.Context, c *Client) ([]RawAppListItem, error) {
	return FetchAppsList(ctx, c)
}

func mergeWorkloadMetrics(apps []WorkloadApp, podData, nsData map[string]format.MonitoringResult) []WorkloadAggregate {
	return MergeWorkloadMetrics(apps, podData, nsData)
}

func podDeploymentName(pod string) string { return PodDeploymentName(pod) }

func hasPknameLabels(rows []LsblkRow) bool { return HasPknameLabels(rows) }

func collectSubtreeByPkname(allRows []LsblkRow, rootName string) []LsblkRow {
	return CollectSubtreeByPkname(allRows, rootName)
}

func resolveParent(r LsblkRow, rootName string, nameSet map[string]bool) string {
	return ResolveParent(r, rootName, nameSet)
}

func buildLsblkTreePrefix(depth int, lastStack []bool) string {
	return BuildLsblkTreePrefix(depth, lastStack)
}

func flattenLsblkHierarchy(rows []LsblkRow, rootName string) []LsblkFlatRow {
	return FlattenLsblkHierarchy(rows, rootName)
}

func hasCUDANode(ctx context.Context, c *Client) (bool, error) {
	return HasCUDANode(ctx, c)
}

func fetchGraphicsList(ctx context.Context, c *Client, filters map[string]string) ([]map[string]any, error) {
	return FetchGraphicsList(ctx, c, filters)
}

func fetchTaskList(ctx context.Context, c *Client, filters map[string]string) ([]map[string]any, error) {
	return FetchTaskList(ctx, c, filters)
}

func fetchGraphicsDetail(ctx context.Context, c *Client, uuid string) (map[string]any, error) {
	return FetchGraphicsDetail(ctx, c, uuid)
}

func fetchTaskDetail(ctx context.Context, c *Client, name, podUID, sharemode string) (map[string]any, error) {
	return FetchTaskDetail(ctx, c, name, podUID, sharemode)
}

func fetchInstantVector(ctx context.Context, c *Client, query string) ([]InstantVectorSample, error) {
	return FetchInstantVector(ctx, c, query)
}

func fetchRangeVector(ctx context.Context, c *Client, query, start, end, step string) ([]RangeVectorSeries, error) {
	return FetchRangeVector(ctx, c, query, start, end, step)
}

// newTestClient is provided by dashboard_test.go (kept there so the
// httptest fixture lives next to the bodies that exercise it).

// Lowercase shape aliases the historic cmd-side test bodies used.
type (
	monitoringWindow    = MonitoringWindow
	workloadAggregate   = WorkloadAggregate
	workloadRequest     = WorkloadRequest
	workloadApp         = WorkloadApp
	rawAppListItem      = RawAppListItem
	lsblkRow            = LsblkRow
	lsblkFlatRow        = LsblkFlatRow
	graphicsListBody    = GraphicsListBody
	instantVectorSample = InstantVectorSample
	rangeVectorSeries   = RangeVectorSeries
	rangeVectorPoint    = RangeVectorPoint
	rangeVectorRange    = RangeVectorRange
)

// _ silences unused-import warnings if a refactor temporarily drops a
// fixture; cheaper than threading explicit ignores per import.
var _ = http.MethodGet
