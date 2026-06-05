package archive

import (
	"fmt"
	"testing"
)

// TestClassifyPasswordError pins the cross-endpoint password-error
// recogniser the cobra layer keys on to re-prompt and retry —
// mirroring TermiPass's isArchivePasswordError([30001, 30002]).
func TestClassifyPasswordError(t *testing.T) {
	cases := []struct {
		name string
		err  error
		want PasswordErrorKind
	}{
		{"nil", nil, PasswordErrorNone},
		{
			"post numeric required",
			&HTTPError{Status: 400, Body: `{"code":30001,"message":"archive password required"}`},
			PasswordErrorRequired,
		},
		{
			"post numeric invalid",
			&HTTPError{Status: 400, Body: `{"code":30002,"message":"archive password incorrect"}`},
			PasswordErrorInvalid,
		},
		{
			"post numeric wrapped",
			fmt.Errorf("extract: %w", &HTTPError{Status: 400, Body: `{"code":30001}`}),
			PasswordErrorRequired,
		},
		{
			"post other numeric code is not a password error",
			&HTTPError{Status: 400, Body: `{"code":40004,"message":"archive corrupt"}`},
			PasswordErrorNone,
		},
		{
			"entries stream string required",
			&EntriesStreamError{Code: CodePasswordRequired, Message: "need pw"},
			PasswordErrorRequired,
		},
		{
			"entries stream string invalid",
			&EntriesStreamError{Code: CodePasswordInvalid, Message: "bad pw"},
			PasswordErrorInvalid,
		},
		{
			"entry string required",
			&EntryError{
				HTTPError: &HTTPError{Status: 401, Body: `{"error":"need pw","code":"password_required"}`},
				Code:      CodePasswordRequired,
				Message:   "need pw",
			},
			PasswordErrorRequired,
		},
		{
			"entry string invalid",
			&EntryError{
				HTTPError: &HTTPError{Status: 401, Body: `{"error":"bad pw","code":"password_invalid"}`},
				Code:      CodePasswordInvalid,
				Message:   "bad pw",
			},
			PasswordErrorInvalid,
		},
		{
			"entry not_found is not a password error",
			&EntryError{
				HTTPError: &HTTPError{Status: 404, Body: `{"error":"nope","code":"not_found"}`},
				Code:      CodeNotFound,
			},
			PasswordErrorNone,
		},
		{
			"plain 404 http error",
			&HTTPError{Status: 404, Body: `{"code":404,"message":"not found"}`},
			PasswordErrorNone,
		},
		{
			"non-json body",
			&HTTPError{Status: 500, Body: `<html>502 bad gateway</html>`},
			PasswordErrorNone,
		},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			if got := ClassifyPasswordError(tc.err); got != tc.want {
				t.Errorf("ClassifyPasswordError(%v) = %d, want %d", tc.err, got, tc.want)
			}
		})
	}
}

// TestParseWireErrorCode exercises the numeric-code extractor on
// the JSON shapes the POST endpoints emit, plus the degenerate
// inputs it must not panic on.
func TestParseWireErrorCode(t *testing.T) {
	cases := []struct {
		body string
		want int
	}{
		{`{"code":30001,"message":"x"}`, 30001},
		{`{"code":0}`, 0},
		{`{"message":"no code"}`, 0},
		{`{"code":"password_required"}`, 0}, // string code → not numeric
		{``, 0},
		{`not json`, 0},
	}
	for _, tc := range cases {
		if got := parseWireErrorCode(tc.body); got != tc.want {
			t.Errorf("parseWireErrorCode(%q) = %d, want %d", tc.body, got, tc.want)
		}
	}
}
