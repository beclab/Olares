package apiserver

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"sync"
	"time"

	sysv1alpha1 "github.com/beclab/Olares/framework/app-service/api/sys.bytetrade.io/v1alpha1"
	"github.com/beclab/Olares/framework/app-service/pkg/apiserver/api"
	"github.com/beclab/Olares/framework/app-service/pkg/constants"
	"github.com/beclab/Olares/framework/app-service/pkg/utils"
	apputils "github.com/beclab/Olares/framework/app-service/pkg/utils/app"
	"github.com/emicklei/go-restful/v3"
	"k8s.io/apimachinery/pkg/types"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

type remoteOptionsProxyRequest struct {
	Endpoint string `json:"endpoint"`
}

func (h *Handler) getAppEnv(req *restful.Request, resp *restful.Response) {
	appName := req.PathParameter(ParamAppName)
	owner := getCurrentUser(req)

	if appName == "" || owner == "" {
		api.HandleBadRequest(resp, req, fmt.Errorf("app name and owner are required"))
		return
	}

	appNamespace, err := utils.AppNamespace(appName, owner, "")
	if err != nil {
		api.HandleInternalError(resp, req, fmt.Errorf("failed to get app namespace: %v", err))
		return
	}

	envs := make([]sysv1alpha1.AppEnvVar, 0)
	var appEnv sysv1alpha1.AppEnv
	if err := client.IgnoreNotFound(h.ctrlClient.Get(req.Request.Context(), types.NamespacedName{Namespace: appNamespace, Name: apputils.FormatAppEnvName(appName, owner)}, &appEnv)); err != nil {
		api.HandleInternalError(resp, req, err)
		return
	}
	if len(appEnv.Envs) > 0 {
		envs = appEnv.Envs
	}

	resp.WriteAsJson(envs)
}

