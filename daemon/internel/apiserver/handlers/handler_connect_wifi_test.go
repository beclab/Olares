package handlers

import (
	"context"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/beclab/Olares/daemon/pkg/cluster/state"
	"github.com/beclab/Olares/daemon/pkg/commands"
	connectwifi "github.com/beclab/Olares/daemon/pkg/commands/connect_wifi"
	"github.com/gofiber/fiber/v2"
)

type wifiCommandRecorder struct {
	commands.Operation
	request *connectwifi.Param
}

func (c *wifiCommandRecorder) Execute(_ context.Context, p any) (any, error) {
	c.request = p.(*connectwifi.Param)
	return nil, nil
}
func TestConnectWifiForwardsEnterpriseAndLegacy(t *testing.T) {
	for _, body := range []string{
		`{"ssid":"office","username":"alice","password":"secret","security":"wpa-eap","enterprise":{"eap":"ttls","phase2Auth":"pap","domainSuffixMatch":"example.com"}}`,
		`{"ssid":"home","password":"12345678"}`,
	} {
		cmd := &wifiCommandRecorder{}
		h := &Handlers{}
		app := fiber.New()
		app.Post("/connect-wifi", func(c *fiber.Ctx) error { return h.PostConnectWifi(c, cmd) })
		req := httptest.NewRequest("POST", "/connect-wifi", strings.NewReader(body))
		req.Header.Set("Content-Type", "application/json")
		resp, err := app.Test(req)
		if err != nil {
			t.Fatal(err)
		}
		resp.Body.Close()
		if resp.StatusCode != 200 || cmd.request == nil {
			t.Fatalf("request not accepted: %d", resp.StatusCode)
		}
		if strings.Contains(body, "alice") && (cmd.request.Username != "alice" || cmd.request.Enterprise.Phase2Auth != "pap") {
			t.Fatal("enterprise parameters not forwarded")
		}
	}
}

// Exercise the registered route so a signature middleware regression cannot be
// hidden by calling PostConnectWifi directly. Malformed JSON avoids network work.
func TestConnectWifiRegisteredRouteDoesNotRequireSignature(t *testing.T) {
	previous := state.CurrentState
	t.Cleanup(func() { state.CurrentState = previous })
	state.CurrentState.TerminusdState = state.Running
	state.CurrentState.TerminusState = state.NotInstalled
	for _, headers := range []map[string]string{nil, {"X-Signature": "unused"}} {
		resp, body := callRegisteredMethod(t, http.MethodPost, "/command/connect-wifi", "{", headers)
		resp.Body.Close()
		if resp.StatusCode != http.StatusBadRequest || !strings.Contains(string(body), "unable to parse body") {
			t.Fatalf("unsigned Wi-Fi request did not reach body validation: status=%d body=%s", resp.StatusCode, body)
		}
	}
	// Removing the Wi-Fi gate must not affect other signed commands.
	resp, body := callRegisteredMethod(t, http.MethodPost, "/command/change-host", "{", nil)
	resp.Body.Close()
	if resp.StatusCode != http.StatusForbidden || !strings.Contains(string(body), "request is forbidden") {
		t.Fatalf("change-host signature gate changed: status=%d body=%s", resp.StatusCode, body)
	}
	state.CurrentState.TerminusdState = ""
	resp, body = callRegisteredMethod(t, http.MethodPost, "/command/connect-wifi", "{", nil)
	resp.Body.Close()
	if resp.StatusCode != http.StatusForbidden || !strings.Contains(string(body), "server is not running") {
		t.Fatalf("Wi-Fi server readiness gate changed: status=%d body=%s", resp.StatusCode, body)
	}
}
