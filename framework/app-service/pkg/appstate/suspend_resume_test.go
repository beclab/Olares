package appstate

import (
	"context"
	"errors"
	"testing"

	"github.com/beclab/Olares/framework/app-service/pkg/appcfg"
	"github.com/beclab/Olares/framework/app-service/pkg/testutil"

	k8sappsv1 "k8s.io/api/apps/v1"
	"k8s.io/apimachinery/pkg/types"
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
	p := &ResumingApp{&baseOperationApp{baseStatefulApp: &baseStatefulApp{manager: am, client: c}}}

	fake := testutil.NewFakeHelmOps()
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
	p := &SuspendingApp{&baseOperationApp{baseStatefulApp: &baseStatefulApp{manager: am, client: c}}}

	fake := testutil.NewFakeHelmOps()
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
	p := &ResumingApp{&baseOperationApp{baseStatefulApp: &baseStatefulApp{manager: am, client: c}}}

	fake := testutil.NewFakeHelmOps()
	fake.ScaleErr = errors.New("scale boom")
	injectHelmOps(t, fake)

	if err := p.scaleOrPatchResume(context.TODO(), false); err == nil {
		t.Fatal("expected scale error to propagate")
	}
}
