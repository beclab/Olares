package v1alpha1

import (
	"context"
	"encoding/json"
	"math"
	"time"

	"k8s.io/klog/v2"

	"bytetrade.io/web3os/bfl/pkg/api/response"
	"bytetrade.io/web3os/bfl/pkg/constants"

	"github.com/emicklei/go-restful/v3"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"k8s.io/client-go/kubernetes/scheme"
	"k8s.io/client-go/rest"
)

var UnitTypes = map[string]struct {
	Conditions []float64
	Units      []string
}{
	"cpu": {
		Conditions: []float64{0.1, 0},
		Units:      []string{"core", "m"},
	},
	"memory": {
		Conditions: []float64{math.Pow(1024, 4), math.Pow(1024, 3), math.Pow(1024, 2), 1024, 0},
		Units:      []string{"Ti", "Gi", "Mi", "Ki", "Bytes"},
	},
	"disk": {
		Conditions: []float64{math.Pow(10240, 4), math.Pow(1024, 3), math.Pow(1024, 2), 1024, 0},
		Units:      []string{"Ti", "Gi", "Mi", "Ki", "Bytes"},
	},
}

type Handler struct {
}

func newHandler() *Handler {
	return &Handler{}
}

func (h *Handler) GetClusterMetric(req *restful.Request, resp *restful.Response) {
	config := rest.Config{
		Host:        constants.KubeSphereAPIHost,
		BearerToken: req.HeaderParameter(constants.UserAuthorizationTokenKey),
		APIPath:     "/kapis",
		ContentConfig: rest.ContentConfig{
			GroupVersion: &schema.GroupVersion{
				Group:   "monitoring.kubesphere.io",
				Version: "v1alpha3",
			},
			NegotiatedSerializer: scheme.Codecs.WithoutConversion(),
		},
	}

	client, err := rest.RESTClientFor(&config)
	if err != nil {
		response.HandleError(resp, err)
		return
	}

	metricParam := "cluster_cpu_usage|cluster_cpu_total|cluster_memory_usage_wo_cache|cluster_memory_total|cluster_disk_size_usage|cluster_disk_size_capacity|cluster_net_bytes_transmitted|cluster_net_bytes_received$"

	ctx, cancel := context.WithTimeout(req.Request.Context(), 2*time.Second)
	defer cancel()

	res := client.Get().Resource("cluster").
		Param("metrics_filter", metricParam).Do(ctx)

	if res.Error() != nil {
		response.HandleError(resp, res.Error())
		return
	}

	var metrics Metrics
	data, err := res.Raw()
	if err != nil {
		response.HandleError(resp, err)
		return
	}
	klog.Infof("metricData: %s", string(data))

	err = json.Unmarshal(data, &metrics)
	if err != nil {
		response.HandleError(resp, err)
		return
	}

	var clusterMetrics ClusterMetrics
	for _, m := range metrics.Results {
		switch m.MetricName {
		case "cluster_cpu_usage":
			clusterMetrics.CPU.Usage = getValue(&m)
		case "cluster_cpu_total":
			clusterMetrics.CPU.Total = getValue(&m)

		case "cluster_disk_size_usage":
			clusterMetrics.Disk.Usage = getValue(&m)
		case "cluster_disk_size_capacity":
			clusterMetrics.Disk.Total = getValue(&m)

		case "cluster_memory_total":
			clusterMetrics.Memory.Total = getValue(&m)
		case "cluster_memory_usage_wo_cache":
			clusterMetrics.Memory.Usage = getValue(&m)

		case "cluster_net_bytes_transmitted":
			clusterMetrics.Net.Transmitted = getValue(&m)

		case "cluster_net_bytes_received":
			clusterMetrics.Net.Received = getValue(&m)
		}
	}

	fmtMetricsValue(&clusterMetrics.CPU, "cpu")
	fmtMetricsValue(&clusterMetrics.Memory, "memory")
	fmtMetricsValue(&clusterMetrics.Disk, "disk")

	response.Success(resp, clusterMetrics)
}

func getValue(m *Metric) float64 {
	if len(m.MetricData.MetricValues) == 0 {
		return 0.0
	}
	return m.MetricData.MetricValues[0].Sample[1]
}

func fmtMetricsValue(v *MetricV, unitType string) {
	v.Ratio = math.Round((v.Usage / v.Total) * 100)

	v.Unit = getSuitableUnit(v.Total, unitType)
	v.Usage = getValueByUnit(v.Usage, v.Unit)
	v.Total = getValueByUnit(v.Total, v.Unit)
}

func getSuitableUnit(value float64, unitType string) string {
	config, ok := UnitTypes[unitType]
	if !ok {
		return ""
	}
	result := config.Units[len(config.Units)-1]

	for i, condition := range config.Conditions {
		if value >= condition {
			result = config.Units[i]
			break
		}
	}
	return result
}

func getValueByUnit(num float64, unit string) float64 {

	switch unit {
	case "", "default":
		return num
	case "%":
		num *= 100
	case "m":
		num *= 1000
		if num < 1 {
			return 0
		}
	case "Ki":
		num /= 1024
	case "Mi":
		num /= math.Pow(1024, 2)
	case "Gi":
		num /= math.Pow(1024, 3)
	case "Ti":
		num /= math.Pow(1024, 4)
	case "Bytes", "B":
	case "K", "KB":
		num /= 1000
	case "M", "MB":
		num /= math.Pow(1000, 2)
	case "G", "GB":
		num /= math.Pow(1000, 3)
	case "T", "TB":
		num /= math.Pow(1000, 4)
	}

	return math.Round(num*100) / 100
}
