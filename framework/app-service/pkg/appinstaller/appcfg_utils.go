package appinstaller

import (
	"encoding/json"

	"github.com/beclab/Olares/framework/app-service/api/app.bytetrade.io/v1alpha1"
	"github.com/beclab/Olares/framework/app-service/pkg/apiserver/api"
	"github.com/beclab/Olares/framework/app-service/pkg/appcfg"
	"github.com/beclab/Olares/framework/app-service/pkg/client/clientset"

	"github.com/emicklei/go-restful/v3"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// GetAppConfigFromCRD et app uninstallation config from crd
func GetAppConfigFromCRD(app, owner string,
	client *clientset.ClientSet, req *restful.Request) (*appcfg.ApplicationConfig, error) {
	// run with request context for incoming client
	applist, err := client.AppClient.AppV1alpha1().Applications().List(req.Request.Context(), metav1.ListOptions{})
	if err != nil {
		return nil, err
	}

	// get by application's owner and name
	for _, a := range applist.Items {
		if a.Spec.Owner == owner && a.Spec.Name == app {
			// TODO: other configs
			return &appcfg.ApplicationConfig{
				AppName:   app,
				Namespace: a.Spec.Namespace,
				//ChartsName: "charts/apps",
				OwnerName: owner,
			}, nil
		}
	}

	return nil, api.ErrResourceNotFound
}

func ToEntrances(s string) (entrances []v1alpha1.Entrance, err error) {
	err = json.Unmarshal([]byte(s), &entrances)
	if err != nil {
		return entrances, err
	}

	return entrances, nil
}

func ToEntrancesLabel(entrances []v1alpha1.Entrance) string {
	serviceLabel, _ := json.Marshal(entrances)
	return string(serviceLabel)
}

func ToAppTCPUDPPorts(ports []v1alpha1.ServicePort) string {
	ret := make([]v1alpha1.ServicePort, 0)
	for _, port := range ports {
		protos := []string{port.Protocol}
		if port.Protocol == "" {
			protos = []string{"tcp", "udp"}
		}
		for _, proto := range protos {
			ret = append(ret, v1alpha1.ServicePort{
				Name:              port.Name,
				Host:              port.Host,
				Port:              port.Port,
				ExposePort:        port.ExposePort,
				Protocol:          proto,
				AddToTailscaleAcl: port.AddToTailscaleAcl,
			})
		}
	}
	portsLabel, _ := json.Marshal(ret)
	return string(portsLabel)
}

func ToTailScale(tailScale v1alpha1.TailScale) string {
	tailScaleLabel, _ := json.Marshal(tailScale)
	return string(tailScaleLabel)
}
