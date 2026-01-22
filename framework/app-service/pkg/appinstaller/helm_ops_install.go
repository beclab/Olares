package appinstaller

import (
	"context"
	"encoding/json"
	"errors"
	"flag"
	"fmt"
	"net/http/httputil"
	"strconv"
	"time"

	appv1alpha1 "github.com/beclab/Olares/framework/app-service/api/app.bytetrade.io/v1alpha1"
	"github.com/beclab/Olares/framework/app-service/pkg/appcfg"
	"github.com/beclab/Olares/framework/app-service/pkg/client/clientset"
	"github.com/beclab/Olares/framework/app-service/pkg/constants"
	"github.com/beclab/Olares/framework/app-service/pkg/errcode"
	"github.com/beclab/Olares/framework/app-service/pkg/generated/clientset/versioned"
	"github.com/beclab/Olares/framework/app-service/pkg/helm"
	"github.com/beclab/Olares/framework/app-service/pkg/kubesphere"
	"github.com/beclab/Olares/framework/app-service/pkg/tapr"
	"github.com/beclab/Olares/framework/app-service/pkg/utils"
	apputils "github.com/beclab/Olares/framework/app-service/pkg/utils/app"

	"github.com/emicklei/go-restful/v3"
	"github.com/go-resty/resty/v2"
	"helm.sh/helm/v3/pkg/action"
	"helm.sh/helm/v3/pkg/cli"
	helmrelease "helm.sh/helm/v3/pkg/release"
	"helm.sh/helm/v3/pkg/storage/driver"
	corev1 "k8s.io/api/core/v1"
	apierrors "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"k8s.io/apimachinery/pkg/types"
	"k8s.io/client-go/dynamic"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/rest"
	"k8s.io/klog/v2"
)

var (
	systemServerHost = ""
	middlewareTypes  = []string{
		tapr.TypePostgreSQL.String(),
		tapr.TypeMongoDB.String(),
		tapr.TypeRedis.String(),
		tapr.TypeNats.String(),
		tapr.TypeMinio.String(),
		tapr.TypeRabbitMQ.String(),
		tapr.TypeElasticsearch.String(),
		tapr.TypeMariaDB.String(),
		tapr.TypeMySQL.String(),
		tapr.TypeClickHouse.String(),
	}
)

func init() {
	flag.StringVar(&systemServerHost, "system-server", "",
		"user's system-server host")
}

// HelmOpsInterface is an interface that defines operations related to helm chart.
type HelmOpsInterface interface {
	// Uninstall is the action for uninstall a release.
	Uninstall() error
	// Install is the action for install a release.
	Install() error
	// Upgrade is the action for upgrade a release.
	Upgrade() error
	// ApplyEnv upgrades only environment variables, reusing existing values.
	ApplyEnv() error
	// RollBack is the action for rollback a release.
	RollBack() error

	WaitForLaunch() (bool, error)

	UninstallAll() error
}

// Opt options for helm ops.
type Opt struct {
	Source       string
	MarketSource string
}

var _ HelmOpsInterface = &HelmOps{}

// HelmOps implements HelmOpsInterface.
type HelmOps struct {
	ctx          context.Context
	kubeConfig   *rest.Config
	actionConfig *action.Configuration
	app          *appcfg.ApplicationConfig
	settings     *cli.EnvSettings
	token        string
	//client       *kubernetes.Clientset
	//dyClient dynamic.Interface
	client  *clientset.ClientSet
	options Opt
}

func (h *HelmOps) install(values map[string]interface{}) error {
	_, err := h.status()
	if err == nil {
		return driver.ErrReleaseExists
	}
	if errors.Is(err, driver.ErrReleaseNotFound) {
		return helm.InstallCharts(h.ctx, h.actionConfig, h.settings, h.app.AppName, h.app.ChartsName, h.app.RepoURL, h.app.Namespace, values)
	}
	return err
}

// NewHelmOps constructs a new helmOps.
func NewHelmOps(ctx context.Context, kubeConfig *rest.Config, app *appcfg.ApplicationConfig, token string, options Opt) (HelmOpsInterface, error) {
	actionConfig, settings, err := helm.InitConfig(kubeConfig, app.Namespace)
	if err != nil {
		return nil, err
	}

	client, err := clientset.New(kubeConfig)
	if err != nil {
		return nil, err
	}
	ops := &HelmOps{
		ctx:          ctx,
		kubeConfig:   kubeConfig,
		app:          app,
		actionConfig: actionConfig,
		settings:     settings,
		token:        token,
		client:       client,
		options:      options,
	}
	return ops, nil
}

