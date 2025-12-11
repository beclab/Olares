package v1alpha1

import (
	"net/http"

	"bytetrade.io/web3os/bfl/pkg/api/response"
	"bytetrade.io/web3os/bfl/pkg/apiserver/runtime"

	restfulspec "github.com/emicklei/go-restful-openapi/v2"
	"github.com/emicklei/go-restful/v3"
)

type CallbackHandler struct {
	BackupNew    func() error
	BackupFinish func() error
}

var (
	MODULE_TAGS = []string{"callbacks"}

	callbackHandlers []*CallbackHandler
)

var ModuleVersion = runtime.ModuleVersion{Name: "callback", Version: "v1alpha1"}

func AddToContainer(c *restful.Container) error {
	ws := runtime.NewWebService(ModuleVersion)
	handler := newHandler()

	ws.Route(ws.POST("/backup/new").
		To(handler.backupNew).
		Doc("Provide system backup phase-new to callback").
		Metadata(restfulspec.KeyOpenAPITags, MODULE_TAGS).
		Returns(http.StatusOK, "Success", response.Response{}))

	ws.Route(ws.POST("/backup/finish").
		To(handler.backupFinish).
		Doc("Provide system backup phase-success / failed / canceled to callback").
		Metadata(restfulspec.KeyOpenAPITags, MODULE_TAGS).
		Returns(http.StatusOK, "Success", response.Response{}))

	c.Add(ws)
	return nil
}

func AddBackupCallbackHandler(backupNew func() error, backupFinish func() error) {
	callbackHandlers = append(callbackHandlers, &CallbackHandler{
		BackupNew:    backupNew,
		BackupFinish: backupFinish,
	})
}
