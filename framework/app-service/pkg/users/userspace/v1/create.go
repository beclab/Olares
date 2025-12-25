package userspace

import (
	"context"
	"crypto"
	"encoding/hex"
	"errors"
	"fmt"
	"math/rand"
	"os"
	"strconv"
	"strings"
	"time"

	"helm.sh/helm/v3/pkg/storage/driver"
	apierrors "k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/labels"
	"k8s.io/apimachinery/pkg/types"

	"github.com/beclab/Olares/framework/app-service/pkg/constants"
	"github.com/beclab/Olares/framework/app-service/pkg/helm"
	"sigs.k8s.io/controller-runtime/pkg/client"

	"github.com/beclab/Olares/framework/app-service/pkg/kubesphere"
	"github.com/beclab/Olares/framework/app-service/pkg/users/userspace"
	"github.com/beclab/Olares/framework/app-service/pkg/utils"

	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/rest"
	"k8s.io/klog/v2"
)

type Creator struct {
	Deleter
	rand *rand.Rand
}

const (
	USER_SPACE_ROLE = "admin"
	MAX_RAND_INT    = 1000000
	OlaresRootPath  = "OLARES_SYSTEM_ROOT_PATH"
	DefaultRootPath = "/olares"
)

var (
	REQUIRE_PERMISSION_APPS = []string{
		// "files",
		"portfolio",
		"vault",
		"desktop",
		"message",
		"rss",
		"search",
	}
)

func init() {
	envRPA := os.Getenv("REQUIRE_PERMISSION_APPS")
	if envRPA != "" {
		apps := strings.Split(envRPA, ",")
		REQUIRE_PERMISSION_APPS = apps
	}
}

func NewCreator(client client.Client, config *rest.Config, user string) *Creator {
	return &Creator{
		Deleter: Deleter{
			Client:    client,
			k8sConfig: config,
			user:      user,
		},
		rand: rand.New(rand.NewSource(time.Now().Unix())),
	}
}

func (c *Creator) CreateUserApps(ctx context.Context) (int32, int32, error) {
	//userspace, userspaceRoleBinding, err := c.createNamespace(ctx)
	userspaceNs := fmt.Sprintf("user-space-%s", c.user)
	//err := c.createNamespace(ctx)
	//if err != nil {
	//	klog.Errorf("failed to create namespace %s", userspaceNs)
	//	return 0, 0, err
	//}

	//if err != nil {
	//	return 0, 0, err
	//}
	//clear = true

	//defer func() {
	//	if clear {
	//		klog.Infof("Start clear process err=%v", err)
	//		if userspace != "" {
	//			// clear context should not be canceled
	//			clearCtx := context.Background()
	//			if launcherReleaseName != "" {
	//				klog.Warningf("Clear launcher failed launcherReleaseName=%s", launcherReleaseName)
	//
	//				err1 := c.clearLauncher(clearCtx, launcherReleaseName, userspace)
	//				if err1 != nil {
	//					klog.Warningf("Clear launcher failed err=%v", err1)
	//				}
	//			}
	//
	//			klog.Warningf("Clear userspace failed userspace=%s", userspace)
	//
	//			err1 := c.deleteNamespace(clearCtx, userspace, userspaceRoleBinding)
	//			if err1 != nil {
	//				klog.Warningf("Failed to delete namespace err=%v", err1)
	//			}
	//		}
	//	}
	//
	//}()

	actionCfg, settings, err := helm.InitConfig(c.k8sConfig, userspaceNs)
	if err != nil {
		return 0, 0, err
	}
	c.helmCfg.ActionCfg = actionCfg
	c.helmCfg.Settings = settings

	_, err = c.installLauncher(ctx, userspaceNs)
	if err != nil {
		klog.Errorf("failed to install launcher in ns %s, %v", userspaceNs, err)
		return 0, 0, err
	}

	var bfl *corev1.Pod
	if bfl, err = c.checkLauncher(ctx, userspaceNs, checkLauncherRunning); err != nil {
		klog.Errorf("check launcher failed %v", err)
		return 0, 0, err
	}
	klog.Infof("c.name: %s", c.user)
	klog.Infof("userspaceNs: %s", userspaceNs)

	klog.Infof("bflname: %s,namespace: %s", bfl.Name, bfl.Namespace)
	err = c.installSysApps(ctx, bfl)
	if err != nil {
		klog.Errorf("failed to install sys apps %v", err)
		return 0, 0, err
	}

	desktopPort, wizardPort, err := c.checkDesktopRunning(ctx, userspaceNs)

	return desktopPort, wizardPort, err
}

