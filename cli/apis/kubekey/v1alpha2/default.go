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

import (
	"fmt"
	"os"
	"strings"

	"bytetrade.io/web3os/installer/pkg/core/util"
)

const (
	DefaultPreDir                   = "kubekey"
	DefaultTmpDir                   = "/tmp/kubekey"
	DefaultSSHPort                  = 22
	DefaultLBPort                   = 6443
	DefaultApiserverPort            = 6443
	DefaultLBDomain                 = "lb.kubesphere.local"
	DefaultNetworkPlugin            = "calico"
	DefaultPodsCIDR                 = "10.233.64.0/18"
	DefaultServiceCIDR              = "10.233.0.0/18"
	DefaultKubeImageNamespace       = "kubesphere"
	DefaultClusterName              = "cluster.local"
	DefaultDNSDomain                = "cluster.local"
	DefaultArch                     = "amd64"
	DefaultSSHTimeout               = 30
	DefaultEtcdVersion              = "v3.4.13"
	DefaultEtcdPort                 = "2379"
	DefaultDockerVersion            = "20.10.8"
	DefaultContainerdVersion        = "1.6.4"
	DefaultRuncVersion              = "v1.1.1"
	DefaultRuncVersion_v_1_1_4      = "v1.1.4"
	DefaultCrictlVersion            = "v1.24.0"
	DefaultKubeVersion              = "v1.23.10"
	DefaultCalicoVersion            = "v3.29.2"
	DefaultFlannelVersion           = "v0.12.0"
	DefaultCniVersion               = "v0.9.1"
	DefaultCniVersion_v_1_1_1       = "v1.1.1"
	DefaultCiliumVersion            = "v1.11.6"
	DefaultKubeovnVersion           = "v1.10.6"
	DefalutMultusVersion            = "v3.8"
	DefaultHelmVersion              = "v3.9.0"
	DefaultDockerComposeVersion     = "v2.2.2"
	DefaultRegistryVersion          = "2"
	DefaultHarborVersion            = "v2.5.3"
	DefaultUbuntu24AppArmonVersion  = "4.0.1"
	DefaultSocatVersion             = "1.7.3.4"
	DefaultFlexVersion              = "2.6.4"
	DefaultConntrackVersion         = "1.4.1"
	DefaultOssUtilVersion           = "v1.7.18"
	DefaultCosUtilVersion           = "v1.0.2"
	DefaultMinioVersion             = "RELEASE.2023-05-04T21-44-30Z"
	DefaultMinioOperatorVersion     = "0.0.1"
	DefaultRedisVersion             = "5.0.14"
	DefaultJuiceFsVersion           = "v11.1.0"
	CudaKeyringVersion1_0           = "1.0"
	CudaKeyringVersion1_1           = "1.1"
	DefaultVeleroVersion            = "v1.11.3"
	DefaultWSLInstallPackageVersion = "2.3.26.0"

	DefaultShutdownGracePeriod             = "30s"
	DefaultShutdownGracePeriodCriticalPods = "10s"
	DefaultMaxPods                         = 200
	DefaultPodPidsLimit                    = 10000
	DefaultNodeCidrMaskSize                = 24
	DefaultIPIPMode                        = "Always"
	DefaultVXLANMode                       = "Never"
	DefaultVethMTU                         = 0
	DefaultBackendMode                     = "vxlan"
	DefaultProxyMode                       = "ipvs"
	DefaultCrioEndpoint                    = "unix:///var/run/crio/crio.sock"
	DefaultContainerdEndpoint              = "unix:///run/containerd/containerd.sock"
	DefaultIsulaEndpoint                   = "unix:///var/run/isulad.sock"
	Etcd                                   = "etcd"
	Master                                 = "master"
	ControlPlane                           = "control-plane"
	Worker                                 = "worker"
	K8s                                    = "k8s"
	Registry                               = "registry"
	DefaultEtcdBackupDir                   = "/var/backups/kube_etcd"
	DefaultEtcdBackupPeriod                = 30
	DefaultKeepBackNumber                  = 5
	DefaultEtcdBackupScriptDir             = "/usr/local/bin/kube-scripts"
	DefaultPodGateway                      = "10.233.64.1"
	DefaultJoinCIDR                        = "100.64.0.0/16"
	DefaultNetworkType                     = "geneve"
	DefaultTunnelType                      = "geneve"
	DefaultPodNicType                      = "veth-pair"
	DefaultModules                         = "kube_ovn_fastpath.ko"
	DefaultRPMs                            = "openvswitch-kmod"
	DefaultVlanID                          = "100"
	DefaultOvnLabel                        = "node-role.kubernetes.io/control-plane"
	DefaultDPDKVersion                     = "19.11"
	DefaultDNSAddress                      = "114.114.114.114"
	DefaultDpdkTunnelIface                 = "br-phy"
	DefaultCNIConfigPriority               = "01"

	Docker     = "docker"
	Containerd = "containerd"
	Crio       = "crio"
	Isula      = "isula"

	Haproxy            = "haproxy"
	Kubevip            = "kube-vip"
	DefaultKubeVipMode = "ARP"
)

