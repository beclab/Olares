package apiserver

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io/ioutil"
	"strconv"
	"strings"
	"time"

	"github.com/beclab/Olares/framework/app-service/pkg/apiserver/api"
	"github.com/beclab/Olares/framework/app-service/pkg/appcfg"
	"github.com/beclab/Olares/framework/app-service/pkg/appinstaller"
	"github.com/beclab/Olares/framework/app-service/pkg/appstate"
	"github.com/beclab/Olares/framework/app-service/pkg/client/clientset"
	"github.com/beclab/Olares/framework/app-service/pkg/constants"
	"github.com/beclab/Olares/framework/app-service/pkg/kubesphere"
	"github.com/beclab/Olares/framework/app-service/pkg/provider"
	"github.com/beclab/Olares/framework/app-service/pkg/tapr"
	apputils "github.com/beclab/Olares/framework/app-service/pkg/utils/app"
	"github.com/beclab/api/api/app.bytetrade.io/v1alpha1"
	"github.com/beclab/api/pkg/generated/clientset/versioned"

	"github.com/emicklei/go-restful/v3"
	kerrors "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/apimachinery/pkg/types"
	"k8s.io/client-go/dynamic"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/util/retry"
	"k8s.io/klog/v2"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

// appLevelSettingKeys lists Spec.Settings keys whose values are scoped
// to the Application as a whole — shared by every user — rather than
// to an individual caller. When any settings handler is about to write
// one of these keys, it lands in Spec.Settings even for v3 (shared)
// apps, overriding the default v3 routing into Spec.UserSettings[caller].
// App-level writes are still gated by gateSharedAppWrite, so on v3 only
// admin/owner can mutate them; on v1/v2 the per-app owner check applies
// as before.
var appLevelSettingKeys = map[string]struct{}{
	"enableOverlayGateway": {},
	"enableLLMGateway":     {},
}

// isAppLevelSettingKey reports whether the given Spec.Settings key is
// scoped to the Application as a whole and must always be persisted in
// Spec.Settings rather than the per-user Spec.UserSettings overlay.
func isAppLevelSettingKey(k string) bool {
	_, ok := appLevelSettingKeys[k]
	return ok
}

// overridePatch builds a JSON merge-patch that stores the given override
// key/values in the correct slot: the per-user Spec.UserSettings[caller] map
// for shared (v3) apps, or the app-global Spec.Settings map for non-shared
// (v1/v2/v3) apps. This mirrors where EffectiveEntrances/EffectiveSettings read
// the overrides from, so writes and reads stay consistent. A merge-patch on the
// scoped sub-map keeps other users' entries and other keys untouched.
func overridePatch(app *v1alpha1.Application, caller string, kv map[string]string) map[string]interface{} {
	if v1alpha1.IsShared(app) {
		return map[string]interface{}{
			"spec": map[string]interface{}{
				"userSettings": map[string]interface{}{
					caller: kv,
				},
			},
		}
	}
	return map[string]interface{}{
		"spec": map[string]interface{}{
			"settings": kv,
		},
	}
}

func (h *Handler) setupApp(req *restful.Request, resp *restful.Response) {
	app, err := getAppByName(req, resp)
	if err != nil {
		klog.Errorf("Failed to get app name=%s err=%v", app.Spec.Name, err)
		// if error, response in function. Do nothing
		return
	}
	if !h.gateSharedAppWrite(req, resp, app) {
		return
	}

	bodyData, err := ioutil.ReadAll(req.Request.Body)
	if err != nil {
		api.HandleError(resp, req, err)
		return
	}

	var settings map[string]interface{}
	err = json.Unmarshal(bodyData, &settings)
	if err != nil {
		api.HandleError(resp, req, err)
		return
	}

	appCopy := app.DeepCopy()
	caller := req.Attribute(constants.UserContextAttribute).(string)
	shared := v1alpha1.IsShared(appCopy)

	// TODO: validate settings keys
	for k, v := range settings {
		var str []byte
		switch v.(type) {
		case map[string]interface{}:
			str, err = json.Marshal(v)
			if err != nil {
				api.HandleError(resp, req, err)
				return
			}
		default:
			str = []byte(v.(string))
		}
		// For shared (v3) apps, non-app-level keys are per-user overrides and
		// land in Spec.UserSettings[caller] so they are layered per user by
		// EffectiveSettings. App-level keys, and every key on non-shared apps,
		// live directly in the app-global Spec.Settings.
		if shared && !isAppLevelSettingKey(k) {
			if appCopy.Spec.UserSettings == nil {
				appCopy.Spec.UserSettings = map[string]map[string]string{}
			}
			if appCopy.Spec.UserSettings[caller] == nil {
				appCopy.Spec.UserSettings[caller] = map[string]string{}
			}
			appCopy.Spec.UserSettings[caller][k] = string(str)
			continue
		}
		if appCopy.Spec.Settings == nil {
			appCopy.Spec.Settings = map[string]string{}
		}
		appCopy.Spec.Settings[k] = string(str)
	}
	client := req.Attribute(constants.KubeSphereClientAttribute).(*clientset.ClientSet)

	appUpdated, err := client.AppClient.AppV1alpha1().Applications().Update(req.Request.Context(), appCopy, metav1.UpdateOptions{})
	if err != nil {
		api.HandleError(resp, req, err)
		return
	}
	// Respond with the caller's effective view so the response reflects
	// both the global Spec.Settings and any per-user overlay just
	// written. For v1/v2 apps EffectiveSettings is just a copy of
	// Spec.Settings, so v1/v2 callers see the same shape as before.
	resp.WriteAsJson(appUpdated.EffectiveSettings(caller))
}

