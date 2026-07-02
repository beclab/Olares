package upgrade

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"
	"time"

	"github.com/beclab/Olares/cli/pkg/common"
	"github.com/beclab/Olares/cli/pkg/core/connector"
	"github.com/beclab/Olares/cli/pkg/core/logger"
	appv1alpha1 "github.com/beclab/Olares/framework/app-service/api/app.bytetrade.io/v1alpha1"
	"github.com/beclab/Olares/framework/app-service/pkg/appcfg"
	"github.com/pkg/errors"
	appsv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
	apierrors "k8s.io/apimachinery/pkg/api/errors"
	apimeta "k8s.io/apimachinery/pkg/api/meta"
	"k8s.io/apimachinery/pkg/api/resource"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/types"
	ctrl "sigs.k8s.io/controller-runtime"
	ctrlclient "sigs.k8s.io/controller-runtime/pkg/client"
)

const (
	legacyGPUBindingAPIVersion = "gpu.bytetrade.io/v1alpha1"
	legacyGPUBindingListKind   = "GPUBindingList"
	gpuAllocationCMNamespace   = "os-framework"
	gpuAllocationCMName        = "app-gpu-allocations"
	gpuAllocationCMKey         = "allocations.json"
	nodeGPUTypeLabel           = "gpu.bytetrade.io/type"
	nodeNvidiaRegistryKey      = "hami.io/node-nvidia-register"
	deviceSplitSymbol          = ":"
	appNameLabelKey            = "applications.app.bytetrade.io/name"
	appOwnerLabelKey           = "applications.app.bytetrade.io/owner"
	appInstallUserLabelKey     = "applications.app.bytetrade.io/install_user"
	managedByLabelKey          = "app.bytetrade.io/managed-by"
	managedByAppService        = "app-service"
	allocationModeLabelKey     = "gpu.bytetrade.io/mode"
	shareModeAnnotationPrefix  = "sharemode.gpu.bytetrade.io/"
	supportTypeExclusive       = "Exclusive"
	supportTypeMemorySlice     = "MemorySlice"
	supportTypeTimeSlice       = "TimeSlice"
	sharedNamespaceSuffix      = "-shared"

	// gpuBindingMemoryBytesPerMiB converts a GPUBinding spec.memory value to
	// bytes. HAMi expresses the vGPU memory limit in MiB (the quantity holds a
	// bare MiB number, e.g. 2048 for 2GiB — the same value app-service writes
	// back via allocation.Memory/mib in compute.createHAMIBinding), while the
	// compute-allocation model stores memory in bytes.
	gpuBindingMemoryBytesPerMiB = int64(1024 * 1024)
)

type migrateLegacyGPUBindings struct {
	common.KubeAction
}

type migratedComputeAllocation struct {
	AppID    string `json:"appId,omitempty"`
	AppName  string `json:"appName"`
	Owner    string `json:"owner,omitempty"`
	Mode     string `json:"mode"`
	NodeName string `json:"nodeName"`
	DeviceID string `json:"deviceId"`
	Memory   int64  `json:"memory"`
}

