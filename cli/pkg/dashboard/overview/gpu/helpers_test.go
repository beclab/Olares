package gpu

import (
	"io"
	"net/http"
	"net/http/httptest"
	"os"
	"strings"
	"testing"

	"github.com/beclab/Olares/cli/pkg/credential"
	pkgdashboard "github.com/beclab/Olares/cli/pkg/dashboard"
	"github.com/beclab/Olares/cli/pkg/dashboard/format"
)

// newTestClient mints a *pkgdashboard.Client pointed at srv.
// Mirrors the parent dashboard_test.go fixture so this subpackage
// stays self-contained.
func newTestClient(srv *httptest.Server) *pkgdashboard.Client {
	rp := &credential.ResolvedProfile{
		OlaresID:     "alice@olares.com",
		DashboardURL: srv.URL,
	}
	return pkgdashboard.NewClient(srv.Client(), rp)
}

// fixtureFlags returns a CommonFlags whose Validate() has run.
func fixtureFlags(t *testing.T) *pkgdashboard.CommonFlags {
	t.Helper()
	cf := &pkgdashboard.CommonFlags{Timezone: format.LocalLocation()}
	if err := cf.Validate(); err != nil {
		t.Fatalf("CommonFlags.Validate: %v", err)
	}
	return cf
}

// adminEnsureUser is the canonical EnsureUser body used by every
// non-empty test: alice is a platform-admin, so GPUAdvisory's
// non-admin branch never trips. The CUDA-supported node check is
// covered by adminEnsureCudaNode.
const adminEnsureUser = `{"clusterRole":"workspaces-manager","user":{"username":"alice","globalrole":"platform-admin","email":"alice@olares.com"}}`

// adminEnsureCudaNode advertises a CUDA-supported node so
// GPUAdvisory's no-CUDA branch never trips either.
const adminEnsureCudaNode = `{"items":[{"metadata":{"name":"node-1","labels":{"gpu.bytetrade.io/cuda-supported":"true"}}}]}`

// gpuStubMux dispatches the endpoints gpu business logic touches.
// Each handler is opt-in. Setting any of them turns "advisory
// admin & CUDA" on by default (so we don't repeat the boilerplate
// in every test); subtests can override by setting the matching
// field explicitly.
type gpuStubMux struct {
	ensureUser    http.HandlerFunc // /capi/app/detail
	ensureNodes   http.HandlerFunc // /kapis/resources.kubesphere.io/v1alpha2/...
	graphicsList  http.HandlerFunc // /hami/api/vgpu/v1/gpus
	graphicsGet   http.HandlerFunc // /hami/api/vgpu/v1/gpu
	taskList      http.HandlerFunc // /hami/api/vgpu/v1/containers
	taskGet       http.HandlerFunc // /hami/api/vgpu/v1/container
	instantVector http.HandlerFunc // /hami/api/vgpu/v1/instant-vector
	rangeVector   http.HandlerFunc // /hami/api/vgpu/v1/range-vector
}

func (m gpuStubMux) server(t *testing.T) *httptest.Server {
	t.Helper()
	if m.ensureUser == nil {
		m.ensureUser = func(w http.ResponseWriter, _ *http.Request) {
			_, _ = w.Write([]byte(adminEnsureUser))
		}
	}
	if m.ensureNodes == nil {
		m.ensureNodes = func(w http.ResponseWriter, _ *http.Request) {
			_, _ = w.Write([]byte(adminEnsureCudaNode))
		}
	}
	return httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch {
		case strings.HasSuffix(r.URL.Path, "/capi/app/detail"):
			m.ensureUser(w, r)
		case strings.Contains(r.URL.Path, "/kapis/resources.kubesphere.io"):
			m.ensureNodes(w, r)
		case strings.HasSuffix(r.URL.Path, "/v1/gpus"):
			if m.graphicsList != nil {
				m.graphicsList(w, r)
				return
			}
			t.Errorf("unstubbed graphicsList %s %s", r.Method, r.URL.Path)
			w.WriteHeader(http.StatusNotFound)
		case strings.HasSuffix(r.URL.Path, "/v1/gpu"):
			if m.graphicsGet != nil {
				m.graphicsGet(w, r)
				return
			}
			t.Errorf("unstubbed graphicsGet %s %s", r.Method, r.URL.Path)
			w.WriteHeader(http.StatusNotFound)
		case strings.HasSuffix(r.URL.Path, "/v1/containers"):
			if m.taskList != nil {
				m.taskList(w, r)
				return
			}
			t.Errorf("unstubbed taskList %s %s", r.Method, r.URL.Path)
			w.WriteHeader(http.StatusNotFound)
		case strings.HasSuffix(r.URL.Path, "/v1/container"):
			if m.taskGet != nil {
				m.taskGet(w, r)
				return
			}
			t.Errorf("unstubbed taskGet %s %s", r.Method, r.URL.Path)
			w.WriteHeader(http.StatusNotFound)
		case strings.HasSuffix(r.URL.Path, "/instant-vector"):
			if m.instantVector != nil {
				m.instantVector(w, r)
				return
			}
			_, _ = w.Write([]byte(`{"data":[]}`))
		case strings.HasSuffix(r.URL.Path, "/range-vector"):
			if m.rangeVector != nil {
				m.rangeVector(w, r)
				return
			}
			_, _ = w.Write([]byte(`{"data":[]}`))
		default:
			t.Errorf("unexpected upstream %s %s", r.Method, r.URL.Path)
			w.WriteHeader(http.StatusNotFound)
		}
	}))
}

// captureStdout redirects os.Stdout for the duration of fn so
// tests can assert on the JSON / table output that Run* helpers
// print.
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