func (h *Handler) setupAppEntranceDomain(req *restful.Request, resp *restful.Response) {
	app, err := getAppByName(req, resp)
	if err != nil {
		klog.Errorf("Failed to get app name=%s err=%v", req.PathParameter(ParamAppName), err)
		// if error, response in function. Do nothing
		return
	}
	// Per-user settings handlers do NOT call gateSharedAppWrite: every
	// authenticated user may set their own customDomain for a v3 shared
	// app. Global Settings mutations stay admin-only via setupApp.
	caller := req.Attribute(constants.UserContextAttribute).(string)
	entranceName := req.PathParameter(ParamEntranceName)
	// Validate against the caller's effective entrances so user-added entrances
	// (e.g. proxylistener-managed "dev-<port>" entrances that live only in the
	// addedEntrances overlay) are accepted too — not just chart-base entrances.
	validName := false
	isDevEntrance := false
	for _, e := range app.EffectiveEntrances(caller) {
		if e.Name == entranceName {
			validName = true
			isDevEntrance = e.Type == constants.EntranceTypeDev
			break
		}
	}
	if !validName {
		api.HandleBadRequest(resp, req, errors.New("invalid entrance name"))
		return
	}
	// A dev entrance (proxylistener-managed "dev-<port>") is reached directly by
	// pod IP, so it cannot own a custom third_party_domain / third_level_domain.
	if isDevEntrance {
		api.HandleBadRequest(resp, req, fmt.Errorf("entrance %q is a dev entrance and cannot set a custom domain", entranceName))
		return
	}

	bodyData, err := ioutil.ReadAll(req.Request.Body)
	if err != nil {
		api.HandleError(resp, req, err)
		return
	}

	var settings map[string]interface{}
	err = json.Unmarshal(bodyData, &settings)
	if err != nil {
		api.HandleError(resp, req, err)
		return
	}
	appCopy := app.DeepCopy()

	kclient := req.Attribute(constants.KubeSphereClientAttribute).(*clientset.ClientSet)

	customDomain, ok := settings["customDomain"].(map[string]interface{})

	existing := appCopy.EffectiveSettings(caller)["customDomain"]

	merge := make(map[string]interface{})

	keys := []string{"third_level_domain", "third_party_domain"}

	// Reject the request if the caller is trying to set a domain that is
	// already taken. third_party_domain (a full custom domain) must be
	// globally unique across every user and app; third_level_domain only
	// needs to be unique within the caller's own apps because it resolves
	// to "<prefix>.<caller-zone>" and the zone is bound to the user. Only
	// the values the caller explicitly provided are checked so a re-save
	// of the untouched sibling field is not flagged against itself.
	if ok {
		reqThirdLevel, _ := customDomain["third_level_domain"].(string)
		reqThirdParty, _ := customDomain["third_party_domain"].(string)
		if err = checkEntranceDomainDuplicate(req.Request.Context(), kclient.AppClient, caller, appCopy.Spec.Name, entranceName, reqThirdLevel, reqThirdParty); err != nil {
			api.HandleBadRequest(resp, req, err)
			return
		}
	}

	if len(existing) > 0 {
		var origins map[string]interface{}
		err = json.Unmarshal([]byte(existing), &origins)
		if err != nil {
			api.HandleError(resp, req, err)
			return
		}
		// do a merge
		// origins {"a":{"third_level_domain":"","third_party__domain":""},"b":{"third_level_domain":"","third_party__domain":""}}
		// {"third_level_domain":"","third_party__domain":""}
		for k, v := range origins {
			originV := v.(map[string]interface{})
			if k != entranceName {
				merge[k] = originV
				continue
			} else {
				for _, key := range keys {
					if ov, ok := originV[key]; ok {
						if _, exists := customDomain[key]; !exists {
							customDomain[key] = ov
						}
					}
				}
			}
		}
	}
	for _, key := range keys {
		if _, exists := customDomain[key]; !exists {
			customDomain[key] = ""
		}
	}
	merge[entranceName] = customDomain

	settingsBytes, err := json.Marshal(merge)
	if err != nil {
		api.HandleError(resp, req, err)
		return
	}

	// customDomain edits land in the override slot: per-user
	// (Spec.UserSettings[caller]) for shared apps, or the app-global
	// Spec.Settings for non-shared apps. A JSON merge-patch leaves other users'
	// entries and other keys untouched.
	patchData := overridePatch(appCopy, caller, map[string]string{
		"customDomain": string(settingsBytes),
	})
	patchByte, err := json.Marshal(patchData)
	if err != nil {
		api.HandleError(resp, req, err)
		return
	}

	appUpdated, err := kclient.AppClient.AppV1alpha1().Applications().Patch(req.Request.Context(), appCopy.Name, types.MergePatchType, patchByte, metav1.PatchOptions{})
	if err != nil {
		api.HandleError(resp, req, err)
		return
	}

	if ok {
		// upgrade app set values
		owner := req.Attribute(constants.UserContextAttribute).(string)
		zone, err := kubesphere.GetUserZone(req.Request.Context(), owner)
		if err != nil {
			api.HandleError(resp, req, err)
			return
		}

		token, err := h.GetUserServiceAccountToken(req.Request.Context(), owner)
		if err != nil {
			klog.Error("Failed to get user service account token: ", err)
			api.HandleError(resp, req, err)
			return
		}

		vals := make(map[string]interface{})
		entries := make(map[string]interface{})
		for i, entrance := range app.Spec.Entrances {
			cfg, ok := customDomain[entrance.Name].(map[string]interface{})
			if !ok {
				continue
			}
			urls := make([]string, 0)
			if cDomain, _ := cfg["third_party_domain"].(string); cDomain != "" {
				urls = append(urls, cDomain)
			}
			if prefix := cfg["third_level_domain"]; prefix != "" {
				urls = append(urls, fmt.Sprintf("%s.%s", prefix, zone))
			}
			var url string
			if len(app.Spec.Entrances) == 1 {
				url = fmt.Sprintf("%s.%s", app.Spec.Appid, zone)
			} else {
				url = fmt.Sprintf("%s%d.%s", app.Spec.Appid, i, zone)
			}
			urls = append(urls, url)

			entries[entrance.Name] = strings.Join(urls, ",")
		}
		vals["domain"] = entries

		appMgr, err := kclient.AppClient.AppV1alpha1().ApplicationManagers().Get(req.Request.Context(), appUpdated.Name, metav1.GetOptions{})
		if err != nil {
			api.HandleError(resp, req, err)
			return
		}

		var appCfg *appcfg.ApplicationConfig
		err = json.Unmarshal([]byte(appMgr.Spec.Config), &appCfg)
		if err != nil {
			api.HandleError(resp, req, err)
			return
		}
		if !appstate.IsOperationAllowed(appMgr.Status.State, v1alpha1.UpgradeOp) {
			err = fmt.Errorf("%s operation is not allowed for %s state", v1alpha1.UpgradeOp, appMgr.Status.State)
			api.HandleBadRequest(resp, req, err)
			return
		}

		appMgrCopy := appMgr.DeepCopy()
		appMgrCopy.Annotations[api.AppTokenKey] = token

		err = h.ctrlClient.Patch(req.Request.Context(), appMgrCopy, client.MergeFrom(appMgr))
		if err != nil {
			api.HandleError(resp, req, err)
			return
		}

		now := metav1.Now()
		opID := strconv.FormatInt(time.Now().Unix(), 10)

		status := v1alpha1.ApplicationManagerStatus{
			OpType:     v1alpha1.UpgradeOp,
			OpID:       opID,
			State:      v1alpha1.Upgrading,
			Message:    fmt.Sprintf("app %s was upgrade via setup domain by user %s", appMgrCopy.Spec.AppName, appMgrCopy.Spec.AppOwner),
			StatusTime: &now,
			UpdateTime: &now,
		}

		_, err = apputils.UpdateAppMgrStatus(appMgr.Name, status)
		if err != nil {
			api.HandleError(resp, req, err)
			return
		}
	}
	// Respond with the caller's effective view so the UI sees their own
	// customDomain entries (the global Spec.Settings is admin-only for v3).
	resp.WriteAsJson(appUpdated.EffectiveSettings(caller))
}

