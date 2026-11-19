package upgrade

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"strings"
	"time"

	"github.com/beclab/Olares/cli/pkg/common"
	"github.com/beclab/Olares/cli/pkg/core/connector"
	"github.com/beclab/Olares/cli/pkg/core/logger"
	"github.com/pkg/errors"
	appsv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
	apierrors "k8s.io/apimachinery/pkg/api/errors"
	apimeta "k8s.io/apimachinery/pkg/api/meta"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/apimachinery/pkg/runtime"
	ctrl "sigs.k8s.io/controller-runtime"
	ctrlclient "sigs.k8s.io/controller-runtime/pkg/client"
)

const (
	intelGPUMode               = "intel-gpu"
	nodeIntelRegisterKey       = "bytetrade.io/node-intel-register"
	nodeIntelHAMIRegisterKey   = "hami.io/node-intel-register"
	nodeIntelCardsKey          = "gpu.bytetrade.io/node-intel-cards"
	legacyIntelDGPUResource    = "gpu.intel.com/xe"
	gpuBindingKind             = "GPUBinding"
	restartedAtAnnotation      = "kubectl.kubernetes.io/restartedAt"
	intelHAMIRegisterWait      = 5 * time.Minute
	intelHAMIRegisterPollDelay = 10 * time.Second
)

// migrateIntelDGPUToHAMI moves discrete Intel GPU apps from the plain
// gpu.intel.com/xe device plugin resource to HAMi. It runs after the HAMi
// chart, the Intel plugin and app-service are upgraded, and every step is
// safe to repeat:
//
//  1. wait for the plugin to publish hami.io/node-intel-register and the
//     card list on every node that has a discrete Intel GPU;
//  2. rewrite intel-gpu allocations from <node>-intel-gpu-<card> to the card
//     UUID, move the share-mode annotations and create the Exclusive
//     GPUBindings HAMi filters on;
//  3. restart the workloads of those apps that still request
//     gpu.intel.com/xe, so the gpu-limit webhook injects intel.com/vgpu.
type migrateIntelDGPUToHAMI struct {
	common.KubeAction
}

type intelCardInfo struct {
	UUID string `json:"uuid"`
	Card string `json:"card"`
	BDF  string `json:"bdf"`
}

func (m *migrateIntelDGPUToHAMI) Execute(_ connector.Runtime) error {
	config, err := ctrl.GetConfig()
	if err != nil {
		return errors.Wrap(errors.WithStack(err), "failed to get kubernetes config")
	}
	scheme := runtime.NewScheme()
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

	ctx, cancel := context.WithTimeout(context.Background(), intelHAMIRegisterWait+3*time.Minute)
	defer cancel()

	nodes, err := waitForIntelHAMIRegister(ctx, c, intelHAMIRegisterWait, intelHAMIRegisterPollDelay)
	if err != nil {
		return err
	}
	return migrateIntelDGPUAllocations(ctx, c, nodes, time.Now())
}

// waitForIntelHAMIRegister returns the nodes once every node with a discrete
// Intel GPU carries the HAMi register and card list, or when timeout passes.
// A node still missing them then keeps its allocations untouched so a later
// run can finish it.
func waitForIntelHAMIRegister(ctx context.Context, c ctrlclient.Client, timeout, delay time.Duration) (*corev1.NodeList, error) {
	deadline := time.Now().Add(timeout)
	for {
		var nodes corev1.NodeList
		if err := c.List(ctx, &nodes); err != nil {
			return nil, errors.Wrap(errors.WithStack(err), "failed to list nodes")
		}
		var pending []string
		for i := range nodes.Items {
			node := &nodes.Items[i]
			if hasLegacyIntelDGPU(node) && len(intelCardUUIDs(node)) == 0 {
				pending = append(pending, node.Name)
			}
		}
		if len(pending) == 0 {
			return &nodes, nil
		}
		if time.Now().After(deadline) {
			logger.Warnf("Intel device plugin has not registered with HAMi on nodes %v; their intel-gpu apps are left for a later upgrade run", pending)
			return &nodes, nil
		}
		logger.Infof("waiting for the Intel device plugin to register with HAMi on nodes %v", pending)
		select {
		case <-ctx.Done():
			return nil, ctx.Err()
		case <-time.After(delay):
		}
	}
}

