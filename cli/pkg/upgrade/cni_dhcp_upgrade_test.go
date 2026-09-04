package upgrade

import (
	"context"
	"errors"
	"fmt"
	"path"
	"strings"
	"testing"
	"time"

	"github.com/beclab/Olares/cli/pkg/core/task"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	k8sruntime "k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/types"
	"k8s.io/client-go/kubernetes/fake"
)

// overlayPod builds a Pod owned by a ReplicaSet, which is how app-service
// deploys Overlay apps, so deleting it is a recreate.
func overlayPod(ns, name, selection string) *corev1.Pod {
	p := &corev1.Pod{ObjectMeta: metav1.ObjectMeta{
		Name:            name,
		Namespace:       ns,
		OwnerReferences: []metav1.OwnerReference{{APIVersion: "apps/v1", Kind: "ReplicaSet", Name: name + "-rs", UID: types.UID("u-" + name)}},
	}}
	if selection != "" {
		p.Annotations = map[string]string{multusNetworksAnnotation: selection}
	}
	return p
}

func TestPodSelectsUnderlay(t *testing.T) {
	cases := []struct {
		name, ns, selection string
		want                bool
	}{
		// JSON array as injected by app-service 0.6.38
		{"json array", "jellyfin-a", `[{"mac":"02:af:0a:43:df:84","name":"underlay-macvlan","namespace":"kube-system"}]`, true},
		{"json object", "jellyfin-a", `{"name":"underlay-macvlan","namespace":"kube-system"}`, true},
		{"json other namespace", "jellyfin-a", `[{"name":"underlay-macvlan","namespace":"default"}]`, false},
		{"json no namespace resolves to pod namespace", "kube-system", `[{"name":"underlay-macvlan"}]`, true},
		{"json no namespace in user namespace", "jellyfin-a", `[{"name":"underlay-macvlan"}]`, false},
		{"json malformed", "jellyfin-a", `[{"name":"underlay-macvlan",`, false},
		// short form as written by the pre-0.6.38 webhook
		{"short form", "jellyfin-b", "kube-system/underlay-macvlan", true},
		{"short form with interface", "jellyfin-b", "kube-system/underlay-macvlan@net1", true},
		{"short form list", "jellyfin-b", "kube-system/some-other-net, kube-system/underlay-macvlan", true},
		{"short form other namespace", "jellyfin-b", "default/underlay-macvlan", false},
		{"short form same prefix", "jellyfin-b", "kube-system/underlay-macvlan-v2", false},
		{"short form no namespace in user namespace", "jellyfin-b", "underlay-macvlan", false},
		{"unrelated network", "user-space-a", "kube-system/some-other-net", false},
		{"no annotation", "user-space-a", "", false},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			if got := podSelectsUnderlay(overlayPod(tc.ns, "p", tc.selection)); got != tc.want {
				t.Fatalf("selection %q in %s: got %v want %v", tc.selection, tc.ns, got, tc.want)
			}
		})
	}
}

func TestRecreateOverlayGatewayPodsSelectsOnlyUnderlayAttachments(t *testing.T) {
	now := metav1.Now()
	terminating := overlayPod("homeassistant-vodevall", "ha-old", "kube-system/underlay-macvlan")
	terminating.DeletionTimestamp = &now
	terminating.Finalizers = []string{"test/keep"}

	standalone := overlayPod("user-space-a", "bare", "kube-system/underlay-macvlan")
	standalone.OwnerReferences = nil

	kube := fake.NewSimpleClientset(
		overlayPod("jellyfin-brucedai", "jf-1", `[{"mac":"02:af:0a:43:df:84","name":"underlay-macvlan","namespace":"kube-system"}]`),
		overlayPod("jellyfin-vodevall", "jf-2", "kube-system/underlay-macvlan"),
		overlayPod("user-space-a", "other", "kube-system/some-other-net"),
		overlayPod("user-space-a", "lookalike", "kube-system/underlay-macvlan-v2"),
		overlayPod("user-space-a", "plain", ""),
		standalone,
		terminating,
	)

	deleted, remaining, err := recreateOverlayGatewayPods(context.Background(), kube)
	if err != nil {
		t.Fatal(err)
	}
	if deleted != 2 || remaining != 0 {
		t.Fatalf("deleted=%d remaining=%d want 2/0", deleted, remaining)
	}
	for _, tc := range []struct {
		ns, name string
		want     bool
	}{
		{"jellyfin-brucedai", "jf-1", false},
		{"jellyfin-vodevall", "jf-2", false},
		{"user-space-a", "other", true},
		{"user-space-a", "lookalike", true},
		{"user-space-a", "plain", true},
		// a Pod no controller owns must survive: deleting it is not a recreate
		{"user-space-a", "bare", true},
		{"homeassistant-vodevall", "ha-old", true},
	} {
		_, err := kube.CoreV1().Pods(tc.ns).Get(context.Background(), tc.name, metav1.GetOptions{})
		if exists := err == nil; exists != tc.want {
			t.Fatalf("%s/%s exists=%v want %v", tc.ns, tc.name, exists, tc.want)
		}
	}
}

