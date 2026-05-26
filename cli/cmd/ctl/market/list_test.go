package market

import (
	"net/http"
	"net/http/httptest"
	"reflect"
	"strings"
	"testing"
)

// TestIsInstalledState locks down the "is this row one of the user's
// apps?" filter used by runListInstalled (the implementation of
// `market list --mine`), aligned VERBATIM to the SPA's "My Terminus"
// filter (`uninstalledApp` / `uninstalledAppStates` in
// apps/packages/app/src/constant/config.ts ~L170). The six states
// asserted as `installed: false` below are exactly the SPA's set —
// the only rows hidden from My Terminus; everything else — including
// transitional in-flight states like `pending`, `downloading`,
// `installing` and their `*Canceling` / `*CancelFailed` variants —
// must stay in the listing so the CLI and the Market UI show the same
// apps. (The `installed` column name is the internal shorthand the
// helper uses; the user-facing semantic is "shown on My Terminus".)
// If the SPA changes that set, update notInstalledStates and this
// table together.
func TestIsInstalledState(t *testing.T) {
	cases := []struct {
		state     string
		installed bool
	}{
		// SPA `uninstalledAppStates` — the six terminal failure /
		// cancel / uninstall states that hide a row from My Terminus.
		{"pendingCanceled", false},
		{"downloadingCanceled", false},
		{"downloadFailed", false},
		{"installFailed", false},
		{"installingCanceled", false},
		{"uninstalled", false},

		// In-flight install path. The SPA keeps these visible on My
		// Terminus because the user clicked install and wants to
		// see / monitor / cancel the row, so `--mine` must show them
		// too — they're part of "my apps" even though they aren't
		// "完成安装" yet.
		{"pending", true},
		{"pendingCanceling", true},
		{"pendingCancelFailed", true},
		{"downloading", true},
		{"downloadingCanceling", true},
		{"downloadingCancelFailed", true},
		{"installing", true},
		{"installingCanceling", true},
		{"installingCancelFailed", true},

		// `initializing` is reused by Resuming / Upgrading /
		// ApplyingEnv transitions, and `initializingCanceling` always
		// transitions to `stopping → stopped`; both stay installed
		// (the SPA never lists them as `uninstalledAppStates`).
		{"initializing", true},
		{"initializingCanceling", true},

		// Terminal-success / post-install / lifecycle-transient
		// states — the SPA always shows these on My Terminus.
		{"running", true},
		{"stopped", true},
		{"stopping", true},
		{"stopFailed", true},
		{"resuming", true},
		{"resumeFailed", true},
		{"resumingCanceling", true},
		{"resumingCanceled", true},
		{"resumingCancelFailed", true},
		{"upgrading", true},
		{"upgradeFailed", true},
		{"upgradingCanceling", true},
		{"upgradingCanceled", true},
		{"upgradingCancelFailed", true},
		{"applyingEnv", true},
		{"applyEnvFailed", true},
		{"applyingEnvCanceling", true},
		{"applyingEnvCanceled", true},
		{"applyingEnvCancelFailed", true},
		{"uninstalling", true},
		{"uninstallFailed", true},

		// Defensive: empty string is treated as "row has no usable
		// state info" and falls into the not-installed bucket so a
		// malformed payload never silently pads the listing.
		{"", false},
	}

	for _, c := range cases {
		t.Run(c.state, func(t *testing.T) {
			if got := isInstalledState(c.state); got != c.installed {
				t.Fatalf("isInstalledState(%q) = %v, want %v", c.state, got, c.installed)
			}
		})
	}
}

