package constants

import (
	"fmt"
	"net/url"
	"strings"
)

var (
	APIServerListenAddress string

	// Namespace is the shared BFL namespace.
	Namespace string

	KubeSphereAPIHost string

	UserContextAttribute = "UserName"

	AnnotationUserCreator = "bytetrade.io/creator"

	UserAnnotationLimitsCpuKey    = "bytetrade.io/user-cpu-limit"
	UserAnnotationLimitsMemoryKey = "bytetrade.io/user-memory-limit"

	MyAppsIsLocalKey = "my-apps-is-local"
)

var (
	KubeSphereJwtKey []byte

	// RequestURLWhiteList api url white list, no token authentication
	RequestURLWhiteList = []string{
		"/bfl/apidocs.json",
		"/bfl/backend/v1/",
		"/bfl/info/v1/",
	}
)

var (
	OlaresRemoteService = "https://api.olares.com"

	APIPrefixCertService string

	APIPrefixDNSOPService string

	APIFormatCertGenerateRequest string

	APIFormatCertGenerateStatus string

	APIFormatCertDownload string

	APIDNSAddRecord string

	APIFormatDNSDeleteRecord string

	APIDNSAddCustomDomain string

	APIDNSCheckCustomDomainCname string

	APIMyExternalIP string

	nameParamters = "name=%s"
)

const (
	EnvOlaresSystemRemoteService = "OLARES_SYSTEM_REMOTE_SERVICE"
	UserAuthorizationTokenKey    = "X-Authorization"
	HeaderBflUserKey             = "X-BFL-USER"
)

var (
	AnnotationGroup = "bytetrade.io"

	UserAnnotationTerminusNameKey = fmt.Sprintf("%s/terminus-name", AnnotationGroup)

	UserAnnotationOwnerRole = fmt.Sprintf("%s/owner-role", AnnotationGroup)

	// L4ListenSSLProxyProtocolPort is the one
	// with "listen xxx proxy_protocol",
	// which reads the proxy protocol header upon tcp initialization
	// for proxy agent with proxy protocol data
	// this port must be used to proxy traffic
	L4ListenSSLProxyProtocolPort = "444"

	ReverseProxyAgentImage        = "beclab/reverse-proxy"
	ReverseProxyAgentImageVersion = "v0.1.0"

	UserspaceNameFormat = "user-space-%s"

	UserLauncherAccessLevel = fmt.Sprintf("%s/launcher-access-level", AnnotationGroup)

	UserLauncherAllowCIDR = fmt.Sprintf("%s/launcher-allow-cidr", AnnotationGroup)

	UserLauncherAuthPolicy = fmt.Sprintf("%s/launcher-auth-policy", AnnotationGroup)

	UserCertManagerJWSToken = fmt.Sprintf("%s/jws-token-signature", AnnotationGroup)

	UserCertManagerDID = fmt.Sprintf("%s/user-did", AnnotationGroup)

	UserDenyAllPolicy = fmt.Sprintf("%s/deny-all", AnnotationGroup)

	UserLocation = fmt.Sprintf("%s/location", AnnotationGroup)

	UserTerminusWizardStatus = fmt.Sprintf("%s/wizard-status", AnnotationGroup)

	UserAvatar = fmt.Sprintf("%s/avatar", AnnotationGroup)

	UserLoginBackground      = fmt.Sprintf("%s/login-background", AnnotationGroup)
	UserLoginBackgroundStyle = fmt.Sprintf("%s/login-background-style", AnnotationGroup)
)

var (
	ReverseProxyTypeFRP    = "frp"
	ReverseProxyTypePublic = "public"
	ReverseProxyTypeNone   = "none"
)

var (
	RoleOwner = "owner"
	RoleAdmin = "admin"
)

var (
	ApplicationCustomDomain = "customDomain"

	ApplicationThirdLevelDomain = "third_level_domain"

	ApplicationThirdPartyDomain = "third_party_domain"

	ApplicationReverseProxyType = "reverse_proxy_type"

	ApplicationCustomDomainCnameTarget       = "cname_target"
	ApplicationCustomDomainCnameTargetStatus = "cname_target_status"
	ApplicationCustomDomainCnameStatus       = "cname_status"
)