func underlayPods(n int) []k8sruntime.Object {
	objs := make([]k8sruntime.Object, 0, n)
	for i := 0; i < n; i++ {
		objs = append(objs, overlayPod("ns", fmt.Sprintf("p%d", i), "kube-system/underlay-macvlan"))
	}
	return objs
}

// stubBatchWait replaces the pause between delete batches for the duration of
// a test. The stub is handed the pause number, starting at 1, and its error is
// what the loop sees; the deletes themselves keep running under a live context.
func stubBatchWait(t *testing.T, fn func(n int) error) *int {
	t.Helper()
	prev := overlayRecreateBatchWait
	pauses := 0
	overlayRecreateBatchWait = func(_ context.Context, d time.Duration) error {
		if d != overlayRecreateBatchPause {
			t.Errorf("batch wait called with %v, want the configured pause %v", d, overlayRecreateBatchPause)
		}
		pauses++
		return fn(pauses)
	}
	t.Cleanup(func() { overlayRecreateBatchWait = prev })
	return &pauses
}

func TestRecreateOverlayGatewayPodsDeletesAllInBatches(t *testing.T) {
	pauses := stubBatchWait(t, func(int) error { return nil })

	total := 2*overlayRecreateBatchLimit + 5
	kube := fake.NewSimpleClientset(underlayPods(total)...)
	deleted, remaining, err := recreateOverlayGatewayPods(context.Background(), kube)
	if err != nil {
		t.Fatal(err)
	}
	if deleted != total || remaining != 0 {
		t.Fatalf("deleted=%d remaining=%d want %d/0", deleted, remaining, total)
	}
	// one pause between each pair of batches, none before the first
	if want := (total - 1) / overlayRecreateBatchLimit; *pauses != want {
		t.Fatalf("paused %d times, want %d for %d pods in batches of %d", *pauses, want, total, overlayRecreateBatchLimit)
	}
	left, _ := kube.CoreV1().Pods("ns").List(context.Background(), metav1.ListOptions{})
	if len(left.Items) != 0 {
		t.Fatalf("%d pods left", len(left.Items))
	}
}

func TestRecreateOverlayGatewayPodsReportsRemainingWhenTimeIsUp(t *testing.T) {
	// the deadline passes during the first pause, so the first batch is
	// deleted and everything after it is reported as remaining
	stubBatchWait(t, func(int) error { return context.DeadlineExceeded })

	total := overlayRecreateBatchLimit + 7
	kube := fake.NewSimpleClientset(underlayPods(total)...)
	deleted, remaining, err := recreateOverlayGatewayPods(context.Background(), kube)
	if !errors.Is(err, context.DeadlineExceeded) {
		t.Fatalf("err = %v, want context.DeadlineExceeded", err)
	}
	if deleted != overlayRecreateBatchLimit || remaining != 7 {
		t.Fatalf("deleted=%d remaining=%d want %d/7", deleted, remaining, overlayRecreateBatchLimit)
	}
	left, _ := kube.CoreV1().Pods("ns").List(context.Background(), metav1.ListOptions{})
	if len(left.Items) != 7 {
		t.Fatalf("%d pods left, want the 7 that were never reached", len(left.Items))
	}
}

