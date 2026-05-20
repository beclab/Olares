// Command app-service-ext-authz is the Envoy ext_authz endpoint referenced by
// the SecurityPolicy that Envoy Gateway applies to the app-gateway Gateway.
// Phase A always returns OK ("allow_all") but is wired with failOpen: false so
// that an outage of this adapter denies all ingress traffic to the data plane
// (F-5: fail-closed).
package main

import (
	"context"
	"errors"
	"flag"
	"log/slog"
	"net"
	"net/http"
	"os"
	"os/signal"
	"strings"
	"sync/atomic"
	"syscall"
	"time"

	authv3 "github.com/envoyproxy/go-control-plane/envoy/service/auth/v3"
	envoytypev3 "github.com/envoyproxy/go-control-plane/envoy/type/v3"
	rpcstatus "google.golang.org/genproto/googleapis/rpc/status"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/health"
	healthpb "google.golang.org/grpc/health/grpc_health_v1"
	"google.golang.org/grpc/reflection"

	"github.com/beclab/Olares/framework/app-gateway/internal/authz"
)

func main() {
	var grpcAddr, httpAddr, mode, hostUserCheck, skipViewers string
	flag.StringVar(&grpcAddr, "grpc-listen", ":9001", "gRPC ext_authz listen address.")
	flag.StringVar(&httpAddr, "http-listen", ":9002", "HTTP listen address for /healthz, /readyz, /metrics.")
	flag.StringVar(&mode, "mode", "allow", "Authorization decision baseline: 'allow' (Phase A default) or 'deny'.")
	flag.StringVar(&hostUserCheck, "host-user-check", "enabled", "Enable v2 host-user enforcement: 'enabled' (default) or 'disabled'.")
	flag.StringVar(&skipViewers, "host-user-skip-viewers", "", "Comma-separated viewer labels that bypass host-user check (admin SAs).")
	flag.Parse()

	logger := slog.New(slog.NewJSONHandler(os.Stdout, &slog.HandlerOptions{Level: slog.LevelInfo}))
	slog.SetDefault(logger)

	cfg := authz.HostUserConfig{
		Enabled:      strings.EqualFold(hostUserCheck, "enabled"),
		SkipPrefixes: parseSkipViewers(skipViewers),
	}
	logger.Info("app-service-ext-authz starting",
		"grpc", grpcAddr, "http", httpAddr, "mode", mode,
		"host_user_check", cfg.Enabled, "skip_viewers", cfg.SkipPrefixes,
		"version", "phase-a-v2-hostuser")

	grpcLis, err := net.Listen("tcp", grpcAddr)
	if err != nil {
		logger.Error("gRPC listen failed", "err", err)
		os.Exit(1)
	}

	allow := mode != "deny"
	srv := grpc.NewServer()
	authv3.RegisterAuthorizationServer(srv, &authzServer{allow: allow, logger: logger, hostUser: cfg})
	healthSvc := health.NewServer()
	healthSvc.SetServingStatus("", healthpb.HealthCheckResponse_SERVING)
	healthSvc.SetServingStatus(authv3.Authorization_ServiceDesc.ServiceName, healthpb.HealthCheckResponse_SERVING)
	healthpb.RegisterHealthServer(srv, healthSvc)
	reflection.Register(srv)

	ready := &atomic.Bool{}
	ready.Store(true)
	httpSrv := newProbeServer(httpAddr, ready)

	ctx, cancel := signal.NotifyContext(context.Background(), syscall.SIGTERM, syscall.SIGINT)
	defer cancel()

	go func() {
		logger.Info("gRPC listening", "addr", grpcAddr)
		if err := srv.Serve(grpcLis); err != nil && !errors.Is(err, grpc.ErrServerStopped) {
			logger.Error("gRPC serve error", "err", err)
			os.Exit(1)
		}
	}()
	go func() {
		logger.Info("HTTP probes listening", "addr", httpAddr)
		if err := httpSrv.ListenAndServe(); err != nil && !errors.Is(err, http.ErrServerClosed) {
			logger.Error("HTTP probe serve error", "err", err)
		}
	}()

	<-ctx.Done()
	logger.Info("shutdown signal received; draining")
	ready.Store(false)

	shutdownCtx, shutdownCancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer shutdownCancel()
	_ = httpSrv.Shutdown(shutdownCtx)

	stopped := make(chan struct{})
	go func() { srv.GracefulStop(); close(stopped) }()
	select {
	case <-stopped:
	case <-shutdownCtx.Done():
		srv.Stop()
	}
}

// authzServer answers Envoy ext_authz Check requests. Phase A v2 wires the
// HostUser decider in front of the legacy allow_all behaviour:
//
//	HostUser=Pass   → fall through to allow/deny baseline
//	HostUser=Allow  → return OK
//	HostUser=Deny   → return Denied(403, INVALID_HOST_USER)
type authzServer struct {
	authv3.UnimplementedAuthorizationServer
	allow    bool
	logger   *slog.Logger
	hostUser authz.HostUserConfig
}

func (s *authzServer) Check(ctx context.Context, req *authv3.CheckRequest) (*authv3.CheckResponse, error) {
	httpReq := req.GetAttributes().GetRequest().GetHttp()
	host := httpReq.GetHost()
	path := httpReq.GetPath()
	method := httpReq.GetMethod()
	headers := httpReq.GetHeaders()
	rid := extractRequestID(headers)

	dec := authz.HostUser(host, headers, s.hostUser)
	switch dec.Action {
	case authz.ActionDeny:
		s.logger.Warn("authz deny",
			"rid", rid, "authority", host, "method", method, "path", path,
			"viewer_host", dec.Viewer, "viewer_authenticated", dec.Username,
			"code", dec.Code, "decision", "deny")
		body := "Forbidden by app-service-ext-authz: " + dec.Code
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
	case authz.ActionAllow:
		s.logger.Info("authz allow",
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

	if s.allow {
		s.logger.Info("authz allow",
			"rid", rid, "authority", host, "method", method, "path", path,
			"decision", "allow_all")
		return &authv3.CheckResponse{
			Status: &rpcstatus.Status{Code: int32(codes.OK)},
			HttpResponse: &authv3.CheckResponse_OkResponse{
				OkResponse: &authv3.OkHttpResponse{},
			},
		}, nil
	}
	s.logger.Warn("authz deny",
		"rid", rid, "authority", host, "method", method, "path", path,
		"decision", "deny")
	return &authv3.CheckResponse{
		Status: &rpcstatus.Status{Code: int32(codes.PermissionDenied), Message: "phase-a deny mode"},
		HttpResponse: &authv3.CheckResponse_DeniedResponse{
			DeniedResponse: &authv3.DeniedHttpResponse{
				Status: &envoytypev3.HttpStatus{Code: envoytypev3.StatusCode_Forbidden},
				Body:   "Forbidden by app-service-ext-authz (phase-a deny mode)",
			},
		},
	}, nil
}

func parseSkipViewers(s string) []string {
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
		_, _ = w.Write([]byte("# Phase A app-service-ext-authz exposes a minimal metric set.\n"))
		_, _ = w.Write([]byte("gateway_authz_phase 1\n"))
	})
	return &http.Server{
		Addr:              addr,
		Handler:           mux,
		ReadHeaderTimeout: 5 * time.Second,
	}
}
