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

package addons

import (
	"net/url"
	"os"
	"path/filepath"
	"strings"

	kubekeyapiv1alpha2 "bytetrade.io/web3os/installer/apis/kubekey/v1alpha2"
	"bytetrade.io/web3os/installer/pkg/common"
	"github.com/pkg/errors"
	"helm.sh/helm/v3/pkg/cli"
	"helm.sh/helm/v3/pkg/getter"
)

func InstallAddons(kubeConf *common.KubeConf, addon *kubekeyapiv1alpha2.Addon, kubeConfig string) error {
	// install chart
	if addon.Sources.Chart.Name != "" {
		_ = os.Setenv("HELM_NAMESPACE", strings.TrimSpace(addon.Namespace))
		if err := InstallChart(kubeConf, addon, kubeConfig); err != nil {
			return err
		}
	}

	// install yaml
	if len(addon.Sources.Yaml.Path) != 0 {
		var settings = cli.New()
		p := getter.All(settings)
		for _, yaml := range addon.Sources.Yaml.Path {
			u, _ := url.Parse(yaml)
			_, err := p.ByScheme(u.Scheme)
			if err != nil {
				fp, err := filepath.Abs(yaml)
				if err != nil {
					return errors.Wrap(err, "Failed to look up current directory")
				}
				yamlPaths := []string{fp}
				if err := InstallYaml(yamlPaths, addon.Namespace, kubeConfig, kubeConf.Cluster.Kubernetes.Version); err != nil {
					return err
				}
			} else {
				yamlPaths := []string{yaml}
				if err := InstallYaml(yamlPaths, addon.Namespace, kubeConfig, kubeConf.Cluster.Kubernetes.Version); err != nil {
					return err
				}
			}
		}
	}
	return nil
}
