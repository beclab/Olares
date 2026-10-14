package proxy

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net"
	"net/http"
	"net/http/httputil"
	"net/url"
	"os"
	"strconv"
	"strings"
	"sync"
	"time"

	"bytetrade.io/web3os/bfl/pkg/apiserver/authn"
	v1alpha1client "bytetrade.io/web3os/bfl/pkg/client/clientset/v1alpha1"
	"bytetrade.io/web3os/bfl/pkg/constants"

	iamv1alpha2 "github.com/beclab/api/iam/v1alpha2"
	"k8s.io/apimachinery/pkg/types"
)

const (
	defaultMaxResponseBytes = int64(32 << 20)
	defaultMaxRequestBytes  = int64(10 << 20)
	defaultMaxStreams       = 8
)

type contextKey int

const (
	identityKey contextKey = iota
	adminKey
)

type roleResolver func(context.Context, string) (bool, error)

type Handler struct {
	next             http.Handler
	proxy            *httputil.ReverseProxy
	resolveAdmin     roleResolver
	maxRequestBytes  int64
	maxResponseBytes int64
	streams          *streamLimiter
}

func New(next http.Handler, kubeClient v1alpha1client.ClientInterface, targetHost string) (http.Handler, error) {
	target, err := parseTarget(targetHost)
	if err != nil {
		return nil, err
	}
	h := &Handler{
		next:             next,
		maxRequestBytes:  defaultMaxRequestBytes,
		maxResponseBytes: defaultMaxResponseBytes,
		streams:          newStreamLimiter(envInt("BFL_PROXY_MAX_STREAMS", defaultMaxStreams)),
		resolveAdmin: func(ctx context.Context, username string) (bool, error) {
			var user iamv1alpha2.User
			if err := kubeClient.CtrlClient().Get(ctx, types.NamespacedName{Name: username}, &user); err != nil {
				return false, fmt.Errorf("get proxy user role: %w", err)
			}
			role := user.Annotations[constants.UserAnnotationOwnerRole]
			return role == constants.RoleOwner || role == constants.RoleAdmin, nil
		},
	}
	h.proxy = &httputil.ReverseProxy{
		Director:       h.director(target),
		Transport:      proxyTransport(),
		ModifyResponse: h.modifyResponse,
		ErrorHandler:   h.proxyError,
	}
	return h, nil
}

func parseTarget(host string) (*url.URL, error) {
	host = strings.TrimRight(host, "/")
	if !strings.HasPrefix(host, "http://") && !strings.HasPrefix(host, "https://") {
		host = "http://" + host
	}
	target, err := url.Parse(host)
	if err != nil || target.Host == "" {
		return nil, fmt.Errorf("invalid kubesphere proxy target %q", host)
	}
	return target, nil
}

func proxyTransport() *http.Transport {
	return &http.Transport{
		Proxy:                 http.ProxyFromEnvironment,
		DialContext:           (&net.Dialer{Timeout: 10 * time.Second, KeepAlive: 30 * time.Second}).DialContext,
		ForceAttemptHTTP2:     true,
		MaxIdleConns:          100,
		IdleConnTimeout:       90 * time.Second,
		TLSHandshakeTimeout:   10 * time.Second,
		ResponseHeaderTimeout: 30 * time.Second,
		ExpectContinueTimeout: time.Second,
	}
}

func (h *Handler) ServeHTTP(w http.ResponseWriter, req *http.Request) {
	if !proxyPath(req.URL.Path) {
		h.next.ServeHTTP(w, req)
		return
	}

	identity, err := authn.Authenticate(req)
	if err != nil || identity.Token == "" {
		writeError(w, http.StatusUnauthorized, "authentication required")
		return
	}
	admin, err := h.resolveAdmin(req.Context(), identity.Username)
	if err != nil {
		writeError(w, http.StatusBadGateway, err.Error())
		return
	}
	if !admin && !allowedUserScope(req.URL.Path, identity.Username) {
		writeError(w, http.StatusForbidden, "requested resource is outside the user scope")
		return
	}

	if !admin && isMonitoringNamespacesPost(req) {
		if err := h.validateJSONBody(w, req); err != nil {
			return
		}
	}

	if !admin && req.Method == http.MethodGet && req.URL.Path == resourcesPodsPath {
		query := req.URL.Query()
		query.Set("limit", "-1")
		req.URL.RawQuery = query.Encode()
	}

	if isUpgrade(req) {
		if !h.streams.acquire(identity.Username) {
			writeError(w, http.StatusTooManyRequests, "too many concurrent streaming requests")
			return
		}
		defer h.streams.release(identity.Username)
	}

	ctx := context.WithValue(req.Context(), identityKey, identity)
	ctx = context.WithValue(ctx, adminKey, admin)
	h.proxy.ServeHTTP(w, req.WithContext(ctx))
}

func proxyPath(path string) bool {
	for _, prefix := range []string{"/kapis", "/api/v1", "/apis"} {
		if path == prefix || strings.HasPrefix(path, prefix+"/") {
			return true
		}
	}
	return false
}

