package domain

import (
	"testing"

	appsv1 "github.com/beclab/api/api/app.bytetrade.io/v1alpha1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

func TestOperationRequestPending(t *testing.T) {
	am := &appsv1.ApplicationManager{
		ObjectMeta: metav1.ObjectMeta{Annotations: map[string]string{
			AnnotationOperationRequestID: "new-stop",
		}},
		Spec:   appsv1.ApplicationManagerSpec{OpType: appsv1.StopOp},
		Status: appsv1.ApplicationManagerStatus{State: appsv1.Running, OpType: appsv1.InstallOp, OpID: "old-op"},
	}
	if !OperationRequestPending(am, appsv1.StopOp) {
		t.Fatal("unconsumed StopOp request should be pending")
	}
	if OperationRequestPending(am, appsv1.ApplyEnvOp) {
		t.Fatal("a StopOp request is not a pending ApplyEnv")
	}

	am.Status.OpID = "new-stop"
	if OperationRequestPending(am, appsv1.StopOp) {
		t.Fatal("request copied into Status.OpID is no longer pending")
	}
	if op, ok := UnconsumedOperation(am); ok {
		t.Fatalf("UnconsumedOperation = %q, true, want none", op)
	}
}

func TestUnconsumedOperation(t *testing.T) {
	am := &appsv1.ApplicationManager{
		ObjectMeta: metav1.ObjectMeta{Annotations: map[string]string{
			AnnotationOperationRequestID: "new-upgrade",
		}},
		Spec:   appsv1.ApplicationManagerSpec{OpType: appsv1.UpgradeOp},
		Status: appsv1.ApplicationManagerStatus{State: appsv1.Running, OpID: "old-op"},
	}
	op, ok := UnconsumedOperation(am)
	if !ok || op != appsv1.UpgradeOp {
		t.Fatalf("UnconsumedOperation = %q, %v, want UpgradeOp, true", op, ok)
	}

	am.Annotations = nil
	if op, ok := UnconsumedOperation(am); ok {
		t.Fatalf("UnconsumedOperation = %q, true, want none", op)
	}
}

func TestStopRemainsAllowed(t *testing.T) {
	if !StopRemainsAllowed(appsv1.Running, appsv1.ApplyEnvOp) {
		t.Fatal("ApplyEnv lands in ApplyingEnv, which still allows Stop")
	}
	for _, op := range []OpType{appsv1.UpgradeOp, appsv1.UninstallOp, appsv1.StopOp} {
		if StopRemainsAllowed(appsv1.Running, op) {
			t.Fatalf("StopRemainsAllowed(Running, %s) = true, want false", op)
		}
	}
	if StopRemainsAllowed(appsv1.ApplyingEnv, appsv1.CancelOp) {
		t.Fatal("cancel from ApplyingEnv lands where Stop is not allowed")
	}
}
