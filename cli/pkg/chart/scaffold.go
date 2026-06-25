package chart

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	oac "github.com/beclab/Olares/framework/oac"
	appv1 "github.com/beclab/api/api/app.bytetrade.io/v1alpha1"
	"github.com/beclab/api/manifest"
	"github.com/kubernetes/kompose/pkg/kobject"

	"helm.sh/helm/v3/pkg/chart"
	appsv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
	kresource "k8s.io/apimachinery/pkg/api/resource"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"sigs.k8s.io/yaml"
)

const (
	defaultIcon    = "https://app.cdn.olares.com/appstore/default/defaulticon.webp"
	appCfgFileName = "OlaresManifest.yaml"

	// configVersion is the olaresManifest.version every scaffold emits:
	// resources live under spec.accelerator[mode=cpu].
	configVersion   = "0.12.0"
	resourceModeCPU = "cpu"

	// olaresSystemDepName / olaresSystemDepVersion are the options.dependencies
	// entry the 0.12.0 schema requires: spec.accelerator and workloadReplicas
	// are 1.12.6-only features, so the constraint must restrict to >=1.12.6-0.
	olaresSystemDepName    = "olares"
	olaresSystemDepVersion = ">=1.12.6-0"

	// entranceAnnotation marks the compose service the user wants exposed as
	// the primary entrance (set via a compose label of the same name).
	entranceAnnotation      = "olares.service.type"
	entranceAnnotationValue = "Entrance"
)

var defaultRequests = corev1.ResourceList{
	corev1.ResourceCPU:    kresource.MustParse("100m"),
	corev1.ResourceMemory: kresource.MustParse("128Mi"),
}

var defaultLimits = corev1.ResourceList{
	corev1.ResourceCPU:    kresource.MustParse("200m"),
	corev1.ResourceMemory: kresource.MustParse("512Mi"),
}

// Options drives a single docker-compose -> Olares chart conversion.
type Options struct {
	// ComposeFiles is one or more docker-compose file paths.
	ComposeFiles []string
	// OutputDir is the chart root to write into; defaults to ./<Name>.
	OutputDir string
	// Name is the Olares app name (also the chart name / metadata.appid).
	Name string
	// Title is the human-facing title; defaults to Name.
	Title string
	// Type is the OlaresManifest type: app | recommend | middleware.
	Type string
	// Profiles / NoInterpolate are passed straight to the kompose loader.
	Profiles      []string
	NoInterpolate bool
}

// FromCompose converts the compose file(s) in opts into an Olares chart
// directory. It is the single entry point used by the CLI command.
func FromCompose(opts Options) error {
	if len(opts.ComposeFiles) == 0 {
		return fmt.Errorf("at least one compose file is required")
	}
	if opts.Name == "" {
		return fmt.Errorf("app name is required")
	}
	if opts.OutputDir == "" {
		opts.OutputDir = "./" + opts.Name
	}
	if opts.Type == "" {
		opts.Type = "app"
	}
	if opts.Title == "" {
		opts.Title = opts.Name
	}

	kopts := kobject.ConvertOptions{
		InputFiles:            opts.ComposeFiles,
		OutFile:               opts.OutputDir,
		CreateD:               true,
		CreateChart:           true,
		WithKomposeAnnotation: true,
		Replicas:              1,
		Profiles:              opts.Profiles,
		NoInterpolate:         opts.NoInterpolate,
	}
	resources, err := composeToK8s(kopts)
	if err != nil {
		return fmt.Errorf("kompose convert failed: %w", err)
	}
	return writeChart(opts, resources)
}

