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
	"bufio"
	"context"
	"encoding/base64"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"sort"
	"strings"
	"time"

	"bytetrade.io/web3os/installer/pkg/storage"
	storagetpl "bytetrade.io/web3os/installer/pkg/storage/templates"

	"bytetrade.io/web3os/installer/pkg/etcd"
	"bytetrade.io/web3os/installer/pkg/manifest"

	"github.com/pkg/errors"
	corev1 "k8s.io/api/core/v1"
	kubeerrors "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	versionutil "k8s.io/apimachinery/pkg/util/version"
	kube "k8s.io/client-go/kubernetes"
	"k8s.io/client-go/tools/clientcmd"

	kubekeyv1alpha2 "bytetrade.io/web3os/installer/apis/kubekey/v1alpha2"
	"bytetrade.io/web3os/installer/pkg/common"
	"bytetrade.io/web3os/installer/pkg/core/action"
	"bytetrade.io/web3os/installer/pkg/core/connector"
	"bytetrade.io/web3os/installer/pkg/core/logger"
	"bytetrade.io/web3os/installer/pkg/core/prepare"
	"bytetrade.io/web3os/installer/pkg/core/task"
	"bytetrade.io/web3os/installer/pkg/core/util"
	"bytetrade.io/web3os/installer/pkg/files"
	"bytetrade.io/web3os/installer/pkg/images"
	"bytetrade.io/web3os/installer/pkg/kubernetes/templates"
	"bytetrade.io/web3os/installer/pkg/kubernetes/templates/v1beta2"
	"bytetrade.io/web3os/installer/pkg/utils"
)

type GetKubeType struct{}

func (g *GetKubeType) Execute() (kubeType string) {
	kubeType = common.K3s
	var getKubeVersion = new(GetKubeVersion)
	_, kubeType, _ = getKubeVersion.Execute()

	if kubeType != "" {
		return
	}

	if util.IsExist("/etc/systemd/system/k3s.service") || util.IsExist("/usr/local/bin/k3s-uninstall.sh") {
		kubeType = common.K3s
	} else {
		kubeType = common.K8s
	}

	return
}

type GetKubeVersion struct{}

func (g *GetKubeVersion) Execute() (string, string, error) {
	var kubectlpath, err = util.GetCommand(common.CommandKubectl)
	if err != nil {
		return "", "", fmt.Errorf("kubectl not found, Olares might not be installed.")
	}

	var ctx, cancel = context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	cmd := exec.CommandContext(ctx, "/bin/sh", "-c", fmt.Sprintf("%s get nodes -l node-role.kubernetes.io/master -o jsonpath='{.items[*].status.nodeInfo.kubeletVersion}'", kubectlpath))
	// the kubectl command may continue running after the context has timed out
	// causing the cmd.Wait() to block for a long time
	cmd.WaitDelay = 3 * time.Second
	output, err := cmd.CombinedOutput()
	if err != nil {
		return "", "", errors.Wrap(errors.WithStack(err), "get kube version failed")
	}

	if output == nil || len(output) == 0 {
		return "", "", fmt.Errorf("get kube version failed")
	}

	var version = string(output)
	var kubeVersion, kubeType = utils.KubeVersionAlias(version)

	return kubeVersion, kubeType, nil
}

type GetClusterStatus struct {
	common.KubeAction
}

func (g *GetClusterStatus) Execute(runtime connector.Runtime) error {
	exist, err := runtime.GetRunner().FileExist("/etc/kubernetes/admin.conf")
	if err != nil {
		return err
	}
	if !exist {
		g.PipelineCache.Set(common.ClusterExist, false)
		return nil
	} else {
		g.PipelineCache.Set(common.ClusterExist, true)

		if v, ok := g.PipelineCache.Get(common.ClusterStatus); ok {
			cluster := v.(*KubernetesStatus)
			if err := cluster.SearchVersion(runtime); err != nil {
				return err
			}
			if err := cluster.SearchKubeConfig(runtime); err != nil {
				return err
			}
			if err := cluster.LoadKubeConfig(runtime, g.KubeConf); err != nil {
				return err
			}
			if err := cluster.SearchClusterInfo(runtime); err != nil {
				return err
			}
			if err := cluster.SearchNodesInfo(runtime); err != nil {
				return err
			}
			if err := cluster.SearchJoinInfo(runtime); err != nil {
				return err
			}

			g.PipelineCache.Set(common.ClusterStatus, cluster)
		} else {
			return errors.New("get kubernetes cluster status by pipeline cache failed")
		}
	}
	return nil
}

type SyncKubeBinary struct {
	common.KubeAction
	manifest.ManifestAction
}

func (i *SyncKubeBinary) Execute(runtime connector.Runtime) error {
	if err := utils.ResetTmpDir(runtime); err != nil {
		return err
	}

	binaryList := []string{"kubeadm", "kubelet", "kubectl", "helm", "cni-plugins"}
	for _, name := range binaryList {
		binary, err := i.Manifest.Get(name)
		if err != nil {
			return fmt.Errorf("get kube binary %s info failed: %w", name, err)
		}

		path := binary.FilePath(i.BaseDir)

		fileName := binary.Filename
		switch name {
		//case "kubelet":
		//	if err := runtime.GetRunner().Scp(binary.Path, fmt.Sprintf("%s/%s", common.TmpDir, binary.Name)); err != nil {
		//		return errors.Wrap(errors.WithStack(err), fmt.Sprintf("sync kube binaries failed"))
		//	}
		case "cni-plugins":
			dst := filepath.Join(common.TmpDir, fileName)
			logger.Debugf("SyncKubeBinary cp %s from %s to %s", name, path, dst)
			if err := runtime.GetRunner().Scp(path, dst); err != nil {
				return errors.Wrap(errors.WithStack(err), fmt.Sprintf("sync kube binaries failed"))
			}
			if _, err := runtime.GetRunner().SudoCmd(fmt.Sprintf("tar -zxf %s -C /opt/cni/bin", dst), false, false); err != nil {
				return err
			}
		case "helm":
			dst := filepath.Join(common.TmpDir, fileName)
			untarDst := filepath.Join(common.TmpDir, strings.TrimSuffix(fileName, ".tar.gz"))
			logger.Debugf("SyncKubeBinary cp %s from %s to %s", name, path, dst)
			if err := runtime.GetRunner().Scp(path, dst); err != nil {
				return errors.Wrap(errors.WithStack(err), fmt.Sprintf("sync kube binaries failed"))
			}

			cmd := fmt.Sprintf("mkdir -p %s && tar -zxf %s -C %s && cd %s/linux-* && mv ./helm /usr/local/bin/.",
				untarDst, dst, untarDst, untarDst)
			if _, err := runtime.GetRunner().SudoCmd(cmd, false, false); err != nil {
				return err
			}
		default:
			dst := filepath.Join(common.BinDir, name)
			if err := runtime.GetRunner().SudoScp(path, dst); err != nil {
				return errors.Wrap(errors.WithStack(err), fmt.Sprintf("sync kube binaries failed"))
			}
			if _, err := runtime.GetRunner().SudoCmd(fmt.Sprintf("chmod +x %s", dst), false, false); err != nil {
				return err
			}
		}
	}
	return nil
}

