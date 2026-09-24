package wifi

import (
	"context"
	"errors"
	"fmt"
	"net"
	"runtime"
	"slices"
	"sort"
	"strings"
	"time"

	"github.com/godbus/dbus/v5"
	"github.com/google/uuid"
)

// Serialize profile mutations shared by HTTP and BLE. Waiting is cancellable.
var connectionGate = make(chan struct{}, 1)

func Connect(ctx context.Context, request ConnectRequest) error {
	if runtime.GOOS != "linux" {
		return errors.New("wifi connection is supported only on linux")
	}
	if err := request.validateBase(); err != nil {
		return err
	}
	ctx, cancel := context.WithTimeout(ctx, 90*time.Second)
	defer cancel()
	select {
	case connectionGate <- struct{}{}:
		defer func() { <-connectionGate }()
	case <-ctx.Done():
		return errors.New("wifi connection timed out or was cancelled")
	}
	conn, err := dbus.ConnectSystemBus()
	if err != nil {
		return errors.New("failed to connect to system bus")
	}
	defer conn.Close()
	b := systemNM{conn}
	candidates, err := scanCandidates(ctx, b, request)
	if err != nil {
		return safeNMError("failed to discover wifi networks", err)
	}
	target, security, err := selectTarget(request, candidates, func(iface string) (bool, error) { return supportsSAE(ctx, conn, iface) })
	if err != nil {
		return err
	}
	request.Security = security
	sections, err := connectionSettings(request)
	if err != nil {
		return err
	}
	return activate(ctx, b, target, request, sections)
}

type candidate struct {
	AccessPoint
	path, device dbus.ObjectPath
}

func scanCandidates(ctx context.Context, b nmBus, r ConnectRequest) ([]candidate, error) {
	var devices []dbus.ObjectPath
	if err := b.call(ctx, nmPath, nmNs+".GetDevices").Store(&devices); err != nil {
		return nil, err
	}
	var result []candidate
	for _, device := range devices {
		var kind uint32
		if err := property(ctx, b, device, deviceInterface, "DeviceType", &kind); err != nil {
			return nil, err
		}
		if kind != 2 {
			continue
		}
		var iface string
		if err := property(ctx, b, device, deviceInterface, "Interface", &iface); err != nil {
			return nil, err
		}
		if r.Interface != "" && r.Interface != iface {
			continue
		}
		// A scan may be rate limited; cached APs remain useful in that case.
		requestScan(ctx, b, device)
		var aps []dbus.ObjectPath
		if err := b.call(ctx, device, wirelessInterface+".GetAllAccessPoints").Store(&aps); err != nil {
			return nil, err
		}
		for _, path := range aps {
			ap, err := readAP(ctx, b, path, iface)
			if err != nil {
				continue
			} // APs can disappear while enumerating.
			if ap.SSID != r.SSID || (r.BSSID != "" && !strings.EqualFold(ap.BSSID, r.BSSID)) {
				continue
			}
			result = append(result, candidate{ap, path, device})
		}
	}
	return result, nil
}

func selectTarget(r ConnectRequest, candidates []candidate, sae func(string) (bool, error)) (candidate, Security, error) {
	var options []candidate
	groups := map[string]bool{}
	for _, c := range candidates {
		if r.Security != "" && !slices.Contains(c.SecurityTypes, r.Security) {
			continue
		}
		options = append(options, c)
		groups[securityKey(c.SecurityTypes)] = true
	}
	if len(options) == 0 {
		return candidate{}, "", errors.New("no matching wifi access point found")
	}
	if r.Security == "" && len(groups) != 1 {
		return candidate{}, "", errors.New("ambiguous wifi authentication; specify security or bssid")
	}
	sort.Slice(options, func(i, j int) bool {
		if options[i].Strength != options[j].Strength {
			return options[i].Strength > options[j].Strength
		}
		return options[i].Interface+options[i].BSSID < options[j].Interface+options[j].BSSID
	})
	selected := options[0]
	if r.Security != "" {
		return selected, r.Security, nil
	}
	types := selected.SecurityTypes
	if slices.Contains(types, Unsupported) {
		return candidate{}, "", errors.New("unsupported wifi authentication")
	}
	if len(types) == 1 {
		return selected, types[0], nil
	}
	if len(types) == 2 && slices.Contains(types, PSK) && slices.Contains(types, SAE) {
		supported, err := sae(selected.Interface)
		if err != nil {
			return candidate{}, "", errors.New("cannot determine sae capability; specify security explicitly")
		}
		if supported {
			return selected, SAE, nil
		}
		return selected, PSK, nil
	}
	return candidate{}, "", errors.New("ambiguous wifi authentication; specify security")
}

