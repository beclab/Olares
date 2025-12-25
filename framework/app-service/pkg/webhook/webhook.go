package webhook

import (
	"context"
	"crypto"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"io/ioutil"
	"net/http"
	"sort"
	"strconv"

	"github.com/beclab/Olares/framework/app-service/api/app.bytetrade.io/v1alpha1"
	"github.com/beclab/Olares/framework/app-service/pkg/apiserver/api"
	"github.com/beclab/Olares/framework/app-service/pkg/appcfg"
	appcfg_mod "github.com/beclab/Olares/framework/app-service/pkg/appcfg"
	"github.com/beclab/Olares/framework/app-service/pkg/constants"
	"github.com/beclab/Olares/framework/app-service/pkg/generated/clientset/versioned"
	"github.com/beclab/Olares/framework/app-service/pkg/provider"
	"github.com/beclab/Olares/framework/app-service/pkg/sandbox/sidecar"
	"github.com/beclab/Olares/framework/app-service/pkg/security"
	"github.com/beclab/Olares/framework/app-service/pkg/utils"

	apputils "github.com/beclab/Olares/framework/app-service/pkg/utils/app"
	"github.com/emicklei/go-restful/v3"
	"github.com/google/uuid"
	"github.com/thoas/go-funk"
	admissionv1 "k8s.io/api/admission/v1"
	corev1 "k8s.io/api/core/v1"
	apierrors "k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/api/resource"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/apimachinery/pkg/labels"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/runtime/serializer"
	"k8s.io/apimachinery/pkg/types"
	"k8s.io/client-go/dynamic"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/rest"
	"k8s.io/klog/v2"
	"k8s.io/utils/ptr"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/webhook/admission"
)

var (
	errEmptyAdmissionRequestBody = fmt.Errorf("empty request admission request body")

	// codecs is the codec factory used by the deserializer.
	codecs = serializer.NewCodecFactory(runtime.NewScheme())

	// Deserializer is used to decode the admission request body.
	Deserializer = codecs.UniversalDeserializer()

	// UUIDAnnotation uuid key for annotation.
	UUIDAnnotation = "sidecar.bytetrade.io/proxy-uuid"
)

// Webhook used to implement a webhook.
type Webhook struct {
	kubeClient    *kubernetes.Clientset
	dynamicClient *versioned.Clientset
}

// New create a webhook client.
func New(config *rest.Config) (*Webhook, error) {
	client, err := kubernetes.NewForConfig(config)
	if err != nil {
		return nil, err
	}
	dynamicClient, err := versioned.NewForConfig(config)
	if err != nil {
		return nil, err
	}

	return &Webhook{
		kubeClient:    client,
		dynamicClient: dynamicClient,
	}, nil
}

// GetAppConfig get app config by namespace.
func (wh *Webhook) GetAppConfig(namespace string) (appMgr *v1alpha1.ApplicationManager, appCfg *appcfg_mod.ApplicationConfig, isShared bool, err error) {
	list, err := wh.dynamicClient.AppV1alpha1().ApplicationManagers().List(context.TODO(), metav1.ListOptions{})
	if err != nil {
		return nil, nil, false, err
	}
	sorted := list.Items
	sort.Slice(sorted, func(i, j int) bool {
		return sorted[j].CreationTimestamp.Before(&sorted[i].CreationTimestamp)
	})

	ns, err := wh.kubeClient.CoreV1().Namespaces().Get(context.TODO(), namespace, metav1.GetOptions{})
	if err != nil {
		klog.Error("failed to get namespace, namespace=", namespace, " err=", err)
		return nil, nil, false, err
	}

	refAppName := ns.Labels[constants.ApplicationNameLabel]
	sharedNamespace := ns.Labels["bytetrade.io/ns-shared"]
	installedUser := ns.Labels[constants.ApplicationInstallUserLabel]

	var appconfig appcfg.ApplicationConfig
	for _, a := range sorted {
		switch {
		case a.Spec.AppNamespace == namespace && (a.Spec.Type == v1alpha1.App || a.Spec.Type == v1alpha1.Middleware),
			// shared server namespace
			sharedNamespace == "true" && a.Spec.AppName == refAppName && a.Spec.AppOwner == installedUser &&
				(a.Spec.Type == v1alpha1.App || a.Spec.Type == v1alpha1.Middleware):
			err = json.Unmarshal([]byte(a.Spec.Config), &appconfig)
			if err != nil {
				return nil, nil, false, err
			}
			return &a, &appconfig, (sharedNamespace == "true" && a.Spec.AppName == refAppName), nil
		}
	}
	return nil, nil, false, api.ErrApplicationManagerNotFound
}