// reservedThirdLevelDomains are subdomain prefixes owned by system apps
// (auth.<zone>, desktop.<zone>, wizard-<user>.<zone>, ...) that a caller
// must never be able to claim as a custom third_level_domain. Keys are
// lower-cased for case-insensitive matching.
var reservedThirdLevelDomains = map[string]struct{}{
	"auth":    {},
	"desktop": {},
	"wizard":  {},
}

// entranceCustomDomain is the per-entrance shape stored under the
// "customDomain" settings key: {"third_level_domain": "", "third_party_domain": ""}.
type entranceCustomDomain struct {
	thirdLevel string
	thirdParty string
}

// parseCustomDomain decodes a "customDomain" settings blob
// ({"<entrance>": {"third_level_domain": "", "third_party_domain": ""}})
// into a per-entrance map. A malformed or empty blob yields an empty map.
func parseCustomDomain(blob string) map[string]entranceCustomDomain {
	out := make(map[string]entranceCustomDomain)
	if blob == "" {
		return out
	}
	var raw map[string]map[string]interface{}
	if err := json.Unmarshal([]byte(blob), &raw); err != nil {
		klog.Warningf("failed to parse customDomain blob for duplicate check: %v", err)
		return out
	}
	for entrance, cfg := range raw {
		tl, _ := cfg["third_level_domain"].(string)
		tp, _ := cfg["third_party_domain"].(string)
		out[entrance] = entranceCustomDomain{thirdLevel: tl, thirdParty: tp}
	}
	return out
}

// ownedCustomDomainBlob is a customDomain blob tagged with the user it belongs
// to: the install owner for the global Spec.Settings entry, or the map key for
// a per-user Spec.UserSettings overlay.
type ownedCustomDomainBlob struct {
	owner string
	blob  string
}

// allCustomDomainBlobs returns every customDomain blob stored on the app —
// the global Spec.Settings entry (owner = install owner) plus every per-user
// overlay in Spec.UserSettings (owner = the user key) — each tagged with its
// owner. Used for the global third_party_domain uniqueness scan, where the
// owner is needed to skip only the caller's own entry: a shared app lets each
// user set their own third_party per entrance, so other users' overlays on the
// same entrance must still be enforced.
func allCustomDomainBlobs(app *v1alpha1.Application) []ownedCustomDomainBlob {
	blobs := make([]ownedCustomDomainBlob, 0)
	if b := app.Spec.Settings["customDomain"]; b != "" {
		blobs = append(blobs, ownedCustomDomainBlob{owner: app.Spec.Owner, blob: b})
	}
	for user, us := range app.Spec.UserSettings {
		if b := us["customDomain"]; b != "" {
			blobs = append(blobs, ownedCustomDomainBlob{owner: user, blob: b})
		}
	}
	return blobs
}

// callerCustomDomainBlob returns the customDomain blob that resolves under the
// caller's zone for the given app. It must match the effective view the domain
// handler itself uses (EffectiveSettings(caller)):
//   - shared app: the caller's Spec.UserSettings[caller] overlay when present,
//     otherwise the global Spec.Settings default — both are effective in the
//     caller's zone, so a prefix present only in the global blob still counts;
//   - per-user app the caller owns: the global Spec.Settings entry.
//
// Other users' entries live under their own zones and are irrelevant to the
// caller's third_level_domain scope, so per-user apps not owned by the caller
// contribute nothing.
func callerCustomDomainBlob(app *v1alpha1.Application, caller string) string {
	if appcfg.IsShared(app) {
		// EffectiveSettings overlays UserSettings[caller] onto Spec.Settings,
		// falling back to the global customDomain when the caller has no
		// per-user override — exactly what resolves in the caller's zone.
		return app.EffectiveSettings(caller)["customDomain"]
	}
	if app.Spec.Owner == caller {
		// Non-shared customDomain lives directly in Spec.Settings;
		// EffectiveSettings returns a copy of it.
		return app.EffectiveSettings(caller)["customDomain"]
	}
	return ""
}

