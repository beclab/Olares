package dashboard

import (
	"context"
	"errors"
	"fmt"
	"net/http"
	"net/url"
	"sort"
	"strings"
	"time"

	"github.com/beclab/Olares/cli/pkg/dashboard/format"
)

// MonitoringWindow encapsulates the time-window flags (--since / --start /
// --end) and the SPA's default fallback. The SPA's
// controlPanelCommon/containers/Monitoring/config.ts uses step=600s,
// times=20 for the cluster headline and step=60s for detail pages; we mirror
// the cluster default and let leaves override.
type MonitoringWindow struct {
	Step       time.Duration // upstream "step" param (Prometheus-flavoured)
	Times      int           // upstream "times" param (sample count)
	DefaultDur time.Duration // fallback when --since / --start are unset
}

// DefaultClusterWindow is the headline cluster window. 600s × 20 = 200min,
// matching the SPA's `getParams({})` defaults for the overview page.
func DefaultClusterWindow() MonitoringWindow {
	return MonitoringWindow{
		Step:       600 * time.Second,
		Times:      20,
		DefaultDur: 200 * time.Minute,
	}
}

// DefaultDetailWindow is the per-detail-page sliding window. 60s × 50 = 50min,
// matching the SPA's CPU/Memory/Pods detail page `timeRangeDefault`.
func DefaultDetailWindow() MonitoringWindow {
	return MonitoringWindow{
		Step:       60 * time.Second,
		Times:      50,
		DefaultDur: 50 * time.Minute,
	}
}

// MonitoringQuery composes the URL query for a /cluster or /nodes monitoring
// fetch. Caller passes `cf` so the helper honours --since / --start / --end
// without reaching into a global; everything else (step, times, instant)
// defaults to the SPA contract.
func MonitoringQuery(cf *CommonFlags, metricsFilter []string, w MonitoringWindow, now time.Time, instant bool) url.Values {
	v := url.Values{}
	v.Set("metrics_filter", strings.Join(metricsFilter, "|")+"$")
	if instant {
		// "step=0s" is the SPA's `step: '0s'` shorthand for "give me the
		// instantaneous reading" (used by Disk detail). The BFF treats
		// it as a value-only query and skips the values[] array.
		v.Set("step", "0s")
		return v
	}
	if cf != nil && cf.HasAbsoluteWindow() {
		v.Set("start", fmt.Sprintf("%d", cf.Start.Unix()))
		v.Set("end", fmt.Sprintf("%d", cf.End.Unix()))
	} else {
		end := now
		dur := time.Duration(0)
		if cf != nil {
			dur = cf.Since
		}
		if dur == 0 {
			dur = w.DefaultDur
		}
		start := end.Add(-dur)
		v.Set("start", fmt.Sprintf("%d", start.Unix()))
		v.Set("end", fmt.Sprintf("%d", end.Unix()))
	}
	v.Set("step", fmt.Sprintf("%ds", int(w.Step.Seconds())))
	v.Set("times", fmt.Sprintf("%d", w.Times))
	return v
}

// FetchClusterMetrics issues GET /kapis/monitoring.kubesphere.io/v1alpha3/cluster
// with metrics_filter and returns the per-metric raw payload (one entry per
// `metric_name`). Used by overview physical / overview ranking sections.
//
// Wire shape (from the SPA's getClusterMonitoring):
//
//	GET /kapis/monitoring.kubesphere.io/v1alpha3/cluster?metrics_filter=...$&start=...&end=...&step=600s&times=20
//
// Response: { results: [ { metric_name, data: { result: [ { metric, values, value } ] } } ] }
func FetchClusterMetrics(ctx context.Context, c *Client, cf *CommonFlags, metrics []string, w MonitoringWindow, now time.Time, instant bool) (map[string]format.MonitoringResult, error) {
	q := MonitoringQuery(cf, metrics, w, now, instant)
	return DoMonitoring(ctx, c, "/kapis/monitoring.kubesphere.io/v1alpha3/cluster", q)
}

// FetchNodeMetrics issues GET /kapis/.../v1alpha3/nodes — used by every
// per-node detail page (CPU / Memory / Disk / Pods / Fan-via-monitoring,
// when applicable). Response shape is identical to FetchClusterMetrics.
func FetchNodeMetrics(ctx context.Context, c *Client, cf *CommonFlags, metrics []string, w MonitoringWindow, now time.Time, instant bool) (map[string]format.MonitoringResult, error) {
	q := MonitoringQuery(cf, metrics, w, now, instant)
	return DoMonitoring(ctx, c, "/kapis/monitoring.kubesphere.io/v1alpha3/nodes", q)
}

// FetchUserMetric issues GET /kapis/.../v1alpha3/users/<username> — used by
// overview user. Same monitoring-response shape as cluster / nodes.
func FetchUserMetric(ctx context.Context, c *Client, cf *CommonFlags, username string, metrics []string, w MonitoringWindow, now time.Time, instant bool) (map[string]format.MonitoringResult, error) {
	if username == "" {
		return nil, errors.New("FetchUserMetric: username is required")
	}
	q := MonitoringQuery(cf, metrics, w, now, instant)
	path := fmt.Sprintf("/kapis/monitoring.kubesphere.io/v1alpha3/users/%s", url.PathEscape(username))
	return DoMonitoring(ctx, c, path, q)
}

// DoMonitoring shells out to the BFF, decodes the standard
// `{ results: [...] }` envelope, and returns a metric_name → MonitoringResult
// map ready for format.GetLastMonitoringData.
func DoMonitoring(ctx context.Context, c *Client, path string, q url.Values) (map[string]format.MonitoringResult, error) {
	var raw struct {
		Results []struct {
			MetricName string                 `json:"metric_name"`
			Data       map[string]interface{} `json:"data"`
		} `json:"results"`
	}
	if err := c.DoJSON(ctx, http.MethodGet, path, q, nil, &raw); err != nil {
		return nil, err
	}
	out := make(map[string]format.MonitoringResult, len(raw.Results))
	for _, r := range raw.Results {
		b, err := jsonMarshal(map[string]interface{}{"data": r.Data})
		if err != nil {
			continue
		}
		var mr format.MonitoringResult
		if err := jsonUnmarshal(b, &mr); err != nil {
			continue
		}
		out[r.MetricName] = mr
	}
	return out, nil
}

// FetchNodesList returns the sorted set of node names. Used by overview
// physical (to label cluster-level rows with hostnames).
func FetchNodesList(ctx context.Context, c *Client) ([]string, error) {
	var raw struct {
		Items []struct {
			Metadata struct {
				Name string `json:"name"`
			} `json:"metadata"`
		} `json:"items"`
	}
	q := url.Values{"sortBy": []string{"createTime"}}
	if err := c.DoJSON(ctx, http.MethodGet, "/kapis/resources.kubesphere.io/v1alpha3/nodes", q, nil, &raw); err != nil {
		return nil, err
	}
	out := make([]string, 0, len(raw.Items))
	for _, it := range raw.Items {
		out = append(out, it.Metadata.Name)
	}
	sort.Strings(out)
	return out, nil
}
