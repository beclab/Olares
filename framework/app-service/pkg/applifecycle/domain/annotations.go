package domain

import appsv1 "github.com/beclab/api/api/app.bytetrade.io/v1alpha1"

// AnnotationOperationRequestID identifies the latest accepted user operation
// request. The Runner copies it to Status.OpID when consuming Spec.OpType.
const AnnotationOperationRequestID = "bytetrade.io/operation-request-id"

// UnconsumedOperation returns the Spec.OpType whose request id the Runner
// has not copied into Status.OpID. Status.State still describes the
// previous operation during that window. The boolean is false when no
// request is waiting.
func UnconsumedOperation(am *appsv1.ApplicationManager) (OpType, bool) {
	if am == nil || am.Spec.OpType == "" || am.Annotations == nil {
		return "", false
	}
	requestID := am.Annotations[AnnotationOperationRequestID]
	if requestID == "" || requestID == am.Status.OpID {
		return "", false
	}
	return am.Spec.OpType, true
}

// OperationRequestPending reports that op was accepted onto Spec and the
// operation-request annotation, and the Runner has not copied that id into
// Status.OpID yet.
func OperationRequestPending(am *appsv1.ApplicationManager, op OpType) bool {
	got, ok := UnconsumedOperation(am)
	return ok && got == op
}
