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

package kubernetes

import (
	"fmt"
	"io/ioutil"
	"path/filepath"
	"regexp"
	"strings"

	"bytetrade.io/web3os/installer/pkg/common"
	"bytetrade.io/web3os/installer/pkg/core/connector"
	"github.com/pkg/errors"
)

type KubernetesStatus struct {
	Version        string
	BootstrapToken string
	CertificateKey string
	ClusterInfo    string
	KubeConfig     string
	NodesInfo      map[string]string
}

func NewKubernetesStatus() *KubernetesStatus {
	return &KubernetesStatus{NodesInfo: make(map[string]string)}
}

func (k *KubernetesStatus) SearchVersion(runtime connector.Runtime) error {
	cmd := "sudo cat /etc/kubernetes/manifests/kube-apiserver.yaml | grep 'image:' | awk -F '[:]' '{print $(NF-0)}'"
	if output, err := runtime.GetRunner().Cmd(cmd, false, true); err != nil {
		return errors.Wrap(errors.WithStack(err), "search current version failed")
	} else {
		if !strings.Contains(output, "No such file or directory") {
			k.Version = output
		}
	}
	return nil
}

func (k *KubernetesStatus) SearchJoinInfo(runtime connector.Runtime) error {
	checkKubeadmConfig, err := runtime.GetRunner().SudoCmd("cat /etc/kubernetes/kubeadm-config.yaml", false, false)
	if err != nil {
		return err
	}
	if (k.BootstrapToken != "" || k.CertificateKey != "") &&
		(!strings.Contains(checkKubeadmConfig, "InitConfiguration") ||
			!strings.Contains(checkKubeadmConfig, "ClusterConfiguration")) {
		return nil
	}

	uploadCertsCmd := "/usr/local/bin/kubeadm init phase upload-certs --upload-certs"
	output, err := runtime.GetRunner().SudoCmd(uploadCertsCmd, false, true)
	if err != nil {
		return errors.Wrap(errors.WithStack(err), "Failed to upload kubeadm certs")
	}
	reg := regexp.MustCompile("[0-9|a-z]{64}")
	k.CertificateKey = reg.FindAllString(output, -1)[0]

	if err := patchKubeadmSecret(runtime); err != nil {
		return err
	}

	tokenCreateMasterCmd := "/usr/local/bin/kubeadm token create"
	token, err := runtime.GetRunner().SudoCmd(tokenCreateMasterCmd, false, true)
	if err != nil {
		return errors.Wrap(errors.WithStack(err), "Failed to get join node cmd")
	}

	reg = regexp.MustCompile("[0-9|a-z]{6}.[0-9|a-z]{16}")
	k.BootstrapToken = reg.FindAllString(token, -1)[0]
	return nil
}

func (k *KubernetesStatus) SearchClusterInfo(runtime connector.Runtime) error {
	output, err := runtime.GetRunner().SudoCmd(
		"/usr/local/bin/kubectl --kubeconfig=/etc/kubernetes/admin.conf --no-headers=true get nodes -o custom-columns=:metadata.name,:status.nodeInfo.kubeletVersion,:status.addresses",
		false, true)
	if err != nil {
		return errors.Wrap(errors.WithStack(err), "get kubernetes cluster info failed")
	}
	k.ClusterInfo = output
	return nil
}

func (k *KubernetesStatus) SearchNodesInfo(_ connector.Runtime) error {
	ipv4Regexp, err := regexp.Compile(common.IPv4Regexp)
	if err != nil {
		return err
	}
	ipv6Regexp, err := regexp.Compile(common.IPv6Regexp)
	if err != nil {
		return err
	}
	tmp := strings.Split(k.ClusterInfo, "\r\n")
	if len(tmp) >= 1 {
		for i := 0; i < len(tmp); i++ {
			if ipv4 := ipv4Regexp.FindStringSubmatch(tmp[i]); len(ipv4) != 0 {
				k.NodesInfo[ipv4[0]] = ipv4[0]
			}
			if ipv6 := ipv6Regexp.FindStringSubmatch(tmp[i]); len(ipv6) != 0 {
				k.NodesInfo[ipv6[0]] = ipv6[0]
			}
			if len(strings.Fields(tmp[i])) > 3 {
				k.NodesInfo[strings.Fields(tmp[i])[0]] = strings.Fields(tmp[i])[1]
			} else {
				k.NodesInfo[strings.Fields(tmp[i])[0]] = ""
			}
		}
	}
	return nil
}

func (k *KubernetesStatus) SearchKubeConfig(runtime connector.Runtime) error {
	kubeCfgCmd := "cat /etc/kubernetes/admin.conf"
	if kubeConfigStr, err := runtime.GetRunner().SudoCmd(kubeCfgCmd, false, false); err != nil {
		return errors.Wrap(errors.WithStack(err), "search cluster kubeconfig failed")
	} else {
		k.KubeConfig = kubeConfigStr
	}
	return nil
}

func (k *KubernetesStatus) LoadKubeConfig(runtime connector.Runtime, kubeConf *common.KubeConf) error {
	kubeConfigPath := filepath.Join(runtime.GetWorkDir(), fmt.Sprintf("config-%s", runtime.GetObjName()))
	kubeConfigStr := k.KubeConfig

	oldServer := fmt.Sprintf("server: https://%s:%d", kubeConf.Cluster.ControlPlaneEndpoint.Domain, kubeConf.Cluster.ControlPlaneEndpoint.Port)
	newServer := fmt.Sprintf("server: https://%s:%d", kubeConf.Cluster.ControlPlaneEndpoint.Address, kubeConf.Cluster.ControlPlaneEndpoint.Port)
	newKubeConfigStr := strings.Replace(kubeConfigStr, oldServer, newServer, -1)

	if err := ioutil.WriteFile(kubeConfigPath, []byte(newKubeConfigStr), 0644); err != nil {
		return err
	}
	return nil
}

func (k *KubernetesStatus) GetNodesInfo() map[string]string {
	return k.NodesInfo
}

// PatchKubeadmSecret is used to patch etcd's certs for kubeadm-certs secret.
func patchKubeadmSecret(runtime connector.Runtime) error {
	externalEtcdCerts := []string{"external-etcd-ca.crt", "external-etcd.crt", "external-etcd.key"}
	for _, cert := range externalEtcdCerts {
		patchData := fmt.Sprintf(`{"data": {"%s": ""}}`, cert)
		_, err := runtime.GetRunner().SudoCmd(
			fmt.Sprintf("/usr/local/bin/kubectl --kubeconfig=/etc/kubernetes/admin.conf patch -n kube-system secret kubeadm-certs -p '%s'", patchData),
			false, true)
		if err != nil {
			return errors.Wrap(errors.WithStack(err), "patch kubeadm secret failed")
		}
	}
	return nil
}
