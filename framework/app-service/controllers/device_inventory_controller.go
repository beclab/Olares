package controllers

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"reflect"
	"sort"
	"strings"
	"sync"
	"time"

	"github.com/beclab/Olares/framework/app-service/pkg/appcfg"
	"github.com/beclab/Olares/framework/app-service/pkg/deviceinventory"
	appevent "github.com/beclab/Olares/framework/app-service/pkg/event"
	"github.com/beclab/Olares/framework/app-service/pkg/users"
	appv1alpha1 "github.com/beclab/api/api/app.bytetrade.io/v1alpha1"
	iamv1alpha2 "github.com/beclab/api/iam/v1alpha2"
	"github.com/google/uuid"

	coordinationv1 "k8s.io/api/coordination/v1"
	corev1 "k8s.io/api/core/v1"
	apierrors "k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/types"
	"k8s.io/klog/v2"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/builder"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/controller"
	"sigs.k8s.io/controller-runtime/pkg/event"
	"sigs.k8s.io/controller-runtime/pkg/handler"
	"sigs.k8s.io/controller-runtime/pkg/predicate"
	"sigs.k8s.io/controller-runtime/pkg/reconcile"
)

const (
	deviceInventorySubject = "os.device.inventory"
	deviceInventoryTopic   = "device.inventory.changed"
	deviceLeaseNamespace   = "kube-node-lease"
	deviceLeasePrefix      = "generic-device-plugin-"
	defaultLeaseDuration   = 30 * time.Second
)

// DeviceState is the observable state of one logical device. Identity fields
// remain on DeviceInventoryChange so previous/current only contain mutable
// state.
type DeviceState struct {
	DisplayName  string   `json:"displayName,omitempty"`
	Online       bool     `json:"online"`
	Selectable   bool     `json:"selectable"`
	KernelPaths  []string `json:"kernelPaths,omitempty"`
	PortUnstable bool     `json:"portUnstable,omitempty"`
	Reason       string   `json:"reason,omitempty"`
	Adapter      string   `json:"adapter,omitempty"`
	Address      string   `json:"address,omitempty"`
	AccessMode   string   `json:"accessMode,omitempty"`
}

type DeviceBinding struct {
	AppName      string   `json:"appName"`
	AppNamespace string   `json:"appNamespace,omitempty"`
	AppOwner     string   `json:"appOwner,omitempty"`
	Workload     string   `json:"workload"`
	Containers   []string `json:"containers"`
}

type DeviceInventoryChange struct {
	Kind         string          `json:"kind"`
	Action       string          `json:"action"`
	Type         string          `json:"type,omitempty"`
	UID          string          `json:"uid,omitempty"`
	ResourceName string          `json:"resourceName,omitempty"`
	DisplayName  string          `json:"displayName,omitempty"`
	Previous     *DeviceState    `json:"previous"`
	Current      *DeviceState    `json:"current"`
	Reason       string          `json:"reason,omitempty"`
	Bindings     []DeviceBinding `json:"bindings,omitempty"`
}

type DeviceInventoryEvent struct {
	SchemaVersion    int                     `json:"schemaVersion"`
	EventID          string                  `json:"eventID"`
	Topic            string                  `json:"topic"`
	OccurredAt       time.Time               `json:"occurredAt"`
	Node             string                  `json:"node"`
	PublisherOnline  bool                    `json:"publisherOnline"`
	PreviousRevision string                  `json:"previousRevision,omitempty"`
	Revision         string                  `json:"revision,omitempty"`
	Changes          []DeviceInventoryChange `json:"changes"`
}

type deviceEventRecipient struct {
	Username         string
	Privileged       bool
	PublisherVisible bool
	DeviceKeys       map[string]struct{}
}

type deviceSnapshot struct {
	Type         string
	UID          string
	ResourceName string
	State        DeviceState
}

type nodeDeviceSnapshot struct {
	NodeName        string
	Revision        string
	PublisherOnline bool
	Devices         map[string]deviceSnapshot
}

// DeviceInventoryController watches the generic-device-plugin Node annotation
// and heartbeat Lease, then publishes normalized changes through the shared
// app-service event queue.
type DeviceInventoryController struct {
	client.Client
	Publish func(subject, msgID string, data interface{})
	Now     func() time.Time

	mu        sync.Mutex
	snapshots map[string]nodeDeviceSnapshot
}

