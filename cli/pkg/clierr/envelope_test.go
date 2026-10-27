package clierr

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"testing"
)

type classified struct {
	code      string
	retryable *bool
	action    string
}

func (c classified) Error() string          { return "the call was refused" }
func (c classified) ErrorCode() string      { return c.code }
func (c classified) Retryable() *bool       { return c.retryable }
func (c classified) RecoveryAction() string { return c.action }

func decode(t *testing.T, raw []byte) map[string]any {
	t.Helper()
	var document map[string]map[string]any
	if err := json.Unmarshal(raw, &document); err != nil {
		t.Fatalf("the envelope is not JSON, which is the one thing it had to be: %v\n%s", err, raw)
	}
	body, ok := document["error"]
	if !ok {
		t.Fatalf("no `error` key: %s", raw)
	}
	return body
}

func TestAClassifiedErrorCarriesItsCodeAndNextStep(t *testing.T) {
	yes := true
	var out bytes.Buffer
	if !WriteEnvelope(&out, classified{code: "model_not_ready", retryable: &yes, action: "wait, then retry"}) {
		t.Fatal("nothing was written")
	}
	body := decode(t, out.Bytes())
	if body["code"] != "model_not_ready" {
		t.Fatalf("code = %v", body["code"])
	}
	if body["retryable"] != true {
		t.Fatalf("retryable = %v", body["retryable"])
	}
	if body["action"] != "wait, then retry" {
		t.Fatalf("action = %v", body["action"])
	}
}

// An unknown answer has to read as unknown. Defaulting `retryable` to
// false would tell a caller to give up on a failure nothing examined.
func TestWhatIsNotKnownIsAbsentRatherThanFalse(t *testing.T) {
	var out bytes.Buffer
	WriteEnvelope(&out, classified{code: "upstream_error"})
	body := decode(t, out.Bytes())
	if _, present := body["retryable"]; present {
		t.Fatalf("retryable was invented: %v", body["retryable"])
	}
	if _, present := body["action"]; present {
		t.Fatalf("action was invented: %v", body["action"])
	}
}

func TestAnUnclassifiedErrorStillParses(t *testing.T) {
	var out bytes.Buffer
	WriteEnvelope(&out, errors.New("connection refused"))
	body := decode(t, out.Bytes())
	if body["code"] != UnclassifiedCode {
		t.Fatalf("code = %v, want %q", body["code"], UnclassifiedCode)
	}
	if body["message"] != "connection refused" {
		t.Fatalf("message = %v", body["message"])
	}
}

// Adding context to an error is the normal thing to do on the way up,
// and it must not cost the classification underneath.
func TestWrappingDoesNotLoseTheClassification(t *testing.T) {
	no := false
	wrapped := fmt.Errorf("send the chat request: %w", classified{code: "quota_exceeded", retryable: &no})
	var out bytes.Buffer
	WriteEnvelope(&out, wrapped)
	body := decode(t, out.Bytes())
	if body["code"] != "quota_exceeded" {
		t.Fatalf("code = %v", body["code"])
	}
	if body["retryable"] != false {
		t.Fatalf("retryable = %v", body["retryable"])
	}
	if body["message"] != "send the chat request: the call was refused" {
		t.Fatalf("the wrapping context was dropped: %v", body["message"])
	}
}

// A deadline is the stdlib's error and belongs to no producer, so it is
// classified here or nowhere. The wrapped case is the realistic one: by
// the time a deadline reaches the entrypoint it has been through the
// layer that was waiting when it fired.
func TestADeadlineIsClassifiedEvenThoughNobodyOwnsIt(t *testing.T) {
	for _, err := range []error{
		context.DeadlineExceeded,
		fmt.Errorf("poll the install: %w", context.DeadlineExceeded),
	} {
		var out bytes.Buffer
		WriteEnvelope(&out, err)
		body := decode(t, out.Bytes())
		if body["code"] != CodeTimeout {
			t.Fatalf("%v: code = %v, want %q", err, body["code"], CodeTimeout)
		}
		if body["retryable"] != true {
			t.Fatalf("%v: a deadline is the one failure worth retrying, got retryable = %v", err, body["retryable"])
		}
	}
}

// An error that classifies itself wins over the deadline check, so a
// tree that gives a timeout a better name than "timeout" keeps it.
func TestAProducersOwnClassificationOutranksTheDeadlineFallback(t *testing.T) {
	var out bytes.Buffer
	WriteEnvelope(&out, fmt.Errorf("%w: %w", classified{code: "model_not_ready"}, context.DeadlineExceeded))
	if body := decode(t, out.Bytes()); body["code"] != "model_not_ready" {
		t.Fatalf("code = %v", body["code"])
	}
}

func TestNilIsNotAnError(t *testing.T) {
	var out bytes.Buffer
	if WriteEnvelope(&out, nil) {
		t.Fatal("nil produced an envelope")
	}
	if out.Len() != 0 {
		t.Fatalf("nil produced output: %q", out.String())
	}
}
