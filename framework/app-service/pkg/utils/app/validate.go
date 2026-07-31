package app

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"math"
	"strconv"
	"strings"
	"sync"
	"time"

	"github.com/beclab/Olares/framework/app-service/pkg/apiserver/api"
	"github.com/beclab/Olares/framework/app-service/pkg/appcfg"
	v1alpha1client "github.com/beclab/Olares/framework/app-service/pkg/client/clientset/v1alpha1"
	"github.com/beclab/Olares/framework/app-service/pkg/constants"
	"github.com/beclab/Olares/framework/app-service/pkg/kubesphere"
	"github.com/beclab/Olares/framework/app-service/pkg/prometheus"
	"github.com/beclab/Olares/framework/app-service/pkg/tapr"
	"github.com/beclab/Olares/framework/app-service/pkg/users/userspace"
	"github.com/beclab/api/api/app.bytetrade.io/v1alpha1"
	sysv1alpha1 "github.com/beclab/api/api/sys.bytetrade.io/v1alpha1"
	"github.com/beclab/api/pkg/generated/clientset/versioned/scheme"

	"github.com/beclab/Olares/framework/app-service/pkg/utils"

	corev1 "k8s.io/api/core/v1"
	apierrors "k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/api/resource"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"k8s.io/apimachinery/pkg/types"
	"k8s.io/apimachinery/pkg/util/sets"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/rest"
	"k8s.io/klog/v2"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

var getNodeInfo = utils.GetNodeInfo

const resourcePressureThreshold = 0.9
const minimumDiskAvailable int64 = 5 << 30

type ResourceState struct{ CPU, Memory, Disk int64 }

type ResourceDimensions struct{ CPU, Memory, Disk bool }

var AllResourceDimensions = ResourceDimensions{CPU: true, Memory: true, Disk: true}

// MetricsUnavailableError identifies an invalid resource metric.
type MetricsUnavailableError struct {
	Resource constants.ResourceType
	Detail   string
}

func (e *MetricsUnavailableError) Error() string {
	return fmt.Sprintf("%s %s is invalid", e.Resource, e.Detail)
}

// MetricsFailureResource extracts the resource dimension from a metric error.
func MetricsFailureResource(err error) (constants.ResourceType, bool) {
	var target *MetricsUnavailableError
	if errors.As(err, &target) {
		return target.Resource, true
	}
	return "", false
}

func metricRequirementFailure(err error, op v1alpha1.OpType) (constants.ResourceType, constants.ResourceConditionType, error, bool) {
	resourceType, ok := MetricsFailureResource(err)
	if !ok {
		return "", "", nil, false
	}
	return resourceType, constants.MetricsUnavailable, fmt.Errorf(constants.MetricsUnavailableMessage, op), true
}

type ResourcePressure struct {
	Resource  string `json:"resource"`
	Required  int64  `json:"required"`
	Used      int64  `json:"used"`
	Capacity  int64  `json:"capacity"`
	Available int64  `json:"available"`
	Pressured bool   `json:"pressured"`
}

func EvaluatePhysicalCapacity(required ResourceState, metrics *prometheus.ClusterMetrics, dimensions ResourceDimensions) ([]ResourcePressure, error) {
	if metrics == nil {
		return nil, fmt.Errorf("metrics are nil")
	}
	values := []struct {
		enabled      bool
		resource     constants.ResourceType
		required     int64
		total, scale float64
	}{
		{dimensions.CPU, constants.CPU, required.CPU, metrics.CPU.Total, 1000},
		{dimensions.Memory, constants.Memory, required.Memory, metrics.Memory.Total, 1},
		{dimensions.Disk, constants.Disk, required.Disk, metrics.Disk.Total, 1},
	}
	var out []ResourcePressure
	for _, value := range values {
		if !value.enabled {
			continue
		}
		capacity, err := metricValue(value.resource, "total", value.total, value.scale, false)
		if err != nil {
			return nil, err
		}
		appendPressure(&out, resourcePressure(string(value.resource), value.required, 0, capacity, false), true)
	}
	return out, nil
}

