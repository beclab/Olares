package apiserver

import (
	"fmt"
	"sort"

	"github.com/beclab/Olares/framework/app-service/pkg/apiserver/api"
	"github.com/beclab/Olares/framework/app-service/pkg/appcfg"

	"github.com/emicklei/go-restful/v3"
)

func (h *Handler) listRegistry(req *restful.Request, resp *restful.Response) {
	apps, err := registryApps(appcfg.ChartsPath)
	if err != nil {
		api.HandleError(resp, req, err)
		return
	}

	appList := make([]*appcfg.AppConfiguration, 0, len(apps))
	for _, a := range apps {
		appList = append(appList, a)
	}

	// sort by rating desc and name asc
	sort.Slice(appList, func(i, j int) bool {
		if appList[i].Metadata.Rating == appList[j].Metadata.Rating {
			return appList[i].Metadata.Name < appList[j].Metadata.Name
		}

		return appList[i].Metadata.Rating > appList[j].Metadata.Rating
	})

	resp.WriteAsJson(appList)
}

func (h *Handler) registryGet(req *restful.Request, resp *restful.Response) {
	app := req.PathParameter(ParamAppName)

	apps, err := registryApps(appcfg.ChartsPath)
	if err != nil {
		api.HandleError(resp, req, err)
		return
	}
	if a, ok := apps[app]; ok {
		resp.WriteAsJson(a)
		return
	}

	api.HandleNotFound(resp, req, fmt.Errorf("app [%s] not found", app))
}
