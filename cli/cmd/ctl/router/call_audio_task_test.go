package router

import (
	"io"
	"strings"
	"testing"
)

// `task list` used to reach Router without --model, be refused with
// model_required, and be told "Router does not remember task  —" with nothing
// where the id belongs: it follows no task. Its own help said the flag was
// "only needed when Router no longer remembers the task", while its Long text
// said the opposite two paragraphs earlier.
//
// A queue belongs to one engine and there is no view across them, so the flag
// is required at parse time and the round trip never happens.
func TestTaskListRefusesToRunWithoutAModel(t *testing.T) {
	root := NewRouterCommand(nil)
	root.SetOut(io.Discard)
	root.SetErr(io.Discard)
	root.SetArgs([]string{"call", "task", "list"})
	err := root.Execute()
	if err == nil {
		t.Fatal("`call task list` ran with no --model")
	}
	if !strings.Contains(err.Error(), "model") {
		t.Errorf("the refusal does not name the missing flag: %v", err)
	}
}

// The same Router code reaches two different readers, and only one of them has
// an id to be told about.
func TestModelRequiredReadsDifferentlyForAListAndForATask(t *testing.T) {
	refusal := &RouterError{Method: "GET", Path: "/v1/tasks", Status: 400, Code: "model_required"}

	list := audioTaskErr(refusal, "").Error()
	if strings.Contains(list, "remember task") {
		t.Errorf("the list refusal talks about a task it does not have: %q", list)
	}
	if !strings.Contains(list, "--model") {
		t.Errorf("the list refusal does not name the flag to pass: %q", list)
	}

	one := audioTaskErr(refusal, "tsk_123").Error()
	if !strings.Contains(one, "tsk_123") {
		t.Errorf("the single-task refusal drops the id: %q", one)
	}
}
