package app_service

import (
	"fmt"
	"os"
	"strings"

	"bytetrade.io/web3os/bfl/pkg/apis/iam/v1alpha1/operator"
	"bytetrade.io/web3os/bfl/pkg/apiserver/runtime"
	"bytetrade.io/web3os/bfl/pkg/utils"

	"github.com/asaskevich/govalidator"
	"github.com/emicklei/go-restful/v3"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/klog/v2"
)

type URLGenerator func(appname, appid string) string
type URLGeneratorMultiEntrance func(appname, appid string, index int, entrances []Entrance, appDomainConfigs []utils.DefaultThirdLevelDomainConfig) string

func AppUrlGenerator(req *restful.Request, user string) (URLGenerator, error) {
	host := req.Request.Host
	clientHost := req.HeaderParameter("Client-Host") // client specific
	if clientHost != "" {
		host = clientHost
	}

	if host != "" {
		host = strings.Split(host, ":")[0]
	}
	var appURL URLGenerator

	klog.Info("get request host is ", host)
	if !govalidator.IsIP(host) {
		// ssl mode
		// Does the user have a domain by himself or not
		userOp, err := operator.NewUserOperator()
		if err != nil {
			return nil, err
		}

		userCR, err := userOp.GetUser(user)
		if err != nil {
			return nil, err
		}

		isEphemeral, zone, err := userOp.GetUserDomainType(userCR)

		if err != nil {
			klog.Error("get user crd error ", user, err)
			return nil, err
		} else {
			if isEphemeral {
				// Found the new user without zone binding, we need proxy user's apps routes via
				// the zone of the user's creator.
				// At the meanwhile, the zone returned by the function is creator's zone.
				klog.Info("new user: ", user)
				appURL = func(appname, appid string) string {
					return fmt.Sprintf("%s-%s.%s", appid, user, zone)
				}
			} else {
				appURL = func(appname, appid string) string {
					return fmt.Sprintf("%s.%s", appid, zone)
				}
			}

		}

	} else {
		// to get the app serve port on the bfl ingress service node port, find the service from
		// current namespace, since ingress and api in the same pod
		// in cluster mode, use k8s downward api
		// spec:
		//   containers:
		//   - env:
		// 	   - name: BFL_NAMESPACE
		// 	     valueFrom:
		// 		   fieldRef:
		// 		     fieldPath: metadata.namespace
		currentNamepspace := os.Getenv("BFL_NAMESPACE")
		client, err := runtime.NewKubeClientInCluster()
		if err != nil {
			return nil, err
		}

		services := client.Kubernetes().CoreV1().Services(currentNamepspace)
		service, err := services.Get(req.Request.Context(), "bfl", metav1.GetOptions{})
		if err != nil {
			return nil, err
		}

		// if service.Spec.Type != corev1.ServiceTypeNodePort {
		// 	return nil, errors.New("bfl service is not a NodePort")
		// }

		// FIXME: in IP mode, apps are invisible
		ip := host

		appUrlMap := make(map[string]string)
		for _, sp := range service.Spec.Ports {
			appUrlMap[sp.Name] = fmt.Sprintf("%s:%d", ip, sp.Port)
		}

		appURL = func(appname, appid string) string {
			url, ok := appUrlMap["app-"+appname]
			if !ok {
				klog.Errorf("app [ %s ]'s ServicePort not found !")
				return ""
			}

			return url
		}

	}

	return appURL, nil
}

func AppUrlGeneratorMultiEntrance(req *restful.Request, user string) (URLGeneratorMultiEntrance, error) {
	host := req.Request.Host
	clientHost := req.HeaderParameter("Client-Host") // client specific
	if clientHost != "" {
		host = clientHost
	}

	if host != "" {
		host = strings.Split(host, ":")[0]
	}
	var appURL URLGeneratorMultiEntrance

	klog.Info("get request host is ", host)
	if !govalidator.IsIP(host) {
		// ssl mode
		// Does the user have a domain by himself or not
		userOp, err := operator.NewUserOperator()
		if err != nil {
			return nil, err
		}

		userCR, err := userOp.GetUser(user)
		if err != nil {
			return nil, err
		}

		isEphemeral, zone, err := userOp.GetUserDomainType(userCR)
		if err != nil {
			klog.Error("get user crd error ", user, err)
			return nil, err
		} else {
			if isEphemeral {
				// Found the new user without zone binding, we need proxy user's apps routes via
				// the zone of the user's creator.
				// At the meanwhile, the zone returned by the function is creator's zone.
				klog.Info("new user: ", user)
				appURL = func(appname, appid string, index int, entrances []Entrance, appDomainConfigs []utils.DefaultThirdLevelDomainConfig) string {
					return fmt.Sprintf("%s%d-%s.%s", appid, index, user, zone)
				}
			} else {
				appURL = func(appname, appid string, index int, entrances []Entrance, appDomainConfigs []utils.DefaultThirdLevelDomainConfig) string {
					for _, adc := range appDomainConfigs {
						if adc.EntranceName == entrances[index].Name && len(adc.ThirdLevelDomain) > 0 {
							return fmt.Sprintf("%s.%s", adc.ThirdLevelDomain, zone)
						}
					}

					return fmt.Sprintf("%s%d.%s", appid, index, zone)
				}
			}

		}

	} else {
		// to get the app serve port on the bfl ingress service node port, find the service from
		// current namespace, since ingress and api in the same pod
		// in cluster mode, use k8s downward api
		// spec:
		//   containers:
		//   - env:
		// 	   - name: BFL_NAMESPACE
		// 	     valueFrom:
		// 		   fieldRef:
		// 		     fieldPath: metadata.namespace
		currentNamepspace := os.Getenv("BFL_NAMESPACE")
		client, err := runtime.NewKubeClientInCluster()
		if err != nil {
			return nil, err
		}

		services := client.Kubernetes().CoreV1().Services(currentNamepspace)
		service, err := services.Get(req.Request.Context(), "bfl", metav1.GetOptions{})
		if err != nil {
			return nil, err
		}

		//ip := utils.GetMyExternalIPAddr()
		ip := host

		appUrlMap := make(map[string]string)
		for _, sp := range service.Spec.Ports {
			appUrlMap[sp.Name] = fmt.Sprintf("%s:%d", ip, sp.Port)
		}

		appURL = func(appname, appid string, index int, entrances []Entrance, appDomainConfigs []utils.DefaultThirdLevelDomainConfig) string {
			url, ok := appUrlMap["app-"+appname]
			if !ok {
				klog.Errorf("app [ %s ]'s ServicePort not found !")
				return ""
			}

			return url
		}

	}

	return appURL, nil
}

// func isLocalHost(host string) bool {
// 	hostSplit := strings.Split(host, ".")
// 	if len(hostSplit) < 4 {
// 		return false
// 	}

// 	/*
// 	 1. com
// 	 2. example
// 	 3. user
// 	 4. local
// 	*/
// 	return hostSplit[len(hostSplit)-4] == "local"
// }
