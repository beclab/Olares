package serviceproxy

import (
	"compress/gzip"
	"context"
	"errors"
	"fmt"
	"io"
	"io/ioutil"
	"net"
	"net/http"
	"net/http/httputil"
	"net/url"
	"strings"
	"time"

	sysv1alpha1 "bytetrade.io/web3os/system-server/pkg/apis/sys/v1alpha1"
	apiv1alpha1 "bytetrade.io/web3os/system-server/pkg/apiserver/v1alpha1/api"
	"bytetrade.io/web3os/system-server/pkg/constants"
	prodiverregistry "bytetrade.io/web3os/system-server/pkg/providerregistry/v1alpha1"
	"bytetrade.io/web3os/system-server/pkg/utils"

	"github.com/emicklei/go-restful/v3"
	"github.com/go-resty/resty/v2"
	"github.com/gorilla/websocket"
	"k8s.io/klog/v2"
)

var (
	SUPPORTED_DATA_TYPE = []string{
		sysv1alpha1.Event,
		sysv1alpha1.Calendar,
		sysv1alpha1.Contact,
		sysv1alpha1.Key,
		sysv1alpha1.Token,
		sysv1alpha1.Message,
	}

	Group_WebSocket string = "websocket."

	hdrContentEncodingKey = http.CanonicalHeaderKey("Content-Encoding")
)

type Proxy struct {
	registry *prodiverregistry.Registry
}

// NewProxy constructs a new Proxy.
func NewProxy(registry *prodiverregistry.Registry) *Proxy {
	proxy := &Proxy{
		registry: registry,
	}

	return proxy
}

// DoRequest send request to provider.
func (p *Proxy) DoRequest(req *restful.Request, op string, proxyrequest *ProxyRequest) (ret map[string]interface{}, statusCode int, err error) {

	klog.Info("send request to provider: ", utils.PrettyJSON(proxyrequest))

	if !utils.ListContains(SUPPORTED_DATA_TYPE, proxyrequest.DataType) {
		klog.Warning("unsupported data type, ", proxyrequest.DataType)
	}

	provider, err := p.registry.GetProvider(req.Request.Context(),
		proxyrequest.DataType,
		proxyrequest.Group,
		proxyrequest.Version,
	)

	if err != nil {
		return nil, http.StatusInternalServerError, err
	}

	authtoken := req.Request.Header.Get(apiv1alpha1.AuthorizationTokenHeader)

	for _, api := range provider.Spec.OpApis {
		requiredOp := sysv1alpha1.DecodeOps(op)
		if api.Name == requiredOp.Op {
			var url string
			if strings.HasPrefix(provider.Spec.Endpoint, "http://") ||
				strings.HasPrefix(provider.Spec.Endpoint, "https://") {
				url = fmt.Sprintf("%s%s", provider.Spec.Endpoint, api.URI)
			} else {
				url = fmt.Sprintf("http://%s%s", provider.Spec.Endpoint, api.URI)
			}

			klog.Info("provider url: ", url)

			client := resty.New()

			resp, err := client.SetTimeout(2*time.Minute).R().
				SetHeader(restful.HEADER_ContentType, restful.MIME_JSON).
				SetHeader(apiv1alpha1.BackendTokenHeader, constants.Nonce).
				SetHeader(apiv1alpha1.AuthorizationTokenHeader, authtoken).
				SetHeader(constants.BflUserKey, constants.Owner).
				SetBody(proxyrequest).
				SetResult(&ret).
				Post(url)

			if err != nil {
				return nil, http.StatusInternalServerError, fmt.Errorf("invoke provider err: %s", err.Error())
			}

			if resp.StatusCode() >= 400 {
				return nil, resp.StatusCode(), fmt.Errorf("invoke provider err: code %d, %s", resp.StatusCode(), string(resp.Body()))
			}

			return ret, resp.StatusCode(), nil
		}
	}

	return nil, http.StatusNotFound, errors.New("provider not found")
}

