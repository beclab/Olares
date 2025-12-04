package appcfg

type PermissionRegister struct {
	App   string              `json:"app"`
	AppID string              `json:"appid"`
	Perm  []PermissionRequire `json:"perm"`
}

type PermissionRequire struct {
	ProviderAppName   string  `json:"provider_app_name"`
	ProviderName      string  `json:"provider_name"`
	ProviderNamespace string  `json:"provider_namespace"`
	ServiceAccount    *string `json:"service_account,omitempty"`
	ProviderDomain    string  `json:"provider_domain,omitempty"`
}

type Header struct {
	Code    int    `json:"code"`
	Message string `json:"message"`
}

type RegisterResp struct {
	Header
	Data RegisterRespData `json:"data"`
}

type RegisterRespData struct {
	AppKey    string `json:"app_key"`
	AppSecret string `json:"app_secret"`
}

type UnregisterResp struct {
	Header
}

type ProviderRegisterRequest struct {
	AppName      string         `json:"app_name"`
	AppNamespace string         `json:"app_namespace"`
	Providers    []*ProviderCfg `json:"providers"`
}

type ProviderCfg struct {
	Provider `json:",inline"`
	Domain   string `json:"domain"`
	Service  string `json:"service"`
}

type PermissionCfg struct {
	*ProviderPermission
	Port   int    // not in yaml, cannot be set in OlaresManifest
	Svc    string // not in yaml, cannot be set in OlaresManifest
	Domain string // not in yaml, cannot be set in OlaresManifest
	Paths  []string
}
