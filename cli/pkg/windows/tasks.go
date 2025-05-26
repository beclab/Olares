package windows

import (
	"bufio"
	"fmt"
	"net"
	"os"
	"os/exec"
	"path/filepath"
	"strconv"
	"strings"

	kubekeyapiv1alpha2 "bytetrade.io/web3os/installer/apis/kubekey/v1alpha2"
	"bytetrade.io/web3os/installer/pkg/common"
	cc "bytetrade.io/web3os/installer/pkg/core/common"
	"bytetrade.io/web3os/installer/pkg/core/connector"
	"bytetrade.io/web3os/installer/pkg/core/logger"
	"bytetrade.io/web3os/installer/pkg/core/util"
	"bytetrade.io/web3os/installer/pkg/files"
	"bytetrade.io/web3os/installer/pkg/utils"
	templates "bytetrade.io/web3os/installer/pkg/windows/templates"
	"github.com/pkg/errors"
)

const (
	windowsAppPath = "AppData\\Local\\Microsoft\\WindowsApps"
	ubuntuTool     = "ubuntu.exe"
	distro         = "Ubuntu"

	OLARES_WINDOWS_FIREWALL_RULE_NAME = "OlaresRule"
)

type DownloadAppxPackage struct {
	common.KubeAction
}

func (d *DownloadAppxPackage) Execute(runtime connector.Runtime) error {
	var systemInfo = runtime.GetSystemInfo()
	var downloadCdnUrl = getDownloadCdnUrl(d.KubeConf.Arg.DownloadCdnUrl)

	appx := files.NewKubeBinary("wsl", systemInfo.GetOsArch(), systemInfo.GetOsType(), systemInfo.GetOsVersion(), systemInfo.GetOsPlatformFamily(), "2204", fmt.Sprintf("%s\\%s\\%s\\%s", systemInfo.GetHomeDir(), cc.DefaultBaseDir, "pkg", "components"), downloadCdnUrl)

	if err := appx.CreateBaseDir(); err != nil {
		return errors.Wrapf(errors.WithStack(err), "create file %s base dir failed", appx.FileName)
	}

	var exists = util.IsExist(appx.Path())
	if exists {
		p := appx.Path()
		output := util.LocalMd5Sum(p)
		if output != appx.Md5sum {
			util.RemoveFile(p)
			exists = false
		}
	}

	if !exists {
		logger.Debugf("download %s, url: %s", appx.FileName, appx.Url)
		if err := appx.Download(); err != nil {
			return fmt.Errorf("Failed to download %s binary: %s error: %w ", appx.ID, appx.Url, err)
		}
	}

	d.PipelineCache.Set(common.WslUbuntuBinaries, appx)
	return nil
}

type InstallAppxPackage struct {
	common.KubeAction
}

func (i *InstallAppxPackage) Execute(runtime connector.Runtime) error {
	wslAppxPackageObj, ok := i.PipelineCache.Get(common.WslUbuntuBinaries)
	if !ok {
		return errors.New("get WSL appx package from pipelinecache failed")
	}

	wslAppxPackage := wslAppxPackageObj.(*files.KubeBinary)

	var ps = &utils.PowerShellCommandExecutor{
		Commands: []string{fmt.Sprintf("Add-AppxPackage %s -ForceUpdateFromAnyVersion", wslAppxPackage.Path())},
	}

	if _, err := ps.Run(); err != nil {
		return fmt.Errorf("Unable to install the package because the resources it modifies are currently in use. Please close the current terminal and try again.")
	}

	return nil
}

type DownloadWSLInstallPackage struct {
	common.KubeAction
}

