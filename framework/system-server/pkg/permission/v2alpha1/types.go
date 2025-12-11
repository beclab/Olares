package v2alpha1

type RegisterResp struct {
}

type PermissionRequire struct {
	ProviderAppName   string  `json:"provider_app_name"`
	ProviderName      string  `json:"provider_name"`
	ProviderNamespace string  `json:"provider_namespace"`
	ServiceAccount    *string `json:"service_account,omitempty"`
	ProviderDomain    string  `json:"provider_domain,omitempty"`
}

type PermissionRegister struct {
	App   string              `json:"app"`
	AppID string              `json:"appid"`
	Perm  []PermissionRequire `json:"perm"`
}
