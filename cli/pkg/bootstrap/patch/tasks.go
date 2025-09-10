package patch

import (
	"fmt"
	"strings"

	"github.com/beclab/Olares/cli/pkg/utils"
	"github.com/pkg/errors"

	"github.com/beclab/Olares/cli/pkg/common"
	"github.com/beclab/Olares/cli/pkg/core/connector"
	"github.com/beclab/Olares/cli/pkg/core/logger"
	"github.com/beclab/Olares/cli/pkg/core/util"
)

type EnableSSHTask struct {
	common.KubeAction
}

func (t *EnableSSHTask) Execute(runtime connector.Runtime) error {
	stdout, _ := runtime.GetRunner().SudoCmd("systemctl is-active ssh", false, false)
	if stdout != "active" {
		if _, err := runtime.GetRunner().SudoCmd("systemctl enable --now ssh", false, false); err != nil {
			return err
		}
	}

	return nil
}

type PatchTask struct {
	common.KubeAction
}

func (t *PatchTask) Execute(runtime connector.Runtime) error {
	var cmd string
	var debianFrontend = "DEBIAN_FRONTEND=noninteractive"
	var pre_reqs = "apt-transport-https ca-certificates curl cifs-utils"

	if _, err := util.GetCommand(common.CommandGPG); err != nil {
		pre_reqs = pre_reqs + " gnupg "
	}
	if _, err := util.GetCommand(common.CommandSudo); err != nil {
		pre_reqs = pre_reqs + " sudo "
	}
	if _, err := util.GetCommand(common.CommandUpdatePciids); err != nil {
		pre_reqs = pre_reqs + " pciutils "
	}
	if _, err := util.GetCommand(common.CommandIptables); err != nil {
		pre_reqs = pre_reqs + " iptables "
	}
	if _, err := util.GetCommand(common.CommandIp6tables); err != nil {
		pre_reqs = pre_reqs + " iptables "
	}
	if _, err := util.GetCommand(common.CommandIpset); err != nil {
		pre_reqs = pre_reqs + " ipset "
	}
	if _, err := util.GetCommand(common.CommandNmcli); err != nil {
		pre_reqs = pre_reqs + " network-manager "
	}

	pre_reqs += " conntrack socat apache2-utils net-tools make gcc bison flex tree unzip "

	var systemInfo = runtime.GetSystemInfo()
	var platformFamily = systemInfo.GetOsPlatformFamily()
	var pkgManager = systemInfo.GetPkgManager()
	switch platformFamily {
	case common.Debian:
		sourceType := "deb"
		repoURL := "https://deb.debian.org/debian"
		suite := systemInfo.GetDebianVersionCode()
		components := []string{"contrib", "non-free"}
		var aptToolAvailable bool
		if _, err := util.GetCommand("add-apt-repository"); err == nil {
			aptToolAvailable = true
		} else {
			if _, err := runtime.GetRunner().SudoCmd("apt install -y software-properties-common", false, true); err == nil {
				aptToolAvailable = true
			} else {
				logger.Infof("software-properties-common not available, try to update apt sources ourself")
			}
		}
		if aptToolAvailable {
			cmd := fmt.Sprintf("add-apt-repository '%s %s %s %s' -y", sourceType, repoURL, suite, strings.Join(components, " "))
			if _, err := runtime.GetRunner().SudoCmd(cmd, false, true); err != nil {
				return err
			}
		} else {
			err := utils.AddAptSource(sourceType, repoURL, suite, components)
			if err != nil {
				return err
			}
		}

		if systemInfo.IsDebianVersionEqual(connector.Debian13) {
			pre_reqs += " systemd-timesyncd "
		} else {
			pre_reqs += " ntpdate "
		}

		fallthrough
	case common.Ubuntu:
		if systemInfo.IsUbuntu() {
			if !systemInfo.IsPveOrPveLxc() && !systemInfo.IsRaspbian() {
				if _, err := runtime.GetRunner().SudoCmd("add-apt-repository universe -y", false, true); err != nil {
					logger.Errorf("add os repo error %v", err)
					return err
				}

				if _, err := runtime.GetRunner().SudoCmd("add-apt-repository multiverse -y", false, true); err != nil {
					logger.Errorf("add os repo error %v", err)
					return err
				}
			}
			pre_reqs += " ntpdate "
		}

		if _, err := runtime.GetRunner().SudoCmd(fmt.Sprintf("%s update -qq", pkgManager), false, true); err != nil {
			logger.Errorf("update os error %v", err)
			return err
		}

		logger.Debug("apt update success")

		if _, err := runtime.GetRunner().SudoCmd(fmt.Sprintf("%s install -y -qq %s", pkgManager, pre_reqs), false, true); err != nil {
			logger.Errorf("install deps %s error %v", pre_reqs, err)
			return err
		}

		if _, err := runtime.GetRunner().SudoCmd(fmt.Sprintf("%s %s install -y %s", debianFrontend, pkgManager, cmd), false, true); err != nil {
			logger.Errorf("install deps %s error %v", cmd, err)
			return err
		}

		if _, err := runtime.GetRunner().SudoCmd("update-pciids", false, true); err != nil {
			return fmt.Errorf("failed to update-pciids: %v", err)
		}

		if _, err := runtime.GetRunner().SudoCmd(fmt.Sprintf("%s %s install -y openssh-server", debianFrontend, pkgManager), false, true); err != nil {
			logger.Errorf("install deps %s error %v", cmd, err)
			return err
		}
	case common.CentOs, common.Fedora, common.RHEl:
		cmd = "conntrack socat httpd-tools ntpdate net-tools make gcc openssh-server"
		if _, err := runtime.GetRunner().SudoCmd(fmt.Sprintf("%s install -y %s", pkgManager, cmd), false, true); err != nil {
			logger.Errorf("install deps %s error %v", cmd, err)
			return err
		}
	}

	return nil
}

