package overview

import (
	"context"
	"fmt"
	"os"
	"time"

	"github.com/spf13/cobra"

	"github.com/beclab/Olares/cli/pkg/cmdutil"
	"github.com/beclab/Olares/cli/pkg/dashboard/format"
)

func newOverviewUserCommand(f *cmdutil.Factory) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "user [<username>]",
		Short: "User-grain CPU / memory quota usage (mirrors the SPA's User Resources panel)",
		Example: `  olares-cli dashboard overview user
  olares-cli dashboard overview user alice    # admin only`,
		Args:          cobra.MaximumNArgs(1),
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
		Flags:       common,
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
