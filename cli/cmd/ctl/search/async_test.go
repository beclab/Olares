package search

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net"
	"net/http"
	"net/http/httptest"
	"os"
	"reflect"
	"strings"
	"sync/atomic"
	"syscall"
	"testing"

	"github.com/gorilla/websocket"
	"github.com/spf13/viper"

	"github.com/beclab/Olares/cli/pkg/cmdutil"
	"github.com/beclab/Olares/cli/pkg/credential"
)

func TestSearchWebSocketURL(t *testing.T) {
	t.Parallel()
	cases := []struct {
		in   string
		want string
	}{
		{"https://desktop.alice.olares.com", "wss://desktop.alice.olares.com/ws"},
		{"http://127.0.0.1:8080/base/", "ws://127.0.0.1:8080/base/ws"},
	}
	for _, tc := range cases {
		got, err := searchWebSocketURL(tc.in)
		if err != nil {
			t.Fatalf("searchWebSocketURL(%q): %v", tc.in, err)
		}
		if got != tc.want {
			t.Fatalf("searchWebSocketURL(%q) = %q, want %q", tc.in, got, tc.want)
		}
	}
}

func TestAsyncHitCollectorDeduplicatesAndOrders(t *testing.T) {
	t.Parallel()
	c := newAsyncHitCollector()
	c.addBatch(appFilesV2, 1, []json.RawMessage{json.RawMessage(`{"title":"one"}`), json.RawMessage(`{"title":"two"}`)})
	c.addBatch(appFilesV2, 0, []json.RawMessage{json.RawMessage(`{"title":"zero"}`), json.RawMessage(`{"title":"duplicate"}`)})

	rows := c.rows([]string{appFilesV2})
	if len(rows) != 3 {
		t.Fatalf("rows = %d, want 3", len(rows))
	}
	var titles []string
	for _, row := range rows {
		var v struct {
			Title string `json:"title"`
		}
		if err := json.Unmarshal(row, &v); err != nil {
			t.Fatal(err)
		}
		titles = append(titles, v.Title)
	}
	if want := []string{"zero", "one", "two"}; !reflect.DeepEqual(titles, want) {
		t.Fatalf("titles = %#v, want %#v", titles, want)
	}
}

func TestAsyncHitCollectorKeepsRowsAfterGap(t *testing.T) {
	t.Parallel()
	c := newAsyncHitCollector()
	c.addBatch(appFilesV2, 0, []json.RawMessage{json.RawMessage(`{"title":"zero"}`)})
	c.addBatch(appFilesV2, 2, []json.RawMessage{json.RawMessage(`{"title":"two"}`)})

	rows := c.rows([]string{appFilesV2})
	if len(rows) != 2 {
		t.Fatalf("rows = %d, want 2", len(rows))
	}
}

func TestAsyncHitCollectorMissingRanges(t *testing.T) {
	t.Parallel()
	c := newAsyncHitCollector()
	c.addBatch(appFilesV2, 1, []json.RawMessage{json.RawMessage(`{}`)})
	c.addBatch(appFilesV2, 4, []json.RawMessage{json.RawMessage(`{}`), json.RawMessage(`{}`)})
	got := c.missingRanges(appFilesV2, 8)
	want := [][]int{{0, 0}, {2, 3}, {6, 7}}
	if !reflect.DeepEqual(got, want) {
		t.Fatalf("missingRanges = %#v, want %#v", got, want)
	}
}

func TestAsyncSearchVersionBoundary(t *testing.T) {
	previous := viper.GetString(cmdutil.FlagOlaresVersion)
	t.Cleanup(func() { viper.Set(cmdutil.FlagOlaresVersion, previous) })

	cases := []struct {
		version string
		want    bool
	}{
		{"1.12.6", false},
		{"1.12.6-20260817", false},
		{"1.12.7", true},
		{"1.12.7-20260817", true},
		{"1.13.0", true},
	}
	for _, tc := range cases {
		viper.Set(cmdutil.FlagOlaresVersion, tc.version)
		got, err := (&cmdutil.Factory{}).OlaresBackendAtLeast(context.Background(), asyncSearchMinOlaresVersion)
		if err != nil {
			t.Fatalf("version %s: %v", tc.version, err)
		}
		if got != tc.want {
			t.Fatalf("version %s: async = %v, want %v", tc.version, got, tc.want)
		}
	}
}

func TestFederatedFileSources(t *testing.T) {
	t.Parallel()
	want := []string{appFilesV2, appGoogleDrive, appDropbox, appSeafile}
	if !reflect.DeepEqual(federatedFileSources, want) {
		t.Fatalf("federatedFileSources = %#v, want %#v", federatedFileSources, want)
	}
}

// `search sync` must stay a narrowed view of what `search drive` already
// covers; a source here that drive does not request would mean the two
// commands report different libraries for the same keyword.
func TestSyncSearchSourcesAreCoveredByDrive(t *testing.T) {
	t.Parallel()
	if want := []string{appSeafile}; !reflect.DeepEqual(syncSearchSources, want) {
		t.Fatalf("syncSearchSources = %#v, want %#v", syncSearchSources, want)
	}
	for _, source := range syncSearchSources {
		found := false
		for _, federated := range federatedFileSources {
			if federated == source {
				found = true
				break
			}
		}
		if !found {
			t.Fatalf("source %q is not part of federatedFileSources %#v", source, federatedFileSources)
		}
	}
}