type SyncKubelet struct {
	common.KubeAction
}

func (s *SyncKubelet) Execute(runtime connector.Runtime) error {
	if _, err := runtime.GetRunner().SudoCmd("chmod +x /usr/local/bin/kubelet", false, false); err != nil {
		return errors.Wrap(errors.WithStack(err), "sync kubelet service failed")
	}
	return nil
}

type GenerateKubeletService struct {
	common.KubeAction
}

func (t *GenerateKubeletService) Execute(runtime connector.Runtime) error {
	tplActions := &action.Template{
		Name:     "GenerateKubeletService",
		Template: templates.KubeletService,
		Dst:      filepath.Join("/etc/systemd/system/", templates.KubeletService.Name()),
		Data: util.Data{
			"JuiceFSPreCheckEnabled": util.IsExist(storage.JuiceFsServiceFile),
			"JuiceFSServiceUnit":     storagetpl.JuicefsService.Name(),
			"JuiceFSBinPath":         storage.JuiceFsFile,
			"JuiceFSMountPoint":      storage.OlaresJuiceFSRootDir,
		},
	}
	return tplActions.Execute(runtime)
}

type EnableKubelet struct {
	common.KubeAction
}

func (e *EnableKubelet) Execute(runtime connector.Runtime) error {
	if _, err := runtime.GetRunner().SudoCmd("systemctl disable kubelet "+
		"&& systemctl enable kubelet "+
		"&& ln -snf /usr/local/bin/kubelet /usr/bin/kubelet", false, false); err != nil {
		return errors.Wrap(errors.WithStack(err), "enable kubelet service failed")
	}
	return nil
}

type GenerateKubeletEnv struct {
	common.KubeAction
}

func (g *GenerateKubeletEnv) Execute(runtime connector.Runtime) error {
	host := runtime.RemoteHost()
	templateAction := action.Template{
		Name:     "GenerateKubeletEnv",
		Template: templates.KubeletEnv,
		Dst:      filepath.Join("/etc/systemd/system/kubelet.service.d", templates.KubeletEnv.Name()),
		Data: util.Data{
			"NodeIP":           host.GetInternalAddress(),
			"Hostname":         host.GetName(),
			"ContainerRuntime": "",
		},
	}

	templateAction.Init(nil, nil)
	if err := templateAction.Execute(runtime); err != nil {
		return err
	}
	return nil
}

type GenerateKubeadmConfig struct {
	common.KubeAction
	IsInitConfiguration     bool
	WithSecurityEnhancement bool
}

func (g *GenerateKubeadmConfig) Execute(runtime connector.Runtime) error {
	host := runtime.RemoteHost()

	localConfig := filepath.Join(runtime.GetWorkDir(), "kubeadm-config.yaml")
	if util.IsExist(localConfig) {
		// todo: if it is necessary?
		if err := runtime.GetRunner().SudoScp(localConfig, "/etc/kubernetes/kubeadm-config.yaml"); err != nil {
			return errors.Wrap(errors.WithStack(err), "scp local kubeadm config failed")
		}
	} else {
		// generate etcd configuration
		var externalEtcd kubekeyv1alpha2.ExternalEtcd
		var endpointsList, etcdCertSANs []string

		switch g.KubeConf.Cluster.Etcd.Type {
		case kubekeyv1alpha2.KubeKey:
			for _, host := range runtime.GetHostsByRole(common.ETCD) {
				endpoint := fmt.Sprintf("https://%s:%s", host.GetInternalAddress(), kubekeyv1alpha2.DefaultEtcdPort)
				endpointsList = append(endpointsList, endpoint)
			}
			externalEtcd.Endpoints = endpointsList

			externalEtcd.CAFile = "/etc/ssl/etcd/ssl/ca.pem"
			externalEtcd.CertFile = fmt.Sprintf("/etc/ssl/etcd/ssl/node-%s.pem", host.GetName())
			externalEtcd.KeyFile = fmt.Sprintf("/etc/ssl/etcd/ssl/node-%s-key.pem", host.GetName())
		case kubekeyv1alpha2.External:
			externalEtcd.Endpoints = g.KubeConf.Cluster.Etcd.External.Endpoints

			if len(g.KubeConf.Cluster.Etcd.External.CAFile) != 0 && len(g.KubeConf.Cluster.Etcd.External.CAFile) != 0 && len(g.KubeConf.Cluster.Etcd.External.CAFile) != 0 {
				externalEtcd.CAFile = fmt.Sprintf("/etc/ssl/etcd/ssl/%s", filepath.Base(g.KubeConf.Cluster.Etcd.External.CAFile))
				externalEtcd.CertFile = fmt.Sprintf("/etc/ssl/etcd/ssl/%s", filepath.Base(g.KubeConf.Cluster.Etcd.External.CertFile))
				externalEtcd.KeyFile = fmt.Sprintf("/etc/ssl/etcd/ssl/%s", filepath.Base(g.KubeConf.Cluster.Etcd.External.KeyFile))
			}
		case kubekeyv1alpha2.Kubeadm:
			altNames := etcd.GenerateAltName(g.KubeConf, &runtime)
			etcdCertSANs = append(etcdCertSANs, altNames.DNSNames...)
			for _, ip := range altNames.IPs {
				etcdCertSANs = append(etcdCertSANs, string(ip))
			}
		}

		_, ApiServerArgs := util.GetArgs(v1beta2.GetApiServerArgs(g.WithSecurityEnhancement), g.KubeConf.Cluster.Kubernetes.ApiServerArgs)
		_, ControllerManagerArgs := util.GetArgs(v1beta2.GetControllermanagerArgs(g.WithSecurityEnhancement), g.KubeConf.Cluster.Kubernetes.ControllerManagerArgs)
		_, SchedulerArgs := util.GetArgs(v1beta2.GetSchedulerArgs(g.WithSecurityEnhancement), g.KubeConf.Cluster.Kubernetes.SchedulerArgs)

		checkCgroupDriver, err := v1beta2.GetKubeletCgroupDriver(runtime, g.KubeConf)
		if err != nil {
			return err
		}

		var (
			bootstrapToken, certificateKey string
			// todo: if port needed
		)
		if !g.IsInitConfiguration {
			if v, ok := g.PipelineCache.Get(common.ClusterStatus); ok {
				cluster := v.(*KubernetesStatus)
				bootstrapToken = cluster.BootstrapToken
				certificateKey = cluster.CertificateKey
			} else {
				return errors.New("get kubernetes cluster status by pipeline cache failed")
			}
		}

		v1beta2.AdjustDefaultFeatureGates(g.KubeConf)
		templateAction := action.Template{
			Name:     "GenerateKubeadmConfig",
			Template: v1beta2.KubeadmConfig,
			Dst:      filepath.Join(common.KubeConfigDir, v1beta2.KubeadmConfig.Name()),
			Data: util.Data{
				"IsInitCluster":          g.IsInitConfiguration,
				"EtcdTypeIsKubeadm":      g.KubeConf.Cluster.Etcd.Type == kubekeyv1alpha2.Kubeadm,
				"EtcdCertSANs":           etcdCertSANs,
				"EtcdRepo":               strings.TrimSuffix(images.GetImage(runtime, g.KubeConf, "etcd").ImageRepo(), "/etcd"),
				"EtcdTag":                images.GetImage(runtime, g.KubeConf, "etcd").Tag,
				"Version":                g.KubeConf.Cluster.Kubernetes.Version,
				"ClusterName":            g.KubeConf.Cluster.Kubernetes.ClusterName,
				"DNSDomain":              g.KubeConf.Cluster.Kubernetes.DNSDomain,
				"AdvertiseAddress":       host.GetInternalAddress(),
				"BindPort":               kubekeyv1alpha2.DefaultApiserverPort,
				"ControlPlaneEndpoint":   fmt.Sprintf("%s:%d", g.KubeConf.Cluster.ControlPlaneEndpoint.Domain, g.KubeConf.Cluster.ControlPlaneEndpoint.Port),
				"PodSubnet":              g.KubeConf.Cluster.Network.KubePodsCIDR,
				"ServiceSubnet":          g.KubeConf.Cluster.Network.KubeServiceCIDR,
				"CertSANs":               g.KubeConf.Cluster.GenerateCertSANs(),
				"ExternalEtcd":           externalEtcd,
				"NodeCidrMaskSize":       g.KubeConf.Cluster.Kubernetes.NodeCidrMaskSize,
				"CriSock":                g.KubeConf.Cluster.Kubernetes.ContainerRuntimeEndpoint,
				"ApiServerArgs":          v1beta2.UpdateFeatureGatesConfiguration(ApiServerArgs, g.KubeConf),
				"ControllerManagerArgs":  v1beta2.UpdateFeatureGatesConfiguration(ControllerManagerArgs, g.KubeConf),
				"SchedulerArgs":          v1beta2.UpdateFeatureGatesConfiguration(SchedulerArgs, g.KubeConf),
				"KubeletConfiguration":   v1beta2.GetKubeletConfiguration(runtime, g.KubeConf, g.KubeConf.Cluster.Kubernetes.ContainerRuntimeEndpoint, g.WithSecurityEnhancement),
				"KubeProxyConfiguration": v1beta2.GetKubeProxyConfiguration(g.KubeConf, runtime.GetSystemInfo().IsPveLxc()),
				"IsControlPlane":         host.IsRole(common.Master),
				"CgroupDriver":           checkCgroupDriver,
				"BootstrapToken":         bootstrapToken,
				"CertificateKey":         certificateKey,
			},
		}

		templateAction.Init(nil, nil)
		if err := templateAction.Execute(runtime); err != nil {
			return err
		}
	}
	return nil
}

