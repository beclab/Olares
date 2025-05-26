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

package k3s

import (
	"context"
	"encoding/base64"
	"fmt"
	"path"
	"path/filepath"
	"strings"
	"time"

	"bytetrade.io/web3os/installer/pkg/storage"
	storagetpl "bytetrade.io/web3os/installer/pkg/storage/templates"

	"bytetrade.io/web3os/installer/pkg/container"
	"bytetrade.io/web3os/installer/pkg/manifest"
	"bytetrade.io/web3os/installer/pkg/registry"

	kubekeyapiv1alpha2 "bytetrade.io/web3os/installer/apis/kubekey/v1alpha2"
	kubekeyregistry "bytetrade.io/web3os/installer/pkg/bootstrap/registry"
	"bytetrade.io/web3os/installer/pkg/common"
	"bytetrade.io/web3os/installer/pkg/core/action"
	"bytetrade.io/web3os/installer/pkg/core/connector"
	"bytetrade.io/web3os/installer/pkg/core/logger"
	"bytetrade.io/web3os/installer/pkg/core/util"
	"bytetrade.io/web3os/installer/pkg/images"
	"bytetrade.io/web3os/installer/pkg/k3s/templates"
	"bytetrade.io/web3os/installer/pkg/utils"
	"github.com/pkg/errors"
	"github.com/shirou/gopsutil/v4/net"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	versionutil "k8s.io/apimachinery/pkg/util/version"
	kube "k8s.io/client-go/kubernetes"
	"k8s.io/client-go/tools/clientcmd"
)

type GetClusterStatus struct {
	common.KubeAction
}

func (g *GetClusterStatus) Execute(runtime connector.Runtime) error {
	exist, err := runtime.GetRunner().FileExist("/etc/systemd/system/k3s.service")
	if err != nil {
		return err
	}

	if !exist {
		g.PipelineCache.Set(common.ClusterExist, false)
		return nil
	} else {
		g.PipelineCache.Set(common.ClusterExist, true)

		if v, ok := g.PipelineCache.Get(common.ClusterStatus); ok {
			cluster := v.(*K3sStatus)
			if err := cluster.SearchVersion(runtime); err != nil {
				return err
			}
			if err := cluster.SearchKubeConfig(runtime); err != nil {
				return err
			}
			if err := cluster.LoadKubeConfig(runtime, g.KubeConf); err != nil {
				return err
			}
			if err := cluster.SearchNodeToken(runtime); err != nil {
				return err
			}
			if err := cluster.SearchInfo(runtime); err != nil {
				return err
			}
			if err := cluster.SearchNodesInfo(runtime); err != nil {
				return err
			}
			g.PipelineCache.Set(common.ClusterStatus, cluster)
		} else {
			return errors.New("get k3s cluster status by pipeline cache failed")
		}
	}
	return nil
}

type SyncKubeBinary struct {
	common.KubeAction
	manifest.ManifestAction
}

func (s *SyncKubeBinary) Execute(runtime connector.Runtime) error {
	if err := utils.ResetTmpDir(runtime); err != nil {
		return err
	}

	binaryList := []string{"k3s", "helm", "cni-plugins"} // kubecni
	for _, name := range binaryList {
		binary, err := s.Manifest.Get(name)
		if err != nil {
			return fmt.Errorf("get kube binary %s info failed: %w", name, err)
		}

		path := binary.FilePath(s.BaseDir)

		fileName := binary.Filename
		switch name {
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
			logger.Debugf("SyncKubeBinary cp %s from %s to %s", name, path, dst)
			if err := runtime.GetRunner().SudoScp(path, dst); err != nil {
				return errors.Wrap(errors.WithStack(err), fmt.Sprintf("sync kube binaries failed"))
			}
			if _, err := runtime.GetRunner().SudoCmd(fmt.Sprintf("chmod +x %s", dst), false, false); err != nil {
				return err
			}
		}
	}

	binaries := []string{"kubectl"}
	var createLinkCMDs []string
	for _, binary := range binaries {
		createLinkCMDs = append(createLinkCMDs, fmt.Sprintf("ln -snf /usr/local/bin/k3s /usr/local/bin/%s", binary))
	}
	if _, err := runtime.GetRunner().SudoCmd(strings.Join(createLinkCMDs, " && "), false, false); err != nil {
		return errors.Wrap(errors.WithStack(err), "create ctl tool link failed")
	}

	return nil
}

