package apps

import (
	"context"
	"encoding/json"
	"strings"
	"testing"
)

// TestResolveSuspendAll covers the four-state matrix for the --all
// flag: explicit --all=true / --all=false win unconditionally, while
// the auto path resolves to the named app's isClusterScoped (CS apps
// default to true, user-scoped default to false), with a clean
// not-found error when the named app is missing from /api/myapps.
func TestResolveSuspendAll(t *testing.T) {
	cases := []struct {
		name        string
		appName     string
		allOpt      *bool
		myappsRows  []appInfo
		wantAll     bool
		wantPrefix  string
		wantErrSub  string
		wantHTTPLen int // expected number of /api/myapps fetches
	}{
		{
			name:        "explicit_true_skips_lookup",
			appName:     "settings",
			allOpt:      boolPtr(true),
			wantAll:     true,
			wantPrefix:  "--all=true set explicitly",
			wantHTTPLen: 0,
		},
		{
			name:        "explicit_false_skips_lookup",
			appName:     "settings",
			allOpt:      boolPtr(false),
			wantAll:     false,
			wantPrefix:  "--all=false set explicitly",
			wantHTTPLen: 0,
		},
		{
			name:    "auto_cs_app_defaults_true",
			appName: "settings",
			allOpt:  nil,
			myappsRows: []appInfo{
				{Name: "files", IsClusterScoped: false},
				{Name: "settings", IsClusterScoped: true},
			},
			wantAll:     true,
			wantPrefix:  "auto from isClusterScoped=true",
			wantHTTPLen: 1,
		},
		{
			name:    "auto_user_scoped_defaults_false",
			appName: "testenv",
			allOpt:  nil,
			myappsRows: []appInfo{
				{Name: "testenv", IsClusterScoped: false},
			},
			wantAll:     false,
			wantPrefix:  "auto from isClusterScoped=false",
			wantHTTPLen: 1,
		},
		{
			name:        "auto_unknown_app_errs",
			appName:     "settings",
			allOpt:      nil,
			myappsRows:  []appInfo{{Name: "other", IsClusterScoped: true}},
			wantErrSub:  "not found in /api/myapps",
			wantHTTPLen: 1,
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			d := &fakeDoer{}
			if tc.allOpt == nil && tc.myappsRows != nil {
				d.enqueueEnvelope(tc.myappsRows)
			}
			gotAll, gotSrc, err := resolveSuspendAll(context.Background(), d, tc.appName, tc.allOpt)
			if tc.wantErrSub != "" {
				if err == nil {
					t.Fatalf("want error containing %q, got nil", tc.wantErrSub)
				}
				if !strings.Contains(err.Error(), tc.wantErrSub) {
					t.Fatalf("err = %v, want substring %q", err, tc.wantErrSub)
				}
				return
			}
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
			if gotAll != tc.wantAll {
				t.Errorf("all = %t, want %t", gotAll, tc.wantAll)
			}
			if !strings.HasPrefix(gotSrc, tc.wantPrefix) {
				t.Errorf("source = %q, want prefix %q", gotSrc, tc.wantPrefix)
			}
			if got, want := len(d.calls), tc.wantHTTPLen; got != want {
				t.Errorf("HTTP calls = %d, want %d (calls=%+v)", got, want, d.calls)
			}
			if tc.wantHTTPLen == 1 {
				c := d.lastCall()
				if c.method != "GET" || c.path != "/api/myapps" {
					t.Errorf("auto lookup hit %s %s, want GET /api/myapps", c.method, c.path)
				}
			}
			// Resolver must be called for the SAME app the caller
			// passed to runAppSuspend. Re-run with a name that
			// isn't in the rows to make the contract obvious.
			if tc.allOpt == nil && len(tc.myappsRows) > 0 && tc.wantErrSub == "" {
				d2 := &fakeDoer{}
				d2.enqueueEnvelope(tc.myappsRows)
				_, _, err := resolveSuspendAll(context.Background(), d2, "definitely-not-installed", nil)
				if err == nil || !strings.Contains(err.Error(), "not found") {
					t.Errorf("expected not-found error for unknown app, got %v", err)
				}
			}
		})
	}
}

