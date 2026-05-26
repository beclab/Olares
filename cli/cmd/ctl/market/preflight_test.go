package market

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

// TestIsUpgradableState pins the CLI's upgradable-state set verbatim to
// the SPA's `isUpgradableAppStates` in apps/.../constant/config.ts.
// Drift in either direction silently changes who can / can't run
// `market upgrade`, so keep this table in lockstep.
func TestIsUpgradableState(t *testing.T) {
	cases := []struct {
		state      string
		upgradable bool
	}{
		// SPA allowlist (5 states):
		{"running", true},
		{"stopped", true},
		{"stopFailed", true},
		{"upgradeFailed", true},
		{"applyEnvFailed", true},

		// Everything else must be denied — sample from the broader
		// backend state machine so a future expansion of the SPA
		// allowlist makes this test red instead of silently relaxing
		// the gate.
		{"", false},
		{"pending", false},
		{"downloading", false},
		{"installing", false},
		{"upgrading", false},
		{"resuming", false},
		{"stopping", false},
		{"applyingEnv", false},
		{"uninstalling", false},
		{"uninstalled", false},
		{"installFailed", false},
		{"resumeFailed", false},
		{"uninstallFailed", false},
		{"installingCanceled", false},
		{"resumingCanceled", false},
		{"upgradingCanceled", false},
		{"unknown-state-xyz", false},
	}
	for _, c := range cases {
		t.Run(c.state, func(t *testing.T) {
			if got := isUpgradable(c.state); got != c.upgradable {
				t.Fatalf("isUpgradable(%q) = %v, want %v", c.state, got, c.upgradable)
			}
		})
	}
}

// TestIsAppSuspended mirrors the SPA's `suspendApp(simpleLatest)`
// predicate: app_simple_info.app_labels containing 'suspend' OR
// 'remove' both flip the upgrade gate to false. Anything else (empty,
// nil, missing keys, wrong types, unrelated labels) leaves the gate
// open.
func TestIsAppSuspended(t *testing.T) {
	cases := []struct {
		name      string
		input     map[string]interface{}
		suspended bool
	}{
		{name: "nil input", input: nil, suspended: false},
		{name: "empty map", input: map[string]interface{}{}, suspended: false},
		{
			name: "missing app_simple_info",
			input: map[string]interface{}{
				"app_info": map[string]interface{}{},
			},
			suspended: false,
		},
		{
			name: "missing app_labels",
			input: map[string]interface{}{
				"app_simple_info": map[string]interface{}{},
			},
			suspended: false,
		},
		{
			name: "empty app_labels",
			input: map[string]interface{}{
				"app_simple_info": map[string]interface{}{
					"app_labels": []interface{}{},
				},
			},
			suspended: false,
		},
		{
			name: "unrelated labels only",
			input: map[string]interface{}{
				"app_simple_info": map[string]interface{}{
					"app_labels": []interface{}{"featured", "nsfw"},
				},
			},
			suspended: false,
		},
		{
			name: "suspend label present",
			input: map[string]interface{}{
				"app_simple_info": map[string]interface{}{
					"app_labels": []interface{}{"featured", "suspend"},
				},
			},
			suspended: true,
		},
		{
			name: "remove label present",
			input: map[string]interface{}{
				"app_simple_info": map[string]interface{}{
					"app_labels": []interface{}{"remove"},
				},
			},
			suspended: true,
		},
		{
			name: "both suspend and remove",
			input: map[string]interface{}{
				"app_simple_info": map[string]interface{}{
					"app_labels": []interface{}{"suspend", "remove"},
				},
			},
			suspended: true,
		},
		{
			name: "labels wrong type (string, not []) → false",
			input: map[string]interface{}{
				"app_simple_info": map[string]interface{}{
					"app_labels": "suspend",
				},
			},
			suspended: false,
		},
	}
	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			if got := isAppSuspended(c.input); got != c.suspended {
				t.Fatalf("isAppSuspended(%s) = %v, want %v", c.name, got, c.suspended)
			}
		})
	}
}