func (p *Proxy) ProxyLegacyAPI(ctx context.Context,
	method string,
	req *restful.Request,
	resp *restful.Response,
) (interface{}, error) {
	klog.Info("send request to legacy api")
	klog.Infof("proxyLegacyAPI: header: %v", req.Request.Header)

	version := req.PathParameter(apiv1alpha1.ParamVersion)
	group := req.PathParameter(apiv1alpha1.ParamGroup)

	provider, err := p.registry.GetProvider(ctx,
		sysv1alpha1.LegacyAPI,
		group,
		version,
	)

	if err != nil {
		return nil, err
	}

	path := req.PathParameter(ParamSubPath)

	if len(provider.Spec.OpApis) > 0 {
		if func() bool {
			for _, op := range provider.Spec.OpApis {
				if strings.ToUpper(op.Name) == method &&
					fmt.Sprintf("/%s", path) == op.URI {
					return false
				}
			}

			// not found in provided apis
			return true
		}() {
			return nil, errors.New("unsupported api of provider")
		}
	}

	var providerURL string
	if strings.HasPrefix(provider.Spec.Endpoint, "http://") ||
		strings.HasPrefix(provider.Spec.Endpoint, "https://") {
		providerURL = fmt.Sprintf("%s/%s", provider.Spec.Endpoint, path)
	} else {
		providerURL = fmt.Sprintf("http://%s/%s", provider.Spec.Endpoint, path)
	}

	klog.Info("provider url: ", providerURL)

	switch {
	// websocket group api
	case method == "GET" && strings.HasPrefix(group, Group_WebSocket):
		wsURL, err := url.Parse(providerURL)
		if err != nil {
			return nil, err
		}

		wsProxy := NewWsProxy()
		wsProxy.Director = func(req *http.Request, header http.Header) {
			header.Add(apiv1alpha1.BackendTokenHeader, constants.Nonce)
			header.Add(constants.BflUserKey, constants.Owner)

			for _, auth := range req.Header[http.CanonicalHeaderKey("Authorization")] {
				header.Add("Authorization", auth)
			}

		}
		return wsProxy.doWs(req.Request, resp, wsURL)
	default:
		dump, err := httputil.DumpRequest(req.Request, true)
		if err != nil {
			klog.Error("dump request err: ", err)
		}
		klog.Info("orig request: ", string(dump))

		client := resty.New()
		bodyData, err := ioutil.ReadAll(req.Request.Body)
		if err != nil {
			return nil, err
		}

		proxyReq := client.SetTimeout(2*time.Second).R().
			SetQueryParamsFromValues(req.Request.URL.Query()).
			SetHeaderMultiValues(req.Request.Header).
			SetHeader(apiv1alpha1.BackendTokenHeader, constants.Nonce).
			SetHeader(constants.BflUserKey, constants.Owner).
			SetBody(bodyData)

		return proxyReq.Execute(method, providerURL)
	}
}

func (p *Proxy) ProxyLegacyAPIV2(ctx context.Context,
	method string,
	req *restful.Request,
	resp *restful.Response,
) (interface{}, error) {
	dataType := req.PathParameter(apiv1alpha1.ParamDataType)
	version := req.PathParameter(apiv1alpha1.ParamVersion)
	group := req.PathParameter(apiv1alpha1.ParamGroup)

	provider, err := p.registry.GetProvider(ctx,
		dataType,
		group,
		version,
	)

	if err != nil {
		return nil, err
	}

	path := req.PathParameter(ParamSubPath)

	var providerURL string
	if strings.HasPrefix(provider.Spec.Endpoint, "http://") ||
		strings.HasPrefix(provider.Spec.Endpoint, "https://") {
		providerURL = fmt.Sprintf("%s/%s", provider.Spec.Endpoint, path)
	} else {
		providerURL = fmt.Sprintf("http://%s/%s", provider.Spec.Endpoint, path)
	}

	klog.Info("provider url: ", providerURL)

	switch {
	// websocket group api
	case method == "GET" && strings.HasPrefix(group, Group_WebSocket):
		wsURL, err := url.Parse(providerURL)
		if err != nil {
			return nil, err
		}

		wsProxy := NewWsProxy()
		wsProxy.Director = func(req *http.Request, header http.Header) {
			header.Add(apiv1alpha1.BackendTokenHeader, constants.Nonce)
			header.Add(constants.BflUserKey, constants.Owner)

			for _, auth := range req.Header[http.CanonicalHeaderKey("Authorization")] {
				header.Add("Authorization", auth)
			}

		}
		return wsProxy.doWs(req.Request, resp, wsURL)
	default:
		dump, err := httputil.DumpRequest(req.Request, true)
		if err != nil {
			klog.Error("dump request err: ", err)
		}
		klog.Info("orig request: ", string(dump))

		client := resty.New()
		bodyData, err := ioutil.ReadAll(req.Request.Body)
		if err != nil {
			return nil, err
		}
		client.SetTransport(&http.Transport{
			DisableCompression: true,
		}).SetDoNotParseResponse(true)

		proxyReq := client.SetTimeout(3600*time.Second).R().
			SetQueryParamsFromValues(req.Request.URL.Query()).
			SetHeaderMultiValues(req.Request.Header).
			SetHeader(apiv1alpha1.BackendTokenHeader, constants.Nonce).
			SetHeader(constants.BflUserKey, constants.Owner).
			SetBody(bodyData)

		acceptEncoding := req.Request.Header.Get("Accept-Encoding")
		if strings.Contains(acceptEncoding, "br") {
			proxyReq.SetHeader("Accept-Encoding", "gzip")
		}

		return proxyReq.Execute(method, providerURL)
	}
}

