package authz

import (
	"context"
	"errors"
	"fmt"
	"log/slog"
	"net"
	"net/http"
	"strings"
	"sync/atomic"
	"time"

	corev3 "github.com/envoyproxy/go-control-plane/envoy/config/core/v3"
	authv3 "github.com/envoyproxy/go-control-plane/envoy/service/auth/v3"
	envoytypev3 "github.com/envoyproxy/go-control-plane/envoy/type/v3"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	rpcstatus "google.golang.org/genproto/googleapis/rpc/status"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/health"
	healthpb "google.golang.org/grpc/health/grpc_health_v1"
	"google.golang.org/grpc/reflection"
	corev1 "k8s.io/api/core/v1"
	ctrlclient "sigs.k8s.io/controller-runtime/pkg/client"

	"github.com/beclab/Olares/framework/app-service/pkg/cluster"
)

// ServerOptions configures the in-process PEP server (gRPC :9001, HTTP :9002)
// for app-gateway SecurityPolicy ext_authz backendRefs.
type ServerOptions struct {
	// Enabled toggles the entire runnable. When false the Start method
	// returns immediately so app-service Pods that do not host the PEP
	// pay no cost.
	Enabled bool

	// Address of the gRPC ext_authz listener (e.g. ":9001"). Empty means
	// keep the legacy default.
	GRPCAddr string
	// Address of the HTTP /healthz, /readyz, /metrics listener
	// (e.g. ":9002"). Empty means keep the legacy default.
	HTTPAddr string

	// AllowMode preserves the allow-all baseline when the
	// HostUser decider returns Pass. Set false to deny by default.
	AllowMode bool

	// HostUser controls the in-process host-user decider.
	HostUser HostUserConfig

	// Logger is the structured logger used for per-request audit lines.
	// nil falls back to slog.Default().
	Logger *slog.Logger

	// SnapshotFunc reads ClusterConfig (default: cluster.GetSnapshot).
	SnapshotFunc cluster.SnapshotFunc

	// K8sClient is a controller-runtime client used by WI-27 request-path
	// wiring to list Namespaces and derive known users. Nil keeps the current
	// DeriveViewer fallback behavior.
	K8sClient ctrlclient.Client
}

// DefaultServerOptions returns the defaults for the in-process PEP.
func DefaultServerOptions() ServerOptions {
	return ServerOptions{
		Enabled:   true,
		GRPCAddr:  ":9001",
		HTTPAddr:  ":9002",
		AllowMode: true,
		HostUser:  DefaultHostUserConfig(),
	}
}

// ParseSkipViewers expands a comma-separated viewer-allowlist (same semantics
// as --gateway-authz-skip-viewers on app-service).
func ParseSkipViewers(s string) []string {
	if strings.TrimSpace(s) == "" {
		return nil
	}
	parts := strings.Split(s, ",")
	out := make([]string, 0, len(parts))
	for _, p := range parts {
		if v := strings.TrimSpace(p); v != "" {
			out = append(out, strings.ToLower(v))
		}
	}
	return out
}

// Server bundles the gRPC ext_authz and HTTP probe listeners. It is meant
// to be added to a controller-runtime manager via mgr.Add(srv), which
// invokes Start(ctx) and triggers graceful drain on ctx cancellation.
type Server struct {
	opts      ServerOptions
	logger    *slog.Logger
	k8sClient ctrlclient.Client
}

// NewServer constructs a Server with the provided options. The struct is
// inert until Start is called.
func NewServer(opts ServerOptions) *Server {
	l := opts.Logger
	if l == nil {
		l = slog.Default()
	}
	if opts.GRPCAddr == "" {
		opts.GRPCAddr = ":9001"
	}
	if opts.HTTPAddr == "" {
		opts.HTTPAddr = ":9002"
	}
	return &Server{
		opts:      opts,
		logger:    l,
		k8sClient: opts.K8sClient,
	}
}