type CustomDomain int

const (
	CustomDomainIgnore CustomDomain = iota
	CustomDomainAdd
	CustomDomainUpdate
	CustomDomainDelete
	CustomDomainCheck
)

const (
	CustomDomainCnameStatusEmpty        = ""
	CustomDomainCnameStatusSet          = "set"     //
	CustomDomainCnameStatusNotset       = "unset"   // notset
	CustomDomainCnameStatusPending      = "pending" // pending
	CustomDomainCnameStatusCertNotFound = "cert-not-found"
	CustomDomainCnameStatusCertInvalid  = "cert-invalid"
	CustomDomainCnameStatusActive       = "active"  // active
	CustomDomainCnameStatusNone         = "none"    // none
	CustomDomainCnameStatusTimeout      = "timeout" // timeout
	CustomDomainCnameStatusError        = "error"   // error
)

var SystemReservedKeyWords = []string{
	"user",
	"system",
	"space",
	"default",
	"os",
	"kubesphere",
	"kube",
	"kubekey",
	"kubernetes",
	"gpu",
	"tapr",
	"bfl",
	"bytetrade",
	"project",
	"pod",
}

// format: user-name@domain-name
type TerminusName string

func (s TerminusName) UserName() string {
	return s.UserAndDomain()[0]
}

func NewTerminusName(username, domainname string) TerminusName {
	return TerminusName(fmt.Sprintf("%s@%s", username, domainname))
}

func (s TerminusName) UserAndDomain() []string {
	return strings.Split(string(s), "@")
}

func (s TerminusName) UserZone() string {
	return strings.Join(s.UserAndDomain(), ".")
}

type WizardStatus string

const (
	WaitActivateVault     WizardStatus = "wait_activate_vault"
	WaitActivateSystem    WizardStatus = "wait_activate_system"
	SystemActivating      WizardStatus = "system_activating"
	WaitActivateNetwork   WizardStatus = "wait_activate_network"
	SystemActivateFailed  WizardStatus = "system_activate_failed"
	NetworkActivating     WizardStatus = "network_activating"
	NetworkActivateFailed WizardStatus = "network_activate_failed"
	WaitResetPassword     WizardStatus = "wait_reset_password"
	Completed             WizardStatus = "completed"
)

func ReloadEnvDependantVars() error {
	APIPrefixCertService, err := url.JoinPath(OlaresRemoteService, "/cert")
	if err != nil {
		return err
	}
	APIPrefixDNSOPService, err := url.JoinPath(OlaresRemoteService, "/dns/op")
	if err != nil {
		return err
	}

	APIMyExternalIP, err = url.JoinPath(OlaresRemoteService, "/myip/ip")
	if err != nil {
		return err
	}

	APIFormatCertGenerateRequest = APIPrefixCertService + "/generate?" + nameParamters

	APIFormatCertGenerateStatus = APIPrefixCertService + "/status?" + nameParamters

	APIFormatCertDownload = APIPrefixCertService + "/download?" + nameParamters

	APIDNSAddRecord = APIPrefixDNSOPService + "/adddnsrecord"

	APIFormatDNSDeleteRecord = APIPrefixDNSOPService + "/deldnsrecord?" + nameParamters

	APIDNSAddCustomDomain = APIPrefixDNSOPService + "/customhostname"

	APIDNSCheckCustomDomainCname = APIPrefixDNSOPService + "/checkcname"

	return nil
}

const (
	AppEntranceCertConfigMapLabel   = "app.bytetrade.io/custom-domain-cert"
	AppEntranceCertConfigMapNameTpl = "%s-ssl-config"
	AppEntranceCertConfigMapCertKey = "cert"
	AppEntranceCertConfigMapKeyKey  = "key"
	AppEntranceCertConfigMapZoneKey = "zone"
)

func init() {
	ReloadEnvDependantVars()
}
