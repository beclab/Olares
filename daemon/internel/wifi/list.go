package wifi

import (
	"context"
	"time"

	"github.com/godbus/dbus/v5"
	"github.com/muka/network_manager"
	log "github.com/sirupsen/logrus"
	"golang.org/x/exp/maps"
)

// Device wrap a NetworkManager_Device instance
type Device struct {
	Interface string
	Path      dbus.ObjectPath
	Device    *network_manager.NetworkManager_Device
}

// AccessPoint wrap AP information
type AccessPoint struct {
	SSID     string `json:"ss"`
	Strength byte   `json:"st"`
}

// GetWifiDevices enumerate WIFI devices
func (m *Manager) GetWifiDevices() ([]Device, error) {

	list := []Device{}

	devices, err := m.networkManager.GetAllDevices(context.Background())
	if err != nil {
		return list, err
	}

	for _, devicePath := range devices {
		device := network_manager.NewNetworkManager_Device(m.conn.Object(nmNs, devicePath))

		deviceType, err := device.GetDeviceType(context.Background())
		if err != nil {
			log.Warnf("Error reading device type %s: %s", devicePath, err)
			continue
		}

		deviceInterface, err := device.GetInterface(context.Background())
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

	wireless := network_manager.NewNetworkManager_Device_Wireless(m.conn.Object(nmNs, devicePath))

	list := make(map[string]AccessPoint)

	err := m.EnableWifi()
	if err != nil {
		return nil, err
	}

	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()
	wireless.RequestScan(ctx, nil)
	accessPoints, err := wireless.GetAllAccessPoints(ctx)
	if err != nil {
		return nil, err
	}

	for _, accessPointPath := range accessPoints {

		accessPoint := network_manager.NewNetworkManager_AccessPoint(m.conn.Object(nmNs, accessPointPath))

		ssid, err := accessPoint.GetSsid(context.Background())
		if err != nil {
			log.Errorf("Error on GetSsid: %s", err)
			continue
		}

		strength, err := accessPoint.GetStrength(context.Background())
		if err != nil {
			log.Errorf("Error on GetStrength: %s", err)
			continue
		}

		addAp := AccessPoint{
			SSID:     string(ssid),
			Strength: strength,
		}

		list[addAp.SSID] = addAp
	}

	return maps.Values(list), err
}
