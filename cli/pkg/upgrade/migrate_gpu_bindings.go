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
			if err := c.Delete(ctx, binding); err != nil && !apierrors.IsNotFound(err) {
				return errors.Wrapf(errors.WithStack(err), "failed to delete stale GPUBinding %s", binding.GetName())
			}
			logger.Infof("deleted legacy GPUBinding %s for app %s owner %s in state %s", binding.GetName(), am.Spec.AppName, am.Spec.AppOwner, am.Status.State)
			deleted++
			continue
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
			memory, err = bindingMemory(memoryRaw)
			if err != nil {
				return errors.Wrapf(errors.WithStack(err), "failed to parse memory of GPUBinding %s", binding.GetName())
			}
			if memory <= 0 {
				logger.Warnf("skip MemorySlice GPUBinding %s because spec.memory is empty", binding.GetName())
				skipped++
				continue
			}
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
	if owner != "" {
		return findApplicationManager(managers, appName, owner)
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
