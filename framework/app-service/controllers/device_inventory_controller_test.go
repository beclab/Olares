package controllers

import (
	"context"
	"encoding/json"
	"testing"
	"time"

	"github.com/beclab/Olares/framework/app-service/pkg/appcfg"
	"github.com/beclab/Olares/framework/app-service/pkg/users"
	appv1alpha1 "github.com/beclab/api/api/app.bytetrade.io/v1alpha1"
	iamv1alpha2 "github.com/beclab/api/iam/v1alpha2"
	coordinationv1 "k8s.io/api/coordination/v1"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/types"
	"k8s.io/utils/ptr"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/client/fake"
)

func TestParseNodeInventoryNormalizesOrdering(t *testing.T) {
	first := nodeWithInventory("node-a", `{"schemaVersion":1,"serial":[{"uid":"s-2","resourceName":"devices.bytetrade.io/serial-s-2","online":true,"selectable":true,"kernelPaths":["/dev/b","/dev/a"]},{"uid":"s-1","resourceName":"devices.bytetrade.io/serial-s-1","online":true,"selectable":true}]}`)
	second := nodeWithInventory("node-a", `{"schemaVersion":1,"serial":[{"uid":"s-1","resourceName":"devices.bytetrade.io/serial-s-1","online":true,"selectable":true},{"uid":"s-2","resourceName":"devices.bytetrade.io/serial-s-2","online":true,"selectable":true,"kernelPaths":["/dev/a","/dev/b"]}]}`)
	a, err := parseNodeInventory(first)
	if err != nil {
		t.Fatal(err)
	}
	b, err := parseNodeInventory(second)
	if err != nil {
		t.Fatal(err)
	}
	if a.Revision != b.Revision {
		t.Fatalf("equivalent inventory revisions differ: %s != %s", a.Revision, b.Revision)
	}
}

func TestDiffSnapshotsAddedRemovedAndUpdated(t *testing.T) {
	previous := snapshotForTest(true,
		deviceForTest("device.serial", "s-removed", "/dev/ttyUSB0"),
		deviceForTest("device.serial", "s-updated", "/dev/ttyUSB1"),
	)
	current := snapshotForTest(true,
		deviceForTest("device.serial", "s-updated", "/dev/ttyUSB2"),
		deviceForTest("device.bluetooth", "p-added", ""),
	)
	changes := diffSnapshots(previous, current)
	if len(changes) != 3 {
		t.Fatalf("changes = %#v", changes)
	}
	byResource := make(map[string]DeviceInventoryChange)
	for _, change := range changes {
		byResource[change.ResourceName] = change
	}
	if got := byResource["devices.bytetrade.io/serial-s-removed"]; got.Action != "removed" || got.Reason != "device_removed" || got.Previous == nil || got.Current != nil {
		t.Fatalf("removed change = %#v", got)
	}
	if got := byResource["devices.bytetrade.io/serial-s-updated"]; got.Action != "updated" || got.Previous == nil || got.Current == nil || got.Current.KernelPaths[0] != "/dev/ttyUSB2" {
		t.Fatalf("updated change = %#v", got)
	}
	if got := byResource["devices.bytetrade.io/bluetooth-p-added"]; got.Action != "added" || got.Previous != nil || got.Current == nil {
		t.Fatalf("added change = %#v", got)
	}
}

func TestDiffSnapshotsPublisherTransitionsDoNotRemoveDevices(t *testing.T) {
	previous := snapshotForTest(true, deviceForTest("device.serial", "s-1", "/dev/ttyUSB0"))
	current := snapshotForTest(false)
	changes := diffSnapshots(previous, current)
	if len(changes) != 1 || changes[0].Kind != "publisher" || changes[0].Action != "offline" {
		t.Fatalf("offline changes = %#v", changes)
	}

	previous.PublisherOnline = false
	current = snapshotForTest(true, deviceForTest("device.serial", "s-1", "/dev/ttyUSB0"))
	changes = diffSnapshots(previous, current)
	if len(changes) != 1 || changes[0].Kind != "publisher" || changes[0].Action != "online" {
		t.Fatalf("online changes = %#v", changes)
	}
}

