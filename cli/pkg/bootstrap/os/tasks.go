/*
 Copyright 2021 The KubeSphere Authors.

 Licensed under the Apache License, Version 2.0 (the "License");
 you may not use this file except in compliance with the License.
 You may obtain a copy of the License at

     http://www.apache.org/licenses/LICENSE-2.0

 Unless required by applicable law or agreed to in writing, software
 distributed under the License is distributed on an "AS IS" BASIS,
 WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
 See the License for the specific language governing permissions and
 limitations under the License.
*/

package os

import (
	"fmt"
	"io/fs"
	"io/ioutil"
	"os"
	"path"
	"path/filepath"
	"strconv"
	"strings"

	osrelease "github.com/dominodatalab/os-release"
	"github.com/pkg/errors"

	"bytetrade.io/web3os/installer/pkg/bootstrap/os/repository"
	"bytetrade.io/web3os/installer/pkg/bootstrap/os/templates"
	"bytetrade.io/web3os/installer/pkg/common"
	cc "bytetrade.io/web3os/installer/pkg/core/common"
	"bytetrade.io/web3os/installer/pkg/core/connector"
	"bytetrade.io/web3os/installer/pkg/core/logger"
	"bytetrade.io/web3os/installer/pkg/core/util"
	"bytetrade.io/web3os/installer/pkg/utils"
)

// pve-lxc
type PatchLxcEnvVars struct {
	common.KubeAction
}

func (p *PatchLxcEnvVars) Execute(runtime connector.Runtime) error {
	var cmd = `sed -n '/export PATH=\"\/usr\/local\/bin:$PATH\"/p' ~/.bashrc`
	if res, _ := runtime.GetRunner().Cmd(cmd, false, false); res == "" {
		if _, err := runtime.GetRunner().Cmd("echo 'export PATH=\"/usr/local/bin:$PATH\"' >> ~/.bashrc", false, false); err != nil {
			return err
		}

		os.Setenv("PATH", "/usr/local/bin:"+os.Getenv("PATH"))
	}
	return nil
}

type PatchLxcInitScript struct {
	common.KubeAction
}

func (p *PatchLxcInitScript) Execute(runtime connector.Runtime) error {
	filePath := path.Join("/", "etc", templates.InitPveLxcTmpl.Name())

	lxcPatchScriptStr, err := util.Render(templates.InitPveLxcTmpl, nil)
	if err != nil {
		return errors.Wrap(errors.WithStack(err), "render lxc patch template failed")
	}

	if err := util.WriteFile(filePath, []byte(lxcPatchScriptStr), cc.FileMode0755); err != nil {
		return errors.Wrap(errors.WithStack(err), fmt.Sprintf("write lxc patch %s failed", filePath))
	}

	_, _ = runtime.GetRunner().Cmd("/etc/rc.local", false, false)
	return nil
}

type RemoveCNDomain struct {
	common.KubeAction
}

func (r *RemoveCNDomain) Execute(runtime connector.Runtime) error {
	if res, _ := runtime.GetRunner().Cmd("sed -n '/search/p' /etc/resolv.conf", false, false); res != "" {
		if _, err := runtime.GetRunner().Cmd("sed -i '/search/d' /etc/resolv.conf", false, false); err != nil {
			return err
		}
	}
	return nil
}

// pve
type PveAptUpdateSourceCheck struct {
	common.KubeAction
}

func (p *PveAptUpdateSourceCheck) Execute(runtime connector.Runtime) error {
	if _, err := runtime.GetRunner().Cmd("apt-get update -qq", false, false); err != nil {

		fmt.Printf("\n\nNOTE: \nThe PVE apt-get update has failed. Please check the source repository. \n\nIf you are a Non-Enterprise user:\n1. Disable the Enterprise Repository in the PVE Control Panel.\n2. Or remove the Enterprise Repository files located in /etc/apt/sources.list.d/.\n\n\n")

		return err
	}

	return nil
}

// general
type UpdateNtpDateTask struct {
	common.KubeAction
}