// websocket proxy
var (
	// DefaultUpgrader specifies the parameters for upgrading an HTTP
	// connection to a WebSocket connection.
	DefaultUpgrader = &websocket.Upgrader{
		ReadBufferSize:  1024,
		WriteBufferSize: 1024,
	}

	// DefaultDialer is a dialer with all fields set to the default zero values.
	DefaultDialer = websocket.DefaultDialer
)

type WebsocketProxy struct {
	Director func(incoming *http.Request, out http.Header)
	Upgrader *websocket.Upgrader
	Dialer   *websocket.Dialer
}

type WsProxyResponse struct {
	Request     *resty.Request
	RawResponse *http.Response

	Body []byte
}

func NewWsProxy() *WebsocketProxy {
	return &WebsocketProxy{}
}

// ServeHTTP implements the http.Handler that proxies WebSocket connections.
func (w *WebsocketProxy) doWs(req *http.Request, resp *restful.Response, backendURL *url.URL) (*WsProxyResponse, error) {
	if backendURL == nil {
		return nil, errors.New(("websocketproxy: backend URL is nil"))
	}

	dialer := w.Dialer
	if w.Dialer == nil {
		dialer = DefaultDialer
	}

	// Pass headers from the incoming request to the dialer to forward them to
	// the final destinations.
	requestHeader := http.Header{}
	if origin := req.Header.Get("Origin"); origin != "" {
		requestHeader.Add("Origin", origin)
	}
	for _, prot := range req.Header[http.CanonicalHeaderKey("Sec-WebSocket-Protocol")] {
		requestHeader.Add("Sec-WebSocket-Protocol", prot)
	}
	for _, cookie := range req.Header[http.CanonicalHeaderKey("Cookie")] {
		requestHeader.Add("Cookie", cookie)
	}
	if req.Host != "" {
		requestHeader.Set("Host", req.Host)
	}

	// Pass X-Forwarded-For headers too, code below is a part of
	// httputil.ReverseProxy. See http://en.wikipedia.org/wiki/X-Forwarded-For
	// for more information
	// TODO: use RFC7239 http://tools.ietf.org/html/rfc7239
	if clientIP, _, err := net.SplitHostPort(req.RemoteAddr); err == nil {
		// If we aren't the first proxy retain prior
		// X-Forwarded-For information as a comma+space
		// separated list and fold multiple headers into one.
		if prior, ok := req.Header["X-Forwarded-For"]; ok {
			clientIP = strings.Join(prior, ", ") + ", " + clientIP
		}
		requestHeader.Set("X-Forwarded-For", clientIP)
	}

	// Set the originating protocol of the incoming HTTP request. The SSL might
	// be terminated on our site and because we doing proxy adding this would
	// be helpful for applications on the backend.
	requestHeader.Set("X-Forwarded-Proto", "http")
	if req.TLS != nil {
		requestHeader.Set("X-Forwarded-Proto", "https")
	}

	// Enable the director to copy any additional headers it desires for
	// forwarding to the remote server.
	if w.Director != nil {
		w.Director(req, requestHeader)
	}

	// Connect to the backend URL, also pass the headers we get from the requst
	// together with the Forwarded headers we prepared above.
	// TODO: support multiplexing on the same backend connection instead of
	// opening a new TCP connection time for each request. This should be
	// optional:
	// http://tools.ietf.org/html/draft-ietf-hybi-websocket-multiplexing-01
	wsURL := strings.Replace(backendURL.String(), "http://", "ws://", 1)
	connBackend, backendResp, err := dialer.Dial(wsURL, requestHeader)
	if err != nil {
		klog.Errorf("websocketproxy: couldn't dial to remote backend url %s", err)
		if backendResp != nil {
			return toResponse(req, backendResp)
		}
		return nil, errors.New(http.StatusText(http.StatusServiceUnavailable))
	}

	upgrader := w.Upgrader
	if w.Upgrader == nil {
		upgrader = DefaultUpgrader
	}

	// Only pass those headers to the upgrader.
	upgradeHeader := http.Header{}
	if hdr := backendResp.Header.Get("Sec-Websocket-Protocol"); hdr != "" {
		upgradeHeader.Set("Sec-Websocket-Protocol", hdr)
	}
	if hdr := backendResp.Header.Get("Set-Cookie"); hdr != "" {
		upgradeHeader.Set("Set-Cookie", hdr)
	}

	// Now upgrade the existing incoming request to a WebSocket connection.
	// Also pass the header that we gathered from the Dial handshake.

	connPub, err := upgrader.Upgrade(resp, req, upgradeHeader)
	if err != nil {
		connBackend.Close()
		return nil, fmt.Errorf("websocketproxy: couldn't upgrade %s", err)
	}

	go func() {
		defer connPub.Close()
		defer connBackend.Close()

		errClient := make(chan error, 1)
		errBackend := make(chan error, 1)
		replicateWebsocketConn := func(dst, src *websocket.Conn, errc chan error) {
			for {
				msgType, msg, err := src.ReadMessage()
				if err != nil {
					m := websocket.FormatCloseMessage(websocket.CloseNormalClosure, fmt.Sprintf("%v", err))
					if e, ok := err.(*websocket.CloseError); ok {
						if e.Code != websocket.CloseNoStatusReceived {
							m = websocket.FormatCloseMessage(e.Code, e.Text)
						}
					}
					errc <- err
					dst.WriteMessage(websocket.CloseMessage, m)
					break
				}
				err = dst.WriteMessage(msgType, msg)
				if err != nil {
					errc <- err
					break
				}
			}
		}

		go replicateWebsocketConn(connPub, connBackend, errClient)
		go replicateWebsocketConn(connBackend, connPub, errBackend)

		var message string
		select {
		case err = <-errClient:
			message = "websocketproxy: Error when copying from backend to client: %v"
		case err = <-errBackend:
			message = "websocketproxy: Error when copying from client to backend: %v"

		}
		if e, ok := err.(*websocket.CloseError); !ok || e.Code == websocket.CloseAbnormalClosure {
			klog.Errorf(message, err)
		}
	}()

	return nil, nil
}

func toResponse(req *http.Request, resp *http.Response) (*WsProxyResponse, error) {
	request := &resty.Request{RawRequest: req}
	response := &WsProxyResponse{Request: request, RawResponse: resp}

	defer closeq(resp.Body)
	body := resp.Body

	// GitHub #142 & #187
	if strings.EqualFold(resp.Header.Get(hdrContentEncodingKey), "gzip") && resp.ContentLength != 0 {
		if _, ok := body.(*gzip.Reader); !ok {
			body, err := gzip.NewReader(body)
			if err != nil {
				return response, err
			}
			defer closeq(body)
		}
	}

	var err error
	if response.Body, err = ioutil.ReadAll(body); err != nil {
		return response, err
	}

	return response, nil
}

func closeq(v interface{}) {
	if c, ok := v.(io.Closer); ok {
		silently(c.Close())
	}
}

func silently(_ ...interface{}) {}
