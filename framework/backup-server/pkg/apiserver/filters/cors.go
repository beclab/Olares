package filters

import (
	"net/http"

	"github.com/emicklei/go-restful/v3"
)

func Cors(req *restful.Request, resp *restful.Response, chain *restful.FilterChain) {
	resp.AddHeader("Access-Control-Allow-Origin", "*")

	resp.AddHeader("Content-Type", "application/json, application/x-www-form-urlencoded")
	resp.AddHeader("Access-Control-Allow-Methods", "POST, GET, OPTIONS, PUT, PATCH, DELETE")
	resp.AddHeader("Access-Control-Allow-Headers", "Accept, Content-Type, Accept-Encoding, Authorization, X-Authorization")

	if req.Request.Method == "OPTIONS" {
		resp.WriteHeader(http.StatusOK)
		resp.Write([]byte("ok"))
		return
	}

	chain.ProcessFilter(req, resp)
}