// TestFetchInstalledAppsParsesStateAndEnrichesCatalog exercises the full
// parsing path: it feeds a fixture matching the wire shape of
// /market/state + /market/data into an httptest server backing a real
// *MarketClient and checks that (a) only the SPA's "My Terminus" rows
// survive the filter (others get dropped — see obsidian / ghost),
// (b) the per-row Version comes from AppStateLatest.Version (the
// version the user picked for this row, NOT the catalog's latest —
// see the firefox case below where the two disagree), (c) catalog
// enrichment fills in title / categories when the row exists in
// /market/data, and (d) rows missing from the catalog still render
// with whatever the state knows.
func TestFetchInstalledAppsParsesStateAndEnrichesCatalog(t *testing.T) {
	srv := newFakeMarketDataServer(t, stateAndDataResponses{
		state: marketStateFixture(),
		data:  marketDataFixture(),
	})
	mc := newTestMarketClient(t, srv.URL)

	got, err := fetchInstalledApps(mc, "", true)
	if err != nil {
		t.Fatalf("fetchInstalledApps: %v", err)
	}

	want := []AppDisplayInfo{
		// `myapp` lives in the `cli` local source and the catalog
		// fixture is silent about it; the state row also lacks a
		// version, so the renderer surfaces an empty version rather
		// than guessing.
		{Name: "myapp", Source: "cli", State: "running"},
		// `firefox` is the canonical "row version != catalog latest"
		// case: the user's state row carries 1.1.0, but /market/data
		// claims the latest is 1.2.3. The CLI MUST surface 1.1.0 —
		// anything else lies to the user about which version their
		// row is on.
		{
			Name:       "firefox",
			Title:      "Firefox",
			Version:    "1.1.0",
			Source:     "market.olares",
			Categories: []string{"Web", "Browser"},
			State:      "running",
		},
		// `kuma`'s state and catalog versions happen to match (0.5.0),
		// confirming we don't accidentally regress the easy case.
		{
			Name:       "kuma",
			Title:      "Kuma",
			Version:    "0.5.0",
			Source:     "market.olares",
			Categories: []string{"Monitoring"},
			State:      "upgradeFailed",
		},
	}

	if !reflect.DeepEqual(got, want) {
		t.Fatalf("fetchInstalledApps mismatch:\n got: %#v\nwant: %#v", got, want)
	}
}