type KubeadmInit struct {
	common.KubeAction
}

func (k *KubeadmInit) Execute(runtime connector.Runtime) error {
	initCmd := "/usr/local/bin/kubeadm init --config=/etc/kubernetes/kubeadm-config.yaml --ignore-preflight-errors=FileExisting-crictl,ImagePull"

	if k.KubeConf.Cluster.Kubernetes.DisableKubeProxy {
		initCmd = initCmd + " --skip-phases=addon/kube-proxy"
	}

	// we manage the creation of coredns ourselves
	initCmd = initCmd + " --skip-phases=addon/coredns"

	if _, err := runtime.GetRunner().SudoCmd(initCmd, false, true); err != nil {
		// kubeadm reset and then retry
		resetCmd := "/usr/local/bin/kubeadm reset -f"
		if k.KubeConf.Cluster.Kubernetes.ContainerRuntimeEndpoint != "" {
			resetCmd = resetCmd + " --cri-socket " + k.KubeConf.Cluster.Kubernetes.ContainerRuntimeEndpoint
		}
		_, _ = runtime.GetRunner().SudoCmd(resetCmd, false, true)
		return errors.Wrap(errors.WithStack(err), "init kubernetes cluster failed")
	}
	return nil
}

type CopyKubeConfigForControlPlane struct {
	common.KubeAction
}

func (c *CopyKubeConfigForControlPlane) Execute(runtime connector.Runtime) error {
	createConfigDirCmd := "mkdir -p /root/.kube"
	getKubeConfigCmd := "cp -f /etc/kubernetes/admin.conf /root/.kube/config"
	cmd := strings.Join([]string{createConfigDirCmd, getKubeConfigCmd}, " && ")
	if _, err := runtime.GetRunner().SudoCmd(cmd, false, false); err != nil {
		return errors.Wrap(errors.WithStack(err), "copy kube config failed")
	}

	userMkdir := "mkdir -p $HOME/.kube"
	if _, err := runtime.GetRunner().Cmd(userMkdir, false, false); err != nil {
		return errors.Wrap(errors.WithStack(err), "user mkdir $HOME/.kube failed")
	}

	userCopyKubeConfig := "cp -f /etc/kubernetes/admin.conf $HOME/.kube/config"
	if _, err := runtime.GetRunner().SudoCmd(userCopyKubeConfig, false, false); err != nil {
		return errors.Wrap(errors.WithStack(err), "user copy /etc/kubernetes/admin.conf to $HOME/.kube/config failed")
	}

	if _, err := runtime.GetRunner().SudoCmd("chmod 0600 $HOME/.kube/config", false, false); err != nil {
		return errors.Wrap(errors.WithStack(err), "chmod $HOME/.kube/config failed")
	}

	// userId, err := runtime.GetRunner().Cmd("echo $(id -u)", false, false)
	// if err != nil {
	// 	return errors.Wrap(errors.WithStack(err), "get user id failed")
	// }

	// userGroupId, err := runtime.GetRunner().Cmd("echo $(id -g)", false, false)
	// if err != nil {
	// 	return errors.Wrap(errors.WithStack(err), "get user group id failed")
	// }

	userId, err := runtime.GetRunner().Cmd("echo $SUDO_UID", false, false)
	if err != nil {
		return errors.Wrap(errors.WithStack(err), "get user id failed")
	}

	userGroupId, err := runtime.GetRunner().Cmd("echo $SUDO_GID", false, false)
	if err != nil {
		return errors.Wrap(errors.WithStack(err), "get user group id failed")
	}

	chownKubeConfig := fmt.Sprintf("chown -R %s:%s $HOME/.kube", userId, userGroupId)
	if _, err := runtime.GetRunner().SudoCmd(chownKubeConfig, false, false); err != nil {
		return errors.Wrap(errors.WithStack(err), "chown user kube config failed")
	}
	return nil
}

