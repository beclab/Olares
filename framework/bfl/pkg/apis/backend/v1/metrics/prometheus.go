package metrics

import (
	"context"
	"fmt"
	"strings"
	"sync"
	"time"

	"github.com/prometheus/client_golang/api"
	apiv1 "github.com/prometheus/client_golang/api/prometheus/v1"
	"github.com/prometheus/common/model"
)

const MeteringDefaultTimeout = 20 * time.Second
const PrometheusEndpoint = "http://prometheus-operated.kubesphere-monitoring-system.svc:9090"

type Level int

const (
	LevelCluster = 1 << iota
	LevelUser
)

type QueryOptions struct {
	Level    Level
	UserName string
}

type UserMetrics struct {
	CPU    Value `json:"cpu"`
	Memory Value `json:"memory"`
}

type Value struct {
	Total float64 `json:"total"`
	Usage float64 `json:"usage"`
}

type Monitoring interface {
	GetNamedMetrics(ctx context.Context, metrics []string, ts time.Time, opts QueryOptions) []Metric
}

// prometheus implements monitoring interface backed by Prometheus
type prometheus struct {
	client apiv1.API
}

func makeUserMetricExpr(tmpl string, username string) string {
	var userSelector string
	if username != "" {
		userSelector = fmt.Sprintf(`user=~"%s"`, username)
	}
	return strings.Replace(tmpl, "$1", userSelector, -1)
}

func makeExpr(metric string, opts QueryOptions) string {
	tmpl := promQLTemplates[metric]
	switch opts.Level {
	case LevelCluster:
		return tmpl
	case LevelUser:
		return makeUserMetricExpr(tmpl, opts.UserName)
	default:
		return tmpl
	}
}

func NewPrometheus(address string) (Monitoring, error) {
	cfg := api.Config{
		Address: address,
	}

	client, err := api.NewClient(cfg)
	return prometheus{client: apiv1.NewAPI(client)}, err
}

func (p prometheus) GetNamedMetrics(ctx context.Context, metrics []string, ts time.Time, opts QueryOptions) []Metric {
	var res []Metric
	var mtx sync.Mutex
	var wg sync.WaitGroup

	for _, metric := range metrics {
		wg.Add(1)
		go func(metric string) {
			parsedResp := Metric{MetricName: metric}

			value, _, err := p.client.Query(ctx, makeExpr(metric, opts), ts)
			if err != nil {
				parsedResp.Error = err.Error()
			} else {
				parsedResp.MetricData = parseQueryResp(value)
			}

			mtx.Lock()
			res = append(res, parsedResp)
			mtx.Unlock()

			wg.Done()
		}(metric)
	}

	wg.Wait()

	return res
}

func parseQueryResp(value model.Value) MetricData {
	res := MetricData{MetricType: MetricTypeVector}

	data, _ := value.(model.Vector)

	for _, v := range data {
		mv := MetricValue{
			Metadata: make(map[string]string),
		}

		for k, v := range v.Metric {
			mv.Metadata[string(k)] = string(v)
		}

		mv.Sample = &Point{float64(v.Timestamp) / 1000, float64(v.Value)}

		res.MetricValues = append(res.MetricValues, mv)
	}

	return res
}

func GetValue(m *Metric) float64 {
	if len(m.MetricData.MetricValues) == 0 {
		return float64(0)
	}
	return m.MetricData.MetricValues[0].Sample[1]
}
