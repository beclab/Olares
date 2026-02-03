package apiserver

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"os"
	"strconv"
	"strings"
	"time"

	"github.com/beclab/Olares/framework/app-service/api/app.bytetrade.io/v1alpha1"
	"github.com/beclab/Olares/framework/app-service/pkg/apiserver/api"
	"github.com/beclab/Olares/framework/app-service/pkg/appcfg"
	"github.com/beclab/Olares/framework/app-service/pkg/client/clientset"
	v1alpha1client "github.com/beclab/Olares/framework/app-service/pkg/client/clientset/v1alpha1"
	"github.com/beclab/Olares/framework/app-service/pkg/constants"
	"github.com/beclab/Olares/framework/app-service/pkg/generated/clientset/versioned/scheme"
	"github.com/beclab/Olares/framework/app-service/pkg/kubesphere"
	"github.com/beclab/Olares/framework/app-service/pkg/middlewareinstaller"
	"github.com/beclab/Olares/framework/app-service/pkg/prometheus"
	"github.com/beclab/Olares/framework/app-service/pkg/tapr"

	"github.com/beclab/Olares/framework/app-service/pkg/utils"
	apputils "github.com/beclab/Olares/framework/app-service/pkg/utils/app"
	"github.com/beclab/Olares/framework/app-service/pkg/workflowinstaller"

	"github.com/emicklei/go-restful/v3"
	"github.com/pkg/errors"
	apierrors "k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/api/resource"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"k8s.io/apimachinery/pkg/util/sets"
	"k8s.io/client-go/dynamic"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/rest"
	"k8s.io/klog/v2"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/yaml"
)

func getAppByName(req *restful.Request, resp *restful.Response) (*v1alpha1.Application, error) {
	appName := req.PathParameter(ParamAppName)
	owner := req.Attribute(constants.UserContextAttribute) // get owner from request token

	client := req.Attribute(constants.KubeSphereClientAttribute).(*clientset.ClientSet)

	// run with request context for incoming client
	applist, err := client.AppClient.AppV1alpha1().Applications().List(req.Request.Context(), metav1.ListOptions{})
	if err != nil {
		api.HandleError(resp, req, err)
		return nil, err
	}

	if applist.Items == nil || len(applist.Items) == 0 {
		api.HandleNotFound(resp, req, errors.New("there is not any application"))
		return nil, err
	}

	for _, app := range applist.Items {
		if app.Spec.Name == appName && app.Spec.Owner == owner {
			return &app, nil
		}
	}

	api.HandleNotFound(resp, req, fmt.Errorf("the application %s not found", appName))
	return nil, err
}

