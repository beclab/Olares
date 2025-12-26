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

package kubesphere

import (
	"fmt"
	"path"

	"github.com/beclab/Olares/cli/pkg/common"
	"github.com/beclab/Olares/cli/pkg/core/connector"
	"github.com/beclab/Olares/cli/pkg/core/logger"
	"github.com/beclab/Olares/cli/pkg/core/util"
	"github.com/pkg/errors"
)

type DeleteKubeSphereCaches struct {
	common.KubeAction
}

func (d *DeleteKubeSphereCaches) Execute(runtime connector.Runtime) error {
	var files = []string{
		path.Join(runtime.GetInstallerDir(), "files"),
		path.Join(runtime.GetInstallerDir(), "cli"),
	}

	for _, f := range files {
		if _, err := runtime.GetRunner().SudoCmd(fmt.Sprintf("rm -rf %s", f), false, true); err != nil {
			return errors.Wrapf(errors.WithStack(err), "delete %s failed", f)
		}
	}

	return nil
}

type CreateNamespace struct {
	common.KubeAction
}

func (c *CreateNamespace) Execute(runtime connector.Runtime) error {
	var kubectl, ok = c.PipelineCache.GetMustString(common.CacheCommandKubectlPath)
	if !ok || kubectl == "" {
		kubectl = path.Join(common.BinDir, "kubectl")
	}

	var cmd = fmt.Sprintf(`cat <<EOF | %s apply -f -
apiVersion: v1
kind: Namespace
metadata:
  name: kubesphere-system
---
apiVersion: v1
kind: Namespace
metadata:
  name: kubesphere-controls-system
---
apiVersion: v1
kind: Namespace
metadata:
  name: kubesphere-monitoring-system
EOF`, kubectl)
	_, err := runtime.GetRunner().SudoCmd(cmd, false, true)
	if err != nil {
		return errors.Wrap(errors.WithStack(err), "create namespace: kubesphere-system and kubesphere-monitoring-system")
	}
	return nil
}

type GetKubeCommand struct {
	common.KubeAction
}

func (t *GetKubeCommand) Execute(runtime connector.Runtime) error {
	kubectlpath, err := util.GetCommand(common.CommandKubectl)
	if err != nil || kubectlpath == "" {
		return fmt.Errorf("kubectl not found")
	}

	t.PipelineCache.Set(common.CacheCommandKubectlPath, kubectlpath)
	logger.InfoInstallationProgress("k8s and kubesphere installation is complete")
	return nil
}

type Check struct {
	common.KubeAction
}

func (c *Check) Execute(runtime connector.Runtime) error {
	var kubectlpath, err = util.GetCommand(common.CommandKubectl)
	if err != nil {
		return fmt.Errorf("kubectl not found")
	}

	var labels = []string{"app=ks-apiserver"}

	for _, label := range labels {
		var cmd = fmt.Sprintf("%s get pod -n %s -l '%s' -o jsonpath='{.items[0].status.phase}'", kubectlpath, common.NamespaceKubesphereSystem, label)
		rphase, _ := runtime.GetRunner().SudoCmd(cmd, false, false)
		if rphase != "Running" {
			return errors.New("Waiting for KubeSphere to be Running")
		}
	}

	//if runtime.GetSystemInfo().IsDarwin() {
	//	epIPCMD := fmt.Sprintf("%s -n kubesphere-system get ep ks-controller-manager -o jsonpath='{.subsets[*].addresses[*].ip}'", kubectlpath)
	//	epIP, _ := runtime.GetRunner().SudoCmd(epIPCMD, false, false)
	//	if net.ParseIP(strings.TrimSpace(epIP)) == nil {
	//		return errors.New("Waiting for ks-controller-manager svc endpoints to be populated")
	//	}
	//	// we can't check the svc connectivity in macOS host
	//	// so just wait for some time for the proxy to take effect
	//	time.Sleep(5 * time.Second)
	//	return nil
	//}
	//
	//svcIPCMD := fmt.Sprintf("%s -n kubesphere-system get svc ks-controller-manager -o jsonpath='{.spec.clusterIP}'", kubectlpath)
	//svcIP, err := runtime.GetRunner().SudoCmd(svcIPCMD, false, false)
	//if err != nil {
	//	return errors.New("Waiting for ks-controller-manager service to be reachable")
	//}
	//
	//conn, err := net.DialTimeout("tcp", net.JoinHostPort(svcIP, strconv.Itoa(443)), 10*time.Second)
	//if err != nil {
	//	return errors.New("Waiting for ks-controller-manager service to be reachable")
	//}
	//defer conn.Close()
	return nil
}