// GetAdmissionRequestBody returns admission request body.
func (wh *Webhook) GetAdmissionRequestBody(req *restful.Request, resp *restful.Response) ([]byte, bool) {
	emptyBodyError := func() ([]byte, bool) {
		klog.Error("Failed to read admission request body err=body is empty")
		api.HandleBadRequest(resp, req, errEmptyAdmissionRequestBody)
		return nil, false
	}

	if req.Request.Body == nil {
		return emptyBodyError()
	}

	admissionRequestBody, err := ioutil.ReadAll(req.Request.Body)
	if err != nil {
		api.HandleInternalError(resp, req, err)
		klog.Errorf("Failed to  read admission request body; Responded to admission request with HTTP=%v err=%v", http.StatusInternalServerError, err)
		return admissionRequestBody, false
	}

	if len(admissionRequestBody) == 0 {
		return emptyBodyError()
	}

	return admissionRequestBody, true
}

// CreatePatch create a patch for a pod.
func (wh *Webhook) CreatePatch(
	ctx context.Context,
	pod *corev1.Pod,
	req *admissionv1.AdmissionRequest,
	proxyUUID uuid.UUID, injectPolicy, injectWs, injectUpload bool,
	injectSharedPod *bool,
	appmgr *v1alpha1.ApplicationManager,
	appcfg *appcfg_mod.ApplicationConfig,
	perms []appcfg.ProviderPermission,
) ([]byte, error) {
	isInjected, prevUUID := isInjectedPod(pod)

	if isInjected {
		// TODO: force mutate
		klog.Infof("Pod is injected with uuid=%s namespace=%s", prevUUID, req.Namespace)
		return makePatches(req, pod)
	}

	// inject sidecar only for the app's namespace
	if req.Namespace == appmgr.Spec.AppNamespace {
		configMapName, err := wh.createSidecarConfigMap(ctx, pod, proxyUUID.String(), req.Namespace, injectPolicy, injectWs, injectUpload, appmgr, appcfg, perms)
		if err != nil {
			return nil, err
		}

		volume := sidecar.GetSidecarVolumeSpec(configMapName)

		if pod.Spec.Volumes == nil {
			pod.Spec.Volumes = []corev1.Volume{}
		}

		pod.Spec.Volumes = append(pod.Spec.Volumes, volume, sidecar.GetEnvoyConfigWorkVolume())

		clusterID := fmt.Sprintf("%s.%s", pod.Spec.ServiceAccountName, req.Name)
		envoyFilename := constants.EnvoyConfigFilePath + "/" + constants.EnvoyConfigFileName
		// pod is not an entrance pod, just inject outbound proxy
		if !injectPolicy {
			envoyFilename = constants.EnvoyConfigFilePath + "/" + constants.EnvoyConfigOnlyOutBoundFileName
		}
		appKey, appSecret, _ := wh.getAppKeySecret(req.Namespace)

		if injectPolicy || len(appcfg.PodsSelectors) == 0 || wh.isSelected(appcfg.PodsSelectors, pod) {
			initContainer := sidecar.GetInitContainerSpec(appcfg)
			pod.Spec.InitContainers = append(pod.Spec.InitContainers, initContainer)
			policySidecar := sidecar.GetEnvoySidecarContainerSpec(clusterID, envoyFilename, appKey, appSecret)
			pod.Spec.Containers = append(pod.Spec.Containers, policySidecar)

			pod.Spec.InitContainers = append(
				[]corev1.Container{
					sidecar.GetInitContainerSpecForWaitFor(appcfg.OwnerName),
					sidecar.GetInitContainerSpecForRenderEnvoyConfig(),
				},
				pod.Spec.InitContainers...)
		}

		if injectWs {
			wsSidecar := sidecar.GetWebSocketSideCarContainerSpec(&appcfg.WsConfig)
			pod.Spec.Containers = append(pod.Spec.Containers, wsSidecar)
		}
		if injectUpload {
			uploadSidecar := sidecar.GetUploadSideCarContainerSpec(pod, &appcfg.Upload)
			if uploadSidecar != nil {
				pod.Spec.Containers = append(pod.Spec.Containers, *uploadSidecar)
			}
		}
	} // end of inject sidecar

	if injectSharedPod != nil {
		if *injectSharedPod {
			if pod.Labels == nil {
				pod.Labels = make(map[string]string)
			}
			pod.Labels[constants.AppSharedEntrancesLabel] = "true"
		} else {
			if pod.Labels != nil {
				delete(pod.Labels, constants.AppSharedEntrancesLabel)
			}
		}
	}

	if pod.Annotations == nil {
		pod.Annotations = make(map[string]string)
	}
	pod.Annotations[UUIDAnnotation] = proxyUUID.String()

	// add header to probes
	if err := wh.patchProbeHeaders(ctx, pod); err != nil {
		klog.Errorf("Failed to patch probe headers for pod=%s/%s err=%v", pod.Namespace, pod.Name, err)
		return nil, err
	}
	return makePatches(req, pod)
}

