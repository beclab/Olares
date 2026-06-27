package utils

import (
	"bufio"
	"bytes"
	"context"
	"errors"
	"fmt"
	"k8s.io/klog"
	"net/url"
	"os"
	"sort"
	"strings"
	"sync"

	bflconst "bytetrade.io/web3os/bfl/pkg/constants"
	"github.com/beclab/Olares/daemon/pkg/commands"
	"github.com/beclab/Olares/daemon/pkg/nets"
	"github.com/joho/godotenv"
	corev1 "k8s.io/api/core/v1"
	apixclientset "k8s.io/apiextensions-apiserver/pkg/client/clientset/clientset"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	k8sruntime "k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"k8s.io/client-go/dynamic"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/util/retry"
	"k8s.io/utils/pointer"
	ctrl "sigs.k8s.io/controller-runtime"

	"github.com/beclab/Olares/framework/app-service/pkg/appcfg"
	appv1alpha1 "github.com/beclab/api/api/app.bytetrade.io/v1alpha1"
	sysv1 "github.com/beclab/api/api/sys.bytetrade.io/v1alpha1"
	iamv1alpha2 "github.com/beclab/api/iam/v1alpha2"
	"github.com/beclab/api/manifest"
	"github.com/beclab/api/pkg/generated/clientset/versioned"
	nadutils "github.com/k8snetworkplumbingwg/network-attachment-definition-client/pkg/utils"
	apierrors "k8s.io/apimachinery/pkg/api/errors"
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

func GetAppClientSet() (versioned.Clientset, error) {
	config, err := ctrl.GetConfig()
	if err != nil {
		klog.Error("get k8s config error, ", err)
		return versioned.Clientset{}, err
	}

	client, err := versioned.NewForConfig(config)
	if err != nil {
		klog.Error("get app clientset error, ", err)
		return versioned.Clientset{}, err
	}

	return *client, nil
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
			case corev1.PodRunning:
				if !isPodReady(&pod) {
					return false, nil
				}
				continue
			case corev1.PodSucceeded:
				continue
			default:
				return false, nil
			}
		}
	}

	return true, nil
}

func IsIpChanged(ctx context.Context, installed bool) bool {
	ips, err := nets.GetInternalIpv4Addr()
	if err != nil {
		klog.Error("get internal ip error, ", err)
		return false
	}

	masterIpFromETCD, err := MasterNodeIp(installed)
	if err != nil {
		klog.Error("get master node ip error, ", err)
		return false
	}

	for _, ip := range ips {
		hostIps, err := nets.LookupHostIps()
		if err != nil {
			klog.Error("get host ip error, ", err)
			return false
		}

		for _, hostIp := range hostIps {
			if hostIp == ip.IP {
				klog.V(8).Info("host ip is the same as internal ip of interface, ", hostIp, ", ", ip.IP)

				if !installed {
					// terminus not installed
					if masterIpFromETCD == "" {
						return false
					}

					if masterIpFromETCD == ip.IP {
						return false
					}

					return true
				}

				kubeClient, err := GetKubeClient()
				if err != nil {
					klog.Error("get kube client error, ", err)
					return false
				}

				_, nodeIp, nodeRole, err := GetThisNodeName(ctx, kubeClient)
				if err != nil {
					klog.Warning("get this node name error, ", err, ", try to compare with etcd ip")
					if masterIpFromETCD == "" {
						klog.Info("master node ip not found, mybe it's a worker node")
						return false
					}

					if masterIpFromETCD == ip.IP {
						return false
					}

					klog.Info("master node ip from etcd is not the same as internal ip of interface, ", masterIpFromETCD, ", ", hostIp, ", ", ip.IP)
					return true

				}

				if nodeRole == "master" && nodeIp == ip.IP {
					return false
				}

				// FIXME:(BUG) worker node will not work with this check
				if nodeRole == "worker" {
					return false
				}

				klog.Info("node is master and node ip is not the same as internal ip of interface, ", nodeIp, ", ", hostIp, ", ", ip.IP)
				return true
			}
		} // end for host ips
	} // end for interface ips

	klog.Info("no host ip is the same as internal ip of interface, ", ips)
	return true
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

	ips, err := nets.LookupHostIps()
	if err != nil {
		klog.Error("get host ip error, ", err)
		return
	}

	for _, node := range nodes.Items {
		var foundIp bool
		for _, address := range node.Status.Addresses {
			switch address.Type {
			case corev1.NodeInternalIP:
				for _, ip := range ips {
					foundIp = address.Address == ip
					if foundIp {
						nodeIp = address.Address
						break
					}
				}
			}

			if foundIp {
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
				status[node.Name] = append(status[node.Name], NodePressure{Type: string(condition.Type), Message: condition.Message})
			}
		}
	}

	return status, nil
}

