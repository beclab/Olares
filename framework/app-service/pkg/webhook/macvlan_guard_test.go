package webhook

import (
	"strings"
	"testing"

	"github.com/beclab/Olares/framework/app-service/pkg/constants"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

func guardPod(labels, annotations map[string]string) *corev1.Pod {
	return &corev1.Pod{
		ObjectMeta: metav1.ObjectMeta{
			Name:        "jellyfin-pod",
			Namespace:   "app-space",
			Labels:      labels,
			Annotations: annotations,
		},
		Spec: corev1.PodSpec{Containers: []corev1.Container{{Name: "jellyfin"}}},
	}
}

func authorizedLabels() map[string]string {
	return map[string]string{
		constants.ApplicationNameLabel:        "jellyfin",
		constants.ApplicationMacvlanInitLabel: "true",
	}
}

// admitAuthorizedPod runs the mutating step so the ledger holds a bound MAC and
// the pod carries the platform selection, exactly as after real admission.
func admitAuthorizedPod(t *testing.T, wh *Webhook, pod *corev1.Pod) {
	t.Helper()
	req := macvlanBypassAdmissionRequest(t, pod)
	if _, err := wh.CreateMacvlanInitPatch(req, pod); err != nil {
		t.Fatalf("CreateMacvlanInitPatch: %v", err)
	}
}

func TestValidateMacvlanAnnotationAcceptsPlatformSelection(t *testing.T) {
	wh := testMacvlanWebhook()
	pod := guardPod(authorizedLabels(), nil)
	admitAuthorizedPod(t, wh, pod)

	if err := wh.ValidateMacvlanAnnotation(t.Context(), pod, pod.Namespace, false); err != nil {
		t.Fatalf("expected platform selection to be accepted: %v", err)
	}
}

func TestValidateMacvlanAnnotationRejectsTamperedSelection(t *testing.T) {
	wh := testMacvlanWebhook()
	pod := guardPod(authorizedLabels(), nil)
	admitAuthorizedPod(t, wh, pod)

	cases := map[string]string{
		"different mac": strings.Replace(pod.Annotations[multusNetworksAnnotation], `"mac":"02:`, `"mac":"02:ff:`, 1),
		"extra space":   strings.Replace(pod.Annotations[multusNetworksAnnotation], `,`, `, `, 1),
		"extra network": strings.TrimSuffix(pod.Annotations[multusNetworksAnnotation], `]`) + `,{"name":"other-net"}]`,
		"short form":    "kube-system/underlay-macvlan",
	}
	for name, value := range cases {
		tampered := pod.DeepCopy()
		tampered.Annotations[multusNetworksAnnotation] = value
		if err := wh.ValidateMacvlanAnnotation(t.Context(), tampered, pod.Namespace, false); err == nil {
			t.Errorf("%s: expected rejection for %q", name, value)
		}
	}
}

func TestValidateMacvlanAnnotationRejectsMissingSelectionWhenAuthorized(t *testing.T) {
	wh := testMacvlanWebhook()
	pod := guardPod(authorizedLabels(), nil)
	admitAuthorizedPod(t, wh, pod)
	delete(pod.Annotations, multusNetworksAnnotation)

	if err := wh.ValidateMacvlanAnnotation(t.Context(), pod, pod.Namespace, false); err == nil {
		t.Fatal("expected rejection when the authorized pod lost its platform selection")
	}
}

func TestValidateMacvlanAnnotationRejectsAuthorizedPodWithoutLedger(t *testing.T) {
	wh := testMacvlanWebhook()
	pod := guardPod(authorizedLabels(), map[string]string{
		multusNetworksAnnotation: platformMacvlanSelection("02:00:00:00:00:01"),
	})

	if err := wh.ValidateMacvlanAnnotation(t.Context(), pod, pod.Namespace, false); err == nil {
		t.Fatal("expected rejection when no MAC is bound in the ledger")
	}
}

func TestValidateMacvlanAnnotationAllowsDryRunWithoutLedger(t *testing.T) {
	wh := testMacvlanWebhook()
	pod := guardPod(authorizedLabels(), map[string]string{
		multusNetworksAnnotation: platformMacvlanSelection("02:00:00:00:00:01"),
	})

	if err := wh.ValidateMacvlanAnnotation(t.Context(), pod, pod.Namespace, true); err != nil {
		t.Fatalf("dry-run must not depend on ledger state: %v", err)
	}
}

func TestValidateMacvlanAnnotationRejectsUnauthorizedSelections(t *testing.T) {
	wh := testMacvlanWebhook()
	cases := map[string]*corev1.Pod{
		"unlabeled short form": guardPod(nil, map[string]string{
			multusNetworksAnnotation: "kube-system/underlay-macvlan",
		}),
		"unlabeled json": guardPod(nil, map[string]string{
			multusNetworksAnnotation: `[{"name":"underlay-macvlan","namespace":"kube-system"}]`,
		}),
		"unlabeled capitalized keys": guardPod(nil, map[string]string{
			multusNetworksAnnotation: `[{"Name":"underlay-macvlan","Namespace":"kube-system"}]`,
		}),
		"label without application": guardPod(map[string]string{
			constants.ApplicationNameLabel:        "unknown-app",
			constants.ApplicationMacvlanInitLabel: "true",
		}, map[string]string{
			multusNetworksAnnotation: platformMacvlanSelection("02:00:00:00:00:01"),
		}),
		"default network override": guardPod(authorizedLabels(), map[string]string{
			multusDefaultNetworkAnnotation: "underlay-macvlan",
		}),
	}
	for name, pod := range cases {
		if err := wh.ValidateMacvlanAnnotation(t.Context(), pod, pod.Namespace, false); err == nil {
			t.Errorf("%s: expected rejection", name)
		}
	}
}

func TestValidateMacvlanAnnotationAllowsPodsWithoutSelection(t *testing.T) {
	wh := testMacvlanWebhook()
	for name, pod := range map[string]*corev1.Pod{
		"plain pod":            guardPod(nil, nil),
		"unrelated annotation": guardPod(nil, map[string]string{"foo": "bar"}),
	} {
		if err := wh.ValidateMacvlanAnnotation(t.Context(), pod, pod.Namespace, false); err != nil {
			t.Errorf("%s: unexpected rejection: %v", name, err)
		}
	}
}

func TestMacvlanSelectionAnnotationsChanged(t *testing.T) {
	base := guardPod(nil, map[string]string{multusNetworksAnnotation: "kube-system/underlay-macvlan", "foo": "bar"})

	unrelated := base.DeepCopy()
	unrelated.Annotations["foo"] = "baz"
	unrelated.Labels = map[string]string{"new": "label"}
	if MacvlanSelectionAnnotationsChanged(base, unrelated) {
		t.Fatal("unrelated annotation and label changes must not count as a selection change")
	}

	edited := base.DeepCopy()
	edited.Annotations[multusNetworksAnnotation] = `[{"name":"underlay-macvlan"}]`
	if !MacvlanSelectionAnnotationsChanged(base, edited) {
		t.Fatal("editing the networks annotation must count as a change")
	}

	added := base.DeepCopy()
	added.Annotations[multusDefaultNetworkAnnotation] = "underlay-macvlan"
	if !MacvlanSelectionAnnotationsChanged(base, added) {
		t.Fatal("adding the default-network annotation must count as a change")
	}

	removed := base.DeepCopy()
	delete(removed.Annotations, multusNetworksAnnotation)
	if !MacvlanSelectionAnnotationsChanged(base, removed) {
		t.Fatal("removing the networks annotation must count as a change")
	}
}

func TestHasMacvlanSelectionAnnotations(t *testing.T) {
	if HasMacvlanSelectionAnnotations(guardPod(nil, nil)) {
		t.Fatal("pod without annotations must not report a selection")
	}
	if HasMacvlanSelectionAnnotations(guardPod(nil, map[string]string{"foo": "bar"})) {
		t.Fatal("unrelated annotation must not report a selection")
	}
	if !HasMacvlanSelectionAnnotations(guardPod(nil, map[string]string{multusDefaultNetworkAnnotation: "x"})) {
		t.Fatal("default-network annotation must report a selection")
	}
	if !HasMacvlanSelectionAnnotations(guardPod(nil, map[string]string{multusNetworksAnnotation: "x"})) {
		t.Fatal("networks annotation must report a selection")
	}
}