func (wh *Webhook) getProbeUA(ctx context.Context, pod *corev1.Pod) (string, error) {
	secret, err := wh.kubeClient.CoreV1().Secrets("os-framework").Get(ctx, "authelia-secrets", metav1.GetOptions{})
	if err != nil {
		klog.Errorf("Failed to get authelia-secrets in os-framework namespace err=%v", err)
		return "", err
	}
	signSecret, ok := secret.Data["probe_secret"]
	if !ok {
		klog.Errorf("Failed to get probe_secret in authelia-secrets")
		return "", fmt.Errorf("probe-secret not found in authelia-secrets")
	}

	uuid := pod.Annotations[UUIDAnnotation]
	MD5 := func(str string) string {
		h := crypto.MD5.New()
		h.Write([]byte(str))
		return hex.EncodeToString(h.Sum(nil))
	}
	sign := MD5(uuid + string(signSecret))
	return fmt.Sprintf("%s/%s", uuid, sign), nil
}

func (wh *Webhook) patchProbeHeaders(ctx context.Context, pod *corev1.Pod) error {
	const UA_HEADER = "User-Agent"
	ua, err := wh.getProbeUA(ctx, pod)
	if err != nil {
		klog.Errorf("Failed to get probe UA for pod=%s/%s err=%v", pod.Namespace, pod.Name, err)
		return err
	}

	setProbeUA := func(action *corev1.HTTPGetAction) {
		for i, h := range action.HTTPHeaders {
			if h.Name == UA_HEADER {
				action.HTTPHeaders[i].Value = ua
				return
			}
		}

		// not found, add new header
		action.HTTPHeaders = append(action.HTTPHeaders, corev1.HTTPHeader{
			Name:  UA_HEADER,
			Value: ua,
		})
	}

	for _, c := range pod.Spec.Containers {
		if c.LivenessProbe != nil && c.LivenessProbe.HTTPGet != nil {
			setProbeUA(c.LivenessProbe.HTTPGet)
		}
		if c.ReadinessProbe != nil && c.ReadinessProbe.HTTPGet != nil {
			setProbeUA(c.ReadinessProbe.HTTPGet)
		}
		if c.StartupProbe != nil && c.StartupProbe.HTTPGet != nil {
			setProbeUA(c.StartupProbe.HTTPGet)
		}
	}

	return nil
}

// PatchAdmissionResponse returns an admission response with patch data.
func (wh *Webhook) PatchAdmissionResponse(resp *admissionv1.AdmissionResponse, patchBytes []byte) {
	resp.Patch = patchBytes
	pt := admissionv1.PatchTypeJSONPatch
	resp.PatchType = &pt
}

// AdmissionError wraps error as AdmissionResponse
func (wh *Webhook) AdmissionError(uid types.UID, err error) *admissionv1.AdmissionResponse {
	return &admissionv1.AdmissionResponse{
		UID: uid,
		Result: &metav1.Status{
			Message: err.Error(),
		},
	}
}

