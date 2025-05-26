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
	"encoding/base64"
	"fmt"
	"os"
	"path/filepath"

	"bytetrade.io/web3os/installer/pkg/core/cache"
	"bytetrade.io/web3os/installer/pkg/core/common"
	"bytetrade.io/web3os/installer/pkg/core/logger"
	"bytetrade.io/web3os/installer/pkg/core/util"
)

type BaseHost struct {
	Name            string `yaml:"name,omitempty" json:"name,omitempty"`
	Address         string `yaml:"address,omitempty" json:"address,omitempty"`
	InternalAddress string `yaml:"internalAddress,omitempty" json:"internalAddress,omitempty"`
	Port            int    `yaml:"port,omitempty" json:"port,omitempty"`
	User            string `yaml:"user,omitempty" json:"user,omitempty"`
	Password        string `yaml:"password,omitempty" json:"password,omitempty"`
	PrivateKey      string `yaml:"privateKey,omitempty" json:"privateKey,omitempty"`
	PrivateKeyPath  string `yaml:"privateKeyPath,omitempty" json:"privateKeyPath,omitempty"`
	Arch            string `yaml:"arch,omitempty" json:"arch,omitempty"`
	Os              string `yaml:"os,omitempty" json:"os,omitempty"`
	Timeout         int64  `yaml:"timeout,omitempty" json:"timeout,omitempty"`
	MiniKubeProfile string `json:"minikubeProfileName,omitempty" json:"minikubeProfileName,omitempty"`

	Roles     []string        `json:"-"`
	RoleTable map[string]bool `json:"-"`
	Cache     *cache.Cache    `json:"-"`
}

func NewHost() *BaseHost {
	return &BaseHost{
		Roles:     make([]string, 0, 0),
		RoleTable: make(map[string]bool),
		Cache:     cache.NewCache(),
	}
}

func (b *BaseHost) GetName() string {
	return b.Name
}

func (b *BaseHost) SetName(name string) {
	b.Name = name
}

func (b *BaseHost) GetAddress() string {
	return b.Address
}

func (b *BaseHost) SetAddress(str string) {
	b.Address = str
}

func (b *BaseHost) GetInternalAddress() string {
	return b.InternalAddress
}

func (b *BaseHost) SetInternalAddress(str string) {
	b.InternalAddress = str
}

func (b *BaseHost) GetPort() int {
	return b.Port
}

func (b *BaseHost) SetPort(port int) {
	b.Port = port
}

func (b *BaseHost) GetUser() string {
	return b.User
}

func (b *BaseHost) SetUser(u string) {
	b.User = u
}

func (b *BaseHost) GetPassword() string {
	return b.Password
}

func (b *BaseHost) SetPassword(password string) {
	b.Password = password
}

func (b *BaseHost) GetPrivateKey() string {
	return b.PrivateKey
}

func (b *BaseHost) SetPrivateKey(privateKey string) {
	b.PrivateKey = privateKey
}

func (b *BaseHost) GetPrivateKeyPath() string {
	return b.PrivateKeyPath
}

func (b *BaseHost) SetPrivateKeyPath(path string) {
	b.PrivateKeyPath = path
}

func (b *BaseHost) GetArch() string {
	return b.Arch
}

func (b *BaseHost) SetArch(arch string) {
	b.Arch = arch
}

func (b *BaseHost) GetOs() string {
	return b.Os
}

func (b *BaseHost) SetOs(osType string) {
	b.Os = osType
}

func (b *BaseHost) SetMinikubeProfile(profile string) {
	b.MiniKubeProfile = profile
}
func (b *BaseHost) GetMinikubeProfile() string {
	return b.MiniKubeProfile
}

func (b *BaseHost) GetTimeout() int64 {
	return b.Timeout
}

func (b *BaseHost) SetTimeout(timeout int64) {
	b.Timeout = timeout
}

func (b *BaseHost) GetRoles() []string {
	return b.Roles
}

func (b *BaseHost) SetRoles(roles []string) {
	b.Roles = roles
}

func (b *BaseHost) SetRole(role string) {
	b.RoleTable[role] = true
	b.Roles = append(b.Roles, role)
}

func (b *BaseHost) IsRole(role string) bool {
	if res, ok := b.RoleTable[role]; ok {
		return res
	}
	return false
}

func (b *BaseHost) GetCache() *cache.Cache {
	return b.Cache
}

func (b *BaseHost) SetCache(c *cache.Cache) {
	b.Cache = c
}

func (b *BaseHost) Exec(ctx context.Context, cmd string, printOutput bool, printLine bool) (stdout string, code int, err error) {
	return util.Exec(ctx, cmd, printOutput, printLine)
}

func (b *BaseHost) ExecExt(cmd string, printOutput bool, printLine bool) (stdout string, code int, err error) {
	return util.Exec(context.Background(), cmd, printOutput, printLine)
}

func (b *BaseHost) Fetch(local, remote string, printOutput bool, printLine bool) error {
	output, _, err := b.Exec(context.Background(), b.SudoPrefixIfNecessary(fmt.Sprintf("cat %s | base64 -w 0", remote)), printOutput, printLine)
	if err != nil {
		return fmt.Errorf("open remote file failed %v, remote path: %s", err, remote)
	}

	err = util.MkFileFullPathDir(local)
	if err != nil {
		return err
	}
	// open local Destination file
	dstFile, err := os.Create(local)
	if err != nil {
		return fmt.Errorf("create local file failed %v", err)
	}
	defer dstFile.Close()
	// copy to local file
	//_, err = srcFile.WriteTo(dstFile)
	if base64Str, err := base64.StdEncoding.DecodeString(output); err != nil {
		return err
	} else {
		if _, err = dstFile.WriteString(string(base64Str)); err != nil {
			return err
		}
	}

	return nil
}

