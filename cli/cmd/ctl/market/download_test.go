package market

import (
	"context"
	"fmt"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

// chartBytes stands in for a real .tgz: the CLI never parses the payload, it
// only has to arrive on disk byte-for-byte.
var chartBytes = []byte("\x1f\x8b\x08 not really gzip, but opaque to the CLI")

type packageRequest struct {
	path    string
	source  string
	version string
	accept  string
}

// newFakeChartPackageServer answers GET /apps/{name}/package with chart bytes
// and records what was asked for. filename is sent through
// Content-Disposition; pass "" to omit the header.
func newFakeChartPackageServer(t *testing.T, filename string, seen *packageRequest) *httptest.Server {
	t.Helper()
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if !strings.HasSuffix(r.URL.Path, "/package") {
			http.NotFound(w, r)
			return
		}
		if seen != nil {
			*seen = packageRequest{
				path:    r.URL.Path,
				source:  r.URL.Query().Get("source"),
				version: r.URL.Query().Get("version"),
				accept:  r.Header.Get("Accept"),
			}
		}
		w.Header().Set("Content-Type", "application/octet-stream")
		if filename != "" {
			w.Header().Set("Content-Disposition", fmt.Sprintf("attachment; filename=%q", filename))
		}
		w.Header().Set("Content-Length", fmt.Sprintf("%d", len(chartBytes)))
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write(chartBytes)
	}))
	t.Cleanup(srv.Close)
	return srv
}

// newFakeChartErrorServer answers with the app-store JSON envelope, the way
// the backend reports a chart it does not hold.
func newFakeChartErrorServer(t *testing.T, status int, message string) *httptest.Server {
	t.Helper()
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(status)
		fmt.Fprintf(w, `{"success":false,"message":%q}`, message)
	}))
	t.Cleanup(srv.Close)
	return srv
}

func TestDownloadChartWritesServerSuggestedFilename(t *testing.T) {
	var seen packageRequest
	srv := newFakeChartPackageServer(t, "amule-0.0.7.tgz", &seen)
	mc := newTestMarketClient(t, srv.URL)

	dir := t.TempDir()
	result, err := fetchChartToDisk(context.Background(), mc, "amule", "upload", "", dir, false)
	if err != nil {
		t.Fatalf("fetchChartToDisk: %v", err)
	}

	want := filepath.Join(dir, "amule-0.0.7.tgz")
	if result.Path != want {
		t.Fatalf("path = %q, want %q", result.Path, want)
	}
	got, err := os.ReadFile(want)
	if err != nil {
		t.Fatalf("read downloaded chart: %v", err)
	}
	if string(got) != string(chartBytes) {
		t.Fatalf("chart bytes differ: got %q", got)
	}
	if result.Bytes != int64(len(chartBytes)) {
		t.Fatalf("bytes = %d, want %d", result.Bytes, len(chartBytes))
	}
	// Version is unreported by the caller here, so it has to come back from
	// the filename the server chose.
	if result.Version != "0.0.7" {
		t.Fatalf("version = %q, want 0.0.7", result.Version)
	}
	if seen.path != "/app-store/api/v2/apps/amule/package" {
		t.Fatalf("path = %q", seen.path)
	}
	if seen.source != "upload" {
		t.Fatalf("source = %q, want upload", seen.source)
	}
	if seen.version != "" {
		t.Fatalf("version query = %q, want it omitted so the server picks the current one", seen.version)
	}
	if !strings.Contains(seen.accept, "application/octet-stream") {
		t.Fatalf("Accept = %q, want it to allow octet-stream", seen.accept)
	}
	if _, err := os.Stat(want + ".tmp"); !os.IsNotExist(err) {
		t.Fatalf("temp file was left behind: %v", err)
	}
}

func TestDownloadChartForwardsSourceAndVersion(t *testing.T) {
	var seen packageRequest
	srv := newFakeChartPackageServer(t, "amule-0.0.5.tgz", &seen)
	mc := newTestMarketClient(t, srv.URL)

	result, err := fetchChartToDisk(context.Background(), mc, "amule", "market.olares", "0.0.5", t.TempDir(), false)
	if err != nil {
		t.Fatalf("fetchChartToDisk: %v", err)
	}
	if seen.source != "market.olares" {
		t.Fatalf("source = %q, want market.olares", seen.source)
	}
	if seen.version != "0.0.5" {
		t.Fatalf("version = %q, want 0.0.5", seen.version)
	}
	if result.Version != "0.0.5" {
		t.Fatalf("result version = %q, want 0.0.5", result.Version)
	}
}