type RemoveMasterTaint struct {
	common.KubeAction
}

func (r *RemoveMasterTaint) Execute(runtime connector.Runtime) error {
	if _, err := runtime.GetRunner().SudoCmd(fmt.Sprintf(
		"/usr/local/bin/kubectl taint nodes %s node-role.kubernetes.io/master=:NoSchedule-",
		runtime.RemoteHost().GetName()), false, true); err != nil {
		logger.Warn(err.Error())
	}
	if _, err := runtime.GetRunner().SudoCmd(fmt.Sprintf(
		"/usr/local/bin/kubectl taint nodes %s node-role.kubernetes.io/control-plane=:NoSchedule-",
		runtime.RemoteHost().GetName()), false, true); err != nil {
		logger.Warn(err.Error())
	}
	return nil
}

type AddWorkerLabel struct {
	common.KubeAction
}

func (a *AddWorkerLabel) Execute(runtime connector.Runtime) error {
	if _, err := runtime.GetRunner().SudoCmd(fmt.Sprintf(
		"/usr/local/bin/kubectl label --overwrite node %s node-role.kubernetes.io/worker=",
		runtime.RemoteHost().GetName()), false, false); err != nil {
		return errors.Wrap(errors.WithStack(err), "add worker label failed")
	}
	return nil
}

type JoinNode struct {
	common.KubeAction
}

func (j *JoinNode) Execute(runtime connector.Runtime) error {
	if _, err := runtime.GetRunner().SudoCmd("/usr/local/bin/kubeadm join --config=/etc/kubernetes/kubeadm-config.yaml --ignore-preflight-errors=FileExisting-crictl,ImagePull",
		true, false); err != nil {
		resetCmd := "/usr/local/bin/kubeadm reset -f"
		if j.KubeConf.Cluster.Kubernetes.ContainerRuntimeEndpoint != "" {
			resetCmd = resetCmd + " --cri-socket " + j.KubeConf.Cluster.Kubernetes.ContainerRuntimeEndpoint
		}
		_, _ = runtime.GetRunner().SudoCmd(resetCmd, true, false)
		return errors.Wrap(errors.WithStack(err), "join node failed")
	}
	return nil
}

type SyncKubeConfigToWorker struct {
	common.KubeAction
}

func (s *SyncKubeConfigToWorker) Execute(runtime connector.Runtime) error {
	if v, ok := s.PipelineCache.Get(common.ClusterStatus); ok {
		cluster := v.(*KubernetesStatus)

		createConfigDirCmd := "mkdir -p /root/.kube"
		if _, err := runtime.GetRunner().SudoCmd(createConfigDirCmd, false, false); err != nil {
			return errors.Wrap(errors.WithStack(err), "create .kube dir failed")
		}

		syncKubeConfigForRootCmd := fmt.Sprintf("echo '%s' > %s", cluster.KubeConfig, "/root/.kube/config")
		if _, err := runtime.GetRunner().SudoCmd(syncKubeConfigForRootCmd, false, false); err != nil {
			return errors.Wrap(errors.WithStack(err), "sync kube config for root failed")
		}

		if _, err := runtime.GetRunner().SudoCmd("chmod 0600 /root/.kube/config", false, false); err != nil {
			return errors.Wrap(errors.WithStack(err), "chmod $HOME/.kube/config failed")
		}

		userConfigDirCmd := "mkdir -p $HOME/.kube"
		if _, err := runtime.GetRunner().Cmd(userConfigDirCmd, false, false); err != nil {
			return errors.Wrap(errors.WithStack(err), "user mkdir $HOME/.kube failed")
		}

		syncKubeConfigForUserCmd := fmt.Sprintf("echo '%s' > %s", cluster.KubeConfig, "$HOME/.kube/config")
		if _, err := runtime.GetRunner().Cmd(syncKubeConfigForUserCmd, false, false); err != nil {
			return errors.Wrap(errors.WithStack(err), "sync kube config for normal user failed")
		}

		// userId, err := runtime.GetRunner().Cmd("echo $(id -u)", false, false)
		// if err != nil {
		// 	return errors.Wrap(errors.WithStack(err), "get user id failed")
		// }

		// userGroupId, err := runtime.GetRunner().Cmd("echo $(id -g)", false, false)
		// if err != nil {
		// 	return errors.Wrap(errors.WithStack(err), "get user group id failed")
		// }

		userId, err := runtime.GetRunner().Cmd("echo $SUDO_UID", false, false)
		if err != nil {
			return errors.Wrap(errors.WithStack(err), "get user id failed")
		}

		userGroupId, err := runtime.GetRunner().Cmd("echo $SUDO_GID", false, false)
		if err != nil {
			return errors.Wrap(errors.WithStack(err), "get user group id failed")
		}

		chownKubeConfig := fmt.Sprintf("chown -R %s:%s -R $HOME/.kube", userId, userGroupId)
		if _, err := runtime.GetRunner().SudoCmd(chownKubeConfig, false, false); err != nil {
			return errors.Wrap(errors.WithStack(err), "chown user kube config failed")
		}
	}
	return nil
}

type KubeadmReset struct {
	common.KubeAction
}

func (k *KubeadmReset) Execute(runtime connector.Runtime) error {
	resetCmd := "/usr/local/bin/kubeadm reset -f"
	if k.KubeConf.Cluster.Kubernetes.ContainerRuntimeEndpoint != "" {
		resetCmd = resetCmd + " --cri-socket " + k.KubeConf.Cluster.Kubernetes.ContainerRuntimeEndpoint
	}
	_, _ = runtime.GetRunner().SudoCmd(resetCmd, false, true)
	return nil
}

type UmountKubelet struct {
	common.KubeAction
}

