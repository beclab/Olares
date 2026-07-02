package controllers

import (
	"context"
	"testing"

	appv1alpha1 "github.com/beclab/api/api/app.bytetrade.io/v1alpha1"
	"k8s.io/apimachinery/pkg/runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/client/fake"

	"github.com/beclab/Olares/framework/app-service/pkg/apiserver/api"
	"github.com/beclab/Olares/framework/app-service/pkg/applifecycle/domain"
	"github.com/beclab/Olares/framework/app-service/pkg/constants"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

type recordOps struct {
	name   string
	action domain.Action
	calls  int
}

func (r *recordOps) Trigger(name string, action domain.Action) {
	r.calls++
	r.name = name
	r.action = action
}

func TestStopAlreadyInFlight_UnconsumedRequest(t *testing.T) {
	am := &appv1alpha1.ApplicationManager{
		ObjectMeta: metav1.ObjectMeta{Annotations: map[string]string{
			domain.AnnotationOperationRequestID: "new-stop",
		}},
		Spec: appv1alpha1.ApplicationManagerSpec{OpType: appv1alpha1.StopOp},
		Status: appv1alpha1.ApplicationManagerStatus{
			State:  appv1alpha1.Running,
			OpType: appv1alpha1.InstallOp,
			OpID:   "old-op",
		},
	}
	if !stopAlreadyInFlight(am) {
		t.Fatal("StopOp request not yet copied into Status should count as in flight")
	}
	am.Status.OpID = "new-stop"
	am.Status.State = appv1alpha1.Running
	am.Status.OpType = appv1alpha1.InstallOp
	if stopAlreadyInFlight(am) {
		t.Fatal("consumed request on a Running install should allow another stop")
	}
}

func TestDeferSuspendForOtherUnconsumedRequest(t *testing.T) {
	for _, op := range []appv1alpha1.OpType{
		appv1alpha1.UninstallOp,
		appv1alpha1.UpgradeOp,
		appv1alpha1.ApplyEnvOp,
	} {
		am := &appv1alpha1.ApplicationManager{
			ObjectMeta: metav1.ObjectMeta{Annotations: map[string]string{
				domain.AnnotationOperationRequestID: "pending-other",
			}},
			Spec: appv1alpha1.ApplicationManagerSpec{OpType: op},
			Status: appv1alpha1.ApplicationManagerStatus{
				State:  appv1alpha1.Running,
				OpType: appv1alpha1.InstallOp,
				OpID:   "old-op",
			},
		}
		if stopAlreadyInFlight(am) {
			t.Fatalf("%s: pending request counted as an in-flight stop", op)
		}
		if !deferSuspend(am) {
			t.Fatalf("%s: suspend should wait for the unconsumed request", op)
		}
	}

	stop := &appv1alpha1.ApplicationManager{
		ObjectMeta: metav1.ObjectMeta{Annotations: map[string]string{
			domain.AnnotationOperationRequestID: "pending-stop",
		}},
		Spec:   appv1alpha1.ApplicationManagerSpec{OpType: appv1alpha1.StopOp},
		Status: appv1alpha1.ApplicationManagerStatus{State: appv1alpha1.Running, OpID: "old-op"},
	}
	if deferSuspend(stop) {
		t.Fatal("a pending Stop is the suspend itself, not a reason to defer")
	}
}

func TestSuspendResolved_PatchesIntentAndTriggersStop(t *testing.T) {
	scheme := runtime.NewScheme()
	if err := appv1alpha1.AddToScheme(scheme); err != nil {
		t.Fatal(err)
	}
	am := &appv1alpha1.ApplicationManager{
		ObjectMeta: metav1.ObjectMeta{Name: "app-alice"},
		Spec: appv1alpha1.ApplicationManagerSpec{
			AppName: "app",
			OpType:  appv1alpha1.InstallOp,
		},
		Status: appv1alpha1.ApplicationManagerStatus{
			State:  appv1alpha1.Running,
			OpType: appv1alpha1.InstallOp,
			OpID:   "old-op",
		},
	}
	c := fake.NewClientBuilder().WithScheme(scheme).WithObjects(am).Build()
	ops := &recordOps{}
	r := &PodAbnormalSuspendAppController{Client: c, Ops: ops}

	err := r.suspendResolved(context.Background(), am, constants.AppUnschedulable, "pending unschedulable timeout on pod: ns/pod", "alice")
	if err != nil {
		t.Fatal(err)
	}

	got := &appv1alpha1.ApplicationManager{}
	if err := c.Get(context.Background(), client.ObjectKey{Name: "app-alice"}, got); err != nil {
		t.Fatal(err)
	}
	if got.Status.State != appv1alpha1.Running || got.Status.OpID != "old-op" {
		t.Fatalf("status was written: state=%s opID=%s", got.Status.State, got.Status.OpID)
	}
	if got.Spec.OpType != appv1alpha1.StopOp {
		t.Fatalf("Spec.OpType = %q, want StopOp", got.Spec.OpType)
	}
	if got.Annotations[api.AppStopAllKey] != "" {
		t.Fatalf("stop-all = %q, want empty for a user namespace", got.Annotations[api.AppStopAllKey])
	}
	if got.Annotations[api.AppStopByControllerDuePendingPod] != "true" {
		t.Fatalf("pending-pod = %q, want true", got.Annotations[api.AppStopByControllerDuePendingPod])
	}
	if got.Annotations[api.AppStopReasonKey] != constants.AppUnschedulable {
		t.Fatalf("stop-reason = %q", got.Annotations[api.AppStopReasonKey])
	}
	if got.Annotations[api.AppStopMessageKey] == "" {
		t.Fatal("stop-message is empty")
	}
	if got.Annotations[domain.AnnotationOperationRequestID] == "" || got.Annotations[domain.AnnotationOperationRequestID] == "old-op" {
		t.Fatalf("operation request id = %q", got.Annotations[domain.AnnotationOperationRequestID])
	}
	if ops.calls != 1 || ops.name != "app-alice" {
		t.Fatalf("Trigger calls=%d name=%s", ops.calls, ops.name)
	}
	op, ok := ops.action.(domain.TriggerOp)
	if !ok || op.Op != appv1alpha1.StopOp {
		t.Fatalf("Trigger action = %#v, want TriggerOp{StopOp}", ops.action)
	}
}

func TestSuspendResolved_SharedEvictionSetsStopAllOnly(t *testing.T) {
	scheme := runtime.NewScheme()
	if err := appv1alpha1.AddToScheme(scheme); err != nil {
		t.Fatal(err)
	}
	am := &appv1alpha1.ApplicationManager{
		ObjectMeta: metav1.ObjectMeta{Name: "app-shared-app"},
		Spec:       appv1alpha1.ApplicationManagerSpec{OpType: appv1alpha1.InstallOp},
		Status:     appv1alpha1.ApplicationManagerStatus{State: appv1alpha1.Running},
	}
	c := fake.NewClientBuilder().WithScheme(scheme).WithObjects(am).Build()
	ops := &recordOps{}
	r := &PodAbnormalSuspendAppController{Client: c, Ops: ops}

	err := r.suspendResolved(context.Background(), am, constants.AppStopDueToEvicted, "evicted pod: app-shared/pod", "app-shared")
	if err != nil {
		t.Fatal(err)
	}
	got := &appv1alpha1.ApplicationManager{}
	if err := c.Get(context.Background(), client.ObjectKey{Name: "app-shared-app"}, got); err != nil {
		t.Fatal(err)
	}
	if got.Annotations[api.AppStopAllKey] != "true" {
		t.Fatalf("stop-all = %q, want true", got.Annotations[api.AppStopAllKey])
	}
	if _, ok := got.Annotations[api.AppStopByControllerDuePendingPod]; ok {
		t.Fatalf("pending-pod set on eviction: %q", got.Annotations[api.AppStopByControllerDuePendingPod])
	}
	if got.Status.State != appv1alpha1.Running {
		t.Fatalf("status state = %s, want Running", got.Status.State)
	}
}
