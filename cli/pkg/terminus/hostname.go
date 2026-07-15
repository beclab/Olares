package terminus

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"sort"
	"strings"
	"time"

	corev1 "k8s.io/api/core/v1"
	kerrors "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/types"
	"k8s.io/client-go/kubernetes"
	ctrl "sigs.k8s.io/controller-runtime"

	"github.com/beclab/Olares/cli/pkg/common"
	"github.com/beclab/Olares/cli/pkg/core/connector"
	"github.com/beclab/Olares/cli/pkg/core/logger"
	"github.com/beclab/Olares/cli/pkg/core/task"
	"github.com/beclab/Olares/cli/pkg/core/util"

	"github.com/pkg/errors"
)

// renamedNodeLabelAllowlistPrefixes / renamedNodeLabelAllowlistKeys control
// which labels are carried over from the old (pre-rename) node object to the
// freshly registered node when the hostname changed.
//
// We intentionally use an allowlist rather than copying everything: labels like
// kubernetes.io/hostname or the k3s-managed role labels must NOT be copied (they
// either belong to the new name or are re-applied automatically by k3s). Only
// labels that Olares applies itself and that are not self-healing need to be
// restored here.
//
// The categories that need explicit restoration on a k3s Olares node are:
//   - GPU labels written directly onto the node by olares-cli (gpu.bytetrade.io/*)
//   - the kube-ovn master role label (a no-op on Calico clusters, which use no
//     node label)
var (
	// Labels olares-cli writes directly onto the node and that are NOT
	// re-applied automatically after the node is re-registered under a new name:
	//   - gpu.bytetrade.io/{driver,cuda,cuda-supported,<mode>} (SetNodeGpuModeLabel)
	//   - node-role.kubernetes.io/worker (AddWorkerLabel; k3s only re-applies the
	//     master/control-plane roles automatically, not worker)
	// Deliberately excluded (they self-heal or belong to the new node):
	//   - kubernetes.io/*, beta.kubernetes.io/* (kubelet built-ins)
	//   - node-role.kubernetes.io/{master,control-plane} (re-applied by k3s)
	//   - gpu.intel.com/*, intel.feature.node.kubernetes.io/* (re-applied by NFD)
	renamedNodeLabelAllowlistPrefixes = []string{
		"gpu.bytetrade.io/",
	}
	renamedNodeLabelAllowlistKeys = []string{
		"node-role.kubernetes.io/worker",
		"kube-ovn/role",
	}
	// Annotations olares/app-service persist as user configuration and that are
	// NOT recreated by any controller:
	//   - sharemode.gpu.bytetrade.io/<gpu-uuid> (per-device GPU share mode set via
	//     app-service; keyed by the stable GPU UUID, so it maps cleanly to the
	//     same hardware on the renamed node)
	// Deliberately excluded (self-heal, or would be stale/wrong after an IP change):
	//   - hami.io/* (re-registered by the HAMi device plugin daemonset)
	//   - projectcalico.org/* (re-populated by calico-node; carry the OLD IP)
	//   - k3s.io/*, nfd.node.kubernetes.io/*, node.alpha.kubernetes.io/ttl,
	//     volumes.kubernetes.io/* (managed by k3s/NFD/kubelet)
	renamedNodeAnnotationAllowlistPrefixes = []string{
		gpuShareModeAnnotationPrefix,
	}
)

func newChangeIPKubeClient() (*kubernetes.Clientset, error) {
	kubeConfig, err := ctrl.GetConfig()
	if err != nil {
		return nil, errors.Wrap(err, "failed to load kubeconfig")
	}
	kubeClient, err := kubernetes.NewForConfig(kubeConfig)
	if err != nil {
		return nil, errors.Wrap(err, "failed to create kube client")
	}
	return kubeClient, nil
}

func nodeIsReady(node *corev1.Node) bool {
	if node == nil {
		return false
	}
	for _, cond := range node.Status.Conditions {
		if cond.Type == corev1.NodeReady {
			return cond.Status == corev1.ConditionTrue
		}
	}
	return false
}

func configuredNodeName() string {
	for _, filePath := range []string{
		"/etc/systemd/system/k3s.service",
		"/etc/systemd/system/kubelet.service.d/10-kubeadm.conf",
	} {
		content, err := os.ReadFile(filePath)
		if err != nil {
			continue
		}
		for _, flag := range []string{"--node-name=", "--hostname-override="} {
			if value := commandFlagValue(string(content), flag); value != "" {
				return strings.ToLower(value)
			}
		}
	}
	return ""
}

