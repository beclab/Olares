package oac

import (
	"errors"
	"strings"
	"testing"

	validation "github.com/go-ozzo/ozzo-validation/v4"
)

func TestValidationError_Format(t *testing.T) {
	cases := []struct {
		name string
		err  *ValidationError
		want []string // substrings that must appear
	}{
		{
			name: "version+field+reason",
			err:  &ValidationError{Version: "v1", Field: "metadata.name", Reason: "is required"},
			want: []string{"manifest validation failed", "apiVersion=v1", "metadata.name", "is required"},
		},
		{
			name: "no field",
			err:  &ValidationError{Version: "v2", Reason: "boom"},
			want: []string{"manifest validation failed", "apiVersion=v2", "boom"},
		},
		{
			name: "no version",
			err:  &ValidationError{Field: "spec", Reason: "x"},
			want: []string{"manifest validation failed", "spec", "x"},
		},
	}
	for _, tc := range cases {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			got := tc.err.Error()
			for _, w := range tc.want {
				if !strings.Contains(got, w) {
					t.Errorf("error %q missing substring %q", got, w)
				}
			}
		})
	}
}

func TestValidationError_Unwrap(t *testing.T) {
	inner := errors.New("root cause")
	v := &ValidationError{Version: "v1", Reason: "wrapped", Inner: inner}
	if !errors.Is(v, inner) {
		t.Fatal("errors.Is must walk to Inner")
	}
}

func TestNewValidationError(t *testing.T) {
	v := NewValidationError("v1", "spec.name", "required")
	if v.Version != "v1" || v.Field != "spec.name" || v.Reason != "required" {
		t.Fatalf("unexpected fields: %+v", v)
	}
}

func TestWrapValidation_Nil(t *testing.T) {
	if err := WrapValidation("v1", nil); err != nil {
		t.Fatalf("nil in -> nil out, got %v", err)
	}
}

func TestWrapValidation_PlainError(t *testing.T) {
	plain := errors.New("ordinary")
	out := WrapValidation("v1", plain)
	var ve *ValidationError
	if !errors.As(out, &ve) {
		t.Fatalf("expected *ValidationError, got %T", out)
	}
	if ve.Version != "v1" || ve.Reason != "ordinary" {
		t.Fatalf("plain wrap: %+v", ve)
	}
	if !errors.Is(out, plain) {
		t.Fatal("must preserve original error in chain")
	}
}

func TestWrapValidation_FlattensValidationErrors(t *testing.T) {
	verrs := validation.Errors{
		"metadata": validation.Errors{
			"name":    errors.New("is required"),
			"version": errors.New("is required"),
		},
		"entrances": errors.New("must have at least one item"),
	}
	out := WrapValidation("v1", verrs)
	var ve *ValidationError
	if !errors.As(out, &ve) {
		t.Fatalf("expected *ValidationError, got %T", out)
	}
	msg := ve.Error()
	for _, want := range []string{"metadata.name", "metadata.version", "entrances"} {
		if !strings.Contains(msg, want) {
			t.Errorf("message missing %q: %s", want, msg)
		}
	}
	// Sorting must be deterministic.
	out2 := WrapValidation("v1", verrs)
	if out2.Error() != out.Error() {
		t.Fatalf("WrapValidation must produce deterministic output:\nfirst:  %s\nsecond: %s", out.Error(), out2.Error())
	}
}

func TestAggregateErrors(t *testing.T) {
	if err := AggregateErrors(nil); err != nil {
		t.Fatalf("nil slice -> nil, got %v", err)
	}
	if err := AggregateErrors([]error{nil, nil}); err != nil {
		t.Fatalf("only-nil slice -> nil, got %v", err)
	}
	one := errors.New("one")
	if err := AggregateErrors([]error{nil, one, nil}); err != one {
		t.Fatalf("single non-nil -> that error, got %v", err)
	}
	a := errors.New("a")
	b := errors.New("b")
	combined := AggregateErrors([]error{a, b})
	if combined == nil {
		t.Fatal("expected combined error")
	}
	if !errors.Is(combined, a) || !errors.Is(combined, b) {
		t.Fatalf("combined error must wrap both: %v", combined)
	}

	// Regression: AggregateErrors must not mutate the caller's slice. A
	// previous implementation used filtered := errs[:0], which aliased the
	// backing array and overwrote leading nil entries with later non-nil
	// errors during filtering.
	x := errors.New("x")
	y := errors.New("y")
	input := []error{nil, x, nil, y}
	snapshot := []error{input[0], input[1], input[2], input[3]}
	_ = AggregateErrors(input)
	for i := range input {
		if input[i] != snapshot[i] {
			t.Fatalf("AggregateErrors mutated caller slice at index %d: got %v, want %v", i, input[i], snapshot[i])
		}
	}
}