//+kubebuilder:rbac:groups="",resources=nodes,verbs=get;list;watch
//+kubebuilder:rbac:groups=coordination.k8s.io,resources=leases,verbs=get;list;watch

func (r *DeviceInventoryController) SetupWithManager(mgr ctrl.Manager) error {
	if r.Publish == nil {
		r.Publish = appevent.PublishToQueueWithID
	}
	if r.Now == nil {
		r.Now = time.Now
	}
	return ctrl.NewControllerManagedBy(mgr).
		Named("device-inventory").
		For(&corev1.Node{}, builder.WithPredicates(deviceInventoryNodePredicate())).
		Watches(
			&coordinationv1.Lease{},
			handler.EnqueueRequestsFromMapFunc(deviceLeaseRequest),
			builder.WithPredicates(deviceLeasePredicate()),
		).
		WithOptions(controller.Options{MaxConcurrentReconciles: 1}).
		Complete(r)
}

func (r *DeviceInventoryController) Reconcile(ctx context.Context, req ctrl.Request) (ctrl.Result, error) {
	if r.Now == nil {
		r.Now = time.Now
	}
	if r.Publish == nil {
		r.Publish = appevent.PublishToQueueWithID
	}
	now := r.Now().UTC()
	var node corev1.Node
	if err := r.Get(ctx, types.NamespacedName{Name: req.Name}, &node); err != nil {
		if apierrors.IsNotFound(err) {
			return ctrl.Result{}, r.handleNodeDeleted(ctx, req.Name, now)
		}
		return ctrl.Result{}, err
	}

	parsed, err := parseNodeInventory(&node)
	if err != nil {
		return ctrl.Result{}, fmt.Errorf("parse device inventory for node %s: %w", node.Name, err)
	}
	online, requeueAfter, err := r.publisherState(ctx, node.Name, now)
	if err != nil {
		return ctrl.Result{}, err
	}
	parsed.PublisherOnline = online

	r.mu.Lock()
	if r.snapshots == nil {
		r.snapshots = make(map[string]nodeDeviceSnapshot)
	}
	previous, existed := r.snapshots[node.Name]
	if !existed {
		r.snapshots[node.Name] = parsed
		r.mu.Unlock()
		return ctrl.Result{RequeueAfter: requeueAfter}, nil
	}
	// While the publisher is offline the annotation is only stale last-known
	// state. Preserve the last accepted inventory and publish one publisher
	// transition rather than a storm of device removals.
	if !online {
		parsed.Devices = previous.Devices
		parsed.Revision = previous.Revision
	}
	changes := diffSnapshots(previous, parsed)
	r.mu.Unlock()

	if err := r.enrichDeviceBindings(ctx, node.Name, changes); err != nil {
		return ctrl.Result{}, err
	}
	if len(changes) > 0 {
		e := buildDeviceInventoryEvent(previous, parsed, changes, now)
		if err := r.publishToRecipients(ctx, e); err != nil {
			return ctrl.Result{}, err
		}
		klog.Infof("generated device inventory event node=%s eventID=%s revision=%s changes=%d", node.Name, e.EventID, e.Revision, len(e.Changes))
	}
	r.mu.Lock()
	r.snapshots[node.Name] = parsed
	r.mu.Unlock()
	return ctrl.Result{RequeueAfter: requeueAfter}, nil
}

