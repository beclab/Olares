package appcfg

import (
	"fmt"
	"path/filepath"

	"bytetrade.io/web3os/app-service/api/app.bytetrade.io/v1alpha1"
	sysv1alpha1 "bytetrade.io/web3os/app-service/api/sys.bytetrade.io/v1alpha1"
	"bytetrade.io/web3os/app-service/pkg/tapr"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// TODO: share the structs in projects
type AppMetaData struct {
	Name        string   `yaml:"name" json:"name"`
	Icon        string   `yaml:"icon" json:"icon"`
	Description string   `yaml:"description" json:"description"`
	AppID       string   `yaml:"appid" json:"appid"`
	Title       string   `yaml:"title" json:"title"`
	Version     string   `yaml:"version" json:"version"`
	Categories  []string `yaml:"categories" json:"categories"`
	Rating      float32  `yaml:"rating" json:"rating"`
	Target      string   `yaml:"target" json:"target"`
	Type        string   `yaml:"type" json:"type"`
}

type AppConfiguration struct {
	ConfigVersion string                  `yaml:"olaresManifest.version" json:"olaresManifest.version"`
	ConfigType    string                  `yaml:"olaresManifest.type" json:"olaresManifest.type"`
	APIVersion    string                  `yaml:"apiVersion" json:"apiVersion"`
	Metadata      AppMetaData             `yaml:"metadata" json:"metadata"`
	Entrances     []v1alpha1.Entrance     `yaml:"entrances" json:"entrances"`
	Ports         []v1alpha1.ServicePort  `yaml:"ports" json:"ports"`
	TailScale     v1alpha1.TailScale      `yaml:"tailscale" json:"tailscale"`
	Spec          AppSpec                 `yaml:"spec" json:"spec"`
	Permission    Permission              `yaml:"permission" json:"permission" description:"app permission request"`
	Middleware    *tapr.Middleware        `yaml:"middleware" json:"middleware" description:"app middleware request"`
	Options       Options                 `yaml:"options" json:"options" description:"app options"`
	Provider      []Provider              `yaml:"provider,omitempty" json:"provider,omitempty" description:"app provider information"`
	Envs          []sysv1alpha1.AppEnvVar `yaml:"envs,omitempty" json:"envs,omitempty"`

	// Only for v2 c/s apps to share the api to other cluster scope apps
	SharedEntrances []v1alpha1.Entrance `yaml:"sharedEntrances,omitempty" json:"sharedEntrances,omitempty"`
}

type AppSpec struct {
	Namespace           string        `yaml:"namespace,omitempty" json:"namespace,omitempty"`
	OnlyAdmin           bool          `yaml:"onlyAdmin,omitempty" json:"onlyAdmin,omitempty"`
	VersionName         string        `yaml:"versionName" json:"versionName"`
	FullDescription     string        `yaml:"fullDescription" json:"fullDescription"`
	UpgradeDescription  string        `yaml:"upgradeDescription" json:"upgradeDescription"`
	PromoteImage        []string      `yaml:"promoteImage" json:"promoteImage"`
	PromoteVideo        string        `yaml:"promoteVideo" json:"promoteVideo"`
	SubCategory         string        `yaml:"subCategory" json:"subCategory"`
	Developer           string        `yaml:"developer" json:"developer"`
	RequiredMemory      string        `yaml:"requiredMemory" json:"requiredMemory"`
	RequiredDisk        string        `yaml:"requiredDisk" json:"requiredDisk"`
	SupportClient       SupportClient `yaml:"supportClient" json:"supportClient"`
	RequiredGPU         string        `yaml:"requiredGpu" json:"requiredGpu"`
	RequiredCPU         string        `yaml:"requiredCpu" json:"requiredCpu"`
	RunAsUser           bool          `yaml:"runAsUser" json:"runAsUser"`
	RunAsInternal       bool          `yaml:"runAsInternal" json:"runAsInternal"`
	PodGPUConsumePolicy string        `yaml:"podGpuConsumePolicy" json:"podGpuConsumePolicy"`
	SubCharts           []Chart       `yaml:"subCharts" json:"subCharts"`
	Hardware            Hardware      `yaml:"hardware" json:"hardware"`
}

type Hardware struct {
	Cpu CpuConfig `yaml:"cpu" json:"cpu"`
	Gpu GpuConfig `yaml:"gpu" json:"gpu"`
}

type CpuConfig struct {
	Vendor string `yaml:"vendor" json:"vendor"`
	Arch   string `yaml:"arch" json:"arch"`
}
type GpuConfig struct {
	Vendor string   `yaml:"vendor" json:"vendor"`
	Arch   []string `yaml:"arch" json:"arch"`
	// SingleMemory specifies the minimum memory size required for a single GPU
	SingleMemory string `yaml:"singleMemory" json:"singleMemory"`
	// TotalMemory specifies the total GPU memory required across all GPUs within one node
	TotalMemory string `yaml:"totalMemory" json:"totalMemory"`
}

type SupportClient struct {
	Edge    string `yaml:"edge" json:"edge"`
	Android string `yaml:"android" json:"android"`
	Ios     string `yaml:"ios" json:"ios"`
	Windows string `yaml:"windows" json:"windows"`
	Mac     string `yaml:"mac" json:"mac"`
	Linux   string `yaml:"linux" json:"linux"`
}

type Permission struct {
	AppData        bool                 `yaml:"appData" json:"appData"  description:"app data permission for writing"`
	AppCache       bool                 `yaml:"appCache" json:"appCache"`
	UserData       []string             `yaml:"userData" json:"userData"`
	Provider       []ProviderPermission `yaml:"provider" json:"provider"  description:"system shared data permission for accessing"`
	ServiceAccount *string              `yaml:"serviceAccount,omitempty" json:"serviceAccount,omitempty" description:"service account for app permission"`
}

type ProviderPermission struct {
	AppName      string                 `yaml:"appName" json:"appName"`
	Namespace    string                 `yaml:"namespace,omitempty" json:"namespace,omitempty"`
	ProviderName string                 `yaml:"providerName" json:"providerName"`
	PodSelectors []metav1.LabelSelector `yaml:"podSelectors" json:"podSelectors"`
}

type Policy struct {
	EntranceName string `yaml:"entranceName" json:"entranceName"`
	Description  string `yaml:"description" json:"description" description:"the description of the policy"`
	URIRegex     string `yaml:"uriRegex" json:"uriRegex" description:"uri regular expression"`
	Level        string `yaml:"level" json:"level"`
	OneTime      bool   `yaml:"oneTime" json:"oneTime"`
	Duration     string `yaml:"validDuration" json:"validDuration"`
}

type Dependency struct {
	Name    string `yaml:"name" json:"name"`
	Version string `yaml:"version" json:"version"`
	// dependency type: system, application.
	Type      string `yaml:"type" json:"type"`
	Mandatory bool   `yaml:"mandatory" json:"mandatory"`
	SelfRely  bool   `yaml:"selfRely" json:"selfRely"`
}

type Conflict struct {
	Name string `yaml:"name" json:"name"`
	// conflict type: application
	Type string `yaml:"type" json:"type"`
}

type Options struct {
	MobileSupported      bool                     `yaml:"mobileSupported" json:"mobileSupported"`
	Policies             []Policy                 `yaml:"policies" json:"policies"`
	ResetCookie          ResetCookie              `yaml:"resetCookie" json:"resetCookie"`
	Dependencies         []Dependency             `yaml:"dependencies" json:"dependencies"`
	Conflicts            []Conflict               `yaml:"conflicts" json:"conflicts"`
	AppScope             AppScope                 `yaml:"appScope" json:"appScope"`
	WsConfig             WsConfig                 `yaml:"websocket" json:"websocket"`
	Upload               Upload                   `yaml:"upload" json:"upload"`
	SyncProvider         []map[string]interface{} `yaml:"syncProvider" json:"syncProvider"`
	OIDC                 OIDC                     `yaml:"oidc" json:"oidc"`
	ApiTimeout           *int64                   `yaml:"apiTimeout" json:"apiTimeout"`
	AllowedOutboundPorts []int                    `yaml:"allowedOutboundPorts" json:"AllowedOutboundPorts"`
	Images               []string                 `yaml:"images" json:"images"`
	AllowMultipleInstall bool                     `yaml:"allowMultipleInstall" json:"allowMultipleInstall"`
}

type ResetCookie struct {
	Enabled bool `yaml:"enabled" json:"enabled"`
}

type AppScope struct {
	ClusterScoped bool     `yaml:"clusterScoped" json:"clusterScoped"`
	AppRef        []string `yaml:"appRef" json:"appRef"`
	SystemService bool     `yaml:"systemService" json:"systemService"`
}

type WsConfig struct {
	Port int    `yaml:"port" json:"port"`
	URL  string `yaml:"url" json:"url"`
}

type Upload struct {
	FileType    []string `yaml:"fileType" json:"fileType"`
	Dest        string   `yaml:"dest" json:"dest"`
	LimitedSize int      `yaml:"limitedSize" json:"limitedSize"`
}

type OIDC struct {
	Enabled      bool   `yaml:"enabled" json:"enabled"`
	RedirectUri  string `yaml:"redirectUri" json:"redirectUri"`
	EntranceName string `yaml:"entranceName" json:"entranceName"`
}

type Chart struct {
	Name   string `yaml:"name" json:"name"`
	Shared bool   `yaml:"shared" json:"shared"`
}

type Provider struct {
	Name     string   `yaml:"name" json:"name"`
	Entrance string   `yaml:"entrance" json:"entrance"`
	Paths    []string `yaml:"paths" json:"paths"`
	Verbs    []string `yaml:"verbs" json:"verbs"`
}

func (c *Chart) Namespace(owner string) string {
	if c.Shared {
		return fmt.Sprintf("%s-%s", c.Name, "shared")
	}
	return fmt.Sprintf("%s-%s", c.Name, owner)
}

func (c *Chart) ChartPath(appName string) string {
	return AppChartPath(filepath.Join(appName, c.Name))
}
