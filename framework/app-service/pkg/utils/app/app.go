package app

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"net"
	"net/http"
	"net/netip"
	"net/url"
	"path/filepath"
	"strings"
	"time"

	corev1 "k8s.io/api/core/v1"

	sysv1alpha1 "github.com/beclab/Olares/framework/app-service/api/sys.bytetrade.io/v1alpha1"
	"github.com/go-viper/mapstructure/v2"
	"sigs.k8s.io/controller-runtime/pkg/client"

	"github.com/beclab/Olares/framework/app-service/api/app.bytetrade.io/v1alpha1"
	"github.com/beclab/Olares/framework/app-service/pkg/appcfg"
	"github.com/beclab/Olares/framework/app-service/pkg/constants"
	"github.com/beclab/Olares/framework/app-service/pkg/generated/clientset/versioned"
	"github.com/beclab/Olares/framework/app-service/pkg/users/userspace"
	"github.com/beclab/Olares/framework/app-service/pkg/utils"
	"github.com/beclab/Olares/framework/app-service/pkg/utils/files"

	"github.com/Masterminds/semver/v3"
	"github.com/go-resty/resty/v2"
	"github.com/hashicorp/go-getter"
	"helm.sh/helm/v3/pkg/action"
	"helm.sh/helm/v3/pkg/chart"
	"helm.sh/helm/v3/pkg/repo"
	apierrors "k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/api/resource"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/util/rand"
	"k8s.io/apimachinery/pkg/util/sets"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/util/retry"
	"k8s.io/klog/v2"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/yaml"
)

const expectedTokenItems = 2
const AppCfgFileName = "OlaresManifest.yaml"

var (
	ErrInvalidAction     = errors.New("invalid action")
	ErrInvalidPortFormat = errors.New("invalid port format")
)
var protectedNamespace = []string{
	"default",
	"kube-node-lease",
	"kube-public",
	"kube-system",
	"kubekey-system",
	"kubesphere-controls-system",
	"kubesphere-monitoring-federated",
	"kubesphere-monitoring-system",
	"kubesphere-system",
	"user-space-",
	"user-system-",
	"os-platform",
	"os-framework",
	"os-network",
}

var forbidNamespace = []string{
	"default",
	"kube-node-lease",
	"kube-public",
	"kube-system",
	"kubekey-system",
	"kubesphere-controls-system",
	"kubesphere-monitoring-federated",
	"kubesphere-monitoring-system",
	"kubesphere-system",
}

var ErrAppNotFoundInChartRepo = errors.New("app not found in chart repo")
var ErrProviderNotFound = errors.New("provider not found")

// UpdateAppState update application status state.
func UpdateAppState(ctx context.Context, am *v1alpha1.ApplicationManager, state string) error {
	client, err := utils.GetClient()
	if err != nil {
		return err
	}
	return retry.RetryOnConflict(retry.DefaultRetry, func() error {
		app, err := client.AppV1alpha1().Applications().Get(ctx, am.Name, metav1.GetOptions{})
		if err != nil {
			if apierrors.IsNotFound(err) {
				// dev mode, try to find app in user-space
				apps, err := client.AppV1alpha1().Applications().List(ctx, metav1.ListOptions{})
				if err != nil {
					return err
				}

				for _, a := range apps.Items {
					if a.Spec.Name == am.Spec.AppName &&
						a.Spec.Owner == am.Spec.AppOwner &&
						a.Spec.Namespace == "user-space-"+a.Spec.Owner {
						app = &a

						break
					}
				}

			} else {
				return err
			}
		}
		now := metav1.Now()
		appCopy := app.DeepCopy()
		appCopy.Status.State = state
		appCopy.Status.StatusTime = &now
		appCopy.Status.UpdateTime = &now

		// set startedTime when app first become running
		if state == v1alpha1.AppRunning.String() && appCopy.Status.StartedTime.IsZero() {
			appCopy.Status.StartedTime = &now
			entranceStatues := make([]v1alpha1.EntranceStatus, 0, len(app.Spec.Entrances))
			for _, e := range app.Spec.Entrances {
				entranceStatues = append(entranceStatues, v1alpha1.EntranceStatus{
					Name:       e.Name,
					State:      v1alpha1.EntranceRunning,
					StatusTime: &now,
					Reason:     v1alpha1.EntranceRunning.String(),
				})
			}
			appCopy.Status.EntranceStatuses = entranceStatues
		}

		if appCopy.Name == "" {
			return nil
		}

		_, err = client.AppV1alpha1().Applications().Get(ctx, appCopy.Name, metav1.GetOptions{})
		if apierrors.IsNotFound(err) {
			return nil
		}

		_, err = client.AppV1alpha1().Applications().UpdateStatus(ctx, appCopy, metav1.UpdateOptions{})

		return err
	})

}

