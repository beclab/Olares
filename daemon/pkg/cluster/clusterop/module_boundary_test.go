package clusterop

import (
	"context"
	"encoding/json"
	"errors"
	"strings"
	"testing"
	"time"

	"github.com/beclab/Olares/daemon/pkg/cluster/nodestatus"
)

// A module is asked two questions outside Run: whether it will carry out a
// request (Validate) and what it makes the cluster look like while it is
// happening (Phase). Both are other people's code, both are reached from a
// request being served, and neither has anything to do with the operation
// record — so a panic in either must end at this package's boundary rather
// than at the process.

// validatePanicDetail is what a module panics with while judging a request.
// It stands for a stack trace, a raw error, or the very params the caller
// sent: none of it may reach whoever asked.
const validatePanicDetail = "boom: params={\"password\":\"hunter2\"}"

// panickingValidateModule cannot answer whether it will carry out anything.
type panickingValidateModule struct {
	fakeModule
	typ Type
}

func (p *panickingValidateModule) Type() Type { return p.typ }

func (p *panickingValidateModule) Validate(CreateRequest) error { panic(validatePanicDetail) }

func createWithParams(t *testing.T, m *Manager, typ Type, requestID, params string) (Operation, error) {
	t.Helper()
	return m.Create(context.Background(), CreateRequest{
		Type:      typ,
		RequestID: requestID,
		Owner:     "alice@olares.com",
		Params:    json.RawMessage(params),
	})
}

// A module that cannot say whether it accepts a request has not refused it
// and has not started it. The daemon stays up, the caller is told only that
// the operation could not be started, and nothing the module was holding
// when it went wrong is repeated back.
func TestCreateContainsAPanicWhileValidating(t *testing.T) {
	module := &panickingValidateModule{typ: Type("cannot-judge")}
	c := newCluster(master("master-1", "10.0.0.1"))
	m, _ := newManagerWith(t, c, registryWith(t, module))

	_, err := createWithParams(t, m, module.typ, "client-1", `{"password":"hunter2"}`)

	if !errors.Is(err, ErrModuleFailed) {
		t.Fatalf("Create() = %v, want ErrModuleFailed", err)
	}
	for _, leak := range []string{"boom", "hunter2", "panic"} {
		if strings.Contains(err.Error(), leak) {
			t.Errorf("Create() = %v, leaked %q", err, leak)
		}
	}
}

// The request never became an operation: a module that could not judge it is
// not a module that agreed to carry it out.
func TestCreateRecordsNothingWhenValidationPanics(t *testing.T) {
	module := &panickingValidateModule{typ: Type("cannot-judge")}
	c := newCluster(master("master-1", "10.0.0.1"))
	m, _ := newManagerWith(t, c, registryWith(t, module))

	if _, err := createWithParams(t, m, module.typ, "client-1", `{}`); err == nil {
		t.Fatal("Create() = nil error, want a refusal")
	}

	if op, ok := m.GetByRequest("client-1"); ok {
		t.Fatalf("a request nothing judged was recorded anyway: %+v", op)
	}
}

// Nor did it take the cluster's single-operation lock. An operation that does
// not exist cannot be the one in progress, and the next caller is not made to
// wait for it.
func TestCreateHoldsNoLockWhenValidationPanics(t *testing.T) {
	panicking := &panickingValidateModule{typ: Type("cannot-judge")}
	healthy := newFake(Type("healthy"))
	c := newCluster(master("master-1", "10.0.0.1"))
	m, _ := newManagerWith(t, c, registryWith(t, panicking, healthy))

	if _, err := createWithParams(t, m, panicking.typ, "client-1", `{}`); err == nil {
		t.Fatal("Create() = nil error, want a refusal")
	}

	next, err := createFake(t, m, healthy.typ, "client-2")
	if err != nil {
		t.Fatalf("Create() after a panicking Validate = %v, want the cluster free", err)
	}
	awaitTerminal(t, m, next.ID)
}

// An ordinary refusal is unchanged by the boundary: a module that answered
// "no" is still a bad request, not a daemon failure.
func TestCreateStillReportsAnOrdinaryRefusalAsARefusal(t *testing.T) {
	module := newFake(Type("refuses"))
	module.validateErr = errors.New("this module never bakes almond")
	c := newCluster(master("master-1", "10.0.0.1"))
	m, _ := newManagerWith(t, c, registryWith(t, module))

	_, err := createFake(t, m, module.typ, "client-1")

	var refused *ModuleValidationError
	if !errors.As(err, &refused) {
		t.Fatalf("Create() = %v, want the module's own refusal", err)
	}
	if errors.Is(err, ErrModuleFailed) {
		t.Error("a module that refused was reported as one that could not answer")
	}
}

// phasePanicDetail is what a module panics with while being asked what the
// cluster looks like.
const phasePanicDetail = "boom: phase panic detail"