// AddApplicationLabelsToDeployment add application label to deployment or statefulset
func (h *HelmOps) AddApplicationLabelsToDeployment() error {
	k8s, err := kubernetes.NewForConfig(h.kubeConfig)
	if err != nil {
		klog.Error("create kubernetes client error, ", err)
		return err
	}

	// add namespace to workspace
	patch := "{\"metadata\": {\"labels\":{\"kubesphere.io/workspace\":\"system-workspace\"}}}"
	_, err = k8s.CoreV1().Namespaces().Patch(h.ctx, h.app.Namespace,
		types.MergePatchType, []byte(patch), metav1.PatchOptions{})
	if err != nil {
		klog.Errorf("patch namespace %s error %v", h.app.Namespace, err)
		return err
	}
	if h.app.Type == appv1alpha1.Middleware.String() {
		err = h.tryToAddApplicationLabelsToCluster()
		if err != nil {
			return err
		}
	}

	services := ToEntrancesLabel(h.app.Entrances)
	ports := ToAppTCPUDPPorts(h.app.Ports)

	tailScale := ToTailScale(h.app.TailScale)

	patchData := map[string]interface{}{
		"metadata": map[string]interface{}{
			"labels": map[string]string{
				constants.ApplicationNameLabel:       h.app.AppName,
				constants.ApplicationRawAppNameLabel: h.app.RawAppName,
				constants.ApplicationOwnerLabel:      h.app.OwnerName,
				constants.ApplicationTargetLabel:     h.app.Target,
				constants.ApplicationRunAsUserLabel:  strconv.FormatBool(h.app.RunAsUser),
				constants.ApplicationMiddlewareLabel: func() string {
					if h.app.Type == appv1alpha1.Middleware.String() {
						return "true"
					}
					return "false"
				}(),
			},
			"annotations": map[string]string{
				constants.ApplicationIconLabel:    h.app.Icon,
				constants.ApplicationTitleLabel:   h.app.Title,
				constants.ApplicationVersionLabel: h.app.Version,
				constants.ApplicationEntrancesKey: services,
				constants.ApplicationPortsKey:     ports,
				constants.ApplicationSourceLabel:  h.options.Source,
				constants.ApplicationTailScaleKey: tailScale,
				constants.ApplicationRequiredGPU:  h.app.RequiredGPU,
			},
		},
	}

	patchByte, err := json.Marshal(patchData)
	if err != nil {
		klog.Errorf("Failed to marshal patch data %v", err)
		return err
	}

	patch = string(patchByte)

	// TODO: add ownerReferences of user
	deployment, err := k8s.AppsV1().Deployments(h.app.Namespace).Get(h.ctx, h.app.AppName, metav1.GetOptions{})
	if err != nil {
		if apierrors.IsNotFound(err) {
			return h.tryToAddApplicationLabelsToStatefulSet(k8s, patch)
		}

		klog.Errorf("Failed to get deployment %s in namespace %s: %v", h.app.AppName, h.app.Namespace, err)
		return err
	}

	_, err = k8s.AppsV1().Deployments(h.app.Namespace).Patch(h.ctx,
		deployment.Name,
		types.MergePatchType,
		[]byte(patch),
		metav1.PatchOptions{})

	if err != nil {
		klog.Errorf("Failed to patch deployment %s in namespace %s: %v", h.app.AppName, h.app.Namespace, err)
		return err
	}
	return nil
}