var (
	appUrlMu    sync.Mutex
	appUrlSig   string
	appUrlCache []string
)

// appUrlSignature returns a stable signature of the inputs that determine the
// application entrance URLs: the set of applications and the set of users (whose
// zone annotation feeds the URLs). Any change to an app spec or a user bumps its
// resourceVersion, so this detects when GetApplicationUrlAll must recompute.
func appUrlSignature(apps []appv1alpha1.Application, users []*unstructured.Unstructured) string {
	appSigs := make([]string, 0, len(apps))
	for i := range apps {
		appSigs = append(appSigs, apps[i].Name+":"+apps[i].ResourceVersion)
	}
	sort.Strings(appSigs)

	userSigs := make([]string, 0, len(users))
	for _, u := range users {
		userSigs = append(userSigs, u.GetName()+":"+u.GetResourceVersion())
	}
	sort.Strings(userSigs)

	var b strings.Builder
	b.WriteString("apps\n")
	for _, s := range appSigs {
		b.WriteString(s)
		b.WriteByte('\n')
	}
	b.WriteString("users\n")
	for _, s := range userSigs {
		b.WriteString(s)
		b.WriteByte('\n')
	}
	return b.String()
}

func GetApplicationUrlAll(ctx context.Context) ([]string, error) {
	clientset, err := GetAppClientSet()
	if err != nil {
		klog.Error("get app clientset error, ", err)
		return nil, err
	}

	apps, err := clientset.AppV1alpha1().Applications().List(ctx, metav1.ListOptions{})
	if err != nil {
		klog.Error("list applications error, ", err)
		return nil, err
	}

	dynamicClient, err := GetDynamicClient()
	if err != nil {
		klog.Error("get dynamic client error, ", err)
		return nil, err
	}
	users, err := ListUsers(ctx, dynamicClient)
	if err != nil {
		klog.Error("list users error, ", err)
		return nil, err
	}

	// Skip the per-app entrance URL recompute (each app otherwise triggers
	// GetUserZone/GenEntranceURL API calls) when neither the apps nor the users
	// changed since the last successful computation.
	sig := appUrlSignature(apps.Items, users)
	appUrlMu.Lock()
	if appUrlCache != nil && sig == appUrlSig {
		cached := make([]string, len(appUrlCache))
		copy(cached, appUrlCache)
		appUrlMu.Unlock()
		return cached, nil
	}
	appUrlMu.Unlock()

	// Owner zone is just the user's zone annotation, so read it from the user
	// list instead of building an IAM REST client and doing a per-app lookup.
	zoneMap := make(map[string]string, len(users))
	for _, u := range users {
		zoneMap[u.GetName()] = u.GetAnnotations()[iamv1alpha2.UserAnnotationZoneKey]
	}

	urls := make([]string, 0)
	for _, app := range apps.Items {
		var entrances []appcfg.Entrance
		var err error
		zone, ok := zoneMap[app.Spec.Owner]
		if !ok {
			continue
		}
		if !appv1alpha1.IsShared(&app) {
			entrances, err = appcfg.GenEntranceURL(ctx, &app)
			if err != nil {
				klog.Error("generate application entrance url error, ", err, ", ", app.Name)
				continue
			}
			if zone == "" {
				continue
			}
			// GenEntranceURL / BatchGenSharedAppEntranceURL only fill in the
			// default appid-based URLs. User-customized third-level domains live
			// in the "customDomain" setting, so without this the intranet (.local)
			// layer never learns about custom subdomains and cannot resolve them.
			thirdLevelURLs := app.ThirdLevelCusDomainURLs(zone, app.Spec.Owner)
			for _, tl := range thirdLevelURLs {
				entrances = append(entrances, appcfg.Entrance{URL: tl})
			}
		} else {

			entrances, err = BatchGenSharedAppEntranceURL(ctx, &app)
			if err != nil {
				klog.Error("generate shared application entrance url error, ", err, ", ", app.Name)
				continue
			}
		}

		for _, entrance := range entrances {
			urls = append(urls, entrance.URL)
		}
	}

	appUrlMu.Lock()
	appUrlSig = sig
	appUrlCache = urls
	appUrlMu.Unlock()

	return urls, nil
}