func (t *UpdateNtpDateTask) Execute(runtime connector.Runtime) error {
	if _, err := runtime.GetRunner().Cmd("apt remove unattended-upgrades -y", false, true); err != nil {
		return err
	}

	var ntpPkg = " ntpdate "
	var systemInfo = runtime.GetSystemInfo()
	if systemInfo.IsUbuntu() && systemInfo.IsUbuntuVersionEqual(connector.Ubuntu24) {
		ntpPkg += " util-linux-extra "
	}
	if _, err := runtime.GetRunner().Cmd(fmt.Sprintf("apt install %s -y", ntpPkg), false, true); err != nil {
		return err
	}

	ntpdateCommand, err := util.GetCommand(common.CommandNtpdate)
	if err != nil {
		return fmt.Errorf("getntpdate command error: %v", err)
	}

	if _, err := runtime.GetRunner().Cmd(fmt.Sprintf("%s -b -u pool.ntp.org", ntpdateCommand), false, true); err != nil {
		return err
	}

	return nil
}

type TimeSyncTask struct {
	common.KubeAction
}

func (t *TimeSyncTask) Execute(runtime connector.Runtime) error {
	// var cmd = `sysctl -w kernel.printk="3 3 1 7"`
	// if _, err := runtime.GetRunner().SudoCmd(cmd, false, true); err != nil {
	// 	logger.Errorf("failed to execute %s: %v", cmd, err)
	// 	return err
	// }

	var hwclockCmd = ""
	ntpdatePath, _ := util.GetCommand(common.CommandNtpdate)
	hwclockCommand, err := util.GetCommand(common.CommandHwclock)
	if err != nil {
		logger.Debugf("hwclock not found")
	} else {
		if _, err := runtime.GetRunner().Cmd(fmt.Sprintf("%s -w", hwclockCommand), false, true); err != nil {
			logger.Debugf("hwclock set the RTC from the system time error %v", err)
		} else {
			hwclockCmd = fmt.Sprintf(" && %s -w", hwclockCommand)
		}
	}

	cronContent := fmt.Sprintf(`#!/bin/sh
%s -b -u pool.ntp.org %s
exit 0`, ntpdatePath, hwclockCmd)

	// cronContent := fmt.Sprintf(`#!/bin/sh
	// %s -b -u pool.ntp.org && %s -w
	// exit 0`, ntpdatePath, hwclockPath)

	// cronContent := fmt.Sprintf(`#!/bin/sh
	// %s -b -u pool.ntp.org
	// exit 0`, ntpdatePath)
	cronFile := path.Join(runtime.GetBaseDir(), "cron.ntpdate")
	if err := ioutil.WriteFile(cronFile, []byte(cronContent), 0700); err != nil {
		logger.Errorf("Failed to write cron.ntpdate: %v", err)
		return err
	}

	if _, err := runtime.GetRunner().SudoCmd(fmt.Sprintf("/bin/sh %s", cronFile), false, true); err != nil {
		logger.Errorf("failed to execute cron.ntpdate: %v", err)
		return err
	}

	var cmd = fmt.Sprintf("cat %s > /etc/cron.daily/ntpdate && chmod 0700 /etc/cron.daily/ntpdate && rm -rf %s", cronFile, cronFile)
	if _, err := runtime.GetRunner().SudoCmd(cmd, false, true); err != nil {
		logger.Errorf("failed to execute %s: %v", cmd, err)
		return err
	}

	return nil
}

type ConfigProxyTask struct {
	common.KubeAction
}

func (t *ConfigProxyTask) Execute(runtime connector.Runtime) error {
	if common.ResolvProxy == "" {
		return nil
	}

	var cmd = fmt.Sprintf("echo nameserver %s > /etc/resolv.conf", common.ResolvProxy)
	if _, err := runtime.GetRunner().SudoCmd(cmd, false, true); err != nil {
		logger.Errorf("failed to execute %s: %v", cmd, err)
		return err
	}

	return nil
}

type NodeConfigureOS struct {
	common.KubeAction
}

