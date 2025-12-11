package permission

import (
	"time"

	sysv1alpha1 "bytetrade.io/web3os/system-server/pkg/apis/sys/v1alpha1"
)

type AccessTokenResponse struct {
	AccessToken string    `json:"access_token"`
	ExpiredAt   time.Time `json:"expired_at"`
}

type AccessTokenRequest struct {
	AppKey    string                        `json:"app_key"`
	Timestamp int64                         `json:"timestamp"`
	Token     string                        `json:"token"`
	Perm      sysv1alpha1.PermissionRequire `json:"perm"`
}

type PermissionControlSet struct {
	Ctrl *PermissionControl
	Mgr  *AccessManager
}

type PermissionRegister struct {
	App   string                          `json:"app"`
	AppID string                          `json:"appid"`
	Perm  []sysv1alpha1.PermissionRequire `json:"perm"`
}

type RegisterResp struct {
	AppKey    string `json:"app_key"`
	AppSecret string `json:"app_secret"`
}
