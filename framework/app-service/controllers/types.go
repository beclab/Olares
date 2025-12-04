package controllers

const (
	applicationSettingsPolicyKey = "policy"
	namespaceFinalizer           = "finalizers.bytetrade.io/namespaces"
	userFinalizer                = "finalizers.bytetrade.io/users"
	creator                      = "bytetrade.io/creator"
)

type applicationSettingsSubPolicy struct {
	URI      string `json:"uri"`
	Policy   string `json:"policy"`
	OneTime  bool   `json:"one_time"`
	Duration int32  `json:"valid_duration"`
}

type applicationSettingsPolicy struct {
	DefaultPolicy string                          `json:"default_policy"`
	SubPolicies   []*applicationSettingsSubPolicy `json:"sub_policies"`
	OneTime       bool                            `json:"one_time"`
	Duration      int32                           `json:"valid_duration"`
}