func (h *HelmOps) tryToAddApplicationLabelsToCluster() error {
	// try to get kubeblocks cluster
	gvr := schema.GroupVersionResource{
		Group:    "apps.kubeblocks.io",
		Version:  "v1",
		Resource: "clusters",
	}

	patchData := map[string]interface{}{
		"metadata": map[string]interface{}{
			"labels": map[string]string{
				constants.ApplicationNameLabel:   h.app.AppName,
				constants.ApplicationOwnerLabel:  h.app.OwnerName,
				constants.ApplicationTargetLabel: h.app.Target,
			},
			"annotations": map[string]string{
				constants.ApplicationIconLabel:    h.app.Icon,
				constants.ApplicationTitleLabel:   h.app.Title,
				constants.ApplicationVersionLabel: h.app.Version,
				constants.ApplicationSourceLabel:  h.options.Source,
			},
		},
	}

	patchBytes, err := json.Marshal(patchData)
	if err != nil {
		klog.Errorf("Failed to marshal patch data for cluster %s in namespace %s: %v", h.app.AppName, h.app.Namespace, err)
		return err
	}

	dyClient, err := dynamic.NewForConfig(h.kubeConfig)
	if err != nil {
		klog.Errorf("Failed to create dynamic client: %v", err)
		return err
	}

	// check whether the cluster exists first
	_, err = dyClient.Resource(gvr).Namespace(h.app.Namespace).Get(h.ctx, h.app.AppName, metav1.GetOptions{})
	if err != nil {
		if apierrors.IsNotFound(err) {
			return nil
		}
		klog.Errorf("Failed to get cluster %s in namespace %s: %v", h.app.AppName, h.app.Namespace, err)
		return err
	}

	_, err = dyClient.Resource(gvr).Namespace(h.app.Namespace).Patch(h.ctx, h.app.AppName, types.MergePatchType, patchBytes, metav1.PatchOptions{})
	if err != nil {
		klog.Errorf("Failed to patch cluster %s in namespace %s: %v", h.app.AppName, h.app.Namespace, err)
		return err
	}

	return nil
}

func (h *HelmOps) tryToAddApplicationLabelsToStatefulSet(k8s *kubernetes.Clientset, patch string) error {
	statefulSet, err := k8s.AppsV1().StatefulSets(h.app.Namespace).Get(h.ctx, h.app.AppName, metav1.GetOptions{})
	if err != nil {
		if apierrors.IsNotFound(err) {
			return nil
		}

		klog.Errorf("Failed to get statefulset %s in namespace %s: %v", h.app.AppName, h.app.Namespace, err)
		return err
	}

	_, err = k8s.AppsV1().StatefulSets(h.app.Namespace).Patch(h.ctx,
		statefulSet.Name,
		types.MergePatchType,
		[]byte(patch),
		metav1.PatchOptions{})

	if err != nil {
		klog.Errorf("Failed to patch statefulset %s in namespace %s: %v", h.app.AppName, h.app.Namespace, err)
	}

	return err
}

func (h *HelmOps) status() (*helmrelease.Release, error) {
	statusClient := action.NewStatus(h.actionConfig)
	status, err := statusClient.Run(h.app.AppName)
	if err != nil {
		klog.Errorf("Failed to get status for release %s: %v", h.app.AppName, err)
		return nil, err
	}
	return status, nil
}

func (h *HelmOps) AddLabelToNamespaceForDependClusterApp() error {
	k8s, err := kubernetes.NewForConfig(h.kubeConfig)
	if err != nil {
		klog.Errorf("Failed to create kubernetes client: %v", err)
		return err
	}

	labels := map[string]string{
		constants.ApplicationClusterDep: h.app.AppName,
	}
	patchData := map[string]interface{}{"metadata": map[string]map[string]string{"labels": labels}}
	patchBytes, _ := json.Marshal(patchData)
	_, err = k8s.CoreV1().Namespaces().Patch(h.ctx, h.app.Namespace,
		types.MergePatchType, patchBytes, metav1.PatchOptions{})
	if err != nil {
		klog.Errorf("Failed to patch namespace %s with labels %v: %v", h.app.Namespace, labels, err)
		return err
	}
	return nil
}

func (h *HelmOps) userZone() (string, error) {
	return kubesphere.GetUserZone(h.ctx, h.app.OwnerName)
}

