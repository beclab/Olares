package wifi

import (
	"context"
	"sort"
	"time"
)

// ListAccessPoints performs a bounded scan independent of the BLE service.
// Like the existing BLE scanner, it enables Wi-Fi before scanning a radio.
func ListAccessPoints(ctx context.Context) ([]AccessPoint, error) {
	ctx, cancel := context.WithTimeout(ctx, 30*time.Second)
	defer cancel()
	manager, err := NewManager()
	if err != nil {
		return nil, safeNMError("failed to connect to networkmanager", err)
	}
	defer manager.Close()
	devices, err := manager.getWifiDevices(ctx)
	if err != nil {
		return nil, safeNMError("failed to list wifi devices", err)
	}
	result := make([]AccessPoint, 0)
	for _, device := range devices {
		aps, err := manager.getAccessPoints(ctx, device.Path)
		if err != nil {
			return nil, safeNMError("failed to scan wifi access points", err)
		}
		result = append(result, aps...)
	}
	if err := ctx.Err(); err != nil {
		return nil, safeNMError("wifi scan timed out or was cancelled", err)
	}
	sort.Slice(result, func(i, j int) bool {
		if result[i].Connected != result[j].Connected {
			return result[i].Connected
		}
		if result[i].Strength != result[j].Strength {
			return result[i].Strength > result[j].Strength
		}
		return result[i].Interface+result[i].BSSID < result[j].Interface+result[j].BSSID
	})
	return result, nil
}
