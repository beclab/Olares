package utils

import (
	"bufio"
	"bytes"
	"context"
	"errors"
	"net/url"
	"os"
	"strings"

	bflconst "bytetrade.io/web3os/bfl/pkg/constants"
	"github.com/beclab/Olares/daemon/pkg/commands"
	"github.com/beclab/Olares/daemon/pkg/nets"
	"github.com/joho/godotenv"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	k8sruntime "k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"k8s.io/client-go/dynamic"
	"k8s.io/client-go/kubernetes"
	"k8s.io/klog/v2"
	"k8s.io/utils/pointer"
	ctrl "sigs.k8s.io/controller-runtime"

	sysv1 "bytetrade.io/web3os/app-service/api/sys.bytetrade.io/v1alpha1"
)

const (
	RoleOwner = "owner"
)

func GetKubeClient() (kubernetes.Interface, error) {
	config, err := ctrl.GetConfig()
	if err != nil {
		klog.Error("get k8s config error, ", err)
		return nil, err
	}

	client, err := kubernetes.NewForConfig(config)
	if err != nil {
		klog.Error("get k8s client error, ", err)
		return nil, err
	}

	return client, nil
}

func GetDynamicClient() (dynamic.Interface, error) {
	config, err := ctrl.GetConfig()
	if err != nil {
		klog.Error("get k8s config error, ", err)
		return nil, err
	}

	client, err := dynamic.NewForConfig(config)
	if err != nil {
		klog.Error("get k8s dynamic client error, ", err)
		return nil, err
	}

	return client, nil
}

func IsTerminusInitialized(ctx context.Context, client dynamic.Interface) (initialized bool, failed bool, err error) {
	users, err := client.Resource(UserGVR).List(ctx, metav1.ListOptions{})
	if err != nil {
		klog.Error("list user error, ", err)
		initialized = false
		failed = false
		return
	}

	for _, u := range users.Items {
		role, ok := u.GetAnnotations()[bflconst.UserAnnotationOwnerRole]
		if !ok {
			continue
		}

		if role == RoleOwner {
			status, ok := u.GetAnnotations()[bflconst.UserTerminusWizardStatus]
			if !ok {
				initialized = false
				failed = false
				return
			}
			initialized = status == string(bflconst.Completed)
			failed = (status == string(bflconst.SystemActivateFailed) ||
				status == string(bflconst.NetworkActivateFailed))
			return
		}
	}

	return
}

func IsTerminusInitializing(ctx context.Context, client dynamic.Interface) (bool, error) {
	user, err := GetAdminUser(ctx, client)
	if err != nil {
		return false, err
	}

	if user == nil {
		return false, nil
	}

	status, ok := user.GetAnnotations()[bflconst.UserTerminusWizardStatus]
	if !ok {
		return false, nil
	}

	return status != string(bflconst.Completed), nil
}

func IsTerminusRunning(ctx context.Context, client kubernetes.Interface) (bool, error) {
	pods, err := client.CoreV1().Pods("").List(ctx, metav1.ListOptions{})
	if err != nil {
		klog.Error("list pods error, ", err)
		return false, err
	}

	for _, pod := range pods.Items {
		if isKeyPod(&pod) {
			switch pod.Status.Phase {
			case corev1.PodRunning, corev1.PodSucceeded:
				continue
			default:
				return false, nil
			}
		}
	}

	return true, nil
}

func IsIpChanged(ctx context.Context, installed bool) (bool, error) {
	ips, err := nets.GetInternalIpv4Addr()
	if err != nil {
		return false, err
	}

	masterIpFromETCD, err := MasterNodeIp(installed)
	if err != nil {
		return false, err
	}

	for _, ip := range ips {
		hostIps, err := nets.LookupHostIps()
		if err != nil {
			return false, err
		}

		for _, hostIp := range hostIps {
			if hostIp == ip.IP {
				klog.V(8).Info("host ip is the same as internal ip of interface, ", hostIp, ", ", ip.IP)

				if !installed {
					// terminus not installed
					if masterIpFromETCD == "" {
						return false, nil
					}

					if masterIpFromETCD == ip.IP {
						return false, nil
					}

					return true, nil
				}

				kubeClient, err := GetKubeClient()
				if err != nil {
					klog.Error("get kube client error, ", err)
					return false, nil
				}

				_, nodeIp, nodeRole, err := GetThisNodeName(ctx, kubeClient)
				if err != nil {
					klog.Warning("get this node name error, ", err, ", try to compare with etcd ip")
					if masterIpFromETCD == "" {
						klog.Info("master node ip not found, mybe it's a worker node")
						return false, nil
					}

					if masterIpFromETCD == ip.IP {
						return false, nil
					}

					klog.Info("master node ip from etcd is not the same as internal ip of interface, ", masterIpFromETCD, ", ", hostIp, ", ", ip.IP)
					return true, nil

				}

				if nodeRole == "master" && nodeIp == ip.IP {
					return false, nil
				}

				// FIXME:(BUG) worker node will not work with this check
				if nodeRole == "worker" {
					return false, nil
				}

				klog.Info("node is master and node ip is not the same as internal ip of interface, ", nodeIp, ", ", hostIp, ", ", ip.IP)
				return true, nil
			}
		} // end for host ips
	} // end for interface ips

	klog.Info("no host ip is the same as internal ip of interface, ", ips)
	return true, nil
}

