package wifi

import (
	"encoding/hex"
	"errors"
	"fmt"
	"net"
	"os"
	"path/filepath"
	"strings"

	"github.com/godbus/dbus/v5"
)

type Security string

const (
	Open        Security = "open"
	OWE         Security = "owe"
	WEP         Security = "wep"
	PSK         Security = "wpa-psk"
	SAE         Security = "sae"
	EAP         Security = "wpa-eap"
	Unsupported Security = "unsupported"
)

// ConnectRequest is the shared HTTP and BLE wire contract.
type ConnectRequest struct {
	SSID        string            `json:"ssid"`
	Password    string            `json:"password"`
	Username    string            `json:"username,omitempty"`
	Security    Security          `json:"security,omitempty"`
	BSSID       string            `json:"bssid,omitempty"`
	Interface   string            `json:"interface,omitempty"`
	Enterprise  *EnterpriseConfig `json:"enterprise,omitempty"`
	WEPKeyType  string            `json:"wepKeyType,omitempty"`
	WEPKeyIndex uint32            `json:"wepKeyIndex,omitempty"`
}
type EnterpriseConfig struct {
	EAP                string `json:"eap,omitempty"`
	Phase2Auth         string `json:"phase2Auth,omitempty"`
	AnonymousIdentity  string `json:"anonymousIdentity,omitempty"`
	CACertPath         string `json:"caCertPath,omitempty"`
	DomainSuffixMatch  string `json:"domainSuffixMatch,omitempty"`
	InsecureSkipVerify bool   `json:"insecureSkipVerify,omitempty"`
}

// SecurityTypes decodes NM80211ApFlags and NM80211ApSecurityFlags.
// Unknown key management is never inferred to be an open network.
func SecurityTypes(flags, wpa, rsn uint32) []Security {
	bits := wpa | rsn
	var result []Security
	for _, entry := range []struct {
		mask     uint32
		security Security
	}{
		{0x100, PSK}, {0x200, EAP}, {0x400, SAE}, {0x1800, OWE},
	} {
		if bits&entry.mask != 0 {
			result = append(result, entry.security)
		}
	}
	if bits & ^uint32(0x1fff) != 0 {
		result = append(result, Unsupported)
	}
	if len(result) != 0 {
		return result
	}
	if bits != 0 {
		return []Security{Unsupported}
	}
	if flags&1 != 0 {
		return []Security{WEP}
	}
	if flags & ^uint32(15) != 0 {
		return []Security{Unsupported}
	}
	return []Security{Open}
}

type settingsMap = map[string]map[string]dbus.Variant

func variants(values map[string]any) map[string]dbus.Variant {
	result := make(map[string]dbus.Variant, len(values))
	for k, v := range values {
		result[k] = dbus.MakeVariant(v)
	}
	return result
}

func (r ConnectRequest) validateBase() error {
	if len(r.SSID) == 0 || len(r.SSID) > 32 {
		return errors.New("ssid must contain 1 to 32 bytes")
	}
	if r.BSSID != "" {
		mac, err := net.ParseMAC(r.BSSID)
		if err != nil || len(mac) != 6 {
			return errors.New("invalid bssid")
		}
	}
	switch r.Security {
	case "", Open, OWE, WEP, PSK, SAE, EAP:
	default:
		return errors.New("unsupported wifi security type")
	}
	return nil
}

