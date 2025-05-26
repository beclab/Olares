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

package registry

import (
	"fmt"
	"path/filepath"
	"strings"

	"bytetrade.io/web3os/installer/pkg/common"
	"bytetrade.io/web3os/installer/pkg/core/connector"
	"bytetrade.io/web3os/installer/pkg/files"
	"bytetrade.io/web3os/installer/pkg/utils"
	"github.com/pkg/errors"
)

type SyncCertsFile struct {
	common.KubeAction
}

func (s *SyncCertsFile) Execute(runtime connector.Runtime) error {
	localCertsDir, ok := s.ModuleCache.Get(LocalCertsDir)
	if !ok {
		return errors.New("get registry local certs dir by module cache failed")
	}
	files, ok := s.ModuleCache.Get(CertsFileList)
	if !ok {
		return errors.New("get registry certs file list by module cache failed")
	}
	dir := localCertsDir.(string)
	fileList := files.([]string)

	for _, fileName := range fileList {
		if err := runtime.GetRunner().SudoScp(filepath.Join(dir, fileName), filepath.Join(common.RegistryCertDir, fileName)); err != nil {
			return errors.Wrap(errors.WithStack(err), "scp registry certs file failed")
		}
	}

	return nil
}

type SyncCertsToAllNodes struct {
	common.KubeAction
}

func (s *SyncCertsToAllNodes) Execute(runtime connector.Runtime) error {
	localCertsDir, ok := s.ModuleCache.Get(LocalCertsDir)
	if !ok {
		return errors.New("get registry local certs dir by module cache failed")
	}
	files, ok := s.ModuleCache.Get(CertsFileList)
	if !ok {
		return errors.New("get registry certs file list by module cache failed")
	}
	dir := localCertsDir.(string)
	fileList := files.([]string)

	for _, fileName := range fileList {
		var dstFileName string
		switch fileName {
		case "ca.pem":
			dstFileName = "ca.crt"
		case "ca-key.pem":
			continue
		default:
			if strings.HasSuffix(fileName, "-key.pem") {
				dstFileName = strings.Replace(fileName, "-key.pem", ".key", -1)
			} else {
				dstFileName = strings.Replace(fileName, ".pem", ".cert", -1)
			}
		}

		if err := runtime.GetRunner().SudoScp(filepath.Join(dir, fileName), filepath.Join(filepath.Join("/etc/docker/certs.d", RegistryCertificateBaseName), dstFileName)); err != nil {
			return errors.Wrap(errors.WithStack(err), "scp registry certs file to /etc/docker/certs.d/ failed")
		}

		if err := runtime.GetRunner().SudoScp(filepath.Join(dir, fileName), filepath.Join(common.RegistryCertDir, dstFileName)); err != nil {
			return errors.Wrap(errors.WithStack(err), fmt.Sprintf("scp registry certs file to %s failed", common.RegistryCertDir))
		}
	}

	return nil
}

type InstallRegistryBinary struct {
	common.KubeAction
}

func (g *InstallRegistryBinary) Execute(runtime connector.Runtime) error {
	if err := utils.ResetTmpDir(runtime); err != nil {
		return err
	}

	binariesMapObj, ok := g.PipelineCache.Get(common.KubeBinaries + "-" + runtime.RemoteHost().GetArch())
	if !ok {
		return errors.New("get KubeBinary by pipeline cache failed")
	}
	binariesMap := binariesMapObj.(map[string]*files.KubeBinary)

	registry, ok := binariesMap[common.Registry]
	if !ok {
		return errors.New("get KubeBinary key registry by pipeline cache failed")
	}

	dst := filepath.Join(common.TmpDir, registry.FileName)
	if err := runtime.GetRunner().Scp(registry.Path(), dst); err != nil {
		return errors.Wrap(errors.WithStack(err), "sync etcd tar.gz failed")
	}

	installCmd := fmt.Sprintf("tar -zxf %s && mv -f registry /usr/local/bin/ && chmod +x /usr/local/bin/registry", dst)
	if _, err := runtime.GetRunner().SudoCmd(installCmd, false, false); err != nil {
		return errors.Wrap(errors.WithStack(err), "install etcd binaries failed")
	}
	return nil
}

