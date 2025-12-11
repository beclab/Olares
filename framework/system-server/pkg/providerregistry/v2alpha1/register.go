package v2alpha1

import (
	"net/http"

	"bytetrade.io/web3os/system-server/pkg/apiserver/v1alpha1/api/response"
	"bytetrade.io/web3os/system-server/pkg/utils/apitools"
	restfulspec "github.com/emicklei/go-restful-openapi/v2"
	"github.com/emicklei/go-restful/v3"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/rest"
)

var (
	MODULE_TAGS  = []string{"provider-registry"}
	MODULE_ROUTE = "/provider/v2alpha1"
)

func AddProviderRegistryToContainer(
	c *restful.Container,
	requireAuth func(f restful.RouteFunction) restful.RouteFunction,
	kubeconfig *rest.Config,
) error {

	client := kubernetes.NewForConfigOrDie(kubeconfig)
	handler := &handler{BaseHandler: &apitools.BaseHandler{}, kubeClient: client}
	ws := newWebService()

	ws.Route(ws.POST("/register").
		To(requireAuth(handler.register)).
		Doc("register an app provider binding").
		Metadata(restfulspec.KeyOpenAPITags, MODULE_TAGS).
		Returns(http.StatusOK, "Success to register a invoker", &response.Response{}))

	ws.Route(ws.POST("/unregister").
		To(requireAuth(handler.unregister)).
		Doc("unregister an app provider binding").
		Metadata(restfulspec.KeyOpenAPITags, MODULE_TAGS).
		Returns(http.StatusOK, "Success to unregister a invoker", &response.Response{}))

	c.Add(ws)
	return nil
}

func newWebService() *restful.WebService {
	webservice := restful.WebService{}

	webservice.Path(MODULE_ROUTE).
		Produces(restful.MIME_JSON)

	return &webservice
}
