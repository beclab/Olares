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

package storage

import (
	"regexp"

	"bytetrade.io/web3os/installer/pkg/common"
	"bytetrade.io/web3os/installer/pkg/core/connector"
	"bytetrade.io/web3os/installer/pkg/core/logger"
	"github.com/pkg/errors"
)

type CheckDefaultStorageClass struct {
	common.KubePrepare
}

func (c *CheckDefaultStorageClass) PreCheck(runtime connector.Runtime) (bool, error) {
	output, err := runtime.GetRunner().SudoCmd(
		"/usr/local/bin/kubectl get sc --no-headers | grep '(default)' | wc -l", false, false)
	if err != nil {
		return false, errors.Wrap(errors.WithStack(err), "check default storageClass failed")
	}

	reg := regexp.MustCompile(`([\d])`)
	defaultStorageClassNum := reg.FindStringSubmatch(output)[0]
	if defaultStorageClassNum == "0" {
		return true, nil
	}
	host := runtime.RemoteHost()
	logger.Info(host.GetName(), "Default storageClass in cluster is not unique!")
	return false, nil
}
