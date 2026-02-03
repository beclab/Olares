package constants

import (
	"flag"
	"os"

	"k8s.io/apimachinery/pkg/util/sets"
)

const (
	KubeSphereAPIScheme                = "http"
	ApplicationDefaultThirdLevelDomain = "applications.app.bytetrade.io/default-thirdlevel-domains"
	ApplicationNameLabel               = "applications.app.bytetrade.io/name"
	ApplicationRawAppNameLabel         = "applications.app.bytetrade.io/raw-app-name"
	ApplicationAppGroupLabel           = "applications.app.bytetrade.io/group"
	ApplicationAuthorLabel             = "applications.app.bytetrade.io/author"
	ApplicationOwnerLabel              = "applications.app.bytetrade.io/owner"
	ApplicationMiddlewareLabel         = "applications.app.bytetrade.io/middleware"
	ApplicationIconLabel               = "applications.app.bytetrade.io/icon"
	ApplicationEntrancesKey            = "applications.app.bytetrade.io/entrances"
	ApplicationPortsKey                = "applications.app.bytetrade.io/ports"
	ApplicationSystemServiceLabel      = "applications.app.bytetrade.io/system_service"
	ApplicationTitleLabel              = "applications.app.bytetrade.io/title"
	ApplicationImageLabel              = "applications.app.bytetrade.io/images"
	ApplicationTargetLabel             = "applications.app.bytetrade.io/target"
	ApplicationRunAsUserLabel          = "applications.apps.bytetrade.io/runasuser"
	ApplicationVersionLabel            = "applications.app.bytetrade.io/version"
	ApplicationSourceLabel             = "applications.app.bytetrade.io/source"
	ApplicationTailScaleKey            = "applications.app.bytetrade.io/tailscale"
	ApplicationRequiredGPU             = "applications.app.bytetrade.io/required_gpu"
	AppPodGPUConsumePolicy             = "gpu.bytetrade.io/app-pod-consume-policy"
	ApplicationPolicies                = "applications.app.bytetrade.io/policies"
	ApplicationMobileSupported         = "applications.app.bytetrade.io/mobile_supported"
	ApplicationClusterDep              = "applications.app.bytetrade.io/need_cluster_scoped_app"
	ApplicationGroupClusterDep         = "applications.app.bytetrade.io/need_cluster_scoped_group"
	UserContextAttribute               = "username"
	KubeSphereClientAttribute          = "ksclient"
	MarketSource                       = "X-Market-Source"
	MarketUser                         = "X-Market-User"
	StudioSource                       = "devbox"
	ApplicationInstallUserLabel        = "applications.app.bytetrade.io/install_user"
	BflUserKey                         = "X-Bfl-User"

	InstanceIDLabel         = "workflows.argoproj.io/controller-instanceid"
	WorkflowOwnerLabel      = "workflows.app.bytetrade.io/owner"
	WorkflowNameLabel       = "workflows.app.bytetrade.io/name"
	WorkflowTitleAnnotation = "workflows.app.bytetrade.io/title"

	OwnerNamespacePrefix = "user-space"
	OwnerNamespaceTempl  = "%s-%s"
	UserSpaceDirPVC      = "userspace-dir"

	UserAppDataDirPVC = "appcache-dir"

	UserChartsPath = "./userapps"

	EnvoyUID                        int64 = 1555
	DefaultEnvoyLogLevel                  = "debug"
	EnvoyImageVersion                     = "beclab/envoy:v1.25.11.1"
	EnvoyContainerName                    = "olares-envoy-sidecar"
	EnvoyAdminPort                        = 15000
	EnvoyAdminPortName                    = "proxy-admin"
	EnvoyInboundListenerPort              = 15003
	EnvoyInboundListenerPortName          = "proxy-inbound"
	EnvoyOutboundListenerPort             = 15001
	EnvoyOutboundListenerPortName         = "proxy-outbound"
	EnvoyLivenessProbePort                = 15008
	EnvoyConfigFileName                   = "envoy.yaml"
	EnvoyConfigFilePath                   = "/config"
	EnvoyConfigOnlyOutBoundFileName       = "envoy2.yaml"
	WsContainerName                       = "olares-ws-sidecar"
	WsContainerImage                      = "WS_CONTAINER_IMAGE"

	UploadContainerName  = "olares-upload-sidecar"
	UploadContainerImage = "UPLOAD_CONTAINER_IMAGE"

	SidecarConfigMapVolumeName = "olares-sidecar-config"
	SidecarInitContainerName   = "olares-sidecar-init"
	EnvoyConfigWorkDirName     = "envoy-config"

	ByteTradeAuthor         = "bytetrade.io"
	NvshareGPU              = "nvshare.com/gpu"
	NvidiaGPU               = "nvidia.com/gpu"
	VirtAiTechVGPU          = "virtaitech.com/gpu"
	PatchOpAdd              = "add"
	PatchOpReplace          = "replace"
	EnvNvshareManagedMemory = "NVSHARE_MANAGED_MEMORY"

	AuthorizationLevelOfPublic  = "public"
	AuthorizationLevelOfPrivate = "private"

	DependencyTypeSystem = "system"
	DependencyTypeApp    = "application"
	AppCacheDirURL       = "http://files-service.os-framework/api/resources/cache/%s/"
	AppDataDirURL        = "http://files-service.os-framework/api/resources/drive/Data/"

	UserSpaceDirKey   = "userspace_hostpath"
	UserAppDataDirKey = "appcache_hostpath"

	OIDCSecret = "oidc-secret"

	AppMarketSourceKey = "bytetrade.io/market-source"

	// EnvRefStatus* constants for AppEnvVar.ValueFrom.Status (used for both SystemEnv and UserEnv references)
	EnvRefStatusPending  = "pending"
	EnvRefStatusSynced   = "synced"
	EnvRefStatusNotFound = "notfound"

	OlaresEnvHelmValuesKey = "olaresEnv"
	SystemEnvHelmValuesKey = "system"
	AppEnvHelmValuesKey    = "app"

	// AppEnvSyncAnnotation triggers AppEnvController to sync environment values from SystemEnv or UserEnv changes
	AppEnvSyncAnnotation = "appenv.bytetrade.io/sync-triggered-by"

	AppForceUninstall      = "ForceUninstall"
	AppForceUninstalled    = "ForceUninstalled"
	AppUnschedulable       = "Unschedulable"
	AppStopByUser          = "StopByUser"
	AppStopDueToInitFailed = "InitFailed"
	AppStopDueToEvicted    = "Evicted"

	AppSharedEntrancesLabel = "app.bytetrade.io/shared-entrance"
	AppMiddlewareLabel      = "app.bytetrade.io/middleware"

	OneContainerMultiDeviceSplitSymbol = ":"
	ArchLabelKey                       = "kubernetes.io/arch"
	CudaVersionLabelKey                = "gpu.bytetrade.io/cuda"
	NodeNvidiaRegistryKey              = "hami.io/node-nvidia-register"
)