func (cfg *ClusterSpec) SetDefaultClusterSpec(incluster bool, macos bool) (*ClusterSpec, map[string][]*KubeHost) {
	clusterCfg := ClusterSpec{}

	clusterCfg.Hosts = SetDefaultHostsCfg(cfg)
	clusterCfg.RoleGroups = cfg.RoleGroups
	clusterCfg.Etcd = SetDefaultEtcdCfg(cfg, macos)
	roleGroups := clusterCfg.GroupHosts()
	clusterCfg.ControlPlaneEndpoint = SetDefaultLBCfg(cfg, roleGroups[Master], incluster)
	clusterCfg.Network = SetDefaultNetworkCfg(cfg)
	clusterCfg.System = cfg.System
	clusterCfg.Kubernetes = SetDefaultClusterCfg(cfg)
	clusterCfg.Registry = cfg.Registry
	clusterCfg.Addons = cfg.Addons
	clusterCfg.KubeSphere = cfg.KubeSphere

	if cfg.Kubernetes.ClusterName == "" {
		clusterCfg.Kubernetes.ClusterName = DefaultClusterName
	}
	if cfg.Kubernetes.Version == "" {
		clusterCfg.Kubernetes.Version = DefaultKubeVersion
	}
	if cfg.Kubernetes.ShutdownGracePeriod == "" {
		clusterCfg.Kubernetes.ShutdownGracePeriod = DefaultShutdownGracePeriod
	}
	if cfg.Kubernetes.ShutdownGracePeriodCriticalPods == "" {
		clusterCfg.Kubernetes.ShutdownGracePeriodCriticalPods = DefaultShutdownGracePeriodCriticalPods
	}
	if cfg.Kubernetes.MaxPods == 0 {
		clusterCfg.Kubernetes.MaxPods = DefaultMaxPods
	}
	if cfg.Kubernetes.PodPidsLimit == 0 {
		clusterCfg.Kubernetes.PodPidsLimit = DefaultPodPidsLimit
	}
	if cfg.Kubernetes.NodeCidrMaskSize == 0 {
		clusterCfg.Kubernetes.NodeCidrMaskSize = DefaultNodeCidrMaskSize
	}
	if cfg.Kubernetes.ProxyMode == "" {
		clusterCfg.Kubernetes.ProxyMode = DefaultProxyMode
	}
	return &clusterCfg, roleGroups
}

func SetDefaultHostsCfg(cfg *ClusterSpec) []HostCfg {
	var hostCfg []HostCfg
	if len(cfg.Hosts) == 0 {
		return nil
	}
	for _, host := range cfg.Hosts {
		if len(host.Address) == 0 && len(host.InternalAddress) > 0 {
			host.Address = host.InternalAddress
		}
		if len(host.InternalAddress) == 0 && len(host.Address) > 0 {
			host.InternalAddress = host.Address
		}
		if host.User == "" {
			host.User = "root"
		}
		if host.Port == 0 {
			host.Port = DefaultSSHPort
		}
		if host.PrivateKey == "" {
			if host.Password == "" && host.PrivateKeyPath == "" {
				host.PrivateKeyPath = "~/.ssh/id_rsa"
			}
			if host.PrivateKeyPath != "" && strings.HasPrefix(strings.TrimSpace(host.PrivateKeyPath), "~/") {
				homeDir, _ := util.Home()
				host.PrivateKeyPath = strings.Replace(host.PrivateKeyPath, "~/", fmt.Sprintf("%s/", homeDir), 1)
			}
		}

		if host.Arch == "" {
			host.Arch = DefaultArch
		}

		if host.Timeout == nil {
			var timeout int64
			timeout = DefaultSSHTimeout
			host.Timeout = &timeout
		}

		hostCfg = append(hostCfg, host)
	}
	return hostCfg
}