// NeedLeaderElection implements manager.LeaderElectionRunnable. Each
// app-service replica must serve ext_authz independently of leader status
// because EG will load-balance across all backing pods.
func (s *Server) NeedLeaderElection() bool { return false }

// Start blocks until ctx is cancelled, then performs a graceful drain
// (readyz flips to 503, gRPC GracefulStop with a 10-second budget).
func (s *Server) Start(ctx context.Context) error {
	if !s.opts.Enabled {
		s.logger.Info("authz server disabled; not starting")
		<-ctx.Done()
		return nil
	}

	grpcLis, err := net.Listen("tcp", s.opts.GRPCAddr)
	if err != nil {
		return fmt.Errorf("authz gRPC listen %s: %w", s.opts.GRPCAddr, err)
	}

	grpcSrv := grpc.NewServer()
	snapFn := s.opts.SnapshotFunc
	if snapFn == nil {
		snapFn = cluster.DefaultSnapshotFunc()
	}
	authv3.RegisterAuthorizationServer(grpcSrv, &authzHandler{
		allow:        s.opts.AllowMode,
		audit:        Auditor{Logger: s.logger},
		hostUser:     s.opts.HostUser,
		snapshotFunc: snapFn,
		k8sClient:    s.k8sClient,
	})
	healthSvc := health.NewServer()
	healthSvc.SetServingStatus("", healthpb.HealthCheckResponse_SERVING)
	healthSvc.SetServingStatus(authv3.Authorization_ServiceDesc.ServiceName, healthpb.HealthCheckResponse_SERVING)
	healthpb.RegisterHealthServer(grpcSrv, healthSvc)
	reflection.Register(grpcSrv)

	ready := &atomic.Bool{}
	ready.Store(true)
	httpSrv := newProbeServer(s.opts.HTTPAddr, ready)

	s.logger.Info("authz server starting",
		"grpc", s.opts.GRPCAddr,
		"http", s.opts.HTTPAddr,
		"allow_mode", s.opts.AllowMode,
		"host_user_check", s.opts.HostUser.Enabled,
		"skip_viewers", s.opts.HostUser.SkipPrefixes,
	)

	errCh := make(chan error, 2)
	go func() {
		if err := grpcSrv.Serve(grpcLis); err != nil && !errors.Is(err, grpc.ErrServerStopped) {
			errCh <- fmt.Errorf("authz gRPC serve: %w", err)
		}
	}()
	go func() {
		if err := httpSrv.ListenAndServe(); err != nil && !errors.Is(err, http.ErrServerClosed) {
			errCh <- fmt.Errorf("authz HTTP probe serve: %w", err)
		}
	}()

	select {
	case <-ctx.Done():
		s.logger.Info("authz server draining")
		ready.Store(false)
		shutdownCtx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
		defer cancel()
		_ = httpSrv.Shutdown(shutdownCtx)
		stopped := make(chan struct{})
		go func() { grpcSrv.GracefulStop(); close(stopped) }()
		select {
		case <-stopped:
		case <-shutdownCtx.Done():
			grpcSrv.Stop()
		}
		return nil
	case err := <-errCh:
		grpcSrv.Stop()
		_ = httpSrv.Close()
		return err
	}
}

// authzHandler is the gRPC ext_authz Check implementation. It wires
// the HostUser decider in front of the allow/deny baseline.
type authzHandler struct {
	authv3.UnimplementedAuthorizationServer
	allow        bool
	audit        Auditor
	hostUser     HostUserConfig
	snapshotFunc cluster.SnapshotFunc
	k8sClient    ctrlclient.Client
}

func (h *authzHandler) Check(ctx context.Context, req *authv3.CheckRequest) (*authv3.CheckResponse, error) {
	start := time.Now()
	httpReq := req.GetAttributes().GetRequest().GetHttp()
	host := httpReq.GetHost()
	path := httpReq.GetPath()
	method := httpReq.GetMethod()
	headers := httpReq.GetHeaders()
	rid := extractRequestID(headers)

	snap, _ := h.snapshotFunc(ctx)
	phase := "baseline"
	if snap.InClusterGatewayEnabled && IsSharedInclusterHost(host) {
		phase = "incluster"
		return h.checkInCluster(ctx, host, path, method, headers, rid, phase, start)
	}
	return h.checkHostUserBaseline(host, path, method, headers, rid, phase, start)
}