func (u *UmountKubelet) Execute(runtime connector.Runtime) error {
	procMountsFile := "/proc/mounts"
	targetPaths := []string{
		"/var/lib/kubelet",
		"/run/netns/cni",
	}
	f, err := os.Open(procMountsFile)
	if err != nil {
		if !os.IsNotExist(err) {
			return fmt.Errorf("failed to open %s: %w", procMountsFile, err)
		}
		return nil
	}
	defer f.Close()

	var mounts []string
	scanner := bufio.NewScanner(f)
	for scanner.Scan() {
		line := scanner.Text()
		fields := strings.Fields(line)
		if len(fields) < 2 {
			continue
		}
		for _, targetPath := range targetPaths {
			if strings.HasPrefix(fields[1], targetPath) {
				mounts = append(mounts, fields[1])
			}
		}
	}
	if err := scanner.Err(); err != nil {
		return fmt.Errorf("failed to scan %s: %w", procMountsFile, err)
	}

	logger.Debugf("kubelet mounts %v", mounts)

	for _, m := range mounts {
		runtime.GetRunner().SudoCmd(fmt.Sprintf("umount %s && rm -rf %s", m, m), false, true)
	}

	return nil
}

type FindNode struct {
	common.KubeAction
}

func (f *FindNode) Execute(runtime connector.Runtime) error {
	var resArr []string
	res, err := runtime.GetRunner().Cmd(
		"sudo -E /usr/local/bin/kubectl get nodes | grep -v NAME | grep -v 'master\\|control-plane' | awk '{print $1}'",
		true, false)
	if err != nil {
		return errors.Wrap(errors.WithStack(err), "kubectl get nodes failed")
	}

	if !strings.Contains(res, "\r\n") {
		resArr = append(resArr, res)
	} else {
		resArr = strings.Split(res, "\r\n")
	}

	workerName := make(map[string]struct{})
	for j := 0; j < len(runtime.GetHostsByRole(common.Worker)); j++ {
		workerName[runtime.GetHostsByRole(common.Worker)[j].GetName()] = struct{}{}
	}

	var node string
	for i := 0; i < len(resArr); i++ {
		if _, ok := workerName[resArr[i]]; ok && resArr[i] == f.KubeConf.Arg.NodeName {
			node = resArr[i]
			break
		}
	}

	if node == "" {
		return errors.New("" +
			"1. check the node name in the config-sample.yaml\n" +
			"2. check the node name in the Kubernetes cluster\n" +
			"3. only support to delete a worker\n")
	}

	f.PipelineCache.Set("dstNode", node)
	return nil
}

type DrainNode struct {
	common.KubeAction
}

func (d *DrainNode) Execute(runtime connector.Runtime) error {
	nodeName, ok := d.PipelineCache.Get("dstNode")
	if !ok {
		return errors.New("get dstNode failed by pipeline cache")
	}
	if _, err := runtime.GetRunner().SudoCmd(fmt.Sprintf(
		"/usr/local/bin/kubectl drain %s --delete-emptydir-data --ignore-daemonsets --timeout=2m --force", nodeName),
		true, false); err != nil {
		return errors.Wrap(err, "drain the node failed")
	}
	return nil
}

type KubectlDeleteCurrentWorkerNode struct {
	common.KubeAction
	FailOnError bool
}

func (k *KubectlDeleteCurrentWorkerNode) Execute(runtime connector.Runtime) error {
	// currently only master node has a redis server
	if util.IsExist(storage.RedisServiceFile) {
		return nil
	}

	kubectl, err := util.GetCommand(common.CommandKubectl)
	// kubernetes very likely not installed
	if err != nil {
		return nil
	}
	nodeName := runtime.GetSystemInfo().GetHostname()
	if _, _, err := util.Exec(context.Background(), fmt.Sprintf(
		"%s delete node %s", kubectl, nodeName),
		true, false); err != nil {
		if k.FailOnError {
			return err
		}
		logger.Infof("failed to delete current node from kubernetes metadata, if this is a worker node, please delete it manually by \"kubectl delete node %s\" on the master to clean up", nodeName)
	}
	return nil
}

type KubectlDeleteNode struct {
	common.KubeAction
}

func (k *KubectlDeleteNode) Execute(runtime connector.Runtime) error {
	nodeName, ok := k.PipelineCache.Get("dstNode")
	if !ok {
		return errors.New("get dstNode failed by pipeline cache")
	}
	if _, err := runtime.GetRunner().SudoCmd(fmt.Sprintf(
		"/usr/local/bin/kubectl delete node %s", nodeName),
		true, false); err != nil {
		return errors.Wrap(err, "delete the node failed")
	}
	return nil
}

type SetUpgradePlan struct {
	common.KubeAction
	Step UpgradeStep
}

func (s *SetUpgradePlan) Execute(_ connector.Runtime) error {
	currentVersion, ok := s.PipelineCache.GetMustString(common.K8sVersion)
	if !ok {
		return errors.New("get current Kubernetes version failed by pipeline cache")
	}

	desiredVersion, ok := s.PipelineCache.GetMustString(common.DesiredK8sVersion)
	if !ok {
		return errors.New("get desired Kubernetes version failed by pipeline cache")
	}
	if cmp, err := versionutil.MustParseSemantic(currentVersion).Compare(desiredVersion); err != nil {
		return err
	} else if cmp == 1 {
		logger.Infof(
			"%s The current version (%s) is greater than the target version (%s)",
			common.LocalHost, currentVersion, desiredVersion)
		os.Exit(0)
	}

	if s.Step == ToV121 {
		v122 := versionutil.MustParseSemantic("v1.22.0")
		atLeast := versionutil.MustParseSemantic(desiredVersion).AtLeast(v122)
		cmp, err := versionutil.MustParseSemantic(currentVersion).Compare("v1.21.5")
		if err != nil {
			return err
		}
		if atLeast && cmp <= 0 {
			desiredVersion = "v1.21.5"
		}
	}

	s.PipelineCache.Set(common.PlanK8sVersion, desiredVersion)
	return nil
}

type CalculateNextVersion struct {
	common.KubeAction
}

func (c *CalculateNextVersion) Execute(_ connector.Runtime) error {
	currentVersion, ok := c.PipelineCache.GetMustString(common.K8sVersion)
	if !ok {
		return errors.New("get current Kubernetes version failed by pipeline cache")
	}
	planVersion, ok := c.PipelineCache.GetMustString(common.PlanK8sVersion)
	if !ok {
		return errors.New("get upgrade plan Kubernetes version failed by pipeline cache")
	}
	nextVersionStr := calculateNextStr(currentVersion, planVersion)
	c.KubeConf.Cluster.Kubernetes.Version = nextVersionStr
	return nil
}

func calculateNextStr(currentVersion, desiredVersion string) string {
	current := versionutil.MustParseSemantic(currentVersion)
	target := versionutil.MustParseSemantic(desiredVersion)
	var nextVersionMinor uint
	if target.Minor() == current.Minor() {
		nextVersionMinor = current.Minor()
	} else {
		nextVersionMinor = current.Minor() + 1
	}

	if nextVersionMinor == target.Minor() {
		return desiredVersion
	} else {
		nextVersionPatchList := make([]int, 0)
		for supportVersionStr := range files.FileSha256["kubeadm"]["amd64"] {
			supportVersion := versionutil.MustParseSemantic(supportVersionStr)
			if supportVersion.Minor() == nextVersionMinor {
				nextVersionPatchList = append(nextVersionPatchList, int(supportVersion.Patch()))
			}
		}
		sort.Ints(nextVersionPatchList)

		nextVersion := current.WithMinor(nextVersionMinor)
		nextVersion = nextVersion.WithPatch(uint(nextVersionPatchList[len(nextVersionPatchList)-1]))

		return fmt.Sprintf("v%s", nextVersion.String())
	}
}

