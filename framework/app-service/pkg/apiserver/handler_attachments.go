package apiserver

import (
	"context"
	"encoding/json"
	"fmt"
	"net/url"
	"path"
	"regexp"
	"sort"
	"strconv"
	"strings"
	"time"

	"github.com/beclab/Olares/framework/app-service/pkg/apiserver/api"
	"github.com/beclab/Olares/framework/app-service/pkg/appcfg"
	"github.com/beclab/Olares/framework/app-service/pkg/client/clientset"
	"github.com/beclab/Olares/framework/app-service/pkg/constants"
	appv1alpha1 "github.com/beclab/api/api/app.bytetrade.io/v1alpha1"

	"github.com/emicklei/go-restful/v3"
	apierrors "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/util/retry"
)

const (
	deviceInventoryAnnotation = "bytetrade.io/devices"
	deviceLeaseNamespace      = "kube-node-lease"
	deviceLeasePrefix         = "generic-device-plugin-"
	attachmentsRevision       = "app.bytetrade.io/attachments-revision"
	bluetoothNodeLabel        = "devices.bytetrade.io/bluetooth-capable"
	deviceLeaseTimeout        = 30 * time.Second
)

var attachmentNamePattern = regexp.MustCompile(`^[a-z][a-z0-9-]{0,31}$`)

type DeviceInventoryItem struct {
	UID          string   `json:"uid"`
	ResourceName string   `json:"resourceName"`
	DisplayName  string   `json:"displayName,omitempty"`
	Online       bool     `json:"online"`
	Selectable   bool     `json:"selectable"`
	KernelPaths  []string `json:"kernelPaths,omitempty"`
	PortUnstable bool     `json:"portUnstable,omitempty"`
	Node         string   `json:"node"`
	Type         string   `json:"type"`
	Reason       string   `json:"reason,omitempty"`
}

type nodeDeviceInventory struct {
	SchemaVersion int                   `json:"schemaVersion"`
	Serial        []DeviceInventoryItem `json:"serial,omitempty"`
	Video         []DeviceInventoryItem `json:"video,omitempty"`
	Audio         []DeviceInventoryItem `json:"audio,omitempty"`
	HID           []DeviceInventoryItem `json:"hid,omitempty"`
	Truncated     bool                  `json:"truncated,omitempty"`
}

type DeviceInventoryResponse struct {
	Devices []DeviceInventoryItem `json:"devices"`
}

type AttachmentCapability struct {
	Type       string   `json:"type"`
	Containers []string `json:"containers"`
}

type AttachmentCapabilityWorkload struct {
	Name         string                 `json:"name"`
	Replicas     *int32                 `json:"replicas"`
	Capabilities []AttachmentCapability `json:"capabilities"`
}

type AttachmentCapabilitiesResponse struct {
	Workloads []AttachmentCapabilityWorkload `json:"workloads"`
}

type AttachmentsResponse struct {
	ResourceVersion string                   `json:"resourceVersion"`
	Attachments     []appv1alpha1.Attachment `json:"attachments"`
	Bluetooth       bool                     `json:"bluetooth"`
}

type UpdateAttachmentsRequest struct {
	ResourceVersion string                   `json:"resourceVersion"`
	Attachments     []appv1alpha1.Attachment `json:"attachments"`
}

func (h *Handler) getAttachmentCapabilities(req *restful.Request, resp *restful.Response) {
	app, err := getAppByName(req, resp)
	if err != nil {
		return
	}
	_, cfg, _, _, err := h.sidecarWebhook.GetAppConfig(app.Spec.Namespace)
	if err != nil {
		api.HandleError(resp, req, err)
		return
	}
	if cfg == nil {
		api.HandleBadRequest(resp, req, fmt.Errorf("application manifest is unavailable"))
		return
	}
	result := attachmentCapabilitiesResponse(cfg)
	if err := resp.WriteAsJson(&result); err != nil {
		api.HandleInternalError(resp, req, err)
	}
}

