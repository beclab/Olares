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

package kubesphere

import (
	kubekeyapiv1alpha2 "bytetrade.io/web3os/installer/apis/kubekey/v1alpha2"
	"bytetrade.io/web3os/installer/pkg/common"
	"bytetrade.io/web3os/installer/pkg/core/connector"
	"bytetrade.io/web3os/installer/pkg/core/logger"
	"bytetrade.io/web3os/installer/pkg/core/util"
	ksv2 "bytetrade.io/web3os/installer/pkg/kubesphere/v2"
	ksv3 "bytetrade.io/web3os/installer/pkg/kubesphere/v3"
	"bytetrade.io/web3os/installer/pkg/version/kubesphere"
	"bytetrade.io/web3os/installer/pkg/version/kubesphere/templates"
	"fmt"
	"github.com/pkg/errors"
	yamlV2 "gopkg.in/yaml.v2"
	"os"
	"path"
	"path/filepath"
	"strings"
)

type DeleteKubeSphereCaches struct {
	common.KubeAction
}

func (d *DeleteKubeSphereCaches) Execute(runtime connector.Runtime) error {
	var files = []string{
		path.Join(runtime.GetInstallerDir(), "files"),
		path.Join(runtime.GetInstallerDir(), "cli"),
	}

	for _, f := range files {
		if _, err := runtime.GetRunner().SudoCmd(fmt.Sprintf("rm -rf %s", f), false, true); err != nil {
			return errors.Wrapf(errors.WithStack(err), "delete %s failed", f)
		}
	}

	return nil
}

type DeleteCache struct {
	common.KubeAction
}

func (t *DeleteCache) Execute(runtime connector.Runtime) error {
	// var cacheDir = path.Join(runtime.GetBaseDir(), cc.ImagesDir)
	// if err := util.RemoveDir(cacheDir); err != nil {
	// 	return err
	// }
	// logger.Debugf("delete caches success")
	return nil
}

type AddInstallerConfig struct {
	common.KubeAction
}

func (a *AddInstallerConfig) Execute(runtime connector.Runtime) error {
	//var ksFilename string

	// if runtime.GetSystemInfo().IsDarwin() {
	// ksFilename = path.Join(common.TmpDir, "/etc/kubernetes/addons/kubesphere.yaml")
	// } else {
	//ksFilename = "/etc/kubernetes/addons/kubesphere.yaml"
	//// }
	//configurationBase64 := base64.StdEncoding.EncodeToString([]byte(a.KubeConf.Cluster.KubeSphere.Configurations))
	//if _, err := runtime.GetRunner().SudoCmd(
	//	fmt.Sprintf("echo %s | base64 -d >> %s", configurationBase64, ksFilename),
	//	false, false); err != nil {
	//	return errors.Wrap(errors.WithStack(err), "add config to ks-installer manifests failed")
	//}
	return nil
}

type CreateNamespace struct {
	common.KubeAction
}

func (c *CreateNamespace) Execute(runtime connector.Runtime) error {
	var kubectl, ok = c.PipelineCache.GetMustString(common.CacheCommandKubectlPath)
	if !ok || kubectl == "" {
		kubectl = path.Join(common.BinDir, "kubectl")
	}

	var cmd = fmt.Sprintf(`cat <<EOF | %s apply -f -
apiVersion: v1
kind: Namespace
metadata:
  name: kubesphere-system
---
apiVersion: v1
kind: Namespace
metadata:
  name: kubesphere-monitoring-system
EOF`, kubectl)
	_, err := runtime.GetRunner().SudoCmd(cmd, false, true)
	if err != nil {
		return errors.Wrap(errors.WithStack(err), "create namespace: kubesphere-system and kubesphere-monitoring-system")
	}
	return nil
}

type Setup struct {
	common.KubeAction
}

