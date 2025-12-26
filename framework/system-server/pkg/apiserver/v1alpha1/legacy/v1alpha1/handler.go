package legacy

import (
	"bufio"
	"errors"
	"fmt"
	"io"
	"io/ioutil"
	"net/http"
	"net/http/httputil"
	"reflect"

	"bytetrade.io/web3os/system-server/pkg/apiserver/v1alpha1/api"
	permission "bytetrade.io/web3os/system-server/pkg/permission/v1alpha1"
	prodiverregistry "bytetrade.io/web3os/system-server/pkg/providerregistry/v1alpha1"
	serviceproxy "bytetrade.io/web3os/system-server/pkg/serviceproxy/v1alpha1"

	"github.com/emicklei/go-restful/v3"
	"github.com/go-resty/resty/v2"
	"k8s.io/klog/v2"
)

type Handler struct {
	method string
	proxy  *serviceproxy.Proxy
}

func newHandler(method string, registry *prodiverregistry.Registry,
) *Handler {
	proxy := serviceproxy.NewProxy(registry)

	return &Handler{
		method: method,
		proxy:  proxy,
	}
}

func (h *Handler) do(req *restful.Request, resp *restful.Response) {
	klog.Info("proxy ", h.method, " /", req.PathParameter(serviceproxy.ParamSubPath))

	proxyRespIntf, err := h.proxy.ProxyLegacyAPI(req.Request.Context(), h.method, req, resp)
	if err != nil && isNil(proxyRespIntf) {
		klog.Info("proxy error: ", err)
		api.HandleError(resp, req, err)
		return
	}

	if err == nil && isNil(proxyRespIntf) {
		klog.Info("websocket proxy connected")
		return
	}

	switch proxyResp := proxyRespIntf.(type) {
	case *resty.Response:
		dump, e := httputil.DumpRequest(proxyResp.Request.RawRequest, true)
		if e != nil {
			klog.Error("dump request err: ", e)
		} else {
			klog.Info("proxy request: ", string(dump))
		}
		if proxyResp.RawResponse == nil {
			klog.Info("proxy error nil raw response: ", err)
			api.HandleError(resp, req, err)
			return
		}

		dump, err = httputil.DumpResponse(proxyResp.RawResponse, false)
		if err != nil {
			klog.Error("dump response err: ", err)
		} else {
			klog.Info("proxy response: ", string(dump))
		}

		for h, values := range proxyResp.Header() {
			for _, v := range values {
				resp.Header().Set(h, v)
			}
		}

		for _, c := range proxyResp.Cookies() {
			http.SetCookie(resp, c)
		}

		resp.WriteHeader(proxyResp.StatusCode())
		resp.Write(proxyResp.Body())

	case *serviceproxy.WsProxyResponse:
		if proxyResp.RawResponse == nil {
			klog.Info("ws proxy error: ", err)
			api.HandleError(resp, req, err)
			return
		}
		resp.WriteHeader(proxyResp.RawResponse.StatusCode)
		resp.Write(proxyResp.Body)
	}

}

func isNil(i interface{}) bool {
	return i == nil || reflect.ValueOf(i).IsNil()
}

func (h *Handler) doV2(req *restful.Request, resp *restful.Response) {
	klog.Info("proxy ", h.method, " /", req.PathParameter(serviceproxy.ParamSubPath))
	appKey := req.HeaderParameter("X-App-Key")
	if appKey == "" {
		api.HandleForbidden(resp, req, errors.New("empty X-App-Key"))
		return
	}

	signature := req.HeaderParameter("X-Auth-Signature")
	if len(signature) == 0 {
		api.HandleForbidden(resp, req, errors.New("invalid signature"))
		return
	}
	err := permission.ValidateAppKeyWithRequest(appKey, req)
	if err != nil {
		if errors.Is(err, prodiverregistry.ErrProviderNotFound) {
			api.HandleNotFound(resp, req, err)
			return
		}
		api.HandleForbidden(resp, req, fmt.Errorf("permission denied: err=%v", err))
		return
	}

	proxyRespIntf, err := h.proxy.ProxyLegacyAPIV2(req.Request.Context(), h.method, req, resp)
	if err != nil && errors.Is(err, prodiverregistry.ErrProviderNotFound) {
		api.HandleNotFound(resp, req, err)
		return
	}
	if err != nil && isNil(proxyRespIntf) {
		klog.Info("proxy error: ", err)
		api.HandleError(resp, req, err)
		return
	}

	if err == nil && isNil(proxyRespIntf) {
		klog.Info("websocket proxy connected")
		return
	}
	switch proxyResp := proxyRespIntf.(type) {
	case *resty.Response:
		dump, e := httputil.DumpRequest(proxyResp.Request.RawRequest, true)
		if e != nil {
			klog.Error("dump request err: ", e)
		} else {
			klog.Info("proxy request: ", string(dump))
		}
		if proxyResp.RawResponse == nil {
			klog.Info("proxy error nil raw response: ", err)
			api.HandleError(resp, req, err)
			return
		}
		dump, err = httputil.DumpResponse(proxyResp.RawResponse, false)
		if err != nil {
			klog.Error("dump response err: ", err)
		} else {
			klog.Info("proxy response: ", string(dump))
		}

		for h, values := range proxyResp.Header() {
			for _, v := range values {
				resp.Header().Set(h, v)
			}
		}
		for _, c := range proxyResp.Cookies() {
			http.SetCookie(resp, c)
		}
		if proxyResp.RawBody() != nil {
			defer proxyResp.RawBody().Close()
		}
		if isSSEOrNdJson(resp.Header()) {
			resp.WriteHeader(proxyResp.StatusCode())
			err = handleSSEOrNdJsonPROXY(resp, proxyResp.RawBody())
			if err != nil {
				klog.Infof("handleSSEOrNdJsonPROXY err=%v", err)
			}
		} else {
			resp.Header().Del("Content-Length")
			resp.WriteHeader(proxyResp.StatusCode())
			content, _ := ioutil.ReadAll(proxyResp.RawBody())
			resp.Write(content)
		}

	case *serviceproxy.WsProxyResponse:
		if proxyResp.RawResponse == nil {
			klog.Info("ws proxy error: ", err)
			api.HandleError(resp, req, err)
			return
		}
		resp.WriteHeader(proxyResp.RawResponse.StatusCode)
		resp.Write(proxyResp.Body)
	}

}

func isSSEOrNdJson(header http.Header) bool {
	return header.Get("Content-Type") == "text/event-stream" || header.Get("Content-Type") == "application/x-ndjson"
}

func handleSSEOrNdJsonPROXY(w *restful.Response, body io.Reader) error {
	reader := bufio.NewReader(body)
	for {
		line, err := reader.ReadBytes('\n')
		if err != nil {
			if err != io.EOF {
				klog.Errorf("Error reading SSE: %v", err)
			}
			return err
		}
		_, err = w.Write(line)
		if err != nil {
			klog.Errorf("Error writing SSE: %v", err)
			return err
		}
		w.Flush()
	}
}
