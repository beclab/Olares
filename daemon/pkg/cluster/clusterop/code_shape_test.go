package clusterop

import (
	"context"
	"encoding/json"
	"strings"
	"testing"
	"time"
)

// A code is the one thing a module writes that a caller is meant to act on:
// it is persisted, returned over HTTP, and matched against by user-service
// and TermiPass. A module chooses it freely, so what a record can hold has to
// be checked here — a code carrying a sentence, a stack trace or a megabyte
// of text is not a code, and neither is one nothing could match on.

// unusableCodes are shapes no caller could act on. Each stands for something
// a module might put in a code field by accident or on purpose: a whole
// sentence, punctuation a matcher would have to escape, text that would break
// a log line or a JSON reader's assumptions, and something far too long to be
// an identifier.
func unusableCodes() map[string]string {
	return map[string]string{
		"a sentence":     "the oven was already on fire",
		"punctuation":    "module.failed!",
		"upper case":     "ModuleFailed",
		"a newline":      "module_failed\nlevel=fatal",
		"leading space":  " module_failed",
		"too long":       strings.Repeat("a", 65),
		"not ascii":      "模块失败",
		"a json snippet": `{"code":"module_failed"}`,
	}
}

// An operation settles on the code its module reported. One nothing could act
// on is recorded as module_failed instead — what is actually known — with the
// reviewed sentence that goes with it, and the module's own text nowhere near
// the record.
func TestAnOperationKeepsOnlyACodeShapedLikeACode(t *testing.T) {
	for name, code := range unusableCodes() {
		t.Run(name, func(t *testing.T) {
			rt, m, id := newTestRuntime(t, StatusRunning)

			if err := rt.Complete(Outcome{Status: StatusFailed, Code: code}); err != nil {
				t.Fatalf("Complete: %v", err)
			}

			got, _ := m.Get(id)
			if got.Code != CodeModuleFailed {
				t.Fatalf("Code = %q, want %s", got.Code, CodeModuleFailed)
			}
			if got.Error != reasonFor(CodeModuleFailed) {
				t.Errorf("Error = %q, want the reviewed sentence %q", got.Error, reasonFor(CodeModuleFailed))
			}
		})
	}
}

// What a module put in an unusable code reaches neither the record a caller
// reads nor the file it is saved to. A code is text this package did not
// write, and a module that puts a token or an address in one must not have
// found a way to persist it.
func TestNothingOfAnUnusableCodeIsKeptOnTheRecord(t *testing.T) {
	const leak = "token=super-secret"
	rt, m, id := newTestRuntime(t, StatusRunning)

	if err := rt.Complete(Outcome{Status: StatusFailed, Code: "module_failed " + leak}); err != nil {
		t.Fatalf("Complete: %v", err)
	}

	got, _ := m.Get(id)
	raw, err := json.Marshal(got)
	if err != nil {
		t.Fatalf("marshal: %v", err)
	}
	if strings.Contains(string(raw), leak) {
		t.Errorf("the record kept what the module put in its code: %s", raw)
	}
}

// A module's own code is not second-guessed for being unfamiliar: an
// operation this package knows nothing about settles on whatever stable code
// it declared, exactly as written.
func TestACodeAModuleDeclaredForItselfIsKeptAsWritten(t *testing.T) {
	for _, code := range []string{"bake_cake_failed", "oven_too_cold", "a", strings.Repeat("a", 64), "code_2"} {
		t.Run(code, func(t *testing.T) {
			rt, m, id := newTestRuntime(t, StatusRunning)

			if err := rt.Complete(Outcome{Status: StatusFailed, Code: code}); err != nil {
				t.Fatalf("Complete: %v", err)
			}

			got, _ := m.Get(id)
			if got.Code != code {
				t.Errorf("Code = %q, want the module's own %q", got.Code, code)
			}
		})
	}
}

// Every code this package has ever published is one a record can still hold.
// The check exists to keep unusable text out, not to break the contract
// callers already read.
func TestTheStableCodesAreAllUsable(t *testing.T) {
	stable := []string{
		CodeInventoryUnavailable, CodeNoMasterNode, CodeNodeUnaddressable, CodeNodeUnreachable,
		CodePowerUnsupported, CodeDispatchFailed, CodeNodeDidNotGoDown, CodeRestartTimeout,
		CodeHostPowerFailed, CodePrecheckFailed, CodeWorkerCommandFailed, CodeWorkerRestartFailed,
		CodeStatePersistenceFailed, CodeRequestInProgress, CodeUnsupportedTopology,
		CodeSelfUnresolved, CodeNodeIdentityUnknown, CodeNodeNotReady, CodeBootIDUnavailable,
		CodeModuleFailed, CodeDaemonRestarted, CodeUnsupportedOperation, CodeInvalidRequest,
		CodeSignatureUnbound, CodeSignatureMismatch, CodeSignatureExpired,
	}
	for _, code := range stable {
		if safeCode(code) != code {
			t.Errorf("safeCode(%q) = %q, want the published code unchanged", code, safeCode(code))
		}
	}
}

