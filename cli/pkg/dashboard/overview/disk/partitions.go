package disk

import (
	"context"
	"io"
	"net/http"
	"os"
	"strings"
	"time"

	pkgdashboard "github.com/beclab/Olares/cli/pkg/dashboard"
)

// RunPartitions is the cmd-side entry point for `dashboard overview
// disk partitions <device>`. The leaf is watch-aware so the
// scrolling lsblk view can pin a long-running shell against a
// rapidly changing partition table; cmd-side stays a one-liner.
func RunPartitions(ctx context.Context, c *pkgdashboard.Client, cf *pkgdashboard.CommonFlags, device string) error {
	r := &pkgdashboard.Runner{
		Flags:       cf,
		Recommended: 60 * time.Second,
		Iter: func(ctx context.Context, iter int, now time.Time) (pkgdashboard.Envelope, error) {
			env, err := BuildPartitionsEnvelope(ctx, c, cf, device, now)
			if err != nil {
				return env, err
			}
			env.Meta = pkgdashboard.NewMeta(now.In(cf.Timezone.Time()), c.OlaresID(), cf.User)
			env.Meta.RecommendedPollSeconds = 60
			if cf.Output == pkgdashboard.OutputJSON {
				return env, nil
			}
			return env, WritePartitionsTable(os.Stdout, env)
		},
	}
	return r.Run(ctx)
}

// BuildPartitionsEnvelope mirrors SPA Overview2/Disk/IndexPage's
// "Occupancy Analysis" popup (`getDiskPartitionRows` in
// Overview2/Disk/config.ts:398).
//
// It pulls `node_disk_lsblk_info` (instant), filters to the rows
// whose `metric.node` == <node> (the SPA uses the node selected via
// the device card); when `node` is empty the CLI auto-resolves it
// by scanning the SMART rows for a row matching the requested
// device.
//
// The candidate subset for the requested root `device` is then
// chosen per the SPA logic:
//
//  1. if any row in the node's lsblk set has a `pkname` label →
//     CollectSubtreeByPkname BFS from `<device>`.
//  2. otherwise fall back to `name == device || name.startsWith(device)`.
//
// Finally FlattenLsblkHierarchy walks the tree pre-order and
// computes the ASCII tree prefix per row (`├── `, `└── `, `│   `,
// ...). Per the user's decision the Display column for `name`
// carries that tree prefix so humans see the topology at a glance.
func BuildPartitionsEnvelope(ctx context.Context, c *pkgdashboard.Client, cf *pkgdashboard.CommonFlags, device string, now time.Time) (pkgdashboard.Envelope, error) {
	q := pkgdashboard.MonitoringQuery(cf, []string{"node_disk_lsblk_info"}, pkgdashboard.DefaultDetailWindow(), now, true)
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
		return pkgdashboard.Envelope{Kind: pkgdashboard.KindOverviewDiskPart}, err
	}

	// Find the node hosting <device>. The SPA pulls this from the
	// SMART card the user clicked; the CLI infers it from any row
	// whose name == <device>. If none, fall through with node="" and
	// match all rows; that mirrors SPA's degraded path when the
	// popup is opened without a node selection.
	var node string
	allRows := []pkgdashboard.LsblkRow{}
	for _, r := range raw.Results {
		if r.MetricName != "node_disk_lsblk_info" {
			continue
		}
		for _, e := range r.Data.Result {
			if e.Metric["name"] == device && node == "" {
				node = e.Metric["node"]
			}
			allRows = append(allRows, pkgdashboard.LsblkRow{
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

	scoped := make([]pkgdashboard.LsblkRow, 0, len(allRows))
	for _, r := range allRows {
		if node == "" || r.Node == node {
			scoped = append(scoped, r)
		}
	}

	var subset []pkgdashboard.LsblkRow
	if pkgdashboard.HasPknameLabels(scoped) {
		subset = pkgdashboard.CollectSubtreeByPkname(scoped, device)
	} else {
		subset = subset[:0]
		for _, r := range scoped {
			if r.Name == device || strings.HasPrefix(r.Name, device) {
				subset = append(subset, r)
			}
		}
	}
	flat := pkgdashboard.FlattenLsblkHierarchy(subset, device)

	items := make([]pkgdashboard.Item, 0, len(flat))
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
		items = append(items, pkgdashboard.Item{Raw: raw, Display: disp})
	}
	return pkgdashboard.Envelope{
		Kind:  pkgdashboard.KindOverviewDiskPart,
		Meta:  pkgdashboard.NewMeta(now.In(cf.Timezone.Time()), c.OlaresID(), cf.User),
		Items: items,
	}, nil
}

// orDash mirrors `displayLsblkCell` in Overview2/Disk/config.ts:30 —
// trim → empty falls through to "-". Private — partitions-only.
func orDash(v string) string {
	t := strings.TrimSpace(v)
	if t == "" {
		return "-"
	}
	return t
}

// WritePartitionsTable renders the lsblk subtree. NAME carries the
// ASCII tree prefix; the other columns are the SPA's
// `getLsblkColumns`.
func WritePartitionsTable(w io.Writer, env pkgdashboard.Envelope) error {
	cols := []pkgdashboard.TableColumn{
		{Header: "NAME", Get: func(it pkgdashboard.Item) string { return pkgdashboard.DisplayString(it, "name") }},
		{Header: "SIZE", Get: func(it pkgdashboard.Item) string { return pkgdashboard.DisplayString(it, "size") }},
		{Header: "FSTYPE", Get: func(it pkgdashboard.Item) string { return pkgdashboard.DisplayString(it, "fstype") }},
		{Header: "MOUNT", Get: func(it pkgdashboard.Item) string { return pkgdashboard.DisplayString(it, "mountpoint") }},
		{Header: "FSUSED", Get: func(it pkgdashboard.Item) string { return pkgdashboard.DisplayString(it, "fsused") }},
		{Header: "FSUSE%", Get: func(it pkgdashboard.Item) string { return pkgdashboard.DisplayString(it, "fsuse_percent") }},
	}
	return pkgdashboard.WriteTable(w, cols, env.Items)
}
