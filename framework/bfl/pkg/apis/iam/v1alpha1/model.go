package v1alpha1

type UserInfo struct {
	UID               string `json:"uid"`
	Name              string `json:"name"`
	DisplayName       string `json:"display_name"`
	Description       string `json:"description"`
	Email             string `json:"email"`
	State             string `json:"state"`
	LastLoginTime     *int64 `json:"last_login_time"`
	CreationTimestamp int64  `json:"creation_timestamp"`
	Avatar            string `json:"avatar"`

	TerminusName   string `json:"terminusName"`
	WizardComplete bool   `json:"wizard_complete"`

	Roles []string `json:"roles"`

	MemoryLimit string `json:"memory_limit"`
	CpuLimit    string `json:"cpu_limit"`
}

type LoginRecord struct {
	Type      string `json:"type"`
	Success   bool   `json:"success"`
	SourceIP  string `json:"sourceIP"`
	UserAgent string `json:"user_agent"`
	Reason    string `json:"reason"`
	LoginTime *int64 `json:"login_time"`
}

type PasswordReset struct {
	CurrentPassword string `json:"current_password"`
	Password        string `json:"password"`
}
