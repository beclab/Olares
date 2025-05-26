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

type RedhatPackageManager struct {
	backup bool
}

func NewRPM() Interface {
	return &RedhatPackageManager{}
}

func (r *RedhatPackageManager) Backup(runtime connector.Runtime) error {
	if _, err := runtime.GetRunner().SudoCmd("mv /etc/yum.repos.d /etc/yum.repos.d.kubekey.bak", false, false); err != nil {
		return err
	}

	if _, err := runtime.GetRunner().SudoCmd("mkdir -p /etc/yum.repos.d", false, false); err != nil {
		return err
	}
	r.backup = true
	return nil
}

func (r *RedhatPackageManager) IsAlreadyBackUp() bool {
	return r.backup
}

func (r *RedhatPackageManager) Add(runtime connector.Runtime, path string) error {
	if !r.IsAlreadyBackUp() {
		return fmt.Errorf("linux repository must be backuped before")
	}

	if _, err := runtime.GetRunner().SudoCmd("rm -rf /etc/yum.repos.d/*", false, false); err != nil {
		return err
	}

	content := fmt.Sprintf(`cat << EOF > /etc/yum.repos.d/CentOS-local.repo
[base-local]
name=CentOS7.6-local

baseurl=file://%s

enabled=1 

gpgcheck=1
gpgkey=file:///etc/pki/rpm-gpg/RPM-GPG-KEY-CentOS-7
EOF
`, path)
	if _, err := runtime.GetRunner().SudoCmd(content, false, false); err != nil {
		return err
	}

	return nil
}

func (r *RedhatPackageManager) Update(runtime connector.Runtime) error {
	if _, err := runtime.GetRunner().SudoCmd("yum clean all && yum makecache", true, false); err != nil {
		return err
	}
	return nil
}

func (r *RedhatPackageManager) Install(runtime connector.Runtime, pkg ...string) error {
	defaultPkg := []string{"openssl", "socat", "conntrack", "ipset", "ebtables", "chrony", "ipvsadm"}
	if len(pkg) == 0 {
		pkg = defaultPkg
	} else {
		pkg = append(pkg, defaultPkg...)
	}

	str := strings.Join(pkg, " ")
	if _, err := runtime.GetRunner().SudoCmd(fmt.Sprintf("yum install -y %s", str), true, false); err != nil {
		return err
	}
	return nil
}

func (r *RedhatPackageManager) Reset(runtime connector.Runtime) error {
	if _, err := runtime.GetRunner().SudoCmd("rm -rf /etc/yum.repos.d", false, false); err != nil {
		return err
	}

	if _, err := runtime.GetRunner().SudoCmd("mv /etc/yum.repos.d.kubekey.bak /etc/yum.repos.d", false, false); err != nil {
		return err
	}

	return nil
}