// writeChart serializes each kompose resource into templates/<kind>-<name>.yaml,
// stamps default resource requests/limits, namespaces every object with the
// release template, and finally writes the manifest trio.
func writeChart(opts Options, resources []runtime.Object) error {
	templatesDir := filepath.Join(opts.OutputDir, "templates")
	if err := os.MkdirAll(templatesDir, os.ModePerm); err != nil {
		return err
	}

	totalRequests := corev1.ResourceList{
		corev1.ResourceCPU:    kresource.MustParse("100m"),
		corev1.ResourceMemory: kresource.MustParse("100Mi"),
	}
	totalLimits := corev1.ResourceList{
		corev1.ResourceCPU:    kresource.MustParse("100m"),
		corev1.ResourceMemory: kresource.MustParse("100Mi"),
	}

	host, port := detectEntrance(resources, opts.Name)

	// The app store lint requires a Deployment/StatefulSet named exactly after
	// the app, so rename the primary workload (matching devbox). Renaming only
	// metadata.name leaves pod-template labels intact, so its Service still
	// selects it.
	renamePrimaryWorkload(resources, opts.Name)

	replicas := manifest.WorkloadReplicas{}
	for i := range resources {
		resource := resources[i]
		addResourcesRequirements(resource)

		// Every kind kompose emits (Deployment/StatefulSet/DaemonSet/Pod/
		// Service/PVC/ConfigMap/Secret/Ingress/...) is namespace-scoped, so we
		// template the namespace unconditionally instead of consulting a
		// cluster RESTMapper the way devbox does.
		if obj, ok := resource.(metav1.Object); ok {
			obj.SetNamespace("{{ .Release.Namespace }}")
		}

		switch obj := resource.(type) {
		case *appsv1.Deployment:
			accumulateContainerResources(obj.Spec.Template.Spec.Containers, totalRequests, totalLimits)
			replicas[obj.GetName()] = 1
		case *appsv1.StatefulSet:
			accumulateContainerResources(obj.Spec.Template.Spec.Containers, totalRequests, totalLimits)
			replicas[obj.GetName()] = 1
		case *appsv1.DaemonSet:
			accumulateContainerResources(obj.Spec.Template.Spec.Containers, totalRequests, totalLimits)
		case *corev1.Pod:
			accumulateContainerResources(obj.Spec.Containers, totalRequests, totalLimits)
		}

		mobj, ok := resource.(metav1.Object)
		if !ok {
			continue
		}
		yml, err := toYAML(resource)
		if err != nil {
			return err
		}
		kind := strings.ToLower(resource.GetObjectKind().GroupVersionKind().Kind)
		filename := filepath.Join(templatesDir, fmt.Sprintf("%s-%s.yaml", kind, mobj.GetName()))
		if err := os.WriteFile(filename, yml, 0644); err != nil {
			return err
		}
	}

	return writeManifest(opts, host, port, totalRequests, totalLimits, replicas)
}

// renamePrimaryWorkload renames one workload to appName so the chart has a
// Deployment/StatefulSet named after the app (required by the app-store lint).
// Preference: the workload annotated as the Entrance, otherwise the first
// Deployment, otherwise the first StatefulSet. If a workload already carries
// the app name, nothing is changed.
func renamePrimaryWorkload(resources []runtime.Object, appName string) {
	var firstDeploy *appsv1.Deployment
	var firstSts *appsv1.StatefulSet
	for _, r := range resources {
		switch obj := r.(type) {
		case *appsv1.Deployment:
			if obj.GetName() == appName {
				return
			}
			if obj.Annotations[entranceAnnotation] == entranceAnnotationValue {
				obj.SetName(appName)
				return
			}
			if firstDeploy == nil {
				firstDeploy = obj
			}
		case *appsv1.StatefulSet:
			if obj.GetName() == appName {
				return
			}
			if obj.Annotations[entranceAnnotation] == entranceAnnotationValue {
				obj.SetName(appName)
				return
			}
			if firstSts == nil {
				firstSts = obj
			}
		}
	}
	if firstDeploy != nil {
		firstDeploy.SetName(appName)
		return
	}
	if firstSts != nil {
		firstSts.SetName(appName)
	}
}

