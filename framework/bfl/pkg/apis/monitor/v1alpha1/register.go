package v1alpha1

import (
	"net/http"

	"bytetrade.io/web3os/bfl/pkg/api/response"
	"bytetrade.io/web3os/bfl/pkg/apiserver/runtime"

	restfulspec "github.com/emicklei/go-restful-openapi/v2"
	"github.com/emicklei/go-restful/v3"
)

var ModuleVersion = runtime.ModuleVersion{Name: "monitor", Version: "v1alpha1"}

var tags = []string{"monitor"}

func AddContainer(c *restful.Container) error {
	ws := runtime.NewWebService(ModuleVersion)
	ws.Consumes(restful.MIME_JSON)
	ws.Produces(restful.MIME_JSON)

	handler := newHandler()

	ws.Route(ws.GET("/cluster").
		To(handler.GetClusterMetric).
		Doc("get the cluster current metrics ( cpu, memory, disk ).").
		Metadata(restfulspec.KeyOpenAPITags, tags).
		Returns(http.StatusOK, "", response.Response{}))

	c.Add(ws)
	return nil
}