// TestFetchInstalledAppsMirrorsSpaUninstalledFilter wires the full
// install-path state space through fetchInstalledApps and asserts the
// listing matches the SPA's "My Terminus" filter exactly: only the six
// `uninstalledAppStates` (apps/packages/app/src/constant/config.ts
// ~L170) drop out; every in-flight install row (`pending`,
// `downloading`, `installing`, plus the `*Canceling` / `*CancelFailed`
// variants) survives because the SPA shows them on My Terminus too.
//
// This pins the alignment: if a future change starts dropping
// `pending` / `downloading` / `installing` (or stops dropping
// `installFailed` / `downloadFailed` etc.), the CLI listing diverges
// from the SPA's "what apps do I have" view and the user's mental
// model breaks.
func TestFetchInstalledAppsMirrorsSpaUninstalledFilter(t *testing.T) {
	srv := newFakeMarketDataServer(t, stateAndDataResponses{
		state: `{
            "user_data": {
                "sources": {
                    "market.olares": {"type": "market", "app_state_latest": [
                        {"version": "1.0.0", "status": {"name": "p1", "state": "pending"}},
                        {"version": "1.0.0", "status": {"name": "p2", "state": "pendingCanceling"}},
                        {"status": {"name": "p3", "state": "pendingCanceled"}},
                        {"version": "1.0.0", "status": {"name": "p4", "state": "pendingCancelFailed"}},

                        {"version": "1.0.0", "status": {"name": "d1", "state": "downloading"}},
                        {"version": "1.0.0", "status": {"name": "d2", "state": "downloadingCanceling"}},
                        {"status": {"name": "d3", "state": "downloadingCanceled"}},
                        {"version": "1.0.0", "status": {"name": "d4", "state": "downloadingCancelFailed"}},
                        {"status": {"name": "d5", "state": "downloadFailed"}},

                        {"version": "1.0.0", "status": {"name": "i1", "state": "installing"}},
                        {"version": "1.0.0", "status": {"name": "i2", "state": "installingCanceling"}},
                        {"status": {"name": "i3", "state": "installingCanceled"}},
                        {"version": "1.0.0", "status": {"name": "i4", "state": "installingCancelFailed"}},
                        {"status": {"name": "i5", "state": "installFailed"}},

                        {"status": {"name": "u1", "state": "uninstalled"}},

                        {"version": "1.0.0", "status": {"name": "init", "state": "initializing"}},
                        {"version": "1.0.0", "status": {"name": "initcanc", "state": "initializingCanceling"}},
                        {"version": "1.0.0", "status": {"name": "running", "state": "running"}}
                    ]}
                }
            }
        }`,
		data: `{"user_data": {"sources": {}}}`,
	})
	mc := newTestMarketClient(t, srv.URL)

	got, err := fetchInstalledApps(mc, "", true)
	if err != nil {
		t.Fatalf("fetchInstalledApps: %v", err)
	}

	// Every row whose state is NOT in the SPA's `uninstalledAppStates`
	// must survive. The 6 SPA-filtered rows above (`p3`, `d3`, `d5`,
	// `i3`, `i5`, `u1`) drop out; the remaining 12 stay. Sort order is
	// alphabetical-by-name within a single source.
	want := []AppDisplayInfo{
		{Name: "d1", Source: "market.olares", Version: "1.0.0", State: "downloading"},
		{Name: "d2", Source: "market.olares", Version: "1.0.0", State: "downloadingCanceling"},
		{Name: "d4", Source: "market.olares", Version: "1.0.0", State: "downloadingCancelFailed"},
		{Name: "i1", Source: "market.olares", Version: "1.0.0", State: "installing"},
		{Name: "i2", Source: "market.olares", Version: "1.0.0", State: "installingCanceling"},
		{Name: "i4", Source: "market.olares", Version: "1.0.0", State: "installingCancelFailed"},
		{Name: "init", Source: "market.olares", Version: "1.0.0", State: "initializing"},
		{Name: "initcanc", Source: "market.olares", Version: "1.0.0", State: "initializingCanceling"},
		{Name: "p1", Source: "market.olares", Version: "1.0.0", State: "pending"},
		{Name: "p2", Source: "market.olares", Version: "1.0.0", State: "pendingCanceling"},
		{Name: "p4", Source: "market.olares", Version: "1.0.0", State: "pendingCancelFailed"},
		{Name: "running", Source: "market.olares", Version: "1.0.0", State: "running"},
	}
	if !reflect.DeepEqual(got, want) {
		t.Fatalf("SPA-aligned filter mismatch:\n got: %#v\nwant: %#v", got, want)
	}
}

// TestFetchInstalledAppsClonesEnrichViaRawAppName locks down the clone
// lookup fix: a cloned app row carries the clone's unique `name`
// (e.g. windowsefe992) plus the source app's name as `rawAppName`
// (e.g. windows). The catalog only knows `windows`, so a naive
// name-based lookup misses every clone and they render with no
// categories. The parser must fall back to `rawAppName` for the
// catalog lookup, while still using the clone's `name` for display.
func TestFetchInstalledAppsClonesEnrichViaRawAppName(t *testing.T) {
	srv := newFakeMarketDataServer(t, stateAndDataResponses{
		state: `{
            "user_data": {
                "sources": {
                    "market.olares": {"type": "market", "app_state_latest": [
                        {"version": "1.1.4", "status": {
                            "name": "windows", "state": "running"
                        }},
                        {"version": "1.1.4", "status": {
                            "name": "windowsefe992",
                            "rawAppName": "windows",
                            "title": "windows c1",
                            "state": "stopped"
                        }}
                    ]}
                }
            }
        }`,
		data: `{
            "user_data": {
                "sources": {
                    "market.olares": {"type": "market", "app_info_latest": [
                        {"timestamp": 1, "version": "1.1.4", "app_simple_info": {
                            "app_name": "windows",
                            "app_title": "Windows",
                            "categories": ["Utilities"]
                        }}
                    ]}
                }
            }
        }`,
	})
	mc := newTestMarketClient(t, srv.URL)

	got, err := fetchInstalledApps(mc, "", true)
	if err != nil {
		t.Fatalf("fetchInstalledApps: %v", err)
	}

	want := []AppDisplayInfo{
		{
			Name:       "windows",
			Title:      "Windows",
			Version:    "1.1.4",
			Source:     "market.olares",
			Categories: []string{"Utilities"},
			State:      "running",
		},
		{
			// Display name keeps the clone's unique identifier and the
			// row's own title; categories MUST come from the source
			// app (windows) via the rawAppName lookup.
			Name:       "windowsefe992",
			Title:      "windows c1",
			Version:    "1.1.4",
			Source:     "market.olares",
			Categories: []string{"Utilities"},
			State:      "stopped",
		},
	}
	if !reflect.DeepEqual(got, want) {
		t.Fatalf("clone enrichment mismatch:\n got: %#v\nwant: %#v", got, want)
	}
}