// fakeMarketBackend serves the two endpoints preflightUpgrade calls:
// GET /market/state and POST /apps. Each endpoint is fed by a per-test
// supplier closure so individual tests can pin one app to "installed
// running", another to "installed but state row missing", etc.
type fakeMarketBackend struct {
	t          *testing.T
	srv        *httptest.Server
	stateRows  map[string]fakeStateRow         // by sourceName
	appLabels  map[string][]string             // by appName: returned via /apps
	appMissing map[string]bool                 // appName -> 404-ish empty apps list
	appsCalls  []map[string]interface{}        // POST /apps payloads received
	stateErr   bool
	appsHook   func(payload map[string]interface{}) (map[string]interface{}, error)
}

type fakeStateRow struct {
	name    string
	rawName string
	state   string
	version string
}

func newFakeMarketBackend(t *testing.T) *fakeMarketBackend {
	t.Helper()
	f := &fakeMarketBackend{
		t:          t,
		stateRows:  map[string]fakeStateRow{},
		appLabels:  map[string][]string{},
		appMissing: map[string]bool{},
	}
	f.srv = httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch {
		case strings.HasSuffix(r.URL.Path, "/market/state") && r.Method == http.MethodGet:
			if f.stateErr {
				http.Error(w, "synthetic state error", http.StatusInternalServerError)
				return
			}
			w.Header().Set("Content-Type", "application/json")
			_, _ = w.Write(f.encodeState())
		case strings.HasSuffix(r.URL.Path, "/apps") && r.Method == http.MethodPost:
			var payload map[string]interface{}
			_ = json.NewDecoder(r.Body).Decode(&payload)
			f.appsCalls = append(f.appsCalls, payload)
			if f.appsHook != nil {
				resp, err := f.appsHook(payload)
				if err != nil {
					http.Error(w, err.Error(), http.StatusInternalServerError)
					return
				}
				w.Header().Set("Content-Type", "application/json")
				_ = json.NewEncoder(w).Encode(resp)
				return
			}
			w.Header().Set("Content-Type", "application/json")
			_, _ = w.Write(f.encodeApps(payload))
		default:
			http.NotFound(w, r)
		}
	}))
	t.Cleanup(f.srv.Close)
	return f
}

func (f *fakeMarketBackend) encodeState() []byte {
	sources := map[string]interface{}{}
	for sourceName, row := range f.stateRows {
		sources[sourceName] = map[string]interface{}{
			"type": "market",
			"app_state_latest": []map[string]interface{}{
				{
					"version": row.version,
					"status": map[string]interface{}{
						"name":       row.name,
						"rawAppName": row.rawName,
						"state":      row.state,
					},
				},
			},
		}
	}
	body, _ := json.Marshal(map[string]interface{}{
		"success": true,
		"data": map[string]interface{}{
			"user_data": map[string]interface{}{
				"sources": sources,
			},
		},
	})
	return body
}

func (f *fakeMarketBackend) encodeApps(payload map[string]interface{}) []byte {
	// The /apps endpoint takes { apps: [{appid, sourceDataName}] }; we
	// honor only the first request element since preflight only ever
	// fetches one app's metadata at a time.
	queries, _ := payload["apps"].([]interface{})
	if len(queries) == 0 {
		return mustJSON(map[string]interface{}{"success": true, "data": map[string]interface{}{"apps": []interface{}{}}})
	}
	q, _ := queries[0].(map[string]interface{})
	appID, _ := q["appid"].(string)
	if f.appMissing[appID] {
		return mustJSON(map[string]interface{}{"success": true, "data": map[string]interface{}{"apps": []interface{}{}}})
	}
	labels := f.appLabels[appID]
	labelIfaces := make([]interface{}, 0, len(labels))
	for _, l := range labels {
		labelIfaces = append(labelIfaces, l)
	}
	return mustJSON(map[string]interface{}{
		"success": true,
		"data": map[string]interface{}{
			"apps": []map[string]interface{}{
				{
					"app_simple_info": map[string]interface{}{
						"app_labels": labelIfaces,
					},
					"app_info": map[string]interface{}{
						"app_entry": map[string]interface{}{
							"apiVersion": "v1",
						},
					},
				},
			},
		},
	})
}

func mustJSON(v interface{}) []byte {
	b, err := json.Marshal(v)
	if err != nil {
		panic(err)
	}
	return b
}

