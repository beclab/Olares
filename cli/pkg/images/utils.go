/*
 Copyright 2022 The KubeSphere Authors.

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

package images

import (
	"strings"

	kubekeyv1alpha2 "bytetrade.io/web3os/installer/apis/kubekey/v1alpha2"
	"bytetrade.io/web3os/installer/pkg/common"
	"bytetrade.io/web3os/installer/pkg/core/connector"
	versionutil "k8s.io/apimachinery/pkg/util/version"
)

var k8sImageRepositoryAddr = "registry.k8s.io"

// GetImage defines the list of all images and gets image object by name.
func GetImage(runtime connector.ModuleRuntime, kubeConf *common.KubeConf, name string) Image {
	var image Image
	pauseTag, corednsTag := "3.2", "1.6.9"

	if versionutil.MustParseSemantic(kubeConf.Cluster.Kubernetes.Version).LessThan(versionutil.MustParseSemantic("v1.21.0")) {
		pauseTag = "3.2"
		corednsTag = "1.6.9"
	}
	if versionutil.MustParseSemantic(kubeConf.Cluster.Kubernetes.Version).AtLeast(versionutil.MustParseSemantic("v1.21.0")) ||
		(kubeConf.Cluster.Kubernetes.ContainerManager != "" && kubeConf.Cluster.Kubernetes.ContainerManager != "docker") {
		pauseTag = "3.4.1"
		corednsTag = "1.8.0"
	}
	if versionutil.MustParseSemantic(kubeConf.Cluster.Kubernetes.Version).AtLeast(versionutil.MustParseSemantic("v1.22.0")) {
		pauseTag = "3.5"
		corednsTag = "1.8.0"
	}
	if versionutil.MustParseSemantic(kubeConf.Cluster.Kubernetes.Version).AtLeast(versionutil.MustParseSemantic("v1.23.0")) {
		pauseTag = "3.6"
		corednsTag = "1.8.6"
	}
	if versionutil.MustParseSemantic(kubeConf.Cluster.Kubernetes.Version).AtLeast(versionutil.MustParseSemantic("v1.24.0")) {
		pauseTag = "3.7"
		corednsTag = "1.8.6"
	}
	if versionutil.MustParseSemantic(kubeConf.Cluster.Kubernetes.Version).AtLeast(versionutil.MustParseSemantic("v1.31.0")) {
		pauseTag = "3.10"
		corednsTag = "1.11.3"
	}

	// logger.Debugf("pauseTag: %s, corednsTag: %s", pauseTag, corednsTag)

	ImageList := map[string]Image{
		"pause":                   {RepoAddr: k8sImageRepositoryAddr, Repo: "pause", Tag: pauseTag, Group: kubekeyv1alpha2.K8s, Enable: true},
		"etcd":                    {RepoAddr: kubeConf.Cluster.Registry.PrivateRegistry, Namespace: kubekeyv1alpha2.DefaultKubeImageNamespace, Repo: "etcd", Tag: kubekeyv1alpha2.DefaultEtcdVersion, Group: kubekeyv1alpha2.Master, Enable: strings.EqualFold(kubeConf.Cluster.Etcd.Type, kubekeyv1alpha2.Kubeadm)},
		"kube-apiserver":          {RepoAddr: kubeConf.Cluster.Registry.PrivateRegistry, Namespace: kubekeyv1alpha2.DefaultKubeImageNamespace, Repo: "kube-apiserver", Tag: kubeConf.Cluster.Kubernetes.Version, Group: kubekeyv1alpha2.Master, Enable: true},
		"kube-controller-manager": {RepoAddr: kubeConf.Cluster.Registry.PrivateRegistry, Namespace: kubekeyv1alpha2.DefaultKubeImageNamespace, Repo: "kube-controller-manager", Tag: kubeConf.Cluster.Kubernetes.Version, Group: kubekeyv1alpha2.Master, Enable: true},
		"kube-scheduler":          {RepoAddr: kubeConf.Cluster.Registry.PrivateRegistry, Namespace: kubekeyv1alpha2.DefaultKubeImageNamespace, Repo: "kube-scheduler", Tag: kubeConf.Cluster.Kubernetes.Version, Group: kubekeyv1alpha2.Master, Enable: true},
		"kube-proxy":              {RepoAddr: kubeConf.Cluster.Registry.PrivateRegistry, Namespace: kubekeyv1alpha2.DefaultKubeImageNamespace, Repo: "kube-proxy", Tag: kubeConf.Cluster.Kubernetes.Version, Group: kubekeyv1alpha2.K8s, Enable: !kubeConf.Cluster.Kubernetes.DisableKubeProxy},

		// network
		"coredns":                 {RepoAddr: kubeConf.Cluster.Registry.PrivateRegistry, Namespace: "coredns", Repo: "coredns", Tag: corednsTag, Group: kubekeyv1alpha2.K8s, Enable: true},
		"k8s-dns-node-cache":      {RepoAddr: kubeConf.Cluster.Registry.PrivateRegistry, Namespace: kubekeyv1alpha2.DefaultKubeImageNamespace, Repo: "k8s-dns-node-cache", Tag: "1.15.12", Group: kubekeyv1alpha2.K8s, Enable: kubeConf.Cluster.Kubernetes.EnableNodelocaldns()},
		"calico-kube-controllers": {RepoAddr: kubeConf.Cluster.Registry.PrivateRegistry, Namespace: "calico", Repo: "kube-controllers", Tag: kubekeyv1alpha2.DefaultCalicoVersion, Group: kubekeyv1alpha2.K8s, Enable: strings.EqualFold(kubeConf.Cluster.Network.Plugin, "calico")},
		"calico-cni":              {RepoAddr: kubeConf.Cluster.Registry.PrivateRegistry, Namespace: "calico", Repo: "cni", Tag: kubekeyv1alpha2.DefaultCalicoVersion, Group: kubekeyv1alpha2.K8s, Enable: strings.EqualFold(kubeConf.Cluster.Network.Plugin, "calico")},
		"calico-node":             {RepoAddr: kubeConf.Cluster.Registry.PrivateRegistry, Namespace: "calico", Repo: "node", Tag: kubekeyv1alpha2.DefaultCalicoVersion, Group: kubekeyv1alpha2.K8s, Enable: strings.EqualFold(kubeConf.Cluster.Network.Plugin, "calico")},
		"calico-flexvol":          {RepoAddr: kubeConf.Cluster.Registry.PrivateRegistry, Namespace: "calico", Repo: "pod2daemon-flexvol", Tag: kubekeyv1alpha2.DefaultCalicoVersion, Group: kubekeyv1alpha2.K8s, Enable: strings.EqualFold(kubeConf.Cluster.Network.Plugin, "calico")},
		"calico-typha":            {RepoAddr: kubeConf.Cluster.Registry.PrivateRegistry, Namespace: "calico", Repo: "typha", Tag: kubekeyv1alpha2.DefaultCalicoVersion, Group: kubekeyv1alpha2.K8s, Enable: strings.EqualFold(kubeConf.Cluster.Network.Plugin, "calico") && len(runtime.GetHostsByRole(common.K8s)) > 50},
		"flannel":                 {RepoAddr: kubeConf.Cluster.Registry.PrivateRegistry, Namespace: kubekeyv1alpha2.DefaultKubeImageNamespace, Repo: "flannel", Tag: kubekeyv1alpha2.DefaultFlannelVersion, Group: kubekeyv1alpha2.K8s, Enable: strings.EqualFold(kubeConf.Cluster.Network.Plugin, "flannel")},
		"cilium":                  {RepoAddr: kubeConf.Cluster.Registry.PrivateRegistry, Namespace: "cilium", Repo: "cilium", Tag: kubekeyv1alpha2.DefaultCiliumVersion, Group: kubekeyv1alpha2.K8s, Enable: strings.EqualFold(kubeConf.Cluster.Network.Plugin, "cilium")},
		"cilium-operator-generic": {RepoAddr: kubeConf.Cluster.Registry.PrivateRegistry, Namespace: "cilium", Repo: "operator-generic", Tag: kubekeyv1alpha2.DefaultCiliumVersion, Group: kubekeyv1alpha2.K8s, Enable: strings.EqualFold(kubeConf.Cluster.Network.Plugin, "cilium")},
		"kubeovn":                 {RepoAddr: kubeConf.Cluster.Registry.PrivateRegistry, Namespace: "kubeovn", Repo: "kube-ovn", Tag: kubekeyv1alpha2.DefaultKubeovnVersion, Group: kubekeyv1alpha2.K8s, Enable: strings.EqualFold(kubeConf.Cluster.Network.Plugin, "kubeovn")},
		"multus":                  {RepoAddr: kubeConf.Cluster.Registry.PrivateRegistry, Namespace: kubekeyv1alpha2.DefaultKubeImageNamespace, Repo: "multus-cni", Tag: kubekeyv1alpha2.DefalutMultusVersion, Group: kubekeyv1alpha2.K8s, Enable: strings.Contains(kubeConf.Cluster.Network.Plugin, "multus")},
		// storage
		"provisioner-localpv": {RepoAddr: kubeConf.Cluster.Registry.PrivateRegistry, Namespace: "openebs", Repo: "provisioner-localpv", Tag: "3.3.0", Group: kubekeyv1alpha2.Worker, Enable: false},
		"linux-utils":         {RepoAddr: kubeConf.Cluster.Registry.PrivateRegistry, Namespace: "openebs", Repo: "linux-utils", Tag: "3.3.0", Group: kubekeyv1alpha2.Worker, Enable: false},
		// load balancer
		"haproxy": {RepoAddr: kubeConf.Cluster.Registry.PrivateRegistry, Namespace: "library", Repo: "haproxy", Tag: "2.3", Group: kubekeyv1alpha2.Worker, Enable: kubeConf.Cluster.ControlPlaneEndpoint.IsInternalLBEnabled()},
		"kubevip": {RepoAddr: kubeConf.Cluster.Registry.PrivateRegistry, Namespace: "plndr", Repo: "kube-vip", Tag: "v0.5.0", Group: kubekeyv1alpha2.Master, Enable: kubeConf.Cluster.ControlPlaneEndpoint.IsInternalLBEnabledVip()},
		// kata-deploy
		"kata-deploy": {RepoAddr: kubeConf.Cluster.Registry.PrivateRegistry, Namespace: kubekeyv1alpha2.DefaultKubeImageNamespace, Repo: "kata-deploy", Tag: "stable", Group: kubekeyv1alpha2.Worker, Enable: kubeConf.Cluster.Kubernetes.EnableKataDeploy()},
		// node-feature-discovery
		"node-feature-discovery": {RepoAddr: kubeConf.Cluster.Registry.PrivateRegistry, Namespace: kubekeyv1alpha2.DefaultKubeImageNamespace, Repo: "node-feature-discovery", Tag: "v0.10.0", Group: kubekeyv1alpha2.K8s, Enable: kubeConf.Cluster.Kubernetes.EnableNodeFeatureDiscovery()},
	}

	image = ImageList[name]
	if kubeConf.Cluster.Registry.NamespaceOverride != "" {
		image.NamespaceOverride = kubeConf.Cluster.Registry.NamespaceOverride
	}
	return image
}
