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
	isShared := appcfg.IsShared(appCopy)

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
		// App-level keys are global to the Application and always land
		// in Spec.Settings; on shared apps any other key overlays into
		// Spec.UserSettings[caller] so per-user views stay isolated.
		// Per-user apps (v1 and v3+per-user) have no UserSettings
		// concept and keep the legacy behavior of writing everything to
		// Spec.Settings.
		if isShared && !isAppLevelSettingKey(k) {
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
	entranceName := req.PathParameter(ParamEntranceName)
	validName := false
	for _, e := range app.Spec.Entrances {
		if e.Name == entranceName {
			validName = true
		}
	}
	if !validName {
		api.HandleBadRequest(resp, req, errors.New("invalid entrance name"))
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
	isShared := appcfg.IsShared(app)

	kclient := req.Attribute(constants.KubeSphereClientAttribute).(*clientset.ClientSet)

	customDomain, ok := settings["customDomain"].(map[string]interface{})

	existing := appCopy.EffectiveSettings(caller)["customDomain"]

	merge := make(map[string]interface{})

	keys := []string{"third_level_domain", "third_party_domain"}

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

	var patchData map[string]interface{}
	if isShared {
		// Patch only Spec.UserSettings[caller]["customDomain"]. JSON
		// merge-patch leaves other users' entries and other Settings
		// keys untouched.
		patchData = map[string]interface{}{
			"spec": map[string]interface{}{
				"userSettings": map[string]interface{}{
					caller: map[string]string{
						"customDomain": string(settingsBytes),
					},
				},
			},
		}
	} else {
		patchData = map[string]interface{}{
			"spec": map[string]interface{}{
				"settings": map[string]string{
					"customDomain": string(settingsBytes),
				},
			},
		}
	}
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

	entranceValid := false
	for _, e := range app.Spec.Entrances {
		if e.Name == entranceName {
			entranceValid = true
			break
		}
	}
	if !entranceValid {
		api.HandleBadRequest(resp, req, fmt.Errorf("invalid entrance name: %q", entranceName))
		return
	}

	caller := req.Attribute(constants.UserContextAttribute).(string)
	kclient := req.Attribute(constants.KubeSphereClientAttribute).(*clientset.ClientSet)

	var updated *v1alpha1.Application
	err = retry.RetryOnConflict(retry.DefaultRetry, func() error {
		current, err := kclient.AppClient.AppV1alpha1().Applications().Get(req.Request.Context(), app.Name, metav1.GetOptions{})
		if err != nil {
			return err
		}
		appCopy := current.DeepCopy()

		if appcfg.IsShared(appCopy) {
			// Shared apps: only Spec.UserSettings[caller]["authLevel"] +
			// ["policy"] for the affected entrance are touched. The
			// global Spec.Entrances[*].AuthLevel stays as the install-
			// time default for everybody else.
			//
			// Uses a JSON merge-patch on spec.userSettings.<caller> so
			// concurrent edits by *other* users (different keys in the
			// same map) don't fight over resourceVersion.
			perEntranceLevel := make(map[string]string)
			if u := appCopy.Spec.UserSettings[caller]; u != nil {
				if raw := u["authLevel"]; raw != "" {
					if err := json.Unmarshal([]byte(raw), &perEntranceLevel); err != nil {
						klog.Warningf("corrupt UserSettings[%s][authLevel] for app %s, resetting: %v", caller, appCopy.Name, err)
						return err
					}
				}
			}

			// Baseline auth used to decide whether default_policy needs
			// to flip back from "public" to "system" on a public→private
			// transition. Prefer the user's previously stored value
			// (their current effective auth); fall back to the global
			// install-time default when this is the first toggle.
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

			policy := make(map[string]map[string]interface{})
			if u := appCopy.Spec.UserSettings[caller]; u != nil {
				if raw := u["policy"]; raw != "" {
					if err := json.Unmarshal([]byte(raw), &policy); err != nil {
						klog.Warningf("corrupt UserSettings[%s][policy] for app %s, resetting: %v", caller, appCopy.Name, err)
						policy = make(map[string]map[string]interface{})
					}
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

			patchData := map[string]interface{}{
				"spec": map[string]interface{}{
					"userSettings": map[string]interface{}{
						caller: map[string]string{
							"authLevel": string(lvlBytes),
							"policy":    string(policyStr),
						},
					},
				},
			}
			patchByte, err := json.Marshal(patchData)
			if err != nil {
				return err
			}
			updated, err = kclient.AppClient.AppV1alpha1().Applications().Patch(req.Request.Context(), appCopy.Name, types.MergePatchType, patchByte, metav1.PatchOptions{})
			return err
		}

		// v1/v2: write Spec.Entrances[i].AuthLevel + Spec.Settings["policy"].
		policy := make(map[string]map[string]interface{})
		if p := appCopy.Spec.Settings["policy"]; p != "" {
			if err := json.Unmarshal([]byte(p), &policy); err != nil {
				return err
			}
		}
		for i := range appCopy.Spec.Entrances {
			e := &appCopy.Spec.Entrances[i]
			if e.Name != entranceName {
				continue
			}
			// Ensure the per-entrance policy entry exists before
			// writing into it — a freshly installed app or an entrance
			// added later may not have one yet, and assigning into a
			// nil sub-map would panic.
			if _, ok := policy[e.Name]; !ok {
				policy[e.Name] = make(map[string]interface{})
			}
			switch {
			case authLevel == constants.AuthorizationLevelOfPublic:
				policy[e.Name]["default_policy"] = constants.AuthorizationLevelOfPublic
			case authLevel == constants.AuthorizationLevelOfPrivate &&
				e.AuthLevel == constants.AuthorizationLevelOfPublic:
				policy[e.Name]["default_policy"] = "system"
			}
			e.AuthLevel = authLevel
			break
		}
		policyStr, err := json.Marshal(policy)
		if err != nil {
			return err
		}
		if appCopy.Spec.Settings == nil {
			appCopy.Spec.Settings = make(map[string]string)
		}
		appCopy.Spec.Settings["policy"] = string(policyStr)
		updated, err = kclient.AppClient.AppV1alpha1().Applications().Update(req.Request.Context(), appCopy, metav1.UpdateOptions{})
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
		isShared := appcfg.IsShared(appCopy)

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

		var patchData map[string]interface{}
		if isShared {
			patchData = map[string]interface{}{
				"spec": map[string]interface{}{
					"userSettings": map[string]interface{}{
						caller: map[string]string{
							"policy": string(settingsBytes),
						},
					},
				},
			}
		} else {
			patchData = map[string]interface{}{
				"spec": map[string]interface{}{
					"settings": map[string]string{
						"policy": string(settingsBytes),
					},
				},
			}
		}
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
