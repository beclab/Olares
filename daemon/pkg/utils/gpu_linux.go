//go:build linux
// +build linux

package utils

import (
	"fmt"
	"strings"

	"github.com/jaypipes/ghw"
	"k8s.io/klog/v2"
	"k8s.io/utils/ptr"
)

func GetGpuInfo() (*string, error) {
	gpu, err := ghw.GPU()
	if err != nil {
		klog.Errorf("Error getting GPU info: %v", err)
		return nil, err
	}

	var first string
	for _, card := range gpu.GraphicsCards {
		if card.DeviceInfo == nil || card.DeviceInfo.Vendor == nil || card.DeviceInfo.Product == nil {
			continue
		}
		info := fmt.Sprintf("%s %s", card.DeviceInfo.Vendor.Name, card.DeviceInfo.Product.Name)
		if strings.Contains(strings.ToLower(info), "nvidia") {
			return ptr.To(info), nil
		}

		if first == "" {
			first = info
		}
	}

	if first == "" {
		return nil, nil
	}

	return ptr.To(first), nil
}