func commandFlagValue(content, flag string) string {
	index := strings.Index(content, flag)
	if index < 0 {
		return ""
	}
	value := content[index+len(flag):]
	if end := strings.IndexAny(value, " \t\r\n\"'"); end >= 0 {
		value = value[:end]
	}
	return strings.TrimSpace(value)
}

// pvPinnedToNode reports whether a PersistentVolume's node affinity restricts it
// to the given node name via the standard kubernetes.io/hostname key, which is
// how node-local provisioners (local-path, openebs-hostpath, ...) pin volumes.
func pvPinnedToNode(pv *corev1.PersistentVolume, nodeName string) bool {
	if pv.Spec.NodeAffinity == nil || pv.Spec.NodeAffinity.Required == nil {
		return false
	}
	for _, term := range pv.Spec.NodeAffinity.Required.NodeSelectorTerms {
		for _, expr := range term.MatchExpressions {
			if expr.Operator != corev1.NodeSelectorOpIn ||
				(expr.Key != corev1.LabelHostname && expr.Key != "beta.kubernetes.io/hostname") {
				continue
			}
			for _, v := range expr.Values {
				if strings.EqualFold(v, nodeName) {
					return true
				}
			}
		}
		for _, field := range term.MatchFields {
			if field.Key != "metadata.name" || field.Operator != corev1.NodeSelectorOpIn {
				continue
			}
			for _, value := range field.Values {
				if strings.EqualFold(value, nodeName) {
					return true
				}
			}
		}
	}
	return false
}

func disposableRenamePV(pv *corev1.PersistentVolume) bool {
	if pv == nil || pv.Spec.ClaimRef == nil {
		return false
	}
	return pv.Spec.ClaimRef.Namespace == "kubesphere-monitoring-system" &&
		strings.HasPrefix(pv.Spec.ClaimRef.Name, "prometheus-k8s-db-prometheus-k8s-")
}

func labelIsAllowlistedForRestore(key string) bool {
	for _, k := range renamedNodeLabelAllowlistKeys {
		if key == k {
			return true
		}
	}
	for _, p := range renamedNodeLabelAllowlistPrefixes {
		if strings.HasPrefix(key, p) {
			return true
		}
	}
	return false
}

func annotationIsAllowlistedForRestore(key string) bool {
	for _, p := range renamedNodeAnnotationAllowlistPrefixes {
		if strings.HasPrefix(key, p) {
			return true
		}
	}
	return false
}

const (
	// gpuShareModeAnnotationPrefix is the node-annotation prefix app-service uses
	// to persist a per-device GPU share mode: sharemode.gpu.bytetrade.io/<deviceID>.
	gpuShareModeAnnotationPrefix = "sharemode.gpu.bytetrade.io/"

	// GPU allocation store written by app-service: a ConfigMap holding a JSON
	// array of allocation rows ({nodeName, deviceId, ...}).
	gpuAllocationsConfigMapNamespace = "os-framework"
	gpuAllocationsConfigMapName      = "app-gpu-allocations"
	gpuAllocationsConfigMapKey       = "allocations.json"
)

// rewriteNodeScopedID rewrites an identifier that embeds the node name as a
// leading "<node>-" segment from the old node name to the new one. app-service
// builds non-HAMI GPU device ids as "<node>-<mode>-0" (compute.nonHAMIDevice),
// so those must follow a hostname change. Identifiers that do not start with
// "<oldNode>-" (e.g. a HAMI GPU UUID like "GPU-....") are returned unchanged.
func rewriteNodeScopedID(id, oldNode, newNode string) string {
	if oldNode == "" || newNode == "" {
		return id
	}
	if !strings.HasPrefix(id, oldNode+"-") {
		return id
	}
	return newNode + strings.TrimPrefix(id, oldNode)
}

// DetectHostnameChangeModule runs before ChangeIPModule and records, in the
// pipeline cache, whether this machine is already registered in Kubernetes under
// a different node name than the current (lowercased) hostname. If so, the
// hostname was changed prior to running change-ip and ChangeIPModule will take
// the extra steps needed to migrate to the new node identity.
type DetectHostnameChangeModule struct {
	common.KubeModule
}