// A stage is reported the same way the operation is. A step carrying an
// unusable code would be read by the same callers, out of the same record.
func TestAStepKeepsOnlyACodeShapedLikeACode(t *testing.T) {
	rt, m, id := newTestRuntime(t, StatusRunning)
	if err := rt.StartStep("work"); err != nil {
		t.Fatalf("StartStep: %v", err)
	}

	if err := rt.FinishStep("work", StepFailed, "the oven was already on fire", "detail"); err != nil {
		t.Fatalf("FinishStep: %v", err)
	}

	got, _ := m.Get(id)
	step := findStep(&got, "work")
	if step == nil {
		t.Fatal("the step is gone")
	}
	if step.Code != CodeModuleFailed {
		t.Fatalf("Code = %q, want %s", step.Code, CodeModuleFailed)
	}
	if step.Error != reasonFor(CodeModuleFailed) {
		t.Errorf("Error = %q, want the reviewed sentence for %s", step.Error, CodeModuleFailed)
	}
}

// So is a node result, whether the module wrote the code when it listed the
// nodes or when it later reported what happened to one.
func TestANodeKeepsOnlyACodeShapedLikeACode(t *testing.T) {
	rt, m, id := newTestRuntime(t, StatusRunning)

	if err := rt.InitNodes([]NodeResult{{
		NodeName: "node-a", Status: NodeFailed, Code: "the oven was already on fire", Error: "detail",
	}}); err != nil {
		t.Fatalf("InitNodes: %v", err)
	}
	got, _ := m.Get(id)
	if node := findNode(&got, "node-a"); node == nil || node.Code != CodeModuleFailed {
		t.Fatalf("InitNodes code = %+v, want %s", node, CodeModuleFailed)
	}

	if err := rt.UpdateNode("node-a", func(n *NodeResult) {
		n.Code = "still on fire, now with a stack trace"
		n.Error = "detail"
	}); err != nil {
		t.Fatalf("UpdateNode: %v", err)
	}

	got, _ = m.Get(id)
	node := findNode(&got, "node-a")
	if node == nil {
		t.Fatal("the node is gone")
	}
	if node.Code != CodeModuleFailed {
		t.Fatalf("Code = %q, want %s", node.Code, CodeModuleFailed)
	}
	if node.Error != reasonFor(CodeModuleFailed) {
		t.Errorf("Error = %q, want the reviewed sentence for %s", node.Error, CodeModuleFailed)
	}
}

// The one settlement that writes a stage, its nodes and the outcome together
// is held to the same rule: a confirmation written as one change must not be
// the way an unusable code reaches the record.
func TestAnAtomicSettlementKeepsOnlyACodeShapedLikeACode(t *testing.T) {
	_, m, id := buildCommandIssuedRuntime(t, time.Hour)
	recovery := newRecoveryRuntime(m, id, context.Background())

	err := recovery.Settle(settlement{
		step:    stepSettlement{name: StepMasterCommand, status: StepFailed},
		nodes:   []nodeSettlement{{name: "node-a", from: NodePending, to: NodeFailed}},
		outcome: failedWith("the oven was already on fire", "a sentence nobody reviewed"),
	})
	if err != nil {
		t.Fatalf("Settle: %v", err)
	}

	got, _ := m.Get(id)
	if got.Code != CodeModuleFailed {
		t.Fatalf("Code = %q, want %s", got.Code, CodeModuleFailed)
	}
	if got.Error != reasonFor(CodeModuleFailed) {
		t.Errorf("Error = %q, want the reviewed sentence for %s", got.Error, CodeModuleFailed)
	}
}

// An operation that settles with no code at all keeps no message either,
// which is what already distinguishes "succeeded" from "failed with
// something the record could not keep".
func TestAnOutcomeWithNoCodeStillKeepsNothing(t *testing.T) {
	rt, m, id := newTestRuntime(t, StatusRunning)

	if err := rt.Complete(Outcome{Status: StatusSucceeded}); err != nil {
		t.Fatalf("Complete: %v", err)
	}

	got, _ := m.Get(id)
	if got.Code != "" || got.Error != "" {
		t.Errorf("code = %q error = %q, want nothing kept", got.Code, got.Error)
	}
}