func (s *Setup) Execute(runtime connector.Runtime) error {
	nodeIp, _ := s.PipelineCache.GetMustString(common.CacheMinikubeNodeIp)
	filePath := filepath.Join(common.KubeAddonsDir, templates.KsInstaller.Name())

	var minikubepath, ok = s.PipelineCache.GetMustString(common.CacheCommandMinikubePath)
	if !ok || minikubepath == "" {
		minikubepath = path.Join(common.BinDir, common.CommandMinikube)
	}

	kubectlpath, ok := s.PipelineCache.GetMustString(common.CacheCommandKubectlPath)
	if !ok || kubectlpath == "" {
		kubectlpath = path.Join(common.BinDir, common.CommandKubectl)
	}

	var addrList []string
	var tlsDisable bool
	var port string
	switch s.KubeConf.Cluster.Etcd.Type {
	case kubekeyapiv1alpha2.KubeKey:
		for _, host := range runtime.GetHostsByRole(common.ETCD) {
			addrList = append(addrList, host.GetInternalAddress())
		}

		caFile := "/etc/ssl/etcd/ssl/ca.pem"
		certFile := fmt.Sprintf("/etc/ssl/etcd/ssl/node-%s.pem", runtime.RemoteHost().GetName())
		keyFile := fmt.Sprintf("/etc/ssl/etcd/ssl/node-%s-key.pem", runtime.RemoteHost().GetName())
		if output, err := runtime.GetRunner().SudoCmd(
			fmt.Sprintf("/usr/local/bin/kubectl -n kubesphere-monitoring-system create secret generic kube-etcd-client-certs "+
				"--from-file=etcd-client-ca.crt=%s "+
				"--from-file=etcd-client.crt=%s "+
				"--from-file=etcd-client.key=%s", caFile, certFile, keyFile), false, false); err != nil {
			if !strings.Contains(output, "exists") {
				return err
			}
		}
	case kubekeyapiv1alpha2.MiniKube:
		var etcdPath = common.KubeEtcdCertDir // path.Join(common.TmpDir, common.KubeEtcdCertDir)
		if !util.IsExist(etcdPath) {
			if _, err := runtime.GetRunner().SudoCmd(fmt.Sprintf("mkdir -p %s", etcdPath), false, false); err != nil {
				return err
			}
		}
		var certfiles = []string{
			"ca.crt",
			"server.crt",
			"server.key",
		}

		for _, certfile := range certfiles {
			var cfile = path.Join(common.MinikubeEtcdCertDir, certfile)
			var cmd = fmt.Sprintf("%s -p %s ssh sudo chmod 644 %s && minikube -p %s cp %s:%s %s", minikubepath,
				runtime.RemoteHost().GetMinikubeProfile(), cfile,
				runtime.RemoteHost().GetMinikubeProfile(), runtime.RemoteHost().GetMinikubeProfile(),
				cfile, path.Join(etcdPath, certfile))
			if _, err := runtime.GetRunner().SudoCmd(cmd, false, false); err != nil {
				return err
			}
		}

		caFile := path.Join(etcdPath, "ca.crt")
		certFile := path.Join(etcdPath, "server.crt")
		keyFile := path.Join(etcdPath, "server.key")

		addrList = append(addrList, nodeIp)
		if output, err := runtime.GetRunner().SudoCmd(
			fmt.Sprintf("%s -n kubesphere-monitoring-system create secret generic kube-etcd-client-certs "+
				"--from-file=%s "+
				"--from-file=%s "+
				"--from-file=%s", kubectlpath, caFile, certFile, keyFile), false, false); err != nil {
			if !strings.Contains(output, "already exists") {
				return err
			}
		}

		//path.Join(common.TmpDir, filepath.Join(common.KubeAddonsDir, templates.KsInstaller.Name()))
		filePath = path.Join(filepath.Join(common.KubeAddonsDir, templates.KsInstaller.Name()))
	case kubekeyapiv1alpha2.Kubeadm:
		for _, host := range runtime.GetHostsByRole(common.Master) {
			addrList = append(addrList, host.GetInternalAddress())
		}

		caFile := "/etc/kubernetes/pki/etcd/ca.crt"
		certFile := "/etc/kubernetes/pki/etcd/healthcheck-client.crt"
		keyFile := "/etc/kubernetes/pki/etcd/healthcheck-client.key"
		if output, err := runtime.GetRunner().SudoCmd(
			fmt.Sprintf("/usr/local/bin/kubectl -n kubesphere-monitoring-system create secret generic kube-etcd-client-certs "+
				"--from-file=etcd-client-ca.crt=%s "+
				"--from-file=etcd-client.crt=%s "+
				"--from-file=etcd-client.key=%s", caFile, certFile, keyFile), false, false); err != nil {
			if !strings.Contains(output, "exists") {
				return err
			}
		}
	case kubekeyapiv1alpha2.External:
		for _, endpoint := range s.KubeConf.Cluster.Etcd.External.Endpoints {
			e := strings.Split(strings.TrimSpace(endpoint), "://")
			s := strings.Split(e[1], ":")
			port = s[1]
			addrList = append(addrList, s[0])
			if e[0] == "http" {
				tlsDisable = true
			}
		}
		if tlsDisable {
			if output, err := runtime.GetRunner().SudoCmd("/usr/local/bin/kubectl -n kubesphere-monitoring-system create secret generic kube-etcd-client-certs", true, false); err != nil {
				if !strings.Contains(output, "exists") {
					return err
				}
			}
		} else {
			caFile := fmt.Sprintf("/etc/ssl/etcd/ssl/%s", filepath.Base(s.KubeConf.Cluster.Etcd.External.CAFile))
			certFile := fmt.Sprintf("/etc/ssl/etcd/ssl/%s", filepath.Base(s.KubeConf.Cluster.Etcd.External.CertFile))
			keyFile := fmt.Sprintf("/etc/ssl/etcd/ssl/%s", filepath.Base(s.KubeConf.Cluster.Etcd.External.KeyFile))
			if output, err := runtime.GetRunner().SudoCmd(
				fmt.Sprintf("/usr/local/bin/kubectl -n kubesphere-monitoring-system create secret generic kube-etcd-client-certs "+
					"--from-file=etcd-client-ca.crt=%s "+
					"--from-file=etcd-client.crt=%s "+
					"--from-file=etcd-client.key=%s", caFile, certFile, keyFile), true, false); err != nil {
				if !strings.Contains(output, "exists") {
					return err
				}
			}
		}
	}

	var sedCommand = runtime.GetCommandSed()
	etcdEndPoint := strings.Join(addrList, ",")
	var cmdEndpoint = fmt.Sprintf("%s '/endpointIps/s/\\:.*/\\: %s/g' %s", sedCommand, etcdEndPoint, filePath)
	if _, err := runtime.GetRunner().SudoCmd(cmdEndpoint, false, false); err != nil {
		return errors.Wrap(errors.WithStack(err), fmt.Sprintf("update etcd endpoint failed"))
	}

	if tlsDisable {
		if _, err := runtime.GetRunner().SudoCmd(
			fmt.Sprintf("%s '/tlsEnable/s/\\:.*/\\: false/g' %s", sedCommand, filePath),
			false, false); err != nil {
			return errors.Wrap(errors.WithStack(err), fmt.Sprintf("update etcd tls failed"))
		}
	}

	if len(port) != 0 {
		if _, err := runtime.GetRunner().SudoCmd(
			fmt.Sprintf("%s 's/2379/%s/g' %s", sedCommand, port, filePath),
			false, false); err != nil {
			return errors.Wrap(errors.WithStack(err), fmt.Sprintf("update etcd tls failed"))
		}
	}

	if s.KubeConf.Cluster.Registry.PrivateRegistry != "" {
		PrivateRegistry := strings.Replace(s.KubeConf.Cluster.Registry.PrivateRegistry, "/", "\\/", -1)
		if _, err := runtime.GetRunner().SudoCmd(
			fmt.Sprintf("%s '/local_registry/s/\\:.*/\\: %s/g' %s", sedCommand, PrivateRegistry, filePath),
			false, false); err != nil {
			return errors.Wrap(errors.WithStack(err), fmt.Sprintf("add private registry: %s failed", s.KubeConf.Cluster.Registry.PrivateRegistry))
		}
	} else {
		if _, err := runtime.GetRunner().SudoCmd(fmt.Sprintf("%s '/local_registry/d' %s", sedCommand, filePath), false, false); err != nil {
			return errors.Wrap(errors.WithStack(err), fmt.Sprintf("remove private registry failed"))
		}
	}

	if s.KubeConf.Cluster.Registry.NamespaceOverride != "" {
		if _, err := runtime.GetRunner().SudoCmd(
			fmt.Sprintf("%s '/namespace_override/s/\\:.*/\\: %s/g' %s", sedCommand, s.KubeConf.Cluster.Registry.NamespaceOverride, filePath),
			false, false); err != nil {
			return errors.Wrap(errors.WithStack(err), fmt.Sprintf("add namespace override: %s failed", s.KubeConf.Cluster.Registry.NamespaceOverride))
		}
	} else {
		if _, err := runtime.GetRunner().SudoCmd(fmt.Sprintf("%s '/namespace_override/d' %s", sedCommand, filePath), false, false); err != nil {
			return errors.Wrap(errors.WithStack(err), fmt.Sprintf("remove namespace override failed"))
		}
	}

	_, ok = kubesphere.CNSource[s.KubeConf.Cluster.KubeSphere.Version]
	if ok && (os.Getenv("KKZONE") == "cn" || s.KubeConf.Cluster.Registry.PrivateRegistry == "registry.cn-beijing.aliyuncs.com") {
		if _, err := runtime.GetRunner().SudoCmd(
			fmt.Sprintf("%s '/zone/s/\\:.*/\\: %s/g' %s", sedCommand, "cn", filePath),
			false, false); err != nil {
			return errors.Wrap(errors.WithStack(err), fmt.Sprintf("add kubekey zone: %s failed", s.KubeConf.Cluster.Registry.PrivateRegistry))
		}
	} else {
		if _, err := runtime.GetRunner().SudoCmd(
			fmt.Sprintf("%s '/zone/d' %s", sedCommand, filePath),
			false, false); err != nil {
			return errors.Wrap(errors.WithStack(err), fmt.Sprintf("remove kubekey zone failed"))
		}
	}

	switch s.KubeConf.Cluster.Kubernetes.ContainerManager {
	case "docker", "containerd", "crio":
		if _, err := runtime.GetRunner().SudoCmd(
			fmt.Sprintf("%s '/containerruntime/s/\\:.*/\\: %s/g' %s", sedCommand, s.KubeConf.Cluster.Kubernetes.ContainerManager, filePath), false, false); err != nil {
			return errors.Wrap(errors.WithStack(err), fmt.Sprintf("set container runtime: %s failed", s.KubeConf.Cluster.Kubernetes.ContainerManager))
		}
	default:
		logger.Infof(
			fmt.Sprintf("%s Currently, the logging module of KubeSphere does not support %s. If %s is used, the logging module will be unavailable.", runtime.RemoteHost().GetName(),
				s.KubeConf.Cluster.Kubernetes.ContainerManager, s.KubeConf.Cluster.Kubernetes.ContainerManager))
	}

	return nil
}

