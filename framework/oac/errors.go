package oac

import (
	"errors"
	"fmt"
	"sort"
	"strings"

	validation "github.com/go-ozzo/ozzo-validation/v4"
)

// errAppConfigUnavailable is returned by LoadAppConfiguration when the
// underlying Strategy produced a Manifest whose Raw() is not
// *AppConfiguration. With the current single-strategy setup this is
// unreachable in practice, but keeping it as a sentinel guards against
// silent type drift if a future Strategy starts returning a different
// concrete type.
var errAppConfigUnavailable = errors.New("oac: manifest is not backed by *AppConfiguration")

// errAppConfigNil is reported by (*Checker).ValidateAppConfiguration when the
// caller passes a nil pointer. Wrapping it in a *ValidationError (via
// WrapValidation) keeps the error model uniform with every other
// Validate* entry point.
var errAppConfigNil = errors.New("oac: AppConfiguration is nil")

// ValidationError describes a single failed manifest validation.
//
// Callers should use errors.As to pull this out of the error chain returned by
// ValidateManifestFile / ValidateManifestContent / LintChart and friends.
type ValidationError struct {
	// Version is the manifest version (apiVersion) that produced the error,
	// e.g. "v1" or "v2".
	Version string
	// Field is the dotted field path that failed, e.g. "metadata.name".
	Field string
	// Reason is a short human-readable explanation.
	Reason string
	// Inner is the underlying error if any (typically validation.Errors from ozzo).
	Inner error
}

func (e *ValidationError) Error() string {
	var b strings.Builder
	b.WriteString("manifest validation failed")
	if e.Version != "" {
		b.WriteString(" (apiVersion=")
		b.WriteString(e.Version)
		b.WriteString(")")
	}
	// Field is skipped when Reason already opens with it, so a single-field
	// failure names its path once. WrapValidation sets Field to the first
	// offending path, and most validators in internal/manifest spell the
	// full path inside their own message ("spec.resources[0].limitedCpu is
	// required to declare ..."), which used to print as
	// "metadata.icon: metadata.icon: metadata.icon is required".
	if e.Field != "" && !reasonNamesField(e.Reason, e.Field) {
		b.WriteString(": ")
		b.WriteString(e.Field)
	}
	if e.Reason != "" {
		b.WriteString(": ")
		b.WriteString(e.Reason)
	}
	return b.String()
}

// reasonNamesField reports whether reason already opens with field, so that
// prefixing it again would repeat the path. The character after the match
// must be a separator: otherwise field "spec" would suppress the prefix on a
// reason like "specification is malformed".
func reasonNamesField(reason, field string) bool {
	if field == "" || !strings.HasPrefix(reason, field) {
		return false
	}
	rest := reason[len(field):]
	return rest == "" || rest[0] == ':' || rest[0] == ' '
}

func (e *ValidationError) Unwrap() error { return e.Inner }

// NewValidationError constructs a single-field ValidationError.
func NewValidationError(version, field, reason string) *ValidationError {
	return &ValidationError{Version: version, Field: field, Reason: reason}
}

// WrapValidation converts an ozzo validation.Errors map into a sorted, stable,
// multi-line ValidationError. If err is nil it returns nil. If err is not a
// validation.Errors it is wrapped as a single-message ValidationError.
func WrapValidation(version string, err error) error {
	if err == nil {
		return nil
	}
	var verrs validation.Errors
	if errors.As(err, &verrs) {
		fields := flattenValidationErrors("", verrs)
		if len(fields) == 0 {
			return &ValidationError{Version: version, Reason: err.Error(), Inner: err}
		}
		sort.Slice(fields, func(i, j int) bool { return fields[i].path < fields[j].path })
		var lines []string
		for _, f := range fields {
			// Validators in internal/manifest spell the full path in their
			// own message; prefixing those would say it twice. Ozzo's
			// built-ins ("cannot be blank") do not, and need the prefix to
			// be actionable at all.
			if reasonNamesField(f.reason, f.path) {
				lines = append(lines, f.reason)
				continue
			}
			lines = append(lines, fmt.Sprintf("%s: %s", f.path, f.reason))
		}
		return &ValidationError{
			Version: version,
			Field:   fields[0].path,
			Reason:  strings.Join(lines, "; "),
			Inner:   err,
		}
	}
	return &ValidationError{Version: version, Reason: err.Error(), Inner: err}
}

type flatField struct {
	path   string
	reason string
}

func flattenValidationErrors(prefix string, errs validation.Errors) []flatField {
	var out []flatField
	for k, v := range errs {
		path := k
		if prefix != "" {
			path = prefix + "." + k
		}
		var nested validation.Errors
		if errors.As(v, &nested) {
			out = append(out, flattenValidationErrors(path, nested)...)
			continue
		}
		out = append(out, flatField{path: path, reason: v.Error()})
	}
	return out
}

// AggregateErrors combines multiple errors into one. Returns nil when the input
// is empty or all entries are nil. The input slice is not modified.
func AggregateErrors(errs []error) error {
	if len(errs) == 0 {
		return nil
	}
	filtered := make([]error, 0, len(errs))
	for _, e := range errs {
		if e != nil {
			filtered = append(filtered, e)
		}
	}
	if len(filtered) == 0 {
		return nil
	}
	if len(filtered) == 1 {
		return filtered[0]
	}
	return errors.Join(filtered...)
}

// ErrNotImplemented is returned by manifest strategies whose validation logic
// has not been implemented yet (e.g. v2 scaffold).
var ErrNotImplemented = errors.New("not implemented")
