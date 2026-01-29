package app

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"strconv"
	"strings"
	"sync"
	"time"

	"github.com/beclab/Olares/framework/app-service/api/app.bytetrade.io/v1alpha1"
	sysv1alpha1 "github.com/beclab/Olares/framework/app-service/api/sys.bytetrade.io/v1alpha1"
	"github.com/beclab/Olares/framework/app-service/pkg/apiserver/api"
	"github.com/beclab/Olares/framework/app-service/pkg/appcfg"
	v1alpha1client "github.com/beclab/Olares/framework/app-service/pkg/client/clientset/v1alpha1"
	"github.com/beclab/Olares/framework/app-service/pkg/constants"
	"github.com/beclab/Olares/framework/app-service/pkg/generated/clientset/versioned/scheme"
	"github.com/beclab/Olares/framework/app-service/pkg/kubesphere"
	"github.com/beclab/Olares/framework/app-service/pkg/prometheus"
	"github.com/beclab/Olares/framework/app-service/pkg/tapr"
	"github.com/beclab/Olares/framework/app-service/pkg/users/userspace"

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

	if appConfig.Requirement.Disk != nil &&
		appConfig.Requirement.Disk.CmpInt64(int64(metrics.Disk.Total*0.9-metrics.Disk.Usage)) > 0 ||
		int64(metrics.Disk.Total*0.9-metrics.Disk.Usage) < 5*1024*1024*1024 {
		return constants.Disk, constants.DiskPressure, fmt.Errorf(constants.DiskPressureMessage, op)
	}

	if appConfig.Requirement.Memory != nil &&
		appConfig.Requirement.Memory.CmpInt64(int64(metrics.Memory.Total*0.9-metrics.Memory.Usage)) > 0 {
		return constants.Memory, constants.SystemMemoryPressure, fmt.Errorf(constants.SystemMemoryPressureMessage, op)
	}
	if appConfig.Requirement.CPU != nil {
		availableCPU, _ := resource.ParseQuantity(strconv.FormatFloat(metrics.CPU.Total*0.9-metrics.CPU.Usage, 'f', -1, 64))
		if appConfig.Requirement.CPU.Cmp(availableCPU) > 0 {
			return constants.CPU, constants.SystemCPUPressure, fmt.Errorf(constants.SystemCPUPressureMessage, op)
		}
	}

	// only support nvidia gpu managment by HAMi for now
	if appConfig.Requirement.GPU != nil && appConfig.SelectedGpuType == utils.NvidiaCardType {
		if !appConfig.Requirement.GPU.IsZero() && metrics.GPU.Total <= 0 {
			return constants.GPU, constants.SystemGPUNotAvailable, fmt.Errorf(constants.SystemGPUNotAvailableMessage, op)

		}
		nodes, err := utils.GetNodeInfo(context.TODO())
		if err != nil {
			klog.Errorf("failed to get node info %v", err)
			return "", "", err
		}
		klog.Infof("nodes info: %#v", nodes)
		var maxNodeGPUMem int64
		for _, n := range nodes {
			var sum int64
			for _, g := range n.GPUS {
				sum += g.Memory
			}
			if sum > maxNodeGPUMem {
				maxNodeGPUMem = sum
			}
		}

		if appConfig.Requirement.GPU.CmpInt64(maxNodeGPUMem) > 0 {
			return constants.GPU, constants.SystemGPUPressure, fmt.Errorf(constants.SystemGPUPressureMessage, op)
		}
	}

	return CheckAppK8sRequestResource(appConfig, op)
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
				} else if quantity, ok = n.Status.Capacity[constants.NvidiaGB10GPU]; ok {
					total += quantity.AsApproximateFloat64()
				} else if quantity, ok = n.Status.Capacity[constants.AMDAPU]; ok {
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
	switch {
	case appConfig.Requirement.Memory != nil && metrics.Memory.Total != 0 &&
		appConfig.Requirement.Memory.CmpInt64(int64(metrics.Memory.Total*0.9-metrics.Memory.Usage)) > 0:
		return constants.Memory, constants.UserMemoryPressure, fmt.Errorf(constants.UserMemoryPressureMessage, op)
	case appConfig.Requirement.CPU != nil && metrics.CPU.Total != 0:
		availableCPU, _ := resource.ParseQuantity(strconv.FormatFloat(metrics.CPU.Total*0.9-metrics.CPU.Usage, 'f', -1, 64))
		if appConfig.Requirement.CPU.Cmp(availableCPU) > 0 {
			return constants.CPU, constants.UserCPUPressure, fmt.Errorf(constants.UserCPUPressureMessage, op)
		}
	}
	return "", "", nil
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

// CheckHardwareRequirement validates whether there exists at least one node
// in the cluster that satisfies the hardware constraints specified by appConfig.Spec.Hardware.
// Returns nil, nil if satisfied (or no constraints specified). Otherwise returns a slice of reasons and an error.
func CheckHardwareRequirement(ctx context.Context, appConfig *appcfg.ApplicationConfig) ([]HardwareUnmetReason, error) {
	hw := appConfig.HardwareRequirement
	// If no hardware constraints are specified, treat as satisfied
	if hw.Cpu.Vendor == "" && hw.Cpu.Arch == "" &&
		hw.Gpu.Vendor == "" && len(hw.Gpu.Arch) == 0 &&
		hw.Gpu.SingleMemory == "" && hw.Gpu.TotalMemory == "" {
		return nil, nil
	}
	var reasons []HardwareUnmetReason

	nodes, err := getNodeInfo(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to get node info: %v", err)
	}
	if hw.Gpu.SingleMemory != "" && hw.Gpu.TotalMemory != "" {
		reasons = append(reasons, HardwareUnmetReason{
			Type:   "gpuMemory",
			Reason: fmt.Sprintf("gpu.singleMemory and gpu.totalMemory are mutually exclusive"),
		})
	}

	requiredSingleMemBytes := int64(0)
	if hw.Gpu.SingleMemory != "" {
		if q, err := resource.ParseQuantity(hw.Gpu.SingleMemory); err == nil {
			requiredSingleMemBytes = int64(q.AsApproximateFloat64())
		} else {
			reasons = append(reasons, HardwareUnmetReason{
				Type:   "gpu.singleMemory",
				Reason: fmt.Sprintf("invalid gpu.singleMemory: %s", hw.Gpu.SingleMemory),
			})
			return reasons, nil
		}
	}
	requiredTotalMemBytes := int64(0)
	if hw.Gpu.TotalMemory != "" {
		if q, err := resource.ParseQuantity(hw.Gpu.TotalMemory); err == nil {
			requiredTotalMemBytes = int64(q.AsApproximateFloat64())
		} else {
			reasons = append(reasons, HardwareUnmetReason{
				Type:   "gpu.totalMemory",
				Reason: fmt.Sprintf("invalid gpu.totalMemory: %s", hw.Gpu.TotalMemory),
			})
			return reasons, nil
		}
	}

	strEq := func(a, b string) bool {
		return strings.EqualFold(strings.TrimSpace(a), strings.TrimSpace(b))
	}

	// Build sets for each constraint, defaulting to all nodes when unspecified
	universe := sets.NewString()
	indexToName := make([]string, 0, len(nodes))
	for i := range nodes {
		name := fmt.Sprintf("idx%d", i)
		indexToName = append(indexToName, name)
		universe.Insert(name)
	}
	newAll := func() sets.String { return universe.Clone() }

	// CPU vendor
	cpuVendorSet := newAll()
	if hw.Cpu.Vendor != "" {
		cpuVendorSet = sets.NewString()
		for i, n := range nodes {
			match := false
			for _, c := range n.CPU {
				if strEq(c.Vendor, hw.Cpu.Vendor) {
					match = true
					break
				}
			}
			if match {
				cpuVendorSet.Insert(indexToName[i])
			}
		}
	}
	// CPU arch
	cpuArchSet := newAll()
	if hw.Cpu.Arch != "" {
		cpuArchSet = sets.NewString()
		for i, n := range nodes {
			match := false
			for _, c := range n.CPU {
				if strEq(c.Arch, hw.Cpu.Arch) {
					match = true
					break
				}
			}
			if match {
				cpuArchSet.Insert(indexToName[i])
			}
		}
	}
	// GPU vendor
	gpuVendorSet := newAll()
	if hw.Gpu.Vendor != "" {
		gpuVendorSet = sets.NewString()
		for i, n := range nodes {
			ok := false
			for _, g := range n.GPUS {
				if strEq(g.Vendor, hw.Gpu.Vendor) {
					ok = true
					break
				}
			}
			if ok {
				gpuVendorSet.Insert(indexToName[i])
			}
		}
	}
	// GPU arch
	gpuArchSet := newAll()
	if len(hw.Gpu.Arch) > 0 {
		gpuArchSet = sets.NewString()
		for i, n := range nodes {
			ok := false
			for _, g := range n.GPUS {
				for _, want := range hw.Gpu.Arch {
					if strEq(g.Architecture, want) {
						ok = true
						break
					}
				}
				if ok {
					break
				}
			}
			if ok {
				gpuArchSet.Insert(indexToName[i])
			}
		}
	}
	// GPU single memory
	gpuSingleMemSet := newAll()
	if requiredSingleMemBytes > 0 {
		gpuSingleMemSet = sets.NewString()
		for i, n := range nodes {
			ok := false
			for _, g := range n.GPUS {
				if g.Memory >= requiredSingleMemBytes {
					ok = true
					break
				}
			}
			if ok {
				gpuSingleMemSet.Insert(indexToName[i])
			}
		}
	}
	// GPU total memory
	gpuTotalMemSet := newAll()
	if requiredTotalMemBytes > 0 {
		gpuTotalMemSet = sets.NewString()
		for i, n := range nodes {
			var total int64
			for _, g := range n.GPUS {
				total += g.Memory
			}
			if total >= requiredTotalMemBytes {
				gpuTotalMemSet.Insert(indexToName[i])
			}
		}
	}

	// Intersect all sets
	res := cpuVendorSet.Intersection(cpuArchSet).
		Intersection(gpuVendorSet).
		Intersection(gpuArchSet).
		Intersection(gpuSingleMemSet).
		Intersection(gpuTotalMemSet)

	if res.Len() > 0 {
		return nil, nil
	}

	// build detailed diagnostics about which conditions failed
	if hw.Cpu.Vendor != "" && cpuVendorSet.Len() == 0 {
		reasons = append(reasons, HardwareUnmetReason{
			Type:   "cpu.vendor",
			Reason: fmt.Sprintf("cpu.vendor=%s has no matching nodes", hw.Cpu.Vendor),
		})
	}
	if hw.Cpu.Arch != "" && cpuArchSet.Len() == 0 {
		reasons = append(reasons, HardwareUnmetReason{
			Type:   "cpu.arch",
			Reason: fmt.Sprintf("cpu.arch=%s has no matching nodes", hw.Cpu.Arch),
		})
	}
	if hw.Gpu.Vendor != "" && gpuVendorSet.Len() == 0 {
		reasons = append(reasons, HardwareUnmetReason{
			Type:   "gpu.vendor",
			Reason: fmt.Sprintf("gpu.vendor=%s has no matching nodes", hw.Gpu.Vendor),
		})
	}
	if len(hw.Gpu.Arch) > 0 && gpuArchSet.Len() == 0 {
		reasons = append(reasons, HardwareUnmetReason{
			Type:   "gpu.architecture",
			Reason: fmt.Sprintf("gpu.arch in %v has no matching nodes", hw.Gpu.Arch),
		})
	}
	if requiredSingleMemBytes > 0 && gpuSingleMemSet.Len() == 0 {
		reasons = append(reasons, HardwareUnmetReason{
			Type:   "gpu.singleMemory",
			Reason: fmt.Sprintf("gpu.singleMemory>=%s has no matching nodes", hw.Gpu.SingleMemory),
		})
	}
	if requiredTotalMemBytes > 0 && gpuTotalMemSet.Len() == 0 {
		reasons = append(reasons, HardwareUnmetReason{
			Type:   "gpu.totalMemory",
			Reason: fmt.Sprintf("gpu.totalMemory>=%s has no matching nodes", hw.Gpu.TotalMemory),
		})
	}
	if len(reasons) > 0 {
		return reasons, nil
	}

	// all individual conditions have candidates, but no single node satisfies all combined
	return []HardwareUnmetReason{{
		Type:   "intersection",
		Reason: fmt.Sprintf("no single node satisfies the combined constraints"),
	}}, nil
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
	initAllocatable := resource.MustParse("0")
	availableResources := resources{
		cpu:    usage{allocatable: &initAllocatable},
		memory: usage{allocatable: &initAllocatable},
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

func CheckAppK8sRequestResource(appConfig *appcfg.ApplicationConfig, op v1alpha1.OpType) (constants.ResourceType, constants.ResourceConditionType, error) {
	availableResources, err := GetClusterAvailableResource()
	if err != nil {
		return "", "", err
	}
	if appConfig == nil {
		return "", "", errors.New("nil appConfig")
	}

	sufficientCPU, sufficientMemory := false, false

	if appConfig.Requirement.CPU == nil {
		sufficientCPU = true
	}
	if appConfig.Requirement.Memory == nil {
		sufficientMemory = true
	}
	if appConfig.Requirement.CPU != nil && availableResources.cpu.allocatable.Cmp(*appConfig.Requirement.CPU) > 0 {
		sufficientCPU = true
	}
	if appConfig.Requirement.Memory != nil && availableResources.memory.allocatable.Cmp(*appConfig.Requirement.Memory) > 0 {
		sufficientMemory = true
	}
	if !sufficientCPU {
		return constants.CPU, constants.K8sRequestCPUPressure, fmt.Errorf(constants.K8sRequestCPUPressureMessage, op)
	}
	if !sufficientMemory {
		return constants.Memory, constants.K8sRequestMemoryPressure, fmt.Errorf(constants.K8sRequestMemoryPressureMessage, op)
	}
	return "", "", nil
}
