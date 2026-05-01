package appcfg

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/beclab/Olares/framework/app-service/pkg/utils"
	"github.com/beclab/Olares/framework/oac"
	manifestsysv1alpha1 "github.com/beclab/api/api/sys.bytetrade.io/v1alpha1"

	"k8s.io/apimachinery/pkg/api/resource"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/klog/v2"
)

type AppPermission interface{}

type AppDataPermission string
type AppCachePermission string
type UserDataPermission string

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
	Entrances            []Entrance
	Ports                []ServicePort
	TailScale            TailScale
	Icon                 string          // base64 icon data
	Permission           []AppPermission // app permission requests
	Requirement          AppRequirement
	Policies             []AppPolicy
	Middleware           *Middleware
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
	Envs                 []manifestsysv1alpha1.AppEnvVar
	Images               []string
	AllowMultipleInstall bool
	RawAppName           string
	PodsSelectors        []metav1.LabelSelector
	HardwareRequirement  Hardware
	SharedEntrances      []Entrance
	SelectedGpuType      string
	Resources            []ResourceMode
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

func (c *ApplicationConfig) GenEntranceURL(ctx context.Context) ([]Entrance, error) {
	app := &Application{
		Spec: ApplicationSpec{
			Owner:     c.OwnerName,
			Name:      c.AppName,
			Entrances: c.Entrances,
		},
	}

	return GenEntranceURL(ctx, app)
}

func (c *ApplicationConfig) GetEntrances(ctx context.Context) (map[string]Entrance, error) {
	entrances, err := c.GenEntranceURL(ctx)
	if err != nil {
		klog.Errorf("failed to generate entrance URL: %v", err)
		return nil, err
	}

	return utils.ListToMap(entrances, func(e Entrance) string {
		return e.Name
	}), nil
}

func (c *ApplicationConfig) GenSharedEntranceURL(ctx context.Context) ([]Entrance, error) {
	app := &Application{
		Spec: ApplicationSpec{
			Owner:           c.OwnerName,
			Name:            c.AppName,
			SharedEntrances: c.SharedEntrances,
		},
	}

	return GenSharedEntranceURL(ctx, app)
}

func (c *ApplicationConfig) GetSelectedGpuTypeValue() string {
	if c.SelectedGpuType == "" {
		return "none"
	}
	return c.SelectedGpuType
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

// ParseResourceRequirement converts the scalar required* fields of a
// ResourceRequirement into an AppRequirement. Empty fields and values that
// fail with resource.ErrFormatWrong are treated as "not set" and produce a
// nil quantity; any other parse error is returned to the caller. This mirrors
// the behaviour of parseLegacyAppRequirement so the two code paths agree on
// what "invalid manifest" means.
func ParseResourceRequirement(req *ResourceRequirement) (AppRequirement, error) {
	parseQty := func(s string) (*resource.Quantity, error) {
		if s == "" {
			return nil, nil
		}
		q, err := resource.ParseQuantity(s)
		if err != nil {
			if errors.Is(err, resource.ErrFormatWrong) {
				return nil, nil
			}
			return nil, err
		}
		return &q, nil
	}

	cpu, err := parseQty(req.RequiredCPU)
	if err != nil {
		return AppRequirement{}, fmt.Errorf("parse required cpu %q: %w", req.RequiredCPU, err)
	}
	mem, err := parseQty(req.RequiredMemory)
	if err != nil {
		return AppRequirement{}, fmt.Errorf("parse required memory %q: %w", req.RequiredMemory, err)
	}
	disk, err := parseQty(req.RequiredDisk)
	if err != nil {
		return AppRequirement{}, fmt.Errorf("parse required disk %q: %w", req.RequiredDisk, err)
	}
	gpu, err := parseQty(req.RequiredGPU)
	if err != nil {
		return AppRequirement{}, fmt.Errorf("parse required gpu %q: %w", req.RequiredGPU, err)
	}

	return AppRequirement{
		CPU:    cpu,
		Memory: mem,
		Disk:   disk,
		GPU:    gpu,
	}, nil
}

// ResolveRequirement returns the effective resource requirement for the app,
// taking the selected GPU type and manifest version into account. It is the
// single source of truth for picking between the new spec.resources matrix
// and the legacy scalar spec.required* fields.
//
//   - New manifest format (>= 0.12.0): pick the ResourceMode that matches
//     selectedGpu from c.Resources (which must be non-empty).
//   - Legacy manifest format (< 0.12.0): return c.Requirement directly. The
//     scalar fields are expected to already be populated (and any
//     supportedGpu special-resource overrides applied) at conversion time.
func (c *ApplicationConfig) ResolveRequirement(selectedGpu string) (*AppRequirement, error) {
	if oac.IsNewOlaresManifestVersion(c.CfgFileVersion) {
		if len(c.Resources) == 0 {
			return nil, fmt.Errorf("empty spec resources")
		}
		mode := resolveResourceMode(c.Resources, selectedGpu)
		if mode == nil {
			return nil, fmt.Errorf("mode %s not found in spec resources", selectedGpu)
		}
		req, err := ParseResourceRequirement(&mode.ResourceRequirement)
		if err != nil {
			return nil, fmt.Errorf("resolve requirement for mode %s: %w", mode.Mode, err)
		}
		return &req, nil
	}

	req := c.Requirement
	return &req, nil
}

// ProviderPermissionNamespace returns the namespace a provider permission
// resolves to for the given owner. It used to be a method on ProviderPermission,
// but that type is now an alias to manifest.ProviderPermission and methods
// cannot be defined on non-local types.
func ProviderPermissionNamespace(p *ProviderPermission, ownerName string) string {
	if p.Namespace != "" {
		if p.Namespace == "user-space" || p.Namespace == "user-system" {
			return fmt.Sprintf("%s-%s", p.Namespace, ownerName)
		} else {
			return p.Namespace
		}
	}

	return fmt.Sprintf("%s-%s", p.AppName, ownerName)
}