func (m *DetectHostnameChangeModule) Init() {
	m.Name = "DetectHostnameChange"
	m.Tasks = []task.Interface{
		&task.LocalTask{
			Name:   "DetectHostnameChange",
			Action: new(DetectHostnameChange),
		},
	}
}

type DetectHostnameChange struct {
	common.KubeAction
}

func (t *DetectHostnameChange) Execute(runtime connector.Runtime) error {
	// default to "not changed" so downstream logic always has a value to read
	t.PipelineCache.Set(common.CacheHostnameChanged, false)

	installed, _ := t.PipelineCache.GetMustBool(common.CacheInstalledState)
	if !installed {
		logger.Info("Olares is not installed, skipping hostname change detection")
		return nil
	}

	currentName := strings.ToLower(runtime.GetSystemInfo().GetHostname())
	previouslyConfiguredName := configuredNodeName()
	machineID := util.GetMachineID()
	systemUUID := util.GetSystemUUID()
	logger.Infof("detecting hostname change for current node name=%q", currentName)
	if machineID == "" && systemUUID == "" {
		return fmt.Errorf("cannot detect a hostname change: both kubelet machineID and systemUUID are empty")
	}
	if machineID == "" || systemUUID == "" {
		logger.Warnf("only one machine identifier is available (machineID=%q, systemUUID=%q); "+
			"detection will match on the available one only, which is less strict against cloned machines",
			machineID, systemUUID)
	}

	kubeClient, err := newChangeIPKubeClient()
	if err != nil {
		if previouslyConfiguredName != "" && strings.EqualFold(previouslyConfiguredName, currentName) {
			logger.Warnf("Kubernetes API is unavailable, but the configured node name %q is unchanged; continuing the original change-ip flow", currentName)
			return nil
		}
		return errors.Wrap(err, "failed to create kube client for hostname change detection")
	}
	nodes, err := kubeClient.CoreV1().Nodes().List(context.Background(), metav1.ListOptions{})
	if err != nil {
		if previouslyConfiguredName != "" && strings.EqualFold(previouslyConfiguredName, currentName) {
			logger.Warnf("Kubernetes API is unavailable, but the configured node name %q is unchanged; continuing the original change-ip flow", currentName)
			return nil
		}
		return errors.Wrap(err, "failed to list nodes for hostname change detection")
	}

	var currentNode *corev1.Node
	var matchingOldNodes []*corev1.Node
	for i := range nodes.Items {
		node := &nodes.Items[i]
		if strings.EqualFold(node.Name, currentName) {
			currentNode = node
			continue
		}
		if !nodeIdentityMatches(node, machineID, systemUUID) {
			continue
		}
		matchingOldNodes = append(matchingOldNodes, node)
	}

	if currentNode != nil && !nodeIdentityMatches(currentNode, machineID, systemUUID) {
		return fmt.Errorf("node %q already exists but belongs to a different machine identity", currentName)
	}
	if len(matchingOldNodes) > 1 {
		names := make([]string, 0, len(matchingOldNodes))
		for _, node := range matchingOldNodes {
			names = append(names, node.Name)
		}
		sort.Strings(names)
		return fmt.Errorf("cannot determine the old hostname: multiple nodes share this machine identity: %s", strings.Join(names, ", "))
	}
	if len(matchingOldNodes) == 0 {
		if currentNode == nil {
			return fmt.Errorf("cannot find a node matching current hostname %q or this machine identity", currentName)
		}
		logger.Infof("no former node found for this machine; hostname is unchanged (current=%q)", currentName)
		return nil
	}

	oldNodeName := matchingOldNodes[0].Name
	if previouslyConfiguredName != "" &&
		!strings.EqualFold(previouslyConfiguredName, currentName) &&
		!strings.EqualFold(previouslyConfiguredName, oldNodeName) {
		return fmt.Errorf("machine identity matched node %q, but the service was configured for node %q", oldNodeName, previouslyConfiguredName)
	}

	logger.Warnf("node %q shares this machine's identity but differs from current hostname %q "+
		"(old node ready=%v); treating this as a hostname change",
		oldNodeName, currentName, nodeIsReady(matchingOldNodes[0]))
	logger.Warnf("hostname change detected: old node=%q, new node=%q; change-ip will restore node "+
		"labels, clean up node-local PVs bound to the old node, and delete the old node after restart",
		oldNodeName, currentName)
	t.PipelineCache.Set(common.CacheHostnameChanged, true)
	t.PipelineCache.Set(common.CacheOldNodeName, oldNodeName)
	t.PipelineCache.Set(common.CacheOldNodeUID, string(matchingOldNodes[0].UID))
	return nil
}

