package v1alpha1

var (
	// accessTokenExpires = 2 * time.Hour

	// refreshTokenExpires = 3 * time.Hour

	defaultUserspaceRole = "admin"

	defaultSystemWorkspace = "system-workspace"

	defaultSystemWorkspaceRole = "system-workspace-admin"
)

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

type KubesphereError struct {
	Error            string `json:"error"`
	ErrorDescription string `json:"error_description"`
}

type TokenResponse struct {
	AccessToken  string `json:"access_token"`
	TokenType    string `json:"token_type"`
	RefreshToken string `json:"refresh_token"`
	ExpiresIn    int    `json:"expires_in"`
	ExpiresAt    int64  `json:"expires_at,omitempty"`
}

type UnauthorizedError struct {
	APIVersion string      `json:"apiVersion"`
	Code       int64       `json:"code"`
	Kind       string      `json:"kind"`
	Message    string      `json:"message"`
	Metadata   interface{} `json:"metadata"`
	Reason     string      `json:"reason"`
	Status     string      `json:"status"`
}

type UserCreate struct {
	Name        string `json:"name"`
	DisplayName string `json:"display_name"`
	OwnerRole   string `json:"owner_role"`
	Password    string `json:"password"`
	Email       string `json:"email"`
	Description string `json:"description"`
	MemoryLimit string `json:"memory_limit"`
	CpuLimit    string `json:"cpu_limit"`
}

type PasswordReset struct {
	CurrentPassword string `json:"current_password"`
	Password        string `json:"password"`
}

type UserResourceLimit struct {
	MemoryLimit string `json:"memory_limit"`
	CpuLimit    string `json:"cpu_limit"`
}

type PostRefreshToken struct {
	Token string `json:"token"`
}

type UserStatusResponse struct {
	Name    string      `json:"name"`
	Status  string      `json:"status"`
	Address UserAddress `json:"address"`
}

type UserAddress struct {
	Desktop string `json:"desktop"`
	Wizard  string `json:"wizard"`
}