// connectionSettings produces a complete authentication section. Replacing the
// section, rather than merging it, prevents stale secrets and EAP settings.
func connectionSettings(r ConnectRequest) (settingsMap, error) {
	if err := r.validateBase(); err != nil {
		return nil, err
	}
	wireless := variants(map[string]any{"ssid": []byte(r.SSID), "mode": "infrastructure"})
	if r.BSSID != "" {
		mac, _ := net.ParseMAC(r.BSSID)
		wireless["bssid"] = dbus.MakeVariant([]byte(mac))
	}
	result := settingsMap{"802-11-wireless": wireless}
	sec := map[string]any{}
	if r.Security != EAP && (r.Username != "" || r.Enterprise != nil) {
		return nil, errors.New("enterprise credentials require wpa-eap security")
	}
	if r.Security != WEP && (r.WEPKeyType != "" || r.WEPKeyIndex != 0) {
		return nil, errors.New("wep parameters require wep security")
	}
	switch r.Security {
	case Open, OWE:
		if r.Password != "" {
			return nil, errors.New("this network does not accept a password")
		}
		if r.Security == Open {
			return result, nil
		}
		sec["key-mgmt"] = "owe"
		sec["pmf"] = uint32(3)
	case PSK:
		valid := len(r.Password) >= 8 && len(r.Password) <= 63
		if len(r.Password) == 64 {
			_, err := hex.DecodeString(r.Password)
			valid = err == nil
		}
		if !valid {
			return nil, errors.New("wpa password must be 8 to 63 bytes or 64 hexadecimal digits")
		}
		sec["key-mgmt"], sec["psk"] = "wpa-psk", r.Password
	case SAE:
		if len(r.Password) == 0 || len(r.Password) > 63 {
			return nil, errors.New("sae password must contain 1 to 63 bytes")
		}
		sec["key-mgmt"], sec["psk"], sec["pmf"] = "sae", r.Password, uint32(3)
	case WEP:
		if r.WEPKeyIndex > 3 {
			return nil, errors.New("wep key index must be between 0 and 3")
		}
		keyType := uint32(1)
		switch r.WEPKeyType {
		case "", "key":
			valid := len(r.Password) == 5 || len(r.Password) == 13
			if len(r.Password) == 10 || len(r.Password) == 26 {
				_, err := hex.DecodeString(r.Password)
				valid = err == nil
			}
			if !valid {
				return nil, errors.New("invalid wep key length or encoding")
			}
		case "passphrase":
			keyType = 2
			if len(r.Password) == 0 || len(r.Password) > 64 {
				return nil, errors.New("wep passphrase must contain 1 to 64 bytes")
			}
		default:
			return nil, errors.New("wep key type must be key or passphrase")
		}
		sec["key-mgmt"], sec["wep-key-type"], sec["wep-tx-keyidx"] = "none", keyType, r.WEPKeyIndex
		sec[fmt.Sprintf("wep-key%d", r.WEPKeyIndex)] = r.Password
	case EAP:
		if strings.TrimSpace(r.Username) == "" || r.Password == "" {
			return nil, errors.New("802.1x username and password are required")
		}
		e := EnterpriseConfig{}
		if r.Enterprise != nil {
			e = *r.Enterprise
		}
		if e.EAP == "" {
			e.EAP = "peap"
		}
		if e.Phase2Auth == "" {
			e.Phase2Auth = "mschapv2"
		}
		switch e.EAP {
		case "peap":
			if e.Phase2Auth != "mschapv2" {
				return nil, errors.New("peap supports mschapv2 only")
			}
		case "ttls":
			switch e.Phase2Auth {
			case "pap", "chap", "mschap", "mschapv2":
			default:
				return nil, errors.New("unsupported ttls inner authentication")
			}
		default:
			return nil, errors.New("unsupported eap method")
		}
		dot1x := map[string]any{"eap": []string{e.EAP}, "identity": r.Username, "password": r.Password, "phase2-auth": e.Phase2Auth}
		if e.AnonymousIdentity != "" {
			dot1x["anonymous-identity"] = e.AnonymousIdentity
		}
		if !e.InsecureSkipVerify {
			if strings.TrimSpace(e.DomainSuffixMatch) == "" {
				return nil, errors.New("802.1x domainSuffixMatch is required for certificate verification")
			}
			dot1x["domain-suffix-match"] = e.DomainSuffixMatch
			if e.CACertPath == "" {
				dot1x["system-ca-certs"] = true
			} else {
				if !filepath.IsAbs(e.CACertPath) || strings.ContainsRune(e.CACertPath, 0) {
					return nil, errors.New("caCertPath must be an absolute local path")
				}
				f, err := os.Open(e.CACertPath)
				if err != nil {
					return nil, errors.New("cannot read ca certificate")
				}
				info, err := f.Stat()
				f.Close()
				if err != nil || !info.Mode().IsRegular() {
					return nil, errors.New("ca certificate must be a regular file")
				}
				dot1x["ca-cert"] = append([]byte("file://"+e.CACertPath), 0)
				dot1x["system-ca-certs"] = false
			}
		} else {
			dot1x["system-ca-certs"] = false
		}
		result["802-1x"] = variants(dot1x)
		sec["key-mgmt"] = "wpa-eap"
	default:
		return nil, errors.New("wifi security could not be determined")
	}
	result["802-11-wireless-security"] = variants(sec)
	return result, nil
}
