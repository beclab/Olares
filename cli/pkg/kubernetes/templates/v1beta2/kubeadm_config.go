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

package v1beta2

import (
	"fmt"
	"strings"
	"text/template"

	"bytetrade.io/web3os/installer/pkg/utils"

	versionutil "k8s.io/apimachinery/pkg/util/version"

	"github.com/lithammer/dedent"
	"github.com/pkg/errors"
	"gopkg.in/yaml.v2"

	"bytetrade.io/web3os/installer/pkg/common"
	"bytetrade.io/web3os/installer/pkg/core/connector"
	"bytetrade.io/web3os/installer/pkg/core/logger"
)

var (
	funcMap = template.FuncMap{"toYaml": utils.ToYAML, "indent": utils.Indent}
	// KubeadmConfig defines the template of kubeadm configuration file.
	KubeadmConfig = template.Must(template.New("kubeadm-config.yaml").Funcs(funcMap).Parse(
		dedent.Dedent(`
{{- if .IsInitCluster -}}
---
apiVersion: kubeadm.k8s.io/v1beta3
kind: ClusterConfiguration
etcd:
{{- if .EtcdTypeIsKubeadm }}
  local:
    imageRepository: {{ .EtcdRepo }}
    imageTag: {{ .EtcdTag }}
    serverCertSANs:
    {{- range .ExternalEtcd.Endpoints }}
    - {{ . }}
    {{- end }}
{{- else }}
  external:
    endpoints:
    {{- range .ExternalEtcd.Endpoints }}
    - {{ . }}
    {{- end }}
{{- if .ExternalEtcd.CAFile }}
    caFile: {{ .ExternalEtcd.CAFile }}
{{- end }}
{{- if .ExternalEtcd.CertFile }}
    certFile: {{ .ExternalEtcd.CertFile }}
{{- end }}
{{- if .ExternalEtcd.KeyFile }}
    keyFile: {{ .ExternalEtcd.KeyFile }}
{{- end }}
{{- end }}
kubernetesVersion: {{ .Version }}
certificatesDir: /etc/kubernetes/pki
clusterName: {{ .ClusterName }}
controlPlaneEndpoint: {{ .ControlPlaneEndpoint }}
networking:
  dnsDomain: {{ .DNSDomain }}
  podSubnet: {{ .PodSubnet }}
  serviceSubnet: {{ .ServiceSubnet }}
apiServer:
  extraArgs:
{{ toYaml .ApiServerArgs | indent 4}}
  certSANs:
    {{- range .CertSANs }}
    - {{ . }}
    {{- end }}
controllerManager:
  extraArgs:
    node-cidr-mask-size: "{{ .NodeCidrMaskSize }}"
{{ toYaml .ControllerManagerArgs | indent 4 }}
  extraVolumes:
  - name: host-time
    hostPath: /etc/localtime
    mountPath: /etc/localtime
    readOnly: true
scheduler:
  extraArgs:
{{ toYaml .SchedulerArgs | indent 4 }}

---
apiVersion: kubeadm.k8s.io/v1beta3
kind: InitConfiguration
localAPIEndpoint:
  advertiseAddress: {{ .AdvertiseAddress }}
  bindPort: {{ .BindPort }}
nodeRegistration:
{{- if .CriSock }}
  criSocket: {{ .CriSock }}
{{- end }}
  kubeletExtraArgs:
    cgroup-driver: {{ .CgroupDriver }}
---
apiVersion: kubeproxy.config.k8s.io/v1alpha1
kind: KubeProxyConfiguration
{{ toYaml .KubeProxyConfiguration }}
---
apiVersion: kubelet.config.k8s.io/v1beta1
kind: KubeletConfiguration
{{ toYaml .KubeletConfiguration }}

{{- else -}}
---
apiVersion: kubeadm.k8s.io/v1beta3
kind: JoinConfiguration
discovery:
  bootstrapToken:
    apiServerEndpoint: {{ .ControlPlaneEndpoint }}
    token: "{{ .BootstrapToken }}"
    unsafeSkipCAVerification: true
  tlsBootstrapToken: "{{ .BootstrapToken }}"
{{- if .IsControlPlane }}
controlPlane:
  localAPIEndpoint:
    advertiseAddress: {{ .AdvertiseAddress }}
    bindPort: {{ .BindPort }}
  certificateKey: {{ .CertificateKey }}
{{- end }}
nodeRegistration:
{{- if .CriSock }}
  criSocket: {{ .CriSock }}
{{- end }}
  kubeletExtraArgs:
    cgroup-driver: {{ .CgroupDriver }}

{{- end }}
    `)))
)