// UpdateAppMgrStatus update applicationmanager status, if filed in parameter status is empty that field will not be set.
func UpdateAppMgrStatus(name string, status v1alpha1.ApplicationManagerStatus) (*v1alpha1.ApplicationManager, error) {
	client, err := utils.GetClient()
	if err != nil {
		return nil, err
	}
	var appMgr *v1alpha1.ApplicationManager

	err = retry.RetryOnConflict(retry.DefaultRetry, func() error {
		appMgr, err = client.AppV1alpha1().ApplicationManagers().Get(context.TODO(), name, metav1.GetOptions{})
		if err != nil {
			return err
		}
		appMgrCopy := appMgr.DeepCopy()

		status.OpGeneration = appMgrCopy.Status.OpGeneration + 1
		status.OpRecords = appMgrCopy.Status.OpRecords

		if status.State == "" {
			status.State = appMgrCopy.Status.State
		}
		if status.Message == "" {
			status.Message = appMgrCopy.Status.Message
		}
		if status.Reason == "" {
			status.Reason = appMgrCopy.Status.Reason
		}
		payload := status.Payload
		if payload == nil {
			payload = make(map[string]string)
		}
		for k, v := range appMgrCopy.Status.Payload {
			if _, ok := payload[k]; !ok {
				payload[k] = v
			}
		}
		status.Payload = payload

		appMgrCopy.Status = status

		appMgr, err = client.AppV1alpha1().ApplicationManagers().Update(context.TODO(), appMgrCopy, metav1.UpdateOptions{})
		return err
	})

	return appMgr, err
}

// ApplyAppEnv applies the environment variable configuration of the app:
// if no existing AppEnv is found for the app, the AppEnv is created
// if existing AppEnv is found for the app, any new configured env is added to the AppEnv
func ApplyAppEnv(ctx context.Context, c client.Client, appConfig *appcfg.ApplicationConfig) (*sysv1alpha1.AppEnv, error) {
	if appConfig == nil {
		return nil, fmt.Errorf("app config is nil")
	}

	if len(appConfig.Envs) == 0 {
		klog.Infof("No envs found for app: %s owner: %s, skip app env apply", appConfig.AppName, appConfig.OwnerName)
		return nil, nil
	}

	var desiredAppEnvs []sysv1alpha1.AppEnvVar
	for _, env := range appConfig.Envs {
		if env.ValueFrom != nil && env.ValueFrom.EnvName == "" {
			env.ValueFrom = nil
		}
		// If ValueFrom is specified, set it up
		if env.ValueFrom != nil && env.ValueFrom.EnvName != "" {
			env.Value = ""
			env.Default = ""
			env.Type = ""
			env.Editable = false
			env.Options = nil
			env.RemoteOptions = ""
			env.Regex = ""
			env.ValueFrom.Status = constants.EnvRefStatusPending
		}

		desiredAppEnvs = append(desiredAppEnvs, env)
	}

	appEnv := new(sysv1alpha1.AppEnv)
	appEnv.SetName(FormatAppEnvName(appConfig.AppName, appConfig.OwnerName))
	appEnv.SetNamespace(appConfig.Namespace)
	if err := c.Get(ctx, client.ObjectKeyFromObject(appEnv), appEnv); err != nil {
		if !apierrors.IsNotFound(err) {
			return nil, err
		}
		klog.Infof("no existing app env found for app: %s owner: %s, create new app env", appConfig.AppName, appConfig.OwnerName)
		appNS := new(corev1.Namespace)
		appNS.SetName(appConfig.Namespace)
		err = c.Create(ctx, appNS)
		if err != nil && !apierrors.IsAlreadyExists(err) {
			return nil, err
		}
		newAppEnv := &sysv1alpha1.AppEnv{
			ObjectMeta: metav1.ObjectMeta{
				Name:      appEnv.GetName(),
				Namespace: appEnv.GetNamespace(),
			},
			AppName:  appConfig.AppName,
			AppOwner: appConfig.OwnerName,
			Envs:     desiredAppEnvs,
		}
		if err := c.Create(ctx, newAppEnv); err != nil {
			return nil, err
		}
		return newAppEnv, nil
	}

	original := appEnv.DeepCopy()
	updated := false

	if len(desiredAppEnvs) > 0 {
		existingEnvs := make(map[string]struct{}, len(appEnv.Envs))
		for _, e := range appEnv.Envs {
			existingEnvs[e.EnvName] = struct{}{}
		}
		for _, e := range desiredAppEnvs {
			if _, ok := existingEnvs[e.EnvName]; !ok {
				appEnv.Envs = append(appEnv.Envs, e)
				updated = true
			}
		}
	}

	if !updated {
		return appEnv, nil
	}

	if err := c.Patch(ctx, appEnv, client.MergeFrom(original)); err != nil {
		return nil, err
	}
	klog.Infof("patched app env for app: %s, owner: %s", appConfig.AppName, appConfig.OwnerName)
	return appEnv, nil
}

// GetDeployedReleaseVersion check whether app has been deployed and return release chart version
func GetDeployedReleaseVersion(actionConfig *action.Configuration, appName string) (string, int, error) {
	client := action.NewGet(actionConfig)
	release, err := client.Run(appName)
	if err != nil {
		return "", 0, err
	}
	return release.Chart.Metadata.Version, release.Version, nil
}