func (h *HelmOps) registerAppPerm(sa *string, ownerName string, perm []appcfg.PermissionCfg) (*appcfg.RegisterResp, error) {
	requires := make([]appcfg.PermissionRequire, 0, len(perm))
	for _, p := range perm {
		requires = append(requires, appcfg.PermissionRequire{
			ProviderAppName:   p.AppName,
			ProviderNamespace: p.GetNamespace(ownerName),
			ServiceAccount:    sa,
			ProviderName:      p.ProviderName,
			ProviderDomain:    p.Domain,
		})
	}
	register := appcfg.PermissionRegister{
		App:   h.app.AppName,
		AppID: h.app.AppID,
		Perm:  requires,
	}

	url := fmt.Sprintf("http://%s/permission/v2alpha1/register", h.systemServerHost())
	client := resty.New()

	body, err := json.Marshal(register)
	if err != nil {
		klog.Errorf("Failed to marshal register request body: %v", err)
		return nil, err
	}

	klog.Infof("Sending app register request with body=%s url=%s", utils.PrettyJSON(string(body)), url)

	resp, err := client.SetTimeout(2*time.Second).R().
		SetHeader(restful.HEADER_ContentType, restful.MIME_JSON).
		SetAuthToken(h.token).
		SetBody(body).Post(url)
	if err != nil {
		klog.Errorf("Failed to make register request: %v", err)
		return nil, err
	}

	if resp.StatusCode() != 200 {
		dump, e := httputil.DumpRequest(resp.Request.RawRequest, true)
		if e == nil {
			klog.Errorf("Failed to get response body=%s url=%s", string(dump), url)
		}

		return nil, errors.New(string(resp.Body()))
	}

	var regResp appcfg.RegisterResp
	err = json.Unmarshal(resp.Body(), &regResp)
	if err != nil {
		klog.Errorf("Failed to unmarshal response body=%s err=%v", string(resp.Body()), err)
		return nil, err
	}

	return &regResp, nil
}

func (h *HelmOps) systemServerHost() string {
	if systemServerHost != "" {
		return systemServerHost
	}

	return fmt.Sprintf("system-server.user-system-%s", h.app.OwnerName)
}

func (h *HelmOps) selectNode() (node string, appCache, userspace string, err error) {
	k8s, err := kubernetes.NewForConfig(h.kubeConfig)
	if err != nil {
		klog.Errorf("Failed to create kubernetes client: %v", err)
		return "", "", "", err
	}

	bflPods, err := k8s.CoreV1().Pods(h.ownerNamespace()).List(h.ctx,
		metav1.ListOptions{LabelSelector: "tier=bfl"})
	if err != nil {
		klog.Errorf("Failed to list pods in namespace %s: %v", h.ownerNamespace(), err)
		return "", "", "", err
	}

	if len(bflPods.Items) > 0 {
		bfl := bflPods.Items[0]

		vols := bfl.Spec.Volumes
		if len(vols) < 1 {
			klog.Error("No volumes found in bfl pod")
			return "", "", "", errors.New("user space not found")
		}

		// find user space pvc
		for _, vol := range vols {
			if vol.Name == constants.UserAppDataDirPVC || vol.Name == constants.UserSpaceDirPVC {
				if vol.PersistentVolumeClaim != nil {
					// find user space path
					pvc, err := k8s.CoreV1().PersistentVolumeClaims(h.ownerNamespace()).Get(h.ctx,
						vol.PersistentVolumeClaim.ClaimName,
						metav1.GetOptions{})
					if err != nil {
						klog.Errorf("Failed to get PVC %s in namespace %s: %v", vol.PersistentVolumeClaim.ClaimName, h.ownerNamespace(), err)
						return "", "", "", err
					}

					pv, err := k8s.CoreV1().PersistentVolumes().Get(h.ctx, pvc.Spec.VolumeName, metav1.GetOptions{})
					if err != nil {
						klog.Errorf("Failed to get PV %s: %v", pvc.Spec.VolumeName, err)
						return "", "", "", err
					}

					var path string
					if pv.Spec.Local != nil {
						path = pv.Spec.Local.Path
					}
					if path == "" {
						path = pv.Spec.HostPath.Path
					}

					switch vol.Name {
					case constants.UserAppDataDirPVC:
						appCache = path
					case constants.UserSpaceDirPVC:
						userspace = path
					}
				}
			}
		}

		if appCache == "" || userspace == "" {
			klog.Error("No user space or app cache found in bfl pod")
			return "", "", "", errors.New("user space not found")
		}

		return bfl.Spec.NodeName, appCache, userspace, nil
	}

	klog.Error("No bfl pod found in namespace", h.ownerNamespace())
	return "", "", "", errors.New("node not found")
}

func (h *HelmOps) ownerNamespace() string {
	return utils.UserspaceName(h.app.OwnerName)
}