func TestDownloadChartLocalPathModes(t *testing.T) {
	srv := newFakeChartPackageServer(t, "amule-0.0.7.tgz", nil)
	mc := newTestMarketClient(t, srv.URL)

	t.Run("explicit file path is used verbatim", func(t *testing.T) {
		dst := filepath.Join(t.TempDir(), "custom.tgz")
		result, err := fetchChartToDisk(context.Background(), mc, "amule", "upload", "", dst, false)
		if err != nil {
			t.Fatalf("fetchChartToDisk: %v", err)
		}
		if result.Path != dst {
			t.Fatalf("path = %q, want %q", result.Path, dst)
		}
	})

	t.Run("trailing slash creates the directory", func(t *testing.T) {
		dir := filepath.Join(t.TempDir(), "charts") + "/"
		result, err := fetchChartToDisk(context.Background(), mc, "amule", "upload", "", dir, false)
		if err != nil {
			t.Fatalf("fetchChartToDisk: %v", err)
		}
		want := filepath.Join(strings.TrimSuffix(dir, "/"), "amule-0.0.7.tgz")
		if result.Path != want {
			t.Fatalf("path = %q, want %q", result.Path, want)
		}
		if _, err := os.Stat(want); err != nil {
			t.Fatalf("chart not written: %v", err)
		}
	})
}

func TestDownloadChartWithoutServerFilenameFallsBackToAppAndVersion(t *testing.T) {
	srv := newFakeChartPackageServer(t, "", nil)
	mc := newTestMarketClient(t, srv.URL)

	dir := t.TempDir()
	result, err := fetchChartToDisk(context.Background(), mc, "amule", "upload", "0.0.7", dir, false)
	if err != nil {
		t.Fatalf("fetchChartToDisk: %v", err)
	}
	if want := filepath.Join(dir, "amule-0.0.7.tgz"); result.Path != want {
		t.Fatalf("path = %q, want %q", result.Path, want)
	}
}

// A filename is a suggestion from the server, and a suggestion carrying path
// separators must not be able to decide where the write lands.
func TestDownloadChartIgnoresFilenameWithPathSeparators(t *testing.T) {
	srv := newFakeChartPackageServer(t, "../../etc/amule.tgz", nil)
	mc := newTestMarketClient(t, srv.URL)

	dir := t.TempDir()
	result, err := fetchChartToDisk(context.Background(), mc, "amule", "upload", "0.0.7", dir, false)
	if err != nil {
		t.Fatalf("fetchChartToDisk: %v", err)
	}
	if want := filepath.Join(dir, "amule-0.0.7.tgz"); result.Path != want {
		t.Fatalf("path = %q, want the download to stay under %q", result.Path, dir)
	}
}

func TestDownloadChartRefusesToClobberWithoutOverwrite(t *testing.T) {
	srv := newFakeChartPackageServer(t, "amule-0.0.7.tgz", nil)
	mc := newTestMarketClient(t, srv.URL)

	dir := t.TempDir()
	dst := filepath.Join(dir, "amule-0.0.7.tgz")
	if err := os.WriteFile(dst, []byte("previous chart"), 0o644); err != nil {
		t.Fatalf("seed existing file: %v", err)
	}

	_, err := fetchChartToDisk(context.Background(), mc, "amule", "upload", "", dir, false)
	if err == nil {
		t.Fatal("expected the download to refuse an existing file")
	}
	if !strings.Contains(err.Error(), "--overwrite") {
		t.Fatalf("error = %v, want it to point at --overwrite", err)
	}
	got, _ := os.ReadFile(dst)
	if string(got) != "previous chart" {
		t.Fatalf("existing file was modified: %q", got)
	}

	result, err := fetchChartToDisk(context.Background(), mc, "amule", "upload", "", dir, true)
	if err != nil {
		t.Fatalf("fetchChartToDisk with overwrite: %v", err)
	}
	got, _ = os.ReadFile(result.Path)
	if string(got) != string(chartBytes) {
		t.Fatalf("overwrite left %q", got)
	}
}