func SetDefaultLBCfg(cfg *ClusterSpec, masterGroup []*KubeHost, incluster bool) ControlPlaneEndpoint {
	if !incluster {
		//The detection is not an HA environment, and the address at LB does not need input
		if len(masterGroup) == 1 && cfg.ControlPlaneEndpoint.Address != "" {
			fmt.Println("When the environment is not HA, the LB address does not need to be entered, so delete the corresponding value.")
			os.Exit(0)
		}

		//Check whether LB should be configured
		if len(masterGroup) >= 3 && !cfg.ControlPlaneEndpoint.IsInternalLBEnabled() && cfg.ControlPlaneEndpoint.Address == "" {
			fmt.Println("When the environment has at least three masters, You must set the value of the LB address or enable the internal loadbalancer.")
			os.Exit(0)
		}

		// Check whether LB address and the internal LB are both enabled
		if cfg.ControlPlaneEndpoint.IsInternalLBEnabled() && cfg.ControlPlaneEndpoint.Address != "" {
			fmt.Println("You cannot set up the internal load balancer and the LB address at the same time.")
			os.Exit(0)
		}
	}

	if cfg.ControlPlaneEndpoint.Address == "" || cfg.ControlPlaneEndpoint.Address == "127.0.0.1" {
		cfg.ControlPlaneEndpoint.Address = masterGroup[0].InternalAddress
	}
	if cfg.ControlPlaneEndpoint.Domain == "" {
		cfg.ControlPlaneEndpoint.Domain = DefaultLBDomain
	}
	if cfg.ControlPlaneEndpoint.Port == 0 {
		cfg.ControlPlaneEndpoint.Port = DefaultLBPort
	}
	if cfg.ControlPlaneEndpoint.KubeVip.Mode == "" {
		cfg.ControlPlaneEndpoint.KubeVip.Mode = DefaultKubeVipMode
	}
	defaultLbCfg := cfg.ControlPlaneEndpoint
	return defaultLbCfg
}

func SetDefaultNetworkCfg(cfg *ClusterSpec) NetworkConfig {
	if cfg.Network.Plugin == "" {
		cfg.Network.Plugin = DefaultNetworkPlugin
	}
	if cfg.Network.KubePodsCIDR == "" {
		cfg.Network.KubePodsCIDR = DefaultPodsCIDR
	}
	if cfg.Network.KubeServiceCIDR == "" {
		cfg.Network.KubeServiceCIDR = DefaultServiceCIDR
	}
	if cfg.Network.Calico.IPIPMode == "" {
		cfg.Network.Calico.IPIPMode = DefaultIPIPMode
	}
	if cfg.Network.Calico.VXLANMode == "" {
		cfg.Network.Calico.VXLANMode = DefaultVXLANMode
	}
	if cfg.Network.Calico.VethMTU == 0 {
		cfg.Network.Calico.VethMTU = DefaultVethMTU
	}
	if cfg.Network.Flannel.BackendMode == "" {
		cfg.Network.Flannel.BackendMode = DefaultBackendMode
	}
	// kube-ovn default config
	if cfg.Network.Kubeovn.KubeOvnController.PodGateway == "" {
		cfg.Network.Kubeovn.KubeOvnController.PodGateway = DefaultPodGateway
	}
	if cfg.Network.Kubeovn.JoinCIDR == "" {
		cfg.Network.Kubeovn.JoinCIDR = DefaultJoinCIDR
	}
	if cfg.Network.Kubeovn.Label == "" {
		cfg.Network.Kubeovn.Label = DefaultOvnLabel
	}
	if cfg.Network.Kubeovn.KubeOvnController.VlanID == "" {
		cfg.Network.Kubeovn.KubeOvnController.VlanID = DefaultVlanID
	}
	if cfg.Network.Kubeovn.KubeOvnController.NetworkType == "" {
		cfg.Network.Kubeovn.KubeOvnController.NetworkType = DefaultNetworkType
	}
	if cfg.Network.Kubeovn.TunnelType == "" {
		cfg.Network.Kubeovn.TunnelType = DefaultTunnelType
	}
	if cfg.Network.Kubeovn.KubeOvnController.PodNicType == "" {
		cfg.Network.Kubeovn.KubeOvnController.PodNicType = DefaultPodNicType
	}
	if cfg.Network.Kubeovn.KubeOvnCni.Modules == "" {
		cfg.Network.Kubeovn.KubeOvnCni.Modules = DefaultModules
	}
	if cfg.Network.Kubeovn.KubeOvnCni.RPMs == "" {
		cfg.Network.Kubeovn.KubeOvnCni.RPMs = DefaultRPMs
	}
	if cfg.Network.Kubeovn.KubeOvnPinger.PingerExternalAddress == "" {
		cfg.Network.Kubeovn.KubeOvnPinger.PingerExternalAddress = DefaultDNSAddress
	}
	if cfg.Network.Kubeovn.Dpdk.DpdkVersion == "" {
		cfg.Network.Kubeovn.Dpdk.DpdkVersion = DefaultDPDKVersion
	}
	if cfg.Network.Kubeovn.Dpdk.DpdkTunnelIface == "" {
		cfg.Network.Kubeovn.Dpdk.DpdkTunnelIface = DefaultDpdkTunnelIface
	}
	if cfg.Network.Kubeovn.KubeOvnCni.CNIConfigPriority == "" {
		cfg.Network.Kubeovn.KubeOvnCni.CNIConfigPriority = DefaultCNIConfigPriority
	}
	defaultNetworkCfg := cfg.Network

	return defaultNetworkCfg
}

