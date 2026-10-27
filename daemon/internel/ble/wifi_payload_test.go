package ble

import (
	"encoding/json"
	"fmt"
	"strings"
	"testing"

	"github.com/beclab/Olares/daemon/internel/wifi"
)

func TestWifiConnectionPayload(t *testing.T) {
	s := &service{}
	request, err := s.parseConnectConext([]byte(`{"ssid":"office","username":"alice","password":"secret","enterprise":{"insecureSkipVerify":true}}`))
	if err != nil || request.Username != "alice" || !request.Enterprise.InsecureSkipVerify {
		t.Fatal("enterprise payload lost")
	}
	if _, err = s.parseConnectConext([]byte(strings.Repeat(" ", 513))); err == nil || !strings.Contains(err.Error(), "512") {
		t.Fatal("oversized BLE request accepted")
	}
	request, err = s.parseConnectConext([]byte(`{"ssid":"home","password":"12345678"}`))
	if err != nil || request.SSID != "home" || request.Security != "" {
		t.Fatal("legacy payload changed")
	}
}
func TestWifiAPListKeepsHTTPComplete(t *testing.T) {
	list := []AccessPoint{}
	for i := 0; i < 20; i++ {
		list = append(list, AccessPoint{AccessPoint: wifi.AccessPoint{SSID: fmt.Sprintf("wifi-%d", i), Strength: 80, SecurityTypes: []wifi.Security{wifi.EAP}, BSSID: "01:02:03:04:05:06", Interface: "wlan0"}})
	}
	list[0].Connected = true
	list[0].AccessPoint.Connected = true
	var httpList []AccessPoint
	s := &service{updateApListCB: func(aps []AccessPoint) { httpList = aps }}
	s.publishAPList(list)
	if len(httpList) != 20 {
		t.Fatal("HTTP scan list truncated")
	}
	if len(s.apList) > 512 {
		t.Fatal("BLE payload exceeds attribute limit")
	}
	var bleList []AccessPoint
	if err := json.Unmarshal([]byte(s.apList), &bleList); err != nil {
		t.Fatal(err)
	}
	if len(bleList) == 0 || len(bleList) >= 20 || bleList[0].SecurityTypes[0] != wifi.EAP {
		t.Fatal("incorrect BLE payload")
	}
	var wireList []map[string]any
	if err := json.Unmarshal([]byte(s.apList), &wireList); err != nil {
		t.Fatal(err)
	}
	for _, ap := range wireList {
		for _, field := range []string{"connected", "co", "bs", "if"} {
			if _, ok := ap[field]; ok {
				t.Fatalf("BLE AP must not include %s", field)
			}
		}
	}
	if len(wireList) < 10 {
		t.Fatalf("compact payload contains only %d APs", len(wireList))
	}
	for _, ap := range wireList {
		if len(ap) != 3 || ap["st"] != float64(80) {
			t.Fatalf("unexpected BLE fields: %v", ap)
		}
	}
	if httpList[0].Strength != 80 || httpList[0].BSSID != "01:02:03:04:05:06" || httpList[0].Interface != "wlan0" {
		t.Fatal("full AP metadata was lost")
	}
	if !httpList[0].Connected {
		t.Fatal("internal connection status was lost")
	}
	s.publishAPList([]AccessPoint{})
	if s.apList != "[]" {
		t.Fatal("empty list encoding changed")
	}
}

func TestWifiAPListDeduplicatesAcrossDevices(t *testing.T) {
	list := []AccessPoint{
		{AccessPoint: wifi.AccessPoint{SSID: "office", Strength: 40, Interface: "wlan0", SecurityTypes: []wifi.Security{wifi.PSK, wifi.SAE}}},
		{AccessPoint: wifi.AccessPoint{SSID: "office", Strength: 90, Interface: "wlan1", SecurityTypes: []wifi.Security{wifi.SAE, wifi.PSK}}},
		{AccessPoint: wifi.AccessPoint{SSID: "office", Strength: 90, Interface: "wlan1", SecurityTypes: []wifi.Security{wifi.EAP}}},
	}
	s := &service{}
	s.publishAPList(list)
	var aps []advertisedAP
	if err := json.Unmarshal([]byte(s.apList), &aps); err != nil {
		t.Fatal(err)
	}
	if len(aps) != 2 || aps[0].Strength != 90 || aps[1].SecurityTypes[0] != wifi.EAP {
		t.Fatalf("incorrect network deduplication: %s", s.apList)
	}
	if list[1].SecurityTypes[0] != wifi.SAE {
		t.Fatal("scan metadata was mutated")
	}
}

func TestWifiAPListFitsEncodedBytesAndFillsRemainingSpace(t *testing.T) {
	list := []AccessPoint{}
	// Escaped SSIDs must be measured as JSON bytes, not characters.
	for i := 0; i < 3; i++ {
		list = append(list, AccessPoint{AccessPoint: wifi.AccessPoint{SSID: strings.Repeat("<", 25) + fmt.Sprint(i), SecurityTypes: []wifi.Security{wifi.EAP}}})
	}
	list = append(list, AccessPoint{AccessPoint: wifi.AccessPoint{SSID: "short", SecurityTypes: []wifi.Security{wifi.Open}}})
	s := &service{}
	s.publishAPList(list)
	var aps []advertisedAP
	if err := json.Unmarshal([]byte(s.apList), &aps); err != nil {
		t.Fatal(err)
	}
	if len(s.apList) > 512 || len(aps) != 3 || aps[2].SSID != "short" {
		t.Fatalf("remaining capacity was not used correctly: %s", s.apList)
	}
	s.publishAPList(nil)
	if s.apList != "[]" {
		t.Fatal("nil input must produce an empty array")
	}
}
