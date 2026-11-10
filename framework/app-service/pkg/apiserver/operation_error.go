package apiserver

import (
	"fmt"

	"github.com/beclab/Olares/framework/app-service/pkg/apiserver/api"
	"github.com/beclab/Olares/framework/app-service/pkg/applifecycle/domain"
	appv1alpha1 "github.com/beclab/api/api/app.bytetrade.io/v1alpha1"
	"github.com/emicklei/go-restful/v3"
)

// explainRequestPending rejects a new operation while Spec still holds a
// request the Runner has not copied into Status.OpID. Status.State still
// describes the previous operation, so AllowsOp alone would accept the
// write and replace the request id.
func explainRequestPending(pending, want appv1alpha1.OpType) error {
	if pending == want {
		return fmt.Errorf("%s operation is already pending", want)
	}
	return fmt.Errorf("%s operation is not allowed while %s is pending", want, pending)
}

// blockIfRequestPending writes a 400 and reports true when am has an
// unconsumed operation. Callers must not patch Spec.OpType after it returns
// true.
func blockIfRequestPending(resp *restful.Response, req *restful.Request, am *appv1alpha1.ApplicationManager, want appv1alpha1.OpType) bool {
	pending, ok := domain.UnconsumedOperation(am)
	if !ok {
		return false
	}
	api.HandleBadRequest(resp, req, explainRequestPending(pending, want))
	return true
}

// explainOperationNotAllowed builds the error returned when op is rejected in
// curState. When the state has an in-progress operation that can be cancelled
// (CancelOp is allowed), the message guides the caller to cancel first instead
// of just reporting "not allowed", which otherwise leaves callers retrying the
// same blocked operation (e.g. uninstall stuck behind an initializing app).
func explainOperationNotAllowed(curState appv1alpha1.ApplicationManagerState, op appv1alpha1.OpType) error {
	if domain.DefaultTable.AllowsOp(curState, appv1alpha1.CancelOp) {
		return fmt.Errorf("%s operation is not allowed while the app is in %s state: an operation is in progress; "+
			"cancel it first via POST /app-service/v1/apps/{name}/cancel and wait until the app reaches a terminal state, then retry %s",
			op, curState, op)
	}
	return fmt.Errorf("%s operation is not allowed for %s state", op, curState)
}
