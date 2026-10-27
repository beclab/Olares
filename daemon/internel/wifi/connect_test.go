package wifi

import (
	"context"
	"errors"
	"reflect"
	"strings"
	"testing"
	"time"

	"github.com/godbus/dbus/v5"
)

type fakeNM struct {
	profile          settingsMap
	secrets          settingsMap
	updated          settingsMap
	restored         settingsMap
	calls            []string
	state            uint32
	fail             string
	deleted          bool
	cleanupCancelled bool
}

func (f *fakeNM) call(ctx context.Context, path dbus.ObjectPath, method string, args ...any) *dbus.Call {
	f.calls = append(f.calls, method)
	result := &dbus.Call{}
	if strings.HasSuffix(method, ".Delete") || strings.HasSuffix(method, ".Update") {
		f.cleanupCancelled = ctx.Err() != nil
	}
	if method == f.fail {
		result.Err = errors.New("remote error with password secret")
		return result
	}
	if ctx.Err() != nil {
		result.Err = ctx.Err()
		return result
	}
	switch method {
	case nmNs + ".Settings.ListConnections":
		paths := []dbus.ObjectPath{}
		if f.profile != nil {
			paths = append(paths, "/profile")
		}
		result.Body = []any{paths}
	case profileInterface + ".GetSettings":
		result.Body = []any{cloneSettings(f.profile)}
	case profileInterface + ".GetSecrets":
		result.Body = []any{f.secrets}
	case profileInterface + ".UpdateUnsaved":
		f.updated = cloneSettings(args[0].(settingsMap))
	case profileInterface + ".Update":
		f.restored = cloneSettings(args[0].(settingsMap))
	case nmNs + ".Settings.AddConnectionUnsaved":
		f.updated = cloneSettings(args[0].(settingsMap))
		result.Body = []any{dbus.ObjectPath("/new")}
	case nmNs + ".ActivateConnection":
		result.Body = []any{dbus.ObjectPath("/active")}
	case "org.freedesktop.DBus.Properties.Get":
		if args[1] == "Unsaved" {
			result.Body = []any{dbus.MakeVariant(false)}
		} else {
			result.Body = []any{dbus.MakeVariant(f.state)}
		}
	case profileInterface + ".Delete":
		f.deleted = true
	}
	return result
}
func savedPSK() settingsMap {
	return settingsMap{
		"connection":               variants(map[string]any{"id": "saved", "uuid": "uuid", "type": "802-11-wireless"}),
		"802-11-wireless":          variants(map[string]any{"ssid": []byte("home")}),
		"802-11-wireless-security": variants(map[string]any{"key-mgmt": "wpa-psk"}),
		"ipv4":                     variants(map[string]any{"method": "manual"}),
	}
}
func TestActivationRollback(t *testing.T) {
	r := ConnectRequest{SSID: "home", Security: PSK, Password: "newpassword"}
	sections, err := connectionSettings(r)
	if err != nil {
		t.Fatal(err)
	}
	target := candidate{AccessPoint: AccessPoint{Interface: "wlan0"}, path: "/ap", device: "/device"}
	for _, existing := range []bool{false, true} {
		for _, failure := range []string{nmNs + ".ActivateConnection", profileInterface + ".Save", "state"} {
			t.Run(strings.ReplaceAll(failure, ".", "_")+map[bool]string{true: "_existing", false: "_new"}[existing], func(t *testing.T) {
				f := &fakeNM{state: 2, fail: failure}
				if existing {
					f.profile = savedPSK()
					f.secrets = settingsMap{"802-11-wireless-security": variants(map[string]any{"psk": "oldpassword"})}
				}
				if failure == "state" {
					f.state = 4
				}
				err := activate(context.Background(), f, target, r, sections)
				if err == nil {
					t.Fatal("expected activation failure")
				}
				if strings.Contains(err.Error(), "secret") {
					t.Fatal("credential leaked")
				}
				if existing {
					if value(f.restored, "802-11-wireless-security", "psk") != "oldpassword" || value(f.restored, "ipv4", "method") != "manual" {
						t.Fatal("rollback lost old secret or IP settings")
					}
				} else if !f.deleted {
					t.Fatal("new profile not removed")
				}
			})
		}
	}
}
func TestActivationCommit(t *testing.T) {
	r := ConnectRequest{SSID: "home", Security: PSK, Password: "newpassword"}
	s, _ := connectionSettings(r)
	f := &fakeNM{profile: savedPSK(), state: 2, secrets: settingsMap{}}
	if err := activate(context.Background(), f, candidate{AccessPoint: AccessPoint{Interface: "wlan0"}, device: "/device", path: "/ap"}, r, s); err != nil {
		t.Fatal(err)
	}
	if f.restored != nil || f.deleted || f.calls[len(f.calls)-1] != profileInterface+".Save" {
		t.Fatal("successful profile was not committed")
	}
	if value(f.updated, "ipv4", "method") != "manual" || value(f.updated, "802-11-wireless-security", "psk") != "newpassword" {
		t.Fatal("profile data incorrect")
	}
}
func TestActivationTimeoutCleansWithFreshContext(t *testing.T) {
	r := ConnectRequest{SSID: "home", Security: Open}
	s, _ := connectionSettings(r)
	f := &fakeNM{state: 1}
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Millisecond)
	defer cancel()
	err := activate(ctx, f, candidate{}, r, s)
	if err == nil || !f.deleted || f.cleanupCancelled {
		t.Fatal("timeout cleanup did not run with an independent context")
	}
}
func TestProfileMatching(t *testing.T) {
	s := savedPSK()
	r := ConnectRequest{SSID: "home", Security: PSK}
	if !profileMatches(s, r, "wlan0") {
		t.Fatal("matching profile not reused")
	}
	s["802-11-wireless"]["bssid"] = dbus.MakeVariant([]byte{1})
	if profileMatches(s, r, "wlan0") {
		t.Fatal("invalid pinned BSSID accepted")
	}
	original := savedPSK()
	copy := cloneSettings(original)
	copy["connection"]["id"] = dbus.MakeVariant("changed")
	if reflect.DeepEqual(original, copy) {
		t.Fatal("clone shares maps")
	}
}