func TestBuildDeviceInventoryEventHasDeterministicID(t *testing.T) {
	previous := snapshotForTest(true)
	previous.Revision = "before"
	current := snapshotForTest(true, deviceForTest("device.serial", "s-1", "/dev/ttyUSB0"))
	current.Revision = "after"
	changes := diffSnapshots(previous, current)
	first := buildDeviceInventoryEvent(previous, current, changes, time.Unix(1, 0))
	second := buildDeviceInventoryEvent(previous, current, changes, time.Unix(2, 0))
	if first.EventID != second.EventID {
		t.Fatalf("event IDs differ: %q != %q", first.EventID, second.EventID)
	}
}

func TestDeviceInventoryControllerInitialObservationIsBaselineOnly(t *testing.T) {
	scheme := runtime.NewScheme()
	if err := corev1.AddToScheme(scheme); err != nil {
		t.Fatal(err)
	}
	if err := coordinationv1.AddToScheme(scheme); err != nil {
		t.Fatal(err)
	}
	now := time.Date(2026, 11, 18, 8, 0, 0, 0, time.UTC)
	node := nodeWithInventory("node-a", `{"schemaVersion":1,"serial":[{"uid":"s-1","resourceName":"devices.bytetrade.io/serial-s-1","online":true,"selectable":true}]}`)
	lease := &coordinationv1.Lease{
		ObjectMeta: metav1.ObjectMeta{Name: deviceLeasePrefix + "node-a", Namespace: deviceLeaseNamespace},
		Spec: coordinationv1.LeaseSpec{
			RenewTime:            &metav1.MicroTime{Time: now},
			LeaseDurationSeconds: ptr.To(int32(30)),
		},
	}
	published := 0
	r := &DeviceInventoryController{
		Client: fake.NewClientBuilder().WithScheme(scheme).WithObjects(node, lease).Build(),
		Now:    func() time.Time { return now },
		Publish: func(string, string, interface{}) {
			published++
		},
	}
	result, err := r.Reconcile(context.Background(), ctrl.Request{NamespacedName: types.NamespacedName{Name: "node-a"}})
	if err != nil {
		t.Fatal(err)
	}
	if published != 0 {
		t.Fatalf("initial baseline published %d events", published)
	}
	if result.RequeueAfter != 30*time.Second {
		t.Fatalf("requeueAfter = %s", result.RequeueAfter)
	}
}

func TestEnrichDeviceBindingsIncludesApplicationsWorkloadsAndContainers(t *testing.T) {
	scheme := runtime.NewScheme()
	if err := appv1alpha1.AddToScheme(scheme); err != nil {
		t.Fatal(err)
	}
	ref := "devices.bytetrade.io/bluetooth-p-1"
	appB, managerB := boundApplicationForTest(t, "b-app", "b-ns", "bob", "worker", "device.bluetooth", ref, []string{"sidecar", "app", "app"})
	appA, managerA := boundApplicationForTest(t, "a-app", "a-ns", "alice", "server", "device.bluetooth", ref, []string{"server"})
	// A duplicate attachment must not duplicate a binding.
	appA.Spec.Attachments = append(appA.Spec.Attachments, appA.Spec.Attachments[0])
	r := &DeviceInventoryController{Client: fake.NewClientBuilder().WithScheme(scheme).WithObjects(appB, managerB, appA, managerA).Build()}
	changes := []DeviceInventoryChange{{
		Kind: "device", Action: "removed", Type: "device.bluetooth", ResourceName: ref,
		Previous: &DeviceState{Online: true, Selectable: true},
	}}
	if err := r.enrichDeviceBindings(context.Background(), "node-a", changes); err != nil {
		t.Fatal(err)
	}
	bindings := changes[0].Bindings
	if len(bindings) != 2 {
		t.Fatalf("bindings = %#v", bindings)
	}
	if bindings[0].AppName != "a-app" || bindings[0].Workload != "server" || len(bindings[0].Containers) != 1 || bindings[0].Containers[0] != "server" {
		t.Fatalf("first binding = %#v", bindings[0])
	}
	if bindings[1].AppName != "b-app" || len(bindings[1].Containers) != 2 || bindings[1].Containers[0] != "app" || bindings[1].Containers[1] != "sidecar" {
		t.Fatalf("second binding = %#v", bindings[1])
	}
}

