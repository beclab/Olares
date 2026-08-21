package search

import (
	"context"
	"encoding/json"
	"reflect"
	"testing"

	"github.com/spf13/viper"

	"github.com/beclab/Olares/cli/pkg/cmdutil"
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