type ChmodScript struct {
	common.KubeAction
}

func (c *ChmodScript) Execute(runtime connector.Runtime) error {
	killAllScript := filepath.Join("/usr/local/bin", templates.K3sKillallScript.Name())
	uninstallScript := filepath.Join("/usr/local/bin", templates.K3sUninstallScript.Name())

	if _, err := runtime.GetRunner().SudoCmd(fmt.Sprintf("chmod +x %s", killAllScript),
		false, false); err != nil {
		return err
	}
	if _, err := runtime.GetRunner().SudoCmd(fmt.Sprintf("chmod +x %s", uninstallScript),
		false, false); err != nil {
		return err
	}
	return nil
}

type GenerateK3sService struct {
	common.KubeAction
}

func (g *GenerateK3sService) Execute(runtime connector.Runtime) error {
	// exist := checkContainerExists(runtime)
	host := runtime.RemoteHost()

	var server string
	if !host.IsRole(common.Master) {
		server = fmt.Sprintf("https://%s:%d", g.KubeConf.Cluster.ControlPlaneEndpoint.Domain, g.KubeConf.Cluster.ControlPlaneEndpoint.Port)
	}

	defaultKubeletArs := map[string]string{
		"kube-reserved":           "cpu=200m,memory=250Mi,ephemeral-storage=1Gi",
		"system-reserved":         "cpu=200m,memory=250Mi,ephemeral-storage=1Gi",
		"eviction-hard":           "memory.available<5%,nodefs.available<10%",
		"config":                  "/etc/rancher/k3s/kubelet.config",
		"containerd":              container.DefaultContainerdCRISocket,
		"cgroup-driver":           "systemd",
		"runtime-request-timeout": "5m",
	}
	defaultKubeProxyArgs := map[string]string{
		"proxy-mode": "ipvs",
	}

	kubeApiserverArgs, _ := util.GetArgs(map[string]string{}, g.KubeConf.Cluster.Kubernetes.ApiServerArgs)
	kubeControllerManager, _ := util.GetArgs(map[string]string{
		"terminated-pod-gc-threshold": "1",
	}, g.KubeConf.Cluster.Kubernetes.ControllerManagerArgs)
	kubeSchedulerArgs, _ := util.GetArgs(map[string]string{}, g.KubeConf.Cluster.Kubernetes.SchedulerArgs)
	kubeletArgs, _ := util.GetArgs(defaultKubeletArs, g.KubeConf.Cluster.Kubernetes.KubeletArgs)
	kubeProxyArgs, _ := util.GetArgs(defaultKubeProxyArgs, g.KubeConf.Cluster.Kubernetes.KubeProxyArgs)

	var data = util.Data{
		"Server":                 server,
		"IsMaster":               host.IsRole(common.Master),
		"NodeIP":                 host.GetInternalAddress(),
		"HostName":               host.GetName(),
		"PodSubnet":              g.KubeConf.Cluster.Network.KubePodsCIDR,
		"ServiceSubnet":          g.KubeConf.Cluster.Network.KubeServiceCIDR,
		"ClusterDns":             g.KubeConf.Cluster.CorednsClusterIP(),
		"CertSANs":               g.KubeConf.Cluster.GenerateCertSANs(),
		"PauseImage":             images.GetImage(runtime, g.KubeConf, "pause").ImageName(),
		"Container":              fmt.Sprintf("unix://%s", container.DefaultContainerdCRISocket),
		"ApiserverArgs":          kubeApiserverArgs,
		"ControllerManager":      kubeControllerManager,
		"SchedulerArgs":          kubeSchedulerArgs,
		"KubeletArgs":            kubeletArgs,
		"KubeProxyArgs":          kubeProxyArgs,
		"JuiceFSPreCheckEnabled": util.IsExist(storage.JuiceFsServiceFile),
		"JuiceFSServiceUnit":     storagetpl.JuicefsService.Name(),
		"JuiceFSBinPath":         storage.JuiceFsFile,
		"JuiceFSMountPoint":      storage.OlaresJuiceFSRootDir,
	}

	templateAction := action.Template{
		Name:     "GenerateK3sService",
		Template: templates.K3sService,
		Dst:      filepath.Join("/etc/systemd/system/", templates.K3sService.Name()),
		Data:     data,
	}

	templateAction.Init(nil, nil)
	if err := templateAction.Execute(runtime); err != nil {
		return err
	}

	templateAction = action.Template{
		Name:     "K3sKubeletConfig",
		Template: templates.K3sKubeletConfig,
		Dst:      filepath.Join("/etc/rancher/k3s/", templates.K3sKubeletConfig.Name()),
		Data: util.Data{
			"ShutdownGracePeriod":             g.KubeConf.Cluster.Kubernetes.ShutdownGracePeriod,
			"ShutdownGracePeriodCriticalPods": g.KubeConf.Cluster.Kubernetes.ShutdownGracePeriodCriticalPods,
			"MaxPods":                         g.KubeConf.Cluster.Kubernetes.MaxPods,
			"EnablePodSwap":                   g.KubeConf.Arg.EnablePodSwap,
		},
	}

	templateAction.Init(nil, nil)
	if err := templateAction.Execute(runtime); err != nil {
		return err
	}

	return nil
}

