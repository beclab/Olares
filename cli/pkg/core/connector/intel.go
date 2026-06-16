package connector

import (
	"os/exec"
	"strings"
)

// hasIntelGPU reports whether the machine exposes an Intel GPU (PCI vendor
// 8086) on a display-class controller.
func hasIntelGPU(cmdExec func(s string) (string, error)) bool {
	// lspci -nn prints the vendor:device id in brackets, e.g.
	//   00:02.0 VGA compatible controller [0300]: Intel Corporation ... [8086:7d55]
	out, err := cmdExec("lspci -nn 2>/dev/null | grep -iE 'VGA compatible controller|3D controller|Display controller' | grep -i '\\[8086:' || true")
	if err != nil {
		return false
	}
	return strings.TrimSpace(out) != ""
}

// HasIntelGPULocal runs the Intel GPU detection against the local machine.
func HasIntelGPULocal() bool {
	return hasIntelGPU(func(s string) (string, error) {
		out, err := exec.Command("sh", "-c", s).Output()
		if err != nil {
			return "", err
		}
		return string(out), nil
	})
}

// HasIntelGPU runs the Intel GPU detection against the given runtime's host.
func HasIntelGPU(execRuntime Runtime) bool {
	return hasIntelGPU(func(s string) (string, error) {
		return execRuntime.GetRunner().SudoCmd(s, false, false)
	})
}