func (h *authzHandler) checkInCluster(ctx context.Context, host, path, method string, headers map[string]string, rid, phase string, start time.Time) (*authv3.CheckResponse, error) {
	known := h.loadKnownUsers(ctx)
	id := InClusterIdentity(host, headers, known)
	switch id.Action {
	case ActionDeny:
		return h.denyResponse(rid, host, method, path, phase, start, headers, id, "")
	case ActionAllow:
		headers = HeadersWithDerivedUser(headers, id)
	}

	hu := HostUser(host, headers, h.hostUser)
	switch hu.Action {
	case ActionDeny:
		return h.denyResponse(rid, host, method, path, phase, start, headers, hu, "")
	}

	sa := InClusterSharedAllow(host)
	if sa.Action == ActionAllow {
		return h.allowResponse(rid, host, method, path, phase, start, headers, &hu, "incluster_shared_allow")
	}
	return h.checkHostUserBaseline(host, path, method, headers, rid, phase, start)
}

func (h *authzHandler) checkHostUserBaseline(host, path, method string, headers map[string]string, rid, phase string, start time.Time) (*authv3.CheckResponse, error) {
	dec := HostUser(host, headers, h.hostUser)
	switch dec.Action {
	case ActionDeny:
		return h.denyResponse(rid, host, method, path, phase, start, headers, dec, "")
	case ActionAllow:
		return h.allowResponse(rid, host, method, path, phase, start, headers, &dec, "host_user")
	}
	if h.allow {
		return h.allowResponse(rid, host, method, path, phase, start, headers, nil, "allow_all")
	}
	elapsed := time.Since(start)
	h.audit.Deny(Event{
		RID: rid, Authority: host, Method: method, Path: path,
		Decision: "deny", Code: "DENY_MODE", Phase: phase,
		L5dPresent: l5dPresent(headers), LatencyMS: elapsed.Milliseconds(),
	})
	recordAuthzMetrics("deny", "DENY_MODE", elapsed)
	return &authv3.CheckResponse{
		Status: &rpcstatus.Status{Code: int32(codes.PermissionDenied), Message: "phase-a deny mode"},
		HttpResponse: &authv3.CheckResponse_DeniedResponse{
			DeniedResponse: &authv3.DeniedHttpResponse{
				Status: &envoytypev3.HttpStatus{Code: envoytypev3.StatusCode_Forbidden},
				Body:   "Forbidden by app-service authz (phase-a deny mode)",
			},
		},
	}, nil
}

// loadKnownUsers lists Namespaces and derives the Olares user set from two
// merged signals: (1) namespace names with the "user-space-" prefix, and (2)
// non-empty values of the bytetrade.io/ns-owner label. Returns nil when the
// handler has no k8sClient (e.g. legacy server_test wiring) or when the List
// call fails; nil keeps InClusterIdentity behavior equivalent to the legacy
// DeriveViewer-only path (the WI-27 app_user_fallback simply does not fire).
// Caller scope: invoked once per ext_authz Check on the in-cluster path; no
// process-wide cache (intentionally simple — NS count is O(users + system NS),
// in the tens to low hundreds for the LiteLLM pilot scale).
func (h *authzHandler) loadKnownUsers(ctx context.Context) map[string]struct{} {
	if h.k8sClient == nil {
		return nil
	}
	var nsList corev1.NamespaceList
	if err := h.k8sClient.List(ctx, &nsList); err != nil {
		h.audit.logger().Warn("authz_load_known_users_failed", "err", err.Error())
		return nil
	}
	out := make(map[string]struct{}, len(nsList.Items))
	for i := range nsList.Items {
		ns := &nsList.Items[i]
		if strings.HasPrefix(ns.Name, "user-space-") {
			if u := strings.TrimPrefix(ns.Name, "user-space-"); u != "" {
				out[strings.ToLower(u)] = struct{}{}
			}
		}
		if owner := strings.TrimSpace(ns.Labels[nsOwnerLabel]); owner != "" {
			out[strings.ToLower(owner)] = struct{}{}
		}
	}
	return out
}

