package fan

import (
	"io"
	"net/http"
	"net/http/httptest"
	"os"
	"testing"

	"github.com/beclab/Olares/cli/pkg/credential"
	pkgdashboard "github.com/beclab/Olares/cli/pkg/dashboard"
	"github.com/beclab/Olares/cli/pkg/dashboard/format"
)

// newTestClient mints a *pkgdashboard.Client pointed at srv.
// Mirrors the parent dashboard_test.go fixture so this subpackage
// stays self-contained without reaching into the parent package's
// _test.go file (cross-package _test.go is not addressable).
func newTestClient(srv *httptest.Server) *pkgdashboard.Client {
	rp := &credential.ResolvedProfile{
		OlaresID:     "alice@olares.com",
		DashboardURL: srv.URL,
	}
	return pkgdashboard.NewClient(srv.Client(), rp)
}

// fixtureFlags returns a CommonFlags whose Validate() has already
// run. Tests override what they need (TempUnit, Output) on top.
func fixtureFlags(t *testing.T) *pkgdashboard.CommonFlags {
	t.Helper()
	cf := &pkgdashboard.CommonFlags{Timezone: format.LocalLocation()}
	if err := cf.Validate(); err != nil {
		t.Fatalf("CommonFlags.Validate: %v", err)
	}
	return cf
}

// fanStubMux dispatches the three endpoints fan business logic
// touches: device profile (Olares One vs. generic), live fan data,
// HAMI graphics list. Each handler is opt-in — leaving a slot nil
// returns 404 (the "no integration" branch that BuildLiveEnvelope
// is supposed to handle).
type fanStubMux struct {
	systemStatus http.HandlerFunc // /user-service/api/system/status
	systemFan    http.HandlerFunc // /user-service/api/mdns/olares-one/cpu-gpu
	graphics     http.HandlerFunc // /hami/api/vgpu/v1/gpus
}

func (m fanStubMux) server(t *testing.T) *httptest.Server {
	t.Helper()
	return httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch r.URL.Path {
		case "/user-service/api/system/status":
			if m.systemStatus != nil {
				m.systemStatus(w, r)
				return
			}
		case "/user-service/api/mdns/olares-one/cpu-gpu":
			if m.systemFan != nil {
				m.systemFan(w, r)
				return
			}
		case "/hami/api/vgpu/v1/gpus":
			if m.graphics != nil {
				m.graphics(w, r)
				return
			}
		}
		t.Errorf("unexpected upstream %s %s", r.Method, r.URL.Path)
		w.WriteHeader(http.StatusNotFound)
	}))
}

// olaresOneStatus is the canonical Olares One device profile reply.
func olaresOneStatus(w http.ResponseWriter, _ *http.Request) {
	_, _ = w.Write([]byte(`{"code":0,"data":{"device_name":"Olares One","host_name":"olares-one"}}`))
}

// genericStatus is the canonical non-Olares-One device profile
// reply (the gate trips → empty envelope w/ EmptyReason="not_olares_one").
func genericStatus(w http.ResponseWriter, _ *http.Request) {
	_, _ = w.Write([]byte(`{"code":0,"data":{"device_name":"Generic Box","host_name":"box"}}`))
}

// captureStdout redirects os.Stdout for the duration of fn so tests
// can assert on the JSON / table output that Run* helpers print.
// Restores the original stream on return; failures inside fn
// propagate via fn's own error return.
func captureStdout(t *testing.T, fn func() error) string {
	t.Helper()
	old := os.Stdout
	r, w, err := os.Pipe()
	if err != nil {
		t.Fatalf("os.Pipe: %v", err)
	}
	os.Stdout = w
	done := make(chan struct {
		out string
		err error
	}, 1)
	go func() {
		buf, rerr := io.ReadAll(r)
		done <- struct {
			out string
			err error
		}{string(buf), rerr}
	}()
	runErr := fn()
	_ = w.Close()
	os.Stdout = old
	res := <-done
	if res.err != nil {
		t.Fatalf("captureStdout read: %v", res.err)
	}
	if runErr != nil {
		t.Fatalf("captured fn err: %v", runErr)
	}
	return res.out
}
