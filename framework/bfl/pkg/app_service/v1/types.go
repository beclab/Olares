package app_service

import (
	"time"

	appv1 "bytetrade.io/web3os/bfl/internal/ingress/api/app.bytetrade.io/v1alpha1"

	k8sv1 "k8s.io/api/apps/v1"
)

type AppInfo struct {
	ID                            string        `json:"id"`
	Name                          string        `json:"name"`
	Namespace                     string        `json:"namespace"`
	DeploymentName                string        `json:"deployment"`
	Owner                         string        `json:"owner"`
	URL                           string        `json:"url"`
	Icon                          string        `json:"icon"`
	Title                         string        `json:"title"`
	Target                        string        `json:"target"`
	Entrances                     []Entrance    `json:"entrances"`
	Ports                         []ServicePort `json:"ports"`
	TailScaleACLs                 []ACL         `json:"tailscaleAcls,omitempty"`
	State                         string        `json:"state"`
	IsSysApp                      bool          `json:"isSysApp"`
	RequiredGpu                   string        `json:"requiredGpu"`
	IsClusterScoped               bool          `json:"isClusterScoped"`
	MobileSupported               bool          `json:"mobileSupported"`
	DefaultThirdLevelDomainConfig string        `json:"defaultThirdLevelDomainConfig"`
	SharedEntrances               []Entrance    `json:"sharedEntrances,omitempty"`
}

type Entrance struct {
	ID         string `json:"id"`
	Name       string `json:"name"`
	Title      string `json:"title"`
	URL        string `json:"url"`
	Icon       string `json:"icon"`
	Invisible  bool   `json:"invisible"`
	AuthLevel  string `json:"authLevel"`
	OpenMethod string `json:"openMethod"`

	State   string `json:"state"`
	Reason  string `json:"reason,omitempty"`
	Message string `json:"message,omitempty"`
}
type ServicePort struct {
	Name string `json:"name"`
	Host string `json:"host"`
	Port int32  `json:"port"`

	ExposePort int32 `json:"exposePort,omitempty"`

	// The protocol for this entrance. Supports "tcp" and "udp".
	// Default is tcp.
	// +default="tcp"
	// +optional
	Protocol string `json:"protocol,omitempty"`
}
type ACL struct {
	Action string   `json:"action,omitempty"`
	Src    []string `json:"src,omitempty"`
	Proto  string   `json:"proto"`
	Dst    []string `json:"dst"`
}

type AppDeploymentInfo struct {
	AppInfo    *AppInfo
	Deployment *k8sv1.Deployment
}

type ApplicationSettingsSubPolicy struct {
	URI      string        `json:"uri"`
	Policy   string        `json:"policy"`
	OneTime  bool          `json:"one_time"`
	Duration time.Duration `json:"valid_duration"`
}

type ApplicationSettingsPolicy struct {
	DefaultPolicy string                          `json:"default_policy"`
	SubPolicies   []*ApplicationSettingsSubPolicy `json:"sub_policies"`
	OneTime       bool                            `json:"one_time"`
	Duration      time.Duration                   `json:"valid_duration"`
}

type InstallOptions struct {
	App     string `json:"appName"`
	Dev     bool   `json:"devMode"`
	RepoURL string `json:"repoUrl"`
	CfgURL  string `json:"cfgUrl"`
	Source  string `json:"source"`
}

type UpgradeOptions struct {
	CfgURL  string `json:"cfgURL,omitempty"`
	RepoURL string `json:"repoURL"`
	Version string `json:"version"`
	Source  string `json:"source"`
}

type ApplicationSettingsDomain struct {
	ThirdLevelDomain string `json:"third_level_domain"`
	ThirdPartyDomain string `json:"third_party_domain"`
	Cert             string `json:"cert"`
	Key              string `json:"key"`
}

const (
	ApplicationSettingsPolicyKey     = "policy"
	oneFactor                        = "one_factor"
	twoFactor                        = "two_factor"
	deny                             = "deny"
	public                           = "public"
	ApplicationSettingsDomainKey     = "customDomain"
	ApplicationAuthorizationLevelKey = "authorizationLevel"
)

type EntrancesWithCustomDomain []EntranceWithCustomDomain

type EntranceWithCustomDomain struct {
	appv1.Entrance
	AppName      string `json:"app_name"`
	CustomDomain string `json:"custom_domain"`
}

type ApplicationsSettings map[string]interface{}

type Entrances []string
