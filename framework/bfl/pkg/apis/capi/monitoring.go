package capi

import (
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"
	"sort"
	"strconv"
	"strings"

	"bytetrade.io/web3os/bfl/pkg/apiserver/authn"

	"github.com/emicklei/go-restful/v3"
)

const (
	clusterMetricFilter = "cluster_cpu_usage|cluster_cpu_total|cluster_cpu_utilisation|cluster_memory_usage_wo_cache|cluster_memory_total|cluster_memory_utilisation|cluster_disk_size_usage|cluster_disk_size_capacity|cluster_disk_size_utilisation|cluster_pod_running_count|cluster_pod_quota$"
	userMetricFilter    = "user_cpu_usage|user_cpu_total|user_cpu_utilisation|user_memory_usage_wo_cache|user_memory_total|user_memory_utilisation|user_disk_size_usage|user_disk_size_capacity|user_disk_size_utilisation|user_pod_running_count|user_pod_count$"
	clusterTotalFilter  = "cluster_cpu_total|cluster_memory_total|cluster_disk_size_capacity$"
)

type metricResponse struct {
	Results []map[string]any `json:"results"`
}

func (h *handler) monitoringMetric(req *restful.Request, resp *restful.Response) {
	identity, err := authn.Authenticate(req.Request)
	if err != nil {
		writeJSON(resp, http.StatusUnauthorized, map[string]string{"message": err.Error()})
		return
	}
	metricType := req.PathParameter("type")
	query := cloneQuery(req.Request.URL.Query())
	admin, err := h.isAdmin(req.Request.Context(), identity.Username)
	if err != nil {
		writeUpstreamError(resp, err)
		return
	}
	if !allowedMetricType(metricType, identity.Username, admin) {
		writeJSON(resp, http.StatusForbidden, map[string]string{"message": "requested metrics are outside the user scope"})
		return
	}

	if metricType == "cluster" {
		query.Set("metrics_filter", clusterMetricFilter)
		var metrics metricResponse
		if err := h.getKSJSON(req.Request.Context(), identity.Token,
			"/kapis/monitoring.kubesphere.io/v1alpha3/cluster", query, &metrics); err != nil {
			writeUpstreamError(resp, err)
			return
		}
		renameMetrics(metrics.Results, "cluster_")
		writeJSON(resp, http.StatusOK, metrics)
		return
	}

	query.Set("metrics_filter", userMetricFilter)
	var userMetrics metricResponse
	userPath := "/kapis/monitoring.kubesphere.io/v1alpha3/users/" + url.PathEscape(metricType)
	if err := h.getKSJSON(req.Request.Context(), identity.Token, userPath, query, &userMetrics); err != nil {
		writeUpstreamError(resp, err)
		return
	}
	userMetrics.Results = collapseSplitMetrics(userMetrics.Results)
	renameMetrics(userMetrics.Results, "user_")

	if admin {
		clusterQuery := cloneQuery(query)
		clusterQuery.Set("metrics_filter", clusterTotalFilter)
		var clusterMetrics metricResponse
		if err := h.getKSJSON(req.Request.Context(), identity.Token,
			"/kapis/monitoring.kubesphere.io/v1alpha3/cluster", clusterQuery, &clusterMetrics); err != nil {
			writeUpstreamError(resp, err)
			return
		}
		replaceMetricTotals(userMetrics.Results, clusterMetrics.Results)
	}
	for _, metric := range userMetrics.Results {
		if metricName(metric) == "cluster_pod_count" {
			metric["metric_name"] = "cluster_pod_quota"
		}
	}
	userMetrics.Results = appendUtilisation(userMetrics.Results, "cluster_cpu_usage", "cluster_cpu_total", "cluster_cpu_utilisation")
	userMetrics.Results = appendUtilisation(userMetrics.Results, "cluster_memory_usage_wo_cache", "cluster_memory_total", "cluster_memory_utilisation")
	writeJSON(resp, http.StatusOK, userMetrics)
}

func allowedMetricType(metricType, username string, admin bool) bool {
	return admin || (metricType != "cluster" && metricType == username)
}

func cloneQuery(source url.Values) url.Values {
	result := make(url.Values, len(source))
	for key, values := range source {
		result[key] = append([]string(nil), values...)
	}
	return result
}

func renameMetrics(metrics []map[string]any, prefix string) {
	for _, metric := range metrics {
		name := metricName(metric)
		metric["metric_name"] = strings.Replace(name, prefix, "cluster_", 1)
	}
}

func metricName(metric map[string]any) string {
	name, _ := metric["metric_name"].(string)
	return name
}

func replaceMetricTotals(userMetrics, clusterMetrics []map[string]any) {
	for _, name := range []string{"cluster_cpu_total", "cluster_memory_total", "cluster_disk_size_capacity"} {
		userMetric := findMetric(userMetrics, name)
		clusterMetric := findMetric(clusterMetrics, name)
		if userMetric == nil || clusterMetric == nil {
			continue
		}
		if sample, ok := firstSeries(clusterMetric); ok {
			if target, ok := firstSeries(userMetric); ok {
				if values, exists := sample["values"]; exists {
					target["values"] = values
				}
				if value, exists := sample["value"]; exists {
					target["value"] = value
				}
			}
		}
	}
}

