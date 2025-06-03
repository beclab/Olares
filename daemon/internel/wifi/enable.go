package wifi

import (
	"context"

	log "github.com/sirupsen/logrus"
)

// IsWifiEnabled return if wifi is enabled
func (m *Manager) IsWifiEnabled() (bool, error) {
	enabled, err := m.networkManager.GetWirelessEnabled(context.Background())
	if err != nil {
		return false, err
	}
	return enabled, nil
}

// EnableWifi unlock wifi via network manager
func (m *Manager) EnableWifi() error {

	enabled, err := m.IsWifiEnabled()
	if err != nil {
		return err
	}

	if !enabled {
		log.Debug("Enabling WIFI...")
		err := m.networkManager.SetWirelessEnabled(context.Background(), true)
		if err != nil {
			return err
		}
		log.Debug("WIFI enabled")
	}

	return nil
}
