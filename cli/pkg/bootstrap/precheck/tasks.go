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

package precheck

import (
	"bufio"
	"bytes"
	"context"
	"fmt"
	"net"
	"os"
	"slices"
	"strings"
	"time"

	"github.com/Masterminds/semver/v3"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/util/sets"
	ctrl "sigs.k8s.io/controller-runtime"

	"github.com/beclab/Olares/cli/pkg/common"
	"github.com/beclab/Olares/cli/pkg/core/action"
	"github.com/beclab/Olares/cli/pkg/core/connector"
	"github.com/beclab/Olares/cli/pkg/core/logger"
	"github.com/beclab/Olares/cli/pkg/core/util"
	"github.com/beclab/Olares/cli/pkg/utils"
	"github.com/pkg/errors"
	kclient "k8s.io/client-go/kubernetes"
)

type RunChecks struct {
	common.KubeAction
	Checkers []Checker
}

type Checker interface {
	Name() string
	Check(runtime connector.Runtime) error
}

func (t *RunChecks) Execute(runtime connector.Runtime) error {
	var errBuffer bytes.Buffer
	for _, checker := range t.Checkers {
		if err := checker.Check(runtime); err != nil {
			errBuffer.WriteString(
				fmt.Sprintf("[%s] %v\n", checker.Name(), err),
			)
		}
	}
	if errBuffer.Len() > 0 {
		logger.Errorf("Some checks have failed:\n%s", errBuffer.String())
		os.Exit(1)
	}
	return nil
}

type SystemSupportCheck struct{}

func (t *SystemSupportCheck) Name() string {
	return "System"
}

func (t *SystemSupportCheck) Check(runtime connector.Runtime) error {
	err := runtime.GetSystemInfo().IsSupport()
	if err == nil {
		return nil
	}
	// Interactive warning instead of outright failure
	fmt.Printf("%v Use at your own risk, would you like to continue? (Y/N): ", err)
	reader, rerr := utils.GetBufIOReaderOfTerminalInput()
	if rerr != nil {
		return fmt.Errorf("could not read terminal input: %v", rerr)
	}
	input, _ := reader.ReadString('\n')
	input = strings.TrimSpace(input)
	if !strings.HasPrefix("yes", strings.ToLower(input)) {
		return err
	}
	return nil
}

type RequiredPortsCheck struct{}

func (t *RequiredPortsCheck) Name() string {
	return "Ports"
}

func (t *RequiredPortsCheck) Check(runtime connector.Runtime) error {
	if !runtime.GetSystemInfo().IsLinux() {
		return nil
	}
	ports := []int{80, 443, 444, 2444, 9100, 30180}
	var unbindablePorts []int
	for _, port := range ports {
		l, err := net.Listen("tcp", fmt.Sprintf(":%d", port))
		if err != nil {
			unbindablePorts = append(unbindablePorts, port)
			continue
		}
		defer l.Close()
	}
	if len(unbindablePorts) > 0 {
		return fmt.Errorf("port %v required by Olares cannot be bound, you can check which process using the command `sudo netstat -tlnp`", unbindablePorts)
	}
	return nil
}

type ConflictingContainerdCheck struct{}

func (t *ConflictingContainerdCheck) Name() string {
	return "Containerd"
}

func (t *ConflictingContainerdCheck) Check(runtime connector.Runtime) error {
	if !runtime.GetSystemInfo().IsLinux() {
		return nil
	}
	kubeRuntime := runtime.(*common.KubeRuntime)
	if kubeRuntime.Arg.IsCloudInstance {
		return nil
	}
	fixMSG := "\nIf it is installed as a component of Docker, it should be uninstalled per the official doc https://docs.docker.com/engine/install/ubuntu/#uninstall-old-versions"
	fixMSG += "\nIf it is left over from a previous installation of Olares, clean it up using the command `sudo olares-cli uninstall --all`"
	containerdBin, err := util.GetCommand("containerd")
	if err == nil && containerdBin != "" {
		return fmt.Errorf("found existing containerd binary: %s, a containerd managed by Olares is required to ensure normal function%s", containerdBin, fixMSG)
	}
	containerdSocket := "/run/containerd/containerd.sock"
	if util.IsExist(containerdSocket) {
		return fmt.Errorf("found existing containerd socket: %s, a containerd managed by Olares is required to ensure normal function%s", containerdSocket, fixMSG)
	}
	return nil
}