func (b *BaseHost) Scp(local, remote string) error {
	var remoteDir = filepath.Dir(remote)
	if !util.IsExist(remoteDir) {
		if err := util.Mkdir(remoteDir); err != nil {
			return err
		}
	}
	var cmd = fmt.Sprintf("cp %s %s", local, remote)
	_, _, err := b.Exec(context.Background(), cmd, false, false)
	return err
}

func (b *BaseHost) SudoScp(local, remote string) error {
	remoteTmp := filepath.Join(common.TmpDir, remote)

	// remoteTmp := remote
	if err := b.Scp(local, remoteTmp); err != nil { // ~ copy
		return err
	}

	baseRemotePath := remote
	if !util.IsDir(local) {
		baseRemotePath = filepath.Dir(remote)
	}

	// todo macos need to test
	if err := b.MkDirAll(baseRemotePath, "755"); err != nil {
		return err
	}

	var remoteDir = filepath.Dir(remote)
	if !util.IsExist(remoteDir) {
		util.Mkdir(remoteDir)
	}

	if _, err := b.SudoCmd(fmt.Sprintf(common.MoveCmd, remoteTmp, remote), false, false); err != nil {
		return err
	}

	if _, err := b.SudoCmd(fmt.Sprintf("rm -rf %s", filepath.Join(common.TmpDir, "*")), false, false); err != nil {
		return err
	}

	return nil
}

func (b *BaseHost) FileExist(f string) (bool, error) {
	return b.pathExist(f)
}
func (b *BaseHost) DirExist(d string) (bool, error) {
	return b.pathExist(d)
}

func (b *BaseHost) pathExist(path string) (bool, error) {
	_, err := os.Stat(path)
	if err == nil {
		return true, nil
	}
	if os.IsNotExist(err) {
		return false, nil
	}
	return false, err
}

func (b *BaseHost) MkDir(path string) error {
	if err := b.MkDirAll(path, ""); err != nil {
		logger.Errorf("make dir %s failed: %v", path, err)
		return err
	}
	return nil
}

func (b *BaseHost) Cmd(cmd string, printOutput bool, printLine bool) (string, error) {
	stdout, _, err := b.Exec(context.Background(), cmd, printOutput, printLine)
	if err != nil {
		return stdout, err
	}
	return stdout, nil
}

func (b *BaseHost) CmdContext(ctx context.Context, cmd string, printOutput bool, printLine bool) (string, error) {
	stdout, _, err := b.Exec(ctx, cmd, printOutput, printLine)
	if err != nil {
		return stdout, err
	}
	return stdout, nil
}

func (b *BaseHost) CmdExt(cmd string, printOutput bool, printLine bool) (string, error) {
	stdout, _, err := util.Exec(context.Background(), cmd, printOutput, printLine)

	return stdout, err
}

func (b *BaseHost) SudoPrefixIfNecessary(cmd string) string {
	if b.GetUser() == "root" {
		return cmd
	}
	return SudoPrefix(cmd)
}

func (b *BaseHost) SudoCmd(cmd string, printOutput bool, printLine bool) (string, error) {
	return b.Cmd(b.SudoPrefixIfNecessary(cmd), printOutput, printLine)
}

func (b *BaseHost) SudoCmdContext(ctx context.Context, cmd string, printOutput bool, printLine bool) (string, error) {
	return b.CmdContext(ctx, b.SudoPrefixIfNecessary(cmd), printOutput, printLine)
}

func (b *BaseHost) CmdExtWithContext(ctx context.Context, cmd string, printOutput bool, printLine bool) (string, error) {
	stdout, _, err := util.ExecWithContext(ctx, cmd, printOutput, printLine)

	if printOutput {
		logger.Infof("[exec] %s CMD: %s, OUTPUT: \n%s", b.GetName(), cmd, stdout)
	}

	logger.Debugf("[exec] %s CMD: %s, OUTPUT: %s", b.GetName(), cmd, stdout)

	return stdout, err
}

func (b *BaseHost) MkDirAll(path string, mode string) error {
	if mode == "" {
		mode = "775"
	}
	mkDstDir := fmt.Sprintf("mkdir -p -m %s %s || true", mode, path)
	if _, _, err := b.Exec(context.Background(), b.SudoPrefixIfNecessary(mkDstDir), false, false); err != nil {
		return err
	}

	return nil
}

func (b *BaseHost) Chmod(path string, mode os.FileMode) error {
	if err := os.Chmod(path, mode); err != nil {
		logger.Errorf("chmod path %s failed: %v", path, err)
		return err
	}
	return nil
}

func (b *BaseHost) FileMd5(path string) (string, error) {
	cmd := fmt.Sprintf("md5sum %s | cut -d\" \" -f1", path)
	out, _, err := b.ExecExt(cmd, false, false)
	if err != nil {
		logger.Errorf("count %s md5 failed: %v", path, err)
		return "", err
	}
	return out, nil
}