func (m *migrateLegacyGPUBindings) Execute(_ connector.Runtime) error {
	config, err := ctrl.GetConfig()
	if err != nil {
		return errors.Wrap(errors.WithStack(err), "failed to get kubernetes config")
	}

	scheme := runtime.NewScheme()
	if err := appv1alpha1.AddToScheme(scheme); err != nil {
		return errors.Wrap(errors.WithStack(err), "failed to add app-service scheme")
	}
	if err := corev1.AddToScheme(scheme); err != nil {
		return errors.Wrap(errors.WithStack(err), "failed to add corev1 scheme")
	}
	if err := appsv1.AddToScheme(scheme); err != nil {
		return errors.Wrap(errors.WithStack(err), "failed to add apps/v1 scheme")
	}

	c, err := ctrlclient.New(config, ctrlclient.Options{Scheme: scheme})
	if err != nil {
		return errors.Wrap(errors.WithStack(err), "failed to create controller-runtime client")
	}

	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Minute)
	defer cancel()

	bindings := &unstructured.UnstructuredList{}
	bindings.SetAPIVersion(legacyGPUBindingAPIVersion)
	bindings.SetKind(legacyGPUBindingListKind)
	if err := c.List(ctx, bindings); err != nil {
		if apierrors.IsNotFound(err) || runtime.IsNotRegisteredError(err) || apimeta.IsNoMatchError(err) {
			logger.Infof("legacy GPUBinding resource not found, skip migration")
			return nil
		}
		return errors.Wrap(errors.WithStack(err), "failed to list legacy GPUBindings")
	}
	if len(bindings.Items) == 0 {
		logger.Infof("no legacy GPUBindings found, skip migration")
		return nil
	}

	var managers appv1alpha1.ApplicationManagerList
	if err := c.List(ctx, &managers); err != nil {
		return errors.Wrap(errors.WithStack(err), "failed to list applicationmanagers")
	}
	var nodes corev1.NodeList
	if err := c.List(ctx, &nodes); err != nil {
		return errors.Wrap(errors.WithStack(err), "failed to list nodes")
	}
	allocations, err := loadMigratedAllocations(ctx, c)
	if err != nil {
		return err
	}
	allocationKeys := make(map[string]struct{}, len(allocations))
	for _, allocation := range allocations {
		allocationKeys[allocationKey(allocation.AppName, allocation.Owner, allocation.DeviceID)] = struct{}{}
	}

	var migrated, deleted, skipped int
	for i := range bindings.Items {
		binding := &bindings.Items[i]
		appName, _, _ := unstructured.NestedString(binding.Object, "spec", "appName")
		deviceID, _, _ := unstructured.NestedString(binding.Object, "spec", "uuid")
		if appName == "" || deviceID == "" {
			logger.Warnf("skip GPUBinding %s with empty appName or uuid", binding.GetName())
			skipped++
			continue
		}

		owner, _, _ := unstructured.NestedString(binding.Object, "spec", "owner")
		namespace, _, _ := unstructured.NestedString(binding.Object, "spec", "namespace")
		am, appConfig, err := resolveBindingApplication(ctx, c, &managers, binding, appName, owner, namespace)
		if err != nil {
			return err
		}
		if am == nil {
			if err := c.Delete(ctx, binding); err != nil && !apierrors.IsNotFound(err) {
				return errors.Wrapf(errors.WithStack(err), "failed to delete orphan GPUBinding %s", binding.GetName())
			}
			logger.Infof("deleted orphan legacy GPUBinding %s for app %s", binding.GetName(), appName)
			deleted++
			continue
		}

		if !shouldMigrateLegacyGPUBinding(am.Status.State.String()) {
			// A non-holding AM state normally means the binding is stale. But
			// for a v2 cluster-shared app the binding belongs to the shared
			// *server*, which keeps holding its GPU as long as its workloads
			// are running — independent of this AM's declared state. The
			// resolved AM (the install_user's) can be Stopped after a
			// client-only stop while the server stays up for other users, and a
			// stop-all would have scaled the server to zero. So only drop the
			// binding when the shared server is actually not running; otherwise
			// migrate it.
			serverRunning := false
			if appConfig != nil && appConfig.IsV2() && appConfig.HasClusterSharedCharts() {
				running, err := sharedServerRunning(ctx, c, appConfig)
				if err != nil {
					return err
				}
				serverRunning = running
			}
			if !serverRunning {
				if err := c.Delete(ctx, binding); err != nil && !apierrors.IsNotFound(err) {
					return errors.Wrapf(errors.WithStack(err), "failed to delete stale GPUBinding %s", binding.GetName())
				}
				logger.Infof("deleted legacy GPUBinding %s for app %s owner %s in state %s", binding.GetName(), am.Spec.AppName, am.Spec.AppOwner, am.Status.State)
				deleted++
				continue
			}
			logger.Infof("keeping legacy GPUBinding %s for app %s owner %s: shared server still running despite AM state %s", binding.GetName(), am.Spec.AppName, am.Spec.AppOwner, am.Status.State)
		}

		nodeName, supportType := findDevicePlacement(&nodes, deviceID)
		if nodeName == "" {
			logger.Warnf("skip GPUBinding %s because device %s was not found on any node", binding.GetName(), deviceID)
			skipped++
			continue
		}
		mode := strings.TrimSpace(appConfig.SelectedGpuType)
		if mode == "" {
			mode = "nvidia"
		}
		var memory int64
		if supportType == supportTypeMemorySlice {
			memoryRaw, _, _ := unstructured.NestedFieldNoCopy(binding.Object, "spec", "memory")
			memoryMiB, memErr := bindingMemory(memoryRaw)
			if memErr != nil {
				return errors.Wrapf(errors.WithStack(memErr), "failed to parse memory of GPUBinding %s", binding.GetName())
			}
			if memoryMiB <= 0 {
				logger.Warnf("skip MemorySlice GPUBinding %s because spec.memory is empty", binding.GetName())
				skipped++
				continue
			}
			// GPUBinding.spec.memory is the vGPU memory limit in MiB, but the
			// compute-allocation model stores Allocation.Memory in bytes (it is
			// subtracted from the byte-scale device capacity during
			// availability/scheduling and divided back by 1MiB when the binding
			// is re-emitted). Convert MiB -> bytes here; without it the migrated
			// allocation is ~1e6x too small — it renders as 0Gi in the UI and
			// makes the device look almost entirely free, risking GPU
			// over-allocation.
			memory = memoryMiB * gpuBindingMemoryBytesPerMiB
		}
		key := allocationKey(am.Spec.AppName, am.Spec.AppOwner, deviceID)
		if _, exists := allocationKeys[key]; !exists {
			allocations = append(allocations, migratedComputeAllocation{
				AppID:    appConfig.AppID,
				AppName:  am.Spec.AppName,
				Owner:    am.Spec.AppOwner,
				Mode:     mode,
				NodeName: nodeName,
				DeviceID: deviceID,
				Memory:   memory,
			})
			allocationKeys[key] = struct{}{}
			migrated++
		}
		if err := patchMigratedGPUBinding(ctx, c, binding, am, mode); err != nil {
			return err
		}
	}

	if err := saveMigratedAllocations(ctx, c, allocations); err != nil {
		return err
	}
	logger.Infof("migrated legacy GPUBindings: migrated=%d deleted=%d skipped=%d", migrated, deleted, skipped)
	return nil
}