var (
	// ref: https://kubernetes.io/docs/reference/command-line-tools-reference/feature-gates/
	FeatureGatesDefaultConfiguration = map[string]bool{
		"RotateKubeletServerCertificate": true, //k8s 1.7+
		"TTLAfterFinished":               true, //k8s 1.12+
		"ExpandCSIVolumes":               true, //k8s 1.14+
		"CSIStorageCapacity":             true, //k8s 1.19+
	}
	FeatureGatesSecurityDefaultConfiguration = map[string]bool{
		"RotateKubeletServerCertificate": true, //k8s 1.7+
		"TTLAfterFinished":               true, //k8s 1.12+
		"ExpandCSIVolumes":               true, //k8s 1.14+
		"CSIStorageCapacity":             true, //k8s 1.19+
		"SeccompDefault":                 true, //kubelet
	}

	ApiServerArgs = map[string]string{
		"bind-address":        "0.0.0.0",
		"audit-log-maxage":    "30",
		"audit-log-maxbackup": "10",
		"audit-log-maxsize":   "100",
	}
	ApiServerSecurityArgs = map[string]string{
		"bind-address":        "0.0.0.0",
		"audit-log-maxage":    "30",
		"audit-log-maxbackup": "10",
		"audit-log-maxsize":   "100",
		"authorization-mode":  "Node,RBAC",
		// --enable-admission-plugins=EventRateLimit must have a configuration file
		"enable-admission-plugins": "AlwaysPullImages,ServiceAccount,NamespaceLifecycle,NodeRestriction,LimitRanger,ResourceQuota,MutatingAdmissionWebhook,ValidatingAdmissionWebhook,PodNodeSelector,PodSecurity",
		// "audit-log-path":      "/var/log/apiserver/audit.log", // need audit policy
		"profiling":              "false",
		"request-timeout":        "120s",
		"service-account-lookup": "true",
		"tls-min-version":        "VersionTLS12",
		"tls-cipher-suites":      "TLS_ECDHE_ECDSA_WITH_AES_128_GCM_SHA256,TLS_ECDHE_RSA_WITH_AES_128_GCM_SHA256,TLS_ECDHE_ECDSA_WITH_CHACHA20_POLY1305",
	}
	ControllermanagerArgs = map[string]string{
		"bind-address":                "0.0.0.0",
		"cluster-signing-duration":    "87600h",
		"terminated-pod-gc-threshold": "1",
	}
	ControllermanagerSecurityArgs = map[string]string{
		"bind-address":                    "127.0.0.1",
		"cluster-signing-duration":        "87600h",
		"profiling":                       "false",
		"terminated-pod-gc-threshold":     "1",
		"use-service-account-credentials": "true",
	}
	SchedulerArgs = map[string]string{
		"bind-address": "0.0.0.0",
	}
	SchedulerSecurityArgs = map[string]string{
		"bind-address": "127.0.0.1",
		"profiling":    "false",
	}
)

func GetApiServerArgs(securityEnhancement bool) map[string]string {
	if securityEnhancement {
		return ApiServerSecurityArgs
	}
	return ApiServerArgs
}

func GetControllermanagerArgs(securityEnhancement bool) map[string]string {
	if securityEnhancement {
		return ControllermanagerSecurityArgs
	}
	return ControllermanagerArgs
}

func GetSchedulerArgs(securityEnhancement bool) map[string]string {
	if securityEnhancement {
		return SchedulerSecurityArgs
	}
	return SchedulerArgs
}

func AdjustDefaultFeatureGates(kubeConf *common.KubeConf) {
	for _, conf := range []map[string]bool{FeatureGatesDefaultConfiguration, FeatureGatesSecurityDefaultConfiguration} {
		// When kubernetes version is less than 1.21,`CSIStorageCapacity` is not recognized and should not be set.
		cmp, _ := versionutil.MustParseSemantic(kubeConf.Cluster.Kubernetes.Version).Compare("v1.21.0")
		if cmp == -1 {
			delete(conf, "CSIStorageCapacity")
		}

		// When kubernetes version is equal to or greater than 1.27, `CSIStorageCapacity` is removed and not recognized
		// the same logic applies to the feature gates below
		cmp, _ = versionutil.MustParseSemantic(kubeConf.Cluster.Kubernetes.Version).Compare("v1.27.0")
		if cmp >= 0 {
			delete(conf, "CSIStorageCapacity")
		}

		cmp, _ = versionutil.MustParseSemantic(kubeConf.Cluster.Kubernetes.Version).Compare("v1.24.0")
		if cmp >= 0 {
			delete(conf, "TTLAfterFinished")
		}

		cmp, _ = versionutil.MustParseSemantic(kubeConf.Cluster.Kubernetes.Version).Compare("v1.26.0")
		if cmp >= 0 {
			delete(conf, "ExpandCSIVolumes")
		}

		cmp, _ = versionutil.MustParseSemantic(kubeConf.Cluster.Kubernetes.Version).Compare("v1.28.0")
		if cmp >= 0 {
			delete(conf, "SeccompDefault")
		}

	}
}

