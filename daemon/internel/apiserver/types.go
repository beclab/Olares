package apiserver

type ClientBindingResp struct {
	UID string `json:"uid"`
}

type ClientBindingReq struct {
	Nonce string `json:"nonce" valid:"required"`
}

type ClientConfirmReq struct {
	UID       string `json:"uid" valid:"required"`
	Challenge string `json:"challenge" valid:"required"`
}
