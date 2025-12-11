package v2alpha1

import (
	"net/http"

	"github.com/emicklei/go-restful/v3"
	"k8s.io/apiserver/pkg/authentication/authenticator"
)

func Auth(authenticator authenticator.Request) func(f restful.RouteFunction) restful.RouteFunction {
	return func(f restful.RouteFunction) restful.RouteFunction {
		return func(req *restful.Request, resp *restful.Response) {
			handlerFunc := func(rw http.ResponseWriter, r *http.Request) {
				f(req, resp)
			}

			handlerFunc = WithUserHeader(nil, handlerFunc)
			handlerFunc = WithAuthentication(authenticator, nil, handlerFunc)

			handlerFunc(resp, req.Request)
		}
	}
}
