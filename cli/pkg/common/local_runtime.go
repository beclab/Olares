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

package common

import (
	"fmt"
	"os"
	"os/exec"
	"os/user"

	"bytetrade.io/web3os/installer/pkg/core/connector"
	"github.com/pkg/errors"
)

type LocalRuntime struct {
	connector.BaseRuntime
}

func NewLocalRuntime(debug, ingoreErr bool) (LocalRuntime, error) {
	var localRuntime LocalRuntime
	u, err := user.Current()
	if err != nil {
		return localRuntime, err
	}
	if u.Username != "root" {
		return localRuntime, fmt.Errorf("current user is %s. Please use root", u.Username)
	}

	if output, err := exec.Command("/bin/sh", "-c", "if [ ! -f \"$HOME/.ssh/id_rsa\" ]; then mkdir -p \"$HOME/.ssh\" && ssh-keygen -t rsa-sha2-512 -P \"\" -f $HOME/.ssh/id_rsa && ls $HOME/.ssh;fi;").CombinedOutput(); err != nil {
		return localRuntime, errors.New(fmt.Sprintf("Failed to generate public key: %v\n%s", err, string(output)))
	}
	if output, err := exec.Command("/bin/sh", "-c", "echo \"\n$(cat $HOME/.ssh/id_rsa.pub)\" >> $HOME/.ssh/authorized_keys && awk ' !x[$0]++{print > \"'$HOME'/.ssh/authorized_keys.tmp\"}' $HOME/.ssh/authorized_keys && mv $HOME/.ssh/authorized_keys.tmp $HOME/.ssh/authorized_keys").CombinedOutput(); err != nil {
		return localRuntime, errors.New(fmt.Sprintf("Failed to copy public key to authorized_keys: %v\n%s", err, string(output)))
	}

	name, err := os.Hostname()
	if err != nil {
		return localRuntime, err
	}
	base := connector.NewBaseRuntime(name, connector.NewDialer(), debug, ingoreErr, nil, "", "", "", false, nil)

	host := connector.NewHost()
	host.Name = name
	host.Address = ""
	host.InternalAddress = ""
	host.Port = 22
	host.User = u.Name
	host.Password = ""
	host.PrivateKeyPath = fmt.Sprintf("%s/.ssh/id_rsa", u.HomeDir)
	host.Arch = ""
	host.SetRole(KubeKey)

	base.AppendHost(host)
	base.AppendRoleMap(host)

	local := LocalRuntime{base}
	return local, nil
}