func (n *NodeConfigureOS) Execute(runtime connector.Runtime) error {
	host := runtime.RemoteHost()
	if err := addUsers(runtime, host); err != nil {
		return errors.Wrap(errors.WithStack(err), "Failed to add users")
	}

	if err := createDirectories(runtime, host); err != nil {
		return err
	}

	if err := utils.ResetTmpDir(runtime); err != nil {
		return err
	}

	if runtime.GetSystemInfo().IsWsl() {
		if _, err := runtime.GetRunner().SudoCmd("chattr -i /etc/hosts", false, true); err != nil {
			return errors.Wrap(err, "failed to change attributes of /etc/hosts")
		}
	}

	// if running in docker container, /etc/hosts file is bind mounting, cannot be replaced via mv command
	if !n.KubeConf.Arg.IsOlaresInContainer {
		_, err1 := runtime.GetRunner().SudoCmd(fmt.Sprintf("hostnamectl set-hostname %s && sed -i '/^127.0.1.1/s/.*/127.0.1.1      %s/g' /etc/hosts", host.GetName(), host.GetName()), false, false)
		if err1 != nil {
			return errors.Wrap(errors.WithStack(err1), "Failed to override hostname")
		}
	}

	if runtime.GetSystemInfo().IsWsl() {
		if _, err := runtime.GetRunner().SudoCmd("chattr +i /etc/hosts", false, true); err != nil {
			return errors.Wrap(err, "failed to change attributes of /etc/hosts")
		}
	}

	return nil
}

type ConfigureSwapTask struct {
	common.KubeAction
}

func (t *ConfigureSwapTask) Execute(runtime connector.Runtime) error {
	if !t.KubeConf.Arg.EnableZRAM && t.KubeConf.Arg.Swappiness == 0 {
		return nil
	}
	if t.KubeConf.Arg.EnableZRAM {
		if _, err := util.GetCommand(common.CommandZRAMCtl); err != nil {
			_, err := runtime.GetRunner().SudoCmd("apt-get install -y util-linux", false, true)
			if err != nil {
				return errors.Wrap(err, "failed to install util-linux to configure zram and swap")
			}
		}

		if t.KubeConf.Arg.ZRAMSize == "" {
			t.KubeConf.Arg.ZRAMSize = strconv.Itoa(int(runtime.GetSystemInfo().GetTotalMemory() / 2))
		}
		if t.KubeConf.Arg.ZRAMSwapPriority == 0 {
			t.KubeConf.Arg.ZRAMSwapPriority = 100
		}
	}
	swapServiceStr, err := util.Render(templates.SwapServiceTmpl, t.KubeConf.Arg.SwapConfig)
	if err != nil {
		return errors.Wrap(err, "failed to generate swap configuring service")
	}

	swapServiceName := templates.SwapServiceTmpl.Name()
	swapServicePath := path.Join("/etc/systemd/system", swapServiceName)

	if err := util.WriteFile(swapServicePath, []byte(swapServiceStr), cc.FileMode0755); err != nil {
		return errors.Wrap(err, "failed to write swap configuring service file")
	}
	if _, err := runtime.GetRunner().SudoCmd("systemctl daemon-reload", false, true); err != nil {
		return errors.Wrap(err, "failed to reload swap configuring service")
	}
	if _, err := runtime.GetRunner().SudoCmd(fmt.Sprintf("systemctl enable %s", swapServiceName), false, true); err != nil {
		return errors.Wrap(err, "failed to enable swap configuring service")
	}

	// the service type is oneshot, issue restart to make it actually execute
	if _, err := runtime.GetRunner().SudoCmd(fmt.Sprintf("systemctl restart %s", swapServiceName), false, true); err != nil {
		return errors.Wrap(err, "failed to start swap configuring service")
	}
	return nil
}

func addUsers(runtime connector.Runtime, node connector.Host) error {
	if _, err := runtime.GetRunner().SudoCmd("useradd -M -c 'Kubernetes user' -s /sbin/nologin -r kube || :", false, false); err != nil {
		return err
	}

	if node.IsRole(common.ETCD) {
		if _, err := runtime.GetRunner().SudoCmd("useradd -M -c 'Etcd user' -s /sbin/nologin -r etcd || :", false, false); err != nil {
			return err
		}
	}

	return nil
}

