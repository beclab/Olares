package legacy

import (
	"net/http"

	"bytetrade.io/web3os/system-server/pkg/apiserver/v1alpha1/api"
	prodiverregistry "bytetrade.io/web3os/system-server/pkg/providerregistry/v1alpha1"
	serviceproxy "bytetrade.io/web3os/system-server/pkg/serviceproxy/v1alpha1"

	restfulspec "github.com/emicklei/go-restful-openapi/v2"
	"github.com/emicklei/go-restful/v3"
	apiruntime "github.com/go-openapi/runtime"
	"github.com/go-resty/resty/v2"
)

var (
	MODULE_TAGS      = []string{"service-proxy-leagcy"}
	ParamSubPathExpr = "{" + serviceproxy.ParamSubPath + ":*}"
	RoutePath        = "/{" + api.ParamGroup + "}/{" + api.ParamVersion + "}/" + ParamSubPathExpr
	RoutePathV2      = "/{" + api.ParamDataType + "}/{" + api.ParamGroup + "}/{" + api.ParamVersion + "}/" + ParamSubPathExpr
)

func AddLegacyAPIToContainer(c *restful.Container,
	registry *prodiverregistry.Registry,
) error {
	ws := newWebService()

	ws.Route(ws.GET(RoutePath).
		To(newHandler(resty.MethodGet, registry).do).
		Doc("Proxy get").
		Metadata(restfulspec.KeyOpenAPITags, MODULE_TAGS).
		Returns(http.StatusOK, "Success to proxy", nil))

	ws.Route(ws.POST(RoutePath).
		To(newHandler(resty.MethodPost, registry).do).
		Doc("Proxy post").
		Metadata(restfulspec.KeyOpenAPITags, MODULE_TAGS).
		Returns(http.StatusOK, "Success to proxy", nil))

	ws.Route(ws.PUT(RoutePath).
		To(newHandler(resty.MethodPut, registry).do).
		Doc("Proxy put").
		Metadata(restfulspec.KeyOpenAPITags, MODULE_TAGS).
		Returns(http.StatusOK, "Success to proxy", nil))

	ws.Route(ws.DELETE(RoutePath).
		To(newHandler(resty.MethodDelete, registry).do).
		Doc("Proxy delete").
		Metadata(restfulspec.KeyOpenAPITags, MODULE_TAGS).
		Returns(http.StatusOK, "Success to proxy", nil))

	ws.Route(ws.PATCH(RoutePath).
		To(newHandler(resty.MethodPatch, registry).do).
		Doc("Proxy patch").
		Metadata(restfulspec.KeyOpenAPITags, MODULE_TAGS).
		Returns(http.StatusOK, "Success to proxy", nil))

	ws.Route(ws.HEAD(RoutePath).
		To(newHandler(resty.MethodHead, registry).do).
		Doc("Proxy head").
		Metadata(restfulspec.KeyOpenAPITags, MODULE_TAGS).
		Returns(http.StatusOK, "Success to proxy", nil))

	ws.Route(ws.OPTIONS(RoutePath).
		To(newHandler(resty.MethodOptions, registry).do).
		Doc("Proxy options").
		Metadata(restfulspec.KeyOpenAPITags, MODULE_TAGS).
		Returns(http.StatusOK, "Success to proxy", nil))

	c.Add(ws)
	return nil
}

func AddLegacyAPIV2ToContainer(c *restful.Container,
	registry *prodiverregistry.Registry,
) error {
	ws := newWebServiceV2()

	ws.Route(ws.GET(RoutePathV2).
		To(newHandler(resty.MethodGet, registry).doV2).
		Doc("Proxy get").
		Metadata(restfulspec.KeyOpenAPITags, MODULE_TAGS).
		Returns(http.StatusOK, "Success to proxy", nil))

	ws.Route(ws.POST(RoutePathV2).
		To(newHandler(resty.MethodPost, registry).doV2).
		Doc("Proxy post").
		Metadata(restfulspec.KeyOpenAPITags, MODULE_TAGS).
		Returns(http.StatusOK, "Success to proxy", nil))

	ws.Route(ws.PUT(RoutePathV2).
		To(newHandler(resty.MethodPut, registry).doV2).
		Doc("Proxy put").
		Metadata(restfulspec.KeyOpenAPITags, MODULE_TAGS).
		Returns(http.StatusOK, "Success to proxy", nil))

	ws.Route(ws.DELETE(RoutePathV2).
		To(newHandler(resty.MethodDelete, registry).doV2).
		Doc("Proxy delete").
		Metadata(restfulspec.KeyOpenAPITags, MODULE_TAGS).
		Returns(http.StatusOK, "Success to proxy", nil))

	ws.Route(ws.PATCH(RoutePathV2).
		To(newHandler(resty.MethodPatch, registry).doV2).
		Doc("Proxy patch").
		Metadata(restfulspec.KeyOpenAPITags, MODULE_TAGS).
		Returns(http.StatusOK, "Success to proxy", nil))

	ws.Route(ws.HEAD(RoutePathV2).
		To(newHandler(resty.MethodHead, registry).doV2).
		Doc("Proxy head").
		Metadata(restfulspec.KeyOpenAPITags, MODULE_TAGS).
		Returns(http.StatusOK, "Success to proxy", nil))

	ws.Route(ws.OPTIONS(RoutePathV2).
		To(newHandler(resty.MethodOptions, registry).doV2).
		Doc("Proxy options").
		Metadata(restfulspec.KeyOpenAPITags, MODULE_TAGS).
		Returns(http.StatusOK, "Success to proxy", nil))

	c.Add(ws)
	return nil
}

func newWebService() *restful.WebService {
	webservice := restful.WebService{}

	webservice.Path(serviceproxy.LEAGCY_PATCH).
		Consumes(restful.MIME_JSON,
			restful.MIME_OCTET,
			restful.MIME_XML,
			restful.MIME_ZIP,
			apiruntime.HTMLMime,
			apiruntime.TextMime,
			apiruntime.MultipartFormMime,
			apiruntime.URLencodedFormMime).
		Produces(restful.MIME_JSON,
			restful.MIME_OCTET,
			restful.MIME_XML,
			restful.MIME_ZIP,
			apiruntime.HTMLMime,
			apiruntime.TextMime,
			apiruntime.MultipartFormMime,
			apiruntime.URLencodedFormMime)
	return &webservice
}

func newWebServiceV2() *restful.WebService {
	webservice := restful.WebService{}

	webservice.Path(serviceproxy.LEAGCY_PATCH_V2).
		Consumes(restful.MIME_JSON,
			restful.MIME_OCTET,
			restful.MIME_XML,
			restful.MIME_ZIP,
			apiruntime.HTMLMime,
			apiruntime.TextMime,
			apiruntime.MultipartFormMime,
			apiruntime.URLencodedFormMime).
		Produces(restful.MIME_JSON,
			restful.MIME_OCTET,
			restful.MIME_XML,
			restful.MIME_ZIP,
			apiruntime.HTMLMime,
			apiruntime.TextMime,
			apiruntime.MultipartFormMime,
			apiruntime.URLencodedFormMime)
	return &webservice
}
