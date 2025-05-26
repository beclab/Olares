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

type NetworkConfig struct {
	Plugin          string     `yaml:"plugin" json:"plugin,omitempty"`
	KubePodsCIDR    string     `yaml:"kubePodsCIDR" json:"kubePodsCIDR,omitempty"`
	KubeServiceCIDR string     `yaml:"kubeServiceCIDR" json:"kubeServiceCIDR,omitempty"`
	Calico          CalicoCfg  `yaml:"calico" json:"calico,omitempty"`
	Flannel         FlannelCfg `yaml:"flannel" json:"flannel,omitempty"`
	Kubeovn         KubeovnCfg `yaml:"kubeovn" json:"kubeovn,omitempty"`
	MultusCNI       MultusCNI  `yaml:"multusCNI" json:"multusCNI,omitempty"`
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
	EnableSSL             bool              `yaml:"enableSSL" json:"enableSSL,omitempty"`
	JoinCIDR              string            `yaml:"joinCIDR" json:"joinCIDR,omitempty"`
	Label                 string            `yaml:"label" json:"label,omitempty"`
	TunnelType            string            `yaml:"tunnelType" json:"tunnelType,omitempty"`
	SvcYamlIpfamilypolicy string            `yaml:"svcYamlIpfamilypolicy" json:"svcYamlIpfamilypolicy,omitempty"`
	Dpdk                  Dpdk              `yaml:"dpdk" json:"dpdk,omitempty"`
	OvsOvn                OvsOvn            `yaml:"ovs-ovn" json:"ovs-ovn,omitempty"`
	KubeOvnController     KubeOvnController `yaml:"kube-ovn-controller" json:"kube-ovn-controller,omitempty"`
	KubeOvnCni            KubeOvnCni        `yaml:"kube-ovn-cni" json:"kube-ovn-cni,omitempty"`
	KubeOvnPinger         KubeOvnPinger     `yaml:"kube-ovn-pinger" json:"kube-ovn-pinger,omitempty"`
}

type Dpdk struct {
	DpdkMode        bool   `yaml:"dpdkMode" json:"dpdkMode,omitempty"`
	DpdkTunnelIface string `yaml:"dpdkTunnelIface" json:"dpdkTunnelIface,omitempty"`
	DpdkVersion     string `yaml:"dpdkVersion" json:"dpdkVersion,omitempty"`
}

type OvsOvn struct {
	HwOffload bool `yaml:"hwOffload" json:"hwOffload,omitempty"`
}

type KubeOvnController struct {
	PodGateway        string `yaml:"podGateway" json:"podGateway,omitempty"`
	CheckGateway      *bool  `yaml:"checkGateway" json:"checkGateway,omitempty"`
	LogicalGateway    bool   `yaml:"logicalGateway" json:"logicalGateway,omitempty"`
	ExcludeIps        string `yaml:"excludeIps" json:"excludeIps,omitempty"`
	NetworkType       string `yaml:"networkType" json:"networkType,omitempty"`
	VlanInterfaceName string `yaml:"vlanInterfaceName" json:"vlanInterfaceName,omitempty"`
	VlanID            string `yaml:"vlanID" json:"vlanID,omitempty"`
	PodNicType        string `yaml:"podNicType" json:"podNicType,omitempty"`
	EnableLB          *bool  `yaml:"enableLB" json:"enableLB,omitempty"`
	EnableNP          *bool  `yaml:"enableNP" json:"enableNP,omitempty"`
	EnableEipSnat     *bool  `yaml:"enableEipSnat" json:"enableEipSnat,omitempty"`
	EnableExternalVPC *bool  `yaml:"enableExternalVPC" json:"enableExternalVPC,omitempty"`
}

type KubeOvnCni struct {
	EnableMirror      bool   `yaml:"enableMirror" json:"enableMirror,omitempty"`
	Iface             string `yaml:"iface" json:"iface,omitempty"`
	CNIConfigPriority string `yaml:"CNIConfigPriority" json:"CNIConfigPriority,omitempty"`
	Modules           string `yaml:"modules" json:"modules,omitempty"`
	RPMs              string `yaml:"RPMs" json:"RPMs,omitempty"`
}

type KubeOvnPinger struct {
	PingerExternalAddress string `yaml:"pingerExternalAddress" json:"pingerExternalAddress,omitempty"`
	PingerExternalDomain  string `yaml:"pingerExternalDomain" json:"pingerExternalDomain,omitempty"`
}

func (k *KubeovnCfg) KubeovnCheckGateway() bool {
	if k.KubeOvnController.CheckGateway == nil {
		return true
	}
	return *k.KubeOvnController.CheckGateway
}

func (k *KubeovnCfg) KubeovnEnableLB() bool {
	if k.KubeOvnController.EnableLB == nil {
		return true
	}
	return *k.KubeOvnController.EnableLB
}

func (k *KubeovnCfg) KubeovnEnableNP() bool {
	if k.KubeOvnController.EnableNP == nil {
		return true
	}
	return *k.KubeOvnController.EnableNP
}

func (k *KubeovnCfg) KubeovnEnableEipSnat() bool {
	if k.KubeOvnController.EnableEipSnat == nil {
		return true
	}
	return *k.KubeOvnController.EnableEipSnat
}

func (k *KubeovnCfg) KubeovnEnableExternalVPC() bool {
	if k.KubeOvnController.EnableExternalVPC == nil {
		return true
	}
	return *k.KubeOvnController.EnableExternalVPC
}

type MultusCNI struct {
	Enabled *bool `yaml:"enabled" json:"enabled,omitempty"`
}

func (n *NetworkConfig) EnableMultusCNI() bool {
	if n.MultusCNI.Enabled == nil {
		return false
	}
	return *n.MultusCNI.Enabled
}
