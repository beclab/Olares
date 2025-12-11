package v1

import (
	"bytetrade.io/web3os/bfl/pkg/constants"
)

type UserInfo struct {
	Name           string `json:"name"`
	OwnerRole      string `json:"owner_role"`
	TerminusName   string `json:"terminusName"`
	IsEphemeral    bool   `json:"is_ephemeral"`
	Zone           string `json:"zone"`
	CreatedUser    string `json:"created_user"`
	WizardComplete bool   `json:"wizard_complete"`
	AccessLevel    *int   `json:"access_level,omitempty"`
}

type IPAddress struct {
	IsNatted bool `json:"is_natted"`

	Internal string `json:"internal"`

	External string `json:"external"`

	MasterInternalIP string `json:"masterInternalIP,omitempty"`

	MasterExternalIP string `json:"masterExternalIP"`
}

// Depreacted
type TerminusInfo struct {
	TerminusName    string                 `json:"terminusName"`
	WizardStatus    constants.WizardStatus `json:"wizardStatus"`
	Selfhosted      bool                   `json:"selfhosted"`
	TailScaleEnable bool                   `json:"tailScaleEnable"`
	OsVersion       string                 `json:"osVersion"`
	LoginBackground string                 `json:"loginBackground"`
	Avatar          string                 `json:"avatar"`
	TerminusID      string                 `json:"terminusId"`
	UserDID         string                 `json:"did"`
	ReverseProxy    string                 `json:"reverseProxy"`
	Terminusd       string                 `json:"terminusd"`
}

type OlaresInfo struct {
	OlaresID           string                 `json:"olaresId"`
	WizardStatus       constants.WizardStatus `json:"wizardStatus"`
	EnableReverseProxy bool                   `json:"enableReverseProxy"`
	TailScaleEnable    bool                   `json:"tailScaleEnable"`
	OsVersion          string                 `json:"osVersion"`
	LoginBackground    string                 `json:"loginBackground"`
	Avatar             string                 `json:"avatar"`
	ID                 string                 `json:"id"`
	UserDID            string                 `json:"did"`
	Olaresd            string                 `json:"olaresd"`
}

type MyAppsParam struct {
	IsLocal bool `json:"isLocal"`
}
