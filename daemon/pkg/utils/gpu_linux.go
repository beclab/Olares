//go:build linux
// +build linux

package utils

import (
	"fmt"
	"io"
	"log"
	"strings"
	"sync"

	"github.com/jaypipes/ghw"
	"k8s.io/klog/v2"
	"k8s.io/utils/ptr"
)

var (
	gpuInfoMu     sync.Mutex
	gpuInfoCached bool
	gpuInfoValue  *string
)

// GetGpuInfo returns the primary GPU description. GPUs are not hot-plugged at
// runtime, but scanning the PCI bus via ghw on every status tick is expensive,
// so the result is cached after the first successful scan. A failed scan is not
// cached and will be retried on the next call.
func GetGpuInfo() (*string, error) {
	gpuInfoMu.Lock()
	defer gpuInfoMu.Unlock()
	if gpuInfoCached {
		return gpuInfoValue, nil
	}

	gpu, err := ghw.GPU(ghw.WithAlerter(log.New(io.Discard, "", 0))) // discard warnings
	if err != nil {
		klog.Errorf("Error getting GPU info: %v", err)
		return nil, err
	}

	var first string
	var result *string
	for _, card := range gpu.GraphicsCards {
		if card.DeviceInfo == nil || card.DeviceInfo.Vendor == nil || card.DeviceInfo.Product == nil {
			continue
		}
		info := fmt.Sprintf("%s %s", card.DeviceInfo.Vendor.Name, card.DeviceInfo.Product.Name)
		if strings.Contains(strings.ToLower(info), "nvidia") {
			first = info
			break
		}

		if first == "" {
			first = info
		}
	}

	if first != "" {
		result = ptr.To(first)
	}

	gpuInfoValue = result
	gpuInfoCached = true
	return result, nil
}
