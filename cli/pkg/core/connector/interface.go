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
	"context"
	"io"
	"os"

	"bytetrade.io/web3os/installer/pkg/core/cache"
	"bytetrade.io/web3os/installer/pkg/core/storage"
)

type Connection interface {
	Exec(cmd string, host Host) (stdout string, code int, err error)
	PExec(cmd string, stdin io.Reader, stdout io.Writer, stderr io.Writer, host Host) (code int, err error)
	Fetch(local, remote string, host Host) error
	Scp(local, remote string, host Host) error
	RemoteFileExist(remote string, host Host) bool
	RemoteDirExist(remote string, host Host) (bool, error)
	MkDirAll(path string, mode string, host Host) error
	Chmod(path string, mode os.FileMode) error
	Close()
}

type Connector interface {
	Connect(host Host) (Connection, error)
	Close(host Host)
}

type ModuleRuntime interface {
	GetObjName() string
	SetObjName(name string)
	GenerateBaseDir(baseDir string) error
	GenerateWorkDir() error
	GetHostWorkDir() string
	GetHomeDir() string
	GetBaseDir() string
	GetInstallerDir() string
	GetWorkDir() string
	GetIgnoreErr() bool
	GetAllHosts() []Host
	SetAllHosts([]Host)
	GetHostsByRole(role string) []Host
	GetLocalHost() Host
	DeleteHost(host Host)
	HostIsDeprecated(host Host) bool
	// InitLogger() error
	GetCommandSed() string
}

type Runtime interface {
	GetRunner() *Runner
	SetRunner(r *Runner)
	GetConnector() Connector
	SetConnector(c Connector)
	SetStorage(s storage.Provider)
	GetStorage() storage.Provider
	RemoteHost() Host
	Copy() Runtime
	GetSystemInfo() Systems
	ModuleRuntime
}

type Host interface {
	GetName() string
	SetName(name string)
	GetAddress() string
	SetAddress(str string)
	GetInternalAddress() string
	SetInternalAddress(str string)
	GetPort() int
	SetPort(port int)
	GetUser() string
	SetUser(u string)
	GetPassword() string
	SetPassword(password string)
	GetPrivateKey() string
	SetPrivateKey(privateKey string)
	GetPrivateKeyPath() string
	SetPrivateKeyPath(path string)
	GetArch() string
	SetArch(arch string)
	GetOs() string
	SetOs(osType string)
	SetMinikubeProfile(profile string)
	GetMinikubeProfile() string
	GetTimeout() int64
	SetTimeout(timeout int64)
	GetRoles() []string
	SetRoles(roles []string)
	IsRole(role string) bool
	GetCache() *cache.Cache
	SetCache(c *cache.Cache)

	Exec(ctx context.Context, cmd string, printOutput bool, printLine bool) (stdout string, code int, err error)
	ExecExt(cmd string, printOutput bool, printLine bool) (stdout string, code int, err error)
	Fetch(local, remote string, printOutput bool, printLine bool) error
	SudoScp(local, remote string) error
	Scp(local, remote string) error
	FileExist(f string) (bool, error)
	DirExist(d string) (bool, error)
	MkDir(path string) error
	Cmd(cmd string, printOutput bool, printLine bool) (string, error)
	CmdExt(cmd string, printOutput bool, printLine bool) (string, error)
	SudoPrefixIfNecessary(cmd string) string
	SudoCmd(cmd string, printOutput bool, printLine bool) (string, error)
	SudoCmdContext(ctx context.Context, cmd string, printOutput bool, printLine bool) (string, error)
	CmdExtWithContext(ctx context.Context, cmd string, printOutput bool, printLine bool) (string, error)
	MkDirAll(path string, mode string) error
	Chmod(path string, mode os.FileMode) error
	FileMd5(path string) (string, error)
}
