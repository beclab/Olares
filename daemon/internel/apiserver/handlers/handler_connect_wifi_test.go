package handlers

import (
	"context"
	"net/http/httptest"
	"strings"
	"testing"

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
