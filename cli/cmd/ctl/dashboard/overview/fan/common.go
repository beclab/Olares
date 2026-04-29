// Package overview hosts the cobra subtree for `olares-cli dashboard
// overview`. Per the settings-style layout (cli/cmd/ctl/settings/<area>/),
// every cobra leaf lives in its own .go file and the directory tree
// mirrors the command tree:
//
//	overview/                       (this package — root.go + 7 leaves)
//	  ├── disk/                     (subgroup; own Go package)
//	  ├── fan/
//	  └── gpu/
//
// This file is the area's deliberate "light duplication" surface: it
// holds the type aliases, kind constants, and trampolines over
// `cli/pkg/dashboard` that every leaf in the area uses. The pkg layer is
// the source of truth; common.go just gives the leaf code its expected
// unqualified names so call sites stay close to the pre-refactor shape.
//
// `var common` is a *pkgdashboard.CommonFlags pointer set by
// NewFanCommand at construction time; cobra's persistent-flag
// inheritance from the dashboard root populates the pointed-at struct
// before any leaf RunE fires.
package fan

import (
	"context"
	"fmt"
	"net/url"
	"os"
	"strings"
	"time"

	"github.com/spf13/cobra"

	"github.com/beclab/Olares/cli/pkg/cmdutil"
	pkgdashboard "github.com/beclab/Olares/cli/pkg/dashboard"
	"github.com/beclab/Olares/cli/pkg/dashboard/format"
)

// ----------------------------------------------------------------------------
// Type aliases
// ----------------------------------------------------------------------------

type (
	Envelope     = pkgdashboard.Envelope
	Item         = pkgdashboard.Item
	Meta         = pkgdashboard.Meta
	TimeWindow   = pkgdashboard.TimeWindow
	Client       = pkgdashboard.Client
	Runner       = pkgdashboard.Runner
	OutputFormat = pkgdashboard.OutputFormat
	TableColumn  = pkgdashboard.TableColumn
	HTTPError    = pkgdashboard.HTTPError
	UserDetail   = pkgdashboard.UserDetail
	SystemStatus = pkgdashboard.SystemStatus

	monitoringWindow    = pkgdashboard.MonitoringWindow
	workloadAggregate   = pkgdashboard.WorkloadAggregate
	workloadRequest     = pkgdashboard.WorkloadRequest
	workloadApp         = pkgdashboard.WorkloadApp
	SystemIFSItem       = pkgdashboard.SystemIFSItem
	SystemFanData       = pkgdashboard.SystemFanData
	fanCurveRow         = pkgdashboard.FanCurveRow
	lsblkRow            = pkgdashboard.LsblkRow
	lsblkFlatRow        = pkgdashboard.LsblkFlatRow
	rawAppListItem      = pkgdashboard.RawAppListItem
	instantVectorSample = pkgdashboard.InstantVectorSample
	rangeVectorSeries   = pkgdashboard.RangeVectorSeries
	rangeVectorPoint    = pkgdashboard.RangeVectorPoint
	rangeVectorRange    = pkgdashboard.RangeVectorRange
)

// ----------------------------------------------------------------------------
// Constants
// ----------------------------------------------------------------------------

const (
	OutputTable = pkgdashboard.OutputTable
	OutputJSON  = pkgdashboard.OutputJSON

	KindOverview               = pkgdashboard.KindOverview
	KindOverviewPhysical       = pkgdashboard.KindOverviewPhysical
	KindOverviewUser           = pkgdashboard.KindOverviewUser
	KindOverviewRanking        = pkgdashboard.KindOverviewRanking
	KindOverviewCPU            = pkgdashboard.KindOverviewCPU
	KindOverviewMemory         = pkgdashboard.KindOverviewMemory
	KindOverviewDisk           = pkgdashboard.KindOverviewDisk
	KindOverviewDiskMain       = pkgdashboard.KindOverviewDiskMain
	KindOverviewDiskPart       = pkgdashboard.KindOverviewDiskPart
	KindOverviewPods           = pkgdashboard.KindOverviewPods
	KindOverviewNetwork        = pkgdashboard.KindOverviewNetwork
	KindOverviewFan            = pkgdashboard.KindOverviewFan
	KindOverviewFanLive        = pkgdashboard.KindOverviewFanLive
	KindOverviewFanCurve       = pkgdashboard.KindOverviewFanCurve
	KindOverviewGPUList        = pkgdashboard.KindOverviewGPUList
	KindOverviewGPUTasks       = pkgdashboard.KindOverviewGPUTasks
	KindOverviewGPUDetail      = pkgdashboard.KindOverviewGPUDetail
	KindOverviewGPUTaskDet     = pkgdashboard.KindOverviewGPUTaskDet
	KindOverviewGPUDetailFull  = pkgdashboard.KindOverviewGPUDetailFull
	KindOverviewGPUTaskDetFull = pkgdashboard.KindOverviewGPUTaskDetFull
	KindOverviewGPUGauges      = pkgdashboard.KindOverviewGPUGauges
	KindOverviewGPUTrends      = pkgdashboard.KindOverviewGPUTrends

	systemFrontendDeployment = pkgdashboard.SystemFrontendDeployment
	fanSpeedMaxCPU           = pkgdashboard.FanSpeedMaxCPU
	fanSpeedMaxGPU           = pkgdashboard.FanSpeedMaxGPU
)

