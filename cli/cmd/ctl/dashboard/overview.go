package dashboard

import (
	"context"
	"fmt"
	"net/http"
	"os"
	"sort"
	"strconv"
	"strings"
	"time"

	"github.com/spf13/cobra"

	"github.com/beclab/Olares/cli/cmd/ctl/dashboard/format"
	"github.com/beclab/Olares/cli/pkg/cmdutil"
)

// ----------------------------------------------------------------------------
// `dashboard overview` — command tree assembly
// ----------------------------------------------------------------------------
//
// Default action emits a sections envelope mirroring the SPA's overview
// page: { physical, user, ranking }. Each leaf below produces the same
// envelope shape it would emit standalone, so consumers can demux on
// `meta.kind` per section.
//
// Endpoint mapping (one helper per section + per leaf — see helpers.go):
//
//	overview (default)            — fan-out: physical + user + ranking
//	overview physical             — GET  /kapis/.../v1alpha3/cluster
//	overview user [<username>]    — GET  /kapis/.../v1alpha3/users/<u>
//	overview ranking              — workload-grain (fetchWorkloadsMetrics)
//	overview cpu                  — GET  /kapis/.../v1alpha3/nodes  (per-node multi-metric)
//	overview memory               — GET  /kapis/.../v1alpha3/nodes  (per-node, --mode physical|swap)
//	overview disk                 — sections: main + per-disk partitions
//	overview disk main            — GET  /kapis/.../v1alpha3/nodes  (per-disk metric)
//	overview disk partitions <d>  — GET  /kapis/.../v1alpha3/nodes  (per-partition metric)
//	overview pods                 — GET  /kapis/.../v1alpha3/nodes  (per-node count)
//	overview network              — GET  /capi/system/ifs           (per-iface system-ifs)
//	overview fan                  — sections: live + curve
//	overview fan live             — GET  /user-service/api/mdns/olares-one/cpu-gpu + graphics list
//	overview fan curve            — hardcoded fanCurveTable (helpers.go)
//	overview gpu list             — POST /hami/api/vgpu/v1/gpus
//	overview gpu tasks            — POST /hami/api/vgpu/v1/containers
//	overview gpu get <uuid>       — GET  /hami/api/vgpu/v1/gpu?uuid=...
//	overview gpu task <name> <uid>— GET  /hami/api/vgpu/v1/container?name=&podUid=
//
// Every leaf consumes CommonFlags (--output / --watch / --since / etc.)
// and returns one Envelope. Watch-able leaves wrap their fetch in Runner.

func newOverviewCommand(f *cmdutil.Factory) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "overview",
		Short: "Sections envelope mirroring the SPA's overview page (physical / user / ranking)",
		Example: `  # Default — emit the three sections in parallel as a single envelope:
  olares-cli dashboard overview -o json

  # Just the workload-grain ranking:
  olares-cli dashboard overview ranking --sort desc`,
		SilenceErrors: true,
		SilenceUsage:  true,
		RunE: func(c *cobra.Command, args []string) error {
			if len(args) > 0 {
				return unknownSubcommandRunE(c, args)
			}
			if err := common.Validate(); err != nil {
				return err
			}
			return runOverviewDefault(c.Context(), f)
		},
	}

	cmd.AddCommand(newOverviewPhysicalCommand(f))
	cmd.AddCommand(newOverviewUserCommand(f))
	cmd.AddCommand(newOverviewRankingCommand(f))
	cmd.AddCommand(newOverviewCPUCommand(f))
	cmd.AddCommand(newOverviewMemoryCommand(f))
	cmd.AddCommand(newOverviewDiskCommand(f))
	cmd.AddCommand(newOverviewPodsCommand(f))
	cmd.AddCommand(newOverviewNetworkCommand(f))
	cmd.AddCommand(newOverviewFanCommand(f))
	cmd.AddCommand(newOverviewGPUCommand(f))
	return cmd
}

// runOverviewDefault is the aggregate action. Fans out the three SECTIONS
// in parallel; per-section failures populate Meta.Error on that section
// without aborting the whole envelope. Mirrors the SPA's "partial degradation
// is fine, surface it" behaviour on the overview page.
func runOverviewDefault(ctx context.Context, f *cmdutil.Factory) error {
	c, err := buildDashboardClient(ctx, f)
	if err != nil {
		return err
	}
	now := time.Now()
	type sectionResult struct {
		key string
		env Envelope
	}
	results := make(chan sectionResult, 3)

	go func() {
		env, err := buildPhysicalEnvelope(ctx, c, now)
		if err != nil {
			env.Kind = KindOverviewPhysical
			env.Meta.Error = err.Error()
			env.Meta.ErrorKind = ClassifyTransportErr(err)
		}
		results <- sectionResult{"physical", env}
	}()

	go func() {
		env, err := buildUserEnvelope(ctx, c, common.User, now)
		if err != nil {
			env.Kind = KindOverviewUser
			env.Meta.Error = err.Error()
			env.Meta.ErrorKind = ClassifyTransportErr(err)
		}
		results <- sectionResult{"user", env}
	}()

	go func() {
		env, err := buildRankingEnvelope(ctx, c, common.User, "desc", now)
		if err != nil {
			env.Kind = KindOverviewRanking
			env.Meta.Error = err.Error()
			env.Meta.ErrorKind = ClassifyTransportErr(err)
		}
		results <- sectionResult{"ranking", env}
	}()

	out := map[string]Envelope{}
	for i := 0; i < 3; i++ {
		r := <-results
		r.env.Meta.FetchedAt = time.Now().In(common.Timezone.Time()).Format(time.RFC3339)
		out[r.key] = r.env
	}

	env := Envelope{
		Kind:     KindOverview,
		Meta:     NewMeta(time.Now().In(common.Timezone.Time()), c.OlaresID(), common.User),
		Sections: out,
	}

	if common.Output == OutputJSON {
		return WriteJSON(os.Stdout, env)
	}
	// Table mode: render each section's table back-to-back, separated
	// by section banners. Lets a human eyeball the three views in one
	// shot.
	return writeOverviewSectionsTable(env)
}

// writeOverviewSectionsTable lays out the three sections back-to-back in
// table mode. Section banners use a leading "==" so a human scanning the
// scrollback can locate them by simple pattern match.
func writeOverviewSectionsTable(env Envelope) error {
	for _, key := range []string{"physical", "user", "ranking"} {
		section, ok := env.Sections[key]
		if !ok {
			continue
		}
		fmt.Fprintf(os.Stdout, "== %s ==\n", strings.ToUpper(key))
		if section.Meta.Error != "" {
			fmt.Fprintf(os.Stdout, "(error: %s)\n\n", section.Meta.Error)
			continue
		}
		switch section.Kind {
		case KindOverviewPhysical:
			if err := writePhysicalTable(section); err != nil {
				return err
			}
		case KindOverviewUser:
			if err := writeUserTable(section); err != nil {
				return err
			}
		case KindOverviewRanking:
			if err := writeRankingTable(section); err != nil {
				return err
			}
		}
		fmt.Fprintln(os.Stdout)
	}
	return nil
}

// ----------------------------------------------------------------------------
// overview physical — 9-row cluster metric table
// ----------------------------------------------------------------------------

// physicalMetric is one row of the SPA's Physical Resources panel. Columns:
// metric / value / unit / utilisation / detail. Names mirror the SPA's
// rendering conventions.
type physicalMetric struct {
	Key         string  // canonical metric key (cpu / memory / disk / pods / net_in / net_out)
	Label       string  // human-friendly metric name shown in column 1
	Value       float64 // headline numeric value (used / running)
	Total       float64 // total / quota
	Unit        string  // SPA unit suffix
	Utilisation float64 // 0..1 ratio
	Detail      string  // free-form detail string (used by net rows)
}

func newOverviewPhysicalCommand(f *cmdutil.Factory) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "physical",
		Short: "9-row cluster-level resource snapshot (CPU/Memory/Disk/Pods/Net + extras)",
		Example: `  olares-cli dashboard overview physical -o json`,
		SilenceErrors: true,
		SilenceUsage:  true,
		RunE: func(c *cobra.Command, args []string) error {
			if err := common.Validate(); err != nil {
				return err
			}
			return runOverviewPhysical(c.Context(), f)
		},
	}
	return cmd
}

func runOverviewPhysical(ctx context.Context, f *cmdutil.Factory) error {
	c, err := buildDashboardClient(ctx, f)
	if err != nil {
		return err
	}
	r := &Runner{
		Flags:       &common,
		Recommended: 60 * time.Second,
		Iter: func(ctx context.Context, iter int, now time.Time) (Envelope, error) {
			env, err := buildPhysicalEnvelope(ctx, c, now)
			if err != nil {
				return env, err
			}
			env.Meta = NewMeta(now.In(common.Timezone.Time()), c.OlaresID(), common.User)
			env.Meta.RecommendedPollSeconds = 60
			if common.Output == OutputJSON {
				return env, nil
			}
			return env, writePhysicalTable(env)
		},
	}
	return r.Run(ctx)
}

// buildPhysicalEnvelope is the shared fetcher used by both `overview
// physical` (standalone) and the `overview` default sections envelope.
func buildPhysicalEnvelope(ctx context.Context, c *Client, now time.Time) (Envelope, error) {
	metrics := []string{
		"cluster_cpu_usage", "cluster_cpu_total", "cluster_cpu_utilisation",
		"cluster_memory_usage_wo_cache", "cluster_memory_total", "cluster_memory_utilisation",
		"cluster_disk_size_usage", "cluster_disk_size_capacity", "cluster_disk_size_utilisation",
		"cluster_pod_running_count", "cluster_pod_quota",
		"cluster_net_bytes_received", "cluster_net_bytes_transmitted",
	}
	res, err := fetchClusterMetrics(ctx, c, metrics, defaultClusterWindow(), now, false)
	if err != nil {
		return Envelope{Kind: KindOverviewPhysical}, err
	}
	last := format.GetLastMonitoringData(res, 0)
	rows := derivePhysicalRows(last)
	items := make([]Item, 0, len(rows))
	for _, r := range rows {
		raw := map[string]any{
			"metric":      r.Key,
			"label":       r.Label,
			"value":       r.Value,
			"total":       r.Total,
			"unit":        r.Unit,
			"utilisation": r.Utilisation,
		}
		if r.Detail != "" {
			raw["detail"] = r.Detail
		}
		display := map[string]any{
			"metric":      r.Label,
			"value":       formatPhysicalValue(r),
			"unit":        r.Unit,
			"utilisation": percentString(r.Utilisation),
			"detail":      r.Detail,
		}
		items = append(items, Item{Raw: raw, Display: display})
	}
	return Envelope{
		Kind:  KindOverviewPhysical,
		Meta:  NewMeta(now.In(common.Timezone.Time()), c.OlaresID(), common.User),
		Items: items,
	}, nil
}