func EvaluateClusterPressure(required ResourceState, metrics *prometheus.ClusterMetrics, dimensions ResourceDimensions) ([]ResourcePressure, error) {
	if metrics == nil {
		return nil, fmt.Errorf("metrics are nil")
	}
	cpuTotal, cpuUsed, err := metricState(dimensions.CPU, constants.CPU, metrics.CPU, 1000, false)
	if err != nil {
		return nil, err
	}
	memoryTotal, memoryUsed, err := metricState(dimensions.Memory, constants.Memory, metrics.Memory, 1, false)
	if err != nil {
		return nil, err
	}
	diskTotal, diskUsed, err := metricState(true, constants.Disk, metrics.Disk, 1, false)
	if err != nil {
		return nil, err
	}
	disk := resourcePressure(string(constants.Disk), required.Disk, diskUsed, thresholdValue(diskTotal), false)
	disk.Pressured = disk.Pressured && dimensions.Disk || disk.Available < minimumDiskAvailable
	memory := resourcePressure(string(constants.Memory), required.Memory, memoryUsed, thresholdValue(memoryTotal), false)
	cpu := resourcePressure(string(constants.CPU), required.CPU, cpuUsed, thresholdValue(cpuTotal), false)
	var out []ResourcePressure
	appendPressure(&out, disk, true)
	appendPressure(&out, memory, dimensions.Memory)
	appendPressure(&out, cpu, dimensions.CPU)
	return out, nil
}

func EvaluateOwnerPressure(required ResourceState, metrics *prometheus.ClusterMetrics, dimensions ResourceDimensions) ([]ResourcePressure, error) {
	if metrics == nil {
		return nil, fmt.Errorf("metrics are nil")
	}
	cpuTotal, cpuUsed, err := metricState(dimensions.CPU, constants.CPU, metrics.CPU, 1000, true)
	if err != nil {
		return nil, err
	}
	memoryTotal, memoryUsed, err := metricState(dimensions.Memory, constants.Memory, metrics.Memory, 1, true)
	if err != nil {
		return nil, err
	}
	memory := resourcePressure(string(constants.Memory), required.Memory, memoryUsed, thresholdValue(memoryTotal), false)
	cpu := resourcePressure(string(constants.CPU), required.CPU, cpuUsed, thresholdValue(cpuTotal), false)
	var out []ResourcePressure
	appendPressure(&out, memory, dimensions.Memory && memoryTotal > 0)
	appendPressure(&out, cpu, dimensions.CPU && cpuTotal > 0)
	return out, nil
}

func EvaluateK8sRequest(required ResourceState, dimensions ResourceDimensions, cpu, memory resource.Quantity) ([]ResourcePressure, error) {
	if cpu.Sign() < 0 || memory.Sign() < 0 {
		return nil, fmt.Errorf("kubernetes available resources must not be negative")
	}
	maxCPU := resource.NewMilliQuantity(math.MaxInt64, resource.DecimalSI)
	maxMemory := resource.NewQuantity(math.MaxInt64, resource.DecimalSI)
	if cpu.Cmp(*maxCPU) > 0 || memory.Cmp(*maxMemory) > 0 {
		return nil, fmt.Errorf("kubernetes available resources overflow int64")
	}
	cpuPressure := resourcePressure(string(constants.CPU), required.CPU, 0, cpu.MilliValue(), true)
	memoryPressure := resourcePressure(string(constants.Memory), required.Memory, 0, memory.Value(), true)
	var out []ResourcePressure
	appendPressure(&out, cpuPressure, dimensions.CPU)
	appendPressure(&out, memoryPressure, dimensions.Memory)
	return out, nil
}

func metricState(enabled bool, resource constants.ResourceType, metric prometheus.Value, scale float64, allowZero bool) (int64, int64, error) {
	if !enabled {
		return 0, 0, nil
	}
	total, err := metricValue(resource, "total", metric.Total, scale, allowZero)
	if err != nil || total == 0 && allowZero {
		return total, 0, err
	}
	used, err := metricValue(resource, "usage", metric.Usage, scale, true)
	return total, used, err
}

func metricValue(resource constants.ResourceType, detail string, value, multiplier float64, allowZero bool) (int64, error) {
	scaled := value * multiplier
	if math.IsNaN(value) || math.IsInf(value, 0) || value < 0 || !allowZero && value == 0 ||
		math.IsNaN(scaled) || math.IsInf(scaled, 0) || scaled >= math.MaxInt64 {
		return 0, &MetricsUnavailableError{Resource: resource, Detail: detail}
	}
	return int64(scaled), nil
}

