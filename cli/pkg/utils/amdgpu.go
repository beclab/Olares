package utils

import (
	"fmt"
	"os"
	"strings"

	"github.com/Masterminds/semver/v3"
	"github.com/beclab/Olares/cli/pkg/core/connector"
)

func HasAmdIGPU(execRuntime connector.Runtime) (bool, error) {
	// Detect by CPU model names that bundle AMD AI NPU/graphics
	targets := []string{
		"AMD Ryzen AI Max+ 395",
		"AMD Ryzen AI Max 390",
		"AMD Ryzen AI Max 385",
		"AMD Ryzen AI 9 HX 375",
		"AMD Ryzen AI 9 HX 370",
		"AMD Ryzen AI 9 365",
	}
	// try lscpu first: extract 'Model name' field
	out, err := execRuntime.GetRunner().SudoCmd("lscpu 2>/dev/null | awk -F': *' '/^Model name/{print $2; exit}' || true", false, false)
	if err != nil {
		return false, err
	}
	if out != "" {
		lo := strings.ToLower(strings.TrimSpace(out))
		for _, t := range targets {
			if strings.Contains(lo, strings.ToLower(t)) {
				return true, nil
			}
		}
	}
	// fallback to /proc/cpuinfo
	out, err = execRuntime.GetRunner().SudoCmd("awk -F': *' '/^model name/{print $2; exit}' /proc/cpuinfo 2>/dev/null || true", false, false)
	if err != nil {
		return false, err
	}
	if out != "" {
		lo := strings.ToLower(strings.TrimSpace(out))
		for _, t := range targets {
			if strings.Contains(lo, strings.ToLower(t)) {
				return true, nil
			}
		}
	}
	return false, nil
}

func RocmVersion() (*semver.Version, error) {
	const rocmVersionFile = "/opt/rocm/.info/version"
	data, err := os.ReadFile(rocmVersionFile)
	if err != nil {
		// no ROCm installed, nothing to check
		if os.IsNotExist(err) {
			return nil, err
		}
		return nil, err
	}
	curStr := strings.TrimSpace(string(data))
	cur, err := semver.NewVersion(curStr)
	if err != nil {
		return nil, fmt.Errorf("invalid rocm version: %s", curStr)
	}
	return cur, nil
}