type UpgradeKubeMaster struct {
	common.KubeAction
	ModuleName string
}

func (u *UpgradeKubeMaster) Execute(runtime connector.Runtime) error {
	host := runtime.RemoteHost()
	if err := KubeadmUpgradeTasks(runtime, u); err != nil {
		return errors.Wrap(errors.WithStack(err), fmt.Sprintf("upgrade cluster using kubeadm failed: %s", host.GetName()))
	}

	if _, err := runtime.GetRunner().SudoCmd("systemctl stop kubelet", false, false); err != nil {
		return errors.Wrap(errors.WithStack(err), fmt.Sprintf("stop kubelet failed: %s", host.GetName()))
	}

	if err := SetKubeletTasks(runtime, u.KubeAction); err != nil {
		return errors.Wrap(errors.WithStack(err), fmt.Sprintf("set kubelet failed: %s", host.GetName()))
	}
	if _, err := runtime.GetRunner().SudoCmd("systemctl daemon-reload && systemctl restart kubelet", true, false); err != nil {
		return errors.Wrap(errors.WithStack(err), fmt.Sprintf("restart kubelet failed: %s", host.GetName()))
	}

	time.Sleep(10 * time.Second)
	return nil
}

type UpgradeKubeWorker struct {
	common.KubeAction
	ModuleName string
}

func (u *UpgradeKubeWorker) Execute(runtime connector.Runtime) error {

	host := runtime.RemoteHost()

	if _, err := runtime.GetRunner().SudoCmd("/usr/local/bin/kubeadm upgrade node", true, false); err != nil {
		return errors.Wrap(errors.WithStack(err), fmt.Sprintf("upgrade node using kubeadm failed: %s", host.GetName()))
	}
	if _, err := runtime.GetRunner().SudoCmd("systemctl stop kubelet", true, false); err != nil {
		return errors.Wrap(errors.WithStack(err), fmt.Sprintf("stop kubelet failed: %s", host.GetName()))
	}
	if err := SetKubeletTasks(runtime, u.KubeAction); err != nil {
		return errors.Wrap(errors.WithStack(err), fmt.Sprintf("set kubelet failed: %s", host.GetName()))
	}
	if _, err := runtime.GetRunner().SudoCmd("systemctl daemon-reload && systemctl restart kubelet", true, false); err != nil {
		return errors.Wrap(errors.WithStack(err), fmt.Sprintf("restart kubelet failed: %s", host.GetName()))
	}
	time.Sleep(10 * time.Second)

	if err := SyncKubeConfigTask(runtime, u.KubeAction); err != nil {
		return errors.Wrap(errors.WithStack(err), fmt.Sprintf("sync kube config to worker failed: %s", host.GetName()))
	}
	return nil
}

func KubeadmUpgradeTasks(runtime connector.Runtime, u *UpgradeKubeMaster) error {
	host := runtime.RemoteHost()

	kubeadmUpgrade := &task.RemoteTask{
		Name:     "KubeadmUpgrade(k8s)",
		Desc:     "Upgrade cluster using kubeadm",
		Hosts:    []connector.Host{host},
		Prepare:  new(NotEqualDesiredVersion),
		Action:   new(KubeadmUpgrade),
		Parallel: false,
		Retry:    3,
	}

	copyKubeConfig := &task.RemoteTask{
		Name:     "CopyKubeConfig",
		Desc:     "Copy admin.conf to ~/.kube/config",
		Hosts:    []connector.Host{host},
		Prepare:  new(NotEqualDesiredVersion),
		Action:   new(CopyKubeConfigForControlPlane),
		Parallel: false,
		Retry:    2,
	}

	tasks := []task.Interface{
		kubeadmUpgrade,
		copyKubeConfig,
	}

	for i := range tasks {
		t := tasks[i]
		t.Init(runtime, u.ModuleCache, u.PipelineCache)
		if res := t.Execute(); res.IsFailed() {
			return res.CombineErr()
		}
	}
	return nil
}

type KubeadmUpgrade struct {
	common.KubeAction
}

func (k *KubeadmUpgrade) Execute(runtime connector.Runtime) error {

	host := runtime.RemoteHost()
	if _, err := runtime.GetRunner().SudoCmd(fmt.Sprintf(
		"timeout -k 600s 600s /usr/local/bin/kubeadm upgrade apply %s -y "+
			"--ignore-preflight-errors=all "+
			"--allow-experimental-upgrades "+
			"--allow-release-candidate-upgrades "+
			"--etcd-upgrade=false "+
			"--certificate-renewal=true ",
		k.KubeConf.Cluster.Kubernetes.Version), false, false); err != nil {
		return errors.Wrap(errors.WithStack(err), fmt.Sprintf("upgrade master failed: %s", host.GetName()))
	}
	return nil
}

func SetKubeletTasks(runtime connector.Runtime, kubeAction common.KubeAction) error {
	host := runtime.RemoteHost()
	syncKubelet := &task.RemoteTask{
		Name:     "SyncKubelet",
		Desc:     "synchronize kubelet",
		Hosts:    []connector.Host{host},
		Prepare:  new(NotEqualDesiredVersion),
		Action:   new(SyncKubelet),
		Parallel: false,
		Retry:    2,
	}

	enableKubelet := &task.RemoteTask{
		Name:     "EnableKubelet",
		Desc:     "enable kubelet service",
		Hosts:    []connector.Host{host},
		Prepare:  new(NotEqualDesiredVersion),
		Action:   new(EnableKubelet),
		Parallel: false,
		Retry:    5,
	}

	tasks := []task.Interface{
		syncKubelet,
		enableKubelet,
	}

	for i := range tasks {
		t := tasks[i]
		t.Init(runtime, kubeAction.ModuleCache, kubeAction.PipelineCache)
		if res := t.Execute(); res.IsFailed() {
			return res.CombineErr()
		}
	}
	return nil
}