// nodeIdentityMatches requires every machine identifier we could read locally to
// match the corresponding field on the node. When both machineID and systemUUID
// are available, both must match (strict, to avoid false positives on cloned
// machines that share only one identifier).
func nodeIdentityMatches(node *corev1.Node, machineID, systemUUID string) bool {
	ni := node.Status.NodeInfo
	matched := 0
	if machineID != "" {
		if !strings.EqualFold(ni.MachineID, machineID) {
			return false
		}
		matched++
	}
	if systemUUID != "" {
		if !strings.EqualFold(ni.SystemUUID, systemUUID) {
			return false
		}
		matched++
	}
	return matched > 0
}

// RestoreLabelsFromRenamedNode copies the allowlisted labels/annotations from the
// old (pre-rename) node onto the newly registered node, so metadata that Olares
// applied during install (and that is not re-applied automatically) survives a
// hostname change. It is idempotent.
type RestoreLabelsFromRenamedNode struct {
	common.KubeAction
	OldNode string
	NewNode string
}

func (a *RestoreLabelsFromRenamedNode) Execute(runtime connector.Runtime) error {
	if a.OldNode == "" || a.NewNode == "" {
		logger.Warnf("skipping label restore: old node=%q, new node=%q", a.OldNode, a.NewNode)
		return nil
	}
	kubeClient, err := newChangeIPKubeClient()
	if err != nil {
		return err
	}
	ctx := context.Background()

	oldNode, err := kubeClient.CoreV1().Nodes().Get(ctx, a.OldNode, metav1.GetOptions{})
	if err != nil {
		if kerrors.IsNotFound(err) {
			// old node already gone (e.g. a re-run after labels were restored)
			logger.Infof("old node %q not found, assuming labels were already restored", a.OldNode)
			return nil
		}
		return errors.Wrapf(err, "failed to get old node %q", a.OldNode)
	}

	// the new node must exist before we can copy labels onto it; return an error
	// so the task is retried until k3s has registered it.
	newNode, err := kubeClient.CoreV1().Nodes().Get(ctx, a.NewNode, metav1.GetOptions{})
	if err != nil {
		return errors.Wrapf(err, "failed to get new node %q (may not be registered yet)", a.NewNode)
	}
	if !nodeIsReady(newNode) {
		return fmt.Errorf("new node %q is not ready yet", a.NewNode)
	}

	if newNode.Labels == nil {
		newNode.Labels = map[string]string{}
	}
	if newNode.Annotations == nil {
		newNode.Annotations = map[string]string{}
	}

	changed := false
	for k, v := range oldNode.Labels {
		if !labelIsAllowlistedForRestore(k) {
			continue
		}
		if existing, ok := newNode.Labels[k]; !ok || existing != v {
			logger.Infof("restoring label %s=%s onto node %q (from old node %q)", k, v, a.NewNode, a.OldNode)
			newNode.Labels[k] = v
			changed = true
		}
	}
	for k, v := range oldNode.Annotations {
		if !annotationIsAllowlistedForRestore(k) {
			continue
		}
		// The share-mode annotation key is sharemode.gpu.bytetrade.io/<deviceID>;
		// for non-HAMI devices the deviceID embeds the node name, so the key must
		// be rewritten to the new node (HAMI UUID keys are left unchanged).
		targetKey := k
		if suffix := strings.TrimPrefix(k, gpuShareModeAnnotationPrefix); suffix != k {
			targetKey = gpuShareModeAnnotationPrefix + rewriteNodeScopedID(suffix, a.OldNode, a.NewNode)
		}
		if existing, ok := newNode.Annotations[targetKey]; !ok || existing != v {
			logger.Infof("restoring annotation %s=%s onto node %q (from old node %q key %q)", targetKey, v, a.NewNode, a.OldNode, k)
			newNode.Annotations[targetKey] = v
			changed = true
		}
	}

	if !changed {
		logger.Infof("no labels/annotations need restoring onto node %q", a.NewNode)
		return nil
	}
	if _, err := kubeClient.CoreV1().Nodes().Update(ctx, newNode, metav1.UpdateOptions{}); err != nil {
		return errors.Wrapf(err, "failed to update node %q with restored labels", a.NewNode)
	}
	return nil
}

