package app_service

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"io/ioutil"
	"net/http"
	"net/url"
	"os"

	"bytetrade.io/web3os/bfl/pkg/constants"
	appv1 "github.com/beclab/api/api/app.bytetrade.io/v1alpha1"
	ctrl "sigs.k8s.io/controller-runtime"

	"k8s.io/klog/v2"
)

func (c *Client) fetchAppInfoFromAppService(appname, token string) (map[string]interface{}, error) {
	appServiceHost := os.Getenv(AppServiceHostEnv)
	appServicePort := os.Getenv(AppServicePortEnv)
	urlStr := fmt.Sprintf(AppServiceGetURLTempl, appServiceHost, appServicePort, appname)

	return c.doHttpGetOne(urlStr, token)
}

func (c *Client) getAppInfoFromData(data map[string]interface{}) (*AppInfo, error) {
	appSpec, ok := data["spec"].(map[string]interface{})
	if !ok {
		klog.Error("get app info error: ", data)
		return nil, errors.New("app info is invalid")
	}

	return &AppInfo{
		ID:             genAppID(appSpec),
		Name:           appSpec["name"].(string),
		Namespace:      appSpec["namespace"].(string),
		DeploymentName: appSpec["deployment"].(string),
		Owner:          appSpec["owner"].(string)}, nil
}

func (c *Client) getAppListFromData(apps []appv1.Application) ([]*AppInfo, error) {

	var res []*AppInfo
	for _, app := range apps {
		var appEntrances []Entrance
		var appSharedEntrances []Entrance
		appPorts := make([]ServicePort, 0)
		appACLs := make([]ACL, 0)

		appSpec := app.Spec

		isSysApp := appSpec.IsSysApp

		// get app settings to filter system service not to list
		var title, target, state, requiredGPU, defaultThirdLevelDomainConfig string
		isClusterScoped, mobileSupported := false, false
		settingsMap := appSpec.Settings
		if settingsMap != nil {
			if _, ok := settingsMap["system_service"]; ok {
				// It is the system service, not app
				continue
			}

			if t, ok := settingsMap["title"]; ok {
				title = t
			}
			if t, ok := settingsMap["clusterScoped"]; ok && t == "true" {
				isClusterScoped = true
			}
			if t, ok := settingsMap["defaultThirdLevelDomainConfig"]; ok {
				defaultThirdLevelDomainConfig = t
			}

			if t, ok := settingsMap["target"]; ok {
				target = t
			}
			if t, ok := settingsMap["mobileSupported"]; ok && t == "true" {
				mobileSupported = true
			}
			if t, ok := settingsMap["requiredGPU"]; ok {
				requiredGPU = t
			}
		}

		entranceStatusesMap := make(map[string]appv1.EntranceStatus)
		state = app.Status.State
		for _, es := range app.Status.EntranceStatuses {
			entranceStatusesMap[es.Name] = es
		}
		klog.Infof("entranceStatusesMap: %v", entranceStatusesMap)

		for _, entrance := range appSpec.Entrances {
			var appEntrance Entrance
			appEntrance.Name = entrance.Name
			appEntrance.Title = entrance.Title
			appEntrance.Icon = entrance.Icon
			if entrance.Invisible {
				appEntrance.Invisible = true
			}
			appEntrance.AuthLevel = entrance.AuthLevel
			if entrance.OpenMethod != "" {
				appEntrance.OpenMethod = entrance.OpenMethod
			} else {
				appEntrance.OpenMethod = "default"
			}
			if t, ok := entranceStatusesMap[appEntrance.Name]; ok {
				appEntrance.State = t.State.String()
				appEntrance.Reason = t.Reason
				appEntrance.Message = t.Message
			} else {
				appEntrance.State = state
			}

			appEntrances = append(appEntrances, appEntrance)
		}

		for _, entrance := range appSpec.SharedEntrances {
			var appEntrance Entrance
			appEntrance.Name = entrance.Name
			appEntrance.URL = entrance.URL
			appEntrance.Title = entrance.Title
			appEntrance.Icon = entrance.Icon
			if entrance.Invisible {
				appEntrance.Invisible = true
			}
			appEntrance.AuthLevel = entrance.AuthLevel
			// appEntrance.State = state

			appSharedEntrances = append(appSharedEntrances, appEntrance)
		}

		for _, p := range appSpec.Ports {
			var appPort ServicePort
			appPort.ExposePort = p.ExposePort
			appPort.Host = p.Host
			appPort.Name = p.Name
			appPort.Port = p.Port
			appPort.Protocol = p.Protocol
			appPorts = append(appPorts, appPort)
		}

		for _, a := range appSpec.TailScaleACLs {
			var tailscaleACL ACL
			tailscaleACL.Action = a.Action
			if a.Src != nil {
				src := make([]string, 0)
				src = append(src, a.Src...)
				tailscaleACL.Src = src
			}
			tailscaleACL.Proto = a.Proto
			if a.Dst != nil {
				dst := make([]string, 0)
				dst = append(dst, a.Dst...)
				tailscaleACL.Dst = dst
			}
			appACLs = append(appACLs, tailscaleACL)
		}

		res = append(res, &AppInfo{
			ID:                            appSpec.Appid,
			Name:                          appSpec.Name,
			RawAppName:                    appSpec.RawAppName,
			Namespace:                     appSpec.Namespace,
			DeploymentName:                appSpec.DeploymentName,
			Owner:                         appSpec.Owner,
			Icon:                          appSpec.Icon,
			Title:                         title,
			Target:                        target,
			Entrances:                     appEntrances,
			Ports:                         appPorts,
			TailScaleACLs:                 appACLs,
			State:                         state,
			IsSysApp:                      isSysApp,
			IsClusterScoped:               isClusterScoped,
			MobileSupported:               mobileSupported,
			RequiredGpu:                   requiredGPU,
			DefaultThirdLevelDomainConfig: defaultThirdLevelDomainConfig,
			SharedEntrances:               appSharedEntrances,
			IsShared:                      app.Labels["app.bytetrade.io/app-shared"] == "true",
			ClonedFrom:                    app.Labels["app.bytetrade.io/app-cloned-from"],
		})

	}

	return res, nil

}