// fanCurveTable mirrors pkgdashboard.FanCurveTable so the fan-related
// leaves keep their tabletop alias.
var fanCurveTable = pkgdashboard.FanCurveTable

// ----------------------------------------------------------------------------
// Pkg-side function bindings
// ----------------------------------------------------------------------------

var (
	NewMeta              = pkgdashboard.NewMeta
	HeadItems            = pkgdashboard.HeadItems
	WriteJSON            = pkgdashboard.WriteJSON
	WriteTable           = pkgdashboard.WriteTable
	DisplayString        = pkgdashboard.DisplayString
	EmitDefault          = pkgdashboard.EmitDefault
	ClassifyTransportErr = pkgdashboard.ClassifyTransportErr
	IsHTTPError          = pkgdashboard.IsHTTPError
	NewClient            = pkgdashboard.NewClient
	ParseStep            = pkgdashboard.ParseStep

	// Numeric formatting (lifted to pkg so cmd subpackages don't dup).
	FormatFloat       = pkgdashboard.FormatFloat
	SafeRatio         = pkgdashboard.SafeRatio
	FormatRateAny     = pkgdashboard.FormatRateAny
	ParseRFCTimestamp = pkgdashboard.ParseRFCTimestamp
	SampleFloat       = pkgdashboard.SampleFloat
)

// Lower-case aliases for legacy leaf code (settings precedent: light
// duplication acceptable; flips happen exclusively at the binding).
var (
	formatFloat       = FormatFloat
	safeRatio         = SafeRatio
	formatRateAny     = FormatRateAny
	sampleFloat       = SampleFloat
	lastSampleFromRow = pkgdashboard.LastSampleFromRow
)

// ----------------------------------------------------------------------------
// Area state — common is the shared CommonFlags pointer for all leaves
// ----------------------------------------------------------------------------

// common is wired by NewFanCommand(f, cf) at construction; reads
// flow through cobra's persistent-flag inheritance which mutates the
// pointed-at struct before any leaf RunE runs.
var common *pkgdashboard.CommonFlags

// ----------------------------------------------------------------------------
// prepareClient — area-private factory
// ----------------------------------------------------------------------------
//
// Each area writes its own prepareClient (settings precedent) so the
// transport seam between cmdutil.Factory (interactive credential
// resolution) and pkgdashboard.Client (pure HTTP) is local and
// inspectable per area.

func prepareClient(ctx context.Context, f *cmdutil.Factory) (*Client, error) {
	if f == nil {
		return nil, fmt.Errorf("internal error: fan not wired with cmdutil.Factory")
	}
	rp, err := f.ResolveProfile(ctx)
	if err != nil {
		return nil, err
	}
	hc, err := f.HTTPClient(ctx)
	if err != nil {
		return nil, err
	}
	return pkgdashboard.NewClient(hc, rp), nil
}

// buildDashboardClient is the legacy alias used throughout the leaf code;
// kept identical to the parent-package signature so leaf RunE bodies
// don't need rewriting.
func buildDashboardClient(ctx context.Context, f *cmdutil.Factory) (*Client, error) {
	return prepareClient(ctx, f)
}

// ----------------------------------------------------------------------------
// Monitoring trampolines (close over `common`)
// ----------------------------------------------------------------------------

func defaultClusterWindow() monitoringWindow { return pkgdashboard.DefaultClusterWindow() }
func defaultDetailWindow() monitoringWindow  { return pkgdashboard.DefaultDetailWindow() }

func monitoringQuery(metricsFilter []string, w monitoringWindow, now time.Time, instant bool) url.Values {
	return pkgdashboard.MonitoringQuery(common, metricsFilter, w, now, instant)
}

func fetchClusterMetrics(ctx context.Context, c *Client, metrics []string, w monitoringWindow, now time.Time, instant bool) (map[string]format.MonitoringResult, error) {
	return pkgdashboard.FetchClusterMetrics(ctx, c, common, metrics, w, now, instant)
}

func fetchNodeMetrics(ctx context.Context, c *Client, metrics []string, w monitoringWindow, now time.Time, instant bool) (map[string]format.MonitoringResult, error) {
	return pkgdashboard.FetchNodeMetrics(ctx, c, common, metrics, w, now, instant)
}