func (d *DownloadWSLInstallPackage) Execute(runtime connector.Runtime) error {
	var systemInfo = runtime.GetSystemInfo()
	var downloadCdnUrl = getDownloadCdnUrl(d.KubeConf.Arg.DownloadCdnUrl)
	var osArch = systemInfo.GetOsArch()
	var osType = systemInfo.GetOsType()
	var osVersion = systemInfo.GetOsVersion()
	var osPlatformFamily = systemInfo.GetOsPlatformFamily()
	wslInstallationPackage := files.NewKubeBinary("wslpackage", osArch, osType, osVersion, osPlatformFamily, kubekeyapiv1alpha2.DefaultWSLInstallPackageVersion, fmt.Sprintf("%s\\%s\\%s\\%s", systemInfo.GetHomeDir(), cc.DefaultBaseDir, "pkg", "components"), downloadCdnUrl)

	if err := wslInstallationPackage.CreateBaseDir(); err != nil {
		return errors.Wrapf(errors.WithStack(err), "create file %s base dir failed", wslInstallationPackage.FileName)
	}

	var exists = util.IsExist(wslInstallationPackage.Path())
	if exists {
		p := wslInstallationPackage.Path()
		output := util.LocalMd5Sum(p)
		if output != wslInstallationPackage.Md5sum {
			util.RemoveFile(p)
			exists = false
		}
	}

	if !exists {
		if err := wslInstallationPackage.Download(); err != nil {
			return fmt.Errorf("Failed to download %s binary: %s error: %w ", wslInstallationPackage.ID, wslInstallationPackage.Url, err)
		}
	}
	d.PipelineCache.Set(common.WslBinaries+"-"+osArch, wslInstallationPackage)
	return nil
}

type UpdateWSL struct {
	common.KubeAction
}

func (u *UpdateWSL) Execute(runtime connector.Runtime) error {
	var disableWslUpdate = os.Getenv("DISABLE_WSL_UPDATE")
	if strings.EqualFold(disableWslUpdate, "1") {
		return nil
	}

	var wslConfigFile = fmt.Sprintf("%s\\%s", runtime.GetSystemInfo().GetHomeDir(), templates.WSLConfigValue.Name())

	file, err := os.Create(wslConfigFile)
	defer file.Close()
	if err != nil {
		return errors.Wrap(errors.WithStack(err), fmt.Sprintf("create wsl config %s failed", wslConfigFile))
	}

	systemInfo := runtime.GetSystemInfo()
	memory := u.getMemroy(systemInfo.GetTotalMemory())
	var data = util.Data{
		"Memory": memory,
	}

	wslConfigStr, err := util.Render(templates.WSLConfigValue, data)
	if err != nil {
		return errors.Wrap(errors.WithStack(err), "render account template failed")
	}

	if _, err = file.WriteString(wslConfigStr); err != nil {
		return errors.Wrap(errors.WithStack(err), fmt.Sprintf("write wsl config %s failed", wslConfigFile))
	}

	wslInstallPackageObj, ok := u.PipelineCache.Get(common.WslBinaries + "-" + systemInfo.GetOsArch())
	if !ok {
		return errors.New("get WSL install package from pipelinecache failed")
	}
	wslInstallPackage := wslInstallPackageObj.(*files.KubeBinary)

	var wslInfo = new(Wsl)
	wslInfo.GetVersion()
	wslInfo.PrintVersion()

	if !wslInfo.IsInstalled() || wslInfo.CompareTo(kubekeyapiv1alpha2.DefaultWSLInstallPackageVersion) < 0 {
		logger.Info("WSL is updating. This process may take a few minutes. Please wait...")
		installoutput, err := wslInfo.Install(wslInstallPackage.Path())
		if err != nil {
			return errors.Wrap(errors.WithStack(err), fmt.Sprintf("Install WSL failed, message: %s", installoutput))
		}

		wslInfo.GetVersion()
		wslInfo.PrintVersion()
	}

	return nil
}

func (u *UpdateWSL) getMemroy(total uint64) uint64 {
	var memory uint64 = 12
	m := os.Getenv("WSL_MEMORY")
	if m == "" {
		return memory
	}

	sets, err := strconv.ParseUint(m, 10, 64)
	if err != nil {
		return memory
	}

	localMemeory := total / 1024 / 1024 / 1024
	if localMemeory < sets {
		if localMemeory > memory {
			return memory
		} else {
			return localMemeory
		}
	}

	return sets
}

type InstallWSLDistro struct {
	common.KubeAction
}

