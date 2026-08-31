package webhook

import (
	"context"
	"crypto/rand"
	"encoding/json"
	"errors"
	"fmt"
	"net"
	"strconv"
	"strings"

	"github.com/beclab/Olares/framework/app-service/pkg/constants"
	apputils "github.com/beclab/Olares/framework/app-service/pkg/utils/app"
	appv1alpha1 "github.com/beclab/api/api/app.bytetrade.io/v1alpha1"
	corev1 "k8s.io/api/core/v1"
	apierrors "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"k8s.io/apimachinery/pkg/types"
	"k8s.io/client-go/util/retry"
	"k8s.io/klog/v2"
)

const (
	overlayMACSetting            = "overlayMacvlanMac"
	overlayMACByOrdinalSetting   = "overlayMacvlanMacByOrdinal"
	overlayMACFinalizer          = "app.bytetrade.io/overlay-mac-claim"
	overlayMACAllocationPlural   = "overlaymacallocations"
	overlayMACMasterNodeLabel    = "node-role.kubernetes.io/control-plane"
	overlayMACAllocationPhase    = "Bound"
	overlayMACAllocationAttempts = 10
)

var overlayMACAllocationGVR = schema.GroupVersionResource{
	Group:    "app.bytetrade.io",
	Version:  "v1alpha1",
	Resource: overlayMACAllocationPlural,
}

func generateOverlayMAC() (string, error) {
	var suffix [5]byte
	if _, err := rand.Read(suffix[:]); err != nil {
		return "", fmt.Errorf("generate overlay MAC: %w", err)
	}
	mac := append([]byte{0x02}, suffix[:]...)
	return net.HardwareAddr(mac).String(), nil
}

func validateOverlayMAC(value string) error {
	mac, err := net.ParseMAC(value)
	if err != nil || len(mac) != 6 {
		return fmt.Errorf("invalid overlay MAC %q", value)
	}
	if mac[0] != 0x02 || mac[0]&0x01 != 0 {
		return fmt.Errorf("overlay MAC %q must be a locally administered unicast 02: address", value)
	}
	return nil
}

func overlayMACKey(mac string) string {
	return strings.ToLower(strings.ReplaceAll(mac, ":", ""))
}

func (wh *Webhook) ensureOverlayMAC(ctx context.Context, pod *corev1.Pod, dryRun bool) (string, error) {
	if wh.dynamicClient == nil || wh.allocationClient == nil {
		return "", errors.New("overlay MAC allocator clients are not configured")
	}
	appName := pod.Labels[constants.ApplicationNameLabel]
	owner := pod.Labels[constants.ApplicationOwnerLabel]
	if appName == "" {
		return "", errors.New("overlay MAC allocation requires an application name label")
	}
	applicationName, err := apputils.FmtAppMgrName(appName, owner, pod.Namespace)
	if err != nil {
		return "", fmt.Errorf("resolve application name: %w", err)
	}
	app, err := wh.dynamicClient.AppV1alpha1().Applications().Get(ctx, applicationName, metav1.GetOptions{})
	if err != nil {
		return "", fmt.Errorf("get application %q: %w", applicationName, err)
	}
	instanceKey, err := wh.overlayMACInstanceKey(ctx, pod, appName)
	if err != nil {
		return "", err
	}
	ordinal, hasOrdinal := "", false
	for _, ownerRef := range pod.OwnerReferences {
		if ownerRef.Controller != nil && *ownerRef.Controller && ownerRef.Kind == "StatefulSet" {
			ordinal, hasOrdinal = statefulSetOrdinal(pod)
			break
		}
	}
	persisted, err := persistedOverlayMAC(app, ordinal, hasOrdinal)
	if err != nil {
		return "", err
	}
	if persisted != "" {
		if dryRun {
			return persisted, nil
		}
		if err := wh.ensureOverlayMACAllocation(ctx, app, instanceKey, persisted); err != nil {
			return "", err
		}
		if err := wh.ensureOverlayMACFinalizer(ctx, app.Name, app.UID); err != nil {
			return "", err
		}
		return persisted, nil
	}
	if dryRun {
		return generateOverlayMAC()
	}
	if app.UID == "" {
		return "", fmt.Errorf("application %q has no UID; refusing to allocate overlay MAC", app.Name)
	}

	for attempt := 0; attempt < overlayMACAllocationAttempts; attempt++ {
		candidate, err := generateOverlayMAC()
		if err != nil {
			return "", err
		}
		created, err := wh.createOverlayMACAllocation(ctx, app, instanceKey, candidate)
		if err != nil {
			if apierrors.IsAlreadyExists(err) {
				continue
			}
			return "", fmt.Errorf("reserve overlay MAC %s: %w", candidate, err)
		}
		if !created {
			continue
		}
		if err := wh.persistOverlayMAC(ctx, app.Name, app.UID, ordinal, hasOrdinal, candidate); err != nil {
			return "", err
		}
		return candidate, nil
	}
	return "", fmt.Errorf("reserve overlay MAC: exhausted %d atomic allocation attempts", overlayMACAllocationAttempts)
}