func attachmentCapabilitiesResponse(cfg *appcfg.ApplicationConfig) AttachmentCapabilitiesResponse {
	names := make([]string, 0, len(cfg.WorkloadOptions))
	for name := range cfg.WorkloadOptions {
		names = append(names, name)
	}
	sort.Strings(names)

	workloads := make([]AttachmentCapabilityWorkload, 0, len(names))
	for _, name := range names {
		option := cfg.WorkloadOptions[name]
		capabilities := make([]AttachmentCapability, 0, len(option.Allow))
		for _, capability := range option.Allow {
			capabilities = append(capabilities, AttachmentCapability{
				Type:       capability.Type,
				Containers: append([]string(nil), capability.Containers...),
			})
		}
		workloads = append(workloads, AttachmentCapabilityWorkload{
			Name:         name,
			Replicas:     option.Replicas,
			Capabilities: capabilities,
		})
	}
	return AttachmentCapabilitiesResponse{Workloads: workloads}
}

func (h *Handler) getAttachments(req *restful.Request, resp *restful.Response) {
	app, err := getAppByName(req, resp)
	if err != nil {
		return
	}
	result := attachmentsResponse(app)
	if err := resp.WriteAsJson(&result); err != nil {
		api.HandleInternalError(resp, req, err)
	}
}

func attachmentsResponse(app *appv1alpha1.Application) AttachmentsResponse {
	attachments := app.Spec.Attachments
	if attachments == nil {
		attachments = []appv1alpha1.Attachment{}
	}
	bluetooth, _ := strconv.ParseBool(app.Spec.Settings["bluetooth"])
	return AttachmentsResponse{ResourceVersion: app.ResourceVersion, Attachments: attachments, Bluetooth: bluetooth}
}

func (h *Handler) getDeviceInventory(req *restful.Request, resp *restful.Response) {
	app, err := getAppByName(req, resp)
	if err != nil {
		return
	}
	workload := req.QueryParameter("workload")
	if workload == "" {
		api.HandleBadRequest(resp, req, fmt.Errorf("workload is required"))
		return
	}
	_, cfg, _, _, err := h.sidecarWebhook.GetAppConfig(app.Spec.Namespace)
	if err != nil {
		api.HandleError(resp, req, err)
		return
	}
	allowed, err := deviceTypesForWorkload(cfg, workload)
	if err != nil {
		api.HandleBadRequest(resp, req, err)
		return
	}
	client := req.Attribute(constants.KubeSphereClientAttribute).(*clientset.ClientSet)
	items, err := collectDeviceInventory(req.Request.Context(), client, allowed)
	if err != nil {
		api.HandleError(resp, req, err)
		return
	}
	if err := applyDeviceReservations(req.Request.Context(), client, app, items); err != nil {
		api.HandleError(resp, req, err)
		return
	}
	seen := make(map[string]struct{}, len(items))
	for _, item := range items {
		seen[item.Node+"/"+item.Type+"/"+item.ResourceName] = struct{}{}
	}
	for _, attachment := range app.Spec.Attachments {
		if attachment.Workload != workload || !strings.HasPrefix(attachment.Type, "device.") {
			continue
		}
		key := deviceKey(attachment)
		if _, ok := seen[key]; ok {
			continue
		}
		items = append(items, DeviceInventoryItem{
			ResourceName: attachment.Ref,
			DisplayName:  attachment.Ref,
			Node:         attachment.Node,
			Type:         attachment.Type,
			Online:       false,
			Selectable:   false,
			Reason:       "bound device is offline",
		})
	}
	sortDeviceInventory(items)
	if err := resp.WriteAsJson(&DeviceInventoryResponse{Devices: items}); err != nil {
		api.HandleInternalError(resp, req, err)
	}
}

