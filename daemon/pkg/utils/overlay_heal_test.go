package utils

import (
	"context"
	"errors"
	apierrors "k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/runtime"
	kubefake "k8s.io/client-go/kubernetes/fake"
	ktesting "k8s.io/client-go/testing"
	"testing"

	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

func healPod(phase corev1.PodPhase, labeled bool, status string) *corev1.Pod {
	p := &corev1.Pod{
		ObjectMeta: metav1.ObjectMeta{Name: "p", Namespace: "ns", Labels: map[string]string{}, Annotations: map[string]string{}},
		Status:     corev1.PodStatus{Phase: phase},
	}
	if labeled {
		p.Labels[overlayMacvlanInitKey] = "true"
	}
	if status != "" {
		p.Annotations["k8s.v1.cni.cncf.io/network-status"] = status
	}
	return p
}

const healWithNet1 = `[{"name":"k8s-pod-network","interface":"eth0","default":true},{"name":"kube-system/underlay-macvlan","interface":"net1","ips":["192.168.50.12"]}]`
const healWithoutNet1 = `[{"name":"k8s-pod-network","interface":"eth0","default":true}]`

func TestOverlayPodMissingNet1(t *testing.T) {
	cases := []struct {
		name string
		pod  *corev1.Pod
		want bool
	}{
		{"running with net1", healPod(corev1.PodRunning, true, healWithNet1), false},
		{"running without net1", healPod(corev1.PodRunning, true, healWithoutNet1), true},
		{"running with no status yet", healPod(corev1.PodRunning, true, ""), true},
		{"pending", healPod(corev1.PodPending, true, healWithoutNet1), false},
		{"not an overlay pod", healPod(corev1.PodRunning, false, healWithoutNet1), false},
	}
	for _, c := range cases {
		if got := overlayPodMissingNet1(c.pod); got != c.want {
			t.Errorf("%s: overlayPodMissingNet1 = %v, want %v", c.name, got, c.want)
		}
	}
}

func TestOverlayHealRetriesFailedEviction(t *testing.T) {
	pod := healPod(corev1.PodRunning, true, healWithoutNet1)
	pod.UID = "retry-pod"
	yes := true
	pod.OwnerReferences = []metav1.OwnerReference{{Kind: "StatefulSet", Name: "app", UID: "app", Controller: &yes}}
	client := kubefake.NewSimpleClientset(pod)
	apps := func(context.Context, string) ([]OverlayGatewaySupportedApp, error) {
		return []OverlayGatewaySupportedApp{{Namespace: "ns", Enabled: true}}, nil
	}
	calls := 0
	client.PrependReactor("create", "pods", func(a ktesting.Action) (bool, runtime.Object, error) {
		calls++
		if calls == 1 {
			return true, nil, apierrors.NewTooManyRequests("budget", 0)
		}
		return true, nil, nil
	})
	defer healedOverlayPods.Delete(pod.UID)
	if n, err := healOverlayPods(t.Context(), client, apps); n != 0 || err == nil {
		t.Fatalf("failure swallowed: %d %v", n, err)
	}
	if n, err := healOverlayPods(t.Context(), client, apps); n != 1 || err != nil {
		t.Fatalf("retry failed: %d %v", n, err)
	}
	if n, err := healOverlayPods(t.Context(), client, apps); n != 0 || err != nil || calls != 2 {
		t.Fatalf("duplicate eviction: %d %v calls=%d", n, err, calls)
	}
}

func TestOverlayHealDefersWhileUpgradeOwnsRecovery(t *testing.T) {
	client := kubefake.NewSimpleClientset(&corev1.ConfigMap{ObjectMeta: metav1.ObjectMeta{Name: "olares-fixed-mac-upgrade", Namespace: "kube-system"}, Data: map[string]string{"phase": "recovering"}})
	_, err := healOverlayPods(t.Context(), client, func(context.Context, string) ([]OverlayGatewaySupportedApp, error) {
		t.Fatal("queried applications while paused")
		return nil, nil
	})
	if !errors.Is(err, errOverlayHealingDeferred) {
		t.Fatalf("deferred recovery reported complete: %v", err)
	}
}