func (c *Creator) createNamespace(ctx context.Context) error {

	// ns and rolebinding creation moves to bfl

	// create namespace user-space-<user>
	userspaceNs := fmt.Sprintf("user-space-%s", c.user)
	userSystemNs := fmt.Sprintf("user-system-%s", c.user)

	// create user-space namespace
	ns := corev1.Namespace{
		ObjectMeta: metav1.ObjectMeta{
			Name: userspaceNs,
			// TODO:hys
			Annotations: map[string]string{
				"kubesphere.io/creator": "",
			},
			Finalizers: []string{
				"finalizers.kubesphere.io/namespaces",
			},
		},
	}
	err := c.Create(ctx, &ns)
	if err != nil && !apierrors.IsAlreadyExists(err) {
		klog.Errorf("failed to create user-space namespace %v", err)
		return err
	}

	// create user-system namespace
	userSystemNamespace := corev1.Namespace{
		ObjectMeta: metav1.ObjectMeta{
			Name: userSystemNs,
			Annotations: map[string]string{
				"kubesphere.io/creator": "",
			},
			Finalizers: []string{
				"finalizers.kubesphere.io/namespaces",
			},
		},
	}
	err = c.Create(ctx, &userSystemNamespace)
	if err != nil && !apierrors.IsAlreadyExists(err) {
		klog.Errorf("failed to create user-system namespace %v", err)
		return err
	}
	return nil
}

func (c *Creator) installSysApps(ctx context.Context, bflPod *corev1.Pod) error {
	vals := make(map[string]interface{})

	vals["bfl"] = map[string]interface{}{
		"username": c.user,
		"nodeName": bflPod.Spec.NodeName,
	}

	vals["global"] = map[string]interface{}{
		"bfl": map[string]string{
			"username": c.user,
		},
	}
	rootPath := DefaultRootPath
	if os.Getenv(OlaresRootPath) != "" {
		rootPath = os.Getenv(OlaresRootPath)
	}
	vals["rootPath"] = rootPath

	pvcData, err := c.findPVC(ctx, bflPod.Namespace)
	if err != nil {
		return err
	}

	var pv corev1.PersistentVolume
	err = c.Get(ctx, types.NamespacedName{Name: pvcData.userspacePv}, &pv)
	if err != nil {
		klog.Errorf("failed to get pv %v", err)
		return err
	}

	// pvc to mount with filebrowser
	pvPath := pv.Spec.HostPath.Path
	vals["pvc"] = map[string]interface{}{
		"userspace": pvcData.userspacePvc,
	}

	vals["userspace"] = map[string]interface{}{
		"appCache": pvcData.appCacheHostPath,
		"userData": fmt.Sprintf("%s/Home", pvPath),
		"appData":  fmt.Sprintf("%s/Data", pvPath),
		"dbdata":   pvcData.dbdataHostPath,
	}

	// generate app key and secret of permission required apps
	osVals := make(map[string]interface{})
	for _, r := range REQUIRE_PERMISSION_APPS {
		k, s := c.genAppkeyAndSecret(r)
		osVals[r] = map[string]interface{}{
			"appKey":    k,
			"appSecret": s,
		}
	}
	if appstoreVal, exists := osVals["appstore"]; exists {
		if appstoreMap, ok := appstoreVal.(map[string]interface{}); ok {
			appstoreMap["marketProvider"] = os.Getenv("MARKET_PROVIDER")
		}
	} else {
		osVals["appstore"] = map[string]interface{}{
			"marketProvider": os.Getenv("MARKET_PROVIDER"),
		}
	}
	vals["os"] = osVals

	var arch string
	var nodes corev1.NodeList

	err = c.List(ctx, &nodes)
	if err != nil {
		return err
	}
	for _, node := range nodes.Items {
		arch = node.Labels["kubernetes.io/arch"]
		break
	}
	vals["cluster"] = map[string]interface{}{
		"arch": arch,
	}

	gpuType, err := utils.FindGpuTypeFromNodes(&nodes)
	if err != nil {
		return err
	}
	vals["gpu"] = gpuType

	userIndex, userSubnet, err := c.getUserSubnet(ctx)
	if err != nil {
		return err
	}
	vals["tailscaleUserIndex"] = userIndex
	vals["tailscaleUserSubnet"] = userSubnet

	sysApps, err := userspace.GetAppsFromDirectory(constants.UserChartsPath + "/apps")
	if err != nil {
		return err
	}
	for _, appname := range sysApps {
		name := helm.ReleaseName(appname, c.user)
		_, err = c.helmCfg.ActionCfg.Releases.Last(name)
		if err != nil {
			if errors.Is(err, driver.ErrReleaseNotFound) {
				installErr := helm.InstallCharts(ctx, c.helmCfg.ActionCfg, c.helmCfg.Settings,
					name, constants.UserChartsPath+"/apps/"+appname, "", bflPod.Namespace, vals)
				if installErr != nil && !errors.Is(installErr, driver.ErrReleaseExists) {
					klog.Errorf("failed to install sys app:%s, %v", name, installErr)
					return installErr
				}
			} else {
				klog.Errorf("failed to get last release for %s", name)
				return err
			}
		}

	}
	return nil
}

