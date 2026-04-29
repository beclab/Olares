package disk

import (
	"context"
	"net/http"
	"os"
	"strings"
	"time"

	"github.com/spf13/cobra"

	"github.com/beclab/Olares/cli/pkg/cmdutil"
)

func newOverviewDiskPartitionsCommand(f *cmdutil.Factory) *cobra.Command {
	cmd := &cobra.Command{
		Use:           "partitions <device>",
		Short:         "Partition-level table for one physical device",
		Args:          cobra.ExactArgs(1),
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
		Flags:       common,
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