func hasLegacyIntelDGPU(node *corev1.Node) bool {
	for _, entry := range strings.Split(node.Annotations[nodeIntelRegisterKey], deviceSplitSymbol) {
		if strings.HasPrefix(entry, "dgpu,") {
			return true
		}
	}
	return false
}

// intelCardUUIDs maps each card name (card0, ...) of the node to its HAMi
// UUID. It is empty until the plugin has registered the node with HAMi.
func intelCardUUIDs(node *corev1.Node) map[string]string {
	if node.Annotations[nodeIntelHAMIRegisterKey] == "" {
		return nil
	}
	var cards []intelCardInfo
	if err := json.Unmarshal([]byte(node.Annotations[nodeIntelCardsKey]), &cards); err != nil {
		return nil
	}
	uuids := make(map[string]string, len(cards))
	for _, card := range cards {
		if card.Card != "" && card.UUID != "" {
			uuids[card.Card] = card.UUID
		}
	}
	return uuids
}

// legacyIntelCardUUID resolves a <node>-intel-gpu-<card> device ID. Builds
// before the per-card register named the only device <node>-intel-gpu-0,
// which maps to the node's card when it has exactly one.
func legacyIntelCardUUID(deviceID, nodeName string, uuids map[string]string) (string, bool) {
	card, ok := strings.CutPrefix(deviceID, nodeName+"-"+intelGPUMode+"-")
	if !ok {
		return "", false
	}
	if uuid, ok := uuids[card]; ok {
		return uuid, true
	}
	if card == "0" && len(uuids) == 1 {
		for _, uuid := range uuids {
			return uuid, true
		}
	}
	return "", false
}

func migrateIntelDGPUAllocations(ctx context.Context, c ctrlclient.Client, nodes *corev1.NodeList, now time.Time) error {
	byName := make(map[string]*corev1.Node, len(nodes.Items))
	for i := range nodes.Items {
		byName[nodes.Items[i].Name] = &nodes.Items[i]
	}

	allocations, err := loadMigratedAllocations(ctx, c)
	if err != nil {
		return err
	}
	var rewritten, unresolved int
	migratedApps := make(map[string]struct{})
	pendingApps := make(map[string]struct{})
	for i := range allocations {
		allocation := &allocations[i]
		if allocation.Mode != intelGPUMode {
			continue
		}
		app := appOwnerKey(allocation.AppName, allocation.Owner)
		var uuids map[string]string
		if node := byName[allocation.NodeName]; node != nil {
			uuids = intelCardUUIDs(node)
		}
		if uuid, ok := legacyIntelCardUUID(allocation.DeviceID, allocation.NodeName, uuids); ok {
			allocation.DeviceID = uuid
			allocation.Memory = 0
			rewritten++
		} else if !isIntelCardUUID(allocation.DeviceID, uuids) {
			logger.Warnf("intel-gpu allocation of app %s owner %s on device %s has no HAMi card yet", allocation.AppName, allocation.Owner, allocation.DeviceID)
			pendingApps[app] = struct{}{}
			unresolved++
			continue
		}
		migratedApps[app] = struct{}{}
	}
	for app := range pendingApps {
		delete(migratedApps, app)
	}
	if rewritten > 0 {
		if err := saveMigratedAllocations(ctx, c, allocations); err != nil {
			return err
		}
	}

	for _, allocation := range allocations {
		if _, ok := migratedApps[appOwnerKey(allocation.AppName, allocation.Owner)]; !ok || allocation.Mode != intelGPUMode {
			continue
		}
		if err := ensureExclusiveGPUBinding(ctx, c, allocation); err != nil {
			return err
		}
	}
	for _, node := range byName {
		if err := migrateIntelShareModeAnnotations(ctx, c, node); err != nil {
			return err
		}
	}
	restarted, err := restartLegacyIntelWorkloads(ctx, c, migratedApps, now)
	if err != nil {
		return err
	}
	logger.Infof("migrated intel-gpu apps to HAMi: allocations rewritten=%d unresolved=%d workloads restarted=%d", rewritten, unresolved, restarted)
	return nil
}