// CheckDependencies check application dependencies, returns unsatisfied dependency.
func CheckDependencies(ctx context.Context, deps []appcfg.Dependency, ctrlClient client.Client, owner string, checkAll bool) ([]appcfg.Dependency, error) {
	unSatisfiedDeps := make([]appcfg.Dependency, 0)
	client, err := utils.GetClient()
	if err != nil {
		klog.Errorf("Failed to get client err=%v", err)
		return unSatisfiedDeps, err
	}

	applist, err := client.AppV1alpha1().Applications().List(ctx, metav1.ListOptions{})
	if err != nil {
		klog.Errorf("Failed to list application err=%v", err)
		return unSatisfiedDeps, err
	}
	appToVersion := make(map[string]string)
	appNames := make([]string, 0, len(applist.Items))
	for _, app := range applist.Items {
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
				klog.Errorf("Failed to get olares version err=%v", err)
				return unSatisfiedDeps, err
			}

			if !utils.MatchVersion(terminus.Spec.Version, dep.Version) {
				unSatisfiedDeps = append(unSatisfiedDeps, dep)
				if !checkAll {
					return unSatisfiedDeps, fmt.Errorf("olares version %s not match dependency %s", terminus.Spec.Version, dep.Version)
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

// CheckAppRequirement check if the cluster has enough resources for application install/upgrade.
func CheckAppRequirement(kubeConfig *rest.Config, token string, appConfig *appcfg.ApplicationConfig) (string, error) {
	metrics, _, err := GetClusterResource(kubeConfig, token)
	if err != nil {
		return "", err
	}

	klog.Infof("Current resource=%s", utils.PrettyJSON(metrics))
	klog.Infof("App required resource=%s", utils.PrettyJSON(appConfig.Requirement))

	switch {
	case appConfig.Requirement.Disk != nil &&
		appConfig.Requirement.Disk.CmpInt64(int64(metrics.Disk.Total-metrics.Disk.Usage)) > 0:
		return "disk", errors.New("The app's DISK requirement cannot be satisfied")
	case appConfig.Requirement.Memory != nil &&
		appConfig.Requirement.Memory.CmpInt64(int64(metrics.Memory.Total*0.9-metrics.Memory.Usage)) > 0:
		return "memory", errors.New("The app's MEMORY requirement cannot be satisfied")
	case appConfig.Requirement.CPU != nil:
		availableCPU, _ := resource.ParseQuantity(strconv.FormatFloat(metrics.CPU.Total*0.9-metrics.CPU.Usage, 'f', -1, 64))
		if appConfig.Requirement.CPU.Cmp(availableCPU) > 0 {
			return "cpu", errors.New("The app's CPU requirement cannot be satisfied")
		}
	case appConfig.Requirement.GPU != nil && !appConfig.Requirement.GPU.IsZero() &&
		metrics.GPU.Total <= 0:
		return "gpu", errors.New("The app's GPU requirement cannot be satisfied")
	}

	allocatedResources, err := getRequestResources()
	if err != nil {
		return "", err
	}
	if len(allocatedResources) == 1 {
		sufficientCPU, sufficientMemory := false, false
		if appConfig.Requirement.CPU == nil {
			sufficientCPU = true
		}
		if appConfig.Requirement.Memory == nil {
			sufficientMemory = true
		}
		for _, v := range allocatedResources {
			if appConfig.Requirement.CPU != nil {
				if v.cpu.allocatable.Cmp(*appConfig.Requirement.CPU) > 0 {
					sufficientCPU = true
				}
			}
			if appConfig.Requirement.Memory != nil {
				if v.memory.allocatable.Cmp(*appConfig.Requirement.Memory) > 0 {
					sufficientMemory = true
				}
			}
		}
		if !sufficientCPU {
			return "cpu", errors.New("The app's CPU requirement specified in the kubernetes requests cannot be satisfied")
		}
		if !sufficientMemory {
			return "memory", errors.New("The app's MEMORY requirement specified in the kubernetes requests cannot be satisfied")
		}
	}

	return "", nil
}

// CheckUserResRequirement check if the user has enough resources for application install/upgrade.
func CheckUserResRequirement(ctx context.Context, appConfig *appcfg.ApplicationConfig, username string) (string, error) {
	metrics, err := prometheus.GetCurUserResource(ctx, username)
	if err != nil {
		return "", err
	}
	switch {
	case appConfig.Requirement.Memory != nil && metrics.Memory.Total != 0 &&
		appConfig.Requirement.Memory.CmpInt64(int64(metrics.Memory.Total*0.9-metrics.Memory.Usage)) > 0:
		return "memory", errors.New("The user's app MEMORY requirement cannot be satisfied")
	case appConfig.Requirement.CPU != nil && metrics.CPU.Total != 0:
		availableCPU, _ := resource.ParseQuantity(strconv.FormatFloat(metrics.CPU.Total*0.9-metrics.CPU.Usage, 'f', -1, 64))
		if appConfig.Requirement.CPU.Cmp(availableCPU) > 0 {
			return "cpu", errors.New("The user's app CPU requirement cannot be satisfied")
		}
	}
	return "", nil
}

func CheckConflicts(ctx context.Context, conflicts []appcfg.Conflict, owner string) ([]string, error) {
	installedConflictApp := make([]string, 0)
	client, err := utils.GetClient()
	if err != nil {
		return installedConflictApp, err
	}
	appSet := sets.NewString()
	applist, err := client.AppV1alpha1().Applications().List(ctx, metav1.ListOptions{})
	if err != nil {
		return installedConflictApp, err
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
	return installedConflictApp, nil
}

type Conflict struct {
	Name string `yaml:"name" json:"name"`
	// conflict type: application
	Type string `yaml:"type" json:"type"`
}
type Options struct {
	Conflicts *[]Conflict `yaml:"conflicts" json:"conflicts"`
}

// GetClusterResource returns cluster resource metrics and cluster arches.
func GetClusterResource(kubeConfig *rest.Config, token string) (*prometheus.ClusterMetrics, []string, error) {
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

	var metrics apputils.Metrics
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

type resources struct {
	cpu    usage
	memory usage
}

type usage struct {
	allocatable *resource.Quantity
}

func getRequestResources() (map[string]resources, error) {
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

func getWorkflowConfigFromRepo(ctx context.Context, options *apputils.ConfigOptions) (*workflowinstaller.WorkflowConfig, error) {
	chartPath, err := apputils.GetIndexAndDownloadChart(ctx, options)
	if err != nil {
		return nil, err
	}

	f, err := os.Open(chartPath + "/" + apputils.AppCfgFileName)
	if err != nil {
		return nil, err
	}
	defer f.Close()

	data, err := io.ReadAll(f)
	if err != nil {
		return nil, err
	}

	var cfg appcfg.AppConfiguration
	if err := yaml.Unmarshal(data, &cfg); err != nil {
		return nil, err
	}

	namespace, _ := utils.AppNamespace(options.App, options.Owner, cfg.Spec.Namespace)

	return &workflowinstaller.WorkflowConfig{
		WorkflowName: options.App,
		ChartsName:   chartPath,
		RepoURL:      options.RepoURL,
		Namespace:    namespace,
		OwnerName:    options.Owner,
		Cfg:          &cfg}, nil
}

func getMiddlewareConfigFromRepo(ctx context.Context, options *apputils.ConfigOptions) (*middlewareinstaller.MiddlewareConfig, error) {
	chartPath, err := apputils.GetIndexAndDownloadChart(ctx, options)
	if err != nil {
		return nil, err
	}

	f, err := os.Open(chartPath + "/OlaresManifest.yaml")
	if err != nil {
		return nil, err
	}
	defer f.Close()

	data, err := io.ReadAll(f)
	if err != nil {
		return nil, err
	}

	var cfg appcfg.AppConfiguration
	if err := yaml.Unmarshal(data, &cfg); err != nil {
		return nil, err
	}

	namespace, _ := utils.AppNamespace(options.App, options.Owner, cfg.Spec.Namespace)

	return &middlewareinstaller.MiddlewareConfig{
		MiddlewareName: options.App,
		Title:          cfg.Metadata.Title,
		Version:        cfg.Metadata.Version,
		ChartsName:     chartPath,
		RepoURL:        options.RepoURL,
		Namespace:      namespace,
		OwnerName:      options.Owner,
		Cfg:            &cfg}, nil
}

func CheckMiddlewareRequirement(ctx context.Context, kubeConfig *rest.Config, middleware *tapr.Middleware) (bool, error) {
	if middleware != nil && middleware.MongoDB != nil {
		dConfig, err := dynamic.NewForConfig(kubeConfig)
		if err != nil {
			return false, err
		}
		dClient, err := middlewareinstaller.NewMiddlewareMongodb(dConfig)
		if err != nil {
			return false, err
		}
		u, err := dClient.Get(ctx, "os-platform", "mongo-cluster", metav1.GetOptions{})
		if err != nil && !apierrors.IsNotFound(err) {
			return false, err
		}
		if u == nil {
			return false, nil
		}
		state, _, err := unstructured.NestedString(u.Object, "status", "state")
		if err != nil {
			return false, err
		}
		if state == "ready" {
			return true, nil
		}
		return false, nil
	}
	return true, nil
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

type ListResult struct {
	Code   int   `json:"code"`
	Data   []any `json:"data"`
	Totals int   `json:"totals"`
}

func NewListResult[T any](items []T) ListResult {
	vs := make([]any, 0)
	if len(items) > 0 {
		for _, item := range items {
			vs = append(vs, item)
		}
	}
	return ListResult{Code: 200, Data: vs, Totals: len(items)}
}

// getCurrentUser extracts the current user from the request context
func getCurrentUser(req *restful.Request) string {
	return req.Attribute(constants.UserContextAttribute).(string)
}
