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
	"bytetrade.io/web3os/installer/pkg/common"
	"bytetrade.io/web3os/installer/pkg/core/connector"
	"github.com/pkg/errors"
	versionutil "k8s.io/apimachinery/pkg/util/version"
)

type ShouldDeleteCache struct {
	common.KubePrepare
}

func (p *ShouldDeleteCache) PreCheck(runtime connector.Runtime) (bool, error) {
	if p.KubeConf.Arg.DeleteCache {
		return true, nil
	}
	return false, nil
}

type VersionBelowV3 struct {
	common.KubePrepare
}

func (v *VersionBelowV3) PreCheck(runtime connector.Runtime) (bool, error) {
	versionStr, ok := v.PipelineCache.GetMustString(common.KubeSphereVersion)
	if !ok {
		return false, errors.New("get current kubesphere version failed by pipeline cache")
	}
	version := versionutil.MustParseSemantic(versionStr)
	v300 := versionutil.MustParseSemantic("v3.0.0")
	if v.KubeConf.Cluster.KubeSphere.Enabled && v.KubeConf.Cluster.KubeSphere.Version == "v3.0.0" && version.LessThan(v300) {
		return true, nil
	}
	return false, nil
}

type NotEqualDesiredVersion struct {
	common.KubePrepare
}

func (n *NotEqualDesiredVersion) PreCheck(runtime connector.Runtime) (bool, error) {
	ksVersion, ok := n.PipelineCache.GetMustString(common.KubeSphereVersion)
	if !ok {
		ksVersion = ""
	}

	if n.KubeConf.Cluster.KubeSphere.Version == ksVersion {
		return false, nil
	}
	return true, nil
}