// derivePhysicalRows turns the last-sample map into the 9 (or so) rows the
// SPA renders. Each row carries the headline number + the unit + a
// utilisation ratio. Ordering mirrors the SPA's panel.
func derivePhysicalRows(last map[string]format.LastMonitoringSample) []physicalMetric {
	cpuUsage := sampleFloat(last["cluster_cpu_usage"])
	cpuTotal := sampleFloat(last["cluster_cpu_total"])
	cpuUtil := sampleFloat(last["cluster_cpu_utilisation"])

	memUsage := sampleFloat(last["cluster_memory_usage_wo_cache"])
	memTotal := sampleFloat(last["cluster_memory_total"])
	memUtil := sampleFloat(last["cluster_memory_utilisation"])

	diskUsage := sampleFloat(last["cluster_disk_size_usage"])
	diskCap := sampleFloat(last["cluster_disk_size_capacity"])
	diskUtil := sampleFloat(last["cluster_disk_size_utilisation"])

	podsRun := sampleFloat(last["cluster_pod_running_count"])
	podsQuota := sampleFloat(last["cluster_pod_quota"])

	netIn := sampleFloat(last["cluster_net_bytes_received"])
	netOut := sampleFloat(last["cluster_net_bytes_transmitted"])

	rows := []physicalMetric{
		{Key: "cpu", Label: "CPU", Value: cpuUsage, Total: cpuTotal, Unit: "core", Utilisation: cpuUtil},
		{Key: "memory", Label: "Memory", Value: memUsage, Total: memTotal, Unit: format.GetSuitableUnit(memTotal, format.UnitTypeMemory), Utilisation: memUtil},
		{Key: "disk", Label: "Disk", Value: diskUsage, Total: diskCap, Unit: format.GetSuitableUnit(diskCap, format.UnitTypeDisk), Utilisation: diskUtil},
		{Key: "pods", Label: "Pods", Value: podsRun, Total: podsQuota, Unit: "", Utilisation: safeRatio(podsRun, podsQuota)},
		{Key: "net_in", Label: "Net In", Value: netIn, Total: 0, Unit: format.GetSuitableUnit(netIn, format.UnitTypeThroughput), Utilisation: 0, Detail: format.GetThroughput(formatFloat(netIn))},
		{Key: "net_out", Label: "Net Out", Value: netOut, Total: 0, Unit: format.GetSuitableUnit(netOut, format.UnitTypeThroughput), Utilisation: 0, Detail: format.GetThroughput(formatFloat(netOut))},
	}
	return rows
}

// formatPhysicalValue renders the headline value column for a physical
// row, mirroring the SPA's "value / total" + unit formatting.
func formatPhysicalValue(r physicalMetric) string {
	switch r.Key {
	case "cpu":
		return fmt.Sprintf("%.2f / %.2f", r.Value, r.Total)
	case "memory":
		return fmt.Sprintf("%s / %s",
			format.GetDiskSize(formatFloat(r.Value)),
			format.GetDiskSize(formatFloat(r.Total)))
	case "disk":
		return fmt.Sprintf("%s / %s",
			format.GetDiskSize(formatFloat(r.Value)),
			format.GetDiskSize(formatFloat(r.Total)))
	case "pods":
		return fmt.Sprintf("%.0f / %.0f", r.Value, r.Total)
	case "net_in", "net_out":
		return r.Detail
	default:
		return formatFloat(r.Value)
	}
}

func writePhysicalTable(env Envelope) error {
	cols := []TableColumn{
		{Header: "METRIC", Get: func(it Item) string { return DisplayString(it, "metric") }},
		{Header: "VALUE", Get: func(it Item) string { return DisplayString(it, "value") }},
		{Header: "UTIL", Get: func(it Item) string { return DisplayString(it, "utilisation") }},
	}
	return WriteTable(os.Stdout, cols, env.Items)
}

// ----------------------------------------------------------------------------
// overview user — CPU / memory quota
// ----------------------------------------------------------------------------

func newOverviewUserCommand(f *cmdutil.Factory) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "user [<username>]",
		Short: "User-grain CPU / memory quota usage (mirrors the SPA's User Resources panel)",
		Example: `  olares-cli dashboard overview user
  olares-cli dashboard overview user alice    # admin only`,
		Args: cobra.MaximumNArgs(1),
		SilenceErrors: true,
		SilenceUsage:  true,
		RunE: func(c *cobra.Command, args []string) error {
			if err := common.Validate(); err != nil {
				return err
			}
			target := common.User
			if len(args) == 1 {
				target = args[0]
			}
			return runOverviewUser(c.Context(), f, target)
		},
	}
	return cmd
}

func runOverviewUser(ctx context.Context, f *cmdutil.Factory, target string) error {
	c, err := buildDashboardClient(ctx, f)
	if err != nil {
		return err
	}
	r := &Runner{
		Flags:       &common,
		Recommended: 60 * time.Second,
		Iter: func(ctx context.Context, iter int, now time.Time) (Envelope, error) {
			env, err := buildUserEnvelope(ctx, c, target, now)
			if err != nil {
				return env, err
			}
			env.Meta = NewMeta(now.In(common.Timezone.Time()), c.OlaresID(), target)
			env.Meta.RecommendedPollSeconds = 60
			if common.Output == OutputJSON {
				return env, nil
			}
			return env, writeUserTable(env)
		},
	}
	return r.Run(ctx)
}

func buildUserEnvelope(ctx context.Context, c *Client, target string, now time.Time) (Envelope, error) {
	user, err := resolveTargetUser(ctx, c, target)
	if err != nil {
		return Envelope{Kind: KindOverviewUser}, err
	}
	metrics := []string{
		"user_cpu_total", "user_cpu_usage", "user_cpu_utilisation",
		"user_memory_total", "user_memory_usage_wo_cache", "user_memory_utilisation",
	}
	res, err := fetchUserMetric(ctx, c, user.Name, metrics, defaultClusterWindow(), now, false)
	if err != nil {
		return Envelope{Kind: KindOverviewUser, Meta: Meta{User: user.Name}}, err
	}
	last := format.GetLastMonitoringData(res, 0)
	cpuTotal := sampleFloat(last["user_cpu_total"])
	cpuUsage := sampleFloat(last["user_cpu_usage"])
	cpuUtil := sampleFloat(last["user_cpu_utilisation"])
	memTotal := sampleFloat(last["user_memory_total"])
	memUsage := sampleFloat(last["user_memory_usage_wo_cache"])
	memUtil := sampleFloat(last["user_memory_utilisation"])

	rows := []map[string]any{
		{
			"metric":      "CPU",
			"used_raw":    cpuUsage,
			"total_raw":   cpuTotal,
			"utilisation": cpuUtil,
			"used":        fmt.Sprintf("%.2f", cpuUsage),
			"total":       fmt.Sprintf("%.2f", cpuTotal),
		},
		{
			"metric":      "Memory",
			"used_raw":    memUsage,
			"total_raw":   memTotal,
			"utilisation": memUtil,
			"used":        format.GetDiskSize(formatFloat(memUsage)),
			"total":       format.GetDiskSize(formatFloat(memTotal)),
		},
	}
	items := make([]Item, 0, len(rows))
	for _, r := range rows {
		items = append(items, Item{
			Raw: map[string]any{
				"metric":      r["metric"],
				"used":        r["used_raw"],
				"total":       r["total_raw"],
				"utilisation": r["utilisation"],
				"user":        user.Name,
			},
			Display: map[string]any{
				"metric":      r["metric"],
				"used":        r["used"],
				"total":       r["total"],
				"utilisation": percentString(toFloat(r["utilisation"])),
			},
		})
	}
	return Envelope{
		Kind:  KindOverviewUser,
		Meta:  NewMeta(now.In(common.Timezone.Time()), c.OlaresID(), user.Name),
		Items: items,
	}, nil
}

func writeUserTable(env Envelope) error {
	cols := []TableColumn{
		{Header: "METRIC", Get: func(it Item) string { return DisplayString(it, "metric") }},
		{Header: "USED", Get: func(it Item) string { return DisplayString(it, "used") }},
		{Header: "TOTAL", Get: func(it Item) string { return DisplayString(it, "total") }},
		{Header: "UTIL", Get: func(it Item) string { return DisplayString(it, "utilisation") }},
	}
	return WriteTable(os.Stdout, cols, env.Items)
}

// ----------------------------------------------------------------------------
// overview ranking — workload-grain ranking
// ----------------------------------------------------------------------------

func newOverviewRankingCommand(f *cmdutil.Factory) *cobra.Command {
	var sortDir string
	cmd := &cobra.Command{
		Use:   "ranking",
		Short: "Workload-grain (per-application) resource ranking (mirrors the SPA's UsageRanking widget)",
		Example: `  olares-cli dashboard overview ranking
  olares-cli dashboard overview ranking --sort asc --head 5`,
		SilenceErrors: true,
		SilenceUsage:  true,
		RunE: func(c *cobra.Command, args []string) error {
			if err := common.Validate(); err != nil {
				return err
			}
			return runOverviewRanking(c.Context(), f, sortDir)
		},
	}
	cmd.Flags().StringVar(&sortDir, "sort", "desc", "sort direction (asc or desc)")
	return cmd
}

func runOverviewRanking(ctx context.Context, f *cmdutil.Factory, sortDir string) error {
	if sortDir != "asc" && sortDir != "desc" {
		return fmt.Errorf("--sort: %q is not asc/desc", sortDir)
	}
	c, err := buildDashboardClient(ctx, f)
	if err != nil {
		return err
	}
	r := &Runner{
		Flags:       &common,
		Recommended: 60 * time.Second,
		Iter: func(ctx context.Context, iter int, now time.Time) (Envelope, error) {
			env, err := buildRankingEnvelope(ctx, c, common.User, sortDir, now)
			if err != nil {
				return env, err
			}
			env.Meta = NewMeta(now.In(common.Timezone.Time()), c.OlaresID(), common.User)
			env.Meta.RecommendedPollSeconds = 60
			env.Items = HeadItems(env.Items, common.Head)
			if common.Output == OutputJSON {
				return env, nil
			}
			return env, writeRankingTable(env)
		},
	}
	return r.Run(ctx)
}

func buildRankingEnvelope(ctx context.Context, c *Client, target, sortDir string, now time.Time) (Envelope, error) {
	return buildRankingEnvelopeBy(ctx, c, target, "cpu", sortDir, now)
}

// buildRankingEnvelopeBy is the common implementation behind both
// `overview ranking` (default sortBy=cpu) and `applications` (which
// exposes --sort-by). Mirrors `formatResult` in Applications2/config.ts:
// fetch via fetchWorkloadsMetrics, then render each row carrying the
// app's title / icon / state / pod count alongside the four metric values.
func buildRankingEnvelopeBy(ctx context.Context, c *Client, target, sortBy, sortDir string, now time.Time) (Envelope, error) {
	apps, userNs, err := loadAppsForRanking(ctx, c, target)
	if err != nil {
		return Envelope{Kind: KindOverviewRanking}, err
	}
	rows, err := fetchWorkloadsMetrics(ctx, c, workloadRequest{
		Apps: apps, UserNamespace: userNs, SortBy: sortBy, Sort: sortDir,
	}, defaultClusterWindow(), now)
	if err != nil {
		return Envelope{Kind: KindOverviewRanking}, err
	}
	items := make([]Item, 0, len(rows))
	for i, r := range rows {
		title := r.Title
		if title == "" {
			title = r.Name
		}
		raw := map[string]any{
			"rank":       i + 1,
			"app":        r.Name,
			"title":      title,
			"icon":       r.Icon,
			"namespace":  r.Namespace,
			"deployment": r.Deployment,
			"owner_kind": r.OwnerKind,
			"state":      r.State,
			"is_system":  r.IsSystem,
			"pods":       r.PodCount,
			"cpu":        r.CPU,
			"memory":     r.Memory,
			"net_in":     r.NetIn,
			"net_out":    r.NetOut,
		}
		state := r.State
		if state == "" {
			state = "Unknown"
		}
		display := map[string]any{
			"rank":      strconv.Itoa(i + 1),
			"app":       title,
			"namespace": r.Namespace,
			"state":     state,
			"pods":      strconv.Itoa(r.PodCount),
			"cpu":       fmt.Sprintf("%.3f", r.CPU),
			"memory":    format.GetDiskSize(formatFloat(r.Memory)),
			"net_in":    format.GetThroughput(formatFloat(r.NetIn)),
			"net_out":   format.GetThroughput(formatFloat(r.NetOut)),
		}
		items = append(items, Item{Raw: raw, Display: display})
	}
	return Envelope{
		Kind:  KindOverviewRanking,
		Meta:  NewMeta(now.In(common.Timezone.Time()), c.OlaresID(), target),
		Items: items,
	}, nil
}