type SystemdCheck struct{}

func (t *SystemdCheck) Name() string {
	return "Systemd"
}

func (t *SystemdCheck) Check(runtime connector.Runtime) error {
	if !runtime.GetSystemInfo().IsLinux() {
		return nil
	}
	if util.IsExist("/run/systemd/system") {
		return nil
	}
	return errors.New("this system is not inited by systemd, which is required by Olares")
}

type MasterNodeReadyCheck struct{}

func (t *MasterNodeReadyCheck) Name() string {
	return "MasterNodeReady"
}

func (t *MasterNodeReadyCheck) Check(runtime connector.Runtime) error {
	config, err := ctrl.GetConfig()
	if err != nil {
		return fmt.Errorf("failed to get kubernetes config: %s", err)
	}
	client, err := kclient.NewForConfig(config)
	if err != nil {
		return fmt.Errorf("failed to create kubernetes client: %s", err)
	}
	ctx, cancel := context.WithTimeout(context.Background(), 1*time.Minute)
	defer cancel()
	nodes, err := client.CoreV1().Nodes().List(ctx, metav1.ListOptions{})
	if err != nil {
		return fmt.Errorf("failed to list nodes: %s", err)
	}
	for _, node := range nodes.Items {
		roles := sets.NewString()
		for k, v := range node.Labels {
			switch {
			case strings.HasPrefix(k, "node-role.kubernetes.io/"):
				if role := strings.TrimPrefix(k, "node-role.kubernetes.io/"); len(role) > 0 {
					roles.Insert(role)
				}

			case k == "kubernetes.io/role" && v != "":
				roles.Insert(v)
			}
		}
		if !roles.HasAny("control-plane", "master") {
			continue
		}
		if node.Spec.Unschedulable {
			return fmt.Errorf("node %s is unschedulable", node.Name)
		}
		var readyConditionExists bool
		for _, condition := range node.Status.Conditions {
			if condition.Type == corev1.NodeReady {
				readyConditionExists = true
				if condition.Status != corev1.ConditionTrue {
					return fmt.Errorf("node %s is not ready", node.Name)
				}
			}
		}
		if !readyConditionExists {
			return fmt.Errorf("node %s's condition is unknown", node.Name)
		}
	}

	return nil
}

type RootPartitionAvailableSpaceCheck struct{}

func (t *RootPartitionAvailableSpaceCheck) Name() string {
	return "RootPartitionAvailableSpace"
}

func (t *RootPartitionAvailableSpaceCheck) Check(runtime connector.Runtime) error {
	return nil
}

type ValidResolvConfCheck struct{}

func (t *ValidResolvConfCheck) Name() string {
	return "ResolvConf"
}

func (t *ValidResolvConfCheck) Check(runtime connector.Runtime) error {
	if !runtime.GetSystemInfo().IsLinux() {
		return nil
	}
	resolvConfFiles := []string{"/etc/resolv.conf", "/run/systemd/resolve/resolv.conf"}
	searchDomainPrefix := "search"
	for _, f := range resolvConfFiles {
		file, err := os.Open(f)
		if err != nil {
			if !os.IsNotExist(err) {
				return fmt.Errorf("failed to open resolv.conf file %s for validity check", f)
			}
			continue
		}
		defer file.Close()
		scanner := bufio.NewScanner(file)
		for scanner.Scan() {
			line := strings.TrimSpace(scanner.Text())
			if !strings.HasPrefix(line, searchDomainPrefix) {
				continue
			}
			logger.Debugf("found search domain list line in file %s: %s", f, line)
			searchDomains := strings.Fields(strings.TrimPrefix(line, searchDomainPrefix))
			if len(searchDomains) == 0 {
				return fmt.Errorf("invalid resolv.conf file %s: syntax error: empty search domain list", f)
			}
			for _, searchDomain := range searchDomains {
				if searchDomain != "" && searchDomain != "." {
					return fmt.Errorf("invalid resolv.conf file %s: search domain other than \".\" causes the malfunction of cluster DNS, please empty it before installation", f)
				}
			}
		}
	}
	return nil
}

