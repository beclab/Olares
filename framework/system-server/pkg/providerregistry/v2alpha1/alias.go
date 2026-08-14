package v2alpha1

import (
	"encoding/json"
	"strings"

	appv1alpha1 "github.com/beclab/api/api/app.bytetrade.io/v1alpha1"
	"k8s.io/apimachinery/pkg/labels"
	"k8s.io/klog/v2"
)

const (
	settingsKeyCustomDomain                  = "customDomain"
	settingsKeyDefaultThirdLevelDomainConfig = "defaultThirdLevelDomainConfig"
	settingsCustomDomainThirdLevelDomain     = "third_level_domain"
	settingsCustomDomainThirdPartyDomain     = "third_party_domain"
)

// ApplicationLister lists Application CRs. Implemented by the generated
// ApplicationLister from github.com/beclab/api.
type ApplicationLister interface {
	List(selector labels.Selector) ([]*appv1alpha1.Application, error)
}

// ResolveProviderRefFromHost resolves a request host to the provider-ref that
// ClusterRoles are annotated with at registration time.
//
// Registration only records the canonical entrance URL (appid /
// defaultThirdLevelDomain). User-set customDomain aliases
// (third_level_domain / third_party_domain) are resolved by scanning
// Application CRs; if no alias matches, falls back to ProviderRefFromHost.
func ResolveProviderRefFromHost(host string, apps ApplicationLister) string {
	host = normalizeHost(host)
	if host == "" {
		return ProviderRefFromHost(host)
	}

	if apps != nil {
		if ref, ok := resolveProviderRefFromAlias(host, apps); ok {
			return ref
		}
	}

	return ProviderRefFromHost(host)
}

func resolveProviderRefFromAlias(host string, apps ApplicationLister) (string, bool) {
	list, err := apps.List(labels.Everything())
	if err != nil {
		klog.Errorf("ResolveProviderRefFromHost: list applications: %v", err)
		return "", false
	}

	hostLabels := strings.Split(host, ".")
	var thirdLevelPrefix, zoneUser string
	if len(hostLabels) >= 2 {
		thirdLevelPrefix = hostLabels[0]
		zoneUser = hostLabels[1]
	}

	for _, app := range list {
		if app == nil {
			continue
		}
		for _, owned := range allCustomDomainBlobs(app) {
			customDomainMap := parseCustomDomainBlob(owned.blob)
			if len(customDomainMap) == 0 {
				continue
			}
			for entranceName, cfg := range customDomainMap {
				if cfg == nil {
					continue
				}
				matched := false
				if thirdParty := strings.ToLower(strings.TrimSpace(cfg[settingsCustomDomainThirdPartyDomain])); thirdParty != "" && thirdParty == host {
					matched = true
				}
				if !matched && thirdLevelPrefix != "" {
					thirdLevel := strings.ToLower(strings.TrimSpace(cfg[settingsCustomDomainThirdLevelDomain]))
					if thirdLevel != "" && thirdLevel == thirdLevelPrefix && zoneUser == strings.ToLower(owned.zoneUser) {
						matched = true
					}
				}
				if !matched {
					continue
				}
				if ref, ok := canonicalProviderRef(app, entranceName); ok {
					klog.V(5).Infof("ResolveProviderRefFromHost: host %q alias of entrance %q app %q -> %q", host, entranceName, app.Spec.Name, ref)
					return ref, true
				}
			}
		}
	}

	return "", false
}

type ownedCustomDomainBlob struct {
	// zoneUser is the user whose zone this blob resolves under:
	// Spec.Settings uses app.Spec.Owner; UserSettings[user] uses user.
	zoneUser string
	blob     string
}

func allCustomDomainBlobs(app *appv1alpha1.Application) []ownedCustomDomainBlob {
	blobs := make([]ownedCustomDomainBlob, 0)
	if app.Spec.Settings != nil {
		if b := app.Spec.Settings[settingsKeyCustomDomain]; b != "" {
			blobs = append(blobs, ownedCustomDomainBlob{zoneUser: app.Spec.Owner, blob: b})
		}
	}
	for user, us := range app.Spec.UserSettings {
		if us == nil {
			continue
		}
		if b := us[settingsKeyCustomDomain]; b != "" {
			blobs = append(blobs, ownedCustomDomainBlob{zoneUser: user, blob: b})
		}
	}
	return blobs
}

func parseCustomDomainBlob(blob string) map[string]map[string]string {
	var out map[string]map[string]string
	if err := json.Unmarshal([]byte(blob), &out); err != nil {
		return nil
	}
	return out
}

func canonicalProviderRef(app *appv1alpha1.Application, entranceName string) (string, bool) {
	owner := strings.TrimSpace(app.Spec.Owner)
	if owner == "" {
		return "", false
	}

	entrances := app.Spec.Entrances
	index := -1
	for i := range entrances {
		if entrances[i].Name == entranceName {
			index = i
			break
		}
	}
	if index < 0 {
		return "", false
	}

	var configs []appv1alpha1.DefaultThirdLevelDomainConfig
	if raw := app.Spec.Settings[settingsKeyDefaultThirdLevelDomainConfig]; raw != "" {
		if err := json.Unmarshal([]byte(raw), &configs); err != nil {
			klog.Warningf("ResolveProviderRefFromHost: parse defaultThirdLevelDomainConfig for app %s: %v", app.Spec.Name, err)
		}
	}

	appid := strings.ToLower(strings.TrimSpace(app.Spec.Appid))
	ptrs := make([]*appv1alpha1.Entrance, len(entrances))
	for i := range entrances {
		ptrs[i] = &entrances[i]
	}
	prefix := appv1alpha1.ResolveEntranceIDWithDefaultThirdLevelDomainOverride(ptrs, index, appid, configs)
	if prefix == "" {
		return "", false
	}

	return ProviderRefName(prefix, owner), true
}

func normalizeHost(host string) string {
	host = strings.TrimSpace(host)
	host = strings.Split(host, ":")[0]
	return strings.ToLower(host)
}