func deviceTypesForWorkload(cfg *appcfg.ApplicationConfig, workload string) (map[string]struct{}, error) {
	if cfg == nil {
		return nil, fmt.Errorf("application manifest is unavailable")
	}
	option, ok := cfg.WorkloadOptions[workload]
	if !ok {
		return nil, fmt.Errorf("workload %q is not declared in workloadOptions", workload)
	}
	allowed := make(map[string]struct{})
	for _, capability := range option.Allow {
		if strings.HasPrefix(capability.Type, "device.") {
			allowed[capability.Type] = struct{}{}
		}
	}
	return allowed, nil
}

func collectDeviceInventory(ctx context.Context, client *clientset.ClientSet, allowed map[string]struct{}) ([]DeviceInventoryItem, error) {
	nodes, err := client.KubeClient.Kubernetes().CoreV1().Nodes().List(ctx, metav1.ListOptions{})
	if err != nil {
		return nil, err
	}
	items := make([]DeviceInventoryItem, 0)
	for i := range nodes.Items {
		node := &nodes.Items[i]
		raw := node.Annotations[deviceInventoryAnnotation]
		if raw == "" {
			continue
		}
		var inventory nodeDeviceInventory
		if err := json.Unmarshal([]byte(raw), &inventory); err != nil || inventory.SchemaVersion != 1 {
			continue
		}
		lease, leaseErr := client.KubeClient.Kubernetes().CoordinationV1().Leases(deviceLeaseNamespace).Get(ctx, deviceLeasePrefix+node.Name, metav1.GetOptions{})
		leaseHealthy := leaseErr == nil && lease.Spec.RenewTime != nil && time.Since(lease.Spec.RenewTime.Time) <= deviceLeaseTimeout
		byType := map[string][]DeviceInventoryItem{
			"device.serial": inventory.Serial,
			"device.video":  inventory.Video,
			"device.audio":  inventory.Audio,
			"device.hid":    inventory.HID,
		}
		for capability, devices := range byType {
			if _, ok := allowed[capability]; !ok {
				continue
			}
			for _, item := range devices {
				item.Node = node.Name
				item.Type = capability
				if !leaseHealthy {
					item.Online = false
					item.Selectable = false
					item.Reason = "device inventory publisher is offline"
				}
				if item.DisplayName == "" {
					item.DisplayName = item.ResourceName
				}
				items = append(items, item)
			}
		}
	}
	sortDeviceInventory(items)
	return items, nil
}

func sortDeviceInventory(items []DeviceInventoryItem) {
	sort.Slice(items, func(i, j int) bool {
		if items[i].Node != items[j].Node {
			return items[i].Node < items[j].Node
		}
		if items[i].Type != items[j].Type {
			return items[i].Type < items[j].Type
		}
		if items[i].UID != items[j].UID {
			return items[i].UID < items[j].UID
		}
		return items[i].ResourceName < items[j].ResourceName
	})
}

func applyDeviceReservations(ctx context.Context, client *clientset.ClientSet, current *appv1alpha1.Application, items []DeviceInventoryItem) error {
	applications, err := client.AppClient.AppV1alpha1().Applications().List(ctx, metav1.ListOptions{})
	if err != nil {
		return err
	}
	reserved := make(map[string]string)
	for i := range applications.Items {
		application := &applications.Items[i]
		if application.Name == current.Name {
			continue
		}
		for _, attachment := range application.Spec.Attachments {
			if strings.HasPrefix(attachment.Type, "device.") {
				reserved[deviceKey(attachment)] = application.Spec.Name
			}
		}
	}
	for i := range items {
		key := items[i].Node + "/" + items[i].Type + "/" + items[i].ResourceName
		if appName, exists := reserved[key]; exists {
			items[i].Selectable = false
			items[i].Reason = fmt.Sprintf("device is attached to application %s", appName)
		}
	}
	return nil
}

