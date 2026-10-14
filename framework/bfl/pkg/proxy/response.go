package proxy

import (
	"bytes"
	"compress/gzip"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"sort"
	"strconv"
	"strings"

	"bytetrade.io/web3os/bfl/pkg/apiserver/authn"
)

const (
	resourcesNamespacesPath = "/kapis/resources.kubesphere.io/v1alpha3/namespaces"
	resourcesPodsPath       = "/kapis/resources.kubesphere.io/v1alpha3/pods"
	monitoringNamespaces    = "/kapis/monitoring.kubesphere.io/v1alpha3/namespaces"
)

var namespacedListPaths = map[string]struct{}{
	resourcesPodsPath: {},
	"/kapis/resources.kubesphere.io/v1alpha3/networkpolicies":        {},
	"/kapis/resources.kubesphere.io/v1alpha3/persistentvolumeclaims": {},
	"/kapis/resources.kubesphere.io/v1alpha3/cronjobs":               {},
	"/kapis/resources.kubesphere.io/v1alpha3/jobs":                   {},
}

func (h *Handler) modifyResponse(resp *http.Response) error {
	admin, _ := resp.Request.Context().Value(adminKey).(bool)
	if admin || resp.StatusCode < http.StatusOK || resp.StatusCode >= http.StatusMultipleChoices {
		return nil
	}
	kind := responseFilter(resp.Request.URL.Path)
	if kind == filterNone {
		return nil
	}

	raw, err := readBounded(resp.Body, h.maxResponseBytes)
	if err != nil {
		return err
	}
	_ = resp.Body.Close()
	decoded, compressed, err := decodeResponse(raw, resp.Header.Get("Content-Encoding"))
	if err != nil {
		return err
	}
	if int64(len(decoded)) > h.maxResponseBytes {
		return fmt.Errorf("decoded protected kubesphere response exceeds %d bytes", h.maxResponseBytes)
	}
	var payload any
	if err := json.Unmarshal(decoded, &payload); err != nil {
		return fmt.Errorf("decode protected kubesphere response: %w", err)
	}
	identity, _ := resp.Request.Context().Value(identityKey).(authn.Identity)
	payload, err = filterResponse(kind, payload, identity.Username)
	if err != nil {
		return err
	}
	encoded, err := json.Marshal(payload)
	if err != nil {
		return fmt.Errorf("encode protected kubesphere response: %w", err)
	}
	if compressed {
		encoded, err = gzipBytes(encoded)
		if err != nil {
			return err
		}
	}
	resp.Body = io.NopCloser(bytes.NewReader(encoded))
	resp.ContentLength = int64(len(encoded))
	resp.Header.Set("Content-Length", strconv.Itoa(len(encoded)))
	resp.Header.Del("Transfer-Encoding")
	return nil
}

type filterType int

const (
	filterNone filterType = iota
	filterNamespaceList
	filterNamespacedList
	filterMonitoringNamespaces
	filterUserMetrics
)

func responseFilter(path string) filterType {
	switch {
	case path == resourcesNamespacesPath:
		return filterNamespaceList
	case path == monitoringNamespaces:
		return filterMonitoringNamespaces
	case userMetricsPath(path):
		return filterUserMetrics
	default:
		if _, ok := namespacedListPaths[path]; ok {
			return filterNamespacedList
		}
		return filterNone
	}
}

func filterResponse(kind filterType, payload any, username string) (any, error) {
	switch kind {
	case filterNamespaceList:
		return filterList(payload, func(item map[string]any) bool {
			return ownedNamespace(nestedString(item, "metadata", "name"), username)
		})
	case filterNamespacedList:
		return filterList(payload, func(item map[string]any) bool {
			return ownedNamespace(nestedString(item, "metadata", "namespace"), username)
		})
	case filterMonitoringNamespaces:
		return filterMonitoring(payload, username)
	case filterUserMetrics:
		return collapseMetricPayload(payload), nil
	default:
		return payload, nil
	}
}

func filterList(payload any, keep func(map[string]any) bool) (any, error) {
	object, ok := payload.(map[string]any)
	if !ok {
		return nil, errorsNewShape("resource list")
	}
	items, ok := object["items"].([]any)
	if !ok {
		return nil, errorsNewShape("resource list items")
	}
	filtered := make([]any, 0, len(items))
	for _, raw := range items {
		item, ok := raw.(map[string]any)
		if !ok {
			return nil, errorsNewShape("resource list item")
		}
		if keep(item) {
			filtered = append(filtered, item)
		}
	}
	originalTotal := object["totalItems"]
	object["items"] = filtered
	object["totalItems"] = len(filtered)
	if originalTotal != nil {
		object["total"] = originalTotal
	}
	return object, nil
}

func filterMonitoring(payload any, username string) (any, error) {
	metrics, ok := payload.([]any)
	if !ok {
		return nil, errorsNewShape("monitoring namespace response")
	}
	for _, rawMetric := range metrics {
		metric, ok := rawMetric.(map[string]any)
		if !ok {
			return nil, errorsNewShape("monitoring metric")
		}
		data, ok := metric["data"].(map[string]any)
		if !ok {
			continue
		}
		results, ok := data["result"].([]any)
		if !ok {
			continue
		}
		filtered := make([]any, 0, len(results))
		for _, rawSeries := range results {
			series, ok := rawSeries.(map[string]any)
			if !ok {
				return nil, errorsNewShape("monitoring series")
			}
			namespace := nestedString(series, "metric", "namespace")
			if ownedNamespace(namespace, username) || strings.HasSuffix(namespace, "-shared") {
				filtered = append(filtered, series)
			}
		}
		data["result"] = filtered
	}
	return metrics, nil
}