func UpdateFeatureGatesConfiguration(args map[string]string, kubeConf *common.KubeConf) map[string]string {
	var featureGates []string

	for k, v := range kubeConf.Cluster.Kubernetes.FeatureGates {
		featureGates = append(featureGates, fmt.Sprintf("%s=%v", k, v))
	}

	for k, v := range FeatureGatesDefaultConfiguration {
		if _, ok := kubeConf.Cluster.Kubernetes.FeatureGates[k]; !ok {
			featureGates = append(featureGates, fmt.Sprintf("%s=%v", k, v))
		}
	}

	args["feature-gates"] = strings.Join(featureGates, ",")

	return args
}

func GetKubeletConfiguration(runtime connector.Runtime, kubeConf *common.KubeConf, criSock string, securityEnhancement bool) map[string]interface{} {
	defaultKubeletConfiguration := map[string]interface{}{
		"clusterDomain":                   kubeConf.Cluster.Kubernetes.DNSDomain,
		"clusterDNS":                      []string{kubeConf.Cluster.CorednsClusterIP()},
		"shutdownGracePeriod":             kubeConf.Cluster.Kubernetes.ShutdownGracePeriod,
		"shutdownGracePeriodCriticalPods": kubeConf.Cluster.Kubernetes.ShutdownGracePeriodCriticalPods,
		"maxPods":                         kubeConf.Cluster.Kubernetes.MaxPods,
		"podPidsLimit":                    kubeConf.Cluster.Kubernetes.PodPidsLimit,
		"rotateCertificates":              true,
		"failSwapOn":                      false,
		"kubeReserved": map[string]string{
			"cpu":    "200m",
			"memory": "250Mi",
		},
		"systemReserved": map[string]string{
			"cpu":    "200m",
			"memory": "250Mi",
		},
		"evictionHard": map[string]string{
			"memory.available": "5%",
			"pid.available":    "10%",
		},
		"evictionSoft": map[string]string{
			"memory.available": "10%",
		},
		"evictionSoftGracePeriod": map[string]string{
			"memory.available": "2m",
		},
		"evictionMaxPodGracePeriod":        120,
		"evictionPressureTransitionPeriod": "30s",
		"featureGates":                     FeatureGatesDefaultConfiguration,
		"runtimeRequestTimeout":            "5m",
	}

	if securityEnhancement {
		defaultKubeletConfiguration["readOnlyPort"] = 0
		defaultKubeletConfiguration["protectKernelDefaults"] = true
		defaultKubeletConfiguration["eventRecordQPS"] = 1
		defaultKubeletConfiguration["streamingConnectionIdleTimeout"] = "5m"
		defaultKubeletConfiguration["makeIPTablesUtilChains"] = true
		defaultKubeletConfiguration["tlsCipherSuites"] = []string{
			"TLS_ECDHE_ECDSA_WITH_AES_128_GCM_SHA256",
			"TLS_ECDHE_RSA_WITH_AES_128_GCM_SHA256",
			"TLS_ECDHE_ECDSA_WITH_CHACHA20_POLY1305",
		}
		defaultKubeletConfiguration["featureGates"] = FeatureGatesSecurityDefaultConfiguration
	}

	cgroupDriver, err := GetKubeletCgroupDriver(runtime, kubeConf)
	if err != nil {
		logger.Fatal(err)
	}
	if len(cgroupDriver) == 0 {
		defaultKubeletConfiguration["cgroupDriver"] = "systemd"
	}

	if len(criSock) != 0 {
		defaultKubeletConfiguration["containerLogMaxSize"] = "5Mi"
		defaultKubeletConfiguration["containerLogMaxFiles"] = 3
	}

	customKubeletConfiguration := make(map[string]interface{})
	if len(kubeConf.Cluster.Kubernetes.KubeletConfiguration.Raw) != 0 {
		err := yaml.Unmarshal(kubeConf.Cluster.Kubernetes.KubeletConfiguration.Raw, &customKubeletConfiguration)
		if err != nil {
			logger.Fatal("failed to parse kubelet configuration")
		}
	}

	kubeletConfiguration := make(map[string]interface{})
	if len(customKubeletConfiguration) != 0 {
		for customArg := range customKubeletConfiguration {
			if _, ok := defaultKubeletConfiguration[customArg]; ok {
				kubeletConfiguration[customArg] = customKubeletConfiguration[customArg]
				delete(defaultKubeletConfiguration, customArg)
				delete(customKubeletConfiguration, customArg)
			} else {
				kubeletConfiguration[customArg] = customKubeletConfiguration[customArg]
			}
		}
	}

	if len(defaultKubeletConfiguration) != 0 {
		for k, v := range defaultKubeletConfiguration {
			kubeletConfiguration[k] = v
		}
	}

	if featureGates, ok := kubeletConfiguration["featureGates"].(map[string]bool); ok {
		for k, v := range kubeConf.Cluster.Kubernetes.FeatureGates {
			if _, ok := featureGates[k]; !ok {
				featureGates[k] = v
			}
		}

		for k, v := range FeatureGatesDefaultConfiguration {
			if _, ok := featureGates[k]; !ok {
				featureGates[k] = v
			}
		}
	}

	if kubeConf.Arg.EnablePodSwap {
		kubeletConfiguration["memorySwap"] = map[string]string{
			"swapBehavior": "LimitedSwap",
		}
	}

	if kubeConf.Arg.Debug {
		logger.Debugf("Set kubeletConfiguration: %v", kubeletConfiguration)
	}

	return kubeletConfiguration
}

