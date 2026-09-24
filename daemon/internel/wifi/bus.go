package wifi

import (
	"context"
	"errors"
	"strings"
	"time"

	"github.com/godbus/dbus/v5"
)

const nmPath dbus.ObjectPath = "/org/freedesktop/NetworkManager"
const settingsPath dbus.ObjectPath = "/org/freedesktop/NetworkManager/Settings"
const deviceInterface = nmNs + ".Device"
const wirelessInterface = deviceInterface + ".Wireless"
const apInterface = nmNs + ".AccessPoint"
const profileInterface = nmNs + ".Settings.Connection"
const activeInterface = nmNs + ".Connection.Active"

// nmBus is kept small so activation and rollback can be tested without a radio.
type nmBus interface {
	call(context.Context, dbus.ObjectPath, string, ...any) *dbus.Call
}
type systemNM struct{ conn *dbus.Conn }

func (b systemNM) call(ctx context.Context, path dbus.ObjectPath, method string, args ...any) *dbus.Call {
	return b.conn.Object(nmNs, path).CallWithContext(ctx, method, 0, args...)
}
func property(ctx context.Context, b nmBus, path dbus.ObjectPath, iface, name string, out any) error {
	var value dbus.Variant
	if err := b.call(ctx, path, "org.freedesktop.DBus.Properties.Get", iface, name).Store(&value); err != nil {
		return err
	}
	return value.Store(out)
}
func properties(ctx context.Context, b nmBus, path dbus.ObjectPath, iface string) (map[string]dbus.Variant, error) {
	var values map[string]dbus.Variant
	err := b.call(ctx, path, "org.freedesktop.DBus.Properties.GetAll", iface).Store(&values)
	return values, err
}
func readAP(ctx context.Context, b nmBus, path dbus.ObjectPath, iface string) (AccessPoint, error) {
	values, err := properties(ctx, b, path, apInterface)
	if err != nil {
		return AccessPoint{}, err
	}
	var ssid []byte
	var flags, wpa, rsn uint32
	var ap AccessPoint
	for name, out := range map[string]any{"Ssid": &ssid, "Flags": &flags, "WpaFlags": &wpa, "RsnFlags": &rsn, "Strength": &ap.Strength, "HwAddress": &ap.BSSID} {
		v, ok := values[name]
		if !ok {
			return AccessPoint{}, errors.New("access point properties are incomplete")
		}
		if err := v.Store(out); err != nil {
			return AccessPoint{}, err
		}
	}
	ap.SSID, ap.Interface, ap.SecurityTypes = string(ssid), iface, SecurityTypes(flags, wpa, rsn)
	return ap, nil
}

func supportsSAE(ctx context.Context, conn *dbus.Conn, iface string) (bool, error) {
	const service = "fi.w1.wpa_supplicant1"
	var path dbus.ObjectPath
	if err := conn.Object(service, "/fi/w1/wpa_supplicant1").CallWithContext(ctx, service+".GetInterface", 0, iface).Store(&path); err != nil {
		return false, err
	}
	var v dbus.Variant
	if err := conn.Object(service, path).CallWithContext(ctx, "org.freedesktop.DBus.Properties.Get", 0, service+".Interface", "Capabilities").Store(&v); err != nil {
		return false, err
	}
	var caps map[string]dbus.Variant
	if err := v.Store(&caps); err != nil {
		return false, err
	}
	var methods []string
	keyMgmt, ok := caps["KeyMgmt"]
	if !ok || keyMgmt.Value() == nil {
		return false, errors.New("key management capabilities are unavailable")
	}
	if err := keyMgmt.Store(&methods); err != nil {
		return false, err
	}
	for _, method := range methods {
		if strings.EqualFold(method, "sae") {
			return true, nil
		}
	}
	return false, nil
}

// RequestScan is asynchronous. Wait for LastScan to advance before consuming
// the AP cache; when NM rate-limits a request, use its current cache.
func requestScan(ctx context.Context, b nmBus, device dbus.ObjectPath) {
	c, cancel := context.WithTimeout(ctx, 10*time.Second)
	defer cancel()
	var previous int64
	readErr := property(c, b, device, wirelessInterface, "LastScan", &previous)
	if b.call(c, device, wirelessInterface+".RequestScan", map[string]dbus.Variant{}).Err != nil || readErr != nil {
		return
	}
	timer := time.NewTicker(250 * time.Millisecond)
	defer timer.Stop()
	for {
		var latest int64
		if property(c, b, device, wirelessInterface, "LastScan", &latest) != nil || latest > previous {
			return
		}
		select {
		case <-c.Done():
			return
		case <-timer.C:
		}
	}
}