func writeRankingTable(env Envelope) error {
	cols := []TableColumn{
		{Header: "RANK", Get: func(it Item) string { return DisplayString(it, "rank") }},
		{Header: "APP", Get: func(it Item) string { return DisplayString(it, "app") }},
		{Header: "NAMESPACE", Get: func(it Item) string { return DisplayString(it, "namespace") }},
		{Header: "CPU", Get: func(it Item) string { return DisplayString(it, "cpu") }},
		{Header: "MEMORY", Get: func(it Item) string { return DisplayString(it, "memory") }},
		{Header: "NET_IN", Get: func(it Item) string { return DisplayString(it, "net_in") }},
		{Header: "NET_OUT", Get: func(it Item) string { return DisplayString(it, "net_out") }},
	}
	return WriteTable(os.Stdout, cols, env.Items)
}

// loadAppsForRanking discovers the active user's app inventory the way the
// SPA does — via /user-service/api/myapps_v2 (fetchAppsList). Each app
// entry is then tagged `IsSystem` based on whether it lives in the user's
// `user-space-<username>` namespace, mirroring
// Applications2/IndexPage.vue:330 (`userNamespace = "user-space-${username}"`).
//
// Returns the apps + the user's `user-space-` namespace so the per-pod
// monitoring fetch can target the right ns.
func loadAppsForRanking(ctx context.Context, c *Client, target string) ([]workloadApp, string, error) {
	user, err := resolveTargetUser(ctx, c, target)
	if err != nil {
		return nil, "", err
	}
	if user.Name == "" {
		return nil, "", fmt.Errorf("loadAppsForRanking: empty username (server response missing user.username)")
	}
	userNs := fmt.Sprintf("user-space-%s", user.Name)

	raws, err := fetchAppsList(ctx, c)
	if err != nil {
		return nil, "", err
	}
	apps := make([]workloadApp, 0, len(raws))
	for _, it := range raws {
		apps = append(apps, workloadApp{
			Name:       it.Name,
			Title:      it.Title,
			Icon:       it.Icon,
			Namespace:  it.Namespace,
			Deployment: it.Deployment,
			OwnerKind:  it.OwnerKind,
			State:      it.State,
			IsSystem:   it.Namespace == userNs,
		})
	}
	return apps, userNs, nil
}

// ----------------------------------------------------------------------------
// overview cpu / memory / pods — per-node multi-metric tables
// ----------------------------------------------------------------------------

func newOverviewCPUCommand(f *cmdutil.Factory) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "cpu",
		Short: "Per-node CPU details (model / freq / cores / utilisation breakdown / temp / load avg)",
		Example: `  olares-cli dashboard overview cpu -o json`,
		SilenceErrors: true,
		SilenceUsage:  true,
		RunE: func(c *cobra.Command, args []string) error {
			if err := common.Validate(); err != nil {
				return err
			}
			return runPerNodeMetric(c.Context(), f, KindOverviewCPU, cpuMetricSet(), cpuColumns(), cpuDisplay)
		},
	}
	return cmd
}

func newOverviewPodsCommand(f *cmdutil.Factory) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "pods",
		Short: "Per-node pod count snapshot (last/avg/max running)",
		Example: `  olares-cli dashboard overview pods -o json`,
		SilenceErrors: true,
		SilenceUsage:  true,
		RunE: func(c *cobra.Command, args []string) error {
			if err := common.Validate(); err != nil {
				return err
			}
			return runPerNodeMetric(c.Context(), f, KindOverviewPods, podsMetricSet(), podsColumns(), podsDisplay)
		},
	}
	return cmd
}

func newOverviewMemoryCommand(f *cmdutil.Factory) *cobra.Command {
	var mode string
	cmd := &cobra.Command{
		Use:   "memory",
		Short: "Per-node memory breakdown (--mode physical | swap)",
		Example: `  olares-cli dashboard overview memory --mode physical
  olares-cli dashboard overview memory --mode swap`,
		SilenceErrors: true,
		SilenceUsage:  true,
		RunE: func(c *cobra.Command, args []string) error {
			if err := common.Validate(); err != nil {
				return err
			}
			switch mode {
			case "", "physical":
				return runPerNodeMetric(c.Context(), f, KindOverviewMemory, memoryPhysicalMetricSet(), memoryPhysicalColumns(), memoryPhysicalDisplay)
			case "swap":
				return runPerNodeMetric(c.Context(), f, KindOverviewMemory, memorySwapMetricSet(), memorySwapColumns(), memorySwapDisplay)
			default:
				return fmt.Errorf("--mode: %q must be physical or swap", mode)
			}
		},
	}
	cmd.Flags().StringVar(&mode, "mode", "physical", "memory view: physical | swap")
	return cmd
}

// runPerNodeMetric is the shared workhorse for cpu / memory / pods. It
// fetches the requested metric set against /v1alpha3/nodes, groups by node
// (the `node` label), and renders one row per node with the columns / display
// the caller specifies.
func runPerNodeMetric(ctx context.Context, f *cmdutil.Factory, kind string, metrics []string, cols []TableColumn, disp func(node string, last map[string]format.LastMonitoringSample) (rawCols, dispCols map[string]any)) error {
	c, err := buildDashboardClient(ctx, f)
	if err != nil {
		return err
	}
	r := &Runner{
		Flags:       &common,
		Recommended: 60 * time.Second,
		Iter: func(ctx context.Context, iter int, now time.Time) (Envelope, error) {
			env, err := buildPerNodeEnvelope(ctx, c, kind, metrics, disp, now)
			if err != nil {
				return env, err
			}
			env.Meta = NewMeta(now.In(common.Timezone.Time()), c.OlaresID(), common.User)
			env.Meta.RecommendedPollSeconds = 60
			if common.Output == OutputJSON {
				return env, nil
			}
			return env, WriteTable(os.Stdout, cols, env.Items)
		},
	}
	return r.Run(ctx)
}

// buildPerNodeEnvelope shells out to /v1alpha3/nodes and groups the results
// by the `node` label. Unlike fetchClusterMetrics (which collapses by
// metric_name), per-node metrics carry one row per node within each metric;
// we transpose into one Item per node.
func buildPerNodeEnvelope(ctx context.Context, c *Client, kind string, metrics []string, disp func(node string, last map[string]format.LastMonitoringSample) (rawCols, dispCols map[string]any), now time.Time) (Envelope, error) {
	q := monitoringQuery(metrics, defaultDetailWindow(), now, false)
	var raw struct {
		Results []struct {
			MetricName string `json:"metric_name"`
			Data       struct {
				Result []struct {
					Metric map[string]string `json:"metric"`
					Values [][]any           `json:"values"`
					Value  []any             `json:"value"`
				} `json:"result"`
			} `json:"data"`
		} `json:"results"`
	}
	if err := c.DoJSON(ctx, http.MethodGet, "/kapis/monitoring.kubesphere.io/v1alpha3/nodes", q, nil, &raw); err != nil {
		return Envelope{Kind: kind}, err
	}
	// Group rows by node label.
	type nodeBucket struct {
		samples map[string]format.LastMonitoringSample
	}
	buckets := map[string]*nodeBucket{}
	order := []string{}
	for _, r := range raw.Results {
		for _, e := range r.Data.Result {
			node := e.Metric["node"]
			if node == "" {
				node = e.Metric["instance"]
			}
			if node == "" {
				continue
			}
			b, ok := buckets[node]
			if !ok {
				b = &nodeBucket{samples: map[string]format.LastMonitoringSample{}}
				buckets[node] = b
				order = append(order, node)
			}
			b.samples[r.MetricName] = lastSampleFromRow(e.Values, e.Value)
		}
	}
	sort.Strings(order)
	items := make([]Item, 0, len(order))
	for _, n := range order {
		raws, disps := disp(n, buckets[n].samples)
		items = append(items, Item{Raw: raws, Display: disps})
	}
	return Envelope{
		Kind:  kind,
		Meta:  NewMeta(now.In(common.Timezone.Time()), c.OlaresID(), common.User),
		Items: items,
	}, nil
}

func lastSampleFromRow(values [][]any, value []any) format.LastMonitoringSample {
	if len(values) > 0 {
		row := values[len(values)-1]
		if len(row) >= 2 {
			ts, _ := row[0].(float64)
			s := fmt.Sprintf("%v", row[1])
			return format.LastMonitoringSample{Timestamp: ts, RawValue: s}
		}
	}
	if len(value) >= 2 {
		ts, _ := value[0].(float64)
		s := fmt.Sprintf("%v", value[1])
		return format.LastMonitoringSample{Timestamp: ts, RawValue: s}
	}
	return format.LastMonitoringSample{Empty: true}
}

// cpuMetricSet — column 1:1 with SPA Overview2/CPU/config.ts.
func cpuMetricSet() []string {
	return []string{
		"node_cpu_total", "node_cpu_utilisation",
		"node_user_cpu_usage", "node_system_cpu_usage", "node_iowait_cpu_usage",
		"node_load1", "node_load5", "node_load15",
		"node_cpu_temp_celsius",
		"node_cpu_base_frequency_hertz_max",
	}
}

func cpuColumns() []TableColumn {
	return []TableColumn{
		{Header: "NODE", Get: func(it Item) string { return DisplayString(it, "node") }},
		{Header: "FREQ", Get: func(it Item) string { return DisplayString(it, "freq") }},
		{Header: "CORES", Get: func(it Item) string { return DisplayString(it, "cores") }},
		{Header: "CPU_UTIL", Get: func(it Item) string { return DisplayString(it, "cpu_util") }},
		{Header: "USER", Get: func(it Item) string { return DisplayString(it, "user") }},
		{Header: "SYSTEM", Get: func(it Item) string { return DisplayString(it, "system") }},
		{Header: "IOWAIT", Get: func(it Item) string { return DisplayString(it, "iowait") }},
		{Header: "LOAD1", Get: func(it Item) string { return DisplayString(it, "load1") }},
		{Header: "LOAD5", Get: func(it Item) string { return DisplayString(it, "load5") }},
		{Header: "LOAD15", Get: func(it Item) string { return DisplayString(it, "load15") }},
		{Header: "TEMP", Get: func(it Item) string { return DisplayString(it, "temp") }},
	}
}

func cpuDisplay(node string, last map[string]format.LastMonitoringSample) (map[string]any, map[string]any) {
	cores := sampleFloat(last["node_cpu_total"])
	cpuUtil := sampleFloat(last["node_cpu_utilisation"])
	userCPU := sampleFloat(last["node_user_cpu_usage"])
	sysCPU := sampleFloat(last["node_system_cpu_usage"])
	iowait := sampleFloat(last["node_iowait_cpu_usage"])
	load1 := sampleFloat(last["node_load1"])
	load5 := sampleFloat(last["node_load5"])
	load15 := sampleFloat(last["node_load15"])
	temp := sampleFloat(last["node_cpu_temp_celsius"])
	freq := sampleFloat(last["node_cpu_base_frequency_hertz_max"])
	raw := map[string]any{
		"node":     node,
		"freq_hz":  freq,
		"cores":    cores,
		"cpu_util": cpuUtil,
		"user":     userCPU,
		"system":   sysCPU,
		"iowait":   iowait,
		"load1":    load1, "load5": load5, "load15": load15,
		"temp_c": temp,
	}
	disp := map[string]any{
		"node":     node,
		"freq":     format.FormatFrequency(freq, "Hz"),
		"cores":    fmt.Sprintf("%.0f", cores),
		"cpu_util": percentString(cpuUtil),
		"user":     percentString(userCPU),
		"system":   percentString(sysCPU),
		"iowait":   percentString(iowait),
		"load1":    fmt.Sprintf("%.2f", load1),
		"load5":    fmt.Sprintf("%.2f", load5),
		"load15":   fmt.Sprintf("%.2f", load15),
		"temp":     renderTemperature(temp, common.TempUnit),
	}
	return raw, disp
}

func memoryPhysicalMetricSet() []string {
	return []string{
		"node_memory_total", "node_memory_usage_wo_cache", "node_memory_available",
		"node_memory_utilisation", "node_memory_cached", "node_memory_buffers",
	}
}