// CreateSysAppMgr create an applicationmanager for the system application.
func CreateSysAppMgr(app, owner string) error {
	client, err := utils.GetClient()
	if err != nil {
		return err
	}
	appNamespace, _ := utils.AppNamespace(app, owner, "user-space")
	now := metav1.Now()
	appMgr := &v1alpha1.ApplicationManager{
		ObjectMeta: metav1.ObjectMeta{
			Name: fmt.Sprintf("%s-%s", appNamespace, app),
		},
		Spec: v1alpha1.ApplicationManagerSpec{
			OpType:       v1alpha1.InstallOp,
			AppName:      app,
			RawAppName:   app,
			AppNamespace: appNamespace,
			AppOwner:     owner,
			Source:       "system",
			Type:         "app",
		},
	}

	a, err := client.AppV1alpha1().ApplicationManagers().Get(context.TODO(), appMgr.Name, metav1.GetOptions{})
	if err != nil && !apierrors.IsNotFound(err) {
		return err
	}
	if apierrors.IsNotFound(err) {
		a, err = client.AppV1alpha1().ApplicationManagers().Create(context.TODO(), appMgr, metav1.CreateOptions{})
		if err != nil && !apierrors.IsAlreadyExists(err) {
			return err
		}
	}

	appMgrCopy := a.DeepCopy()
	status := v1alpha1.ApplicationManagerStatus{
		OpType:       v1alpha1.InstallOp,
		State:        v1alpha1.Running,
		OpGeneration: int64(0),
		Message:      "sys app install completed",
		UpdateTime:   &now,
		StatusTime:   &now,
	}
	appMgrCopy.Status = status
	_, err = client.AppV1alpha1().ApplicationManagers().Update(context.TODO(), appMgrCopy, metav1.UpdateOptions{})
	return err
}

// GetAppMgrStatus returns status of an applicationmanager.
func GetAppMgrStatus(name string) (*v1alpha1.ApplicationManagerStatus, error) {
	client, err := utils.GetClient()
	if err != nil {
		return nil, err
	}
	appMgr, err := client.AppV1alpha1().ApplicationManagers().Get(context.TODO(), name, metav1.GetOptions{})
	if err != nil {
		return nil, err
	}

	return &appMgr.Status, nil
}

// FmtAppMgrName returns applicationmanager name for application.
func FmtAppMgrName(app, owner, ns string) (string, error) {
	namespace, err := utils.AppNamespace(app, owner, ns)
	if err != nil {
		return "", err
	}
	return fmt.Sprintf("%s-%s", namespace, app), nil
}

// TryConnect try to connect to a service with specified host and port.
func TryConnect(host string, port string) bool {
	timeout := time.Second
	conn, err := net.DialTimeout("tcp", net.JoinHostPort(host, port), timeout)
	if err != nil {
		klog.Errorf("Try to connect: %s:%s err=%v", host, port, err)
		return false
	}
	if conn != nil {
		defer conn.Close()
		return true
	}

	return false
}

// GetAppID returns appID for an application.
// for system app appID equals name, otherwise appID equals md5(name)[:8].
func GetAppID(name string) string {
	if userspace.IsSysApp(name) {
		return name
	}
	return utils.Md5String(name)[:8]
}

// GetPendingOrRunningTask returns pending and running state applicationmanager.
func GetPendingOrRunningTask(ctx context.Context) (ams []v1alpha1.ApplicationManager, err error) {
	ams = make([]v1alpha1.ApplicationManager, 0)
	client, err := utils.GetClient()
	if err != nil {
		return ams, err
	}
	list, err := client.AppV1alpha1().ApplicationManagers().List(ctx, metav1.ListOptions{})
	if err != nil {
		return ams, err
	}
	for _, am := range list.Items {
		if am.Status.State == v1alpha1.Pending || am.Status.State == v1alpha1.Installing ||
			am.Status.State == v1alpha1.Uninstalling || am.Status.State == v1alpha1.Upgrading {
			ams = append(ams, am)
		}
	}

	return ams, nil
}

// UpdateStatus update application state and applicationmanager state.
func UpdateStatus(appMgr *v1alpha1.ApplicationManager, state v1alpha1.ApplicationManagerState,
	opRecord *v1alpha1.OpRecord, message string) error {
	client, _ := utils.GetClient()
	var err error
	return retry.RetryOnConflict(retry.DefaultBackoff, func() error {
		appMgr, err = client.AppV1alpha1().ApplicationManagers().Get(context.TODO(), appMgr.Name, metav1.GetOptions{})
		if err != nil {
			return err
		}
		now := metav1.Now()
		appMgrCopy := appMgr.DeepCopy()
		appMgrCopy.Status.State = state
		appMgrCopy.Status.Message = message
		appMgrCopy.Status.StatusTime = &now
		appMgrCopy.Status.UpdateTime = &now
		if opRecord != nil {
			appMgrCopy.Status.OpRecords = append([]v1alpha1.OpRecord{*opRecord}, appMgr.Status.OpRecords...)
		}
		if len(appMgr.Status.OpRecords) > 20 {
			appMgrCopy.Status.OpRecords = appMgr.Status.OpRecords[:20:20]
		}
		//klog.Infof("utils: UpdateStatus: %v", appMgrCopy.Status.Conditions)

		_, err = client.AppV1alpha1().ApplicationManagers().Update(context.TODO(), appMgrCopy, metav1.UpdateOptions{})
		if err != nil {
			return err
		}
		//if len(appState) > 0 {
		//	err = UpdateAppState(context.TODO(), appMgr, appState.String())
		//	if err != nil {
		//		return err
		//	}
		//}
		return err
	})
}

func IsProtectedNamespace(namespace string) bool {
	for _, n := range protectedNamespace {
		if strings.HasPrefix(namespace, n) {
			return true
		}
	}
	return false
}

func IsForbidNamespace(namespace string) bool {
	for _, n := range forbidNamespace {
		if namespace == n {
			return true
		}
	}
	return false
}