// MigrateRenamedNodeGPUAllocations rewrites the app-service GPU allocation store
// (ConfigMap os-framework/app-gpu-allocations) so bound-GPU apps keep their
// bindings after a hostname change. app-service matches an allocation row to a
// device by exact (nodeName, deviceId) (compute.attachBindings), and for
// non-HAMI devices the deviceId embeds the node name ("<node>-<mode>-0"), so
// without this rewrite every binding would be dropped and share modes would
// reset. Each row's nodeName is moved old->new and its node-scoped deviceId is
// rewritten; HAMI GPU UUIDs are left unchanged. The row is parsed generically to
// preserve any fields we do not model. Idempotent and tolerant of a
// missing/empty ConfigMap.
type MigrateRenamedNodeGPUAllocations struct {
	common.KubeAction
	OldNode string
	NewNode string
}

func (a *MigrateRenamedNodeGPUAllocations) Execute(_ connector.Runtime) error {
	if a.OldNode == "" || a.NewNode == "" {
		return nil
	}
	kubeClient, err := newChangeIPKubeClient()
	if err != nil {
		return err
	}
	ctx := context.Background()
	cm, err := kubeClient.CoreV1().ConfigMaps(gpuAllocationsConfigMapNamespace).Get(ctx, gpuAllocationsConfigMapName, metav1.GetOptions{})
	if err != nil {
		if kerrors.IsNotFound(err) {
			logger.Infof("no %s/%s ConfigMap; no GPU allocations to migrate", gpuAllocationsConfigMapNamespace, gpuAllocationsConfigMapName)
			return nil
		}
		return errors.Wrapf(err, "failed to get %s ConfigMap", gpuAllocationsConfigMapName)
	}

	raw := ""
	if cm.Data != nil {
		raw = cm.Data[gpuAllocationsConfigMapKey]
	}
	if strings.TrimSpace(raw) == "" {
		logger.Infof("GPU allocation store is empty; nothing to migrate")
		return nil
	}

	// Parse generically to preserve fields we do not model.
	var allocations []map[string]json.RawMessage
	if err := json.Unmarshal([]byte(raw), &allocations); err != nil {
		return errors.Wrapf(err, "failed to parse %s", gpuAllocationsConfigMapKey)
	}

	changed := false
	for i := range allocations {
		row := allocations[i]
		if newName, ok := rewriteStringField(row, "nodeName", func(v string) string {
			if v == a.OldNode {
				return a.NewNode
			}
			return v
		}); ok {
			logger.Infof("migrating GPU allocation nodeName -> %q", newName)
			changed = true
		}
		if newID, ok := rewriteStringField(row, "deviceId", func(v string) string {
			return rewriteNodeScopedID(v, a.OldNode, a.NewNode)
		}); ok {
			logger.Infof("migrating GPU allocation deviceId -> %q", newID)
			changed = true
		}
	}
	if !changed {
		logger.Infof("no GPU allocations reference old node %q; nothing to migrate", a.OldNode)
		return nil
	}

	updated, err := json.Marshal(allocations)
	if err != nil {
		return errors.Wrap(err, "failed to serialize migrated GPU allocations")
	}
	if cm.Data == nil {
		cm.Data = map[string]string{}
	}
	cm.Data[gpuAllocationsConfigMapKey] = string(updated)
	if _, err := kubeClient.CoreV1().ConfigMaps(gpuAllocationsConfigMapNamespace).Update(ctx, cm, metav1.UpdateOptions{}); err != nil {
		return errors.Wrapf(err, "failed to update %s ConfigMap", gpuAllocationsConfigMapName)
	}
	logger.Warnf("migrated GPU allocations from old node %q to new node %q", a.OldNode, a.NewNode)
	return nil
}

// rewriteStringField applies fn to a string field of a generic JSON object and
// writes it back when it changed. It returns the new value and whether it
// changed. Non-string / missing fields are left untouched.
func rewriteStringField(row map[string]json.RawMessage, field string, fn func(string) string) (string, bool) {
	raw, ok := row[field]
	if !ok {
		return "", false
	}
	var value string
	if err := json.Unmarshal(raw, &value); err != nil {
		return "", false
	}
	updated := fn(value)
	if updated == value {
		return value, false
	}
	encoded, err := json.Marshal(updated)
	if err != nil {
		return "", false
	}
	row[field] = encoded
	return updated, true
}

