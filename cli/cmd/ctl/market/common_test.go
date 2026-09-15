package market

import (
	"context"
	"errors"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

// An app lookup that comes back empty has two causes with different next
// steps, and the backend answers both the same way. A mistyped -s used to read
// as "app 'clitest' not found in source 'bogussource'", which describes a
// missing app rather than a source that was never there.
func TestFetchAppInfoSeparatesUnknownSourceFromMissingApp(t *testing.T) {
	const sources = `{
        "user_data": {
            "sources": {
                "market.olares": {"type": "market", "app_info_latest": []},
                "market.test":   {"type": "market", "app_info_latest": []},
                "upload":        {"type": "local",  "app_info_latest": []}
            }
        }
    }`

	t.Run("unknown source names the ones that exist", func(t *testing.T) {
		srv := newAppLookupServer(t, sources, `{"apps": []}`)
		_, err := fetchAppInfo(context.Background(), newTestMarketClient(t, srv.URL), "clitest", "bogussource")

		var unknown *unknownSourceError
		if !errors.As(err, &unknown) {
			t.Fatalf("error = %v, want unknownSourceError", err)
		}
		// market.test is the one a hand-written list omitted, so its presence
		// here is what says the list came from the cluster.
		for _, want := range []string{"bogussource", "market.olares", "market.test", "upload"} {
			if !strings.Contains(err.Error(), want) {
				t.Fatalf("error %q does not mention %q", err.Error(), want)
			}
		}
	})

	t.Run("real source with no such app stays an app error", func(t *testing.T) {
		srv := newAppLookupServer(t, sources, `{"apps": []}`)
		_, err := fetchAppInfo(context.Background(), newTestMarketClient(t, srv.URL), "clitest", "market.test")

		var notInSource *appNotInSourceError
		if !errors.As(err, &notInSource) {
			t.Fatalf("error = %v, want appNotInSourceError", err)
		}
	})

	// The source list is a best-effort second call; losing it must not turn a
	// missing app into a claim about the source.
	t.Run("unreachable source list falls back to the app error", func(t *testing.T) {
		srv := newAppLookupServer(t, "", `{"apps": []}`)
		_, err := fetchAppInfo(context.Background(), newTestMarketClient(t, srv.URL), "clitest", "bogussource")

		var notInSource *appNotInSourceError
		if !errors.As(err, &notInSource) {
			t.Fatalf("error = %v, want appNotInSourceError", err)
		}
	})
}

// "cannot determine version ... (use --version to specify)" is the right frame
// only for a failure --version would actually change. It is wrong for a missing
// app, and on delete it is harmful: --version does not narrow a delete, and the
// backend reports deleting an app it never held as success, so following the
// advice converts a correct exit 1 into a false success.
func TestResolveVersionInSourceFramesFailureByCause(t *testing.T) {
	const sources = `{"user_data": {"sources": {"upload": {"type": "local", "app_info_latest": []}}}}`

	t.Run("missing app is reported as itself", func(t *testing.T) {
		srv := newAppLookupServer(t, sources, `{"apps": []}`)
		_, err := resolveVersionInSource(newTestMarketClient(t, srv.URL), "clitest", "upload", true)

		if strings.Contains(err.Error(), "cannot determine version") {
			t.Fatalf("error %q buries the cause under the version wrapper", err)
		}
		if !strings.Contains(err.Error(), "not found in source 'upload'") {
			t.Fatalf("error = %q", err)
		}
	})

	t.Run("delete never suggests --version", func(t *testing.T) {
		// A malformed payload is a failure --version could plausibly work
		// around, so it reaches the wrapper on both verbs; only the hint differs.
		srv := newAppLookupServer(t, sources, `{"apps": [["not-an-object"]]}`)
		mc := newTestMarketClient(t, srv.URL)

		_, installErr := resolveVersionInSource(mc, "clitest", "upload", true)
		if !strings.Contains(installErr.Error(), "use --version to specify") {
			t.Fatalf("install error %q dropped the hint", installErr)
		}
		_, deleteErr := resolveVersionInSource(mc, "clitest", "upload", false)
		if strings.Contains(deleteErr.Error(), "--version") {
			t.Fatalf("delete error %q suggests --version", deleteErr)
		}
		if !strings.Contains(deleteErr.Error(), "cannot determine version") {
			t.Fatalf("delete error %q lost the cause", deleteErr)
		}
	})
}

// newAppLookupServer answers POST /apps with appsJSON and GET /market/data with
// sourcesJSON; an empty sourcesJSON makes the source list unavailable.
func newAppLookupServer(t *testing.T, sourcesJSON, appsJSON string) *httptest.Server {
	t.Helper()
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		switch {
		case strings.HasSuffix(r.URL.Path, "/apps"):
			_, _ = w.Write([]byte(wrapEnvelope(true, appsJSON)))
		case strings.HasSuffix(r.URL.Path, "/market/data"):
			if sourcesJSON == "" {
				w.WriteHeader(http.StatusInternalServerError)
				return
			}
			_, _ = w.Write([]byte(wrapEnvelope(true, sourcesJSON)))
		default:
			http.NotFound(w, r)
		}
	}))
	t.Cleanup(srv.Close)
	return srv
}

