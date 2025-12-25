package apiserver

import (
	"github.com/beclab/Olares/framework/app-service/pkg/apiserver/api"
	"github.com/beclab/Olares/framework/app-service/pkg/constants"
	"github.com/beclab/Olares/framework/app-service/pkg/helm"
	"github.com/beclab/Olares/framework/app-service/pkg/utils"
	apputils "github.com/beclab/Olares/framework/app-service/pkg/utils/app"
	"github.com/emicklei/go-restful/v3"
)

func (h *Handler) releaseVersion(req *restful.Request, resp *restful.Response) {
	appName := req.PathParameter(ParamAppName)
	owner := req.Attribute(constants.UserContextAttribute)
	appNamespace, err := utils.AppNamespace(appName, owner.(string), "")
	if err != nil {
		api.HandleError(resp, req, err)
		return
	}

	actionConfig, _, err := helm.InitConfig(h.kubeConfig, appNamespace)
	if err != nil {
		api.HandleError(resp, req, err)
		return
	}
	version, _, err := apputils.GetDeployedReleaseVersion(actionConfig, appName)
	if err != nil {
		api.HandleError(resp, req, err)
		return
	}
	resp.WriteAsJson(api.ReleaseVersionResponse{
		Response: api.Response{Code: 200},
		Data:     api.ReleaseVersionData{Version: version},
	})
}