// MustInject checks which inject operation should do for a pod.
func (wh *Webhook) MustInject(ctx context.Context, pod *corev1.Pod, namespace string) (
	injectPolicy, injectWs, injectUpload bool, injectSharedPod *bool, perms []appcfg.ProviderPermission,
	appCfg *appcfg_mod.ApplicationConfig, appMgr *v1alpha1.ApplicationManager, err error) {
	var isShared bool

	perms = make([]appcfg.ProviderPermission, 0)
	if !isNamespaceInjectable(namespace) {
		return
	}

	// TODO: uninject annotation

	// get appLabel from namespace
	_, err = wh.kubeClient.CoreV1().Namespaces().Get(ctx, namespace, metav1.GetOptions{})
	if err != nil {
		klog.Errorf("Failed to get namespace=%s err=%v", namespace, err)
		return
	}

	appMgr, appCfg, isShared, err = wh.GetAppConfig(namespace)
	if err != nil {
		if errors.Is(err, api.ErrApplicationManagerNotFound) {
			err = nil
		} else {
			klog.Errorf("Failed to get app config err=%v", err)
			return
		}
	}

	if appCfg == nil {
		klog.Infof("Unknown namespace=%s, do not inject", namespace)
		return
	}
	if appCfg.IsMiddleware() {
		return
	}

	if !isShared {
		if appCfg.WsConfig.URL != "" && appCfg.WsConfig.Port > 0 {
			injectWs = true
		}
		if appCfg.Upload.Dest != "" {
			injectUpload = true
		}
		for _, p := range appCfg.Permission {
			klog.Info("found permission: ", p)
			if providerP, ok := p.([]interface{}); ok {
				for _, v := range providerP {
					provider := v.(map[string]interface{})
					var ns string
					if val, ok := provider["namespace"].(string); ok {
						ns = val
					}
					providerAppName := provider["appName"].(string)
					providerName := provider["providerName"].(string)
					perms = append(perms, appcfg_mod.ProviderPermission{
						AppName:      providerAppName,
						Namespace:    ns,
						ProviderName: providerName,
					})

				}
			}

		}

		injectPolicy = false
		for _, e := range appCfg.Entrances {
			var isEntrancePod bool
			isEntrancePod, err = wh.isAppEntrancePod(ctx, appCfg.AppName, e.Host, pod, namespace)
			klog.Infof("entranceName=%s isEntrancePod=%v", e.Name, isEntrancePod)
			if err != nil {
				return false, false, false, nil, perms, nil, nil, err
			}

			if isEntrancePod {
				injectPolicy = true
				break
			}
		}
	} // end of non-shared namespace's pod

	for _, e := range appCfg.SharedEntrances {
		var isEntrancePod bool
		isEntrancePod, err = wh.isAppEntrancePod(ctx, appCfg.AppName, e.Host, pod, namespace)
		klog.Infof("entranceName=%s isEntrancePod=%v", e.Name, isEntrancePod)
		if err != nil {
			return false, false, false, nil, perms, nil, nil, err
		}

		if isEntrancePod {
			injectSharedPod = ptr.To(true)
			break
		}
	}

	// not a shared entrance pod, should not have the shared entrance label
	if injectSharedPod == nil && pod.Labels != nil {
		if v, ok := pod.Labels[constants.AppSharedEntrancesLabel]; ok && v == "false" {
			injectSharedPod = ptr.To(false)
		}
	}

	return
}

func (wh *Webhook) isAppEntrancePod(ctx context.Context, appname, host string, pod *corev1.Pod, namespace string) (bool, error) {
	service, err := wh.kubeClient.CoreV1().Services(namespace).Get(ctx, host, metav1.GetOptions{})
	if err != nil {
		klog.Errorf("Failed to get app service appName=%s host=%s err=%v", appname, host, err)
		if apierrors.IsNotFound(err) {
			return false, nil
		}
		return false, err
	}

	selector, err := labels.ValidatedSelectorFromSet(service.Spec.Selector)
	if err != nil {
		klog.Errorf("Failed to get service selector appName=%s host=%s err=%v", appname, host, err)
		return false, err
	}

	return selector.Matches(labels.Set(pod.GetLabels())), nil
}