func (c *Creator) installLauncher(ctx context.Context, userspace string) (string, error) {
	vals := make(map[string]interface{})

	k, s := c.genAppkeyAndSecret("bfl")
	vals["bfl"] = map[string]interface{}{
		"username":  c.user,
		"appKey":    k,
		"appSecret": s,
	}
	rootPath := DefaultRootPath
	if os.Getenv(OlaresRootPath) != "" {
		rootPath = os.Getenv(OlaresRootPath)
	}
	vals["rootPath"] = rootPath
	name := helm.ReleaseName("launcher", c.user)
	_, err := c.helmCfg.ActionCfg.Releases.Last(name)
	if err != nil {
		if errors.Is(err, driver.ErrReleaseNotFound) {
			installErr := helm.InstallCharts(ctx, c.helmCfg.ActionCfg, c.helmCfg.Settings, name, constants.UserChartsPath+"/launcher", "", userspace, vals)
			if installErr != nil && !errors.Is(installErr, driver.ErrReleaseExists) {
				klog.Errorf("failed to install launcher %v", err)
				return "", installErr
			}
		} else {
			klog.Errorf("failed to get last release for %s", name)
			return "", err
		}
	}

	return name, nil
}

func (c *Creator) findBflAPIPort(ctx context.Context, namespace string) (int32, error) {
	var bflSvc corev1.Service
	err := c.Get(ctx, types.NamespacedName{Namespace: namespace, Name: "bfl"}, &bflSvc)
	if err != nil {
		return 0, err
	}

	var port int32 = 0
	for _, p := range bflSvc.Spec.Ports {
		if p.Name == "api" {
			port = p.NodePort
			break
		}
	}

	return port, nil
}

func (c *Creator) getUserSubnet(ctx context.Context) (string, string, error) {
	userIndex, err := kubesphere.GetUserIndexByName(ctx, c.user)
	if err != nil {
		return "0", "", err
	}
	userNum := os.Getenv("OLARES_MAX_USERS")
	if userNum == "" {
		userNum = "1024"
	}
	num, err := strconv.ParseInt(userNum, 10, 64)
	if err != nil {
		return "0", "", err
	}
	userSubnet := utils.SubnetSplit(int(num))[userIndex]
	if userSubnet == nil || userSubnet.String() == "" {
		return "0", "", errors.New("empty userSubnet")
	}
	return userIndex, userSubnet.String(), nil
}

func (c *Creator) checkDesktopRunning(ctx context.Context, userspace string) (int32, int32, error) {
	ticker := time.NewTicker(1 * time.Second)
	timeout := time.NewTimer(5 * time.Minute)

	defer func() {
		ticker.Stop()
		timeout.Stop()
	}()

	for {
		select {
		case <-ticker.C:
			var pods corev1.PodList
			selector, _ := labels.Parse("app=wizard")
			err := c.List(ctx, &pods, &client.ListOptions{LabelSelector: selector})

			if err != nil {
				return 0, 0, err
			}

			if len(pods.Items) == 0 {
				klog.Warningf("Failed to find user's wizard userspace=%s", userspace)
				continue
			}

			wizard := pods.Items[0] // a single bfl per user
			if wizard.Status.Phase == "Running" {
				// find desktop port
				// desktop, err := c.clientSet.KubeClient.Kubernetes().CoreV1().Services(userspace).Get(ctx, "edge-desktop", metav1.GetOptions{})
				// if err != nil {
				// 	return 0, 0, err
				// }

				//wizardPort := wizard.Spec.Ports[0].NodePort

				return 0, 0, nil
			}
		case <-timeout.C:
			return 0, 0, fmt.Errorf("user's wizard checking timeout error. [%s]", userspace)
		}
	}
}

func (c *Creator) genAppkeyAndSecret(app string) (string, string) {
	key := fmt.Sprintf("bytetrade_%s_%d", app, c.rand.Intn(MAX_RAND_INT))
	secret := md5(fmt.Sprintf("%s|%d", key, time.Now().Unix()))

	return key, secret[:16]
}

func md5(str string) string {
	h := crypto.MD5.New()
	h.Write([]byte(str))
	return hex.EncodeToString(h.Sum(nil))
}