type GenerateK3sServiceEnv struct {
	common.KubeAction
}

func (g *GenerateK3sServiceEnv) Execute(runtime connector.Runtime) error {
	host := runtime.RemoteHost()

	clusterStatus, ok := g.PipelineCache.Get(common.ClusterStatus)
	if !ok {
		return errors.New("get cluster status by pipeline cache failed")
	}
	cluster := clusterStatus.(*K3sStatus)

	var externalEtcd kubekeyapiv1alpha2.ExternalEtcd
	var endpointsList []string
	var externalEtcdEndpoints, token string

	switch g.KubeConf.Cluster.Etcd.Type {
	case kubekeyapiv1alpha2.External:
		externalEtcd.Endpoints = g.KubeConf.Cluster.Etcd.External.Endpoints

		if len(g.KubeConf.Cluster.Etcd.External.CAFile) != 0 && len(g.KubeConf.Cluster.Etcd.External.CAFile) != 0 && len(g.KubeConf.Cluster.Etcd.External.CAFile) != 0 {
			externalEtcd.CAFile = fmt.Sprintf("/etc/ssl/etcd/ssl/%s", filepath.Base(g.KubeConf.Cluster.Etcd.External.CAFile))
			externalEtcd.CertFile = fmt.Sprintf("/etc/ssl/etcd/ssl/%s", filepath.Base(g.KubeConf.Cluster.Etcd.External.CertFile))
			externalEtcd.KeyFile = fmt.Sprintf("/etc/ssl/etcd/ssl/%s", filepath.Base(g.KubeConf.Cluster.Etcd.External.KeyFile))
		}
	default:
		for _, node := range runtime.GetHostsByRole(common.ETCD) {
			endpoint := fmt.Sprintf("https://%s:%s", node.GetInternalAddress(), kubekeyapiv1alpha2.DefaultEtcdPort)
			endpointsList = append(endpointsList, endpoint)
		}
		externalEtcd.Endpoints = endpointsList

		externalEtcd.CAFile = "/etc/ssl/etcd/ssl/ca.pem"
		externalEtcd.CertFile = fmt.Sprintf("/etc/ssl/etcd/ssl/node-%s.pem", runtime.GetHostsByRole(common.Master)[0].GetName())
		externalEtcd.KeyFile = fmt.Sprintf("/etc/ssl/etcd/ssl/node-%s-key.pem", runtime.GetHostsByRole(common.Master)[0].GetName())
	}

	externalEtcdEndpoints = strings.Join(externalEtcd.Endpoints, ",")

	v121 := versionutil.MustParseSemantic("v1.21.0")
	atLeast := versionutil.MustParseSemantic(g.KubeConf.Cluster.Kubernetes.Version).AtLeast(v121)
	if atLeast {
		token = cluster.NodeToken
	} else {
		if !host.IsRole(common.Master) {
			token = cluster.NodeToken
		}
	}

	templateAction := action.Template{
		Name:     "K3sServiceEnv",
		Template: templates.K3sServiceEnv,
		Dst:      filepath.Join("/etc/systemd/system/", templates.K3sServiceEnv.Name()),
		Data: util.Data{
			"DataStoreEndPoint": externalEtcdEndpoints,
			"DataStoreCaFile":   externalEtcd.CAFile,
			"DataStoreCertFile": externalEtcd.CertFile,
			"DataStoreKeyFile":  externalEtcd.KeyFile,
			"IsMaster":          host.IsRole(common.Master),
			"Token":             token,
		},
	}

	templateAction.Init(nil, nil)
	if err := templateAction.Execute(runtime); err != nil {
		return err
	}
	return nil
}

