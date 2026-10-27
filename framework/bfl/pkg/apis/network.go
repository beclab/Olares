package apis

import (
	"net/http"

	"bytetrade.io/web3os/bfl/pkg/api/response"
	"github.com/emicklei/go-restful/v3"
)

func (*Base) NetworkUnavailable(req *restful.Request, resp *restful.Response) {
	resp.WriteHeaderAndEntity(http.StatusNotImplemented, response.Header{
		Code:    http.StatusNotImplemented,
		Message: "this network feature is not yet supported by the shared BFL",
	})
}
