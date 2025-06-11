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

package container

import (
	"bufio"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/beclab/Olares/cli/pkg/common"
	"github.com/beclab/Olares/cli/pkg/container/templates"
	cc "github.com/beclab/Olares/cli/pkg/core/common"
	"github.com/beclab/Olares/cli/pkg/core/connector"
	"github.com/beclab/Olares/cli/pkg/core/logger"
	"github.com/beclab/Olares/cli/pkg/core/util"
	"github.com/beclab/Olares/cli/pkg/manifest"
	"github.com/beclab/Olares/cli/pkg/utils"
	"github.com/pkg/errors"
)

type CreateZfsMount struct {
	common.KubeAction
}

func (t *CreateZfsMount) Execute(runtime connector.Runtime) error {
	systemInfo := runtime.GetSystemInfo()
	if systemInfo.GetFsType() != "zfs" {
		return nil
	}
	var cmd = fmt.Sprintf("zfs create -o mountpoint=%s %s/containerd", cc.ZfsSnapshotter, systemInfo.GetDefaultZfsPrefixName())
	if _, err := runtime.GetRunner().SudoCmd(cmd, false, true); err != nil {
		if strings.Contains(err.Error(), "already exists") {
			logger.Debugf("zfs %s/containerd already exists", systemInfo.GetDefaultZfsPrefixName())
			return nil
		}
		logger.Errorf("create zfs mount error %v", err)
	}
	return nil
}

type ZfsReset struct {
	common.KubeAction
}

func (t *ZfsReset) Execute(runtime connector.Runtime) error {
	if _, err := util.GetCommand("zfs"); err != nil {
		return err
	}
	var systemInfo = runtime.GetSystemInfo()
	res, _ := runtime.GetRunner().SudoCmd("zfs list -t all", false, false)
	if res != "" {
		scanner := bufio.NewScanner(strings.NewReader(res))
		for scanner.Scan() {
			line := scanner.Text()
			fields := strings.Fields(line)
			if len(fields) < 5 {
				continue
			}

			var name = fields[0]

			if !strings.Contains(name, fmt.Sprintf("%s/containerd", systemInfo.GetDefaultZfsPrefixName())) {
				continue
			}
			var mp = fields[4]
			if !strings.Contains(mp, "legacy") {
				continue
			}

			if _, err := runtime.GetRunner().SudoCmd(fmt.Sprintf("zfs destroy %s -frR", name), false, false); err == nil {
				fmt.Printf("delete zfs device %s\n", name)
			}
		}
	}

	runtime.GetRunner().SudoCmd(fmt.Sprintf("zfs destroy %s/containerd -frR", systemInfo.GetDefaultZfsPrefixName()), false, false)

	return nil
}

type SyncContainerd struct {
	common.KubeAction
	manifest.ManifestAction
}

func (s *SyncContainerd) Execute(runtime connector.Runtime) error {
	if err := utils.ResetTmpDir(runtime); err != nil {
		return err
	}

	containerd, err := s.Manifest.Get("containerd")
	if err != nil {
		return err
	}

	path := containerd.FilePath(s.BaseDir)

	dst := filepath.Join(common.TmpDir, containerd.Filename)
	if err := runtime.GetRunner().Scp(path, dst); err != nil {
		return errors.Wrap(errors.WithStack(err), fmt.Sprintf("sync containerd binaries failed"))
	}

	if _, err := runtime.GetRunner().SudoCmd(
		fmt.Sprintf("mkdir -p /usr/bin && tar -zxf %s --strip-components=1 -C /usr/bin", dst),
		false, false); err != nil {
		return errors.Wrap(errors.WithStack(err), fmt.Sprintf("install containerd binaries failed"))
	}
	return nil
}

type SyncCrictlBinaries struct {
	common.KubeAction
	manifest.ManifestAction
}