func (h *Handler) director(target *url.URL) func(*http.Request) {
	return func(req *http.Request) {
		identity, _ := req.Context().Value(identityKey).(authn.Identity)
		req.URL.Scheme = target.Scheme
		req.URL.Host = target.Host
		req.Host = target.Host

		for _, header := range []string{
			"Authorization", constants.HeaderBflUserKey, constants.UserAuthorizationTokenKey,
			"Cookie", "Forwarded", "X-Forwarded-For", "X-Forwarded-Host", "X-Forwarded-Proto", "Referer",
		} {
			req.Header.Del(header)
		}
		req.Header.Set("Authorization", "Bearer "+identity.Token)
	}
}

func (h *Handler) validateJSONBody(w http.ResponseWriter, req *http.Request) error {
	if !strings.Contains(strings.ToLower(req.Header.Get("Content-Type")), "application/json") {
		return nil
	}
	body := http.MaxBytesReader(w, req.Body, h.maxRequestBytes)
	defer body.Close()
	decoder := json.NewDecoder(body)
	decoder.UseNumber()
	var value map[string]any
	if err := decoder.Decode(&value); err != nil && !errors.Is(err, io.EOF) {
		if strings.Contains(err.Error(), "request body too large") {
			writeError(w, http.StatusRequestEntityTooLarge, "payload too large")
		} else {
			writeError(w, http.StatusBadRequest, "invalid JSON body")
		}
		return err
	}
	if value == nil {
		value = map[string]any{}
	}
	var trailing any
	if err := decoder.Decode(&trailing); !errors.Is(err, io.EOF) {
		writeError(w, http.StatusBadRequest, "invalid JSON body")
		if err == nil {
			return errors.New("multiple JSON values")
		}
		return err
	}
	encoded, err := json.Marshal(value)
	if err != nil {
		writeError(w, http.StatusBadRequest, "invalid JSON body")
		return err
	}
	req.Body = ioNopCloser(encoded)
	req.ContentLength = int64(len(encoded))
	req.Header.Set("Content-Length", strconv.Itoa(len(encoded)))
	req.Header.Del("Transfer-Encoding")
	return nil
}

func (h *Handler) proxyError(w http.ResponseWriter, _ *http.Request, err error) {
	status := http.StatusBadGateway
	if errors.Is(err, context.DeadlineExceeded) || errors.Is(err, os.ErrDeadlineExceeded) {
		status = http.StatusGatewayTimeout
	}
	writeError(w, status, "kubesphere proxy failed")
}

func writeError(w http.ResponseWriter, status int, message string) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	_ = json.NewEncoder(w).Encode(map[string]string{"message": message})
}

func envInt(name string, fallback int) int {
	value, err := strconv.Atoi(os.Getenv(name))
	if err != nil || value <= 0 {
		return fallback
	}
	return value
}

func isUpgrade(req *http.Request) bool {
	if strings.TrimSpace(req.Header.Get("Upgrade")) == "" {
		return false
	}
	for _, token := range strings.Split(req.Header.Get("Connection"), ",") {
		if strings.EqualFold(strings.TrimSpace(token), "upgrade") {
			return true
		}
	}
	return false
}

func allowedUserScope(path, username string) bool {
	if user, ok := userMetricTarget(path); ok && user != username {
		return false
	}
	if namespace, ok := namespaceTarget(path); ok {
		return ownedNamespace(namespace, username) || strings.HasSuffix(namespace, "-shared")
	}
	return true
}

func userMetricTarget(path string) (string, bool) {
	const prefix = "/kapis/monitoring.kubesphere.io/v1alpha3/users/"
	if !strings.HasPrefix(path, prefix) {
		return "", false
	}
	user := strings.TrimPrefix(path, prefix)
	if user == "" || strings.Contains(user, "/") {
		return "", false
	}
	decoded, err := url.PathUnescape(user)
	return decoded, err == nil
}

func namespaceTarget(path string) (string, bool) {
	parts := strings.Split(strings.Trim(path, "/"), "/")
	for index := 0; index+1 < len(parts); index++ {
		if parts[index] != "namespaces" || parts[index+1] == "" {
			continue
		}
		decoded, err := url.PathUnescape(parts[index+1])
		return decoded, err == nil
	}
	return "", false
}

type streamLimiter struct {
	mu     sync.Mutex
	limit  int
	active map[string]int
}

func newStreamLimiter(limit int) *streamLimiter {
	return &streamLimiter{limit: limit, active: make(map[string]int)}
}

func (l *streamLimiter) acquire(username string) bool {
	l.mu.Lock()
	defer l.mu.Unlock()
	if l.active[username] >= l.limit {
		return false
	}
	l.active[username]++
	return true
}

func (l *streamLimiter) release(username string) {
	l.mu.Lock()
	defer l.mu.Unlock()
	if l.active[username] <= 1 {
		delete(l.active, username)
		return
	}
	l.active[username]--
}
