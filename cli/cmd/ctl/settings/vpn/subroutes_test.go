package vpn

import (
	"bytes"
	"encoding/json"
	"strings"
	"testing"
)

// TestDecodeSubroutesStatus mirrors TestDecodeSSHStatus: lock down the
// fact that the BFL envelope from /api/acl/subroutes/status is
// transparently unwrapped to the inner []string the SPA reads via its
// boot/axios.ts interceptor, while still tolerating both forward-looking
// "user-service starts stripping" and defensive null/empty shapes.
func TestDecodeSubroutesStatus(t *testing.T) {
	cases := []struct {
		name       string
		body       string
		want       []string
		assertNil  bool
		wantErrSub string
	}{
		{
			// Live wire shape today: response.Success(resp, subRoutes)
			// from handle_headscale.go:288 emits `{code, message,
			// data:[]string}` and user-service forwards verbatim.
			name: "envelope wrapped with routes",
			body: `{"code":0,"message":"success","data":["10.96.0.0/12","10.244.0.0/16"]}`,
			want: []string{"10.96.0.0/12", "10.244.0.0/16"},
		},
		{
			// SPA treats this as `allow_subroutes=false`.
			name:      "envelope wrapped with empty data array",
			body:      `{"code":0,"message":"success","data":[]}`,
			assertNil: false,
			want:      []string{},
		},
		{
			// Forward-looking: if user-service ever strips the envelope
			// the way it already does for /api/launcher-public-domain-
			// access-policy, the CLI must keep working without flipping
			// to "always empty".
			name: "unwrapped naked array",
			body: `["10.0.0.0/8"]`,
			want: []string{"10.0.0.0/8"},
		},
		{
			name:      "empty body",
			body:      ``,
			assertNil: true,
		},
		{
			name:      "null body",
			body:      `null`,
			assertNil: true,
		},
		{
			name:      "envelope with null data",
			body:      `{"code":0,"data":null}`,
			assertNil: true,
		},
		{
			name:      "envelope with omitted data",
			body:      `{"code":0,"message":"success"}`,
			assertNil: true,
		},
		{
			name:       "envelope non-zero code with message surfaces upstream error",
			body:       `{"code":-1,"message":"app not found"}`,
			wantErrSub: "app not found",
		},
		{
			name:       "envelope non-zero code without message falls back to code number",
			body:       `{"code":42}`,
			wantErrSub: "code=42",
		},
		{
			name:       "garbage body",
			body:       `not json`,
			wantErrSub: "decode acl subroutes status",
		},
	}
	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			got, err := decodeSubroutesStatus(json.RawMessage(c.body))
			if c.wantErrSub != "" {
				if err == nil {
					t.Fatalf("want error containing %q, got nil (routes=%v)", c.wantErrSub, got)
				}
				if !strings.Contains(err.Error(), c.wantErrSub) {
					t.Fatalf("err = %q, want substring %q", err.Error(), c.wantErrSub)
				}
				return
			}
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
			if c.assertNil {
				if got != nil {
					t.Fatalf("want nil routes, got %v", got)
				}
				return
			}
			if len(got) != len(c.want) {
				t.Fatalf("routes length = %d (%v), want %d (%v)", len(got), got, len(c.want), c.want)
			}
			for i, r := range c.want {
				if got[i] != r {
					t.Errorf("routes[%d] = %q, want %q", i, got[i], r)
				}
			}
		})
	}
}

// TestRenderSubroutesStatus locks down the table view contract: a single
// "Allow sub-routes" line (mirroring the SPA's `allow_subroutes` toggle)
// followed by the route list. Disabled state always prints "(none)"
// instead of an empty list to keep the column alignment intact.
func TestRenderSubroutesStatus(t *testing.T) {
	cases := []struct {
		name      string
		status    subroutesStatus
		wantSubs  []string
		wantSkips []string
	}{
		{
			name: "enabled with routes",
			status: subroutesStatus{
				AllowSubRoutes: true,
				Routes:         []string{"10.96.0.0/12", "10.244.0.0/16"},
			},
			wantSubs: []string{
				"Allow sub-routes:",
				"enabled",
				"Routes:",
				"(2)",
				"- 10.96.0.0/12",
				"- 10.244.0.0/16",
			},
			wantSkips: []string{"(none)"},
		},
		{
			name: "disabled with no routes",
			status: subroutesStatus{
				AllowSubRoutes: false,
				Routes:         nil,
			},
			wantSubs: []string{
				"Allow sub-routes:",
				"disabled",
				"Routes:",
				"(none)",
			},
			wantSkips: []string{"- "},
		},
	}
	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			var buf bytes.Buffer
			if err := renderSubroutesStatus(&buf, c.status); err != nil {
				t.Fatalf("renderSubroutesStatus: %v", err)
			}
			out := buf.String()
			for _, s := range c.wantSubs {
				if !strings.Contains(out, s) {
					t.Errorf("output missing %q\n---\n%s---", s, out)
				}
			}
			for _, s := range c.wantSkips {
				if strings.Contains(out, s) {
					t.Errorf("output unexpectedly contains %q\n---\n%s---", s, out)
				}
			}
		})
	}
}