func createDirectories(runtime connector.Runtime, node connector.Host) error {
	dirs := []string{
		common.BinDir,
		common.KubeConfigDir,
		common.KubeCertDir,
		common.KubeManifestDir,
		common.KubeScriptDir,
		common.KubeletFlexvolumesPluginsDir,
	}

	for _, dir := range dirs {
		if _, err := runtime.GetRunner().SudoCmd(fmt.Sprintf("mkdir -p %s", dir), false, false); err != nil {
			return err
		}
		if dir == common.KubeletFlexvolumesPluginsDir {
			if _, err := runtime.GetRunner().SudoCmd(fmt.Sprintf("chown kube -R %s", "/usr/libexec/kubernetes"), false, false); err != nil {
				return err
			}
		} else {
			if _, err := runtime.GetRunner().SudoCmd(fmt.Sprintf("chown kube -R %s", dir), false, false); err != nil {
				return err
			}
		}
	}

	if _, err := runtime.GetRunner().SudoCmd(fmt.Sprintf("mkdir -p %s && chown kube -R %s", "/etc/cni/net.d", "/etc/cni"), false, false); err != nil {
		return err
	}

	if _, err := runtime.GetRunner().SudoCmd(fmt.Sprintf("mkdir -p %s && chown kube -R %s", "/opt/cni/bin", "/opt/cni"), false, false); err != nil {
		return err
	}

	if _, err := runtime.GetRunner().SudoCmd(fmt.Sprintf("mkdir -p %s && chown kube -R %s", "/var/lib/calico", "/var/lib/calico"), false, false); err != nil {
		return err
	}

	if node.IsRole(common.ETCD) {
		if _, err := runtime.GetRunner().SudoCmd(fmt.Sprintf("mkdir -p %s && chown etcd -R %s", "/var/lib/etcd", "/var/lib/etcd"), false, false); err != nil {
			return err
		}
	}

	return nil
}

type NodeExecScript struct {
	common.KubeAction
}

func (n *NodeExecScript) Execute(runtime connector.Runtime) error {
	if _, err := runtime.GetRunner().SudoCmd(fmt.Sprintf("chmod +x %s/initOS.sh", common.KubeScriptDir), false, false); err != nil {
		return errors.Wrap(errors.WithStack(err), "Failed to chmod +x init os script")
	}

	if _, err := runtime.GetRunner().SudoCmd(fmt.Sprintf("%s/initOS.sh", common.KubeScriptDir), false, true); err != nil {
		return errors.Wrap(errors.WithStack(err), "Failed to configure operating system")
	}
	return nil
}

var (
	etcdFiles = []string{
		"/usr/local/bin/etcd",
		"/etc/ssl/etcd",
		"/var/lib/etcd",
		"/etc/etcd.env",
	}
	clusterFiles = []string{
		"/etc/kubernetes",
		"/etc/systemd/system/backup-etcd.timer",
		"/etc/systemd/system/backup-etcd.service",
		"/etc/systemd/system/etcd.service",
		"/var/log/calico",
		"/etc/cni",
		"/var/log/pods/",
		"/var/lib/cni",
		"/var/lib/calico",
		"/var/lib/kubelet",
		"/run/calico",
		"/run/flannel",
		"/etc/flannel",
		"/var/openebs",
		"/etc/systemd/system/kubelet.service",
		"/etc/systemd/system/kubelet.service.d",
		"/usr/local/bin/kubelet",
		"/usr/local/bin/kubeadm",
		"/usr/bin/kubelet",
		"/var/lib/rook",
		"/tmp/kubekey",
		"/etc/kubekey",
		"/etc/kke/version",
		"/etc/systemd/system/olares-swap.service",
	}

	networkResetCmds = []string{
		"ip netns show 2>/dev/null | grep cni- | xargs -r -t -n 1 ip netns delete",
		"ipvsadm -C",
		"ip link del kube-ipvs0",
		"rm -rf /var/lib/cni",
		"iptables-save | grep -v KUBE- | grep -v CALICO- | iptables-restore",
		"ip6tables-save | grep -v KUBE- | grep -v CALICO- | ip6tables-restore",
		"ipset x",
	}
)

