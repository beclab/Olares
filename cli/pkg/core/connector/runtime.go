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

package connector

import (
	"fmt"
	"os"
	"os/user"
	"path"
	"path/filepath"

	"github.com/pkg/errors"
	"k8s.io/client-go/kubernetes"

	"bytetrade.io/web3os/installer/pkg/core/common"
	"bytetrade.io/web3os/installer/pkg/core/logger"
	"bytetrade.io/web3os/installer/pkg/core/storage"
	"bytetrade.io/web3os/installer/pkg/core/util"
)

type BaseRuntime struct {
	ObjName         string
	connector       Connector
	runner          *Runner
	storage         storage.Provider
	homeDir         string
	baseDir         string
	installerDir    string
	workDir         string
	verbose         bool
	ignoreErr       bool
	allHosts        []Host
	roleHosts       map[string][]Host
	deprecatedHosts map[string]string
	cmdSed          string
	olaresVersion   string
	systemInfo      Systems
	k8sClient       *kubernetes.Clientset
}

func NewBaseRuntime(name string, connector Connector, verbose bool, ignoreErr bool, sqlProvider storage.Provider, baseDir string, olaresVersion string, consoleLogFileName string, consoleLogTruncate bool, systemInfo Systems) BaseRuntime {
	base := BaseRuntime{
		ObjName:         name,
		connector:       connector,
		storage:         sqlProvider,
		verbose:         verbose,
		ignoreErr:       ignoreErr,
		allHosts:        make([]Host, 0, 0),
		roleHosts:       make(map[string][]Host),
		deprecatedHosts: make(map[string]string),
		cmdSed:          util.FormatSed(systemInfo.IsDarwin()),
		systemInfo:      systemInfo,
		olaresVersion:   olaresVersion,
	}

	if systemInfo.IsWindows() {
		baseDir = fmt.Sprintf("%s\\%s", systemInfo.GetHomeDir(), common.DefaultBaseDir)
	}
	if err := base.GenerateBaseDir(baseDir); err != nil {
		fmt.Printf("[ERRO]: Failed to create base dir: %s\n", err)
		os.Exit(1)
	}
	if err := base.GenerateWorkDir(); err != nil {
		fmt.Printf("[ERRO]: Failed to create work dir: %s\n", err)
		os.Exit(1)
	}
	if err := base.InitLogger(consoleLogFileName, consoleLogTruncate); err != nil {
		fmt.Printf("[ERRO]: Failed to init log entry: %s\n", err)
		os.Exit(1)
	}

	return base
}

func (b *BaseRuntime) GetSystemInfo() Systems {
	return b.systemInfo
}

func (b *BaseRuntime) GetObjName() string {
	return b.ObjName
}

func (b *BaseRuntime) SetObjName(name string) {
	b.ObjName = name
}

func (b *BaseRuntime) GetRunner() *Runner {
	return b.runner
}

func (b *BaseRuntime) SetRunner(r *Runner) {
	b.runner = r
}

func (b *BaseRuntime) GetStorage() storage.Provider {
	return b.storage
}

func (b *BaseRuntime) SetStorage(s storage.Provider) {
	b.storage = s
}

func (b *BaseRuntime) GetConnector() Connector {
	return b.connector
}

func (b *BaseRuntime) SetConnector(c Connector) {
	b.connector = c
}

func (b *BaseRuntime) GenerateBaseDir(baseDir string) error {
	usr, err := user.Current()
	if err != nil {
		return errors.Wrap(err, "get current user failed")
	}
	homeDir := usr.HomeDir
	b.homeDir = homeDir

	b.baseDir = baseDir
	if baseDir == "" {
		b.baseDir = path.Join(homeDir, common.DefaultBaseDir)
	}

	return nil
}

func (b *BaseRuntime) GenerateWorkDir() error {
	installerPath := filepath.Join(b.baseDir, "versions", fmt.Sprintf("v%s", b.olaresVersion))
	if err := util.CreateDir(installerPath); err != nil {
		return errors.Wrap(err, "create wizard dir failed")
	}
	b.installerDir = installerPath

	rootPath := filepath.Join(installerPath, common.Cli)
	if err := util.CreateDir(rootPath); err != nil {
		return errors.Wrap(err, "create work dir failed")
	}
	b.workDir = rootPath

	for i := range b.allHosts {
		subPath := filepath.Join(rootPath, b.allHosts[i].GetName())
		if err := util.CreateDir(subPath); err != nil {
			return errors.Wrap(err, "create work dir failed")
		}
	}
	return nil
}