func patchMigratedGPUBinding(ctx context.Context, c ctrlclient.Client, binding *unstructured.Unstructured, am *appv1alpha1.ApplicationManager, mode string) error {
	patch := binding.DeepCopy()
	labels := patch.GetLabels()
	if labels == nil {
		labels = make(map[string]string)
	}
	labels[managedByLabelKey] = managedByAppService
	labels[appNameLabelKey] = am.Spec.AppName
	labels[appOwnerLabelKey] = am.Spec.AppOwner
	labels[allocationModeLabelKey] = mode
	patch.SetLabels(labels)

	if err := unstructured.SetNestedField(patch.Object, am.Spec.AppName, "spec", "appName"); err != nil {
		return errors.Wrapf(errors.WithStack(err), "failed to patch appName for GPUBinding %s", binding.GetName())
	}
	if err := unstructured.SetNestedField(patch.Object, am.Spec.AppOwner, "spec", "owner"); err != nil {
		return errors.Wrapf(errors.WithStack(err), "failed to patch owner for GPUBinding %s", binding.GetName())
	}
	if err := unstructured.SetNestedStringMap(patch.Object, map[string]string{
		appNameLabelKey:  am.Spec.AppName,
		appOwnerLabelKey: am.Spec.AppOwner,
	}, "spec", "podSelector", "matchLabels"); err != nil {
		return errors.Wrapf(errors.WithStack(err), "failed to patch podSelector for GPUBinding %s", binding.GetName())
	}

	if err := c.Patch(ctx, patch, ctrlclient.MergeFrom(binding)); err != nil {
		return errors.Wrapf(errors.WithStack(err), "failed to patch migrated GPUBinding %s", binding.GetName())
	}
	return nil
}

