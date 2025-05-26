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

package v1alpha2

type Addon struct {
	Name      string  `yaml:"name" json:"name,omitempty"`
	Namespace string  `yaml:"namespace" json:"namespace,omitempty"`
	Sources   Sources `yaml:"sources" json:"sources,omitempty"`
	Retries   int     `yaml:"retries" json:"retries,omitempty"`
	Delay     int     `yaml:"delay" json:"delay,omitempty"`
}

type Sources struct {
	Chart Chart `yaml:"chart" json:"chart,omitempty"`
	Yaml  Yaml  `yaml:"yaml" json:"yaml,omitempty"`
}

type Chart struct {
	Name       string   `yaml:"name" json:"name,omitempty"`
	Repo       string   `yaml:"repo" json:"repo,omitempty"`
	Path       string   `yaml:"path" json:"path,omitempty"`
	Version    string   `yaml:"version" json:"version,omitempty"`
	ValuesFile string   `yaml:"valuesFile" json:"valuesFile,omitempty"`
	Values     []string `yaml:"values" json:"values,omitempty"`
}

type Yaml struct {
	Path []string `yaml:"path" json:"path,omitempty"`
}