type Apply struct {
	common.KubeAction
}

func (a *Apply) Execute(runtime connector.Runtime) error {
	var kubectlpath, ok = a.PipelineCache.GetMustString(common.CacheCommandKubectlPath)
	if !ok || kubectlpath == "" {
		kubectlpath = path.Join(common.BinDir, common.CommandKubectl)
	}

	filePath := filepath.Join(common.KubeAddonsDir, templates.KsInstaller.Name())
	// if runtime.GetSystemInfo().IsDarwin() {
	// 	filePath = path.Join(common.TmpDir, filePath)
	// }

	deployKubesphereCmd := fmt.Sprintf("%s apply -f %s --force", kubectlpath, filePath)
	if _, err := runtime.GetRunner().Cmd(deployKubesphereCmd, false, true); err != nil {
		return errors.Wrapf(errors.WithStack(err), "deploy %s failed", filePath)
	}
	return nil
}

type GetKubeCommand struct {
	common.KubeAction
}

func (t *GetKubeCommand) Execute(runtime connector.Runtime) error {
	kubectlpath, err := util.GetCommand(common.CommandKubectl)
	if err != nil || kubectlpath == "" {
		return fmt.Errorf("kubectl not found")
	}

	t.PipelineCache.Set(common.CacheCommandKubectlPath, kubectlpath)
	logger.InfoInstallationProgress("k8s and kubesphere installation is complete")
	return nil
}