// TestFetchInstalledAppsScopedToSource confirms that passing showAll=false
// + a source name only returns rows belonging to that source. The studio
// row is deliberately in `installFailed` state (one of the six SPA
// `uninstalledAppStates`) so the test also re-proves the state filter
// is honored on the scoped path — `ghost` is dropped both by source
// scope and by the state filter, leaving only the `cli` row.
func TestFetchInstalledAppsScopedToSource(t *testing.T) {
	srv := newFakeMarketDataServer(t, stateAndDataResponses{
		state: `{
            "user_data": {
                "sources": {
                    "market.olares": {"type": "market", "app_state_latest": [
                        {"version": "1.0.0", "status": {"name": "firefox", "state": "running"}}
                    ]},
                    "cli": {"type": "local", "app_state_latest": [
                        {"version": "0.9.0", "status": {"name": "myapp", "state": "running"}}
                    ]},
                    "studio": {"type": "local", "app_state_latest": [
                        {"status": {"name": "ghost", "state": "installFailed"}}
                    ]}
                }
            }
        }`,
		data: `{"user_data": {"sources": {}}}`,
	})
	mc := newTestMarketClient(t, srv.URL)

	got, err := fetchInstalledApps(mc, "cli", false)
	if err != nil {
		t.Fatalf("fetchInstalledApps: %v", err)
	}

	want := []AppDisplayInfo{
		{Name: "myapp", Source: "cli", State: "running", Version: "0.9.0"},
	}
	if !reflect.DeepEqual(got, want) {
		t.Fatalf("scoped fetchInstalledApps mismatch:\n got: %#v\nwant: %#v", got, want)
	}
}

// TestFetchInstalledAppsCatalogFailureFallback proves that a broken
// /market/data does not blank out the "my apps" listing: enrichment
// just degrades to "name + state + source + whatever title the state
// row already carries" for every row.
func TestFetchInstalledAppsCatalogFailureFallback(t *testing.T) {
	srv := newFakeMarketDataServer(t, stateAndDataResponses{
		state: `{
            "user_data": {
                "sources": {
                    "market.olares": {"type": "market", "app_state_latest": [
                        {"version": "1.1.0", "status": {"name": "firefox", "title": "Firefox", "state": "running"}}
                    ]}
                }
            }
        }`,
		dataStatus: http.StatusInternalServerError,
		data:       `{"success": false, "message": "catalog unreachable"}`,
	})
	mc := newTestMarketClient(t, srv.URL)

	got, err := fetchInstalledApps(mc, "", true)
	if err != nil {
		t.Fatalf("fetchInstalledApps: %v", err)
	}
	// Even with the catalog broken, the row's version still surfaces
	// because it comes from /market/state, not /market/data.
	want := []AppDisplayInfo{
		{Name: "firefox", Title: "Firefox", Version: "1.1.0", Source: "market.olares", State: "running"},
	}
	if !reflect.DeepEqual(got, want) {
		t.Fatalf("fallback fetchInstalledApps mismatch:\n got: %#v\nwant: %#v", got, want)
	}
}

