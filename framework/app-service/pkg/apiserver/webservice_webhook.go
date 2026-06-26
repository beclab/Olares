package apiserver

import (
	"net/http"

	restfulspec "github.com/emicklei/go-restful-openapi/v2"
	"github.com/emicklei/go-restful/v3"
)

func registerContainerFilters(c *restful.Container, handler *Handler) {
	c.Filter(handler.createClientSet)
	c.Filter(handler.authenticate)
}

// addWebhookRoutesToContainer registers admission webhook routes only. These are
// registered before informer cache sync so the TLS listener can serve apiserver
// webhook calls as early as possible during startup.
func addWebhookRoutesToContainer(c *restful.Container, handler *Handler) error {
	ws := newWebService()

	ws.Route(ws.POST("/sandbox/inject").
		To(handler.sandboxInject).
		Doc("mutating webhook for sandbox sidecar injection ").
		Metadata(restfulspec.KeyOpenAPITags, MODULE_TAGS).
		Returns(http.StatusOK, "Success to inject", nil)).
		Consumes(restful.MIME_JSON)

	ws.Route(ws.POST("/appns/validate").
		To(handler.appNamespaceValidate).
		Doc("validating webhook for validate app install namespace").
		Metadata(restfulspec.KeyOpenAPITags, MODULE_TAGS).
		Returns(http.StatusOK, "App namespace validated success", nil)).
		Consumes(restful.MIME_JSON)

	ws.Route(ws.POST("/runasuser/inject").
		To(handler.handleRunAsUser).
		Doc("mutating webhook for inject runasuser 1000 for third party app's pod").
		Metadata(restfulspec.KeyOpenAPITags, MODULE_TAGS).
		Returns(http.StatusOK, "inject runasuser success", nil)).
		Consumes(restful.MIME_JSON)

	ws.Route(ws.POST("/workflow/inject").
		To(handler.cronWorkflowInject).
		Doc("mutating webhook for cron workflow").
		Metadata(restfulspec.KeyOpenAPITags, MODULE_TAGS).
		Returns(http.StatusOK, "cron workflow inject success", nil)).
		Consumes(restful.MIME_JSON)

	ws.Route(ws.POST("/workflow/validate").
		To(handler.argoResourcesValidate).
		Doc("validating webhook for argo workflow resources namespace").
		Metadata(restfulspec.KeyOpenAPITags, MODULE_TAGS).
		Returns(http.StatusOK, "argo workflow resources validate success", nil)).
		Consumes(restful.MIME_JSON)

	ws.Route(ws.POST("/gpulimit/inject").
		To(handler.gpuLimitInject).
		Doc("add resources limits for deployment/statefulset").
		Metadata(restfulspec.KeyOpenAPITags, MODULE_TAGS).
		Returns(http.StatusOK, "add limit success", nil)).
		Consumes(restful.MIME_JSON)

	ws.Route(ws.POST("/app-label/inject").
		To(handler.appLabelInject).
		Doc("add resources limits for deployment/statefulset").
		Metadata(restfulspec.KeyOpenAPITags, MODULE_TAGS).
		Returns(http.StatusOK, "add limit success", nil)).
		Consumes(restful.MIME_JSON)

	ws.Route(ws.POST("/macvlan-init/inject").
		To(handler.macvlanInitInject).
		Doc("mutating webhook to inject macvlan reply-via-eth0 init container").
		Metadata(restfulspec.KeyOpenAPITags, MODULE_TAGS).
		Returns(http.StatusOK, "Success to inject", nil)).
		Consumes(restful.MIME_JSON)

	ws.Route(ws.POST("/provider-registry/validate").
		To(handler.providerRegistryValidate).
		Doc("validating webhook for validate app install namespace").
		Metadata(restfulspec.KeyOpenAPITags, MODULE_TAGS).
		Returns(http.StatusOK, "provider registry validated success", nil)).
		Consumes(restful.MIME_JSON)

	ws.Route(ws.POST("/user/validate").
		To(handler.userValidate).
		Doc("validating webhook for validate user creation").
		Metadata(restfulspec.KeyOpenAPITags, MODULE_TAGS).
		Returns(http.StatusOK, "user validated success", nil)).
		Consumes(restful.MIME_JSON)

	ws.Route(ws.POST("/metrics/highload").
		To(handler.highload).
		Doc("Provide system resources high load event to callback").
		Metadata(restfulspec.KeyOpenAPITags, MODULE_TAGS).
		Returns(http.StatusOK, "Success", nil))

	ws.Route(ws.POST("/metrics/user/highload").
		To(handler.userHighLoad).
		Doc("Provide user resources high load event to callback").
		Metadata(restfulspec.KeyOpenAPITags, MODULE_TAGS).
		Returns(http.StatusOK, "Success", nil))

	ws.Route(ws.POST("/applicationmanager/validate").
		To(handler.applicationManagerValidate).
		Doc("validating webhook for validate user creation").
		Metadata(restfulspec.KeyOpenAPITags, MODULE_TAGS).
		Returns(http.StatusOK, "user validated success", nil)).
		Consumes(restful.MIME_JSON)

	c.Add(ws)
	return nil
}
