package clusterop

import (
	"context"
	"errors"
	"strings"
	"testing"
)

// nodeRequestFor is one node-local command, as a master's fan-out puts it on
// the wire.
func nodeRequestFor(typ Type) NodeRequest {
	return NodeRequest{PeerRequest: PeerRequest{
		Type: typ, OperationID: "op-1", RequestID: "client-1",
	}}
}

// moduleDetail is what a module says when it fails. It stands for an address,
// a token, or the text of some lower-level error: this package did not write
// it and cannot vouch for it, so no test may find it in something a caller
// receives.
const moduleDetail = "token=super-secret from a module nobody reviewed"

// The two power operations are the ones this package implements itself, and
// they are exactly the ones the legacy endpoint serves. Anything else that
// registers itself later is not, and does not become one by existing.
func TestOnlyTheBuiltInPowerOperationsAreServedByTheLegacyEndpoint(t *testing.T) {
	for _, typ := range []Type{TypeReboot, TypeShutdown} {
		if !isBuiltInPowerOperation(typ) {
			t.Errorf("%q is not served by /command/power-node, but old masters send it there", typ)
		}
	}
	if isBuiltInPowerOperation(Type("bake-cake")) {
		t.Error("an operation this package never wrote is served by the legacy endpoint")
	}

	// The two modules say so themselves, rather than being named in a list
	// somewhere else that could drift from them.
	for _, typ := range []Type{TypeReboot, TypeShutdown} {
		module, ok := DefaultRegistry().Lookup(typ)
		if !ok {
			t.Fatalf("the built-in %q module is not registered", typ)
		}
		if _, ok := module.(builtInPowerOperation); !ok {
			t.Errorf("the %q module does not declare itself a built-in power operation", typ)
		}
	}
}

// The legacy power endpoint predates module validation: it hands a request
// straight to a module without ever asking whether the module accepts it. So
// it may only reach operations whose node half this package wrote. A module
// registered later is refused before it is called at all — otherwise that
// endpoint would be a way around the Validate the generic one performs.
func TestExecutePowerNodeRefusesAnOperationItDoesNotServe(t *testing.T) {
	module := &nodeCapableModule{typ: Type("bake-cake")}
	reg := registryWith(t, module)

	err := ExecutePowerNode(context.Background(), reg, nodeRequestFor(module.typ))

	if powerCode(t, err) != CodeUnsupportedOperation {
		t.Fatalf("ExecutePowerNode() code = %q, want %s", powerCode(t, err), CodeUnsupportedOperation)
	}
	if module.callCount() != 0 {
		t.Errorf("an operation the legacy endpoint does not serve was carried out %d times",
			module.callCount())
	}
}

func TestExecutePowerNodeCarriesOutABuiltInPowerOperation(t *testing.T) {
	for _, typ := range []Type{TypeReboot, TypeShutdown} {
		t.Run(string(typ), func(t *testing.T) {
			module := &nodeCapableModule{typ: typ}
			reg := registryWith(t, module)

			if err := ExecutePowerNode(context.Background(), reg, nodeRequestFor(typ)); err != nil {
				t.Fatalf("ExecutePowerNode(%q): %v", typ, err)
			}
			if module.callCount() != 1 {
				t.Errorf("ExecutePowerNode(%q) calls = %d, want 1", typ, module.callCount())
			}
		})
	}
}

// What a power operation refuses with is text this package wrote and
// reviewed, and callers have been reading those codes since before cluster
// operations existed. It reaches them unchanged.
func TestExecutePowerNodeKeepsABuiltInRefusalExactlyAsItWasWritten(t *testing.T) {
	refusal := &PowerError{
		Code:    CodePowerUnsupported,
		Message: "olaresd runs in a container on this node, so it cannot power the machine",
	}
	module := &nodeCapableModule{typ: TypeReboot, err: refusal}
	reg := registryWith(t, module)

	err := ExecutePowerNode(context.Background(), reg, nodeRequestFor(TypeReboot))

	var pe *PowerError
	if !errors.As(err, &pe) {
		t.Fatalf("ExecutePowerNode() = %v, want the operation's own refusal", err)
	}
	if pe.Code != refusal.Code || pe.Message != refusal.Message {
		t.Errorf("refusal = %q/%q, want %q/%q", pe.Code, pe.Message, refusal.Code, refusal.Message)
	}
}

// A module chooses its own error, including its code and its message, and
// both reach an HTTP caller if they are passed along. A module could
// therefore pick a code that means something else here, or put an address or
// a token in a message the caller reads. What it says goes to this node's
// log; what goes back is the one stable code that describes what is actually
// known — that the module failed.
func TestExecuteNodeDoesNotRepeatWhatAModuleFailedWith(t *testing.T) {
	for _, tc := range []struct {
		name string
		err  error
	}{
		{"a plain error", errors.New(moduleDetail)},
		{"an error dressed as this package's own", &PowerError{
			Code:    CodePowerUnsupported,
			Message: moduleDetail,
		}},
	} {
		t.Run(tc.name, func(t *testing.T) {
			module := &nodeCapableModule{typ: Type("bake-cake"), err: tc.err}
			reg := registryWith(t, module)

			err := ExecuteNode(context.Background(), reg, nodeRequestFor(module.typ))

			if powerCode(t, err) != CodeModuleFailed {
				t.Errorf("ExecuteNode() code = %q, want %s", powerCode(t, err), CodeModuleFailed)
			}
			if strings.Contains(err.Error(), moduleDetail) {
				t.Errorf("ExecuteNode() = %v, repeated what the module said", err)
			}
			if module.callCount() != 1 {
				t.Errorf("ExecuteNode calls = %d, want 1", module.callCount())
			}
		})
	}
}

// Sanitizing a module's failure must not sanitize this package's own
// refusals: a type nothing holds is still reported as one.
func TestExecuteNodeStillReportsAnUnknownTypeAsUnsupported(t *testing.T) {
	reg := registryWith(t, &nodeCapableModule{typ: Type("bake-cake")})

	err := ExecuteNode(context.Background(), reg, nodeRequestFor(Type("no-such-type")))

	if powerCode(t, err) != CodeUnsupportedOperation {
		t.Fatalf("ExecuteNode() code = %q, want %s", powerCode(t, err), CodeUnsupportedOperation)
	}
}

// The generic helper carries the built-in operations too — the master's own
// node runs one through it — and their refusals are this package's own text,
// so those still reach the caller unchanged.
func TestExecuteNodeKeepsABuiltInPowerRefusal(t *testing.T) {
	refusal := &PowerError{Code: CodeHostPowerFailed, Message: "this node could not be powered"}
	module := &nodeCapableModule{typ: TypeShutdown, err: refusal}
	reg := registryWith(t, module)

	err := ExecuteNode(context.Background(), reg, nodeRequestFor(TypeShutdown))

	if powerCode(t, err) != CodeHostPowerFailed {
		t.Fatalf("ExecuteNode() code = %q, want %s", powerCode(t, err), CodeHostPowerFailed)
	}
}
