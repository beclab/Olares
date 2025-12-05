package apiserver

import (
	"github.com/emicklei/go-restful/v3"
)

func newWebService() *restful.WebService {
	webservice := restful.WebService{}

	webservice.Path("/app-service/v1").
		Produces(restful.MIME_JSON)

	return &webservice
}