type ResetNetworkConfig struct {
	common.KubeAction
}

func (r *ResetNetworkConfig) Execute(runtime connector.Runtime) error {
	for _, cmd := range networkResetCmds {
		_, _ = runtime.GetRunner().SudoCmd(cmd, false, false)
	}
	return nil
}

type UninstallETCD struct {
	common.KubeAction
}

func (s *UninstallETCD) Execute(runtime connector.Runtime) error {
	_, _ = runtime.GetRunner().SudoCmd("systemctl stop etcd && exit 0", false, false)
	for _, file := range etcdFiles {
		_, _ = runtime.GetRunner().SudoCmd(fmt.Sprintf("rm -rf %s", file), false, false)
	}
	return nil
}

type RemoveNodeFiles struct {
	common.KubeAction
}

func (r *RemoveNodeFiles) Execute(runtime connector.Runtime) error {
	nodeFiles := []string{
		"/etc/kubernetes",
		"/etc/systemd/system/etcd.service",
		"/var/log/calico",
		"/etc/cni",
		"/var/log/pods/",
		"/var/lib/cni",
		"/var/lib/calico",
		"/var/lib/kubelet",
		"/run/calico",
		"/run/flannel",
		"/etc/flannel",
		"/etc/systemd/system/kubelet.service",
		"/etc/systemd/system/kubelet.service.d",
		"/usr/local/bin/kubelet",
		"/usr/local/bin/kubeadm",
		"/usr/bin/kubelet",
		"/tmp/kubekey",
		"/etc/kubekey",
		"/var/openebs",
	}

	for _, file := range nodeFiles {
		_, _ = runtime.GetRunner().SudoCmd(fmt.Sprintf("rm -rf %s", file), false, false)
	}
	return nil
}

type RemoveClusterFiles struct {
	common.KubeAction
}

func (r *RemoveClusterFiles) Execute(runtime connector.Runtime) error {
	masterHostConfigFile := filepath.Join(runtime.GetBaseDir(), common.MasterHostConfigFile)
	clusterFiles = append(clusterFiles, masterHostConfigFile)
	for _, file := range clusterFiles {
		_, _ = runtime.GetRunner().SudoCmd(fmt.Sprintf("rm -rf %s", file), false, false)
	}
	return nil
}

type BackupDirBase struct {
	BackupDir string
}

func (b *BackupDirBase) InitPath(runtime connector.Runtime) error {
	b.BackupDir = path.Clean(b.BackupDir)
	if b.BackupDir == "." {
		return errors.New("backup dir is empty")
	}
	if !strings.HasSuffix(b.BackupDir, runtime.GetWorkDir()) {
		logger.Warnf("backup dir does not in workdir %s, prepending the path prefix for safety", runtime.GetWorkDir())
		b.BackupDir = path.Join(runtime.GetWorkDir(), b.BackupDir)
	}
	if err := util.CreateDir(b.BackupDir); err != nil {
		return errors.Wrapf(err, "failed to create backup dir %s", b.BackupDir)
	}
	return nil
}

type BackupFilesToDir struct {
	common.KubeAction
	*BackupDirBase
	Files []string
}

func (a *BackupFilesToDir) Execute(runtime connector.Runtime) error {
	if err := a.InitPath(runtime); err != nil {
		return err
	}
	for _, file := range a.Files {
		if file == "" {
			continue
		}
		if !util.IsExist(file) {
			logger.Warnf("backup target file does not exist: %s", file)
			continue
		}
		if !filepath.IsAbs(file) {
			var err error
			file, err = filepath.Abs(file)
			if err != nil {
				return errors.Wrapf(err, "failed to get absolute path of %s", file)
			}
		}
		if err := util.CreateDir(path.Join(a.BackupDir, path.Dir(file))); err != nil {
			return errors.Wrapf(err, "failed to create backup dir %s for file %s", path.Dir(file), path.Base(file))
		}
		logger.Debugf("copying file %s to backup dir %s", file, a.BackupDir)
		if err := util.CopyFile(file, path.Join(a.BackupDir, file)); err != nil {
			return errors.Wrapf(err, "failed to copy file %s to backup dir", file)
		}
	}
	return nil
}

