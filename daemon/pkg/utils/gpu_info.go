package utils

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/jaypipes/ghw"
)

// collectGpuInfos returns "Vendor Product" for every graphics card that has
// complete device metadata. NVIDIA cards come first; order is otherwise
// preserved. Always returns a non-nil slice (empty when nothing qualifies).
func collectGpuInfos(cards []*ghw.GraphicsCard) []string {
	nvidia := make([]string, 0, len(cards))
	others := make([]string, 0, len(cards))
	for _, card := range cards {
		if card == nil || card.DeviceInfo == nil || card.DeviceInfo.Vendor == nil || card.DeviceInfo.Product == nil {
			continue
		}
		info := fmt.Sprintf("%s %s", card.DeviceInfo.Vendor.Name, card.DeviceInfo.Product.Name)
		if strings.Contains(strings.ToLower(info), "nvidia") {
			nvidia = append(nvidia, info)
		} else {
			others = append(others, info)
		}
	}
	return append(nvidia, others...)
}

// filterSriovVirtualFunctions removes cards that sysfs identifies as SR-IOV
// virtual functions. Cards are retained when their VF status cannot be
// confirmed.
func filterSriovVirtualFunctions(cards []*ghw.GraphicsCard, sysfsDevices string) []*ghw.GraphicsCard {
	filtered := make([]*ghw.GraphicsCard, 0, len(cards))
	for _, card := range cards {
		address := ""
		if card != nil {
			address = card.Address
			if address == "" && card.DeviceInfo != nil {
				address = card.DeviceInfo.Address
			}
		}
		if address != "" {
			physfn := filepath.Join(sysfsDevices, address, "physfn")
			if _, err := os.Lstat(physfn); err == nil {
				continue
			}
		}
		filtered = append(filtered, card)
	}
	return filtered
}