func nestedString(object map[string]any, path ...string) string {
	var value any = object
	for _, part := range path {
		mapping, ok := value.(map[string]any)
		if !ok {
			return ""
		}
		value = mapping[part]
	}
	result, _ := value.(string)
	return result
}

func ownedNamespace(namespace, username string) bool {
	return namespace != "" && username != "" && strings.HasSuffix(namespace, "-"+username)
}

func userMetricsPath(path string) bool {
	_, ok := userMetricTarget(path)
	return ok
}

func isMonitoringNamespacesPost(req *http.Request) bool {
	return req.Method == http.MethodPost && req.URL.Path == monitoringNamespaces
}

func readBounded(reader io.Reader, limit int64) ([]byte, error) {
	data, err := io.ReadAll(io.LimitReader(reader, limit+1))
	if err != nil {
		return nil, fmt.Errorf("read protected kubesphere response: %w", err)
	}
	if int64(len(data)) > limit {
		return nil, fmt.Errorf("protected kubesphere response exceeds %d bytes", limit)
	}
	return data, nil
}

func decodeResponse(body []byte, encoding string) ([]byte, bool, error) {
	if encoding == "" || strings.EqualFold(encoding, "identity") {
		return body, false, nil
	}
	if !strings.EqualFold(encoding, "gzip") {
		return nil, false, fmt.Errorf("unsupported protected response encoding %q", encoding)
	}
	reader, err := gzip.NewReader(bytes.NewReader(body))
	if err != nil {
		return nil, false, fmt.Errorf("open gzip response: %w", err)
	}
	defer reader.Close()
	decoded, err := io.ReadAll(reader)
	if err != nil {
		return nil, false, fmt.Errorf("read gzip response: %w", err)
	}
	return decoded, true, nil
}

func gzipBytes(body []byte) ([]byte, error) {
	var buffer bytes.Buffer
	writer := gzip.NewWriter(&buffer)
	if _, err := writer.Write(body); err != nil {
		return nil, err
	}
	if err := writer.Close(); err != nil {
		return nil, err
	}
	return buffer.Bytes(), nil
}

func errorsNewShape(name string) error {
	return fmt.Errorf("unexpected %s shape", name)
}

var volatileLabels = map[string]struct{}{
	"instance": {}, "pod": {}, "container": {}, "endpoint": {}, "uid": {}, "job": {},
}

func collapseMetricPayload(payload any) any {
	object, objectOK := payload.(map[string]any)
	var metrics []any
	if objectOK {
		metrics, _ = object["results"].([]any)
	} else {
		metrics, _ = payload.([]any)
	}
	for _, rawMetric := range metrics {
		metric, ok := rawMetric.(map[string]any)
		if !ok {
			continue
		}
		data, _ := metric["data"].(map[string]any)
		results, _ := data["result"].([]any)
		if len(results) < 2 {
			continue
		}
		groups := make(map[string][]map[string]any)
		order := make([]string, 0)
		for _, rawSeries := range results {
			series, ok := rawSeries.(map[string]any)
			if !ok {
				continue
			}
			key := seriesKey(series)
			if _, exists := groups[key]; !exists {
				order = append(order, key)
			}
			groups[key] = append(groups[key], series)
		}
		merged := make([]any, 0, len(order))
		for _, key := range order {
			merged = append(merged, mergeSeries(groups[key]))
		}
		data["result"] = merged
	}
	return payload
}

func seriesKey(series map[string]any) string {
	metric, _ := series["metric"].(map[string]any)
	keys := make([]string, 0, len(metric))
	for key := range metric {
		if _, volatile := volatileLabels[key]; !volatile {
			keys = append(keys, key)
		}
	}
	sort.Strings(keys)
	var result strings.Builder
	for _, key := range keys {
		result.WriteString(key)
		result.WriteByte('=')
		result.WriteString(fmt.Sprint(metric[key]))
		result.WriteByte(0)
	}
	return result.String()
}

func mergeSeries(series []map[string]any) map[string]any {
	if len(series) == 1 {
		return series[0]
	}
	base := cloneMap(series[0])
	metric, _ := base["metric"].(map[string]any)
	for key := range metric {
		if _, volatile := volatileLabels[key]; volatile {
			delete(metric, key)
		}
	}
	points := make(map[string][]any)
	var latest []any
	for _, item := range series {
		for _, raw := range asSlice(item["values"]) {
			point := asSlice(raw)
			if len(point) >= 2 {
				points[fmt.Sprint(point[0])] = point
			}
		}
		value := asSlice(item["value"])
		if len(value) >= 2 && (len(latest) == 0 || number(value[0]) > number(latest[0])) {
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
		base["values"] = values
	}
	if len(latest) > 0 {
		base["value"] = latest
	}
	return base
}

func cloneMap(source map[string]any) map[string]any {
	data, _ := json.Marshal(source)
	var result map[string]any
	_ = json.Unmarshal(data, &result)
	return result
}

func asSlice(value any) []any {
	result, _ := value.([]any)
	return result
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