func fetchUserMetric(ctx context.Context, c *Client, username string, metrics []string, w monitoringWindow, now time.Time, instant bool) (map[string]format.MonitoringResult, error) {
	return pkgdashboard.FetchUserMetric(ctx, c, common, username, metrics, w, now, instant)
}

func doMonitoring(ctx context.Context, c *Client, path string, q url.Values) (map[string]format.MonitoringResult, error) {
	return pkgdashboard.DoMonitoring(ctx, c, path, q)
}

func fetchNodesList(ctx context.Context, c *Client) ([]string, error) {
	return pkgdashboard.FetchNodesList(ctx, c)
}

// ----------------------------------------------------------------------------
// Capability gates
// ----------------------------------------------------------------------------

func hasCUDANode(ctx context.Context, c *Client) (bool, error) {
	return pkgdashboard.HasCUDANode(ctx, c)
}

func gateOlaresOne(ctx context.Context, c *Client, kind string, now time.Time) (Envelope, bool) {
	return pkgdashboard.GateOlaresOne(ctx, c, common, kind, now, os.Stderr)
}

func gpuAdvisory(ctx context.Context, c *Client) (note, reason string) {
	return pkgdashboard.GPUAdvisory(ctx, c, common, os.Stderr)
}

func vgpuUnavailableFromError(c *Client, err error, kind string, now time.Time) (Envelope, bool) {
	return pkgdashboard.VgpuUnavailableFromError(c, common, err, kind, now, os.Stderr)
}

func extractHAMIMessage(body string) string { return pkgdashboard.ExtractHAMIMessage(body) }
func displayRole(r string) string           { return pkgdashboard.DisplayRole(r) }

// ----------------------------------------------------------------------------
// Workload aggregation
// ----------------------------------------------------------------------------

func podDeploymentName(pod string) string { return pkgdashboard.PodDeploymentName(pod) }

func fetchWorkloadsMetrics(ctx context.Context, c *Client, req workloadRequest, w monitoringWindow, now time.Time) ([]workloadAggregate, error) {
	return pkgdashboard.FetchWorkloadsMetrics(ctx, c, common, req, w, now)
}

func buildSystemDeploymentFilter(systemApps []workloadApp) string {
	return pkgdashboard.BuildSystemDeploymentFilter(systemApps)
}

func buildCustomNamespaceFilter(customApps []workloadApp) string {
	return pkgdashboard.BuildCustomNamespaceFilter(customApps)
}

func mergeWorkloadMetrics(apps []workloadApp, podData, nsData map[string]format.MonitoringResult) []workloadAggregate {
	return pkgdashboard.MergeWorkloadMetrics(apps, podData, nsData)
}

func aggregateByDeployment(data map[string]format.MonitoringResult, metric string) map[string]float64 {
	return pkgdashboard.AggregateByDeployment(data, metric)
}

func podCountByDeployment(data map[string]format.MonitoringResult) map[string]int {
	return pkgdashboard.PodCountByDeployment(data)
}

func aggregateByNamespace(data map[string]format.MonitoringResult, metric string) map[string]float64 {
	return pkgdashboard.AggregateByNamespace(data, metric)
}

func lastValueOfRow(value []interface{}, values [][]interface{}) (float64, bool) {
	return pkgdashboard.LastValueOfRow(value, values)
}

func scalarFloat(v interface{}) (float64, bool) {
	return pkgdashboard.ScalarFloat(v)
}

func sortWorkloadAggregates(rows []workloadAggregate, sortBy, dir string) {
	pkgdashboard.SortWorkloadAggregates(rows, sortBy, dir)
}

// ----------------------------------------------------------------------------
// System / lsblk / apps
// ----------------------------------------------------------------------------

func fetchSystemIFS(ctx context.Context, c *Client, testConnectivity bool) ([]SystemIFSItem, error) {
	return pkgdashboard.FetchSystemIFS(ctx, c, testConnectivity)
}

func fetchSystemFan(ctx context.Context, c *Client) (*SystemFanData, error) {
	return pkgdashboard.FetchSystemFan(ctx, c)
}

func hasPknameLabels(rows []lsblkRow) bool { return pkgdashboard.HasPknameLabels(rows) }

func collectSubtreeByPkname(allRows []lsblkRow, rootName string) []lsblkRow {
	return pkgdashboard.CollectSubtreeByPkname(allRows, rootName)
}

func resolveParent(r lsblkRow, rootName string, nameSet map[string]bool) string {
	return pkgdashboard.ResolveParent(r, rootName, nameSet)
}

func buildLsblkTreePrefix(depth int, lastStack []bool) string {
	return pkgdashboard.BuildLsblkTreePrefix(depth, lastStack)
}

