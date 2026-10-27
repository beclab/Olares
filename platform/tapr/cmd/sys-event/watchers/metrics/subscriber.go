package metrics

import (
	"context"

	"bytetrade.io/web3os/tapr/cmd/sys-event/watchers"
	aprv1 "bytetrade.io/web3os/tapr/pkg/apis/apr/v1alpha1"
	aprclientset "bytetrade.io/web3os/tapr/pkg/generated/clientset/versioned"

	"k8s.io/client-go/rest"
	"k8s.io/klog/v2"
)

type Subscriber struct {
	aprClient *aprclientset.Clientset
	invoker   *watchers.CallbackInvoker
}

func (s *Subscriber) WithKubeConfig(config *rest.Config) *Subscriber {
	s.aprClient = aprclientset.NewForConfigOrDie(config)
	s.invoker = &watchers.CallbackInvoker{
		AprClient: s.aprClient,
		Retriable: func(err error) bool { return true },
	}
	return s
}

func (s *Subscriber) Do(ctx context.Context, obj interface{}, action watchers.Action) error {
	metric := obj.(*Metric)

	switch metric.MetricName {
	case "cluster_cpu_utilisation":
		klog.Info("CPU load is HIGH")
		postMetricInfo := &struct {
			CPU    float64 `json:"cpu"`
			Memory float64 `json:"memory"`
		}{
			CPU:    GetValue(metric),
			Memory: 0,
		}
		err := s.invoker.Invoke(ctx,
			func(cb *aprv1.SysEventRegistry) bool {
				return cb.Spec.Type == aprv1.Subscriber && cb.Spec.Event == aprv1.CPUHigh
			},
			postMetricInfo,
		)

		if err != nil {
			klog.Warning(err)
		}

	case "cluster_memory_utilisation":
		klog.Info("Memory usage is HIGH")
		postMetricInfo := &struct {
			CPU    float64 `json:"cpu"`
			Memory float64 `json:"memory"`
		}{
			CPU:    0,
			Memory: GetValue(metric),
		}
		err := s.invoker.Invoke(ctx,
			func(cb *aprv1.SysEventRegistry) bool {
				return cb.Spec.Type == aprv1.Subscriber && cb.Spec.Event == aprv1.MemoryHigh
			},
			postMetricInfo,
		)

		if err != nil {
			klog.Warning(err)
		}

	case "user_cpu_usage":
		user := metric.MetricValues[0].Metadata["user"]
		klog.InfoS("User's cpu usage is HIGH", "user", user)
		postMetricInfo := &struct {
			CPU    float64 `json:"cpu"`
			Memory float64 `json:"memory"`
			User   string  `json:"user"`
		}{
			CPU:    GetValue(metric),
			Memory: 0,
			User:   user,
		}
		err := s.invoker.Invoke(ctx,
			func(cb *aprv1.SysEventRegistry) bool {
				return cb.Spec.Type == aprv1.Subscriber && cb.Spec.Event == aprv1.UserCPUHigh
			},
			postMetricInfo,
		)

		if err != nil {
			klog.Warning(err)
		}

	case "user_memory_usage":
		user := metric.MetricValues[0].Metadata["user"]
		klog.InfoS("User's memory usage is HIGH", "user", user)
		postMetricInfo := &struct {
			CPU    float64 `json:"cpu"`
			Memory float64 `json:"memory"`
			User   string  `json:"user"`
		}{
			CPU:    0,
			Memory: GetValue(metric),
			User:   user,
		}
		err := s.invoker.Invoke(ctx,
			func(cb *aprv1.SysEventRegistry) bool {
				return cb.Spec.Type == aprv1.Subscriber && cb.Spec.Event == aprv1.UserMemoryHigh
			},
			postMetricInfo,
		)

		if err != nil {
			klog.Warning(err)
		}

	}

	return nil
}
