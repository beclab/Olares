package market

import (
	"net/http"
	"reflect"
	"testing"

	resumev5 "github.com/beclab/Olares/cli/cmd/ctl/market/resume/v1_12_5"
	resumev6 "github.com/beclab/Olares/cli/cmd/ctl/market/resume/v1_12_6"
	stopv5 "github.com/beclab/Olares/cli/cmd/ctl/market/stop/v1_12_5"
	stopv6 "github.com/beclab/Olares/cli/cmd/ctl/market/stop/v1_12_6"
	uninstallv5 "github.com/beclab/Olares/cli/cmd/ctl/market/uninstall/v1_12_5"
	uninstallv6 "github.com/beclab/Olares/cli/cmd/ctl/market/uninstall/v1_12_6"
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
	m, p, b := stopv5.Stop("firefox", "market.olares", true)
	assertReq(t, m, p, b, http.MethodPost, "/apps/stop", map[string]any{
		"appName": "firefox", "all": true,
	})

	m, p, b = stopv6.Stop("firefox", "market.olares", true)
	assertReq(t, m, p, b, http.MethodPost, "/apps/stop", map[string]any{
		"app_name": "firefox", "source": "market.olares", "all": true,
	})
}

func TestResumeWireFormat(t *testing.T) {
	m, p, b := resumev5.Resume("firefox", "market.olares")
	assertReq(t, m, p, b, http.MethodPost, "/apps/resume", map[string]any{
		"appName": "firefox",
	})

	m, p, b = resumev6.Resume("firefox", "market.olares")
	assertReq(t, m, p, b, http.MethodPost, "/apps/resume", map[string]any{
		"app_name": "firefox", "source": "market.olares",
	})
}

func TestUninstallWireFormat(t *testing.T) {
	m, p, b := uninstallv5.Uninstall("firefox", "market.olares", "1.2.3", true, true)
	assertReq(t, m, p, b, http.MethodDelete, "/apps/firefox", map[string]any{
		"sync": true, "all": true, "deleteData": true,
	})

	// 1.12.6 with no version supplied: body omits "version".
	m, p, b = uninstallv6.Uninstall("firefox", "market.olares", "", false, false)
	assertReq(t, m, p, b, http.MethodDelete, "/apps/firefox", map[string]any{
		"app_name": "firefox", "source": "market.olares",
		"sync": true, "all": false, "deleteData": false,
	})

	// 1.12.6 with a version: body includes "version".
	m, p, b = uninstallv6.Uninstall("firefox", "market.olares", "1.2.3", true, true)
	assertReq(t, m, p, b, http.MethodDelete, "/apps/firefox", map[string]any{
		"app_name": "firefox", "source": "market.olares",
		"sync": true, "all": true, "deleteData": true, "version": "1.2.3",
	})
}