func (wh *Webhook) createSidecarConfigMap(
	ctx context.Context, pod *corev1.Pod,
	proxyUUID, namespace string, injectPolicy, injectWs, injectUpload bool,
	appmgr *v1alpha1.ApplicationManager, appcfg *appcfg_mod.ApplicationConfig,
	perms []appcfg_mod.ProviderPermission,
) (string, error) {
	configMapName := fmt.Sprintf("%s-%s", constants.SidecarConfigMapVolumeName, proxyUUID)
	if deployName := utils.GetDeploymentName(pod); deployName != "" {
		configMapName = fmt.Sprintf("%s-%s", constants.SidecarConfigMapVolumeName, deployName)
	}
	cm, e := wh.kubeClient.CoreV1().ConfigMaps(namespace).Get(ctx, configMapName, metav1.GetOptions{})
	if e != nil && !apierrors.IsNotFound(e) {
		return "", e
	}

	permCfg, err := apputils.ProviderPermissionsConvertor(perms).ToPermissionCfg(ctx, appcfg.OwnerName, appmgr.GetMarketSource())
	if err != nil {
		klog.Errorf("Failed to convert permissions for app %s: %v", appcfg.AppName, err)
		return "", err
	}

	newConfigMap := sidecar.GetSidecarConfigMap(configMapName, namespace, appcfg, injectPolicy, injectWs, injectUpload, pod, permCfg)
	if e == nil {
		// configmap found
		cm.Data = newConfigMap.Data
		if _, err := wh.kubeClient.CoreV1().ConfigMaps(namespace).Update(ctx, cm, metav1.UpdateOptions{}); err != nil {
			klog.Errorf("Failed to update sidecar configmap=%s in namespace=%s err=%v", configMapName, namespace, err)
			return "", err
		}
	} else {
		if _, err := wh.kubeClient.CoreV1().ConfigMaps(namespace).Create(ctx, newConfigMap, metav1.CreateOptions{}); err != nil {
			klog.Errorf("Failed to create sidecar configmap=%s in namespace=%s err=%v", configMapName, namespace, err)
			return "", err
		}
	}

	return configMapName, nil
}

func isNamespaceInjectable(namespace string) bool {
	if security.IsUnderLayerNamespace(namespace) {
		return false
	}

	if security.IsOSSystemNamespace(namespace) {
		return false
	}

	if ok, _ := security.IsUserInternalNamespaces(namespace); ok {
		return false
	}

	return true
}

func isInjectedPod(pod *corev1.Pod) (bool, string) {
	if pod.Annotations != nil {
		if proxyUUID, ok := pod.Annotations[UUIDAnnotation]; ok {
			for _, c := range pod.Spec.Containers {
				if c.Name == constants.EnvoyContainerName {
					return true, proxyUUID
				}
			}
		}
	}

	for _, c := range pod.Spec.InitContainers {
		if c.Name == constants.SidecarInitContainerName {
			return true, ""
		}
	}

	return false, ""
}

func makePatches(req *admissionv1.AdmissionRequest, pod *corev1.Pod) ([]byte, error) {
	original := req.Object.Raw
	current, err := json.Marshal(pod)
	if err != nil {
		klog.Errorf("Failed to  marshal pod with UID=%s", pod.ObjectMeta.UID)
	}
	admissionResponse := admission.PatchResponseFromRaw(original, current)
	return json.Marshal(admissionResponse.Patches)
}

type patchOp struct {
	Op    string      `json:"op"`
	Path  string      `json:"path"`
	Value interface{} `json:"value,omitempty"`
}

var resourcePath = "/spec/template/spec/containers/%d/resources/limits"
var envPath = "/spec/template/spec/containers/%d/env/%s"
var runtimeClassPath = "/spec/template/spec/runtimeClassName"

type EnvKeyValue struct {
	Key   string
	Value string
}

// CreatePatchForDeployment add gpu env for deployment and returns patch bytes.
func CreatePatchForDeployment(tpl *corev1.PodTemplateSpec, namespace string, gpuRequired *resource.Quantity, typeKey string, envKeyValues []EnvKeyValue) ([]byte, error) {
	patches, err := addResourceLimits(tpl, namespace, gpuRequired, typeKey, envKeyValues)
	if err != nil {
		return []byte{}, err
	}
	return json.Marshal(patches)
}

