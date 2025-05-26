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

package repository

import (
	"fmt"
	"strings"

	"bytetrade.io/web3os/installer/pkg/core/connector"
)

type Debian struct {
	backup bool
}

func NewDeb() Interface {
	return &Debian{}
}

func (d *Debian) Backup(runtime connector.Runtime) error {
	if _, err := runtime.GetRunner().SudoCmd("mv /etc/apt/sources.list /etc/apt/sources.list.kubekey.bak", false, false); err != nil {
		return err
	}

	if _, err := runtime.GetRunner().SudoCmd("mv /etc/apt/sources.list.d /etc/apt/sources.list.d.kubekey.bak", false, false); err != nil {
		return err
	}

	if _, err := runtime.GetRunner().SudoCmd("mkdir -p /etc/apt/sources.list.d", false, false); err != nil {
		return err
	}
	d.backup = true
	return nil
}

func (d *Debian) IsAlreadyBackUp() bool {
	return d.backup
}

func (d *Debian) Add(runtime connector.Runtime, path string) error {
	if !d.IsAlreadyBackUp() {
		return fmt.Errorf("linux repository must be backuped before")
	}

	if _, err := runtime.GetRunner().SudoCmd("rm -rf /etc/apt/sources.list.d/*", false, false); err != nil {
		return err
	}

	if _, err := runtime.GetRunner().SudoCmd(fmt.Sprintf("echo 'deb [trusted=yes]  file://%s   /' > /etc/apt/sources.list.d/kubekey.list", path),
		true, false); err != nil {
		return err
	}
	return nil
}

func (d *Debian) Update(runtime connector.Runtime) error {
	if _, err := runtime.GetRunner().Cmd("sudo apt-get update", true, false); err != nil {
		return err
	}
	return nil
}

func (d *Debian) Install(runtime connector.Runtime, pkg ...string) error {
	defaultPkg := []string{"socat", "conntrack", "ipset", "ebtables", "chrony", "ipvsadm"}
	if len(pkg) == 0 {
		pkg = defaultPkg
	} else {
		pkg = append(pkg, defaultPkg...)
	}

	str := strings.Join(pkg, " ")
	if _, err := runtime.GetRunner().SudoCmd(fmt.Sprintf("apt install -y %s", str), true, false); err != nil {
		return err
	}
	return nil
}

func (d *Debian) Reset(runtime connector.Runtime) error {
	if _, err := runtime.GetRunner().SudoCmd("rm -rf /etc/apt/sources.list.d", false, false); err != nil {
		return err
	}

	if _, err := runtime.GetRunner().SudoCmd("mv /etc/apt/sources.list.kubekey.bak /etc/apt/sources.list", false, false); err != nil {
		return err
	}

	if _, err := runtime.GetRunner().SudoCmd("mv /etc/apt/sources.list.d.kubekey.bak /etc/apt/sources.list.d", false, false); err != nil {
		return err
	}

	return nil
}