func (b *BaseRuntime) GetHostWorkDir() string {
	return filepath.Join(b.workDir, b.RemoteHost().GetName())
}

func (b *BaseRuntime) GetHomeDir() string {
	return b.homeDir
}

func (b *BaseRuntime) GetBaseDir() string {
	return b.baseDir
}

func (b *BaseRuntime) GetInstallerDir() string {
	return b.installerDir
}

func (b *BaseRuntime) GetWorkDir() string {
	return b.workDir
}

func (b *BaseRuntime) GetIgnoreErr() bool {
	return b.ignoreErr
}

func (b *BaseRuntime) GetAllHosts() []Host {
	hosts := make([]Host, 0, 0)
	for i := range b.allHosts {
		if b.allHosts[i] == nil || b.HostIsDeprecated(b.allHosts[i]) {
			continue
		}
		hosts = append(hosts, b.allHosts[i])
	}
	return hosts
}

func (b *BaseRuntime) SetAllHosts(hosts []Host) {
	b.allHosts = hosts
}

func (b *BaseRuntime) GetHostsByRole(role string) []Host {
	if _, ok := b.roleHosts[role]; ok {
		return b.roleHosts[role]
	} else {
		return []Host{}
	}
}

func (b *BaseRuntime) GetLocalHost() Host {
	si := b.GetSystemInfo()
	for i := range b.allHosts {
		if b.allHosts[i].GetName() == si.GetHostname() {
			return b.allHosts[i]
		}
	}
	return &BaseHost{
		Name:            common.LocalHost,
		User:            si.GetUsername(),
		Address:         si.GetLocalIp(),
		InternalAddress: si.GetLocalIp(),
		Arch:            si.GetOsArch(),
		Os:              si.GetOsType(),
	}
}

func (b *BaseRuntime) RemoteHost() Host {
	return b.GetRunner().Host
}

func (b *BaseRuntime) DeleteHost(host Host) {
	i := 0
	for j := range b.allHosts {
		if b.allHosts[j].GetName() != host.GetName() {
			b.allHosts[i] = b.allHosts[j]
			i++
		}
	}
	b.allHosts[i] = nil
	b.allHosts = b.allHosts[:i]
	b.RoleMapDelete(host)
	b.deprecatedHosts[host.GetName()] = ""
}

func (b *BaseRuntime) HostIsDeprecated(host Host) bool {
	if _, ok := b.deprecatedHosts[host.GetName()]; ok {
		return true
	}
	return false
}

func (b *BaseRuntime) InitLogger(consoleLogFileName string, consoleLogTruncate bool) error {
	if consoleLogFileName == "" {
		consoleLogFileName = common.InstallLogFile
	}
	// the JSON-structured logs under .terminus/logs/yyyy-mm-dd_hh-mm-ss.log
	// and the console formatted logs under .terminus/versions/v{version}/install.log (for backward compatibility)
	logger.InitLog(path.Join(b.baseDir, common.LogsDir), path.Join(b.installerDir, common.LogsDir, consoleLogFileName), consoleLogTruncate)
	return nil
}

func (b *BaseRuntime) GetCommandSed() string {
	return b.cmdSed
}

func (b *BaseRuntime) Copy() Runtime {
	runtime := *b
	return &runtime
}

func (b *BaseRuntime) GenerateRoleMap() {
	for i := range b.allHosts {
		b.AppendRoleMap(b.allHosts[i])
	}
}

func (b *BaseRuntime) AppendHost(host Host) {
	b.allHosts = append(b.allHosts, host)
}

func (b *BaseRuntime) AppendRoleMap(host Host) {
	for _, r := range host.GetRoles() {
		if hosts, ok := b.roleHosts[r]; ok {
			hosts = append(hosts, host)
			b.roleHosts[r] = hosts
		} else {
			first := make([]Host, 0, 0)
			first = append(first, host)
			b.roleHosts[r] = first
		}
	}
}

func (b *BaseRuntime) RoleMapDelete(host Host) {
	for role, hosts := range b.roleHosts {
		i := 0
		for j := range hosts {
			if hosts[j].GetName() != host.GetName() {
				hosts[i] = hosts[j]
				i++
			}
		}
		hosts = hosts[:i]
		b.roleHosts[role] = hosts
	}
}
