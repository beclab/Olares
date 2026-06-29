//go:build linux
// +build linux

package utils

import (
	"fmt"
	"io"
	"log"
	"strings"
	"sync"
	"time"

	"github.com/jaypipes/ghw"
	"k8s.io/klog/v2"
	"k8s.io/utils/ptr"
)

// gpuEmptyRescanInterval bounds how often an empty GPU scan is retried, so a
// node without a GPU does not re-scan the PCI bus on every status tick while a
// GPU that only appears after boot is still eventually detected.
const gpuEmptyRescanInterval = 5 * time.Minute

var (
	gpuInfoMu     sync.Mutex
	gpuInfoCached bool
	gpuInfoValue  *string
	gpuLastScan   time.Time
)

// GetGpuInfo returns the primary GPU description. GPUs are not hot-plugged at
// runtime, but scanning the PCI bus via ghw on every status tick is expensive.
// A found GPU is cached for the process lifetime; an empty scan is retried at
// most once per gpuEmptyRescanInterval, and a failed scan is retried on the
// next call.
func GetGpuInfo() (*string, error) {
	gpuInfoMu.Lock()
	defer gpuInfoMu.Unlock()
	if gpuInfoCached {
		return gpuInfoValue, nil
	}
	if !gpuLastScan.IsZero() && time.Since(gpuLastScan) < gpuEmptyRescanInterval {
		return gpuInfoValue, nil
	}

	gpu, err := ghw.GPU(ghw.WithAlerter(log.New(io.Discard, "", 0))) // discard warnings
	if err != nil {
		klog.Errorf("Error getting GPU info: %v", err)
		return nil, err
	}
	gpuLastScan = time.Now()

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
	if result != nil {
		gpuInfoCached = true
	}
	return result, nil
}
