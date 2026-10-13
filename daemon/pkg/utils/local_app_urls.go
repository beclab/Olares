package utils

import (
	"encoding/json"
	"fmt"
	"strings"

	appv1 "github.com/beclab/api/api/app.bytetrade.io/v1alpha1"
)

// localAppEntrances uses the machine's local zone rather than the viewer's
// optional public zone. Effective settings keep shared per-user overrides in
// sync with the proxy; the informer object is never modified.
func localAppEntrances(app *appv1.Application, viewer, zone string) ([]appv1.Entrance, error) {
	entrances := app.EffectiveEntrances(viewer)
	settings := app.EffectiveSettings(viewer)
	var domains []appv1.DefaultThirdLevelDomainConfig
	if raw := settings["defaultThirdLevelDomainConfig"]; raw != "" {
		if err := json.Unmarshal([]byte(raw), &domains); err != nil {
			return nil, err
		}
	}
	refs := make([]*appv1.Entrance, len(entrances))
	for i := range entrances {
		refs[i] = &entrances[i]
	}
	appid := strings.ToLower(strings.TrimSpace(app.Spec.Appid))
	for i := range entrances {
		prefix := appv1.ResolveEntranceIDWithDefaultThirdLevelDomainOverride(refs, i, appid, domains)
		entrances[i].URL = fmt.Sprintf("%s.%s", prefix, zone)
	}
	return entrances, nil
}
