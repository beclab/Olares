package apis

import (
	"encoding/json"

	appv1alpha1 "github.com/beclab/api/api/app.bytetrade.io/v1alpha1"

	"bytetrade.io/web3os/bfl/pkg/api/response"
	"bytetrade.io/web3os/bfl/pkg/apis/iam/v1alpha1/operator"
	"bytetrade.io/web3os/bfl/pkg/apiserver/runtime"
	"bytetrade.io/web3os/bfl/pkg/app_service/v1"
	"bytetrade.io/web3os/bfl/pkg/constants"
	"bytetrade.io/web3os/bfl/pkg/userenv"
	"github.com/emicklei/go-restful/v3"
	"github.com/pkg/errors"
	"k8s.io/klog/v2"
)

type Base struct {
}

type PostLocale struct {
	Language string `json:"language"`
	Location string `json:"location"`
	Timezone string `json:"timezone"`
	// dark/light
	Theme string `json:"theme"`
}

func (b *Base) GetAppForUser(username string, appService *app_service.Client) (string, []*app_service.AppInfo, error) {
	apps, err := appService.ListAppInfosByUser(username)

	return username, apps, err
}

func (h *Base) GetAppListAndServicePort(req *restful.Request, getApps func() (string, []*app_service.AppInfo, error)) ([]*app_service.AppInfo, error) {

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

		var appDomainConfigs []appv1alpha1.DefaultThirdLevelDomainConfig
		if len(app.DefaultThirdLevelDomainConfig) > 0 {
			err := json.Unmarshal([]byte(app.DefaultThirdLevelDomainConfig), &appDomainConfigs)
			if err != nil {
				klog.Errorf("unmarshal defaultThirdLevelDomainConfig error %v", err)
			}

		}
		apiEntrances := make([]*appv1alpha1.Entrance, len(apps[i].Entrances))
		for k := range apps[i].Entrances {
			e := apps[i].Entrances[k]
			apiEntrances[k] = &appv1alpha1.Entrance{
				Name:       e.Name,
				Icon:       e.Icon,
				Title:      e.Title,
				AuthLevel:  e.AuthLevel,
				Invisible:  e.Invisible,
				URL:        e.URL,
				OpenMethod: e.OpenMethod,
			}
		}

		// return all entrances, let the frontend filter invisible or not
		// filteredEntrances := make([]app_service.Entrance, 0)
		for j := range apps[i].Entrances {
			apps[i].Entrances[j].ID = appv1alpha1.ResolveEntranceIDWithDefaultThirdLevelDomainOverride(apiEntrances, j, app.ID, appDomainConfigs)
			apps[i].Entrances[j].URL = appURLMulti(app.Name, app.ID, j, apiEntrances, appDomainConfigs)
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

func (b *Base) HandleGetSysConfig(req *restful.Request, resp *restful.Response) {
	username, userErr := runtime.RequestUser(req)
	if userErr != nil {
		response.HandleBadRequest(resp, userErr)
		return
	}
	kc, err := runtime.NewKubeClientInCluster()
	if err != nil {
		response.HandleError(resp, errors.Errorf("get user sys config err: new kube client err, %v", err))
		return
	}
	ctx := req.Request.Context()
	c := kc.CtrlClient()

	// language/theme/timezone are owned by the per-user UserEnvs (see
	// build/user-env.yaml) and read directly from the UserEnv CRs; an unset
	// env reads as empty.
	cfg := PostLocale{}
	targets := []struct {
		env string
		dst *string
	}{
		{userenv.EnvLanguage, &cfg.Language},
		{userenv.EnvTimezone, &cfg.Timezone},
		{userenv.EnvTheme, &cfg.Theme},
	}
	for _, t := range targets {
		v, _, e := userenv.GetValue(ctx, c, username, t.env)
		if e != nil {
			response.HandleError(resp, errors.Errorf("get user sys config err: read userenv %s err, %v", t.env, e))
			return
		}
		*t.dst = v
	}

	// Location has no backing UserEnv; it still lives on the user annotation.
	userOp, err := operator.NewUserOperatorWithContext(req.Request.Context())
	if err != nil {
		response.HandleError(resp, errors.Errorf("get user sys config err: new user operator err, %v", err))
		return
	}
	user, err := userOp.GetUser(username)
	if err != nil {
		response.HandleError(resp, errors.Errorf("get user sys config err: get user err, %v", err))
		return
	}
	cfg.Location = user.Annotations[constants.UserLocation]

	response.Success(resp, &cfg)
}
