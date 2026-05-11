package applications

import (
	"context"
	"fmt"
	"os"
	"time"

	"github.com/spf13/cobra"

	"github.com/beclab/Olares/cli/pkg/cmdutil"
	pkgdashboard "github.com/beclab/Olares/cli/pkg/dashboard"
)

// ----------------------------------------------------------------------------
// `dashboard applications` — workload-grain table (single leaf, no subverbs).
// ----------------------------------------------------------------------------
//
// Default action: workload-grain table — same data source as `overview
// ranking` (fetchWorkloadsMetrics) with the `state` and `pods` columns
// added. The SPA's Applications2/IndexPage renders this view; we mirror
// the row shape so consumers can join `applications` and `overview ranking`
// on the (app, namespace) tuple.
//
// The deprecated `applications list / users / containers / pods` leaves are
// gone: `list` is now the default action; `users` was admin-only and
// rarely useful (the same data shows up in `overview user --user`);
// `containers` was a stub that nobody depended on; `pods` duplicated
// `kubectl get pods -n <ns>` and never carried first-class agent
// semantics.

func NewApplicationsCommand(f *cmdutil.Factory, cf *pkgdashboard.CommonFlags) *cobra.Command {
	common = cf
	var sortDir, sortBy string
	cmd := &cobra.Command{
		Use:           "applications",
		Aliases:       []string{"apps"},
		Short:         "Workload-grain application table (mirrors the SPA's Applications page)",
		SilenceErrors: true,
		SilenceUsage:  true,
		RunE: func(c *cobra.Command, args []string) error {
			if len(args) > 0 {
				return unknownSubcommandRunE(c, args)
			}
			if err := common.Validate(); err != nil {
				return err
			}
			return runApplicationsList(c.Context(), f, sortBy, sortDir)
		},
	}
	cmd.Flags().StringVar(&sortDir, "sort", "desc", "sort direction (asc or desc)")
	cmd.Flags().StringVar(&sortBy, "sort-by", "cpu", "sort key: cpu | memory | net_in | net_out")
	return cmd
}

// runApplicationsList is the default workload-grain view. Reuses
// buildRankingEnvelopeBy's wire path (fetchWorkloadsMetrics) so the first
// row of `applications` and `overview ranking` match by construction. The
// envelope kind is re-tagged so consumers can demux. State / pod-count
// already ride along inside the ranking envelope (sourced from the
// AppListItem + namespace_pod_count metric); no extra round-trip is
// needed.
func runApplicationsList(ctx context.Context, f *cmdutil.Factory, sortBy, sortDir string) error {
	if sortDir != "asc" && sortDir != "desc" {
		return fmt.Errorf("--sort: %q is not asc/desc", sortDir)
	}
	switch sortBy {
	case "cpu", "memory", "net_in", "net_out":
	default:
		return fmt.Errorf("--sort-by: %q is not cpu|memory|net_in|net_out", sortBy)
	}
	c, err := buildDashboardClient(ctx, f)
	if err != nil {
		return err
	}
	r := &Runner{
		Flags:       common,
		Recommended: 60 * time.Second,
		Iter: func(ctx context.Context, iter int, now time.Time) (Envelope, error) {
			rankEnv, err := buildRankingEnvelopeBy(ctx, c, common.User, sortBy, sortDir, now)
			if err != nil {
				return rankEnv, err
			}
			env := Envelope{
				Kind:  KindApplicationsList,
				Meta:  NewMeta(now.In(common.Timezone.Time()), c.OlaresID(), common.User),
				Items: rankEnv.Items,
			}
			env.Meta.RecommendedPollSeconds = 60
			env.Items = HeadItems(env.Items, common.Head)
			if common.Output == OutputJSON {
				return env, nil
			}
			return env, writeApplicationsListTable(env)
		},
	}
	return r.Run(ctx)
}

func writeApplicationsListTable(env Envelope) error {
	cols := []TableColumn{
		{Header: "RANK", Get: func(it Item) string { return DisplayString(it, "rank") }},
		{Header: "APP", Get: func(it Item) string { return DisplayString(it, "app") }},
		{Header: "NAMESPACE", Get: func(it Item) string { return DisplayString(it, "namespace") }},
		{Header: "STATE", Get: func(it Item) string { return DisplayString(it, "state") }},
		{Header: "PODS", Get: func(it Item) string { return DisplayString(it, "pods") }},
		{Header: "CPU", Get: func(it Item) string { return DisplayString(it, "cpu") }},
		{Header: "MEMORY", Get: func(it Item) string { return DisplayString(it, "memory") }},
		{Header: "NET_IN", Get: func(it Item) string { return DisplayString(it, "net_in") }},
		{Header: "NET_OUT", Get: func(it Item) string { return DisplayString(it, "net_out") }},
	}
	return WriteTable(os.Stdout, cols, env.Items)
}
