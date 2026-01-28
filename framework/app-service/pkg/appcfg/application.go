package appcfg

import (
	"context"
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
