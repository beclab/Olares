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
	"fmt"
	"path/filepath"
	"strings"
	"text/template"

	"github.com/beclab/Olares/cli/pkg/storage"
	storagetpl "github.com/beclab/Olares/cli/pkg/storage/templates"

	"github.com/beclab/Olares/cli/pkg/container"
	"github.com/beclab/Olares/cli/pkg/manifest"
	"github.com/beclab/Olares/cli/pkg/registry"

	kubekeyapiv1alpha2 "github.com/beclab/Olares/cli/apis/kubekey/v1alpha2"
	kubekeyregistry "github.com/beclab/Olares/cli/pkg/bootstrap/registry"
	"github.com/beclab/Olares/cli/pkg/common"
	"github.com/beclab/Olares/cli/pkg/core/action"
	"github.com/beclab/Olares/cli/pkg/core/connector"
	"github.com/beclab/Olares/cli/pkg/core/logger"
	"github.com/beclab/Olares/cli/pkg/core/util"
	"github.com/beclab/Olares/cli/pkg/images"
	"github.com/beclab/Olares/cli/pkg/k3s/templates"
	"github.com/beclab/Olares/cli/pkg/utils"
	"github.com/pkg/errors"
	versionutil "k8s.io/apimachinery/pkg/util/version"
)