// writeManifest assembles the OlaresManifest.yaml + Chart.yaml + values.yaml.
func writeManifest(opts Options, entranceHost string, entrancePort int32, totalRequests, totalLimits corev1.ResourceList, replicas manifest.WorkloadReplicas) error {
	cpuReq := totalRequests[corev1.ResourceCPU]
	memReq := totalRequests[corev1.ResourceMemory]
	cpuLim := totalLimits[corev1.ResourceCPU]
	memLim := totalLimits[corev1.ResourceMemory]

	appcfg := manifest.AppConfiguration{
		ConfigVersion: configVersion,
		ConfigType:    opts.Type,
		Metadata: manifest.AppMetaData{
			Name:        opts.Name,
			Icon:        defaultIcon,
			Description: fmt.Sprintf("app %s", opts.Name),
			AppID:       opts.Name,
			Version:     "0.0.1",
			Title:       opts.Title,
			Categories:  []string{"Utilities"},
		},
		Spec: manifest.AppSpec{
			VersionName: "0.0.1",
			SupportArch: []string{"amd64", "arm64"},
		},
		Options: manifest.Options{
			AppScope: manifest.AppScope{AppRef: []string{}},
			Dependencies: []manifest.Dependency{{
				Name:    olaresSystemDepName,
				Version: olaresSystemDepVersion,
				Type:    "system",
			}},
		},
	}
	if len(replicas) > 0 {
		appcfg.WorkloadReplicas = &replicas
	}
	applyAppResources(&appcfg.Spec, oac.ManifestResourceLimits{
		RequiredCPU:    cpuReq.String(),
		RequiredMemory: memReq.String(),
		RequiredDisk:   "50Mi",
		LimitedDisk:    "5Gi",
		LimitedCPU:     cpuLim.String(),
		LimitedMemory:  memLim.String(),
	})
	appcfg.Entrances = []appv1.Entrance{{
		Name:       opts.Name,
		Host:       entranceHost,
		Port:       entrancePort,
		Title:      opts.Title,
		Icon:       defaultIcon,
		AuthLevel:  "private",
		OpenMethod: "default",
	}}

	if err := os.MkdirAll(opts.OutputDir, os.ModePerm); err != nil {
		return err
	}
	if err := writeYAMLFile(filepath.Join(opts.OutputDir, appCfgFileName), appcfg); err != nil {
		return err
	}

	meta := chart.Metadata{
		APIVersion:  "v2",
		Name:        opts.Name,
		Description: fmt.Sprintf("app %s", opts.Name),
		Type:        "application",
		Version:     "0.0.1",
		AppVersion:  "0.0.1",
	}
	if err := writeYAMLFile(filepath.Join(opts.OutputDir, "Chart.yaml"), meta); err != nil {
		return err
	}
	return writeValuesFile(filepath.Join(opts.OutputDir, "values.yaml"), replicas)
}

// writeValuesFile seeds values.yaml with workloads.<name>.replicaCount for
// every workload in replicas. The 0.12.0 lint sources each workload's replica
// count from .Values.workloads.<name>.replicaCount, so an empty values.yaml
// would fail validation as soon as workloadReplicas is declared.
func writeValuesFile(path string, replicas manifest.WorkloadReplicas) error {
	if len(replicas) == 0 {
		return os.WriteFile(path, []byte{}, 0644)
	}
	workloads := make(map[string]map[string]int32, len(replicas))
	for name, count := range replicas {
		workloads[name] = map[string]int32{"replicaCount": count}
	}
	out, err := yaml.Marshal(map[string]any{"workloads": workloads})
	if err != nil {
		return err
	}
	return os.WriteFile(path, out, 0644)
}

// detectEntrance picks the primary entrance for the scaffolded manifest:
//  1. a workload annotated olares.service.type=Entrance, resolved to its Service;
//  2. otherwise the first Service exposing a port;
//  3. otherwise a placeholder (the app name on port 80) for the user to fix.
func detectEntrance(resources []runtime.Object, appName string) (string, int32) {
	services := make([]*corev1.Service, 0)
	for _, r := range resources {
		if s, ok := r.(*corev1.Service); ok {
			services = append(services, s)
		}
	}

	for _, r := range resources {
		var ann, labels map[string]string
		switch obj := r.(type) {
		case *appsv1.Deployment:
			ann, labels = obj.Annotations, obj.Spec.Template.Labels
		case *appsv1.StatefulSet:
			ann, labels = obj.Annotations, obj.Spec.Template.Labels
		default:
			continue
		}
		if ann[entranceAnnotation] != entranceAnnotationValue {
			continue
		}
		if name, port, ok := matchService(services, labels); ok {
			return name, port
		}
	}

	for _, s := range services {
		if len(s.Spec.Ports) > 0 && s.Spec.Ports[0].Port > 0 {
			return s.Name, s.Spec.Ports[0].Port
		}
	}
	return appName, 80
}