func value(s settingsMap, section, key string) any { return s[section][key].Value() }
func profileMatches(s settingsMap, r ConnectRequest, iface string) bool {
	ssid, ok := value(s, "802-11-wireless", "ssid").([]byte)
	if !ok || string(ssid) != r.SSID || value(s, "connection", "type") != "802-11-wireless" {
		return false
	}
	if bound, ok := value(s, "connection", "interface-name").(string); ok && bound != "" && bound != iface {
		return false
	}
	if mac, ok := value(s, "802-11-wireless", "bssid").([]byte); ok && len(mac) > 0 {
		// A pinned profile is only reusable for the same explicitly selected BSSID.
		if len(mac) != 6 || r.BSSID == "" || !strings.EqualFold(net.HardwareAddr(mac).String(), r.BSSID) {
			return false
		}
	}
	key, _ := value(s, "802-11-wireless-security", "key-mgmt").(string)
	switch r.Security {
	case Open:
		return key == ""
	case WEP:
		return key == "none"
	case EAP:
		return key == "wpa-eap" && value(s, "802-1x", "identity") == r.Username
	default:
		return key == string(r.Security)
	}
}

func existingProfile(ctx context.Context, b nmBus, r ConnectRequest, iface string) (dbus.ObjectPath, settingsMap, error) {
	var paths []dbus.ObjectPath
	if err := b.call(ctx, settingsPath, nmNs+".Settings.ListConnections").Store(&paths); err != nil {
		return "", nil, err
	}
	sort.Slice(paths, func(i, j int) bool { return paths[i] < paths[j] })
	for _, path := range paths {
		var original settingsMap
		if err := b.call(ctx, path, profileInterface+".GetSettings").Store(&original); err != nil {
			return "", nil, err
		}
		if !profileMatches(original, r, iface) {
			continue
		}
		// GetSettings intentionally omits secrets. Fetch them before modifying so
		// rollback cannot silently erase a saved password.
		for _, section := range []string{"802-11-wireless-security", "802-1x"} {
			if _, ok := original[section]; !ok {
				continue
			}
			var secrets settingsMap
			if err := b.call(ctx, path, profileInterface+".GetSecrets", section).Store(&secrets); err != nil {
				return "", nil, err
			}
			for group, values := range secrets {
				if original[group] == nil {
					original[group] = map[string]dbus.Variant{}
				}
				for key, v := range values {
					original[group][key] = v
				}
			}
		}
		return path, original, nil
	}
	return "", nil, nil
}

func cloneSettings(original settingsMap) settingsMap {
	result := settingsMap{}
	for group, values := range original {
		result[group] = map[string]dbus.Variant{}
		for key, v := range values {
			result[group][key] = v
		}
	}
	return result
}