func thresholdValue(value int64) int64 {
	return value/10*9 + value%10*9/10
}

func resourcePressure(name string, required, used, capacity int64, strict bool) ResourcePressure {
	headroom := capacity - used
	if headroom < 0 {
		headroom = 0
	}
	pressured := used > capacity || required > headroom
	if strict {
		pressured = required >= headroom
	}
	return ResourcePressure{Resource: name, Required: required, Used: used, Capacity: capacity, Available: headroom, Pressured: pressured}
}

func appendPressure(out *[]ResourcePressure, pressure ResourcePressure, enabled bool) {
	if enabled && pressure.Pressured {
		*out = append(*out, pressure)
	}
}

func resourceStateFromConfig(appConfig *appcfg.ApplicationConfig) (ResourceState, ResourceDimensions) {
	if appConfig == nil {
		return ResourceState{}, ResourceDimensions{}
	}
	var state ResourceState
	var dimensions ResourceDimensions
	if appConfig.Requirement.CPU != nil {
		state.CPU, dimensions.CPU = appConfig.Requirement.CPU.MilliValue(), true
	}
	if appConfig.Requirement.Memory != nil {
		state.Memory, dimensions.Memory = appConfig.Requirement.Memory.Value(), true
	}
	if appConfig.Requirement.Disk != nil {
		state.Disk, dimensions.Disk = appConfig.Requirement.Disk.Value(), true
	}
	return state, dimensions
}

func CheckChartSource(source api.AppSource) error {
	if source != api.Market && source != api.Custom && source != api.DevBox && source != api.System {
		return fmt.Errorf("unsupported chart source: %s", source)
	}
	return nil
}

// CheckDependencies check application dependencies, returns unsatisfied dependency.
func CheckDependencies(ctx context.Context, ctrlClient client.Client, deps []appcfg.Dependency, owner string, checkAll bool) ([]appcfg.Dependency, error) {
	unSatisfiedDeps := make([]appcfg.Dependency, 0)
	var appList v1alpha1.ApplicationList
	err := ctrlClient.List(ctx, &appList)
	if err != nil {
		return unSatisfiedDeps, err
	}

	appToVersion := make(map[string]string)
	appNames := make([]string, 0, len(appList.Items))
	for _, app := range appList.Items {
		clusterScoped, _ := strconv.ParseBool(app.Spec.Settings["clusterScoped"])
		// add app name to list if app is cluster scoped or owner equal app.Spec.Name
		if clusterScoped || owner == app.Spec.Owner {
			appNames = append(appNames, app.Spec.Name)
			appToVersion[app.Spec.Name] = app.Spec.Settings["version"]
		}
	}
	set := sets.NewString(appNames...)

	for _, dep := range deps {
		if dep.Type == constants.DependencyTypeSystem {
			terminus, err := utils.GetTerminus(ctx, ctrlClient)
			if err != nil {
				return unSatisfiedDeps, err
			}

			if !utils.MatchVersion(terminus.Spec.Version, dep.Version) {
				unSatisfiedDeps = append(unSatisfiedDeps, dep)
				if !checkAll {
					return unSatisfiedDeps, fmt.Errorf("terminus version %s not match dependency %s", terminus.Spec.Version, dep.Version)
				}
			}
		}
		if dep.Type == constants.DependencyTypeApp {
			if dep.SelfRely == true {
				continue
			}
			if !set.Has(dep.Name) && dep.Mandatory {
				unSatisfiedDeps = append(unSatisfiedDeps, dep)
				if !checkAll {
					return unSatisfiedDeps, fmt.Errorf("dependency application %s not existed", dep.Name)
				}
			}
			if !utils.MatchVersion(appToVersion[dep.Name], dep.Version) && dep.Mandatory {
				unSatisfiedDeps = append(unSatisfiedDeps, dep)
				if !checkAll {
					return unSatisfiedDeps, fmt.Errorf("%s version: %s not match dependency %s", dep.Name, appToVersion[dep.Name], dep.Version)
				}
			}
		}
	}
	if len(unSatisfiedDeps) > 0 {
		return unSatisfiedDeps, fmt.Errorf("some dependency not satisfied")
	}
	return unSatisfiedDeps, nil
}