// ACLProto If the ACL proto field is empty, it allows ICMPv4, ICMPv6, TCP, and UDP as per Tailscale behaviour
var ACLProto = sets.NewString("", "igmp", "ipv4", "ip-in-ip", "tcp", "egp", "igp", "udp", "gre", "esp", "ah", "sctp", "icmp")

func CheckTailScaleACLs(acls []v1alpha1.ACL) error {
	if len(acls) == 0 {
		return nil
	}
	var err error
	// fill default value fro ACL
	for i := range acls {
		acls[i].Action = "accept"
		acls[i].Src = []string{"*"}
	}
	for _, acl := range acls {
		err = parseProtocol(acl.Proto)
		if err != nil {
			return err
		}
		for _, dest := range acl.Dst {
			_, _, err = parseDestination(dest)
			if err != nil {
				return err
			}
		}
	}
	return nil
}

func parseProtocol(protocol string) error {
	if ACLProto.Has(protocol) {
		return nil
	}
	return fmt.Errorf("unsupported protocol: %v", protocol)
}

// parseDestination from
// https://github.com/juanfont/headscale/blob/770f3dcb9334adac650276dcec90cd980af53c6e/hscontrol/policy/acls.go#L475
func parseDestination(dest string) (string, string, error) {
	var tokens []string

	// Check if there is a IPv4/6:Port combination, IPv6 has more than
	// three ":".
	tokens = strings.Split(dest, ":")
	if len(tokens) < expectedTokenItems || len(tokens) > 3 {
		port := tokens[len(tokens)-1]

		maybeIPv6Str := strings.TrimSuffix(dest, ":"+port)

		filteredMaybeIPv6Str := maybeIPv6Str
		if strings.Contains(maybeIPv6Str, "/") {
			networkParts := strings.Split(maybeIPv6Str, "/")
			filteredMaybeIPv6Str = networkParts[0]
		}

		if maybeIPv6, err := netip.ParseAddr(filteredMaybeIPv6Str); err != nil && !maybeIPv6.Is6() {

			return "", "", fmt.Errorf(
				"failed to parse destination, tokens %v: %w",
				tokens,
				ErrInvalidPortFormat,
			)
		} else {
			tokens = []string{maybeIPv6Str, port}
		}
	}

	var alias string
	// We can have here stuff like:
	// git-server:*
	// 192.168.1.0/24:22
	// fd7a:115c:a1e0::2:22
	// fd7a:115c:a1e0::2/128:22
	// tag:montreal-webserver:80,443
	// tag:api-server:443
	// example-host-1:*
	if len(tokens) == expectedTokenItems {
		alias = tokens[0]
	} else {
		alias = fmt.Sprintf("%s:%s", tokens[0], tokens[1])
	}

	return alias, tokens[len(tokens)-1], nil
}

func TryToGetAppdataDirFromDeployment(ctx context.Context, namespace, name, owner string) (appdirs []string, err error) {
	userspaceNs := utils.UserspaceName(owner)
	config, err := ctrl.GetConfig()
	if err != nil {
		return
	}
	clientset, err := kubernetes.NewForConfig(config)
	if err != nil {
		return
	}
	sts, err := clientset.AppsV1().StatefulSets(userspaceNs).Get(ctx, "bfl", metav1.GetOptions{})
	if err != nil {
		return
	}
	appName := fmt.Sprintf("%s-%s", namespace, name)
	appCachePath := sts.GetAnnotations()["appcache_hostpath"]
	if len(appCachePath) == 0 {
		err = errors.New("empty appcache_hostpath")
		return
	}
	if !strings.HasSuffix(appCachePath, "/") {
		appCachePath += "/"
	}
	dClient, err := versioned.NewForConfig(config)
	if err != nil {
		return
	}
	appCRD, err := dClient.AppV1alpha1().Applications().Get(ctx, appName, metav1.GetOptions{})
	if err != nil {
		return
	}
	deploymentName := appCRD.Spec.DeploymentName
	deployment, err := clientset.AppsV1().Deployments(namespace).
		Get(context.Background(), deploymentName, metav1.GetOptions{})
	if err != nil {
		if apierrors.IsNotFound(err) {
			return tryToGetAppdataDirFromSts(ctx, namespace, deploymentName, appCachePath)
		}
		return
	}
	appDirSet := sets.NewString()
	for _, v := range deployment.Spec.Template.Spec.Volumes {
		if v.HostPath != nil && strings.HasPrefix(v.HostPath.Path, appCachePath) && len(v.HostPath.Path) > len(appCachePath) {
			appDir := GetFirstSubDir(v.HostPath.Path, appCachePath)
			if appDir != "" {
				if appDirSet.Has(appDir) {
					continue
				}
				appdirs = append(appdirs, appDir)
				appDirSet.Insert(appDir)
			}
		}
	}
	return appdirs, nil
}

func tryToGetAppdataDirFromSts(ctx context.Context, namespace, stsName, baseDir string) (appdirs []string, err error) {
	config, err := ctrl.GetConfig()
	if err != nil {
		return
	}
	clientset, err := kubernetes.NewForConfig(config)
	if err != nil {
		return
	}

	sts, err := clientset.AppsV1().StatefulSets(namespace).
		Get(ctx, stsName, metav1.GetOptions{})
	if err != nil {
		return
	}
	appDirSet := sets.NewString()
	for _, v := range sts.Spec.Template.Spec.Volumes {
		if v.HostPath != nil && strings.HasPrefix(v.HostPath.Path, baseDir) && len(v.HostPath.Path) > len(baseDir) {
			appDir := GetFirstSubDir(v.HostPath.Path, baseDir)
			if appDir != "" {
				if appDirSet.Has(appDir) {
					continue
				}
				appdirs = append(appdirs, appDir)
				appDirSet.Insert(appDir)
			}
		}
	}
	return appdirs, nil
}