// defaultThirdLevelPrefixes returns the live default third-level subdomain
// prefix of every entrance of the app, mirroring the URL generation in
// GenEntranceURL / EntrancesWithZone (and the gateway's resolveEntrancePrefix):
//   - a single-entrance app always uses the bare "<appid>" and ignores any
//     defaultThirdLevelDomainConfig override;
//   - a multi-entrance app uses the configured thirdLevelDomain for an entrance
//     when present, otherwise the positional "<appid><i>".
//
// These are the domains a caller-supplied third_level_domain must not collide
// with. Resolving per entrance (rather than pre-filling then overriding) keeps
// the result limited to entrances that actually exist on the app.
func defaultThirdLevelPrefixes(app *v1alpha1.Application) map[string]string {
	out := make(map[string]string)
	appid := strings.ToLower(strings.TrimSpace(app.Spec.Appid))
	if appid == "" {
		return out
	}

	var cfgs []appcfg.DefaultThirdLevelDomainConfig
	if raw := app.Spec.Settings["defaultThirdLevelDomainConfig"]; raw != "" {
		if err := json.Unmarshal([]byte(raw), &cfgs); err != nil {
			klog.Warningf("failed to parse defaultThirdLevelDomainConfig for duplicate check: %v", err)
		}
	}

	for i := range app.Spec.Entrances {
		out[app.Spec.Entrances[i].Name] = resolveDefaultThirdLevelPrefix(app.Spec.Entrances, i, appid, app.Spec.Name, cfgs)
	}
	return out
}

func resolveDefaultThirdLevelPrefix(entrances []v1alpha1.Entrance, index int, appid, appName string, cfgs []appcfg.DefaultThirdLevelDomainConfig) string {
	for _, cfg := range cfgs {
		if cfg.AppName == appName && cfg.EntranceName == entrances[index].Name && cfg.ThirdLevelDomain != "" {
			return cfg.ThirdLevelDomain
		}
	}
	if len(entrances) == 1 {
		return appid
	}
	return fmt.Sprintf("%s%d", appid, index)
}

// checkEntranceDomainDuplicate reports an error when the requested
// third_level_domain / third_party_domain for (currentApp, currentEntrance)
// is already taken. thirdLevel/thirdParty are the raw requested values; an
// empty value is skipped. The (currentApp, currentEntrance) entrance itself
// is excluded so re-saving an unchanged value is never a conflict.
func checkEntranceDomainDuplicate(ctx context.Context, appClient versioned.Interface, caller, currentApp, currentEntrance, thirdLevel, thirdParty string) error {
	thirdLevel = strings.TrimSpace(thirdLevel)
	thirdParty = strings.TrimSpace(thirdParty)
	if thirdLevel == "" && thirdParty == "" {
		return nil
	}

	applist, err := appClient.AppV1alpha1().Applications().List(ctx, metav1.ListOptions{})
	if err != nil {
		return err
	}

	return findEntranceDomainConflict(applist.Items, caller, currentApp, currentEntrance, thirdLevel, thirdParty)
}

// domainRecord is one already-claimed entrance domain, flattened out of the
// whole cluster so the conflict check is a single linear scan.
//   - thirdParty is compared globally (across every user and app);
//   - thirdLevel is only compared when inCallerZone, i.e. the value resolves
//     under the caller's own zone, because third_level_domain expands to
//     "<prefix>.<caller-zone>" and the zone is bound to the user;
//   - isSelf marks the exact entry the caller is editing (same app + entrance,
//     owned by the caller) so re-saving an unchanged value never conflicts;
//   - isDefault marks a synthesized default subdomain (appid / appid<i>) rather
//     than a user-set value, so the error can say so.
type domainRecord struct {
	app          string
	entrance     string
	thirdLevel   string
	thirdParty   string
	inCallerZone bool
	isSelf       bool
	isDefault    bool
}

// collectDomainRecords flattens every claimed entrance domain in the cluster
// into a single slice, tagging each record with the scope flags the conflict
// scan needs (see domainRecord). currentApp/currentEntrance identify the entry
// being edited so it can be excluded.
func collectDomainRecords(apps []v1alpha1.Application, caller, currentApp, currentEntrance string) []domainRecord {
	records := make([]domainRecord, 0)
	for i := range apps {
		app := &apps[i]

		// Global third_party scope: every stored blob, tagged with the user
		// that actually set it so only the caller's own target entry is
		// treated as self (a same-named install owned by another user, or
		// another user's overlay on a shared app, still counts).
		for _, ob := range allCustomDomainBlobs(app) {
			for entrance, cfg := range parseCustomDomain(ob.blob) {
				records = append(records, domainRecord{
					app:        app.Spec.Name,
					entrance:   entrance,
					thirdParty: cfg.thirdParty,
					isSelf:     app.Spec.Name == currentApp && entrance == currentEntrance && ob.owner == caller,
				})
			}
		}

		// Caller-zone third_level scope: the blob effective in the caller's
		// zone (their overlay, or the shared/global default they inherit).
		for entrance, cfg := range parseCustomDomain(callerCustomDomainBlob(app, caller)) {
			records = append(records, domainRecord{
				app:          app.Spec.Name,
				entrance:     entrance,
				thirdLevel:   cfg.thirdLevel,
				inCallerZone: true,
				isSelf:       app.Spec.Name == currentApp && entrance == currentEntrance,
			})
		}

		// Default subdomains (appid / appid<i>) resolve in the caller's zone
		// for shared apps and for per-user apps the caller owns; other users'
		// per-user apps render under their own zones and cannot collide.
		if appcfg.IsShared(app) || app.Spec.Owner == caller {
			for entrance, prefix := range defaultThirdLevelPrefixes(app) {
				records = append(records, domainRecord{
					app:          app.Spec.Name,
					entrance:     entrance,
					thirdLevel:   prefix,
					inCallerZone: true,
					isDefault:    true,
					isSelf:       app.Spec.Name == currentApp && entrance == currentEntrance,
				})
			}
		}
	}
	return records
}

