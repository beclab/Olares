package v1alpha1

import (
	"net/http"

	"bytetrade.io/web3os/bfl/pkg/api/response"
	"bytetrade.io/web3os/bfl/pkg/apiserver/runtime"

	iamV1alpha2 "github.com/beclab/api/iam/v1alpha2"
	restfulspec "github.com/emicklei/go-restful-openapi/v2"
	"github.com/emicklei/go-restful/v3"
	aruntime "k8s.io/apimachinery/pkg/runtime"
	utilruntime "k8s.io/apimachinery/pkg/util/runtime"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

var (
	scheme = aruntime.NewScheme()
)

func init() {
	utilruntime.Must(iamV1alpha2.AddToScheme(scheme))
}

var ModuleVersion = runtime.ModuleVersion{Name: "iam", Version: "v1alpha1"}

var (
	iamTags = []string{"iam"}

	userTags = []string{"users"}
)

func AddToContainer(c *restful.Container) error {
	ws := runtime.NewWebService(ModuleVersion)
	config, err := ctrl.GetConfig()
	if err != nil {
		return err
	}
	ctrlClient, err := client.New(config, client.Options{Scheme: scheme})
	if err != nil {
		return err
	}

	handler := New(ctrlClient)

	ws.Route(ws.GET("/users").
		To(handler.handleListUsers).
		Doc("List users.").
		Metadata(restfulspec.KeyOpenAPITags, userTags).
		Produces(restful.MIME_JSON).
		Returns(http.StatusOK, "", response.Response{}))

	ws.Route(ws.GET("/users/{user}/login-records").
		To(handler.handleListUserLoginRecords).
		Doc("List user login records.").
		Metadata(restfulspec.KeyOpenAPITags, userTags).
		Param(ws.PathParameter("user", "user name").DataType("string").Required(true)).
		Produces(restful.MIME_JSON).
		Returns(http.StatusOK, "", response.Response{}))

	ws.Route(ws.PUT("/users/{user}/password").
		To(handler.handleResetUserPassword).
		Doc("Reset user password.").
		Metadata(restfulspec.KeyOpenAPITags, userTags).
		Param(ws.PathParameter("user", "user name").DataType("string").Required(true)).
		Reads(PasswordReset{}).
		Produces(restful.MIME_JSON).
		Returns(http.StatusOK, "Reset password", response.Response{}))

	ws.Route(ws.GET("/users/{user}/metrics").
		To(handler.handleGetUserMetrics).
		Doc("get user's metrics").
		Metadata(restfulspec.KeyOpenAPITags, userTags).
		Param(ws.PathParameter("user", "user name").DataType("string").Required(true)).
		Produces(restful.MIME_JSON).
		Returns(http.StatusOK, "get user's metrics", nil))

	ws.Route(ws.GET("/roles").
		To(handler.handleListUserRoles).
		Doc("List user roles.").
		Metadata(restfulspec.KeyOpenAPITags, userTags).
		Produces(restful.MIME_JSON).
		Returns(http.StatusOK, "", response.Response{}))

	c.Add(ws)
	return nil
}