func GetFirstSubDir(fullPath, basePath string) string {
	if basePath == "" {
		return ""
	}
	if !strings.HasSuffix(basePath, "/") {
		basePath += "/"
	}

	if !strings.HasPrefix(fullPath, basePath) {
		return ""
	}

	relPath := strings.TrimPrefix(fullPath, basePath)
	if relPath == "" {
		return ""
	}

	parts := strings.Split(relPath, "/")
	if len(parts) == 0 {
		return ""
	}
	return parts[0]
}

type ConfigOptions struct {
	App          string
	RepoURL      string
	Owner        string
	Version      string
	Token        string
	Admin        string
	MarketSource string
	IsAdmin      bool
	RawAppName   string
	SelectedGpu  string
}

// GetAppConfig get app installation configuration from app store
func GetAppConfig(ctx context.Context, options *ConfigOptions) (*appcfg.ApplicationConfig, string, error) {
	if options == nil {
		return nil, "", fmt.Errorf("options parameter is nil")
	}
	if options.RepoURL == "" {
		return nil, "", fmt.Errorf("url info is empty, repo [%s]", options.RepoURL)
	}

	var (
		appcfg    *appcfg.ApplicationConfig
		chartPath string
		err       error
	)

	appcfg, chartPath, err = getAppConfigFromRepo(ctx, options)
	if err != nil {
		return nil, chartPath, err
	}

	// set appcfg.Namespace to specified namespace by OlaresManifests.Spec
	var namespace string
	if appcfg.Namespace != "" {
		namespace, _ = utils.AppNamespace(options.App, options.Owner, appcfg.Namespace)
	} else {
		namespace = fmt.Sprintf("%s-%s", options.App, options.Owner)
	}

	if appcfg.IsMiddleware() {
		namespace = fmt.Sprintf("%s-%s", appcfg.AppName, "middleware")
		klog.Infof("appcfg.IsMiddlewarew, set namespace=%s", namespace)

	}

	appcfg.Namespace = namespace
	appcfg.OwnerName = options.Owner
	appcfg.RepoURL = options.RepoURL
	return appcfg, chartPath, nil
}

func GetApiVersionFromAppConfig(ctx context.Context, options *ConfigOptions) (appcfg.APIVersion, *appcfg.ApplicationConfig, error) {

	cfg, _, err := GetAppConfig(ctx, options)
	if err != nil {
		return "", nil, fmt.Errorf("failed to get app config: %w", err)
	}

	// default version is v1
	if cfg.APIVersion == "" {
		return appcfg.V1, cfg, nil
	}

	return cfg.APIVersion, cfg, nil
}

func getAppConfigFromRepo(ctx context.Context, options *ConfigOptions) (*appcfg.ApplicationConfig, string, error) {
	chartPath, err := GetIndexAndDownloadChart(ctx, options)
	if err != nil {
		klog.Errorf("failed to get index and download chart app: %s, rawAppName: %s, %v", options.App, options.RawAppName, err)
		return nil, chartPath, err
	}
	return getAppConfigFromConfigurationFile(options, chartPath)
}

