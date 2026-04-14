package appcfg

import (
	"context"
	"encoding/json"
	"fmt"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	"github.com/beclab/Olares/framework/app-service/pkg/utils"
)

const (
	ChartsPath = "./charts"
)

func AppChartPath(app string) string {
	return ChartsPath + "/" + app
}

// GetAppInstallationConfig get app installation configuration from app store
func GetAppInstallationConfig(app, owner string) (*ApplicationConfig, error) {
	//chart := AppChartPath(rawAppName)
	appcfg, err := getAppConfigFromAppMgrConfig(app, owner)
	if err != nil {
		return nil, err
	}

	// TODO: app installation namespace
	var namespace string
	if appcfg.Namespace != "" {
		namespace, _ = utils.AppNamespace(app, owner, appcfg.Namespace)
	} else {
		namespace = fmt.Sprintf("%s-%s", app, owner)
	}

	appcfg.Namespace = namespace
	appcfg.OwnerName = owner

	return appcfg, nil
}

func getAppConfigFromAppMgrConfig(appName, owner string) (*ApplicationConfig, error) {
	kclient, err := utils.GetClient()
	if err != nil {
		return nil, err
	}
	name := fmt.Sprintf("%s-%s-%s", appName, owner, appName)
	am, err := kclient.AppV1alpha1().ApplicationManagers().Get(context.TODO(), name, metav1.GetOptions{})
	if err != nil {
		return nil, err
	}
	appConfig := ApplicationConfig{}
	err = json.Unmarshal([]byte(am.Spec.Config), &appConfig)
	if err != nil {
		return nil, err
	}
	return &appConfig, nil

}
