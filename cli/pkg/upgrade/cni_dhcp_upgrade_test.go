package upgrade

import (
	"context"
	"fmt"
	"testing"

	"github.com/beclab/Olares/cli/pkg/core/task"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	k8sruntime "k8s.io/apimachinery/pkg/runtime"
	"k8s.io/client-go/kubernetes/fake"
)

func overlayPod(ns, name, selection string) *corev1.Pod {
	p := &corev1.Pod{ObjectMeta: metav1.ObjectMeta{Name: name, Namespace: ns}}
	if selection != "" {
		p.Annotations = map[string]string{multusNetworksAnnotation: selection}
	}
	return p
}

func TestRecreateOverlayGatewayPodsSelectsOnlyUnderlayAttachments(t *testing.T) {
	now := metav1.Now()
	terminating := overlayPod("homeassistant-vodevall", "ha-old", "kube-system/underlay-macvlan")
	terminating.DeletionTimestamp = &now
	terminating.Finalizers = []string{"test/keep"}

	kube := fake.NewSimpleClientset(
		// JSON selection as injected by app-service 0.6.38
		overlayPod("jellyfin-brucedai", "jf-1", `[{"mac":"02:af:0a:43:df:84","name":"underlay-macvlan","namespace":"kube-system"}]`),
		// short-form selection as written by the pre-0.6.38 webhook
		overlayPod("jellyfin-vodevall", "jf-2", "kube-system/underlay-macvlan"),
		// unrelated secondary network must be left alone
		overlayPod("user-space-a", "other", "kube-system/some-other-net"),
		// no Multus annotation at all
		overlayPod("user-space-a", "plain", ""),
		terminating,
	)

	n, err := recreateOverlayGatewayPods(context.Background(), kube)
	if err != nil {
		t.Fatal(err)
	}
	if n != 2 {
		t.Fatalf("deleted=%d want 2", n)
	}
	for _, tc := range []struct {
		ns, name string
		want     bool
	}{
		{"jellyfin-brucedai", "jf-1", false},
		{"jellyfin-vodevall", "jf-2", false},
		{"user-space-a", "other", true},
		{"user-space-a", "plain", true},
		{"homeassistant-vodevall", "ha-old", true},
	} {
		_, err := kube.CoreV1().Pods(tc.ns).Get(context.Background(), tc.name, metav1.GetOptions{})
		if exists := err == nil; exists != tc.want {
			t.Fatalf("%s/%s exists=%v want %v", tc.ns, tc.name, exists, tc.want)
		}
	}
}

func TestRecreateOverlayGatewayPodsRespectsBatchLimit(t *testing.T) {
	objs := make([]k8sruntime.Object, 0, overlayRecreateBatchLimit+5)
	for i := 0; i < overlayRecreateBatchLimit+5; i++ {
		objs = append(objs, overlayPod("ns", fmt.Sprintf("p%d", i), "kube-system/underlay-macvlan"))
	}
	kube := fake.NewSimpleClientset(objs...)
	n, err := recreateOverlayGatewayPods(context.Background(), kube)
	if err != nil {
		t.Fatal(err)
	}
	if n != overlayRecreateBatchLimit {
		t.Fatalf("deleted=%d want batch limit %d", n, overlayRecreateBatchLimit)
	}
}

func TestUpgrader20260905PhaseOrder(t *testing.T) {
	u := upgrader_1_12_7_20260905{}
	if u.Version().String() != "1.12.7-20260905" {
		t.Fatalf("version = %s", u.Version())
	}
	if !u.AddedBreakingChange() {
		t.Fatal("daily upgrader must be registered as breaking")
	}
	first := func(ts []task.Interface) string {
		if len(ts) == 0 {
			return ""
		}
		return ts[0].GetName()
	}
	if got := first(u.PrepareForUpgrade()); got != "UpgradeCniDhcpBinary" {
		t.Fatalf("PrepareForUpgrade must start with the binary swap, got %q", got)
	}
	if got := first(u.PostUpgrade()); got != "RecreateOverlayGatewayPods" {
		t.Fatalf("PostUpgrade must start with the pod recreate, got %q", got)
	}
	found := false
	for _, tk := range u.UpgradeSystemComponents() {
		if tk.GetName() == "GenerateMultusDefine" {
			found = true
		}
	}
	if !found {
		t.Fatal("UpgradeSystemComponents must re-render the underlay-macvlan NAD (upgradeMultus)")
	}
}