func GetKubeletCgroupDriver(runtime connector.Runtime, kubeConf *common.KubeConf) (string, error) {
	var cmd, kubeletCgroupDriver string
	switch kubeConf.Cluster.Kubernetes.ContainerManager {
	case common.Docker, "":
		cmd = "docker info | grep 'Cgroup Driver'"
	case common.Crio:
		cmd = "crio config | grep cgroup_manager"
	case common.Containerd:
		cmd = "containerd config dump | grep SystemdCgroup"
	case common.Isula:
		cmd = "isula info | grep 'Cgroup Driver'"
	default:
		kubeletCgroupDriver = ""
	}

	checkResult, err := runtime.GetRunner().SudoCmd(cmd, false, false)
	if err != nil {
		return "", errors.Wrap(errors.WithStack(err), "Failed to get container runtime cgroup driver.")
	}
	if strings.Contains(checkResult, "systemd") || strings.Contains(checkResult, "SystemdCgroup = true") {
		kubeletCgroupDriver = "systemd"
	} else if strings.Contains(checkResult, "cgroupfs") || strings.Contains(checkResult, "SystemdCgroup = false") {
		kubeletCgroupDriver = "cgroupfs"
	} else {
		return "", errors.Errorf("Failed to get container runtime cgroup driver from %s by run %s", checkResult, cmd)
	}
	return kubeletCgroupDriver, nil
}

func GetKubeProxyConfiguration(kubeConf *common.KubeConf, isPveLxc bool) map[string]interface{} {
	defaultKubeProxyConfiguration := map[string]interface{}{
		"clusterCIDR": kubeConf.Cluster.Network.KubePodsCIDR,
		"mode":        kubeConf.Cluster.Kubernetes.ProxyMode,
		"iptables": map[string]interface{}{
			"masqueradeAll": kubeConf.Cluster.Kubernetes.MasqueradeAll,
			"masqueradeBit": 14,
			"minSyncPeriod": "0s",
			"syncPeriod":    "30s",
		},
	}

	if isPveLxc {
		defaultKubeProxyConfiguration["conntrack"] = map[string]interface{}{
			"maxPerCore": 0,
		}
	}

	customKubeProxyConfiguration := make(map[string]interface{})
	if len(kubeConf.Cluster.Kubernetes.KubeProxyConfiguration.Raw) != 0 {
		err := yaml.Unmarshal(kubeConf.Cluster.Kubernetes.KubeProxyConfiguration.Raw, &customKubeProxyConfiguration)
		if err != nil {
			logger.Fatal("failed to parse kube-proxy's configuration")
		}
	}

	kubeProxyConfiguration := make(map[string]interface{})
	if len(customKubeProxyConfiguration) != 0 {
		for customArg := range customKubeProxyConfiguration {
			if _, ok := defaultKubeProxyConfiguration[customArg]; ok {
				kubeProxyConfiguration[customArg] = customKubeProxyConfiguration[customArg]
				delete(defaultKubeProxyConfiguration, customArg)
				delete(customKubeProxyConfiguration, customArg)
			} else {
				kubeProxyConfiguration[customArg] = customKubeProxyConfiguration[customArg]
			}
		}
	}

	if len(defaultKubeProxyConfiguration) != 0 {
		for defaultArg := range defaultKubeProxyConfiguration {
			kubeProxyConfiguration[defaultArg] = defaultKubeProxyConfiguration[defaultArg]
		}
	}

	return kubeProxyConfiguration
}
