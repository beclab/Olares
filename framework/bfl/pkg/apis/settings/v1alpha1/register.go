package v1alpha1

import (
	"net/http"

	"bytetrade.io/web3os/bfl/pkg/api/response"
	"bytetrade.io/web3os/bfl/pkg/apis"
	"bytetrade.io/web3os/bfl/pkg/apiserver/runtime"
	"bytetrade.io/web3os/bfl/pkg/app_service/v1"

	restfulspec "github.com/emicklei/go-restful-openapi/v2"
	"github.com/emicklei/go-restful/v3"
)

const (
	ParamServiceName  = "service"
	ParamAppName      = "app"
	ParamEntranceName = "entrance_name"
	ParamDataType     = "dataType"
	ParamGroup        = "group"
	ParamVersion      = "version"
)

var ModuleVersion = runtime.ModuleVersion{Name: "settings", Version: "v1alpha1"}

var tags = []string{"settings"}

func AddContainer(c *restful.Container) error {
	ws := runtime.NewWebService(ModuleVersion)
	ws.Consumes(restful.MIME_JSON)
	ws.Produces(restful.MIME_JSON)

	handler := New()

	ws.Route(ws.POST("/binding-zone").
		To(handler.handleBindingUserZone).
		Doc("Binding user zone.").
		Reads(PostTerminusName{}).
		Metadata(restfulspec.KeyOpenAPITags, tags).
		Returns(http.StatusOK, "", response.Response{}))

	// FIXME: only for testing, noqa
	ws.Route(ws.GET("/unbind-zone").
		To(handler.handleUnbindingUserZone).
		Doc("Unbinding user zone.").
		Metadata(restfulspec.KeyOpenAPITags, tags).
		Returns(http.StatusOK, "", response.Response{}))

	ws.Route(ws.POST("/activate").
		To(handler.handleActivate).
		Doc("Activate system.").
		Metadata(restfulspec.KeyOpenAPITags, tags).
		Returns(http.StatusOK, "", response.Response{}))

	ws.Route(ws.POST("/reverse-proxy").
		To(handler.handleChangeReverseProxyConfig).
		Doc("Change the current reverse proxy settings.").
		Reads(ReverseProxyConfig{}).
		Metadata(restfulspec.KeyOpenAPITags, tags).
		Returns(http.StatusOK, "", response.Response{}))

	ws.Route(ws.GET("/reverse-proxy").
		To(handler.handleGetReverseProxyConfig).
		Doc("Get the current reverse proxy settings.").
		Metadata(restfulspec.KeyOpenAPITags, tags).
		Returns(http.StatusOK, "", response.Response{}))

	ws.Route(ws.GET("/external-network").
		To(handler.handleGetExternalNetworkSwitch).
		Doc("Get external network switch status.").
		Metadata(restfulspec.KeyOpenAPITags, tags).
		Returns(http.StatusOK, "", response.Response{}))

	ws.Route(ws.POST("/external-network").
		To(handler.handleUpdateExternalNetworkSwitch).
		Doc("Enable/Disable external network access (owner only).").
		Reads(ExternalNetworkSwitchUpdateRequest{}).
		Metadata(restfulspec.KeyOpenAPITags, tags).
		Returns(http.StatusOK, "", response.Response{}))

	ws.Route(ws.GET("/launcher-acc-policy").
		To(handler.handleGetLauncherAccessPolicy).
		Doc("Get launcher access policy.").
		Metadata(restfulspec.KeyOpenAPITags, tags).
		Returns(http.StatusOK, "", response.Response{}))

	ws.Route(ws.POST("/launcher-acc-policy").
		To(handler.handleUpdateLauncherAccessPolicy).
		Doc("Get launcher access policy.").
		Metadata(restfulspec.KeyOpenAPITags, tags).
		Reads(LauncherAccessPolicy{}).
		Returns(http.StatusOK, "", response.Response{}))

	ws.Route(ws.GET("/launcher-public-domain-access-policy").
		To(handler.handleGetPublicDomainAccessPolicy).
		Doc("Get public domain access policy.").
		Metadata(restfulspec.KeyOpenAPITags, tags).
		Reads(PublicDomainAccessPolicy{}).
		Returns(http.StatusOK, "", response.Response{}))

	ws.Route(ws.POST("/launcher-public-domain-access-policy").
		To(handler.handleUpdatePublicDomainAccessPolicy).
		Doc("Update public domain access policy.").
		Metadata(restfulspec.KeyOpenAPITags, tags).
		Reads(PublicDomainAccessPolicy{}).
		Returns(http.StatusOK, "", response.Response{}))

	ws.Route(ws.POST("/config-system").
		To(handler.handleUpdateLocale).
		Doc("Update user locale.").
		Metadata(restfulspec.KeyOpenAPITags, tags).
		Reads(apis.PostLocale{}).
		Returns(http.StatusOK, "", response.Response{}))

	ws.Route(ws.GET("/config-system").
		To(handler.HandleGetSysConfig).
		Doc("get user locale.").
		Metadata(restfulspec.KeyOpenAPITags, tags).
		Returns(http.StatusOK, "", response.Response{}))

	ws.Route(ws.POST("/set-login-background").
		To(handler.handlerUpdateUserLoginBackground).
		Doc("Update user login background.").
		Metadata(restfulspec.KeyOpenAPITags, tags).
		Returns(http.StatusOK, "", response.Response{}))

	ws.Route(ws.POST("/set-avatar").
		To(handler.handlerUpdateUserAvatar).
		Doc("Update user avatar.").
		Metadata(restfulspec.KeyOpenAPITags, tags).
		Returns(http.StatusOK, "", response.Response{}))

	// app settings
	ws.Route(ws.POST("/applications/{"+ParamAppName+"}/setup/policy").
		To(handler.setupAppPolicy).
		Doc("Setup application access policy.").
		Param(ws.PathParameter(ParamAppName, "app name").DataType("string").Required(true)).
		Reads(app_service.ApplicationSettingsPolicy{}).
		Metadata(restfulspec.KeyOpenAPITags, tags).
		Returns(http.StatusOK, "", response.Response{Data: app_service.ApplicationsSettings{}}))

	ws.Route(ws.GET("/applications/{"+ParamAppName+"}/setup/policy").
		To(handler.getAppPolicy).
		Doc("Get application access policy.").
		Param(ws.PathParameter(ParamAppName, "app name").DataType("string").Required(true)).
		Metadata(restfulspec.KeyOpenAPITags, tags).
		Returns(http.StatusOK, "", response.Response{Data: app_service.ApplicationsSettings{}}))

	ws.Route(ws.POST("/applications/{"+ParamAppName+"}/{"+ParamEntranceName+"}/setup/policy").
		To(handler.setupAppEntrancePolicy).
		Doc("Setup application entrance policy").
		Param(ws.PathParameter(ParamAppName, "app name").DataType("string").Required(true)).
		Param(ws.PathParameter(ParamEntranceName, "entrance name").DataType("string").Required(true)).
		Reads(app_service.ApplicationSettingsDomain{}).
		Metadata(restfulspec.KeyOpenAPITags, tags).
		Returns(http.StatusOK, "", response.Response{Data: app_service.ApplicationsSettings{}}))

	// app set custom domain
	ws.Route(ws.POST("/applications/{"+ParamAppName+"}/{"+ParamEntranceName+"}/setup/domain").
		To(handler.setupAppCustomDomain).
		Doc("Setup application domain").
		Param(ws.PathParameter(ParamAppName, "app name").DataType("string").Required(true)).
		Param(ws.PathParameter(ParamEntranceName, "entrance name").DataType("string").Required(true)).
		Reads(app_service.ApplicationSettingsDomain{}).
		Metadata(restfulspec.KeyOpenAPITags, tags).
		Returns(http.StatusOK, "", response.Response{Data: app_service.ApplicationsSettings{}}))

	ws.Route(ws.GET("/applications/entrances/setup/domain").
		To(handler.listEntrancesWithCustomDomain).
		Doc("List application entrances with a custom third party domain set").
		Metadata(restfulspec.KeyOpenAPITags, tags).
		Returns(http.StatusOK, "", response.Response{Data: app_service.EntrancesWithCustomDomain{}}))

	ws.Route(ws.GET("/applications/{"+ParamAppName+"}/setup/domain").
		To(handler.getAppCustomDomain).
		Doc("Get application domain settings").
		Param(ws.PathParameter(ParamAppName, "app name").DataType("string").Required(true)).
		Metadata(restfulspec.KeyOpenAPITags, tags).
		Returns(http.StatusOK, "", response.Response{Data: app_service.ApplicationsSettings{}}))

	ws.Route(ws.GET("/applications/{"+ParamAppName+"}/{"+ParamEntranceName+"}/setup/domain/finish").
		To(handler.finishAppCustomDomainCnameTarget).
		Doc("Finish application domain cname target setting").
		Param(ws.PathParameter(ParamAppName, "app name").DataType("string").Required(true)).
		Param(ws.PathParameter(ParamEntranceName, "entrance name").DataType("string").Required(true)).
		Metadata(restfulspec.KeyOpenAPITags, tags).
		Returns(http.StatusOK, "", response.Response{Data: app_service.ApplicationsSettings{}}))

	ws.Route(ws.POST("/applications/{"+ParamAppName+"}/{"+ParamEntranceName+"}/setup/auth-level").
		To(handler.setupAppAuthorizationLevel).
		Doc("Setup application auth level").
		Param(ws.PathParameter(ParamAppName, "app name").DataType("string").Required(true)).
		Param(ws.PathParameter(ParamEntranceName, "entrance name").DataType("string").Required(true)).
		Metadata(restfulspec.KeyOpenAPITags, tags).
		Returns(http.StatusOK, "", response.Response{Data: app_service.ApplicationsSettings{}}))

	ws.Route(ws.GET("/applications/{"+ParamAppName+"}/entrances").
		To(handler.getAppEntrances).
		Doc("Get application entrances").
		Param(ws.PathParameter(ParamAppName, "app name").DataType("string").Required(true)).
		Metadata(restfulspec.KeyOpenAPITags, tags).
		Returns(http.StatusOK, "", response.Response{Data: app_service.Entrances{}}))

	ws.Route(ws.GET("/apps/permissions").
		To(handler.applicationPermissionList).
		Doc("Get application permission list").
		Metadata(restfulspec.KeyOpenAPITags, tags).
		Returns(http.StatusOK, "", nil))

	ws.Route(ws.GET("/apps/permissions/{"+ParamAppName+"}").
		To(handler.applicationPermission).
		Doc("Get application permission list").
		Param(ws.PathParameter(ParamAppName, "app name").DataType("string").Required(true)).
		Metadata(restfulspec.KeyOpenAPITags, tags).
		Returns(http.StatusOK, "", nil))

	ws.Route(ws.GET("/apps/provider-registry/{"+ParamAppName+"}").
		To(handler.getApplicationProviderList).
		Doc("Get application provider-registry list").
		Metadata(restfulspec.KeyOpenAPITags, tags).
		Returns(http.StatusOK, "", nil))

	ws.Route(ws.GET("/apps/{"+ParamAppName+"}/subject").
		To(handler.getApplicationSubjectList).
		Doc("Get application subject list").
		Metadata(restfulspec.KeyOpenAPITags, tags).
		Returns(http.StatusOK, "", nil))

	ws.Route(ws.GET("/apps/provider-registry/{"+ParamDataType+"}/{"+ParamGroup+"}/{"+ParamVersion+"}").
		To(handler.getProviderRegistry).
		Doc("Get an provider registry").
		Param(ws.PathParameter(ParamDataType, "dataType").DataType("string").Required(true)).
		Param(ws.PathParameter(ParamGroup, "group").DataType("string").Required(true)).
		Param(ws.PathParameter(ParamVersion, "version").DataType("string").Required(true)).
		Metadata(restfulspec.KeyOpenAPITags, tags).
		Returns(http.StatusOK, "", nil))

	// headscale
	ws.Route(ws.GET("/headscale/ssh/acl").
		To(handler.handleGetHeadscaleSshAcl).
		Doc("get headscale ssh acl").
		Metadata(restfulspec.KeyOpenAPITags, []string{"headscale"}).
		Returns(http.StatusOK, "", &response.Response{}))

	ws.Route(ws.POST("/headscale/enable/ssh").
		To(handler.handleEnableHeadscaleSshAcl).
		Doc("enable headscale ssh acl").
		Metadata(restfulspec.KeyOpenAPITags, []string{"headscale"}).
		Returns(http.StatusOK, "", &response.Response{}))

	ws.Route(ws.POST("/headscale/disable/ssh").
		To(handler.handleDisableHeadscaleSshAcl).
		Doc("disable headscale ssh acl").
		Metadata(restfulspec.KeyOpenAPITags, []string{"headscale"}).
		Returns(http.StatusOK, "", &response.Response{}))

	ws.Route(ws.GET("/headscale/{"+ParamAppName+"}/acl").
		To(handler.handleGetHeadscaleAppAcl).
		Doc("get app's headscale acl").
		Metadata(restfulspec.KeyOpenAPITags, []string{"headscale"}).
		Param(ws.PathParameter(ParamAppName, "app name").DataType("string").Required(true)).
		Returns(http.StatusOK, "", &response.Response{}))

	ws.Route(ws.POST("/headscale/{"+ParamAppName+"}/acl").
		To(handler.handleUpdateHeadscaleAppAcl).
		Doc("set app's headscale acl").
		Metadata(restfulspec.KeyOpenAPITags, []string{"headscale"}).
		Param(ws.PathParameter(ParamAppName, "app name").DataType("string").Required(true)).
		Returns(http.StatusOK, "", &response.Response{}))

	ws.Route(ws.GET("/headscale/acls").
		To(handler.handleHeadscaleACLList).
		Doc("get app's headscale acl list").
		Metadata(restfulspec.KeyOpenAPITags, []string{"headscale"}).
		Returns(http.StatusOK, "", &response.Response{}))

	ws.Route(ws.POST("/tailscale/enable/subroutes").
		To(handler.handleEnableTailScaleSubnet).
		Doc("enable tailscale subroutes").
		Returns(http.StatusOK, "", &response.Response{}))

	ws.Route(ws.POST("/tailscale/disable/subroutes").
		To(handler.handleDisableTailScaleSubnet).
		Doc("enable tailscale subroutes").
		Returns(http.StatusOK, "", &response.Response{}))

	ws.Route(ws.GET("/tailscale/subroutes").
		To(handler.handleGetTailScaleSubnet).
		Doc("get tailscale subroutes").
		Returns(http.StatusOK, "", &response.Response{}))

	c.Add(ws)
	return nil
}