func isIntelCardUUID(deviceID string, uuids map[string]string) bool {
	for _, uuid := range uuids {
		if uuid == deviceID {
			return true
		}
	}
	return false
}

func appOwnerKey(appName, owner string) string {
	return appName + "\x00" + owner
}

// ensureExclusiveGPUBinding writes the binding app-service creates when it
// binds an intel-gpu app; the name and labels must match its
// compute.createHAMIBinding so a later rebind replaces this one.
func ensureExclusiveGPUBinding(ctx context.Context, c ctrlclient.Client, allocation migratedComputeAllocation) error {
	obj := &unstructured.Unstructured{}
	obj.SetAPIVersion(legacyGPUBindingAPIVersion)
	obj.SetKind(gpuBindingKind)
	obj.SetName(gpuBindingName(allocation.AppName, allocation.Owner, allocation.DeviceID))
	obj.SetLabels(map[string]string{
		managedByLabelKey:      managedByAppService,
		appNameLabelKey:        allocation.AppName,
		appOwnerLabelKey:       allocation.Owner,
		allocationModeLabelKey: allocation.Mode,
	})
	obj.Object["spec"] = map[string]any{
		"uuid":    allocation.DeviceID,
		"appName": allocation.AppName,
		"owner":   allocation.Owner,
		"podSelector": map[string]any{
			"matchLabels": map[string]any{
				appNameLabelKey:  allocation.AppName,
				appOwnerLabelKey: allocation.Owner,
			},
		},
	}
	err := c.Create(ctx, obj)
	if apierrors.IsAlreadyExists(err) {
		current := &unstructured.Unstructured{}
		current.SetAPIVersion(legacyGPUBindingAPIVersion)
		current.SetKind(gpuBindingKind)
		if err := c.Get(ctx, ctrlclient.ObjectKey{Name: obj.GetName()}, current); err != nil {
			return errors.Wrapf(errors.WithStack(err), "failed to get GPUBinding %s", obj.GetName())
		}
		current.SetLabels(obj.GetLabels())
		current.Object["spec"] = obj.Object["spec"]
		err = c.Update(ctx, current)
	}
	if err != nil {
		if apimeta.IsNoMatchError(err) {
			return errors.Wrap(err, "GPUBinding CRD is not installed")
		}
		return errors.Wrapf(errors.WithStack(err), "failed to write GPUBinding for app %s owner %s on %s", allocation.AppName, allocation.Owner, allocation.DeviceID)
	}
	return nil
}

func gpuBindingName(appName, owner, deviceID string) string {
	raw := fmt.Sprintf("%s-%s-%s", appName, owner, deviceID)
	sum := sha256.Sum256([]byte(raw))
	suffix := "-" + hex.EncodeToString(sum[:])[:8]

	var b strings.Builder
	for _, r := range strings.ToLower(raw) {
		switch {
		case r >= 'a' && r <= 'z', r >= '0' && r <= '9', r == '-', r == '.':
			b.WriteRune(r)
		default:
			b.WriteRune('-')
		}
	}
	out := strings.Trim(b.String(), "-.")
	if maxOut := 63 - len(suffix); len(out) > maxOut {
		out = strings.Trim(out[:maxOut], "-.")
	}
	if out == "" {
		return "gpu-binding"
	}
	return out + suffix
}