func TestEnrichDeviceBindingsKeepsBindingWhenConfigUnavailable(t *testing.T) {
	scheme := runtime.NewScheme()
	if err := appv1alpha1.AddToScheme(scheme); err != nil {
		t.Fatal(err)
	}
	ref := "devices.bytetrade.io/serial-s-1"
	application := &appv1alpha1.Application{
		ObjectMeta: metav1.ObjectMeta{Name: "homeassistant-olaresid-homeassistant"},
		Spec: appv1alpha1.ApplicationSpec{
			Name: "homeassistant", Namespace: "homeassistant-olaresid", Owner: "olaresid",
			Attachments: []appv1alpha1.Attachment{{Workload: "homeassistant", Type: "device.serial", Ref: ref, Node: "node-a"}},
		},
	}
	r := &DeviceInventoryController{Client: fake.NewClientBuilder().WithScheme(scheme).WithObjects(application).Build()}
	changes := []DeviceInventoryChange{{Kind: "device", Action: "removed", Type: "device.serial", ResourceName: ref}}
	if err := r.enrichDeviceBindings(context.Background(), "node-a", changes); err != nil {
		t.Fatal(err)
	}
	if len(changes[0].Bindings) != 1 || changes[0].Bindings[0].Containers == nil || len(changes[0].Bindings[0].Containers) != 0 {
		t.Fatalf("binding without config = %#v", changes[0].Bindings)
	}
}

func TestEnrichDeviceBindingsForAddedAndOnlineTransitions(t *testing.T) {
	scheme := runtime.NewScheme()
	if err := appv1alpha1.AddToScheme(scheme); err != nil {
		t.Fatal(err)
	}
	ref := "devices.bytetrade.io/serial-s-1"
	application, manager := boundApplicationForTest(t, "homeassistant", "homeassistant-alice", "alice", "homeassistant", "device.serial", ref, []string{"homeassistant"})
	r := &DeviceInventoryController{Client: fake.NewClientBuilder().WithScheme(scheme).WithObjects(application, manager).Build()}
	tests := []struct {
		name        string
		change      DeviceInventoryChange
		wantBinding bool
	}{
		{
			name: "added",
			change: DeviceInventoryChange{Kind: "device", Action: "added", Type: "device.serial", ResourceName: ref,
				Current: &DeviceState{Online: true}},
			wantBinding: true,
		},
		{
			name: "updated offline to online",
			change: DeviceInventoryChange{Kind: "device", Action: "updated", Type: "device.serial", ResourceName: ref,
				Previous: &DeviceState{Online: false}, Current: &DeviceState{Online: true}},
			wantBinding: true,
		},
		{
			name: "updated online to offline",
			change: DeviceInventoryChange{Kind: "device", Action: "updated", Type: "device.serial", ResourceName: ref,
				Previous: &DeviceState{Online: true}, Current: &DeviceState{Online: false}},
			wantBinding: true,
		},
		{
			name: "updated without online transition",
			change: DeviceInventoryChange{Kind: "device", Action: "updated", Type: "device.serial", ResourceName: ref,
				Previous: &DeviceState{Online: true, DisplayName: "old"}, Current: &DeviceState{Online: true, DisplayName: "new"}},
			wantBinding: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			changes := []DeviceInventoryChange{tt.change}
			if err := r.enrichDeviceBindings(context.Background(), "node-a", changes); err != nil {
				t.Fatal(err)
			}
			if got := len(changes[0].Bindings) > 0; got != tt.wantBinding {
				t.Fatalf("bindings = %#v, wantBinding=%t", changes[0].Bindings, tt.wantBinding)
			}
		})
	}
}

