package apiserver

import (
	"bytes"
	"fmt"
	"net/http"
	"runtime"
	"time"

	"bytetrade.io/web3os/bfl/internal/log"
	"bytetrade.io/web3os/bfl/pkg/api/response"
	apiRuntime "bytetrade.io/web3os/bfl/pkg/apiserver/runtime"
	"bytetrade.io/web3os/bfl/pkg/constants"
	"bytetrade.io/web3os/bfl/pkg/utils"

	"github.com/emicklei/go-restful/v3"
)

func logStackOnRecover(panicReason interface{}, w http.ResponseWriter) {
	var buffer bytes.Buffer
	buffer.WriteString(fmt.Sprintf("recover from panic situation: - %v\r\n", panicReason))
	for i := 2; ; i += 1 {
		_, file, line, ok := runtime.Caller(i)
		if !ok {
			break
		}
		buffer.WriteString(fmt.Sprintf("    %s:%d\r\n", file, line))
	}
	log.Error(buffer.String())
}

func logRequestAndResponse(req *restful.Request, resp *restful.Response, chain *restful.FilterChain) {
	start := time.Now()
	chain.ProcessFilter(req, resp)

	// Always log error response
	log.Infof("%s - \"%s %s %s\" %d %d %dms",
		utils.RemoteIp(req.Request),
		req.Request.Method,
		req.Request.URL,
		req.Request.Proto,
		resp.StatusCode(),
		resp.ContentLength(),
		time.Since(start)/time.Millisecond,
	)
}

func cors(req *restful.Request, resp *restful.Response, chain *restful.FilterChain) {
	resp.AddHeader("Access-Control-Allow-Origin", "*")

	resp.AddHeader("Content-Type", "application/json, application/x-www-form-urlencoded")
	resp.AddHeader("Access-Control-Allow-Methods", "POST, GET, OPTIONS, PUT, DELETE")
	resp.AddHeader("Access-Control-Allow-Headers", "Accept, Content-Type, Accept-Encoding, X-Authorization")

	if req.Request.Method == "OPTIONS" {
		resp.WriteHeader(http.StatusOK)
		resp.Write([]byte("ok"))
		return
	}

	chain.ProcessFilter(req, resp)
}

func authenticate(req *restful.Request, resp *restful.Response, chain *restful.FilterChain) {
	// Ignore uris, because do not need authentication
	needAuth, reqPath := true, req.Request.URL.Path
	for _, p := range constants.RequestURLWhiteList {
		if len(reqPath) >= len(p) && reqPath[:len(p)] == p {
			needAuth = false
			break
		}
	}

	if needAuth {
		log.Debugw("request headers", "requestURL", req.Request.URL, "headers", req.Request.Header)

		user := req.HeaderParameter(constants.HeaderBflUserKey)
		if user == "" {
			tokenStr := req.HeaderParameter(constants.UserAuthorizationTokenKey)
			if tokenStr == "" {
				response.HandleUnauthorized(resp, response.NewTokenValidationError("user not found"))
				return
			}

			claims, err := apiRuntime.ParseToken(tokenStr)
			if err != nil {
				response.HandleUnauthorized(resp, response.NewTokenValidationError("parse token", err))
				return
			}

			user = claims.Username
		}

		req.SetAttribute(constants.UserContextAttribute, user)
	}

	chain.ProcessFilter(req, resp)
}