func persistedOverlayMAC(app *appv1alpha1.Application, ordinal string, hasOrdinal bool) (string, error) {
	if app.Spec.Settings == nil {
		return "", nil
	}
	if hasOrdinal {
		raw := app.Spec.Settings[overlayMACByOrdinalSetting]
		if raw == "" {
			return "", nil
		}
		values := map[string]string{}
		if err := json.Unmarshal([]byte(raw), &values); err != nil {
			return "", fmt.Errorf("invalid persisted overlay MAC map: %w", err)
		}
		value, ok := values[ordinal]
		if !ok {
			return "", nil
		}
		if err := validateOverlayMAC(value); err != nil {
			return "", err
		}
		return value, nil
	}
	value := app.Spec.Settings[overlayMACSetting]
	if value == "" {
		return "", nil
	}
	if err := validateOverlayMAC(value); err != nil {
		return "", err
	}
	return value, nil
}

func (wh *Webhook) createOverlayMACAllocation(ctx context.Context, app *appv1alpha1.Application, instanceKey, mac string) (bool, error) {
	key := overlayMACKey(mac)
	allocation := &unstructured.Unstructured{Object: map[string]interface{}{
		"apiVersion": "app.bytetrade.io/v1alpha1",
		"kind":       "OverlayMACAllocation",
		"metadata": map[string]interface{}{
			"name": key,
			"ownerReferences": []interface{}{map[string]interface{}{
				"apiVersion":         "app.bytetrade.io/v1alpha1",
				"kind":               "Application",
				"name":               app.Name,
				"uid":                string(app.UID),
				"blockOwnerDeletion": true,
			}},
		},
		"spec": map[string]interface{}{
			"mac":            mac,
			"instanceKey":    instanceKey,
			"applicationUID": string(app.UID),
			"applicationRef": app.Name,
			"phase":          overlayMACAllocationPhase,
		},
	}}
	_, err := wh.allocationClient.Resource(overlayMACAllocationGVR).Create(ctx, allocation, metav1.CreateOptions{})
	if err != nil {
		return false, err
	}
	klog.Infof("overlay-mac: action=allocate app=%s instance=%s mac=%s", app.Name, instanceKey, mac)
	return true, nil
}

func (wh *Webhook) ensureOverlayMACAllocation(ctx context.Context, app *appv1alpha1.Application, instanceKey, mac string) error {
	allocation, err := wh.allocationClient.Resource(overlayMACAllocationGVR).Get(ctx, overlayMACKey(mac), metav1.GetOptions{})
	if apierrors.IsNotFound(err) {
		_, err = wh.createOverlayMACAllocation(ctx, app, instanceKey, mac)
		if apierrors.IsAlreadyExists(err) {
			allocation, err = wh.allocationClient.Resource(overlayMACAllocationGVR).Get(ctx, overlayMACKey(mac), metav1.GetOptions{})
			if err != nil {
				return err
			}
			return validateOverlayMACAllocation(allocation, app, instanceKey, mac)
		}
		return err
	}
	if err != nil {
		return fmt.Errorf("get overlay MAC allocation %s: %w", overlayMACKey(mac), err)
	}
	if err := validateOverlayMACAllocation(allocation, app, instanceKey, mac); err != nil {
		return err
	}
	klog.Infof("overlay-mac: action=reuse app=%s instance=%s mac=%s", app.Name, instanceKey, mac)
	return nil
}

func validateOverlayMACAllocation(allocation *unstructured.Unstructured, app *appv1alpha1.Application, instanceKey, mac string) error {
	claimedMAC, _, _ := unstructured.NestedString(allocation.Object, "spec", "mac")
	claimedInstance, _, _ := unstructured.NestedString(allocation.Object, "spec", "instanceKey")
	claimedUID, _, _ := unstructured.NestedString(allocation.Object, "spec", "applicationUID")
	if claimedMAC != mac || claimedInstance != instanceKey || claimedUID != string(app.UID) {
		return fmt.Errorf("overlay MAC allocation %s is owned by another instance", allocation.GetName())
	}
	return nil
}