func memoryPhysicalColumns() []TableColumn {
	return []TableColumn{
		{Header: "NODE", Get: func(it Item) string { return DisplayString(it, "node") }},
		{Header: "TOTAL", Get: func(it Item) string { return DisplayString(it, "total") }},
		{Header: "USED", Get: func(it Item) string { return DisplayString(it, "used") }},
		{Header: "AVAIL", Get: func(it Item) string { return DisplayString(it, "avail") }},
		{Header: "BUFFERS", Get: func(it Item) string { return DisplayString(it, "buffers") }},
		{Header: "CACHED", Get: func(it Item) string { return DisplayString(it, "cached") }},
		{Header: "UTIL", Get: func(it Item) string { return DisplayString(it, "util") }},
	}
}

func memoryPhysicalDisplay(node string, last map[string]format.LastMonitoringSample) (map[string]any, map[string]any) {
	total := sampleFloat(last["node_memory_total"])
	used := sampleFloat(last["node_memory_usage_wo_cache"])
	avail := sampleFloat(last["node_memory_available"])
	util := sampleFloat(last["node_memory_utilisation"])
	cached := sampleFloat(last["node_memory_cached"])
	buffers := sampleFloat(last["node_memory_buffers"])
	raw := map[string]any{
		"node": node, "total": total, "used": used, "avail": avail,
		"util": util, "cached": cached, "buffers": buffers, "mode": "physical",
	}
	disp := map[string]any{
		"node":    node,
		"total":   format.GetDiskSize(formatFloat(total)),
		"used":    format.GetDiskSize(formatFloat(used)),
		"avail":   format.GetDiskSize(formatFloat(avail)),
		"buffers": format.GetDiskSize(formatFloat(buffers)),
		"cached":  format.GetDiskSize(formatFloat(cached)),
		"util":    percentString(util),
	}
	return raw, disp
}

func memorySwapMetricSet() []string {
	return []string{
		"node_memory_swap_total", "node_memory_swap_used",
		"node_memory_pgpgin_rate", "node_memory_pgpgout_rate",
	}
}

func memorySwapColumns() []TableColumn {
	return []TableColumn{
		{Header: "NODE", Get: func(it Item) string { return DisplayString(it, "node") }},
		{Header: "TOTAL", Get: func(it Item) string { return DisplayString(it, "total") }},
		{Header: "USED", Get: func(it Item) string { return DisplayString(it, "used") }},
		{Header: "PG_IN", Get: func(it Item) string { return DisplayString(it, "pg_in") }},
		{Header: "PG_OUT", Get: func(it Item) string { return DisplayString(it, "pg_out") }},
		{Header: "UTIL", Get: func(it Item) string { return DisplayString(it, "util") }},
	}
}

func memorySwapDisplay(node string, last map[string]format.LastMonitoringSample) (map[string]any, map[string]any) {
	total := sampleFloat(last["node_memory_swap_total"])
	used := sampleFloat(last["node_memory_swap_used"])
	pgIn := sampleFloat(last["node_memory_pgpgin_rate"])
	pgOut := sampleFloat(last["node_memory_pgpgout_rate"])
	util := safeRatio(used, total)
	raw := map[string]any{
		"node": node, "total": total, "used": used, "pg_in": pgIn, "pg_out": pgOut,
		"util": util, "mode": "swap",
	}
	disp := map[string]any{
		"node":   node,
		"total":  format.GetDiskSize(formatFloat(total)),
		"used":   format.GetDiskSize(formatFloat(used)),
		"pg_in":  format.WorthValue(formatFloat(pgIn)),
		"pg_out": format.WorthValue(formatFloat(pgOut)),
		"util":   percentString(util),
	}
	return raw, disp
}

func podsMetricSet() []string {
	return []string{
		"node_pod_running_count", "node_pod_quota", "node_pod_utilisation",
	}
}

func podsColumns() []TableColumn {
	return []TableColumn{
		{Header: "NODE", Get: func(it Item) string { return DisplayString(it, "node") }},
		{Header: "RUNNING", Get: func(it Item) string { return DisplayString(it, "running") }},
		{Header: "QUOTA", Get: func(it Item) string { return DisplayString(it, "quota") }},
		{Header: "UTIL", Get: func(it Item) string { return DisplayString(it, "util") }},
	}
}

func podsDisplay(node string, last map[string]format.LastMonitoringSample) (map[string]any, map[string]any) {
	running := sampleFloat(last["node_pod_running_count"])
	quota := sampleFloat(last["node_pod_quota"])
	util := sampleFloat(last["node_pod_utilisation"])
	raw := map[string]any{"node": node, "running": running, "quota": quota, "util": util}
	disp := map[string]any{
		"node":    node,
		"running": fmt.Sprintf("%.0f", running),
		"quota":   fmt.Sprintf("%.0f", quota),
		"util":    percentString(util),
	}
	return raw, disp
}

// ----------------------------------------------------------------------------
// overview disk — sections (main + per-disk partitions)
// ----------------------------------------------------------------------------

func newOverviewDiskCommand(f *cmdutil.Factory) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "disk",
		Short: "Sections envelope: main = per-disk table; partitions = per-device partition tables",
		Example: `  olares-cli dashboard overview disk -o json
  olares-cli dashboard overview disk main
  olares-cli dashboard overview disk partitions sda`,
		SilenceErrors: true,
		SilenceUsage:  true,
		RunE: func(c *cobra.Command, args []string) error {
			if len(args) > 0 {
				return unknownSubcommandRunE(c, args)
			}
			if err := common.Validate(); err != nil {
				return err
			}
			return runOverviewDiskDefault(c.Context(), f)
		},
	}
	cmd.AddCommand(newOverviewDiskMainCommand(f))
	cmd.AddCommand(newOverviewDiskPartitionsCommand(f))
	return cmd
}

func runOverviewDiskDefault(ctx context.Context, f *cmdutil.Factory) error {
	c, err := buildDashboardClient(ctx, f)
	if err != nil {
		return err
	}
	now := time.Now()

	mainEnv, mainErr := buildDiskMainEnvelope(ctx, c, now)
	partitionEnvs := map[string]Envelope{}

	if mainErr == nil {
		// One partitions section per device row in main.
		for _, it := range mainEnv.Items {
			device := DisplayString(it, "device")
			if device == "-" || device == "" {
				continue
			}
			env, err := buildDiskPartitionsEnvelope(ctx, c, device, now)
			if err != nil {
				env = Envelope{Kind: KindOverviewDiskPart}
				env.Meta.Error = err.Error()
				env.Meta.ErrorKind = ClassifyTransportErr(err)
			}
			env.Meta.FetchedAt = time.Now().In(common.Timezone.Time()).Format(time.RFC3339)
			partitionEnvs[device] = env
		}
	}

	sections := map[string]Envelope{
		"main": mainEnv,
	}
	if mainErr != nil {
		mainEnv.Kind = KindOverviewDiskMain
		mainEnv.Meta.Error = mainErr.Error()
		mainEnv.Meta.ErrorKind = ClassifyTransportErr(mainErr)
		sections["main"] = mainEnv
	}
	// Embed per-device partitions under a single envelope whose Sections
	// field is the device→partitions map. Lets consumers walk
	// sections.partitions.sda just like sections.main.
	partsEnv := Envelope{Kind: KindOverviewDiskPart, Sections: partitionEnvs}
	sections["partitions"] = partsEnv

	env := Envelope{
		Kind:     KindOverviewDisk,
		Meta:     NewMeta(time.Now().In(common.Timezone.Time()), c.OlaresID(), common.User),
		Sections: sections,
	}

	if common.Output == OutputJSON {
		return WriteJSON(os.Stdout, env)
	}
	// Table mode: render main, then for each device its partitions.
	fmt.Fprintln(os.Stdout, "== MAIN ==")
	if mainErr != nil {
		fmt.Fprintf(os.Stdout, "(error: %s)\n", mainErr)
	} else {
		_ = writeDiskMainTable(mainEnv)
	}
	for device, pEnv := range partitionEnvs {
		fmt.Fprintf(os.Stdout, "\n== PARTITIONS: %s ==\n", device)
		if pEnv.Meta.Error != "" {
			fmt.Fprintf(os.Stdout, "(error: %s)\n", pEnv.Meta.Error)
			continue
		}
		_ = writeDiskPartitionsTable(pEnv)
	}
	return nil
}

func newOverviewDiskMainCommand(f *cmdutil.Factory) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "main",
		Short: "Per-physical-disk table (device / type / health / total / used / avail / temp / model / serial / firmware ...)",
		SilenceErrors: true,
		SilenceUsage:  true,
		RunE: func(c *cobra.Command, args []string) error {
			if err := common.Validate(); err != nil {
				return err
			}
			return runOverviewDiskMain(c.Context(), f)
		},
	}
	return cmd
}

func runOverviewDiskMain(ctx context.Context, f *cmdutil.Factory) error {
	c, err := buildDashboardClient(ctx, f)
	if err != nil {
		return err
	}
	r := &Runner{
		Flags:       &common,
		Recommended: 60 * time.Second,
		Iter: func(ctx context.Context, iter int, now time.Time) (Envelope, error) {
			env, err := buildDiskMainEnvelope(ctx, c, now)
			if err != nil {
				return env, err
			}
			env.Meta = NewMeta(now.In(common.Timezone.Time()), c.OlaresID(), common.User)
			env.Meta.RecommendedPollSeconds = 60
			if common.Output == OutputJSON {
				return env, nil
			}
			return env, writeDiskMainTable(env)
		},
	}
	return r.Run(ctx)
}