func (h *authzHandler) denyResponse(rid, host, method, path, phase string, start time.Time, headers map[string]string, dec Decision, via string) (*authv3.CheckResponse, error) {
	elapsed := time.Since(start)
	code := dec.Code
	if code == "" {
		code = "-"
	}
	h.audit.Deny(Event{
		RID: rid, Authority: host, Method: method, Path: path,
		Decision: "deny", Code: code, Via: via, Phase: phase,
		ViewerHost: dec.Viewer, ViewerAuthenticated: dec.Username,
		L5dPresent: l5dPresent(headers), LatencyMS: elapsed.Milliseconds(),
	})
	recordAuthzMetrics("deny", code, elapsed)
	body := "Forbidden by app-service authz: " + dec.Code
	if dec.Message != "" {
		body += " — " + dec.Message
	}
	return &authv3.CheckResponse{
		Status: &rpcstatus.Status{Code: int32(codes.PermissionDenied), Message: dec.Code},
		HttpResponse: &authv3.CheckResponse_DeniedResponse{
			DeniedResponse: &authv3.DeniedHttpResponse{
				Status: &envoytypev3.HttpStatus{Code: envoytypev3.StatusCode_Forbidden},
				Body:   body,
			},
		},
	}, nil
}

func (h *authzHandler) allowResponse(rid, host, method, path, phase string, start time.Time, headers map[string]string, dec *Decision, via string) (*authv3.CheckResponse, error) {
	elapsed := time.Since(start)
	ev := Event{
		RID: rid, Authority: host, Method: method, Path: path,
		Decision: "allow", Via: via, Phase: phase,
		L5dPresent: l5dPresent(headers), LatencyMS: elapsed.Milliseconds(),
	}
	if dec != nil {
		ev.ViewerHost = dec.Viewer
		ev.ViewerAuthenticated = dec.Username
	}
	h.audit.Allow(ev)
	recordAuthzMetrics("allow", "-", elapsed)
	ok := &authv3.OkHttpResponse{}
	if v := headerValue(headers, "x-bfl-user"); v != "" {
		ok.Headers = []*corev3.HeaderValueOption{
			{Header: &corev3.HeaderValue{Key: "x-bfl-user", Value: v}},
		}
	}
	return &authv3.CheckResponse{
		Status:       &rpcstatus.Status{Code: int32(codes.OK)},
		HttpResponse: &authv3.CheckResponse_OkResponse{OkResponse: ok},
	}, nil
}

func extractRequestID(h map[string]string) string {
	if h == nil {
		return ""
	}
	for _, k := range []string{"x-request-id", "x-amzn-trace-id", "traceparent"} {
		if v, ok := h[k]; ok && v != "" {
			return v
		}
	}
	return ""
}

func newProbeServer(addr string, ready *atomic.Bool) *http.Server {
	mux := http.NewServeMux()
	mux.HandleFunc("/healthz", func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte("ok"))
	})
	mux.HandleFunc("/readyz", func(w http.ResponseWriter, _ *http.Request) {
		if ready.Load() {
			w.WriteHeader(http.StatusOK)
			_, _ = w.Write([]byte("ready"))
			return
		}
		w.WriteHeader(http.StatusServiceUnavailable)
		_, _ = w.Write([]byte("draining"))
	})
	mux.Handle("/metrics", promhttp.Handler())
	return &http.Server{
		Addr:              addr,
		Handler:           mux,
		ReadHeaderTimeout: 5 * time.Second,
	}
}
