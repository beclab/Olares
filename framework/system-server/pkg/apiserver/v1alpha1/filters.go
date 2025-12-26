package apiserver

import (
	"bytes"
	"errors"
	"fmt"
	"net/http"
	"runtime"
	"time"

	"bytetrade.io/web3os/system-server/pkg/apiserver/v1alpha1/api"
	"bytetrade.io/web3os/system-server/pkg/utils"

	"github.com/emicklei/go-restful/v3"
	"k8s.io/klog/v2"
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
	klog.Error(errors.New(buffer.String()), "panic error")

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
	logWithVerbose := klog.V(4)
	if resp.StatusCode() > http.StatusBadRequest {
		logWithVerbose = klog.V(0)
	}

	logWithVerbose.Infof("%s - \"%s %s %s\" %d %d %dms",
		utils.RemoteIP(req.Request),
		req.Request.Method,
		req.Request.URL,
		req.Request.Proto,
		resp.StatusCode(),
		resp.ContentLength(),
		time.Since(start)/time.Millisecond,
	)

}

// func (h *Handler) createClientSet(req *restful.Request, resp *restful.Response, chain *restful.FilterChain) {
// 	kubeconfig := newKubeConfigFromRequest(req, h.kubeHost)
// 	client, err := clientset.NewClientSet(kubeconfig)
// 	if err != nil {
// 		api.HandleError(resp, req, err)
// 		return
// 	}

// 	req.SetAttribute(constants.KubeSphereClientAttribute, client)
// 	chain.ProcessFilter(req, resp)
// }

func (h *Handler) authenticate(req *restful.Request, resp *restful.Response, chain *restful.FilterChain) {
	// Ignore uris, because do not need authentication
	trustPaths := []string{
		"/system-server/v1alpha1/apidocs.json",
	}

	needAuth := true
	for _, p := range trustPaths {
		if req.Request.URL.Path == p {
			needAuth = false
			break
		}
	}

	if needAuth {
		token := req.Request.Header.Get(api.AccessTokenHeader)
		if token == "" {
			api.HandleUnauthorized(resp, req, errors.New("no authentication token error"))
			return
		}

		// postpone permission verify
	}

	chain.ProcessFilter(req, resp)
}
