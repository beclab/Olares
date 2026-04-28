package dashboard

import (
	"context"
	"errors"
	"fmt"
	"net/http"
	"net/url"
	"os"
	"strconv"
	"strings"
	"time"

	"github.com/spf13/cobra"

	"github.com/beclab/Olares/cli/cmd/ctl/dashboard/format"
	"github.com/beclab/Olares/cli/pkg/cmdutil"
)

// ----------------------------------------------------------------------------
// `dashboard applications` — workload-grain table + per-namespace pods leaf.
// ----------------------------------------------------------------------------
//
// Default action: workload-grain table — same data source as `overview
// ranking` (fetchWorkloadsMetrics) with the `state` and `pods` columns
// added. The SPA's Applications2/IndexPage renders this view; we mirror
// the row shape so consumers can join `applications` and `overview ranking`
// on the (app, namespace) tuple.
//
// Subcommands:
//
//	applications pods <namespace>  — per-pod table for one namespace
//
// The deprecated `applications list / users / containers` leaves are gone:
// `list` is now the default action; `users` was admin-only and rarely
// useful (the same data shows up in `overview user --user`); `containers`
// was a stub that nobody depended on.

func newApplicationsCommand(f *cmdutil.Factory) *cobra.Command {
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
	cmd.AddCommand(newApplicationsPodsCommand(f))
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
		Flags:       &common,
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

// ----------------------------------------------------------------------------
// applications pods <namespace>
// ----------------------------------------------------------------------------

func newApplicationsPodsCommand(f *cmdutil.Factory) *cobra.Command {
	cmd := &cobra.Command{
		Use:     "pods <namespace>",
		Short:   "Per-pod table for a single namespace",
		Example: `  olares-cli dashboard applications pods user-system-alice -o json`,
		Args:    cobra.ExactArgs(1),
		SilenceErrors: true,
		SilenceUsage:  true,
		RunE: func(c *cobra.Command, args []string) error {
			if err := common.Validate(); err != nil {
				return err
			}
			return runApplicationsPods(c.Context(), f, args[0])
		},
	}
	return cmd
}

func runApplicationsPods(ctx context.Context, f *cmdutil.Factory, namespace string) error {
	if strings.TrimSpace(namespace) == "" {
		return errors.New("namespace argument is required")
	}
	c, err := buildDashboardClient(ctx, f)
	if err != nil {
		return err
	}
	r := &Runner{
		Flags:       &common,
		Recommended: 60 * time.Second,
		Iter: func(ctx context.Context, iter int, now time.Time) (Envelope, error) {
			env, err := buildPodsEnvelope(ctx, c, namespace, now)
			if err != nil {
				return env, err
			}
			env.Meta = NewMeta(now.In(common.Timezone.Time()), c.OlaresID(), common.User)
			env.Meta.RecommendedPollSeconds = 60
			env.Items = HeadItems(env.Items, common.Head)
			if common.Output == OutputJSON {
				return env, nil
			}
			return env, writePodsTable(env)
		},
	}
	return r.Run(ctx)
}

func buildPodsEnvelope(ctx context.Context, c *Client, namespace string, now time.Time) (Envelope, error) {
	q := url.Values{}
	if common.Limit > 0 {
		q.Set("limit", strconv.Itoa(common.Limit))
	}
	if common.Page > 0 {
		q.Set("page", strconv.Itoa(common.Page))
	}
	var raw struct {
		Items []struct {
			Metadata struct {
				Name              string `json:"name"`
				Namespace         string `json:"namespace"`
				CreationTimestamp string `json:"creationTimestamp"`
				UID               string `json:"uid"`
			} `json:"metadata"`
			Spec struct {
				NodeName   string `json:"nodeName"`
				Containers []struct {
					Name      string `json:"name"`
					Resources struct {
						Limits   map[string]string `json:"limits"`
						Requests map[string]string `json:"requests"`
					} `json:"resources"`
				} `json:"containers"`
			} `json:"spec"`
			Status struct {
				Phase             string `json:"phase"`
				ContainerStatuses []struct {
					RestartCount int `json:"restartCount"`
				} `json:"containerStatuses"`
			} `json:"status"`
		} `json:"items"`
	}
	path := fmt.Sprintf("/kapis/resources.kubesphere.io/v1alpha3/namespaces/%s/pods", url.PathEscape(namespace))
	if err := c.DoJSON(ctx, http.MethodGet, path, q, nil, &raw); err != nil {
		return Envelope{Kind: KindApplicationsPods}, err
	}
	items := make([]Item, 0, len(raw.Items))
	for _, p := range raw.Items {
		restarts := 0
		for _, cs := range p.Status.ContainerStatuses {
			restarts += cs.RestartCount
		}
		items = append(items, Item{
			Raw: map[string]any{
				"name":         p.Metadata.Name,
				"namespace":    p.Metadata.Namespace,
				"node":         p.Spec.NodeName,
				"phase":        p.Status.Phase,
				"created_at":   p.Metadata.CreationTimestamp,
				"uid":          p.Metadata.UID,
				"restart_sum":  restarts,
				"container_n":  len(p.Spec.Containers),
			},
			Display: map[string]any{
				"name":      p.Metadata.Name,
				"namespace": p.Metadata.Namespace,
				"node":      p.Spec.NodeName,
				"phase":     p.Status.Phase,
				"created":   format.FormatTime(parseRFCTimestamp(p.Metadata.CreationTimestamp), false, common.Timezone),
				"uid":       p.Metadata.UID,
				"restarts":  strconv.Itoa(restarts),
			},
		})
	}
	return Envelope{
		Kind:  KindApplicationsPods,
		Meta:  NewMeta(now.In(common.Timezone.Time()), c.OlaresID(), common.User),
		Items: items,
	}, nil
}

func writePodsTable(env Envelope) error {
	cols := []TableColumn{
		{Header: "NAME", Get: func(it Item) string { return DisplayString(it, "name") }},
		{Header: "NODE", Get: func(it Item) string { return DisplayString(it, "node") }},
		{Header: "PHASE", Get: func(it Item) string { return DisplayString(it, "phase") }},
		{Header: "RESTARTS", Get: func(it Item) string { return DisplayString(it, "restarts") }},
		{Header: "CREATED", Get: func(it Item) string { return DisplayString(it, "created") }},
		{Header: "UID", Get: func(it Item) string { return DisplayString(it, "uid") }},
	}
	return WriteTable(os.Stdout, cols, env.Items)
}

// ----------------------------------------------------------------------------
// Misc helpers
// ----------------------------------------------------------------------------

// parseRFCTimestamp converts an RFC3339 timestamp string to milliseconds
// since epoch (the unit format.FormatTime expects). 0 on parse failure.
func parseRFCTimestamp(s string) int64 {
	t, err := time.Parse(time.RFC3339, s)
	if err != nil {
		return 0
	}
	return t.UnixMilli()
}

// preserve the cmdutil import in case future leaves bypass buildDashboardClient.
var _ = cmdutil.Factory{}