func (i *InstallWSLDistro) Execute(runtime connector.Runtime) error {
	var homeDir = runtime.GetSystemInfo().GetHomeDir()
	var installerPath = filepath.Join(homeDir, windowsAppPath, ubuntuTool)

	logger.Infof("%s path: %s", ubuntuTool, installerPath)

	var checkInstallerPs = &utils.PowerShellCommandExecutor{
		Commands: []string{fmt.Sprintf("Test-Path %s", installerPath)},
	}
	installerExists, err := checkInstallerPs.Run()
	if err != nil {
		logger.Errorf("check %s path %s failed, error: %s", ubuntuTool, installerPath, err.Error())
		return err
	}

	installerExists = strings.TrimSpace(installerExists)
	if installerExists != "True" {
		logger.Errorf("%s not found in %s", ubuntuTool, installerPath)
		return fmt.Errorf("%s not found in %s", ubuntuTool, installerPath)
	}

	var cmd = &utils.DefaultCommandExecutor{
		Commands:    []string{"install", "--root"},
		PrintOutput: true,
	}
	logger.Infof("install ubuntu distro...")
	output, err := cmd.RunCmd(installerPath, utils.UTF8)
	if err != nil {
		return showUbuntuErrorMsg(output, err)
	}

	logger.Infof("Install WSL Ubuntu Distro %s successd\n", distro)

	return nil
}

type MoveDistro struct {
	common.KubeAction
}

func (m *MoveDistro) Execute(runtime connector.Runtime) error {
	distroStoreDriver, _ := m.PipelineCache.GetMustString(common.CacheWindowsDistroStoreLocation)
	distroStoreLocationNums, _ := m.PipelineCache.GetMustString(common.CacheWindowsDistroStoreLocationNums)
	if distroStoreDriver == "" {
		return errors.New("get distro location failed")
	}
	if distroStoreLocationNums == "1" {
		logger.Infof("distro store default location: %s", distroStoreDriver)
		return nil
	}
	var distroStorePath = fmt.Sprintf("%s:\\.olares\\distro", distroStoreDriver)
	if !utils.IsExist(distroStorePath) {
		if err := utils.CreateDir(distroStorePath); err != nil {
			return errors.WithStack(fmt.Errorf("create dir %s failed: %v", distroStorePath, err))
		}
	}

	var si = runtime.GetSystemInfo()
	var aclCmd = &utils.DefaultCommandExecutor{
		Commands: []string{fmt.Sprintf("%s:\\.olares", distroStoreDriver), "/grant", fmt.Sprintf("*S-1-5-32-545:(OI)(CI)F")},
	}

	logger.Infof("distro store path: %s, user: %s", distroStorePath, si.GetUsername())

	if aclRes, err := aclCmd.RunCmd("icacls", utils.UTF8); err != nil {
		logger.Debugf("icacls exec failed, err: %v, message: %s", err, aclRes)
		return err
	}

	var cmd = &utils.DefaultCommandExecutor{
		Commands: []string{"--shutdown"},
	}

	_, _ = cmd.RunCmd("wsl", utils.DEFAULT)

	var removeCmd = &utils.DefaultCommandExecutor{
		Commands: []string{"--manage", distro, "--move", distroStorePath},
	}

	if _, err := removeCmd.RunCmd("wsl", utils.DEFAULT); err != nil {
		return err
	}

	cmd = &utils.DefaultCommandExecutor{
		Commands: []string{"--shutdown"},
	}

	_, _ = cmd.RunCmd("wsl", utils.DEFAULT)

	return nil
}

type ConfigWslConf struct {
	common.KubeAction
}

func (c *ConfigWslConf) Execute(runtime connector.Runtime) error {
	var cmd = &utils.DefaultCommandExecutor{
		Commands: []string{"-d", distro, "-u", "root", "bash", "-c", "echo -e '[boot]\\nsystemd=true\\ncommand=\"mount --make-rshared /\"\\n[network]\\ngenerateHosts=false\\ngenerateResolvConf=false\\nhostname=terminus' > /etc/wsl.conf"},
	}
	if _, err := cmd.RunCmd("wsl", utils.DEFAULT); err != nil {
		return errors.Wrap(errors.WithStack(err), fmt.Sprintf("config wsl %s hosts and dns failed", distro))
	}

	cmd = &utils.DefaultCommandExecutor{
		Commands: []string{"--shutdown", distro},
	}
	if _, err := cmd.RunCmd("wsl", utils.DEFAULT); err != nil {
		return errors.Wrap(errors.WithStack(err), fmt.Sprintf("shutdown wsl %s failed", distro))
	}

	return nil
}

