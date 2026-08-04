//go:build linux
// +build linux

package utils

import (
	"io"
	"log"
	"sync"
	"time"

	"github.com/jaypipes/ghw"
	"k8s.io/klog/v2"
)

// gpuEmptyRescanInterval bounds how often an empty GPU scan is retried, so a
// node without a GPU does not re-scan the PCI bus on every status tick while a
// GPU that only appears after boot is still eventually detected.
const gpuEmptyRescanInterval = 5 * time.Minute

var (
	gpuInfoMu     sync.Mutex
	gpuInfoCached bool
	gpuInfoValue  []string
	gpuLastScan   time.Time
)

// GetGpuInfo returns descriptions for all detected GPUs. GPUs are not
// hot-plugged at runtime, but scanning the PCI bus via ghw on every status
// tick is expensive. A non-empty result is cached for the process lifetime; an
// empty scan is retried at most once per gpuEmptyRescanInterval, and a failed
// scan is retried on the next call. The returned slice is never nil.
func GetGpuInfo() ([]string, error) {
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

	result := collectGpuInfos(gpu.GraphicsCards)
	gpuInfoValue = result
	if len(result) > 0 {
		gpuInfoCached = true
	}
	return result, nil
}