type StartRegistryService struct {
	common.KubeAction
}

func (g *StartRegistryService) Execute(runtime connector.Runtime) error {
	installCmd := "systemctl daemon-reload && systemctl enable registry && systemctl restart registry"
	if _, err := runtime.GetRunner().SudoCmd(installCmd, false, false); err != nil {
		return errors.Wrap(errors.WithStack(err), "start registry service failed")
	}

	fmt.Println()
	fmt.Println("Local image registry created successfully. Address: dockerhub.kubekey.local")
	fmt.Println()

	return nil
}

type InstallDockerCompose struct {
	common.KubeAction
}

func (g *InstallDockerCompose) Execute(runtime connector.Runtime) error {
	if err := utils.ResetTmpDir(runtime); err != nil {
		return err
	}

	binariesMapObj, ok := g.PipelineCache.Get(common.KubeBinaries + "-" + runtime.RemoteHost().GetArch())
	if !ok {
		return errors.New("get KubeBinary by pipeline cache failed")
	}
	binariesMap := binariesMapObj.(map[string]*files.KubeBinary)

	compose, ok := binariesMap[common.DockerCompose]
	if !ok {
		return errors.New("get KubeBinary key docker-compose by pipeline cache failed")
	}

	dst := filepath.Join(common.TmpDir, compose.FileName)
	if err := runtime.GetRunner().Scp(compose.Path(), dst); err != nil {
		return errors.Wrap(errors.WithStack(err), "sync docker-compose failed")
	}

	installCmd := fmt.Sprintf("mv -f %s /usr/local/bin/docker-compose && chmod +x /usr/local/bin/docker-compose", dst)
	if _, err := runtime.GetRunner().SudoCmd(installCmd, false, false); err != nil {
		return errors.Wrap(errors.WithStack(err), "install dokcer-compose failed")
	}

	return nil
}

type SyncHarborPackage struct {
	common.KubeAction
}

func (g *SyncHarborPackage) Execute(runtime connector.Runtime) error {
	if err := utils.ResetTmpDir(runtime); err != nil {
		return err
	}

	binariesMapObj, ok := g.PipelineCache.Get(common.KubeBinaries + "-" + runtime.RemoteHost().GetArch())
	if !ok {
		return errors.New("get KubeBinary by pipeline cache failed")
	}
	binariesMap := binariesMapObj.(map[string]*files.KubeBinary)

	harbor, ok := binariesMap[common.Harbor]
	if !ok {
		return errors.New("get KubeBinary key harbor by pipeline cache failed")
	}

	dst := filepath.Join(common.TmpDir, harbor.FileName)
	if err := runtime.GetRunner().Scp(harbor.Path(), dst); err != nil {
		return errors.Wrap(errors.WithStack(err), "sync harbor package failed")
	}

	installCmd := fmt.Sprintf("tar -zxvf %s -C /opt", dst)
	if _, err := runtime.GetRunner().SudoCmd(installCmd, false, false); err != nil {
		return errors.Wrap(errors.WithStack(err), "unzip harbor package failed")
	}

	return nil
}

type StartHarbor struct {
	common.KubeAction
}

func (g *StartHarbor) Execute(runtime connector.Runtime) error {
	startCmd := "cd /opt/harbor && chmod +x install.sh && export PATH=$PATH:/usr/local/bin; ./install.sh --with-notary --with-trivy --with-chartmuseum && systemctl daemon-reload && systemctl enable harbor && systemctl restart harbor"
	if _, err := runtime.GetRunner().SudoCmd(startCmd, false, false); err != nil {
		return errors.Wrap(errors.WithStack(err), "start harbor failed")
	}

	fmt.Println()
	fmt.Println("Local image registry created successfully. Address: dockerhub.kubekey.local")
	fmt.Println()

	return nil
}