// findEntranceDomainConflict is the pure comparison core of
// checkEntranceDomainDuplicate: given the full set of applications it reports
// the first conflict for the requested third_level_domain / third_party_domain
// of (currentApp, currentEntrance). thirdLevel/thirdParty are already trimmed
// and at least one is non-empty. The (currentApp, currentEntrance) entrance is
// excluded so re-saving an unchanged value is never a conflict.
func findEntranceDomainConflict(apps []v1alpha1.Application, caller, currentApp, currentEntrance, thirdLevel, thirdParty string) error {
	// Reserved system subdomains (auth.<zone>, desktop.<zone>,
	// wizard-<user>.<zone>, ...) can never be claimed as a custom
	// third-level domain regardless of which apps exist.
	if thirdLevel != "" {
		if _, ok := reservedThirdLevelDomains[strings.ToLower(thirdLevel)]; ok {
			return fmt.Errorf("third_level_domain %q is reserved and cannot be used", thirdLevel)
		}
	}

	for _, r := range collectDomainRecords(apps, caller, currentApp, currentEntrance) {
		if r.isSelf {
			continue
		}
		if thirdParty != "" && r.thirdParty != "" && strings.EqualFold(r.thirdParty, thirdParty) {
			return fmt.Errorf("third_party_domain %q is already used by entrance %q of app %q", thirdParty, r.entrance, r.app)
		}
		if thirdLevel != "" && r.inCallerZone && r.thirdLevel != "" && strings.EqualFold(r.thirdLevel, thirdLevel) {
			if r.isDefault {
				return fmt.Errorf("third_level_domain %q conflicts with the default domain of entrance %q of app %q", thirdLevel, r.entrance, r.app)
			}
			return fmt.Errorf("third_level_domain %q is already used by entrance %q of app %q", thirdLevel, r.entrance, r.app)
		}
	}
	return nil
}

func (h *Handler) getAppEntrances(req *restful.Request, resp *restful.Response) {
	app, err := getAppByName(req, resp)
	if err != nil {
		klog.Errorf("Failed to get app name=%s err=%v", app.Spec.Name, err)
		// if error, response in function. Do nothing
		return
	}
	caller := req.Attribute(constants.UserContextAttribute).(string)
	resp.WriteAsJson(app.EffectiveEntrances(caller))
}

func (h *Handler) getAppEntrancesSettings(req *restful.Request, resp *restful.Response) {
	app, err := getAppByName(req, resp)
	if err != nil {
		klog.Errorf("Failed to get app name=%s err=%v", app.Spec.Name, err)
		// if error, response in function. Do nothing
		return
	}
	caller := req.Attribute(constants.UserContextAttribute).(string)
	resp.WriteAsJson(app.EffectiveSettings(caller))
}

func (h *Handler) getAppSettings(req *restful.Request, resp *restful.Response) {
	app, err := getAppByName(req, resp)
	if err != nil {
		klog.Errorf("Failed to get app name=%s err=%v", app.Spec.Name, err)
		// if error, response in function. Do nothing
		return
	}
	caller := req.Attribute(constants.UserContextAttribute).(string)
	resp.WriteAsJson(app.EffectiveSettings(caller))
}

func getRepoURL() string {
	return constants.CHART_REPO_URL
}

func (h *Handler) setupAppAuthLevel(req *restful.Request, resp *restful.Response) {
	app, err := getAppByName(req, resp)
	if err != nil {
		klog.Errorf("Failed to get app name=%s err=%v", req.PathParameter(ParamAppName), err)
		// if error, response in function. Do nothing
		return
	}

	entranceName := req.PathParameter(ParamEntranceName)

	bodyData, err := ioutil.ReadAll(req.Request.Body)
	if err != nil {
		api.HandleError(resp, req, err)
		return
	}

	var data map[string]map[string]string
	if err = json.Unmarshal(bodyData, &data); err != nil {
		api.HandleError(resp, req, err)
		return
	}

	authLevel := data["authorizationLevel"]["authorization_level"]
	switch authLevel {
	case constants.AuthorizationLevelOfPublic, constants.AuthorizationLevelOfPrivate, constants.AuthorizationLevelOfInternal:
	default:
		api.HandleBadRequest(resp, req, fmt.Errorf("invalid authorization_level: %q", authLevel))
		return
	}

	caller := req.Attribute(constants.UserContextAttribute).(string)

	// Validate against the caller's effective entrances so user-added
	// entrances (stored in the overlay) are accepted too.
	entranceValid := false
	for _, e := range app.EffectiveEntrances(caller) {
		if e.Name == entranceName {
			entranceValid = true
			break
		}
	}
	if !entranceValid {
		api.HandleBadRequest(resp, req, fmt.Errorf("invalid entrance name: %q", entranceName))
		return
	}

	kclient := req.Attribute(constants.KubeSphereClientAttribute).(*clientset.ClientSet)

	var updated *v1alpha1.Application
	err = retry.RetryOnConflict(retry.DefaultRetry, func() error {
		current, err := kclient.AppClient.AppV1alpha1().Applications().Get(req.Request.Context(), app.Name, metav1.GetOptions{})
		if err != nil {
			return err
		}
		appCopy := current.DeepCopy()

		// User auth-level edits always live in the override slot — per-user
		// (Spec.UserSettings[caller]) for shared apps, or the app-global
		// Spec.Settings["authLevel"] for non-shared apps — so they survive the
		// reconciler reprojecting chart truth into Spec.Entrances. The base
		// Spec.Entrances[*].AuthLevel stays the install-time default;
		// EffectiveEntrances layers the override at read time.
		perEntranceLevel := make(map[string]string)
		if raw := appCopy.EffectiveSettings(caller)["authLevel"]; raw != "" {
			if err := json.Unmarshal([]byte(raw), &perEntranceLevel); err != nil {
				klog.Warningf("corrupt authLevel override for app %s, resetting: %v", appCopy.Name, err)
				return err
			}
		}

		// Baseline auth used to decide whether default_policy needs to flip
		// back from "public" to "system" on a public→private transition.
		// Prefer the entrance's current effective value; fall back to the
		// chart-default in Spec.Entrances when this is the first toggle.
		baselineAuth, ok := perEntranceLevel[entranceName]
		if !ok {
			for _, e := range appCopy.Spec.Entrances {
				if e.Name == entranceName {
					baselineAuth = e.AuthLevel
					break
				}
			}
		}

		perEntranceLevel[entranceName] = authLevel
		lvlBytes, err := json.Marshal(perEntranceLevel)
		if err != nil {
			return err
		}

		// Seed the policy blob from the caller's effective view so existing
		// per-entrance policies (chart base or prior overlay) are preserved
		// even before the one-time migration has run.
		policy := make(map[string]map[string]interface{})
		if raw := appCopy.EffectiveSettings(caller)["policy"]; raw != "" {
			if err := json.Unmarshal([]byte(raw), &policy); err != nil {
				klog.Warningf("corrupt policy for app %s, resetting: %v", appCopy.Name, err)
				policy = make(map[string]map[string]interface{})
			}
		}
		if _, ok := policy[entranceName]; !ok {
			policy[entranceName] = make(map[string]interface{})
		}
		switch {
		case authLevel == constants.AuthorizationLevelOfPublic:
			policy[entranceName]["default_policy"] = constants.AuthorizationLevelOfPublic
		case authLevel == constants.AuthorizationLevelOfPrivate &&
			baselineAuth == constants.AuthorizationLevelOfPublic:
			policy[entranceName]["default_policy"] = "system"
		}
		policyStr, err := json.Marshal(policy)
		if err != nil {
			return err
		}

		patchData := overridePatch(appCopy, caller, map[string]string{
			"authLevel": string(lvlBytes),
			"policy":    string(policyStr),
		})
		patchByte, err := json.Marshal(patchData)
		if err != nil {
			return err
		}
		updated, err = kclient.AppClient.AppV1alpha1().Applications().Patch(req.Request.Context(), appCopy.Name, types.MergePatchType, patchByte, metav1.PatchOptions{})
		return err
	})
	if err != nil {
		api.HandleError(resp, req, err)
		return
	}
	resp.WriteAsJson(updated.EffectiveSettings(caller))
}

