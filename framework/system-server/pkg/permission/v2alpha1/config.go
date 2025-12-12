package v2alpha1

import "github.com/brancz/kube-rbac-proxy/pkg/authn"

type AuthnConfig struct {
	authn.AuthnConfig
	LLDAP LLDAPConfig
}

type LLDAPConfig struct {
	Server string
	Port   int
}