func resolveBindingApplication(ctx context.Context, c ctrlclient.Client, managers *appv1alpha1.ApplicationManagerList, binding *unstructured.Unstructured, appName, owner, namespace string) (*appv1alpha1.ApplicationManager, *appcfg.ApplicationConfig, error) {
	// An explicitly recorded owner (new-schema binding, or a prior migration
	// pass) is authoritative.
	if owner != "" {
		return findApplicationManager(managers, appName, owner)
	}
	// v2 client-server apps with a cluster-shared subchart can be installed by
	// multiple users (one ApplicationManager each) while only a single server
	// side actually holds the GPU. app-service attributes that GPU to the
	// "shared server owner" — the install_user label on the shared namespace
	// (see compute.sharedServerOwner). This is the authoritative owner and must
	// be resolved BEFORE the namespace / pod-selector heuristics below: those
	// inspect the (legacy) binding's podSelector, which can carry a client
	// user's owner label or match a client user's pods/namespace, and would
	// otherwise attribute the shared server's GPU to a client user instead of
	// the install_user. ownerFromSharedServer returns "" for non-shared apps,
	// so this is a no-op outside the v2 cluster-shared case. Only adopt it when
	// the install_user actually has a matching ApplicationManager; otherwise
	// (e.g. a stale install_user label) fall through to the heuristics.
	if sharedOwner := ownerFromSharedServer(ctx, c, managers, appName); sharedOwner != "" {
		if am, cfg, err := findApplicationManager(managers, appName, sharedOwner); err != nil || am != nil {
			return am, cfg, err
		}
	}
	if namespace != "" {
		if am, cfg, err := findApplicationManagerByNamespace(managers, namespace); err != nil || am != nil {
			return am, cfg, err
		}
		owner = ownerFromNamespace(ctx, c, namespace, appName)
		if owner != "" {
			return findApplicationManager(managers, appName, owner)
		}
	}
	if owner = ownerFromBindingSelector(ctx, c, binding, managers); owner != "" {
		return findApplicationManager(managers, appName, owner)
	}
	return findSingleApplicationManager(managers, appName)
}

// ownerFromSharedServer resolves the shared-server owner for a v2
// client-server app whose GPU is held by a single, cluster-shared server side.
// It mirrors compute.sharedServerOwner: for each ApplicationManager matching
// appName that declares a v2 cluster-shared chart, it reads the install_user
// label on the shared namespace (<sharedChart>-shared) and returns it. Returns
// "" when the app is not a v2 shared app or the shared namespace is missing.
func ownerFromSharedServer(ctx context.Context, c ctrlclient.Client, managers *appv1alpha1.ApplicationManagerList, appName string) string {
	for i := range managers.Items {
		am := &managers.Items[i]
		if am.Spec.AppName != appName {
			continue
		}
		cfg, err := decodeAppConfig(am)
		if err != nil || cfg == nil {
			continue
		}
		if !cfg.IsV2() || !cfg.HasClusterSharedCharts() {
			continue
		}
		for j := range cfg.SubCharts {
			chart := cfg.SubCharts[j]
			if !chart.Shared {
				continue
			}
			// Mirror appcfg.ChartNamespace for shared charts: the shared
			// server always lives in "<chartName>-shared" regardless of
			// the installing user.
			var ns corev1.Namespace
			if err := c.Get(ctx, types.NamespacedName{Name: chart.Name + sharedNamespaceSuffix}, &ns); err != nil {
				continue
			}
			if name := ns.Labels[appNameLabelKey]; name != "" && name != appName {
				continue
			}
			if owner := ns.Labels[appInstallUserLabelKey]; owner != "" {
				return owner
			}
		}
	}
	return ""
}

// sharedServerRunning reports whether the v2 cluster-shared server side of
// appConfig still has running workloads (any shared-chart Deployment /
// StatefulSet in the <chart>-shared namespace with a non-zero replica count).
// A nil replica count is treated as running (defaults to 1). It is the
// migration-side counterpart to compute.SharedServerSuspended and lets the
// GPUBinding migration keep a shared server's binding while the server is up,
// even when the resolved ApplicationManager reports a stopped state (a
// client-only stop leaves the server running for other users).
func sharedServerRunning(ctx context.Context, c ctrlclient.Client, appConfig *appcfg.ApplicationConfig) (bool, error) {
	if appConfig == nil || !appConfig.IsV2() || !appConfig.HasClusterSharedCharts() {
		return false, nil
	}
	for _, chart := range appConfig.SubCharts {
		if !chart.Shared {
			continue
		}
		namespace := chart.Name + sharedNamespaceSuffix
		var deployments appsv1.DeploymentList
		if err := c.List(ctx, &deployments, ctrlclient.InNamespace(namespace)); err != nil {
			if apierrors.IsNotFound(err) {
				continue
			}
			return false, errors.Wrapf(errors.WithStack(err), "failed to list deployments in shared namespace %s", namespace)
		}
		for i := range deployments.Items {
			if replicas := deployments.Items[i].Spec.Replicas; replicas == nil || *replicas > 0 {
				return true, nil
			}
		}
		var statefulSets appsv1.StatefulSetList
		if err := c.List(ctx, &statefulSets, ctrlclient.InNamespace(namespace)); err != nil {
			if apierrors.IsNotFound(err) {
				continue
			}
			return false, errors.Wrapf(errors.WithStack(err), "failed to list statefulsets in shared namespace %s", namespace)
		}
		for i := range statefulSets.Items {
			if replicas := statefulSets.Items[i].Spec.Replicas; replicas == nil || *replicas > 0 {
				return true, nil
			}
		}
	}
	return false, nil
}

