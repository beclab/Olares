package connector

import (
	"fmt"
	"os"
	"os/exec"
	"strings"

	"github.com/Masterminds/semver/v3"
)

func hasAmdAPU(cmdExec func(s string) (string, error)) (bool, error) {
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

func hasAmdAPUOrGPU(cmdExec func(s string) (string, error)) (bool, error) {
	out, err := cmdExec("lspci -d '1002:' 2>/dev/null | grep 'AMD' || true")
	if err != nil {
		return false, err
	}
	if out != "" {
		return true, nil
	}
	out, err = cmdExec("lshw -c display -numeric -disable network 2>/dev/null | grep 'vendor: .* \\[1002\\]' || true")
	if err != nil {
		return false, err
	}
	if out != "" {
		return true, nil
	}
	return false, nil
}

func HasAmdAPU(execRuntime Runtime) (bool, error) {
	return hasAmdAPU(func(s string) (string, error) {
		return execRuntime.GetRunner().SudoCmd(s, false, false)
	})
}

func HasAmdAPULocal() (bool, error) {
	return hasAmdAPU(func(s string) (string, error) {
		out, err := exec.Command("sh", "-c", s).Output()
		if err != nil {
			return "", err
		}
		return string(out), nil
	})
}

func HasAmdAPUOrGPULocal() (bool, error) {
	return hasAmdAPUOrGPU(func(s string) (string, error) {
		out, err := exec.Command("sh", "-c", s).Output()
		if err != nil {
			return "", err
		}
		return string(out), nil
	})
}

func HasAmdAPUOrGPU(execRuntime Runtime) (bool, error) {
	return hasAmdAPUOrGPU(func(s string) (string, error) {
		return execRuntime.GetRunner().SudoCmd(s, false, false)
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
