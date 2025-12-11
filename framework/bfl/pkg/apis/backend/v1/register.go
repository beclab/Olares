package v1

import (
	"net/http"

	"bytetrade.io/web3os/bfl/pkg/api/response"
	iamV1alpha1 "bytetrade.io/web3os/bfl/pkg/apis/iam/v1alpha1"
	"bytetrade.io/web3os/bfl/pkg/apiserver/runtime"

	restfulspec "github.com/emicklei/go-restful-openapi/v2"
	"github.com/emicklei/go-restful/v3"
)

var ModuleVersion = runtime.ModuleVersion{Name: "backend", Version: "v1"}
var wizardModuleVersion = runtime.ModuleVersion{Name: "info", Version: "v1"}

var tags = []string{"backend"}

func AddContainer(c *restful.Container) error {
	ws := runtime.NewWebService(ModuleVersion)
	ws.Consumes(restful.MIME_JSON)
	ws.Produces(restful.MIME_JSON)

	handler := New()

	ws.Route(ws.POST("/verify-password").
		To(handler.handleVerifyUserPassword).
		Doc("Verify user password.").
		Reads(iamV1alpha1.UserPassword{}).
		Metadata(restfulspec.KeyOpenAPITags, tags).
		Returns(http.StatusOK, "", response.Response{}))

	ws.Route(ws.GET("/user-info").
		To(handler.handleUserInfo).
		Doc("User information.").
		Metadata(restfulspec.KeyOpenAPITags, tags).
		Returns(http.StatusOK, "", response.Response{}))

	ws.Route(ws.GET("/terminus-info").
		To(handler.handleTerminusInfo).
		Doc("terminus information.").
		Metadata(restfulspec.KeyOpenAPITags, tags).
		Returns(http.StatusOK, "", response.Response{}))

	ws.Route(ws.GET("/olares-info").
		To(handler.handleOlaresInfo).
		Doc("olares information.").
		Metadata(restfulspec.KeyOpenAPITags, tags).
		Returns(http.StatusOK, "", response.Response{}))

	ws.Route(ws.GET("/ip").
		To(handler.handleGetIPAddress).
		Doc("IP Address.").
		Param(ws.QueryParameter("master", "get master nodeIP only").DataType("string").Required(false)).
		Metadata(restfulspec.KeyOpenAPITags, tags).
		Returns(http.StatusOK, "", response.Response{}))

	ws.Route(ws.GET("/re-download-cert").
		To(handler.handleReDownloadCert).
		Doc("Re-download ssl certificate").
		Metadata(restfulspec.KeyOpenAPITags, tags).
		Returns(http.StatusOK, "", response.Response{}))

	ws.Route(ws.POST("/myapps").
		To(handler.myapps).
		Doc("List user's apps (Only for Provider) ").
		Metadata(restfulspec.KeyOpenAPITags, tags).
		Returns(http.StatusOK, "", response.Response{}))

	ws.Route(ws.GET("/cluster").
		To(handler.getClusterMetric).
		Doc("get the cluster current metrics ( cpu, memory, disk ).").
		Metadata(restfulspec.KeyOpenAPITags, tags).
		Returns(http.StatusOK, "", response.Response{}))

	ws.Route(ws.GET("/config-system").
		To(handler.HandleGetSysConfig).
		Doc("get user locale.").
		Metadata(restfulspec.KeyOpenAPITags, tags).
		Returns(http.StatusOK, "", response.Response{}))

	c.Add(ws)

	wsWizard := runtime.NewWebService(wizardModuleVersion)
	wsWizard.Consumes(restful.MIME_JSON)
	wsWizard.Produces(restful.MIME_JSON)

	wsWizard.Route(wsWizard.GET("/terminus-info").
		To(handler.handleTerminusInfo).
		Doc("terminus information.").
		Metadata(restfulspec.KeyOpenAPITags, tags).
		Returns(http.StatusOK, "", response.Response{}))

	wsWizard.Route(wsWizard.GET("/olares-info").
		To(handler.handleOlaresInfo).
		Doc("olares information.").
		Metadata(restfulspec.KeyOpenAPITags, tags).
		Returns(http.StatusOK, "", response.Response{}))

	c.Add(wsWizard)
	return nil
}