func (h *Handler) updateAppEnv(req *restful.Request, resp *restful.Response) {
	appName := req.PathParameter(ParamAppName)
	owner := getCurrentUser(req)

	if appName == "" || owner == "" {
		api.HandleBadRequest(resp, req, fmt.Errorf("app name and owner are required"))
		return
	}

	var updatedEnvs []sysv1alpha1.AppEnvVar
	if err := req.ReadEntity(&updatedEnvs); err != nil {
		api.HandleBadRequest(resp, req, fmt.Errorf("failed to parse request body: %v", err))
		return
	}

	appNamespace, err := utils.AppNamespace(appName, owner, "")
	if err != nil {
		api.HandleInternalError(resp, req, fmt.Errorf("failed to get app namespace: %v", err))
		return
	}

	var targetAppEnv sysv1alpha1.AppEnv
	if err := h.ctrlClient.Get(req.Request.Context(), types.NamespacedName{Namespace: appNamespace, Name: apputils.FormatAppEnvName(appName, owner)}, &targetAppEnv); err != nil {
		api.HandleInternalError(resp, req, err)
		return
	}

	var refEnvOnce sync.Once
	var listErr error
	refEnvs := make(map[string]string)

	updated := false
	original := targetAppEnv.DeepCopy()
	for i, existingEnv := range targetAppEnv.Envs {
		for _, env := range updatedEnvs {
			if existingEnv.EnvName != env.EnvName {
				continue
			}
			if !existingEnv.Editable {
				api.HandleBadRequest(resp, req, fmt.Errorf("app env '%s' is not editable", env.EnvName))
				return
			}
			if existingEnv.Required && existingEnv.Default == "" && env.Value == "" && (env.ValueFrom == nil || env.ValueFrom.EnvName == "") {
				api.HandleBadRequest(resp, req, fmt.Errorf("app env '%s' is required", env.EnvName))
				return
			}
			if env.ValueFrom != nil && env.ValueFrom.EnvName != "" && (existingEnv.ValueFrom == nil || existingEnv.ValueFrom.EnvName != env.ValueFrom.EnvName) {
				refEnvOnce.Do(func() {
					sysenvs := new(sysv1alpha1.SystemEnvList)
					listErr = h.ctrlClient.List(req.Request.Context(), sysenvs)
					if listErr != nil {
						return
					}
					userenvs := new(sysv1alpha1.UserEnvList)
					listErr = h.ctrlClient.List(req.Request.Context(), userenvs, client.InNamespace(utils.UserspaceName(owner)))
					for _, sysenv := range sysenvs.Items {
						refEnvs[sysenv.EnvName] = sysenv.GetEffectiveValue()
					}
					for _, userenv := range userenvs.Items {
						refEnvs[userenv.EnvName] = userenv.GetEffectiveValue()
					}
				})
				if listErr != nil {
					api.HandleInternalError(resp, req, fmt.Errorf("failed to list referenced envs: %s", listErr))
					return
				}
				value, ok := refEnvs[env.ValueFrom.EnvName]
				if !ok {
					api.HandleBadRequest(resp, req, fmt.Errorf("app env '%s' references unknown env '%s'", env.EnvName, env.ValueFrom.EnvName))
					return
				}
				if existingEnv.Required && value == "" {
					api.HandleBadRequest(resp, req, fmt.Errorf("required app env '%s' references empty env '%s'", env.EnvName, env.ValueFrom.EnvName))
					return
				}
				if existingEnv.ValidateValue(value) != nil {
					api.HandleBadRequest(resp, req, fmt.Errorf("app env '%s' references invalid value '%s' from '%s': %v", env.EnvName, value, env.ValueFrom.EnvName, err))
					return
				}
				targetAppEnv.Envs[i].ValueFrom = env.ValueFrom
				targetAppEnv.Envs[i].Value = value
				targetAppEnv.Envs[i].ValueFrom.Status = constants.EnvRefStatusSynced
				updated = true
			} else if existingEnv.Value != env.Value {
				if err := existingEnv.ValidateValue(env.Value); err != nil {
					api.HandleBadRequest(resp, req, fmt.Errorf("failed to update app env '%s': %v", env.EnvName, err))
					return
				}
				targetAppEnv.Envs[i].Value = env.Value
				updated = true
			}
			if updated && existingEnv.ApplyOnChange {
				targetAppEnv.NeedApply = true
			}
			break
		}
	}

	if updated {
		if err := h.ctrlClient.Patch(req.Request.Context(), &targetAppEnv, client.MergeFrom(original)); err != nil {
			api.HandleInternalError(resp, req, err)
			return
		}
	}

	resp.WriteAsJson(targetAppEnv.Envs)
}

func (h *Handler) proxyRemoteOptions(req *restful.Request, resp *restful.Response) {
	var body remoteOptionsProxyRequest
	if err := req.ReadEntity(&body); err != nil {
		api.HandleBadRequest(resp, req, err)
		return
	}
	if body.Endpoint == "" {
		api.HandleBadRequest(resp, req, fmt.Errorf("endpoint is required"))
		return
	}
	u, err := url.Parse(body.Endpoint)
	if err != nil {
		api.HandleBadRequest(resp, req, fmt.Errorf("invalid endpoint: %w", err))
		return
	}
	if u.Scheme != "http" && u.Scheme != "https" {
		api.HandleBadRequest(resp, req, fmt.Errorf("unsupported scheme: %s", u.Scheme))
		return
	}

	client := &http.Client{Timeout: 10 * time.Second}
	httpReq, err := http.NewRequestWithContext(req.Request.Context(), http.MethodGet, body.Endpoint, nil)
	if err != nil {
		api.HandleError(resp, req, err)
		return
	}

	r, err := client.Do(httpReq)
	if err != nil {
		api.HandleError(resp, req, err)
		return
	}
	defer r.Body.Close()
	if r.StatusCode < 200 || r.StatusCode >= 300 {
		api.HandleBadRequest(resp, req, fmt.Errorf("unexpected status code: %d", r.StatusCode))
		return
	}
	data, err := io.ReadAll(r.Body)
	if err != nil {
		api.HandleError(resp, req, err)
		return
	}

	var items []sysv1alpha1.EnvValueOptionItem
	if err := json.Unmarshal(data, &items); err != nil {
		api.HandleBadRequest(resp, req, fmt.Errorf("invalid RemoteOptions body: %w", err))
		return
	}
	resp.WriteAsJson(items)
}