// buildDiskMainEnvelope mirrors SPA Overview2/Disk/IndexPage's "main"
// view (`getDiskOptions` in Overview2/Disk/config.ts:131). The driver
// metric is `node_disk_smartctl_info`; for each smartctl row we attach
// the matching samples from the auxiliary metrics by `(device, node)`
// label match (config.ts:142–151).
//
// Auxiliary metrics:
//
//   - node_disk_temp_celsius          → temperature
//   - node_one_disk_capacity_size     → reported "used capacity" total
//   - node_one_disk_avail_size        → free space (used = capacity - avail)
//   - node_disk_power_on_hours        → power-on hours
//   - node_one_disk_data_bytes_written → lifetime write volume
//
// All queries are instant (`step=0s`) — the SPA passes
// `step:'0s'` in IndexPage.vue:113.
//
// Per the user's policy decision the per-second IOPS / throughput
// columns are intentionally NOT emitted; the table is otherwise a 1:1
// of the SPA card content.
func buildDiskMainEnvelope(ctx context.Context, c *Client, now time.Time) (Envelope, error) {
	metrics := []string{
		"node_disk_smartctl_info",
		"node_disk_temp_celsius",
		"node_one_disk_capacity_size",
		"node_one_disk_avail_size",
		"node_disk_power_on_hours",
		"node_one_disk_data_bytes_written",
	}
	q := monitoringQuery(metrics, defaultDetailWindow(), now, true)
	var raw struct {
		Results []struct {
			MetricName string `json:"metric_name"`
			Data       struct {
				Result []struct {
					Metric map[string]string `json:"metric"`
					Values [][]any           `json:"values"`
					Value  []any             `json:"value"`
				} `json:"result"`
			} `json:"data"`
		} `json:"results"`
	}
	if err := c.DoJSON(ctx, http.MethodGet, "/kapis/monitoring.kubesphere.io/v1alpha3/nodes", q, nil, &raw); err != nil {
		return Envelope{Kind: KindOverviewDiskMain}, err
	}

	// Build a lookup of all auxiliary samples + the SMART rows.
	type smartRow struct {
		labels map[string]string
	}
	type auxSample struct {
		labels map[string]string
		sample format.LastMonitoringSample
	}
	smarts := []smartRow{}
	aux := map[string][]auxSample{}
	for _, r := range raw.Results {
		if r.MetricName == "node_disk_smartctl_info" {
			for _, e := range r.Data.Result {
				smarts = append(smarts, smartRow{labels: e.Metric})
			}
			continue
		}
		for _, e := range r.Data.Result {
			aux[r.MetricName] = append(aux[r.MetricName], auxSample{
				labels: e.Metric,
				sample: lastSampleFromRow(e.Values, e.Value),
			})
		}
	}

	// findAux mirrors `getLastMonitoringDataWithPath` (utils/monitoring)
	// + the predicate in config.ts:143–151:
	// metric.device | metric.disk_name must contain smart.device, AND
	// metric.node must equal smart.node.
	findAux := func(metricName string, smartDevice, smartNode string) format.LastMonitoringSample {
		samples, ok := aux[metricName]
		if !ok {
			return format.LastMonitoringSample{Empty: true}
		}
		for _, s := range samples {
			dev := s.labels["device"]
			if dev == "" {
				dev = s.labels["disk_name"]
			}
			node := s.labels["node"]
			if dev != "" && smartDevice != "" && strings.Contains(dev, smartDevice) && node == smartNode {
				return s.sample
			}
		}
		return format.LastMonitoringSample{Empty: true}
	}

	// Stable order: first by node, then by device, matching the SPA's
	// implicit order (smartctl rows arrive in the BFF's natural order
	// but we sort to keep table output deterministic in tests).
	sort.SliceStable(smarts, func(i, j int) bool {
		if smarts[i].labels["node"] != smarts[j].labels["node"] {
			return smarts[i].labels["node"] < smarts[j].labels["node"]
		}
		return smarts[i].labels["device"] < smarts[j].labels["device"]
	})

	items := make([]Item, 0, len(smarts))
	for _, s := range smarts {
		dev := s.labels["device"]
		node := s.labels["node"]
		name := s.labels["name"]
		rotational := s.labels["rotational"]
		logicalBlk := s.labels["logical_block_size"]
		if logicalBlk == "" {
			logicalBlk = "512"
		}
		physicalBlk := s.labels["physical_block_size"]
		if physicalBlk == "" {
			physicalBlk = "512"
		}
		const fourK = "4096"
		is4K := (rotational == "0" && logicalBlk == fourK) ||
			(rotational == "1" && logicalBlk == fourK && physicalBlk == fourK)
		typeStr := "SSD"
		if rotational == "1" {
			typeStr = "HDD"
		}
		healthOK := s.labels["health_ok"] == "true"
		healthStr := "Exception"
		if healthOK {
			healthStr = "Normal"
		}

		capLabel := s.labels["capacity"]
		capLabelF, _ := strconv.ParseFloat(capLabel, 64)

		capSample := findAux("node_one_disk_capacity_size", dev, node)
		availSample := findAux("node_one_disk_avail_size", dev, node)
		tempSample := findAux("node_disk_temp_celsius", dev, node)
		powerSample := findAux("node_disk_power_on_hours", dev, node)
		writtenSample := findAux("node_one_disk_data_bytes_written", dev, node)

		capUsed := sampleFloat(capSample)
		availUsed := sampleFloat(availSample)
		usedSize := capUsed - availUsed
		ratio := safeRatio(usedSize, capUsed)
		celsius := sampleFloat(tempSample)
		powerHours := sampleFloat(powerSample)
		written := sampleFloat(writtenSample)

		raw := map[string]any{
			"device":           dev,
			"node":             node,
			"name":             name,
			"type":             typeStr,
			"rotational":       rotational,
			"health_ok":        healthOK,
			"capacity_total":   capLabelF,
			"capacity_used":    capUsed,
			"capacity_avail":   availUsed,
			"used":             usedSize,
			"used_ratio":       ratio,
			"model":            s.labels["model"],
			"serial":           s.labels["serial"],
			"protocol":         s.labels["protocol"],
			"firmware":         s.labels["firmware"],
			"logical_block":    logicalBlk,
			"physical_block":   physicalBlk,
			"is_4k_native":     is4K,
			"temperature_c":    celsius,
			"power_on_hours":   powerHours,
			"data_bytes_written": written,
		}
		disp := map[string]any{
			"device":         dispOrDash(dev),
			"node":           dispOrDash(node),
			"type":           typeStr,
			"health":         healthStr,
			"total":          format.GetDiskSize(capLabel),
			"used":           format.GetDiskSize(formatFloat(usedSize)),
			"avail":          format.GetDiskSize(formatFloat(availUsed)),
			"util":           percentString(ratio),
			"temperature":    renderDiskTemperature(celsius, common.TempUnit),
			"model":          dispOrDash(s.labels["model"]),
			"serial":         dispOrDash(s.labels["serial"]),
			"protocol":       dispOrDash(s.labels["protocol"]),
			"firmware":       dispOrDash(s.labels["firmware"]),
			"is_4k_native":   ifYesNo(is4K),
			"power_on_hours": renderHoursOrDash(powerHours, powerSample.Empty),
			"write_volume":   format.GetDiskSize(formatFloat(written)),
		}
		items = append(items, Item{Raw: raw, Display: disp})
	}
	return Envelope{
		Kind:  KindOverviewDiskMain,
		Meta:  NewMeta(now.In(common.Timezone.Time()), c.OlaresID(), common.User),
		Items: items,
	}, nil
}

// renderDiskTemperature mirrors `renderTemperature` but only emits the
// active --temp-unit value (no dual celsius/fahrenheit display).
// Empty/zero celsius prints "-/-" the way the SPA does (config.ts:219).
func renderDiskTemperature(celsius float64, target format.TempUnit) string {
	if celsius == 0 {
		return "-"
	}
	return renderTemperature(celsius, target)
}

func renderHoursOrDash(hours float64, empty bool) string {
	if empty {
		return "-"
	}
	return fmt.Sprintf("%vh", strconv.FormatFloat(hours, 'f', -1, 64))
}

func dispOrDash(s string) string {
	if strings.TrimSpace(s) == "" {
		return "-"
	}
	return s
}

func ifYesNo(b bool) string {
	if b {
		return "Yes"
	}
	return "No"
}

// writeDiskMainTable renders the per-disk summary. Per the user's
// decision IOPS / throughput columns are dropped; everything that fits
// the SPA card lives here in column form.
func writeDiskMainTable(env Envelope) error {
	cols := []TableColumn{
		{Header: "DEVICE", Get: func(it Item) string { return DisplayString(it, "device") }},
		{Header: "NODE", Get: func(it Item) string { return DisplayString(it, "node") }},
		{Header: "TYPE", Get: func(it Item) string { return DisplayString(it, "type") }},
		{Header: "HEALTH", Get: func(it Item) string { return DisplayString(it, "health") }},
		{Header: "TOTAL", Get: func(it Item) string { return DisplayString(it, "total") }},
		{Header: "USED", Get: func(it Item) string { return DisplayString(it, "used") }},
		{Header: "AVAIL", Get: func(it Item) string { return DisplayString(it, "avail") }},
		{Header: "UTIL", Get: func(it Item) string { return DisplayString(it, "util") }},
		{Header: "TEMP", Get: func(it Item) string { return DisplayString(it, "temperature") }},
		{Header: "MODEL", Get: func(it Item) string { return DisplayString(it, "model") }},
		{Header: "SERIAL", Get: func(it Item) string { return DisplayString(it, "serial") }},
		{Header: "PROTOCOL", Get: func(it Item) string { return DisplayString(it, "protocol") }},
		{Header: "FIRMWARE", Get: func(it Item) string { return DisplayString(it, "firmware") }},
		{Header: "4K_NATIVE", Get: func(it Item) string { return DisplayString(it, "is_4k_native") }},
		{Header: "POWER_ON", Get: func(it Item) string { return DisplayString(it, "power_on_hours") }},
		{Header: "WRITTEN", Get: func(it Item) string { return DisplayString(it, "write_volume") }},
	}
	return WriteTable(os.Stdout, cols, env.Items)
}

func newOverviewDiskPartitionsCommand(f *cmdutil.Factory) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "partitions <device>",
		Short: "Partition-level table for one physical device",
		Args:  cobra.ExactArgs(1),
		SilenceErrors: true,
		SilenceUsage:  true,
		RunE: func(c *cobra.Command, args []string) error {
			if err := common.Validate(); err != nil {
				return err
			}
			return runOverviewDiskPartitions(c.Context(), f, args[0])
		},
	}
	return cmd
}

func runOverviewDiskPartitions(ctx context.Context, f *cmdutil.Factory, device string) error {
	c, err := buildDashboardClient(ctx, f)
	if err != nil {
		return err
	}
	r := &Runner{
		Flags:       &common,
		Recommended: 60 * time.Second,
		Iter: func(ctx context.Context, iter int, now time.Time) (Envelope, error) {
			env, err := buildDiskPartitionsEnvelope(ctx, c, device, now)
			if err != nil {
				return env, err
			}
			env.Meta = NewMeta(now.In(common.Timezone.Time()), c.OlaresID(), common.User)
			env.Meta.RecommendedPollSeconds = 60
			if common.Output == OutputJSON {
				return env, nil
			}
			return env, writeDiskPartitionsTable(env)
		},
	}
	return r.Run(ctx)
}

// buildDiskPartitionsEnvelope mirrors SPA Overview2/Disk/IndexPage's
// "Occupancy Analysis" popup (`getDiskPartitionRows` in
// Overview2/Disk/config.ts:398).
//
// It pulls `node_disk_lsblk_info` (instant), filters to the rows whose
// `metric.node` == <node> (the SPA uses the node selected via the
// device card); when `node` is empty the CLI auto-resolves it by
// scanning the SMART rows for a row matching the requested device.
//
// The candidate subset for the requested root `device` is then chosen
// per the SPA logic:
//
//  1. if any row in the node's lsblk set has a `pkname` label →
//     `collectSubtreeByPkname` BFS from `<device>`.
//  2. otherwise fall back to `name == device || name.startsWith(device)`.
//
// Finally `flattenLsblkHierarchy` walks the tree pre-order and
// computes the ASCII tree prefix per row (`├── `, `└── `, `│   `, ...).
// Per the user's decision the Display column for `name` carries that
// tree prefix so humans see the topology at a glance.
func buildDiskPartitionsEnvelope(ctx context.Context, c *Client, device string, now time.Time) (Envelope, error) {
	q := monitoringQuery([]string{"node_disk_lsblk_info"}, defaultDetailWindow(), now, true)
	var raw struct {
		Results []struct {
			MetricName string `json:"metric_name"`
			Data       struct {
				Result []struct {
					Metric map[string]string `json:"metric"`
					Values [][]any           `json:"values"`
					Value  []any             `json:"value"`
				} `json:"result"`
			} `json:"data"`
		} `json:"results"`
	}
	if err := c.DoJSON(ctx, http.MethodGet, "/kapis/monitoring.kubesphere.io/v1alpha3/nodes", q, nil, &raw); err != nil {
		return Envelope{Kind: KindOverviewDiskPart}, err
	}

	// Find the node hosting <device>. The SPA pulls this from the
	// SMART card the user clicked; the CLI infers it from any row
	// whose name == <device>. If none, fall through with node="" and
	// match all rows; that mirrors SPA's degraded path when the popup
	// is opened without a node selection.
	var node string
	allRows := []lsblkRow{}
	for _, r := range raw.Results {
		if r.MetricName != "node_disk_lsblk_info" {
			continue
		}
		for _, e := range r.Data.Result {
			if e.Metric["name"] == device && node == "" {
				node = e.Metric["node"]
			}
			allRows = append(allRows, lsblkRow{
				Name:         e.Metric["name"],
				Node:         e.Metric["node"],
				Pkname:       e.Metric["pkname"],
				Size:         e.Metric["size"],
				Fstype:       e.Metric["fstype"],
				Mountpoint:   e.Metric["mountpoint"],
				Fsused:       e.Metric["fsused"],
				FsusePercent: e.Metric["fsuse_percent"],
			})
		}
	}

	scoped := make([]lsblkRow, 0, len(allRows))
	for _, r := range allRows {
		if node == "" || r.Node == node {
			scoped = append(scoped, r)
		}
	}

	var subset []lsblkRow
	if hasPknameLabels(scoped) {
		subset = collectSubtreeByPkname(scoped, device)
	} else {
		subset = subset[:0]
		for _, r := range scoped {
			if r.Name == device || strings.HasPrefix(r.Name, device) {
				subset = append(subset, r)
			}
		}
	}
	flat := flattenLsblkHierarchy(subset, device)

	items := make([]Item, 0, len(flat))
	for _, fr := range flat {
		raw := map[string]any{
			"name":          fr.Row.Name,
			"node":          fr.Row.Node,
			"parent":        fr.Parent,
			"depth":         fr.Depth,
			"size":          fr.Row.Size,
			"fstype":        fr.Row.Fstype,
			"mountpoint":    fr.Row.Mountpoint,
			"fsused":        fr.Row.Fsused,
			"fsuse_percent": fr.Row.FsusePercent,
		}
		disp := map[string]any{
			"name":          fr.TreePrefix + fr.Row.Name,
			"size":          orDash(fr.Row.Size),
			"fstype":        orDash(fr.Row.Fstype),
			"mountpoint":    orDash(fr.Row.Mountpoint),
			"fsused":        orDash(fr.Row.Fsused),
			"fsuse_percent": orDash(fr.Row.FsusePercent),
		}
		items = append(items, Item{Raw: raw, Display: disp})
	}
	return Envelope{
		Kind:  KindOverviewDiskPart,
		Meta:  NewMeta(now.In(common.Timezone.Time()), c.OlaresID(), common.User),
		Items: items,
	}, nil
}

