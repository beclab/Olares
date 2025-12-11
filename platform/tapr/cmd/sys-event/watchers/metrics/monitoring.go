package metrics

import (
	"context"
	"time"

	"bytetrade.io/web3os/tapr/cmd/sys-event/watchers"
	"bytetrade.io/web3os/tapr/pkg/utils"

	"k8s.io/apimachinery/pkg/api/resource"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/util/wait"
	"k8s.io/client-go/rest"
	"k8s.io/klog/v2"
)

var DefaultMetricThreshold = 0.9

type Watcher struct {
	ctx           context.Context
	subscriber    *Subscriber
	eventWatchers *watchers.Watchers
	monitoring    Monitoring
}

func NewMetricsWatcherOrDie(ctx context.Context,
	w *watchers.Watchers, n *watchers.Notification, kubeconfig *rest.Config) *Watcher {
	return &Watcher{
		ctx:           ctx,
		subscriber:    (&Subscriber{notification: n}).WithKubeConfig(kubeconfig),
		eventWatchers: w,
		monitoring:    utils.ValueMust[Monitoring](NewPrometheus(PrometheusEndpoint)),
	}
}

func (w *Watcher) Run() {
	klog.Info("start cluster metrics watcher")

	latestSend := map[string]time.Time{
		"cluster_cpu_utilisation":    time.Now().Add(-2 * time.Minute),
		"cluster_memory_utilisation": time.Now().Add(-2 * time.Minute),
		"user_cpu_usage":             time.Now().Add(-2 * time.Minute),
		"user_memory_usage":          time.Now().Add(-2 * time.Minute),
	}

	wait.PollImmediateInfiniteWithContext(w.ctx, 5*time.Second, func(ctx context.Context) (done bool, err error) {
		if ctx.Err() != nil {
			return true, nil
		}
		opts := QueryOptions{
			Level: LevelCluster,
		}

		metrics := w.monitoring.GetNamedMetrics(ctx, []string{"cluster_cpu_utilisation", "cluster_memory_utilisation"}, time.Now(), opts)

		for _, m := range metrics {
			if GetValue(&m) > DefaultMetricThreshold {
				now := time.Now()
				if now.Sub(latestSend[m.MetricName]) > 1*time.Minute {
					w.eventWatchers.Enqueue(watchers.EnqueueObj{
						Obj:       &m,
						Action:    watchers.UNKNOWN,
						Subscribe: w.subscriber,
					})
					latestSend[m.MetricName] = now
				}
			}
		}
		users, err := w.subscriber.notification.DynamicClient.Resource(watchers.UserSchemeGroupVersionResource).List(ctx, metav1.ListOptions{})
		if err != nil {
			klog.Error("get user list error, ", err)
			return false, err
		}

		userResourceMap := make(map[string]UserMetrics)

		filteredUsers := make([]string, 0)
		for _, user := range users.Items {
			c, err := resource.ParseQuantity(user.GetAnnotations()["bytetrade.io/user-cpu-limit"])
			if err != nil {
				continue
			}
			m, err := resource.ParseQuantity(user.GetAnnotations()["bytetrade.io/user-memory-limit"])
			if err != nil {
				continue
			}
			userResourceMap[user.GetName()] = UserMetrics{
				CPU: Value{
					Total: c.AsApproximateFloat64(),
				},
				Memory: Value{
					Total: m.AsApproximateFloat64(),
				},
			}
			filteredUsers = append(filteredUsers, user.GetName())
		}

		for _, username := range filteredUsers {
			opts = QueryOptions{
				Level:    LevelUser,
				UserName: username,
			}
			metrics = w.monitoring.GetNamedMetrics(ctx, []string{"user_cpu_usage", "user_memory_usage"}, time.Now(), opts)
			for _, m := range metrics {
				switch m.MetricName {
				case "user_cpu_usage":
					if GetValue(&m) > userResourceMap[username].CPU.Total*DefaultMetricThreshold {
						now := time.Now()
						if now.Sub(latestSend[m.MetricName]) > 1*time.Minute {
							w.eventWatchers.Enqueue(watchers.EnqueueObj{
								Obj:       &m,
								Action:    watchers.UNKNOWN,
								Subscribe: w.subscriber,
							})
							latestSend[m.MetricName] = now
						}
					}
				case "user_memory_usage":
					if GetValue(&m) > userResourceMap[username].Memory.Total*DefaultMetricThreshold {
						now := time.Now()
						if now.Sub(latestSend[m.MetricName]) > 1*time.Minute {
							w.eventWatchers.Enqueue(watchers.EnqueueObj{
								Obj:       &m,
								Action:    watchers.UNKNOWN,
								Subscribe: w.subscriber,
							})
							latestSend[m.MetricName] = now
						}
					}

				}

			}
		}

		return false, nil
	})
}