func toApplicationConfig(app, chart, rawAppName, selectedGpu string, cfg *appcfg.AppConfiguration) (*appcfg.ApplicationConfig, string, error) {
	var permission []appcfg.AppPermission
	if cfg.Permission.AppData {
		permission = append(permission, appcfg.AppDataRW)
	}
	if cfg.Permission.AppCache {
		permission = append(permission, appcfg.AppCacheRW)
	}
	if len(cfg.Permission.UserData) > 0 {
		permission = append(permission, appcfg.UserDataRW)
	}

	if len(cfg.Permission.Provider) > 0 {
		var perm []appcfg.ProviderPermission
		for _, s := range cfg.Permission.Provider {
			perm = append(perm, appcfg.ProviderPermission(s))
		}
		permission = append(permission, perm)
	}

	valuePtr := func(v resource.Quantity, err error) (*resource.Quantity, error) {
		if errors.Is(err, resource.ErrFormatWrong) {
			return nil, nil
		}

		return &v, nil
	}

	mem, err := valuePtr(resource.ParseQuantity(cfg.Spec.RequiredMemory))
	if err != nil {
		return nil, chart, err
	}

	disk, err := valuePtr(resource.ParseQuantity(cfg.Spec.RequiredDisk))
	if err != nil {
		return nil, chart, err
	}

	cpu, err := valuePtr(resource.ParseQuantity(cfg.Spec.RequiredCPU))
	if err != nil {
		return nil, chart, err
	}

	gpu, err := valuePtr(resource.ParseQuantity(cfg.Spec.RequiredGPU))
	if err != nil {
		return nil, chart, err
	}

	// if requiredGPU is set and no supportedGpu specified, set suppertedGpu to ["nvidia"]
	if cfg.Spec.RequiredGPU != "" && len(cfg.Spec.SupportedGpu) == 0 {
		cfg.Spec.SupportedGpu = []interface{}{"nvidia"}
	}

	// try to get selected GPU type special resource requirement
	if selectedGpu != "" {
		found := false
		for _, supportedGpu := range cfg.Spec.SupportedGpu {
			if str, ok := supportedGpu.(string); ok && str == selectedGpu {
				found = true
				break
			}

			if supportedGpuResourceMap, ok := supportedGpu.(map[string]interface{}); ok {
				if resourceRequirement, ok := supportedGpuResourceMap[selectedGpu].(map[string]interface{}); ok {
					found = true
					var specialResource appcfg.SpecialResource
					err := mapstructure.Decode(resourceRequirement, &specialResource)
					if err != nil {
						return nil, chart, fmt.Errorf("failed to decode special resource for selected GPU type %s: %v", selectedGpu, err)
					}

					for _, resSetter := range []struct {
						v **resource.Quantity
						s *string
					}{
						{v: &mem, s: specialResource.RequiredMemory},
						{v: &disk, s: specialResource.RequiredDisk},
						{v: &cpu, s: specialResource.RequiredCPU},
						{v: &gpu, s: specialResource.RequiredGPU},
					} {

						if resSetter.s != nil && *resSetter.s != "" {
							*resSetter.v, err = valuePtr(resource.ParseQuantity(*resSetter.s))
							if err != nil {
								return nil, chart, fmt.Errorf("failed to parse special resource quantity %s: %v", *resSetter.s, err)
							}
						}
					}

					break
				} // end if selected gpu's resource requirement found
			} // end if supportedGpu is map
		} // end for supportedGpu

		if !found {
			return nil, chart, fmt.Errorf("selected GPU type %s is not supported", selectedGpu)
		}
	}

	// transform from Policy to AppPolicy
	var policies []appcfg.AppPolicy
	for _, p := range cfg.Options.Policies {
		d, _ := time.ParseDuration(p.Duration)
		policies = append(policies, appcfg.AppPolicy{
			EntranceName: p.EntranceName,
			URIRegex:     p.URIRegex,
			Level:        p.Level,
			OneTime:      p.OneTime,
			Duration:     d,
		})
	}

	// check dependencies version format
	for _, dep := range cfg.Options.Dependencies {
		if err = checkVersionFormat(dep.Version); err != nil {
			return nil, chart, err
		}
	}

	if cfg.Middleware != nil && cfg.Middleware.Redis != nil {
		if len(cfg.Middleware.Redis.Namespace) == 0 {
			return nil, chart, errors.New("middleware of Redis namespace can not be empty")
		}
	}
	var appid string
	if userspace.IsSysApp(app) {
		appid = app
	} else {
		appid = utils.Md5String(app)[:8]
	}

	if appcfg.APIVersion(cfg.APIVersion) == appcfg.V2 {
		// V2: additional validation for app config
	}
	podSelectors := make([]metav1.LabelSelector, 0)
	for _, p := range cfg.Permission.Provider {
		podSelectors = append(podSelectors, p.PodSelectors...)
	}

	return &appcfg.ApplicationConfig{
		AppID:          appid,
		APIVersion:     appcfg.APIVersion(cfg.APIVersion),
		CfgFileVersion: cfg.ConfigVersion,
		AppName:        app,
		RawAppName:     rawAppName,
		Title:          cfg.Metadata.Title,
		Version:        cfg.Metadata.Version,
		Target:         cfg.Metadata.Target,
		ChartsName:     chart,
		Entrances:      cfg.Entrances,
		Ports:          cfg.Ports,
		TailScale:      cfg.TailScale,
		Icon:           cfg.Metadata.Icon,
		Permission:     permission,
		Requirement: appcfg.AppRequirement{
			Memory: mem,
			CPU:    cpu,
			Disk:   disk,
			GPU:    gpu,
		},
		Policies:             policies,
		Middleware:           cfg.Middleware,
		ResetCookieEnabled:   cfg.Options.ResetCookie.Enabled,
		Dependencies:         cfg.Options.Dependencies,
		Conflicts:            cfg.Options.Conflicts,
		AppScope:             cfg.Options.AppScope,
		WsConfig:             cfg.Options.WsConfig,
		Upload:               cfg.Options.Upload,
		OnlyAdmin:            cfg.Spec.OnlyAdmin,
		Namespace:            cfg.Spec.Namespace,
		MobileSupported:      cfg.Options.MobileSupported,
		OIDC:                 cfg.Options.OIDC,
		ApiTimeout:           cfg.Options.ApiTimeout,
		RunAsUser:            cfg.Spec.RunAsUser,
		AllowedOutboundPorts: cfg.Options.AllowedOutboundPorts,
		RequiredGPU:          cfg.Spec.RequiredGPU,
		PodGPUConsumePolicy:  cfg.Spec.PodGPUConsumePolicy,
		Internal:             cfg.Spec.RunAsInternal,
		SubCharts:            cfg.Spec.SubCharts,
		ServiceAccountName:   cfg.Permission.ServiceAccount,
		Provider:             cfg.Provider,
		Type:                 cfg.ConfigType,
		Envs:                 cfg.Envs,
		Images:               cfg.Options.Images,
		AllowMultipleInstall: cfg.Options.AllowMultipleInstall,
		PodsSelectors:        podSelectors,
		HardwareRequirement:  cfg.Spec.Hardware,
		SharedEntrances:      cfg.SharedEntrances,
		SelectedGpuType:      selectedGpu,
	}, chart, nil
}