type ConfigWSLForwardRules struct {
	common.KubeAction
}

func (c *ConfigWSLForwardRules) ipFormat(wslIp string) string {
	var pip net.IP
	wslIp = strings.ReplaceAll(wslIp, "\n", "\r")
	var ipStrs = strings.Split(wslIp, "\r")
	if len(ipStrs) == 0 {
		return ""
	}

	for _, ipStr := range ipStrs {
		var tmp = strings.TrimSpace(ipStr)
		if tmp == "" {
			continue
		}

		pip = net.ParseIP(tmp)
		if pip != nil {
			break
		}
	}

	if pip == nil {
		return ""
	}

	return pip.To4().String()
}

func (c *ConfigWSLForwardRules) Execute(runtime connector.Runtime) error {
	var cmd = &utils.DefaultCommandExecutor{
		Commands: []string{"wsl", "-d", distro, "bash", "-c", "ip address show eth0 | grep inet | grep -v inet6 | cut -d ' ' -f 6 | cut -d '/' -f 1"},
	}

	ip, err := cmd.Run()
	if err != nil {
		return errors.Wrap(errors.WithStack(err), fmt.Sprintf("get wsl %s ip failed", distro))
	}

	ip = c.ipFormat(ip)
	if ip == "" {
		return fmt.Errorf("wsl ip address not found")
	}

	logger.Infof("wsl %s, ip: %s", distro, ip)

	cmd = &utils.DefaultCommandExecutor{
		Commands: []string{fmt.Sprintf("netsh interface portproxy add v4tov4 listenport=80 listenaddress=0.0.0.0 connectport=80 connectaddress=%s", ip)},
	}
	if output, err := cmd.Run(); err != nil {
		logger.Debugf("set portproxy listenport 80 failed %v, message: %s", err, output)
		return errors.Wrap(errors.WithStack(err), fmt.Sprintf("config wsl %s forward rules failed", distro))
	}

	cmd = &utils.DefaultCommandExecutor{
		Commands: []string{fmt.Sprintf("netsh interface portproxy add v4tov4 listenport=443 listenaddress=0.0.0.0 connectport=443 connectaddress=%s", ip)},
	}
	if output, err := cmd.Run(); err != nil {
		logger.Debugf("set portproxy listenport 443 failed %v, message: %s", err, output)
		return errors.Wrap(errors.WithStack(err), fmt.Sprintf("config wsl %s forward rules failed", distro))
	}

	cmd = &utils.DefaultCommandExecutor{
		Commands: []string{fmt.Sprintf("netsh interface portproxy add v4tov4 listenport=30180 listenaddress=0.0.0.0 connectport=30180 connectaddress=%s", ip)},
	}
	if output, err := cmd.Run(); err != nil {
		logger.Debugf("set portproxy listenport 30180 failed %v, message: %s", err, output)
		return errors.Wrap(errors.WithStack(err), fmt.Sprintf("config wsl %s forward rules failed", distro))
	}

	return nil
}

type ConfigWSLHostsAndDns struct {
	common.KubeAction
}

func (c *ConfigWSLHostsAndDns) Execute(runtime connector.Runtime) error {
	var cmd = &utils.DefaultCommandExecutor{
		Commands: []string{"-d", distro, "-u", "root", "bash", "-c", "chattr -i /etc/hosts /etc/resolv.conf && "},
	}
	_, _ = cmd.RunCmd("wsl", utils.DEFAULT)

	cmd = &utils.DefaultCommandExecutor{
		Commands: []string{"-d", distro, "-u", "root", "bash", "-c", "echo -e '127.0.0.1 localhost\\n$(ip -4 addr show eth0 | grep -oP '(?<=inet\\s)\\d+(\\.\\d+){3}') $(hostname)' > /etc/hosts && echo -e 'nameserver 1.1.1.1\\nnameserver 1.0.0.1' > /etc/resolv.conf"},
	}

	if _, err := cmd.RunCmd("wsl", utils.DEFAULT); err != nil {
		return errors.Wrap(errors.WithStack(err), fmt.Sprintf("config wsl %s hosts and dns failed", distro))
	}

	return nil
}

type ConfigWindowsFirewallRule struct {
	common.KubeAction
}