// CleanupRenamedNodeLocalPVs removes node-local PersistentVolumes that are pinned
// (via node affinity) to the old node name, together with their bound PVCs and
// the pods consuming them. On a single-node Olares the only node-local PV belongs
// to Prometheus (an operator-managed StatefulSet with a fixed PVC name); its data
// is disposable monitoring data. The stale PV must be removed so that the
// StatefulSet's recreated PVC (same fixed name) can bind to a fresh PV on the new
// node instead of the old, now-unschedulable one.
//
// The action is idempotent (tolerates already-deleted objects). It first lets the
// PVC/PV clear on their own: once the (scheduled) consuming pod terminates, the
// pvc-protection finalizer is released naturally (the StatefulSet-recreated pod
// is Pending/unscheduled and does not count as a consumer). Only if a PVC/PV is
// still stuck terminating after a grace period does it force-remove the protection
// finalizers as a fallback.
type CleanupRenamedNodeLocalPVs struct {
	common.KubeAction
	OldNode string
}

func (a *CleanupRenamedNodeLocalPVs) Execute(runtime connector.Runtime) error {
	if a.OldNode == "" {
		logger.Warn("skipping node-local PV cleanup: old node name is empty")
		return nil
	}
	kubeClient, err := newChangeIPKubeClient()
	if err != nil {
		return err
	}
	ctx := context.Background()

	pvs, err := kubeClient.CoreV1().PersistentVolumes().List(ctx, metav1.ListOptions{})
	if err != nil {
		return errors.Wrap(err, "failed to list persistent volumes")
	}

	var targets []corev1.PersistentVolume
	var unsupported []string
	for i := range pvs.Items {
		pv := &pvs.Items[i]
		if !pvPinnedToNode(pv, a.OldNode) {
			continue
		}
		if !disposableRenamePV(pv) {
			unsupported = append(unsupported, pv.Name)
			continue
		}
		targets = append(targets, *pv)
	}
	if len(unsupported) > 0 {
		sort.Strings(unsupported)
		return fmt.Errorf("refusing to delete non-disposable local PVs pinned to old node %q: %s", a.OldNode, strings.Join(unsupported, ", "))
	}
	if len(targets) == 0 {
		logger.Infof("no node-local PVs pinned to old node %q, nothing to clean up", a.OldNode)
		return nil
	}

	for i := range targets {
		pv := &targets[i]
		logger.Warnf("cleaning up node-local PV %q pinned to old node %q (reclaimPolicy=%s)",
			pv.Name, a.OldNode, pv.Spec.PersistentVolumeReclaimPolicy)

		if ref := pv.Spec.ClaimRef; ref != nil && ref.Name != "" {
			if err := validatePVCBinding(ctx, kubeClient, ref.Namespace, ref.Name, pv.Name, ref.UID); err != nil {
				return errors.Wrapf(err, "refusing to clean up stale binding for PV %q", pv.Name)
			}
			if err := deleteConsumingPods(ctx, kubeClient, ref.Namespace, ref.Name); err != nil {
				return errors.Wrapf(err, "failed to delete pods consuming PVC %s/%s", ref.Namespace, ref.Name)
			}
			if err := forceDeletePVC(ctx, kubeClient, ref.Namespace, ref.Name); err != nil {
				return errors.Wrapf(err, "failed to delete PVC %s/%s", ref.Namespace, ref.Name)
			}
		}
		if err := forceDeletePV(ctx, kubeClient, pv.Name); err != nil {
			return errors.Wrapf(err, "failed to delete PV %s", pv.Name)
		}
	}
	return nil
}

func validatePVCBinding(ctx context.Context, kubeClient kubernetes.Interface, namespace, name, pvName string, claimUID types.UID) error {
	pvc, err := kubeClient.CoreV1().PersistentVolumeClaims(namespace).Get(ctx, name, metav1.GetOptions{})
	if err != nil {
		if kerrors.IsNotFound(err) {
			return nil
		}
		return err
	}
	if pvc.Spec.VolumeName != pvName {
		return fmt.Errorf("PVC %s/%s is bound to PV %q, expected %q", namespace, name, pvc.Spec.VolumeName, pvName)
	}
	if claimUID != "" && pvc.UID != claimUID {
		return fmt.Errorf("PVC %s/%s UID is %q, expected claim UID %q", namespace, name, pvc.UID, claimUID)
	}
	return nil
}

