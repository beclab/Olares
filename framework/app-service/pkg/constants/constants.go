package constants

import (
	"flag"
	"os"
	"strings"
	"time"

	"k8s.io/apimachinery/pkg/util/sets"
)

const (
	KubeSphereAPIScheme                = "http"
	ApplicationDefaultThirdLevelDomain = "applications.app.bytetrade.io/default-thirdlevel-domains"
	ApplicationNameLabel               = "applications.app.bytetrade.io/name"
	ApplicationMacvlanInitLabel        = "applications.app.bytetrade.io/macvlan-init"
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

	ByteTradeAuthor = "bytetrade.io"
	PatchOpAdd      = "add"
	PatchOpReplace  = "replace"
	PatchOpRemove   = "remove"
	EnvGPUType      = "GPU_TYPE"

	// gpu resource keys
	NvidiaGPU    = "nvidia.com/gpu"
	NvidiaGPUMem = "nvidia.com/gpumem"
	//	NvidiaGB10GPU = "nvidia.com/gb10"
	AMDAPU   = "amd.com/apu"
	AMDGPU   = "amd.com/gpu"
	IntelGPU = "gpu.intel.com/i915"

	AuthorizationLevelOfPublic   = "public"
	AuthorizationLevelOfPrivate  = "private"
	AuthorizationLevelOfInternal = "internal"

	DependencyTypeSystem = "system"
	DependencyTypeApp    = "application"
	AppCacheDirURL       = "http://files-service.os-framework/api/resources/cache/%s/"
	AppDataDirURL        = "http://files-service.os-framework/api/resources/drive/Data/"

	UserSpaceDirKey   = "userspace_hostpath"
	UserAppDataDirKey = "appcache_hostpath"

	OIDCSecret = "oidc-secret"

	AppMarketSourceKey = "bytetrade.io/market-source"

	// AppChartOwnerKey records the user who uploaded the chart the app was
	// installed from. It is only meaningful for uploaded (non-market) apps;
	// market apps leave it empty and fall back to the installing user. Stamped
	// at install time and read when building push events (chartOwner field).
	AppChartOwnerKey = "app.bytetrade.io/chart-owner"

	// EnvRefStatus* constants for AppEnvVar.ValueFrom.Status (used for both SystemEnv and UserEnv references)
	EnvRefStatusPending  = "pending"
	EnvRefStatusSynced   = "synced"
	EnvRefStatusNotFound = "notfound"

	OlaresEnvHelmValuesKey = "olaresEnv"
	SystemEnvHelmValuesKey = "system"
	AppEnvHelmValuesKey    = "app"

	// AppEnvSyncAnnotation triggers AppEnvController to sync environment values from SystemEnv or UserEnv changes
	AppEnvSyncAnnotation = "appenv.bytetrade.io/sync-triggered-by"

	AppForceUninstall         = "ForceUninstall"
	AppForceUninstalled       = "ForceUninstalled"
	AppUnschedulable          = "Unschedulable"
	AppHamiSchedulable        = "HamiUnschedulable"
	AppStopByUser             = "StopByUser"
	AppStopDueToInitFailed    = "InitFailed"
	AppStopDueToStartUpFailed = "StartUpFailed"
	AppStopDueToEvicted       = "Evicted"

	AppSharedEntrancesLabel = "app.bytetrade.io/shared-entrance"
	AppMiddlewareLabel      = "app.bytetrade.io/middleware"

	// AppApiVersionLabel marks an Application / ApplicationManager / namespace
	// / workload with the OlaresManifest schema version (currently only v3 is
	// stamped). This is a SCHEMA marker — a v3 app may be either a shared
	// cluster-wide singleton or a regular per-user app, depending on
	// options.shared in the manifest. Use AppSharedLabel to discriminate
	// between those two; api-version=v3 alone does NOT imply shared.
	AppApiVersionLabel = "app.bytetrade.io/api-version"
	AppVersionV3       = "v3"

	// AppSharedLabel marks an Application / ApplicationManager / namespace /
	// workload as a shared cluster-wide app. Stamped at install time by the
	// v3 install handler when ApplicationConfig.Shared is true (i.e.
	// apiVersion: v3 + options.shared: true) and propagated by the
	// Application controller. Drives admin-only lifecycle, cluster-wide
	// visibility, NATS fan-out, NetworkPolicy fast-path, etc. Per-user v3
	// apps do NOT carry this label and are handled like v1 apps.
	AppSharedLabel        = "app.bytetrade.io/app-shared"
	AppSharedTrue         = "true"
	AppClonedFromKey      = "app.bytetrade.io/app-cloned-from"
	AppClonedFromTemplate = "template"
	AppClonedFromApp      = "app"

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

	// RemoteOptionsDomainWhitelist restricts which domains the remote-options
	// proxy (proxyRemoteOptions) may fetch from. The proxy lets any normal user
	// have app-service issue an outbound GET, so without a whitelist it becomes
	// an open proxy / SSRF vector to arbitrary URLs. Validating only against the
	// remoteOptions URL declared in the app manifest is insufficient, because a
	// user can bypass that by uploading a custom chart, so we gate on an
	// explicit domain whitelist instead. Currently only the Olares app CDN is
	// used, but it is kept as a list for future entries. Override at runtime
	// with the comma-separated REMOTE_OPTIONS_DOMAIN_WHITELIST env var.
	RemoteOptionsDomainWhitelist = []string{"app.cdn.olares.com"}
)

