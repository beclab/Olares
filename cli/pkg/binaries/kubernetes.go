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

package binaries

import (
	"fmt"
	"os/exec"

	kubekeyapiv1alpha2 "bytetrade.io/web3os/installer/apis/kubekey/v1alpha2"
	"bytetrade.io/web3os/installer/pkg/common"
	"bytetrade.io/web3os/installer/pkg/core/cache"
	"bytetrade.io/web3os/installer/pkg/core/logger"
	"bytetrade.io/web3os/installer/pkg/core/util"
	"bytetrade.io/web3os/installer/pkg/files"
	"github.com/pkg/errors"
)

// CriDownloadHTTP defines the kubernetes' binaries that need to be downloaded in advance and downloads them.
func CriDownloadHTTP(kubeConf *common.KubeConf, path, arch, osType, osVersion, osPlatformFamily string, pipelineCache *cache.Cache) error {

	binaries := []*files.KubeBinary{}
	switch kubeConf.Arg.Type {
	case common.Docker:
		docker := files.NewKubeBinary("docker", arch, osType, osVersion, osPlatformFamily, kubekeyapiv1alpha2.DefaultDockerVersion, path, "")
		binaries = append(binaries, docker)
	case common.Containerd:
		containerd := files.NewKubeBinary("containerd", arch, osType, osVersion, osPlatformFamily, kubekeyapiv1alpha2.DefaultContainerdVersion, path, "")
		runc := files.NewKubeBinary("runc", arch, osType, osVersion, osPlatformFamily, kubekeyapiv1alpha2.DefaultRuncVersion, path, "")
		crictl := files.NewKubeBinary("crictl", arch, osType, osVersion, osPlatformFamily, kubekeyapiv1alpha2.DefaultCrictlVersion, path, "")
		binaries = append(binaries, containerd, runc, crictl)
	default:
	}
	binariesMap := make(map[string]*files.KubeBinary)
	for _, binary := range binaries {
		if err := binary.CreateBaseDir(); err != nil {
			return errors.Wrapf(errors.WithStack(err), "create file %s base dir failed", binary.FileName)
		}

		logger.Infof("%s downloading %s %s %s ...", common.LocalHost, arch, binary.ID, binary.Version)

		binariesMap[binary.ID] = binary
		if util.IsExist(binary.Path()) {
			// download it again if it's incorrect
			if err := binary.SHA256Check(); err != nil {
				p := binary.Path()
				_ = exec.Command("/bin/sh", "-c", fmt.Sprintf("rm -f %s", p)).Run()
			} else {
				logger.Infof("%s %s is existed", common.LocalHost, binary.ID)
				continue
			}
		}

		if err := binary.Download(); err != nil {
			return fmt.Errorf("Failed to download %s binary: %s error: %w ", binary.ID, binary.Url, err)
		}
	}

	pipelineCache.Set(common.KubeBinaries+"-"+arch, binariesMap)
	return nil
}