type NvidiaCardArchChecker struct{}

func (t *NvidiaCardArchChecker) Name() string {
	return "NvidiaCardArch"
}

func (t *NvidiaCardArchChecker) Check(runtime connector.Runtime) error {
	supportedArchs := []string{"Blackwell", "Hopper", "Ada Lovelace", "Ampere", "Turing"}
	model, arch, err := utils.DetectNvidiaModelAndArch(runtime)
	if err != nil {
		return err
	}
	if strings.TrimSpace(model) == "" {
		return nil
	}
	if !slices.Contains(supportedArchs, arch) {
		return fmt.Errorf("unsupported NVIDIA card %s of architecture: %s, Olares only supports the following architectures: %s", model, arch, strings.Join(supportedArchs, ", "))
	}
	return nil
}

// NouveauChecker checks whether nouveau is loaded and has modeset=1 or -1.
// This check only runs when an NVIDIA GPU is present.
type NouveauChecker struct{}

func (n *NouveauChecker) Name() string {
	return "NouveauKernelModule"
}

func (n *NouveauChecker) Check(runtime connector.Runtime) error {
	if !runtime.GetSystemInfo().IsLinux() {
		return nil
	}
	model, _, err := utils.DetectNvidiaModelAndArch(runtime)
	if err != nil {
		fmt.Println("Error detecting NVIDIA card:", err)
		os.Exit(1)
	}
	if strings.TrimSpace(model) == "" {
		return nil
	}

	if !util.IsExist("/sys/module/nouveau") {
		return nil
	}

	const modesetPath = "/sys/module/nouveau/parameters/modeset"
	data, err := os.ReadFile(modesetPath)
	if err != nil {
		fmt.Printf("Error reading modeset parameter of nouveau kernel module by reading file %s: %v", modesetPath, err)
		os.Exit(1)
	}
	val := strings.TrimSpace(string(data))
	if val == "1" || val == "-1" {
		return fmt.Errorf("detected nouveau kernel module loaded with modeset=%s; this conflicts with the NVIDIA driver that Olares will install, please disable it by running `sudo olares-cli gpu disable-nouveau`, REBOOT your machine, and try again", val)
	}
	return nil
}

type CudaChecker struct{}

func (c *CudaChecker) Name() string {
	return "CUDA"
}

func (c *CudaChecker) Check(runtime connector.Runtime) error {
	if !runtime.GetSystemInfo().IsLinux() ||
		// Skip check on NVIDIA DGX Spark systems, which have their own GPU management
		runtime.GetSystemInfo().IsGB10Chip() {
		return nil
	}

	st, err := utils.GetNvidiaStatus(runtime)
	if err != nil {
		return err
	}
	if st == nil || !st.Installed {
		if st != nil && st.Running {
			return ErrKernelDriverUninstalledButRunning
		}
		logger.Info("NVIDIA driver is not installed")
		return nil
	}
	if st.Mismatch {
		return ErrDriverLibraryVersionMismatch
	}
	if st.InstallMethod != utils.GPUDriverInstallMethodRunfile && !runtime.GetSystemInfo().IsWsl() {
		return ErrNotInstalledByRunfile
	}
	logger.Infof("NVIDIA driver is installed, version: %s, cuda version: %s", st.DriverVersion, st.CudaVersion)
	oldestVer := semver.MustParse(supportedCudaVersions[0])
	newestVer := semver.MustParse(supportedCudaVersions[len(supportedCudaVersions)-1])
	currentVer := semver.MustParse(st.CudaVersion)
	if oldestVer.GreaterThan(currentVer) {
		return ErrUnsupportedCudaVersion
	}
	if newestVer.LessThan(currentVer) {
		logger.Info("CUDA version is too new, there might be compatibility issues with some applications, use at your own risk")
	}
	return nil
}

// RocmChecker checks AMD ROCm version for AMD GPU on Ubuntu 22.04/24.04 only.
type RocmChecker struct{}

func (r *RocmChecker) Name() string {
	return "ROCm"
}

