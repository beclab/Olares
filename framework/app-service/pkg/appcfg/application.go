package appcfg

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/beclab/Olares/framework/app-service/pkg/utils"
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
	Memory        *resource.Quantity
	Disk          *resource.Quantity
	GPU           *resource.Quantity
	CPU           *resource.Quantity
	LimitedMemory *resource.Quantity
	LimitedDisk   *resource.Quantity
	LimitedGPU    *resource.Quantity
	LimitedCPU    *resource.Quantity
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
	V3 APIVersion = "v3"
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
	Accelerator          []ResourceMode
	// NeedsSharedAccess signals that the app needs cross-namespace access to a
	// v3 app's services (e.g. for service-mesh sidecar injection).
	// Force-set to true for v3 apps in toApplicationConfig regardless of
	// manifest value because v3 apps are themselves the destination of shared
	// traffic and naturally need the same treatment.
	NeedsSharedAccess       bool
	OverlayGatewaySupported bool
	LLMGatewaySupported     bool
}

func (c *ApplicationConfig) IsMiddleware() bool {
	return c.Type == "middleware"
}

func (c *ApplicationConfig) IsV2() bool {
	return c.APIVersion == V2
}

// IsV3 reports whether the app is declared with apiVersion: v3.
func (c *ApplicationConfig) IsV3() bool {
	return c.APIVersion == V3
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

func (c *ApplicationConfig) SelectedResourceMode() (ResourceMode, bool) {
	modes := c.ComputeResourceModes()
	if len(modes) == 1 && len(c.Accelerator) == 0 {
		return modes[0], true
	}
	mode := findResourceMode(modes, c.SelectedGpuType)
	if mode == nil {
		return ResourceMode{}, false
	}
	return *mode, true
}

// ComputeResourceModes returns the modes the app declares support for.
// New-format manifests (>= 0.12.0) carry an explicit spec.resources matrix
// and are returned verbatim. Legacy manifests (< 0.12.0) don't have a
// per-mode matrix, so we synthesize a single ResourceMode whose Mode is
// derived from c.SelectedGpuType / c.Requirement.GPU — see
// legacyComputeMode for the exact rules.
func (c *ApplicationConfig) ComputeResourceModes() []ResourceMode {
	if len(c.Accelerator) > 0 {
		return c.Accelerator
	}
	mode := legacyComputeMode(c)
	return []ResourceMode{{
		Mode:                mode,
		ResourceRequirement: c.Requirement.ResourceRequirement(mode),
	}}
}

// legacyComputeMode picks the synthesized mode for a legacy manifest:
//
//   - If SelectedGpuType is set, use it verbatim. This covers the path
//     where the install caller explicitly picked a GPU type (or the
//     auto-selector did so on their behalf and we re-loaded the appCfg
//     with the chosen mode).
//   - If SelectedGpuType is empty and the manifest has a non-zero
//     Requirement.GPU, fall back to nvidia — the only legacy-supported
//     GPU type. This is what the install pre-check and the first
//     GetAppConfig call see before auto-selection runs.
//   - Otherwise the app is cpu-only.
func legacyComputeMode(c *ApplicationConfig) string {
	if c.SelectedGpuType != "" {
		return c.SelectedGpuType
	}
	if c.Requirement.GPU != nil && !c.Requirement.GPU.IsZero() {
		return utils.NvidiaCardType
	}
	return utils.CPUType
}

func (r AppRequirement) ResourceRequirement(mode string) ResourceRequirement {
	req := ResourceRequirement{
		RequiredCPU:    quantityString(r.CPU),
		LimitedCPU:     quantityString(r.LimitedCPU),
		RequiredMemory: quantityString(r.Memory),
		LimitedMemory:  quantityString(r.LimitedMemory),
		RequiredDisk:   quantityString(r.Disk),
		LimitedDisk:    quantityString(r.LimitedDisk),
	}
	if mode == utils.NvidiaCardType {
		req.RequiredGPU = quantityString(r.GPU)
		req.LimitedGPU = quantityString(r.LimitedGPU)
	}
	return req
}

func quantityString(q *resource.Quantity) string {
	if q == nil {
		return ""
	}
	return q.String()
}

// resolveResourceMode picks the ResourceMode the install pipeline should
// load against. The fallback rules differ depending on whether the caller
// is expressing "no preference" or "I picked this specific mode":
//
//   - selectedGpu == ""  →  caller hasn't decided yet (install pre-check or
//     the first GetAppConfig before auto-select runs). Prefer the cpu mode
//     so legacy-style chart values stay stable; if the manifest doesn't
//     declare a cpu mode (e.g. a GPU-only new-format app), fall back to
//     the first declared mode as a placeholder. Either way, the install
//     handler's auto-select step reloads the chart with the real chosen
//     mode immediately afterwards, so the placeholder Requirement never
//     reaches AllocateForInstall / AppInstallable.
//
//   - selectedGpu != ""  →  caller explicitly picked this mode. Match
//     exactly and surface a "not declared" error if the manifest doesn't
//     have it; silently falling back to cpu / first-mode would leave
//     SelectedGpuType and Requirement out of sync and produce a misleading
//     "compute resource is not enough" message at the AppInstallable gate.
//
// Returns nil only when modes is empty (the caller turns that into an
// "empty compute resources" error) or when an explicitly selected mode
// is not declared in the manifest.
func resolveResourceMode(modes []ResourceMode, selectedGpu string) *ResourceMode {
	if selectedGpu == "" {
		if cpu := findResourceMode(modes, utils.CPUType); cpu != nil {
			return cpu
		}
		// GPU-only new-format manifest: no cpu mode to fall back to. Use
		// the first declared mode as a load-time placeholder so the chart
		// loader can succeed; the install handler's auto-select step will
		// pick the real mode and reload before any downstream code sees
		// this Requirement.
		if len(modes) > 0 {
			return &modes[0]
		}
		return nil
	}
	return findResourceMode(modes, selectedGpu)
}

func findResourceMode(modes []ResourceMode, mode string) *ResourceMode {
	for i := range modes {
		if modes[i].Mode == mode {
			return &modes[i]
		}
	}
	return nil
}

// ParseResourceRequirement converts the scalar required*/limited* fields of a
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
	limCPU, err := parseQty(req.LimitedCPU)
	if err != nil {
		return AppRequirement{}, fmt.Errorf("parse limited cpu %q: %w", req.LimitedCPU, err)
	}
	limMem, err := parseQty(req.LimitedMemory)
	if err != nil {
		return AppRequirement{}, fmt.Errorf("parse limited memory %q: %w", req.LimitedMemory, err)
	}
	limDisk, err := parseQty(req.LimitedDisk)
	if err != nil {
		return AppRequirement{}, fmt.Errorf("parse limited disk %q: %w", req.LimitedDisk, err)
	}
	limGPU, err := parseQty(req.LimitedGPU)
	if err != nil {
		return AppRequirement{}, fmt.Errorf("parse limited gpu %q: %w", req.LimitedGPU, err)
	}

	return AppRequirement{
		CPU:           cpu,
		Memory:        mem,
		Disk:          disk,
		GPU:           gpu,
		LimitedCPU:    limCPU,
		LimitedMemory: limMem,
		LimitedDisk:   limDisk,
		LimitedGPU:    limGPU,
	}, nil
}