func CheckDependencies2(ctx context.Context, ctrlClient client.Client, deps []appcfg.Dependency, owner string, checkAll bool) error {
	unSatisfiedDeps, err := CheckDependencies(ctx, ctrlClient, deps, owner, checkAll)
	if err != nil {
		return err
	}
	if len(unSatisfiedDeps) > 0 {
		return FormatDependencyError(unSatisfiedDeps)
	}
	return nil
}

func FormatDependencyError(deps []appcfg.Dependency) error {
	var systemDeps, appDeps []string

	for _, dep := range deps {
		depInfo := fmt.Sprintf("%s version=%s",
			dep.Name, dep.Version)

		if dep.Type == "system" {
			systemDeps = append(systemDeps, depInfo)
		} else if dep.Type == "application" {
			appDeps = append(appDeps, depInfo)
		}
	}

	var errMsg strings.Builder
	errMsg.WriteString("Missing dependencies:\n")

	if len(systemDeps) > 0 {
		errMsg.WriteString("\nSystem Dependencies:\n")
		for _, dep := range systemDeps {
			errMsg.WriteString(fmt.Sprintf("- %s\n", dep))
		}
	}

	if len(appDeps) > 0 {
		errMsg.WriteString("\nApplication Dependencies:\n")
		for _, dep := range appDeps {
			errMsg.WriteString(fmt.Sprintf("- %s\n", dep))
		}
	}

	return errors.New(errMsg.String())
}

func CheckConflicts(ctx context.Context, conflicts []appcfg.Conflict, owner string) error {
	installedConflictApp := make([]string, 0)
	client, err := utils.GetClient()
	if err != nil {
		return err
	}
	appSet := sets.NewString()
	applist, err := client.AppV1alpha1().Applications().List(ctx, metav1.ListOptions{})
	if err != nil {
		return err
	}
	for _, app := range applist.Items {
		if app.Spec.Owner != owner {
			continue
		}
		appSet.Insert(app.Spec.Name)
	}
	for _, cf := range conflicts {
		if cf.Type != "application" {
			continue
		}
		if appSet.Has(cf.Name) {
			installedConflictApp = append(installedConflictApp, cf.Name)
		}
	}
	if len(installedConflictApp) > 0 {
		return fmt.Errorf("this app conflict with those installed app: %v", installedConflictApp)
	}
	return nil
}

func CheckCfgFileVersion(version, constraint string) error {
	if !utils.MatchVersion(version, constraint) {
		return fmt.Errorf("olaresManifest.version must >= %s", constraint)
	}
	return nil
}

func CheckNamespace(ns string) error {
	if IsForbidNamespace(ns) {
		return fmt.Errorf("unsupported namespace: %s", ns)
	}
	return nil
}

func CheckUserRole(appConfig *appcfg.ApplicationConfig, owner string) error {
	role, err := kubesphere.GetUserRole(context.TODO(), owner)
	if err != nil {
		return err
	}
	if (appConfig.OnlyAdmin || appConfig.AppScope.ClusterScoped) && role != "owner" && role != "admin" {
		return errors.New("only admin user can install this app")
	}
	return nil
}

// CheckAppRequirement check if the cluster has enough resources for application install/upgrade.
func CheckAppRequirement(token string, appConfig *appcfg.ApplicationConfig, op v1alpha1.OpType) (constants.ResourceType, constants.ResourceConditionType, error) {
	metrics, _, err := GetClusterResource(token)
	if err != nil {
		return "", "", err
	}

	klog.Infof("start to %s app %s", op, appConfig.AppName)
	klog.Infof("Current resource=%s", utils.PrettyJSON(metrics))
	klog.Infof("App required resource=%s", utils.PrettyJSON(appConfig.Requirement))

	required, dimensions := resourceStateFromConfig(appConfig)
	pressure, err := EvaluateClusterPressure(required, metrics, dimensions)
	if err != nil {
		if resourceType, reason, responseErr, ok := metricRequirementFailure(err, op); ok {
			return resourceType, reason, responseErr
		}
		return "", "", err
	}
	if len(pressure) == 0 {
		return "", "", nil
	}
	switch pressure[0].Resource {
	case string(constants.Disk):
		return constants.Disk, constants.DiskPressure, fmt.Errorf(constants.DiskPressureMessage, op)
	case string(constants.Memory):
		return constants.Memory, constants.SystemMemoryPressure, fmt.Errorf(constants.SystemMemoryPressureMessage, op)
	default:
		return constants.CPU, constants.SystemCPUPressure, fmt.Errorf(constants.SystemCPUPressureMessage, op)
	}
}

