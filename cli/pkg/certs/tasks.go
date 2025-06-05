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

package certs

import (
	"fmt"
	"path"
	"path/filepath"

	"bytetrade.io/web3os/installer/pkg/certs/templates"
	"bytetrade.io/web3os/installer/pkg/common"
	cc "bytetrade.io/web3os/installer/pkg/core/common"
	"bytetrade.io/web3os/installer/pkg/core/connector"
	"bytetrade.io/web3os/installer/pkg/core/util"
	"github.com/pkg/errors"
)

type EnableRenewService struct {
	common.KubeAction
}

func (e *EnableRenewService) Execute(runtime connector.Runtime) error {
	if _, err := runtime.GetRunner().SudoCmd(
		"chmod +x /usr/local/bin/kube-scripts/k8s-certs-renew.sh && systemctl enable --now k8s-certs-renew.timer",
		false, false); err != nil {
		return errors.Wrap(errors.WithStack(err), "enable k8s renew certs service failed")
	}
	return nil
}

type UninstallAutoRenewCerts struct {
	common.KubeAction
}

func (u *UninstallAutoRenewCerts) Execute(runtime connector.Runtime) error {
	_, _ = runtime.GetRunner().SudoCmd("systemctl disable k8s-certs-renew.timer 1>/dev/null 2>/dev/null", false, false)
	_, _ = runtime.GetRunner().SudoCmd("systemctl stop k8s-certs-renew.timer 1>/dev/null 2>/dev/null", false, false)

	files := []string{
		filepath.Join("/usr/local/bin/kube-scripts/", templates.K8sCertsRenewScript.Name()),
		filepath.Join("/etc/systemd/system/", templates.K8sCertsRenewService.Name()),
		filepath.Join("/etc/systemd/system/", templates.K8sCertsRenewTimer.Name()),
	}
	for _, file := range files {
		_, _ = runtime.GetRunner().SudoCmd(fmt.Sprintf("rm -rf %s", file), false, false)
	}

	return nil
}

type UninstallCertsFiles struct {
	common.KubeAction
}

func (t *UninstallCertsFiles) Execute(runtime connector.Runtime) error {
	var p = path.Join(runtime.GetBaseDir(), cc.Cli)
	if util.IsExist(p) {
		return util.RemoveDir(p)
	}
	return nil
}
