package metrics

import (
	"context"
	"fmt"
	"time"

	"bytetrade.io/web3os/tapr/cmd/sys-event/watchers"
	aprv1 "bytetrade.io/web3os/tapr/pkg/apis/apr/v1alpha1"
	aprclientset "bytetrade.io/web3os/tapr/pkg/generated/clientset/versioned"
	"bytetrade.io/web3os/tapr/pkg/utils"

	"k8s.io/client-go/rest"
	"k8s.io/klog/v2"
)

var latestSend = map[string]time.Time{
	"user_cpu_usage":    time.Now().Add(-2 * time.Hour),
	"user_memory_usage": time.Now().Add(-2 * time.Hour),
}

type Subscriber struct {
	notification *watchers.Notification
	aprClient    *aprclientset.Clientset
	invoker      *watchers.CallbackInvoker
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
	admin, err := s.notification.AdminUser(ctx)
	if err != nil {
		return err
	}

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
		err = s.invoker.Invoke(ctx,
			func(cb *aprv1.SysEventRegistry) bool {
				return cb.Spec.Type == aprv1.Subscriber && cb.Spec.Event == aprv1.CPUHigh
			},
			postMetricInfo,
		)

		if err != nil {
			klog.Warning(err)
		}

		if s.notification != nil {
			return s.notification.Send(ctx, admin, "High CPU load alert", &watchers.EventPayload{
				Type: string(aprv1.CPUHigh),
			})
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
		err = s.invoker.Invoke(ctx,
			func(cb *aprv1.SysEventRegistry) bool {
				return cb.Spec.Type == aprv1.Subscriber && cb.Spec.Event == aprv1.MemoryHigh
			},
			postMetricInfo,
		)

		if err != nil {
			klog.Warning(err)
		}

		if s.notification != nil {
			return s.notification.Send(ctx, admin, "High memory usage alert", &watchers.EventPayload{
				Type: string(aprv1.MemoryHigh),
			})
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
		err = s.invoker.Invoke(ctx,
			func(cb *aprv1.SysEventRegistry) bool {
				return cb.Spec.Type == aprv1.Subscriber && cb.Spec.Event == aprv1.UserCPUHigh
			},
			postMetricInfo,
		)

		if err != nil {
			klog.Warning(err)
		}

		if s.notification != nil {
			var errs = make([]error, 0)
			now := time.Now()
			if now.Sub(latestSend[metric.MetricName]) > 1*time.Hour {
				err = s.notification.Send(ctx, admin, fmt.Sprintf("user: %s High cpu usage alert", user), &watchers.EventPayload{
					Type: string(aprv1.UserCPUHigh),
				})
				if err != nil {
					errs = append(errs, err)
				}
				err = s.notification.Send(ctx, user, "High cpu usage alert", &watchers.EventPayload{
					Type: string(aprv1.UserCPUHigh),
				})
				if err != nil {
					errs = append(errs, err)
				}
				latestSend[metric.MetricName] = now
				return utils.AggregateErrs(errs)
			}
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
		err = s.invoker.Invoke(ctx,
			func(cb *aprv1.SysEventRegistry) bool {
				return cb.Spec.Type == aprv1.Subscriber && cb.Spec.Event == aprv1.UserMemoryHigh
			},
			postMetricInfo,
		)

		if err != nil {
			klog.Warning(err)
		}

		if s.notification != nil {
			var errs = make([]error, 0)
			now := time.Now()
			if now.Sub(latestSend[metric.MetricName]) > 1*time.Hour {
				err = s.notification.Send(ctx, admin, fmt.Sprintf("user: %s High memory usage alert", user), &watchers.EventPayload{
					Type: string(aprv1.UserMemoryHigh),
				})
				if err != nil {
					errs = append(errs, err)
				}
				err = s.notification.Send(ctx, user, "High memory usage alert", &watchers.EventPayload{
					Type: string(aprv1.UserMemoryHigh),
				})
				if err != nil {
					errs = append(errs, err)
				}
				latestSend[metric.MetricName] = now
				return utils.AggregateErrs(errs)
			}

		}
	}

	return nil
}
