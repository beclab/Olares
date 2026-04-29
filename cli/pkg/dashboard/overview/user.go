package overview

import (
	"context"
	"fmt"
	"io"
	"os"
	"time"

	pkgdashboard "github.com/beclab/Olares/cli/pkg/dashboard"
	"github.com/beclab/Olares/cli/pkg/dashboard/format"
)

// RunUser is the cmd-side entry point for `dashboard overview user
// [<username>]`. target is the optional positional override (admin
// only); empty defers to cf.User which itself defaults to the active
// profile owner.
func RunUser(ctx context.Context, c *pkgdashboard.Client, cf *pkgdashboard.CommonFlags, target string) error {
	r := &pkgdashboard.Runner{
		Flags:       cf,
		Recommended: 60 * time.Second,
		Iter: func(ctx context.Context, iter int, now time.Time) (pkgdashboard.Envelope, error) {
			env, err := BuildUserEnvelope(ctx, c, cf, target, now)
			if err != nil {
				return env, err
			}
			env.Meta = pkgdashboard.NewMeta(now.In(cf.Timezone.Time()), c.OlaresID(), target)
			env.Meta.RecommendedPollSeconds = 60
			if cf.Output == pkgdashboard.OutputJSON {
				return env, nil
			}
			return env, WriteUserTable(os.Stdout, env)
		},
	}
	return r.Run(ctx)
}

// BuildUserEnvelope resolves the target user, fetches user-grain
// CPU / memory quota metrics, and returns the SPA-aligned 2-row
// envelope (one row per resource type). Surfaces a typed admin-
// required error from ResolveTargetUser without modifying it so the
// agent-facing diagnostic stays 1:1.
func BuildUserEnvelope(ctx context.Context, c *pkgdashboard.Client, cf *pkgdashboard.CommonFlags, target string, now time.Time) (pkgdashboard.Envelope, error) {
	user, err := pkgdashboard.ResolveTargetUser(ctx, c, target)
	if err != nil {
		return pkgdashboard.Envelope{Kind: pkgdashboard.KindOverviewUser}, err
	}
	metrics := []string{
		"user_cpu_total", "user_cpu_usage", "user_cpu_utilisation",
		"user_memory_total", "user_memory_usage_wo_cache", "user_memory_utilisation",
	}
	res, err := pkgdashboard.FetchUserMetric(ctx, c, cf, user.Name, metrics, pkgdashboard.DefaultClusterWindow(), now, false)
	if err != nil {
		return pkgdashboard.Envelope{Kind: pkgdashboard.KindOverviewUser, Meta: pkgdashboard.Meta{User: user.Name}}, err
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
	items := make([]pkgdashboard.Item, 0, len(rows))
	for _, r := range rows {
		items = append(items, pkgdashboard.Item{
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
	return pkgdashboard.Envelope{
		Kind:  pkgdashboard.KindOverviewUser,
		Meta:  pkgdashboard.NewMeta(now.In(cf.Timezone.Time()), c.OlaresID(), user.Name),
		Items: items,
	}, nil
}

// WriteUserTable renders env.Items as the SPA-aligned 4-column
// per-user resource table (METRIC / USED / TOTAL / UTIL).
func WriteUserTable(w io.Writer, env pkgdashboard.Envelope) error {
	cols := []pkgdashboard.TableColumn{
		{Header: "METRIC", Get: func(it pkgdashboard.Item) string { return pkgdashboard.DisplayString(it, "metric") }},
		{Header: "USED", Get: func(it pkgdashboard.Item) string { return pkgdashboard.DisplayString(it, "used") }},
		{Header: "TOTAL", Get: func(it pkgdashboard.Item) string { return pkgdashboard.DisplayString(it, "total") }},
		{Header: "UTIL", Get: func(it pkgdashboard.Item) string { return pkgdashboard.DisplayString(it, "utilisation") }},
	}
	return pkgdashboard.WriteTable(w, cols, env.Items)
}
