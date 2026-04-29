package overview

import (
	"context"
	"fmt"
	"os"
	"time"

	"github.com/spf13/cobra"

	"github.com/beclab/Olares/cli/pkg/cmdutil"
)

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
		Flags:       common,
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

// buildRankingEnvelopeBy was hoisted to pkgdashboard.BuildRankingEnvelope
// (cross-area share with `applications`). The area-local trampoline of
// the same name lives in common.go.

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

// loadAppsForRanking was hoisted to pkgdashboard.LoadAppsForRanking (used
// internally by BuildRankingEnvelope).

// ----------------------------------------------------------------------------
// overview cpu / memory / pods — per-node multi-metric tables
// ----------------------------------------------------------------------------