type EnableK3sService struct {
	common.KubeAction
}

func (e *EnableK3sService) Execute(runtime connector.Runtime) error {
	if _, err := runtime.GetRunner().SudoCmd("systemctl daemon-reload && systemctl enable --now k3s",
		false, false); err != nil {
		return errors.Wrap(errors.WithStack(err), "enable k3s failed")
	}
	return nil
}

// type PreloadImagesService struct {
// 	common.KubeAction
// }

// func (p *PreloadImagesService) Execute(runtime connector.Runtime) error {
// 	if utils.IsExist(common.K3sImageDir) {
// 		if err := util.CreateDir(common.K3sImageDir); err != nil {
// 			logger.Errorf("create dir %s failed: %v", common.K3sImageDir, err)
// 			return err
// 		}
// 	}

// 	fileInfos, err := os.ReadDir(common.K3sImageDir)
// 	if err != nil {
// 		logger.Errorf("Unable to read images in %s: %v", common.K3sImageDir, err)
// 		return nil
// 	}

// 	var loadingImages images.LocalImages
// 	for _, fileInfo := range fileInfos {
// 		if fileInfo.IsDir() {
// 			continue
// 		}

// 		filePath := filepath.Join(common.K3sImageDir, fileInfo.Name())

// 		loadingImages = append(loadingImages, images.LocalImage{Filename: filePath})
// 	}

// 	if err := loadingImages.LoadImages(runtime, p.KubeConf); err != nil {
// 		return errors.Wrap(errors.WithStack(err), "preload image failed")
// 	}
// 	return nil
// }

type CopyK3sKubeConfig struct {
	common.KubeAction
}

