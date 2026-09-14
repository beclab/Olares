package apiserver

import (
	"testing"

	"github.com/beclab/Olares/framework/app-service/pkg/appcfg"
	appv1alpha1 "github.com/beclab/api/api/app.bytetrade.io/v1alpha1"
)

func int32Ptr(value int32) *int32 { return &value }

func attachmentTestConfig() *appcfg.ApplicationConfig {
	return &appcfg.ApplicationConfig{
		AppName: "demo",
		WorkloadOptions: appcfg.WorkloadOptions{
			"server": {
				Replicas: int32Ptr(1),
				Allow: []appcfg.WorkloadCapability{
					{Type: "folder", Containers: []string{"server"}},
					{Type: "device.video", Containers: []string{"server"}},
				},
			},
		},
	}
}

func TestAttachmentCapabilitiesResponse(t *testing.T) {
	zero := int32(0)
	cfg := attachmentTestConfig()
	cfg.WorkloadOptions["background"] = appcfg.WorkloadOption{Replicas: &zero}

	got := attachmentCapabilitiesResponse(cfg)
	if len(got.Workloads) != 2 {
		t.Fatalf("workloads=%#v", got.Workloads)
	}
	if got.Workloads[0].Name != "background" || got.Workloads[0].Replicas == nil || *got.Workloads[0].Replicas != 0 {
		t.Fatalf("first workload=%#v", got.Workloads[0])
	}
	server := got.Workloads[1]
	if server.Name != "server" || server.Replicas == nil || *server.Replicas != 1 {
		t.Fatalf("server workload=%#v", server)
	}
	if len(server.Capabilities) != 2 || server.Capabilities[0].Type != "folder" || server.Capabilities[1].Type != "device.video" {
		t.Fatalf("server capabilities=%#v", server.Capabilities)
	}
	if len(server.Capabilities[0].Containers) != 1 || server.Capabilities[0].Containers[0] != "server" {
		t.Fatalf("folder containers=%#v", server.Capabilities[0].Containers)
	}
}

func TestValidateAttachments(t *testing.T) {
	cfg := attachmentTestConfig()
	inventory := []DeviceInventoryItem{{
		Node: "node-a", Type: "device.video", ResourceName: "devices.bytetrade.io/video-camera", Online: true, Selectable: true,
	}}
	next := []appv1alpha1.Attachment{
		{Workload: "server", Type: "folder", Ref: "drive/Home/photos", Name: "photos"},
		{Workload: "server", Type: "device.video", Ref: "devices.bytetrade.io/video-camera", Node: "node-a"},
	}
	if err := validateAttachments(next, nil, cfg, inventory); err != nil {
		t.Fatalf("valid attachments rejected: %v", err)
	}
}

func TestValidateAttachmentsAllowsExistingOfflineDevice(t *testing.T) {
	cfg := attachmentTestConfig()
	attachment := appv1alpha1.Attachment{Workload: "server", Type: "device.video", Ref: "devices.bytetrade.io/video-camera", Node: "node-a"}
	if err := validateAttachments([]appv1alpha1.Attachment{attachment}, []appv1alpha1.Attachment{attachment}, cfg, nil); err != nil {
		t.Fatalf("existing offline attachment rejected: %v", err)
	}
}

func TestValidateAttachmentsRejectsUnsafeFolderAndNewOfflineDevice(t *testing.T) {
	cfg := attachmentTestConfig()
	tests := []appv1alpha1.Attachment{
		{Workload: "server", Type: "folder", Ref: "drive/Home/%2e%2e/secrets", Name: "secrets"},
		{Workload: "server", Type: "device.video", Ref: "devices.bytetrade.io/video-missing", Node: "node-a"},
	}
	for _, attachment := range tests {
		if err := validateAttachments([]appv1alpha1.Attachment{attachment}, nil, cfg, nil); err == nil {
			t.Fatalf("invalid attachment accepted: %#v", attachment)
		}
	}
}

func TestValidateAttachmentsExternalFolderRequiresMatchingNode(t *testing.T) {
	cfg := attachmentTestConfig()
	tests := []appv1alpha1.Attachment{
		{Workload: "server", Type: "folder", Ref: "external/node-a/usb/photos", Name: "photos"},
		{Workload: "server", Type: "folder", Ref: "external/node-a/usb/photos", Name: "photos", Node: "node-b"},
	}
	for _, attachment := range tests {
		if err := validateAttachments([]appv1alpha1.Attachment{attachment}, nil, cfg, nil); err == nil {
			t.Fatalf("external folder with invalid node accepted: %#v", attachment)
		}
	}
	valid := appv1alpha1.Attachment{Workload: "server", Type: "folder", Ref: "external/node-a/usb/photos", Name: "photos", Node: "node-a"}
	if err := validateAttachments([]appv1alpha1.Attachment{valid}, nil, cfg, nil); err != nil {
		t.Fatalf("valid external folder rejected: %v", err)
	}
}

func TestValidateAttachmentsRejectsFolderDeviceNodeConflict(t *testing.T) {
	cfg := attachmentTestConfig()
	inventory := []DeviceInventoryItem{{
		Node: "node-b", Type: "device.video", ResourceName: "devices.bytetrade.io/video-camera", Online: true, Selectable: true,
	}}
	attachments := []appv1alpha1.Attachment{
		{Workload: "server", Type: "folder", Ref: "external/node-a/usb/photos", Name: "photos", Node: "node-a"},
		{Workload: "server", Type: "device.video", Ref: "devices.bytetrade.io/video-camera", Node: "node-b"},
	}
	if err := validateAttachments(attachments, nil, cfg, inventory); err == nil {
		t.Fatal("folder/device node conflict was accepted")
	}
}

func TestAffectedWorkloadsSortedAndDeduplicated(t *testing.T) {
	old := []appv1alpha1.Attachment{{Workload: "z"}, {Workload: "a"}}
	next := []appv1alpha1.Attachment{{Workload: "a"}, {Workload: "m"}}
	got := affectedWorkloads(old, next)
	want := []string{"a", "m", "z"}
	for i := range want {
		if got[i] != want[i] {
			t.Fatalf("affectedWorkloads=%v, want %v", got, want)
		}
	}
}