type Check struct {
	common.KubeAction
}

func (c *Check) Execute(runtime connector.Runtime) error {
	var kubectlpath, err = util.GetCommand(common.CommandKubectl)
	if err != nil {
		return fmt.Errorf("kubectl not found")
	}

	var labels = []string{"app=ks-apiserver", "app=ks-controller-manager"}

	for _, label := range labels {
		var cmd = fmt.Sprintf("%s get pod -n %s -l '%s' -o jsonpath='{.items[0].status.phase}'", kubectlpath, common.NamespaceKubesphereSystem, label)
		rphase, _ := runtime.GetRunner().SudoCmd(cmd, false, false)
		if rphase != "Running" {
			return errors.New("Waiting for KubeSphere to be Running")
		}
	}

	//if runtime.GetSystemInfo().IsDarwin() {
	//	epIPCMD := fmt.Sprintf("%s -n kubesphere-system get ep ks-controller-manager -o jsonpath='{.subsets[*].addresses[*].ip}'", kubectlpath)
	//	epIP, _ := runtime.GetRunner().SudoCmd(epIPCMD, false, false)
	//	if net.ParseIP(strings.TrimSpace(epIP)) == nil {
	//		return errors.New("Waiting for ks-controller-manager svc endpoints to be populated")
	//	}
	//	// we can't check the svc connectivity in macOS host
	//	// so just wait for some time for the proxy to take effect
	//	time.Sleep(5 * time.Second)
	//	return nil
	//}
	//
	//svcIPCMD := fmt.Sprintf("%s -n kubesphere-system get svc ks-controller-manager -o jsonpath='{.spec.clusterIP}'", kubectlpath)
	//svcIP, err := runtime.GetRunner().SudoCmd(svcIPCMD, false, false)
	//if err != nil {
	//	return errors.New("Waiting for ks-controller-manager service to be reachable")
	//}
	//
	//conn, err := net.DialTimeout("tcp", net.JoinHostPort(svcIP, strconv.Itoa(443)), 10*time.Second)
	//if err != nil {
	//	return errors.New("Waiting for ks-controller-manager service to be reachable")
	//}
	//defer conn.Close()
	return nil
}