func findApplicationManager(managers *appv1alpha1.ApplicationManagerList, appName, owner string) (*appv1alpha1.ApplicationManager, *appcfg.ApplicationConfig, error) {
	for i := range managers.Items {
		am := &managers.Items[i]
		if am.Spec.AppName != appName || am.Spec.AppOwner != owner {
			continue
		}
		cfg, err := decodeAppConfig(am)
		return am, cfg, err
	}
	return nil, nil, nil
}

func findApplicationManagerByNamespace(managers *appv1alpha1.ApplicationManagerList, namespace string) (*appv1alpha1.ApplicationManager, *appcfg.ApplicationConfig, error) {
	for i := range managers.Items {
		am := &managers.Items[i]
		if am.Spec.AppNamespace != namespace {
			continue
		}
		cfg, err := decodeAppConfig(am)
		return am, cfg, err
	}
	return nil, nil, nil
}

func findSingleApplicationManager(managers *appv1alpha1.ApplicationManagerList, appName string) (*appv1alpha1.ApplicationManager, *appcfg.ApplicationConfig, error) {
	var found *appv1alpha1.ApplicationManager
	for i := range managers.Items {
		am := &managers.Items[i]
		if am.Spec.AppName != appName {
			continue
		}
		if found != nil {
			return nil, nil, fmt.Errorf("cannot infer owner for legacy GPUBinding app %s: multiple applicationmanagers found", appName)
		}
		found = am
	}
	if found == nil {
		return nil, nil, nil
	}
	cfg, err := decodeAppConfig(found)
	return found, cfg, err
}

func decodeAppConfig(am *appv1alpha1.ApplicationManager) (*appcfg.ApplicationConfig, error) {
	if am.Spec.Config == "" {
		return &appcfg.ApplicationConfig{}, nil
	}
	var cfg appcfg.ApplicationConfig
	if err := json.Unmarshal([]byte(am.Spec.Config), &cfg); err != nil {
		return nil, errors.Wrapf(errors.WithStack(err), "failed to unmarshal config for applicationmanager %s", am.Name)
	}
	return &cfg, nil
}

func ownerFromNamespace(ctx context.Context, c ctrlclient.Client, namespace, appName string) string {
	var ns corev1.Namespace
	if err := c.Get(ctx, types.NamespacedName{Name: namespace}, &ns); err != nil {
		return ""
	}
	if appName != "" && ns.Labels[appNameLabelKey] != "" && ns.Labels[appNameLabelKey] != appName {
		return ""
	}
	if owner := ns.Labels[appInstallUserLabelKey]; owner != "" {
		return owner
	}
	return ns.Labels[appOwnerLabelKey]
}

func ownerFromBindingSelector(ctx context.Context, c ctrlclient.Client, binding *unstructured.Unstructured, managers *appv1alpha1.ApplicationManagerList) string {
	labels, _, _ := unstructured.NestedStringMap(binding.Object, "spec", "podSelector", "matchLabels")
	if owner := labels[appOwnerLabelKey]; owner != "" {
		return owner
	}
	selector := ctrlclient.MatchingLabels(labels)
	var pods corev1.PodList
	if len(labels) == 0 || c.List(ctx, &pods, selector) != nil {
		return ""
	}
	for _, pod := range pods.Items {
		if am, _, _ := findApplicationManagerByNamespace(managers, pod.Namespace); am != nil {
			return am.Spec.AppOwner
		}
		if owner := ownerFromNamespace(ctx, c, pod.Namespace, ""); owner != "" {
			return owner
		}
	}
	return ""
}

