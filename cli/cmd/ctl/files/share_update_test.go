package files

import (
	"strings"
	"testing"

	"github.com/beclab/Olares/cli/internal/files/share"
)

// TestRequireShareType_Mismatches pins the share-type rejection that
// every share-update verb (set-password / set-members / set-smb)
// runs after fetching the share record by id. Without this gate, a
// `set-password` against an internal share would land on the server
// as an opaque 4xx with no recovery path; the gate surfaces the
// type mismatch up front and points at the matching update verb.
//
// The test exercises three properties that ALL need to hold for the
// error to be useful to the user:
//
//  1. The wire `share_type` discriminator (e.g. `"external"` for
//     Public) is NEVER what the user sees first — friendly names
//     come from shareFlavorFriendlyName, so the error must say
//     "public" / "internal" / "smb".
//
//  2. The wire value IS still mentioned (in quotes) for support /
//     debugging, but as a secondary detail.
//
//  3. The recovery hint must list every update verb so the user
//     can map "I have an internal share" → `share set-members`,
//     etc., in one read of the error.
func TestRequireShareType_Mismatches(t *testing.T) {
	cases := []struct {
		name      string
		actual    share.Type
		want      share.Type
		verb      string
		shareID   string
		wantSubs  []string
		mustNot   []string
	}{
		{
			// set-password against an internal share is the most
			// common foot-gun (internal shares have no password,
			// the user might mis-target the verb). Anchor the
			// friendly name "internal" + the recovery pointer to
			// `share set-members`.
			name:    "set-password against internal",
			actual:  share.TypeInternal,
			want:    share.TypePublic,
			verb:    "set the password of",
			shareID: "abc-123",
			wantSubs: []string{
				"refusing to set the password of share abc-123",
				"the share is internal",
				`(wire type "internal")`,
				"not public",
				"`share set-password` for public shares",
				"`share set-members` for internal shares",
				"`share set-smb` for SMB shares",
			},
			mustNot: []string{
				// Critical: the wire value for Public is the
				// historically confusing "external" string. The
				// error must NOT lead with that, or the user
				// will hunt for a non-existent `share external`
				// command.
				"not external",
			},
		},
		{
			// set-members against an SMB share is the cross-flavor
			// case where Internal vs SMB members address by
			// different keys (name vs SMB-account ID). The error
			// must steer to set-smb so the user's `--users
			// smb-uid-1` invocation reaches the right endpoint.
			name:    "set-members against SMB",
			actual:  share.TypeSMB,
			want:    share.TypeInternal,
			verb:    "set the member list of",
			shareID: "def-456",
			wantSubs: []string{
				"the share is smb",
				`(wire type "smb")`,
				"not internal",
				"`share set-smb` for SMB shares",
			},
		},
		{
			// set-smb against a public share is unlikely but
			// possible; check the friendly-name translation works
			// for the wire-confusing direction (Public's wire
			// value is "external", but the user sees "public").
			name:    "set-smb against public",
			actual:  share.TypePublic,
			want:    share.TypeSMB,
			verb:    "update the member list of",
			shareID: "ghi-789",
			wantSubs: []string{
				"the share is public",
				`(wire type "external")`,
				"not smb",
			},
			mustNot: []string{
				// Friendly-name leakage in the other direction:
				// the rendered share TYPE label must read
				// "public", not the wire value.
				"the share is external",
			},
		},
	}
	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			rec := &share.Result{ID: c.shareID, ShareType: c.actual}
			err := requireShareType(rec, c.want, c.verb, c.shareID)
			if err == nil {
				t.Fatalf("requireShareType(%v, %v): expected refusal", c.actual, c.want)
			}
			msg := err.Error()
			for _, s := range c.wantSubs {
				if !strings.Contains(msg, s) {
					t.Errorf("error must contain %q; got: %v", s, err)
				}
			}
			for _, banned := range c.mustNot {
				if strings.Contains(msg, banned) {
					t.Errorf("error must NOT contain %q; got: %v", banned, err)
				}
			}
		})
	}
}

// TestRequireShareType_Match locks in the happy-path no-op: when
// the share's type matches the verb's expectation, requireShareType
// returns nil. Trivial assertion but pins the contract — a refactor
// that accidentally returns a "matches" error would silently break
// every update verb.
func TestRequireShareType_Match(t *testing.T) {
	cases := []struct {
		name string
		t    share.Type
	}{
		{"internal", share.TypeInternal},
		{"public (wire external)", share.TypePublic},
		{"smb", share.TypeSMB},
	}
	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			rec := &share.Result{ID: "id-1", ShareType: c.t}
			if err := requireShareType(rec, c.t, "do something to", "id-1"); err != nil {
				t.Errorf("requireShareType(matching types): unexpected error: %v", err)
			}
		})
	}
}

// TestRequireShareType_NilRecord covers the "share not found" branch
// — Query returns (nil, nil) for an unknown id, and the cobra layer
// hands that straight to requireShareType. The error must clearly
// state the id wasn't found rather than producing a misleading
// type-mismatch message against a zero-value share.Result.
func TestRequireShareType_NilRecord(t *testing.T) {
	err := requireShareType(nil, share.TypePublic, "set the password of", "missing-id")
	if err == nil {
		t.Fatal("requireShareType(nil, ...): expected error")
	}
	msg := err.Error()
	for _, want := range []string{"share missing-id", "not found"} {
		if !strings.Contains(msg, want) {
			t.Errorf("error must contain %q; got: %v", want, err)
		}
	}
}