func (c *ConfigWindowsFirewallRule) Execute(runtime connector.Runtime) error {
	var setFirewallRule bool = false
	var autoSetFirewallRule = os.Getenv(common.ENV_AUTO_ADD_FIREWALL_RULES)
	autoSetFirewallRule = strings.TrimSpace(autoSetFirewallRule)

	switch {
	case autoSetFirewallRule != "":
		setFirewallRule = true
		break
	default:
		scanner := bufio.NewScanner(os.Stdin)

		for {
			fmt.Print("\nAccessing Olares requires setting up firewall rules, specifically adding TCP inbound rules for ports 80, 443, and 30180.\nDo you want to set up the firewall rules? (yes/no): ")
			scanner.Scan()
			confirmation := scanner.Text()
			confirmation = strings.TrimSpace(confirmation)
			confirmation = strings.ToLower(confirmation)

			switch confirmation {
			case "y", "yes":
				setFirewallRule = true
				break
			case "n", "no":
				break
			default:
				continue
			}
			break
		}
	}

	if !setFirewallRule {
		fmt.Printf("\nFirewall settings have been skipped. \nIf you want to access the Olares application, please go to the Windows Defender Firewall rules and add an inbound rule for TCP protocol with port numbers 80, 443, and 30180.\n\n\n")
		return nil
	}

	var ps = &utils.PowerShellCommandExecutor{
		Commands: []string{fmt.Sprintf("Get-NetFirewallRule | Where-Object { $_.DisplayName -eq \"%s\" -and $_.Enabled -eq 'True'} | Get-NetFirewallPortFilter | Where-Object { $_.LocalPort -eq 80 -and $_.LocalPort -eq 443 -and $_.LocalPort -eq 30180 -and $_.Protocol -eq 'TCP' } ", OLARES_WINDOWS_FIREWALL_RULE_NAME)},
	}
	rules, _ := ps.Run()
	rules = strings.TrimSpace(rules)
	if rules == "" {
		ps = &utils.PowerShellCommandExecutor{
			Commands: []string{fmt.Sprintf("New-NetFirewallRule -DisplayName \"%s\" -Direction Inbound -Protocol TCP -LocalPort 80,443,30180 -Action Allow", OLARES_WINDOWS_FIREWALL_RULE_NAME)},
		}
		if _, err := ps.Run(); err != nil {
			return errors.Wrap(errors.WithStack(err), fmt.Sprintf("config windows firewall rule %s failed", OLARES_WINDOWS_FIREWALL_RULE_NAME))
		}
	}

	return nil
}

type InstallTerminus struct {
	common.KubeAction
}