// IsRemoteOptionsHostAllowed reports whether host is permitted by the
// remote-options domain whitelist. A whitelist entry matches the exact host or
// any of its subdomains (both compared case-insensitively). An empty whitelist
// denies everything.
func IsRemoteOptionsHostAllowed(host string) bool {
	host = strings.ToLower(strings.TrimSuffix(strings.TrimSpace(host), "."))
	if host == "" {
		return false
	}
	for _, d := range RemoteOptionsDomainWhitelist {
		d = strings.ToLower(strings.TrimSuffix(strings.TrimSpace(d), "."))
		if d == "" {
			continue
		}
		if host == d || strings.HasSuffix(host, "."+d) {
			return true
		}
	}
	return false
}

// AppMgrTerminalRetention bounds how long an ApplicationManager is allowed to
// linger in a safely-deletable terminal state (Uninstalled / InstallingCanceled
// / PendingCanceled / DownloadingCanceled / DownloadFailed / InstallFailed)
// before the GC controller reclaims it. The retention gives operators a window
// to inspect the failure reason / op record and gives the install-failure
// cleanup helper plenty of time to converge the rare NS-finalizer-stuck case
// before the AM disappears. 60min is conservative enough for both; the
// retention can be overridden at runtime via the APPMGR_TERMINAL_RETENTION
// env var on the controller pod.
var AppMgrTerminalRetention = func() time.Duration {
	if v := os.Getenv("APPMGR_TERMINAL_RETENTION"); v != "" {
		if d, err := time.ParseDuration(v); err == nil && d > 0 {
			return d
		}
	}
	return 60 * time.Minute
}()

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
	// Cluster*Insufficient are emitted by clusterCapacityValidator and
	// indicate the app's declared requirement exceeds the cluster's
	// total (schedulable) capacity, irrespective of current usage. This
	// is the "the cluster physically cannot host this app" case and
	// should be distinguished from K8sRequest*Pressure (which means
	// "current allocatable - already-scheduled requests is too small").
	ClusterCPUInsufficient    ResourceConditionType = "ClusterCPUInsufficient"
	ClusterMemoryInsufficient ResourceConditionType = "ClusterMemoryInsufficient"
	ClusterDiskInsufficient   ResourceConditionType = "ClusterDiskInsufficient"

	// ComputeAllocationFailed is emitted by computeAllocationValidator
	// when compute.AllocateForInstall cannot place the app's
	// workloads on any node under the selected GPU mode (or the
	// underlying scheduler returns an error). The downstream install
	// path uses this to land the app in `stopped` with
	// AppUnschedulable as the user-facing reason.
	ComputeAllocationFailed ResourceConditionType = "ComputeAllocationFailed"
	// ComputeModeUnavailable is emitted by computeModeValidator when the
	// app's selected GPU/compute mode has no runnable placement on the
	// cluster (compute.AppInstallable returned false).
	ComputeModeUnavailable ResourceConditionType = "ComputeModeUnavailable"
	// NodePressure is emitted by nodePressureValidator when no single
	// node can take the app's CPU/memory request while staying under the
	// pressure threshold.
	NodePressure ResourceConditionType = "NodePressure"

	DiskPressureMessage              string = "Insufficient disk space. Unable to %s the application. Please stop other running applications to free up storage."
	SystemCPUPressureMessage         string = "Insufficient system CPU. Unable to %s the application. Please stop other running applications to free up resources."
	SystemMemoryPressureMessage      string = "Insufficient system memory. Unable to %s the application. Please stop other running applications to free up memory."
	SystemGPUNotAvailableMessage     string = "No available GPU found. Unable to %s the application."
	SystemGPUPressureMessage         string = "Available GPU is insufficient to %s this application. The requested GPU memory cannot exceed the maximum GPU memory of the node."
	K8sRequestCPUPressureMessage     string = "Available CPU is insufficient to %s this application. Please stop other applications to free up resources."
	K8sRequestMemoryPressureMessage  string = "Available memory is insufficient to %s this application. Please stop other applications to free up resources."
	UserCPUPressureMessage           string = "Insufficient user CPU. Unable to %s the application. Please stop other running applications to free up resources."
	UserMemoryPressureMessage        string = "Insufficient user memory. Unable to %s the application. Please stop other running applications to free up memory."
	ClusterCPUInsufficientMessage    string = "Cluster total schedulable CPU is smaller than the app's request. Unable to %s this application."
	ClusterMemoryInsufficientMessage string = "Cluster total schedulable memory is smaller than the app's request. Unable to %s this application."
	ClusterDiskInsufficientMessage   string = "Cluster total schedulable ephemeral storage is smaller than the app's request. Unable to %s this application."
)

func (rct ResourceConditionType) String() string {
	return string(rct)
}

type ResourceType string

const (
	Disk     ResourceType = "disk"
	CPU      ResourceType = "cpu"
	Memory   ResourceType = "memory"
	GPU      ResourceType = "gpu"
	Hardware ResourceType = "hardware"
	Compute  ResourceType = "compute"
	Node     ResourceType = "node"
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

	if wl := os.Getenv("REMOTE_OPTIONS_DOMAIN_WHITELIST"); wl != "" {
		domains := make([]string, 0)
		for _, d := range strings.Split(wl, ",") {
			if d = strings.TrimSpace(d); d != "" {
				domains = append(domains, d)
			}
		}
		if len(domains) > 0 {
			RemoteOptionsDomainWhitelist = domains
		}
	}
}
