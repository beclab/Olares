package chart

import (
	"fmt"
	"os"
	"path/filepath"
	"regexp"
	"sort"
	"strings"

	oac "github.com/beclab/Olares/framework/oac"
	appv1 "github.com/beclab/api/api/app.bytetrade.io/v1alpha1"
	"github.com/beclab/api/manifest"
	"github.com/kubernetes/kompose/pkg/kobject"
	"github.com/kubernetes/kompose/pkg/transformer/kubernetes"

	"helm.sh/helm/v3/pkg/chart"
	appsv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
	networkingv1 "k8s.io/api/networking/v1"
	kresource "k8s.io/apimachinery/pkg/api/resource"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/util/validation"
	"sigs.k8s.io/yaml"
)

const (
	defaultIcon    = "https://app.cdn.olares.com/appstore/default/defaulticon.webp"
	appCfgFileName = "OlaresManifest.yaml"
	appAPIVersion  = "v3"

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

	// komposeControllerLabel is the compose label that overrides the workload
	// kind the conversion otherwise pins to Deployment.
	komposeControllerLabel = "kompose.controller.type"
)

var (
	// invalidNameCharRE matches everything a Kubernetes object name may not
	// contain; dnsLabelSepRE additionally folds the dots a DNS label may not
	// contain either. dns1035RE is the whole-name rule a Service has to satisfy,
	// which is also the manifest's own constraint on entrance hosts.
	invalidNameCharRE = regexp.MustCompile(`[^a-z0-9.-]`)
	dnsLabelSepRE     = regexp.MustCompile(`[^a-z0-9-]`)
	dns1035RE         = regexp.MustCompile(`^[a-z]([-a-z0-9]*[a-z0-9])?$`)

	// datastoreImageKeywords / datastorePorts drive the entrance heuristic: a
	// bundled database is never the app's UI, and should become system
	// middleware rather than stay in the chart at all.
	datastoreImageKeywords = []string{
		"postgres", "mysql", "mariadb", "redis", "valkey", "mongo", "memcached",
		"rabbitmq", "nats", "kafka", "zookeeper", "etcd", "cassandra",
		"elasticsearch", "opensearch", "clickhouse", "influxdb", "couchdb",
		"minio", "qdrant", "milvus", "weaviate", "chroma",
	}
	datastorePorts = map[int32]bool{
		3306:  true, // mysql / mariadb
		5432:  true, // postgres
		5672:  true, // rabbitmq
		6379:  true, // redis
		9042:  true, // cassandra
		9200:  true, // elasticsearch
		11211: true, // memcached
		27017: true, // mongodb
	}
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

// Result reports the guesses the conversion had to make. The CLI prints it so
// the decisions the scaffold cannot get right on its own — which service becomes
// the entrance, which volumes need re-modeling — do not stay silent.
type Result struct {
	// EntranceHost / EntrancePort are the entrance written to the manifest.
	EntranceHost string
	EntrancePort int32
	// EntranceReason explains how the entrance was picked, and EntranceGuessed
	// is false only when the compose file labeled the service explicitly.
	EntranceReason  string
	EntranceGuessed bool
	// Notices are the follow-ups the user has to act on, in reading order.
	Notices []string
}

// FromCompose converts the compose file(s) in opts into an Olares chart
// directory. It is the single entry point used by the CLI command.
func FromCompose(opts Options) (*Result, error) {
	if len(opts.ComposeFiles) == 0 {
		return nil, fmt.Errorf("at least one compose file is required")
	}
	if opts.Name == "" {
		return nil, fmt.Errorf("app name is required")
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
	resources, composeObj, err := composeToK8s(kopts)
	if err != nil {
		return nil, fmt.Errorf("kompose convert failed: %w", err)
	}
	return writeChart(opts, resources, composeObj)
}

// writeChart serializes each kompose resource into templates/<kind>-<name>.yaml,
// stamps default resource requests/limits, namespaces every object with the
// release template, and finally writes the manifest trio.
func writeChart(opts Options, resources []runtime.Object, composeObj kobject.KomposeObject) (*Result, error) {
	result := &Result{}

	// kompose leaves the raw compose service name on Service objects (web_app),
	// which is neither a valid Kubernetes object name nor a valid Olares
	// entrance host, so normalize before anything reads a name.
	renameNotices, err := normalizeResourceNames(resources)
	if err != nil {
		return nil, err
	}
	result.Notices = append(result.Notices, renameNotices...)

	totalRequests := corev1.ResourceList{
		corev1.ResourceCPU:    kresource.MustParse("100m"),
		corev1.ResourceMemory: kresource.MustParse("100Mi"),
	}
	totalLimits := corev1.ResourceList{
		corev1.ResourceCPU:    kresource.MustParse("100m"),
		corev1.ResourceMemory: kresource.MustParse("100Mi"),
	}

	// A 0.12.0 manifest must declare a non-empty workloadReplicas, and the app
	// store lint wants one of those workloads named after the app, so a compose
	// file that renders no Deployment/StatefulSet cannot produce a valid chart.
	// Checked before anything is written so a rejected run leaves no partial chart.
	services := collectServices(resources)
	workloads := collectWorkloads(resources)
	if len(workloads) == 0 {
		return nil, fmt.Errorf("no Deployment or StatefulSet was rendered from %s: "+
			"Olares scales apps through workloadReplicas, so at least one service must become a replica-controlled workload "+
			"(check for kompose.controller.type labels on the compose services)",
			strings.Join(opts.ComposeFiles, ", "))
	}

	primary := pickPrimary(services, workloads, opts.Name)
	host, port := opts.Name, int32(80)
	if primary.service != nil {
		host, port = primary.service.GetName(), servicePort(primary.service)
	}
	result.EntranceHost, result.EntrancePort = host, port
	result.EntranceReason, result.EntranceGuessed = primary.reason, !primary.labeled
	if primary.service == nil {
		result.Notices = append(result.Notices, fmt.Sprintf(
			"no service exposes a port, so entrance %s:%d is a placeholder: point it at a real service and port before deploying", host, port))
	}
	if !dns1035RE.MatchString(host) {
		result.Notices = append(result.Notices, fmt.Sprintf(
			"entrance host %q does not match the Olares pattern %s: rename the compose service or fix the entrance by hand", host, dns1035RE))
	}

	// Renaming only metadata.name leaves the pod-template labels intact, so the
	// workload's Service keeps selecting it.
	if primary.renameTarget != nil {
		result.Notices = append(result.Notices, fmt.Sprintf(
			"renamed workload %q to %q: the app store requires a Deployment/StatefulSet named after the app", primary.renameTarget.GetName(), opts.Name))
		primary.renameTarget.SetName(opts.Name)
	}
	if primary.renameBlocked {
		result.Notices = append(result.Notices, fmt.Sprintf(
			"workload %q already carries the app name, so the workload behind entrance %q kept its own name and nothing was renamed", opts.Name, host))
	}

	templatesDir := filepath.Join(opts.OutputDir, "templates")
	if err := os.MkdirAll(templatesDir, os.ModePerm); err != nil {
		return nil, err
	}

	replicas := manifest.WorkloadReplicas{}
	for i := range resources {
		resource := resources[i]
		addResourcesRequirements(resource)
		normalizeRestartPolicy(resource)
		// After stamping, so the manifest totals count the defaults too.
		if spec := podSpecOf(resource); spec != nil {
			accumulateContainerResources(spec.Containers, totalRequests, totalLimits)
		}

		// Every kind kompose emits (Deployment/StatefulSet/DaemonSet/Pod/
		// Service/PVC/ConfigMap/Secret/Ingress/...) is namespace-scoped, so we
		// template the namespace unconditionally instead of consulting a
		// cluster RESTMapper the way devbox does.
		if obj, ok := resource.(metav1.Object); ok {
			obj.SetNamespace("{{ .Release.Namespace }}")
		}

		// wireReplica is set for the workload kinds whose spec.replicas must be
		// driven by .Values.workloads.<name>.replicaCount (Deployment/StatefulSet).
		wireReplica := false
		switch obj := resource.(type) {
		case *appsv1.Deployment:
			obj.Spec.Replicas = ptrInt32(1)
			replicas[obj.GetName()] = 1
			wireReplica = true
		case *appsv1.StatefulSet:
			obj.Spec.Replicas = ptrInt32(1)
			replicas[obj.GetName()] = 1
			wireReplica = true
		}

		mobj, ok := resource.(metav1.Object)
		if !ok {
			continue
		}
		yml, err := toYAML(resource)
		if err != nil {
			return nil, err
		}
		if wireReplica {
			// app-service drives replica counts (install/suspend/resume) purely
			// through .Values.workloads.<name>.replicaCount, so the template must
			// reference it instead of the literal kompose replica count.
			yml = wireReplicasValue(yml, mobj.GetName())
		}
		kind := strings.ToLower(resource.GetObjectKind().GroupVersionKind().Kind)
		filename := filepath.Join(templatesDir, fmt.Sprintf("%s-%s.yaml", kind, mobj.GetName()))
		if err := os.WriteFile(filename, yml, 0644); err != nil {
			return nil, err
		}
	}

	result.Notices = append(result.Notices, composeNotices(composeObj)...)
	result.Notices = append(result.Notices, storageNotices(resources)...)

	if err := writeManifest(opts, host, port, totalRequests, totalLimits, replicas); err != nil {
		return nil, err
	}
	return result, nil
}

// workloadRef is a Deployment/StatefulSet together with what is needed to link
// it to a Service (the pod-template labels a selector must match) and to judge
// what it runs (its container images).
type workloadRef struct {
	obj         metav1.Object
	annotations map[string]string
	podLabels   map[string]string
	images      []string
	isDeploy    bool
}

// primaryChoice holds two related but separate outcomes: the Service backing the
// app entrance, and the workload to rename to the app name.
type primaryChoice struct {
	service *corev1.Service
	// renameTarget is nil when a workload already carries the app name.
	renameTarget metav1.Object
	reason       string
	// labeled records that the compose file named the entrance itself, so the
	// caller knows the choice is not a guess.
	labeled bool
	// renameBlocked records that a workload already carries the app name while
	// the entrance is fronting a different one, so nothing was renamed.
	renameBlocked bool
}

// pickPrimary resolves the entrance first, then the rename target. The two are
// separate decisions: a workload that happens to carry the app name must not
// override an entrance the compose file labeled explicitly.
func pickPrimary(services []*corev1.Service, workloads []workloadRef, appName string) primaryChoice {
	choice := resolveEntrance(services, workloads, appName)

	// A workload already named after the app satisfies the lint on its own, and
	// renaming a second one onto that name would collide in workloadReplicas and
	// in the shared template file name.
	for _, w := range workloads {
		if w.obj.GetName() != appName {
			continue
		}
		// Only worth reporting when the entrance really does front another
		// workload: with no entrance service there is nothing surprising about
		// keeping the name the app-named workload already has.
		entrance := workloadForService(workloads, choice.service)
		choice.renameBlocked = entrance != nil && entrance.obj.GetName() != appName
		return choice
	}

	if w := workloadForService(workloads, choice.service); w != nil {
		choice.renameTarget = w.obj
		return choice
	}
	choice.renameTarget = firstWorkload(workloads)
	return choice
}

// resolveEntrance picks the entrance service:
//  1. the service fronting a workload labeled olares.service.type=Entrance;
//  2. the service fronting the workload already named after the app;
//  3. whatever guessEntranceService settles on.
func resolveEntrance(services []*corev1.Service, workloads []workloadRef, appName string) primaryChoice {
	for _, w := range workloads {
		if w.annotations[entranceAnnotation] != entranceAnnotationValue {
			continue
		}
		if svc := matchService(services, w.podLabels); svc != nil {
			return primaryChoice{
				service: svc,
				reason:  fmt.Sprintf("%s=%s labels this compose service", entranceAnnotation, entranceAnnotationValue),
				labeled: true,
			}
		}
	}

	for _, w := range workloads {
		if w.obj.GetName() != appName {
			continue
		}
		if svc := matchService(services, w.podLabels); svc != nil {
			return primaryChoice{
				service: svc,
				reason:  fmt.Sprintf("it fronts the workload already named %q", appName),
			}
		}
	}

	svc, reason := guessEntranceService(services, workloads)
	return primaryChoice{service: svc, reason: reason}
}

// guessEntranceService prefers the service that most plausibly serves the app's
// UI: datastores are skipped because a bundled database exposing 5432 would
// otherwise win on compose service name order alone.
func guessEntranceService(services []*corev1.Service, workloads []workloadRef) (*corev1.Service, string) {
	var best, fallback *corev1.Service
	for _, s := range services {
		port := servicePort(s)
		if port == 0 {
			continue
		}
		if fallback == nil {
			fallback = s
		}
		if isDatastore(s, workloadForService(workloads, s)) {
			continue
		}
		if best == nil || servicePort(best) > port {
			best = s
		}
	}
	switch {
	case best != nil:
		return best, "lowest TCP port among the services that do not look like a datastore"
	case fallback != nil:
		return fallback, "first service exposing a TCP port; every service looks like a datastore"
	default:
		return nil, "no service exposes a TCP port"
	}
}

// isDatastore reports whether a service fronts a bundled database, cache or
// queue, judged by the images it runs and then by its port.
func isDatastore(svc *corev1.Service, workload *workloadRef) bool {
	if workload != nil {
		for _, image := range workload.images {
			if isDatastoreImage(image) {
				return true
			}
		}
	}
	for _, p := range svc.Spec.Ports {
		if datastorePorts[p.Port] {
			return true
		}
	}
	return false
}

func isDatastoreImage(image string) bool {
	name := strings.ToLower(image)
	for _, keyword := range datastoreImageKeywords {
		if strings.Contains(name, keyword) {
			return true
		}
	}
	return false
}

func collectServices(resources []runtime.Object) []*corev1.Service {
	services := make([]*corev1.Service, 0)
	for _, r := range resources {
		if s, ok := r.(*corev1.Service); ok {
			services = append(services, s)
		}
	}
	return services
}

func collectWorkloads(resources []runtime.Object) []workloadRef {
	workloads := make([]workloadRef, 0)
	for _, r := range resources {
		switch obj := r.(type) {
		case *appsv1.Deployment:
			workloads = append(workloads, workloadRef{
				obj:         obj,
				annotations: obj.Annotations,
				podLabels:   obj.Spec.Template.Labels,
				images:      containerImages(obj.Spec.Template.Spec.Containers),
				isDeploy:    true,
			})
		case *appsv1.StatefulSet:
			workloads = append(workloads, workloadRef{
				obj:         obj,
				annotations: obj.Annotations,
				podLabels:   obj.Spec.Template.Labels,
				images:      containerImages(obj.Spec.Template.Spec.Containers),
			})
		}
	}
	return workloads
}

// firstWorkload prefers a Deployment over a StatefulSet, matching devbox.
func firstWorkload(workloads []workloadRef) metav1.Object {
	for _, w := range workloads {
		if w.isDeploy {
			return w.obj
		}
	}
	if len(workloads) > 0 {
		return workloads[0].obj
	}
	return nil
}

func workloadForService(workloads []workloadRef, svc *corev1.Service) *workloadRef {
	if svc == nil {
		return nil
	}
	for i := range workloads {
		if isSelectorMatch(workloads[i].podLabels, svc.Spec.Selector) {
			return &workloads[i]
		}
	}
	return nil
}

func containerImages(containers []corev1.Container) []string {
	images := make([]string, 0, len(containers))
	for _, c := range containers {
		images = append(images, c.Image)
	}
	return images
}

// servicePort returns the first TCP port a service exposes, or 0: an entrance
// has to be reachable over HTTP, so a UDP-only service is not a candidate.
func servicePort(svc *corev1.Service) int32 {
	for _, p := range svc.Spec.Ports {
		if p.Port > 0 && (p.Protocol == "" || p.Protocol == corev1.ProtocolTCP) {
			return p.Port
		}
	}
	return 0
}

// normalizeResourceNames makes every object name a valid Kubernetes name and
// normalizes the references to those objects the same way, so both sides keep
// pointing at each other. kompose normalizes neither consistently: it names
// Service objects after the raw compose service (web_app) while labeling the pods
// with the normalized form, it lowercases a PVC's own name but not the claimName
// and pod volume name pointing at it (PGData), and it leaves the case of
// env-file derived ConfigMap names alone (Prod-env).
//
// A collision fails the conversion instead of being renamed, because two objects
// of one kind sharing a name would silently overwrite each other's template.
func normalizeResourceNames(resources []runtime.Object) ([]string, error) {
	notices := make([]string, 0)
	taken := make(map[string]string)

	for _, r := range resources {
		obj, ok := r.(metav1.Object)
		if !ok {
			continue
		}
		kind := r.GetObjectKind().GroupVersionKind().Kind
		name := obj.GetName()

		_, isService := r.(*corev1.Service)
		normalized := normalizeResourceName(name)
		if isService {
			normalized = normalizeDNSLabel(name)
		}

		if previous, clash := taken[kind+"/"+normalized]; clash {
			return nil, fmt.Errorf("%s %q and %q both become %q once compose names are normalized to valid Kubernetes names: "+
				"they would overwrite each other in templates/%s-%s.yaml, so rename one of them in the compose file",
				kind, previous, name, normalized, strings.ToLower(kind), normalized)
		}
		taken[kind+"/"+normalized] = name

		if normalized != name {
			obj.SetName(normalized)
			notice := fmt.Sprintf("%s %q was renamed to %q to be a valid Kubernetes name", kind, name, normalized)
			if isService {
				// Only a Service rename moves a hostname other containers may dial.
				notice += ": update any hostname hard-coded in a compose environment or command"
			}
			notices = append(notices, notice)
		}

		// Folding the invalid characters is not enough for a Service, whose whole
		// name has to be a DNS-1035 label: compose accepts a leading digit, and
		// nothing here can repair that. Renaming the Service alone would not
		// either, since kompose derived the pod labels it selects on and any
		// Ingress backend from the same name.
		if isService && !dns1035RE.MatchString(normalized) {
			notices = append(notices, fmt.Sprintf(
				"Service %q is not a valid Kubernetes name, which has to start with a letter and match %s: rename the compose service, because the pod selector and any Ingress backend carry this name too",
				normalized, dns1035RE))
		}
	}

	for _, r := range resources {
		switch obj := r.(type) {
		case *appsv1.StatefulSet:
			// The governing Service is another name kompose takes from the raw
			// compose service, so the Service object no longer carries it.
			obj.Spec.ServiceName = normalizeDNSLabel(obj.Spec.ServiceName)
			// A claim template's name doubles as the volume name its mounts use.
			for i := range obj.Spec.VolumeClaimTemplates {
				claim := &obj.Spec.VolumeClaimTemplates[i]
				claim.Name = normalizeDNSLabel(claim.Name)
			}
		case *networkingv1.Ingress:
			normalizeIngressBackends(obj)
		}

		spec := podSpecOf(r)
		if spec == nil {
			continue
		}
		if err := normalizePodSpecNames(spec, describeObject(r)); err != nil {
			return nil, err
		}
	}
	return notices, nil
}

func normalizeResourceName(name string) string {
	normalized := strings.Trim(invalidNameCharRE.ReplaceAllString(strings.ToLower(name), "-"), "-.")
	if normalized == "" {
		return name
	}
	return normalized
}

// normalizeDNSLabel is for the names that must be a DNS label rather than a
// subdomain, so unlike normalizeResourceName it also folds dots: Service names
// and container names. For a Service it matches kompose's own
// normalizeServiceNames, which is the form kompose already used for the pod
// labels and for the Ingress backend, so any other rule would point those at a
// name no object carries.
func normalizeDNSLabel(name string) string {
	normalized := strings.Trim(dnsLabelSepRE.ReplaceAllString(strings.ToLower(name), "-"), "-")
	if normalized == "" {
		return name
	}
	return normalized
}

// normalizeIngressBackends normalizes the Service names a kompose.service.expose
// Ingress routes to, which kompose writes with its own service-name rule rather
// than the one it named the Service object with.
func normalizeIngressBackends(ing *networkingv1.Ingress) {
	for i := range ing.Spec.Rules {
		http := ing.Spec.Rules[i].HTTP
		if http == nil {
			continue
		}
		for j := range http.Paths {
			if backend := http.Paths[j].Backend.Service; backend != nil {
				backend.Name = normalizeDNSLabel(backend.Name)
			}
		}
	}
}

// normalizePodSpecNames normalizes every name a pod template points at, with the
// same rule the object it names was normalized with. Pod volume and container
// names have to be valid on their own too — and as DNS labels, not the subdomains
// an object name may be — so a renamed volume takes its mounts along.
func normalizePodSpecNames(spec *corev1.PodSpec, owner string) error {
	volumeNames := make(map[string]string)
	for i := range spec.Volumes {
		vol := &spec.Volumes[i]
		if claim := vol.PersistentVolumeClaim; claim != nil {
			claim.ClaimName = normalizeResourceName(claim.ClaimName)
		}
		if cm := vol.ConfigMap; cm != nil {
			cm.Name = normalizeResourceName(cm.Name)
		}
		if vol.Projected != nil {
			for j := range vol.Projected.Sources {
				if cm := vol.Projected.Sources[j].ConfigMap; cm != nil {
					cm.Name = normalizeResourceName(cm.Name)
				}
			}
		}

		// A volume name is a DNS label, unlike the PVC or ConfigMap name it
		// points at, which may legally carry dots.
		normalized := normalizeDNSLabel(vol.Name)
		if previous, clash := volumeNames[normalized]; clash {
			return fmt.Errorf("%s mounts volumes %q and %q, which both become %q once normalized to valid Kubernetes names: "+
				"rename one of them in the compose file", owner, previous, vol.Name, normalized)
		}
		volumeNames[normalized] = vol.Name
		if normalized != vol.Name {
			renameVolumeMounts(spec.Containers, vol.Name, normalized)
			renameVolumeMounts(spec.InitContainers, vol.Name, normalized)
			vol.Name = normalized
		}
	}

	for _, containers := range [][]corev1.Container{spec.Containers, spec.InitContainers} {
		for i := range containers {
			container := &containers[i]
			// kompose only lowercases the container name, so a dotted compose
			// service still leaves an invalid one behind.
			container.Name = normalizeDNSLabel(container.Name)
			for j := range container.EnvFrom {
				if ref := container.EnvFrom[j].ConfigMapRef; ref != nil {
					ref.Name = normalizeResourceName(ref.Name)
				}
			}
			for j := range container.Env {
				if container.Env[j].ValueFrom == nil {
					continue
				}
				if ref := container.Env[j].ValueFrom.ConfigMapKeyRef; ref != nil {
					ref.Name = normalizeResourceName(ref.Name)
				}
			}
		}
	}
	return nil
}

func describeObject(resource runtime.Object) string {
	kind := resource.GetObjectKind().GroupVersionKind().Kind
	if obj, ok := resource.(metav1.Object); ok {
		return fmt.Sprintf("%s %q", kind, obj.GetName())
	}
	return kind
}

func renameVolumeMounts(containers []corev1.Container, from, to string) {
	for i := range containers {
		for j := range containers[i].VolumeMounts {
			if containers[i].VolumeMounts[j].Name == from {
				containers[i].VolumeMounts[j].Name = to
			}
		}
	}
}

// podSpecOf returns the pod template of every kind the conversion can emit that
// carries containers, and is the single place that list lives: name
// normalization, the resource defaults and the manifest totals all go through it.
func podSpecOf(resource runtime.Object) *corev1.PodSpec {
	switch obj := resource.(type) {
	case *appsv1.Deployment:
		return &obj.Spec.Template.Spec
	case *appsv1.StatefulSet:
		return &obj.Spec.Template.Spec
	case *appsv1.DaemonSet:
		return &obj.Spec.Template.Spec
	case *corev1.Pod:
		return &obj.Spec
	}
	return nil
}

// composeNotices reports the compose-level decisions the rendered templates
// cannot show on their own.
func composeNotices(composeObj kobject.KomposeObject) []string {
	notices := make([]string, 0)
	names := make([]string, 0, len(composeObj.ServiceConfigs))
	for name := range composeObj.ServiceConfigs {
		names = append(names, name)
	}
	sort.Strings(names)

	for _, name := range names {
		service := composeObj.ServiceConfigs[name]
		// kompose normalized this name itself to build the selector it stamps on
		// every object of the service, so normalizing the object names cannot
		// repair it: a value Kubernetes rejects as a label fails the whole
		// workload at install time, and local lint has no way to see it.
		if len(validation.IsValidLabelValue(name)) > 0 {
			notices = append(notices, fmt.Sprintf(
				"compose service %q becomes the pod selector value %q, which Kubernetes rejects: rename the compose service to start and end with a letter or digit", service.Name, name))
		}
		switch {
		case service.Image == "":
			notices = append(notices, fmt.Sprintf(
				"service %q declares no image, so its template references %q, which cannot be pulled: build and push an image first", name, name))
		case service.Build != "":
			notices = append(notices, fmt.Sprintf(
				"service %q builds %q locally: push that tag to a registry Olares can reach, for every architecture in spec.supportArch", name, service.Image))
		}
		// Pinning the controller costs the same compose semantics whichever
		// workload kind a service ends up as, so the notices below name that kind
		// rather than assuming the Deployment the conversion defaults to.
		if kind := renderedWorkloadKind(service); kind != "" {
			if service.CronJobSchedule != "" {
				notices = append(notices, fmt.Sprintf(
					"service %q sets kompose.cronjob.schedule %q, which was dropped: it renders as an always-on %s, so either schedule the work inside the container or add a CronJob template by hand and leave it out of workloadReplicas", name, service.CronJobSchedule, kind))
			}
			if service.Restart == "no" || service.Restart == "on-failure" {
				notices = append(notices, fmt.Sprintf(
					"service %q sets restart: %s, but was rendered as a %s with restartPolicy Always: Olares installs, suspends and resumes apps by scaling replicas", name, service.Restart, kind))
			}
			if service.DeployMode == "global" && kind != "DaemonSet" {
				notices = append(notices, fmt.Sprintf(
					"service %q sets deploy.mode: global, but was rendered as a %s rather than a DaemonSet, for the same reason", name, kind))
			}
		}
		if isDatastoreImage(service.Image) {
			notices = append(notices, fmt.Sprintf(
				"service %q runs %q, which looks like a bundled datastore: replace it with Olares system middleware plus an options.dependencies entry, and drop its workload and volumes", name, service.Image))
		}
		notices = append(notices, bindMountNotices(name, service)...)
	}
	return notices
}

// bindMountNotices flags compose bind mounts, which kompose either skips, drops
// or copies into the chart; none of the three behaves like app data on Olares.
func bindMountNotices(name string, service kobject.ServiceConfig) []string {
	notices := make([]string, 0)
	for _, vol := range service.Volumes {
		if vol.Host == "" {
			continue
		}
		remodel := "Re-model it as a userspace volume ({{ .Values.userspace.appData }}) with the matching permission block"
		var detail string
		switch {
		case strings.HasSuffix(vol.Host, ".sock"):
			detail = "kompose skips socket paths, so the container gets no volume and no mount at all"
			remodel = "An Olares app cannot reach a host socket, so drop the mount and rethink whatever needed it"
		case hostPathBecomesConfigMap(vol.Host):
			detail = "kompose copied the host path contents into a ConfigMap in the chart instead of mounting it"
		default:
			detail = "kompose dropped the host path and generated a PVC, so the container starts against an empty volume"
		}
		notices = append(notices, fmt.Sprintf("service %q bind-mounts %s at %s: %s. %s",
			name, vol.Host, vol.Container, detail, remodel))
	}
	return notices
}

// hostPathBecomesConfigMap mirrors kompose's isConfigFile: only a regular file or
// a non-empty directory is copied into a ConfigMap, so an empty or missing
// directory still ends up as a PVC.
func hostPathBecomesConfigMap(path string) bool {
	info, err := os.Stat(path)
	if err != nil {
		return false
	}
	if info.Mode().IsRegular() {
		return true
	}
	entries, err := os.ReadDir(path)
	if err != nil {
		return false
	}
	return len(entries) > 0
}

// renderedWorkloadKind names the workload a service comes out as: the conversion
// pins Deployment and only a kompose.controller.type label moves it. kompose
// compares that label verbatim against its own lowercase controller names, so
// every other value — a differently cased one included — leaves the service with
// no workload at all, which is reported as the empty string.
func renderedWorkloadKind(service kobject.ServiceConfig) string {
	controller, labeled := service.Labels[komposeControllerLabel]
	if !labeled {
		return "Deployment"
	}
	switch controller {
	case kubernetes.DeploymentController:
		return "Deployment"
	case kubernetes.StatefulStateController:
		return "StatefulSet"
	case kubernetes.DaemonSetController:
		return "DaemonSet"
	}
	return ""
}

// storageNotices flags the PVCs kompose derives from compose volumes: Olares apps
// normally persist under the user's own directories instead.
func storageNotices(resources []runtime.Object) []string {
	claims := make([]string, 0)
	for _, r := range resources {
		if pvc, ok := r.(*corev1.PersistentVolumeClaim); ok {
			claims = append(claims, pvc.GetName())
		}
	}
	if len(claims) == 0 {
		return nil
	}
	return []string{fmt.Sprintf(
		"chart declares %d PersistentVolumeClaim(s) (%s), each requesting kompose's 100Mi default: map the ones holding app data onto {{ .Values.userspace.appData }} or appCache with the matching permission block and strategy: Recreate, and delete the ones belonging to a bundled datastore",
		len(claims), strings.Join(claims, ", "))}
}

// writeManifest assembles the OlaresManifest.yaml + Chart.yaml + values.yaml.
func writeManifest(opts Options, entranceHost string, entrancePort int32, totalRequests, totalLimits corev1.ResourceList, replicas manifest.WorkloadReplicas) error {
	cpuReq := totalRequests[corev1.ResourceCPU]
	memReq := totalRequests[corev1.ResourceMemory]
	cpuLim := totalLimits[corev1.ResourceCPU]
	memLim := totalLimits[corev1.ResourceMemory]

	appcfg := manifest.AppConfiguration{
		APIVersion:    appAPIVersion,
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

// matchService returns the service selecting podLabels over a TCP port.
func matchService(services []*corev1.Service, podLabels map[string]string) *corev1.Service {
	for _, s := range services {
		if isSelectorMatch(podLabels, s.Spec.Selector) && servicePort(s) > 0 {
			return s
		}
	}
	return nil
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

// normalizeRestartPolicy resets the pod-template restart policy kompose copies
// from compose: Kubernetes only accepts Always on a replica-controlled template,
// so a service with restart: no would otherwise render an unschedulable workload
// that local lint still accepts.
func normalizeRestartPolicy(resource runtime.Object) {
	switch obj := resource.(type) {
	case *appsv1.Deployment:
		obj.Spec.Template.Spec.RestartPolicy = corev1.RestartPolicyAlways
	case *appsv1.StatefulSet:
		obj.Spec.Template.Spec.RestartPolicy = corev1.RestartPolicyAlways
	case *appsv1.DaemonSet:
		obj.Spec.Template.Spec.RestartPolicy = corev1.RestartPolicyAlways
	}
}

func addResourcesRequirements(resource runtime.Object) {
	if spec := podSpecOf(resource); spec != nil {
		addResourcesToContainers(spec.Containers, defaultRequests, defaultLimits)
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

func ptrInt32(v int32) *int32 { return &v }

var (
	replicasLineRE = regexp.MustCompile(`(?m)^(\s*)replicas: \d+\s*$`)

	// templateFieldNameRE matches the workload names Helm can reach with dotted
	// field syntax. Anything else — notably the dashed names kompose derives
	// from compose services like web_app — has to go through index, because
	// text/template rejects a dash inside a field name.
	templateFieldNameRE = regexp.MustCompile(`^[A-Za-z_][A-Za-z0-9_]*$`)
)

// wireReplicasValue rewrites a serialized Deployment/StatefulSet's literal
// spec.replicas line into a Helm reference to .Values.workloads.<name>.replicaCount.
// app-service installs and suspends/resumes apps by overriding that value, so a
// hardcoded replicas count would make the lifecycle scale operations inert.
func wireReplicasValue(yml []byte, name string) []byte {
	repl := fmt.Sprintf("${1}replicas: {{ %s }}", replicaCountRef(name))
	return replicasLineRE.ReplaceAll(yml, []byte(repl))
}

// replicaCountRef renders the Helm expression for one workload's replica count.
func replicaCountRef(name string) string {
	if templateFieldNameRE.MatchString(name) {
		return fmt.Sprintf(".Values.workloads.%s.replicaCount", name)
	}
	return fmt.Sprintf("(index .Values.workloads %q).replicaCount", name)
}

func writeYAMLFile(path string, v any) error {
	data, err := toYAML(v)
	if err != nil {
		return err
	}
	return os.WriteFile(path, data, 0644)
}
