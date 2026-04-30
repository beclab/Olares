package overview

import (
	"context"
	"net/http"
	"os"
	"sort"
	"time"

	pkgdashboard "github.com/beclab/Olares/cli/pkg/dashboard"
	"github.com/beclab/Olares/cli/pkg/dashboard/format"
)

// PerNodeDisplayFn renders one node's last-sample bucket into the
// (raw, display) maps a single Item carries. Each per-node leaf
// (cpu / memory / pods) supplies its own implementation; the
// scaffold below is metric-set-agnostic.
type PerNodeDisplayFn func(node string, last map[string]format.LastMonitoringSample) (raw, display map[string]any)

// RunPerNodeMetric is the shared workhorse for cpu / memory / pods.
// It owns the watch-aware Runner so the four per-node leaves stay
// pure data-shape declarations + a column list. The cmd-side leaves
// (cli/cmd/ctl/dashboard/overview/{cpu,memory,pods}.go) just forward
// here with their per-leaf metric set / column / display closures.
//
// kind is the envelope's Kind (KindOverviewCPU / KindOverviewMemory /
// KindOverviewPods); metrics is the /v1alpha3/nodes filter; cols is
// the table render schema; disp is the per-row Item builder.
func RunPerNodeMetric(
	ctx context.Context,
	c *pkgdashboard.Client,
	cf *pkgdashboard.CommonFlags,
	kind string,
	metrics []string,
	cols []pkgdashboard.TableColumn,
	disp PerNodeDisplayFn,
) error {
	r := &pkgdashboard.Runner{
		Flags:       cf,
		Recommended: 60 * time.Second,
		Iter: func(ctx context.Context, iter int, now time.Time) (pkgdashboard.Envelope, error) {
			env, err := BuildPerNodeEnvelope(ctx, c, cf, kind, metrics, disp, now)
			if err != nil {
				return env, err
			}
			env.Meta = pkgdashboard.NewMeta(now.In(cf.Timezone.Time()), c.OlaresID(), cf.User)
			env.Meta.RecommendedPollSeconds = 60
			if cf.Output == pkgdashboard.OutputJSON {
				return env, nil
			}
			return env, pkgdashboard.WriteTable(os.Stdout, cols, env.Items)
		},
	}
	return r.Run(ctx)
}

// BuildPerNodeEnvelope shells out to /v1alpha3/nodes and groups the
// results by the `node` label. Unlike FetchClusterMetrics (which
// collapses by metric_name), per-node metrics carry one row per
// node within each metric; we transpose into one Item per node.
//
// Sort order is alphabetical on the node label so the row index is
// deterministic for agents that scrape stdout. Empty `node` labels
// fall back to `instance`; rows missing both are dropped (rather
// than silently bucketed into "").
func BuildPerNodeEnvelope(
	ctx context.Context,
	c *pkgdashboard.Client,
	cf *pkgdashboard.CommonFlags,
	kind string,
	metrics []string,
	disp PerNodeDisplayFn,
	now time.Time,
) (pkgdashboard.Envelope, error) {
	q := pkgdashboard.MonitoringQuery(cf, metrics, pkgdashboard.DefaultDetailWindow(), now, false)
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
		return pkgdashboard.Envelope{Kind: kind}, err
	}
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
	items := make([]pkgdashboard.Item, 0, len(order))
	for _, n := range order {
		raws, disps := disp(n, buckets[n].samples)
		items = append(items, pkgdashboard.Item{Raw: raws, Display: disps})
	}
	return pkgdashboard.Envelope{
		Kind:  kind,
		Meta:  pkgdashboard.NewMeta(now.In(cf.Timezone.Time()), c.OlaresID(), cf.User),
		Items: items,
	}, nil
}
