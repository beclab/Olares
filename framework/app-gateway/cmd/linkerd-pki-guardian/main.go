// Command linkerd-pki-guardian is a minimal always-on controller that
// periodically checks and rotates the Linkerd identity issuer. It replaces the
// legacy CronJob so rotation stays reliable under intermittent power.
package main

import (
	"context"
	"errors"
	"fmt"
	"log/slog"
	"net"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/beclab/Olares/framework/app-gateway/pkg/linkerdpki"
	appsv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/client-go/rest"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

func main() {
	if err := run(); err != nil {
		slog.Error("linkerd-pki-guardian fatal", "error", err)
		os.Exit(1)
	}
}

func run() error {
	interval, err := parseInterval(getenv("GUARDIAN_INTERVAL", "24h"))
	if err != nil {
		return err
	}
	ns := getenv("GUARDIAN_LINKERD_NS", linkerdpki.DefaultLinkerdNamespace)
	mode := getenv("GUARDIAN_MODE", "legacy")
	addr := getenv("GUARDIAN_HTTP_ADDR", ":8080")

	cl, err := buildClient()
	if err != nil {
		return err
	}

	probes := linkerdpki.NewProbeState(interval)
	metrics := linkerdpki.NewMetrics()
	ctrl := linkerdpki.NewController(cl, ns, interval, probes, metrics)

	mux := http.NewServeMux()
	mux.HandleFunc("/startupz", probes.StartupHandler)
	mux.HandleFunc("/healthz", probes.HealthHandler)
	mux.HandleFunc("/readyz", probes.ReadyHandler)
	mux.Handle("/metrics", metrics.Handler())

	// Bind synchronously so a port clash fails fast (PKI-E* / 详设 §2.2).
	ln, err := net.Listen("tcp", addr)
	if err != nil {
		return fmt.Errorf("listen %s: %w", addr, err)
	}
	srv := &http.Server{Handler: mux, ReadHeaderTimeout: 5 * time.Second}
	go func() {
		if serr := srv.Serve(ln); serr != nil && !errors.Is(serr, http.ErrServerClosed) {
			slog.Error("probe server stopped", "error", serr)
		}
	}()

	ctx, stop := signal.NotifyContext(context.Background(), syscall.SIGTERM, syscall.SIGINT)
	defer stop()

	probes.MarkClientReady()
	slog.Info("linkerd-pki-guardian started", "interval", interval.String(), "namespace", ns, "mode", mode)

	ctrl.Run(ctx)

	shutdownCtx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	return srv.Shutdown(shutdownCtx)
}

func buildClient() (client.Client, error) {
	cfg, err := rest.InClusterConfig()
	if err != nil {
		return nil, fmt.Errorf("in-cluster config: %w", err)
	}
	scheme := runtime.NewScheme()
	if err := corev1.AddToScheme(scheme); err != nil {
		return nil, err
	}
	if err := appsv1.AddToScheme(scheme); err != nil {
		return nil, err
	}
	cl, err := client.New(cfg, client.Options{Scheme: scheme})
	if err != nil {
		return nil, fmt.Errorf("build client: %w", err)
	}
	return cl, nil
}

func parseInterval(raw string) (time.Duration, error) {
	d, err := time.ParseDuration(raw)
	if err != nil {
		return 0, fmt.Errorf("parse GUARDIAN_INTERVAL %q: %w", raw, err)
	}
	if d < time.Minute {
		return 0, fmt.Errorf("GUARDIAN_INTERVAL must be >= 1m, got %s", d)
	}
	return d, nil
}

func getenv(key, def string) string {
	if v := os.Getenv(key); v != "" {
		return v
	}
	return def
}