func getAppConfigFromConfigurationFile(opt *ConfigOptions, chartPath string) (*appcfg.ApplicationConfig, string, error) {
	data, err := utils.RenderManifest(filepath.Join(chartPath, AppCfgFileName), opt.Owner, opt.Admin, opt.IsAdmin)
	if err != nil {
		return nil, chartPath, err
	}
	var cfg appcfg.AppConfiguration
	if err := yaml.Unmarshal([]byte(data), &cfg); err != nil {
		return nil, chartPath, err
	}

	return toApplicationConfig(opt.App, chartPath, opt.RawAppName, opt.SelectedGpu, &cfg)
}

func checkVersionFormat(constraint string) error {
	_, err := semver.NewConstraint(constraint)
	if err != nil {
		return err
	}
	return nil
}

// GetIndexAndDownloadChart download a chart and returns download chart path.
func GetIndexAndDownloadChart(ctx context.Context, options *ConfigOptions) (string, error) {
	client := resty.New().SetTimeout(10*time.Second).
		SetAuthToken(options.Token).
		SetHeader(constants.MarketUser, options.Owner).
		SetHeader(constants.MarketSource, options.MarketSource)
	indexFileURL := options.RepoURL
	if options.RepoURL[len(options.RepoURL)-1] != '/' {
		indexFileURL += "/"
	}
	klog.Infof("GetIndexAndDownloadChart: user: %v, source: %v", options.Owner, options.MarketSource)

	indexFileURL += "static-index.yaml"
	resp, err := client.R().Get(indexFileURL)
	if err != nil {
		return "", err
	}

	if resp.StatusCode() >= 400 {
		return "", fmt.Errorf("get app config from repo returns unexpected status code, %d", resp.StatusCode())
	}

	index, err := loadIndex(resp.Body())
	if err != nil {
		klog.Errorf("Failed to load chart index err=%v", err)
		return "", err
	}

	klog.Infof("Success to load chart index from %s", indexFileURL)
	// get specified version chart, if version is empty return the chart with the latest stable version
	chartVersion, err := index.Get(options.RawAppName, options.Version)

	if err != nil {
		klog.Errorf("Failed to get chart version err=%v app=%s rawApp=%s version=%s", err, options.App, options.RawAppName, options.Version)
		if errors.Is(err, repo.ErrNoChartName) {
			return "", ErrAppNotFoundInChartRepo
		}
		return "", err
	}

	klog.Infof("Success to find app chart from index app=%s rawApp=%s version=%s", options.App, options.RawAppName, options.Version)
	chartURL, err := repo.ResolveReferenceURL(options.RepoURL, chartVersion.URLs[0])
	if err != nil {
		return "", err
	}

	url, err := url.Parse(chartURL)
	if err != nil {
		return "", err
	}

	// assume the chart path is app name
	chartPath := appcfg.ChartsPath + "/" + options.RawAppName
	if files.IsExist(chartPath) {
		if err := files.RemoveAll(chartPath); err != nil {
			return "", err
		}
	}
	_, err = downloadAndUnpack(ctx, url, options.Token, options.Owner, options.MarketSource)
	if err != nil {
		return "", err
	}
	return chartPath, nil
}

func downloadAndUnpack(ctx context.Context, tgz *url.URL, token, owner, marketSource string) (string, error) {
	dst := appcfg.ChartsPath
	g := new(getter.HttpGetter)
	g.Header = make(http.Header)
	g.Header.Set("Authorization", "Bearer "+token)
	g.Header.Set(constants.MarketUser, owner)
	g.Header.Set(constants.MarketSource, marketSource)

	downloader := &getter.Client{
		Ctx:       ctx,
		Dst:       dst,
		Src:       tgz.String(),
		Mode:      getter.ClientModeDir,
		Detectors: getter.Detectors,
		Getters: map[string]getter.Getter{
			"http": g,
			"file": new(getter.FileGetter),
		},
	}

	//download the files
	if err := downloader.Get(); err != nil {
		klog.Errorf("Failed to get path=%s err=%v", downloader.Src, err)
		return "", err
	}

	return dst, nil
}

func loadIndex(data []byte) (*repo.IndexFile, error) {
	i := &repo.IndexFile{}

	if len(data) == 0 {
		return i, repo.ErrEmptyIndexYaml
	}

	if err := yaml.UnmarshalStrict(data, i); err != nil {
		return i, err
	}

	for name, cvs := range i.Entries {
		for idx := len(cvs) - 1; idx >= 0; idx-- {
			if cvs[idx].APIVersion == "" {
				cvs[idx].APIVersion = chart.APIVersionV1
			}
			if err := cvs[idx].Validate(); err != nil {
				klog.Infof("Skipping loading invalid entry for chart name=%q version=%q err=%v", name, cvs[idx].Version, err)
				cvs = append(cvs[:idx], cvs[idx+1:]...)
			}
		}
	}
	i.SortEntries()
	if i.APIVersion == "" {
		return i, repo.ErrNoAPIVersion
	}
	return i, nil
}

func UpdateAppMgrState(ctx context.Context, name string, state v1alpha1.ApplicationManagerState) error {
	client, err := utils.GetClient()
	if err != nil {
		return err
	}
	var am *v1alpha1.ApplicationManager
	err = retry.RetryOnConflict(retry.DefaultRetry, func() error {
		am, err = client.AppV1alpha1().ApplicationManagers().Get(ctx, name, metav1.GetOptions{})
		if err != nil {
			return err
		}
		now := metav1.Now()
		amCopy := am.DeepCopy()
		amCopy.Status.State = state
		amCopy.Status.UpdateTime = &now
		amCopy.Status.StatusTime = &now
		amCopy.Status.OpGeneration += 1
		_, err = client.AppV1alpha1().ApplicationManagers().Update(ctx, amCopy, metav1.UpdateOptions{})
		return err
	})
	return err
}

