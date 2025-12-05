package apiserver

import (
	"bytes"
	"errors"
	"fmt"
	"net/http"
	"runtime"
	"time"

	"bytetrade.io/web3os/app-service/pkg/apiserver/api"
	"bytetrade.io/web3os/app-service/pkg/client/clientset"
	"bytetrade.io/web3os/app-service/pkg/constants"
	"bytetrade.io/web3os/app-service/pkg/utils"

	"github.com/emicklei/go-restful/v3"
	ctrl "sigs.k8s.io/controller-runtime"
)

func logStackOnRecover(panicReason interface{}, w http.ResponseWriter) {
	var buffer bytes.Buffer
	buffer.WriteString(fmt.Sprintf("recover from panic situation: - %v\r\n", panicReason))
	for i := 2; ; i++ {
		_, file, line, ok := runtime.Caller(i)
		if !ok {
			break
		}
		buffer.WriteString(fmt.Sprintf("    %s:%d\r\n", file, line))
	}
	ctrl.Log.Error(errors.New(buffer.String()), "panic error")

	headers := http.Header{}
	if ct := w.Header().Get("Content-Type"); len(ct) > 0 {
		headers.Set("Accept", ct)
	}

	w.WriteHeader(http.StatusInternalServerError)
	w.Write([]byte("Internal server error"))
}

func logRequestAndResponse(req *restful.Request, resp *restful.Response, chain *restful.FilterChain) {
	start := time.Now()
	chain.ProcessFilter(req, resp)

	// Always log error response
	if resp.StatusCode() != http.StatusOK {
		ctrl.Log.Info("request",
			"IP",
			utils.RemoteIP(req.Request),
			"method",
			req.Request.Method,
			"URL",
			req.Request.URL,
			"proto",
			req.Request.Proto,
			"code",
			resp.StatusCode(),
			"length",
			resp.ContentLength(),
			"timestamp",
			time.Since(start)/time.Millisecond,
		)
	}

}

func (h *Handler) createClientSet(req *restful.Request, resp *restful.Response, chain *restful.FilterChain) {
	client, err := clientset.New(h.kubeConfig)
	if err != nil {
		api.HandleError(resp, req, err)
		return
	}

	req.SetAttribute(constants.KubeSphereClientAttribute, client)
	chain.ProcessFilter(req, resp)
}

func (h *Handler) authenticate(req *restful.Request, resp *restful.Response, chain *restful.FilterChain) {
	// Ignore uris, because do not need authentication
	trustPaths := []string{
		"/app-service/v1/apidocs.json",
		"/app-service/v1/sandbox/inject",
		"/app-service/v1/appns/validate",
		"/app-service/v1/gpulimit/inject",
		"/app-service/v1/backup/new",
		"/app-service/v1/backup/finish",
		"/app-service/v1/metrics/highload",
		"/app-service/v1/metrics/user/highload",
		"/app-service/v1/user-apps/",
		"/app-service/v1/apidocs.json",
		"/app-service/v1/recommenddev/",
		"/app-service/v1/provider-registry/validate",
		"/app-service/v1/pods/kubelet/eviction",
		"/app-service/v1/workflow/inject",
		"/app-service/v1/runasuser/inject",
		"/app-service/v1/terminus/version",
		"/app-service/v1/app-label/inject",
		"/app-service/v1/apps/image-info",
		"/app-service/v1/all/apps",
		"/app-service/v1/apps/oamvalues",
		"/app-service/v1/users/admin/username",
		"/app-service/v1/user/validate",
		"/app-service/v1/applicationmanager/inject",
		"/app-service/v1/applicationmanager/validate",
		"/app-service/v1/users/admins",
		"/app-service/v1/middlewares/status",
		"/app-service/v1/workflow/validate",
		"/app-service/v1/all/appmanagers",
		"/app-service/v1/cluster/node_info",
	}

	needAuth := true
	func() {
		for _, p := range trustPaths {
			switch {
			case req.Request.URL.Path == p:
				needAuth = false
				return
			case p[len(p)-1] == '/':
				if len(req.Request.URL.Path) > len(p) && req.Request.URL.Path[:len(p)] == p {
					needAuth = false
					return
				}
			}
		}
	}()

	if needAuth {
		username := req.Request.Header.Get(constants.BflUserKey)
		if username == "" {
			api.HandleUnauthorized(resp, req, errors.New("no authentication info error"))
			return
		}

		req.SetAttribute(constants.UserContextAttribute, username)
	}

	chain.ProcessFilter(req, resp)
}