func deleteConsumingPods(ctx context.Context, kubeClient kubernetes.Interface, namespace, pvcName string) error {
	if namespace == "" {
		return nil
	}
	pods, err := kubeClient.CoreV1().Pods(namespace).List(ctx, metav1.ListOptions{})
	if err != nil {
		return err
	}
	for i := range pods.Items {
		pod := &pods.Items[i]
		if !podUsesPVC(pod, pvcName) {
			continue
		}
		logger.Infof("deleting pod %s/%s that consumes PVC %s", pod.Namespace, pod.Name, pvcName)
		if err := kubeClient.CoreV1().Pods(pod.Namespace).Delete(ctx, pod.Name, metav1.DeleteOptions{}); err != nil && !kerrors.IsNotFound(err) {
			return err
		}
	}
	return nil
}

func podUsesPVC(pod *corev1.Pod, pvcName string) bool {
	for _, vol := range pod.Spec.Volumes {
		if vol.PersistentVolumeClaim != nil && vol.PersistentVolumeClaim.ClaimName == pvcName {
			return true
		}
	}
	return false
}

// renamePVReclaimGrace is how long we let a PVC/PV clear on its own (via the
// normal protection-finalizer machinery) after requesting deletion, before
// force-removing the finalizers. It is intentionally short: the volume holds
// disposable Prometheus monitoring data, and the consuming pod can have a large
// terminationGracePeriodSeconds (e.g. 600s) that would otherwise keep the
// pvc-protection finalizer held for a long time, so we prefer to force after a
// brief wait rather than block the whole change-ip.
const (
	renamePVReclaimGrace  = 20 * time.Second
	renamePVReclaimPoll   = 3 * time.Second
	renamePVPostForceWait = 15 * time.Second
)

func forceDeletePVC(ctx context.Context, kubeClient kubernetes.Interface, namespace, name string) error {
	return deleteAndReclaim(ctx, "PVC", namespace+"/"+name,
		func() (metav1.Object, error) {
			return kubeClient.CoreV1().PersistentVolumeClaims(namespace).Get(ctx, name, metav1.GetOptions{})
		},
		func() error {
			return kubeClient.CoreV1().PersistentVolumeClaims(namespace).Delete(ctx, name, metav1.DeleteOptions{})
		},
		func(patch []byte) error {
			_, err := kubeClient.CoreV1().PersistentVolumeClaims(namespace).Patch(ctx, name, types.MergePatchType, patch, metav1.PatchOptions{})
			return err
		})
}

func forceDeletePV(ctx context.Context, kubeClient kubernetes.Interface, name string) error {
	return deleteAndReclaim(ctx, "PV", name,
		func() (metav1.Object, error) {
			return kubeClient.CoreV1().PersistentVolumes().Get(ctx, name, metav1.GetOptions{})
		},
		func() error {
			return kubeClient.CoreV1().PersistentVolumes().Delete(ctx, name, metav1.DeleteOptions{})
		},
		func(patch []byte) error {
			_, err := kubeClient.CoreV1().PersistentVolumes().Patch(ctx, name, types.MergePatchType, patch, metav1.PatchOptions{})
			return err
		})
}

// deleteAndReclaim requests deletion of an object, waits up to
// renamePVReclaimGrace for it to disappear on its own (letting the protection
// finalizers clear naturally once the terminating consumer is gone), and only
// force-removes the finalizers if it is still stuck terminating after that.
func deleteAndReclaim(ctx context.Context, kind, name string,
	get func() (metav1.Object, error),
	del func() error,
	patch func([]byte) error) error {

	obj, err := get()
	if err != nil {
		if kerrors.IsNotFound(err) {
			return nil
		}
		return err
	}
	if obj.GetDeletionTimestamp() == nil {
		logger.Infof("deleting %s %s", kind, name)
		if err := del(); err != nil && !kerrors.IsNotFound(err) {
			return err
		}
	}

	// phase 1: give the normal finalizer machinery a chance to reclaim it.
	gone, err := waitObjectGone(ctx, get, renamePVReclaimGrace, renamePVReclaimPoll)
	if err != nil {
		return err
	}
	if gone {
		logger.Infof("%s %s reclaimed naturally", kind, name)
		return nil
	}

	// phase 2: still terminating after the grace period -> force-remove finalizers.
	obj, err = get()
	if err != nil {
		if kerrors.IsNotFound(err) {
			return nil
		}
		return err
	}
	if len(obj.GetFinalizers()) > 0 {
		logger.Warnf("%s %s still terminating after %s; force-removing finalizers %v",
			kind, name, renamePVReclaimGrace, obj.GetFinalizers())
		if err := patch([]byte(`{"metadata":{"finalizers":null}}`)); err != nil && !kerrors.IsNotFound(err) {
			return err
		}
	}
	gone, err = waitObjectGone(ctx, get, renamePVPostForceWait, renamePVReclaimPoll)
	if err != nil {
		return err
	}
	if !gone {
		return fmt.Errorf("%s %s still exists after finalizers were removed", kind, name)
	}
	return nil
}

