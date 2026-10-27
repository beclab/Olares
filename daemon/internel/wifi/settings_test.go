package wifi

import (
	"context"
	"reflect"
	"testing"

	"github.com/godbus/dbus/v5"
)

func wireIPv6Settings(populated bool) map[string]dbus.Variant {
	addresses, routes := [][]any{}, [][]any{}
	if populated {
		address := []byte{0x20, 0x01, 0x0d, 0xb8, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 1}
		gateway := []byte{0xfe, 0x80, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 1}
		addresses = append(addresses, []any{address, uint32(64), gateway})
		routes = append(routes, []any{address, uint32(128), gateway, uint32(42)})
	}
	return map[string]dbus.Variant{
		"method":    dbus.MakeVariant("auto"),
		"addresses": dbus.MakeVariantWithSignature(addresses, dbus.ParseSignatureMust("a(ayuay)")),
		"routes":    dbus.MakeVariantWithSignature(routes, dbus.ParseSignatureMust("a(ayuayu)")),
	}
}

func TestSavedIPv6WireTypesOnUpdateAndRollback(t *testing.T) {
	for _, populated := range []bool{false, true} {
		for _, failure := range []string{"", nmNs + ".ActivateConnection"} {
			profile := savedPSK()
			profile["ipv6"] = wireIPv6Settings(populated)
			f := &fakeNM{profile: profile, state: 2, fail: failure, secrets: settingsMap{}}
			r := ConnectRequest{SSID: "home", Security: PSK, Password: "newpassword"}
			sections, _ := connectionSettings(r)
			err := activate(context.Background(), f, candidate{AccessPoint: AccessPoint{Interface: "wlan0"}}, r, sections)
			if (err != nil) != (failure != "") {
				t.Fatalf("unexpected activation result: %v", err)
			}
			check := func(settings settingsMap) {
				for key, signature := range map[string]string{"addresses": "a(ayuay)", "routes": "a(ayuayu)"} {
					v := settings["ipv6"][key]
					if v.Signature().String() != signature {
						t.Fatalf("%s signature = %s, want %s", key, v.Signature(), signature)
					}
				}
				var addresses []ipv6Address
				var routes []ipv6Route
				if settings["ipv6"]["addresses"].Store(&addresses) != nil || settings["ipv6"]["routes"].Store(&routes) != nil {
					t.Fatal("invalid typed payload")
				}
				if populated {
					if len(addresses) != 1 || addresses[0].Prefix != 64 || len(routes) != 1 || routes[0].Metric != 42 {
						t.Fatal("IPv6 configuration changed")
					}
				} else if len(addresses) != 0 || len(routes) != 0 {
					t.Fatal("empty arrays changed")
				}
			}
			check(f.updated)
			if failure != "" {
				check(f.restored)
			}
			if !reflect.DeepEqual(profile["ipv6"], wireIPv6Settings(populated)) {
				t.Fatal("source settings changed")
			}
		}
	}
}

func TestNormalizeIPv6SettingsRejectsMalformedTuples(t *testing.T) {
	for _, key := range []string{"addresses", "routes"} {
		s := settingsMap{"ipv6": {key: dbus.MakeVariant([][]any{{"secret"}})}}
		if normalizeIPv6Settings(s) == nil {
			t.Fatal("invalid tuple accepted")
		}
	}
	if err := normalizeIPv6Settings(settingsMap{}); err != nil {
		t.Fatal(err)
	}
}