func (h *Handler) setupAppEntrancePolicy(req *restful.Request, resp *restful.Response) {
	app, err := getAppByName(req, resp)
	if err != nil {
		klog.Errorf("Failed to get app name=%s err=%v", app.Spec.Name, err)
		// if error, response in function. Do nothing
		return
	}

	entranceName := req.PathParameter(ParamEntranceName)
	caller := req.Attribute(constants.UserContextAttribute).(string)

	bodyData, err := ioutil.ReadAll(req.Request.Body)
	if err != nil {
		api.HandleError(resp, req, err)
		return
	}

	var data map[string]interface{}
	err = json.Unmarshal(bodyData, &data)
	if err != nil {
		api.HandleError(resp, req, err)
		return
	}

	settings := data["policy"].(map[string]interface{})

	client := req.Attribute(constants.KubeSphereClientAttribute).(*clientset.ClientSet)

	var updated *v1alpha1.Application
	err = retry.RetryOnConflict(retry.DefaultRetry, func() error {
		current, err := client.AppClient.AppV1alpha1().Applications().Get(req.Request.Context(), app.Name, metav1.GetOptions{})
		if err != nil {
			return err
		}
		appCopy := current.DeepCopy()

		// Policy edits land in the override slot: per-user
		// (Spec.UserSettings[caller]) for shared apps, or the app-global
		// Spec.Settings["policy"] for non-shared apps. For non-shared apps the
		// reconciler no longer reprojects chart policy over it on redeploy, so
		// the edit persists.
		effSettings := appCopy.EffectiveSettings(caller)

		originBlob := effSettings["policy"]

		origin := make(map[string]interface{})
		if originBlob != "" {
			if err = json.Unmarshal([]byte(originBlob), &origin); err != nil {
				return err
			}
		}

		merge := make(map[string]interface{})
		merge[entranceName] = settings
		for k, v := range origin {
			if k != entranceName {
				merge[k] = v.(map[string]interface{})
				continue
			}
		}
		settingsBytes, err := json.Marshal(merge)
		if err != nil {
			return err
		}

		patchData := overridePatch(appCopy, caller, map[string]string{
			"policy": string(settingsBytes),
		})
		patchByte, err := json.Marshal(patchData)
		if err != nil {
			return err
		}
		updated, err = client.AppClient.AppV1alpha1().Applications().Patch(req.Request.Context(), appCopy.Name, types.MergePatchType, patchByte, metav1.PatchOptions{})
		return err
	})
	if err != nil {
		api.HandleError(resp, req, err)
		return
	}

	resp.WriteAsJson(updated.EffectiveSettings(caller))
}

func (h *Handler) tryToPatchDeploymentAnnotations(patchData map[string]interface{}, app *v1alpha1.Application) error {
	clientset, err := kubernetes.NewForConfig(h.kubeConfig)
	if err != nil {
		return err
	}
	patchByte, err := json.Marshal(patchData)
	if err != nil {
		return err
	}
	deployment, err := clientset.AppsV1().Deployments(app.Spec.Namespace).
		Get(context.TODO(), app.Spec.DeploymentName, metav1.GetOptions{})
	if err != nil {
		if kerrors.IsNotFound(err) {
			return h.tryToPatchStatefulSetAnnotations(patchData, app)
		}
		return err
	}
	a, err := clientset.AppsV1().Deployments(app.Spec.Namespace).
		Patch(context.TODO(), deployment.Name,
			types.MergePatchType,
			patchByte,
			metav1.PatchOptions{})
	klog.Infof("update annotations: %v", a.Annotations)
	return err
}

func (h *Handler) tryToPatchStatefulSetAnnotations(patchData map[string]interface{}, app *v1alpha1.Application) error {
	clientset, err := kubernetes.NewForConfig(h.kubeConfig)
	if err != nil {
		return err
	}
	patchByte, err := json.Marshal(patchData)
	if err != nil {
		return err
	}
	statefulSet, err := clientset.AppsV1().StatefulSets(app.Spec.Namespace).
		Get(context.TODO(), app.Spec.DeploymentName, metav1.GetOptions{})
	if err != nil {
		if kerrors.IsNotFound(err) {
			return nil
		}
		return err
	}
	_, err = clientset.AppsV1().StatefulSets(app.Spec.Namespace).
		Patch(context.TODO(), statefulSet.Name,
			types.MergePatchType,
			patchByte,
			metav1.PatchOptions{})

	return err
}

type permission struct {
	DataType string   `json:"dataType"`
	Group    string   `json:"group"`
	Version  string   `json:"version"`
	Ops      []string `json:"ops"`
}

type applicationPermission struct {
	App         string       `json:"app"`
	Owner       string       `json:"owner"`
	Permissions []permission `json:"permissions"`
}

