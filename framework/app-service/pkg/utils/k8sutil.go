package utils

import (
	"context"
	"errors"
	"fmt"
	"net"
	"net/url"
	"strconv"
	"strings"
	"time"

	sysv1alpha1 "bytetrade.io/web3os/app-service/api/sys.bytetrade.io/v1alpha1"
	"bytetrade.io/web3os/app-service/pkg/apiserver/api"
	"bytetrade.io/web3os/app-service/pkg/client/clientset"
	"bytetrade.io/web3os/app-service/pkg/constants"
	"bytetrade.io/web3os/app-service/pkg/prometheus"
	"bytetrade.io/web3os/app-service/pkg/users"
	iamv1alpha2 "github.com/beclab/api/iam/v1alpha2"

	srvconfig "github.com/containerd/containerd/services/server/config"
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

// CalicoTunnelAddrAnnotation annotation key for calico tunnel address.
const CalicoTunnelAddrAnnotation = "projectcalico.org/IPv4IPIPTunnelAddr"
const DefaultRegistry = "https://registry-1.docker.io"

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

func FindGpuTypeFromNodes(nodes *corev1.NodeList) (string, error) {
	gpuType := "none"
	if nodes == nil {
		return gpuType, errors.New("empty node list")
	}
	for _, n := range nodes.Items {
		if _, ok := n.Status.Capacity[constants.NvidiaGPU]; ok {
			if _, ok = n.Status.Capacity[constants.NvshareGPU]; ok {
				return "nvshare", nil

			}
			gpuType = "nvidia"
		}
		if _, ok := n.Status.Capacity[constants.VirtAiTechVGPU]; ok {
			return "virtaitech", nil
		}
	}
	return gpuType, nil
}

func IsNodeReady(node *corev1.Node) bool {
	for _, c := range node.Status.Conditions {
		if c.Type == corev1.NodeReady {
			return c.Status == corev1.ConditionTrue
		}
	}
	return false
}

func GetMirrorsEndpoint() (ep []string) {
	config := &srvconfig.Config{}
	err := srvconfig.LoadConfig("/etc/containerd/config.toml", config)
	if err != nil {
		klog.Infof("load mirrors endpoint failed err=%v", err)
		return
	}
	plugins := config.Plugins["io.containerd.grpc.v1.cri"]
	r := plugins.GetPath([]string{"registry", "mirrors", "docker.io", "endpoint"})
	if r == nil {
		return
	}
	for _, e := range r.([]interface{}) {
		ep = append(ep, e.(string))
	}
	return ep
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
