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

package network

import (
	"embed"
	"fmt"
	"io"
	"os"
	"path/filepath"

	"bytetrade.io/web3os/installer/apis/kubekey/v1alpha2"
	"bytetrade.io/web3os/installer/pkg/common"
	"bytetrade.io/web3os/installer/pkg/core/action"
	"bytetrade.io/web3os/installer/pkg/core/connector"
	"bytetrade.io/web3os/installer/pkg/core/util"
	"bytetrade.io/web3os/installer/pkg/images"
	"bytetrade.io/web3os/installer/pkg/plugins/network/templates"
	"github.com/pkg/errors"
)

//go:embed cilium-1.11.6.tgz
var f embed.FS

type ReleaseCiliumChart struct {
	common.KubeAction
}

func (r *ReleaseCiliumChart) Execute(runtime connector.Runtime) error {
	fs, err := os.Create(fmt.Sprintf("%s/cilium.tgz", runtime.GetWorkDir()))
	if err != nil {
		return err
	}
	chartFile, err := f.Open("cilium-1.11.6.tgz")
	if err != nil {
		return err
	}
	defer chartFile.Close()

	_, err = io.Copy(fs, chartFile)
	if err != nil {
		return err
	}

	return nil
}

type SyncCiliumChart struct {
	common.KubeAction
}

func (s *SyncCiliumChart) Execute(runtime connector.Runtime) error {
	src := filepath.Join(runtime.GetWorkDir(), "cilium.tgz")
	dst := filepath.Join(common.TmpDir, "cilium.tgz")
	if err := runtime.GetRunner().Scp(src, dst); err != nil {
		return errors.Wrap(errors.WithStack(err), fmt.Sprintf("sync cilium chart failed"))
	}
	if _, err := runtime.GetRunner().SudoCmd(fmt.Sprintf("mv %s/cilium.tgz /etc/kubernetes", common.TmpDir), true, false); err != nil {
		return errors.Wrap(errors.WithStack(err), "sync cilium chart failed")
	}
	return nil
}

type DeployCilium struct {
	common.KubeAction
}

func (d *DeployCilium) Execute(runtime connector.Runtime) error {
	ciliumImage := images.GetImage(runtime, d.KubeConf, "cilium").ImageName()
	ciliumOperatorImage := images.GetImage(runtime, d.KubeConf, "cilium-operator-generic").ImageName()

	cmd := fmt.Sprintf("/usr/local/bin/helm upgrade --install cilium /etc/kubernetes/cilium.tgz --namespace kube-system "+
		"--set operator.image.override=%s "+
		"--set operator.replicas=1 "+
		"--set image.override=%s "+
		"--set ipam.operator.clusterPoolIPv4PodCIDR=%s", ciliumOperatorImage, ciliumImage, d.KubeConf.Cluster.Network.KubePodsCIDR)

	if d.KubeConf.Cluster.Kubernetes.DisableKubeProxy {
		cmd = fmt.Sprintf("%s --set kubeProxyReplacement=strict --set k8sServiceHost=%s --set k8sServicePort=%d", cmd, d.KubeConf.Cluster.ControlPlaneEndpoint.Address, d.KubeConf.Cluster.ControlPlaneEndpoint.Port)
	}

	if _, err := runtime.GetRunner().SudoCmd(cmd, true, false); err != nil {
		return errors.Wrap(errors.WithStack(err), "deploy cilium failed")
	}
	return nil
}

type DeployNetworkPlugin struct {
	common.KubeAction
}

func (d *DeployNetworkPlugin) Execute(runtime connector.Runtime) error {
	if _, err := runtime.GetRunner().SudoCmd(
		"/usr/local/bin/kubectl apply -f /etc/kubernetes/network-plugin.yaml --force", false, true); err != nil {
		return errors.Wrap(errors.WithStack(err), "deploy network plugin failed")
	}
	return nil
}

type DeployKubeovnPlugin struct {
	common.KubeAction
}

func (d *DeployKubeovnPlugin) Execute(runtime connector.Runtime) error {
	if _, err := runtime.GetRunner().SudoCmd(
		"/usr/local/bin/kubectl apply -f /etc/kubernetes/kube-ovn-crd.yaml --force", true, false); err != nil {
		return errors.Wrap(errors.WithStack(err), "deploy kube-ovn-crd.yaml failed")
	}
	if _, err := runtime.GetRunner().SudoCmd(
		"/usr/local/bin/kubectl apply -f /etc/kubernetes/ovn.yaml --force", true, false); err != nil {
		return errors.Wrap(errors.WithStack(err), "deploy ovn.yaml failed")
	}
	if _, err := runtime.GetRunner().SudoCmd(
		"/usr/local/bin/kubectl apply -f /etc/kubernetes/kube-ovn.yaml --force", true, false); err != nil {
		return errors.Wrap(errors.WithStack(err), "deploy kube-ovn.yaml failed")
	}
	return nil
}