func GetRequestResources() (map[string]resources, error) {
	config, err := ctrl.GetConfig()
	if err != nil {
		return nil, err
	}
	client, err := kubernetes.NewForConfig(config)
	if err != nil {
		return nil, err
	}
	nodes, err := client.CoreV1().Nodes().List(context.TODO(), metav1.ListOptions{})
	if err != nil {
		return nil, err
	}
	allocatedResources := make(map[string]resources)
	for _, node := range nodes.Items {
		allocatedResources[node.Name] = resources{cpu: usage{allocatable: node.Status.Allocatable.Cpu()},
			memory: usage{allocatable: node.Status.Allocatable.Memory()}}
		fieldSelector := fmt.Sprintf("spec.nodeName=%s,status.phase!=Failed,status.phase!=Succeeded", node.Name)
		pods, err := client.CoreV1().Pods("").List(context.TODO(), metav1.ListOptions{
			FieldSelector: fieldSelector,
		})
		if err != nil {
			return nil, err
		}
		for _, pod := range pods.Items {
			for _, container := range pod.Spec.Containers {
				allocatedResources[node.Name].cpu.allocatable.Sub(*container.Resources.Requests.Cpu())
				allocatedResources[node.Name].memory.allocatable.Sub(*container.Resources.Requests.Memory())
			}
		}
	}
	return allocatedResources, nil
}

type resources struct {
	cpu    usage
	memory usage
}

type usage struct {
	allocatable *resource.Quantity
}

// GetClusterResource returns cluster resource metrics and cluster arches.
func GetClusterResource(token string) (*prometheus.ClusterMetrics, []string, error) {
	supportArch := make([]string, 0)
	arches := sets.String{}

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
		return nil, supportArch, err
	}

	metricParam := "cluster_cpu_usage|cluster_cpu_total|cluster_memory_usage_wo_cache|cluster_memory_total|cluster_disk_size_usage|cluster_disk_size_capacity|cluster_pod_running_count|cluster_pod_quota$"

	client.Client.Timeout = 2 * time.Second
	res := client.Get().Resource("cluster").
		Param("metrics_filter", metricParam).Do(context.TODO())

	if res.Error() != nil {
		return nil, supportArch, res.Error()
	}

	var metrics Metrics
	data, err := res.Raw()
	if err != nil {
		return nil, supportArch, err
	}

	err = json.Unmarshal(data, &metrics)
	if err != nil {
		return nil, supportArch, err
	}

	var clusterMetrics prometheus.ClusterMetrics
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

	// get k8s client with node list privileges
	kubeConfig, err := ctrl.GetConfig()
	if err != nil {
		return nil, supportArch, err
	}

	k8sClient, err := v1alpha1client.NewKubeClient("", kubeConfig)
	if err != nil {
		klog.Errorf("Failed to create k8s client err=%v", err)
	} else {
		nodes, err := k8sClient.Kubernetes().CoreV1().Nodes().List(
			context.TODO(),
			metav1.ListOptions{},
		)

		if err != nil && !apierrors.IsNotFound(err) {
			klog.Errorf("Failed to list node err=%v", err)
		}

		if apierrors.IsNotFound(err) {
			clusterMetrics.GPU.Total = 0
		} else {
			var total float64 = 0
			for _, n := range nodes.Items {
				arches.Insert(n.Labels["kubernetes.io/arch"])
				if quantity, ok := n.Status.Capacity[constants.NvidiaGPU]; ok {
					total += quantity.AsApproximateFloat64()
					// } else if quantity, ok = n.Status.Capacity[constants.NvidiaGB10GPU]; ok {
					// 	total += quantity.AsApproximateFloat64()
				} else if quantity, ok = n.Status.Capacity[constants.AMDGPU]; ok {
					total += quantity.AsApproximateFloat64()
				}
			}

			clusterMetrics.GPU.Total = total
		}

	}
	for arch := range arches {
		supportArch = append(supportArch, arch)
	}
	return &clusterMetrics, supportArch, nil
}

func getValue(m *kubesphere.Metric) float64 {
	if len(m.MetricData.MetricValues) == 0 {
		return 0.0
	}
	return m.MetricData.MetricValues[0].Sample[1]
}

