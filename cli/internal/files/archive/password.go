// password.go: cross-endpoint classification of archive
// password errors, so the cobra layer can re-prompt for a
// password and retry — mirroring TermiPass's flow:
//
//	// utils/interface/archive.ts
//	export const isArchivePasswordError = (error: any) =>
//	    [30001, 30002].includes(getArchiveErrorCode(error));
//
// The three archive endpoints report a missing / wrong password in
// two different shapes:
//
//   - POST compress / extract  → standard JSON body 4xx with a
//     NUMERIC code: {"code":30001,"message":"archive password
//     required"} / {"code":30002,...}. Surfaces as *HTTPError.
//   - GET entries (pre-walk open failure) → same standard JSON 4xx
//     body, also *HTTPError with the numeric code.
//   - GET entry → typed *EntryError whose Code is the STRING enum
//     (password_required / password_invalid).
//   - GET entries (mid-walk) → typed *EntriesStreamError with the
//     same string Code (defensive; password failures normally
//     happen pre-walk).
//
// ClassifyPasswordError folds all of them into one verdict.
package archive

import "encoding/json"

// Numeric `code` values the POST endpoints (and the entries
// pre-walk JSON error body) use for password problems. These are
// the same constants TermiPass keys on in isArchivePasswordError.
const (
	wireCodeArchivePasswordRequired = 30001
	wireCodeArchivePasswordInvalid  = 30002
)

// PasswordErrorKind is the verdict of ClassifyPasswordError.
type PasswordErrorKind int

const (
	// PasswordErrorNone means err is not a password problem (or is
	// nil) — the caller should NOT prompt for a password.
	PasswordErrorNone PasswordErrorKind = iota
	// PasswordErrorRequired means the archive is encrypted and no
	// (or an empty) password was supplied.
	PasswordErrorRequired
	// PasswordErrorInvalid means a password was supplied but the
	// server rejected it as wrong.
	PasswordErrorInvalid
)

// ClassifyPasswordError inspects err across every archive error
// shape and reports whether it is a password-required /
// password-incorrect condition. Returns PasswordErrorNone for nil
// or any non-password error.
//
// The string-coded typed errors (EntryError / EntriesStreamError)
// are checked before the generic *HTTPError branch because an
// EntryError unwraps to an *HTTPError whose body carries the
// STRING code — the numeric parse would simply miss it, but the
// explicit ordering keeps the intent obvious.
func ClassifyPasswordError(err error) PasswordErrorKind {
	if err == nil {
		return PasswordErrorNone
	}

	// GET entry — typed error with the string enum code.
	var entryErr *EntryError
	if asEntryError(err, &entryErr) {
		if k := passwordKindFromStringCode(entryErr.Code); k != PasswordErrorNone {
			return k
		}
	}

	// GET entries — mid-walk in-band sentinel (defensive).
	if se, ok := IsEntriesStreamError(err); ok {
		if k := passwordKindFromStringCode(se.Code); k != PasswordErrorNone {
			return k
		}
	}

	// POST compress / extract, and entries' pre-walk open failure:
	// standard JSON body with the numeric code.
	var he *HTTPError
	if asHTTPError(err, &he) {
		switch parseWireErrorCode(he.Body) {
		case wireCodeArchivePasswordRequired:
			return PasswordErrorRequired
		case wireCodeArchivePasswordInvalid:
			return PasswordErrorInvalid
		}
	}

	return PasswordErrorNone
}

// passwordKindFromStringCode maps the documented string enum codes
// (used by the entry / entries stream endpoints) onto the kind.
func passwordKindFromStringCode(code string) PasswordErrorKind {
	switch code {
	case CodePasswordRequired:
		return PasswordErrorRequired
	case CodePasswordInvalid:
		return PasswordErrorInvalid
	}
	return PasswordErrorNone
}

// parseWireErrorCode pulls the numeric `code` out of a JSON error
// body like {"code":30001,"message":"archive password required"}.
// Returns 0 when the body is empty, not JSON, or carries a
// non-numeric code (e.g. the string-coded entry error shape).
func parseWireErrorCode(body string) int {
	if body == "" {
		return 0
	}
	var env struct {
		Code *int `json:"code"`
	}
	if json.Unmarshal([]byte(body), &env) != nil || env.Code == nil {
		return 0
	}
	return *env.Code
}

// asEntryError unwraps err into *EntryError, walking the fmt.Errorf
// wrap chain. Mirrors asHTTPError's manual walk so this package
// stays free of the "errors" import for its predicates.
func asEntryError(err error, dst **EntryError) bool {
	for cur := err; cur != nil; {
		if e, ok := cur.(*EntryError); ok {
			*dst = e
			return true
		}
		type unwrapper interface{ Unwrap() error }
		u, ok := cur.(unwrapper)
		if !ok {
			return false
		}
		cur = u.Unwrap()
	}
	return false
}