// hostMemoryProtectionDropIn is the file name every memory protection drop-in
// gets, inside its own <unit>.d directory.
const hostMemoryProtectionDropIn = "10-olares-memory.conf"

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

	// Resolved once: it reaches the host, and the reserve below and the
	// protection further down must not disagree about what this node runs.
	runsControlPlane := utils.NodeRunsControlPlane(runtime)

	defaultKubeletArs := map[string]string{
		"kube-reserved":   fmt.Sprintf("cpu=200m,memory=%s,ephemeral-storage=1Gi", utils.KubeReservedMemory(runsControlPlane)),
		"system-reserved": fmt.Sprintf("cpu=200m,memory=%s,ephemeral-storage=1Gi", utils.SystemReservedMemory(runtime, g.KubeConf.Arg.EnablePodSwap)),
		// Soft eviction gives kubelet a runway. Reaching a hard threshold means
		// reclaim has already failed, and terminating a pod from there takes long
		// enough that the OOM killer can decide first — and it does not choose
		// with the cluster in mind.
		//
		// The soft threshold is on allocatableMemory.available rather than
		// memory.available because that is the signal that answers "have the pods
		// used up what they were given", which is the condition worth reacting to
		// early. kubelet will not create it: addAllocatableThresholds in
		// pkg/kubelet/eviction/helpers.go copies only *hard* memory thresholds onto
		// that signal. Its capacity is kubepods.slice's own allowance, so it is
		// unaffected by the host releasing or taking page cache, and it moves with
		// the reserve above rather than with the size of the machine.
		//
		// It has to stay a percentage, and above the 5% below. eviction-hard's
		// memory.available<5% is copied onto the allocatable signal as 5% of that
		// same allowance, so an absolute value here could not stay above it across
		// machine sizes: whatever value sits above 5% on a small node sits below
		// it on a large one, and there the hard copy fires first and the grace
		// period below is never reached.
		//
		// Nothing is reserved for this. Only hard thresholds are subtracted from
		// Node Allocatable (hardEvictionReservation in pkg/kubelet/cm/helpers.go,
		// whose switch does not even have a case for the allocatable signal), so
		// the cost of the 10% is that pods cannot sit above 90% of their allowance
		// indefinitely, not that the 10% goes unscheduled.
		"eviction-hard":                       "memory.available<5%,nodefs.available<5%,imagefs.available<5%,pid.available<10%",
		"eviction-soft":                       "allocatableMemory.available<10%",
		"eviction-soft-grace-period":          "allocatableMemory.available=1m",
		"eviction-max-pod-grace-period":       "120",
		"eviction-pressure-transition-period": "30s",
		"config":                              "/etc/rancher/k3s/kubelet.config",
		"containerd":                          container.DefaultContainerdCRISocket,
		"cgroup-driver":                       "systemd",
		"runtime-request-timeout":             "5m",
		"image-gc-high-threshold":             "96",
		"image-gc-low-threshold":              "95",
		"housekeeping_interval":               "5s",
	}
	defaultKubeProxyArgs := map[string]string{
		"proxy-mode": "ipvs",
	}

	defaultKubeApiServerArgs := map[string]string{
		"service-node-port-range": "445-32767",
	}

	kubeApiserverArgs, _ := util.GetArgs(defaultKubeApiServerArgs, g.KubeConf.Cluster.Kubernetes.ApiServerArgs)
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

	// Written here rather than as its own task so that the reserve above and the
	// protection below cannot be changed independently, and so that the upgrade
	// path, which regenerates the unit through this task, refreshes both.
	//
	// Sized through NodeRunsControlPlane rather than the IsMaster above. That one
	// decides what this unit is told to be, so it has to take the declared role
	// at face value; this one only has to size what the node actually runs, and
	// so gets to confirm a declared master against the host.
	if err := g.generateHostMemoryProtection(runtime, runsControlPlane); err != nil {
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

// generateHostMemoryProtection installs the systemd drop-ins that keep k3s,
// containerd and the shims resident while the node is under memory pressure.
//
// All three are written together on purpose. The per-service values are inert
// without the one on system.slice, and a node where only the services are
// configured reclaims them just as readily as one with nothing configured at
// all, so there is no useful halfway state to leave behind on a partial failure.
func (g *GenerateK3sService) generateHostMemoryProtection(runtime connector.Runtime, runsControlPlane bool) error {
	dropIns := []struct {
		tmpl *template.Template
		unit string
		data util.Data
	}{
		{
			tmpl: templates.HostMemoryProtectionSlice,
			unit: "system.slice",
			data: util.Data{
				"SystemSliceMemoryMin": fmt.Sprintf("%dM", utils.HostDaemonsMemoryMi(runsControlPlane)),
			},
		},
		{
			tmpl: templates.HostMemoryProtectionK3s,
			unit: "k3s.service",
			data: util.Data{
				"K3sMemoryMin": fmt.Sprintf("%dM", utils.K3sMemoryMi(runsControlPlane)),
			},
		},
		{
			tmpl: templates.HostMemoryProtectionContainerd,
			unit: "containerd.service",
			data: util.Data{
				"ContainerdMemoryMin": fmt.Sprintf("%dM", utils.ContainerdMemoryMi()),
			},
		},
	}

	// Only a control plane has an etcd.service to protect, and writing a drop-in
	// for a unit that does not exist would leave a file that quietly starts
	// applying if one ever appears.
	if runsControlPlane {
		dropIns = append(dropIns, struct {
			tmpl *template.Template
			unit string
			data util.Data
		}{
			tmpl: templates.HostMemoryProtectionEtcd,
			unit: "etcd.service",
			data: util.Data{
				"EtcdMemoryMin": fmt.Sprintf("%dM", utils.EtcdMemoryMi(runsControlPlane)),
			},
		})
	}

	for _, dropIn := range dropIns {
		dir := filepath.Join("/etc/systemd/system", dropIn.unit+".d")
		if _, err := runtime.GetRunner().SudoCmd(fmt.Sprintf("mkdir -p %s", dir), false, false); err != nil {
			return errors.Wrapf(errors.WithStack(err), "create drop-in dir %s failed", dir)
		}

		templateAction := action.Template{
			Name:     "GenerateHostMemoryProtection",
			Template: dropIn.tmpl,
			Dst:      filepath.Join(dir, hostMemoryProtectionDropIn),
			Data:     dropIn.data,
		}
		templateAction.Init(nil, nil)
		if err := templateAction.Execute(runtime); err != nil {
			return errors.Wrapf(errors.WithStack(err), "write memory protection for %s failed", dropIn.unit)
		}
	}

	if _, err := runtime.GetRunner().SudoCmd("systemctl daemon-reload", false, false); err != nil {
		return errors.Wrap(errors.WithStack(err), "systemctl daemon-reload failed")
	}

	// daemon-reload is not enough on its own. It updates systemd's own view but
	// does not write a running service's cgroup attributes back to the kernel,
	// which happens when the unit next starts: after a reload that changes these
	// values, systemctl show reports the new MemoryMin while the unit's
	// memory.min still holds the old one. Waiting for a restart is not an answer
	// either, because system.slice is the ancestor the others depend on — cgroup
	// v2 caps a cgroup's effective protection at what its parents protect — and a
	// slice cannot be restarted.
	//
	// set-property leaves a drop-in under /run that outranks the one under /etc
	// for the rest of the boot. That is safe only because both are written here
	// from the same numbers on every run; writing just one of them is what makes
	// a stale value stick until reboot.
	applied := []string{
		fmt.Sprintf("system.slice MemoryMin=%dM", utils.HostDaemonsMemoryMi(runsControlPlane)),
		fmt.Sprintf("k3s.service MemoryMin=%dM MemorySwapMax=0", utils.K3sMemoryMi(runsControlPlane)),
		fmt.Sprintf("containerd.service MemoryMin=%dM MemorySwapMax=0", utils.ContainerdMemoryMi()),
	}
	if runsControlPlane {
		applied = append(applied,
			fmt.Sprintf("etcd.service MemoryMin=%dM MemorySwapMax=0", utils.EtcdMemoryMi(runsControlPlane)))
	}
	for _, props := range applied {
		cmd := fmt.Sprintf("systemctl set-property --runtime %s", props)
		if _, err := runtime.GetRunner().SudoCmd(cmd, false, false); err != nil {
			// The drop-in under /etc still governs from the next boot, so this is
			// worth reporting rather than failing an install over.
			logger.Warnf("could not apply memory protection to the running unit (%s), "+
				"it will take effect on reboot: %v", props, err)
		}
	}

	logger.Infof("host memory protection: system.slice %dMi, k3s %dMi, containerd %dMi, etcd %dMi",
		utils.HostDaemonsMemoryMi(runsControlPlane), utils.K3sMemoryMi(runsControlPlane),
		utils.ContainerdMemoryMi(), utils.EtcdMemoryMi(runsControlPlane))

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
	targetHome, targetUID, targetGID, err := utils.ResolveSudoUserHomeAndIDs(runtime)
	if err != nil {
		return err
	}

	cmds := []string{
		"mkdir -p /root/.kube",
		"cp -f /etc/rancher/k3s/k3s.yaml /root/.kube/config",
		"chmod 0600 /root/.kube/config",
		fmt.Sprintf("mkdir -p %s", filepath.Join(targetHome, ".kube")),
		fmt.Sprintf("cp -f /etc/rancher/k3s/k3s.yaml %s", filepath.Join(targetHome, ".kube", "config")),
		fmt.Sprintf("chmod 0600 %s", filepath.Join(targetHome, ".kube", "config")),
		fmt.Sprintf("chown -R %s:%s %s", targetUID, targetGID, filepath.Join(targetHome, ".kube")),
	}
	if _, err := runtime.GetRunner().SudoCmd(strings.Join(cmds, " && "), false, false); err != nil {
		return errors.Wrap(errors.WithStack(err), "copy k3s kube config failed")
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

		oldServer := "server: https://127.0.0.1:6443"
		newServer := fmt.Sprintf("server: https://%s:%d",
			s.KubeConf.Cluster.ControlPlaneEndpoint.Domain,
			s.KubeConf.Cluster.ControlPlaneEndpoint.Port)
		newKubeConfig := strings.Replace(cluster.KubeConfig, oldServer, newServer, -1)

		targetHome, targetUID, targetGID, err := utils.ResolveSudoUserHomeAndIDs(runtime)
		if err != nil {
			return err
		}
		targetKubeConfigPath := filepath.Join(targetHome, ".kube", "config")

		cmds := []string{
			"mkdir -p /root/.kube",
			fmt.Sprintf("echo '%s' > %s", newKubeConfig, "/root/.kube/config"),
			"chmod 0600 /root/.kube/config",
			fmt.Sprintf("mkdir -p %s", filepath.Join(targetHome, ".kube")),
			fmt.Sprintf("echo '%s' > %s", newKubeConfig, targetKubeConfigPath),
			fmt.Sprintf("chmod 0600 %s", targetKubeConfigPath),
			fmt.Sprintf("chown -R %s:%s %s", targetUID, targetGID, filepath.Join(targetHome, ".kube")),
		}
		if _, err := runtime.GetRunner().SudoCmd(strings.Join(cmds, " && "), false, false); err != nil {
			return errors.Wrap(errors.WithStack(err), "sync kube config failed")
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