func activate(ctx context.Context, b nmBus, target candidate, r ConnectRequest, sections settingsMap) (err error) {
	path, original, e := existingProfile(ctx, b, r, target.Interface)
	if e != nil {
		return safeNMError("failed to read saved wifi profiles", e)
	}
	created := path == ""
	var wasUnsaved bool
	if !created {
		if e := property(ctx, b, path, profileInterface, "Unsaved", &wasUnsaved); e != nil {
			return safeNMError("failed to read wifi profile persistence", e)
		}
	}
	updated := cloneSettings(original)
	if created {
		updated["connection"] = variants(map[string]any{"id": r.SSID, "uuid": uuid.NewString(), "type": "802-11-wireless", "interface-name": target.Interface, "autoconnect": true})
		updated["ipv4"] = variants(map[string]any{"method": "auto"})
		updated["ipv6"] = variants(map[string]any{"method": "auto"})
	}
	// Preserve IP/proxy settings, but replace all authentication and SSID data.
	delete(updated, "802-11-wireless-security")
	delete(updated, "802-1x")
	for group, values := range sections {
		if group == "802-11-wireless" && updated[group] != nil {
			for key, v := range values {
				updated[group][key] = v
			}
			delete(updated[group], "security")
		} else {
			updated[group] = values
		}
	}
	var active dbus.ObjectPath
	changed := false
	activationAttempted := false
	defer func() {
		if err == nil || !changed {
			return
		}
		cleanup, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()
		if activationAttempted && active == "" {
			var current, profile dbus.ObjectPath
			if property(cleanup, b, target.device, deviceInterface, "ActiveConnection", &current) == nil && current != "/" && current != "" {
				if property(cleanup, b, current, activeInterface, "Connection", &profile) == nil && profile == path {
					active = current
				}
			}
		}
		if active != "" && active != "/" {
			_ = b.call(cleanup, nmPath, nmNs+".DeactivateConnection", active).Err
		}
		var rollback error
		if created {
			rollback = b.call(cleanup, path, profileInterface+".Delete").Err
		} else {
			method := profileInterface + ".Update"
			if wasUnsaved {
				method = profileInterface + ".UpdateUnsaved"
			}
			rollback = b.call(cleanup, path, method, original).Err
		}
		if rollback != nil {
			err = errors.Join(err, safeNMError("failed to restore wifi profile", rollback))
		}
	}()
	if created {
		if e = b.call(ctx, settingsPath, nmNs+".Settings.AddConnectionUnsaved", updated).Store(&path); e != nil {
			createErr := safeNMError("failed to create wifi profile", e)
			cleanup, cancel := context.WithTimeout(context.Background(), 5*time.Second)
			defer cancel()
			var createdPath dbus.ObjectPath
			lookupErr := b.call(cleanup, settingsPath, nmNs+".Settings.GetConnectionByUuid", value(updated, "connection", "uuid")).Store(&createdPath)
			if lookupErr == nil {
				if deleteErr := b.call(cleanup, createdPath, profileInterface+".Delete").Err; deleteErr != nil {
					return errors.Join(createErr, safeNMError("failed to remove unconfirmed wifi profile", deleteErr))
				}
			}
			return createErr
		}
		changed = true
	} else {
		changed = true // Restore even if the update reply times out.
		if e = b.call(ctx, path, profileInterface+".UpdateUnsaved", updated).Err; e != nil {
			return safeNMError("failed to update wifi profile", e)
		}
	}
	activationAttempted = true
	if e = b.call(ctx, nmPath, nmNs+".ActivateConnection", path, target.device, target.path).Store(&active); e != nil {
		return safeNMError("failed to activate wifi connection", e)
	}
	if e = waitActivated(ctx, b, active); e != nil {
		return e
	}
	if e = b.call(ctx, path, profileInterface+".Save").Err; e != nil {
		return safeNMError("failed to save wifi profile", e)
	}
	return nil
}
func waitActivated(ctx context.Context, b nmBus, path dbus.ObjectPath) error {
	timer := time.NewTicker(250 * time.Millisecond)
	defer timer.Stop()
	for {
		var state uint32
		if err := property(ctx, b, path, activeInterface, "State", &state); err != nil {
			return safeNMError("failed to read wifi activation state", err)
		}
		switch state {
		case 2:
			return nil
		case 3, 4:
			return errors.New("wifi authentication or activation failed")
		}
		select {
		case <-ctx.Done():
			return errors.New("wifi connection timed out or was cancelled")
		case <-timer.C:
		}
	}
}

// Do not propagate arbitrary service messages: they can contain credentials or
// localized text. D-Bus error names remain useful for diagnostics.
func safeNMError(message string, err error) error {
	if errors.Is(err, context.Canceled) || errors.Is(err, context.DeadlineExceeded) {
		return fmt.Errorf("%s: operation timed out or was cancelled", message)
	}
	var remote dbus.Error
	if errors.As(err, &remote) {
		return fmt.Errorf("%s (%s)", message, remote.Name)
	}
	return errors.New(message)
}
