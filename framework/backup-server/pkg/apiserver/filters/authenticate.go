package filters

import (
	"github.com/emicklei/go-restful/v3"
)

func Authenticate(req *restful.Request, resp *restful.Response, chain *restful.FilterChain) {
	// TODO: authenticate

	chain.ProcessFilter(req, resp)
}
