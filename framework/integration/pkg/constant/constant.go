package constant

const (
	InfisicalNamespace = "os-protected"
	InfisicalService   = "http://infisical-service." + InfisicalNamespace
	DefaultEnvironment = "prod"
	DefaultWorkspace   = "settings"

	HeaderUserName  = "X-Bfl-User"
	EnvInfisicalUrl = "INFISICAL_URL"
	EnvEnvironment  = "ENVIRONMENT"
	EnvWorkspace    = "WORKSPACE"
)

var (
	InfisicalAddr string
	InfisicalPort = "8080"
	Environment   string
	Workspace     string
)