func flattenLsblkHierarchy(rows []lsblkRow, rootName string) []lsblkFlatRow {
	return pkgdashboard.FlattenLsblkHierarchy(rows, rootName)
}

func fetchAppsList(ctx context.Context, c *Client) ([]rawAppListItem, error) {
	return pkgdashboard.FetchAppsList(ctx, c)
}

// ----------------------------------------------------------------------------
// GPU (HAMI)
// ----------------------------------------------------------------------------

func fetchGraphicsList(ctx context.Context, c *Client, filters map[string]string) ([]map[string]any, error) {
	return pkgdashboard.FetchGraphicsList(ctx, c, filters)
}

func fetchTaskList(ctx context.Context, c *Client, filters map[string]string) ([]map[string]any, error) {
	return pkgdashboard.FetchTaskList(ctx, c, filters)
}

func fetchGraphicsDetail(ctx context.Context, c *Client, uuid string) (map[string]any, error) {
	return pkgdashboard.FetchGraphicsDetail(ctx, c, uuid)
}

func fetchTaskDetail(ctx context.Context, c *Client, name, podUID, sharemode string) (map[string]any, error) {
	return pkgdashboard.FetchTaskDetail(ctx, c, name, podUID, sharemode)
}

func fetchInstantVector(ctx context.Context, c *Client, query string) ([]instantVectorSample, error) {
	return pkgdashboard.FetchInstantVector(ctx, c, query)
}

func fetchRangeVector(ctx context.Context, c *Client, query, start, end, step string) ([]rangeVectorSeries, error) {
	return pkgdashboard.FetchRangeVector(ctx, c, query, start, end, step)
}

func gpuTrendStep(start, end time.Time) string { return pkgdashboard.GPUTrendStep(start, end) }
func gpuStepPreset(minutes int) (string, bool) { return pkgdashboard.GPUStepPreset(minutes) }
func gpuTrendTimestampISO(t time.Time) string  { return pkgdashboard.GPUTrendTimestampISO(t) }

func classifyKind(status int) string { return pkgdashboard.ClassifyKind(status) }

func renderTemperature(celsius float64, target format.TempUnit) string {
	return pkgdashboard.RenderTemperature(celsius, target)
}

func percentString(ratio float64) string { return pkgdashboard.PercentString(ratio) }
func percentDirect(pct float64) string   { return pkgdashboard.PercentDirect(pct) }
func gpuModeLabel(raw any) string        { return pkgdashboard.GPUModeLabel(raw) }
func gpuHealthLabel(raw any) string      { return pkgdashboard.GPUHealthLabel(raw) }
func firstAnyInArray(v any) any          { return pkgdashboard.FirstAnyInArray(v) }
func gpuVRAMHuman(mibVal any) string     { return pkgdashboard.GPUVRAMHuman(mibVal) }

// ----------------------------------------------------------------------------
// User
// ----------------------------------------------------------------------------

func resolveTargetUser(ctx context.Context, c *Client, requested string) (*UserDetail, error) {
	return pkgdashboard.ResolveTargetUser(ctx, c, requested)
}

func toFloat(v any) float64 { return pkgdashboard.ToFloat(v) }

// ----------------------------------------------------------------------------
// Ranking — the one legitimate horizontal share between overview + apps
// ----------------------------------------------------------------------------

// buildRankingEnvelopeBy is the area-local trampoline over
// pkgdashboard.BuildRankingEnvelope. Both overview/ranking.go and
// applications/root.go go through this name.
func buildRankingEnvelopeBy(ctx context.Context, c *Client, target, sortBy, sortDir string, now time.Time) (Envelope, error) {
	return pkgdashboard.BuildRankingEnvelope(ctx, c, common, target, sortBy, sortDir, now)
}

// unknownSubcommandRunE mirrors the cmd-root helper of the same name —
// duplicated per area so each parent dispatch behaves identically when
// the user mistypes a verb (e.g. `dashboard overview pods` is fine, but
// `dashboard overview podz` should hint instead of silently rendering
// help with exit 0). Returns pkgdashboard.ErrAlreadyReported so the
// dashboard root's leaf-error wrapper skips redundant Fprintln while
// still propagating the error up for cobra's non-zero exit.
func unknownSubcommandRunE(c *cobra.Command, args []string) error {
	if len(args) == 0 {
		return c.Help()
	}
	msg := fmt.Sprintf("Error: unknown subcommand %q for %q", args[0], c.CommandPath())
	if suggestions := c.SuggestionsFor(args[0]); len(suggestions) > 0 {
		msg += "\n\nDid you mean this?\n\t" + strings.Join(suggestions, "\n\t")
	}
	fmt.Fprintln(c.ErrOrStderr(), msg)
	fmt.Fprintf(c.ErrOrStderr(), "\nRun '%s --help' for usage.\n", c.CommandPath())
	return pkgdashboard.ErrAlreadyReported
}