func (h *Handler) updateAttachments(req *restful.Request, resp *restful.Response) {
	app, err := getAppByName(req, resp)
	if err != nil {
		return
	}
	if !h.gateSharedAppWrite(req, resp, app) {
		return
	}
	var body UpdateAttachmentsRequest
	if err := req.ReadEntity(&body); err != nil {
		api.HandleBadRequest(resp, req, err)
		return
	}
	if body.ResourceVersion == "" || body.ResourceVersion != app.ResourceVersion {
		api.HandleConflict(resp, req, fmt.Errorf("application resourceVersion changed; reload attachments and retry"))
		return
	}
	_, cfg, _, _, err := h.sidecarWebhook.GetAppConfig(app.Spec.Namespace)
	if err != nil {
		api.HandleError(resp, req, err)
		return
	}
	client := req.Attribute(constants.KubeSphereClientAttribute).(*clientset.ClientSet)
	allowedDeviceTypes := allDeviceTypes(cfg)
	inventory, err := collectDeviceInventory(req.Request.Context(), client, allowedDeviceTypes)
	if err != nil {
		api.HandleError(resp, req, err)
		return
	}
	if err := applyDeviceReservations(req.Request.Context(), client, app, inventory); err != nil {
		api.HandleError(resp, req, err)
		return
	}
	if err := validateAttachments(body.Attachments, app.Spec.Attachments, cfg, inventory); err != nil {
		api.HandleUnprocessableEntity(resp, req, err)
		return
	}
	for _, attachment := range body.Attachments {
		if attachment.Type != "folder" {
			continue
		}
		if _, err := h.resolveAttachmentPath(req.Request.Context(), cfg, attachment); err != nil {
			api.HandleUnprocessableEntity(resp, req, err)
			return
		}
	}
	bluetoothEnabled, _ := strconv.ParseBool(app.Spec.Settings["bluetooth"])
	if bluetoothEnabled {
		workloads := bluetoothWorkloadSet(cfg)
		if err := validateBluetoothAttachmentNodes(req.Request.Context(), client, body.Attachments, workloads); err != nil {
			api.HandleUnprocessableEntity(resp, req, err)
			return
		}
	}

	updated := app.DeepCopy()
	updated.Spec.Attachments = append([]appv1alpha1.Attachment(nil), body.Attachments...)
	updated, err = client.AppClient.AppV1alpha1().Applications().Update(req.Request.Context(), updated, metav1.UpdateOptions{})
	if apierrors.IsConflict(err) {
		api.HandleConflict(resp, req, err)
		return
	}
	if err != nil {
		api.HandleError(resp, req, err)
		return
	}
	workloads := affectedWorkloads(app.Spec.Attachments, body.Attachments)
	if err := restartAttachmentWorkloads(req.Request.Context(), client, app.Spec.Namespace, workloads); err != nil {
		api.HandleError(resp, req, err)
		return
	}
	result := attachmentsResponse(updated)
	if err := resp.WriteAsJson(&result); err != nil {
		api.HandleInternalError(resp, req, err)
	}
}

func bluetoothWorkloadSet(cfg *appcfg.ApplicationConfig) map[string]struct{} {
	result := make(map[string]struct{})
	if cfg == nil {
		return result
	}
	for workload, option := range cfg.WorkloadOptions {
		if allowsCapability(option, "bluetooth") {
			result[workload] = struct{}{}
		}
	}
	return result
}

func validateBluetoothAttachmentNodes(ctx context.Context, client *clientset.ClientSet, attachments []appv1alpha1.Attachment, workloads map[string]struct{}) error {
	checked := make(map[string]struct{})
	for _, attachment := range attachments {
		if _, enabled := workloads[attachment.Workload]; !enabled || attachment.Node == "" {
			continue
		}
		if _, ok := checked[attachment.Node]; ok {
			continue
		}
		node, err := client.KubeClient.Kubernetes().CoreV1().Nodes().Get(ctx, attachment.Node, metav1.GetOptions{})
		if err != nil {
			return fmt.Errorf("check bluetooth node %s: %w", attachment.Node, err)
		}
		if node.Labels[bluetoothNodeLabel] != "true" {
			return fmt.Errorf("workload %q is pinned to node %q, which is not bluetooth capable", attachment.Workload, attachment.Node)
		}
		checked[attachment.Node] = struct{}{}
	}
	return nil
}