func (r *DeviceInventoryController) enrichDeviceBindings(ctx context.Context, nodeName string, changes []DeviceInventoryChange) error {
	affected := make(map[string][]int)
	for i := range changes {
		change := &changes[i]
		if changeNeedsBindings(change) {
			key := changeBindingKey(change.Type, change.ResourceName)
			affected[key] = append(affected[key], i)
		}
	}
	if len(affected) == 0 {
		return nil
	}

	var applications appv1alpha1.ApplicationList
	if err := r.List(ctx, &applications); err != nil {
		return fmt.Errorf("list applications for device bindings: %w", err)
	}
	var managers appv1alpha1.ApplicationManagerList
	if err := r.List(ctx, &managers); err != nil {
		klog.Warningf("list application managers for device binding containers: %v", err)
	}
	managerByName := make(map[string]*appv1alpha1.ApplicationManager, len(managers.Items))
	managerByApp := make(map[string]*appv1alpha1.ApplicationManager, len(managers.Items))
	for i := range managers.Items {
		manager := &managers.Items[i]
		managerByName[manager.Name] = manager
		managerByApp[applicationIdentity(manager.Spec.AppName, manager.Spec.AppNamespace, manager.Spec.AppOwner)] = manager
	}

	bindingsByChange := make(map[int]map[string]DeviceBinding)
	for i := range applications.Items {
		application := &applications.Items[i]
		var containersByCapability map[string][]string
		for _, attachment := range application.Spec.Attachments {
			indexes := affected[changeBindingKey(attachment.Type, attachment.Ref)]
			if len(indexes) == 0 || attachment.Node != nodeName {
				continue
			}
			if containersByCapability == nil {
				manager := managerByName[application.Name]
				if manager == nil {
					manager = managerByApp[applicationIdentity(application.Spec.Name, application.Spec.Namespace, application.Spec.Owner)]
				}
				if manager == nil {
					klog.Warningf("application manager config unavailable for device binding application=%s", application.Name)
				}
				containersByCapability = workloadContainers(manager)
			}
			containers := append([]string(nil), containersByCapability[attachment.Workload+"\x00"+attachment.Type]...)
			if containers == nil {
				containers = []string{}
			}
			sort.Strings(containers)
			binding := DeviceBinding{
				AppName: application.Spec.Name, AppNamespace: application.Spec.Namespace,
				AppOwner: application.Spec.Owner, Workload: attachment.Workload, Containers: containers,
			}
			bindingKey := strings.Join([]string{binding.AppNamespace, binding.AppName, binding.AppOwner, binding.Workload, strings.Join(binding.Containers, ",")}, "\x00")
			for _, index := range indexes {
				if bindingsByChange[index] == nil {
					bindingsByChange[index] = make(map[string]DeviceBinding)
				}
				bindingsByChange[index][bindingKey] = binding
			}
		}
	}
	for index, bindings := range bindingsByChange {
		ordered := make([]DeviceBinding, 0, len(bindings))
		for _, binding := range bindings {
			ordered = append(ordered, binding)
		}
		sort.Slice(ordered, func(i, j int) bool {
			left := strings.Join([]string{ordered[i].AppNamespace, ordered[i].AppName, ordered[i].Workload}, "\x00")
			right := strings.Join([]string{ordered[j].AppNamespace, ordered[j].AppName, ordered[j].Workload}, "\x00")
			return left < right
		})
		changes[index].Bindings = ordered
	}
	return nil
}

func changeNeedsBindings(change *DeviceInventoryChange) bool {
	if change == nil || change.Kind != "device" {
		return false
	}
	switch change.Action {
	case "added", "removed":
		return true
	case "updated":
		return change.Previous != nil && change.Current != nil && change.Previous.Online != change.Current.Online
	default:
		return false
	}
}

func changeBindingKey(deviceType, resourceName string) string {
	return deviceType + "\x00" + resourceName
}

func applicationIdentity(name, namespace, owner string) string {
	return strings.Join([]string{name, namespace, owner}, "\x00")
}

func workloadContainers(manager *appv1alpha1.ApplicationManager) map[string][]string {
	result := make(map[string][]string)
	if manager == nil || manager.Spec.Config == "" {
		return result
	}
	var cfg appcfg.ApplicationConfig
	if err := json.Unmarshal([]byte(manager.Spec.Config), &cfg); err != nil {
		klog.Warningf("parse application manager %s workload config for device binding: %v", manager.Name, err)
		return result
	}
	for workload, option := range cfg.WorkloadOptions {
		for _, capability := range option.Allow {
			key := workload + "\x00" + capability.Type
			result[key] = append(result[key], capability.Containers...)
		}
	}
	for key, containers := range result {
		sort.Strings(containers)
		result[key] = compactStrings(containers)
	}
	return result
}

func compactStrings(values []string) []string {
	if len(values) < 2 {
		return values
	}
	result := values[:1]
	for _, value := range values[1:] {
		if value != result[len(result)-1] {
			result = append(result, value)
		}
	}
	return result
}

func (r *DeviceInventoryController) publisherState(ctx context.Context, nodeName string, now time.Time) (bool, time.Duration, error) {
	var lease coordinationv1.Lease
	err := r.Get(ctx, types.NamespacedName{Namespace: deviceLeaseNamespace, Name: deviceLeasePrefix + nodeName}, &lease)
	if apierrors.IsNotFound(err) {
		return false, 0, nil
	}
	if err != nil {
		return false, 0, fmt.Errorf("get device publisher lease for node %s: %w", nodeName, err)
	}
	if lease.Spec.RenewTime == nil {
		return false, 0, nil
	}
	duration := defaultLeaseDuration
	if lease.Spec.LeaseDurationSeconds != nil && *lease.Spec.LeaseDurationSeconds > 0 {
		duration = time.Duration(*lease.Spec.LeaseDurationSeconds) * time.Second
	}
	expiresAt := lease.Spec.RenewTime.Time.Add(duration)
	if !now.Before(expiresAt) {
		return false, 0, nil
	}
	return true, expiresAt.Sub(now), nil
}

