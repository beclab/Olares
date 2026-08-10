package clusterop

import (
	"context"
	"encoding/json"
	"time"

	"github.com/beclab/Olares/daemon/pkg/cluster/nodestatus"
)

type Outcome struct {
	Status             Status
	Code               string
	Error              string
	CommandIssuedUntil time.Time
}

func (o Outcome) valid() bool {
	switch o.Status {
	case StatusSucceeded, StatusPartiallyFailed, StatusFailed, StatusCommandIssued:
		return true
	default:
		return false
	}
}

type RunRequest struct {
	Creds  Credentials
	Params json.RawMessage
}

type NodeRequest struct {
	PeerRequest
	Params json.RawMessage `json:"params,omitempty"`
}

type Runtime interface {
	Operation() (Operation, bool)
	CanContinue() bool
	StartStep(string) error
	FinishStep(string, StepStatus, string, string) error
	InitNodes([]NodeResult) error
	UpdateNode(string, func(*NodeResult)) error
	SetHostBootID(string) error
	SetModuleState(json.RawMessage) error
	SetCommandIssuedUntil(time.Time) error
	Complete(Outcome) error
	Now() time.Time
	Context() context.Context
}

type OperationModule interface {
	Type() Type
	Validate(CreateRequest) error
	Phase(Operation) (nodestatus.Phase, bool)
	Run(context.Context, Runtime, RunRequest) Outcome
}

type RecoverableModule interface {
	Recover(context.Context, Runtime, Operation)
}

type NodeOperationModule interface {
	ExecuteNode(context.Context, NodeRequest) error
}
