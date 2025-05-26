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
	"os"
	"text/template"

	"bytetrade.io/web3os/installer/pkg/core/util"
	"bytetrade.io/web3os/installer/pkg/version/kubesphere/templates"
	versionutil "k8s.io/apimachinery/pkg/util/version"
)

type KsInstaller struct {
	Version                      string
	CRDTemplate                  *template.Template
	ClusterConfigurationTemplate *template.Template
	K8sSupportVersions           []string
	UpgradeSupportVersions       []string
}

func (k *KsInstaller) CCToString() string {
	str, err := util.Render(k.ClusterConfigurationTemplate, util.Data{
		"Tag": k.Version,
	})
	if err != nil {
		os.Exit(0)
	}
	return str
}

func (k *KsInstaller) K8sSupport(version string) bool {
	k8sVersion := versionutil.MustParseSemantic(version)
	for i := range k.K8sSupportVersions {
		if k.K8sSupportVersions[i] == fmt.Sprintf("v%v.%v", k8sVersion.Major(), k8sVersion.Minor()) {
			return true
		}
	}
	return false
}

func (k *KsInstaller) UpgradeSupport(version string) bool {
	for i := range k.UpgradeSupportVersions {
		if k.UpgradeSupportVersions[i] == version {
			return true
		}
	}
	return false
}

var KsV211 = &KsInstaller{
	Version:                      V211.String(),
	CRDTemplate:                  templates.KsInstaller,
	ClusterConfigurationTemplate: templates.V211,
	K8sSupportVersions: []string{
		"v1.15",
		"v1.16",
		"v1.17",
		"v1.18",
	},
	UpgradeSupportVersions: []string{
		V211.String(),
	},
}

var KsV300 = &KsInstaller{
	Version:                      V300.String(),
	CRDTemplate:                  templates.KsInstaller,
	ClusterConfigurationTemplate: templates.V300,
	K8sSupportVersions: []string{
		"v1.15",
		"v1.16",
		"v1.17",
		"v1.18",
	},
	UpgradeSupportVersions: []string{
		V211.String(),
	},
}

var KsV310 = &KsInstaller{
	Version:                      V310.String(),
	CRDTemplate:                  templates.KsInstaller,
	ClusterConfigurationTemplate: templates.V310,
	K8sSupportVersions: []string{
		"v1.15",
		"v1.16",
		"v1.17",
		"v1.18",
		"v1.19",
		"v1.20",
	},
	UpgradeSupportVersions: []string{
		V300.String(),
	},
}

var KsV311 = &KsInstaller{
	Version:                      V311.String(),
	CRDTemplate:                  templates.KsInstaller,
	ClusterConfigurationTemplate: templates.V311,
	K8sSupportVersions: []string{
		"v1.15",
		"v1.16",
		"v1.17",
		"v1.18",
		"v1.19",
		"v1.20",
	},
	UpgradeSupportVersions: []string{
		V300.String(),
		V310.String(),
	},
}

var KsV320 = &KsInstaller{
	Version:                      V320.String(),
	CRDTemplate:                  templates.KsInstaller,
	ClusterConfigurationTemplate: templates.V320,
	K8sSupportVersions: []string{
		"v1.19",
		"v1.20",
		"v1.21",
		"v1.22",
	},
	UpgradeSupportVersions: []string{
		V310.String(),
		V311.String(),
	},
}

var KsV321 = &KsInstaller{
	Version:                      V321.String(),
	CRDTemplate:                  templates.KsInstaller,
	ClusterConfigurationTemplate: templates.V321,
	K8sSupportVersions: []string{
		"v1.19",
		"v1.20",
		"v1.21",
		"v1.22",
	},
	UpgradeSupportVersions: []string{
		V310.String(),
		V311.String(),
		V320.String(),
	},
}

var KsV330 = &KsInstaller{
	Version:                      V330.String(),
	CRDTemplate:                  templates.KsInstaller,
	ClusterConfigurationTemplate: templates.V330,
	K8sSupportVersions: []string{
		"v1.19",
		"v1.20",
		"v1.21",
		"v1.22",
		"v1.23",
	},
	UpgradeSupportVersions: []string{
		V320.String(),
		V321.String(),
	},
}

var KsV331 = &KsInstaller{
	Version:                      V331.String(),
	CRDTemplate:                  templates.KsInstaller,
	ClusterConfigurationTemplate: templates.V331,
	K8sSupportVersions: []string{
		"v1.19",
		"v1.20",
		"v1.21",
		"v1.22",
		"v1.23",
	},
	UpgradeSupportVersions: []string{
		V330.String(),
		V320.String(),
		V321.String(),
	},
}