func (wh *Webhook) persistOverlayMAC(ctx context.Context, applicationName string, uid types.UID, ordinal string, hasOrdinal bool, mac string) error {
	return retry.RetryOnConflict(retry.DefaultRetry, func() error {
		app, err := wh.dynamicClient.AppV1alpha1().Applications().Get(ctx, applicationName, metav1.GetOptions{})
		if err != nil {
			return err
		}
		if app.UID != uid {
			return fmt.Errorf("application %q UID changed while allocating overlay MAC", applicationName)
		}
		copy := app.DeepCopy()
		if copy.Spec.Settings == nil {
			copy.Spec.Settings = map[string]string{}
		}
		if hasOrdinal {
			values := map[string]string{}
			if raw := copy.Spec.Settings[overlayMACByOrdinalSetting]; raw != "" {
				if err := json.Unmarshal([]byte(raw), &values); err != nil {
					return fmt.Errorf("invalid persisted overlay MAC map: %w", err)
				}
			}
			if existing := values[ordinal]; existing != "" && existing != mac {
				return fmt.Errorf("ordinal %s already has a different overlay MAC", ordinal)
			}
			values[ordinal] = mac
			raw, err := json.Marshal(values)
			if err != nil {
				return err
			}
			copy.Spec.Settings[overlayMACByOrdinalSetting] = string(raw)
		} else {
			if existing := copy.Spec.Settings[overlayMACSetting]; existing != "" && existing != mac {
				return fmt.Errorf("application already has a different overlay MAC")
			}
			copy.Spec.Settings[overlayMACSetting] = mac
		}
		addString(&copy.Finalizers, overlayMACFinalizer)
		_, err = wh.dynamicClient.AppV1alpha1().Applications().Update(ctx, copy, metav1.UpdateOptions{})
		return err
	})
}

func (wh *Webhook) ensureOverlayMACFinalizer(ctx context.Context, applicationName string, uid types.UID) error {
	return retry.RetryOnConflict(retry.DefaultRetry, func() error {
		app, err := wh.dynamicClient.AppV1alpha1().Applications().Get(ctx, applicationName, metav1.GetOptions{})
		if err != nil {
			return err
		}
		if app.UID != uid {
			return fmt.Errorf("application %q UID changed while ensuring overlay MAC finalizer", applicationName)
		}
		if containsString(app.Finalizers, overlayMACFinalizer) {
			return nil
		}
		copy := app.DeepCopy()
		addString(&copy.Finalizers, overlayMACFinalizer)
		_, err = wh.dynamicClient.AppV1alpha1().Applications().Update(ctx, copy, metav1.UpdateOptions{})
		return err
	})
}

func addString(values *[]string, value string) {
	for _, existing := range *values {
		if existing == value {
			return
		}
	}
	*values = append(*values, value)
}

func containsString(values []string, value string) bool {
	for _, existing := range values {
		if existing == value {
			return true
		}
	}
	return false
}

func (wh *Webhook) overlayMACInstanceKey(ctx context.Context, pod *corev1.Pod, appName string) (string, error) {
	base := pod.Namespace + "/" + appName
	for _, owner := range pod.OwnerReferences {
		if owner.Controller == nil || !*owner.Controller {
			continue
		}
		switch owner.Kind {
		case "StatefulSet":
			ordinal, ok := statefulSetOrdinal(pod)
			if !ok {
				return "", fmt.Errorf("statefulset pod %s/%s has no valid ordinal", pod.Namespace, pod.Name)
			}
			return base + "/" + ordinal, nil
		case "ReplicaSet":
			if wh.kubeClient == nil {
				return "", errors.New("kubernetes client is required to validate Deployment replicas")
			}
			rs, err := wh.kubeClient.AppsV1().ReplicaSets(pod.Namespace).Get(ctx, owner.Name, metav1.GetOptions{})
			if err != nil {
				return "", fmt.Errorf("get pod ReplicaSet %s: %w", owner.Name, err)
			}
			hasDeploymentOwner := false
			for _, rsOwner := range rs.OwnerReferences {
				if rsOwner.Controller == nil || !*rsOwner.Controller || rsOwner.Kind != "Deployment" {
					continue
				}
				hasDeploymentOwner = true
				deployment, err := wh.kubeClient.AppsV1().Deployments(pod.Namespace).Get(ctx, rsOwner.Name, metav1.GetOptions{})
				if err != nil {
					return "", fmt.Errorf("get pod Deployment %s: %w", rsOwner.Name, err)
				}
				if deployment.Spec.Replicas != nil && *deployment.Spec.Replicas > 1 {
					return "", fmt.Errorf("Deployment %s has %d replicas; stable per-replica identity is required for macvlan", deployment.Name, *deployment.Spec.Replicas)
				}
			}
			if !hasDeploymentOwner {
				return "", fmt.Errorf("ReplicaSet %s has no Deployment owner; stable per-replica identity is required for macvlan", rs.Name)
			}
		}
	}
	return base, nil
}