type DeployNetworkMultusPlugin struct {
	common.KubeAction
}

func (d *DeployNetworkMultusPlugin) Execute(runtime connector.Runtime) error {
	if _, err := runtime.GetRunner().SudoCmd(
		"/usr/local/bin/kubectl apply -f /etc/kubernetes/multus-network-plugin.yaml --force", true, false); err != nil {
		return errors.Wrap(errors.WithStack(err), "deploy multus network plugin failed")
	}
	return nil
}

type LabelNode struct {
	common.KubeAction
}

func (l *LabelNode) Execute(runtime connector.Runtime) error {
	if _, err := runtime.GetRunner().SudoCmd(
		fmt.Sprintf("/usr/local/bin/kubectl label no -l%s kube-ovn/role=master --overwrite",
			l.KubeConf.Cluster.Network.Kubeovn.Label),
		true, false); err != nil {
		return errors.Wrap(errors.WithStack(err), "label kubeovn/role=master in master node failed")
	}
	return nil
}

type GenerateSSL struct {
	common.KubeAction
}

func (g *GenerateSSL) Execute(runtime connector.Runtime) error {
	if exist, err := runtime.GetRunner().SudoCmd(
		"/usr/local/bin/kubectl get secret -n kube-system kube-ovn-tls --ignore-not-found",
		true, false); err != nil {
		return errors.Wrap(errors.WithStack(err), "find ovn secret failed")
	} else if exist != "" {
		return nil
	}

	if _, err := runtime.GetRunner().SudoCmd(
		fmt.Sprintf("docker run --rm -v %s:/etc/ovn %s bash generate-ssl.sh",
			runtime.GetWorkDir(), images.GetImage(runtime, g.KubeConf, "kubeovn").ImageName()),
		true, false); err != nil {
		return errors.Wrap(errors.WithStack(err), "generate ovn secret failed")
	}

	if _, err := runtime.GetRunner().SudoCmd(
		fmt.Sprintf("/usr/local/bin/kubectl create secret generic -n kube-system kube-ovn-tls "+
			"--from-file=cacert=%s/cacert.pem "+
			"--from-file=cert=%s/ovn-cert.pem "+
			"--from-file=key=%s/ovn-privkey.pem",
			runtime.GetWorkDir(), runtime.GetWorkDir(), runtime.GetWorkDir()),
		true, false); err != nil {
		return errors.Wrap(errors.WithStack(err), "create ovn secret failed")
	}

	if _, err := runtime.GetRunner().SudoCmd(
		fmt.Sprintf("rm -rf %s/cacert.pem %s/ovn-cert.pem %s/ovn-privkey.pem %s/ovn-req.pem",
			runtime.GetWorkDir(), runtime.GetWorkDir(), runtime.GetWorkDir(), runtime.GetWorkDir()),
		true, false); err != nil {
		return errors.Wrap(errors.WithStack(err), "remove generated ovn secret file failed")
	}

	return nil
}

type GenerateKubeOVN struct {
	common.KubeAction
}

