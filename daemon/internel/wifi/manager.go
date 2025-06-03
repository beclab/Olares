package wifi

import (
	"github.com/godbus/dbus/v5"
	"github.com/muka/network_manager"
	log "github.com/sirupsen/logrus"
)

// NM 802-1x configuration settings
// https://developer.gnome.org/NetworkManager/1.0/ref-settings.html

const nmNs = network_manager.InterfaceNetworkManager

// NewManager initialiaze an instance of wifi manager
func NewManager() (*Manager, error) {
	wifiManager := new(Manager)

	log.Debug("Connecting to system DBus")
	conn, err := dbus.SystemBus()
	if err != nil {
		return nil, err
	}

	wifiManager.conn = conn

	wifiManager.networkManager = network_manager.NewNetworkManager(
		conn.Object(nmNs, dbus.ObjectPath("/org/freedesktop/NetworkManager")),
	)

	settings := network_manager.NewNetworkManager_Settings(
		conn.Object(nmNs, dbus.ObjectPath("/org/freedesktop/NetworkManager/Settings")),
	)

	wifiManager.settings = settings

	return wifiManager, nil
}

// Manager wrap WIFI management functionalities
type Manager struct {
	conn           *dbus.Conn
	networkManager *network_manager.NetworkManager
	settings       *network_manager.NetworkManager_Settings
}