func MasterNodeIp(installed bool) (addr string, err error) {
	if installed {
		// get master node ip from etcd
		var (
			envs map[string]string
			url  *url.URL
		)
		etcEnvPath := "/etc/etcd.env"
		envs, err = godotenv.Read(etcEnvPath)
		if err != nil {
			if os.IsNotExist(err) {
				return "", nil
			}

			klog.Error("read etcd env file error, ", err)
			return
		}

		etcdListen, ok := envs["ETCD_LISTEN_PEER_URLS"]
		if !ok {
			err = errors.New("cannot find the cluster ip")
			klog.Error(err)

			return
		}

		url, err = url.Parse(etcdListen)
		if err != nil {
			klog.Error("etcd listen url is invalid, ", err, ", ", etcdListen)
			return
		}

		addr = url.Hostname()
		return
	} else {
		// get master node ip from redis
		var (
			data []byte
		)
		data, err = os.ReadFile(commands.REDIS_CONF)
		if err != nil {
			if os.IsNotExist(err) {
				// juicefs not installed
				return nets.GetHostIp()
			}
			klog.Error("read redis file error, ", err)
			return
		}

		r := bufio.NewReader(bytes.NewBuffer(data))
		for {
			var line string
			line, err = r.ReadString('\n')
			if err != nil {
				if err.Error() != "EOF" {
					klog.Errorf("redis conf read error: %s", err)
					return
				}

				// end of file
				err = nil
				return
			}

			token := strings.Split(strings.TrimSpace(line), " ")
			if len(token) < 2 {
				continue
			}

			if token[0] == "bind" {
				return token[1], nil
			}
		}
	}
}

func GetAdminUserJws(ctx context.Context, client dynamic.Interface) (string, error) {
	user, err := GetAdminUser(ctx, client)
	if err != nil {
		return "", err
	}

	if user == nil {
		return "", errors.New("user not found")
	}

	jws, ok := user.GetAnnotations()[bflconst.UserCertManagerJWSToken]
	if !ok {
		return "", errors.New("jws not found")
	}

	return jws, nil

}

func GetAdminUserTerminusName(ctx context.Context, client dynamic.Interface) (string, error) {
	user, err := GetAdminUser(ctx, client)
	if err != nil {
		return "", err
	}

	if user == nil {
		return "", errors.New("user not found")
	}

	name, ok := user.GetAnnotations()[bflconst.UserAnnotationTerminusNameKey]
	if !ok {
		return "", errors.New("olares name not found")
	}

	return name, nil

}

type Filter func(u *unstructured.Unstructured) bool

func GetAdminUser(ctx context.Context, client dynamic.Interface) (*unstructured.Unstructured, error) {
	u, err := ListUsers(ctx, client, func(u *unstructured.Unstructured) bool {
		role, ok := u.GetAnnotations()[bflconst.UserAnnotationOwnerRole]
		if !ok {
			return false
		}
		return role == RoleOwner
	})
	if err != nil {
		klog.Error("list user error, ", err)
		return nil, err
	}

	if len(u) == 0 {
		klog.Info("admin user not found")
		return nil, nil
	}

	return u[0], nil
}

func ListUsers(ctx context.Context, client dynamic.Interface, filters ...Filter) ([]*unstructured.Unstructured, error) {
	users, err := client.Resource(UserGVR).List(ctx, metav1.ListOptions{})
	if err != nil {
		klog.Error("list user error, ", err)
		return nil, err
	}

	var userList []*unstructured.Unstructured
	for _, u := range users.Items {
		var skip bool
		for _, filter := range filters {
			if !filter(&u) {
				skip = true
				break
			}
		}
		if skip {
			continue
		}

		userList = append(userList, &u)
	}

	return userList, nil
}

func isKeyPod(pod *corev1.Pod) bool {
	return strings.HasPrefix(pod.Namespace, "user-space") ||
		strings.HasPrefix(pod.Namespace, "user-system") ||
		pod.Namespace == "os-framework" ||
		pod.Namespace == "os-network" ||
		pod.Namespace == "os-platform" ||
		pod.Namespace == "os-gpu"
}