type CleanCC struct {
	common.KubeAction
}

func (c *CleanCC) Execute(runtime connector.Runtime) error {
	c.KubeConf.Cluster.KubeSphere.Configurations = "\n"
	return nil
}

type ConvertV2ToV3 struct {
	common.KubeAction
}

func (c *ConvertV2ToV3) Execute(runtime connector.Runtime) error {
	configV2Str, err := runtime.GetRunner().SudoCmd(
		"/usr/local/bin/kubectl get cm -n kubesphere-system ks-installer -o jsonpath='{.data.ks-config\\.yaml}'",
		false, false)
	if err != nil {
		return err
	}

	clusterCfgV2 := ksv2.V2{}
	clusterCfgV3 := ksv3.V3{}
	if err := yamlV2.Unmarshal([]byte(configV2Str), &clusterCfgV2); err != nil {
		return err
	}

	configV3, err := MigrateConfig2to3(&clusterCfgV2, &clusterCfgV3)
	if err != nil {
		return err
	}
	c.KubeConf.Cluster.KubeSphere.Configurations = "---\n" + configV3
	return nil
}

func MigrateConfig2to3(v2 *ksv2.V2, v3 *ksv3.V3) (string, error) {
	v3.Etcd = ksv3.Etcd(v2.Etcd)
	v3.Persistence = ksv3.Persistence(v2.Persistence)
	v3.Alerting = ksv3.Alerting(v2.Alerting)
	v3.Notification = ksv3.Notification(v2.Notification)
	v3.LocalRegistry = v2.LocalRegistry
	v3.Servicemesh = ksv3.Servicemesh(v2.Servicemesh)
	v3.Devops = ksv3.Devops(v2.Devops)
	v3.Openpitrix = ksv3.Openpitrix(v2.Openpitrix)
	v3.Console = ksv3.Console(v2.Console)

	if v2.MetricsServerNew.Enabled == "" {
		if v2.MetricsServerOld.Enabled == "true" || v2.MetricsServerOld.Enabled == "True" {
			v3.MetricsServer.Enabled = true
		} else {
			v3.MetricsServer.Enabled = false
		}
	} else {
		if v2.MetricsServerNew.Enabled == "true" || v2.MetricsServerNew.Enabled == "True" {
			v3.MetricsServer.Enabled = true
		} else {
			v3.MetricsServer.Enabled = false
		}
	}

	v3.Monitoring.PrometheusMemoryRequest = v2.Monitoring.PrometheusMemoryRequest
	//v3.Monitoring.PrometheusReplicas = v2.Monitoring.PrometheusReplicas
	v3.Monitoring.PrometheusVolumeSize = v2.Monitoring.PrometheusVolumeSize
	//v3.Monitoring.AlertmanagerReplicas = 1

	v3.Common.EtcdVolumeSize = v2.Common.EtcdVolumeSize
	v3.Common.MinioVolumeSize = v2.Common.MinioVolumeSize
	v3.Common.MysqlVolumeSize = v2.Common.MysqlVolumeSize
	v3.Common.OpenldapVolumeSize = v2.Common.OpenldapVolumeSize
	v3.Common.RedisVolumSize = v2.Common.RedisVolumSize
	//v3.Common.ES.ElasticsearchDataReplicas = v2.Logging.ElasticsearchDataReplicas
	//v3.Common.ES.ElasticsearchMasterReplicas = v2.Logging.ElasticsearchMasterReplicas
	v3.Common.ES.ElkPrefix = v2.Logging.ElkPrefix
	v3.Common.ES.LogMaxAge = v2.Logging.LogMaxAge
	if v2.Logging.ElasticsearchVolumeSize == "" {
		v3.Common.ES.ElasticsearchDataVolumeSize = v2.Logging.ElasticsearchDataVolumeSize
		v3.Common.ES.ElasticsearchMasterVolumeSize = v2.Logging.ElasticsearchMasterVolumeSize
	} else {
		v3.Common.ES.ElasticsearchMasterVolumeSize = "4Gi"
		v3.Common.ES.ElasticsearchDataVolumeSize = v2.Logging.ElasticsearchVolumeSize
	}

	v3.Logging.Enabled = v2.Logging.Enabled
	v3.Logging.LogsidecarReplicas = v2.Logging.LogsidecarReplicas

	v3.Authentication.JwtSecret = ""
	v3.Multicluster.ClusterRole = "none"
	v3.Events.Ruler.Replicas = 2

	var clusterConfiguration = ksv3.ClusterConfig{
		ApiVersion: "installer.kubesphere.io/v1alpha1",
		Kind:       "ClusterConfiguration",
		Metadata: ksv3.Metadata{
			Name:      "ks-installer",
			Namespace: "kubesphere-system",
			Label:     ksv3.Label{Version: "v3.0.0"},
		},
		Spec: v3,
	}

	configV3, err := yamlV2.Marshal(clusterConfiguration)
	if err != nil {
		return "", err
	}

	return string(configV3), nil
}