func TestPublishDeviceInventoryEventFansOutToAdminsAndBoundOwners(t *testing.T) {
	scheme := runtime.NewScheme()
	if err := appv1alpha1.AddToScheme(scheme); err != nil {
		t.Fatal(err)
	}
	if err := iamv1alpha2.AddToScheme(scheme); err != nil {
		t.Fatal(err)
	}
	ref := "devices.bytetrade.io/serial-s-1"
	objects := []client.Object{
		userForDeviceEventTest("alice", "owner"),
		userForDeviceEventTest("carol", "admin"),
		userForDeviceEventTest("bob", "normal"),
		userForDeviceEventTest("dave", "normal"),
		&appv1alpha1.Application{
			ObjectMeta: metav1.ObjectMeta{Name: "bound-app"},
			Spec: appv1alpha1.ApplicationSpec{
				Name: "homeassistant", Namespace: "homeassistant-bob", Owner: "bob",
				Attachments: []appv1alpha1.Attachment{{Workload: "homeassistant", Type: "device.serial", Ref: ref, Node: "node-a"}},
			},
		},
	}
	type publishedEvent struct {
		subject string
		msgID   string
		event   DeviceInventoryEvent
	}
	published := make([]publishedEvent, 0)
	r := &DeviceInventoryController{
		Client: fake.NewClientBuilder().WithScheme(scheme).WithObjects(objects...).Build(),
		Publish: func(subject, msgID string, data interface{}) {
			published = append(published, publishedEvent{subject: subject, msgID: msgID, event: data.(DeviceInventoryEvent)})
		},
	}
	event := DeviceInventoryEvent{
		SchemaVersion: 1, EventID: "device-inventory:node-a:base", Node: "node-a",
		Changes: []DeviceInventoryChange{{Kind: "device", Action: "added", Type: "device.serial", ResourceName: ref}},
	}
	if err := r.publishToRecipients(context.Background(), event); err != nil {
		t.Fatal(err)
	}
	wantRecipients := []string{"alice", "bob", "carol"}
	if len(published) != len(wantRecipients) {
		t.Fatalf("published = %#v", published)
	}
	for i, recipient := range wantRecipients {
		wantSubject := "os.device.inventory." + recipient
		wantID := event.EventID + ":" + recipient
		if published[i].subject != wantSubject || published[i].msgID != wantID || published[i].event.EventID != wantID {
			t.Fatalf("published[%d] = %#v, want subject=%q id=%q", i, published[i], wantSubject, wantID)
		}
	}
}

func TestPublisherEventIncludesOwnersOfAllAttachmentsOnNode(t *testing.T) {
	scheme := runtime.NewScheme()
	if err := appv1alpha1.AddToScheme(scheme); err != nil {
		t.Fatal(err)
	}
	if err := iamv1alpha2.AddToScheme(scheme); err != nil {
		t.Fatal(err)
	}
	application := &appv1alpha1.Application{
		ObjectMeta: metav1.ObjectMeta{Name: "bound-app"},
		Spec: appv1alpha1.ApplicationSpec{
			Name: "homeassistant", Namespace: "homeassistant-bob", Owner: "bob",
			Attachments: []appv1alpha1.Attachment{
				{Workload: "homeassistant", Type: "device.serial", Ref: "devices.bytetrade.io/serial-s-1", Node: "node-a"},
			},
		},
	}
	r := &DeviceInventoryController{Client: fake.NewClientBuilder().WithScheme(scheme).WithObjects(application).Build()}
	recipients, err := r.eventRecipients(context.Background(), "node-a", []DeviceInventoryChange{{Kind: "publisher", Action: "offline"}})
	if err != nil {
		t.Fatal(err)
	}
	if len(recipients) != 1 || recipients[0].Username != "bob" || !recipients[0].PublisherVisible {
		t.Fatalf("recipients = %#v", recipients)
	}
}

