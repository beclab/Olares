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

import (
	"fmt"
	"os"
	"strings"

	"bytetrade.io/web3os/installer/pkg/core/util"
)

const (
	DefaultPreDir              = "kubekey"
	DefaultTmpDir              = "/tmp/kubekey"
	DefaultSSHPort             = 22
	DefaultLBPort              = 6443
	DefaultLBDomain            = "lb.kubesphere.local"
	DefaultNetworkPlugin       = "calico"
	DefaultPodsCIDR            = "10.233.64.0/18"
	DefaultServiceCIDR         = "10.233.0.0/18"
	DefaultKubeImageNamespace  = "kubesphere"
	DefaultClusterName         = "cluster.local"
	DefaultArch                = "amd64"
	DefaultEtcdVersion         = "v3.4.13"
	DefaultEtcdPort            = "2379"
	DefaultDockerVersion       = "20.10.8"
	DefaultCrictlVersion       = "v1.22.0"
	DefaultKubeVersion         = "v1.21.5"
	DefaultCalicoVersion       = "v3.20.0"
	DefaultFlannelVersion      = "v0.12.0"
	DefaultCniVersion          = "v0.9.1"
	DefaultCiliumVersion       = "v1.8.3"
	DefaultKubeovnVersion      = "v1.5.0"
	DefaultHelmVersion         = "v3.6.3"
	DefaultMaxPods             = 200
	DefaultNodeCidrMaskSize    = 24
	DefaultIPIPMode            = "Always"
	DefaultVXLANMode           = "Never"
	DefaultVethMTU             = 1440
	DefaultBackendMode         = "vxlan"
	DefaultProxyMode           = "ipvs"
	DefaultCrioEndpoint        = "unix:///var/run/crio/crio.sock"
	DefaultContainerdEndpoint  = "unix:///run/containerd/containerd.sock"
	DefaultIsulaEndpoint       = "unix:///var/run/isulad.sock"
	Etcd                       = "etcd"
	Master                     = "master"
	Worker                     = "worker"
	K8s                        = "k8s"
	DefaultEtcdBackupDir       = "/var/backups/kube_etcd"
	DefaultEtcdBackupPeriod    = 30
	DefaultKeepBackNumber      = 5
	DefaultEtcdBackupScriptDir = "/usr/local/bin/kube-scripts"
	DefaultJoinCIDR            = "100.64.0.0/16"
	DefaultNetworkType         = "geneve"
	DefaultVlanID              = "100"
	DefaultOvnLabel            = "node-role.kubernetes.io/control-plane"
	DefaultDPDKVersion         = "19.11"
	DefaultDNSAddress          = "114.114.114.114"

	Docker     = "docker"
	Containerd = "containerd"
	Crio       = "crio"
	Isula      = "isula"
)

func (cfg *ClusterSpec) SetDefaultClusterSpec(incluster bool) (*ClusterSpec, *HostGroups, error) {
	clusterCfg := ClusterSpec{}

	clusterCfg.Hosts = SetDefaultHostsCfg(cfg)
	clusterCfg.RoleGroups = cfg.RoleGroups
	hostGroups, err := clusterCfg.GroupHosts()
	if err != nil {
		return nil, nil, err
	}
	clusterCfg.ControlPlaneEndpoint = SetDefaultLBCfg(cfg, hostGroups.Master, incluster)
	clusterCfg.Network = SetDefaultNetworkCfg(cfg)
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
	if cfg.Kubernetes.MaxPods == 0 {
		clusterCfg.Kubernetes.MaxPods = DefaultMaxPods
	}
	if cfg.Kubernetes.NodeCidrMaskSize == 0 {
		clusterCfg.Kubernetes.NodeCidrMaskSize = DefaultNodeCidrMaskSize
	}
	if cfg.Kubernetes.ProxyMode == "" {
		clusterCfg.Kubernetes.ProxyMode = DefaultProxyMode
	}
	return &clusterCfg, hostGroups, nil
}

func SetDefaultHostsCfg(cfg *ClusterSpec) []HostCfg {
	var hostscfg []HostCfg
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
		hostscfg = append(hostscfg, host)
	}
	return hostscfg
}

func SetDefaultLBCfg(cfg *ClusterSpec, masterGroup []HostCfg, incluster bool) ControlPlaneEndpoint {
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
	if cfg.Network.Kubeovn.JoinCIDR == "" {
		cfg.Network.Kubeovn.JoinCIDR = DefaultJoinCIDR
	}
	if cfg.Network.Kubeovn.Label == "" {
		cfg.Network.Kubeovn.Label = DefaultOvnLabel
	}
	if cfg.Network.Kubeovn.VlanID == "" {
		cfg.Network.Kubeovn.VlanID = DefaultVlanID
	}
	if cfg.Network.Kubeovn.NetworkType == "" {
		cfg.Network.Kubeovn.NetworkType = DefaultNetworkType
	}
	if cfg.Network.Kubeovn.PingerExternalAddress == "" {
		cfg.Network.Kubeovn.PingerExternalAddress = DefaultDNSAddress
	}
	if cfg.Network.Kubeovn.DpdkVersion == "" {
		cfg.Network.Kubeovn.DpdkVersion = DefaultDPDKVersion
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
	if cfg.Kubernetes.ClusterName == "" {
		cfg.Kubernetes.ClusterName = DefaultClusterName
	}
	if cfg.Kubernetes.EtcdBackupDir == "" {
		cfg.Kubernetes.EtcdBackupDir = DefaultEtcdBackupDir
	}
	if cfg.Kubernetes.EtcdBackupPeriod == 0 {
		cfg.Kubernetes.EtcdBackupPeriod = DefaultEtcdBackupPeriod
	}
	if cfg.Kubernetes.KeepBackupNumber == 0 {
		cfg.Kubernetes.KeepBackupNumber = DefaultKeepBackNumber
	}
	if cfg.Kubernetes.EtcdBackupScriptDir == "" {
		cfg.Kubernetes.EtcdBackupScriptDir = DefaultEtcdBackupScriptDir
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
