package wifi

import (
	"context"
	"sort"
	"strings"
	"time"

	"github.com/godbus/dbus/v5"
	"github.com/muka/network_manager"
	log "github.com/sirupsen/logrus"
)

// Device wrap a NetworkManager_Device instance
type Device struct {
	Interface string
	Path      dbus.ObjectPath
	Device    *network_manager.NetworkManager_Device
}

// AccessPoint wrap AP information
type AccessPoint struct {
	SSID          string     `json:"ss"`
	Strength      byte       `json:"st"`
	SecurityTypes []Security `json:"se"`
	BSSID         string     `json:"bs"`
	Interface     string     `json:"if"`
	Connected     bool       `json:"-"`
}

// GetWifiDevices enumerate WIFI devices
func (m *Manager) GetWifiDevices() ([]Device, error) {
	return m.getWifiDevices(context.Background())
}

func (m *Manager) getWifiDevices(ctx context.Context) ([]Device, error) {

	list := []Device{}

	devices, err := m.networkManager.GetAllDevices(ctx)
	if err != nil {
		return list, err
	}

	for _, devicePath := range devices {
		device := network_manager.NewNetworkManager_Device(m.conn.Object(nmNs, devicePath))

		deviceType, err := device.GetDeviceType(ctx)
		if err != nil {
			log.Warnf("Error reading device type %s: %s", devicePath, err)
			continue
		}

		deviceInterface, err := device.GetInterface(ctx)
		if err != nil {
			log.Warnf("Error reading device interface %s: %s", devicePath, err)
			continue
		}

		if network_manager.NM_DEVICE_TYPE_WIFI == deviceType {
			list = append(list, Device{
				Path:      devicePath,
				Device:    device,
				Interface: deviceInterface,
			})
		}
	}

	return list, nil
}

// GetAccessPoints return a list of Access Points
func (m *Manager) GetAccessPoints(devicePath dbus.ObjectPath) ([]AccessPoint, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()
	return m.getAccessPoints(ctx, devicePath)
}

func (m *Manager) getAccessPoints(ctx context.Context, devicePath dbus.ObjectPath) ([]AccessPoint, error) {

	wireless := network_manager.NewNetworkManager_Device_Wireless(m.conn.Object(nmNs, devicePath))

	list := make(map[string]AccessPoint)

	enabled, err := m.networkManager.GetWirelessEnabled(ctx)
	if err != nil {
		return nil, err
	}
	if !enabled {
		if err := m.networkManager.SetWirelessEnabled(ctx, true); err != nil {
			return nil, err
		}
	}
	requestScan(ctx, systemNM{m.conn}, devicePath)
	accessPoints, err := wireless.GetAllAccessPoints(ctx)
	if err != nil {
		return nil, err
	}

	var iface string
	if err := property(ctx, systemNM{m.conn}, devicePath, deviceInterface, "Interface", &iface); err != nil {
		return nil, err
	}
	var active dbus.ObjectPath
	if err := property(ctx, systemNM{m.conn}, devicePath, wirelessInterface, "ActiveAccessPoint", &active); err != nil {
		return nil, err
	}
	for _, path := range accessPoints {
		ap, err := readAP(ctx, systemNM{m.conn}, path, iface)
		if err != nil {
			log.Error("Failed to read access point properties")
			continue
		}
		if ap.SSID == "" {
			continue
		}
		ap.Connected = path == active
		mergeAP(list, ap)
	}
	result := make([]AccessPoint, 0, len(list))
	for _, ap := range list {
		result = append(result, ap)
	}
	sort.Slice(result, func(i, j int) bool {
		if result[i].Strength != result[j].Strength {
			return result[i].Strength > result[j].Strength
		}
		return result[i].BSSID < result[j].BSSID
	})
	return result, nil
}

func securityKey(types []Security) string {
	values := make([]string, len(types))
	for i, s := range types {
		values[i] = string(s)
	}
	return strings.Join(values, ",")
}
func mergeAP(list map[string]AccessPoint, ap AccessPoint) {
	key := ap.Interface + "\x00" + ap.SSID + "\x00" + securityKey(ap.SecurityTypes)
	previous, ok := list[key]
	connected := previous.Connected || ap.Connected
	if !ok || ap.Strength > previous.Strength || (ap.Strength == previous.Strength && ap.BSSID < previous.BSSID) {
		ap.Connected = connected
		list[key] = ap
	} else {
		previous.Connected = connected
		list[key] = previous
	}
}
