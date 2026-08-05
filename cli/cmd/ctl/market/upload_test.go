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
