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
	"strings"

	"github.com/beclab/Olares/framework/app-service/pkg/apiserver/api"
	"github.com/beclab/Olares/framework/app-service/pkg/appcfg"
	"github.com/beclab/Olares/framework/app-service/pkg/constants"
<<<<<<< HEAD
	"github.com/beclab/Olares/framework/app-service/pkg/gateway"
	"github.com/beclab/Olares/framework/app-service/pkg/mesh"
=======
	"github.com/beclab/Olares/framework/app-service/pkg/gateway/calleragent"
>>>>>>> ed949b076 (feat(app-service): scaffold Shared HTTPS caller agent webhook injection)
	"github.com/beclab/Olares/framework/app-service/pkg/provider"
	"github.com/beclab/Olares/framework/app-service/pkg/sandbox/sidecar"
	"github.com/beclab/Olares/framework/app-service/pkg/security"
	"github.com/beclab/Olares/framework/app-service/pkg/utils"
	"github.com/beclab/api/api/app.bytetrade.io/v1alpha1"
	"github.com/beclab/api/pkg/generated/clientset/versioned"

	apputils "github.com/beclab/Olares/framework/app-service/pkg/utils/app"
	"github.com/emicklei/go-restful/v3"
	"github.com/google/uuid"
	"github.com/thoas/go-funk"
	admissionv1 "k8s.io/api/admission/v1"
	corev1 "k8s.io/api/core/v1"
	apierrors "k8s.io/apimachinery/pkg/api/errors"
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
	kubeClient    kubernetes.Interface
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
func (wh *Webhook) GetAppConfig(namespace string) (appMgr *v1alpha1.ApplicationManager, appConfig *appcfg.ApplicationConfig, isShared bool, err error) {
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
	proxyUUID uuid.UUID, injectPolicy, injectWs, injectUpload, injectCallerAgent bool,
	injectSharedPod *bool,
	appmgr *v1alpha1.ApplicationManager,
	appConfig *appcfg.ApplicationConfig,
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
		needsEnvoySidecar := wh.shouldInjectEnvoySidecar(ctx, injectPolicy, appConfig, pod)

		configMapName, err := wh.createSidecarConfigMap(ctx, pod, proxyUUID.String(), req.Namespace, injectPolicy, injectWs, injectUpload, appmgr, appConfig, perms)
		if err != nil {
			return nil, err
		}

		volume := sidecar.GetSidecarVolumeSpec(configMapName)

		if pod.Spec.Volumes == nil {
			pod.Spec.Volumes = []corev1.Volume{}
		}

		if needsEnvoySidecar {
			pod.Spec.Volumes = append(pod.Spec.Volumes, volume, sidecar.GetEnvoyConfigWorkVolume())

			clusterID := fmt.Sprintf("%s.%s", pod.Spec.ServiceAccountName, req.Name)
			envoyFilename := constants.EnvoyConfigFilePath + "/" + constants.EnvoyConfigFileName
			// pod is not an entrance pod, just inject outbound proxy
			if !injectPolicy {
				envoyFilename = constants.EnvoyConfigFilePath + "/" + constants.EnvoyConfigOnlyOutBoundFileName
			}
			appKey, appSecret, _ := wh.getAppKeySecret(req.Namespace)

			// If the owning Application enables overlay-gateway, multus will
			// attach a macvlan NIC (net1) to the pod. Tell the iptables init
			// container to install bypass RETURN rules for that interface so
			// north/south traffic on net1 doesn't get redirected to envoy.
			injectMacvlan, err := wh.ShouldInjectMacvlanInit(ctx, pod, req.Namespace)
			if err != nil {
				klog.Errorf("Failed to evaluate macvlan-init for sidecar pod=%s/%s err=%v", req.Namespace, pod.Name, err)
				return nil, err
			}
			initContainer := sidecar.GetInitContainerSpec(appConfig, injectMacvlan)
			pod.Spec.InitContainers = append(pod.Spec.InitContainers, initContainer)
			policySidecar := sidecar.GetEnvoySidecarContainerSpec(clusterID, envoyFilename, appKey, appSecret)
			pod.Spec.Containers = append(pod.Spec.Containers, policySidecar)

			pod.Spec.InitContainers = append(
				[]corev1.Container{
					sidecar.GetInitContainerSpecForWaitFor(appConfig.OwnerName),
					sidecar.GetInitContainerSpecForRenderEnvoyConfig(),
				},
				pod.Spec.InitContainers...)
		} else if injectWs || injectUpload {
			pod.Spec.Volumes = append(pod.Spec.Volumes, volume)
		}

		if injectWs {
			wsSidecar := sidecar.GetWebSocketSideCarContainerSpec(&appConfig.WsConfig)
			pod.Spec.Containers = append(pod.Spec.Containers, wsSidecar)
		}
		if injectUpload {
			uploadSidecar := sidecar.GetUploadSideCarContainerSpec(pod, &appConfig.Upload)
			if uploadSidecar != nil {
				pod.Spec.Containers = append(pod.Spec.Containers, *uploadSidecar)
			}
		}
		if injectCallerAgent {
			pod.Spec.Volumes = append(pod.Spec.Volumes, calleragent.JWTSecretVolume())
			pod.Spec.Containers = append(pod.Spec.Containers, calleragent.ContainerSpec())
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

func (wh *Webhook) shouldInjectEnvoySidecar(ctx context.Context, injectPolicy bool, appConfig *appcfg.ApplicationConfig, pod *corev1.Pod) bool {
	if !injectPolicy {
		if wh != nil && wh.kubeClient != nil && mesh.ShouldSkipEnvoySidecar(ctx, wh.kubeClient) {
			return false
		}
	}
	if appConfig == nil {
		return injectPolicy
	}
	return injectPolicy || len(appConfig.PodsSelectors) == 0 || wh.isSelected(appConfig.PodsSelectors, pod)
}

func (wh *Webhook) shouldSkipInboundEntranceSidecar(ctx context.Context, appConfig *appcfg.ApplicationConfig, ns, entranceName string) (bool, error) {
	if appConfig == nil || wh == nil || wh.kubeClient == nil {
		return false, nil
	}
	applicationName, err := apputils.FmtAppMgrName(appConfig.AppName, appConfig.OwnerName, ns)
	if err != nil {
		return false, err
	}
	app, err := wh.dynamicClient.AppV1alpha1().Applications().Get(ctx, applicationName, metav1.GetOptions{})
	if err != nil {
		if apierrors.IsNotFound(err) {
			return false, nil
		}
		return false, err
	}
	appid := strings.ToLower(strings.TrimSpace(app.Spec.Appid))
	if appid == "" {
		return false, nil
	}
	srrName := gateway.ResourceNameForEntranceApp(appid, entranceName)
	return mesh.ShouldSkipInboundEntranceSidecar(ctx, wh.kubeClient, app.Spec.Namespace, srrName), nil
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
	injectPolicy, injectWs, injectUpload, injectCallerAgent bool, injectSharedPod *bool, perms []appcfg.ProviderPermission,
	appConfig *appcfg.ApplicationConfig, appMgr *v1alpha1.ApplicationManager, err error) {
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

	appMgr, appConfig, isShared, err = wh.GetAppConfig(namespace)
	if err != nil {
		if errors.Is(err, api.ErrApplicationManagerNotFound) {
			err = nil
		} else {
			klog.Errorf("Failed to get app config err=%v", err)
			return
		}
	}

	if appConfig == nil {
		klog.Infof("Unknown namespace=%s, do not inject", namespace)
		return
	}
	if appConfig.IsMiddleware() {
		return
	}

	if !isShared {
		if appConfig.WsConfig.URL != "" && appConfig.WsConfig.Port > 0 {
			injectWs = true
		}
		if appConfig.Upload.Dest != "" {
			injectUpload = true
		}
		for _, p := range appConfig.Permission {
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
					perms = append(perms, appcfg.ProviderPermission{
						AppName:      providerAppName,
						Namespace:    ns,
						ProviderName: providerName,
					})

				}
			}

		}

		injectPolicy = false
		for _, e := range appConfig.Entrances {
			var isEntrancePod bool
			isEntrancePod, err = wh.isAppEntrancePod(ctx, appConfig.AppName, e.Host, pod, namespace)
			klog.Infof("entranceName=%s isEntrancePod=%v", e.Name, isEntrancePod)
			if err != nil {
				return false, false, false, false, nil, perms, nil, nil, err
			}

			if isEntrancePod {
				skip, skipErr := wh.shouldSkipInboundEntranceSidecar(ctx, appConfig, namespace, e.Name)
				if skipErr != nil {
					return false, false, false, nil, perms, nil, nil, skipErr
				}
				if !skip {
					injectPolicy = true
				}
				break
			}
		}
	} // end of non-shared namespace's pod

	for _, e := range appConfig.SharedEntrances {
		var isEntrancePod bool
		isEntrancePod, err = wh.isAppEntrancePod(ctx, appConfig.AppName, e.Host, pod, namespace)
		klog.Infof("entranceName=%s isEntrancePod=%v", e.Name, isEntrancePod)
		if err != nil {
			return false, false, false, false, nil, perms, nil, nil, err
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

	injectCallerAgent, err = wh.shouldInjectCallerAgent(ctx, appConfig, namespace, isShared)
	if err != nil {
		return false, false, false, false, nil, perms, nil, nil, err
	}

	return
}

func (wh *Webhook) shouldInjectCallerAgent(ctx context.Context, appConfig *appcfg.ApplicationConfig, ns string, isShared bool) (bool, error) {
	if appConfig == nil || isShared {
		return false, nil
	}
	applicationName, err := apputils.FmtAppMgrName(appConfig.AppName, appConfig.OwnerName, ns)
	if err != nil {
		return false, err
	}
	app, err := wh.dynamicClient.AppV1alpha1().Applications().Get(ctx, applicationName, metav1.GetOptions{})
	if err != nil {
		if apierrors.IsNotFound(err) {
			return false, nil
		}
		return false, err
	}
	return calleragent.ShouldInject(app, isShared), nil
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
	appmgr *v1alpha1.ApplicationManager, appConfig *appcfg.ApplicationConfig,
	perms []appcfg.ProviderPermission,
) (string, error) {
	configMapName := fmt.Sprintf("%s-%s", constants.SidecarConfigMapVolumeName, proxyUUID)
	if deployName := utils.GetDeploymentName(pod); deployName != "" {
		configMapName = fmt.Sprintf("%s-%s", constants.SidecarConfigMapVolumeName, deployName)
	}
	cm, e := wh.kubeClient.CoreV1().ConfigMaps(namespace).Get(ctx, configMapName, metav1.GetOptions{})
	if e != nil && !apierrors.IsNotFound(e) {
		return "", e
	}

	permCfg, err := apputils.ProviderPermissionsConvertor(perms).ToPermissionCfg(ctx, appConfig.OwnerName, appcfg.GetMarketSource(appmgr))
	if err != nil {
		klog.Errorf("Failed to convert permissions for app %s: %v", appConfig.AppName, err)
		return "", err
	}

	newConfigMap := sidecar.GetSidecarConfigMap(configMapName, namespace, appConfig, injectPolicy, injectWs, injectUpload, pod, permCfg)
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
func CreatePatchForDeployment(tpl *corev1.PodTemplateSpec, injectAll bool, injectContainer []string, gpuTypeKey string, gpumem *string, envKeyValues []EnvKeyValue) ([]byte, error) {
	patches, err := addGpuResourceLimits(tpl, injectAll, injectContainer, gpuTypeKey, gpumem)
	if err != nil {
		return []byte{}, err
	}
	patches = append(patches, addEnvToPatch(tpl, envKeyValues)...)
	return json.Marshal(patches)
}

// CreateCleanupPatchForDeployment scans the workload template for any
// GPU-related fields that may have been added by a previous run of the
// gpu-limit mutating webhook (nvidia.com/gpu, nvidia.com/gpumem,
// amd.com/gpu, amd.com/apu in both resources.limits and
// resources.requests, plus runtimeClassName="nvidia") and returns an
// RFC 6902 JSON Patch that removes them. Returns nil when nothing needs
// to be cleaned up.
//
// This is the counterpart of addGpuResourceLimits: when an app no longer
// needs GPU (e.g., its OlaresManifest dropped requiredGpu on upgrade),
// the inject path used to early-return without emitting any patch. Helm
// upgrade then preserves the previously-injected GPU keys as "live-only"
// fields via strategic merge, leaving stale resources on the pod. By
// emitting explicit remove ops the desired object that K8s sees no
// longer carries the GPU keys, so 3-way merge can drop them.
func CreateCleanupPatchForDeployment(tpl *corev1.PodTemplateSpec) ([]byte, error) {
	patches := removeGpuResources(tpl)
	if len(patches) == 0 {
		return nil, nil
	}
	return json.Marshal(patches)
}

// gpuResourceKeys lists every extended-resource key the gpu-limit
// webhook may have injected on a prior install. Keep in sync with
// addGpuResourceLimits / getGPUResourceTypeKey.
var gpuResourceKeys = []string{
	constants.NvidiaGPU,
	constants.NvidiaGPUMem,
	constants.AMDGPU,
	constants.AMDAPU,
}

// jsonPointerEscape escapes a string per RFC 6901 so it is safe to use
// as a JSON Pointer path segment. The order matters: "~" must be
// replaced before "/".
func jsonPointerEscape(s string) string {
	s = strings.ReplaceAll(s, "~", "~0")
	s = strings.ReplaceAll(s, "/", "~1")
	return s
}

func removeGpuResources(tpl *corev1.PodTemplateSpec) []patchOp {
	if tpl == nil {
		return nil
	}
	var patches []patchOp

	for i := range tpl.Spec.Containers {
		container := &tpl.Spec.Containers[i]
		for _, key := range gpuResourceKeys {
			resName := corev1.ResourceName(key)
			escaped := jsonPointerEscape(key)
			if _, ok := container.Resources.Limits[resName]; ok {
				patches = append(patches, patchOp{
					Op:   constants.PatchOpRemove,
					Path: fmt.Sprintf("/spec/template/spec/containers/%d/resources/limits/%s", i, escaped),
				})
			}
			if _, ok := container.Resources.Requests[resName]; ok {
				patches = append(patches, patchOp{
					Op:   constants.PatchOpRemove,
					Path: fmt.Sprintf("/spec/template/spec/containers/%d/resources/requests/%s", i, escaped),
				})
			}
		}
	}

	if tpl.Spec.RuntimeClassName != nil && *tpl.Spec.RuntimeClassName == "nvidia" {
		patches = append(patches, patchOp{
			Op:   constants.PatchOpRemove,
			Path: runtimeClassPath,
		})
	}

	return patches
}

func addGpuResourceLimits(tpl *corev1.PodTemplateSpec, injectAll bool, injectContainer []string, typeKey string, gpumem *string) (patch []patchOp, err error) {
	if typeKey == "" {
		klog.Warning("No gpu type selected, skip adding resource limits")
		return patch, nil
	}

	// add runtime class for nvidia gpu, HAMi runtime class is "nvidia"
	if typeKey == constants.NvidiaGPU {
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
		if !injectAll && !funk.Contains(injectContainer, container.Name) {
			continue
		}

		if len(container.Resources.Limits) == 0 {
			limitsValues := map[string]interface{}{
				typeKey: "1",
			}

			if gpumem != nil && *gpumem != "" && typeKey == constants.NvidiaGPU {
				limitsValues[constants.NvidiaGPUMem] = *gpumem
			}

			patch = append(patch, patchOp{
				Op:    constants.PatchOpAdd,
				Path:  fmt.Sprintf(resourcePath, i),
				Value: limitsValues,
			})

		} else {
			t := make(map[string]map[string]string)
			t["limits"] = map[string]string{}
			for k, v := range container.Resources.Limits {
				if k.String() == constants.NvidiaGPU ||
					k.String() == constants.NvidiaGPUMem ||
					k.String() == constants.AMDAPU {
					// unset all previous gpu limits
					continue
				}
				t["limits"][k.String()] = v.String()
			}
			t["limits"][typeKey] = "1"
			if gpumem != nil && *gpumem != "" && typeKey == constants.NvidiaGPU {
				t["limits"][constants.NvidiaGPUMem] = *gpumem
			}
			patch = append(patch, patchOp{
				Op:    constants.PatchOpReplace,
				Path:  fmt.Sprintf(resourcePath, i),
				Value: t["limits"],
			})
		}
	}

	return patch, nil
}

func addEnvToPatch(tpl *corev1.PodTemplateSpec, envKeyValues []EnvKeyValue) (patch []patchOp) {
	for i := range tpl.Spec.Containers {
		container := tpl.Spec.Containers[i]

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

	return patch
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
	_, appConfig, isShared, err := wh.GetAppConfig(namespace)
	if err != nil {
		klog.Errorf("Failed to get app config err=%v", err)
		return "", "", err
	}

	if isShared {
		// shared namespace, no need to get appkey/secret
		return "", "", nil
	}

	apClient := provider.NewApplicationPermissionRequest(client)
	ap, err := apClient.Get(context.TODO(), "user-system-"+appConfig.OwnerName, appConfig.AppName, metav1.GetOptions{})
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

// MacvlanInitContainerName is the name of the init container injected for pods
// that need to reply via eth0 in macvlan setups.
const MacvlanInitContainerName = "macvlan-reply-via-eth0"

// macvlanInitScript is the shell script run inside the macvlan-init container.
// It waits for an IPv4 address on eth0, then creates a dedicated routing table
// and a `from <pod-ip>` rule so that reply traffic is sent out via eth0
// instead of the default pod gateway.
const macvlanInitScript = `set -eu
TABLE=100
PRI=100
POD_IP=""
i=0
while [ "$i" -lt 30 ]; do
  POD_IP=$(ip -4 addr show dev eth0 2>/dev/null | awk '/inet /{print $2}' | cut -d/ -f1 | head -1)
  test -n "$POD_IP" && break
  i=$((i + 1))
  sleep 1
done
test -n "$POD_IP" || { echo "no eth0 address after wait"; exit 1; }
GW=$(ip -4 route show dev eth0 | awk '/^default/{print $3; exit}')
test -n "${GW:-}" || GW=169.254.1.1
ip -4 route replace default via "$GW" dev eth0 table "$TABLE"
if ip -4 rule list | grep -Fq "from $POD_IP lookup $TABLE"; then exit 0; fi
ip -4 rule add from "$POD_IP/32" lookup "$TABLE" priority "$PRI"
`

// GetMacvlanInitContainer returns the init container spec used to set up
// a dedicated routing table so that reply traffic flows back via eth0 for
// pods participating in a macvlan / overlay-gateway setup.
func GetMacvlanInitContainer() corev1.Container {
	runAsNonRoot := false
	allowPrivilegeEscalation := false
	runAsUser := int64(0)
	return corev1.Container{
		Name:            MacvlanInitContainerName,
		Image:           "docker.io/beclab/aboveos-busybox:1.37.0",
		ImagePullPolicy: corev1.PullIfNotPresent,
		SecurityContext: &corev1.SecurityContext{
			RunAsUser:                &runAsUser,
			RunAsNonRoot:             &runAsNonRoot,
			AllowPrivilegeEscalation: &allowPrivilegeEscalation,
			Capabilities: &corev1.Capabilities{
				Drop: []corev1.Capability{"ALL"},
				Add:  []corev1.Capability{"NET_ADMIN"},
			},
		},
		Command:                  []string{"sh", "-c", macvlanInitScript},
		Resources:                corev1.ResourceRequirements{},
		TerminationMessagePath:   "/dev/termination-log",
		TerminationMessagePolicy: corev1.TerminationMessageReadFile,
	}
}

// ShouldInjectMacvlanInit reports whether the macvlan init container should be
// injected for the given pod. It returns true only when the owning Application
// can be resolved from the pod's app name/owner labels and has
// `spec.settings.enableOverlayGateway == "true"`.
func (wh *Webhook) ShouldInjectMacvlanInit(ctx context.Context, pod *corev1.Pod, ns string) (bool, error) {
	if pod == nil || pod.Labels == nil {
		return false, nil
	}
	if pod.Labels[constants.ApplicationMacvlanInitLabel] != "true" {
		return false, nil
	}
	appName := pod.Labels[constants.ApplicationNameLabel]
	owner := pod.Labels[constants.ApplicationOwnerLabel]
	if appName == "" {
		klog.Infof("macvlan-init: skip pod=%s/%s missing app labels", ns, pod.Name)
		return false, nil
	}
	klog.Infof("ShouldInjectMacvlanInit: pod.Namespace: %s", ns)
	applicationName, err := apputils.FmtAppMgrName(appName, owner, ns)
	if err != nil {
		klog.Errorf("macvlan-init: failed to format application name app=%s owner=%s ns=%s err=%v", appName, owner, ns, err)
		return false, err
	}
	app, err := wh.dynamicClient.AppV1alpha1().Applications().Get(ctx, applicationName, metav1.GetOptions{})
	if err != nil {
		if apierrors.IsNotFound(err) {
			klog.Infof("macvlan-init: application=%s not found for pod=%s/%s", applicationName, ns, pod.Name)
			return false, nil
		}
		klog.Errorf("macvlan-init: failed to get application=%s err=%v", applicationName, err)
		return false, err
	}
	enabled := app.Spec.Settings["enableOverlayGateway"] == "true"
	if !enabled {
		klog.Infof("macvlan-init: application=%s enableOverlayGateway is not true, skip pod=%s/%s", applicationName, ns, pod.Name)
	}
	return enabled, nil
}

// CreateMacvlanInitPatch appends the macvlan init container to the pod's
// init containers (idempotent — does nothing if the container is already
// present) and returns the JSON merge patch to send back in the admission
// response.
func (wh *Webhook) CreateMacvlanInitPatch(req *admissionv1.AdmissionRequest, pod *corev1.Pod) ([]byte, error) {
	if pod.Annotations == nil {
		pod.Annotations = make(map[string]string)
	}
	pod.Annotations["k8s.v1.cni.cncf.io/networks"] = "kube-system/underlay-macvlan"

	for _, c := range pod.Spec.InitContainers {
		if c.Name == MacvlanInitContainerName {
			klog.Infof("macvlan-init: container already present in pod=%s/%s, skip", pod.Namespace, pod.Name)
			return makePatches(req, pod)
		}
	}
	// Append after any existing init containers (e.g. sidecar wait-for /
	// render-envoy-config) so we run after them but still before the main
	// app containers.
	pod.Spec.InitContainers = append(pod.Spec.InitContainers, GetMacvlanInitContainer())
	return makePatches(req, pod)
}