type CorrectHostname struct {
	common.KubeAction
}

func (t *CorrectHostname) Execute(runtime connector.Runtime) error {
	hostName := runtime.GetSystemInfo().GetHostname()
	if !utils.ContainsUppercase(hostName) {
		return nil
	}
	hostname := strings.ToLower(hostName)
	if _, err := runtime.GetRunner().SudoCmd(fmt.Sprintf("hostnamectl set-hostname %s", hostname), false, false); err != nil {
		return err
	}
	runtime.GetSystemInfo().SetHostname(hostname)
	return nil
}

type RaspbianCheckTask struct {
	common.KubeAction
}

func (t *RaspbianCheckTask) Execute(runtime connector.Runtime) error {
	// if util.IsExist(common.RaspbianCmdlineFile) || util.IsExist(common.RaspbianFirmwareFile) {
	systemInfo := runtime.GetSystemInfo()
	if systemInfo.IsRaspbian() {
		if _, err := util.GetCommand(common.CommandIptables); err != nil {
			_, err = runtime.GetRunner().SudoCmd("apt install -y iptables", false, false)
			if err != nil {
				logger.Errorf("%s install iptables error %v", common.Raspbian, err)
				return err
			}

			_, err = runtime.GetRunner().Cmd("systemctl disable --user gvfs-udisks2-volume-monitor", false, true)
			if err != nil {
				logger.Errorf("%s exec error %v", common.Raspbian, err)
				return err
			}

			_, err = runtime.GetRunner().Cmd("systemctl stop --user gvfs-udisks2-volume-monitor", false, true)
			if err != nil {
				logger.Errorf("%s exec error %v", common.Raspbian, err)
				return err
			}

			if !systemInfo.CgroupCpuEnabled() || !systemInfo.CgroupMemoryEnabled() {
				return fmt.Errorf("cpu or memory cgroups disabled, please edit /boot/cmdline.txt or /boot/firmware/cmdline.txt and reboot to enable it")
			}
		}
	}
	return nil
}

type DisableLocalDNSTask struct {
	common.KubeAction
}