func (r *RocmChecker) Check(runtime connector.Runtime) error {
	if !runtime.GetSystemInfo().IsLinux() {
		return nil
	}
	si := runtime.GetSystemInfo()
	if !si.IsUbuntu() || !(si.IsUbuntuVersionEqual(connector.Ubuntu2204) || si.IsUbuntuVersionEqual(connector.Ubuntu2404)) {
		return nil
	}

	// detect AMD GPU presence
	amdGPUExists, err := connector.HasAmdIGPU(runtime)
	if err != nil {
		return err
	}
	// no AMD GPU found, no need to check rocm
	if !amdGPUExists {
		return nil
	}

	curV, err := connector.RocmVersion()
	if err != nil && !os.IsNotExist(err) {
		return err
	}

	if os.IsNotExist(err) {
		return nil
	}

	min := semver.MustParse("7.1.1")
	if curV.LessThan(min) {
		return fmt.Errorf("detected ROCm version %s, which is lower than required %s; please uninstall existing ROCm/AMDGPU components before installation with command: olares-cli amdgpu uninstall", curV.Original(), min.Original())
	}
	return nil
}

//////////////////////////////////////////////
// precheck - task

type GreetingsTask struct {
	action.BaseAction
}

func (h *GreetingsTask) Execute(runtime connector.Runtime) error {
	_, err := runtime.GetRunner().Cmd("echo 'Greetings, Olares'", false, true)
	if err != nil {
		return err
	}

	return nil
}

type NodePreCheck struct {
	common.KubeAction
}

func (n *NodePreCheck) Execute(runtime connector.Runtime) error {
	var results = make(map[string]string)
	results["name"] = runtime.RemoteHost().GetName()
	for _, software := range baseSoftware {
		var (
			cmd string
		)

		switch software {
		case docker:
			cmd = "docker version --format '{{.Server.Version}}'"
		case containerd:
			cmd = "containerd --version | cut -d ' ' -f 3"
		default:
			cmd = fmt.Sprintf("which %s", software)
		}

		switch software {
		case sudo:
			// sudo skip sudo prefix
		default:
			cmd = runtime.RemoteHost().SudoPrefixIfNecessary(cmd)
		}

		res, err := runtime.GetRunner().Cmd(cmd, false, false)
		switch software {
		case showmount:
			software = nfs
		case rbd:
			software = ceph
		case glusterfs:
			software = glusterfs
		}
		if err != nil || strings.Contains(res, "not found") {
			results[software] = ""
		} else {
			// software in path
			if strings.Contains(res, "bin/") {
				results[software] = "y"
			} else {
				// get software version, e.g. docker, containerd, etc.
				results[software] = res
			}
		}
	}

	output, err := runtime.GetRunner().Cmd("date +\"%Z %H:%M:%S\"", false, false)
	if err != nil {
		results["time"] = ""
	} else {
		results["time"] = strings.TrimSpace(output)
	}

	host := runtime.RemoteHost()
	if res, ok := host.GetCache().Get(common.NodePreCheck); ok {
		m := res.(map[string]string)
		m = results
		host.GetCache().Set(common.NodePreCheck, m)
	} else {
		host.GetCache().Set(common.NodePreCheck, results)
	}
	return nil
}

type GetKubernetesNodesStatus struct {
	common.KubeAction
}

func (g *GetKubernetesNodesStatus) Execute(runtime connector.Runtime) error {
	nodeStatus, err := runtime.GetRunner().SudoCmd("/usr/local/bin/kubectl get node -o wide", false, false)
	if err != nil {
		return err
	}
	g.PipelineCache.Set(common.ClusterNodeStatus, nodeStatus)

	cri, err := runtime.GetRunner().SudoCmd("/usr/local/bin/kubectl get node -o jsonpath=\"{.items[*].status.nodeInfo.containerRuntimeVersion}\"", false, false)
	if err != nil {
		return err
	}
	g.PipelineCache.Set(common.ClusterNodeCRIRuntimes, cri)
	return nil
}

type GetStorageKeyTask struct {
	common.KubeAction
}