func (g *GenerateKubeOVN) Execute(runtime connector.Runtime) error {
	address, err := runtime.GetRunner().Cmd(
		"/usr/local/bin/kubectl get no -lkube-ovn/role=master --no-headers -o wide | awk '{print $6}' | tr \\\\n ','",
		true, false)
	if err != nil {
		return errors.Wrap(errors.WithStack(err), "get kube-ovn label node address failed")
	}

	count, err := runtime.GetRunner().Cmd(
		fmt.Sprintf("/usr/local/bin/kubectl get no -l%s --no-headers -o wide | wc -l | sed 's/ //g'",
			g.KubeConf.Cluster.Network.Kubeovn.Label), true, false)
	if err != nil {
		return errors.Wrap(errors.WithStack(err), "count kube-ovn label nodes num failed")
	}

	if count == "0" {
		return fmt.Errorf("no node with label: %s", g.KubeConf.Cluster.Network.Kubeovn.Label)
	}

	templateAction := action.Template{
		Name:     "GenerateKubeOVNCrd",
		Template: templates.KubeOvnCrd,
		Dst:      filepath.Join(common.KubeConfigDir, templates.KubeOvnCrd.Name()),
	}
	templateAction.Init(nil, nil)
	if err := templateAction.Execute(runtime); err != nil {
		return err
	}

	templateAction = action.Template{
		Name:     "GenerateOVN",
		Template: templates.OVN,
		Dst:      filepath.Join(common.KubeConfigDir, templates.OVN.Name()),
		Data: util.Data{
			"Address":               address,
			"Count":                 count,
			"KubeovnImage":          images.GetImage(runtime, g.KubeConf, "kubeovn").ImageName(),
			"TunnelType":            g.KubeConf.Cluster.Network.Kubeovn.TunnelType,
			"DpdkMode":              g.KubeConf.Cluster.Network.Kubeovn.Dpdk.DpdkMode,
			"DpdkVersion":           g.KubeConf.Cluster.Network.Kubeovn.Dpdk.DpdkVersion,
			"OvnVersion":            v1alpha2.DefaultKubeovnVersion,
			"EnableSSL":             g.KubeConf.Cluster.Network.Kubeovn.EnableSSL,
			"HwOffload":             g.KubeConf.Cluster.Network.Kubeovn.OvsOvn.HwOffload,
			"SvcYamlIpfamilypolicy": g.KubeConf.Cluster.Network.Kubeovn.SvcYamlIpfamilypolicy,
		},
	}
	templateAction.Init(nil, nil)
	if err := templateAction.Execute(runtime); err != nil {
		return err
	}

	templateAction = action.Template{
		Name:     "GenerateKubeOVN",
		Template: templates.KubeOvn,
		Dst:      filepath.Join(common.KubeConfigDir, templates.KubeOvn.Name()),
		Data: util.Data{
			"Address":               address,
			"Count":                 count,
			"KubeovnImage":          images.GetImage(runtime, g.KubeConf, "kubeovn").ImageName(),
			"PodCIDR":               g.KubeConf.Cluster.Network.KubePodsCIDR,
			"SvcCIDR":               g.KubeConf.Cluster.Network.KubeServiceCIDR,
			"JoinCIDR":              g.KubeConf.Cluster.Network.Kubeovn.JoinCIDR,
			"PodGateway":            g.KubeConf.Cluster.Network.Kubeovn.KubeOvnController.PodGateway,
			"CheckGateway":          g.KubeConf.Cluster.Network.Kubeovn.KubeovnCheckGateway(),
			"LogicalGateway":        g.KubeConf.Cluster.Network.Kubeovn.KubeOvnController.LogicalGateway,
			"PingExternalAddress":   g.KubeConf.Cluster.Network.Kubeovn.KubeOvnPinger.PingerExternalAddress,
			"PingExternalDNS":       g.KubeConf.Cluster.Network.Kubeovn.KubeOvnPinger.PingerExternalDomain,
			"NetworkType":           g.KubeConf.Cluster.Network.Kubeovn.KubeOvnController.NetworkType,
			"TunnelType":            g.KubeConf.Cluster.Network.Kubeovn.TunnelType,
			"ExcludeIps":            g.KubeConf.Cluster.Network.Kubeovn.KubeOvnController.ExcludeIps,
			"PodNicType":            g.KubeConf.Cluster.Network.Kubeovn.KubeOvnController.PodNicType,
			"VlanID":                g.KubeConf.Cluster.Network.Kubeovn.KubeOvnController.VlanID,
			"VlanInterfaceName":     g.KubeConf.Cluster.Network.Kubeovn.KubeOvnController.VlanInterfaceName,
			"Iface":                 g.KubeConf.Cluster.Network.Kubeovn.KubeOvnCni.Iface,
			"EnableSSL":             g.KubeConf.Cluster.Network.Kubeovn.EnableSSL,
			"EnableMirror":          g.KubeConf.Cluster.Network.Kubeovn.KubeOvnCni.EnableMirror,
			"EnableLB":              g.KubeConf.Cluster.Network.Kubeovn.KubeovnEnableLB(),
			"EnableNP":              g.KubeConf.Cluster.Network.Kubeovn.KubeovnEnableNP(),
			"EnableEipSnat":         g.KubeConf.Cluster.Network.Kubeovn.KubeovnEnableEipSnat(),
			"EnableExternalVPC":     g.KubeConf.Cluster.Network.Kubeovn.KubeovnEnableExternalVPC(),
			"SvcYamlIpfamilypolicy": g.KubeConf.Cluster.Network.Kubeovn.SvcYamlIpfamilypolicy,
			"DpdkTunnelIface":       g.KubeConf.Cluster.Network.Kubeovn.Dpdk.DpdkTunnelIface,
			"CNIConfigPriority":     g.KubeConf.Cluster.Network.Kubeovn.KubeOvnCni.CNIConfigPriority,
			"Modules":               g.KubeConf.Cluster.Network.Kubeovn.KubeOvnCni.Modules,
			"RPMs":                  g.KubeConf.Cluster.Network.Kubeovn.KubeOvnCni.RPMs,
		},
	}
	templateAction.Init(nil, nil)
	if err := templateAction.Execute(runtime); err != nil {
		return err
	}

	return nil
}

type ChmodKubectlKo struct {
	common.KubeAction
}

func (c *ChmodKubectlKo) Execute(runtime connector.Runtime) error {
	if _, err := runtime.GetRunner().SudoCmd(
		fmt.Sprintf("chmod +x %s", filepath.Join(common.BinDir, templates.KubectlKo.Name())), false, false); err != nil {
		return errors.Wrap(errors.WithStack(err), "chmod +x kubectl-ko failed")
	}
	return nil
}