type ClearBackUpDir struct {
	common.KubeAction
	*BackupDirBase
}

func (a *ClearBackUpDir) Execute(runtime connector.Runtime) error {
	if err := a.InitPath(runtime); err != nil {
		return err
	}
	if err := util.RemoveDir(a.BackupDir); err != nil {
		return errors.Wrapf(err, "failed to remove backup dir %s", a.BackupDir)
	}
	return nil
}

type RestoreBackedUpFiles struct {
	common.KubeAction
	*BackupDirBase
}

func (a *RestoreBackedUpFiles) Execute(runtime connector.Runtime) error {
	if err := a.InitPath(runtime); err != nil {
		return err
	}
	return filepath.WalkDir(a.BackupDir, func(path string, d fs.DirEntry, err error) error {
		if err != nil {
			return err
		}
		if d.IsDir() {
			return nil
		}
		originalPath, err := filepath.Rel(a.BackupDir, path)
		if err != nil {
			return errors.Wrapf(err, "failed to get original path of backed up file %s", path)
		}
		if err := util.CreateDir(filepath.Dir(originalPath)); err != nil {
			return errors.Wrapf(err, "failed to create original dir of backed up file %s", originalPath)
		}
		if err := util.CopyFile(path, originalPath); err != nil {
			return errors.Wrapf(err, "failed to restore backed up file %s", originalPath)
		}
		return nil
	})
}

type DaemonReload struct {
	common.KubeAction
}

func (d *DaemonReload) Execute(runtime connector.Runtime) error {
	if _, err := runtime.GetRunner().SudoCmd("systemctl daemon-reload && exit 0", false, false); err != nil {
		return errors.Wrap(errors.WithStack(err), "systemctl daemon-reload failed")
	}

	// try to restart the cotainerd after /etc/cni has been removed
	_, _ = runtime.GetRunner().SudoCmd("systemctl restart containerd", false, false)
	return nil
}

type GetOSData struct {
	common.KubeAction
}

func (g *GetOSData) Execute(runtime connector.Runtime) error {
	osReleaseStr, err := runtime.GetRunner().SudoCmd("cat /etc/os-release", false, false)
	if err != nil {
		return err
	}

	osrData := osrelease.Parse(strings.Replace(osReleaseStr, "\r\n", "\n", -1))

	host := runtime.RemoteHost()
	// type: *osrelease.data
	host.GetCache().Set(Release, osrData)
	return nil
}

type SyncRepositoryFile struct {
	common.KubeAction
}

func (s *SyncRepositoryFile) Execute(runtime connector.Runtime) error {
	if err := utils.ResetTmpDir(runtime); err != nil {
		return errors.Wrap(err, "reset tmp dir failed")
	}

	host := runtime.RemoteHost()
	release, ok := host.GetCache().Get(Release)
	if !ok {
		return errors.New("get os release failed by root cache")
	}
	r := release.(*osrelease.Data)

	fileName := fmt.Sprintf("%s-%s-%s.iso", r.ID, r.VersionID, host.GetArch())
	src := filepath.Join(runtime.GetWorkDir(), "repository", host.GetArch(), r.ID, r.VersionID, fileName)
	dst := filepath.Join(common.TmpDir, fileName)
	if err := runtime.GetRunner().Scp(src, dst); err != nil {
		return errors.Wrapf(errors.WithStack(err), "scp %s to %s failed", src, dst)
	}

	host.GetCache().Set("iso", fileName)
	return nil
}

type MountISO struct {
	common.KubeAction
}

func (m *MountISO) Execute(runtime connector.Runtime) error {
	mountPath := filepath.Join(common.TmpDir, "iso")
	if err := runtime.GetRunner().MkDir(mountPath); err != nil {
		return errors.Wrapf(errors.WithStack(err), "create mount dir failed")
	}

	host := runtime.RemoteHost()
	isoFile, _ := host.GetCache().GetMustString("iso")
	path := filepath.Join(common.TmpDir, isoFile)
	mountCmd := fmt.Sprintf("sudo mount -t iso9660 -o loop %s %s", path, mountPath)
	if _, err := runtime.GetRunner().Cmd(mountCmd, false, false); err != nil {
		return errors.Wrapf(errors.WithStack(err), "mount %s at %s failed", path, mountPath)
	}
	return nil
}

