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

func (c *Client) getAppListFromData(apps []map[string]interface{}) ([]*AppInfo, error) {

	var res []*AppInfo
	for _, data := range apps {
		var appEntrances []Entrance
		var appSharedEntrances []Entrance
		appPorts := make([]ServicePort, 0)
		appACLs := make([]ACL, 0)

		appSpec, ok := data["spec"].(map[string]interface{})
		if !ok {
			klog.Error("get app info error: ", data)
			return nil, errors.New("app info is invalid")
		}

		isSysApp := appSpec["isSysApp"].(bool)

		// get app settings to filter system service not to list
		var title, target, state, requiredGPU, defaultThirdLevelDomainConfig string
		isClusterScoped, mobileSupported := false, false
		settings, ok := appSpec["settings"]
		if ok {
			settingsMap := settings.(map[string]interface{})
			_, ok = settingsMap["system_service"]
			if ok {
				// It is the system service, not app
				continue
			} // end ok

			if t, ok := settingsMap["title"]; ok {
				title = t.(string)
			}
			if t, ok := settingsMap["clusterScoped"]; ok && t == "true" {
				isClusterScoped = true
			}
			if t, ok := settingsMap["defaultThirdLevelDomainConfig"]; ok {
				defaultThirdLevelDomainConfig = t.(string)
			}

			if t, ok := settingsMap["target"]; ok {
				target = t.(string)
			}
			if t, ok := settingsMap["mobileSupported"]; ok && t == "true" {
				mobileSupported = true
			}
			if t, ok := settingsMap["requiredGPU"]; ok {
				requiredGPU = t.(string)
			}
		}

		entranceStatusesMap := make(map[string]map[string]interface{})
		status, ok := data["status"].(map[string]interface{})
		if ok {
			if t, ok := status["state"]; ok {
				state = t.(string)
			}
			entranceStatuses, ok := status["entranceStatuses"].([]interface{})
			if ok {
				for _, es := range entranceStatuses {
					if e, ok := es.(map[string]interface{}); ok {
						entranceStatusesMap[e["name"].(string)] = e
					}

				}
			} else {
				klog.Infof("error: data[entranceStatues], %v", ok)
			}
		}
		klog.Infof("entranceStatusesMap: %v", entranceStatusesMap)

		entrances, ok := appSpec["entrances"]
		if ok {
			entrancesInterface := entrances.([]interface{})
			for _, entranceInterface := range entrancesInterface {
				entranceMap := entranceInterface.(map[string]interface{})
				var appEntrance Entrance
				if t, ok := entranceMap["name"]; ok {
					appEntrance.Name = stringOrEmpty(t)
				}

				if t, ok := entranceMap["title"]; ok {
					appEntrance.Title = stringOrEmpty(t)
				}

				if t, ok := entranceMap["icon"]; ok {
					appEntrance.Icon = stringOrEmpty(t)
				}
				if t, ok := entranceMap["invisible"]; ok && t.(bool) == true {
					appEntrance.Invisible = true
				}
				if t, ok := entranceMap["authLevel"]; ok {
					appEntrance.AuthLevel = stringOrEmpty(t)
				}
				if t, ok := entranceMap["openMethod"]; ok {
					appEntrance.OpenMethod = stringOrEmpty(t)
				} else {
					appEntrance.OpenMethod = "default"
				}
				if t, ok := entranceStatusesMap[appEntrance.Name]; ok {
					entranceState := t["state"]
					appEntrance.State = stringOrEmpty(entranceState)
					appEntrance.Reason = stringOrEmpty(t["reason"])
					appEntrance.Message = stringOrEmpty(t["message"])
				} else {
					appEntrance.State = state
				}

				appEntrances = append(appEntrances, appEntrance)
			}
		}

		sharedEntrances, ok := appSpec["sharedEntrances"]
		if ok {
			entrancesInterface := sharedEntrances.([]interface{})
			for _, entranceInterface := range entrancesInterface {
				entranceMap := entranceInterface.(map[string]interface{})
				var appEntrance Entrance
				if t, ok := entranceMap["name"]; ok {
					appEntrance.Name = stringOrEmpty(t)
				}
				if t, ok := entranceMap["url"]; ok {
					appEntrance.URL = stringOrEmpty(t)
				}
				if t, ok := entranceMap["title"]; ok {
					appEntrance.Title = stringOrEmpty(t)
				}
				if t, ok := entranceMap["icon"]; ok {
					appEntrance.Icon = stringOrEmpty(t)
				}
				if t, ok := entranceMap["invisible"]; ok && t.(bool) == true {
					appEntrance.Invisible = true
				}
				if t, ok := entranceMap["authLevel"]; ok {
					appEntrance.AuthLevel = stringOrEmpty(t)
				}
				// appEntrance.State = state

				appSharedEntrances = append(appSharedEntrances, appEntrance)
			}
		}

		ports, ok := appSpec["ports"]
		if ok {
			portsInterface := ports.([]interface{})
			for _, p := range portsInterface {
				portsMap := p.(map[string]interface{})
				var appPort ServicePort
				if t, ok := portsMap["exposePort"]; ok {
					appPort.ExposePort = int32(t.(float64))
				}
				if t, ok := portsMap["host"]; ok {
					appPort.Host = stringOrEmpty(t)
				}
				if t, ok := portsMap["name"]; ok {
					appPort.Name = stringOrEmpty(t)
				}
				if t, ok := portsMap["port"]; ok {
					appPort.Port = int32(t.(float64))
				}
				if t, ok := portsMap["protocol"]; ok {
					appPort.Protocol = stringOrEmpty(t)
				}
				appPorts = append(appPorts, appPort)
			}
		}
		acls, ok := appSpec["tailscaleAcls"]
		if ok {
			aclInterface := acls.([]interface{})
			for _, a := range aclInterface {
				aclMap := a.(map[string]interface{})
				var tailscaleACL ACL
				if t, ok := aclMap["action"]; ok {
					tailscaleACL.Action = stringOrEmpty(t)
				}
				if t, ok := aclMap["src"]; ok {
					srcInterface := t.([]interface{})
					src := make([]string, 0)
					for _, s := range srcInterface {
						src = append(src, s.(string))
					}
					tailscaleACL.Src = src
				}
				if t, ok := aclMap["proto"]; ok {
					tailscaleACL.Proto = stringOrEmpty(t)
				}
				if t, ok := aclMap["dst"]; ok {
					dstInterface := t.([]interface{})
					dst := make([]string, 0)
					for _, d := range dstInterface {
						dst = append(dst, d.(string))
					}
					tailscaleACL.Dst = dst
				}
				appACLs = append(appACLs, tailscaleACL)
			}
		}

		res = append(res, &AppInfo{
			ID:                            genAppID(appSpec),
			Name:                          stringOrEmpty(appSpec["name"]),
			RawAppName:                    stringOrEmpty(appSpec["rawAppName"]),
			Namespace:                     stringOrEmpty(appSpec["namespace"]),
			DeploymentName:                stringOrEmpty(appSpec["deployment"]),
			Owner:                         stringOrEmpty(appSpec["owner"]),
			Icon:                          stringOrEmpty(appSpec["icon"]),
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
