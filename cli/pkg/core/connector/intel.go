package connector

import (
	"os/exec"
	"strings"
)

// IntelGPU is a single Intel display-class PCI device detected on the host.
type IntelGPU struct {
	// ID is the PCI device id in uppercase hex without the 0x prefix (e.g. "7D55").
	ID string
	// Discrete reports whether the GPU is discrete (dGPU); false means it is an
	// integrated GPU (iGPU).
	Discrete bool
}

// intelGPUDetectScript walks the PCI devices in sysfs and, for every Intel
// (vendor 0x8086) display-class controller (VGA 0x0300, 3D 0x0302 or Display
// 0x0380), prints a line "<kind> <deviceid>" where kind is iGPU or dGPU and
// deviceid is the PCI device id without the 0x prefix (e.g. "iGPU 7d55").
//
// The integrated vs discrete decision mirrors the PCI topology: an integrated
// GPU hangs directly off the root complex (its parent in sysfs is the
// "pci0000:00" host bridge), whereas a discrete GPU sits behind a PCIe root
// port / bridge (any other parent).
const intelGPUDetectScript = `
for d in /sys/bus/pci/devices/*; do
  [ -e "$d/class" ] || continue
  cls=$(cat "$d/class" 2>/dev/null)
  case "$cls" in 0x0300*|0x0302*|0x0380*) ;; *) continue;; esac
  ven=$(cat "$d/vendor" 2>/dev/null)
  [ "$ven" = "0x8086" ] || continue
  dev=$(cat "$d/device" 2>/dev/null)
  dev=${dev#0x}
  parent=$(basename "$(dirname "$(readlink -f "$d")")")
  case "$parent" in
    pci0000:00) echo "iGPU $dev" ;;
    *)          echo "dGPU $dev" ;;
  esac
done
`

// detectIntelGPUs runs intelGPUDetectScript and returns every Intel display-class
// GPU on the host, classified as integrated or discrete.
func detectIntelGPUs(cmdExec func(s string) (string, error)) []IntelGPU {
	out, err := cmdExec(intelGPUDetectScript)
	if err != nil {
		return nil
	}
	var gpus []IntelGPU
	for _, line := range strings.Split(out, "\n") {
		fields := strings.Fields(line)
		if len(fields) != 2 {
			continue
		}
		kind, id := fields[0], strings.ToUpper(fields[1])
		switch kind {
		case "iGPU":
			gpus = append(gpus, IntelGPU{ID: id, Discrete: false})
		case "dGPU":
			gpus = append(gpus, IntelGPU{ID: id, Discrete: true})
		}
	}
	return gpus
}

// intelLocalExec runs a shell snippet against the local machine.
func intelLocalExec(s string) (string, error) {
	out, err := exec.Command("sh", "-c", s).Output()
	if err != nil {
		return "", err
	}
	return string(out), nil
}

func intelRuntimeExec(execRuntime Runtime) func(string) (string, error) {
	return func(s string) (string, error) {
		return execRuntime.GetRunner().SudoCmd(s, false, false)
	}
}

// IntelGPUsLocal returns the Intel GPUs detected on the local machine.
func IntelGPUsLocal() []IntelGPU {
	return detectIntelGPUs(intelLocalExec)
}

// IntelGPUs returns the Intel GPUs detected on the given runtime's host.
func IntelGPUs(execRuntime Runtime) []IntelGPU {
	return detectIntelGPUs(intelRuntimeExec(execRuntime))
}

func hasIntelIGPU(gpus []IntelGPU) bool {
	for _, g := range gpus {
		if !g.Discrete {
			return true
		}
	}
	return false
}

func hasIntelDGPU(gpus []IntelGPU) bool {
	for _, g := range gpus {
		if g.Discrete {
			return true
		}
	}
	return false
}

// HasIntelIGPULocal reports whether the local machine exposes an Intel
// integrated GPU (the "intel" unified-memory mode).
func HasIntelIGPULocal() bool {
	return hasIntelIGPU(IntelGPUsLocal())
}

// HasIntelIGPU reports whether the given runtime's host exposes an Intel
// integrated GPU (the "intel" unified-memory mode).
func HasIntelIGPU(execRuntime Runtime) bool {
	return hasIntelIGPU(IntelGPUs(execRuntime))
}

// HasIntelDGPULocal reports whether the local machine exposes an Intel discrete
// GPU (e.g. Arc/Data Center GPU on a PCIe slot).
func HasIntelDGPULocal() bool {
	return hasIntelDGPU(IntelGPUsLocal())
}

// HasIntelDGPU reports whether the given runtime's host exposes an Intel
// discrete GPU (e.g. Arc/Data Center GPU on a PCIe slot).
func HasIntelDGPU(execRuntime Runtime) bool {
	return hasIntelDGPU(IntelGPUs(execRuntime))
}
