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
	admissionregistrationv1 "k8s.io/api/admissionregistration/v1"
	appsv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/client-go/rest"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

func main() {
	if len(os.Args) > 1 && os.Args[1] == "bootstrap" {
		if err := runBootstrapOneshot(); err != nil {
			slog.Error("linkerd-pki-guardian bootstrap fatal", "error", err)
			os.Exit(1)
		}
		os.Exit(0)
	}
	if len(os.Args) > 1 && os.Args[1] == "webhook-certgen" {
		overwrite := false
		for _, a := range os.Args[2:] {
			if a == "--overwrite" {
				overwrite = true
			}
		}
		if err := runWebhookCertgenOneshot(overwrite); err != nil {
			slog.Error("linkerd-pki-guardian webhook-certgen fatal", "error", err)
			os.Exit(1)
		}
		os.Exit(0)
	}
	if err := run(); err != nil {
		slog.Error("linkerd-pki-guardian fatal", "error", err)
		os.Exit(1)
	}
}

func runWebhookCertgenOneshot(overwrite bool) error {
	ns := getenv("GUARDIAN_LINKERD_NS", linkerdpki.DefaultLinkerdNamespace)
	cl, err := buildClient()
	if err != nil {
		return err
	}
	ctx := context.Background()
	if err := linkerdpki.EnsureWebhookCerts(ctx, cl, ns, overwrite); err != nil {
		return err
	}
	slog.Info("linkerd-pki-guardian webhook-certgen complete",
		"op", "webhook-certgen", "namespace", ns, "overwrite", overwrite)
	return nil
}

func runBootstrapOneshot() error {
	ns := getenv("GUARDIAN_LINKERD_NS", linkerdpki.DefaultLinkerdNamespace)
	cl, err := buildClient()
	if err != nil {
		return err
	}
	ctx := context.Background()
	created, err := linkerdpki.BootstrapIfMissing(ctx, cl, ns)
	if err != nil {
		return err
	}
	syncChanged, err := linkerdpki.SyncIdentityToLinkerd(ctx, cl, ns)
	if err != nil {
		return err
	}
	slog.Info("linkerd-pki-guardian bootstrap oneshot complete",
		"op", "bootstrap", "namespace", ns, "created", created, "sync_changed", syncChanged)
	return nil
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

	// Bind synchronously so a port clash fails fast (see Linkerd-PKI guardian detailed design §2.2).
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
	if err := admissionregistrationv1.AddToScheme(scheme); err != nil {
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
