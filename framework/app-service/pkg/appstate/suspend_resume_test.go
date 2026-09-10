package appstate

import (
	"context"
	"errors"
	"strings"
	"testing"

	"github.com/beclab/Olares/framework/app-service/pkg/appcfg"
	"github.com/beclab/Olares/framework/app-service/pkg/testutil"
	appv1alpha1 "github.com/beclab/api/api/app.bytetrade.io/v1alpha1"

	k8sappsv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/types"
	"k8s.io/utils/ptr"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

func TestSuspendThenResumeViaPatch(t *testing.T) {
	dep := testutil.NewDeployment("nginx", "nginx-alice", 1)
	sts := testutil.NewStatefulSet("nginx-sts", "nginx-alice", 2)
	am := testutil.NewAppManager("nginx", testutil.WithNamespace("nginx-alice"))
	c := testutil.NewFakeClient(dep, sts, am)

	if err := suspendV1AppOrV2Client(context.TODO(), c, am); err != nil {
		t.Fatalf("suspend: %v", err)
	}
	var d k8sappsv1.Deployment
	if err := c.Get(context.TODO(), types.NamespacedName{Name: "nginx", Namespace: "nginx-alice"}, &d); err != nil {
		t.Fatalf("get deployment: %v", err)
	}
	if *d.Spec.Replicas != 0 {
		t.Errorf("suspended replicas=%d want 0", *d.Spec.Replicas)
	}
	if d.Annotations[suspendAnnotation] != "app-service" {
		t.Errorf("missing suspend annotation: %v", d.Annotations)
	}
	var s k8sappsv1.StatefulSet
	if err := c.Get(context.TODO(), types.NamespacedName{Name: "nginx-sts", Namespace: "nginx-alice"}, &s); err != nil {
		t.Fatalf("get sts: %v", err)
	}
	if *s.Spec.Replicas != 0 {
		t.Errorf("suspended sts replicas=%d want 0", *s.Spec.Replicas)
	}

	if err := resumeV1AppOrV2AppClient(context.TODO(), c, am); err != nil {
		t.Fatalf("resume: %v", err)
	}
	if err := c.Get(context.TODO(), types.NamespacedName{Name: "nginx", Namespace: "nginx-alice"}, &d); err != nil {
		t.Fatalf("get deployment: %v", err)
	}
	if *d.Spec.Replicas != 1 {
		t.Errorf("resumed replicas=%d want 1", *d.Spec.Replicas)
	}
}

// workloadReplicas config routes suspend/resume through helm Scale.
func cfgWithReplicas() *appcfg.ApplicationConfig {
	wr := appcfg.WorkloadReplicas{"nginx": 1}
	return &appcfg.ApplicationConfig{AppName: "nginx", Namespace: "nginx-alice", OwnerName: "alice", WorkloadReplicas: &wr}
}

func TestScaleOrPatchResumeUsesHelmScaleUp(t *testing.T) {
	am := testutil.NewAppManager("nginx",
		testutil.WithNamespace("nginx-alice"),
		testutil.WithConfig(t, cfgWithReplicas()),
	)
	c := testutil.NewFakeClient(am)
	fake := testutil.NewFakeHelmOps()
	p := &ResumingApp{&baseOperationApp{baseStatefulApp: &baseStatefulApp{manager: am, client: c, deps: fakeDeps(fake)}}}

	injectHelmOps(t, fake)

	if err := p.scaleOrPatchResume(context.TODO(), false); err != nil {
		t.Fatalf("scaleOrPatchResume: %v", err)
	}
	if got := fake.ScaleReplicas(); len(got) != 1 || got[0] != -1 {
		t.Errorf("Scale args=%v want [-1]", got)
	}
}

func TestScaleOrPatchSuspendUsesHelmScaleToZero(t *testing.T) {
	am := testutil.NewAppManager("nginx",
		testutil.WithNamespace("nginx-alice"),
		testutil.WithConfig(t, cfgWithReplicas()),
	)
	c := testutil.NewFakeClient(am)
	fake := testutil.NewFakeHelmOps()
	p := &SuspendingApp{&baseOperationApp{baseStatefulApp: &baseStatefulApp{manager: am, client: c, deps: fakeDeps(fake)}}}

	injectHelmOps(t, fake)

	if err := p.scaleOrPatchSuspend(context.TODO(), false); err != nil {
		t.Fatalf("scaleOrPatchSuspend: %v", err)
	}
	if got := fake.ScaleReplicas(); len(got) != 1 || got[0] != 0 {
		t.Errorf("Scale args=%v want [0]", got)
	}
}

func TestScaleOrPatchResumePropagatesScaleError(t *testing.T) {
	am := testutil.NewAppManager("nginx",
		testutil.WithNamespace("nginx-alice"),
		testutil.WithConfig(t, cfgWithReplicas()),
	)
	c := testutil.NewFakeClient(am)
	fake := testutil.NewFakeHelmOps()
	fake.ScaleErr = errors.New("scale boom")
	p := &ResumingApp{&baseOperationApp{baseStatefulApp: &baseStatefulApp{manager: am, client: c, deps: fakeDeps(fake)}}}

	injectHelmOps(t, fake)

	if err := p.scaleOrPatchResume(context.TODO(), false); err == nil {
		t.Fatal("expected scale error to propagate")
	}
}

func newPodForStop(name string, phase corev1.PodPhase) *corev1.Pod {
	return &corev1.Pod{
		ObjectMeta: metav1.ObjectMeta{Name: name, Namespace: "nginx-alice"},
		Status:     corev1.PodStatus{Phase: phase},
	}
}

func suspendingAppFor(am *appv1alpha1.ApplicationManager, c client.Client) *SuspendingApp {
	return &SuspendingApp{&baseOperationApp{baseStatefulApp: &baseStatefulApp{manager: am, client: c}}}
}

// A chart leaves completed hook and job pods behind on purpose, and counting
// them as live made every stop of such an app burn the whole Stopping TTL
// before reporting a timeout.
func TestStopWaitIgnoresPodsThatAlreadyFinished(t *testing.T) {
	am := testutil.NewAppManager("nginx", testutil.WithNamespace("nginx-alice"))
	done := newPodForStop("nginx-predelete", corev1.PodSucceeded)
	failed := newPodForStop("nginx-migrate", corev1.PodFailed)
	p := suspendingAppFor(am, testutil.NewFakeClient(am, done, failed))

	if err := p.waitForPodsGone(context.TODO()); err != nil {
		t.Fatalf("waitForPodsGone: %v", err)
	}
}

// Waiting out the TTL told an operator only that something timed out. The pod
// is the symptom and its owner is what has to change, so both are named.
func TestStopWaitNamesThePodItCannotRemove(t *testing.T) {
	restore := podStopGrace
	podStopGrace = 0
	t.Cleanup(func() { podStopGrace = restore })

	am := testutil.NewAppManager("nginx", testutil.WithNamespace("nginx-alice"))
	engine := newPodForStop("fs-eng-abc", corev1.PodRunning)
	engine.OwnerReferences = []metav1.OwnerReference{
		{APIVersion: "apps/v1", Kind: "ReplicaSet", Name: "fs-eng-abc-7d9", UID: "u", Controller: ptr.To(true)},
	}
	p := suspendingAppFor(am, testutil.NewFakeClient(am, engine))

	err := p.waitForPodsGone(context.TODO())
	if err == nil {
		t.Fatal("expected a pod nobody asked to terminate to fail the stop")
	}
	for _, want := range []string{"fs-eng-abc", "ReplicaSet/fs-eng-abc-7d9", "nginx-alice"} {
		if !strings.Contains(err.Error(), want) {
			t.Errorf("error %q does not mention %q", err, want)
		}
	}
}

// A pod carrying a DeletionTimestamp is terminating legitimately and must keep
// the stop waiting; only a pod nobody asked to stop is a blocker.
func TestClassifyPodsForStopSeparatesTerminatingFromStuck(t *testing.T) {
	now := metav1.Now()
	terminating := *newPodForStop("nginx-1", corev1.PodRunning)
	terminating.DeletionTimestamp = &now
	stuck := *newPodForStop("nginx-2", corev1.PodRunning)

	count, blockers := classifyPodsForStop([]corev1.Pod{terminating, stuck})
	if count != 1 {
		t.Errorf("terminating=%d want 1", count)
	}
	if len(blockers) != 1 || blockers[0].name != "nginx-2" {
		t.Errorf("blockers=%v want only nginx-2", blockers)
	}
}

func TestDescribePodOwnerPrefersTheController(t *testing.T) {
	pod := newPodForStop("p", corev1.PodRunning)
	pod.OwnerReferences = []metav1.OwnerReference{
		{Kind: "Deployment", Name: "not-the-controller"},
		{Kind: "ReplicaSet", Name: "the-controller", Controller: ptr.To(true)},
	}
	if got := describePodOwner(pod); got != "ReplicaSet/the-controller" {
		t.Errorf("owner=%q want ReplicaSet/the-controller", got)
	}
	// An unowned pod is the case where scaling a workload down cannot help.
	if got := describePodOwner(newPodForStop("p", corev1.PodRunning)); got != "no owner" {
		t.Errorf("owner=%q want %q", got, "no owner")
	}
}
