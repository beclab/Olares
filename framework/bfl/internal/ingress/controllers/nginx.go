package controllers

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"

	"bytetrade.io/web3os/bfl/internal/ingress/controllers/config"
	"bytetrade.io/web3os/bfl/pkg/constants"
	"bytetrade.io/web3os/bfl/pkg/utils"

	"k8s.io/klog/v2"
)

var nonAppServers = []NonAppServer{
	{
		Name:        "auth",
		SvcEndpoint: "http://authelia-svc.%s.svc.cluster.local:80",
		AuthEnabled: false,
	},
	{
		Name:        "desktop",
		SvcEndpoint: "http://edge-desktop.%s.svc.cluster.local:80",
		AuthEnabled: true,
	},
	{
		Name:        "wizard",
		SvcEndpoint: "http://wizard.%s.svc.cluster.local:80",
		AuthEnabled: true,
	},
}

type NonAppServer struct {
	Name           string
	LocationPrefix string
	SvcEndpoint    string
	AuthEnabled    bool
}

func (r *NginxController) genNonAppServers(zone string, isEphemeral bool, language string) []config.Server {
	servers := make([]config.Server, 0)
	for _, app := range nonAppServers {
		hostname := fmt.Sprintf("%s.%s", app.Name, zone)
		if isEphemeral {
			hostname = fmt.Sprintf("%s-%s.%s", app.Name, constants.Username, zone)
		}

		servers = append(servers, config.Server{
			Hostname:   hostname,
			Aliases:    []string{r.makeLocalHost(hostname)},
			EnableAuth: app.AuthEnabled,
			EnableSSL:  true,
			Locations: []config.Location{
				{
					Prefix:    utils.StringOrDefault(app.LocationPrefix, "/"),
					ProxyPass: fmt.Sprintf(app.SvcEndpoint, constants.Namespace),
				},
			},
			Language: language,
		})
	}
	return servers
}

func (r *NginxController) addDomainServers(ctx context.Context, isEphemeral bool, zone string, language string) []config.Server {
	servers := make([]config.Server, 0)
	zoneToken := strings.Split(zone, ".")
	zoneAliases := []string{}
	if len(zoneToken) > 1 {
		zoneAliases = append(zoneAliases, fmt.Sprintf("%s.%s", zoneToken[0], "olares.local"))
	}

	profile := config.Server{
		Hostname:  zone,
		Aliases:   zoneAliases,
		EnableSSL: true,
		Locations: []config.Location{
			{
				Prefix:    "/",
				ProxyPass: fmt.Sprintf(constants.IndexAppEndpoint, constants.Username),
			},
		},
		EnableAuth: false,
		Language:   language,
	}

	formatDomain := func(customPrefixDomain string) []string {
		var d []string

		if customPrefixDomain != "" {
			extAppHostName := fmt.Sprintf("%s.%s", customPrefixDomain, zone)

			if isEphemeral {
				extAppHostName = fmt.Sprintf("%s-%s.%s", customPrefixDomain, constants.Username, zone)
			}

			d = append(d, extAppHostName, r.makeLocalHost(extAppHostName))
		}
		return d
	}

	servers = append(servers, profile)

	// add apps servers
	for _, app := range r.apps {
		if len(app.Spec.Entrances) == 0 {
			continue
		}
		var appDomainConfigs []utils.DefaultThirdLevelDomainConfig
		if len(app.Spec.Settings["defaultThirdLevelDomainConfig"]) > 0 {
			err := json.Unmarshal([]byte(app.Spec.Settings["defaultThirdLevelDomainConfig"]), &appDomainConfigs)
			if err != nil {
				klog.Errorf("unmarshal defaultThirdLevelDomainConfig error %v", err)
			}

		}

		//entrancecounts := len(app.Spec.Entrances)
		for index, entrance := range app.Spec.Entrances {
			if entrance.Host == "" {
				continue
			}
			prefix := getAppEntrancesHostName(app.Spec.Entrances, index, app.Spec.Name, appDomainConfigs)
			customPrefixDomainName := ""

			customDomainEntrancesMap, err := getSettingsMap(&app, constants.ApplicationCustomDomain)
			if err != nil {
				klog.Warningf("failed to unmarshal application custom domain, %q, %s, %s, %v", prefix, app.Spec.Name, app.Spec.Appid, err)
			}

			customDomainEntranceMap, ok := customDomainEntrancesMap[entrance.Name]

			if app.Spec.Appid != "" && !app.Spec.IsSysApp { // third-party application
				prefix = getAppEntrancesHostName(app.Spec.Entrances, index, app.Spec.Appid, appDomainConfigs)
				if ok {
					if customDomainEntranceMap != nil {
						customPrefixDomainName = customDomainEntranceMap[constants.ApplicationThirdLevelDomain]
					}
				}
			}
			klog.Infof("add domain server, app prefix: %q, %s", prefix, utils.ToJSON(app))

			appHostname := fmt.Sprintf("%s.%s", prefix, zone)
			if isEphemeral {
				appHostname = fmt.Sprintf("%s-%s.%s", prefix, constants.Username, zone)
			}

			appLocalHost := r.makeLocalHost(appHostname)

			_, enableOIDC := app.Spec.Settings["oidc.client.id"]

			s := config.Server{
				Hostname:   appHostname,
				Aliases:    append([]string{appLocalHost}, formatDomain(customPrefixDomainName)...),
				EnableSSL:  true,
				EnableAuth: true,
				Locations: []config.Location{
					{
						Prefix:    "/",
						ProxyPass: fmt.Sprintf("http://%s.%s.svc.cluster.local:%d", entrance.Host, app.Spec.Namespace, entrance.Port),
					},
				},
				EnableOIDC:            enableOIDC,
				EnableWindowPushState: entrance.WindowPushState,
				Language:              language,
			}

			// server patches
			if patches, ok := patches[prefix]; ok {
				for _, patch := range patches {
					_, err = patch(ctx, r, &s)
					if err != nil {
						klog.Errorf("failed to apply patch for app %s, %v", app.Spec.Name, err)
					}
				}
			}

			servers = append(servers, s)
		}
	}

	// add non application servers
	_servers := r.genNonAppServers(zone, isEphemeral, language)

	if len(_servers) > 0 {
		servers = append(servers, _servers...)
	}

	return servers
}

func (r *NginxController) makeLocalHost(hostname string) string {
	hostToken := strings.Split(hostname, ".")
	if len(hostToken) < 2 {
		return hostname
	}
	return strings.Join([]string{hostToken[0], hostToken[1], "olares", "local"}, ".")
}
