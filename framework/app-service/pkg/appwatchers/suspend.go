package appwatchers

import (
	"context"
	"os"
	"strconv"
	"strings"
	"time"

	"github.com/beclab/Olares/framework/app-service/api/app.bytetrade.io/v1alpha1"
	"github.com/beclab/Olares/framework/app-service/pkg/kubesphere"
	"github.com/beclab/Olares/framework/app-service/pkg/prometheus"

	"golang.org/x/exp/maps"
	appsv1 "k8s.io/api/apps/v1"
	apierrors "k8s.io/apimachinery/pkg/api/errors"
	apimeta "k8s.io/apimachinery/pkg/api/meta"
	"k8s.io/apimachinery/pkg/api/resource"
	"k8s.io/klog/v2"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

const suspendAnnotation = "bytetrade.io/suspend-by"
const suspendCauseAnnotation = "bytetrade.io/suspend-cause"

// SuspendTopApp suspends the top application in cluster level that with high load.
func SuspendTopApp(ctx context.Context, client client.Client) error {
	clusterCpuThreshold := getThreshold("CLUSTER_CPU_THRESHOLD")
	clusterMemoryThreshold := getThreshold("CLUSTER_MEMORY_THRESHOLD")
	if clusterCpuThreshold == 0 && clusterMemoryThreshold == 0 {
		return nil
	}

	applications, err := getAllApplications(ctx, client, "")
	if err != nil {
		return err
	}

	toploadAppNamspaces, err := getTopLoadApps(ctx, applications, clusterCpuThreshold, clusterMemoryThreshold)
	if err != nil {
		return err
	}

	for _, ns := range toploadAppNamspaces {
		err = suspendApp(ctx, client, ns)
		if err != nil {
			return err
		}
	}

	return nil
}

// SuspendUserTopApp suspends the top application int user level that with high load.
func SuspendUserTopApp(ctx context.Context, client client.Client, user string) error {
	userCpuThreshold := getThreshold("USER_CPU_THRESHOLD")
	userMemoryThreshold := getThreshold("USER_MEMORY_THRESHOLD")

	if userCpuThreshold == 0 && userMemoryThreshold == 0 {
		return nil
	}

	applications, err := getAllApplications(ctx, client, user)
	if err != nil {
		return err
	}
	topLoadUserAppNs, err := getUserTopLoadApps(ctx, applications, user, userCpuThreshold, userMemoryThreshold)
	if err != nil {
		return err
	}
	for _, ns := range topLoadUserAppNs {
		err = suspendApp(ctx, client, ns)
		if err != nil {
			return err
		}
	}
	return nil
}

func getAllApplications(ctx context.Context, client client.Client, user string) (map[string]*v1alpha1.Application, error) {
	var appList v1alpha1.ApplicationList
	err := client.List(ctx, &appList)
	if err != nil && !apierrors.IsNotFound(err) {
		klog.Errorf("Failed to list application err=%v", err)
		return nil, err
	}

	res := make(map[string]*v1alpha1.Application)
	for _, app := range appList.Items {
		if user != "" && app.Spec.Owner != user {
			continue
		}
		res[app.Spec.Namespace] = &app
	}

	return res, nil
}

func getTopLoadApps(ctx context.Context, applications map[string]*v1alpha1.Application, clusterCpuThreshold, clusterMemoryThreshold int64) ([]string, error) {
	prom, err := prometheus.New(prometheus.Endpoint)
	if err != nil {
		klog.Errorf("Failed to connect to prometheus service err=%v", err)
		return nil, err
	}
	metrics := prom.GetNamedMetrics(ctx, []string{"cluster_cpu_utilisation", "cluster_memory_utilisation"}, time.Now(), prometheus.QueryOptions{Level: prometheus.LevelCluster})
	disableCpuMonitor := false
	disableMemoryMonitor := false
	for _, m := range metrics {
		if m.MetricName == "cluster_cpu_utilisation" && prometheus.GetValue(&m) < float64(clusterCpuThreshold)*0.01 {
			disableCpuMonitor = true
		}
		if m.MetricName == "cluster_memory_utilisation" && prometheus.GetValue(&m) < float64(clusterMemoryThreshold)*0.01 {
			disableMemoryMonitor = true
		}
	}

	metrics = prom.GetNamedMetrics(ctx, []string{"namespaces_memory_usage", "namespaces_cpu_usage"}, time.Now(), prometheus.QueryOptions{Level: prometheus.LevelCluster})
	namespaces := make(map[string]bool)
	for _, m := range metrics {
		if m.MetricName == "namespaces_cpu_usage" && disableCpuMonitor {
			continue
		}
		if m.MetricName == "namespaces_memory_usage" && disableMemoryMonitor {
			continue
		}
		sorted := prometheus.GetSortedNamespaceMetrics(&m)
		// find top 2 non-builtin apps
		top := 0
		for _, s := range sorted {
			if _, ok := applications[s.Namespace]; ok && !strings.HasPrefix(s.Namespace, "user-space-") {
				namespaces[s.Namespace] = true
				if len(namespaces) == 2 {
					return maps.Keys(namespaces), nil
				}
				top++
			}

			if top == 2 {
				break
			}
		}
	} // end of metrics loop

	return maps.Keys(namespaces), nil
}

func suspendApp(ctx context.Context, cli client.Client, namespace string) error {
	suspend := func(list client.ObjectList) error {
		err := cli.List(ctx, list, client.InNamespace(namespace))
		if err != nil && !apierrors.IsNotFound(err) {
			klog.Errorf("Failed to get workload namespace=%s err=%v", namespace, err)
			return err
		}

		listObjects, err := apimeta.ExtractList(list)
		if err != nil {
			klog.Errorf("Failed to extract list namespace=%s err=%v", namespace, err)
			return err
		}

		var zeroReplica int32 = 0
		for _, w := range listObjects {
			workloadName := ""
			switch workload := w.(type) {
			case *appsv1.Deployment:
				if workload.Annotations == nil {
					workload.Annotations = make(map[string]string)
				}
				workload.Annotations[suspendAnnotation] = "app-service"
				workload.Annotations[suspendCauseAnnotation] = "system high load"
				workload.Spec.Replicas = &zeroReplica
				workloadName = workload.Namespace + "/" + workload.Name
			case *appsv1.StatefulSet:
				if workload.Annotations == nil {
					workload.Annotations = make(map[string]string)
				}
				workload.Annotations[suspendAnnotation] = "app-service"
				workload.Annotations[suspendCauseAnnotation] = "system high load"
				workload.Spec.Replicas = &zeroReplica
				workloadName = workload.Namespace + "/" + workload.Name
			}

			klog.Infof("Try to suspend workload=%s", workloadName)
			err := cli.Update(ctx, w.(client.Object))
			if err != nil {
				klog.Errorf("Failed to scale workload=%s err=%v", workloadName, err)
				return err
			}

			klog.Infof("Success to suspend workload=%s", workloadName)
		} // end list object loop

		return nil
	} // end of suspend func

	var deploymentList appsv1.DeploymentList
	err := suspend(&deploymentList)
	if err != nil {
		return err
	}

	var statefulsetList appsv1.StatefulSetList
	err = suspend(&statefulsetList)

	return err
}

func getUserTopLoadApps(ctx context.Context, applications map[string]*v1alpha1.Application, user string, userCpuThreshold, userMemoryThreshold int64) ([]string, error) {
	prom, err := prometheus.New(prometheus.Endpoint)
	if err != nil {
		klog.Errorf("Failed to connect to prometheus service err=%v", err)
		return nil, err
	}
	opts := prometheus.QueryOptions{
		Level:    prometheus.LevelUser,
		UserName: user,
	}
	userCpuLimitStr, err := kubesphere.GetUserCPULimit(ctx, user)
	if err != nil {
		return nil, err
	}
	cpuLimit, _ := resource.ParseQuantity(userCpuLimitStr)
	userMemoryLimitStr, err := kubesphere.GetUserMemoryLimit(ctx, user)
	if err != nil {
		return nil, err
	}
	memoryLimit, _ := resource.ParseQuantity(userMemoryLimitStr)

	metrics := prom.GetNamedMetrics(ctx, []string{"user_cpu_usage", "user_memory_usage"}, time.Now(), opts)
	disableCpuMonitor := false
	disableMemoryMonitor := false
	for _, m := range metrics {
		if m.MetricName == "user_cpu_usage" && prometheus.GetValue(&m) < cpuLimit.AsApproximateFloat64()*float64(userCpuThreshold)*0.01 {
			disableCpuMonitor = true
		}
		if m.MetricName == "user_memory_usage" && prometheus.GetValue(&m) < memoryLimit.AsApproximateFloat64()*float64(userMemoryThreshold)*0.01 {
			disableMemoryMonitor = true
		}
	}
	metrics = prom.GetNamedMetrics(ctx, []string{"namespaces_memory_usage", "namespaces_cpu_usage"}, time.Now(), prometheus.QueryOptions{Level: prometheus.LevelCluster})
	klog.Infof("disableCpuMonitor:%v,disableMemoryMonitor:%v", disableCpuMonitor, disableMemoryMonitor)
	namespaces := make(map[string]bool)
	for _, m := range metrics {
		if m.MetricName == "namespaces_cpu_usage" && disableCpuMonitor {
			continue
		}
		if m.MetricName == "namespaces_memory_usage" && disableMemoryMonitor {
			continue
		}

		sorted := prometheus.GetSortedNamespaceMetrics(&m)
		// find top 2 non-builtin apps
		top := 0
		for _, s := range sorted {
			if _, ok := applications[s.Namespace]; ok && !strings.HasPrefix(s.Namespace, "user-space-") {
				namespaces[s.Namespace] = true
				if len(namespaces) == 2 {
					return maps.Keys(namespaces), nil
				}
				top++
			}

			if top == 2 {
				break
			}
		}
	} // end of metrics loop

	return maps.Keys(namespaces), nil
}

func getThreshold(name string) int64 {
	thresholdStr := os.Getenv(name)
	if thresholdStr == "0" {
		return 0
	}
	threshold, err := strconv.ParseInt(thresholdStr, 10, 64)
	if err != nil || threshold < 90 {
		threshold = 90
	}
	return threshold
}
