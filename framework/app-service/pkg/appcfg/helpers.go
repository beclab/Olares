package appcfg

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
	appv1alpha1 "github.com/beclab/api/api/app.bytetrade.io/v1alpha1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/klog/v2"
)

var getUserZone = kubesphere.GetUserZone

// Helpers in this file replace methods that used to live on the in-tree
// Application / ApplicationManager types (see api/app.bytetrade.io/v1alpha1
// /helper.go before the migration). Because those types are now aliases to
// github.com/beclab/api types, methods cannot be defined on them directly,
// so the helpers are exposed as package-level functions instead.

// DefaultThirdLevelDomainConfig is re-exported here for backwards
// compatibility with call sites that referenced the in-tree alias.
type DefaultThirdLevelDomainConfig struct {
	AppName          string `json:"appName"`
	EntranceName     string `json:"entranceName"`
	ThirdLevelDomain string `json:"thirdLevelDomain"`
}

// IsV3 reports whether the given object (Application or
// ApplicationManager) carries the v3 SCHEMA marker label. This does NOT
// imply the app is shared: a v3 app may be either a shared cluster-wide
// singleton or a regular per-user app depending on options.shared in the
// manifest. Use IsShared to gate shared-app behaviors (admin-only lifecycle,
// cluster-wide visibility, NATS fan-out, etc.).
func IsV3(o metav1.Object) bool {
	return appv1alpha1.IsV3(o)
}

// IsShared reports whether the given object (Application /
// ApplicationManager / namespace / workload) is a shared cluster-wide app.
// The marker is stamped at install time by the v3 install handler when
// ApplicationConfig.Shared is true (apiVersion: v3 + options.shared: true)
// and propagated by the Application controller. v1 and v3+per-user apps do
// NOT carry the label.
func IsShared(o metav1.Object) bool {
	if o == nil {
		return false
	}
	return o.GetLabels()[constants.AppSharedLabel] == constants.AppSharedTrue
}

// IsClusterScoped reports whether the given application is cluster scoped,
// mirroring the old (*Application).IsClusterScoped() method.
func IsClusterScoped(a *Application) bool {
	if a == nil || a.Spec.Settings == nil {
		return false
	}
	if v, ok := a.Spec.Settings["clusterScoped"]; ok && v == "true" {
		return true
	}
	return false
}

// GetAppConfig decodes the JSON-encoded Spec.Config of the given manager into
// appConfig.
func GetAppConfig(a *ApplicationManager, appConfig any) error {
	if err := json.Unmarshal([]byte(a.Spec.Config), appConfig); err != nil {
		klog.Errorf("unmarshal to appConfig failed %v", err)
		return err
	}
	return nil
}

// SetAppConfig marshals appConfig and stores it into a.Spec.Config.
func SetAppConfig(a *ApplicationManager, appConfig any) error {
	configBytes, err := json.Marshal(appConfig)
	if err != nil {
		klog.Errorf("marshal appConfig failed %v", err)
		return err
	}
	a.Spec.Config = string(configBytes)
	return nil
}

// GetMarketSource returns the market source annotation stored on the manager.
func GetMarketSource(a *ApplicationManager) string {
	if a == nil || a.Annotations == nil {
		return ""
	}
	return a.Annotations[constants.AppMarketSourceKey]
}

// appSourceMarket is the Spec.Source value for apps installed from a market
// (matches api.Market). Apps with any other source are treated as uploaded.
const appSourceMarket = "market"

// GetChartOwner returns the owner of the chart the app was installed from,
// used to populate the chartOwner field on push events:
//   - market apps (Spec.Source == "market"): the installing user (AppOwner).
//   - uploaded apps: the user who uploaded the chart, persisted at install
//     time in the origin-owner annotation; falls back to AppOwner when the
//     annotation is absent (e.g. a self-uploaded chart or a pre-existing AM
//     installed before this field was introduced).
func GetChartOwner(a *ApplicationManager) string {
	if a == nil {
		return ""
	}
	if a.Spec.Source != appSourceMarket && a.Labels != nil {
		if owner := a.Labels[constants.AppChartOwnerKey]; owner != "" {
			return owner
		}
	}
	return a.Spec.AppOwner
}

// AppName provides helpers to derive app IDs and to check the well-known
// classes of app names (system, generated, etc.). It is a local named string
// type so helper methods can hang off of it.
type AppName string

// GetAppID returns the stable ID for the given app name. System apps use the
// raw name whereas user apps use the first 8 hex characters of the name's MD5
// digest.
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

// GenEntranceURL fills in entrance URLs on app.Spec.Entrances based on the
// user's zone. It is a package-level re-implementation of the in-tree
// (*Application).GenEntranceURL method.
func GenEntranceURL(ctx context.Context, app *Application) ([]Entrance, error) {
	zone, err := getUserZone(ctx, app.Spec.Owner)
	if err != nil {
		klog.Errorf("failed to get user zone: %v", err)
	}

	if len(zone) > 0 {
		var appDomainConfigs []DefaultThirdLevelDomainConfig
		if defaultThirdLevelDomainConfig, ok := app.Spec.Settings["defaultThirdLevelDomainConfig"]; ok && len(defaultThirdLevelDomainConfig) > 0 {
			if err := json.Unmarshal([]byte(defaultThirdLevelDomainConfig), &appDomainConfigs); err != nil {
				klog.Errorf("unmarshal defaultThirdLevelDomainConfig error %v", err)
				return nil, err
			}
		}

		appid := strings.ToLower(strings.TrimSpace(app.Spec.Appid))
		if len(app.Spec.Entrances) == 1 {
			app.Spec.Entrances[0] = app.Spec.Entrances[0].ForZone(appid, zone, 0, 1)
		} else {
			entrancesForZone := appv1alpha1.Entrances(app.Spec.Entrances).ForZone(appid, zone)
			for i := range app.Spec.Entrances {
				app.Spec.Entrances[i] = entrancesForZone[i]
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

// GenSharedEntranceURL fills in URLs for the app's shared entrances.
func GenSharedEntranceURL(ctx context.Context, app *Application) ([]Entrance, error) {
	zone, err := getUserZone(ctx, app.Spec.Owner)
	if err != nil {
		klog.Errorf("failed to get user zone: %v", err)
	}

	if len(zone) > 0 {
		tokens := strings.Split(zone, ".")
		tokens[0] = "shared"
		sharedZone := strings.Join(tokens, ".")

		appid := strings.ToLower(strings.TrimSpace(app.Spec.Appid))
		sharedEntrancesForZone := appv1alpha1.Entrances(app.Spec.SharedEntrances).SharedForZone(appid, sharedZone)
		for i := range sharedEntrancesForZone {
			app.Spec.SharedEntrances[i] = sharedEntrancesForZone[i]
			if app.Spec.SharedEntrances[i].Port > 0 {
				app.Spec.SharedEntrances[i].URL = fmt.Sprintf("%s:%d", app.Spec.SharedEntrances[i].URL, app.Spec.SharedEntrances[i].Port)
			}
		}
	}

	return app.Spec.SharedEntrances, nil
}
