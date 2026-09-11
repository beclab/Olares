package webhook

import (
	"strings"
	"testing"

	"github.com/beclab/Olares/framework/app-service/pkg/constants"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

func TestStripMacvlanAnnotationsRemovesBothSelectionKeys(t *testing.T) {
	wh := testMacvlanWebhook()
	pod := guardPod(nil, map[string]string{
		multusNetworksAnnotation:       `[{"Name":"underlay-macvlan","mac":"02:de:ad:be:ef:01"}]`,
		multusDefaultNetworkAnnotation: "underlay-macvlan",
		"foo":                          "bar",
	})
	req := macvlanBypassAdmissionRequest(t, pod)

	patch, removed, err := wh.StripMacvlanAnnotations(req, pod)
	if err != nil {
		t.Fatalf("StripMacvlanAnnotations: %v", err)
	}
	if len(removed) != 2 {
		t.Fatalf("removed = %v, want both selection keys", removed)
	}
	if _, ok := pod.Annotations[multusNetworksAnnotation]; ok {
		t.Fatal("networks annotation survived the strip")
	}
	if _, ok := pod.Annotations[multusDefaultNetworkAnnotation]; ok {
		t.Fatal("default-network annotation survived the strip")
	}
	if pod.Annotations["foo"] != "bar" {
		t.Fatal("unrelated annotation must be preserved")
	}
	for _, want := range []string{`"op":"remove"`, `k8s.v1.cni.cncf.io~1networks`, `v1.multus-cni.io~1default-network`} {
		if !strings.Contains(string(patch), want) {
			t.Fatalf("patch %s lacks %s", patch, want)
		}
	}
	if strings.Contains(string(patch), "initContainers") {
		t.Fatalf("strip must not inject containers, got %s", patch)
	}
}

func TestStripMacvlanAnnotationsIsNoopWithoutSelection(t *testing.T) {
	wh := testMacvlanWebhook()
	pod := guardPod(nil, map[string]string{"foo": "bar"})
	req := macvlanBypassAdmissionRequest(t, pod)

	patch, removed, err := wh.StripMacvlanAnnotations(req, pod)
	if err != nil {
		t.Fatalf("StripMacvlanAnnotations: %v", err)
	}
	if patch != nil || removed != nil {
		t.Fatalf("expected no-op, got patch=%s removed=%v", patch, removed)
	}
}

func TestRecordMacvlanEventCreatesNamespacedEvent(t *testing.T) {
	wh := testMacvlanWebhook()
	pod := guardPod(map[string]string{constants.ApplicationNameLabel: "jellyfin"}, nil)
	pod.Name = ""
	pod.GenerateName = "jellyfin-7c9d-"

	wh.RecordMacvlanEvent(t.Context(), pod, pod.Namespace, EventReasonOverlayGatewayDisabled, "removed", false)

	events, err := wh.kubeClient.CoreV1().Events(pod.Namespace).List(t.Context(), metav1.ListOptions{})
	if err != nil {
		t.Fatalf("list events: %v", err)
	}
	if len(events.Items) != 1 {
		t.Fatalf("expected one event, got %d", len(events.Items))
	}
	event := events.Items[0]
	if event.Reason != EventReasonOverlayGatewayDisabled || event.InvolvedObject.Kind != "Pod" || event.InvolvedObject.Name != "jellyfin-7c9d-" {
		t.Fatalf("unexpected event %+v", event)
	}
}

func TestRecordMacvlanEventSkipsDryRun(t *testing.T) {
	wh := testMacvlanWebhook()
	pod := guardPod(nil, nil)

	wh.RecordMacvlanEvent(t.Context(), pod, pod.Namespace, EventReasonDefaultNetworkNotAllowed, "removed", true)

	events, err := wh.kubeClient.CoreV1().Events(pod.Namespace).List(t.Context(), metav1.ListOptions{})
	if err != nil {
		t.Fatalf("list events: %v", err)
	}
	if len(events.Items) != 0 {
		t.Fatalf("dry-run must not record events, got %d", len(events.Items))
	}
}
