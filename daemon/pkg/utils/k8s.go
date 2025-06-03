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
	"bytetrade.io/web3os/terminusd/pkg/commands"
	"bytetrade.io/web3os/terminusd/pkg/nets"
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

		if role == bflconst.RolePlatformAdmin {
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

	for _, ip := range ips {
		hostIp, err := nets.GetHostIp()
		if err != nil {
			return false, err
		}

		masterIp, err := MasterNodeIp(installed)
		if err != nil {
			return false, err
		}

		if hostIp == ip.IP {
			if masterIp == "" {
				// terminus not installed
				return false, nil
			}

			if masterIp == ip.IP {
				return false, nil
			}

			klog.Info("get master node ip, ", masterIp, ", ", hostIp, ", ", ip.IP)
		}
	}

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

func GetAdminUser(ctx context.Context, client dynamic.Interface) (*unstructured.Unstructured, error) {
	users, err := client.Resource(UserGVR).List(ctx, metav1.ListOptions{})
	if err != nil {
		klog.Error("list user error, ", err)
		return nil, err
	}

	for _, u := range users.Items {
		role, ok := u.GetAnnotations()[bflconst.UserAnnotationOwnerRole]
		if !ok {
			continue
		}

		if role == bflconst.RolePlatformAdmin {
			return &u, nil
		}
	}

	return nil, nil
}

func isKeyPod(pod *corev1.Pod) bool {
	return strings.HasPrefix(pod.Namespace, "user-space") ||
		strings.HasPrefix(pod.Namespace, "user-system") ||
		pod.Namespace == "os-system"
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
	deploy, err := client.AppsV1().Deployments("os-system").
		Get(ctx, "l4-bfl-proxy", metav1.GetOptions{})
	if err != nil {
		klog.Error("get deploy error, ", err)
		return nil, err
	}

	return pointer.Int64(deploy.CreationTimestamp.Unix()), nil
}

func GetThisNodeName(ctx context.Context, client kubernetes.Interface) (nodeName, nodeIp string, err error) {
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

	ip, err := nets.GetHostIp()
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
				foundIp = address.Address == ip
				if foundIp {
					nodeIp = address.Address
				}
			}

			if foundHost && foundIp {
				nodeName = node.Name
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
