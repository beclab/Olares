package utils

import (
	"bufio"
	"fmt"
	"strconv"
	"strings"

	"github.com/beclab/Olares/cli/pkg/core/connector"
	"github.com/beclab/Olares/cli/pkg/core/logger"
)

const (
	// kubeReservedMemoryMi is what kubelet holds back for the Kubernetes
	// daemons themselves (kubelet, the container runtime). It stays a flat
	// amount because those daemons' footprint does not scale with the size of
	// the machine.
	kubeReservedMemoryMi = 250

	// reservedMemoryPercent is the share of physical RAM held back for the OS:
	// the kernel, page cache, and on a GPU node the driver allocations that
	// back unified memory.
	reservedMemoryPercent = 3

	// maxReservedMemoryPercent caps the total reserve, so the swap term can add
	// at most two more points of physical RAM on top of the base share. Without
	// it a large swap device — a ZRAM one defaults to half of RAM — would eat an
	// arbitrary fraction of what the node can schedule.
	maxReservedMemoryPercent = 5

	// minTotalReservedMemoryMi floors the total reserve at the flat 500Mi that
	// preceded this calculation, so a small machine never ends up reserving
	// less than it used to.
	minTotalReservedMemoryMi = 500
)

// KubeReservedMemory is kubeReservedMemoryMi as a Kubernetes quantity.
var KubeReservedMemory = fmt.Sprintf("%dMi", kubeReservedMemoryMi)

// SystemReservedMemory returns the value for kubelet's system-reserved memory,
// as a Kubernetes quantity string.
//
// The total a node holds back is 3% of physical RAM plus, when pods may not use
// swap, the size of the swap device, capped at 5% of RAM and floored at the
// legacy 500Mi. KubeReservedMemory is part of that total, so this returns the
// remainder.
//
// The flat 250Mi + 250Mi it replaces was nowhere near enough on a large machine:
// a 96Gi host was left with 500Mi of headroom for the kernel, the container
// runtime and page cache combined, so it would happily commit itself to the point
// where reclaim could no longer keep up. That is unrecoverable in a way a per-pod
// OOM kill is not — the box stalls with every cgroup still inside its own limit,
// which is exactly how a GPU node dies once several unified-memory apps swap
// their VRAM into host RAM at once.
//
// Swap counts toward the reserve because with NoSwap pods it is not usable
// capacity, yet a node that commits everything still thrashes against it. When
// pods are allowed to swap it is real capacity and is left out.
//
// Determining the host's memory needs a command on that host, so a failure here
// is reported as a warning and the legacy flat value is returned rather than
// failing the install outright.
func SystemReservedMemory(runtime connector.Runtime, podSwapEnabled bool) string {
	out, err := runtime.GetRunner().Cmd("cat /proc/meminfo", false, false)
	if err == nil {
		var memTotal, swapTotal int64
		memTotal, swapTotal, err = parseMemInfo(out)
		if err == nil {
			return fmt.Sprintf("%dMi", systemReservedMemoryMi(memTotal, swapTotal, podSwapEnabled))
		}
	}
	logger.Warnf("failed to read host memory, reserving a flat %s for the system: %v", KubeReservedMemory, err)
	return KubeReservedMemory
}

// systemReservedMemoryMi is the arithmetic behind SystemReservedMemory, in MiB,
// taking MemTotal and SwapTotal in KiB as /proc/meminfo reports them.
func systemReservedMemoryMi(memTotalKiB, swapTotalKiB int64, podSwapEnabled bool) int64 {
	memMi := memTotalKiB / 1024
	totalMi := memMi * reservedMemoryPercent / 100
	if !podSwapEnabled {
		totalMi += swapTotalKiB / 1024
	}
	// The cap comes before the floor: on a machine small enough for the two to
	// disagree, never reserving less than the legacy flat amount wins.
	if capMi := memMi * maxReservedMemoryPercent / 100; totalMi > capMi {
		totalMi = capMi
	}
	if totalMi < minTotalReservedMemoryMi {
		totalMi = minTotalReservedMemoryMi
	}

	// kubeReservedMemoryMi is reserved separately and counts toward the same
	// total, so hand back only what is left for the OS.
	systemMi := totalMi - kubeReservedMemoryMi
	if systemMi < 1 {
		systemMi = 1
	}
	return systemMi
}

// parseMemInfo returns MemTotal and SwapTotal in KiB, the unit /proc/meminfo
// reports them in.
func parseMemInfo(out string) (memTotal, swapTotal int64, err error) {
	scanner := bufio.NewScanner(strings.NewReader(out))
	for scanner.Scan() {
		key, value, ok := strings.Cut(scanner.Text(), ":")
		if !ok {
			continue
		}
		switch key {
		case "MemTotal", "SwapTotal":
		default:
			continue
		}
		// The value looks like "  98231692 kB"; SwapTotal is 0 with no unit
		// when there is no swap device.
		fields := strings.Fields(value)
		if len(fields) == 0 {
			continue
		}
		parsed, parseErr := strconv.ParseInt(fields[0], 10, 64)
		if parseErr != nil {
			return 0, 0, fmt.Errorf("unexpected %s in /proc/meminfo: %q", key, strings.TrimSpace(value))
		}
		if key == "MemTotal" {
			memTotal = parsed
		} else {
			swapTotal = parsed
		}
	}
	if err := scanner.Err(); err != nil {
		return 0, 0, err
	}
	if memTotal <= 0 {
		return 0, 0, fmt.Errorf("no usable MemTotal in /proc/meminfo")
	}
	return memTotal, swapTotal, nil
}
