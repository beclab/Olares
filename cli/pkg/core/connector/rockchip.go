package connector

import "strings"

func isRockChip(execRuntime Runtime) bool {
	// check the /proc/device-tree/model exists and contains "rockchip"
	out, err := execRuntime.GetRunner().SudoCmd("cat /proc/device-tree/model 2>/dev/null || true", false, false)
	if err != nil {
		return false
	}
	if out != "" && strings.Contains(strings.ToLower(out), "rockchip") {
		return true
	}
	return false
}

func isRockChipRK3588(execRuntime Runtime) bool {
	if !isRockChip(execRuntime) {
		return false
	}
	out, err := execRuntime.GetRunner().SudoCmd("cat /proc/device-tree/model 2>/dev/null || true", false, false)
	if err != nil {
		return false
	}
	if out != "" && strings.Contains(strings.ToLower(out), "rk3588") {
		return true
	}
	return false
}

func getRockChipModel(execRuntime Runtime) string {
	out, err := execRuntime.GetRunner().SudoCmd("cat /proc/device-tree/model 2>/dev/null || true", false, false)
	if err != nil {
		return ""
	}
	return strings.TrimSpace(out)
}
