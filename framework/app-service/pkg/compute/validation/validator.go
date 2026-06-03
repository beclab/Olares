// Package validation aggregates resource feasibility checks for install
// and resume. Legacy installs run InstallRuntimePressureValidators before
// helm; two-phase installs (workloadReplicas) run the same chain after
// helm at replicas=0 and before Scale(-1). Upgrade does not use this
// package.
//
// Each individual check (cluster pressure, per-user quota, k8s request
// availability, per-node pressure, GPU compute plan) is wrapped in a
// Validator. The orchestrator (Run / AppInstallable wrapper) picks the
// applicable subset for the given op and runs them in order, returning
// the first non-OK Decision. This replaces a previous pattern where each
// call site duplicated subsets of the same checks inline.
package validation

import (
	"context"

	"github.com/beclab/Olares/framework/app-service/pkg/appcfg"
	"github.com/beclab/Olares/framework/app-service/pkg/constants"
	"github.com/beclab/api/api/app.bytetrade.io/v1alpha1"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

// Op is an alias for v1alpha1.OpType so callers don't need to reach into
// the api module just to express which validators apply.
type Op = v1alpha1.OpType

// Input bundles everything a Validator might need. Token is optional and
// only the cluster-pressure validator currently uses it; the others
// ignore unset fields.
type Input struct {
	Client    client.Client
	AppConfig *appcfg.ApplicationConfig
	Op        Op
	Token     string
}

// Decision is the structured outcome of a single validator (or the chain
// as a whole). When OK is false the remaining fields describe which
// resource bucket failed and why, in a shape the HTTP handlers can map
// straight to api.RequirementResp.
type Decision struct {
	OK        bool
	Resource  constants.ResourceType
	Reason    constants.ResourceConditionType
	Message   string
	Validator string // populated by Run to surface which Validator decided
}

// ok is the standard success value shared by all validators.
func ok() Decision { return Decision{OK: true} }

// Validator wraps a single resource check.
//
//   - Name identifies the validator in logs and Decision.Validator.
//   - AppliesTo lets the chain executor filter by op (e.g. cluster
//     pressure runs for install/resume; upgrade is out of scope).
//   - Validate performs the check and returns a Decision. A non-nil
//     error means the check itself couldn't be evaluated and the caller
//     should treat the result as fatal.
type Validator interface {
	Name() string
	AppliesTo(op Op) bool
	Validate(ctx context.Context, in Input) (Decision, error)
}
