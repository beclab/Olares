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
	"fmt"
	"os"
	"path/filepath"

	"bytetrade.io/web3os/installer/pkg/core/common"
	"bytetrade.io/web3os/installer/pkg/core/logger"
	"bytetrade.io/web3os/installer/pkg/core/util"
)

type Runner struct {
	Conn  Connection
	Debug bool
	Host  Host
	Index int
}

func (r *Runner) Exec(cmd string, printOutput, printLine bool) (string, int, error) {
	if r.Conn == nil {
		return r.Host.ExecExt(cmd, printOutput, printLine)
	}

	stdout, code, err := r.Conn.Exec(cmd, r.Host)
	logger.Debugf("command: [%s]\n%s", r.Host.GetName(), cmd)
	if stdout != "" {
		logger.Debugf("stdout: [%s]\n%s", r.Host.GetName(), stdout)
	}

	if err != nil {
		logger.Errorf("[exec] %s CMD: %s, ERROR: %s", r.Host.GetName(), cmd, err)
	}

	if printOutput {
		if stdout != "" {
			fmt.Printf("stdout: [%s]\n%s\n", r.Host.GetName(), stdout)
		}
	}
	return stdout, code, err
}

func (r *Runner) Cmd(cmd string, printOutput, printLine bool) (string, error) {
	stdout, _, err := r.Exec(cmd, printOutput, printLine)
	if err != nil {
		return stdout, err
	}
	return stdout, nil
}

func (r *Runner) CmdContext(ctx context.Context, cmd string, printOutput bool, printLine bool) (string, error) {
	if r.Conn == nil {
		return r.Host.CmdExtWithContext(ctx, cmd, printOutput, printLine)
	}
	return r.Cmd(cmd, printOutput, printLine)
}

func (r *Runner) SudoCmd(cmd string, printOutput, printLine bool) (string, error) {
	return r.Cmd(r.Host.SudoPrefixIfNecessary(cmd), printOutput, printLine)
}

func (r *Runner) SudoCmdContext(ctx context.Context, cmd string, printOutput, printLine bool) (string, error) {
	return r.CmdContext(ctx, r.Host.SudoPrefixIfNecessary(cmd), printOutput, printLine)
}

func (r *Runner) Fetch(local, remote string, printOutput, printLine bool) error {
	if r.Conn == nil {
		return r.Host.Fetch(local, remote, printOutput, printLine)
	}

	if err := r.Conn.Fetch(local, remote, r.Host); err != nil {
		logger.Errorf("fetch remote file %s to local %s failed: %v", remote, local, err)
		return err
	}
	logger.Infof("fetch remote file %s to local %s success", remote, local)
	return nil
}

func (r *Runner) Scp(local, remote string) error {
	if r.Conn == nil {
		return r.Host.Scp(local, remote)
	}

	if err := r.Conn.Scp(local, remote, r.Host); err != nil {
		logger.Debugf("scp local file %s to remote %s failed: %v", local, remote, err)
		return err
	}
	logger.Debugf("scp local file %s to remote %s success", local, remote)
	return nil
}

func (r *Runner) SudoScp(local, remote string) error {
	if r.Conn == nil {
		return r.Host.SudoScp(local, remote)
	}

	// scp to tmp dir
	remoteTmp := filepath.Join(common.TmpDir, remote)
	//remoteTmp := remote
	if err := r.Scp(local, remoteTmp); err != nil {
		return err
	}

	baseRemotePath := remote
	if !util.IsDir(local) {
		baseRemotePath = filepath.Dir(remote)
	}
	if err := r.Conn.MkDirAll(baseRemotePath, "", r.Host); err != nil {
		return err
	}

	if _, err := r.SudoCmd(fmt.Sprintf(common.MoveCmd, remoteTmp, remote), false, false); err != nil {
		return err
	}

	if _, err := r.SudoCmd(fmt.Sprintf("rm -rf %s", filepath.Join(common.TmpDir, "*")), false, false); err != nil {
		return err
	}
	return nil
}

func (r *Runner) FileExist(remote string) (bool, error) {
	if r.Conn == nil {
		return r.Host.FileExist(remote)
	}

	ok := r.Conn.RemoteFileExist(remote, r.Host)
	logger.Debugf("check remote file exist: %v", ok)
	return ok, nil
}

func (r *Runner) DirExist(remote string) (bool, error) {
	if r.Conn == nil {
		return r.Host.DirExist(remote)
	}

	ok, err := r.Conn.RemoteDirExist(remote, r.Host)
	if err != nil {
		logger.Debugf("check remote dir exist failed: %v", err)
		return false, err
	}
	logger.Debugf("check remote dir exist: %v", ok)
	return ok, nil
}

func (r *Runner) MkDir(path string) error {
	if r.Conn == nil {
		return r.Host.MkDir(path)
	}

	if err := r.Conn.MkDirAll(path, "", r.Host); err != nil {
		logger.Errorf("make remote dir %s failed: %v", path, err)
		return err
	}
	return nil
}

func (r *Runner) Chmod(path string, mode os.FileMode) error {
	if r.Conn == nil {
		return r.Host.Chmod(path, mode)
	}

	if err := r.Conn.Chmod(path, mode); err != nil {
		logger.Errorf("chmod remote path %s failed: %v", path, err)
		return err
	}
	return nil
}

func (r *Runner) FileMd5(path string) (string, error) {
	if r.Conn == nil {
		return r.Host.FileMd5(path)
	}

	cmd := fmt.Sprintf("md5sum %s | cut -d\" \" -f1", path)
	out, _, err := r.Conn.Exec(cmd, r.Host)
	if err != nil {
		logger.Errorf("count remote %s md5 failed: %v", path, err)
		return "", err
	}
	return out, nil
}
