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

	"bytetrade.io/web3os/installer/pkg/common"
	"bytetrade.io/web3os/installer/pkg/core/connector"
	"bytetrade.io/web3os/installer/pkg/core/logger"
	"bytetrade.io/web3os/installer/pkg/manifest"
	"github.com/pkg/errors"
)

type InstallAppArmorTask struct {
	common.KubeAction
	manifest.ManifestAction
}

func (t *InstallAppArmorTask) Execute(runtime connector.Runtime) error {
	fileName, err := GetUbutun24AppArmor(t.BaseDir, t.Manifest)
	if err != nil {
		logger.Fatal("failed to download apparmor: %v", err)
	}

	if _, err := runtime.GetRunner().SudoCmd(fmt.Sprintf("dpkg -i %s", fileName), false, true); err != nil {
		logger.Errorf("failed to install apparmor: %v", err)
		return err
	}

	return nil
}

type CriDownload struct {
	common.KubeAction
	manifest.ManifestAction
}

func (d *CriDownload) Execute(runtime connector.Runtime) error {
	cfg := d.KubeConf.Cluster
	archMap := make(map[string]bool)
	for _, host := range cfg.Hosts {
		switch host.Arch {
		case "amd64":
			archMap["amd64"] = true
		case "arm64":
			archMap["arm64"] = true
		default:
			return errors.New(fmt.Sprintf("Unsupported architecture: %s", host.Arch))
		}
	}

	var systemInfo = runtime.GetSystemInfo()
	var osType = systemInfo.GetOsType()
	var osPlatformFamily = systemInfo.GetOsPlatformFamily()
	var osVersion = systemInfo.GetOsVersion()
	for arch := range archMap {
		if err := CriDownloadHTTP(d.KubeConf, runtime.GetWorkDir(), arch, osType, osVersion, osPlatformFamily, d.PipelineCache); err != nil {
			return err
		}
	}
	return nil
}

// TODO: install helm