// shouldMigrateLegacyGPUBinding decides whether a legacy GPUBinding should be
// carried over to the new compute-allocation configmap (true) or dropped as
// stale (false). The list intentionally covers every state in which the app
// is still considered to be holding its GPU allocation, including the various
// *Failed states where the in-progress op aborted but the binding wasn't
// torn down.
//
// IMPORTANT: app states are lowercased ("running", "stopFailed", ...) per
// `api/app.bytetrade.io/v1alpha1/appmanager_states.go`. Earlier versions of
// this function compared against capitalized names and therefore matched
// nothing, deleting every legacy binding instead of migrating them.
func shouldMigrateLegacyGPUBinding(state string) bool {
	switch state {
	case "running",
		"initializing",
		"resuming",
		"installing",
		"upgrading",
		"applyingEnv",
		"stopFailed",
		"resumeFailed",
		"upgradeFailed",
		"applyEnvFailed":
		return true
	default:
		return false
	}
}

func bindingMemory(raw any) (int64, error) {
	if raw == nil {
		return 0, nil
	}
	switch value := raw.(type) {
	case int64:
		return value, nil
	case int:
		return int64(value), nil
	case int32:
		return int64(value), nil
	case float64:
		return int64(value), nil
	case string:
		if value == "" {
			return 0, nil
		}
		q, err := resource.ParseQuantity(value)
		if err != nil {
			return 0, err
		}
		return q.Value(), nil
	default:
		return 0, fmt.Errorf("unsupported memory type %T", raw)
	}
}

func findDevicePlacement(nodes *corev1.NodeList, deviceID string) (string, string) {
	for _, node := range nodes.Items {
		raw := node.Annotations[nodeNvidiaRegistryKey]
		for _, encoded := range strings.Split(raw, deviceSplitSymbol) {
			fields := strings.Split(encoded, ",")
			if len(fields) > 0 && fields[0] == deviceID {
				return node.Name, shareModeToSupportType(node.Labels[nodeGPUTypeLabel], node.Annotations[shareModeAnnotationPrefix+deviceID])
			}
		}
	}
	return "", ""
}

func shareModeToSupportType(gpuType, mode string) string {
	switch mode {
	case "0":
		return supportTypeExclusive
	case "1":
		return supportTypeMemorySlice
	default:
		if strings.TrimSpace(gpuType) == "nvidia-gb10" {
			return supportTypeMemorySlice
		}
		return supportTypeTimeSlice
	}
}

func loadMigratedAllocations(ctx context.Context, c ctrlclient.Client) ([]migratedComputeAllocation, error) {
	var cm corev1.ConfigMap
	err := c.Get(ctx, types.NamespacedName{Namespace: gpuAllocationCMNamespace, Name: gpuAllocationCMName}, &cm)
	if apierrors.IsNotFound(err) {
		return nil, nil
	}
	if err != nil {
		return nil, errors.Wrap(errors.WithStack(err), "failed to get GPU allocation configmap")
	}
	raw := cm.Data[gpuAllocationCMKey]
	if raw == "" {
		return nil, nil
	}
	var allocations []migratedComputeAllocation
	if err := json.Unmarshal([]byte(raw), &allocations); err != nil {
		return nil, errors.Wrap(errors.WithStack(err), "failed to unmarshal GPU allocations")
	}
	return allocations, nil
}

func saveMigratedAllocations(ctx context.Context, c ctrlclient.Client, allocations []migratedComputeAllocation) error {
	data, err := json.Marshal(allocations)
	if err != nil {
		return errors.Wrap(errors.WithStack(err), "failed to marshal GPU allocations")
	}
	var cm corev1.ConfigMap
	err = c.Get(ctx, types.NamespacedName{Namespace: gpuAllocationCMNamespace, Name: gpuAllocationCMName}, &cm)
	if apierrors.IsNotFound(err) {
		cm = corev1.ConfigMap{
			ObjectMeta: metav1.ObjectMeta{
				Namespace: gpuAllocationCMNamespace,
				Name:      gpuAllocationCMName,
			},
			Data: map[string]string{gpuAllocationCMKey: string(data)},
		}
		return c.Create(ctx, &cm)
	}
	if err != nil {
		return errors.Wrap(errors.WithStack(err), "failed to get GPU allocation configmap")
	}
	if cm.Data == nil {
		cm.Data = make(map[string]string)
	}
	cm.Data[gpuAllocationCMKey] = string(data)
	return c.Update(ctx, &cm)
}

func allocationKey(appName, owner, deviceID string) string {
	return appName + "\x00" + owner + "\x00" + deviceID
}