func TestPreflightUpgrade(t *testing.T) {
	cases := []struct {
		name        string
		setup       func(f *fakeMarketBackend)
		appName     string
		target      string
		source      string
		wantErrSub  string // substring expected in the returned error (empty == expect nil)
		wantSoftWar bool   // true if we expect the soft-fail warning path (no error, suspend gate skipped)
	}{
		{
			name:    "happy path: state=running, target > installed, no suspend label",
			appName: "firefox",
			target:  "1.0.12",
			source:  "market.olares",
			setup: func(f *fakeMarketBackend) {
				f.stateRows["market.olares"] = fakeStateRow{name: "firefox", state: "running", version: "1.0.11"}
			},
		},
		{
			name:       "not installed → bail with install hint",
			appName:    "firefox",
			target:     "1.0.12",
			source:     "market.olares",
			setup:      func(f *fakeMarketBackend) {},
			wantErrSub: "is not installed",
		},
		{
			name:    "non-upgradable state (installing) → bail",
			appName: "firefox",
			target:  "1.0.12",
			source:  "market.olares",
			setup: func(f *fakeMarketBackend) {
				f.stateRows["market.olares"] = fakeStateRow{name: "firefox", state: "installing", version: "1.0.11"}
			},
			wantErrSub: "in state 'installing'",
		},
		{
			name:    "upgradeFailed retry path is allowed",
			appName: "firefox",
			target:  "1.0.12",
			source:  "market.olares",
			setup: func(f *fakeMarketBackend) {
				f.stateRows["market.olares"] = fakeStateRow{name: "firefox", state: "upgradeFailed", version: "1.0.11"}
			},
		},
		{
			name:    "state row has no version → bail",
			appName: "firefox",
			target:  "1.0.12",
			source:  "market.olares",
			setup: func(f *fakeMarketBackend) {
				f.stateRows["market.olares"] = fakeStateRow{name: "firefox", state: "running", version: ""}
			},
			wantErrSub: "no version recorded",
		},
		{
			name:    "target == installed → bail (nothing to do)",
			appName: "firefox",
			target:  "1.0.11",
			source:  "market.olares",
			setup: func(f *fakeMarketBackend) {
				f.stateRows["market.olares"] = fakeStateRow{name: "firefox", state: "running", version: "1.0.11"}
			},
			wantErrSub: "already installed",
		},
		{
			name:    "target < installed → bail (downgrade)",
			appName: "firefox",
			target:  "1.0.10",
			source:  "market.olares",
			setup: func(f *fakeMarketBackend) {
				f.stateRows["market.olares"] = fakeStateRow{name: "firefox", state: "running", version: "1.0.11"}
			},
			wantErrSub: "is older than installed",
		},
		{
			name:    "invalid installed version → comparison error",
			appName: "firefox",
			target:  "1.0.12",
			source:  "market.olares",
			setup: func(f *fakeMarketBackend) {
				f.stateRows["market.olares"] = fakeStateRow{name: "firefox", state: "running", version: "not-semver"}
			},
			wantErrSub: "version comparison failed",
		},
		{
			name:    "suspend label present → bail (chart withdrawn)",
			appName: "firefox",
			target:  "1.0.12",
			source:  "market.olares",
			setup: func(f *fakeMarketBackend) {
				f.stateRows["market.olares"] = fakeStateRow{name: "firefox", state: "running", version: "1.0.11"}
				f.appLabels["firefox"] = []string{"suspend"}
			},
			wantErrSub: "marked 'suspend' or 'remove'",
		},
		{
			name:    "remove label present → bail",
			appName: "firefox",
			target:  "1.0.12",
			source:  "market.olares",
			setup: func(f *fakeMarketBackend) {
				f.stateRows["market.olares"] = fakeStateRow{name: "firefox", state: "running", version: "1.0.11"}
				f.appLabels["firefox"] = []string{"remove"}
			},
			wantErrSub: "marked 'suspend' or 'remove'",
		},
		{
			name:    "catalog /apps returns empty apps → soft-fail (proceed)",
			appName: "firefox",
			target:  "1.0.12",
			source:  "market.olares",
			setup: func(f *fakeMarketBackend) {
				f.stateRows["market.olares"] = fakeStateRow{name: "firefox", state: "running", version: "1.0.11"}
				f.appMissing["firefox"] = true
			},
			wantSoftWar: true,
		},
		{
			name:    "clone path: lookup uses rawAppName for catalog",
			appName: "windowsefe992",
			target:  "1.1.5",
			source:  "market.olares",
			setup: func(f *fakeMarketBackend) {
				f.stateRows["market.olares"] = fakeStateRow{
					name: "windowsefe992", rawName: "windows", state: "running", version: "1.1.4",
				}
				// Suspend label is registered under the SOURCE app
				// (windows), not the clone instance name.
				f.appLabels["windows"] = []string{"suspend"}
			},
			wantErrSub: "marked 'suspend' or 'remove'",
		},
	}

	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			fb := newFakeMarketBackend(t)
			c.setup(fb)
			mc := newTestMarketClient(t, fb.srv.URL)
			opts := &MarketOptions{Source: c.source, Output: "json", Quiet: true}

			err := preflightUpgrade(context.Background(), opts, mc, c.appName, c.target, c.source)
			switch {
			case c.wantErrSub == "" && !c.wantSoftWar:
				if err != nil {
					t.Fatalf("preflightUpgrade unexpected error: %v", err)
				}
			case c.wantSoftWar:
				if err != nil {
					t.Fatalf("preflightUpgrade soft-fail case returned error: %v", err)
				}
			default:
				if err == nil {
					t.Fatalf("preflightUpgrade returned nil; want error containing %q", c.wantErrSub)
				}
				if !strings.Contains(err.Error(), c.wantErrSub) {
					t.Fatalf("preflightUpgrade error %q does not contain %q", err.Error(), c.wantErrSub)
				}
			}
		})
	}
}

