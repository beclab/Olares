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
	"fmt"
	"strings"

	kubekeyv1alpha2 "bytetrade.io/web3os/installer/apis/kubekey/v1alpha2"
	"bytetrade.io/web3os/installer/pkg/common"
	"bytetrade.io/web3os/installer/pkg/core/connector"
	"bytetrade.io/web3os/installer/pkg/core/logger"
	"github.com/containerd/containerd/platforms"
	"github.com/containers/image/v5/types"
	manifesttypes "github.com/estesp/manifest-tool/v2/pkg/types"
	ocispec "github.com/opencontainers/image-spec/specs-go/v1"
	versionutil "k8s.io/apimachinery/pkg/util/version"
)

var defaultUserAgent = "kubekey"
var k8sImageRepositoryAddr = "registry.k8s.io"

type dockerImageOptions struct {
	arch           string
	os             string
	variant        string
	username       string
	password       string
	dockerCertPath string
	SkipTLSVerify  bool
}

type ImageInspect struct {
	Status ImageInspectStatus `json:"status"`
}

type ImageInspectStatus struct {
	Id          string   `json:"id"`
	RepoTags    []string `json:"repoTags"`
	RepoDigests []string `json:"repoDigests"`
	Size        string   `json:"size"`
	Uid         struct{} `json:"uid"`
	Username    string   `json:"username"`
	Spec        struct{} `json:"spec"`
}

func (d *dockerImageOptions) systemContext() *types.SystemContext {
	ctx := &types.SystemContext{
		ArchitectureChoice:          d.arch,
		OSChoice:                    d.os,
		VariantChoice:               d.variant,
		DockerRegistryUserAgent:     defaultUserAgent,
		DockerInsecureSkipTLSVerify: types.NewOptionalBool(d.SkipTLSVerify),
	}
	return ctx
}

type srcImageOptions struct {
	dockerImage   dockerImageOptions
	imageName     string
	sharedBlobDir string
}

func (s *srcImageOptions) systemContext() *types.SystemContext {
	ctx := s.dockerImage.systemContext()
	ctx.DockerCertPath = s.dockerImage.dockerCertPath
	ctx.OCISharedBlobDirPath = s.sharedBlobDir
	ctx.DockerAuthConfig = &types.DockerAuthConfig{
		Username: s.dockerImage.username,
		Password: s.dockerImage.password,
	}

	return ctx
}

type destImageOptions struct {
	dockerImage dockerImageOptions
	imageName   string
}

func (d *destImageOptions) systemContext() *types.SystemContext {
	ctx := d.dockerImage.systemContext()
	ctx.DockerCertPath = d.dockerImage.dockerCertPath
	ctx.DockerAuthConfig = &types.DockerAuthConfig{
		Username: d.dockerImage.username,
		Password: d.dockerImage.password,
	}

	return ctx
}

// ParseArchVariant
// Ex:
// amd64 returns amd64, ""
// arm/v8 returns arm, v8
func ParseArchVariant(platform string) (string, string) {
	osArchArr := strings.Split(platform, "/")

	variant := ""
	arch := osArchArr[0]
	if len(osArchArr) > 1 {
		variant = osArchArr[1]
	}
	return arch, variant
}

func ParseImageWithArchTag(ref string) (string, ocispec.Platform) {
	n := strings.LastIndex(ref, "-")
	if n < 0 {
		logger.Fatalf("get arch or variant index failed: %s", ref)
	}
	archOrVariant := ref[n+1:]

	// try to parse the arch-only case
	specifier := fmt.Sprintf("linux/%s", archOrVariant)
	if p, err := platforms.Parse(specifier); err == nil && isKnownArch(p.Architecture) {
		return ref[:n], p
	}

	archStr := ref[:n]
	a := strings.LastIndex(archStr, "-")
	if a < 0 {
		logger.Fatalf("get arch index failed: %s", ref)
	}
	arch := archStr[a+1:]

	// parse the case where both arch and variant exist
	specifier = fmt.Sprintf("linux/%s/%s", arch, archOrVariant)
	p, err := platforms.Parse(specifier)
	if err != nil {
		logger.Fatalf("parse image %s failed: %s", ref, err.Error())
	}

	return ref[:a], p
}

func isKnownArch(arch string) bool {
	switch arch {
	case "386", "amd64", "amd64p32", "arm", "armbe", "arm64", "arm64be", "ppc64", "ppc64le", "loong64", "mips", "mipsle", "mips64", "mips64le", "mips64p32", "mips64p32le", "ppc", "riscv", "riscv64", "s390", "s390x", "sparc", "sparc64", "wasm":
		return true
	}
	return false
}

// ParseImageTag
// Get a repos name and returns the right reposName + tag
// The tag can be confusing because of a port in a repository name.
//
//	Ex: localhost.localdomain:5000/samalba/hipache:latest
func ParseImageTag(repos string) (string, string) {
	n := strings.LastIndex(repos, ":")
	if n < 0 {
		return repos, ""
	}
	if tag := repos[n+1:]; !strings.Contains(tag, "/") {
		return repos[:n], tag
	}
	return repos, ""
}

func NewManifestSpec(image string, entries []manifesttypes.ManifestEntry) manifesttypes.YAMLInput {
	var srcImages []manifesttypes.ManifestEntry

	for _, e := range entries {
		srcImages = append(srcImages, manifesttypes.ManifestEntry{
			Image:    e.Image,
			Platform: e.Platform,
		})
	}

	return manifesttypes.YAMLInput{
		Image:     image,
		Manifests: srcImages,
	}
}

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