func SetDefaultClusterCfg(cfg *ClusterSpec) Kubernetes {
	if cfg.Kubernetes.Version == "" {
		cfg.Kubernetes.Version = DefaultKubeVersion
	} else {
		s := strings.Split(cfg.Kubernetes.Version, "-")
		if len(s) > 1 {
			cfg.Kubernetes.Version = s[0]
			cfg.Kubernetes.Type = s[1]
		}
	}
	if cfg.Kubernetes.Type == "" {
		cfg.Kubernetes.Type = "kubernetes"
	}
	if cfg.Kubernetes.ClusterName == "" {
		cfg.Kubernetes.ClusterName = DefaultClusterName
	}
	if cfg.Kubernetes.DNSDomain == "" {
		cfg.Kubernetes.DNSDomain = DefaultDNSDomain
	}
	if cfg.Kubernetes.ContainerManager == "" {
		cfg.Kubernetes.ContainerManager = Docker
	}
	if cfg.Kubernetes.ContainerRuntimeEndpoint == "" {
		switch cfg.Kubernetes.ContainerManager {
		case Docker:
			cfg.Kubernetes.ContainerRuntimeEndpoint = ""
		case Crio:
			cfg.Kubernetes.ContainerRuntimeEndpoint = DefaultCrioEndpoint
		case Containerd:
			cfg.Kubernetes.ContainerRuntimeEndpoint = DefaultContainerdEndpoint
		case Isula:
			cfg.Kubernetes.ContainerRuntimeEndpoint = DefaultIsulaEndpoint
		default:
			cfg.Kubernetes.ContainerRuntimeEndpoint = ""
		}
	}
	defaultClusterCfg := cfg.Kubernetes

	return defaultClusterCfg
}

func SetDefaultEtcdCfg(cfg *ClusterSpec, macos bool) EtcdCluster {
	if macos {
		cfg.Etcd.Type = MiniKube
	} else if cfg.Etcd.Type == "" || ((cfg.Kubernetes.Type == "k3s" || (len(strings.Split(cfg.Kubernetes.Version, "-")) > 1) && strings.Split(cfg.Kubernetes.Version, "-")[1] == "k3s") && cfg.Etcd.Type == Kubeadm) {
		cfg.Etcd.Type = KubeKey
	}
	if cfg.Etcd.BackupDir == "" {
		cfg.Etcd.BackupDir = DefaultEtcdBackupDir
	}
	if cfg.Etcd.BackupPeriod == 0 {
		cfg.Etcd.BackupPeriod = DefaultEtcdBackupPeriod
	}
	if cfg.Etcd.KeepBackupNumber == 0 {
		cfg.Etcd.KeepBackupNumber = DefaultKeepBackNumber
	}
	if cfg.Etcd.BackupScriptDir == "" {
		cfg.Etcd.BackupScriptDir = DefaultEtcdBackupScriptDir
	}

	return cfg.Etcd
}
