/*
Copyright 2020 The KubeSphere Authors.

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

package v1alpha1

type NetworkConfig struct {
	Plugin          string     `yaml:"plugin" json:"plugin,omitempty"`
	KubePodsCIDR    string     `yaml:"kubePodsCIDR" json:"kubePodsCIDR,omitempty"`
	KubeServiceCIDR string     `yaml:"kubeServiceCIDR" json:"kubeServiceCIDR,omitempty"`
	Calico          CalicoCfg  `yaml:"calico" json:"calico,omitempty"`
	Flannel         FlannelCfg `yaml:"flannel" json:"flannel,omitempty"`
	Kubeovn         KubeovnCfg `yaml:"kubeovn" json:"kubeovn,omitempty"`
}

type CalicoCfg struct {
	IPIPMode  string `yaml:"ipipMode" json:"ipipMode,omitempty"`
	VXLANMode string `yaml:"vxlanMode" json:"vxlanMode,omitempty"`
	VethMTU   int    `yaml:"vethMTU" json:"vethMTU,omitempty"`
}

type FlannelCfg struct {
	BackendMode   string `yaml:"backendMode" json:"backendMode,omitempty"`
	Directrouting bool   `yaml:"directRouting" json:"directRouting,omitempty"`
}

type KubeovnCfg struct {
	JoinCIDR              string `yaml:"joinCIDR" json:"joinCIDR,omitempty"`
	NetworkType           string `yaml:"networkType" json:"networkType,omitempty"`
	Label                 string `yaml:"label" json:"label,omitempty"`
	Iface                 string `yaml:"iface" json:"iface,omitempty"`
	VlanInterfaceName     string `yaml:"vlanInterfaceName" json:"vlanInterfaceName,omitempty"`
	VlanID                string `yaml:"vlanID" json:"vlanID,omitempty"`
	DpdkMode              bool   `yaml:"dpdkMode" json:"dpdkMode,omitempty"`
	EnableSSL             bool   `yaml:"enableSSL" json:"enableSSL,omitempty"`
	EnableMirror          bool   `yaml:"enableMirror" json:"enableMirror,omitempty"`
	HwOffload             bool   `yaml:"hwOffload" json:"hwOffload,omitempty"`
	DpdkVersion           string `yaml:"dpdkVersion" json:"dpdkVersion,omitempty"`
	PingerExternalAddress string `yaml:"pingerExternalAddress" json:"pingerExternalAddress,omitempty"`
	PingerExternalDomain  string `yaml:"pingerExternalDomain" json:"pingerExternalDomain,omitempty"`
}
