/*
 Copyright 2022 The KubeSphere Authors.

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
	"path/filepath"

	"bytetrade.io/web3os/installer/pkg/bootstrap/os/repository"
	"bytetrade.io/web3os/installer/pkg/common"
	"bytetrade.io/web3os/installer/pkg/core/connector"
	"bytetrade.io/web3os/installer/pkg/core/ending"
	"github.com/pkg/errors"
)

type RollbackUmount struct {
	common.KubeRollback
}

func (r *RollbackUmount) Execute(runtime connector.Runtime, result *ending.ActionResult) error {
	mountPath := filepath.Join(common.TmpDir, "iso")
	umountCmd := fmt.Sprintf("umount %s", mountPath)
	if _, err := runtime.GetRunner().SudoCmd(umountCmd, false, false); err != nil {
		return errors.Wrapf(errors.WithStack(err), "umount %s failed", mountPath)
	}
	return nil
}

type RecoverBackupSuccessNode struct {
	common.KubeRollback
}

func (r *RecoverBackupSuccessNode) Execute(runtime connector.Runtime, result *ending.ActionResult) error {
	if result.Status == ending.SUCCESS {
		host := runtime.RemoteHost()
		repo, ok := host.GetCache().Get("repo")
		if !ok {
			return errors.New("get repo failed by host cache")
		}

		re := repo.(repository.Interface)
		if err := re.Reset(runtime); err != nil {
			return errors.Wrapf(errors.WithStack(err), "reset repository failed")
		}
	}

	mountPath := filepath.Join(common.TmpDir, "iso")
	umountCmd := fmt.Sprintf("umount %s", mountPath)
	if _, err := runtime.GetRunner().SudoCmd(umountCmd, false, false); err != nil {
		return errors.Wrapf(errors.WithStack(err), "umount %s failed", mountPath)
	}
	return nil
}

type RecoverRepository struct {
	common.KubeRollback
}

func (r *RecoverRepository) Execute(runtime connector.Runtime, result *ending.ActionResult) error {
	host := runtime.RemoteHost()
	repo, ok := host.GetCache().Get("repo")
	if !ok {
		return errors.New("get repo failed by host cache")
	}

	re := repo.(repository.Interface)
	_ = re.Reset(runtime)

	mountPath := filepath.Join(common.TmpDir, "iso")
	umountCmd := fmt.Sprintf("umount %s", mountPath)
	if _, err := runtime.GetRunner().SudoCmd(umountCmd, false, false); err != nil {
		return errors.Wrapf(errors.WithStack(err), "umount %s failed", mountPath)
	}
	return nil
}