func SyncKubeConfigTask(runtime connector.Runtime, kubeAction common.KubeAction) error {
	host := runtime.RemoteHost()
	syncKubeConfig := &task.RemoteTask{
		Name:  "SyncKubeConfig",
		Desc:  "synchronize kube config to worker",
		Hosts: []connector.Host{host},
		Prepare: &prepare.PrepareCollection{
			new(NotEqualDesiredVersion),
			new(common.OnlyWorker),
		},
		Action:   new(SyncKubeConfigToWorker),
		Parallel: true,
		Retry:    3,
	}

	tasks := []task.Interface{
		syncKubeConfig,
	}

	for i := range tasks {
		t := tasks[i]
		t.Init(runtime, kubeAction.ModuleCache, kubeAction.PipelineCache)
		if res := t.Execute(); res.IsFailed() {
			return res.CombineErr()
		}
	}
	return nil
}

type SetCurrentK8sVersion struct {
	common.KubeAction
}

func (s *SetCurrentK8sVersion) Execute(_ connector.Runtime) error {
	s.PipelineCache.Set(common.K8sVersion, s.KubeConf.Cluster.Kubernetes.Version)
	return nil
}

type SaveKubeConfig struct {
	common.KubeAction
}

func (s *SaveKubeConfig) Execute(runtime connector.Runtime) error {
	status, ok := s.PipelineCache.Get(common.ClusterStatus)
	if !ok {
		return errors.New("get kubernetes status failed by pipeline cache")
	}
	cluster := status.(*KubernetesStatus)
	kubeConfigStr := cluster.KubeConfig

	clusterPublicAddress := s.KubeConf.Cluster.ControlPlaneEndpoint.Address
	master1 := runtime.GetHostsByRole(common.Master)[0]
	if clusterPublicAddress == master1.GetInternalAddress() {
		clusterPublicAddress = master1.GetAddress()
	}

	oldServer := fmt.Sprintf("https://%s:%d", s.KubeConf.Cluster.ControlPlaneEndpoint.Domain, s.KubeConf.Cluster.ControlPlaneEndpoint.Port)
	newServer := fmt.Sprintf("https://%s:%d", clusterPublicAddress, s.KubeConf.Cluster.ControlPlaneEndpoint.Port)
	newKubeConfigStr := strings.Replace(kubeConfigStr, oldServer, newServer, -1)
	kubeConfigBase64 := base64.StdEncoding.EncodeToString([]byte(newKubeConfigStr))

	config, err := clientcmd.NewClientConfigFromBytes([]byte(newKubeConfigStr))
	if err != nil {
		return err
	}
	restConfig, err := config.ClientConfig()
	if err != nil {
		return err
	}
	clientsetForCluster, err := kube.NewForConfig(restConfig)
	if err != nil {
		return err
	}

	namespace := &corev1.Namespace{
		ObjectMeta: metav1.ObjectMeta{
			Name: "kubekey-system",
		},
	}
	if _, err := clientsetForCluster.
		CoreV1().
		Namespaces().
		Get(context.TODO(), namespace.Name, metav1.GetOptions{}); kubeerrors.IsNotFound(err) {
		if _, err := clientsetForCluster.
			CoreV1().
			Namespaces().
			Create(context.TODO(), namespace, metav1.CreateOptions{}); err != nil {
			return err
		}
	} else {
		return err
	}

	cm := &corev1.ConfigMap{
		ObjectMeta: metav1.ObjectMeta{
			Name: fmt.Sprintf("%s-kubeconfig", s.KubeConf.ClusterName),
		},
		Data: map[string]string{
			"kubeconfig": kubeConfigBase64,
		},
	}

	if _, err := clientsetForCluster.
		CoreV1().
		ConfigMaps("kubekey-system").
		Get(context.TODO(), cm.Name, metav1.GetOptions{}); kubeerrors.IsNotFound(err) {
		if _, err := clientsetForCluster.
			CoreV1().
			ConfigMaps("kubekey-system").
			Create(context.TODO(), cm, metav1.CreateOptions{}); err != nil {
			return err
		}
	} else {
		if _, err := clientsetForCluster.
			CoreV1().
			ConfigMaps("kubekey-system").
			Update(context.TODO(), cm, metav1.UpdateOptions{}); err != nil {
			return err
		}
	}
	return nil
}

type ConfigureKubernetes struct {
	common.KubeAction
}

func (c *ConfigureKubernetes) Execute(runtime connector.Runtime) error {
	host := runtime.RemoteHost()
	kubeHost := host.(*kubekeyv1alpha2.KubeHost)
	for k, v := range kubeHost.Labels {
		labelCmd := fmt.Sprintf("/usr/local/bin/kubectl label --overwrite node %s %s=%s", host.GetName(), k, v)
		_, err := runtime.GetRunner().SudoCmd(labelCmd, true, false)
		if err != nil {
			return err
		}
	}
	return nil
}

type EtcdSecurityEnhancemenAction struct {
	common.KubeAction
	ModuleName string
}

func (s *EtcdSecurityEnhancemenAction) Execute(runtime connector.Runtime) error {
	chmodEtcdCertsDirCmd := "chmod 700 /etc/ssl/etcd/ssl"
	chmodEtcdCertsCmd := "chmod 600 /etc/ssl/etcd/ssl/*"
	chmodEtcdDataDirCmd := "chmod 700 /var/lib/etcd"
	chmodEtcdCmd := "chmod 550 /usr/local/bin/etcd*"

	chownEtcdCertsDirCmd := "chown root:root /etc/ssl/etcd/ssl"
	chownEtcdCertsCmd := "chown root:root /etc/ssl/etcd/ssl/*"
	chownEtcdDataDirCmd := "chown etcd:etcd /var/lib/etcd"
	chownEtcdCmd := "chown root:root /usr/local/bin/etcd*"

	ETCDcmds := []string{chmodEtcdCertsDirCmd, chmodEtcdCertsCmd, chmodEtcdDataDirCmd, chmodEtcdCmd, chownEtcdCertsDirCmd, chownEtcdCertsCmd, chownEtcdDataDirCmd, chownEtcdCmd}

	if _, err := runtime.GetRunner().SudoCmd(strings.Join(ETCDcmds, " && "), true, false); err != nil {
		return errors.Wrap(errors.WithStack(err), "Updating permissions failed.")
	}

	return nil
}

type MasterSecurityEnhancemenAction struct {
	common.KubeAction
	ModuleName string
}

