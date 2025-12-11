package v2alpha1

import (
	"strings"

	"k8s.io/klog/v2"
)

type ProviderRegisterRequest struct {
	AppName      string     `json:"app_name"`
	AppNamespace string     `json:"app_namespace"`
	Providers    []Provider `json:"providers"`
}

type Provider struct {
	Name    string   `json:"name"`
	Domain  string   `json:"domain"`
	Service string   `json:"service"`
	Paths   []string `json:"paths"`
	Verbs   []string `json:"verbs"`
}

func GetRoleNameForRef(ref string) string {
	roleName := strings.Join(strings.Split(ref, "/"), ":")
	klog.Info("Getting role name for provider reference, ", ref, ", role name: ", roleName)
	return roleName
}