func (c *CopyK3sKubeConfig) Execute(runtime connector.Runtime) error {
	createConfigDirCmd := "mkdir -p /root/.kube && mkdir -p $HOME/.kube"
	getKubeConfigCmd := "cp -f /etc/rancher/k3s/k3s.yaml /root/.kube/config"
	chmodKubeConfigCmd := "chmod 0600 /root/.kube/config"

	cmd := strings.Join([]string{createConfigDirCmd, getKubeConfigCmd, chmodKubeConfigCmd}, " && ")
	if _, err := runtime.GetRunner().SudoCmd(cmd, false, false); err != nil {
		return errors.Wrap(errors.WithStack(err), "copy k3s kube config failed")
	}

	userMkdir := "mkdir -p $HOME/.kube"
	if _, err := runtime.GetRunner().Cmd(userMkdir, false, false); err != nil {
		return errors.Wrap(errors.WithStack(err), "user mkdir $HOME/.kube failed")
	}

	userCopyKubeConfig := "cp -f /etc/rancher/k3s/k3s.yaml $HOME/.kube/config"
	if _, err := runtime.GetRunner().SudoCmd(userCopyKubeConfig, false, false); err != nil {
		return errors.Wrap(errors.WithStack(err), "user copy /etc/rancher/k3s/k3s.yaml to $HOME/.kube/config failed")
	}

	if _, err := runtime.GetRunner().SudoCmd("chmod 0600 $HOME/.kube/config", false, false); err != nil {
		return errors.Wrap(errors.WithStack(err), "chmod k3s $HOME/.kube/config 0600 failed")
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

type AddMasterTaint struct {
	common.KubeAction
}

func (a *AddMasterTaint) Execute(runtime connector.Runtime) error {
	host := runtime.RemoteHost()

	cmd := fmt.Sprintf(
		"/usr/local/bin/kubectl taint nodes %s node-role.kubernetes.io/master=effect:NoSchedule --overwrite",
		host.GetName())

	if _, err := runtime.GetRunner().SudoCmd(cmd, false, false); err != nil {
		return errors.Wrap(errors.WithStack(err), "add master NoSchedule taint failed")
	}
	return nil
}

type AddWorkerLabel struct {
	common.KubeAction
}

func (a *AddWorkerLabel) Execute(runtime connector.Runtime) error {
	host := runtime.RemoteHost()

	cmd := fmt.Sprintf(
		"/usr/local/bin/kubectl label --overwrite node %s node-role.kubernetes.io/worker=",
		host.GetName())

	var out string
	var err error
	if out, err = runtime.GetRunner().SudoCmd(cmd, false, false); err != nil {
		return fmt.Errorf("waiting for node ready")
		// return errors.Wrap(errors.WithStack(err), "add master NoSchedule taint failed")
	}
	logger.Debugf("AddWorkerLabel successed: %s", out)
	return nil
}

type SyncKubeConfigToWorker struct {
	common.KubeAction
}

func (s *SyncKubeConfigToWorker) Execute(runtime connector.Runtime) error {
	if v, ok := s.PipelineCache.Get(common.ClusterStatus); ok {
		cluster := v.(*K3sStatus)

		createConfigDirCmd := "mkdir -p /root/.kube"
		if _, err := runtime.GetRunner().SudoCmd(createConfigDirCmd, false, false); err != nil {
			return errors.Wrap(errors.WithStack(err), "create .kube dir failed")
		}

		oldServer := "server: https://127.0.0.1:6443"
		newServer := fmt.Sprintf("server: https://%s:%d",
			s.KubeConf.Cluster.ControlPlaneEndpoint.Domain,
			s.KubeConf.Cluster.ControlPlaneEndpoint.Port)
		newKubeConfig := strings.Replace(cluster.KubeConfig, oldServer, newServer, -1)

		syncKubeConfigForRootCmd := fmt.Sprintf("echo '%s' > %s", newKubeConfig, "/root/.kube/config")
		if _, err := runtime.GetRunner().SudoCmd(syncKubeConfigForRootCmd, false, false); err != nil {
			return errors.Wrap(errors.WithStack(err), "sync kube config for root failed")
		}

		if _, err := runtime.GetRunner().SudoCmd("chmod 0600 /root/.kube/config", false, false); err != nil {
			return errors.Wrap(errors.WithStack(err), "chmod k3s $HOME/.kube/config failed")
		}

		userConfigDirCmd := "mkdir -p $HOME/.kube"
		if _, err := runtime.GetRunner().Cmd(userConfigDirCmd, false, false); err != nil {
			return errors.Wrap(errors.WithStack(err), "user mkdir $HOME/.kube failed")
		}

		syncKubeConfigForUserCmd := fmt.Sprintf("echo '%s' > %s", newKubeConfig, "$HOME/.kube/config")
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

type ExecKillAllScript struct {
	common.KubeAction
}

func (t *ExecKillAllScript) Execute(runtime connector.Runtime) error {
	if _, err := runtime.GetRunner().SudoCmd("systemctl daemon-reload && /usr/local/bin/k3s-killall.sh",
		true, false); err != nil {
		return errors.Wrap(errors.WithStack(err), "add master NoSchedule taint failed")
	}
	return nil
}

type ExecUninstallScript struct {
	common.KubeAction
}

func (e *ExecUninstallScript) Execute(runtime connector.Runtime) error {
	if _, err := runtime.GetRunner().SudoCmd("systemctl daemon-reload && /usr/local/bin/k3s-uninstall.sh",
		true, false); err != nil {
		return errors.Wrap(errors.WithStack(err), "add master NoSchedule taint failed")
	}
	return nil
}

type SaveKubeConfig struct {
	common.KubeAction
}

func (s *SaveKubeConfig) Execute(_ connector.Runtime) error {
	status, ok := s.PipelineCache.Get(common.ClusterStatus)
	if !ok {
		return errors.New("get kubernetes status failed by pipeline cache")
	}
	cluster := status.(*K3sStatus)

	oldServer := fmt.Sprintf("https://%s:%d", s.KubeConf.Cluster.ControlPlaneEndpoint.Domain, s.KubeConf.Cluster.ControlPlaneEndpoint.Port)
	newServer := fmt.Sprintf("https://%s:%d", s.KubeConf.Cluster.ControlPlaneEndpoint.Address, s.KubeConf.Cluster.ControlPlaneEndpoint.Port)
	newKubeConfigStr := strings.Replace(cluster.KubeConfig, oldServer, newServer, -1)
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
		Create(context.TODO(), namespace, metav1.CreateOptions{}); err != nil {
		// return err
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
		Create(context.TODO(), cm, metav1.CreateOptions{}); err != nil {
		// return err
	}
	return nil
}

type GenerateK3sRegistryConfig struct {
	common.KubeAction
}

func (g *GenerateK3sRegistryConfig) Execute(runtime connector.Runtime) error {
	dockerioMirror := registry.Mirror{}
	registryConfigs := map[string]registry.RegistryConfig{}

	auths := registry.DockerRegistryAuthEntries(g.KubeConf.Cluster.Registry.Auths)

	dockerioMirror.Endpoints = g.KubeConf.Cluster.Registry.RegistryMirrors

	if g.KubeConf.Cluster.Registry.NamespaceOverride != "" {
		dockerioMirror.Rewrites = map[string]string{
			"^rancher/(.*)": fmt.Sprintf("%s/$1", g.KubeConf.Cluster.Registry.NamespaceOverride),
		}
	}

	for k, v := range auths {
		registryConfigs[k] = registry.RegistryConfig{
			Auth: &registry.AuthConfig{
				Username: v.Username,
				Password: v.Password,
			},
			TLS: &registry.TLSConfig{
				CAFile:             v.CAFile,
				CertFile:           v.CertFile,
				KeyFile:            v.KeyFile,
				InsecureSkipVerify: v.SkipTLSVerify,
			},
		}
	}

	_, ok := registryConfigs[kubekeyregistry.RegistryCertificateBaseName]

	if !ok && g.KubeConf.Cluster.Registry.PrivateRegistry == kubekeyregistry.RegistryCertificateBaseName {
		registryConfigs[g.KubeConf.Cluster.Registry.PrivateRegistry] = registry.RegistryConfig{TLS: &registry.TLSConfig{InsecureSkipVerify: true}}
	}

	k3sRegistries := registry.Registry{
		Mirrors: map[string]registry.Mirror{"docker.io": dockerioMirror},
		Configs: registryConfigs,
	}

	templateAction := action.Template{
		Name:     "K3sRegistryConfigTempl",
		Template: templates.K3sRegistryConfigTempl,
		Dst:      filepath.Join("/etc/rancher/k3s", templates.K3sRegistryConfigTempl.Name()),
		Data: util.Data{
			"Registries": k3sRegistries,
		},
	}

	templateAction.Init(nil, nil)
	if err := templateAction.Execute(runtime); err != nil {
		return err
	}
	return nil
}

type UninstallK3s struct {
	common.KubeAction
}

func (t *UninstallK3s) Execute(runtime connector.Runtime) error {
	var scriptPath = path.Join(common.BinDir, "k3s-uninstall.sh")
	if _, err := runtime.GetRunner().SudoCmd(scriptPath, false, true); err != nil {
		return err
	}
	return nil
}

type DeleteCalicoCNI struct {
	common.KubeAction
}

func (t *DeleteCalicoCNI) Execute(runtime connector.Runtime) error {
	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Minute)
	defer cancel()

	ifInfo, _ := net.InterfacesWithContext(ctx)
	if ifInfo == nil {
		return nil
	}

	for _, i := range ifInfo {
		var name = i.Name
		if len(name) < 5 || name[0:4] != "cali" {
			continue
		}
		if _, err := runtime.GetRunner().Cmd(fmt.Sprintf("ip link delete %s", name), false, false); err != nil {
			logger.Errorf("delete ip link %s error %v", name, err)
		}
		time.Sleep(200 * time.Millisecond)
	}

	return nil
}