func (i *InstallTerminus) Execute(runtime connector.Runtime) error {
	var systemInfo = runtime.GetSystemInfo()
	// var windowsUserPath = convertPath(systemInfo.GetHomeDir())

	var envs = []string{
		fmt.Sprintf("export %s=%s", common.ENV_KUBE_TYPE, i.KubeConf.Arg.Kubetype),
		fmt.Sprintf("export %s=%s", common.ENV_REGISTRY_MIRRORS, i.KubeConf.Arg.RegistryMirrors),
		fmt.Sprintf("export %s=%s", common.ENV_CLOUDFLARE_ENABLE, i.KubeConf.Arg.Cloudflare.Enable),
		fmt.Sprintf("export %s=%s", common.ENV_FRP_ENABLE, i.KubeConf.Arg.Frp.Enable),
		fmt.Sprintf("export %s=%s", common.ENV_FRP_SERVER, i.KubeConf.Arg.Frp.Server),
		fmt.Sprintf("export %s=%s", common.ENV_FRP_PORT, i.KubeConf.Arg.Frp.Port),
		fmt.Sprintf("export %s=%s", common.ENV_FRP_AUTH_METHOD, i.KubeConf.Arg.Frp.AuthMethod),
		fmt.Sprintf("export %s=%s", common.ENV_FRP_AUTH_TOKEN, i.KubeConf.Arg.Frp.AuthToken),
		fmt.Sprintf("export %s=%d", common.ENV_TOKEN_MAX_AGE, i.KubeConf.Arg.TokenMaxAge),
		fmt.Sprintf("export %s=%s", common.ENV_PREINSTALL, os.Getenv(common.ENV_PREINSTALL)),
		fmt.Sprintf("export %s=%s", common.ENV_MARKET_PROVIDER, i.KubeConf.Arg.MarketProvider),
		fmt.Sprintf("export %s=%s", common.ENV_TERMINUS_CERT_SERVICE_API, i.KubeConf.Arg.TerminusCertServiceAPI),
		fmt.Sprintf("export %s=%s", common.ENV_TERMINUS_DNS_SERVICE_API, i.KubeConf.Arg.TerminusDNSServiceAPI),
		fmt.Sprintf("export %s=%s", common.ENV_HOST_IP, systemInfo.GetLocalIp()),
		fmt.Sprintf("export %s=%s", common.ENV_DISABLE_HOST_IP_PROMPT, os.Getenv(common.ENV_DISABLE_HOST_IP_PROMPT)),
		fmt.Sprintf("export %s=%s", common.ENV_DOWNLOAD_CDN_URL, i.KubeConf.Arg.DownloadCdnUrl),
		fmt.Sprintf("export %s=%s", common.ENV_NVIDIA_CONTAINER_REPO_MIRROR, os.Getenv(common.ENV_NVIDIA_CONTAINER_REPO_MIRROR)),
	}

	var bashUrl = fmt.Sprintf("https://%s", cc.DefaultBashUrl)
	var defaultDomainName = os.Getenv(common.ENV_TERMINUS_OS_DOMAINNAME)
	if !utils.IsValidDomain(defaultDomainName) {
		defaultDomainName = ""
	}
	if defaultDomainName != "" {
		envs = append(envs, fmt.Sprintf("export %s=%s", common.ENV_TERMINUS_OS_DOMAINNAME, defaultDomainName))
		bashUrl = fmt.Sprintf("https://%s", defaultDomainName)
	}

	for key, val := range common.TerminusGlobalEnvs {
		envs = append(envs, fmt.Sprintf("export %s=%s", key, val))
	}

	var downloadUrl = i.KubeConf.Arg.DownloadCdnUrl
	if downloadUrl == "" {
		downloadUrl = cc.DownloadUrl
	}
	var installScript = fmt.Sprintf("curl -fsSL %s | bash -", bashUrl)
	if i.KubeConf.Arg.OlaresVersion != "" {
		var installFile = fmt.Sprintf("install-wizard-v%s.tar.gz", i.KubeConf.Arg.OlaresVersion)
		installScript = fmt.Sprintf("curl -fsSLO %s/%s && tar -xf %s -C ./ ./install.sh && rm -rf %s && bash ./install.sh",
			downloadUrl, installFile, installFile, installFile)
	}

	var params = strings.Join(envs, " && ")

	var cmd = &utils.DefaultCommandExecutor{
		Commands:  []string{"wsl", "-d", distro, "-u", "root", "--cd", "/root", "bash", "-c", fmt.Sprintf("%s && %s", params, installScript)},
		PrintLine: true,
	}

	if _, err := cmd.Exec(); err != nil {
		return errors.Wrap(errors.WithStack(err), fmt.Sprintf("install Olares %s failed", distro))
	}

	exec.Command("cmd", "/C", "wsl", "-d", distro, "--exec", "dbus-launch", "true").Run()

	return nil
}

