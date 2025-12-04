package prometheus

import (
	"errors"
	"fmt"
	"strconv"
	"time"

	jsoniter "github.com/json-iterator/go"
)

const (
	MetricTypeMatrix = "matrix"
	MetricTypeVector = "vector"
)

type ClusterMetrics struct {
	CPU    Value `json:"cpu"`
	Memory Value `json:"memory"`
	Disk   Value `json:"disk"`
	GPU    Value `json:"gpu"`
}

type Value struct {
	Total float64 `json:"total"`
	Usage float64 `json:"usage"`
	Ratio float64 `json:"ratio"`
	Unit  string  `json:"unit"`
}

type Metadata struct {
	Metric string `json:"metric,omitempty" description:"metric name"`
	Type   string `json:"type,omitempty" description:"metric type"`
	Help   string `json:"help,omitempty" description:"metric description"`
}

type Metric struct {
	MetricName string `json:"metric_name,omitempty" description:"metric name, eg. scheduler_up_sum" csv:"metric_name"`
	MetricData `json:"data,omitempty" description:"actual metric result"`
	Error      string `json:"error,omitempty" csv:"-"`
}

type MetricValues []MetricValue

type MetricData struct {
	MetricType   string `json:"resultType,omitempty" description:"result type, one of matrix, vector" csv:"metric_type"`
	MetricValues `json:"result,omitempty" description:"metric data including labels, time series and values" csv:"metric_values"`
}

type DashboardEntity struct {
	GrafanaDashboardUrl     string `json:"grafanaDashboardUrl,omitempty"`
	GrafanaDashboardContent string `json:"grafanaDashboardContent,omitempty"`
	Description             string `json:"description,omitempty"`
	Namespace               string `json:"namespace,omitempty"`
}

// Point The first element is the timestamp, the second is the metric value.
// eg, [1585658599.195, 0.528]
type Point [2]float64

type MetricValue struct {
	Metadata map[string]string `json:"metric,omitempty" description:"time series labels"`
	// The type of Point is a float64 array with fixed length of 2.
	// So Point will always be initialized as [0, 0], rather than nil.
	// To allow empty Sample, we should declare Sample to type *Point
	Sample         *Point        `json:"value,omitempty" description:"time series, values of vector type"`
	Series         []Point       `json:"values,omitempty" description:"time series, values of matrix type"`
	ExportSample   *ExportPoint  `json:"exported_value,omitempty" description:"exported time series, values of vector type"`
	ExportedSeries []ExportPoint `json:"exported_values,omitempty" description:"exported time series, values of matrix type"`

	MinValue     string `json:"min_value" description:"minimum value from monitor points"`
	MaxValue     string `json:"max_value" description:"maximum value from monitor points"`
	AvgValue     string `json:"avg_value" description:"average value from monitor points"`
	SumValue     string `json:"sum_value" description:"sum value from monitor points"`
	Fee          string `json:"fee" description:"resource fee"`
	ResourceUnit string `json:"resource_unit"`
	CurrencyUnit string `json:"currency_unit"`
}

type ExportPoint [2]float64

func (p ExportPoint) Timestamp() string {
	return time.Unix(int64(p[0]), 0).Format("2006-01-02 03:04:05 PM")
}

func (p ExportPoint) Value() float64 {
	return p[1]
}

func (p ExportPoint) Format() string {
	return p.Timestamp() + " " + strconv.FormatFloat(p.Value(), 'f', -1, 64)
}

func (p Point) Timestamp() float64 {
	return p[0]
}

func (p Point) Value() float64 {
	return p[1]
}

func (p Point) transferToExported() ExportPoint {
	return ExportPoint{p[0], p[1]}
}

func (p Point) Add(other Point) Point {
	return Point{p[0], p[1] + other[1]}
}

// MarshalJSON implements json.Marshaler. It will be called when writing JSON to HTTP response
// Inspired by prometheus/client_golang
func (p Point) MarshalJSON() ([]byte, error) {
	t, err := jsoniter.Marshal(p.Timestamp())
	if err != nil {
		return nil, err
	}
	v, err := jsoniter.Marshal(strconv.FormatFloat(p.Value(), 'f', -1, 64))
	if err != nil {
		return nil, err
	}
	return []byte(fmt.Sprintf("[%s,%s]", t, v)), nil
}

// UnmarshalJSON implements json.Unmarshaler. This is for unmarshaling test data.
func (p *Point) UnmarshalJSON(b []byte) error {
	var v []interface{}
	if err := jsoniter.Unmarshal(b, &v); err != nil {
		return err
	}

	if v == nil {
		return nil
	}

	if len(v) != 2 {
		return errors.New("unsupported array length")
	}

	ts, ok := v[0].(float64)
	if !ok {
		return errors.New("failed to unmarshal [timestamp]")
	}
	valstr, ok := v[1].(string)
	if !ok {
		return errors.New("failed to unmarshal [value]")
	}
	valf, err := strconv.ParseFloat(valstr, 64)
	if err != nil {
		return err
	}

	p[0] = ts
	p[1] = valf
	return nil
}

var promQLTemplates = map[string]string{
	"namespaces_cpu_usage":         `round(namespace:container_cpu_usage_seconds_total:sum_rate{namespace!=""}, 0.001)`,
	"namespaces_memory_usage":      `namespace:container_memory_usage_bytes:sum{namespace!=""}`,
	"user_cpu_usage":               `round(sum by (user) (user:container_cpu_usage_seconds_total:sum_rate{namespace!="", $1}), 0.001)`,
	"user_memory_usage":            `sum by (user) (user:container_memory_usage_bytes:sum{namespace!="", $1})`,
	"cluster_cpu_total":            `sum(node:node_num_cpu:sum)`,
	"cluster_memory_total":         `sum(node:node_memory_bytes_total:sum)`,
	"cluster_cpu_utilisation":      ":node_cpu_utilisation:avg1m",
	"cluster_memory_utilisation":   ":node_memory_utilisation:",
	"node_cpu_frequency_max_hertz": "node_cpu_frequency_max_hertz",
	"node_cpu_info":                "node_cpu_info",
}

type NamespaceMetricSlice []struct {
	Namespace string
	Value     float64
}

func (b NamespaceMetricSlice) Len() int           { return len(b) }
func (b NamespaceMetricSlice) Less(i, j int) bool { return b[j].Value < b[i].Value } // desc
func (b NamespaceMetricSlice) Swap(i, j int)      { b[i], b[j] = b[j], b[i] }

type UserMetricSlice []struct {
	User  string
	Value float64
}

func (u UserMetricSlice) Len() int           { return len(u) }
func (u UserMetricSlice) Less(i, j int) bool { return u[j].Value < u[i].Value }
func (u UserMetricSlice) Swap(i, j int)      { u[i], u[j] = u[j], u[i] }