// CheckUserResRequirement check if the user has enough resources for application install/upgrade.
func CheckUserResRequirement(ctx context.Context, appConfig *appcfg.ApplicationConfig, op v1alpha1.OpType) (constants.ResourceType, constants.ResourceConditionType, error) {
	metrics, err := prometheus.GetCurUserResource(ctx, appConfig.OwnerName)
	if err != nil {
		return "", "", err
	}
	required, dimensions := resourceStateFromConfig(appConfig)
	pressure, err := EvaluateOwnerPressure(required, metrics, dimensions)
	if err != nil {
		if resourceType, reason, responseErr, ok := metricRequirementFailure(err, op); ok {
			return resourceType, reason, responseErr
		}
		return "", "", err
	}
	if len(pressure) == 0 {
		return "", "", nil
	}
	if pressure[0].Resource == string(constants.Memory) {
		return constants.Memory, constants.UserMemoryPressure, fmt.Errorf(constants.UserMemoryPressureMessage, op)
	}
	return constants.CPU, constants.UserCPUPressure, fmt.Errorf(constants.UserCPUPressureMessage, op)
}

func CheckMiddlewareRequirement(ctx context.Context, ctrlClient client.Client, middleware *tapr.Middleware) (bool, error) {
	if middleware != nil {
		if middleware.MongoDB != nil {
			var am v1alpha1.ApplicationManager
			err := ctrlClient.Get(ctx, types.NamespacedName{Name: "mongodb-middleware-mongodb"}, &am)
			if err != nil {
				return false, err
			}
			if am.Status.State == "running" {
				return true, nil
			}
			return false, nil
		}
		if middleware.Minio != nil {
			var am v1alpha1.ApplicationManager
			err := ctrlClient.Get(ctx, types.NamespacedName{Name: "minio-middleware-minio"}, &am)
			if err != nil {
				return false, err
			}
			if am.Status.State == "running" {
				return true, nil
			}
			return false, nil
		}
		if middleware.MySQL != nil {
			var am v1alpha1.ApplicationManager
			err := ctrlClient.Get(ctx, types.NamespacedName{Name: "mysql-middleware-mysql"}, &am)
			if err != nil {
				return false, err
			}
			if am.Status.State == "running" {
				return true, nil
			}
			return false, nil
		}

		if middleware.RabbitMQ != nil {
			var am v1alpha1.ApplicationManager
			err := ctrlClient.Get(ctx, types.NamespacedName{Name: "rabbitmq-middleware-rabbitmq"}, &am)
			if err != nil {
				return false, err
			}
			if am.Status.State == "running" {
				return true, nil
			}
			return false, nil
		}
		if middleware.Elasticsearch != nil {
			var am v1alpha1.ApplicationManager
			err := ctrlClient.Get(ctx, types.NamespacedName{Name: "elasticsearch-middleware-elasticsearch"}, &am)
			if err != nil {
				return false, err
			}
			if am.Status.State == "running" {
				return true, nil
			}
			return false, nil
		}
		if middleware.MariaDB != nil {
			var am v1alpha1.ApplicationManager
			err := ctrlClient.Get(ctx, types.NamespacedName{Name: "mariadb-middleware-mariadb"}, &am)
			if err != nil {
				return false, err
			}
			if am.Status.State == "running" {
				return true, nil
			}
			return false, nil
		}
		if middleware.Argo != nil && middleware.Argo.Required {
			var am v1alpha1.ApplicationManager
			err := ctrlClient.Get(ctx, types.NamespacedName{Name: "argo-middleware-argo"}, &am)
			if err != nil {
				return false, err
			}
			if am.Status.State == "running" {
				return true, nil
			}
			return false, nil
		}

		return true, nil

	}
	return true, nil
}

// HardwareUnmetReason describes one unmet hardware condition.
type HardwareUnmetReason struct {
	Type   string `json:"type"`
	Reason string `json:"reason"`
}

