package market

import (
	"net/http"
	"reflect"
	"testing"

	"github.com/beclab/Olares/cli/cmd/ctl/market/cancel"
	"github.com/beclab/Olares/cli/cmd/ctl/market/resume"
	"github.com/beclab/Olares/cli/cmd/ctl/market/stop"
	"github.com/beclab/Olares/cli/cmd/ctl/market/uninstall"
)

// These tests pin the per-version wire format of the stop/resume/uninstall
// request builders so a change to either the 1.12.5 baseline or the 1.12.6
// (TermiPass PR #1162) body shape is caught without a live backend — the
// regression net for "1.12.5 and 1.12.6 both have to be verified".

func assertReq(t *testing.T, gotMethod, gotPath string, gotBody any, wantMethod, wantPath string, wantBody map[string]any) {
	t.Helper()
	if gotMethod != wantMethod {
		t.Errorf("method = %q, want %q", gotMethod, wantMethod)
	}
	if gotPath != wantPath {
		t.Errorf("path = %q, want %q", gotPath, wantPath)
	}
	if !reflect.DeepEqual(gotBody, any(wantBody)) {
		t.Errorf("body = %#v, want %#v", gotBody, wantBody)
	}
}

func TestStopWireFormat(t *testing.T) {
	m, p, b := stop.Build(false, "firefox", "market.olares", true)
	assertReq(t, m, p, b, http.MethodPost, "/apps/stop", map[string]any{
		"appName": "firefox", "all": true,
	})

	m, p, b = stop.Build(true, "firefox", "market.olares", true)
	assertReq(t, m, p, b, http.MethodPost, "/apps/stop", map[string]any{
		"app_name": "firefox", "source": "market.olares", "all": true,
	})
}

func TestResumeWireFormat(t *testing.T) {
	// 1.12.5: {appName}; source + computeBinding are ignored on this version.
	m, p, b := resume.Build(false, "firefox", "market.olares", nil)
	assertReq(t, m, p, b, http.MethodPost, "/apps/resume", map[string]any{
		"appName": "firefox",
	})

	// 1.12.6 without a binding: minimal {app_name, source} body.
	m, p, b = resume.Build(true, "firefox", "market.olares", nil)
	assertReq(t, m, p, b, http.MethodPost, "/apps/resume", map[string]any{
		"app_name": "firefox", "source": "market.olares",
	})

	// 1.12.6 with a binding: computeBinding is added.
	binding := []BindingSelection{{NodeName: "node-1", DeviceID: "gpu-0", Memory: 8 * giBytes}}
	m, p, b = resume.Build(true, "firefox", "market.olares", binding)
	assertReq(t, m, p, b, http.MethodPost, "/apps/resume", map[string]any{
		"app_name": "firefox", "source": "market.olares", "computeBinding": binding,
	})

	// 1.12.5 must NOT carry a binding even if one is passed (untouched path).
	m, p, b = resume.Build(false, "firefox", "market.olares", binding)
	assertReq(t, m, p, b, http.MethodPost, "/apps/resume", map[string]any{
		"appName": "firefox",
	})
}

func TestCancelWireFormat(t *testing.T) {
	m, p, b := cancel.Build(false, "firefox", "market.olares", "1.2.3")
	assertReq(t, m, p, b, http.MethodDelete, "/apps/firefox/install", map[string]any{
		"sync": true,
	})

	// 1.12.6 with no version supplied: body omits "version".
	m, p, b = cancel.Build(true, "firefox", "market.olares", "")
	assertReq(t, m, p, b, http.MethodDelete, "/apps/firefox/install", map[string]any{
		"app_name": "firefox", "source": "market.olares", "sync": true,
	})

	// 1.12.6 with a version: body includes "version".
	m, p, b = cancel.Build(true, "firefox", "market.olares", "1.2.3")
	assertReq(t, m, p, b, http.MethodDelete, "/apps/firefox/install", map[string]any{
		"app_name": "firefox", "source": "market.olares",
		"sync": true, "version": "1.2.3",
	})
}

func TestUninstallWireFormat(t *testing.T) {
	m, p, b := uninstall.Build(false, "firefox", "market.olares", "1.2.3", true, true)
	assertReq(t, m, p, b, http.MethodDelete, "/apps/firefox", map[string]any{
		"sync": true, "all": true, "deleteData": true,
	})

	// 1.12.6 with no version supplied: body omits "version".
	m, p, b = uninstall.Build(true, "firefox", "market.olares", "", false, false)
	assertReq(t, m, p, b, http.MethodDelete, "/apps/firefox", map[string]any{
		"app_name": "firefox", "source": "market.olares",
		"sync": true, "all": false, "deleteData": false,
	})

	// 1.12.6 with a version: body includes "version".
	m, p, b = uninstall.Build(true, "firefox", "market.olares", "1.2.3", true, true)
	assertReq(t, m, p, b, http.MethodDelete, "/apps/firefox", map[string]any{
		"app_name": "firefox", "source": "market.olares",
		"sync": true, "all": true, "deleteData": true, "version": "1.2.3",
	})
}