// orDash mirrors `displayLsblkCell` in Overview2/Disk/config.ts:30 —
// trim → empty falls through to "-".
func orDash(v string) string {
	t := strings.TrimSpace(v)
	if t == "" {
		return "-"
	}
	return t
}

// writeDiskPartitionsTable renders the lsblk subtree. NAME carries the
// ASCII tree prefix; the other columns are the SPA's `getLsblkColumns`.
func writeDiskPartitionsTable(env Envelope) error {
	cols := []TableColumn{
		{Header: "NAME", Get: func(it Item) string { return DisplayString(it, "name") }},
		{Header: "SIZE", Get: func(it Item) string { return DisplayString(it, "size") }},
		{Header: "FSTYPE", Get: func(it Item) string { return DisplayString(it, "fstype") }},
		{Header: "MOUNT", Get: func(it Item) string { return DisplayString(it, "mountpoint") }},
		{Header: "FSUSED", Get: func(it Item) string { return DisplayString(it, "fsused") }},
		{Header: "FSUSE%", Get: func(it Item) string { return DisplayString(it, "fsuse_percent") }},
	}
	return WriteTable(os.Stdout, cols, env.Items)
}

// ----------------------------------------------------------------------------
// overview network — per-iface system-ifs table
// ----------------------------------------------------------------------------

func newOverviewNetworkCommand(f *cmdutil.Factory) *cobra.Command {
	var testConn bool
	cmd := &cobra.Command{
		Use:   "network",
		Short: "Per-physical-NIC table from capi /system/ifs",
		SilenceErrors: true,
		SilenceUsage:  true,
		RunE: func(c *cobra.Command, args []string) error {
			if err := common.Validate(); err != nil {
				return err
			}
			return runOverviewNetwork(c.Context(), f, testConn)
		},
	}
	cmd.Flags().BoolVar(&testConn, "test-connectivity", true, "ask the BFF to probe internet/IPv6 connectivity per interface")
	return cmd
}

func runOverviewNetwork(ctx context.Context, f *cmdutil.Factory, testConn bool) error {
	c, err := buildDashboardClient(ctx, f)
	if err != nil {
		return err
	}
	r := &Runner{
		Flags:       &common,
		Recommended: 60 * time.Second,
		Iter: func(ctx context.Context, iter int, now time.Time) (Envelope, error) {
			ifs, err := fetchSystemIFS(ctx, c, testConn)
			if err != nil {
				return Envelope{Kind: KindOverviewNetwork}, err
			}
			items := make([]Item, 0, len(ifs))
			for i, it := range ifs {
				port := fmt.Sprintf("Port-%d", i+1)
				status := "down"
				if it.InternetConnected {
					status = "up"
				}
				raw := map[string]any{
					"port":            port,
					"iface":           it.Iface,
					"status":          status,
					"is_host_ip":      it.IsHostIp,
					"hostname":        it.Hostname,
					"method":          it.Method,
					"mtu":             it.MTU,
					"ip":              it.IP,
					"ipv4_mask":       it.IPv4Mask,
					"ipv4_gateway":    it.IPv4Gateway,
					"ipv4_dns":        it.IPv4DNS,
					"ipv6_address":    it.IPv6Address,
					"ipv6_gateway":    it.IPv6Gateway,
					"ipv6_dns":        it.IPv6DNS,
					"ipv4_connected":  it.InternetConnected,
					"ipv6_connected":  it.IPv6Connectivity,
					"tx_rate_raw":     it.TxRate,
					"rx_rate_raw":     it.RxRate,
				}
				disp := map[string]any{
					"port":         port,
					"iface":        it.Iface,
					"status":       status,
					"tx":           formatRateAny(it.TxRate),
					"rx":           formatRateAny(it.RxRate),
					"mtu":          fmt.Sprintf("%v", it.MTU),
					"method":       it.Method,
					"host":         it.Hostname,
					"ipv4":         it.IP,
					"ipv4_mask":    it.IPv4Mask,
					"ipv4_gateway": it.IPv4Gateway,
					"ipv4_dns":     it.IPv4DNS,
					"ipv6":         it.IPv6Address,
					"ipv6_gateway": it.IPv6Gateway,
					"ipv6_dns":     it.IPv6DNS,
				}
				items = append(items, Item{Raw: raw, Display: disp})
			}
			env := Envelope{
				Kind:  KindOverviewNetwork,
				Meta:  NewMeta(now.In(common.Timezone.Time()), c.OlaresID(), common.User),
				Items: HeadItems(items, common.Head),
			}
			env.Meta.RecommendedPollSeconds = 60
			if common.Output == OutputJSON {
				return env, nil
			}
			return env, writeNetworkTable(env)
		},
	}
	return r.Run(ctx)
}

func writeNetworkTable(env Envelope) error {
	cols := []TableColumn{
		{Header: "PORT", Get: func(it Item) string { return DisplayString(it, "port") }},
		{Header: "IFACE", Get: func(it Item) string { return DisplayString(it, "iface") }},
		{Header: "STATUS", Get: func(it Item) string { return DisplayString(it, "status") }},
		{Header: "TX", Get: func(it Item) string { return DisplayString(it, "tx") }},
		{Header: "RX", Get: func(it Item) string { return DisplayString(it, "rx") }},
		{Header: "MTU", Get: func(it Item) string { return DisplayString(it, "mtu") }},
		{Header: "METHOD", Get: func(it Item) string { return DisplayString(it, "method") }},
		{Header: "HOST", Get: func(it Item) string { return DisplayString(it, "host") }},
		{Header: "IPV4", Get: func(it Item) string { return DisplayString(it, "ipv4") }},
		{Header: "MASK", Get: func(it Item) string { return DisplayString(it, "ipv4_mask") }},
		{Header: "GW4", Get: func(it Item) string { return DisplayString(it, "ipv4_gateway") }},
		{Header: "DNS4", Get: func(it Item) string { return DisplayString(it, "ipv4_dns") }},
		{Header: "IPV6", Get: func(it Item) string { return DisplayString(it, "ipv6") }},
		{Header: "GW6", Get: func(it Item) string { return DisplayString(it, "ipv6_gateway") }},
		{Header: "DNS6", Get: func(it Item) string { return DisplayString(it, "ipv6_dns") }},
	}
	return WriteTable(os.Stdout, cols, env.Items)
}

// ----------------------------------------------------------------------------
// overview fan — sections (live + curve)
// ----------------------------------------------------------------------------

func newOverviewFanCommand(f *cmdutil.Factory) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "fan",
		Short: "Sections envelope: live = real-time fan/temperature/power; curve = hardcoded fan-curve spec",
		SilenceErrors: true,
		SilenceUsage:  true,
		RunE: func(c *cobra.Command, args []string) error {
			if len(args) > 0 {
				return unknownSubcommandRunE(c, args)
			}
			if err := common.Validate(); err != nil {
				return err
			}
			return runOverviewFanDefault(c.Context(), f)
		},
	}
	cmd.AddCommand(newOverviewFanLiveCommand(f))
	cmd.AddCommand(newOverviewFanCurveCommand(f))
	return cmd
}

func runOverviewFanDefault(ctx context.Context, f *cmdutil.Factory) error {
	c, err := buildDashboardClient(ctx, f)
	if err != nil {
		return err
	}
	now := time.Now()

	// Capability gate: Fan / cooling features are Olares One-only.
	// Mirrors `FanStore.isOlaresOneDevice` in
	// `Overview2/ClusterResource.vue:238`. Both `live` and `curve`
	// share the gate per the user's policy decision (curve is
	// hardware-specific spec, not portable reference data).
	if gated, ok := gateOlaresOne(ctx, c, KindOverviewFan, now); ok {
		// The aggregate envelope mirrors the live + curve sections,
		// each carrying the same `not_olares_one` reason so consumers
		// can demux either at the top or per-section.
		liveGated := gated
		liveGated.Kind = KindOverviewFanLive
		curveGated := gated
		curveGated.Kind = KindOverviewFanCurve
		gated.Sections = map[string]Envelope{
			"live":  liveGated,
			"curve": curveGated,
		}
		if common.Output == OutputJSON {
			return WriteJSON(os.Stdout, gated)
		}
		return nil
	}

	live, lerr := buildFanLiveEnvelope(ctx, c, now)
	if lerr != nil {
		live.Kind = KindOverviewFanLive
		live.Meta.Error = lerr.Error()
		live.Meta.ErrorKind = ClassifyTransportErr(lerr)
	}
	curve := buildFanCurveEnvelope(now, c.OlaresID())

	env := Envelope{
		Kind: KindOverviewFan,
		Meta: NewMeta(time.Now().In(common.Timezone.Time()), c.OlaresID(), common.User),
		Sections: map[string]Envelope{
			"live":  live,
			"curve": curve,
		},
	}
	if common.Output == OutputJSON {
		return WriteJSON(os.Stdout, env)
	}
	fmt.Fprintln(os.Stdout, "== LIVE ==")
	if lerr != nil {
		fmt.Fprintf(os.Stdout, "(error: %s)\n", lerr)
	} else {
		_ = writeFanLiveTable(live)
	}
	fmt.Fprintln(os.Stdout, "\n== CURVE ==")
	return writeFanCurveTable(curve)
}

func newOverviewFanLiveCommand(f *cmdutil.Factory) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "live",
		Short: "1-row real-time fan / temperature / power snapshot (Olares One)",
		SilenceErrors: true,
		SilenceUsage:  true,
		RunE: func(c *cobra.Command, args []string) error {
			if err := common.Validate(); err != nil {
				return err
			}
			return runOverviewFanLive(c.Context(), f)
		},
	}
	return cmd
}