func CheckAppEnvs(ctx context.Context, ctrlClient client.Client, envs []sysv1alpha1.AppEnvVar, owner string) (*api.AppEnvCheckResult, error) {
	if len(envs) == 0 {
		return nil, nil
	}
	result := new(api.AppEnvCheckResult)
	referencedEnvs := make(map[string]string)
	var once sync.Once
	for _, env := range envs {
		if env.ValueFrom != nil && env.ValueFrom.EnvName != "" && env.Required {
			var listErr error
			once.Do(func() {
				sysenvs := new(sysv1alpha1.SystemEnvList)
				listErr = ctrlClient.List(ctx, sysenvs)
				if listErr != nil {
					return
				}
				userenvs := new(sysv1alpha1.UserEnvList)
				listErr = ctrlClient.List(ctx, userenvs, client.InNamespace(utils.UserspaceName(owner)))
				for _, sysenv := range sysenvs.Items {
					referencedEnvs[sysenv.EnvName] = sysenv.GetEffectiveValue()
				}
				for _, userenv := range userenvs.Items {
					referencedEnvs[userenv.EnvName] = userenv.GetEffectiveValue()
				}
			})
			if listErr != nil {
				return nil, fmt.Errorf("failed to list referenced envs: %s", listErr)
			}
			if value, ok := referencedEnvs[env.ValueFrom.EnvName]; !ok || value == "" {
				result.MissingRefs = append(result.MissingRefs, env)
			}
			continue
		}
		effectiveValue := env.GetEffectiveValue()
		if env.Required && effectiveValue == "" {
			result.MissingValues = append(result.MissingValues, env)
			continue
		}
		if err := env.ValidateValue(effectiveValue); err != nil {
			result.InvalidValues = append(result.InvalidValues, env)
			continue
		}
	}
	if len(result.MissingValues) > 0 || len(result.InvalidValues) > 0 || len(result.MissingRefs) > 0 {
		return result, nil
	}
	return nil, nil

}

func CheckCloneEntrances(ctrlClient client.Client, appConfig *appcfg.ApplicationConfig, insReq *api.InstallRequest) (*api.AppEntranceCheckResult, error) {
	if appConfig == nil {
		return nil, nil
	}
	// Only check when app itself supports multiple install and this installation is a clone
	if !appConfig.AllowMultipleInstall || insReq.RawAppName == "" {
		return nil, nil
	}

	result := new(api.AppEntranceCheckResult)

	reqEntranceMap := make(map[string]bool)
	for _, e := range insReq.Entrances {
		reqEntranceMap[e.Name] = true
	}

	for _, e := range appConfig.Entrances {
		if e.Invisible {
			continue
		}
		if _, ok := reqEntranceMap[e.Name]; !ok {
			result.MissingValues = append(result.MissingValues, api.EntranceClone{
				Name:  e.Name,
				Title: e.Title,
			})
			continue
		}
	}

	entranceMap := make(map[string]bool)
	titleMap := make(map[string]bool)

	var amList v1alpha1.ApplicationManagerList
	err := ctrlClient.List(context.TODO(), &amList)
	if err != nil {
		return nil, err
	}
	for _, am := range amList.Items {
		if am.Status.State == v1alpha1.Uninstalled ||
			am.Status.State == v1alpha1.InstallFailed ||
			am.Status.State == v1alpha1.DownloadingCanceled ||
			am.Status.State == v1alpha1.DownloadFailed ||
			am.Status.State == v1alpha1.PendingCanceled ||
			am.Status.State == v1alpha1.InstallingCanceled {
			continue
		}
		if am.Spec.AppOwner != appConfig.OwnerName {
			continue
		}
		if am.Spec.AppName == appConfig.AppName {
			continue
		}
		if userspace.IsSysApp(am.Spec.AppName) {
			continue
		}

		var cfg appcfg.ApplicationConfig
		err = json.Unmarshal([]byte(am.Spec.Config), &cfg)
		if err != nil {
			return nil, err
		}
		titleMap[cfg.Title] = true
		for _, e := range cfg.Entrances {
			entranceMap[e.Title] = true
		}
	}

	for _, e := range insReq.Entrances {
		if entranceMap[e.Title] {
			result.InvalidValues = append(result.InvalidValues, api.EntranceClone{
				Name:    e.Name,
				Title:   e.Title,
				Message: fmt.Sprintf("entrance %s title is duplicated", e.Name),
			})
		} else if len(e.Title) > 30 {
			result.InvalidValues = append(result.InvalidValues, api.EntranceClone{
				Name:    e.Name,
				Title:   e.Title,
				Message: fmt.Sprintf("entrance %s title cannot exceed 30 characters", e.Name),
			})
		} else if len(e.Title) == 0 {
			result.InvalidValues = append(result.InvalidValues, api.EntranceClone{
				Name:    e.Name,
				Title:   e.Title,
				Message: fmt.Sprintf("entrance %s title cannot be empty", e.Name),
			})
		}
	}

	if titleMap[insReq.Title] {
		result.TitleValidation = api.AppTitle{
			Title:   insReq.Title,
			IsValid: false,
			Message: fmt.Sprintf("title %s is duplicated", insReq.Title),
		}
	} else if len(insReq.Title) > 30 {
		result.TitleValidation = api.AppTitle{
			Title:   insReq.Title,
			IsValid: false,
			Message: "Title cannot exceed 30 characters",
		}
	} else if len(insReq.Title) == 0 {
		result.TitleValidation = api.AppTitle{
			Title:   insReq.Title,
			IsValid: false,
			Message: "Title cannot be empty",
		}
	} else {
		result.TitleValidation = api.AppTitle{
			Title:   insReq.Title,
			IsValid: true,
		}
	}

	if len(result.MissingValues) > 0 || len(result.InvalidValues) > 0 || !result.TitleValidation.IsValid {
		return result, nil
	}

	return nil, nil
}

