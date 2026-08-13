package clusterop

import (
	"context"
	"encoding/json"
	"time"

	"github.com/beclab/Olares/daemon/pkg/cluster/nodestatus"
)

type Outcome struct {
	Status Status
	Code   string

	// Error is detail, for the log only. It is whatever the module saw —
	// an address, a token, the text of a lower-level error — so the record
	// never keeps it; a reviewed sentence keyed by Code is kept instead.
	// See safeReason.
	Error string

	CommandIssuedUntil time.Time

	// reason is a sentence that has already been reviewed, and which the
	// record keeps exactly as written. It is unexported on purpose: only
	// this package can set it, so the only text that reaches a record
	// verbatim is text written here — the sentences the built-in power
	// operations have always refused with. A module outside this package
	// has Error, which is detail, and Code, which chooses a reviewed
	// sentence. See settledWith.
	reason string

	// recorded says the module has already put this outcome on the record
	// itself, so the manager must not write it again. One operation needs
	// this: a node shutdown has to say the command went out while the
	// machine is still on its way down, then keep watching. Without it the
	// manager would write the same outcome a second time and depend on
	// being refused, which is control flow by error.
	recorded bool
}

// alreadyRecorded marks an outcome the module has put on the record itself.
// It is returned, rather than acted on, so the sequence still says what the
// operation ended as — see Manager.settle.
func (o Outcome) alreadyRecorded() Outcome {
	o.recorded = true
	return o
}

// valid checks Outcome against now in addition to Status: command_issued is
// the one status that carries a live lock deadline, and that deadline is
// meaningless unless it is both present and still in the future. Every other
// outcome is final the moment it is recorded, so it must not carry one.
func (o Outcome) valid(now time.Time) bool {
	switch o.Status {
	case StatusSucceeded, StatusPartiallyFailed, StatusFailed:
		return o.CommandIssuedUntil.IsZero()
	case StatusCommandIssued:
		return !o.CommandIssuedUntil.IsZero() && o.CommandIssuedUntil.After(now)
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

// OperationModule is one cluster operation: what it is called, what it will
// accept, what it makes the cluster look like while it happens, and how it
// is carried out. Everything this daemon can do is a module, and adding one
// is the only thing adding an operation requires — no switch, no list and no
// type constant anywhere else names it.
//
// A module is asked things by two different machines. The master asks
// Validate before an operation exists and then calls Run; a node asks
// Validate again for its own half of the request and then calls
// NodeOperationModule.ExecuteNode. Both are reached only with the owner's
// signature, and neither hop signs the params.
type OperationModule interface {
	Type() Type

	// Validate says whether the module will carry out the request as asked.
	// It is the only judgement made about Params, which no signature covers
	// at any hop: whoever can reach the route chose those bytes.
	//
	// It must be deterministic and free of side effects. Two things depend
	// on that. The same request is judged more than once — once by the
	// master before the operation is recorded, once again by each node that
	// receives its half — and a module that answered differently would
	// leave a recorded operation whose nodes refuse it. And the owner's
	// signature is spent before the answer is known: it is consumed whether
	// this returns nil, returns a refusal, or panics, so a caller can drive
	// this method exactly once per signature and must not be able to make
	// anything happen by doing so.
	//
	// The error is detail for the log. It is text written outside this
	// package, about params nothing checked, so it never reaches a caller.
	Validate(CreateRequest) error

	// Phase is what the operation makes the cluster as a whole look like
	// while it is happening, or false to impose nothing. It is asked about
	// a copy and must not mutate anything.
	Phase(Operation) (nodestatus.Phase, bool)

	// Run carries the operation out on the master. It records what it does
	// through the Runtime and returns the outcome the record settles on.
	Run(context.Context, Runtime, RunRequest) Outcome
}

type RecoverableModule interface {
	Recover(context.Context, Runtime, Operation)
}

type NodeOperationModule interface {
	ExecuteNode(context.Context, NodeRequest) error
}
