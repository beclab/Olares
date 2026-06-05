package market

import (
	"context"
	"strings"
	"testing"
)

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