func (h *HelmOps) createOIDCClient(values map[string]interface{}, userZone, namespace string) error {
	client, err := kubernetes.NewForConfig(h.kubeConfig)
	if err != nil {
		return err
	}

	id := h.app.AppID + "." + h.app.OwnerName
	secret := utils.GetRandomCharacters()

	values["oidc"] = map[string]interface{}{
		"client": map[string]interface{}{
			"id":     id,
			"secret": secret,
		},
		"issuer": "https://auth." + userZone,
	}

	oidcSecret := &corev1.Secret{
		ObjectMeta: metav1.ObjectMeta{
			Name:      constants.OIDCSecret,
			Namespace: namespace,
		},
		StringData: map[string]string{
			"id":     id,
			"secret": secret,
		},
	}
	_, err = client.CoreV1().Namespaces().Get(h.ctx, namespace, metav1.GetOptions{})
	if apierrors.IsNotFound(err) {
		ns := &corev1.Namespace{
			TypeMeta: metav1.TypeMeta{
				APIVersion: "v1",
				Kind:       "Namespace",
			},
			ObjectMeta: metav1.ObjectMeta{
				Name: namespace,
				Labels: map[string]string{
					"name": namespace,
				},
			},
		}
		_, err = client.CoreV1().Namespaces().Create(h.ctx, ns, metav1.CreateOptions{})
		if err != nil {
			return err
		}
	}

	_, err = client.CoreV1().Secrets(namespace).Get(h.ctx, oidcSecret.Name, metav1.GetOptions{})
	if err != nil {
		if !apierrors.IsNotFound(err) {
			return err
		}
	} else {
		err = client.CoreV1().Secrets(namespace).Delete(h.ctx, oidcSecret.Name, metav1.DeleteOptions{})
		if err != nil {
			return err
		}
	}

	_, err = client.CoreV1().Secrets(namespace).Create(h.ctx, oidcSecret, metav1.CreateOptions{})
	if err != nil {
		klog.Error("create oidc secret error, ", err)
		return err
	}

	return nil
}

func (h *HelmOps) WaitForStartUp() (bool, error) {
	if h.options.Source == constants.StudioSource {
		time.Sleep(5 * time.Second)
		return true, nil
	}
	timer := time.NewTicker(1 * time.Second)
	for {
		select {
		case <-timer.C:
			startedUp, err := h.isStartUp()
			klog.Infof("wait for app %s start up", h.app.AppName)
			if startedUp {
				name, _ := apputils.FmtAppMgrName(h.app.AppName, h.app.OwnerName, h.app.Namespace)
				err := apputils.UpdateAppMgrState(h.ctx, name, appv1alpha1.Initializing)
				if err != nil {
					klog.Errorf("update appmgr state failed %v", err)
				}
				return true, nil
			}
			if errors.Is(err, errcode.ErrPodPending) {
				return false, err
			}

		case <-h.ctx.Done():
			klog.Infof("Waiting for app startup canceled appName=%s", h.app.AppName)
			return false, nil
		}
	}
}

func (h *HelmOps) isStartUp() (bool, error) {
	pods, err := h.findAppSelectedPods()
	if err != nil {
		return false, err
	}
	return checkIfStartup(pods)
}

func (h *HelmOps) findAppSelectedPods() (*corev1.PodList, error) {
	var labelSelector string
	deployment, err := h.client.KubeClient.Kubernetes().AppsV1().Deployments(h.app.Namespace).
		Get(h.ctx, h.app.AppName, metav1.GetOptions{})

	if err == nil {
		labelSelector = metav1.FormatLabelSelector(deployment.Spec.Selector)
	}

	if apierrors.IsNotFound(err) {
		sts, err := h.client.KubeClient.Kubernetes().AppsV1().StatefulSets(h.app.Namespace).
			Get(h.ctx, h.app.AppName, metav1.GetOptions{})
		if err != nil {
			return nil, err
		}
		labelSelector = metav1.FormatLabelSelector(sts.Spec.Selector)
	}
	pods, err := h.client.KubeClient.Kubernetes().CoreV1().Pods(h.app.Namespace).
		List(h.ctx, metav1.ListOptions{LabelSelector: labelSelector})

	if err != nil {
		klog.Errorf("app %s get pods err %v", h.app.AppName, err)
		return nil, err
	}
	return pods, nil
}

func checkIfStartup(pods *corev1.PodList) (bool, error) {
	if len(pods.Items) == 0 {
		return false, errors.New("no pod found")
	}
	for _, pod := range pods.Items {
		creationTime := pod.GetCreationTimestamp()
		pendingDuration := time.Since(creationTime.Time)

		if pod.Status.Phase == corev1.PodPending && pendingDuration > time.Minute*10 {
			return false, errcode.ErrPodPending
		}
		totalContainers := len(pod.Spec.Containers)
		startedContainers := 0
		for i := len(pod.Status.ContainerStatuses) - 1; i >= 0; i-- {
			container := pod.Status.ContainerStatuses[i]
			if *container.Started {
				startedContainers++
			}
		}
		if startedContainers == totalContainers {
			return true, nil
		}
	}
	return false, nil
}