func allDeviceTypes(cfg *appcfg.ApplicationConfig) map[string]struct{} {
	result := make(map[string]struct{})
	for _, option := range cfg.WorkloadOptions {
		for _, capability := range option.Allow {
			if strings.HasPrefix(capability.Type, "device.") {
				result[capability.Type] = struct{}{}
			}
		}
	}
	return result
}

func validateAttachments(next, current []appv1alpha1.Attachment, cfg *appcfg.ApplicationConfig, inventory []DeviceInventoryItem) error {
	if cfg == nil {
		return fmt.Errorf("application manifest is unavailable")
	}
	currentDevices := make(map[string]struct{})
	for _, attachment := range current {
		if strings.HasPrefix(attachment.Type, "device.") {
			currentDevices[deviceKey(attachment)] = struct{}{}
		}
	}
	selectable := make(map[string]struct{})
	for _, item := range inventory {
		if item.Online && item.Selectable {
			selectable[item.Node+"/"+item.Type+"/"+item.ResourceName] = struct{}{}
		}
	}
	counts := make(map[string]int)
	folderNames := make(map[string]struct{})
	folderRefs := make(map[string]struct{})
	devices := make(map[string]struct{})
	workloadNodes := make(map[string]string)
	for i, attachment := range next {
		option, ok := cfg.WorkloadOptions[attachment.Workload]
		if !ok {
			return fmt.Errorf("attachments[%d]: workload %q is not declared", i, attachment.Workload)
		}
		if !allowsCapability(option, attachment.Type) {
			return fmt.Errorf("attachments[%d]: workload %q does not allow %q", i, attachment.Workload, attachment.Type)
		}
		countKey := attachment.Workload + "/" + attachment.Type
		counts[countKey]++
		if counts[countKey] > 8 {
			return fmt.Errorf("workload %q allows at most 8 %q attachments", attachment.Workload, attachment.Type)
		}
		switch {
		case attachment.Type == "folder":
			if !attachmentNamePattern.MatchString(attachment.Name) {
				return fmt.Errorf("attachments[%d]: invalid folder name %q", i, attachment.Name)
			}
			if err := validateFolderRef(attachment.Ref, attachment.Node, cfg.AppName); err != nil {
				return fmt.Errorf("attachments[%d]: %w", i, err)
			}
			nameKey := attachment.Workload + "/" + attachment.Name
			if _, exists := folderNames[nameKey]; exists {
				return fmt.Errorf("attachments[%d]: duplicate folder name %q in workload %q", i, attachment.Name, attachment.Workload)
			}
			folderNames[nameKey] = struct{}{}
			refKey := attachment.Workload + "/" + attachment.Ref
			if _, exists := folderRefs[refKey]; exists {
				return fmt.Errorf("attachments[%d]: duplicate folder ref %q in workload %q", i, attachment.Ref, attachment.Workload)
			}
			folderRefs[refKey] = struct{}{}
			if attachment.Node != "" {
				if previous := workloadNodes[attachment.Workload]; previous != "" && previous != attachment.Node {
					return fmt.Errorf("workload %q attachments require different nodes", attachment.Workload)
				}
				workloadNodes[attachment.Workload] = attachment.Node
			}
		case strings.HasPrefix(attachment.Type, "device."):
			if attachment.Name != "" || attachment.Node == "" || attachment.Ref == "" {
				return fmt.Errorf("attachments[%d]: device requires node/ref and forbids name", i)
			}
			key := deviceKey(attachment)
			if _, exists := devices[key]; exists {
				return fmt.Errorf("attachments[%d]: device %q is already bound", i, key)
			}
			devices[key] = struct{}{}
			if _, existed := currentDevices[key]; !existed {
				if _, ok := selectable[key]; !ok {
					return fmt.Errorf("attachments[%d]: device %q is offline or not selectable", i, key)
				}
			}
			if previous := workloadNodes[attachment.Workload]; previous != "" && previous != attachment.Node {
				return fmt.Errorf("workload %q attachments require different nodes", attachment.Workload)
			}
			workloadNodes[attachment.Workload] = attachment.Node
		default:
			return fmt.Errorf("attachments[%d]: unsupported type %q", i, attachment.Type)
		}
	}
	return nil
}