// newFakeSearchWS serves the Desktop /ws endpoint, refusing the first
// failures upgrades with failStatus and answering the login on the rest.
func newFakeSearchWS(t *testing.T, failures int, failStatus int) (*httptest.Server, *atomic.Int32) {
	t.Helper()
	var upgrader websocket.Upgrader
	var attempts atomic.Int32
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if int(attempts.Add(1)) <= failures {
			http.Error(w, "upstream unavailable", failStatus)
			return
		}
		conn, err := upgrader.Upgrade(w, r, nil)
		if err != nil {
			return
		}
		defer conn.Close()
		if _, _, err := conn.ReadMessage(); err != nil {
			return
		}
		if err := conn.WriteJSON(map[string]string{"event": "pong"}); err != nil {
			return
		}
		for {
			if _, _, err := conn.ReadMessage(); err != nil {
				return
			}
		}
	}))
	t.Cleanup(server.Close)
	return server, &attempts
}

func TestOpenSearchWebSocketRetriesTransientHandshake(t *testing.T) {
	server, attempts := newFakeSearchWS(t, 1, http.StatusBadGateway)
	conn, err := openSearchWebSocket(context.Background(), &credential.ResolvedProfile{DesktopURL: server.URL}, "token")
	if err != nil {
		t.Fatalf("openSearchWebSocket: %v", err)
	}
	defer conn.Close()
	if got := attempts.Load(); got != 2 {
		t.Fatalf("upgrade attempts = %d, want 2 (one 502 + one retry)", got)
	}
}

// An upgrade the server refused deliberately must fail on the spot: retrying a
// rejected session only delays the "run profile login" the user needs to see.
func TestOpenSearchWebSocketDoesNotRetryFinalStatus(t *testing.T) {
	cases := []struct {
		status     int
		wantErrHas string
	}{
		{http.StatusUnauthorized, "authentication failed"},
		{459, "authentication failed"},
		{http.StatusNotFound, "handshake failed (HTTP 404)"},
	}
	for _, tc := range cases {
		server, attempts := newFakeSearchWS(t, 1, tc.status)
		conn, err := openSearchWebSocket(context.Background(), &credential.ResolvedProfile{DesktopURL: server.URL}, "token")
		if err == nil {
			conn.Close()
			t.Fatalf("status %d: expected error", tc.status)
		}
		if !strings.Contains(err.Error(), tc.wantErrHas) {
			t.Fatalf("status %d: error = %q, want it to contain %q", tc.status, err, tc.wantErrHas)
		}
		if got := attempts.Load(); got != 1 {
			t.Fatalf("status %d: upgrade attempts = %d, want 1", tc.status, got)
		}
	}
}

func TestOpenSearchWebSocketGivesUpAfterAttemptLimit(t *testing.T) {
	server, attempts := newFakeSearchWS(t, asyncSearchDialAttempts, http.StatusServiceUnavailable)
	conn, err := openSearchWebSocket(context.Background(), &credential.ResolvedProfile{DesktopURL: server.URL}, "token")
	if err == nil {
		conn.Close()
		t.Fatal("expected error")
	}
	if got := attempts.Load(); int(got) != asyncSearchDialAttempts {
		t.Fatalf("upgrade attempts = %d, want %d", got, asyncSearchDialAttempts)
	}
	if !strings.Contains(err.Error(), fmt.Sprintf("after %d attempts", asyncSearchDialAttempts)) {
		t.Fatalf("error = %q, want it to report the attempt count", err)
	}
}

func TestIsTransientSearchDialError(t *testing.T) {
	t.Parallel()
	cases := []struct {
		name string
		err  error
		want bool
	}{
		{"handshake timeout", context.DeadlineExceeded, true},
		{"tcp i/o timeout", &net.OpError{Op: "dial", Err: os.ErrDeadlineExceeded}, true},
		{"connection refused", &net.OpError{Op: "dial", Err: syscall.ECONNREFUSED}, true},
		{"connection reset", fmt.Errorf("read: %w", syscall.ECONNRESET), true},
		{"truncated response", io.ErrUnexpectedEOF, true},
		{"bad gateway", &searchHandshakeError{status: http.StatusBadGateway}, true},
		{"not found", &searchHandshakeError{status: http.StatusNotFound}, false},
		{"unsupported scheme", errors.New("unsupported Desktop URL scheme"), false},
		{"cancelled", context.Canceled, false},
	}
	for _, tc := range cases {
		if got := isTransientSearchDialError(tc.err); got != tc.want {
			t.Fatalf("%s: isTransientSearchDialError = %v, want %v", tc.name, got, tc.want)
		}
	}
}

func TestAsyncSearchTerminalError(t *testing.T) {
	t.Parallel()
	for _, status := range []string{"completed", "completed_with_partial_failures"} {
		if err := asyncSearchTerminalError(asyncSearchMessage{JobStatus: status}); err != nil {
			t.Fatalf("status %s: unexpected error: %v", status, err)
		}
	}
	for _, status := range []string{"failed", "timeout", "cancelled"} {
		if err := asyncSearchTerminalError(asyncSearchMessage{JobStatus: status}); err == nil {
			t.Fatalf("status %s: expected error", status)
		}
	}
}
