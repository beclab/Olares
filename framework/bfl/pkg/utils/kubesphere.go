package utils

import (
	"context"
	"encoding/json"
	"time"

	"bytetrade.io/web3os/bfl/pkg/apis/monitor/v1alpha1"
	"bytetrade.io/web3os/bfl/pkg/constants"

	"k8s.io/apimachinery/pkg/runtime/schema"
	"k8s.io/client-go/kubernetes/scheme"
	"k8s.io/client-go/rest"
)

func GetCurrentResource(token string) (*v1alpha1.ClusterMetrics, error) {
	config := rest.Config{
		Host:        constants.KubeSphereAPIHost,
		BearerToken: token,
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
		return nil, err
	}
	metricParam := "cluster_cpu_usage|cluster_cpu_total|cluster_memory_usage_wo_cache|cluster_memory_total|cluster_disk_size_usage|cluster_disk_size_capacity|cluster_pod_running_count|cluster_pod_quota$"

	client.Client.Timeout = 2 * time.Second
	res := client.Get().Resource("cluster").
		Param("metrics_filter", metricParam).Do(context.TODO())
	if res.Error() != nil {
		return nil, err
	}
	var metrics v1alpha1.Metrics
	data, err := res.Raw()
	if err != nil {
		return nil, err
	}
	err = json.Unmarshal(data, &metrics)
	if err != nil {
		return nil, err
	}
	var clusterMetrics v1alpha1.ClusterMetrics
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

		}
	}
	return &clusterMetrics, nil
}

func getValue(m *v1alpha1.Metric) float64 {
	return m.MetricData.MetricValues[0].Sample[1]
}