// marketStateFixture and marketDataFixture are extracted so both the
// happy-path test and any future scenarios can reuse the same shape
// without copy/pasting JSON literals.
func marketStateFixture() string {
	// firefox carries version 1.1.0 at the AppStateLatest level (the
	// version on the user's state row). The catalog fixture
	// intentionally advertises 1.2.3 as the latest available — the
	// test asserts the CLI surfaces the state's 1.1.0, not the
	// catalog's 1.2.3.
	return `{
        "user_data": {
            "sources": {
                "market.olares": {
                    "type": "market",
                    "app_state_latest": [
                        {"type": "app", "version": "1.1.0", "status": {"name": "firefox", "title": "Firefox", "state": "running", "opType": "install"}},
                        {"type": "app", "version": "2.0.0", "status": {"name": "obsidian", "state": "uninstalled"}},
                        {"type": "app", "version": "0.5.0", "status": {"name": "kuma", "state": "upgradeFailed"}}
                    ]
                },
                "cli": {
                    "type": "local",
                    "app_state_latest": [
                        {"type": "app", "status": {"name": "myapp", "state": "running"}}
                    ]
                },
                "studio": {
                    "type": "local",
                    "app_state_latest": [
                        {"type": "app", "status": {"name": "ghost", "state": "installFailed"}}
                    ]
                }
            },
            "hash": "h1"
        }
    }`
}

func marketDataFixture() string {
	// /market/data is the CATALOG: it reflects the latest version
	// available upstream. firefox is at 1.2.3 here even though the
	// state fixture says the user's row is on 1.1.0 — that gap is
	// the exact scenario the parser must surface as the row's
	// version, not the catalog one.
	return `{
        "user_data": {
            "sources": {
                "market.olares": {
                    "type": "market",
                    "app_info_latest": [
                        {"timestamp": 1, "version": "1.2.3", "app_simple_info": {
                            "app_name": "firefox",
                            "app_title": {"en-US": "Firefox", "zh-CN": "huoxudou"},
                            "categories": ["Web", "Browser"]
                        }},
                        {"timestamp": 1, "version": "0.5.0", "app_simple_info": {
                            "app_name": "kuma",
                            "app_title": "Kuma",
                            "categories": ["Monitoring"]
                        }}
                    ]
                }
            },
            "hash": "h2"
        }
    }`
}

// stateAndDataResponses wires both endpoints into one server so the
// test cases stay readable; either field can be left blank to fall back
// to a permissive default.
type stateAndDataResponses struct {
	state       string
	stateStatus int
	data        string
	dataStatus  int
}

// newFakeMarketDataServer stands up a tiny httptest.Server that answers
// the two endpoints fetchInstalledApps reaches (/market/state and
// /market/data). The handler intentionally returns the raw `data`
// envelope expected by APIResponse so the real MarketClient parses it
// end-to-end.
func newFakeMarketDataServer(t *testing.T, resp stateAndDataResponses) *httptest.Server {
	t.Helper()
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		switch {
		case strings.HasSuffix(r.URL.Path, "/market/state"):
			status := resp.stateStatus
			if status == 0 {
				status = http.StatusOK
			}
			w.WriteHeader(status)
			body := resp.state
			if body == "" {
				body = `{"user_data": {"sources": {}}}`
			}
			_, _ = w.Write([]byte(wrapEnvelope(status == http.StatusOK, body)))
		case strings.HasSuffix(r.URL.Path, "/market/data"):
			status := resp.dataStatus
			if status == 0 {
				status = http.StatusOK
			}
			w.WriteHeader(status)
			body := resp.data
			if body == "" {
				body = `{"user_data": {"sources": {}}}`
			}
			if status != http.StatusOK {
				_, _ = w.Write([]byte(body))
				return
			}
			_, _ = w.Write([]byte(wrapEnvelope(true, body)))
		default:
			http.NotFound(w, r)
		}
	}))
	t.Cleanup(srv.Close)
	return srv
}

// wrapEnvelope mirrors the canonical APIResponse{success,message,data}
// shape the market backend wraps every response in; the body argument
// is dropped verbatim into the `data` field (no extra quoting).
func wrapEnvelope(success bool, dataJSON string) string {
	successStr := "true"
	if !success {
		successStr = "false"
	}
	return `{"success":` + successStr + `,"message":"","data":` + dataJSON + `}`
}