func (t *DisableLocalDNSTask) Execute(runtime connector.Runtime) error {
	switch runtime.GetSystemInfo().GetOsPlatformFamily() {
	case common.Ubuntu, common.Debian:
		stdout, _ := runtime.GetRunner().SudoCmd("systemctl is-active systemd-resolved", false, false)
		if stdout == "active" {
			_, _ = runtime.GetRunner().SudoCmd("systemctl stop systemd-resolved.service", false, true)
			_, _ = runtime.GetRunner().SudoCmd("systemctl disable systemd-resolved.service", false, true)

			if utils.IsExist("/usr/bin/systemd-resolve") {
				_, _ = runtime.GetRunner().SudoCmd("mv /usr/bin/systemd-resolve /usr/bin/systemd-resolve.bak", false, true)
			}
			ok, err := utils.IsSymLink("/etc/resolv.conf")
			if err != nil {
				logger.Errorf("check /etc/resolv.conf error %v", err)
				return err
			}
			if ok {
				if _, err := runtime.GetRunner().SudoCmd("unlink /etc/resolv.conf && touch /etc/resolv.conf", false, true); err != nil {
					logger.Errorf("unlink /etc/resolv.conf error %v", err)
					return err
				}
			}

			if err = t.configResolvConf(runtime); err != nil {
				logger.Errorf("config /etc/resolv.conf error %v", err)
				return err
			}
		} else {
			if _, err := runtime.GetRunner().SudoCmd("cat /etc/resolv.conf > /etc/resolv.conf.bak", false, true); err != nil {
				logger.Errorf("backup /etc/resolv.conf error %v", err)
				return err
			}

			httpCode, _ := utils.GetHttpStatus("https://www.apple.com")
			if httpCode != 200 {
				if err := t.configResolvConf(runtime); err != nil {
					logger.Errorf("config /etc/resolv.conf error %v", err)
					return err
				}
			}

		}
	}

	sysInfo := runtime.GetSystemInfo()
	localIp := sysInfo.GetLocalIp()
	hostname := sysInfo.GetHostname()
	if stdout, _ := runtime.GetRunner().SudoCmd("hostname -i &>/dev/null", false, true); stdout == "" {
		if _, err := runtime.GetRunner().SudoCmd(fmt.Sprintf("echo %s %s >> /etc/hosts", localIp, hostname), false, true); err != nil {
			return errors.Wrap(err, "failed to set hostname mapping")
		}
	}

	if runtime.GetSystemInfo().IsWsl() {
		_, _ = runtime.GetRunner().SudoCmd("chattr +i /etc/hosts /etc/resolv.conf", false, false)
	}

	return nil
}

func (t *DisableLocalDNSTask) configResolvConf(runtime connector.Runtime) error {
	var err error
	var cmd string
	var secondNameserverOp string
	overrideOp := ">"
	appendOp := ">>"

	if common.CloudVendor == common.CloudVendorAliYun {
		secondNameserverOp = appendOp
		cmd = `echo 'nameserver 100.100.2.136' > /etc/resolv.conf`
		if _, err = runtime.GetRunner().SudoCmd(cmd, false, true); err != nil {
			logger.Errorf("exec %s error %v", cmd, err)
			return err
		}
	} else {
		secondNameserverOp = overrideOp
	}

	primaryDNSServer, secondaryDNSServer := "1.1.1.1", "114.114.114.114"
	if strings.Contains(t.KubeConf.Arg.RegistryMirrors, common.OlaresRegistryMirrorHost) || strings.Contains(t.KubeConf.Arg.RegistryMirrors, common.OlaresRegistryMirrorHostLegacy) {
		primaryDNSServer, secondaryDNSServer = secondaryDNSServer, primaryDNSServer
	}

	cmd = fmt.Sprintf("echo 'nameserver %s' %s /etc/resolv.conf", primaryDNSServer, secondNameserverOp)
	if _, err = runtime.GetRunner().SudoCmd(cmd, false, true); err != nil {
		logger.Errorf("exec %s error %v", cmd, err)
		return err
	}

	cmd = fmt.Sprintf("echo 'nameserver %s' >> /etc/resolv.conf", secondaryDNSServer)
	if _, err = runtime.GetRunner().SudoCmd(cmd, false, true); err != nil {
		logger.Errorf("exec %s error %v", cmd, err)
		return err
	}
	return nil
}
