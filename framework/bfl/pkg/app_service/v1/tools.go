package app_service

import (
	"fmt"
	"strings"

	"bytetrade.io/web3os/bfl/pkg/client/dynamic_client/users"
	appv1 "github.com/beclab/api/api/app.bytetrade.io/v1alpha1"
	"github.com/emicklei/go-restful/v3"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

type URLGenerator func(appname, appid string) string
type URLGeneratorMultiEntrance func(appname, appid string, index int, entrances []*appv1.Entrance, appDomainConfigs []appv1.DefaultThirdLevelDomainConfig) string

func appURLGenerators(req *restful.Request) (URLGenerator, URLGeneratorMultiEntrance, error) {
	client, err := users.NewResourceTerminusClient()
	if err != nil {
		return nil, nil, err
	}
	terminus, err := client.Get(req.Request.Context(), "terminus", metav1.GetOptions{})
	if err != nil {
		return nil, nil, err
	}
	zone := strings.TrimSuffix(strings.TrimSpace(terminus.Spec.Settings[users.SettingsDomainNameKey]), ".")
	if requestUsesLocalURLs(req) || zone == "" || zone == "olares.local" {
		return localURLGenerators(terminus.Spec.LocalName)
	}
	single := func(appname, appid string) string { return fmt.Sprintf("%s.%s", appid, zone) }
	multi := func(appname, appid string, index int, entrances []*appv1.Entrance, domains []appv1.DefaultThirdLevelDomainConfig) string {
		prefix := appv1.ResolveEntranceIDWithDefaultThirdLevelDomainOverride(entrances, index, appid, domains)
		return fmt.Sprintf("%s.%s", prefix, zone)
	}
	return single, multi, nil
}

func AppUrlGenerator(req *restful.Request, _ string) (URLGenerator, error) {
	single, _, err := appURLGenerators(req)
	return single, err
}

func AppUrlGeneratorMultiEntrance(req *restful.Request, _ string) (URLGeneratorMultiEntrance, error) {
	_, multi, err := appURLGenerators(req)
	return multi, err
}
