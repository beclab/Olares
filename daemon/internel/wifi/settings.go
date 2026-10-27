package wifi

import (
	"errors"

	"github.com/godbus/dbus/v5"
)

// GetSettings decodes D-Bus structs as []any. Call.Store then reconstructs
// variants from Go types, turning these tuple arrays into aav (even when empty).
// Restore their concrete wire types before both UpdateUnsaved and rollback.
type ipv6Address struct {
	Address []byte
	Prefix  uint32
	Gateway []byte
}

type ipv6Route struct {
	Destination []byte
	Prefix      uint32
	NextHop     []byte
	Metric      uint32
}

func normalizeIPv6Settings(settings settingsMap) error {
	ipv6 := settings["ipv6"]
	if value, ok := ipv6["addresses"]; ok {
		var addresses []ipv6Address
		if err := value.Store(&addresses); err != nil {
			return errors.New("invalid saved ipv6 address settings")
		}
		ipv6["addresses"] = dbus.MakeVariant(addresses)
	}
	if value, ok := ipv6["routes"]; ok {
		var routes []ipv6Route
		if err := value.Store(&routes); err != nil {
			return errors.New("invalid saved ipv6 route settings")
		}
		ipv6["routes"] = dbus.MakeVariant(routes)
	}
	return nil
}