type applicationSettingsSubPolicy struct {
	URI      string `json:"uri"`
	Policy   string `json:"policy"`
	OneTime  bool   `json:"one_time"`
	Duration int32  `json:"valid_duration"`
}

type applicationSettingsPolicy struct {
	DefaultPolicy string                          `json:"default_policy"`
	SubPolicies   []*applicationSettingsSubPolicy `json:"sub_policies"`
	OneTime       bool                            `json:"one_time"`
	Duration      int32                           `json:"valid_duration"`
}

func getApplicationPolicy(policies []appcfg.AppPolicy, entrances []appv1alpha1.Entrance) (string, error) {
	subPolicy := make(map[string][]*applicationSettingsSubPolicy)

	for _, p := range policies {
		subPolicy[p.EntranceName] = append(subPolicy[p.EntranceName],
			&applicationSettingsSubPolicy{
				URI:      p.URIRegex,
				Policy:   p.Level,
				OneTime:  p.OneTime,
				Duration: int32(p.Duration / time.Second),
			})
	}

	policy := make(map[string]applicationSettingsPolicy)
	for _, e := range entrances {
		defaultPolicy := "system"
		sp := subPolicy[e.Name]
		if e.AuthLevel == constants.AuthorizationLevelOfPublic {
			defaultPolicy = constants.AuthorizationLevelOfPublic
		}
		policy[e.Name] = applicationSettingsPolicy{
			DefaultPolicy: defaultPolicy,
			OneTime:       false,
			Duration:      0,
			SubPolicies:   sp,
		}
	}

	policyStr, err := json.Marshal(policy)
	if err != nil {
		return "", err
	}
	return string(policyStr), nil
}

func parseAppPermission(data []appcfg.AppPermission) []appcfg.AppPermission {
	permissions := make([]appcfg.AppPermission, 0)
	for _, p := range data {
		switch perm := p.(type) {
		case string:
			if perm == "appdata-perm" {
				permissions = append(permissions, appcfg.AppDataRW)
			}
			if perm == "appcache-perm" {
				permissions = append(permissions, appcfg.AppCacheRW)
			}
			if perm == "userdata-perm" {
				permissions = append(permissions, appcfg.UserDataRW)
			}
		case appcfg.AppDataPermission:
			permissions = append(permissions, appcfg.AppDataRW)
		case appcfg.AppCachePermission:
			permissions = append(permissions, appcfg.AppCacheRW)
		case appcfg.UserDataPermission:
			permissions = append(permissions, appcfg.UserDataRW)
		case []appcfg.ProviderPermission:
			permissions = append(permissions, p)
		case []interface{}:
			var sps []appcfg.ProviderPermission
			for _, item := range perm {
				if m, ok := item.(map[string]interface{}); ok {
					var sp appcfg.ProviderPermission
					if appName, ok := m["appName"].(string); ok {
						sp.AppName = appName
					}
					if providerName, ok := m["providerName"].(string); ok {
						sp.ProviderName = providerName
					}
					if ns, ok := m["namespace"].(string); ok {
						sp.Namespace = ns
					}
					sps = append(sps, sp)
				}
			}
			permissions = append(permissions, sps)
		}
	}
	return permissions
}

