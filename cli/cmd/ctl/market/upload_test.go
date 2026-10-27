package market

import (
	"errors"
	"fmt"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestUploadReportsArchitectureMismatch(t *testing.T) {
	srv := newUploadErrorServer(t, http.StatusUnprocessableEntity, `{
		"success": false,
		"code": "architecture_incompatible",
		"message": "manifest architectures [arm64] do not match cluster architectures [amd64]",
		"data": {
			"manifest_architectures": ["arm64"],
			"cluster_architectures": ["amd64"]
		}
	}`)
	err := doUploadFile(&MarketOptions{Quiet: true}, newTestMarketClient(t, srv.URL), testChart(t), chartUploadSource)

	var architectureErr *uploadArchitectureError
	if !errors.As(err, &architectureErr) {
		t.Fatalf("error = %v, want uploadArchitectureError", err)
	}
	if architectureErr.Code != "architecture_incompatible" {
		t.Fatalf("code = %q", architectureErr.Code)
	}
	message := err.Error()
	for _, want := range []string{"manifest supports [arm64]", "cluster provides [amd64]", "do not retry the unchanged package"} {
		if !strings.Contains(message, want) {
			t.Fatalf("error %q does not contain %q", message, want)
		}
	}
}

func TestUploadReportsClusterArchitectureUnavailable(t *testing.T) {
	srv := newUploadErrorServer(t, http.StatusServiceUnavailable, `{
		"success": false,
		"code": "cluster_arch_unavailable",
		"message": "Cluster node architectures are unavailable"
	}`)
	err := doUploadFile(&MarketOptions{Quiet: true}, newTestMarketClient(t, srv.URL), testChart(t), chartUploadSource)

	var architectureErr *uploadArchitectureError
	if !errors.As(err, &architectureErr) {
		t.Fatalf("error = %v, want uploadArchitectureError", err)
	}
	if architectureErr.Code != "cluster_arch_unavailable" {
		t.Fatalf("code = %q", architectureErr.Code)
	}
	message := err.Error()
	for _, want := range []string{"node discovery", "keep the current package and version"} {
		if !strings.Contains(message, want) {
			t.Fatalf("error %q does not contain %q", message, want)
		}
	}
}

func TestUploadUnknownCodeKeepsGenericAPIError(t *testing.T) {
	srv := newUploadErrorServer(t, http.StatusBadRequest, `{
		"success": false,
		"code": "other_error",
		"message": "other failure"
	}`)
	err := doUploadFile(&MarketOptions{Quiet: true}, newTestMarketClient(t, srv.URL), testChart(t), chartUploadSource)

	var architectureErr *uploadArchitectureError
	if errors.As(err, &architectureErr) {
		t.Fatalf("unexpected uploadArchitectureError: %v", err)
	}
	if !strings.Contains(err.Error(), "API error (HTTP 400): other failure") {
		t.Fatalf("error = %v", err)
	}
}

// A directory upload's exit code is the aggregate verdict, and the output mode
// must not change it. -o json used to return the encoder's own error, so a run
// where every chart was rejected still exited 0 and CI read it as a clean
// publish.
func TestUploadDirExitCodeIsIndependentOfOutputMode(t *testing.T) {
	srv := newUploadPerFileServer(t)

	modes := []struct {
		name string
		opts func() *MarketOptions
	}{
		{"table", func() *MarketOptions { return &MarketOptions{Output: "table"} }},
		{"json", func() *MarketOptions { return &MarketOptions{Output: "json"} }},
		{"quiet", func() *MarketOptions { return &MarketOptions{Output: "table", Quiet: true} }},
	}
	mixes := []struct {
		name       string
		charts     []string
		wantFailed bool
	}{
		{"partial failure", []string{"good-1.0.0.tgz", "bad-1.0.0.tgz"}, true},
		{"total failure", []string{"bad-1.0.0.tgz", "bad-2.0.0.tgz"}, true},
		{"no failure", []string{"good-1.0.0.tgz", "good-2.0.0.tgz"}, false},
	}

	for _, mode := range modes {
		for _, mix := range mixes {
			t.Run(mode.name+"/"+mix.name, func(t *testing.T) {
				dir := chartDir(t, mix.charts...)
				var err error
				discardStdout(t, func() {
					err = uploadDir(mode.opts(), newTestMarketClient(t, srv.URL), dir, chartUploadSource)
				})

				if mix.wantFailed && !errors.Is(err, errReported) {
					t.Fatalf("error = %v, want errReported so the process exits non-zero", err)
				}
				if !mix.wantFailed && err != nil {
					t.Fatalf("error = %v, want nil", err)
				}
			})
		}
	}
}

// newUploadPerFileServer accepts a chart whose filename starts with "good" and
// rejects one starting with "bad", so a single server serves both the partial
// and the total failure mixes.
func newUploadPerFileServer(t *testing.T) *httptest.Server {
	t.Helper()
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/app-store/api/v2/apps/upload" {
			http.NotFound(w, r)
			return
		}
		if err := r.ParseMultipartForm(1 << 20); err != nil {
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}
		name := ""
		if files := r.MultipartForm.File["chart"]; len(files) > 0 {
			name = files[0].Filename
		}
		w.Header().Set("Content-Type", "application/json")
		if strings.HasPrefix(name, "bad") {
			w.WriteHeader(http.StatusBadRequest)
			_, _ = fmt.Fprint(w, `{"success": false, "message": "Failed to process uploaded package"}`)
			return
		}
		_, _ = fmt.Fprint(w, `{"success": true, "message": "uploaded", "data": {}}`)
	}))
	t.Cleanup(srv.Close)
	return srv
}

func chartDir(t *testing.T, names ...string) string {
	t.Helper()
	dir := t.TempDir()
	for _, name := range names {
		if err := os.WriteFile(filepath.Join(dir, name), []byte("chart"), 0o600); err != nil {
			t.Fatal(err)
		}
	}
	return dir
}

// discardStdout keeps the JSON report out of `go test` output. printJSON writes
// to os.Stdout directly, so the only way to silence it is to replace the file.
func discardStdout(t *testing.T, run func()) {
	t.Helper()
	saved := os.Stdout
	devNull, err := os.OpenFile(os.DevNull, os.O_WRONLY, 0)
	if err != nil {
		t.Fatal(err)
	}
	os.Stdout = devNull
	defer func() {
		os.Stdout = saved
		devNull.Close()
	}()
	run()
}

func newUploadErrorServer(t *testing.T, status int, body string) *httptest.Server {
	t.Helper()
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/app-store/api/v2/apps/upload" {
			http.NotFound(w, r)
			return
		}
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(status)
		_, _ = fmt.Fprint(w, body)
	}))
	t.Cleanup(srv.Close)
	return srv
}

func testChart(t *testing.T) string {
	t.Helper()
	path := filepath.Join(t.TempDir(), "demo-1.0.0.tgz")
	if err := os.WriteFile(path, []byte("chart"), 0o600); err != nil {
		t.Fatal(err)
	}
	return path
}