func statefulSetOrdinal(pod *corev1.Pod) (string, bool) {
	if value := pod.Labels["apps.kubernetes.io/pod-index"]; value != "" {
		if _, err := strconv.Atoi(value); err == nil && !strings.HasPrefix(value, "-") {
			return value, true
		}
	}
	parts := strings.Split(pod.Name, "-")
	if len(parts) < 2 {
		return "", false
	}
	ordinal := parts[len(parts)-1]
	if _, err := strconv.Atoi(ordinal); err != nil || strings.HasPrefix(ordinal, "-") {
		return "", false
	}
	return ordinal, true
}

func (wh *Webhook) ensureMasterPlacement(ctx context.Context, pod *corev1.Pod) error {
	if pod.Spec.NodeName != "" {
		if wh.kubeClient == nil {
			return errors.New("kubernetes client is required to validate the assigned master node")
		}
		node, err := wh.kubeClient.CoreV1().Nodes().Get(ctx, pod.Spec.NodeName, metav1.GetOptions{})
		if err != nil {
			return fmt.Errorf("get assigned node %s: %w", pod.Spec.NodeName, err)
		}
		if _, ok := node.Labels[overlayMACMasterNodeLabel]; !ok {
			return fmt.Errorf("macvlan pod is assigned to non-master node %s", pod.Spec.NodeName)
		}
		return nil
	}
	if pod.Spec.Affinity == nil {
		pod.Spec.Affinity = &corev1.Affinity{}
	}
	if pod.Spec.Affinity.NodeAffinity == nil {
		pod.Spec.Affinity.NodeAffinity = &corev1.NodeAffinity{}
	}
	required := pod.Spec.Affinity.NodeAffinity.RequiredDuringSchedulingIgnoredDuringExecution
	if required == nil {
		required = &corev1.NodeSelector{}
		pod.Spec.Affinity.NodeAffinity.RequiredDuringSchedulingIgnoredDuringExecution = required
	}
	if len(required.NodeSelectorTerms) == 0 {
		required.NodeSelectorTerms = []corev1.NodeSelectorTerm{{}}
	}
	for i := range required.NodeSelectorTerms {
		term := &required.NodeSelectorTerms[i]
		found := false
		for _, requirement := range term.MatchExpressions {
			if requirement.Key != overlayMACMasterNodeLabel {
				continue
			}
			found = true
			if requirement.Operator == corev1.NodeSelectorOpNotIn || requirement.Operator == corev1.NodeSelectorOpDoesNotExist {
				return fmt.Errorf("existing node affinity excludes master nodes")
			}
		}
		if !found {
			term.MatchExpressions = append(term.MatchExpressions, corev1.NodeSelectorRequirement{
				Key:      overlayMACMasterNodeLabel,
				Operator: corev1.NodeSelectorOpExists,
			})
		}
	}
	return nil
}

func macvlanNetworkAnnotation(existing, mac string) (string, error) {
	selection := []map[string]interface{}{}
	switch strings.TrimSpace(existing) {
	case "", "kube-system/underlay-macvlan":
		selection = append(selection, map[string]interface{}{
			"name":      "underlay-macvlan",
			"namespace": "kube-system",
			"mac":       mac,
		})
	default:
		if err := json.Unmarshal([]byte(existing), &selection); err != nil {
			return "", fmt.Errorf("unsupported existing Multus networks annotation: %w", err)
		}
		found := false
		for i := range selection {
			name, _, _ := unstructured.NestedString(selection[i], "name")
			namespace, _, _ := unstructured.NestedString(selection[i], "namespace")
			if name == "underlay-macvlan" && (namespace == "" || namespace == "kube-system") {
				selection[i]["namespace"] = "kube-system"
				selection[i]["mac"] = mac
				found = true
			}
		}
		if !found {
			selection = append(selection, map[string]interface{}{
				"name":      "underlay-macvlan",
				"namespace": "kube-system",
				"mac":       mac,
			})
		}
	}
	raw, err := json.Marshal(selection)
	if err != nil {
		return "", fmt.Errorf("marshal Multus networks annotation: %w", err)
	}
	return string(raw), nil
}