func (h *HelmOps) Install() error {
	var err error
	values, err := h.SetValues()
	if err != nil {
		klog.Errorf("set values err %v", err)
		return err
	}

	err = h.TaprApply(values, "")
	if err != nil {
		return err
	}

	err = h.install(values)
	if err != nil && !errors.Is(err, driver.ErrReleaseExists) {
		klog.Errorf("Failed to install chart err=%v", err)
		h.Uninstall()
		return err
	}

	err = h.AddApplicationLabelsToDeployment()
	if err != nil {
		h.Uninstall()
		return err
	}

	isDepClusterScopedApp := false
	client, err := versioned.NewForConfig(h.kubeConfig)
	if err != nil {
		return err
	}
	apps, err := client.AppV1alpha1().Applications().List(h.ctx, metav1.ListOptions{})
	if err != nil {
		return err
	}

	for _, dep := range h.app.Dependencies {
		if dep.Type == constants.DependencyTypeSystem {
			continue
		}
		for _, app := range apps.Items {
			if app.Spec.Name == dep.Name && app.Spec.Settings["clusterScoped"] == "true" {
				isDepClusterScopedApp = true
				break
			}
		}

	}
	if isDepClusterScopedApp {
		err = h.AddLabelToNamespaceForDependClusterApp()
		if err != nil {
			h.Uninstall()
			return err
		}
	}

	if err = h.RegisterOrUnregisterAppProvider(Register); err != nil {
		klog.Errorf("Failed to register app provider err=%v", err)
		h.Uninstall()
		return err
	}

	if h.app.Type == appv1alpha1.Middleware.String() {
		return nil
	}
	ok, err := h.WaitForStartUp()
	if err != nil && errors.Is(err, errcode.ErrPodPending) {
		return err
	}
	if !ok {
		h.Uninstall()
		return err
	}

	return nil
}

func (h *HelmOps) WaitForLaunch() (bool, error) {
	if h.options.Source == constants.StudioSource {
		return true, nil
	}

	timer := time.NewTicker(1 * time.Second)
	entrances := h.app.Entrances
	entranceCount := 0
	for _, e := range entrances {
		if !e.Skip {
			entranceCount++
		}
	}
	for {
		select {
		case <-timer.C:
			count := 0
			for _, e := range entrances {
				if e.Skip {
					continue
				}
				klog.Info("Waiting service for launch :", e.Host)
				host := fmt.Sprintf("%s.%s", e.Host, h.app.Namespace)
				if apputils.TryConnect(host, strconv.Itoa(int(e.Port))) {
					count++
				}
			}
			if entranceCount == count {
				return true, nil
			}

		case <-h.ctx.Done():
			klog.Infof("Waiting for launch canceled appName=%s", h.app.AppName)
			return false, h.ctx.Err()
		}
	}
}

func (h *HelmOps) App() *appcfg.ApplicationConfig {
	return h.app
}

func (h *HelmOps) KubeConfig() *rest.Config {
	return h.kubeConfig
}

func (h *HelmOps) ActionConfig() *action.Configuration {
	return h.actionConfig
}

func (h *HelmOps) Settings() *cli.EnvSettings {
	return h.settings
}

func (h *HelmOps) Context() context.Context {
	return h.ctx
}

func (h *HelmOps) Token() string {
	return h.token
}

func (h *HelmOps) Client() *clientset.ClientSet {
	return h.client
}

func (h *HelmOps) Options() *Opt {
	return &h.options
}

func (h *HelmOps) RegisterOrUnregisterAppProvider(operation ProviderOperation) error {
	var providers []*appcfg.ProviderCfg
	appEntrances, err := h.app.GetEntrances(h.ctx)
	if err != nil {
		klog.Errorf("Failed to get app entrances for app %s: %v", h.app.AppName, err)
		return err
	}
	for _, provider := range h.app.Provider {
		providerEntrance, ok := appEntrances[provider.Entrance]
		if !ok {
			err = fmt.Errorf("entrance %s not found for the provider of app %s", provider.Entrance, h.app.AppName)
			klog.Error(err)
			return err
		}

		domain := providerEntrance.URL
		service := fmt.Sprintf("%s.%s:%d", providerEntrance.Host, h.app.Namespace, providerEntrance.Port)

		providers = append(providers, &appcfg.ProviderCfg{
			Service:  service,
			Domain:   domain,
			Provider: provider,
		})
	}

	if len(providers) > 0 {
		register := appcfg.ProviderRegisterRequest{
			AppName:      h.app.AppName,
			AppNamespace: h.app.Namespace,
			Providers:    providers,
		}

		url := fmt.Sprintf("http://%s/provider/v2alpha1/%s", h.systemServerHost(), operation)
		client := resty.New()

		resp, err := client.SetTimeout(2*time.Second).R().
			SetHeader(restful.HEADER_ContentType, restful.MIME_JSON).
			SetAuthToken(h.token).
			SetBody(register).Post(url)
		if err != nil {
			return err
		}

		if resp.StatusCode() != 200 {
			dump, e := httputil.DumpRequest(resp.Request.RawRequest, true)
			if e == nil {
				klog.Errorf("Failed to get response body=%s url=%s", string(dump), url)
			}

			return errors.New(string(resp.Body()))
		}
	}

	return nil
}