func addResourceLimits(tpl *corev1.PodTemplateSpec, namespace string, gpuRequired *resource.Quantity, typeKey string, envKeyValues []EnvKeyValue) (patch []patchOp, err error) {
	if typeKey == constants.NvidiaGPU || typeKey == constants.NvshareGPU {
		if tpl.Spec.RuntimeClassName != nil {
			patch = append(patch, patchOp{
				Op:    constants.PatchOpReplace,
				Path:  runtimeClassPath,
				Value: "nvidia",
			})
		} else {
			patch = append(patch, patchOp{
				Op:    constants.PatchOpAdd,
				Path:  runtimeClassPath,
				Value: "nvidia",
			})
		}
	}

	for i := range tpl.Spec.Containers {
		container := tpl.Spec.Containers[i]

		if len(container.Resources.Limits) == 0 {
			patch = append(patch, patchOp{
				Op:   constants.PatchOpAdd,
				Path: fmt.Sprintf(resourcePath, i),
				Value: map[string]interface{}{
					typeKey: "1",
				},
			})
		} else {
			t := make(map[string]map[string]string)
			t["limits"] = map[string]string{}
			for k, v := range container.Resources.Limits {
				if k.String() == constants.NvidiaGPU || k.String() == constants.NvshareGPU || k.String() == constants.VirtAiTechVGPU {
					continue
				}
				t["limits"][k.String()] = v.String()
			}
			t["limits"][typeKey] = "1"
			patch = append(patch, patchOp{
				Op:    constants.PatchOpReplace,
				Path:  fmt.Sprintf(resourcePath, i),
				Value: t["limits"],
			})
		}
		envNames := make([]string, 0)
		if len(container.Env) == 0 {
			value := make([]map[string]string, 0)
			for _, e := range envKeyValues {
				if e.Value == "" {
					continue
				}
				envNames = append(envNames, e.Key)
				value = append(value, map[string]string{
					"name":  e.Key,
					"value": e.Value,
				})
			}
			op := patchOp{
				Op:    "add",
				Path:  fmt.Sprintf("/spec/template/spec/containers/%d/env", i),
				Value: value,
			}
			patch = append(patch, op)
		} else {
			for envIdx, env := range container.Env {
				for _, e := range envKeyValues {
					if e.Value == "" {
						continue
					}
					if env.Name == e.Key {
						envNames = append(envNames, env.Name)
						patch = append(patch, genPatchesForEnv(constants.PatchOpReplace, i, envIdx, e.Key, e.Value)...)
					}
				}
			}
		}
		for _, env := range envKeyValues {
			if !funk.Contains(envNames, env.Key) {
				patch = append(patch, genPatchesForEnv(constants.PatchOpAdd, i, -1, env.Key, env.Value)...)
			}
		}

	}

	return patch, nil
}

func genPatchesForEnv(op string, containerIdx, envIdx int, name, value string) (patch []patchOp) {
	envIndexString := "-"
	if op == constants.PatchOpReplace {
		envIndexString = strconv.Itoa(envIdx)
	}
	patch = append(patch, patchOp{
		Op:   op,
		Path: fmt.Sprintf(envPath, containerIdx, envIndexString),
		Value: map[string]string{
			"name":  name,
			"value": value,
		},
	})
	return patch
}

func (wh *Webhook) getAppKeySecret(namespace string) (string, string, error) {
	config, err := ctrl.GetConfig()
	if err != nil {
		return "", "", err
	}
	client, err := dynamic.NewForConfig(config)
	if err != nil {
		return "", "", err
	}
	_, appcfg, isShared, err := wh.GetAppConfig(namespace)
	if err != nil {
		klog.Errorf("Failed to get app config err=%v", err)
		return "", "", err
	}

	if isShared {
		// shared namespace, no need to get appkey/secret
		return "", "", nil
	}

	apClient := provider.NewApplicationPermissionRequest(client)
	ap, err := apClient.Get(context.TODO(), "user-system-"+appcfg.OwnerName, appcfg.AppName, metav1.GetOptions{})
	if err != nil {
		return "", "", err
	}
	var appKey, appSecret string
	if ap != nil {
		appKey, _, _ = unstructured.NestedString(ap.Object, "spec", "key")
		appSecret, _, _ = unstructured.NestedString(ap.Object, "spec", "secret")
		return appKey, appSecret, nil
	}
	return "", "", errors.New("nil applicationpermission object")
}

func (wh *Webhook) isSelected(podSelectors []metav1.LabelSelector, pod *corev1.Pod) bool {
	for _, ps := range podSelectors {
		ls, err := metav1.LabelSelectorAsSelector(&ps)
		if err != nil {
			continue
		}
		selected := ls.Matches(labels.Set(pod.Labels))
		if selected {
			return true
		}
	}
	return false
}
