package v1

import (
	"net/http"

	"bytetrade.io/web3os/bfl/pkg/api/response"
	"bytetrade.io/web3os/bfl/pkg/apiserver/runtime"

	"github.com/emicklei/go-restful/v3"
)

var ModuleVersion = runtime.ModuleVersion{Name: "backend", Version: "v1"}
var wizardModuleVersion = runtime.ModuleVersion{Name: "info", Version: "v1"}

func AddContainer(c *restful.Container) error {
	ws := runtime.NewWebService(ModuleVersion)
	ws.Consumes(restful.MIME_JSON)
	ws.Produces(restful.MIME_JSON)

	handler := New()

	ws.Route(ws.GET("/user-info").
		To(handler.handleUserInfo).
		Doc("User information.").
		Returns(http.StatusOK, "", response.Response{}))

	ws.Route(ws.GET("/terminus-info").
		To(handler.handleTerminusInfo).
		Doc("terminus information.").
		Returns(http.StatusOK, "", response.Response{}))

	ws.Route(ws.GET("/olares-info").
		To(handler.handleOlaresInfo).
		Doc("olares information.").
		Returns(http.StatusOK, "", response.Response{}))

	ws.Route(ws.GET("/re-download-cert").
		To(handler.handleReDownloadCert).
		Doc("Re-download ssl certificate").
		Returns(http.StatusOK, "", response.Response{}))

	ws.Route(ws.POST("/myapps").
		To(handler.myapps).
		Doc("List user's apps (Only for Provider) ").
		Returns(http.StatusOK, "", response.Response{}))

	ws.Route(ws.GET("/cluster").
		To(handler.getClusterMetric).
		Doc("get the cluster current metrics ( cpu, memory, disk ).").
		Returns(http.StatusOK, "", response.Response{}))

	ws.Route(ws.GET("/config-system").
		To(handler.HandleGetSysConfig).
		Doc("get user locale.").
		Returns(http.StatusOK, "", response.Response{}))

	c.Add(ws)

	wsWizard := runtime.NewWebService(wizardModuleVersion)
	wsWizard.Consumes(restful.MIME_JSON)
	wsWizard.Produces(restful.MIME_JSON)

	wsWizard.Route(wsWizard.GET("/olares-info").
		To(handler.handleOlaresInfo).
		Doc("olares information.").
		Returns(http.StatusOK, "", response.Response{}))

	c.Add(wsWizard)
	return nil
}
