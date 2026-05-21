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

	authv3 "github.com/envoyproxy/go-control-plane/envoy/service/auth/v3"
	envoytypev3 "github.com/envoyproxy/go-control-plane/envoy/type/v3"
	rpcstatus "google.golang.org/genproto/googleapis/rpc/status"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/health"
	healthpb "google.golang.org/grpc/health/grpc_health_v1"
	"google.golang.org/grpc/reflection"
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
	opts   ServerOptions
	logger *slog.Logger
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
	return &Server{opts: opts, logger: l}
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
	authv3.RegisterAuthorizationServer(grpcSrv, &authzHandler{
		allow:    s.opts.AllowMode,
		logger:   s.logger,
		hostUser: s.opts.HostUser,
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
	allow    bool
	logger   *slog.Logger
	hostUser HostUserConfig
}

func (h *authzHandler) Check(_ context.Context, req *authv3.CheckRequest) (*authv3.CheckResponse, error) {
	httpReq := req.GetAttributes().GetRequest().GetHttp()
	host := httpReq.GetHost()
	path := httpReq.GetPath()
	method := httpReq.GetMethod()
	headers := httpReq.GetHeaders()
	rid := extractRequestID(headers)

	dec := HostUser(host, headers, h.hostUser)
	switch dec.Action {
	case ActionDeny:
		h.logger.Warn("authz deny",
			"rid", rid, "authority", host, "method", method, "path", path,
			"viewer_host", dec.Viewer, "viewer_authenticated", dec.Username,
			"code", dec.Code, "decision", "deny")
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
	case ActionAllow:
		h.logger.Info("authz allow",
			"rid", rid, "authority", host, "method", method, "path", path,
			"viewer_host", dec.Viewer, "viewer_authenticated", dec.Username,
			"decision", "allow")
		return &authv3.CheckResponse{
			Status: &rpcstatus.Status{Code: int32(codes.OK)},
			HttpResponse: &authv3.CheckResponse_OkResponse{
				OkResponse: &authv3.OkHttpResponse{},
			},
		}, nil
	}

	if h.allow {
		h.logger.Info("authz allow",
			"rid", rid, "authority", host, "method", method, "path", path,
			"decision", "allow_all")
		return &authv3.CheckResponse{
			Status: &rpcstatus.Status{Code: int32(codes.OK)},
			HttpResponse: &authv3.CheckResponse_OkResponse{
				OkResponse: &authv3.OkHttpResponse{},
			},
		}, nil
	}
	h.logger.Warn("authz deny",
		"rid", rid, "authority", host, "method", method, "path", path,
		"decision", "deny")
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
	mux.HandleFunc("/metrics", func(w http.ResponseWriter, _ *http.Request) {
		w.Header().Set("Content-Type", "text/plain; charset=utf-8")
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte("# app-service in-process PEP metrics\n"))
		_, _ = w.Write([]byte("app_service_ext_authz_phase 1\n"))
	})
	return &http.Server{
		Addr:              addr,
		Handler:           mux,
		ReadHeaderTimeout: 5 * time.Second,
	}
}