// Deprecated
func (h *Handler) applicationPermissionList(req *restful.Request, resp *restful.Response) {
	owner := req.Attribute(constants.UserContextAttribute).(string)
	//token := req.HeaderParameter(constants.AuthorizationTokenKey)
	client, err := dynamic.NewForConfig(h.kubeConfig)
	if err != nil {
		api.HandleError(resp, req, err)
		return
	}
	ret := make([]applicationPermission, 0)
	apClient := provider.NewApplicationPermissionRequest(client)
	aps, err := apClient.List(req.Request.Context(), metav1.NamespaceAll, metav1.ListOptions{})
	if err != nil {
		api.HandleError(resp, req, err)
		return
	}
	for _, ap := range aps.Items {
		if ap.Object == nil {
			continue
		}
		app, _, _ := unstructured.NestedString(ap.Object, "spec", "app")
		perms, _, _ := unstructured.NestedSlice(ap.Object, "spec", "permissions")
		klog.Infof("perms Type: %T, perms: %v", perms, perms)
		permissions := make([]permission, 0)
		for _, p := range perms {
			if perm, ok := p.(map[string]interface{}); ok {
				ops := make([]string, 0)
				for _, op := range perm["ops"].([]interface{}) {
					if opStr, ok := op.(string); ok {
						ops = append(ops, opStr)
					}
				}
				permissions = append(permissions, permission{
					DataType: perm["dataType"].(string),
					Group:    perm["group"].(string),
					Version:  perm["version"].(string),
					Ops:      ops,
				})
			}

		}
		ret = append(ret, applicationPermission{
			App:         app,
			Owner:       owner,
			Permissions: permissions,
		})
	}
	resp.WriteAsJson(ret)
}

func (h *Handler) getApplicationPermission(req *restful.Request, resp *restful.Response) {
	app := req.PathParameter(ParamAppName)
	owner := req.Attribute(constants.UserContextAttribute).(string)
	// Use ResolveAppMgrName so shared apps installed under a different admin
	// are still found by their canonical {app}-shared-{app} name.
	name, _, err := apputils.ResolveAppMgrName(req.Request.Context(), app, owner)
	if err != nil {
		api.HandleError(resp, req, err)
		return
	}
	var am v1alpha1.ApplicationManager
	err = h.ctrlClient.Get(req.Request.Context(), types.NamespacedName{Name: name}, &am)
	if err != nil {
		api.HandleError(resp, req, err)
		return
	}

	// sys app does not have app config
	if am.Spec.Config == "" {
		ret := &applicationPermission{
			App:         am.Spec.AppName,
			Owner:       owner,
			Permissions: []permission{},
		}

		resp.WriteAsJson(ret)
		return
	}

	var appConfig appcfg.ApplicationConfig
	err = appcfg.GetAppConfig(&am, &appConfig)
	if err != nil {
		klog.Errorf("Failed to get app config err=%v", err)
		api.HandleError(resp, req, err)
		return
	}

	var ret *applicationPermission
	permissions := appinstaller.ParseAppPermission(appConfig.Permission)
	for _, ap := range permissions {
		if perms, ok := ap.([]appcfg.ProviderPermission); ok {
			permissions := make([]permission, 0)
			for _, p := range perms {
				permissions = append(permissions, permission{
					DataType: p.ProviderName,
					Group:    p.AppName,
				})
			}
			ret = &applicationPermission{
				App:         am.Spec.AppName,
				Owner:       owner,
				Permissions: permissions,
			}
			break
		}
	}
	if ret == nil {
		api.HandleNotFound(resp, req, errors.New("application permission not found"))
		return
	}
	resp.WriteAsJson(ret)
}

type providerRegistry struct {
	DataType    string  `json:"dataType"`
	Deployment  string  `json:"deployment"`
	Description string  `json:"description"`
	Endpoint    string  `json:"endpoint"`
	Group       string  `json:"group"`
	Kind        string  `json:"kind"`
	Namespace   string  `json:"namespace"`
	OpApis      []opApi `json:"opApis"`
	Version     string  `json:"version"`
}

type opApi struct {
	Name string `json:"name"`
	URI  string `json:"uri"`
}

// Deprecated
func (h *Handler) getProviderRegistry(req *restful.Request, resp *restful.Response) {
	dataTypeReq := req.PathParameter(ParamDataType)
	groupReq := req.PathParameter(ParamGroup)
	versionReq := req.PathParameter(ParamVersion)
	owner := req.Attribute(constants.UserContextAttribute).(string)
	client, err := dynamic.NewForConfig(h.kubeConfig)
	if err != nil {
		api.HandleError(resp, req, err)
		return
	}
	var ret *providerRegistry
	rClient := provider.NewRegistryRequest(client)
	namespace := fmt.Sprintf("user-system-%s", owner)
	prs, err := rClient.List(req.Request.Context(), namespace, metav1.ListOptions{})
	if err != nil {
		api.HandleError(resp, req, err)
		return
	}
	for _, ap := range prs.Items {
		if ap.Object == nil {
			continue
		}
		dataType, _, _ := unstructured.NestedString(ap.Object, "spec", "dataType")
		group, _, _ := unstructured.NestedString(ap.Object, "spec", "group")
		version, _, _ := unstructured.NestedString(ap.Object, "spec", "version")
		kind, _, _ := unstructured.NestedString(ap.Object, "spec", "kind")

		if dataType == dataTypeReq && group == groupReq && version == versionReq && kind == "provider" {
			deployment, _, _ := unstructured.NestedString(ap.Object, "spec", "deployment")
			description, _, _ := unstructured.NestedString(ap.Object, "spec", "description")
			endpoint, _, _ := unstructured.NestedString(ap.Object, "spec", "endpoint")
			ns, _, _ := unstructured.NestedString(ap.Object, "spec", "namespace")
			opApis := make([]opApi, 0)
			opApiList, _, _ := unstructured.NestedSlice(ap.Object, "spec", "opApis")
			for _, op := range opApiList {
				if aop, ok := op.(map[string]interface{}); ok {
					opApis = append(opApis, opApi{
						Name: aop["name"].(string),
						URI:  aop["uri"].(string),
					})
				}
			}
			ret = &providerRegistry{
				DataType:    dataType,
				Deployment:  deployment,
				Description: description,
				Endpoint:    endpoint,
				Kind:        kind,
				Group:       group,
				Namespace:   ns,
				OpApis:      opApis,
				Version:     version,
			}
			break
		}
	}
	if ret == nil {
		api.HandleNotFound(resp, req, errors.New("provider registry not found"))
		return
	}
	resp.WriteAsJson(ret)
}

