package webhook

import (
	"encoding/json"
	"testing"

	"github.com/beclab/Olares/framework/app-service/pkg/constants"
	"github.com/beclab/Olares/framework/app-service/pkg/gateway/meshinagent"
	"github.com/beclab/Olares/framework/app-service/pkg/sandbox/sidecar"
	appv1alpha1 "github.com/beclab/api/api/app.bytetrade.io/v1alpha1"
	appfake "github.com/beclab/api/pkg/generated/clientset/versioned/fake"

	admissionv1 "k8s.io/api/admission/v1"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	dynamicfake "k8s.io/client-go/dynamic/fake"
	k8sfake "k8s.io/client-go/kubernetes/fake"
)

func macvlanBypassAdmissionRequest(t *testing.T, pod *corev1.Pod) *admissionv1.AdmissionRequest {
	t.Helper()
	raw, err := json.Marshal(pod)
	if err != nil {
		t.Fatalf("marshal pod: %v", err)
	}
	return &admissionv1.AdmissionRequest{
		Namespace: pod.Namespace,
		Object:    runtime.RawExtension{Raw: raw},
	}
}

func initNames(pod *corev1.Pod) []string {
	names := make([]string, 0, len(pod.Spec.InitContainers))
	for _, c := range pod.Spec.InitContainers {
		names = append(names, c.Name)
	}
	return names
}

func indexOfInit(pod *corev1.Pod, name string) int {
	for i, c := range pod.Spec.InitContainers {
		if c.Name == name {
			return i
		}
	}
	return -1
}

func testMacvlanWebhook() *Webhook {
	appClient := appfake.NewSimpleClientset(&appv1alpha1.Application{
		ObjectMeta: metav1.ObjectMeta{
			Name: "app-space-jellyfin",
			UID:  "application-uid",
		},
		Spec: appv1alpha1.ApplicationSpec{
			Name:      "jellyfin",
			Namespace: "app-space",
			Settings:  map[string]string{"enableOverlayGateway": "true"},
		},
	})
	return &Webhook{
		kubeClient:       k8sfake.NewSimpleClientset(),
		dynamicClient:    appClient,
		allocationClient: dynamicfake.NewSimpleDynamicClient(runtime.NewScheme()),
	}
}

func TestCreateMacvlanInitPatchPutsBypassAfterReplyInit(t *testing.T) {
	pod := &corev1.Pod{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "jellyfin-0",
			Namespace: "app-space",
			Labels: map[string]string{
				constants.ApplicationNameLabel: "jellyfin",
			},
		},
		Spec: corev1.PodSpec{
			InitContainers: []corev1.Container{{Name: "linkerd-init"}},
			Containers:     []corev1.Container{{Name: "jellyfin"}},
		},
	}
	req := macvlanBypassAdmissionRequest(t, pod)

	if _, err := testMacvlanWebhook().CreateMacvlanInitPatch(req, pod); err != nil {
		t.Fatalf("CreateMacvlanInitPatch: %v", err)
	}

	names := initNames(pod)
	if last := names[len(names)-1]; last != sidecar.MacvlanBypassInitContainerName {
		t.Fatalf("bypass must be the last init, got order %v", names)
	}
	if indexOfInit(pod, MacvlanInitContainerName) < 0 {
		t.Fatalf("macvlan reply init missing, order %v", names)
	}
	if indexOfInit(pod, MacvlanInitContainerName) > indexOfInit(pod, sidecar.MacvlanBypassInitContainerName) {
		t.Fatalf("bypass must run after the reply-route init, order %v", names)
	}
}

func TestCreateMacvlanInitPatchIsIdempotent(t *testing.T) {
	wh := testMacvlanWebhook()
	pod := &corev1.Pod{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "jellyfin-0",
			Namespace: "app-space",
			Labels: map[string]string{
				constants.ApplicationNameLabel: "jellyfin",
			},
		},
		Spec: corev1.PodSpec{Containers: []corev1.Container{{Name: "jellyfin"}}},
	}
	req := macvlanBypassAdmissionRequest(t, pod)

	if _, err := wh.CreateMacvlanInitPatch(req, pod); err != nil {
		t.Fatalf("first patch: %v", err)
	}
	if _, err := wh.CreateMacvlanInitPatch(req, pod); err != nil {
		t.Fatalf("second patch: %v", err)
	}

	names := initNames(pod)
	if len(names) != 2 {
		t.Fatalf("expected reply init + bypass only, got %v", names)
	}
	if names[1] != sidecar.MacvlanBypassInitContainerName {
		t.Fatalf("bypass must stay last, got %v", names)
	}
}

// The bypass only works when it writes iptables after every other writer in the
// pod; mesh-in inserts its own REDIRECT at the head of nat OUTPUT.
func TestMacvlanBypassRunsAfterMeshInIptablesInit(t *testing.T) {
	pod := &corev1.Pod{
		Spec: corev1.PodSpec{
			InitContainers: []corev1.Container{
				{Name: "linkerd-init"},
				{Name: MacvlanInitContainerName},
			},
		},
	}
	sidecar.EnsureMacvlanBypassLast(pod)
	pod.Spec.InitContainers = append(pod.Spec.InitContainers, meshinagent.InitContainerSpec())

	sidecar.EnsureMacvlanBypassLast(pod)

	names := initNames(pod)
	if indexOfInit(pod, meshinagent.InitContainerName) > indexOfInit(pod, sidecar.MacvlanBypassInitContainerName) {
		t.Fatalf("bypass must run after mesh-in iptables init, order %v", names)
	}
	if names[len(names)-1] != sidecar.MacvlanBypassInitContainerName {
		t.Fatalf("bypass must be last, order %v", names)
	}
}
