package v1alpha1

type AccessLevel uint64

const (
	_ AccessLevel = iota

	WorldWide
	Public
	Protected
	Private
)

type AuthPolicy string

const (
	OneFactor = "one_factor"
	TwoFactor = "two_factor"
)

const DefaultAuthPolicy = TwoFactor

const DefaultPodsCIDR = "10.233.64.0/18"

type PostTerminusName struct {
	JWSSignature string `json:"jws_signature"`
	DID          string `json:"did"`
}

type LauncherAccessPolicy struct {
	AccessLevel AccessLevel `json:"access_level"`
	AuthPolicy  AuthPolicy  `json:"auth_policy"`
	AllowCIDRs  []string    `json:"allow_cidrs,omitempty"`
}

type PublicDomainAccessPolicy struct {
	DenyAll int `json:"deny_all"`
	// AllowedDomains []string `json:"allowed_domains"`
}
