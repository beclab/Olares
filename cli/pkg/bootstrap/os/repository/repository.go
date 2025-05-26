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

type Interface interface {
	Backup(runtime connector.Runtime) error
	IsAlreadyBackUp() bool
	Add(runtime connector.Runtime, path string) error
	Update(runtime connector.Runtime) error
	Install(runtime connector.Runtime, pkg ...string) error
	Reset(runtime connector.Runtime) error
}

func New(os string) (Interface, error) {
	switch strings.ToLower(os) {
	case "ubuntu", "debian":
		return NewDeb(), nil
	case "centos", "rhel":
		return NewRPM(), nil
	default:
		return nil, fmt.Errorf("unsupported operation system %s", os)
	}
}