func GetTerminusInfo(ctx context.Context, client dynamic.Interface) (*sysv1.Terminus, error) {
	gvr := schema.GroupVersionResource{
		Group:    sysv1.GroupVersion.Group,
		Version:  sysv1.GroupVersion.Version,
		Resource: "terminus",
	}

	data, err := client.Resource(gvr).Get(ctx, "terminus", metav1.GetOptions{})
	if err != nil {
		klog.Error("cannot get terminus cr, ", err)
		return nil, err
	}

	var terminus sysv1.Terminus
	err = k8sruntime.DefaultUnstructuredConverter.FromUnstructured(data.Object, &terminus)
	if err != nil {
		klog.Error("decode data error, ", err)
		return nil, err
	}

	return &terminus, nil
}

func GetTerminusVersion(ctx context.Context, client dynamic.Interface) (*string, error) {
	terminus, err := GetTerminusInfo(ctx, client)
	if err != nil {
		return nil, err
	}

	return &terminus.Spec.Version, nil
}

func GetTerminusInstalledTime(ctx context.Context, dynamicClient dynamic.Interface, client kubernetes.Interface) (*int64, error) {
	// FIXME: record the time
	adminUser, err := GetAdminUser(ctx, dynamicClient)
	if err != nil {
		klog.Error("get admin user error, ", err)
		return nil, err
	}

	if adminUser == nil {
		return nil, nil
	}

	deploy, err := client.AppsV1().Deployments("user-system-"+adminUser.GetName()).
		Get(ctx, "system-server", metav1.GetOptions{})
	if err != nil {
		klog.Error("get deploy error, ", err)
		return nil, err
	}

	return pointer.Int64(deploy.CreationTimestamp.Unix()), nil
}

func GetTerminusInitializedTime(ctx context.Context, client kubernetes.Interface) (*int64, error) {
	deploy, err := client.AppsV1().Deployments("os-network").
		Get(ctx, "l4-bfl-proxy", metav1.GetOptions{})
	if err != nil {
		klog.Error("get deploy error, ", err)
		return nil, err
	}

	return pointer.Int64(deploy.CreationTimestamp.Unix()), nil
}

func GetThisNodeName(ctx context.Context, client kubernetes.Interface) (nodeName, nodeIp, nodeRole string, err error) {
	nodes, err := client.CoreV1().Nodes().List(ctx, metav1.ListOptions{})
	if err != nil {
		klog.Error("list nodes error, ", err)
		return
	}

	hostname, err := os.Hostname()
	if err != nil {
		klog.Error("get hostname error, ", err)
		return
	}

	ips, err := nets.LookupHostIps()
	if err != nil {
		klog.Error("get host ip error, ", err)
		return
	}

	for _, node := range nodes.Items {
		var foundIp, foundHost bool
		for _, address := range node.Status.Addresses {
			switch address.Type {
			case corev1.NodeHostName:
				foundHost = address.Address == hostname
			case corev1.NodeInternalIP:
				for _, ip := range ips {
					foundIp = address.Address == ip
					if foundIp {
						nodeIp = address.Address
						break
					}
				}
			}

			if foundHost && foundIp {
				nodeName = node.Name

				if cp, ok := node.Labels["node-role.kubernetes.io/control-plane"]; ok && cp != "false" {
					nodeRole = "master"
				} else if master, ok := node.Labels["node-role.kubernetes.io/master"]; ok && master != "false" {
					nodeRole = "master"
				} else {
					nodeRole = "worker"
				}
				return
			}
		}
	}

	err = os.ErrNotExist
	return
}

func GetUserspacePvcHostPath(ctx context.Context, user string, client kubernetes.Interface) (string, error) {
	namespace := "user-space-" + user
	bfl, err := client.AppsV1().StatefulSets(namespace).Get(ctx, "bfl", metav1.GetOptions{})
	if err != nil {
		klog.Error("find bfl error, ", err)
		return "", err
	}

	hostpath, ok := bfl.Annotations["userspace_hostpath"]
	if !ok {
		return "", errors.New("hostpath not found")
	}

	return hostpath, nil
}

func GetNodesPressure(ctx context.Context, client kubernetes.Interface) (map[string][]NodePressure, error) {
	nodes, err := client.CoreV1().Nodes().List(ctx, metav1.ListOptions{})
	if err != nil {
		klog.Error("list nodes error, ", err)
		return nil, err
	}

	status := make(map[string][]NodePressure)
	for _, node := range nodes.Items {
		for _, condition := range node.Status.Conditions {
			if condition.Type != corev1.NodeReady && condition.Status == corev1.ConditionTrue {
				status[node.Name] = append(status[node.Name], NodePressure{Type: condition.Type, Message: condition.Message})
			}
		}
	}

	return status, nil
}
