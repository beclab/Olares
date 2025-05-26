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
	"text/template"

	"bytetrade.io/web3os/installer/pkg/utils"
	"github.com/lithammer/dedent"
)

var (
	funcMap = template.FuncMap{"toYaml": utils.ToYAML, "indent": utils.Indent}
	// k3sRegistryConfigTempl defines the template of k3s' registry.
	K3sRegistryConfigTempl = template.Must(template.New("registries.yaml").Funcs(funcMap).Parse(
		dedent.Dedent(`{{ toYaml .Registries }}`)))
)