func TestDownloadChartReportsBackendError(t *testing.T) {
	srv := newFakeChartErrorServer(t, http.StatusNotFound, "Chart not found")
	mc := newTestMarketClient(t, srv.URL)

	dir := t.TempDir()
	_, err := fetchChartToDisk(context.Background(), mc, "amule", "upload", "", dir, false)
	if err == nil {
		t.Fatal("expected an error for a chart the market does not hold")
	}
	if !strings.Contains(err.Error(), "404") || !strings.Contains(err.Error(), "Chart not found") {
		t.Fatalf("error = %v, want the status and the backend message", err)
	}
	entries, _ := os.ReadDir(dir)
	if len(entries) != 0 {
		t.Fatalf("a failed download wrote %d file(s)", len(entries))
	}
}

func TestDownloadChartRewritesAuthFailure(t *testing.T) {
	srv := newFakeChartErrorServer(t, http.StatusUnauthorized, "unauthorized")
	mc := newTestMarketClient(t, srv.URL)

	_, err := fetchChartToDisk(context.Background(), mc, "amule", "upload", "", t.TempDir(), false)
	if err == nil {
		t.Fatal("expected an error on 401")
	}
	if !strings.Contains(err.Error(), "profile login") {
		t.Fatalf("error = %v, want the profile login CTA", err)
	}
}

// A 200 carrying the JSON envelope means the request reached something other
// than the package route. Writing that body into a .tgz would only fail later,
// at helm install time.
func TestDownloadChartRejectsJSONSuccessBody(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		fmt.Fprint(w, `{"success":true,"message":"not implemented"}`)
	}))
	t.Cleanup(srv.Close)
	mc := newTestMarketClient(t, srv.URL)

	dir := t.TempDir()
	_, err := fetchChartToDisk(context.Background(), mc, "amule", "upload", "", dir, false)
	if err == nil {
		t.Fatal("expected an error for a JSON body on the package route")
	}
	if !strings.Contains(err.Error(), "not implemented") {
		t.Fatalf("error = %v, want the envelope message", err)
	}
	entries, _ := os.ReadDir(dir)
	if len(entries) != 0 {
		t.Fatalf("wrote %d file(s) for a JSON response", len(entries))
	}
}

func TestParseContentDispositionFilename(t *testing.T) {
	cases := []struct {
		in   string
		want string
	}{
		{`attachment; filename="amule-0.0.7.tgz"`, "amule-0.0.7.tgz"},
		{`attachment; filename=amule-0.0.7.tgz`, "amule-0.0.7.tgz"},
		{`attachment; filename=amule.tgz; foo=bar`, "amule.tgz"},
		{`attachment`, ""},
		{``, ""},
	}
	for _, tc := range cases {
		if got := parseContentDispositionFilename(tc.in); got != tc.want {
			t.Fatalf("parseContentDispositionFilename(%q) = %q, want %q", tc.in, got, tc.want)
		}
	}
}

func TestVersionFromChartFilename(t *testing.T) {
	cases := []struct {
		filename string
		app      string
		want     string
	}{
		{"amule-0.0.7.tgz", "amule", "0.0.7"},
		{"my-app-1.2.3.tgz", "my-app", "1.2.3"},
		{"other-1.0.0.tgz", "amule", ""},
		{"", "amule", ""},
	}
	for _, tc := range cases {
		if got := versionFromChartFilename(tc.filename, tc.app); got != tc.want {
			t.Fatalf("versionFromChartFilename(%q, %q) = %q, want %q", tc.filename, tc.app, got, tc.want)
		}
	}
}

func TestMarketDownloadIsRegistered(t *testing.T) {
	cmd := NewMarketCommand(nil)
	for _, sub := range cmd.Commands() {
		if sub.Name() == "download" {
			if sub.Flag("source") == nil || sub.Flag("version") == nil || sub.Flag("overwrite") == nil {
				t.Fatal("download is missing one of --source / --version / --overwrite")
			}
			return
		}
	}
	t.Fatal("market download is not registered on the market command")
}