func GetClusterAvailableResource() (*resources, error) {
	config, err := ctrl.GetConfig()
	if err != nil {
		return nil, err
	}
	client, err := kubernetes.NewForConfig(config)
	if err != nil {
		return nil, err
	}
	nodes, err := client.CoreV1().Nodes().List(context.TODO(), metav1.ListOptions{})
	if err != nil {
		return nil, err
	}
	initAllocatableCPU := resource.MustParse("0")
	initAllocatableMemory := resource.MustParse("0")
	availableResources := resources{
		cpu:    usage{allocatable: &initAllocatableCPU},
		memory: usage{allocatable: &initAllocatableMemory},
	}
	nodeList := make([]corev1.Node, 0)
	for _, node := range nodes.Items {
		if !utils.IsNodeReady(&node) || node.Spec.Unschedulable {
			continue
		}
		nodeList = append(nodeList, node)
	}
	if len(nodeList) == 0 {
		return nil, errors.New("cluster has no suitable node to schedule")
	}
	for _, node := range nodeList {
		availableResources.cpu.allocatable.Add(*node.Status.Allocatable.Cpu())
		availableResources.memory.allocatable.Add(*node.Status.Allocatable.Memory())
		fieldSelector := fmt.Sprintf("spec.nodeName=%s,status.phase!=Failed,status.phase!=Succeeded", node.Name)
		pods, err := client.CoreV1().Pods("").List(context.TODO(), metav1.ListOptions{
			FieldSelector: fieldSelector,
		})
		if err != nil {
			return nil, err
		}
		for _, pod := range pods.Items {
			for _, container := range pod.Spec.Containers {
				availableResources.cpu.allocatable.Sub(*container.Resources.Requests.Cpu())
				availableResources.memory.allocatable.Sub(*container.Resources.Requests.Memory())
			}
		}
	}
	return &availableResources, nil
}

func GetClusterAvailableResourceQuantities() (resource.Quantity, resource.Quantity, error) {
	available, err := GetClusterAvailableResource()
	if err != nil {
		return resource.Quantity{}, resource.Quantity{}, err
	}
	return available.cpu.allocatable.DeepCopy(), available.memory.allocatable.DeepCopy(), nil
}

func CheckAppK8sRequestResource(appConfig *appcfg.ApplicationConfig, op v1alpha1.OpType) (constants.ResourceType, constants.ResourceConditionType, error) {
	availableResources, err := GetClusterAvailableResource()
	if err != nil {
		return "", "", err
	}
	if appConfig == nil {
		return "", "", errors.New("nil appConfig")
	}

	required, dimensions := resourceStateFromConfig(appConfig)
	pressure, err := EvaluateK8sRequest(required, dimensions, *availableResources.cpu.allocatable, *availableResources.memory.allocatable)
	if err != nil {
		return "", "", err
	}
	if len(pressure) == 0 {
		return "", "", nil
	}
	if pressure[0].Resource == string(constants.CPU) {
		return constants.CPU, constants.K8sRequestCPUPressure, fmt.Errorf(constants.K8sRequestCPUPressureMessage, op)
	}
	return constants.Memory, constants.K8sRequestMemoryPressure, fmt.Errorf(constants.K8sRequestMemoryPressureMessage, op)
}
