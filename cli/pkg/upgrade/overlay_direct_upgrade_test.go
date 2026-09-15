package upgrade

import (
	"context"
	"strings"
	"testing"
	"time"

	"github.com/beclab/Olares/cli/pkg/core/task"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes/fake"
)

func TestUpgrader20260916TaskOrder(t *testing.T) {
	u := upgrader_1_12_7_20260916{}
	if u.Version().String() != "1.12.7-20260916" {
		t.Fatalf("version = %s", u.Version())
	}
	if !u.AddedBreakingChange() {
		t.Fatal("daily upgrader must be registered as breaking")
	}
	names := func(ts []task.Interface) []string {
		out := make([]string, 0, len(ts))
		for _, tk := range ts {
			out = append(out, tk.GetName())
		}
		return out
	}
	index := func(list []string, name string) int {
		for i, n := range list {
			if n == name {
				return i
			}
		}
		return -1
	}
	sys := names(u.UpgradeSystemComponents())
	order := []string{"MigrateOverlayBridgeToDirect", "ReassertHostsAfterMigration", "EnsureOverlayAltname", "GenerateMultusDefine", "DeployMultusDefine", "WriteOverlayDesiredState", "RecreateOverlayGatewayPodsAfterMigration", "WaitOverlayGatewayPodsNet1", "ClearOverlayMigrationMarker"}
	prev := -1
	for _, name := range order {
		at := index(sys, name)
		if at < 0 {
			t.Fatalf("UpgradeSystemComponents must contain %s, got %v", name, sys)
		}
		if at <= prev {
			t.Fatalf("%s is out of order in %v", name, sys)
		}
		prev = at
	}
	if index(sys, "RecreateOverlayGatewayPods") >= 0 {
		t.Fatal("the best-effort recreate must not run alongside the required one")
	}
	for _, phase := range []struct {
		name  string
		tasks []task.Interface
	}{{"PrepareForUpgrade", u.PrepareForUpgrade()}, {"PostUpgrade", u.PostUpgrade()}} {
		got := names(phase.tasks)
		for _, name := range order {
			if index(got, name) >= 0 {
				t.Fatalf("%s must not carry overlay parent steps, got %v", phase.name, got)
			}
		}
	}
}

func migratedOverlayPod(ns, name string, phase corev1.PodPhase, networkStatus string) *corev1.Pod {
	p := &corev1.Pod{
		ObjectMeta: metav1.ObjectMeta{
			Name:      name,
			Namespace: ns,
			Annotations: map[string]string{
				multusNetworksAnnotation: "kube-system/underlay-macvlan",
			},
			OwnerReferences: []metav1.OwnerReference{{Kind: "ReplicaSet", Name: "rs"}},
		},
		Status: corev1.PodStatus{Phase: phase},
	}
	if networkStatus != "" {
		p.Annotations[multusNetworkStatusAnnotation] = networkStatus
	}
	return p
}

const withNet1 = `[{"name":"k8s-pod-network","interface":"eth0","default":true},{"name":"kube-system/underlay-macvlan","interface":"net1","ips":["192.168.50.12"]}]`
const withoutNet1 = `[{"name":"k8s-pod-network","interface":"eth0","default":true}]`

func TestPodHasNet1(t *testing.T) {
	if !podHasNet1(migratedOverlayPod("a", "p", corev1.PodRunning, withNet1)) {
		t.Fatal("net1 present must be detected")
	}
	if podHasNet1(migratedOverlayPod("a", "p", corev1.PodRunning, withoutNet1)) {
		t.Fatal("eth0-only status must not count as net1")
	}
	if podHasNet1(migratedOverlayPod("a", "p", corev1.PodRunning, "not json")) {
		t.Fatal("unparsable status must not count as net1")
	}
}

func TestWaitOverlayGatewayPodsNet1ReturnsWhenAllReady(t *testing.T) {
	kube := fake.NewSimpleClientset(
		migratedOverlayPod("jellyfin-a", "jellyfin-1", corev1.PodRunning, withNet1),
		migratedOverlayPod("ha-a", "ha-1", corev1.PodRunning, withNet1),
	)
	if err := waitOverlayGatewayPodsNet1(context.Background(), kube); err != nil {
		t.Fatalf("expected immediate success, got %v", err)
	}
}

func TestWaitOverlayGatewayPodsNet1ReportsMissingPods(t *testing.T) {
	kube := fake.NewSimpleClientset(
		migratedOverlayPod("jellyfin-a", "jellyfin-1", corev1.PodRunning, withNet1),
		migratedOverlayPod("ha-a", "ha-1", corev1.PodPending, ""),
	)
	orig := overlayNet1PollWait
	defer func() { overlayNet1PollWait = orig }()
	polls := 0
	overlayNet1PollWait = func(ctx context.Context, d time.Duration) error {
		polls++
		if polls >= 2 {
			return context.DeadlineExceeded
		}
		return nil
	}
	err := waitOverlayGatewayPodsNet1(context.Background(), kube)
	if err == nil || !strings.Contains(err.Error(), "ha-a/ha-1") {
		t.Fatalf("expected the pending pod to be reported, got %v", err)
	}
	if polls != 2 {
		t.Fatalf("expected the loop to poll until the deadline, polled %d", polls)
	}
}