func GetApixClient() (apixclientset.Interface, error) {
	config, err := ctrl.GetConfig()
	if err != nil {
		klog.Error("get k8s config error, ", err)
		return nil, err
	}

	client, err := apixclientset.NewForConfig(config)
	if err != nil {
		klog.Error("get k8s apix client error, ", err)
		return nil, err
	}

	return client, nil
}

func GetOverlayGatewaySupportedApps(ctx context.Context, user string) ([]OverlayGatewaySupportedApp, error) {
	clientset, err := GetAppClientSet()
	if err != nil {
		klog.Error("get app clientset error, ", err)
		return nil, err
	}

	kubeClient, err := GetKubeClient()
	if err != nil {
		klog.Error("get kube client error, ", err)
		return nil, err
	}

	appMgrs, err := clientset.AppV1alpha1().ApplicationManagers().List(ctx, metav1.ListOptions{})
	if err != nil {
		klog.Error("list applications error, ", err)
		return nil, err
	}

	var supportedApps []OverlayGatewaySupportedApp
	for _, appMgr := range appMgrs.Items {
		if appMgr.Spec.Config == "" {
			continue
		}

		// check if the app is supported by the overlay gateway
		var appConfig appcfg.ApplicationConfig
		err := appcfg.GetAppConfig(&appMgr, &appConfig)
		if err != nil {
			klog.Error("get app config error, ", err)
			continue
		}

		// check if the app is supported by the overlay gateway
		if appConfig.OverlayGateway.Enable {
			// check if the app is enabled
			app, err := clientset.AppV1alpha1().Applications().Get(ctx, appMgr.Name, metav1.GetOptions{})
			if err != nil {
				klog.Error("get app error, ", err)
				continue
			}

			sharedApp := appv1alpha1.IsShared(app)
			if !sharedApp {
				if user != "" && app.Spec.Owner != user {
					continue
				}
			}

			enabled := strings.ToLower(app.Spec.Settings["enableOverlayGateway"]) == "true"
			sa := OverlayGatewaySupportedApp{
				AppResourceName: app.Name,
				AppName:         app.Spec.Name,
				Enabled:         enabled,
				Owner:           app.Spec.Owner,
				SharedApp:       sharedApp,
				Namespace:       app.Spec.Namespace,
				AppID:           app.Spec.Appid,
			}
			if enabled {
				pods, err := kubeClient.CoreV1().Pods(app.Spec.Namespace).List(ctx, metav1.ListOptions{})
				if err != nil {
					klog.Error("list pods error, ", err)
				} else {
					for _, pod := range pods.Items {
						if pod.Labels["applications.app.bytetrade.io/macvlan-init"] == "true" {
							statuses, err := nadutils.GetNetworkStatus(&pod)
							if err != nil {
								klog.Error("get network status error, ", err)
								continue
							}

							for _, status := range statuses {
								if status.Interface == "net1" {
									if len(status.IPs) == 0 {
										klog.Warningf("pod %s has no ip in net1, skip", pod.Name)
										continue
									}
									un := UnderlayNetwork{
										IP: status.IPs[0],
									}

									var entrances []manifest.OverlayEntrance
									for _, e := range appConfig.OverlayGateway.Entrances {
										if e.Workload == GetWorkloadNameFromPod(&pod) {
											entrances = append(entrances, e)
										}
									}
									un.Ports = entrances
									sa.UnderlayNetworks = append(sa.UnderlayNetworks, un)
									break
								}
							}

						}
					}
				}
			}
			supportedApps = append(supportedApps, sa)
		}

	}

	return supportedApps, nil
}