// TestLookupInstalledAppDisambiguatesPrimaryFromClones is the
// regression test for the clone-vs-primary ambiguity bug in
// lookupInstalledApp. Before the fix, the matching predicate was
// `rowName != appName && rawName != appName` — i.e. either rowName
// OR rawName equal to appName counted as a match. That made the
// lookup non-deterministic whenever the source app AND one or more
// clones of it were both installed: a query for the source-app name
// (`windows`) would also match every clone (each carries
// `RawName=windows`), and the function returned whichever row
// happened to come first in slice / map iteration.
//
// The state fixture below puts the clone (windowsefe992,
// RawName=windows, state=stopFailed, version=1.0.5) FIRST in the
// app_state_latest slice and the primary (windows, state=running,
// version=1.2.0) second. Slice iteration is deterministic in Go, so
// the old buggy code reliably returned the clone for
// `lookupInstalledApp("windows")` — exactly the failure mode the
// bug report describes for `market upgrade windows` (gate 2/3 would
// see the clone's state/version instead of the primary's) and
// `shouldAutoCascade("windows")` (would probe the catalog with the
// clone's source).
//
// The new code matches strictly on `row.Name == appName`. The clone
// is skipped because its Name is `windowsefe992`, the primary is
// returned, and the same fixture also confirms a direct lookup for
// `windowsefe992` still returns the clone row.
func TestLookupInstalledAppDisambiguatesPrimaryFromClones(t *testing.T) {
	state := `{
        "user_data": {
            "sources": {
                "market.olares": {
                    "type": "market",
                    "app_state_latest": [
                        {"version": "1.0.5", "status": {"name": "windowsefe992", "rawAppName": "windows", "state": "stopFailed"}},
                        {"version": "1.2.0", "status": {"name": "windows", "rawAppName": "windows", "state": "running"}}
                    ]
                }
            }
        }
    }`

	srv := newFakeMarketDataServer(t, stateAndDataResponses{state: state})
	mc := newTestMarketClient(t, srv.URL)

	t.Run("source-app name returns the primary, never the clone", func(t *testing.T) {
		row, err := lookupInstalledApp(context.Background(), mc, "windows")
		if err != nil {
			t.Fatalf("lookupInstalledApp: %v", err)
		}
		if row == nil {
			t.Fatalf("lookupInstalledApp returned nil for an installed app")
		}
		if row.Name != "windows" {
			t.Fatalf("lookup(\"windows\") returned the wrong row: name=%q (probably the clone — the bug); want %q (the primary)", row.Name, "windows")
		}
		if row.State != "running" || row.Version != "1.2.0" {
			t.Fatalf("primary row data wrong: state=%q version=%q (want state=running version=1.2.0)", row.State, row.Version)
		}
	})

	t.Run("clone instance name returns the clone (unchanged)", func(t *testing.T) {
		row, err := lookupInstalledApp(context.Background(), mc, "windowsefe992")
		if err != nil {
			t.Fatalf("lookupInstalledApp: %v", err)
		}
		if row == nil {
			t.Fatalf("lookupInstalledApp returned nil for the clone")
		}
		if row.Name != "windowsefe992" {
			t.Fatalf("clone lookup returned wrong row: name=%q want %q", row.Name, "windowsefe992")
		}
		if row.RawName != "windows" {
			t.Fatalf("clone row must carry rawAppName=windows; got %q", row.RawName)
		}
		if row.State != "stopFailed" || row.Version != "1.0.5" {
			t.Fatalf("clone row data wrong: state=%q version=%q (want state=stopFailed version=1.0.5)", row.State, row.Version)
		}
	})

	t.Run("source-name only matches a primary, even when only clones exist for that source", func(t *testing.T) {
		// Edge case: user types the source-app name but ONLY a clone
		// of that source is installed (no primary `windows` row). The
		// strict-Name rule must return nil — the user typed a name
		// that doesn't identify any installed row, and we'd rather
		// surface "not installed" than silently operate on a clone
		// the user didn't ask for.
		stateCloneOnly := `{
            "user_data": {
                "sources": {
                    "market.olares": {
                        "type": "market",
                        "app_state_latest": [
                            {"version": "1.0.5", "status": {"name": "windowsefe992", "rawAppName": "windows", "state": "running"}}
                        ]
                    }
                }
            }
        }`
		srvC := newFakeMarketDataServer(t, stateAndDataResponses{state: stateCloneOnly})
		mcC := newTestMarketClient(t, srvC.URL)
		row, err := lookupInstalledApp(context.Background(), mcC, "windows")
		if err != nil {
			t.Fatalf("lookupInstalledApp: %v", err)
		}
		if row != nil {
			t.Fatalf("source-name lookup must return nil when only a clone exists; got %+v", row)
		}
	})

	t.Run("legacy edge case: row with Name=\"\" but RawName populated still matches", func(t *testing.T) {
		// Conservative fallback for older backends that may surface
		// rows with empty Name. Match only when row.Name is empty
		// AND row.RawName == appName — never when row.Name is
		// populated with a different value (the clone disambiguation
		// rule above).
		stateLegacy := `{
            "user_data": {
                "sources": {
                    "market.olares": {
                        "type": "market",
                        "app_state_latest": [
                            {"version": "0.9.0", "status": {"name": "", "rawAppName": "vault", "state": "running"}}
                        ]
                    }
                }
            }
        }`
		srvL := newFakeMarketDataServer(t, stateAndDataResponses{state: stateLegacy})
		mcL := newTestMarketClient(t, srvL.URL)
		row, err := lookupInstalledApp(context.Background(), mcL, "vault")
		if err != nil {
			t.Fatalf("lookupInstalledApp: %v", err)
		}
		if row == nil {
			t.Fatalf("legacy Name-empty row must still match by RawName fallback")
		}
		if row.Name != "vault" || row.Version != "0.9.0" {
			t.Fatalf("legacy fallback row: name=%q version=%q (want vault / 0.9.0)", row.Name, row.Version)
		}
	})
}

// TestPreflightUpgrade_SourceMismatchWarns confirms that preflight does
// NOT fail when the user passes -s pointing at a different source from
// the one the app is currently installed from — it should only emit a
// stderr warning. The backend (not the CLI) has the final say on
// cross-source upgrades.
func TestPreflightUpgrade_SourceMismatchWarns(t *testing.T) {
	fb := newFakeMarketBackend(t)
	fb.stateRows["market.test"] = fakeStateRow{name: "firefox", state: "running", version: "1.0.11"}

	mc := newTestMarketClient(t, fb.srv.URL)
	opts := &MarketOptions{Source: "market.olares", Output: "json", Quiet: true}

	err := preflightUpgrade(context.Background(), opts, mc, "firefox", "1.0.12", "market.olares")
	if err != nil {
		t.Fatalf("preflightUpgrade returned error on source mismatch: %v", err)
	}
}