func (r *DeviceInventoryController) handleNodeDeleted(ctx context.Context, nodeName string, now time.Time) error {
	r.mu.Lock()
	if r.snapshots == nil {
		r.mu.Unlock()
		return nil
	}
	previous, existed := r.snapshots[nodeName]
	r.mu.Unlock()
	if !existed {
		return nil
	}
	current := nodeDeviceSnapshot{NodeName: nodeName, PublisherOnline: false, Devices: map[string]deviceSnapshot{}}
	changes := []DeviceInventoryChange{{Kind: "publisher", Action: "offline", Reason: "node_deleted"}}
	e := buildDeviceInventoryEvent(previous, current, changes, now)
	if err := r.publishToRecipients(ctx, e); err != nil {
		return err
	}
	r.mu.Lock()
	delete(r.snapshots, nodeName)
	r.mu.Unlock()
	return nil
}

// publishToRecipients fans a cluster device event out to every owner/admin
// user and to owners of applications bound to an affected device. Publisher
// transitions affect every attachment on the node. A recipient-specific event
// ID is required because JetStream message de-duplication is stream-wide, not
// scoped to an individual subject.
func (r *DeviceInventoryController) publishToRecipients(ctx context.Context, event DeviceInventoryEvent) error {
	recipients, err := r.eventRecipients(ctx, event.Node, event.Changes)
	if err != nil {
		return err
	}
	if len(recipients) == 0 {
		klog.Infof("device inventory fan-out: no recipients for node %s, dropping event %s", event.Node, event.EventID)
		return nil
	}
	for _, recipient := range recipients {
		recipientEvent := eventForRecipient(event, recipient)
		if len(recipientEvent.Changes) == 0 {
			continue
		}
		recipientEvent.EventID = event.EventID + ":" + recipient.Username
		r.Publish(deviceInventorySubject+"."+recipient.Username, recipientEvent.EventID, recipientEvent)
	}
	return nil
}

func (r *DeviceInventoryController) eventRecipients(ctx context.Context, nodeName string, changes []DeviceInventoryChange) ([]deviceEventRecipient, error) {
	recipients := make(map[string]*deviceEventRecipient)
	var userList iamv1alpha2.UserList
	if err := r.List(ctx, &userList); err != nil {
		return nil, fmt.Errorf("list owner/admin users for device inventory event: %w", err)
	}
	for i := range userList.Items {
		user := &userList.Items[i]
		role := user.Annotations[users.UserAnnotationOwnerRole]
		if role == "owner" || role == "admin" {
			recipient := ensureDeviceEventRecipient(recipients, user.Name)
			if recipient != nil {
				recipient.Privileged = true
			}
		}
	}

	deviceKeys := make(map[string]struct{})
	publisherChanged := false
	for i := range changes {
		change := &changes[i]
		if change.Kind == "publisher" {
			publisherChanged = true
			continue
		}
		if change.Kind == "device" {
			deviceKeys[changeBindingKey(change.Type, change.ResourceName)] = struct{}{}
		}
	}
	if len(deviceKeys) == 0 && !publisherChanged {
		return sortedDeviceEventRecipients(recipients), nil
	}

	var applications appv1alpha1.ApplicationList
	if err := r.List(ctx, &applications); err != nil {
		return nil, fmt.Errorf("list bound applications for device inventory recipients: %w", err)
	}
	for i := range applications.Items {
		application := &applications.Items[i]
		for _, attachment := range application.Spec.Attachments {
			if attachment.Node != nodeName {
				continue
			}
			key := changeBindingKey(attachment.Type, attachment.Ref)
			_, deviceAffected := deviceKeys[key]
			if !publisherChanged && !deviceAffected {
				continue
			}
			recipient := ensureDeviceEventRecipient(recipients, application.Spec.Owner)
			if recipient == nil {
				continue
			}
			if publisherChanged {
				recipient.PublisherVisible = true
			}
			if deviceAffected {
				recipient.DeviceKeys[key] = struct{}{}
			}
		}
	}
	return sortedDeviceEventRecipients(recipients), nil
}

