package apiserver

import (
	"github.com/google/uuid"

	"github.com/beclab/Olares/framework/app-service/pkg/applifecycle/domain"
)

// stampDurableOperationRequest returns the annotations plus the request ID it
// stamped. The Runner copies that ID into Status.OpID when it consumes the
// TriggerOp (see runner.stampTriggerOperation), so handlers can hand it back
// to the caller as the operation identity before the first Reconcile lands.
func stampDurableOperationRequest(annotations map[string]string) (map[string]string, string) {
	if annotations == nil {
		annotations = make(map[string]string)
	}
	requestID := uuid.NewString()
	annotations[domain.AnnotationOperationRequestID] = requestID
	return annotations, requestID
}