var (
	empty = sets.Empty{}
	// Sources represents the source of the application.
	Sources = sets.String{"market": empty, "custom": empty, "devbox": empty, "system": empty, "unknown": empty}
	// ResourceTypes represents the type of application system supported.
	ResourceTypes = sets.String{"app": empty, "recommend": empty, "model": empty, "agent": empty, "middleware": empty}
)

var (
	// APIServerListenAddress server address for api server.
	APIServerListenAddress = ":6755"
	// WebhookServerListenAddress server address for webhook server.
	WebhookServerListenAddress = ":8433"
	// KubeSphereAPIHost kubesphere api host.
	KubeSphereAPIHost string

	CHART_REPO_URL string = "http://chart-repo-service.os-framework:82/"

	OLARES_APP_NAME = "olares-app"
)

type ResourceConditionType string

const (
	DiskPressure             ResourceConditionType = "DiskPressure"
	SystemCPUPressure        ResourceConditionType = "SystemCPUPressure"
	SystemMemoryPressure     ResourceConditionType = "SystemMemoryPressure"
	SystemGPUNotAvailable    ResourceConditionType = "SystemGPUNotAvailable"
	SystemGPUPressure        ResourceConditionType = "SystemGPUPressure"
	K8sRequestCPUPressure    ResourceConditionType = "K8sReqeustCPUPressure"
	K8sRequestMemoryPressure ResourceConditionType = "K8sRequestMemoryPressure"
	UserCPUPressure          ResourceConditionType = "UserCPUPressure"
	UserMemoryPressure       ResourceConditionType = "UserMemoryPressure"

	DiskPressureMessage             string = "Insufficient disk space. Unable to %s the application. Please stop other running applications to free up storage."
	SystemCPUPressureMessage        string = "Insufficient system CPU. Unable to %s the application. Please stop other running applications to free up resources."
	SystemMemoryPressureMessage     string = "Insufficient system memory. Unable to %s the application. Please stop other running applications to free up memory."
	SystemGPUNotAvailableMessage    string = "No available GPU found. Unable to %s the application."
	SystemGPUPressureMessage        string = "Available GPU is insufficient to %s this application. The requested GPU memory cannot exceed the maximum GPU memory of the node."
	K8sRequestCPUPressureMessage    string = "Available CPU is insufficient to %s this application. Please stop other applications to free up resources."
	K8sRequestMemoryPressureMessage string = "Available memory is insufficient to %s this application. Please stop other applications to free up resources."
	UserCPUPressureMessage          string = "Insufficient user CPU. Unable to %s the application. Please stop other running applications to free up resources."
	UserMemoryPressureMessage       string = "Insufficient user memory. Unable to %s the application. Please stop other running applications to free up memory."
)

func (rct ResourceConditionType) String() string {
	return string(rct)
}

type ResourceType string

const (
	Disk   ResourceType = "disk"
	CPU    ResourceType = "cpu"
	Memory ResourceType = "memory"
	GPU    ResourceType = "gpu"
)

func (rt ResourceType) String() string {
	return string(rt)
}

func init() {
	flag.StringVar(&APIServerListenAddress, "listen", ":6755",
		"app-service listening address")
	flag.StringVar(&WebhookServerListenAddress, "webhook-listen", ":8433",
		"webhook listening address")
	flag.StringVar(&KubeSphereAPIHost, "ks-apiserver", "ks-apiserver.kubesphere-system",
		"kubesphere api server")

	url := os.Getenv("CHART_REPO_URL")
	if url != "" {
		CHART_REPO_URL = url
	}
}