func (t *GetStorageKeyTask) Execute(runtime connector.Runtime) error {
	kubectl, err := util.GetCommand(common.CommandKubectl)
	if err != nil {
		return fmt.Errorf("kubectl not found")
	}
	var storageAccessKey, storageSecretKey, storageToken, storageClusterId string
	var ctx, cancel = context.WithTimeout(context.Background(), 5*time.Minute)
	defer cancel()

	if stdout, err := runtime.GetRunner().CmdContext(ctx, fmt.Sprintf("%s get terminus terminus -o jsonpath='{.metadata.annotations.bytetrade\\.io/s3-ak}'", kubectl), false, false); err != nil {
		storageAccessKey = t.KubeConf.Arg.Storage.StorageAccessKey
		if storageAccessKey == "" {
			logger.Errorf("storage access key not found")
		}
	} else {
		storageAccessKey = stdout
	}

	if stdout, err := runtime.GetRunner().CmdContext(ctx, fmt.Sprintf("%s get terminus terminus -o jsonpath='{.metadata.annotations.bytetrade\\.io/s3-sk}'", kubectl), false, false); err != nil {
		storageSecretKey = t.KubeConf.Arg.Storage.StorageSecretKey
		if storageSecretKey == "" {
			logger.Errorf("storage secret key not found")
		}
	} else {
		storageSecretKey = stdout
	}

	if stdout, err := runtime.GetRunner().CmdContext(ctx, fmt.Sprintf("%s get terminus terminus -o jsonpath='{.metadata.annotations.bytetrade\\.io/s3-sts}'", kubectl), false, false); err != nil {
		storageToken = t.KubeConf.Arg.Storage.StorageToken
		if storageToken == "" {
			logger.Errorf("storage token not found")
		}
	} else {
		storageToken = stdout
	}

	if stdout, err := runtime.GetRunner().CmdContext(ctx, fmt.Sprintf("%s get terminus terminus -o jsonpath='{.metadata.labels.bytetrade\\.io/cluster-id}'", kubectl), false, false); err != nil {
		storageClusterId = t.KubeConf.Arg.Storage.StorageClusterId
		if storageClusterId == "" {
			logger.Errorf("storage cluster id not found")
		}
	} else {
		storageClusterId = stdout
	}

	t.PipelineCache.Set(common.CacheAccessKey, storageAccessKey)
	t.PipelineCache.Set(common.CacheSecretKey, storageSecretKey)
	t.PipelineCache.Set(common.CacheToken, storageToken)
	t.PipelineCache.Set(common.CacheClusterId, storageClusterId)

	logger.Infof("storage: cloud: %v, type: %s, bucket: %s, ak: %s, sk: %s, tk: %s, id: %s",
		t.KubeConf.Arg.IsCloudInstance, t.KubeConf.Arg.Storage.StorageType, t.KubeConf.Arg.Storage.StorageBucket,
		storageAccessKey, storageSecretKey, storageToken, storageClusterId)

	return nil
}

type AddWSLChattr struct {
	common.KubeAction
}

func (a *AddWSLChattr) Execute(runtime connector.Runtime) error {
	if !runtime.GetSystemInfo().IsWsl() {
		return nil
	}
	runtime.GetRunner().SudoCmd("chattr +i /etc/hosts /etc/resolv.conf", false, false)
	return nil
}

type RemoveWSLChattr struct {
	common.KubeAction
}

func (t *RemoveWSLChattr) Execute(runtime connector.Runtime) error {
	if !runtime.GetSystemInfo().IsWsl() {
		return nil
	}
	runtime.GetRunner().SudoCmd("chattr -i /etc/hosts", false, true)
	runtime.GetRunner().SudoCmd("chattr -i /etc/resolv.conf", false, true)
	return nil
}

var ErrUnsupportedCudaVersion = errors.New("unsupported cuda version, please uninstall it using the command `sudo olares-cli gpu uninstall`, REBOOT your machine, and try again")
var ErrKernelDriverUninstalledButRunning = errors.New("NVIDIA driver is uninstalled, but the kernel driver is still running, please REBOOT your machine, and try again")
var ErrNotInstalledByRunfile = errors.New("NVIDIA driver is installed, but not installed by runfile, please uninstall it using the command `sudo olares-cli gpu uninstall`, REBOOT your machine, and try again")
var ErrDriverLibraryVersionMismatch = errors.New("NVIDIA driver is installed, but the library version with the running version is mismatched, please REBOOT your machine, and try again")
var supportedCudaVersions = []string{common.CurrentVerifiedCudaVersion}