func formatCacheDir(s string) string {
	if s == "" {
		return s
	}
	if !strings.HasPrefix(s, "/") {
		s = "/" + s
	}
	if !strings.HasSuffix(s, "/") {
		s = s + "/"
	}
	return s
}

func FormatCacheDirs(dirs []string) []string {
	ret := make([]string, 0, len(dirs))
	for _, dir := range dirs {
		ret = append(ret, formatCacheDir(dir))
	}
	return ret
}

func FormatAppEnvName(appName, appowner string) string {
	return fmt.Sprintf("%s-%s", appName, appowner)
}

func IsClonedApp(appName, rawAppName string) bool {
	if rawAppName != "" && rawAppName != appName {
		return true
	}
	return false
}

func AppTitle(config string) string {
	cfg := unmarshalApplicationConfig(config)
	if cfg == nil {
		return ""
	}
	return cfg.Title
}
func AppIcon(config string) string {
	cfg := unmarshalApplicationConfig(config)
	if cfg == nil {
		return ""
	}
	return cfg.Icon
}

func unmarshalApplicationConfig(config string) *appcfg.ApplicationConfig {
	var cfg appcfg.ApplicationConfig
	err := json.Unmarshal([]byte(config), &cfg)
	if err != nil {
		return nil
	}
	return &cfg
}

func GetRawAppName(AppName, rawAppName string) string {
	if rawAppName == "" {
		return AppName
	}
	return rawAppName

}

type portKey struct {
	port     int32
	protocol string
}

func genPort(protos []string) (int32, error) {
	exposePort := int32(rand.IntnRange(46800, 50000))
	for _, proto := range protos {
		if !utils.IsPortAvailable(proto, int(exposePort)) {
			return 0, fmt.Errorf("failed to allocate an available port after 5 attempts")
		}
	}
	return exposePort, nil
}

// SetExposePorts sets expose ports for app config.
func SetExposePorts(ctx context.Context, appConfig *appcfg.ApplicationConfig, prevPortsMap map[string]int32) error {
	existPorts := make(map[portKey]struct{})
	client, err := utils.GetClient()
	if err != nil {
		return err
	}
	apps, err := client.AppV1alpha1().Applications().List(ctx, metav1.ListOptions{})
	if err != nil {
		return err
	}

	for _, app := range apps.Items {
		for _, p := range app.Spec.Ports {
			protos := []string{p.Protocol}
			if p.Protocol == "" {
				protos = []string{"tcp", "udp"}
			}
			for _, proto := range protos {
				key := portKey{
					port:     p.ExposePort,
					protocol: proto,
				}
				existPorts[key] = struct{}{}
			}
		}
	}
	klog.Infof("existPorts: %v", existPorts)

	for i := range appConfig.Ports {
		port := &appConfig.Ports[i]

		// For upgrade: if port with same name exists in prevPortsMap, preserve its ExposePort
		if prevPortsMap != nil && port.Name != "" {
			if prevExposePort, exists := prevPortsMap[port.Name]; exists && prevExposePort != 0 {
				klog.Infof("preserving ExposePort %d for port %s from previous config", prevExposePort, port.Name)
				port.ExposePort = prevExposePort
				continue
			}
		}

		if port.ExposePort == 0 {
			var exposePort int32
			protos := []string{port.Protocol}
			if port.Protocol == "" {
				protos = []string{"tcp", "udp"}
			}

			for i := 0; i < 5; i++ {
				exposePort, err = genPort(protos)
				if err != nil {
					continue
				}
				for _, proto := range protos {
					key := portKey{port: exposePort, protocol: proto}
					if _, ok := existPorts[key]; !ok && err == nil {
						break
					}
				}
			}
			for _, proto := range protos {
				key := portKey{port: exposePort, protocol: proto}
				if _, ok := existPorts[key]; ok || err != nil {
					return fmt.Errorf("%d port is not available", key.port)
				}
				existPorts[key] = struct{}{}
				port.ExposePort = exposePort
			}
		}
	}

	// add exposePort to tailscale acls
	for i := range appConfig.Ports {
		if appConfig.Ports[i].AddToTailscaleAcl {
			appConfig.TailScale.ACLs = append(appConfig.TailScale.ACLs, v1alpha1.ACL{
				Action: "accept",
				Src:    []string{"*"},
				Proto:  appConfig.Ports[i].Protocol,
				Dst:    []string{fmt.Sprintf("*:%d", appConfig.Ports[i].ExposePort)},
			})
		}
	}
	klog.Infof("appConfig.TailScale: %v", appConfig.TailScale)
	return nil
}

// BuildPrevPortsMap builds a map of port name -> expose port from previous config.
func BuildPrevPortsMap(prevConfig *appcfg.ApplicationConfig) map[string]int32 {
	if prevConfig == nil {
		return nil
	}
	m := make(map[string]int32)
	for _, p := range prevConfig.Ports {
		if p.Name != "" && p.ExposePort != 0 {
			m[p.Name] = p.ExposePort
		}
	}
	return m
}
