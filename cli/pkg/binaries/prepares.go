package binaries

import (
	"strings"

	"bytetrade.io/web3os/installer/pkg/core/connector"
	"bytetrade.io/web3os/installer/pkg/core/logger"
	"bytetrade.io/web3os/installer/pkg/core/prepare"
)

type Ubuntu24AppArmorCheck struct {
	prepare.BasePrepare
}

func (p *Ubuntu24AppArmorCheck) PreCheck(runtime connector.Runtime) (bool, error) {
	sysInfo := runtime.GetSystemInfo()
	if !sysInfo.IsLinux() || !sysInfo.IsUbuntu() {
		return false, nil
	}

	if !sysInfo.IsUbuntuVersionEqual(connector.Ubuntu24) {
		return false, nil
	}

	cmd := "apparmor_parser --version"
	stdout, err := runtime.GetRunner().SudoCmd(cmd, false, true)
	if err != nil {
		logger.Errorf("check apparmor version error %v", err)
		return false, nil
	}

	if strings.Index(stdout, "4.0.1") < 0 {
		return true, nil
	}

	return false, nil
}
