package apiserver

import (
	"context"
	"net/http"

	"bytetrade.io/web3os/system-server/pkg/apiserver/v1alpha1/api"
	permission "bytetrade.io/web3os/system-server/pkg/permission/v1alpha1"
	prodiverregistry "bytetrade.io/web3os/system-server/pkg/providerregistry/v1alpha1"

	restfulspec "github.com/emicklei/go-restful-openapi/v2"
	"github.com/emicklei/go-restful/v3"
	"k8s.io/client-go/rest"
)

var (
	MODULE_TAGS = []string{"service-proxy"}
)

func addServiceToContainer(ctx context.Context,
	c *restful.Container,
	kubeconfig *rest.Config,
	registry *prodiverregistry.Registry,
	ctrlSet *permission.PermissionControlSet) error {
	handler, err := newAPIHandler(ctx, kubeconfig, registry, ctrlSet)
	if err != nil {
		return err
	}

	ws := newWebService()

	ws.Filter(handler.authenticate)

	ws.Route(ws.GET("/{"+api.ParamDataType+"}/{"+api.ParamGroup+"}/{"+api.ParamVersion+"}/{"+api.ParamDataID+"}").
		To(handler.get).
		Doc("Get data").
		Metadata(restfulspec.KeyOpenAPITags, MODULE_TAGS).
		Param(ws.PathParameter(api.ParamDataType, "the data type")).
		Param(ws.PathParameter(api.ParamGroup, "the data group")).
		Param(ws.PathParameter(api.ParamVersion, "the data version")).
		Param(ws.PathParameter(api.ParamDataID, "the data id")).
		Param(ws.HeaderParameter(api.AccessTokenHeader, "Access token")).
		Returns(http.StatusOK, "Success to get a data with group and version", nil))

	ws.Route(ws.GET("/{"+api.ParamDataType+"}/{"+api.ParamGroup+"}/{"+api.ParamVersion+"}").
		To(handler.list).
		Doc("list data").
		Metadata(restfulspec.KeyOpenAPITags, MODULE_TAGS).
		Param(ws.PathParameter(api.ParamDataType, "the data type")).
		Param(ws.PathParameter(api.ParamGroup, "the data group")).
		Param(ws.PathParameter(api.ParamVersion, "the data version")).
		Param(ws.HeaderParameter(api.AccessTokenHeader, "Access token")).
		Returns(http.StatusOK, "Success to get the data list with group and version", nil))

	ws.Route(ws.POST("/{"+api.ParamDataType+"}/{"+api.ParamGroup+"}/{"+api.ParamVersion+"}").
		To(handler.create).
		Doc("create data").
		Metadata(restfulspec.KeyOpenAPITags, MODULE_TAGS).
		Param(ws.PathParameter(api.ParamDataType, "the data type")).
		Param(ws.PathParameter(api.ParamGroup, "the data group")).
		Param(ws.PathParameter(api.ParamVersion, "the data version")).
		Param(ws.HeaderParameter(api.AccessTokenHeader, "Access token")).
		Returns(http.StatusOK, "Success to create the data with group and version", nil))

	ws.Route(ws.POST("/{"+api.ParamDataType+"}/{"+api.ParamGroup+"}/{"+api.ParamVersion+"}/{"+api.ParamAction+"}").
		To(handler.action).
		Doc("data customize action").
		Metadata(restfulspec.KeyOpenAPITags, MODULE_TAGS).
		Param(ws.PathParameter(api.ParamDataType, "the data type")).
		Param(ws.PathParameter(api.ParamGroup, "the data group")).
		Param(ws.PathParameter(api.ParamVersion, "the data version")).
		Param(ws.PathParameter(api.ParamAction, "the data action")).
		Param(ws.HeaderParameter(api.AccessTokenHeader, "Access token")).
		Returns(http.StatusOK, "Success to create the data with group and version", nil))

	ws.Route(ws.PUT("/{"+api.ParamDataType+"}/{"+api.ParamGroup+"}/{"+api.ParamVersion+"}/{"+api.ParamDataID+"}").
		To(handler.update).
		Doc("Update data").
		Metadata(restfulspec.KeyOpenAPITags, MODULE_TAGS).
		Param(ws.PathParameter(api.ParamDataType, "the data type")).
		Param(ws.PathParameter(api.ParamGroup, "the data group")).
		Param(ws.PathParameter(api.ParamVersion, "the data version")).
		Param(ws.PathParameter(api.ParamDataID, "the data id")).
		Param(ws.HeaderParameter(api.AccessTokenHeader, "Access token")).
		Returns(http.StatusOK, "Success to update a data with group and version", nil))

	ws.Route(ws.DELETE("/{"+api.ParamDataType+"}/{"+api.ParamGroup+"}/{"+api.ParamVersion+"}/{"+api.ParamDataID+"}").
		To(handler.delete).
		Doc("Delete data").
		Metadata(restfulspec.KeyOpenAPITags, MODULE_TAGS).
		Param(ws.PathParameter(api.ParamDataType, "the data type")).
		Param(ws.PathParameter(api.ParamGroup, "the data group")).
		Param(ws.PathParameter(api.ParamVersion, "the data version")).
		Param(ws.PathParameter(api.ParamDataID, "the data id")).
		Param(ws.HeaderParameter(api.AccessTokenHeader, "Access token")).
		Returns(http.StatusOK, "Success to delete a data with group and version", nil))

	c.Add(ws)

	return nil
}