func ensureDeviceEventRecipient(recipients map[string]*deviceEventRecipient, username string) *deviceEventRecipient {
	username = strings.TrimSpace(username)
	if username == "" {
		return nil
	}
	recipient := recipients[username]
	if recipient == nil {
		recipient = &deviceEventRecipient{Username: username, DeviceKeys: make(map[string]struct{})}
		recipients[username] = recipient
	}
	return recipient
}

func sortedDeviceEventRecipients(recipients map[string]*deviceEventRecipient) []deviceEventRecipient {
	result := make([]deviceEventRecipient, 0, len(recipients))
	for recipient := range recipients {
		result = append(result, *recipients[recipient])
	}
	sort.Slice(result, func(i, j int) bool { return result[i].Username < result[j].Username })
	return result
}

func eventForRecipient(event DeviceInventoryEvent, recipient deviceEventRecipient) DeviceInventoryEvent {
	if recipient.Privileged {
		return event
	}
	result := event
	result.Changes = make([]DeviceInventoryChange, 0, len(event.Changes))
	for i := range event.Changes {
		change := event.Changes[i]
		switch change.Kind {
		case "publisher":
			if !recipient.PublisherVisible {
				continue
			}
		case "device":
			if _, visible := recipient.DeviceKeys[changeBindingKey(change.Type, change.ResourceName)]; !visible {
				continue
			}
			if len(change.Bindings) > 0 {
				bindings := make([]DeviceBinding, 0, len(change.Bindings))
				for _, binding := range change.Bindings {
					if binding.AppOwner == recipient.Username {
						bindings = append(bindings, binding)
					}
				}
				change.Bindings = bindings
			}
		default:
			continue
		}
		result.Changes = append(result.Changes, change)
	}
	return result
}

func parseNodeInventory(node *corev1.Node) (nodeDeviceSnapshot, error) {
	result := nodeDeviceSnapshot{NodeName: node.Name, Devices: make(map[string]deviceSnapshot)}
	raw := ""
	if node.Annotations != nil {
		raw = node.Annotations[deviceinventory.Annotation]
	}
	var inventory deviceinventory.Inventory
	if raw == "" {
		inventory.SchemaVersion = 1
	} else if err := json.Unmarshal([]byte(raw), &inventory); err != nil {
		return result, err
	}
	if inventory.SchemaVersion != 1 {
		return result, fmt.Errorf("unsupported schemaVersion %d", inventory.SchemaVersion)
	}
	byType := []struct {
		typeName string
		items    []deviceinventory.Item
	}{
		{"device.serial", inventory.Serial},
		{"device.video", inventory.Video},
		{"device.audio", inventory.Audio},
		{"device.hid", inventory.HID},
		{"device.bluetooth", inventory.Bluetooth},
	}
	for _, group := range byType {
		for _, item := range group.items {
			if item.ResourceName == "" {
				return result, fmt.Errorf("%s item has empty resourceName", group.typeName)
			}
			if _, duplicate := result.Devices[item.ResourceName]; duplicate {
				return result, fmt.Errorf("duplicate resourceName %q", item.ResourceName)
			}
			paths := append([]string(nil), item.KernelPaths...)
			sort.Strings(paths)
			result.Devices[item.ResourceName] = deviceSnapshot{
				Type: group.typeName, UID: item.UID, ResourceName: item.ResourceName,
				State: DeviceState{
					DisplayName: item.DisplayName, Online: item.Online, Selectable: item.Selectable,
					KernelPaths: paths, PortUnstable: item.PortUnstable, Reason: item.Reason,
					Adapter: item.Adapter, Address: item.Address, AccessMode: item.AccessMode,
				},
			}
		}
	}
	revision, err := snapshotRevision(result.Devices)
	if err != nil {
		return result, err
	}
	result.Revision = revision
	return result, nil
}

func snapshotRevision(devices map[string]deviceSnapshot) (string, error) {
	ordered := make([]deviceSnapshot, 0, len(devices))
	for _, device := range devices {
		ordered = append(ordered, device)
	}
	sort.Slice(ordered, func(i, j int) bool {
		if ordered[i].Type != ordered[j].Type {
			return ordered[i].Type < ordered[j].Type
		}
		return ordered[i].ResourceName < ordered[j].ResourceName
	})
	data, err := json.Marshal(ordered)
	if err != nil {
		return "", err
	}
	sum := sha256.Sum256(data)
	return hex.EncodeToString(sum[:]), nil
}