func (c *Client) addTokenHeader(req *http.Request, token string) (*http.Request, error) {
	if req.Header == nil {
		req.Header = make(http.Header)
	}
	if len(token) > 0 {
		req.Header.Add(constants.UserAuthorizationTokenKey, token)
	} else {
		config, err := ctrl.GetConfig()
		if err != nil {
			klog.Error("get kube config error: ", err)
			return nil, err
		}

		req.Header.Add("Authorization", "Bearer "+config.BearerToken)
	}

	return req, nil
}

func (c *Client) doHttpGetResponse(urlStr, token string) (*http.Response, error) {
	url, err := url.Parse(urlStr)
	if err != nil {
		return nil, err
	}

	req := &http.Request{
		Method: http.MethodGet,
		URL:    url,
	}

	req, err = c.addTokenHeader(req, token)
	if err != nil {
		return nil, err
	}

	resp, err := c.httpClient.Do(req)
	if err != nil {
		klog.Error("do request error: ", err)
		return nil, err
	}

	if resp.StatusCode != http.StatusOK {
		klog.Error("response not ok, ", resp.Status)
		data, err := ioutil.ReadAll(resp.Body)
		if err != nil {
			return nil, err
		}
		resp.Body.Close()
		return nil, fmt.Errorf("response error, code %d, msg: %s", resp.StatusCode, string(data))
	}

	return resp, nil
}

func (c *Client) readHttpResponse(resp *http.Response) (map[string]interface{}, error) {
	data, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	app := make(map[string]interface{}) // simple get. TODO: application struct
	err = json.Unmarshal(data, &app)
	if err != nil {
		klog.Error("parse response error: ", err, string(data))
		return nil, err
	}

	return app, nil

}

func (c *Client) readHttpResponseList(resp *http.Response) ([]map[string]interface{}, error) {
	data, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	var apps []map[string]interface{} // simple get. TODO: application struct
	err = json.Unmarshal(data, &apps)
	if err != nil {
		klog.Error("parse response error: ", err, string(data))
		return nil, err
	}

	return apps, nil
}

func (c *Client) doHttpGetOne(urlStr, token string) (map[string]interface{}, error) {
	resp, err := c.doHttpGetResponse(urlStr, token)
	if err != nil {
		return nil, err
	}

	return c.readHttpResponse(resp)
}

func (c *Client) doHttpGetList(urlStr, token string) ([]map[string]interface{}, error) {
	resp, err := c.doHttpGetResponse(urlStr, token)
	if err != nil {
		return nil, err
	}

	return c.readHttpResponseList(resp)
}

func (c *Client) doHttpGetApplicationList(urlStr, token string) ([]appv1.Application, error) {
	resp, err := c.doHttpGetResponse(urlStr, token)
	if err != nil {
		return nil, err
	}

	data, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	var apps []appv1.Application
	if err := json.Unmarshal(data, &apps); err != nil {
		klog.Error("parse response error: ", err, string(data))
		return nil, err
	}

	return apps, nil
}

func (c *Client) doHttpPost(urlStr, token string, bodydata interface{}) (map[string]interface{}, error) {
	var data io.Reader
	if bodydata != nil {
		jsonData, err := json.Marshal(bodydata)
		if err != nil {
			return nil, errors.New("body data parse error")
		}

		data = bytes.NewBuffer(jsonData)
	}

	req, err := http.NewRequest(http.MethodPost, urlStr, data)
	if err != nil {
		return nil, err
	}
	req, err = c.addTokenHeader(req, token)
	if err != nil {
		return nil, err
	}
	req.Header.Add("content-type", "application/json")

	resp, err := c.httpClient.Do(req)
	if err != nil {
		klog.Error("do request error: ", err)
		return nil, err
	}

	if resp.StatusCode != http.StatusOK {
		klog.Error("response not ok, ", resp.Status)
		data, err := ioutil.ReadAll(resp.Body)
		if err != nil {
			return nil, err
		}
		resp.Body.Close()
		return nil, fmt.Errorf("response error, code %d, msg: %s", resp.StatusCode, string(data))
	}

	return c.readHttpResponse(resp)
}

func stringOrEmpty(value interface{}) string {
	if value == nil {
		return ""
	}

	return value.(string)
}

// TODO: get app listing id
func genAppID(appSpec map[string]interface{}) string {
	return appSpec["appid"].(string)
}