func (k *MasterSecurityEnhancemenAction) Execute(runtime connector.Runtime) error {
	// Control-plane Security Enhancemen
	chmodKubernetesDirCmd := "chmod 644 /etc/kubernetes"
	chownKubernetesDirCmd := "chown root:root /etc/kubernetes"

	chmodKubernetesConfigCmd := "chmod 600 -R /etc/kubernetes"
	chownKubernetesConfigCmd := "chown root:root -R /etc/kubernetes/*"

	chmodKubenretesManifestsDirCmd := "chmod 644 /etc/kubernetes/manifests"
	chownKubenretesManifestsDirCmd := "chown root:root /etc/kubernetes/manifests"

	chmodKubenretesCertsDirCmd := "chmod 644 /etc/kubernetes/pki"
	chownKubenretesCertsDirCmd := "chown root:root /etc/kubernetes/pki"

	// node Security Enhancemen
	chmodCniConfigDir := "chmod 600 -R /etc/cni/net.d"
	chownCniConfigDir := "chown root:root -R /etc/cni/net.d"

	chmodBinDir := "chmod 550 /usr/local/bin/"
	chownBinDir := "chown root:root /usr/local/bin/"

	chmodKubeCmd := "chmod 550 -R /usr/local/bin/kube*"
	chownKubeCmd := "chown root:root -R /usr/local/bin/kube*"

	chmodHelmCmd := "chmod 550 /usr/local/bin/helm"
	chownHelmCmd := "chown root:root /usr/local/bin/helm"

	chmodCniDir := "chmod 550 -R /opt/cni/bin"
	chownCniDir := "chown root:root -R /opt/cni/bin"

	chmodKubeletConfig := "chmod 640 /var/lib/kubelet/config.yaml && chmod 640 -R /etc/systemd/system/kubelet.service*"
	chownKubeletConfig := "chown root:root /var/lib/kubelet/config.yaml && chown root:root -R /etc/systemd/system/kubelet.service*"

	chmodCertsRenew := "chmod 640 /etc/systemd/system/k8s-certs-renew*"
	chownCertsRenew := "chown root:root /etc/systemd/system/k8s-certs-renew*"

	chmodMasterCmds := []string{chmodKubernetesConfigCmd, chmodKubernetesDirCmd, chmodKubenretesManifestsDirCmd, chmodKubenretesCertsDirCmd}
	if _, err := runtime.GetRunner().SudoCmd(strings.Join(chmodMasterCmds, " && "), true, false); err != nil {
		return errors.Wrap(errors.WithStack(err), "Updating permissions failed.")
	}
	chownMasterCmds := []string{chownKubernetesConfigCmd, chownKubernetesDirCmd, chownKubenretesManifestsDirCmd, chownKubenretesCertsDirCmd}
	if _, err := runtime.GetRunner().SudoCmd(strings.Join(chownMasterCmds, " && "), true, false); err != nil {
		return errors.Wrap(errors.WithStack(err), "Updating permissions failed.")
	}

	chmodNodesCmds := []string{chmodBinDir, chmodKubeCmd, chmodHelmCmd, chmodCniDir, chmodCniConfigDir, chmodKubeletConfig, chmodCertsRenew}
	if _, err := runtime.GetRunner().SudoCmd(strings.Join(chmodNodesCmds, " && "), true, false); err != nil {
		return errors.Wrap(errors.WithStack(err), "Updating permissions failed.")
	}
	chownNodesCmds := []string{chownBinDir, chownKubeCmd, chownHelmCmd, chownCniDir, chownCniConfigDir, chownKubeletConfig, chownCertsRenew}
	if _, err := runtime.GetRunner().SudoCmd(strings.Join(chownNodesCmds, " && "), true, false); err != nil {
		return errors.Wrap(errors.WithStack(err), "Updating permissions failed.")
	}

	return nil
}

type NodesSecurityEnhancemenAction struct {
	common.KubeAction
	ModuleName string
}

func (n *NodesSecurityEnhancemenAction) Execute(runtime connector.Runtime) error {
	// Control-plane Security Enhancemen
	chmodKubernetesDirCmd := "chmod 644 /etc/kubernetes"
	chownKubernetesDirCmd := "chown root:root /etc/kubernetes"

	chmodKubernetesConfigCmd := "chmod 600 -R /etc/kubernetes"
	chownKubernetesConfigCmd := "chown root:root -R /etc/kubernetes"

	chmodKubenretesManifestsDirCmd := "chmod 644 /etc/kubernetes/manifests"
	chownKubenretesManifestsDirCmd := "chown root:root /etc/kubernetes/manifests"

	chmodKubenretesCertsDirCmd := "chmod 644 /etc/kubernetes/pki"
	chownKubenretesCertsDirCmd := "chown root:root /etc/kubernetes/pki"

	// node Security Enhancemen
	chmodCniConfigDir := "chmod 600 -R /etc/cni/net.d"
	chownCniConfigDir := "chown root:root -R /etc/cni/net.d"

	chmodBinDir := "chmod 550 /usr/local/bin/"
	chownBinDir := "chown root:root /usr/local/bin/"

	chmodKubeCmd := "chmod 550 -R /usr/local/bin/kube*"
	chownKubeCmd := "chown root:root -R /usr/local/bin/kube*"

	chmodHelmCmd := "chmod 550 /usr/local/bin/helm"
	chownHelmCmd := "chown root:root /usr/local/bin/helm"

	chmodCniDir := "chmod 550 -R /opt/cni/bin"
	chownCniDir := "chown root:root -R /opt/cni/bin"

	chmodKubeletConfig := "chmod 640 /var/lib/kubelet/config.yaml && chmod 640 -R /etc/systemd/system/kubelet.service*"
	chownKubeletConfig := "chown root:root /var/lib/kubelet/config.yaml && chown root:root -R /etc/systemd/system/kubelet.service*"

	chmodMasterCmds := []string{chmodKubernetesConfigCmd, chmodKubernetesDirCmd, chmodKubenretesManifestsDirCmd, chmodKubenretesCertsDirCmd}
	if _, err := runtime.GetRunner().SudoCmd(strings.Join(chmodMasterCmds, " && "), true, false); err != nil {
		return errors.Wrap(errors.WithStack(err), "Updating permissions failed.")
	}
	chownMasterCmds := []string{chownKubernetesConfigCmd, chownKubernetesDirCmd, chownKubenretesManifestsDirCmd, chownKubenretesCertsDirCmd}
	if _, err := runtime.GetRunner().SudoCmd(strings.Join(chownMasterCmds, " && "), true, false); err != nil {
		return errors.Wrap(errors.WithStack(err), "Updating permissions failed.")
	}

	chmodNodesCmds := []string{chmodBinDir, chmodKubeCmd, chmodHelmCmd, chmodCniDir, chmodCniConfigDir, chmodKubeletConfig}
	if _, err := runtime.GetRunner().SudoCmd(strings.Join(chmodNodesCmds, " && "), true, false); err != nil {
		return errors.Wrap(errors.WithStack(err), "Updating permissions failed.")
	}
	chownNodesCmds := []string{chownBinDir, chownKubeCmd, chownHelmCmd, chownCniDir, chownCniConfigDir, chownKubeletConfig}
	if _, err := runtime.GetRunner().SudoCmd(strings.Join(chownNodesCmds, " && "), true, false); err != nil {
		return errors.Wrap(errors.WithStack(err), "Updating permissions failed.")
	}

	return nil
}
