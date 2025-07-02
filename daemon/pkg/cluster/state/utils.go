package state

import (
	"context"
	"errors"
	"os"
	"os/exec"
	"strconv"
	"strings"

	"github.com/Masterminds/semver/v3"

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

func GetOlaresUpgradeTarget() (*semver.Version, error) {
	b, err := os.ReadFile(commands.UPGRADE_TARGET_FILE)
	if err != nil {
		if os.IsNotExist(err) {
			return nil, nil
		}
		return nil, err
	}
	target := strings.TrimSpace(string(b))
	version, err := semver.NewVersion(target)
	if err != nil {
		klog.Errorf("invalid upgrade target %s: %s, removing", target, err)
		err = os.Remove(commands.UPGRADE_TARGET_FILE)
		if err != nil && !os.IsNotExist(err) {
			return nil, err
		}
		return nil, nil
	}
	return version, nil
}

func IsUpgradeDownloadOnly() (bool, error) {
	_, err := os.Stat(commands.UPGRADE_DOWNLOADONLY_FILE)
	if err != nil {
		if os.IsNotExist(err) {
			return false, nil
		}
		return false, err
	}
	return true, nil
}

func IsUpgradeDownloadCompleted() (bool, error) {
	_, err := os.Stat(commands.UPGRADE_DOWNLOADED_FILE)
	if err != nil {
		if os.IsNotExist(err) {
			return false, nil
		}
		return false, err
	}
	return true, nil
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
	cmd := exec.CommandContext(ctx, cli.TERMINUS_CLI, "info", "show")

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
