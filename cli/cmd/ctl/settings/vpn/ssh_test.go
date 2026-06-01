package vpn

import (
	"encoding/json"
	"strings"
	"testing"
)

func TestDecodeSSHStatus(t *testing.T) {
	cases := []struct {
		name        string
		body        string
		wantState   string
		wantAllow   bool
		wantErrSub  string
		assertEmpty bool
	}{
		{
			// Live wire shape after the user-service round-trip. This is
			// the case the original code silently broke on — the body
			// arrives as the BFL envelope with the inner shape under
			// `data`, but the decoder unmarshaled into sshStatus directly
			// and `state` / `allow_ssh` ended up zero-valued no matter
			// what the cluster actually said.
			name:      "envelope wrapped with allow_ssh true",
			body:      `{"code":0,"message":"success","data":{"state":"applied","allow_ssh":true}}`,
			wantState: "applied",
			wantAllow: true,
		},
		{
			name:      "envelope wrapped with allow_ssh false",
			body:      `{"code":0,"message":"success","data":{"state":"applied","allow_ssh":false}}`,
			wantState: "applied",
			wantAllow: false,
		},
		{
			// Forward-looking shape: if user-service ever starts
			// stripping the envelope for this path (as it already does
			// for /api/launcher-public-domain-access-policy), the CLI
			// must NOT regress to always-empty output.
			name:      "unwrapped inner shape",
			body:      `{"state":"applied","allow_ssh":true}`,
			wantState: "applied",
			wantAllow: true,
		},
		{
			name:      "success code with flat status fields",
			body:      `{"code":0,"message":"success","state":"applied","allow_ssh":true}`,
			wantState: "applied",
			wantAllow: true,
		},
		{
			name:        "empty body",
			body:        ``,
			assertEmpty: true,
		},
		{
			name:        "null body",
			body:        `null`,
			assertEmpty: true,
		},
		{
			name:        "envelope with null data",
			body:        `{"code":0,"data":null}`,
			assertEmpty: true,
		},
		{
			name:        "envelope with omitted data",
			body:        `{"code":0,"message":"success"}`,
			assertEmpty: true,
		},
		{
			name:       "envelope non-zero code with message surfaces upstream error",
			body:       `{"code":-1,"message":"ACL CRD not reconciled yet"}`,
			wantErrSub: "ACL CRD not reconciled yet",
		},
		{
			name:       "envelope non-zero code without message falls back to code number",
			body:       `{"code":42}`,
			wantErrSub: "code=42",
		},
		{
			name:      "envelope with extra unknown fields",
			body:      `{"code":0,"data":{"state":"applied","allow_ssh":true,"extra":"ignored"}}`,
			wantState: "applied",
			wantAllow: true,
		},
		{
			name:       "garbage body",
			body:       `not json`,
			wantErrSub: "decode acl ssh status",
		},
	}
	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			got, err := decodeSSHStatus(json.RawMessage(c.body))
			if c.wantErrSub != "" {
				if err == nil {
					t.Fatalf("want error containing %q, got nil (state=%q, allow=%v)", c.wantErrSub, got.State, got.AllowSSH)
				}
				if !strings.Contains(err.Error(), c.wantErrSub) {
					t.Fatalf("err = %q, want substring %q", err.Error(), c.wantErrSub)
				}
				return
			}
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
			if c.assertEmpty {
				if got.State != "" || got.AllowSSH {
					t.Fatalf("want zero-valued status, got state=%q allow=%v", got.State, got.AllowSSH)
				}
				return
			}
			if got.State != c.wantState {
				t.Errorf("state = %q, want %q", got.State, c.wantState)
			}
			if got.AllowSSH != c.wantAllow {
				t.Errorf("allow_ssh = %v, want %v", got.AllowSSH, c.wantAllow)
			}
		})
	}
}

func TestEnvelopeLooksWrapped(t *testing.T) {
	cases := []struct {
		name string
		body string
		want bool
	}{
		{name: "envelope with data", body: `{"code":0,"data":{"x":1}}`, want: true},
		{name: "envelope with null data", body: `{"code":0,"data":null}`, want: true},
		{name: "envelope with empty data object", body: `{"code":0,"data":{}}`, want: true},
		{name: "no data key", body: `{"state":"applied","allow_ssh":true}`, want: false},
		{name: "non-object body", body: `[]`, want: false},
	}
	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			var env bflEnvelope
			_ = json.Unmarshal([]byte(c.body), &env)
			got := envelopeLooksWrapped([]byte(c.body), env)
			if got != c.want {
				t.Errorf("envelopeLooksWrapped(%q) = %v, want %v", c.body, got, c.want)
			}
		})
	}
}
