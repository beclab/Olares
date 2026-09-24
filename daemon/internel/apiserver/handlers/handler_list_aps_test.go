package handlers

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/beclab/Olares/daemon/internel/apiserver/server"
	"github.com/beclab/Olares/daemon/internel/wifi"
)

func TestListAPsRoute(t *testing.T) {
	original := listWifiAccessPoints
	t.Cleanup(func() { listWifiAccessPoints = original })
	for _, scenario := range []string{"full", "empty", "failure"} {
		t.Run(scenario, func(t *testing.T) {
			calls := 0
			listWifiAccessPoints = func(context.Context) ([]wifi.AccessPoint, error) {
				calls++
				if scenario == "failure" {
					return nil, errors.New("sensitive remote detail")
				}
				if scenario == "empty" {
					return nil, nil
				}
				aps := make([]wifi.AccessPoint, 20)
				for i := range aps {
					aps[i] = wifi.AccessPoint{SSID: fmt.Sprintf("Office-%d", i), Strength: 85, SecurityTypes: []wifi.Security{wifi.EAP}, BSSID: "00:11:22:33:44:55", Interface: "wlan0", Connected: i == 0}
				}
				return aps, nil
			}
			response, err := server.API.App.Test(httptest.NewRequest("GET", "/system/list-aps", nil))
			if err != nil {
				t.Fatal(err)
			}
			defer response.Body.Close()
			body, err := io.ReadAll(response.Body)
			if err != nil {
				t.Fatal(err)
			}
			if calls != 1 {
				t.Fatal("route did not scan directly")
			}
			if scenario == "failure" {
				if response.StatusCode != 503 || strings.Contains(string(body), "sensitive remote detail") {
					t.Fatalf("unexpected error response: %s", body)
				}
				return
			}
			var result struct {
				Code int              `json:"code"`
				Data []map[string]any `json:"data"`
			}
			if err := json.Unmarshal(body, &result); err != nil {
				t.Fatal(err)
			}
			if response.StatusCode != 200 || result.Code != 200 || result.Data == nil {
				t.Fatalf("unexpected response: %s", body)
			}
			if scenario == "empty" {
				if len(result.Data) != 0 {
					t.Fatal("expected empty array")
				}
				return
			}
			if len(result.Data) != 20 || len(body) <= 512 {
				t.Fatal("HTTP AP list was truncated")
			}
			first := result.Data[0]
			if first["ss"] != "Office-0" || first["connected"] != true || first["bs"] != "00:11:22:33:44:55" || first["if"] != "wlan0" || first["st"] != float64(85) {
				t.Fatalf("incorrect AP fields: %v", first)
			}
			if result.Data[1]["connected"] != false {
				t.Fatal("disconnected AP must include connected: false")
			}
			for _, ap := range result.Data {
				if _, ok := ap["co"]; ok {
					t.Fatal("legacy co field must not be returned")
				}
			}
			security, ok := first["se"].([]any)
			if !ok || len(security) != 1 || security[0] != "wpa-eap" {
				t.Fatal("security metadata missing")
			}
		})
	}
}