func TestCniDhcpInstallCommandsTouchOnlyTheDhcpPlugin(t *testing.T) {
	const archive = "/tmp/kubekey/cni-plugins-v1.6.2-olares2.tgz"
	cmds := cniDhcpInstallCommands(archive)

	var descs, all []string
	for _, c := range cmds {
		descs = append(descs, c.desc)
		all = append(all, c.cmd)
	}
	joined := strings.Join(all, "\n")

	// Every tar invocation must name a member explicitly. Extracting the whole
	// archive over /opt/cni/bin would downgrade plugins that belong to other
	// components or were replaced by a later build.
	for _, cmd := range all {
		if !strings.Contains(cmd, "tar ") {
			continue
		}
		for _, inv := range strings.Split(cmd, "||") {
			inv = strings.TrimSpace(inv)
			if !strings.HasSuffix(inv, " ./dhcp") && !strings.HasSuffix(inv, " dhcp") {
				t.Fatalf("tar invocation must extract only the dhcp member: %q", inv)
			}
		}
	}

	// No other plugin, and no file shipped alongside them, may be named.
	for _, other := range []string{"macvlan", "bridge", "host-local", "portmap", "bandwidth", "ipvlan", "loopback", "tuning", "calico", "multus", "LICENSE", "README"} {
		if strings.Contains(joined, other) {
			t.Fatalf("commands must not mention %q:\n%s", other, joined)
		}
	}

	// The staging directory must be a sibling of the plugin directory: same
	// filesystem, so putting the binary in place is a rename, but nothing that
	// is not a plugin is ever written into the plugin directory.
	if parent := path.Dir(cniBinDir); path.Dir(cniDhcpStageDir) != parent {
		t.Fatalf("staging dir %q must sit next to %q, under %q", cniDhcpStageDir, cniBinDir, parent)
	}
	if strings.HasPrefix(cniDhcpStageDir, cniBinDir+"/") {
		t.Fatalf("staging dir %q must not be inside the plugin directory", cniDhcpStageDir)
	}
	if want := "mv -f " + cniDhcpStageDir + "/dhcp " + cniBinDir + "/dhcp"; !strings.Contains(joined, want) {
		t.Fatalf("the plugin must be installed with %q:\n%s", want, joined)
	}

	// Stage, check it runs, install, restart: in that order.
	wantOrder := []string{
		"rm -rf " + cniDhcpStageDir,
		"tar -zxf " + archive,
		cniDhcpStageDir + "/dhcp --version",
		"mv -f ",
		"systemctl restart cni-dhcp",
	}
	at := 0
	for _, want := range wantOrder {
		found := -1
		for i := at; i < len(all); i++ {
			if strings.Contains(all[i], want) {
				found = i
				break
			}
		}
		if found < 0 {
			t.Fatalf("missing or out-of-order step %q in %v", want, all)
		}
		at = found
	}
	if len(cmds) != len(wantOrder) {
		t.Fatalf("unexpected extra steps: %v", descs)
	}
}

func TestUpgrader20260907PhaseOrder(t *testing.T) {
	u := upgrader_1_12_7_20260907{}
	if u.Version().String() != "1.12.7-20260907" {
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
	unit, render, deploy := index(sys, "GenerateMultusDhcpService"), index(sys, "GenerateMultusDefine"), index(sys, "DeployMultusDefine")
	swap, recreate := index(sys, "UpgradeCniDhcpBinary"), index(sys, "RecreateOverlayGatewayPods")
	for name, at := range map[string]int{"GenerateMultusDhcpService": unit, "GenerateMultusDefine": render, "DeployMultusDefine": deploy, "UpgradeCniDhcpBinary": swap, "RecreateOverlayGatewayPods": recreate} {
		if at < 0 {
			t.Fatalf("UpgradeSystemComponents must contain %s, got %v", name, sys)
		}
	}
	// The restart drops the daemon's leases, so it must come after the NAD and
	// the unit are in place and immediately before the recreate.
	if !(unit < swap && render < deploy && deploy < swap && swap < recreate) {
		t.Fatalf("order must be render unit/NAD < deploy NAD < swap plugin < recreate pods, got unit=%d render=%d deploy=%d swap=%d recreate=%d", unit, render, deploy, swap, recreate)
	}
	if n := recreate - swap; n != 1 {
		t.Fatalf("the recreate must directly follow the plugin swap, %d task(s) in between", n-1)
	}
	for _, phase := range []struct {
		name  string
		tasks []task.Interface
	}{{"PrepareForUpgrade", u.PrepareForUpgrade()}, {"PostUpgrade", u.PostUpgrade()}} {
		got := names(phase.tasks)
		if index(got, "UpgradeCniDhcpBinary") >= 0 || index(got, "RecreateOverlayGatewayPods") >= 0 {
			t.Fatalf("%s must not carry cni-dhcp steps (they run back to back in UpgradeSystemComponents), got %v", phase.name, got)
		}
	}
}