func convertPath(windowsPath string) string {
	linuxPath := strings.ReplaceAll(windowsPath, `\`, `/`)
	if len(linuxPath) > 1 && linuxPath[1] == ':' {
		drive := strings.ToLower(string(linuxPath[0]))
		linuxPath = "/mnt/" + drive + linuxPath[2:]
	}

	return linuxPath
}

type UninstallOlares struct {
	common.KubeAction
}

func (u *UninstallOlares) Execute(runtime connector.Runtime) error {
	var cmd = &utils.DefaultCommandExecutor{
		Commands: []string{"--unregister", "Ubuntu"},
	}
	_, _ = cmd.RunCmd("wsl", utils.DEFAULT)

	return nil
}

type RemoveFirewallRule struct {
	common.KubeAction
}

func (r *RemoveFirewallRule) Execute(runtime connector.Runtime) error {
	(&utils.PowerShellCommandExecutor{
		Commands: []string{fmt.Sprintf("Remove-NetFirewallRule -DisplayName \"%s\"", OLARES_WINDOWS_FIREWALL_RULE_NAME)},
	}).Run()

	return nil
}

type RemovePortProxy struct {
	common.KubeAction
}

func (r *RemovePortProxy) Execute(runtime connector.Runtime) error {
	var ports = []string{"80", "443", "30180"}
	for _, port := range ports {
		(&utils.DefaultCommandExecutor{
			Commands: []string{"interface", "portproxy", "delete", "v4tov4", fmt.Sprintf("listenport=%s", port), "listenaddress=0.0.0.0"}}).RunCmd("netsh", utils.DEFAULT)
	}

	return nil
}

type GetDiskPartition struct {
	common.KubeAction
}

func (g *GetDiskPartition) Execute(runtime connector.Runtime) error {
	var partitions []string
	paths := utils.GetDrives()
	if paths == nil || len(paths) == 0 {
		return fmt.Errorf("Unable to retrieve disk information")
	}

	for _, path := range paths {
		_, free, err := utils.GetDiskSpace(path)
		if err != nil {
			continue
		}
		partitions = append(partitions, fmt.Sprintf("%s_%s", path, utils.FormatBytes(int64(free))))
	}

	if len(partitions) == 0 {
		return fmt.Errorf("Unable to retrieve disk space information")
	}
	fmt.Printf("\nInstalling Olares will create a WSL Ubuntu Distro and occupy at least 80 GB of disk space. \nPlease select the drive where you want to install it. \nAvailable drives and free space:\n")
	for _, v := range partitions {
		var tmp = strings.Split(v, "_")
		fmt.Printf("%s  Free Disk: %s\n", tmp[0], tmp[1])
	}

	var enterPath string
	var useDefaultDisk = os.Getenv(common.ENV_DEFAULT_WSL_DISTRO_LOCATION)
	useDefaultDisk = strings.TrimSpace(useDefaultDisk)

	switch {
	case useDefaultDisk != "":
		enterPath = "C"
		break
	default:
		scanner := bufio.NewScanner(os.Stdin)

		for {
			fmt.Printf("\nPlease enter the drive letter (e.g., C):")

			scanner.Scan()
			enterPath = scanner.Text()
			enterPath = strings.TrimSpace(enterPath)
			checkPathValid := g.checkEnter(enterPath, partitions)
			if !checkPathValid {
				continue
			}
			break
		}
	}

	g.PipelineCache.Set(common.CacheWindowsDistroStoreLocation, enterPath)
	g.PipelineCache.Set(common.CacheWindowsDistroStoreLocationNums, len(partitions))
	fmt.Printf("\n")

	return nil
}

func (g *GetDiskPartition) checkEnter(enterPath string, partitions []string) bool {
	var res bool = false
	for _, v := range partitions {
		var tmp = fmt.Sprintf("%s:\\", enterPath)
		var p = strings.Split(v, "_")
		if tmp == p[0] {
			res = true
			break
		}
	}

	return res
}

func getDownloadCdnUrl(downloadCdnUrlFromEnv string) string {
	downloadCdnUrl := strings.TrimSuffix(downloadCdnUrlFromEnv, "/")
	if downloadCdnUrl == "" {
		downloadCdnUrl = cc.DownloadUrl
	}
	return downloadCdnUrl
}

func showUbuntuErrorMsg(msg string, err error) error {
	if msg == "" {
		fmt.Printf(`
Stop Installation !!!!!!!

Installing Windows Olares will use the Ubuntu Distro. It has been detected that there is already an existing Ubuntu Distro in the system.
You can use the 'wsl -l --all' command to view the list of WSL Distros.
To proceed with the installation of Olares, you need to unregister the existing Ubuntu Distro. If your Ubuntu Distro contains important information, please back it up first, then unregister the Ubuntu Distro.
Uninstallation command: 'wsl --unregister Ubuntu'
After the unregister Ubuntu Distro is complete, please reinstall Olares.

Error message: %v
`, err)
		return fmt.Errorf("need to unregister Ubuntu Distro")
	}

	fmt.Printf(`
Stop Installation !!!!!!!

An unknown error occurred while updating WSL.
Please check the Control Panel > Programs > Windows Features to ensure that Windows Subsystem for Linux and Virtual Machine Platform are enabled, then and reinstall them. Error message: %s %v
`, msg, err)

	return fmt.Errorf("need to check system status")
}
