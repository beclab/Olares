package apis

import (
	"context"
	"encoding/json"
	"fmt"

	"bytetrade.io/web3os/bfl/pkg/api/response"
	"bytetrade.io/web3os/bfl/pkg/apis/iam/v1alpha1/operator"
	"bytetrade.io/web3os/bfl/pkg/apiserver/runtime"
	"bytetrade.io/web3os/bfl/pkg/app_service/v1"
	"bytetrade.io/web3os/bfl/pkg/constants"
	"bytetrade.io/web3os/bfl/pkg/utils"

	iamV1alpha2 "github.com/beclab/api/iam/v1alpha2"
	"github.com/emicklei/go-restful/v3"
	"github.com/pkg/errors"
	"k8s.io/apimachinery/pkg/types"
	"k8s.io/klog/v2"
)

type Base struct {
}

type PostLocale struct {
	Language string `json:"language"`
	Location string `json:"location"`
	// dark/light
	Theme string `json:"theme"`
}

func (b *Base) GetAppViaOwner(appService *app_service.Client) (string, []*app_service.AppInfo, error) {
	apps, err := appService.ListAppInfosByUser(constants.Username)

	return constants.Username, apps, err
}

func (h *Base) GetAppListAndServicePort(req *restful.Request, appService *app_service.Client, getApps func() (string, []*app_service.AppInfo, error)) ([]*app_service.AppInfo, error) {

	user, apps, err := getApps()
	if err != nil {
		return nil, err
	}

	appURL, err := app_service.AppUrlGenerator(req, user)
	if err != nil {
		return nil, err
	}

	appURLMulti, err := app_service.AppUrlGeneratorMultiEntrance(req, user)
	if err != nil {
		return nil, err
	}

	for i, app := range apps {
		apps[i].URL = appURL(app.Name, app.ID)

		if len(apps[i].Entrances) == 0 {
			continue
		}

		if len(apps[i].Entrances) == 1 {
			apps[i].Entrances[0].ID = app.ID
			apps[i].Entrances[0].URL = app.URL
			if apps[i].Entrances[0].Icon == "" {
				apps[i].Entrances[0].Icon = app.Icon
			}
			continue
		}

		var appDomainConfigs []utils.DefaultThirdLevelDomainConfig
		if len(app.DefaultThirdLevelDomainConfig) > 0 {
			err := json.Unmarshal([]byte(app.DefaultThirdLevelDomainConfig), &appDomainConfigs)
			if err != nil {
				klog.Errorf("unmarshal defaultThirdLevelDomainConfig error %v", err)
			}

		}
		// return all entrances, let the frontend filter invisible or not
		// filteredEntrances := make([]app_service.Entrance, 0)
		for j := range apps[i].Entrances {
			apps[i].Entrances[j].ID = fmt.Sprintf("%s%d", app.ID, j)
			for _, adc := range appDomainConfigs {
				if adc.EntranceName == apps[i].Entrances[j].Name && len(adc.ThirdLevelDomain) > 0 {
					apps[i].Entrances[j].ID = adc.ThirdLevelDomain
				}
			}
			apps[i].Entrances[j].URL = appURLMulti(app.Name, app.ID, j, apps[i].Entrances, appDomainConfigs)
			if apps[i].Entrances[j].Icon == "" {
				apps[i].Entrances[j].Icon = app.Icon
			}
			// if !apps[i].Entrances[j].Invisible {
			// 	filteredEntrances = append(filteredEntrances, apps[i].Entrances[j])
			// }
		}
		//		apps[i].Entrances = filteredEntrances
	}

	return apps, nil
}

func (b *Base) IsAdminUser(ctx context.Context) (bool, error) {
	kc, err := runtime.NewKubeClientInCluster()
	if err != nil {
		return false, err
	}

	var user iamV1alpha2.User
	err = kc.CtrlClient().Get(ctx, types.NamespacedName{Name: constants.Username}, &user)
	if err != nil {
		return false, err
	}
	role, ok := user.Annotations[constants.UserAnnotationOwnerRole]
	if !ok {
		return false, errors.Errorf("invalid user %q, no owner role annotation", user.Name)
	}
	return role == constants.RoleOwner || role == constants.RoleAdmin, nil
}

func (b *Base) HandleGetSysConfig(_ *restful.Request, resp *restful.Response) {
	userOp, err := operator.NewUserOperator()
	if err != nil {
		response.HandleError(resp, errors.Errorf("new user operator err, %v", err))
		return
	}
	user, err := userOp.GetUser(constants.Username)
	if err != nil {
		response.HandleError(resp, errors.Errorf("get user sys config err: get user err, %v", err))
		return
	}
	cfg := PostLocale{
		Language: user.Annotations[constants.UserLanguage],
		Location: user.Annotations[constants.UserLocation],
		Theme:    user.Annotations[constants.UserTheme],
	}
	response.Success(resp, &cfg)
}
