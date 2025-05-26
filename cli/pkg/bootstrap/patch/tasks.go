package patch

import (
	"fmt"
	"path"
	"strings"

	"bytetrade.io/web3os/installer/pkg/utils"
	"github.com/pkg/errors"

	kubekeyapiv1alpha2 "bytetrade.io/web3os/installer/apis/kubekey/v1alpha2"
	"bytetrade.io/web3os/installer/pkg/binaries"
	"bytetrade.io/web3os/installer/pkg/common"
	"bytetrade.io/web3os/installer/pkg/core/connector"
	"bytetrade.io/web3os/installer/pkg/core/logger"
	"bytetrade.io/web3os/installer/pkg/core/util"
	"bytetrade.io/web3os/installer/pkg/manifest"
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

	var systemInfo = runtime.GetSystemInfo()
	var platformFamily = systemInfo.GetOsPlatformFamily()
	var pkgManager = systemInfo.GetPkgManager()
	switch platformFamily {
	case common.Debian:
		if _, err := util.GetCommand("add-apt-repository"); err != nil {
			if _, err := runtime.GetRunner().SudoCmd("apt install -y software-properties-common", false, true); err != nil {
				logger.Errorf("install add-apt-repository error %v", err)
				return err
			}
		}

		var cmd = fmt.Sprintf("add-apt-repository 'deb http://deb.debian.org/debian %s contrib non-free' -y", systemInfo.GetDebianVersionCode())
		if _, err := runtime.GetRunner().SudoCmd(cmd, false, true); err != nil {
			logger.Errorf("add os repo error %v", err)
			return err
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

		var cmd = "conntrack socat apache2-utils ntpdate net-tools make gcc bison flex tree unzip"
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

type SocatTask struct {
	common.KubeAction
	manifest.ManifestAction
}

func (t *SocatTask) Execute(runtime connector.Runtime) error {
	filePath, fileName, err := binaries.GetSocat(t.BaseDir, t.Manifest)
	if err != nil {
		logger.Errorf("failed to download socat: %v", err)
		return err
	}
	f := path.Join(filePath, fileName)
	if _, err := runtime.GetRunner().SudoCmd(fmt.Sprintf("tar xzvf %s -C %s", f, filePath), false, false); err != nil {
		logger.Errorf("failed to extract %s %v", f, err)
		return err
	}

	tp := path.Join(filePath, fmt.Sprintf("socat-%s", kubekeyapiv1alpha2.DefaultSocatVersion))
	if err := util.ChangeDir(tp); err == nil {
		if _, err := runtime.GetRunner().SudoCmd("./configure --prefix=/usr && make -j4 && make install && strip socat", false, false); err != nil {
			logger.Errorf("failed to install socat %v", err)
			return err
		}
	}
	if err := util.ChangeDir(runtime.GetBaseDir()); err != nil {
		logger.Errorf("failed to change dir %v", err)
		return err
	}

	return nil
}

type ConntrackTask struct {
	common.KubeAction
	manifest.ManifestAction
}

func (t *ConntrackTask) Execute(runtime connector.Runtime) error {
	flexFilePath, flexFileName, err := binaries.GetFlex(t.BaseDir, t.Manifest)
	if err != nil {
		logger.Errorf("failed to download flex: %v", err)
		return err
	}
	filePath, fileName, err := binaries.GetConntrack(t.BaseDir, t.Manifest)
	if err != nil {
		logger.Errorf("failed to download conntrack: %v", err)
		return err
	}
	fl := path.Join(flexFilePath, flexFileName)
	f := path.Join(filePath, fileName)

	if _, err := runtime.GetRunner().SudoCmd(fmt.Sprintf("tar xzvf %s -C %s", fl, filePath), false, true); err != nil {
		logger.Errorf("failed to extract %s %v", flexFilePath, err)
		return err
	}

	if _, err := runtime.GetRunner().SudoCmd(fmt.Sprintf("tar xzvf %s -C %s", f, filePath), false, true); err != nil {
		logger.Errorf("failed to extract %s %v", f, err)
		return err
	}

	// install
	fp := path.Join(flexFilePath, fmt.Sprintf("flex-%s", kubekeyapiv1alpha2.DefaultFlexVersion))
	if err := util.ChangeDir(fp); err == nil {
		if _, err := runtime.GetRunner().SudoCmd("autoreconf -i && ./configure --prefix=/usr && make -j4 && make install", false, true); err != nil {
			logger.Errorf("failed to install flex %v", err)
			return err
		}
	}

	tp := path.Join(filePath, fmt.Sprintf("conntrack-tools-conntrack-tools-%s", kubekeyapiv1alpha2.DefaultConntrackVersion))
	if err := util.ChangeDir(tp); err == nil {
		if _, err := runtime.GetRunner().SudoCmd("autoreconf -i && ./configure --prefix=/usr && make -j4 && make install", false, true); err != nil {
			logger.Errorf("failed to install conntrack %v", err)
			return err
		}
	}
	if err := util.ChangeDir(runtime.GetBaseDir()); err != nil {
		logger.Errorf("failed to change dir %v", err)
		return err
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
