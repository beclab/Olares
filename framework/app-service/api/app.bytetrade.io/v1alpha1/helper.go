package v1alpha1

import (
	"context"
	"crypto/md5"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"strings"

	"github.com/beclab/Olares/framework/app-service/pkg/constants"
	"github.com/beclab/Olares/framework/app-service/pkg/kubesphere"
	"github.com/beclab/Olares/framework/app-service/pkg/users/userspace"
	"k8s.io/klog/v2"
)

type DefaultThirdLevelDomainConfig struct {
	AppName          string `json:"appName"`
	EntranceName     string `json:"entranceName"`
	ThirdLevelDomain string `json:"thirdLevelDomain"`
}

func (a *Application) IsClusterScoped() bool {
	if a.Spec.Settings == nil {
		return false
	}
	if v, ok := a.Spec.Settings["clusterScoped"]; ok && v == "true" {
		return true
	}
	return false
}

func (a *ApplicationManager) GetAppConfig(appConfig any) (err error) {
	err = json.Unmarshal([]byte(a.Spec.Config), appConfig)
	if err != nil {
		klog.Errorf("unmarshal to appConfig failed %v", err)
		return err
	}

	return
}

func (a *ApplicationManager) SetAppConfig(appConfig any) error {
	configBytes, err := json.Marshal(appConfig)
	if err != nil {
		klog.Errorf("marshal appConfig failed %v", err)
		return err
	}

	a.Spec.Config = string(configBytes)

	return nil
}

func (a *ApplicationManager) GetMarketSource() string {
	return a.Annotations[constants.AppMarketSourceKey]
}

type AppName string

func (s AppName) GetAppID() string {
	if s.IsSysApp() {
		return string(s)
	}
	hash := md5.Sum([]byte(s))
	hashString := hex.EncodeToString(hash[:])
	return hashString[:8]
}

func (s AppName) String() string {
	return string(s)
}

func (s AppName) IsSysApp() bool {
	return userspace.IsSysApp(string(s))
}

func (s AppName) IsGeneratedApp() bool {
	return userspace.IsGeneratedApp(string(s))
}

func (s AppName) SharedEntranceIdPrefix() string {
	hash := md5.Sum([]byte(s.GetAppID() + "shared"))
	hashString := hex.EncodeToString(hash[:])
	return hashString[:8]
}

func (app *Application) GenEntranceURL(ctx context.Context) ([]Entrance, error) {
	zone, err := kubesphere.GetUserZone(ctx, app.Spec.Owner)
	if err != nil {
		klog.Errorf("failed to get user zone: %v", err)
	}

	if len(zone) > 0 {
		var appDomainConfigs []DefaultThirdLevelDomainConfig
		if defaultThirdLevelDomainConfig, ok := app.Spec.Settings["defaultThirdLevelDomainConfig"]; ok && len(defaultThirdLevelDomainConfig) > 0 {
			err := json.Unmarshal([]byte(app.Spec.Settings["defaultThirdLevelDomainConfig"]), &appDomainConfigs)
			if err != nil {
				klog.Errorf("unmarshal defaultThirdLevelDomainConfig error %v", err)
				return nil, err
			}
		}

		appid := AppName(app.Spec.Name).GetAppID()
		if len(app.Spec.Entrances) == 1 {
			app.Spec.Entrances[0].URL = fmt.Sprintf("%s.%s", appid, zone)
		} else {
			for i := range app.Spec.Entrances {
				app.Spec.Entrances[i].URL = fmt.Sprintf("%s%d.%s", appid, i, zone)
				for _, adc := range appDomainConfigs {
					if adc.AppName == app.Spec.Name && adc.EntranceName == app.Spec.Entrances[i].Name && len(adc.ThirdLevelDomain) > 0 {
						app.Spec.Entrances[i].URL = fmt.Sprintf("%s.%s", adc.ThirdLevelDomain, zone)
					}
				}
			}
		}
	}
	return app.Spec.Entrances, nil
}

func (app *Application) GenSharedEntranceURL(ctx context.Context) ([]Entrance, error) {
	zone, err := kubesphere.GetUserZone(ctx, app.Spec.Owner)
	if err != nil {
		klog.Errorf("failed to get user zone: %v", err)
	}

	if len(zone) > 0 {
		tokens := strings.Split(zone, ".")
		tokens[0] = "shared"
		sharedZone := strings.Join(tokens, ".")

		appName := AppName(app.Spec.Name)
		sharedEntranceIdPrefix := appName.SharedEntranceIdPrefix()
		for i := range app.Spec.SharedEntrances {
			if app.Spec.SharedEntrances[i].Port > 0 {
				app.Spec.SharedEntrances[i].URL = fmt.Sprintf("%s%d.%s:%d", sharedEntranceIdPrefix, i, sharedZone, app.Spec.SharedEntrances[i].Port)
			} else {
				app.Spec.SharedEntrances[i].URL = fmt.Sprintf("%s%d.%s", sharedEntranceIdPrefix, i, sharedZone)
			}
		}
	}

	return app.Spec.SharedEntrances, nil
}
