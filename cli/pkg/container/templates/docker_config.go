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

package templates

import (
	"fmt"
	"strings"
	"text/template"

	"bytetrade.io/web3os/installer/pkg/common"
	"github.com/lithammer/dedent"
)

var DockerConfig = template.Must(template.New("daemon.json").Parse(
	dedent.Dedent(`{
  "log-opts": {
    "max-size": "5m",
    "max-file":"3"
  },
  {{- if .DataRoot }}
  "data-root": {{ .DataRoot }},
  {{- end}}
  {{- if .Mirrors }}
  "registry-mirrors": [{{ .Mirrors }}],
  {{- end}}
  {{- if .InsecureRegistries }}
  "insecure-registries": [{{ .InsecureRegistries }}],
  {{- end}}
  "exec-opts": ["native.cgroupdriver=systemd"]
}
    `)))

func Mirrors(kubeConf *common.KubeConf) string {
	var mirrors string
	if kubeConf.Cluster.Registry.RegistryMirrors != nil {
		var mirrorsArr []string
		for _, mirror := range kubeConf.Cluster.Registry.RegistryMirrors {
			mirrorsArr = append(mirrorsArr, fmt.Sprintf("\"%s\"", mirror))
		}
		mirrors = strings.Join(mirrorsArr, ", ")
	}
	return mirrors
}

func InsecureRegistries(kubeConf *common.KubeConf) string {
	var insecureRegistries string
	if kubeConf.Cluster.Registry.InsecureRegistries != nil {
		var registriesArr []string
		for _, repo := range kubeConf.Cluster.Registry.InsecureRegistries {
			registriesArr = append(registriesArr, fmt.Sprintf("\"%s\"", repo))
		}
		insecureRegistries = strings.Join(registriesArr, ", ")
	}
	return insecureRegistries
}

func DataRoot(kubeConf *common.KubeConf) string {
	var dataRoot string
	if kubeConf.Cluster.Registry.DataRoot != "" {
		dataRoot = fmt.Sprintf("\"%s\"", kubeConf.Cluster.Registry.DataRoot)
	}
	return dataRoot
}