func (s *SyncCrictlBinaries) Execute(runtime connector.Runtime) error {
	if err := utils.ResetTmpDir(runtime); err != nil {
		return err
	}

	crictl, err := s.Manifest.Get("crictl")
	if err != nil {
		return err
	}

	path := crictl.FilePath(s.BaseDir)

	dst := filepath.Join(common.TmpDir, crictl.Filename)

	if err := runtime.GetRunner().Scp(path, dst); err != nil {
		return errors.Wrap(errors.WithStack(err), fmt.Sprintf("sync crictl binaries failed"))
	}

	if _, err := runtime.GetRunner().SudoCmd(
		fmt.Sprintf("mkdir -p /usr/bin && tar -zxf %s -C /usr/bin ", dst),
		false, false); err != nil {
		return errors.Wrap(errors.WithStack(err), fmt.Sprintf("install crictl binaries failed"))
	}
	return nil
}

type EnableContainerd struct {
	common.KubeAction
	manifest.ManifestAction
}

func (e *EnableContainerd) Execute(runtime connector.Runtime) error {
	if _, err := runtime.GetRunner().SudoCmd(
		"systemctl daemon-reload && systemctl enable containerd && systemctl start containerd",
		false, false); err != nil {
		return errors.Wrap(errors.WithStack(err), fmt.Sprintf("enable and start containerd failed"))
	}

	// install runc
	if err := utils.ResetTmpDir(runtime); err != nil {
		return err
	}

	runcKey := common.Runc
	containerd, err := e.Manifest.Get(runcKey)
	if err != nil {
		return errors.New("get KubeBinary key runc by manifest error")
	}

	path := containerd.FilePath(e.BaseDir)

	dst := filepath.Join(common.TmpDir, containerd.Filename)
	if err := runtime.GetRunner().Scp(path, dst); err != nil {
		return errors.Wrap(errors.WithStack(err), fmt.Sprintf("sync runc binaries failed"))
	}

	if _, err := runtime.GetRunner().SudoCmd(
		fmt.Sprintf("install -m 755 %s /usr/local/sbin/runc", dst),
		false, false); err != nil {
		return errors.Wrap(errors.WithStack(err), fmt.Sprintf("install runc binaries failed"))
	}
	return nil
}

type DisableContainerd struct {
	common.KubeAction
}

func (d *DisableContainerd) Execute(runtime connector.Runtime) error {
	if stdout, err := runtime.GetRunner().SudoCmd("systemctl status containerd", false, false); err != nil {
		if !strings.Contains(stdout, "could not be found") {
			return err
		}
	} else {
		_, _ = runtime.GetRunner().SudoCmd("systemctl disable containerd && systemctl stop containerd", false, false)
	}

	if err := umountPoints(runtime); err != nil {
		return err
	}

	// remove containerd related files
	files := []string{
		"/usr/local/sbin/runc",
		"/usr/bin/crictl",
		"/usr/bin/containerd*",
		"/usr/bin/ctr",
		"/usr/local/bin/crictl",      // cloud version
		"/usr/local/bin/containerd*", // cloud version
		"/usr/local/bin/ctr",         // cloud version
		"/etc/systemd/system/containerd.service",
		"/lib/systemd/system/containerd.service", // apt installed
		"/run/containerd",                        //
		filepath.Join("/etc/systemd/system", templates.ContainerdService.Name()),
		filepath.Join("/etc/containerd", templates.ContainerdConfig.Name()),
		filepath.Join("/etc", templates.CrictlConfig.Name()),
	}
	if d.KubeConf.Cluster.Registry.DataRoot != "" {
		files = append(files, d.KubeConf.Cluster.Registry.DataRoot)
	} else {
		files = append(files, "/var/lib/containerd")
	}

	for _, file := range files {
		_, _ = runtime.GetRunner().SudoCmd(fmt.Sprintf("rm -rf %s", file), false, true)
	}
	return nil
}

func getProcessIds(pid string, runtime connector.Runtime) []string {
	var c []string
	var childs = getChildPids(pid, runtime)
	if childs != nil && len(childs) > 0 {
		for _, child := range childs {
			t := getProcessIds(child, runtime)
			if t == nil || len(t) == 0 {
				continue
			}
			c = append(c, t...)
		}
		c = append(c, childs...)
		return c
	}
	return nil
}