// ResolveRequirement returns the effective resource requirement for the app,
// taking the selected GPU type and manifest version into account. It is the
// single source of truth for picking between the new spec.resources matrix
// and the legacy scalar spec.required* fields.
//
//   - New manifest format (>= 0.12.0): pick the ResourceMode whose Mode
//     equals selectedGpu from c.Resources. If selectedGpu is empty
//     (pre-check / first GetAppConfig before auto-select runs), prefer
//     the cpu mode; if the manifest declares no cpu mode, fall back to
//     the first declared mode as a placeholder so the chart loader has
//     something to bind against. The install handler's auto-select step
//     reloads with the real chosen mode immediately afterwards. If
//     selectedGpu names a mode the manifest doesn't declare, return an
//     error rather than silently using cpu / first-mode.
//   - Legacy manifest format (< 0.12.0): synthesize a single ResourceMode
//     from c.Requirement. Apps without requiredGpu become cpu mode; apps
//     with requiredGpu become nvidia mode (or whatever SelectedGpuType
//     is set to).
func (c *ApplicationConfig) ResolveRequirement(selectedGpu string) (*AppRequirement, error) {
	modes := c.ComputeResourceModes()
	if len(modes) == 0 {
		return nil, fmt.Errorf("empty compute resources")
	}
	if len(c.Accelerator) == 0 && len(modes) == 1 {
		req, err := ParseResourceRequirement(&modes[0].ResourceRequirement)
		if err != nil {
			return nil, fmt.Errorf("resolve requirement for mode %s: %w", modes[0].Mode, err)
		}
		return &req, nil
	}
	mode := resolveResourceMode(modes, selectedGpu)
	if mode == nil {
		// resolveResourceMode only returns nil when an *explicit* mode
		// wasn't found in the manifest; the empty-selectedGpu case is
		// handled by the placeholder fallback inside it.
		return nil, fmt.Errorf("mode %q is not declared in spec.resources", selectedGpu)
	}
	req, err := ParseResourceRequirement(&mode.ResourceRequirement)
	if err != nil {
		return nil, fmt.Errorf("resolve requirement for mode %s: %w", mode.Mode, err)
	}
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