// The suggested remediation command has to be runnable as printed. Without -s,
// `market get` resolves market.olares, which is not where an uploaded app is.
func TestEnvValidationErrorPinsTheSourceInItsRemediation(t *testing.T) {
	withSource := (&envValidationError{
		AppName:       "clitestenv",
		Source:        "upload",
		MissingValues: []string{"CLITEST_TOKEN"},
	}).Error()

	if !strings.Contains(withSource, "market get clitestenv -s upload'") {
		t.Fatalf("error %q does not pin the source in the suggested command", withSource)
	}

	// A verb with no resolved source must not print a bare -s.
	withoutSource := (&envValidationError{
		AppName:       "clitestenv",
		MissingValues: []string{"CLITEST_TOKEN"},
	}).Error()

	if !strings.Contains(withoutSource, "market get clitestenv'") {
		t.Fatalf("error %q is malformed without a source", withoutSource)
	}
}

// TestResolveInstalledSource verifies the source-resolution guard for the
// implicit-source verbs (stop / resume / uninstall). It mirrors the SPA's
// appStore.findAppByName(): a state row is only a valid source when it is
// "installed" (`!uninstalledApp(status)`), an explicit --source bypasses the
// /market/state lookup entirely, and a missing / terminal-not-installed row
// fails fast with a clear message instead of dispatching a doomed request.
func TestResolveInstalledSource(t *testing.T) {
	const runningState = `{
        "user_data": {
            "sources": {
                "market.olares": {
                    "type": "market",
                    "app_state_latest": [
                        {"version": "1.0.12", "status": {"name": "firefox", "rawAppName": "firefox", "state": "running"}}
                    ]
                }
            }
        }
    }`

	t.Run("running app resolves its source", func(t *testing.T) {
		srv := newFakeMarketDataServer(t, stateAndDataResponses{state: runningState})
		mc := newTestMarketClient(t, srv.URL)

		got, err := resolveInstalledSource(context.Background(), &MarketOptions{}, mc, "firefox")
		if err != nil {
			t.Fatalf("resolveInstalledSource: %v", err)
		}
		if got != "market.olares" {
			t.Fatalf("source = %q, want market.olares", got)
		}
	})

	t.Run("same name in two sources resolves to the installed one", func(t *testing.T) {
		// firefox lingers as `uninstalled` in market.olares but is
		// `stopped` (installed) in market.test. resolveInstalledSource
		// must resolve the live instance's source so `resume firefox`
		// works, instead of failing on the uninstalled row.
		const multiSource = `{
            "user_data": {
                "sources": {
                    "market.olares": {
                        "type": "market",
                        "app_state_latest": [
                            {"version": "1.0.11", "status": {"name": "firefox", "rawAppName": "firefox", "state": "uninstalled"}}
                        ]
                    },
                    "market.test": {
                        "type": "market",
                        "app_state_latest": [
                            {"version": "1.2.11", "status": {"name": "firefox", "rawAppName": "firefox", "state": "stopped"}}
                        ]
                    }
                }
            }
        }`
		srv := newFakeMarketDataServer(t, stateAndDataResponses{state: multiSource})
		mc := newTestMarketClient(t, srv.URL)

		got, err := resolveInstalledSource(context.Background(), &MarketOptions{}, mc, "firefox")
		if err != nil {
			t.Fatalf("resolveInstalledSource on multi-source app: %v", err)
		}
		if got != "market.test" {
			t.Fatalf("source = %q, want market.test (the installed instance)", got)
		}
	})

	t.Run("explicit --source bypasses the state lookup", func(t *testing.T) {
		// Point at a server that has NO matching row; an explicit source
		// must still win without consulting /market/state.
		srv := newFakeMarketDataServer(t, stateAndDataResponses{})
		mc := newTestMarketClient(t, srv.URL)

		got, err := resolveInstalledSource(context.Background(), &MarketOptions{Source: "  custom.source  "}, mc, "firefox")
		if err != nil {
			t.Fatalf("resolveInstalledSource with explicit source: %v", err)
		}
		if got != "custom.source" {
			t.Fatalf("source = %q, want custom.source (trimmed)", got)
		}
	})

	t.Run("no installed row fails fast", func(t *testing.T) {
		srv := newFakeMarketDataServer(t, stateAndDataResponses{})
		mc := newTestMarketClient(t, srv.URL)

		_, err := resolveInstalledSource(context.Background(), &MarketOptions{}, mc, "firefox")
		if err == nil {
			t.Fatalf("expected an error for an app with no state row")
		}
		if !strings.Contains(err.Error(), "is not installed for this user") {
			t.Fatalf("error = %q, want it to say the app is not installed", err.Error())
		}
	})

	// Every state in the SPA's uninstalledAppStates set (mirrored by
	// notInstalledStates in types.go) must be treated as "not installed":
	// the row may linger in /market/state, but findAppByName would skip it,
	// so the CLI must not hand its source to a stop/resume/uninstall call.
	for _, state := range []string{
		"pendingCanceled",
		"downloadingCanceled",
		"downloadFailed",
		"installFailed",
		"installingCanceled",
		"uninstalled",
	} {
		state := state
		t.Run("not-installed state "+state+" is rejected", func(t *testing.T) {
			stateJSON := `{
                "user_data": {
                    "sources": {
                        "market.olares": {
                            "type": "market",
                            "app_state_latest": [
                                {"version": "1.0.12", "status": {"name": "firefox", "rawAppName": "firefox", "state": "` + state + `"}}
                            ]
                        }
                    }
                }
            }`
			srv := newFakeMarketDataServer(t, stateAndDataResponses{state: stateJSON})
			mc := newTestMarketClient(t, srv.URL)

			_, err := resolveInstalledSource(context.Background(), &MarketOptions{}, mc, "firefox")
			if err == nil {
				t.Fatalf("state %q: expected a not-installed error", state)
			}
			if !strings.Contains(err.Error(), "not an installed app") {
				t.Fatalf("state %q: error = %q, want it to say not an installed app", state, err.Error())
			}
		})
	}
}