type NewRepoClient struct {
	common.KubeAction
}

func (n *NewRepoClient) Execute(runtime connector.Runtime) error {
	host := runtime.RemoteHost()
	release, ok := host.GetCache().Get(Release)
	if !ok {
		return errors.New("get os release failed by host cache")
	}
	r := release.(*osrelease.Data)

	repo, err := repository.New(r.ID)
	if err != nil {
		checkDeb, debErr := runtime.GetRunner().SudoCmd("which apt", false, false)
		if debErr == nil && strings.Contains(checkDeb, "bin") {
			repo = repository.NewDeb()
		}
		checkRPM, rpmErr := runtime.GetRunner().SudoCmd("which yum", false, false)
		if rpmErr == nil && strings.Contains(checkRPM, "bin") {
			repo = repository.NewRPM()
		}

		if debErr != nil && rpmErr != nil {
			return errors.Wrap(errors.WithStack(err), "new repository manager failed")
		} else if debErr == nil && rpmErr == nil {
			return errors.New("can't detect the main package repository, only one of apt or yum is supported")
		}
	}

	host.GetCache().Set("repo", repo)
	return nil
}

type BackupOriginalRepository struct {
	common.KubeAction
}

func (b *BackupOriginalRepository) Execute(runtime connector.Runtime) error {
	host := runtime.RemoteHost()
	r, ok := host.GetCache().Get("repo")
	if !ok {
		return errors.New("get repo failed by host cache")
	}
	repo := r.(repository.Interface)

	if err := repo.Backup(runtime); err != nil {
		return errors.Wrap(errors.WithStack(err), "backup repository failed")
	}

	return nil
}

type AddLocalRepository struct {
	common.KubeAction
}

func (a *AddLocalRepository) Execute(runtime connector.Runtime) error {
	host := runtime.RemoteHost()
	r, ok := host.GetCache().Get("repo")
	if !ok {
		return errors.New("get repo failed by host cache")
	}
	repo := r.(repository.Interface)

	if installErr := repo.Add(runtime, filepath.Join(common.TmpDir, "iso")); installErr != nil {
		return errors.Wrap(errors.WithStack(installErr), "add local repository failed")
	}
	if installErr := repo.Update(runtime); installErr != nil {
		return errors.Wrap(errors.WithStack(installErr), "update local repository failed")
	}

	return nil
}

type InstallPackage struct {
	common.KubeAction
}

func (i *InstallPackage) Execute(runtime connector.Runtime) error {
	host := runtime.RemoteHost()
	repo, ok := host.GetCache().Get("repo")
	if !ok {
		return errors.New("get repo failed by host cache")
	}
	r := repo.(repository.Interface)

	var pkg []string
	if _, ok := r.(*repository.Debian); ok {
		pkg = i.KubeConf.Cluster.System.Debs
	} else if _, ok := r.(*repository.RedhatPackageManager); ok {
		pkg = i.KubeConf.Cluster.System.Rpms
	}

	if installErr := r.Install(runtime, pkg...); installErr != nil {
		return errors.Wrap(errors.WithStack(installErr), "install repository package failed")
	}
	return nil
}

type ResetRepository struct {
	common.KubeAction
}

func (r *ResetRepository) Execute(runtime connector.Runtime) error {
	host := runtime.RemoteHost()
	repo, ok := host.GetCache().Get("repo")
	if !ok {
		return errors.New("get repo failed by host cache")
	}
	re := repo.(repository.Interface)

	var resetErr error
	defer func() {
		if resetErr != nil {
			mountPath := filepath.Join(common.TmpDir, "iso")
			umountCmd := fmt.Sprintf("umount %s", mountPath)
			_, _ = runtime.GetRunner().SudoCmd(umountCmd, false, false)
		}
	}()

	if resetErr = re.Reset(runtime); resetErr != nil {
		return errors.Wrap(errors.WithStack(resetErr), "reset repository failed")
	}

	return nil
}