// panickingPhaseModule runs until the test lets it finish, so there is a live
// operation to ask about, and panics whenever it is asked what that operation
// makes the cluster look like.
type panickingPhaseModule struct {
	fakeModule
	typ     Type
	release chan struct{}
}

func (p *panickingPhaseModule) Type() Type { return p.typ }

func (p *panickingPhaseModule) Phase(Operation) (nodestatus.Phase, bool) { panic(phasePanicDetail) }

func (p *panickingPhaseModule) Run(context.Context, Runtime, RunRequest) Outcome {
	<-p.release
	return Outcome{Status: StatusSucceeded}
}

// The cluster summary asks the operation in flight what phase it imposes. A
// module that panics there imposes nothing — the summary keeps whatever phase
// it already had — and the request that asked for it is still answered.
func TestActivePhaseContainsAPanicInPhase(t *testing.T) {
	module := &panickingPhaseModule{typ: Type("cannot-say"), release: make(chan struct{})}
	c := newCluster(master("master-1", "10.0.0.1"))
	m, _ := newManagerWith(t, c, registryWith(t, module))

	op, err := createFake(t, m, module.typ, "client-1")
	if err != nil {
		t.Fatalf("Create: %v", err)
	}
	t.Cleanup(func() {
		close(module.release)
		awaitTerminal(t, m, op.ID)
	})

	phase, ok := m.ActivePhase()

	if ok || phase != "" {
		t.Fatalf("ActivePhase() = %q,%v, want nothing imposed by a module that could not answer", phase, ok)
	}
}

// The operation itself is untouched by a module that could not describe it:
// describing a cluster is a read, and a failed read settles nothing.
func TestAPanicInPhaseDoesNotSettleTheOperation(t *testing.T) {
	module := &panickingPhaseModule{typ: Type("cannot-say"), release: make(chan struct{})}
	c := newCluster(master("master-1", "10.0.0.1"))
	m, _ := newManagerWith(t, c, registryWith(t, module))

	op, err := createFake(t, m, module.typ, "client-1")
	if err != nil {
		t.Fatalf("Create: %v", err)
	}
	m.ActivePhase()

	got, ok := m.Get(op.ID)
	if !ok {
		t.Fatal("the operation is gone")
	}
	if got.Status.Terminal() {
		t.Fatalf("status = %q, want the operation still in flight", got.Status)
	}

	close(module.release)
	settled := awaitTerminal(t, m, op.ID)
	if settled.Status != StatusSucceeded {
		t.Errorf("status = %q, want the operation to finish normally", settled.Status)
	}
}

// SafeValidate is the one place a module is asked to judge a request. The
// node endpoint and the manager both go through it, so the two cannot answer
// a module that panicked differently from each other.
func TestSafeValidateReportsWhetherTheModuleAnswered(t *testing.T) {
	refusal := errors.New("this module never bakes almond")
	for _, tc := range []struct {
		name         string
		module       OperationModule
		wantRefusal  error
		wantAnswered bool
	}{
		{"accepts", newFake(Type("accepts")), nil, true},
		{"refuses", &fakeModule{typ: Type("refuses"), validateErr: refusal}, refusal, true},
		{"panics", &panickingValidateModule{typ: Type("panics")}, nil, false},
	} {
		t.Run(tc.name, func(t *testing.T) {
			got, answered := SafeValidate(tc.module, CreateRequest{
				Type: tc.module.Type(), RequestID: "client-1", Owner: "alice@olares.com",
			})

			if answered != tc.wantAnswered {
				t.Fatalf("answered = %v, want %v", answered, tc.wantAnswered)
			}
			if !errors.Is(got, tc.wantRefusal) {
				t.Errorf("refusal = %v, want %v", got, tc.wantRefusal)
			}
		})
	}
}

// A settled operation is not asked about at all, so a panicking Phase there
// is not even reached — but the framework must not depend on that: the record
// carries on saying what it settled as.
func TestAPanicInPhaseLeavesASettledRecordReadable(t *testing.T) {
	module := &panickingPhaseModule{typ: Type("cannot-say"), release: make(chan struct{})}
	close(module.release)
	c := newCluster(master("master-1", "10.0.0.1"))
	m, _ := newManagerWith(t, c, registryWith(t, module))

	op, err := createFake(t, m, module.typ, "client-1")
	if err != nil {
		t.Fatalf("Create: %v", err)
	}
	settled := awaitTerminal(t, m, op.ID)
	// Past the grace window a command_issued record would have held, so
	// nothing is active any more however the clock has moved.
	time.Sleep(time.Millisecond)

	if phase, ok := m.ActivePhase(); ok {
		t.Errorf("ActivePhase() = %q, want nothing once the operation is over", phase)
	}
	if got, ok := m.Get(op.ID); !ok || got.Status != settled.Status {
		t.Errorf("the settled record changed: %+v", got)
	}
}
