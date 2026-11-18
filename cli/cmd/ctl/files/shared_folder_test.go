package files

import (
	"strings"
	"testing"
)

func TestSharedFolderConflict(t *testing.T) {
	body := `{"code":1,"message":"The folder, or a folder in it, is shared.","shares":[` +
		`{"id":"a","share_type":"smb","owner":"alice","path":"/P/","name":"P"},` +
		`{"id":"b","share_type":"external","owner":"bob","path":"/P/q/","name":""}]}`
	msg, ok := sharedFolderConflict(409, body)
	if !ok {
		t.Fatal("expected a shared-folder conflict")
	}
	for _, want := range []string{"P (smb, by alice)", "/P/q/ (public, by bob)", "--unshare"} {
		if !strings.Contains(msg, want) {
			t.Errorf("message must contain %q; got: %s", want, msg)
		}
	}

	many := `{"shares":[` + strings.Repeat(`{"share_type":"smb","owner":"o","name":"n"},`, 7) +
		`{"share_type":"smb","owner":"o","name":"n"}]}`
	if msg, _ := sharedFolderConflict(409, many); !strings.Contains(msg, "and 3 more") {
		t.Errorf("long lists must be cut; got: %s", msg)
	}

	for _, c := range []struct {
		status int
		body   string
	}{
		{409, `{"code":1,"message":"file exists"}`},
		{409, `not json`},
		{400, body},
	} {
		if _, ok := sharedFolderConflict(c.status, c.body); ok {
			t.Errorf("sharedFolderConflict(%d, %q): not a shared-folder conflict", c.status, c.body)
		}
	}
}
