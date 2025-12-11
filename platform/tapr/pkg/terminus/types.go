package terminus

import "time"

// BFL http types.
type Header struct {
	Code    int    `json:"code"`
	Message string `json:"message"`
}

type Response struct {
	Header

	Data any `json:"data,omitempty"` // data field, optional, object or list.
}

type ApplicationSettingsSubPolicy struct {
	URI      string        `json:"uri"`
	Policy   string        `json:"policy"`
	OneTime  bool          `json:"one_time"`
	Duration time.Duration `json:"valid_duration"`
}

type ApplicationSettingsPolicy struct {
	DefaultPolicy string                          `json:"default_policy"`
	SubPolicies   []*ApplicationSettingsSubPolicy `json:"sub_policies"`
	OneTime       bool                            `json:"one_time"`
	Duration      time.Duration                   `json:"valid_duration"`
}

const (
	ApplicationSettingsPolicyKey = "policy"
	oneFactor                    = "one_factor"
	twoFactor                    = "two_factor"
	deny                         = "deny"
	public                       = "public"
)
