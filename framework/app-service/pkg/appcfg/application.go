package appcfg

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"github.com/beclab/Olares/framework/app-service/api/app.bytetrade.io/v1alpha1"
	sysv1alpha1 "github.com/beclab/Olares/framework/app-service/api/sys.bytetrade.io/v1alpha1"
	"github.com/beclab/Olares/framework/app-service/pkg/tapr"
	"github.com/beclab/Olares/framework/app-service/pkg/utils"

	"k8s.io/apimachinery/pkg/api/resource"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/klog/v2"
)

type AppPermission interface{}

type AppDataPermission string
type AppCachePermission string
type UserDataPermission string

type Middleware interface{}

type AppRequirement struct {
	Memory *resource.Quantity
	Disk   *resource.Quantity
	GPU    *resource.Quantity
	CPU    *resource.Quantity
}

type AppPolicy struct {
	EntranceName string        `yaml:"entranceName" json:"entranceName"`
	URIRegex     string        `yaml:"uriRegex" json:"uriRegex" description:"uri regular expression"`
	Level        string        `yaml:"level" json:"level"`
	OneTime      bool          `yaml:"oneTime" json:"oneTime"`
	Duration     time.Duration `yaml:"validDuration" json:"validDuration"`
}

const (
	AppDataRW  AppDataPermission  = "appdata-perm"
	AppCacheRW AppCachePermission = "appcache-perm"
	UserDataRW UserDataPermission = "userdata-perm"
)

type APIVersion string

const (
	V1 APIVersion = "v1"
	V2 APIVersion = "v2"
)

type ApplicationConfig struct {
	AppID                string
	APIVersion           APIVersion
	CfgFileVersion       string
	Namespace            string
	MiddlewareName       string
	ChartsName           string
	RepoURL              string
	Title                string
	Version              string
	Target               string
	AppName              string // name of application displayed on shortcut
	OwnerName            string // name of owner who installed application
	Entrances            []v1alpha1.Entrance
	Ports                []v1alpha1.ServicePort
	TailScale            v1alpha1.TailScale
	Icon                 string          // base64 icon data
	Permission           []AppPermission // app permission requests
	Requirement          AppRequirement
	Policies             []AppPolicy
	Middleware           *tapr.Middleware
	ResetCookieEnabled   bool
	Dependencies         []Dependency
	Conflicts            []Conflict
	AppScope             AppScope
	WsConfig             WsConfig
	Upload               Upload
	OnlyAdmin            bool
	MobileSupported      bool
	OIDC                 OIDC
	ApiTimeout           *int64
	RunAsUser            bool
	AllowedOutboundPorts []int
	RequiredGPU          string
	PodGPUConsumePolicy  string
	Release              []string
	ClusterRelease       []string
	Internal             bool
	SubCharts            []Chart
	ServiceAccountName   *string
	Provider             []Provider
	Type                 string
	Envs                 []sysv1alpha1.AppEnvVar
	Images               []string
	AllowMultipleInstall bool
	RawAppName           string
	PodsSelectors        []metav1.LabelSelector
	HardwareRequirement  Hardware
	SharedEntrances      []v1alpha1.Entrance
	SelectedGpuType      string
	Resources            []ResourceMode
	InstallType          string
	Client               *ConfigOverlay
	Server               *ConfigOverlay
}

func (c *ApplicationConfig) IsMiddleware() bool {
	return c.Type == "middleware"
}

func (c *ApplicationConfig) IsV2() bool {
	return c.APIVersion == V2
}

func (c *ApplicationConfig) IsMultiCharts() bool {
	return len(c.SubCharts) > 1
}

func (c *ApplicationConfig) HasClusterSharedCharts() bool {
	for _, chart := range c.SubCharts {
		if chart.Shared {
			return true
		}
	}
	return false
}

func (c *ApplicationConfig) GenEntranceURL(ctx context.Context) ([]v1alpha1.Entrance, error) {
	app := &v1alpha1.Application{
		Spec: v1alpha1.ApplicationSpec{
			Owner:     c.OwnerName,
			Name:      c.AppName,
			Entrances: c.Entrances,
		},
	}

	return app.GenEntranceURL(ctx)
}

func (c *ApplicationConfig) GetEntrances(ctx context.Context) (map[string]v1alpha1.Entrance, error) {
	entrances, err := c.GenEntranceURL(ctx)
	if err != nil {
		klog.Errorf("failed to generate entrance URL: %v", err)
		return nil, err
	}

	return utils.ListToMap(entrances, func(e v1alpha1.Entrance) string {
		return e.Name
	}), nil
}

func (c *ApplicationConfig) GenSharedEntranceURL(ctx context.Context) ([]v1alpha1.Entrance, error) {
	app := &v1alpha1.Application{
		Spec: v1alpha1.ApplicationSpec{
			Owner:           c.OwnerName,
			Name:            c.AppName,
			SharedEntrances: c.SharedEntrances,
		},
	}

	return app.GenSharedEntranceURL(ctx)
}

func (c *ApplicationConfig) GetSelectedGpuTypeValue() string {
	if c.SelectedGpuType == "" {
		return "none"
	}
	return c.SelectedGpuType
}

func (c *ApplicationConfig) DeepCopy() *ApplicationConfig {
	data, err := json.Marshal(c)
	if err != nil {
		klog.Errorf("failed to marshal ApplicationConfig for deep copy: %v", err)
		return nil
	}
	out := &ApplicationConfig{}
	if err := json.Unmarshal(data, out); err != nil {
		klog.Errorf("failed to unmarshal ApplicationConfig for deep copy: %v", err)
		return nil
	}
	return out
}