func appendUtilisation(metrics []map[string]any, usageName, totalName, resultName string) []map[string]any {
	usage := findMetric(metrics, usageName)
	total := findMetric(metrics, totalName)
	if usage == nil || total == nil {
		return metrics
	}
	copyMetric := deepCopyMap(usage)
	usageSeries, usageOK := firstSeries(copyMetric)
	totalSeries, totalOK := firstSeries(total)
	if !usageOK || !totalOK {
		return metrics
	}
	if values, ok := usageSeries["values"].([]any); ok {
		usageSeries["values"] = ratioRange(values, asArray(totalSeries["values"]))
	} else if value := asArray(usageSeries["value"]); len(value) >= 2 {
		usageSeries["value"] = ratioInstant(value, asArray(totalSeries["value"]))
	} else {
		return metrics
	}
	copyMetric["metric_name"] = resultName
	return append(metrics, copyMetric)
}

func ratioRange(usage, total []any) []any {
	totals := make(map[string]float64, len(total))
	for _, raw := range total {
		point := asArray(raw)
		if len(point) >= 2 {
			totals[fmt.Sprint(point[0])] = number(point[1])
		}
	}
	result := make([]any, 0, len(usage))
	for _, raw := range usage {
		point := asArray(raw)
		if len(point) < 2 {
			continue
		}
		denominator := totals[fmt.Sprint(point[0])]
		ratio := 0.0
		if denominator != 0 {
			ratio = number(point[1]) / denominator
		}
		result = append(result, []any{point[0], ratio})
	}
	return result
}

func ratioInstant(usage, total []any) []any {
	ratio := 0.0
	if len(total) >= 2 && number(total[1]) != 0 {
		ratio = number(usage[1]) / number(total[1])
	}
	return []any{usage[0], ratio}
}

func number(value any) float64 {
	switch typed := value.(type) {
	case float64:
		return typed
	case json.Number:
		result, _ := typed.Float64()
		return result
	case string:
		result, _ := strconv.ParseFloat(typed, 64)
		return result
	default:
		return 0
	}
}

func findMetric(metrics []map[string]any, name string) map[string]any {
	for _, metric := range metrics {
		if metricName(metric) == name {
			return metric
		}
	}
	return nil
}

func firstSeries(metric map[string]any) (map[string]any, bool) {
	data, ok := metric["data"].(map[string]any)
	if !ok {
		return nil, false
	}
	results := asArray(data["result"])
	if len(results) == 0 {
		return nil, false
	}
	series, ok := results[0].(map[string]any)
	return series, ok
}

func asArray(value any) []any {
	values, _ := value.([]any)
	return values
}

func deepCopyMap(source map[string]any) map[string]any {
	data, _ := json.Marshal(source)
	var target map[string]any
	_ = json.Unmarshal(data, &target)
	return target
}

var volatileMetricLabels = map[string]struct{}{
	"instance": {}, "pod": {}, "container": {}, "endpoint": {}, "uid": {}, "job": {},
}

func collapseSplitMetrics(metrics []map[string]any) []map[string]any {
	for _, metric := range metrics {
		data, ok := metric["data"].(map[string]any)
		if !ok {
			continue
		}
		results := asArray(data["result"])
		if len(results) < 2 {
			continue
		}
		groups := make(map[string][]map[string]any)
		order := make([]string, 0)
		for _, raw := range results {
			series, ok := raw.(map[string]any)
			if !ok {
				continue
			}
			identity := seriesIdentity(series)
			if _, exists := groups[identity]; !exists {
				order = append(order, identity)
			}
			groups[identity] = append(groups[identity], series)
		}
		merged := make([]any, 0, len(order))
		for _, identity := range order {
			merged = append(merged, mergeSeries(groups[identity]))
		}
		data["result"] = merged
	}
	return metrics
}

func seriesIdentity(series map[string]any) string {
	metric, _ := series["metric"].(map[string]any)
	keys := make([]string, 0, len(metric))
	for key := range metric {
		if _, volatile := volatileMetricLabels[key]; !volatile {
			keys = append(keys, key)
		}
	}
	sort.Strings(keys)
	var builder strings.Builder
	for _, key := range keys {
		builder.WriteString(key)
		builder.WriteByte('=')
		builder.WriteString(fmt.Sprint(metric[key]))
		builder.WriteByte('\x00')
	}
	return builder.String()
}

func mergeSeries(series []map[string]any) map[string]any {
	if len(series) == 1 {
		return series[0]
	}
	merged := deepCopyMap(series[0])
	cleanMetric := make(map[string]any)
	if metric, ok := merged["metric"].(map[string]any); ok {
		for key, value := range metric {
			if _, volatile := volatileMetricLabels[key]; !volatile {
				cleanMetric[key] = value
			}
		}
	}
	merged["metric"] = cleanMetric
	points := make(map[string][]any)
	var latest []any
	for _, item := range series {
		for _, raw := range asArray(item["values"]) {
			point := asArray(raw)
			if len(point) >= 2 {
				points[fmt.Sprint(point[0])] = point
			}
		}
		if value := asArray(item["value"]); len(value) >= 2 && (len(latest) == 0 || number(value[0]) > number(latest[0])) {
			latest = value
		}
	}
	if len(points) > 0 {
		keys := make([]string, 0, len(points))
		for key := range points {
			keys = append(keys, key)
		}
		sort.Slice(keys, func(i, j int) bool { return number(keys[i]) < number(keys[j]) })
		values := make([]any, 0, len(keys))
		for _, key := range keys {
			values = append(values, points[key])
		}
		merged["values"] = values
	}
	if len(latest) > 0 {
		merged["value"] = latest
	}
	return merged
}