func getChildPids(pid string, runtime connector.Runtime) []string {
	var childs []string
	var cmd = fmt.Sprintf("pgrep -P %s", pid)
	chpids, err := runtime.GetRunner().SudoCmd(cmd, false, false)
	if err == nil && chpids != "" {
		scanner := bufio.NewScanner(strings.NewReader(chpids))
		for scanner.Scan() {
			line := scanner.Text()
			if line != "" {
				childs = append(childs, line)
			}
		}
	}
	return childs
}

func umountPoints(runtime connector.Runtime) error {
	file, err := os.Open("/proc/mounts")
	if err != nil {
		return fmt.Errorf("failed to open /proc/mounts: %w", err)
	}
	defer file.Close()

	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		line := scanner.Text()
		if !strings.Contains(line, "containerd") {
			continue
		}
		fields := strings.Fields(line)
		if len(fields) >= 2 {
			p := fields[1]
			if util.IsExist(p) {
				runtime.GetRunner().SudoCmd(fmt.Sprintf("umount %s", p), false, true)
			}
		}
	}

	if err := scanner.Err(); err != nil {
		return fmt.Errorf("error reading /proc/mounts: %w", err)
	}

	return nil
}

type KillContainerdProcess struct {
	common.KubeAction
	Signal        string
	Timeout       time.Duration
	CheckInterval time.Duration
}

// getContainerdPids returns all containerd-shim process IDs and their child processes
func getContainerdPids(runtime connector.Runtime) ([]string, error) {
	var pids []string
	var childpids []string
	var cmd = "ps -ef | grep containerd-shim"
	stdout, err := runtime.GetRunner().SudoCmd(cmd, false, false)
	if err != nil {
		return nil, err
	}

	scanner := bufio.NewScanner(strings.NewReader(stdout))
	for scanner.Scan() {
		line := scanner.Text()
		if strings.Contains(line, "grep") {
			continue
		}
		fields := strings.Fields(line)
		if len(fields) > 1 {
			pid := fields[1]
			pids = append(pids, pid)
		}
	}

	if len(pids) > 0 {
		for _, pid := range pids {
			var p = getProcessIds(pid, runtime)
			childpids = append(childpids, p...)
		}
	}

	var allPids []string
	allPids = append(allPids, childpids...)
	allPids = append(allPids, pids...)

	return allPids, nil
}

func (t *KillContainerdProcess) Execute(runtime connector.Runtime) error {
	if t.Signal == "" {
		t.Signal = "TERM"
	}
	if t.Timeout == 0 {
		t.Timeout = 1 * time.Minute
	}
	if t.CheckInterval == 0 {
		t.CheckInterval = 10 * time.Second
	}

	allPids, err := getContainerdPids(runtime)
	if err != nil {
		return errors.Wrap(err, "get container pids failed")
	}

	if len(allPids) == 0 {
		return nil
	}

	// first try with the specified signal
	for _, pid := range allPids {
		runtime.GetRunner().SudoCmd(fmt.Sprintf("kill -%s %s", t.Signal, pid), false, false)
	}

	// if signal is KILL, just return immediately
	// otherwise, poll until timeout to check if processes are gone
	if t.Signal == "KILL" || t.Signal == "9" {
		return nil
	}
	deadline := time.Now().Add(t.Timeout)

	for time.Now().Before(deadline) {
		remainingPids, err := getContainerdPids(runtime)
		if err != nil {
			continue
		}

		// If no processes remain, we're done
		if len(remainingPids) == 0 {
			return nil
		}

		// Wait for the check interval before next poll
		time.Sleep(t.CheckInterval)
	}

	// force kill remaining processes
	remainingPids, err := getContainerdPids(runtime)
	if err != nil {
		return err
	}

	for _, pid := range remainingPids {
		runtime.GetRunner().SudoCmd(fmt.Sprintf("kill -9 %s", pid), false, false)
	}

	return nil
}
