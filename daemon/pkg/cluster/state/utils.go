package state

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"github.com/Masterminds/semver/v3"
	"os"
	"os/exec"
	"strconv"
	"strings"

	"github.com/beclab/Olares/daemon/pkg/cli"
	"github.com/beclab/Olares/daemon/pkg/commands"
	"github.com/beclab/Olares/daemon/pkg/utils"
	"k8s.io/klog/v2"
)

var ErrInstallFailed error = errors.New("install failed")
var ErrProcessFailed error = errors.New("process failed")
var ErrChangeIpFailed error = errors.New("change ip failed")

func IsK3SRunning(ctx context.Context) (bool, error) {
	p, err := utils.FindProcByName(ctx, "k3s-server")
	if err != nil {
		if os.IsNotExist(err) {
			return false, nil
		}

		return false, err
	}

	return len(p) > 0, nil

}

func IsTerminusInstalled() (bool, error) {
	info, err := os.Stat(commands.INSTALL_LOCK)
	if err != nil {
		if os.IsNotExist(err) {
			return false, nil
		}
		klog.Error(err)
		return false, err
	}

	if !info.IsDir() {
		return true, nil
	}

	return false, nil
}

func IsSystemShuttingdown() (bool, error) {
	_, isShutdown, err := utils.GetSystemPendingShutdowm()
	if err != nil {
		return false, err
	}

	return isShutdown, nil
}

func IsSystemRebooting() (bool, error) {
	mode, isShutdown, err := utils.GetSystemPendingShutdowm()
	if err != nil {
		return false, err
	}

	if !isShutdown {
		return isShutdown, nil
	}

	return mode == "reboot", nil
}

func isProcessRunning(pidfile string) (bool, error) {
	_, err := os.Stat(pidfile)
	if err != nil {
		if os.IsNotExist(err) {
			return false, nil
		}

		return false, err
	}

	pidData, err := os.ReadFile(pidfile)
	if err != nil {
		return false, err
	}

	if len(strings.TrimSpace(string(pidData))) == 0 {
		return false, nil
	}

	pid, err := strconv.Atoi(string(pidData))
	if err != nil {
		return false, err
	}

	if pid != 0 {
		p, err := utils.ProcessExists(pid)
		if err != nil {
			klog.Error("find process error, ", err)
			return false, err
		}

		if !p {
			return false, ErrProcessFailed
		}

		return true, nil
	}

	return false, nil

}

func IsTerminusInstalling() (bool, error) {
	running, err := isProcessRunning(commands.INSTALLING_PID_FILE)
	if err != nil {
		if err == ErrProcessFailed {
			err = ErrInstallFailed
		}
	}

	return running, err
}

func IsIpChangeRunning() (bool, error) {
	running, err := isProcessRunning(commands.CHANGINGIP_PID_FILE)
	if err != nil {
		if err == ErrProcessFailed {
			err = ErrChangeIpFailed
		}
	}

	return running, err
}

func GetMachineInfo(ctx context.Context) (osType, osInfo, osArch, osVersion, osKernel string, err error) {
	cmd := exec.CommandContext(ctx, cli.TERMINUS_CLI, "osinfo", "show")

	if output, err := cmd.Output(); err == nil {
		lines := strings.Split(string(output), "\n")
		for _, line := range lines {
			kv := strings.Split(line, "=")
			if len(kv) < 2 {
				continue
			}
			switch strings.TrimSpace(kv[0]) {
			case "OS_TYPE":
				osType = kv[1]
			case "OS_INFO":
				osInfo = kv[1]
			case "OS_ARCH":
				osArch = kv[1]
			case "OS_VERSION":
				osVersion = kv[1]
			case "OS_KERNEL":
				osKernel = kv[1]
			}
		}
	}

	return
}

type UpgradeTarget struct {
	Version      semver.Version `json:"version"`
	WizardURL    string         `json:"wizardURL"`
	CliURL       string         `json:"cliURL"`
	DownloadOnly bool           `json:"downloadOnly"`
	Downloaded   bool           `json:"downloaded"`
}

func (t *UpgradeTarget) IsValidRequest() error {
	existingTarget, err := GetOlaresUpgradeTarget()
	if err == nil && existingTarget != nil && !t.Version.Equal(&existingTarget.Version) {
		return fmt.Errorf("different upgrade version: %s already exists, please cancel it first", existingTarget.Version)
	}
	if CurrentState.TerminusVersion != nil {
		current, err := semver.NewVersion(*CurrentState.TerminusVersion)
		if err != nil {
			return fmt.Errorf("invalid current version %s: %v", CurrentState.TerminusVersion, err)
		}
		if !current.LessThan(&t.Version) {
			return fmt.Errorf("target version should be greater than current version: %s", current)
		}
	}
	return nil
}

func (t *UpgradeTarget) Save() error {
	content, err := json.Marshal(t)
	if err != nil {
		return fmt.Errorf("failed to marshal target: %v", err)
	}
	err = os.WriteFile(commands.UPGRADE_TARGET_FILE, content, 0644)
	if err != nil {
		return fmt.Errorf("failed to write target file: %v", err)
	}
	return nil
}

func GetOlaresUpgradeTarget() (*UpgradeTarget, error) {
	b, err := os.ReadFile(commands.UPGRADE_TARGET_FILE)
	if err != nil {
		if os.IsNotExist(err) {
			return nil, nil
		}
		return nil, fmt.Errorf("failed to read upgrade target file: %v", err)
	}
	target := &UpgradeTarget{}
	err = json.Unmarshal(b, target)
	if err != nil {
		vstr := strings.TrimSpace(string(b))
		var v *semver.Version
		v, err = semver.NewVersion(vstr)
		if err == nil {
			target.Version = *v
		}
	}
	if err != nil {
		klog.Errorf("invalid upgrade target file content(%v): %s, removing target file", err, string(b))
		err = os.Remove(commands.UPGRADE_TARGET_FILE)
		if err != nil && !os.IsNotExist(err) {
			klog.Errorf("error removing invalid upgrade target file: %v", err)
		}
		return nil, nil
	}
	return target, nil
}