// ApplyOverlay mutates c in place, merging fields from the Server/Client
// overlay (selected by installType) into the top-level config. It's a no-op
// when c is nil or no overlay is defined for the given installType.
func (c *ApplicationConfig) ApplyOverlay(installType string) {
	if c == nil {
		return
	}
	c.InstallType = installType
	c.applyConfigOverlay(installType)
}

func (c *ApplicationConfig) applyConfigOverlay(installType string) {
	var overlay *ConfigOverlay
	klog.Infof("applyConfigOverlay: installType: %v", installType)
	switch installType {
	case InstallServerAndClient:
		overlay = c.Server
	case InstallClientOnly:
		overlay = c.Client
	}
	if overlay == nil {
		return
	}

	c.Entrances = overlay.Entrances
	c.Middleware = overlay.Middleware

	var permission []AppPermission
	if overlay.Permission.AppData {
		permission = append(permission, AppDataRW)
	}
	if overlay.Permission.AppCache {
		permission = append(permission, AppCacheRW)
	}
	if len(overlay.Permission.UserData) > 0 {
		permission = append(permission, UserDataRW)
	}
	if len(overlay.Permission.Provider) > 0 {
		var perm []ProviderPermission
		for _, s := range overlay.Permission.Provider {
			perm = append(perm, ProviderPermission(s))
		}
		permission = append(permission, perm)
	}
	c.Permission = permission

	c.ResetCookieEnabled = overlay.Options.ResetCookie.Enabled
	c.Dependencies = overlay.Options.Dependencies
	c.Conflicts = overlay.Options.Conflicts
	c.AppScope = overlay.Options.AppScope
	c.WsConfig = overlay.Options.WsConfig
	c.Upload = overlay.Options.Upload
	c.MobileSupported = overlay.Options.MobileSupported
	c.OIDC = overlay.Options.OIDC
	c.ApiTimeout = overlay.Options.ApiTimeout
	c.AllowedOutboundPorts = overlay.Options.AllowedOutboundPorts
	c.Images = overlay.Options.Images
	c.AllowMultipleInstall = overlay.Options.AllowMultipleInstall
	c.Provider = overlay.Provider
	c.Envs = overlay.Envs
}

func (c *ApplicationConfig) ResolveRequirement(selectedGpu, installType string) (*AppRequirement, error) {
	if len(c.Resources) == 0 {
		return nil, fmt.Errorf("empty spec resources")
	}

	mode := resolveResourceMode(c.Resources, selectedGpu)
	if mode == nil {
		return nil, fmt.Errorf("mode %s not found in spec resources", selectedGpu)
	}
	if c.APIVersion == V1 {
		req := parseResourceRequirement(&mode.ResourceRequirement)
		return &req, nil
	}

	if mode.Client == nil && mode.Server == nil {
		if mode.ResourceRequirement == (ResourceRequirement{}) {
			return nil, fmt.Errorf("empty resource requirement")
		}
		req := parseResourceRequirement(&mode.ResourceRequirement)
		return &req, nil
	}

	switch installType {
	case InstallClientOnly:
		if mode.Client == nil {
			return nil, fmt.Errorf("client resource requirement can not be empty")
		}
		req := parseResourceRequirement(mode.Client)
		return &req, nil
	case InstallServerAndClient:
		req := sumResourceRequirements(mode.Server, mode.Client)
		return &req, nil
	}
	return nil, fmt.Errorf("no resource requirement found")
}

func resolveResourceMode(modes []ResourceMode, selectedGpu string) *ResourceMode {
	targetMode := selectedGpu
	if targetMode == "" || targetMode == "none" {
		targetMode = utils.CPUType
	}

	for i := range modes {
		if modes[i].Mode == targetMode {
			return &modes[i]
		}
	}

	// no target mode found fallback to cpu mode
	for i := range modes {
		if modes[i].Mode == utils.CPUType {
			return &modes[i]
		}
	}

	return nil
}

func parseResourceRequirement(req *ResourceRequirement) AppRequirement {
	parseQty := func(s string) *resource.Quantity {
		if s == "" {
			return nil
		}
		q, err := resource.ParseQuantity(s)
		if err != nil {
			return nil
		}
		return &q
	}

	return AppRequirement{
		CPU:    parseQty(req.RequiredCPU),
		Memory: parseQty(req.RequiredMemory),
		Disk:   parseQty(req.RequiredDisk),
		GPU:    parseQty(req.RequiredGPU),
	}
}

func sumResourceRequirements(parts ...*ResourceRequirement) AppRequirement {
	parseAndAdd := func(existing *resource.Quantity, s string) *resource.Quantity {
		if s == "" {
			return existing
		}
		q, err := resource.ParseQuantity(s)
		if err != nil {
			return existing
		}
		if existing == nil {
			return &q
		}
		existing.Add(q)
		return existing
	}

	var cpu, mem, disk, gpu *resource.Quantity
	for _, p := range parts {
		if p == nil {
			continue
		}
		cpu = parseAndAdd(cpu, p.RequiredCPU)
		mem = parseAndAdd(mem, p.RequiredMemory)
		disk = parseAndAdd(disk, p.RequiredDisk)
		gpu = parseAndAdd(gpu, p.RequiredGPU)
	}
	return AppRequirement{CPU: cpu, Memory: mem, Disk: disk, GPU: gpu}
}

func (p *ProviderPermission) GetNamespace(ownerName string) string {
	if p.Namespace != "" {
		if p.Namespace == "user-space" || p.Namespace == "user-system" {
			return fmt.Sprintf("%s-%s", p.Namespace, ownerName)
		} else {
			return p.Namespace
		}
	}

	return fmt.Sprintf("%s-%s", p.AppName, ownerName)
}
