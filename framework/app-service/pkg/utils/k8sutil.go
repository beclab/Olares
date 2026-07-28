package utils

import (
	"context"
	"errors"
	"fmt"
	"net"
	"net/http"
	"net/url"
	"os"
	"sort"
	"strconv"
	"strings"
	"time"

	"k8s.io/apimachinery/pkg/fields"

	"github.com/go-resty/resty/v2"
	"github.com/pelletier/go-toml"

	"github.com/beclab/Olares/framework/app-service/pkg/apiserver/api"
	"github.com/beclab/Olares/framework/app-service/pkg/client/clientset"
	"github.com/beclab/Olares/framework/app-service/pkg/constants"
	"github.com/beclab/Olares/framework/app-service/pkg/prometheus"
	"github.com/beclab/Olares/framework/app-service/pkg/users"
	sysv1alpha1 "github.com/beclab/api/api/sys.bytetrade.io/v1alpha1"
	iamv1alpha2 "github.com/beclab/api/iam/v1alpha2"

	authv1 "k8s.io/api/authentication/v1"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"k8s.io/apimachinery/pkg/types"
	"k8s.io/client-go/dynamic"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/rest"
	"k8s.io/klog/v2"
	"k8s.io/utils/ptr"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

const (
	// CalicoTunnelAddrAnnotation annotation key for calico tunnel address.
	CalicoTunnelAddrAnnotation = "projectcalico.org/IPv4IPIPTunnelAddr"
	DefaultRegistry            = "https://registry-1.docker.io"

	minikubeLabelPrefix = "minikube.k8s.io/"
	// docker or  podman driver's providerID prefix
	minikubeKICProvider = "kic://"
)

// GetAllNodesPodCIDRs returns all node pod's cidr.
func GetAllNodesPodCIDRs() (cidrs []string) {
	cidrs = make([]string, 0)

	config, err := ctrl.GetConfig()
	if err != nil {
		klog.Errorf("Failed to get kube config err=%v", err)
		return
	}
	c, err := clientset.New(config)
	if err != nil {
		klog.Errorf("Failed to create new client err=%v", err)
		return
	}

	nodes, err := c.KubeClient.Kubernetes().CoreV1().Nodes().
		List(context.TODO(), metav1.ListOptions{})
	if err != nil {
		klog.Errorf("Failed to list nodes err=%v", err)
		return
	}

	for _, node := range nodes.Items {
		if node.Spec.PodCIDR != "" {
			cidrs = append(cidrs, node.Spec.PodCIDR)
		}
	}
	return cidrs
}

// GetAllNodesTunnelIPCIDRs returns all node tunnel's ip cidr.
func GetAllNodesTunnelIPCIDRs() (cidrs []string) {
	cidrs = make([]string, 0)

	config, err := ctrl.GetConfig()
	if err != nil {
		klog.Errorf("Failed to get kube config: %v", err)
		return
	}
	c, err := clientset.New(config)
	if err != nil {
		klog.Errorf("Failed to create new client: %v", err)
		return
	}

	nodes, err := c.KubeClient.Kubernetes().CoreV1().Nodes().
		List(context.TODO(), metav1.ListOptions{})
	if err != nil {
		klog.Errorf("Failed to list nodes err=%v", err)
		return
	}

	for _, node := range nodes.Items {
		if ip, ok := node.Annotations[CalicoTunnelAddrAnnotation]; ok {
			cidrs = append(cidrs, ip+"/32")
		}
	}

	return cidrs
}

// func FindGpuTypeFromNodes(nodes *corev1.NodeList) (string, error) {
// 	gpuType := "none"
// 	if nodes == nil {
// 		return gpuType, errors.New("empty node list")
// 	}
// 	for _, n := range nodes.Items {
// 		if _, ok := n.Status.Capacity[constants.NvidiaGPU]; ok {
// 			if _, ok = n.Status.Capacity[constants.NvshareGPU]; ok {
// 				return "nvshare", nil

// 			}
// 			gpuType = "nvidia"
// 		}
// 		if _, ok := n.Status.Capacity[constants.VirtAiTechVGPU]; ok {
// 			return "virtaitech", nil
// 		}
// 	}
// 	return gpuType, nil
// }

// GetAllGpuTypesFromNodes returns the set of non-cpu GPU modes declared across
// the cluster's nodes. Each node may advertise several modes at once via the
// existence-based per-mode labels gpu.bytetrade.io/<mode> (and the legacy
// single-value gpu.bytetrade.io/type label is still honored for backward
// compatibility) — see NodeSupportedGPUTypes. Nodes without any GPU label
// contribute nothing, so a pure-CPU cluster returns an empty map. Callers that
// need to surface "CPU" as a user-selectable option should add it on top of
// this set themselves; mixing it in here would break the chart-render
// auto-detect path which expects len(gpuTypes)==1 to mean "this cluster has
// exactly one GPU flavour".
func GetAllGpuTypesFromNodes(nodes *corev1.NodeList) (map[string]struct{}, error) {
	gpuTypes := make(map[string]struct{})
	if nodes == nil {
		return gpuTypes, errors.New("empty node list")
	}
	for i := range nodes.Items {
		for _, mode := range NodeSupportedGPUTypes(&nodes.Items[i]) {
			gpuTypes[mode] = struct{}{} // TODO: add driver version info
		}
	}
	return gpuTypes, nil
}

func IsNodeReady(node *corev1.Node) bool {
	for _, c := range node.Status.Conditions {
		if c.Type == corev1.NodeReady {
			return c.Status == corev1.ConditionTrue
		}
	}
	return false
}

// dockerHostsTOMLPath is containerd's registry hosts config for docker.io under
// config_path (/etc/containerd/certs.d). Registry mirrors moved here in the
// containerd v3 config; the inline registry.mirrors in config.toml is deprecated
// and ignored by containerd 2.x once config_path is set.
const dockerHostsTOMLPath = "/etc/containerd/certs.d/docker.io/hosts.toml"

// GetMirrorsEndpoint returns the docker.io pull-through mirror endpoints
// configured for containerd, in priority (file) order. They are read from
// containerd's certs.d hosts.toml `[host."<url>"]` entries, the containerd v3
// replacement for the deprecated inline registry.mirrors config.
func GetMirrorsEndpoint() (ep []string) {
	data, err := os.ReadFile(dockerHostsTOMLPath)
	if err != nil {
		klog.Infof("load mirrors endpoint from %s failed err=%v", dockerHostsTOMLPath, err)
		return
	}
	tree, err := toml.LoadBytes(data)
	if err != nil {
		klog.Infof("parse %s failed err=%v", dockerHostsTOMLPath, err)
		return
	}
	hostTree, ok := tree.Get("host").(*toml.Tree)
	if !ok {
		return
	}
	// TOML tables are unordered; recover mirror precedence from source line number,
	// the same way containerd's own resolver orders hosts.
	hosts := hostTree.Keys()
	// Note: use GetPath (single path element), not Get — Get treats the key as a
	// dot-separated path, which would mangle host URLs that contain dots/colons.
	sort.SliceStable(hosts, func(i, j int) bool {
		ti, _ := hostTree.GetPath([]string{hosts[i]}).(*toml.Tree)
		tj, _ := hostTree.GetPath([]string{hosts[j]}).(*toml.Tree)
		if ti == nil || tj == nil {
			return false
		}
		return ti.Position().Line < tj.Position().Line
	})
	return hosts
}

// ReplacedImageRef return replaced image ref and true if mirror is support http
func ReplacedImageRef(mirrorsEndpoint []string, oldImageRef string, checkConnection bool) (string, bool) {
	if len(mirrorsEndpoint) == 0 {
		return oldImageRef, false
	}
	plainHTTP := false
	for _, ep := range mirrorsEndpoint {
		if ep != "" && ep != DefaultRegistry {
			url, err := url.Parse(ep)
			if err != nil {
				continue
			}
			if url.Scheme == "http" {
				plainHTTP = true
			}
			if checkConnection {
				host := url.Host
				if !hasPort(url.Host) {
					if url.Scheme == "https" {
						host = net.JoinHostPort(url.Host, "443")
					} else {
						host = net.JoinHostPort(url.Host, "80")
					}
				}
				conn, err := net.DialTimeout("tcp", host, 2*time.Second)
				if err != nil {
					continue
				}
				if conn != nil {
					conn.Close()
				}
			}

			parts := strings.Split(oldImageRef, "/")
			klog.Infof("parts: %s", parts)
			if parts[0] == "docker.io" {
				parts[0] = url.Host
			}
			klog.Infof("parts2: %s", parts)
			return strings.Join(parts, "/"), plainHTTP
		}
	}
	return oldImageRef, false
}

func hasPort(s string) bool { return strings.LastIndex(s, ":") > strings.LastIndex(s, "]") }

func FindOwnerUser(ctrlClient client.Client, user *iamv1alpha2.User) (*iamv1alpha2.User, error) {
	creator := user.Annotations[users.AnnotationUserCreator]
	if creator != "cli" {
		var creatorUser iamv1alpha2.User
		err := ctrlClient.Get(context.TODO(), types.NamespacedName{Name: creator}, &creatorUser)
		if err != nil {
			return nil, err
		}
		return &creatorUser, nil
	}

	var userList iamv1alpha2.UserList
	err := ctrlClient.List(context.TODO(), &userList)
	if err != nil {
		klog.Errorf("failed to list user %v", err)
		return nil, err
	}
	for _, u := range userList.Items {
		if u.Annotations[users.UserAnnotationOwnerRole] == "owner" {
			return &u, nil
		}
	}
	return nil, errors.New("user with owner role not found")
}

func GetDeploymentName(pod *corev1.Pod) string {
	if pod == nil {
		return ""
	}

	replicaSetHash := pod.Labels["pod-template-hash"]
	if replicaSetHash == "" {
		return ""
	}

	replicaSetSuffix := fmt.Sprintf("-%s", replicaSetHash)
	return strings.Split(pod.GenerateName, replicaSetSuffix)[0] // pod.name not exists
}

var serviceAccountToken string

func GetServerServiceAccountToken() (string, error) {
	if serviceAccountToken != "" {
		return serviceAccountToken, nil
	}

	config, err := ctrl.GetConfig()
	if err != nil {
		klog.Errorf("Failed to get config: %v", err)
		return "", err
	}

	serviceAccountToken = config.BearerToken

	return serviceAccountToken, nil
}

func GetUserServiceAccountToken(ctx context.Context, client kubernetes.Interface, user string) (string, error) {
	namespace := fmt.Sprintf("user-system-%s", user)
	tr := &authv1.TokenRequest{
		Spec: authv1.TokenRequestSpec{
			Audiences:         []string{"https://kubernetes.default.svc.cluster.local"},
			ExpirationSeconds: ptr.To[int64](86400), // one day
		},
	}

	token, err := client.CoreV1().ServiceAccounts(namespace).
		CreateToken(ctx, "user-backend", tr, metav1.CreateOptions{})
	if err != nil {
		klog.Errorf("Failed to create token for user %s in namespace %s: %v", user, namespace, err)
		return "", err
	}

	return token.Status.Token, nil
}

func GetTerminusVersion(ctx context.Context, config *rest.Config) (*sysv1alpha1.Terminus, error) {
	dynamicClient, err := dynamic.NewForConfig(config)
	if err != nil {
		return nil, fmt.Errorf("failed to create dynamic client: %v", err)
	}

	terminusGVR := schema.GroupVersionResource{
		Group:    "sys.bytetrade.io",
		Version:  "v1alpha1",
		Resource: "terminus",
	}

	unstructuredTerminus, err := dynamicClient.Resource(terminusGVR).Get(ctx, "terminus", metav1.GetOptions{})
	if err != nil {
		return nil, fmt.Errorf("failed to get terminus resource: %v", err)
	}

	var terminus sysv1alpha1.Terminus
	err = runtime.DefaultUnstructuredConverter.FromUnstructured(unstructuredTerminus.Object, &terminus)
	if err != nil {
		return nil, fmt.Errorf("failed to convert unstructured to Terminus: %v", err)
	}

	return &terminus, nil
}

func GetTerminus(ctx context.Context, ctrlClient client.Client) (*sysv1alpha1.Terminus, error) {
	var terminus sysv1alpha1.Terminus
	if err := ctrlClient.Get(ctx, types.NamespacedName{Name: "terminus"}, &terminus); err != nil {
		return nil, err
	}
	return &terminus, nil
}

var ArchNameMap = map[int64]string{
	2:  "Kepler",
	3:  "Maxwell",
	4:  "Pascal",
	5:  "Volta",
	6:  "Turing",
	7:  "Ampere",
	8:  "Ada Lovelace",
	9:  "Hopper",
	10: "Blackwell",
}

func DecodeNodeGPU(str string) ([]api.GPUInfo, error) {
	if !strings.Contains(str, constants.OneContainerMultiDeviceSplitSymbol) {
		return []api.GPUInfo{}, errors.New("node annotations not decode successfully")
	}
	gpusStr := strings.Split(str, constants.OneContainerMultiDeviceSplitSymbol)
	var ret []api.GPUInfo
	for _, val := range gpusStr {
		if strings.Contains(val, ",") {
			items := strings.Split(val, ",")
			if len(items) == 7 || len(items) == 9 || len(items) == 10 {
				architecture := int64(0)
				modelName := items[4]
				memory, _ := strconv.ParseInt(items[2], 10, 32)
				if len(items) == 10 {
					architecture, _ = strconv.ParseInt(items[9], 10, 32)
				}
				archStr := ArchNameMap[architecture]
				info := api.GPUInfo{
					Vendor: "NVIDIA",
					Architecture: func() string {
						if archStr != "" {
							return archStr
						}
						return "unknown"
					}(),
					Memory:    memory,
					ModelName: modelName,
					Model:     ExtractGPUVersion(modelName),
				}
				ret = append(ret, info)

			}
		}
	}
	return ret, nil
}

func GetNodeInfo(ctx context.Context) (ret []api.NodeInfo, err error) {
	config, err := ctrl.GetConfig()
	if err != nil {
		return nil, err
	}
	client, err := kubernetes.NewForConfig(config)
	if err != nil {
		return nil, err
	}

	cpuMap, err := prometheus.GetNodeCpuResource(ctx)
	if err != nil {
		return nil, err
	}
	klog.Errorf("CpuMaP: %#v", cpuMap)
	nodes, err := client.CoreV1().Nodes().List(ctx, metav1.ListOptions{})
	if err != nil {
		return nil, err
	}
	for _, n := range nodes.Items {
		cpuInfo, ok := cpuMap[n.Name]
		if !ok {
			cpuInfo = api.CPUInfo{}
		}
		cpuInfo.Arch = n.Labels[constants.ArchLabelKey]
		coreNumber, _ := n.Status.Capacity.Cpu().AsInt64()
		cpuInfo.CoreNumber = int(coreNumber)
		cudaVersion := n.Labels[constants.CudaVersionLabelKey]
		gpus, _ := DecodeNodeGPU(n.Annotations[constants.NodeNvidiaRegistryKey])
		for i := range gpus {
			gpus[i].Memory *= 1024 * 1024
		}

		ret = append(ret, api.NodeInfo{
			CudaVersion: cudaVersion,
			CPU:         []api.CPUInfo{cpuInfo},
			Memory: api.MemInfo{
				Total: func() int64 {
					total, _ := n.Status.Capacity.Memory().AsInt64()
					return total
				}(),
			},
			GPUS: gpus,
		})
	}
	return
}

type SystemStatusResponse struct {
	Code    int    `json:"code"`
	Message string `json:"message"`
	Data    struct {
		TerminusdState            string `json:"terminusdState"`
		TerminusState             string `json:"terminusState"`
		TerminusName              string `json:"terminusName"`
		TerminusVersion           string `json:"terminusVersion"`
		InstalledTime             int64  `json:"installedTime"`
		InitializedTime           int64  `json:"initializedTime"`
		OlaresdVersion            string `json:"olaresdVersion"`
		DeviceName                string `json:"device_name"`
		HostName                  string `json:"host_name"`
		OsType                    string `json:"os_type"`
		OsArch                    string `json:"os_arch"`
		OsInfo                    string `json:"os_info"`
		OsVersion                 string `json:"os_version"`
		CpuInfo                   string `json:"cpu_info"`
		GpuInfo                   string `json:"gpu_info"`
		Memory                    string `json:"memory"`
		Disk                      string `json:"disk"`
		WifiConnected             bool   `json:"wifiConnected"`
		WiredConnected            bool   `json:"wiredConnected"`
		HostIp                    string `json:"hostIp"`
		ExternalIp                string `json:"externalIp"`
		InstallingState           string `json:"installingState"`
		InstallingProgress        string `json:"installingProgress"`
		UninstallingState         string `json:"uninstallingState"`
		UninstallingProgress      string `json:"uninstallingProgress"`
		UpgradingTarget           string `json:"upgradingTarget"`
		UpgradingRetryNum         int    `json:"upgradingRetryNum"`
		UpgradingState            string `json:"upgradingState"`
		UpgradingStep             string `json:"upgradingStep"`
		UpgradingProgress         string `json:"upgradingProgress"`
		UpgradingError            string `json:"upgradingError"`
		UpgradingDownloadState    string `json:"upgradingDownloadState"`
		UpgradingDownloadStep     string `json:"upgradingDownloadStep"`
		UpgradingDownloadProgress string `json:"upgradingDownloadProgress"`
		UpgradingDownloadError    string `json:"upgradingDownloadError"`
		CollectingLogsState       string `json:"collectingLogsState"`
		CollectingLogsError       string `json:"collectingLogsError"`
		DefaultFrpServer          string `json:"defaultFrpServer"`
		FrpEnable                 string `json:"frpEnable"`
	} `json:"data"`
}

func GetDeviceName() (string, error) {
	url := fmt.Sprintf("http://%s/system/status", os.Getenv("OLARESD_HOST"))
	var result SystemStatusResponse
	client := resty.New()
	resp, err := client.R().SetResult(&result).Get(url)
	if err != nil {
		klog.Errorf("failed to send request to olaresd %v", err)
		return "", err
	}
	if resp.StatusCode() != http.StatusOK {
		klog.Errorf("failed to get system status from olaresd %v", err)
		return "", errors.New(string(resp.Body()))
	}
	if result.Code != http.StatusOK {
		return "", fmt.Errorf("not exepcted result code: %v,message: %v", result.Code, result.Message)
	}
	klog.Infof("getDeviceName: %#v", result.Data)
	return result.Data.DeviceName, nil
}

const (
	PendingKindInSufficientGPU = "InsufficientGPU"
)

func GetPendingKind(ctrlClient client.Client, pod *corev1.Pod) (string, error) {
	fieldSelector := fields.OneTermEqualSelector("involvedObject.name", pod.Name)
	var eventList corev1.EventList
	err := ctrlClient.List(context.TODO(), &eventList, &client.ListOptions{
		FieldSelector: fieldSelector,
	})
	if err != nil {
		return "", err
	}

	pendingKind := ""
	for _, event := range eventList.Items {
		if event.Reason == PendingKindInSufficientGPU {
			pendingKind = PendingKindInSufficientGPU
		}
	}
	return pendingKind, nil
}

func IsMinikubeNode(node *corev1.Node) bool {
	if node == nil {
		return false
	}
	for k := range node.Labels {
		if strings.HasPrefix(k, minikubeLabelPrefix) {
			return true
		}
	}
	if strings.HasPrefix(node.Spec.ProviderID, minikubeKICProvider) {
		return true
	}
	if strings.Contains(strings.ToLower(node.Name), "minikube") {
		return true
	}
	return false
}

func IsRunningOnMinikube(ctx context.Context, client kubernetes.Interface) (bool, error) {
	if nodeName := os.Getenv("NODE_NAME"); nodeName != "" {
		node, err := client.CoreV1().Nodes().Get(ctx, nodeName, metav1.GetOptions{})
		if err != nil {
			return false, err
		}
		return IsMinikubeNode(node), nil
	}
	nodes, err := client.CoreV1().Nodes().List(ctx, metav1.ListOptions{})
	if err != nil {
		return false, err
	}
	for i := range nodes.Items {
		if IsMinikubeNode(&nodes.Items[i]) {
			return true, nil
		}
	}
	return false, nil
}
