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

// SharedEntranceIdPrefix returns the 8-char md5 prefix used as the base for
// shared entrance URLs for the app.
func (s AppName) SharedEntranceIdPrefix() string {
	hash := md5.Sum([]byte(s.GetAppID() + "shared"))
	hashString := hex.EncodeToString(hash[:])
	return hashString[:8]
}

// EIDError carries structured error codes from SharedEntranceID-family helpers.
type EIDError struct {
	Code string
	Msg  string
}

func (e *EIDError) Error() string { return e.Code + ": " + e.Msg }

// SharedEntranceID returns the first DNS label for v3 shared entrances:
// appid when count==1, appid+index when count>1.
func SharedEntranceID(appid string, entranceIndex, entranceCount int) (string, error) {
	appid = strings.ToLower(strings.TrimSpace(appid))
	if appid == "" {
		return "", &EIDError{Code: "EID_EMPTY_APPID", Msg: "appid is empty"}
	}
	if entranceCount < 1 || entranceCount > 10 {
		return "", &EIDError{Code: "EID_TOO_MANY_ENTRANCES", Msg: "entrance count must be in [1,10]"}
	}
	if entranceIndex < 0 || entranceIndex >= entranceCount {
		return "", &EIDError{Code: "EID_INDEX_OUT_OF_RANGE", Msg: "entrance index is out of range"}
	}
	if entranceCount == 1 {
		return appid, nil
	}
	return fmt.Sprintf("%s%d", appid, entranceIndex), nil
}

// GenSharedEntranceURLForUser composes the per-viewer Shared entrance URL.
func GenSharedEntranceURLForUser(appid string, entranceIndex, entranceCount int,
	viewer, platformDomain string) (string, error) {
	viewer = strings.ToLower(strings.TrimSpace(viewer))
	platformDomain = strings.ToLower(strings.TrimSpace(strings.TrimSuffix(platformDomain, ".")))
	if viewer == "" || platformDomain == "" {
		return "", &EIDError{Code: "EID_INCOMPLETE_URL_INPUT", Msg: "viewer or platformDomain is empty"}
	}
	entranceID, err := SharedEntranceID(appid, entranceIndex, entranceCount)
	if err != nil {
		return "", err
	}
	return "https://" + entranceID + "." + viewer + "." + platformDomain, nil
}

// LogicalHostPattern returns the SRR hostPattern for a shared entrance
// One logical pattern covers all viewers of an entrance:
//
//	<entranceid>.*.<platformDomain>
//
// The literal "*" segment is the marker app-service route control uses when
// building HTTPRoute hostnames + Host RegularExpression header match.
func LogicalHostPattern(appid string, entranceIndex, entranceCount int, platformDomain string) (string, error) {
	platformDomain = strings.ToLower(strings.TrimSpace(strings.TrimSuffix(platformDomain, ".")))
	if platformDomain == "" {
		return "", &EIDError{Code: "EID_INCOMPLETE_URL_INPUT", Msg: "platformDomain is empty"}
	}
	entranceID, err := SharedEntranceID(appid, entranceIndex, entranceCount)
	if err != nil {
		return "", err
	}
	return entranceID + ".*." + platformDomain, nil
}

// GenEntranceURL fills in entrance URLs on app.Spec.Entrances based on the
// user's zone. It is a package-level re-implementation of the in-tree
// (*Application).GenEntranceURL method.
func GenEntranceURL(ctx context.Context, app *Application) ([]Entrance, error) {
	zone, err := kubesphere.GetUserZone(ctx, app.Spec.Owner)
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

// GenSharedEntranceURL fills in URLs for the app's shared entrances.
func GenSharedEntranceURL(ctx context.Context, app *Application) ([]Entrance, error) {
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
