package app_service

import (
	"bytetrade.io/web3os/bfl/pkg/constants"
	"fmt"
	"net/url"
	"strings"

	"bytetrade.io/web3os/bfl/pkg/client/dynamic_client/users"
	appv1 "github.com/beclab/api/api/app.bytetrade.io/v1alpha1"
	"github.com/emicklei/go-restful/v3"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/util/validation"
)

// Local accounts need neither a public zone nor a human owner. Read the same
// machine local name used by olaresd's mDNS publisher and the shared ingress.
func localAppURLGenerators(req *restful.Request) (URLGenerator, URLGeneratorMultiEntrance, error) {
	client, err := users.NewResourceTerminusClient()
	if err != nil {
		return nil, nil, err
	}
	terminus, err := client.Get(req.Request.Context(), "terminus", metav1.GetOptions{})
	if err != nil {
		return nil, nil, err
	}
	return localURLGenerators(terminus.Spec.LocalName)
}

func localURLGenerators(localName string) (URLGenerator, URLGeneratorMultiEntrance, error) {
	name := strings.ToLower(strings.TrimSpace(localName))
	if name == "" || len(validation.IsDNS1123Label(name)) != 0 {
		return nil, nil, fmt.Errorf("machine local name is missing or invalid")
	}
	zone := name + ".olares.local"
	single := func(appname, appid string) string { return fmt.Sprintf("%s.%s", appid, zone) }
	multi := func(appname, appid string, index int, entrances []*appv1.Entrance, domains []appv1.DefaultThirdLevelDomainConfig) string {
		prefix := appv1.ResolveEntranceIDWithDefaultThirdLevelDomainOverride(entrances, index, appid, domains)
		return fmt.Sprintf("%s.%s", prefix, zone)
	}
	return single, multi, nil
}

// The request may arrive through a shared frontend or a Provider.
func requestUsesLocalURLs(req *restful.Request) bool {
	if local, _ := req.Attribute(constants.MyAppsIsLocalKey).(bool); local {
		return true
	}
	host := req.HeaderParameter("Client-Host")
	if host == "" {
		host = req.HeaderParameter("X-Forwarded-Host")
	}
	if host == "" {
		host = req.Request.Host
	}
	u, err := url.Parse("//" + host)
	if err != nil {
		return false
	}
	hostname := strings.TrimSuffix(strings.ToLower(u.Hostname()), ".")
	return hostname == "olares.local" || strings.HasSuffix(hostname, ".olares.local")
}
