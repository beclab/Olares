package clierr

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
)

// yes is addressable so it can be handed to envelope.Retryable, which is
// a pointer precisely so that "we do not know" stays distinct from "no".
var yes = true

// A caller that asked for `-o json` asked to be read by a program, and
// then a failure hands it `Error: POST /v1/chat/completions: HTTP 503
// (model_not_ready): ...`. Everything it needs is in that line and none
// of it is addressable: whether to wait or give up, what to run next,
// which refusal this actually was. So the same invocations that answer
// JSON on success answer JSON on failure.
//
// It goes to stderr, not stdout. Some commands stream, so stdout may
// already hold a partial document, and a caller redirecting stdout to a
// file is not asking for the error to land in it.

// Structured is implemented by an error that already knows its own
// machine-readable shape. The dependency points this way on purpose:
// this package still imports nothing of the CLI's, and each subtree
// classifies its own failures, since only it knows what its codes mean.
type Structured interface {
	error

	// ErrorCode is the stable identifier a caller can branch on.
	ErrorCode() string

	// Retryable answers whether running the same command again could
	// succeed. Nil is the honest answer where the error does not know:
	// the field is then absent rather than guessed at as false.
	Retryable() *bool

	// RecoveryAction is one concrete next step, or empty when there is
	// none worth naming. Never a restatement of the message.
	RecoveryAction() string
}

// UnclassifiedCode is what an error that does not implement Structured
// reports. Naming it is better than an empty string, which reads as a
// code the producer forgot to set.
const UnclassifiedCode = "unclassified"

type envelope struct {
	Code      string `json:"code"`
	Message   string `json:"message"`
	Retryable *bool  `json:"retryable,omitempty"`
	Action    string `json:"action,omitempty"`
}

// WriteEnvelope renders err as `{"error": {...}}`. It reports whether
// anything was written, so a caller can fall back to the plain line.
func WriteEnvelope(w io.Writer, err error) bool {
	if err == nil {
		return false
	}
	body := envelope{Code: UnclassifiedCode, Message: err.Error()}
	switch structured, ok := AsStructured(err); {
	case ok:
		if code := structured.ErrorCode(); code != "" {
			body.Code = code
		}
		body.Retryable = structured.Retryable()
		body.Action = structured.RecoveryAction()
	case errors.Is(err, context.DeadlineExceeded):
		// Handled here rather than by each producer because there is no
		// producer to fix: this is the stdlib's error, returned from
		// wherever the context happened to be checked, and any tree that
		// passes a context can surface it. It is also the one failure a
		// caller most often wants to distinguish, since a timeout says
		// nothing about whether the request was refused or merely slow.
		body.Code = CodeTimeout
		body.Retryable = &yes
	}
	encoded, marshalErr := json.MarshalIndent(map[string]envelope{"error": body}, "", "  ")
	if marshalErr != nil {
		return false
	}
	if _, writeErr := fmt.Fprintln(w, string(encoded)); writeErr != nil {
		return false
	}
	return true
}

// AsStructured unwraps to the first error in the tree that classifies
// itself, so wrapping a classified error with context does not lose the
// classification.
//
// errors.As rather than a hand-rolled loop over Unwrap() error: an
// error joined from several — which `fmt.Errorf("%w: %w", ...)` and
// errors.Join both produce — unwraps to a slice, and a loop that only
// knows the single-error form walks straight past the classification it
// was looking for.
func AsStructured(err error) (Structured, bool) {
	var structured Structured
	if errors.As(err, &structured) {
		return structured, true
	}
	return nil, false
}
