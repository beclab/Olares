package connector

import (
	"fmt"
	"os"
	"os/exec"
	"strings"

	"github.com/Masterminds/semver/v3"
)

func hasAmdIGPU(cmdExec func(s string) (string, error)) (bool, error) {
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
	out, err := cmdExec("lscpu 2>/dev/null | awk -F': *' '/^Model name/{print $2; exit}' || true")
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
	out, err = cmdExec("awk -F': *' '/^model name/{print $2; exit}' /proc/cpuinfo 2>/dev/null || true")
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

func HasAmdIGPU(execRuntime Runtime) (bool, error) {
	return hasAmdIGPU(func(s string) (string, error) {
		return execRuntime.GetRunner().SudoCmd(s, false, false)
	})
}

func HasAmdIGPULocal() (bool, error) {
	return hasAmdIGPU(func(s string) (string, error) {
		out, err := exec.Command("sh", "-c", s).Output()
		if err != nil {
			return "", err
		}
		return string(out), nil
	})
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