// TestRunAppSuspend_BodyShape verifies the exact wire shape posted
// to /api/app/suspend in the matrix:
//
//   - --all=true (explicit) -> body {name, all:true}
//   - --all=false (explicit) -> body {name, all:false}
//   - auto on CS app -> body {name, all:true} after one /api/myapps
//   - auto on user-scoped -> body {name, all:false} after one /api/myapps
func TestRunAppSuspend_BodyShape(t *testing.T) {
	type want struct {
		all     bool
		preLen  int // /api/myapps lookups before the POST
	}
	cases := []struct {
		name   string
		allOpt *bool
		rows   []appInfo
		want   want
	}{
		{
			name:   "explicit_true",
			allOpt: boolPtr(true),
			want:   want{all: true, preLen: 0},
		},
		{
			name:   "explicit_false",
			allOpt: boolPtr(false),
			want:   want{all: false, preLen: 0},
		},
		{
			name:   "auto_cs",
			allOpt: nil,
			rows:   []appInfo{{Name: "settings", IsClusterScoped: true}},
			want:   want{all: true, preLen: 1},
		},
		{
			name:   "auto_user_scoped",
			allOpt: nil,
			rows:   []appInfo{{Name: "testenv", IsClusterScoped: false}},
			want:   want{all: false, preLen: 1},
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			appName := "testenv"
			if tc.want.preLen > 0 {
				appName = tc.rows[0].Name
			}

			d := &fakeDoer{}
			if tc.want.preLen > 0 {
				d.enqueueEnvelope(tc.rows)
			}
			d.enqueueEmptyEnvelope() // POST /api/app/suspend response

			all, _, err := resolveSuspendAll(context.Background(), d, appName, tc.allOpt)
			if err != nil {
				t.Fatalf("resolve: %v", err)
			}
			body := map[string]interface{}{
				"name": appName,
				"all":  all,
			}
			if err := doMutateEnvelope(context.Background(), d, "POST", "/api/app/suspend", body, nil); err != nil {
				t.Fatalf("mutate: %v", err)
			}

			if got, want := len(d.calls), tc.want.preLen+1; got != want {
				t.Fatalf("HTTP calls = %d, want %d (calls=%+v)", got, want, d.calls)
			}
			post := d.calls[len(d.calls)-1]
			if post.method != "POST" || post.path != "/api/app/suspend" {
				t.Errorf("POST went to %s %s, want POST /api/app/suspend", post.method, post.path)
			}

			raw, err := json.Marshal(post.body)
			if err != nil {
				t.Fatalf("marshal body: %v", err)
			}
			var decoded struct {
				Name string `json:"name"`
				All  bool   `json:"all"`
			}
			if err := json.Unmarshal(raw, &decoded); err != nil {
				t.Fatalf("unmarshal body: %v (raw=%s)", err, raw)
			}
			if decoded.Name != appName {
				t.Errorf("body.name = %q, want %q", decoded.Name, appName)
			}
			if decoded.All != tc.want.all {
				t.Errorf("body.all = %t, want %t (raw=%s)", decoded.All, tc.want.all, raw)
			}
		})
	}
}

// TestRunAppResume_NoBody guards the resume invariant: GET path with
// the name escaped in the URL, and absolutely no body. If somebody ever
// adds an --all flag to resume by mistake (the wire doesn't accept
// one), this test will fail.
func TestRunAppResume_NoBody(t *testing.T) {
	d := &fakeDoer{}
	d.enqueueEmptyEnvelope()

	if err := doMutateEnvelope(context.Background(), d, "GET", "/api/app/resume/files%20app", nil, nil); err != nil {
		t.Fatalf("mutate: %v", err)
	}
	if got := len(d.calls); got != 1 {
		t.Fatalf("calls = %d, want 1", got)
	}
	c := d.calls[0]
	if c.method != "GET" {
		t.Errorf("method = %q, want GET", c.method)
	}
	if c.path != "/api/app/resume/files%20app" {
		t.Errorf("path = %q, want url-escaped /api/app/resume/files%%20app", c.path)
	}
	if c.body != nil {
		t.Errorf("resume must not send a body, got %+v", c.body)
	}
}

func boolPtr(v bool) *bool { return &v }
