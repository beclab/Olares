package utils

import (
	"fmt"
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
