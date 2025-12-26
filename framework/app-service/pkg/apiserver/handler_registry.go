package apiserver

import (
	"fmt"
	"io/ioutil"
	"os"
	"sort"

	"github.com/beclab/Olares/framework/app-service/pkg/apiserver/api"
	"github.com/beclab/Olares/framework/app-service/pkg/appcfg"

	"github.com/emicklei/go-restful/v3"
	"k8s.io/klog/v2"
)

func (h *Handler) listRegistry(req *restful.Request, resp *restful.Response) {
	charts, err := ioutil.ReadDir(appcfg.ChartsPath)
	if err != nil {
		api.HandleError(resp, req, err)
		return
	}

	var appList []*appcfg.AppConfiguration

	for _, c := range charts {
		if c.IsDir() {
			// read app info from chart
			a, err := h.readAppInfo(c)
			if err != nil {
				klog.Errorf("Failed to read app chart err=%v", err)
				continue
			}

			appList = append(appList, a)
		}
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

	charts, err := os.Open(appcfg.AppChartPath(app))
	if err != nil {
		api.HandleError(resp, req, err)
		return
	}
	defer charts.Close()

	if c, err := charts.Stat(); err == nil {
		if c.IsDir() {
			// read app info from chart
			a, err := h.readAppInfo(c)
			if err != nil {
				api.HandleError(resp, req, err)
				return
			}

			resp.WriteAsJson(a)
			return
		}
	} else {
		klog.Errorf("Failed to read dir appName=%s err=%v", app, err)
	}

	api.HandleNotFound(resp, req, fmt.Errorf("app [%s] not found", app))
}