func UpdateApplicationSettings(ctx context.Context, appName string, option string, value string) error {
	clientset, err := GetAppClientSet()
	if err != nil {
		klog.Error("get app clientset error, ", err)
		return err
	}

	retry.RetryOnConflict(retry.DefaultRetry, func() error {
		app, err := clientset.AppV1alpha1().Applications().Get(ctx, appName, metav1.GetOptions{})
		if err != nil {
			klog.Error("get application error, ", err)
			return err
		}

		app.Spec.Settings[option] = value
		_, err = clientset.AppV1alpha1().Applications().Update(ctx, app, metav1.UpdateOptions{})
		if err != nil {
			klog.Error("update application settings error, ", err)
			return err
		}

		return nil
	})

	return nil
}

func RestartOverlayGatewaySupportedApps(ctx context.Context, apps []OverlayGatewaySupportedApp) error {
	client, err := GetKubeClient()
	if err != nil {
		klog.Error("get dynamic client error, ", err)
		return err
	}

	for _, app := range apps {
		// restart the app
		pods, err := client.CoreV1().Pods(app.Namespace).List(ctx, metav1.ListOptions{})
		if err != nil {
			if apierrors.IsNotFound(err) {
				continue
			}
			klog.Error("list pods error, ", err)
			return err
		}

		for _, pod := range pods.Items {
			if pod.Labels["applications.app.bytetrade.io/macvlan-init"] == "true" {
				err = client.CoreV1().Pods(app.Namespace).Delete(ctx, pod.Name, metav1.DeleteOptions{})
				if err != nil {
					klog.Error("delete pod error, ", err)
				}
			}
		}

	}

	return nil
}

func GetWorkloadNameFromPod(pod *corev1.Pod) string {
	podTemplateHash := pod.Labels["pod-template-hash"]
	podNameTokens := strings.Split(pod.Name, fmt.Sprintf("-%s-", podTemplateHash))
	return podNameTokens[0]
}

func BatchGenSharedAppEntranceURL(ctx context.Context, app *appv1alpha1.Application) ([]appcfg.Entrance, error) {
	if !appv1alpha1.IsShared(app) {
		return nil, nil
	}

	client, err := GetDynamicClient()
	if err != nil {
		klog.Error("get dynamic client error, ", err)
		return nil, err
	}

	users, err := ListUsers(ctx, client, func(u *unstructured.Unstructured) bool {
		return true
	})
	if err != nil {
		klog.Error("list user error, ", err)
		return nil, err
	}

	var entrances []appcfg.Entrance
	for _, u := range users {
		a := app.DeepCopy()
		a.Spec.Owner = u.GetName()
		_, err := appcfg.GenEntranceURL(ctx, a)
		if err != nil {
			klog.Error("generate entrance url error, ", err)
			continue
		}

		entrances = append(entrances, a.Spec.Entrances...)
		zone := u.GetAnnotations()["bytetrade.io/zone"]
		if zone == "" {
			continue
		}
		thirdLevelURLs := app.ThirdLevelCusDomainURLs(zone, u.GetName())
		for _, tl := range thirdLevelURLs {
			entrances = append(entrances, appcfg.Entrance{URL: tl})
		}
	}

	return entrances, nil
}

func GetUserRole(ctx context.Context, username string, client dynamic.Interface) (string, error) {
	user, err := client.Resource(UserGVR).Get(ctx, username, metav1.GetOptions{})
	if err != nil {
		klog.Error("get user error, ", err)
		return "", err
	}

	role, ok := user.GetAnnotations()[bflconst.UserAnnotationOwnerRole]
	if !ok {
		return "", errors.New("user role not found")
	}

	return role, nil
}

func isPodReady(pod *corev1.Pod) bool {
	hasReadyCondition := false
	for _, cond := range pod.Status.Conditions {
		// K8s 1.28+
		if cond.Type == "PodReadyToStartContainers" && cond.Status != corev1.ConditionTrue {
			return false
		}
		if cond.Type == corev1.PodReady {
			if cond.Status == corev1.ConditionTrue {
				hasReadyCondition = true
			}
		}
	}

	if !hasReadyCondition {
		return false
	}

	if len(pod.Status.ContainerStatuses) == 0 {
		return false
	}

	for _, containerStatus := range pod.Status.ContainerStatuses {
		if containerStatus.State.Running == nil {
			if containerStatus.State.Waiting != nil {
				return false
			}
			return false
		}

		if !containerStatus.Ready {
			return false
		}

		if containerStatus.State.Running.StartedAt.IsZero() {
			return false
		}
	}
	return true
}