// migrateIntelShareModeAnnotations re-keys the share mode of each card from
// its legacy device ID to its UUID. The legacy mode only offered Exclusive.
func migrateIntelShareModeAnnotations(ctx context.Context, c ctrlclient.Client, node *corev1.Node) error {
	uuids := intelCardUUIDs(node)
	if len(uuids) == 0 {
		return nil
	}
	patched := node.DeepCopy()
	changed := false
	for key := range node.Annotations {
		deviceID, ok := strings.CutPrefix(key, shareModeAnnotationPrefix)
		if !ok {
			continue
		}
		uuid, ok := legacyIntelCardUUID(deviceID, node.Name, uuids)
		if !ok {
			continue
		}
		delete(patched.Annotations, key)
		if _, exists := patched.Annotations[shareModeAnnotationPrefix+uuid]; !exists {
			patched.Annotations[shareModeAnnotationPrefix+uuid] = "0"
		}
		changed = true
	}
	if !changed {
		return nil
	}
	if err := c.Patch(ctx, patched, ctrlclient.MergeFrom(node)); err != nil {
		return errors.Wrapf(errors.WithStack(err), "failed to migrate share-mode annotations of node %s", node.Name)
	}
	return nil
}

// restartLegacyIntelWorkloads bumps the pod template of each migrated app's
// workload that still requests gpu.intel.com/xe. Integrated Intel apps on
// the xe driver request the same resource and are matched out by app.
func restartLegacyIntelWorkloads(ctx context.Context, c ctrlclient.Client, apps map[string]struct{}, now time.Time) (int, error) {
	if len(apps) == 0 {
		return 0, nil
	}
	selected := func(tpl *corev1.PodTemplateSpec) bool {
		_, ok := apps[appOwnerKey(tpl.Labels[appNameLabelKey], tpl.Labels[appOwnerLabelKey])]
		return ok && templateRequestsLegacyIntelDGPU(tpl)
	}
	restartedAt := now.UTC().Format(time.RFC3339)
	restarted := 0

	var deployments appsv1.DeploymentList
	if err := c.List(ctx, &deployments); err != nil {
		return restarted, errors.Wrap(errors.WithStack(err), "failed to list deployments")
	}
	for i := range deployments.Items {
		d := &deployments.Items[i]
		if !selected(&d.Spec.Template) {
			continue
		}
		patched := d.DeepCopy()
		setRestartedAt(&patched.Spec.Template, restartedAt)
		if err := c.Patch(ctx, patched, ctrlclient.MergeFrom(d)); err != nil {
			return restarted, errors.Wrapf(errors.WithStack(err), "failed to restart deployment %s/%s", d.Namespace, d.Name)
		}
		logger.Infof("restarted deployment %s/%s to move it from %s to HAMi", d.Namespace, d.Name, legacyIntelDGPUResource)
		restarted++
	}

	var statefulSets appsv1.StatefulSetList
	if err := c.List(ctx, &statefulSets); err != nil {
		return restarted, errors.Wrap(errors.WithStack(err), "failed to list statefulsets")
	}
	for i := range statefulSets.Items {
		s := &statefulSets.Items[i]
		if !selected(&s.Spec.Template) {
			continue
		}
		patched := s.DeepCopy()
		setRestartedAt(&patched.Spec.Template, restartedAt)
		if err := c.Patch(ctx, patched, ctrlclient.MergeFrom(s)); err != nil {
			return restarted, errors.Wrapf(errors.WithStack(err), "failed to restart statefulset %s/%s", s.Namespace, s.Name)
		}
		logger.Infof("restarted statefulset %s/%s to move it from %s to HAMi", s.Namespace, s.Name, legacyIntelDGPUResource)
		restarted++
	}
	return restarted, nil
}

func templateRequestsLegacyIntelDGPU(tpl *corev1.PodTemplateSpec) bool {
	for _, container := range tpl.Spec.Containers {
		for _, resources := range []corev1.ResourceList{container.Resources.Limits, container.Resources.Requests} {
			if _, ok := resources[legacyIntelDGPUResource]; ok {
				return true
			}
		}
	}
	return false
}

func setRestartedAt(tpl *corev1.PodTemplateSpec, restartedAt string) {
	if tpl.Annotations == nil {
		tpl.Annotations = make(map[string]string)
	}
	tpl.Annotations[restartedAtAnnotation] = restartedAt
}