func matchService(services []*corev1.Service, podLabels map[string]string) (string, int32, bool) {
	for _, s := range services {
		if isSelectorMatch(podLabels, s.Spec.Selector) && len(s.Spec.Ports) > 0 && s.Spec.Ports[0].Port > 0 {
			return s.Name, s.Spec.Ports[0].Port, true
		}
	}
	return "", 0, false
}

func isSelectorMatch(podLabels, selector map[string]string) bool {
	if len(selector) == 0 {
		return false
	}
	for k, v := range selector {
		if podLabels[k] != v {
			return false
		}
	}
	return true
}

// applyAppResources projects r into spec.Accelerator[mode=cpu] (the modern
// 0.12.0 resource envelope); the legacy flat spec.RequiredX/LimitedX fields
// are intentionally left empty.
func applyAppResources(spec *manifest.AppSpec, r oac.ManifestResourceLimits) {
	mode := manifest.ResourceMode{
		Mode: resourceModeCPU,
		ResourceRequirement: manifest.ResourceRequirement{
			RequiredCPU:    r.RequiredCPU,
			RequiredMemory: r.RequiredMemory,
			RequiredDisk:   r.RequiredDisk,
			LimitedDisk:    r.LimitedDisk,
			LimitedCPU:     r.LimitedCPU,
			LimitedMemory:  r.LimitedMemory,
		},
	}
	for i := range spec.Accelerator {
		if spec.Accelerator[i].Mode == resourceModeCPU {
			spec.Accelerator[i] = mode
			return
		}
	}
	spec.Accelerator = append(spec.Accelerator, mode)
}

func addResourcesRequirements(resource runtime.Object) {
	switch obj := resource.(type) {
	case *appsv1.Deployment:
		addResourcesToContainers(obj.Spec.Template.Spec.Containers, defaultRequests, defaultLimits)
	case *appsv1.StatefulSet:
		addResourcesToContainers(obj.Spec.Template.Spec.Containers, defaultRequests, defaultLimits)
	case *appsv1.DaemonSet:
		addResourcesToContainers(obj.Spec.Template.Spec.Containers, defaultRequests, defaultLimits)
	case *corev1.Pod:
		addResourcesToContainers(obj.Spec.Containers, defaultRequests, defaultLimits)
	}
}

func addResourcesToContainers(containers []corev1.Container, requests, limits corev1.ResourceList) {
	for i := range containers {
		container := &containers[i]
		if container.Resources.Requests == nil {
			container.Resources.Requests = make(corev1.ResourceList)
		}
		if container.Resources.Limits == nil {
			container.Resources.Limits = make(corev1.ResourceList)
		}
		for key, value := range requests {
			if _, exists := container.Resources.Requests[key]; !exists {
				container.Resources.Requests[key] = value
			}
		}
		for key, value := range limits {
			if _, exists := container.Resources.Limits[key]; !exists {
				container.Resources.Limits[key] = value
			}
		}
	}
}

func accumulateContainerResources(containers []corev1.Container, totalRequests, totalLimits corev1.ResourceList) {
	for i := range containers {
		c := containers[i]
		for key, value := range c.Resources.Requests {
			if existing, ok := totalRequests[key]; ok {
				existing.Add(value)
				totalRequests[key] = existing
			} else {
				totalRequests[key] = value.DeepCopy()
			}
		}
		for key, value := range c.Resources.Limits {
			if existing, ok := totalLimits[key]; ok {
				existing.Add(value)
				totalLimits[key] = existing
			} else {
				totalLimits[key] = value.DeepCopy()
			}
		}
	}
}

func toYAML(v any) ([]byte, error) {
	return yaml.Marshal(v)
}

func writeYAMLFile(path string, v any) error {
	data, err := toYAML(v)
	if err != nil {
		return err
	}
	return os.WriteFile(path, data, 0644)
}
