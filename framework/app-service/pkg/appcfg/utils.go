package appcfg

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"path/filepath"
	"time"

	"bytetrade.io/web3os/app-service/pkg/kubesphere"
	"bytetrade.io/web3os/app-service/pkg/utils"
	"gopkg.in/yaml.v2"
	"k8s.io/apimachinery/pkg/api/resource"
	"k8s.io/klog/v2"
	ctrl "sigs.k8s.io/controller-runtime"
)

const (
	ChartsPath = "./charts"
)

func AppChartPath(app string) string {
	return ChartsPath + "/" + app
}

// GetAppInstallationConfig get app installation configuration from app store
func GetAppInstallationConfig(app, owner string) (*ApplicationConfig, error) {
	//chart := AppChartPath(rawAppName)
	appcfg, err := getAppConfigFromAppMgrConfig(app, owner)
	if err != nil {
		return nil, err
	}

	// TODO: app installation namespace
	var namespace string
	if appcfg.Namespace != "" {
		namespace, _ = utils.AppNamespace(app, owner, appcfg.Namespace)
	} else {
		namespace = fmt.Sprintf("%s-%s", app, owner)
	}

	appcfg.Namespace = namespace
	appcfg.OwnerName = owner

	return appcfg, nil
}

func getAppConfigFromAppMgrConfig(appName, owner string) (*ApplicationConfig, error) {
	kclient, err := utils.GetClient()
	if err != nil {
		return nil, err
	}
	name := fmt.Sprintf("%s-%s-%s", appName, owner, appName)
	am, err := kclient.AppV1alpha1().ApplicationManagers().Get(context.TODO(), name, metav1.GetOptions{})
	if err != nil {
		return nil, err
	}
	appConfig := ApplicationConfig{}
	err = json.Unmarshal([]byte(am.Spec.Config), &appConfig)
	if err != nil {
		return nil, err
	}
	return &appConfig, nil

}

func getAppConfigFromConfigurationFile(app, chart, owner string) (*ApplicationConfig, error) {
	//f, err := os.Open(filepath.Join(chart, "OlaresManifest.yaml"))
	//if err != nil {
	//	return nil, err
	//}
	//defer f.Close()
	//data, err := ioutil.ReadAll(f)
	//if err != nil {
	//	return nil, err
	//}
	config, err := ctrl.GetConfig()
	if err != nil {
		return nil, err
	}
	admin, err := kubesphere.GetAdminUsername(context.TODO(), config)
	if err != nil {
		return nil, err
	}
	isAdmin, err := kubesphere.IsAdmin(context.TODO(), config, owner)
	if err != nil {
		return nil, err
	}
	if isAdmin {
		admin = owner
	}
	data, err := utils.RenderManifest(filepath.Join(chart, "OlaresManifest.yaml"), owner, admin, isAdmin)
	if err != nil {
		return nil, err
	}
	var cfg AppConfiguration
	if err := yaml.Unmarshal([]byte(data), &cfg); err != nil {
		return nil, err
	}

	var permission []AppPermission
	if cfg.Permission.AppData {
		permission = append(permission, AppDataRW)
	}
	if cfg.Permission.AppCache {
		permission = append(permission, AppCacheRW)
	}
	if len(cfg.Permission.UserData) > 0 {
		permission = append(permission, UserDataRW)
	}

	if len(cfg.Permission.Provider) > 0 {
		var perm []ProviderPermission
		for _, s := range cfg.Permission.Provider {
			perm = append(perm, ProviderPermission{
				AppName:      s.AppName,
				Namespace:    s.Namespace,
				ProviderName: s.ProviderName,
			})
		}
		permission = append(permission, perm)
	}

	valuePtr := func(v resource.Quantity, err error) (*resource.Quantity, error) {
		if errors.Is(err, resource.ErrFormatWrong) {
			return nil, nil
		}

		return &v, nil
	}

	mem, err := valuePtr(resource.ParseQuantity(cfg.Spec.RequiredMemory))
	if err != nil {
		return nil, err
	}

	disk, err := valuePtr(resource.ParseQuantity(cfg.Spec.RequiredDisk))
	if err != nil {
		return nil, err
	}

	cpu, err := valuePtr(resource.ParseQuantity(cfg.Spec.RequiredCPU))
	if err != nil {
		return nil, err
	}
	gpu, err := valuePtr(resource.ParseQuantity(cfg.Spec.RequiredGPU))
	if err != nil {
		return nil, err
	}

	var polices []AppPolicy
	if len(cfg.Options.Policies) > 0 {
		for _, p := range cfg.Options.Policies {
			duration, err := time.ParseDuration(p.Duration)
			if err != nil {
				klog.Errorf("Failed to parse app cfg options policy duration err=%v", err)
			}
			polices = append(polices, AppPolicy{
				EntranceName: p.EntranceName,
				URIRegex:     p.URIRegex,
				Level:        p.Level,
				OneTime:      p.OneTime,
				Duration:     duration,
			})
		}
	}

	return &ApplicationConfig{
		AppID:          cfg.Metadata.AppID,
		CfgFileVersion: cfg.ConfigVersion,
		AppName:        app,
		Title:          cfg.Metadata.Title,
		Version:        cfg.Metadata.Version,
		Target:         cfg.Metadata.Target,
		ChartsName:     chart,
		Entrances:      cfg.Entrances,
		Ports:          cfg.Ports,
		TailScale:      cfg.TailScale,
		Icon:           cfg.Metadata.Icon,
		Permission:     permission,
		Requirement: AppRequirement{
			Memory: mem,
			CPU:    cpu,
			Disk:   disk,
			GPU:    gpu,
		},
		Policies:             polices,
		ResetCookieEnabled:   cfg.Options.ResetCookie.Enabled,
		Dependencies:         cfg.Options.Dependencies,
		Conflicts:            cfg.Options.Conflicts,
		AppScope:             cfg.Options.AppScope,
		OnlyAdmin:            cfg.Spec.OnlyAdmin,
		Namespace:            cfg.Spec.Namespace,
		MobileSupported:      cfg.Options.MobileSupported,
		OIDC:                 cfg.Options.OIDC,
		ApiTimeout:           cfg.Options.ApiTimeout,
		AllowedOutboundPorts: cfg.Options.AllowedOutboundPorts,
		RequiredGPU:          cfg.Spec.RequiredGPU,
		PodGPUConsumePolicy:  cfg.Spec.PodGPUConsumePolicy,
		Internal:             cfg.Spec.RunAsInternal,
		Type:                 cfg.ConfigType,
		Envs:                 cfg.Envs,
		Images:               cfg.Options.Images,
		HardwareRequirement:  cfg.Spec.Hardware,
	}, nil
}