func runOverviewFanLive(ctx context.Context, f *cmdutil.Factory) error {
	c, err := buildDashboardClient(ctx, f)
	if err != nil {
		return err
	}
	r := &Runner{
		Flags:       &common,
		Recommended: 5 * time.Second,
		Iter: func(ctx context.Context, iter int, now time.Time) (Envelope, error) {
			// Capability gate runs each iteration so a `--watch`
			// stream against the wrong device terminates with a
			// clear empty envelope per tick rather than silent zeros.
			if gated, ok := gateOlaresOne(ctx, c, KindOverviewFanLive, now); ok {
				return gated, nil
			}
			env, err := buildFanLiveEnvelope(ctx, c, now)
			if err != nil {
				return env, err
			}
			env.Meta = NewMeta(now.In(common.Timezone.Time()), c.OlaresID(), common.User)
			env.Meta.RecommendedPollSeconds = 5
			if common.Output == OutputJSON {
				return env, nil
			}
			return env, writeFanLiveTable(env)
		},
	}
	return r.Run(ctx)
}

func buildFanLiveEnvelope(ctx context.Context, c *Client, now time.Time) (Envelope, error) {
	fan, err := fetchSystemFan(ctx, c)
	if err != nil {
		// 404 → no fan integration. Surface the empty envelope so the
		// caller's three-state branch works.
		if he, ok := IsHTTPError(err); ok && he.Status == http.StatusNotFound {
			env := Envelope{Kind: KindOverviewFanLive}
			env.Meta.Empty = true
			env.Meta.EmptyReason = "no_fan_integration"
			env.Meta.HTTPStatus = he.Status
			return env, nil
		}
		return Envelope{Kind: KindOverviewFanLive}, err
	}
	gpuPower, gpuPowerLimit := 0.0, 0.0
	if list, _ := fetchGraphicsList(ctx, c, nil); len(list) > 0 {
		if v, ok := list[0]["power"].(float64); ok {
			gpuPower = v
		}
		if v, ok := list[0]["powerLimit"].(float64); ok {
			gpuPowerLimit = v
		}
	}

	raw := map[string]any{
		"cpu_fan_rpm":     fan.CPUFanSpeed,
		"cpu_fan_rpm_max": fanSpeedMaxCPU,
		"cpu_temp_c":      fan.CPUTemperature,
		"gpu_fan_rpm":     fan.GPUFanSpeed,
		"gpu_fan_rpm_max": fanSpeedMaxGPU,
		"gpu_temp_c":      fan.GPUTemperature,
		"gpu_power":       gpuPower,
		"gpu_power_limit": gpuPowerLimit,
	}
	disp := map[string]any{
		"cpu_fan":        fmt.Sprintf("%.0f / %d RPM", fan.CPUFanSpeed, fanSpeedMaxCPU),
		"cpu_temp":       renderTemperature(fan.CPUTemperature, common.TempUnit),
		"gpu_fan":        fmt.Sprintf("%.0f / %d RPM", fan.GPUFanSpeed, fanSpeedMaxGPU),
		"gpu_temp":       renderTemperature(fan.GPUTemperature, common.TempUnit),
		"gpu_power":      fmt.Sprintf("%.2f W", gpuPower),
		"gpu_power_lim":  fmt.Sprintf("%.0f W", gpuPowerLimit),
	}
	return Envelope{
		Kind:  KindOverviewFanLive,
		Meta:  NewMeta(now.In(common.Timezone.Time()), c.OlaresID(), common.User),
		Items: []Item{{Raw: raw, Display: disp}},
	}, nil
}

func writeFanLiveTable(env Envelope) error {
	cols := []TableColumn{
		{Header: "CPU_FAN", Get: func(it Item) string { return DisplayString(it, "cpu_fan") }},
		{Header: "CPU_TEMP", Get: func(it Item) string { return DisplayString(it, "cpu_temp") }},
		{Header: "GPU_FAN", Get: func(it Item) string { return DisplayString(it, "gpu_fan") }},
		{Header: "GPU_TEMP", Get: func(it Item) string { return DisplayString(it, "gpu_temp") }},
		{Header: "GPU_POWER", Get: func(it Item) string { return DisplayString(it, "gpu_power") }},
		{Header: "POWER_LIM", Get: func(it Item) string { return DisplayString(it, "gpu_power_lim") }},
	}
	return WriteTable(os.Stdout, cols, env.Items)
}

func newOverviewFanCurveCommand(f *cmdutil.Factory) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "curve",
		Short: "10-row hardcoded fan-curve specification (RPM ↔ temperature range)",
		SilenceErrors: true,
		SilenceUsage:  true,
		RunE: func(c *cobra.Command, args []string) error {
			if err := common.Validate(); err != nil {
				return err
			}
			return runOverviewFanCurve(c.Context(), f)
		},
	}
	return cmd
}

func runOverviewFanCurve(ctx context.Context, f *cmdutil.Factory) error {
	c, err := buildDashboardClient(ctx, f)
	if err != nil {
		return err
	}
	now := time.Now()
	if gated, ok := gateOlaresOne(ctx, c, KindOverviewFanCurve, now); ok {
		if common.Output == OutputJSON {
			return WriteJSON(os.Stdout, gated)
		}
		return nil
	}
	env := buildFanCurveEnvelope(now, c.OlaresID())
	if common.Output == OutputJSON {
		return WriteJSON(os.Stdout, env)
	}
	return writeFanCurveTable(env)
}

func buildFanCurveEnvelope(now time.Time, olaresID string) Envelope {
	items := make([]Item, 0, len(fanCurveTable))
	for _, r := range fanCurveTable {
		raw := map[string]any{
			"step":           r.Step,
			"cpu_fan_rpm":    r.CPUFanRPM,
			"gpu_fan_rpm":    r.GPUFanRPM,
			"cpu_temp_range": r.CPUTempRange,
			"gpu_temp_range": r.GPUTempRange,
		}
		disp := map[string]any{
			"step":           strconv.Itoa(r.Step),
			"cpu_fan_rpm":    strconv.Itoa(r.CPUFanRPM),
			"gpu_fan_rpm":    strconv.Itoa(r.GPUFanRPM),
			"cpu_temp_range": r.CPUTempRange,
			"gpu_temp_range": r.GPUTempRange,
		}
		items = append(items, Item{Raw: raw, Display: disp})
	}
	return Envelope{
		Kind:  KindOverviewFanCurve,
		Meta:  NewMeta(now.In(common.Timezone.Time()), olaresID, common.User),
		Items: items,
	}
}

func writeFanCurveTable(env Envelope) error {
	cols := []TableColumn{
		{Header: "STEP", Get: func(it Item) string { return DisplayString(it, "step") }},
		{Header: "CPU_RPM", Get: func(it Item) string { return DisplayString(it, "cpu_fan_rpm") }},
		{Header: "GPU_RPM", Get: func(it Item) string { return DisplayString(it, "gpu_fan_rpm") }},
		{Header: "CPU_TEMP", Get: func(it Item) string { return DisplayString(it, "cpu_temp_range") }},
		{Header: "GPU_TEMP", Get: func(it Item) string { return DisplayString(it, "gpu_temp_range") }},
	}
	return WriteTable(os.Stdout, cols, env.Items)
}

// ----------------------------------------------------------------------------
// overview gpu — list / tasks / get / task
// ----------------------------------------------------------------------------

func newOverviewGPUCommand(f *cmdutil.Factory) *cobra.Command {
	cmd := &cobra.Command{
		Use:           "gpu",
		Short:         "vGPU views: list / tasks / get / task / detail / task-detail",
		SilenceErrors: true,
		SilenceUsage:  true,
		RunE: func(c *cobra.Command, args []string) error {
			if len(args) > 0 {
				return unknownSubcommandRunE(c, args)
			}
			// Default action: forward to `list`.
			if err := common.Validate(); err != nil {
				return err
			}
			return runOverviewGPUList(c.Context(), f)
		},
	}
	cmd.AddCommand(newOverviewGPUListCommand(f))
	cmd.AddCommand(newOverviewGPUTasksCommand(f))
	cmd.AddCommand(newOverviewGPUGetCommand(f))
	cmd.AddCommand(newOverviewGPUTaskCommand(f))
	cmd.AddCommand(newOverviewGPUDetailFullCommand(f))
	cmd.AddCommand(newOverviewGPUTaskDetailFullCommand(f))
	return cmd
}

func newOverviewGPUListCommand(f *cmdutil.Factory) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "list",
		Short: "List discovered vGPUs (Graphics management tab; 404 = HAMI not installed)",
		SilenceErrors: true,
		SilenceUsage:  true,
		RunE: func(c *cobra.Command, args []string) error {
			if err := common.Validate(); err != nil {
				return err
			}
			return runOverviewGPUList(c.Context(), f)
		},
	}
	return cmd
}

func runOverviewGPUList(ctx context.Context, f *cmdutil.Factory) error {
	c, err := buildDashboardClient(ctx, f)
	if err != nil {
		return err
	}
	now := time.Now()
	advisoryNote, advisoryReason := gpuAdvisory(ctx, c)
	list, err := fetchGraphicsList(ctx, c, nil)
	env := Envelope{Kind: KindOverviewGPUList, Meta: NewMeta(now.In(common.Timezone.Time()), c.OlaresID(), common.User)}
	if advisoryNote != "" {
		env.Meta.Note = advisoryNote
	}
	if err != nil {
		if he, ok := IsHTTPError(err); ok && he.Status == http.StatusNotFound {
			env.Meta.Empty = true
			env.Meta.EmptyReason = "no_vgpu_integration"
			env.Meta.HTTPStatus = he.Status
			if common.Output == OutputJSON {
				return WriteJSON(os.Stdout, env)
			}
			fmt.Fprintln(os.Stdout, "(no vGPUs detected — HAMI integration absent)")
			return nil
		}
		if unavail, ok := vgpuUnavailableFromError(c, err, KindOverviewGPUList, now); ok {
			if advisoryNote != "" {
				// Stack the advisory ahead of the unavailability
				// note; humans see "GPU sidebar hidden + HAMI down"
				// in one shot. Agents still get both as a single
				// `meta.note` string separated by " | ".
				unavail.Meta.Note = advisoryNote + " | " + unavail.Meta.Note
			}
			if common.Output == OutputJSON {
				return WriteJSON(os.Stdout, unavail)
			}
			return nil
		}
		return err
	}
	if len(list) == 0 {
		env.Meta.Empty = true
		env.Meta.EmptyReason = "no_gpu_detected"
		if common.Output == OutputJSON {
			return WriteJSON(os.Stdout, env)
		}
		fmt.Fprintln(os.Stdout, "(no GPUs reported — HAMI integration up but no devices)")
		return nil
	}
	_ = advisoryReason // stash; intentionally not surfaced separately, the SPA does not either
	// Field names below match HAMI's actual response shape (see SPA's
	// `Graphics` interface in src/apps/dashboard/types/gpu.ts and the
	// fixture captured from olarestest005). Earlier revisions guessed
	// at field names like "modelName" / "hostname" / "totalMem" — none
	// of which HAMI ever returns; the table silently rendered "<nil>".
	// We expose the entire HAMI object under `Raw` so agents can pull
	// fields the table doesn't surface (vgpuUsed/vgpuTotal, nodeUid,
	// memoryUtilizedPercent, etc.).
	for _, g := range list {
		raw := map[string]any{}
		for k, v := range g {
			raw[k] = v
		}
		disp := map[string]any{
			"gpu_id":      fmt.Sprintf("%v", g["uuid"]),
			"model":       fmt.Sprintf("%v", g["type"]),
			"mode":        gpuModeLabel(g["shareMode"]),
			"host_node":   fmt.Sprintf("%v", g["nodeName"]),
			"health":      gpuHealthLabel(g["health"]),
			"core_util":   percentDirect(toFloat(g["coreUtilizedPercent"])),
			"vram_total":  gpuVRAMHuman(g["memoryTotal"]),
			"vram_used":   gpuVRAMHuman(g["memoryUsed"]),
			"power":       fmt.Sprintf("%.2f W", toFloat(g["power"])),
			"temperature": renderTemperature(toFloat(g["temperature"]), common.TempUnit),
		}
		env.Items = append(env.Items, Item{Raw: raw, Display: disp})
	}
	if common.Output == OutputJSON {
		return WriteJSON(os.Stdout, env)
	}
	cols := []TableColumn{
		{Header: "GPU_ID", Get: func(it Item) string { return DisplayString(it, "gpu_id") }},
		{Header: "MODEL", Get: func(it Item) string { return DisplayString(it, "model") }},
		{Header: "MODE", Get: func(it Item) string { return DisplayString(it, "mode") }},
		{Header: "HOST", Get: func(it Item) string { return DisplayString(it, "host_node") }},
		{Header: "HEALTH", Get: func(it Item) string { return DisplayString(it, "health") }},
		{Header: "CORE_UTIL", Get: func(it Item) string { return DisplayString(it, "core_util") }},
		{Header: "VRAM", Get: func(it Item) string { return DisplayString(it, "vram_total") }},
		{Header: "POWER", Get: func(it Item) string { return DisplayString(it, "power") }},
		{Header: "TEMP", Get: func(it Item) string { return DisplayString(it, "temperature") }},
	}
	return WriteTable(os.Stdout, cols, env.Items)
}