type UmountISO struct {
	common.KubeAction
}

func (u *UmountISO) Execute(runtime connector.Runtime) error {
	mountPath := filepath.Join(common.TmpDir, "iso")
	umountCmd := fmt.Sprintf("umount %s", mountPath)
	if _, err := runtime.GetRunner().SudoCmd(umountCmd, false, false); err != nil {
		return errors.Wrapf(errors.WithStack(err), "umount %s failed", mountPath)
	}
	return nil
}

type NodeConfigureNtpServer struct {
	common.KubeAction
}

func (n *NodeConfigureNtpServer) Execute(runtime connector.Runtime) error {
	currentHost := runtime.RemoteHost()
	release, ok := currentHost.GetCache().Get(Release)
	if !ok {
		return errors.New("get os release failed by host cache")
	}
	r := release.(*osrelease.Data)

	chronyConfigFile := "/etc/chrony.conf"
	chronyService := "chronyd.service"
	if r.ID == "ubuntu" || r.ID == "debian" {
		chronyConfigFile = "/etc/chrony/chrony.conf"
		chronyService = "chrony.service"
	}

	// if NtpServers was configured
	for _, server := range n.KubeConf.Cluster.System.NtpServers {

		serverAddr := strings.Trim(server, " \"")
		if serverAddr == currentHost.GetName() || serverAddr == currentHost.GetInternalAddress() {
			allowClientCmd := fmt.Sprintf(`sed -i '/#allow/ a\allow 0.0.0.0/0' %s`, chronyConfigFile)
			if _, err := runtime.GetRunner().SudoCmd(allowClientCmd, false, false); err != nil {
				return errors.Wrapf(err, "change host:%s chronyd conf failed, please check file %s", serverAddr, chronyConfigFile)
			}
		}

		// use internal ip to client chronyd server
		for _, host := range runtime.GetAllHosts() {
			if serverAddr == host.GetName() {
				serverAddr = host.GetInternalAddress()
				break
			}
		}

		checkOrAddCmd := fmt.Sprintf(`grep -q '^server %s iburst' %s||sed '1a server %s iburst' -i %s`, serverAddr, chronyConfigFile, serverAddr, chronyConfigFile)
		if _, err := runtime.GetRunner().SudoCmd(checkOrAddCmd, false, false); err != nil {
			return errors.Wrapf(err, "set ntpserver: %s failed, please check file %s", serverAddr, chronyConfigFile)
		}

	}

	// if Timezone was configured
	if len(n.KubeConf.Cluster.System.Timezone) > 0 {
		setTimeZoneCmd := fmt.Sprintf("timedatectl set-timezone %s", n.KubeConf.Cluster.System.Timezone)
		if _, err := runtime.GetRunner().SudoCmd(setTimeZoneCmd, false, false); err != nil {
			return errors.Wrapf(err, "set timezone: %s failed", n.KubeConf.Cluster.System.Timezone)
		}

		if _, err := runtime.GetRunner().SudoCmd("timedatectl set-ntp true", false, false); err != nil {
			return errors.Wrap(err, "timedatectl set-ntp true failed")
		}
	}

	// ensure chronyd was enabled and work normally
	if len(n.KubeConf.Cluster.System.NtpServers) > 0 || len(n.KubeConf.Cluster.System.Timezone) > 0 {
		startChronyCmd := fmt.Sprintf("systemctl enable %s && systemctl restart %s", chronyService, chronyService)
		if _, err := runtime.GetRunner().SudoCmd(startChronyCmd, false, false); err != nil {
			return errors.Wrap(err, "restart chronyd failed")
		}

		// tells chronyd to cancel any remaining correction that was being slewed and jump the system clock by the equivalent amount, making it correct immediately.
		if _, err := runtime.GetRunner().SudoCmd("chronyc makestep > /dev/null && chronyc sources", false, true); err != nil {
			return errors.Wrap(err, "chronyc makestep failed")
		}
	}

	return nil
}
