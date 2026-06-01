package appinstaller

import (
	"encoding/json"

	"github.com/beclab/api/api/app.bytetrade.io/v1alpha1"
)

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