func (h *Handler) getApplicationProviderList(req *restful.Request, resp *restful.Response) {
	owner := req.Attribute(constants.UserContextAttribute).(string)
	app := req.PathParameter(ParamAppName)

	// Use ResolveAppMgrName so shared apps installed under a different admin
	// are still found by their canonical {app}-shared-{app} name.
	name, _, err := apputils.ResolveAppMgrName(req.Request.Context(), app, owner)
	if err != nil {
		api.HandleError(resp, req, err)
		return
	}
	var am v1alpha1.ApplicationManager
	err = h.ctrlClient.Get(req.Request.Context(), types.NamespacedName{Name: name}, &am)
	if err != nil {
		api.HandleError(resp, req, err)
		return
	}

	// sys app does not have app config
	if am.Spec.Config == "" {
		resp.WriteAsJson([]providerRegistry{})
		return
	}

	var appConfig appcfg.ApplicationConfig
	err = appcfg.GetAppConfig(&am, &appConfig)
	if err != nil {
		klog.Errorf("Failed to get app config err=%v", err)
		api.HandleError(resp, req, err)
		return
	}

	ret := make([]providerRegistry, 0)
	ns := am.Spec.AppNamespace
	for _, ap := range appConfig.Provider {
		dataType := ap.Name
		endpoint := ap.Entrance
		opApis := make([]opApi, 0)
		for _, op := range ap.Paths {
			opApis = append(opApis, opApi{
				URI: op,
			})
		}
		ret = append(ret, providerRegistry{
			DataType:  dataType,
			Endpoint:  endpoint,
			Namespace: ns,
			OpApis:    opApis,
		})
	}
	resp.WriteAsJson(ret)
}

func (h *Handler) getApplicationSubject(req *restful.Request, resp *restful.Response) {
	app := req.PathParameter(ParamAppName)
	owner := req.Attribute(constants.UserContextAttribute).(string)
	client, err := dynamic.NewForConfig(h.kubeConfig)
	if err != nil {
		api.HandleError(resp, req, err)
		return
	}
	dc, err := tapr.NewMiddlewareRequest(client)
	if err != nil {
		api.HandleError(resp, req, err)
		return
	}
	namespace := fmt.Sprintf("user-system-%s", owner)
	mrs, err := dc.List(req.Request.Context(), namespace, metav1.ListOptions{})
	if err != nil {
		api.HandleError(resp, req, err)
		return
	}
	ret := make([]tapr.NatsConfig, 0)
	klog.Infof("get Application Subject...............")
	klog.Infof("mrs.Items:len: %v", len(mrs.Items))
	if len(mrs.Items) > 0 {
		for _, mr := range mrs.Items {
			if mr.Object == nil {
				continue
			}
			middlewareType, _, _ := unstructured.NestedString(mr.Object, "spec", "middleware")
			klog.Infof("middlewareType: %v", middlewareType)
			if middlewareType != "nats" {
				continue
			}
			appName, _, _ := unstructured.NestedString(mr.Object, "spec", "app")
			if appName != app {
				continue
			}
			username, _, _ := unstructured.NestedString(mr.Object, "spec", "nats", "user")

			klog.Infof("appName: %v", appName)
			natsCfg := tapr.NatsConfig{}
			natsCfg.Username = username
			nats, _, _ := unstructured.NestedMap(mr.Object, "spec", "nats")
			subjects, _, _ := unstructured.NestedSlice(nats, "subjects")
			klog.Infof("subjects: %v", subjects)
			natsCfg.Subjects = make([]tapr.Subject, 0)
			for _, s := range subjects {
				subject := tapr.Subject{}
				subjectMap := s.(map[string]interface{})
				subject.Name, _, _ = unstructured.NestedString(subjectMap, "name")

				permission, _, _ := unstructured.NestedMap(subjectMap, "permission")
				subject.Permission = tapr.Permission{
					Pub: permission["pub"].(string),
					Sub: permission["sub"].(string),
				}
				subject.Export = make([]tapr.Permission, 0)
				export, found, _ := unstructured.NestedSlice(subjectMap, "export")
				if found {
					for _, e := range export {
						exportMap := e.(map[string]interface{})
						subject.Export = append(subject.Export,
							tapr.Permission{
								AppName: exportMap["appName"].(string),
								Pub:     exportMap["pub"].(string),
								Sub:     exportMap["sub"].(string),
							},
						)
					}
				}
				natsCfg.Subjects = append(natsCfg.Subjects, subject)
			}
			natsCfg.Refs = make([]tapr.Ref, 0)
			refs, _, _ := unstructured.NestedSlice(nats, "refs")
			for _, r := range refs {
				ref := tapr.Ref{}
				refMap := r.(map[string]interface{})
				ref.AppName, _, _ = unstructured.NestedString(refMap, "appName")
				ref.AppNamespace, _, _ = unstructured.NestedString(refMap, "appNamespace")

				refSubjects, _, _ := unstructured.NestedSlice(refMap, "subjects")
				for _, rs := range refSubjects {
					refSubject := tapr.RefSubject{}
					rsMap := rs.(map[string]interface{})
					refSubject.Name, _, _ = unstructured.NestedString(rsMap, "name")
					refSubject.Pub, _, _ = unstructured.NestedString(rsMap, "pub")
					refSubject.Sub, _, _ = unstructured.NestedString(rsMap, "sub")

					ref.Subjects = append(ref.Subjects, refSubject)
				}

				natsCfg.Refs = append(natsCfg.Refs, ref)
			}
			ret = append(ret, natsCfg)
		}
	}
	resp.WriteAsJson(ret)
}