func allowsCapability(option appcfg.WorkloadOption, capabilityType string) bool {
	for _, capability := range option.Allow {
		if capability.Type == capabilityType {
			return true
		}
	}
	return false
}

func validateFolderRef(ref, node, appName string) error {
	if ref == "" || strings.ContainsAny(ref, "\\\x00") || strings.HasPrefix(ref, "/") || path.Clean(ref) != ref {
		return fmt.Errorf("invalid folder ref %q", ref)
	}
	decoded, err := url.PathUnescape(ref)
	if err != nil || decoded != ref {
		return fmt.Errorf("encoded folder ref is not allowed")
	}
	parts := strings.Split(ref, "/")
	for _, part := range parts {
		if part == "" || part == "." || part == ".." {
			return fmt.Errorf("invalid folder ref %q", ref)
		}
	}
	switch {
	case len(parts) >= 3 && parts[0] == "drive" && parts[1] == "Home":
	case len(parts) >= 4 && parts[0] == "drive" && parts[1] == "Data" && parts[2] == appName:
	case len(parts) >= 3 && parts[0] == "drive" && parts[1] == "Common":
	case len(parts) >= 4 && parts[0] == "external" && parts[1] != "" && parts[2] != "":
		if node == "" || node != parts[1] {
			return fmt.Errorf("external folder node must match ref")
		}
	default:
		return fmt.Errorf("folder ref %q is outside an allowed root", ref)
	}
	return nil
}

func deviceKey(attachment appv1alpha1.Attachment) string {
	return attachment.Node + "/" + attachment.Type + "/" + attachment.Ref
}

func affectedWorkloads(old, next []appv1alpha1.Attachment) []string {
	set := make(map[string]struct{})
	for _, attachment := range append(append([]appv1alpha1.Attachment{}, old...), next...) {
		set[attachment.Workload] = struct{}{}
	}
	result := make([]string, 0, len(set))
	for workload := range set {
		result = append(result, workload)
	}
	sort.Strings(result)
	return result
}

func restartAttachmentWorkloads(ctx context.Context, client *clientset.ClientSet, namespace string, workloads []string) error {
	for _, workload := range workloads {
		revision := time.Now().UTC().Format(time.RFC3339Nano)
		err := retry.RetryOnConflict(retry.DefaultRetry, func() error {
			deployment, err := client.KubeClient.Kubernetes().AppsV1().Deployments(namespace).Get(ctx, workload, metav1.GetOptions{})
			if apierrors.IsNotFound(err) {
				return err
			}
			if err != nil {
				return err
			}
			if deployment.Spec.Template.Annotations == nil {
				deployment.Spec.Template.Annotations = map[string]string{}
			}
			deployment.Spec.Template.Annotations[attachmentsRevision] = revision
			_, err = client.KubeClient.Kubernetes().AppsV1().Deployments(namespace).Update(ctx, deployment, metav1.UpdateOptions{})
			return err
		})
		if apierrors.IsNotFound(err) {
			err = retry.RetryOnConflict(retry.DefaultRetry, func() error {
				statefulSet, getErr := client.KubeClient.Kubernetes().AppsV1().StatefulSets(namespace).Get(ctx, workload, metav1.GetOptions{})
				if getErr != nil {
					return getErr
				}
				if statefulSet.Spec.Template.Annotations == nil {
					statefulSet.Spec.Template.Annotations = map[string]string{}
				}
				statefulSet.Spec.Template.Annotations[attachmentsRevision] = revision
				_, updateErr := client.KubeClient.Kubernetes().AppsV1().StatefulSets(namespace).Update(ctx, statefulSet, metav1.UpdateOptions{})
				return updateErr
			})
		}
		if err != nil {
			return fmt.Errorf("restart workload %s: %w", workload, err)
		}
	}
	return nil
}