// waitObjectGone polls until get() reports NotFound (returns true) or the timeout
// elapses (returns false).
func waitObjectGone(ctx context.Context, get func() (metav1.Object, error), timeout, poll time.Duration) (bool, error) {
	deadline := time.Now().Add(timeout)
	for {
		_, err := get()
		if err != nil {
			if kerrors.IsNotFound(err) {
				return true, nil
			}
			return false, err
		}
		if time.Now().After(deadline) {
			return false, nil
		}
		select {
		case <-ctx.Done():
			return false, ctx.Err()
		case <-time.After(poll):
		}
	}
}

// WaitForRenamedOldNodeDrained blocks until no live (non-terminating,
// non-terminal) pods remain assigned to the old node, i.e. the old node's
// workloads have started being garbage-collected and recreated elsewhere.
//
// This gate exists to prevent the subsequent cluster-wide readiness check from
// passing prematurely: immediately after the old Node object is deleted, its
// pods still report Running (stale status) for a short window, so a readiness
// check run right away would trivially succeed before anything actually
// migrated. Unlike a pre-captured pod count, this waits on real observed state,
// so transient/Job/scaled-down pods that never come back cannot block it.
type WaitForRenamedOldNodeDrained struct {
	common.KubeAction
	OldNode string
}

func (a *WaitForRenamedOldNodeDrained) Execute(_ connector.Runtime) error {
	if a.OldNode == "" {
		return nil
	}
	kubeClient, err := newChangeIPKubeClient()
	if err != nil {
		return err
	}
	pods, err := kubeClient.CoreV1().Pods(corev1.NamespaceAll).List(context.Background(), metav1.ListOptions{})
	if err != nil {
		return errors.Wrap(err, "failed to list pods while draining old node")
	}
	remaining := 0
	for i := range pods.Items {
		pod := &pods.Items[i]
		// A pod that is already terminating (DeletionTimestamp set) or terminal
		// is on its way out and its replacement is being created, so it does not
		// count as "still occupying" the old node.
		if pod.Spec.NodeName != a.OldNode ||
			pod.DeletionTimestamp != nil ||
			pod.Status.Phase == corev1.PodSucceeded ||
			pod.Status.Phase == corev1.PodFailed {
			continue
		}
		remaining++
	}
	if remaining > 0 {
		return fmt.Errorf("waiting for %d pod(s) to drain from old node %q", remaining, a.OldNode)
	}
	logger.Infof("old node %q is drained; workloads are rescheduling onto the new node", a.OldNode)
	return nil
}

// DeleteRenamedOldNode removes the stale Kubernetes Node object for the old
// hostname. Deleting it triggers garbage collection of the pods still assigned to
// it so their controllers reschedule them onto the new node. Idempotent.
type DeleteRenamedOldNode struct {
	common.KubeAction
	OldNode string
}

func (a *DeleteRenamedOldNode) Execute(runtime connector.Runtime) error {
	if a.OldNode == "" {
		logger.Warn("skipping old node deletion: old node name is empty")
		return nil
	}
	kubeClient, err := newChangeIPKubeClient()
	if err != nil {
		return err
	}
	ctx := context.Background()
	logger.Warnf("deleting old node %q", a.OldNode)
	deleteOptions := metav1.DeleteOptions{}
	if oldNodeUID, ok := a.PipelineCache.GetMustString(common.CacheOldNodeUID); ok && oldNodeUID != "" {
		uid := types.UID(oldNodeUID)
		deleteOptions.Preconditions = &metav1.Preconditions{UID: &uid}
	}
	if err := kubeClient.CoreV1().Nodes().Delete(ctx, a.OldNode, deleteOptions); err != nil && !kerrors.IsNotFound(err) {
		return errors.Wrapf(err, "failed to delete old node %q", a.OldNode)
	}
	return nil
}
