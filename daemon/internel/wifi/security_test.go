package wifi

import (
	"encoding/json"
	"os"
	"path/filepath"
	"reflect"
	"strings"
	"testing"
)

func TestSecurityTypes(t *testing.T) {
	for _, tt := range []struct {
		name            string
		flags, wpa, rsn uint32
		want            []Security
	}{
		{"open", 0, 0, 0, []Security{Open}}, {"wps", 14, 0, 0, []Security{Open}},
		{"wep", 1, 0, 0, []Security{WEP}}, {"wpa", 1, 0x100, 0, []Security{PSK}},
		{"wpa2", 1, 0, 0x108, []Security{PSK}}, {"transition", 1, 0, 0x500, []Security{PSK, SAE}},
		{"sae", 1, 0, 0x400, []Security{SAE}}, {"enterprise", 1, 0, 0x200, []Security{EAP}},
		{"owe", 0, 0, 0x800, []Security{OWE}}, {"owe transition", 0, 0, 0x1000, []Security{OWE}},
		{"suite b", 1, 0, 0x2000, []Security{Unsupported}}, {"unknown", 0, 0, 0x4000, []Security{Unsupported}},
		{"cipher only", 0, 0, 8, []Security{Unsupported}},
	} {
		t.Run(tt.name, func(t *testing.T) {
			if got := SecurityTypes(tt.flags, tt.wpa, tt.rsn); !reflect.DeepEqual(got, tt.want) {
				t.Fatalf("got %v want %v", got, tt.want)
			}
		})
	}
}
func enterpriseRequest() ConnectRequest {
	return ConnectRequest{SSID: "office", Security: EAP, Username: "alice", Password: "secret", Enterprise: &EnterpriseConfig{DomainSuffixMatch: "radius.example.com"}}
}
func TestEnterpriseSettings(t *testing.T) {
	r := enterpriseRequest()
	s, err := connectionSettings(r)
	if err != nil {
		t.Fatal(err)
	}
	if !reflect.DeepEqual(value(s, "802-1x", "eap"), []string{"peap"}) || value(s, "802-1x", "phase2-auth") != "mschapv2" || value(s, "802-1x", "identity") != "alice" || value(s, "802-1x", "password") != "secret" || value(s, "802-1x", "system-ca-certs") != true {
		t.Fatal("incorrect default enterprise configuration")
	}
	for _, inner := range []string{"pap", "chap", "mschap", "mschapv2"} {
		r.Enterprise.EAP = "ttls"
		r.Enterprise.Phase2Auth = inner
		s, err = connectionSettings(r)
		if err != nil {
			t.Fatal(err)
		}
		if value(s, "802-1x", "phase2-auth") != inner {
			t.Fatal("inner auth lost")
		}
	}
	r.Enterprise = &EnterpriseConfig{InsecureSkipVerify: true}
	s, err = connectionSettings(r)
	if err != nil {
		t.Fatal(err)
	}
	if _, ok := s["802-1x"]["ca-cert"]; ok {
		t.Fatal("unexpected CA")
	}
	if value(s, "802-1x", "system-ca-certs") != false {
		t.Fatal("skip verify ignored")
	}
	path := filepath.Join(t.TempDir(), "ca.pem")
	if err := os.WriteFile(path, []byte("test certificate"), 0600); err != nil {
		t.Fatal(err)
	}
	r.Enterprise = &EnterpriseConfig{CACertPath: path, DomainSuffixMatch: "radius.example.com"}
	s, err = connectionSettings(r)
	if err != nil {
		t.Fatal(err)
	}
	if !reflect.DeepEqual(value(s, "802-1x", "ca-cert"), append([]byte("file://"+path), 0)) {
		t.Fatal("incorrect CA path encoding")
	}
}
func TestRejectInvalidCredentials(t *testing.T) {
	for _, tt := range []struct {
		name   string
		change func(*ConnectRequest)
	}{
		{"username", func(r *ConnectRequest) { r.Username = "" }}, {"password", func(r *ConnectRequest) { r.Password = "" }},
		{"domain", func(r *ConnectRequest) { r.Enterprise = nil }}, {"tls", func(r *ConnectRequest) { r.Enterprise.EAP = "tls" }},
		{"inner", func(r *ConnectRequest) { r.Enterprise.Phase2Auth = "pap" }},
		{"ca", func(r *ConnectRequest) { r.Enterprise.CACertPath = "relative.pem" }},
		{"bssid", func(r *ConnectRequest) { r.BSSID = "bad" }},
	} {
		t.Run(tt.name, func(t *testing.T) {
			r := enterpriseRequest()
			tt.change(&r)
			if _, err := connectionSettings(r); err == nil {
				t.Fatal("expected validation error")
			}
		})
	}
}
func TestPersonalSettings(t *testing.T) {
	for _, r := range []ConnectRequest{
		{SSID: "a", Security: Open}, {SSID: "a", Security: OWE},
		{SSID: "a", Security: PSK, Password: "12345678"}, {SSID: "a", Security: SAE, Password: "secret"},
		{SSID: "a", Security: WEP, Password: "abcde"}, {SSID: "a", Security: WEP, Password: "passphrase", WEPKeyType: "passphrase", WEPKeyIndex: 3},
	} {
		if _, err := connectionSettings(r); err != nil {
			t.Fatal(err)
		}
	}
	for _, r := range []ConnectRequest{
		{SSID: "a", Security: Open, Password: "secret"}, {SSID: "a", Security: PSK, Password: "short"},
		{SSID: "a", Security: WEP, Password: "abcde", WEPKeyIndex: 4}, {SSID: "a", Security: Unsupported},
	} {
		if _, err := connectionSettings(r); err == nil {
			t.Fatal("invalid request accepted")
		}
	}
}
func TestLegacyAndEnterpriseJSON(t *testing.T) {
	var r ConnectRequest
	if err := json.Unmarshal([]byte(`{"ssid":"home","password":"12345678"}`), &r); err != nil {
		t.Fatal(err)
	}
	if r.Security != "" || r.SSID != "home" {
		t.Fatal("legacy contract changed")
	}
	if err := json.Unmarshal([]byte(`{"ssid":"office","username":"alice","password":"secret","enterprise":{"insecureSkipVerify":true}}`), &r); err != nil {
		t.Fatal(err)
	}
	r.Security = EAP
	if _, err := connectionSettings(r); err != nil {
		t.Fatal(err)
	}
}
func TestAPGroupingAndSelection(t *testing.T) {
	aps := map[string]AccessPoint{}
	for _, ap := range []AccessPoint{
		{SSID: "same", Interface: "wlan0", BSSID: "a", Strength: 20, SecurityTypes: []Security{Open}},
		{SSID: "same", Interface: "wlan0", BSSID: "b", Strength: 80, SecurityTypes: []Security{Open}},
		{SSID: "same", Interface: "wlan0", BSSID: "c", Strength: 90, SecurityTypes: []Security{PSK}},
	} {
		mergeAP(aps, ap)
	}
	if len(aps) != 2 {
		t.Fatal("different authentication was merged")
	}
	for _, ap := range aps {
		if ap.BSSID == "a" {
			t.Fatal("weak AP retained")
		}
	}
	list := []candidate{}
	for _, ap := range aps {
		list = append(list, candidate{AccessPoint: ap})
	}
	if _, _, err := selectTarget(ConnectRequest{}, list, nil); err == nil {
		t.Fatal("ambiguous network accepted")
	}
	_, sec, err := selectTarget(ConnectRequest{Security: PSK}, list, nil)
	if err != nil || sec != PSK {
		t.Fatal("explicit security ignored")
	}
	mixed := []candidate{{AccessPoint: AccessPoint{Interface: "wlan0", SecurityTypes: []Security{PSK, SAE}}}}
	for _, supported := range []bool{true, false} {
		_, sec, err = selectTarget(ConnectRequest{}, mixed, func(string) (bool, error) { return supported, nil })
		if err != nil || (supported && sec != SAE) || (!supported && sec != PSK) {
			t.Fatal("incorrect mixed mode choice")
		}
	}
}
func TestErrorsNeverEchoCredentials(t *testing.T) {
	for _, message := range []string{"secret", "密码 secret"} {
		err := safeNMError("failed to activate wifi connection", &testError{message})
		if strings.Contains(err.Error(), message) {
			t.Fatal("remote message leaked")
		}
	}
}

type testError struct{ message string }

func (e *testError) Error() string { return e.message }