func newOverviewGPUTasksCommand(f *cmdutil.Factory) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "tasks",
		Short: "List vGPU tasks (Task management tab)",
		SilenceErrors: true,
		SilenceUsage:  true,
		RunE: func(c *cobra.Command, args []string) error {
			if err := common.Validate(); err != nil {
				return err
			}
			return runOverviewGPUTasks(c.Context(), f)
		},
	}
	return cmd
}

func runOverviewGPUTasks(ctx context.Context, f *cmdutil.Factory) error {
	c, err := buildDashboardClient(ctx, f)
	if err != nil {
		return err
	}
	now := time.Now()
	advisoryNote, _ := gpuAdvisory(ctx, c)
	list, err := fetchTaskList(ctx, c, nil)
	env := Envelope{Kind: KindOverviewGPUTasks, Meta: NewMeta(now.In(common.Timezone.Time()), c.OlaresID(), common.User)}
	if advisoryNote != "" {
		env.Meta.Note = advisoryNote
	}
	if err != nil {
		if he, ok := IsHTTPError(err); ok && he.Status == http.StatusNotFound {
			env.Meta.Empty = true
			env.Meta.EmptyReason = "no_vgpu_integration"
			env.Meta.HTTPStatus = he.Status
			if common.Output == OutputJSON {
				return WriteJSON(os.Stdout, env)
			}
			fmt.Fprintln(os.Stdout, "(no vGPU tasks)")
			return nil
		}
		if unavail, ok := vgpuUnavailableFromError(c, err, KindOverviewGPUTasks, now); ok {
			if advisoryNote != "" {
				unavail.Meta.Note = advisoryNote + " | " + unavail.Meta.Note
			}
			if common.Output == OutputJSON {
				return WriteJSON(os.Stdout, unavail)
			}
			return nil
		}
		return err
	}
	if len(list) == 0 {
		env.Meta.Empty = true
		env.Meta.EmptyReason = "no_gpu_detected"
		if common.Output == OutputJSON {
			return WriteJSON(os.Stdout, env)
		}
		fmt.Fprintln(os.Stdout, "(no vGPU tasks)")
		return nil
	}
	// HAMI's task entries match the SPA's `TaskItem` interface. The
	// "core util / mem used" columns are arrays (one element per
	// allocated device) — SPA uses index 0 too. Raw envelope retains
	// the full array so multi-GPU tasks aren't silently truncated.
	for _, t := range list {
		raw := map[string]any{}
		for k, v := range t {
			raw[k] = v
		}
		shareModeFirst := firstAnyInArray(t["deviceShareModes"])
		coreUtilFirst := firstAnyInArray(t["devicesCoreUtilizedPercent"])
		memMiBFirst := firstAnyInArray(t["devicesMemUtilized"])
		disp := map[string]any{
			"task_name": fmt.Sprintf("%v", t["name"]),
			"status":    fmt.Sprintf("%v", t["status"]),
			"mode":      gpuModeLabel(shareModeFirst),
			"host_node": fmt.Sprintf("%v", t["nodeName"]),
			"core_util": percentDirect(toFloat(coreUtilFirst)),
			"mem_used":  gpuVRAMHuman(memMiBFirst),
			"pod_uid":   fmt.Sprintf("%v", t["podUid"]),
		}
		env.Items = append(env.Items, Item{Raw: raw, Display: disp})
	}
	if common.Output == OutputJSON {
		return WriteJSON(os.Stdout, env)
	}
	cols := []TableColumn{
		{Header: "TASK", Get: func(it Item) string { return DisplayString(it, "task_name") }},
		{Header: "STATUS", Get: func(it Item) string { return DisplayString(it, "status") }},
		{Header: "MODE", Get: func(it Item) string { return DisplayString(it, "mode") }},
		{Header: "HOST", Get: func(it Item) string { return DisplayString(it, "host_node") }},
		{Header: "CORE_UTIL", Get: func(it Item) string { return DisplayString(it, "core_util") }},
		{Header: "MEM", Get: func(it Item) string { return DisplayString(it, "mem_used") }},
		{Header: "POD_UID", Get: func(it Item) string { return DisplayString(it, "pod_uid") }},
	}
	return WriteTable(os.Stdout, cols, env.Items)
}

func newOverviewGPUGetCommand(f *cmdutil.Factory) *cobra.Command {
	cmd := &cobra.Command{
		Use:     "get <uuid>",
		Short:   "Per-GPU detail by UUID",
		Args:    cobra.ExactArgs(1),
		SilenceErrors: true,
		SilenceUsage:  true,
		RunE: func(c *cobra.Command, args []string) error {
			if err := common.Validate(); err != nil {
				return err
			}
			return runOverviewGPUGet(c.Context(), f, args[0])
		},
	}
	return cmd
}

func runOverviewGPUGet(ctx context.Context, f *cmdutil.Factory, uuid string) error {
	c, err := buildDashboardClient(ctx, f)
	if err != nil {
		return err
	}
	now := time.Now()
	advisoryNote, _ := gpuAdvisory(ctx, c)
	detail, err := fetchGraphicsDetail(ctx, c, uuid)
	env := Envelope{
		Kind: KindOverviewGPUDetail,
		Meta: NewMeta(now.In(common.Timezone.Time()), c.OlaresID(), common.User),
	}
	if advisoryNote != "" {
		env.Meta.Note = advisoryNote
	}
	if err != nil {
		if he, ok := IsHTTPError(err); ok && he.Status == http.StatusNotFound {
			env.Meta.Empty = true
			env.Meta.EmptyReason = "no_vgpu_integration"
			env.Meta.HTTPStatus = he.Status
			if common.Output == OutputJSON {
				return WriteJSON(os.Stdout, env)
			}
			fmt.Fprintln(os.Stdout, "(GPU not found — HAMI integration absent or UUID invalid)")
			return nil
		}
		if unavail, ok := vgpuUnavailableFromError(c, err, KindOverviewGPUDetail, now); ok {
			if advisoryNote != "" {
				unavail.Meta.Note = advisoryNote + " | " + unavail.Meta.Note
			}
			if common.Output == OutputJSON {
				return WriteJSON(os.Stdout, unavail)
			}
			return nil
		}
		return err
	}
	if len(detail) == 0 {
		env.Meta.Empty = true
		env.Meta.EmptyReason = "no_gpu_detected"
		if common.Output == OutputJSON {
			return WriteJSON(os.Stdout, env)
		}
		fmt.Fprintln(os.Stdout, "(no detail returned for this GPU UUID)")
		return nil
	}
	env.Items = []Item{{
		Raw:     detail,
		Display: detail,
	}}
	if common.Output == OutputJSON {
		return WriteJSON(os.Stdout, env)
	}
	return EmitDefault(env, common.Output)
}

func newOverviewGPUTaskCommand(f *cmdutil.Factory) *cobra.Command {
	var sharemode string
	cmd := &cobra.Command{
		Use:     "task <name> <pod-uid>",
		Short:   "Per-task detail (pod-uid from `dashboard applications pods`)",
		Args:    cobra.ExactArgs(2),
		SilenceErrors: true,
		SilenceUsage:  true,
		RunE: func(c *cobra.Command, args []string) error {
			if err := common.Validate(); err != nil {
				return err
			}
			return runOverviewGPUTask(c.Context(), f, args[0], args[1], sharemode)
		},
	}
	cmd.Flags().StringVar(&sharemode, "sharemode", "", "task share mode (passed to /v1/container?sharemode=)")
	return cmd
}

func runOverviewGPUTask(ctx context.Context, f *cmdutil.Factory, name, podUID, sharemode string) error {
	c, err := buildDashboardClient(ctx, f)
	if err != nil {
		return err
	}
	now := time.Now()
	advisoryNote, _ := gpuAdvisory(ctx, c)
	detail, err := fetchTaskDetail(ctx, c, name, podUID, sharemode)
	env := Envelope{
		Kind: KindOverviewGPUTaskDet,
		Meta: NewMeta(now.In(common.Timezone.Time()), c.OlaresID(), common.User),
	}
	if advisoryNote != "" {
		env.Meta.Note = advisoryNote
	}
	if err != nil {
		if he, ok := IsHTTPError(err); ok && he.Status == http.StatusNotFound {
			env.Meta.Empty = true
			env.Meta.EmptyReason = "no_vgpu_integration"
			env.Meta.HTTPStatus = he.Status
			if common.Output == OutputJSON {
				return WriteJSON(os.Stdout, env)
			}
			fmt.Fprintln(os.Stdout, "(task not found — HAMI integration absent or pod-uid invalid)")
			return nil
		}
		if unavail, ok := vgpuUnavailableFromError(c, err, KindOverviewGPUTaskDet, now); ok {
			if advisoryNote != "" {
				unavail.Meta.Note = advisoryNote + " | " + unavail.Meta.Note
			}
			if common.Output == OutputJSON {
				return WriteJSON(os.Stdout, unavail)
			}
			return nil
		}
		return err
	}
	if len(detail) == 0 {
		env.Meta.Empty = true
		env.Meta.EmptyReason = "no_gpu_detected"
		if common.Output == OutputJSON {
			return WriteJSON(os.Stdout, env)
		}
		fmt.Fprintln(os.Stdout, "(no detail returned for this task)")
		return nil
	}
	env.Items = []Item{{Raw: detail, Display: detail}}
	if common.Output == OutputJSON {
		return WriteJSON(os.Stdout, env)
	}
	return EmitDefault(env, common.Output)
}

// ----------------------------------------------------------------------------
// helpers reused only by overview leaves
// ----------------------------------------------------------------------------

func sampleFloat(s format.LastMonitoringSample) float64 {
	if s.Empty {
		return 0
	}
	v, err := strconv.ParseFloat(s.RawValue, 64)
	if err != nil {
		return 0
	}
	return v
}

func formatFloat(v float64) string {
	return strconv.FormatFloat(v, 'f', -1, 64)
}

func safeRatio(num, den float64) float64 {
	if den == 0 {
		return 0
	}
	return num / den
}

// formatRateAny coerces an arbitrary tx/rx value (could be number or string)
// to a SPA-style "X B/s" throughput line. The system-ifs payload returns
// strings on some Olares versions and numbers on others.
func formatRateAny(v any) string {
	if v == nil {
		return "-"
	}
	switch x := v.(type) {
	case string:
		if x == "" {
			return "-"
		}
		return format.GetThroughput(x)
	case float64:
		return format.GetThroughput(formatFloat(x))
	case int, int64, int32:
		return format.GetThroughput(fmt.Sprintf("%d", x))
	default:
		return fmt.Sprintf("%v", x)
	}
}

func toFloat(v any) float64 {
	switch x := v.(type) {
	case float64:
		return x
	case int:
		return float64(x)
	case int64:
		return float64(x)
	case string:
		f, err := strconv.ParseFloat(x, 64)
		if err != nil {
			return 0
		}
		return f
	default:
		return 0
	}
}
