package controllers

import (
	"context"
	"testing"

	appv1alpha1 "github.com/beclab/api/api/app.bytetrade.io/v1alpha1"
	sysv1alpha1 "github.com/beclab/api/api/sys.bytetrade.io/v1alpha1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/client/fake"

	"github.com/beclab/Olares/framework/app-service/pkg/apiserver/api"
	"github.com/beclab/Olares/framework/app-service/pkg/applifecycle/domain"
)

func TestTriggerApplyEnv_PatchesIntentAndTriggers(t *testing.T) {
	scheme := runtime.NewScheme()
	if err := appv1alpha1.AddToScheme(scheme); err != nil {
		t.Fatal(err)
	}
	const name = "demo-ns-demo"
	am := &appv1alpha1.ApplicationManager{
		ObjectMeta: metav1.ObjectMeta{Name: name},
		Spec: appv1alpha1.ApplicationManagerSpec{
			AppName: "demo",
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
	r := &AppEnvController{Client: c, Ops: ops}

	err := r.triggerApplyEnv(context.Background(), &sysv1alpha1.AppEnv{
		ObjectMeta: metav1.ObjectMeta{Namespace: "demo-ns"},
		AppName:    "demo",
		AppOwner:   "alice",
	})
	if err != nil {
		t.Fatal(err)
	}

	got := &appv1alpha1.ApplicationManager{}
	if err := c.Get(context.Background(), client.ObjectKey{Name: name}, got); err != nil {
		t.Fatal(err)
	}
	if got.Status.State != appv1alpha1.Running || got.Status.OpID != "old-op" {
		t.Fatalf("status was written: state=%s opID=%s", got.Status.State, got.Status.OpID)
	}
	if got.Spec.OpType != appv1alpha1.ApplyEnvOp {
		t.Fatalf("Spec.OpType = %q, want ApplyEnvOp", got.Spec.OpType)
	}
	if got.Annotations[api.AppPreUpgradeStateKey] != string(appv1alpha1.Running) {
		t.Fatalf("pre-upgrade state = %q", got.Annotations[api.AppPreUpgradeStateKey])
	}
	if got.Annotations[domain.AnnotationOperationRequestID] == "" {
		t.Fatal("operation request id is empty")
	}
	if ops.calls != 1 || ops.name != name {
		t.Fatalf("Trigger calls=%d name=%s", ops.calls, ops.name)
	}
	op, ok := ops.action.(domain.TriggerOp)
	if !ok || op.Op != appv1alpha1.ApplyEnvOp {
		t.Fatalf("Trigger action = %#v, want TriggerOp{ApplyEnvOp}", ops.action)
	}
}

func TestTriggerApplyEnv_SkipsUnconsumedRequest(t *testing.T) {
	scheme := runtime.NewScheme()
	if err := appv1alpha1.AddToScheme(scheme); err != nil {
		t.Fatal(err)
	}
	const name = "demo-ns-demo"
	am := &appv1alpha1.ApplicationManager{
		ObjectMeta: metav1.ObjectMeta{
			Name: name,
			Annotations: map[string]string{
				domain.AnnotationOperationRequestID: "pending-apply",
			},
		},
		Spec: appv1alpha1.ApplicationManagerSpec{
			AppName: "demo",
			OpType:  appv1alpha1.ApplyEnvOp,
		},
		Status: appv1alpha1.ApplicationManagerStatus{
			State:  appv1alpha1.Running,
			OpType: appv1alpha1.InstallOp,
			OpID:   "old-op",
		},
	}
	c := fake.NewClientBuilder().WithScheme(scheme).WithObjects(am).Build()
	ops := &recordOps{}
	r := &AppEnvController{Client: c, Ops: ops}

	err := r.triggerApplyEnv(context.Background(), &sysv1alpha1.AppEnv{
		ObjectMeta: metav1.ObjectMeta{Namespace: "demo-ns"},
		AppName:    "demo",
		AppOwner:   "alice",
	})
	if err != nil {
		t.Fatal(err)
	}
	if ops.calls != 0 {
		t.Fatalf("Trigger calls=%d, want 0", ops.calls)
	}
	got := &appv1alpha1.ApplicationManager{}
	if err := c.Get(context.Background(), client.ObjectKey{Name: name}, got); err != nil {
		t.Fatal(err)
	}
	if got.Annotations[domain.AnnotationOperationRequestID] != "pending-apply" {
		t.Fatalf("request id = %q, want pending-apply", got.Annotations[domain.AnnotationOperationRequestID])
	}
}

func TestTriggerApplyEnv_LeavesOtherUnconsumedRequest(t *testing.T) {
	for _, op := range []appv1alpha1.OpType{
		appv1alpha1.StopOp,
		appv1alpha1.UninstallOp,
		appv1alpha1.UpgradeOp,
	} {
		t.Run(string(op), func(t *testing.T) {
			scheme := runtime.NewScheme()
			if err := appv1alpha1.AddToScheme(scheme); err != nil {
				t.Fatal(err)
			}
			const name = "demo-ns-demo"
			am := &appv1alpha1.ApplicationManager{
				ObjectMeta: metav1.ObjectMeta{
					Name: name,
					Annotations: map[string]string{
						domain.AnnotationOperationRequestID: "pending-other",
					},
				},
				Spec: appv1alpha1.ApplicationManagerSpec{
					AppName: "demo",
					OpType:  op,
				},
				Status: appv1alpha1.ApplicationManagerStatus{
					State:  appv1alpha1.Running,
					OpType: appv1alpha1.InstallOp,
					OpID:   "old-op",
				},
			}
			c := fake.NewClientBuilder().WithScheme(scheme).WithObjects(am).Build()
			ops := &recordOps{}
			r := &AppEnvController{Client: c, Ops: ops}

			err := r.triggerApplyEnv(context.Background(), &sysv1alpha1.AppEnv{
				ObjectMeta: metav1.ObjectMeta{Namespace: "demo-ns"},
				AppName:    "demo",
				AppOwner:   "alice",
			})
			if err == nil {
				t.Fatal("triggerApplyEnv err = nil, want deferred")
			}
			if ops.calls != 0 {
				t.Fatalf("Trigger calls=%d, want 0", ops.calls)
			}
			got := &appv1alpha1.ApplicationManager{}
			if err := c.Get(context.Background(), client.ObjectKey{Name: name}, got); err != nil {
				t.Fatal(err)
			}
			if got.Spec.OpType != op {
				t.Fatalf("Spec.OpType = %q, want %q", got.Spec.OpType, op)
			}
			if got.Annotations[domain.AnnotationOperationRequestID] != "pending-other" {
				t.Fatalf("request id = %q, want pending-other", got.Annotations[domain.AnnotationOperationRequestID])
			}
		})
	}
}