func TestDeviceInventoryEventIsFilteredForNormalRecipients(t *testing.T) {
	aliceBinding := DeviceBinding{AppName: "alice-app", AppOwner: "alice", Workload: "server", Containers: []string{"server"}}
	bobBinding := DeviceBinding{AppName: "bob-app", AppOwner: "bob", Workload: "worker", Containers: []string{"worker"}}
	aliceRef := "devices.bytetrade.io/serial-s-alice"
	bobRef := "devices.bytetrade.io/serial-s-bob"
	event := DeviceInventoryEvent{Changes: []DeviceInventoryChange{
		{Kind: "device", Action: "added", Type: "device.serial", ResourceName: aliceRef, Bindings: []DeviceBinding{aliceBinding}},
		{Kind: "device", Action: "removed", Type: "device.serial", ResourceName: bobRef, Bindings: []DeviceBinding{aliceBinding, bobBinding}},
	}}
	recipient := deviceEventRecipient{
		Username: "bob",
		DeviceKeys: map[string]struct{}{
			changeBindingKey("device.serial", bobRef): {},
		},
	}
	filtered := eventForRecipient(event, recipient)
	if len(filtered.Changes) != 1 || filtered.Changes[0].ResourceName != bobRef {
		t.Fatalf("changes = %#v", filtered.Changes)
	}
	if len(filtered.Changes[0].Bindings) != 1 || filtered.Changes[0].Bindings[0].AppOwner != "bob" {
		t.Fatalf("bindings = %#v", filtered.Changes[0].Bindings)
	}

	privileged := eventForRecipient(event, deviceEventRecipient{Username: "admin", Privileged: true})
	if len(privileged.Changes) != 2 || len(privileged.Changes[1].Bindings) != 2 {
		t.Fatalf("privileged event = %#v", privileged)
	}
}

func userForDeviceEventTest(name, role string) *iamv1alpha2.User {
	return &iamv1alpha2.User{ObjectMeta: metav1.ObjectMeta{
		Name: name, Annotations: map[string]string{users.UserAnnotationOwnerRole: role},
	}}
}

func nodeWithInventory(name, inventory string) *corev1.Node {
	return &corev1.Node{ObjectMeta: metav1.ObjectMeta{Name: name, Annotations: map[string]string{"bytetrade.io/devices": inventory}}}
}

func deviceForTest(typeName, uid, path string) deviceSnapshot {
	resourceType := typeName[len("device."):]
	state := DeviceState{Online: true, Selectable: true}
	if path != "" {
		state.KernelPaths = []string{path}
	}
	return deviceSnapshot{Type: typeName, UID: uid, ResourceName: "devices.bytetrade.io/" + resourceType + "-" + uid, State: state}
}

func snapshotForTest(online bool, devices ...deviceSnapshot) nodeDeviceSnapshot {
	result := nodeDeviceSnapshot{NodeName: "node-a", PublisherOnline: online, Devices: make(map[string]deviceSnapshot)}
	for _, device := range devices {
		result.Devices[device.ResourceName] = device
	}
	return result
}

func boundApplicationForTest(t *testing.T, name, namespace, owner, workload, deviceType, ref string, containers []string) (*appv1alpha1.Application, *appv1alpha1.ApplicationManager) {
	t.Helper()
	objectName := name + "-object"
	application := &appv1alpha1.Application{
		ObjectMeta: metav1.ObjectMeta{Name: objectName},
		Spec: appv1alpha1.ApplicationSpec{
			Name: name, Namespace: namespace, Owner: owner,
			Attachments: []appv1alpha1.Attachment{{Workload: workload, Type: deviceType, Ref: ref, Node: "node-a"}},
		},
	}
	cfg := appcfg.ApplicationConfig{WorkloadOptions: appcfg.WorkloadOptions{
		workload: {Allow: []appcfg.WorkloadCapability{{Type: deviceType, Containers: containers}}},
	}}
	data, err := json.Marshal(cfg)
	if err != nil {
		t.Fatal(err)
	}
	manager := &appv1alpha1.ApplicationManager{
		ObjectMeta: metav1.ObjectMeta{Name: objectName},
		Spec: appv1alpha1.ApplicationManagerSpec{
			AppName: name, AppNamespace: namespace, AppOwner: owner, Config: string(data),
		},
	}
	return application, manager
}
