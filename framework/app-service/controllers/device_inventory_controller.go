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

	"github.com/beclab/Olares/framework/app-service/pkg/deviceinventory"
	appevent "github.com/beclab/Olares/framework/app-service/pkg/event"

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
			r.handleNodeDeleted(req.Name, now)
			return ctrl.Result{}, nil
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
	r.snapshots[node.Name] = parsed
	r.mu.Unlock()

	if len(changes) > 0 {
		e := buildDeviceInventoryEvent(previous, parsed, changes, now)
		r.Publish(deviceInventorySubject, e.EventID, e)
		klog.Infof("generated device inventory event node=%s eventID=%s revision=%s changes=%d", node.Name, e.EventID, e.Revision, len(e.Changes))
	}
	return ctrl.Result{RequeueAfter: requeueAfter}, nil
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

func (r *DeviceInventoryController) handleNodeDeleted(nodeName string, now time.Time) {
	r.mu.Lock()
	if r.snapshots == nil {
		r.mu.Unlock()
		return
	}
	previous, existed := r.snapshots[nodeName]
	if existed {
		delete(r.snapshots, nodeName)
	}
	r.mu.Unlock()
	if !existed {
		return
	}
	current := nodeDeviceSnapshot{NodeName: nodeName, PublisherOnline: false, Devices: map[string]deviceSnapshot{}}
	changes := []DeviceInventoryChange{{Kind: "publisher", Action: "offline", Reason: "node_deleted"}}
	e := buildDeviceInventoryEvent(previous, current, changes, now)
	r.Publish(deviceInventorySubject, e.EventID, e)
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
	eventIDInput := strings.Join([]string{
		current.NodeName, previous.Revision, current.Revision,
		fmt.Sprintf("%t", previous.PublisherOnline), fmt.Sprintf("%t", current.PublisherOnline),
	}, "\x00")
	sum := sha256.Sum256([]byte(eventIDInput))
	eventID := "device-inventory:" + current.NodeName + ":" + hex.EncodeToString(sum[:])
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