func diffSnapshots(previous, current nodeDeviceSnapshot) []DeviceInventoryChange {
	changes := make([]DeviceInventoryChange, 0)
	switch {
	case previous.PublisherOnline && !current.PublisherOnline:
		changes = append(changes, DeviceInventoryChange{Kind: "publisher", Action: "offline"})
		return changes
	case !previous.PublisherOnline && current.PublisherOnline:
		changes = append(changes, DeviceInventoryChange{Kind: "publisher", Action: "online"})
	case !current.PublisherOnline:
		return changes
	}

	keys := make(map[string]struct{}, len(previous.Devices)+len(current.Devices))
	for key := range previous.Devices {
		keys[key] = struct{}{}
	}
	for key := range current.Devices {
		keys[key] = struct{}{}
	}
	orderedKeys := make([]string, 0, len(keys))
	for key := range keys {
		orderedKeys = append(orderedKeys, key)
	}
	sort.Strings(orderedKeys)
	for _, key := range orderedKeys {
		before, hadBefore := previous.Devices[key]
		after, hasAfter := current.Devices[key]
		switch {
		case !hadBefore && hasAfter:
			currentState := cloneDeviceState(after.State)
			changes = append(changes, changeForDevice("added", after, nil, &currentState, ""))
		case hadBefore && !hasAfter:
			previousState := cloneDeviceState(before.State)
			changes = append(changes, changeForDevice("removed", before, &previousState, nil, "device_removed"))
		case hadBefore && hasAfter && (before.Type != after.Type || before.UID != after.UID || !reflect.DeepEqual(before.State, after.State)):
			previousState := cloneDeviceState(before.State)
			currentState := cloneDeviceState(after.State)
			changes = append(changes, changeForDevice("updated", after, &previousState, &currentState, ""))
		}
	}
	return changes
}

func changeForDevice(action string, device deviceSnapshot, previous, current *DeviceState, reason string) DeviceInventoryChange {
	displayName := device.State.DisplayName
	if current == nil && previous != nil {
		displayName = previous.DisplayName
	}
	return DeviceInventoryChange{
		Kind: "device", Action: action, Type: device.Type, UID: device.UID,
		ResourceName: device.ResourceName, DisplayName: displayName,
		Previous: previous, Current: current, Reason: reason,
	}
}

func cloneDeviceState(state DeviceState) DeviceState {
	state.KernelPaths = append([]string(nil), state.KernelPaths...)
	return state
}

func buildDeviceInventoryEvent(previous, current nodeDeviceSnapshot, changes []DeviceInventoryChange, now time.Time) DeviceInventoryEvent {
	eventID := "device-inventory:" + current.NodeName + ":" + uuid.NewString()
	return DeviceInventoryEvent{
		SchemaVersion: 1, EventID: eventID, Topic: deviceInventoryTopic, OccurredAt: now,
		Node: current.NodeName, PublisherOnline: current.PublisherOnline,
		PreviousRevision: previous.Revision, Revision: current.Revision, Changes: changes,
	}
}

func deviceInventoryNodePredicate() predicate.Predicate {
	return predicate.Funcs{
		CreateFunc: func(event.CreateEvent) bool { return true },
		UpdateFunc: func(e event.UpdateEvent) bool {
			oldValue, newValue := "", ""
			if e.ObjectOld.GetAnnotations() != nil {
				oldValue = e.ObjectOld.GetAnnotations()[deviceinventory.Annotation]
			}
			if e.ObjectNew.GetAnnotations() != nil {
				newValue = e.ObjectNew.GetAnnotations()[deviceinventory.Annotation]
			}
			return oldValue != newValue
		},
		DeleteFunc: func(event.DeleteEvent) bool { return true },
	}
}

func deviceLeasePredicate() predicate.Predicate {
	return predicate.NewPredicateFuncs(func(obj client.Object) bool {
		return obj.GetNamespace() == deviceLeaseNamespace && strings.HasPrefix(obj.GetName(), deviceLeasePrefix)
	})
}

func deviceLeaseRequest(_ context.Context, obj client.Object) []reconcile.Request {
	if obj.GetNamespace() != deviceLeaseNamespace || !strings.HasPrefix(obj.GetName(), deviceLeasePrefix) {
		return nil
	}
	return []reconcile.Request{{NamespacedName: types.NamespacedName{Name: strings.TrimPrefix(obj.GetName(), deviceLeasePrefix)}}}
}
