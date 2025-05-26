package files

import (
	"fmt"
	"strings"
)

var (
	EtcdUrl   = "https://github.com/coreos/etcd/releases/download/%s/etcd-%s-linux-%s.tar.gz"
	EtcdUrlCN = "https://kubernetes-release.pek3b.qingstor.com/etcd/release/download/%s/etcd-%s-linux-%s.tar.gz"

	KubeadmUrl   = "https://storage.googleapis.com/kubernetes-release/release/%s/bin/linux/%s/kubeadm"
	KubeadmUrlCN = "https://kubernetes-release.pek3b.qingstor.com/release/%s/bin/linux/%s/kubeadm"

	KubeletUrl   = "https://storage.googleapis.com/kubernetes-release/release/%s/bin/linux/%s/kubelet"
	KubeletUrlCN = "https://kubernetes-release.pek3b.qingstor.com/release/%s/bin/linux/%s/kubelet"

	KubectlUrl   = "https://storage.googleapis.com/kubernetes-release/release/%s/bin/linux/%s/kubectl"
	KubectlUrlCN = "https://kubernetes-release.pek3b.qingstor.com/release/%s/bin/linux/%s/kubectl"

	KubecniUrl   = "https://github.com/containernetworking/plugins/releases/download/%s/cni-plugins-linux-%s-%s.tgz"
	KubecniUrlCN = "https://containernetworking.pek3b.qingstor.com/plugins/releases/download/%s/cni-plugins-linux-%s-%s.tgz"

	HelmUrl   = "https://get.helm.sh/helm-%s-%s-%s.tar.gz"
	HelmUrlCN = "https://kubernetes-helm.pek3b.qingstor.com/linux-%s/%s/helm"

	DockerUrl   = "https://download.docker.com/linux/static/stable/%s/docker-%s.tgz"
	DockerUrlCN = "https://mirrors.aliyun.com/docker-ce/linux/static/stable/%s/docker-%s.tgz"

	CrictlUrl   = "https://github.com/kubernetes-sigs/cri-tools/releases/download/%s/crictl-%s-linux-%s.tar.gz"
	CrictlUrlCN = "https://kubernetes-release.pek3b.qingstor.com/cri-tools/releases/download/%s/crictl-%s-linux-%s.tar.gz"

	K3sUrl    = "https://github.com/k3s-io/k3s/releases/download/%s+k3s1/k3s"
	K3sUrlCN  = "https://kubernetes-release.pek3b.qingstor.com/k3s/releases/download/%s+k3s1/linux/%s/k3s"
	K3sArmUrl = "https://github.com/k3s-io/k3s/releases/download/%s+k3s1/k3s-%s"

	K8eUrl    = "https://github.com/xiaods/k8e/releases/download/%s+k8e2/k8e"
	K8eArmUrl = "https://github.com/xiaods/k8e/releases/download/%s+k8e2/k8e-%s"

	RegistryUrl   = "https://github.com/kubesphere/kubekey/releases/download/v2.0.0-alpha.1/registry-%s-linux-%s.tar.gz"
	RegistryUrlCN = "https://kubernetes-release.pek3b.qingstor.com/registry/%s/registry-%s-linux-%s.tar.gz"

	HarborUrl   = "https://github.com/goharbor/harbor/releases/download/%s/harbor-offline-installer-%s.tgz"
	HarborUrlCN = "https://kubernetes-release.pek3b.qingstor.com/harbor/releases/download/%s/harbor-offline-installer-%s.tgz"

	ComposeUrl   = "https://github.com/docker/compose/releases/download/%s/docker-compose-linux-x86_64"
	ComposeUrlCN = "https://kubernetes-release.pek3b.qingstor.com/docker/compose/releases/download/%s/docker-compose-linux-x86_64"

	ContainerdUrl   = "https://github.com/containerd/containerd/releases/download/v%s/containerd-%s-linux-%s.tar.gz"
	ContainerdUrlCN = "https://kubernetes-release.pek3b.qingstor.com/containerd/containerd/releases/download/v%s/containerd-%s-linux-%s.tar.gz"

	RuncUrl  = "https://github.com/opencontainers/runc/releases/download/%s/runc.%s"
	RunUrlCN = "https://kubernetes-release.pek3b.qingstor.com/opencontainers/runc/releases/download/%s/runc.%s"

	AppArmorUrl      = "https://launchpad.net/ubuntu/+source/apparmor/%s-0ubuntu1/+build/28428840/+files/apparmor_%s-0ubuntu1_amd64.deb"
	AppArmorArmUrl   = "https://launchpad.net/ubuntu/+source/apparmor/%s-0ubuntu1/+build/28430859/+files/apparmor_%s-0ubuntu1_armhf.deb"
	AppArmorArm64Url = "https://launchpad.net/ubuntu/+source/apparmor/%s-0ubuntu1/+build/28428841/+files/apparmor_%s-0ubuntu1_arm64.deb"
	AppArmorPPC64Url = "https://launchpad.net/ubuntu/+source/apparmor/%s-0ubuntu1/+build/28428843/+files/apparmor_%s-0ubuntu1_ppc64el.deb"

	AWSCliUrl  = "https://awscli.amazonaws.com/awscli-exe-linux-x86_64.zip"
	OSSUtilUrl = "https://github.com/aliyun/ossutil/releases/download/%s/%s"
	COSUtilUrl = "https://github.com/tencentyun/coscli/releases/download/%s/coscli-%s-%s"

	MinioUrl         = "https://dl.min.io/server/minio/release/linux-%s/archive/minio.%s"
	MinioOperatorUrl = "https://github.com/beclab/minio-operator/releases/download/v%s/minio-operator-v%s-linux-%s.tar.gz"

	RedisUrl = "%s/%s"

	JuiceFsUrl = "https://github.com/beclab/juicefs-ext/releases/download/%s/juicefs-%s-linux-%s.tar.gz"

	VeleroUrl = "https://github.com/beclab/velero/releases/download/%s/velero-%s-linux-%s.tar.gz"

	CudaKeyringUbuntu24Amd64Url = "https://developer.download.nvidia.com/compute/cuda/repos/ubuntu2404/x86_64/cuda-keyring_1.1-1_all.deb"
	CudaKeyringUbuntu24Arm64Url = "https://developer.download.nvidia.com/compute/cuda/repos/ubuntu2404/arm64/cuda-keyring_1.1-1_all.deb"

	// https://developer.download.nvidia.com/compute/cuda/repos/ubuntu2204/x86_64/cuda-keyring_1.1-1_all.deb
	CudaKeyringUbuntu22Amd64Url = "https://developer.download.nvidia.com/compute/cuda/repos/ubuntu2204/x86_64/cuda-keyring_1.0-1_all.deb"
	// https://developer.download.nvidia.com/compute/cuda/repos/ubuntu2204/arm64/cuda-keyring_1.1-1_all.deb
	CudaKeyringUbuntu22Arm64Url = "https://developer.download.nvidia.com/compute/cuda/repos/ubuntu2204/arm64/cuda-keyring_1.0-1_all.deb"

	// https://developer.download.nvidia.com/compute/cuda/repos/ubuntu2004/x86_64/cuda-keyring_1.1-1_all.deb
	CudaKeyringUbuntu20Amd64Url = "https://developer.download.nvidia.com/compute/cuda/repos/ubuntu2004/x86_64/cuda-keyring_1.0-1_all.deb"
	// https://developer.download.nvidia.com/compute/cuda/repos/ubuntu2004/arm64/cuda-keyring_1.1-1_all.deb
	CudaKeyringUbuntu20Arm64Url = "https://developer.download.nvidia.com/compute/cuda/repos/ubuntu2004/arm64/cuda-keyring_1.0-1_all.deb"

	CudaKeyringUrl    = "https://developer.download.nvidia.com/compute/cuda/repos/%s/%s/cuda-keyring_%s-1_all.deb" // arm64  x86_64
	CudaKeyringCNDUrl = "%s/%s"
	CudaGpgKeyUrl     = "https://nvidia.github.io/libnvidia-container/gpgkey"
	CudaLibNvidiaUrl  = "https://nvidia.github.io/libnvidia-container/%s%s/libnvidia-container.list"
	// "https://wslstorestorage.blob.core.windows.net/wslblob/Ubuntu2204-221101.AppxBundle"
	WslImageUrl = "%s/%s"
)

func getGpuCDNPrefix(arch string, fileHash string) string {
	var res = fileHash
	if arch == "arm64" {
		res = "arm64/" + fileHash
	}
	return res
}

func getNvidiaLibUrl(platform string, os string) string {
	return fmt.Sprintf(CudaLibNvidiaUrl, platform, os)
}

func getCudaKeyringUrlExt(arch string, fileHash string) string {
	var res = fileHash
	if arch == "arm64" {
		res = "arm64/" + fileHash
	}
	return res
}

func getCudaKeyringUrl(arch string, os string, cudaversion string) string {
	var version string

	if strings.Contains(os, "24.") {
		version = "ubuntu2404"
	} else if strings.Contains(os, "22.") {
		version = "ubuntu2204"
	} else {
		version = "ubuntu2004"
	}

	return fmt.Sprintf(CudaKeyringUrl, version, arch, cudaversion)
}
