// Command access-admission serves ValidatingAdmissionWebhook for access.olares.io
// and reconciles AppAccessPolicy projections. Product posture has no kill switch.
package main

import (
	"context"
	"crypto/tls"
	"crypto/x509"
	"fmt"
	"io"
	"log/slog"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/beclab/Olares/framework/app-gateway/pkg/access"
	"github.com/beclab/Olares/framework/app-gateway/pkg/uniauth"
	admissionregistrationv1 "k8s.io/api/admissionregistration/v1"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"k8s.io/apimachinery/pkg/types"
	"k8s.io/client-go/rest"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

func main() {
	addr := getenv("ACCESS_ADMISSION_ADDR", ":9443")
	mux := http.NewServeMux()
	mux.HandleFunc("/healthz", func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte("ok"))
	})
	mux.HandleFunc("/validate", validateHandler)

	srv := &http.Server{Addr: addr, Handler: mux, ReadHeaderTimeout: 5 * time.Second}
	ctx, stop := signal.NotifyContext(context.Background(), syscall.SIGTERM, syscall.SIGINT)
	defer stop()

	cl, err := buildAdmissionClient()
	if err != nil {
		slog.Error("admission client unavailable", "error", err)
		os.Exit(1)
	}
	bootstrapCtx, cancelBootstrap := context.WithTimeout(ctx, 90*time.Second)
	ca, err := access.PrepareAdmissionTLS(bootstrapCtx, cl)
	if err != nil {
		slog.Error("prepare admission TLS", "error", err)
		os.Exit(1)
	}
	cert, err := loadAdmissionTLS(bootstrapCtx, cl)
	if err != nil {
		slog.Error("load admission TLS", "error", err)
		os.Exit(1)
	}
	srv.TLSConfig = &tls.Config{MinVersion: tls.VersionTLS12, Certificates: []tls.Certificate{cert}}
	serverErr := make(chan error, 1)
	go func() {
		slog.Info("access-admission listening tls", "addr", addr)
		serverErr <- srv.ListenAndServeTLS("", "")
	}()
	serviceURL := "https://" + access.AdmissionServiceName + "." + access.AdmissionTLSNamespace + ".svc/healthz"
	if err := waitForAdmissionEndpoint(bootstrapCtx, serviceURL, ca, serverErr); err != nil {
		slog.Error("admission service not ready", "error", err)
		os.Exit(1)
	}
	if err := access.PromoteAdmissionFailurePolicy(bootstrapCtx, cl, ca); err != nil {
		slog.Error("promote admission failurePolicy", "error", err)
		os.Exit(1)
	}
	cancelBootstrap()
	slog.Info("admission webhook policy reconciled to Fail")

	if uniauth.AccessPolicyReconcileDesired() {
		go runPolicyReconcileLoop(ctx)
	}

	select {
	case <-ctx.Done():
	case err := <-serverErr:
		slog.Error("admission server stopped", "error", err)
		os.Exit(1)
	}
	shutdown, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	_ = srv.Shutdown(shutdown)
}

func validateHandler(w http.ResponseWriter, r *http.Request) {
	body, err := io.ReadAll(io.LimitReader(r.Body, 1<<20))
	if err != nil {
		http.Error(w, "read body", http.StatusBadRequest)
		return
	}
	out, err := access.AdmitAdmissionReview(body)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	w.Header().Set("Content-Type", "application/json")
	_, _ = w.Write(out)
}

func runPolicyReconcileLoop(ctx context.Context) {
	cfg, err := rest.InClusterConfig()
	if err != nil {
		slog.Error("policy reconcile: in-cluster config", "error", err)
		return
	}
	scheme := runtime.NewScheme()
	cl, err := client.New(cfg, client.Options{Scheme: scheme})
	if err != nil {
		slog.Error("policy reconcile: client", "error", err)
		return
	}
	ticker := time.NewTicker(30 * time.Second)
	defer ticker.Stop()
	for {
		select {
		case <-ctx.Done():
			return
		case <-ticker.C:
			reconcileAllPolicies(ctx, cl)
		}
	}
}

func reconcileAllPolicies(ctx context.Context, cl client.Client) {
	list := &unstructured.UnstructuredList{}
	list.SetGroupVersionKind(schema.GroupVersionKind{
		Group: "access.olares.io", Version: "v1alpha1", Kind: "AppAccessPolicyList",
	})
	if err := cl.List(ctx, list); err != nil {
		slog.Warn("list AppAccessPolicy", "error", err)
		return
	}
	enabled := uniauth.PolicyCompilationDesired()
	for i := range list.Items {
		item := &list.Items[i]
		if err := access.ReconcileAccessPolicy(ctx, cl, item, enabled); err != nil {
			slog.Warn("reconcile policy", "name", item.GetName(), "error", err)
		}
	}
}

func buildAdmissionClient() (client.Client, error) {
	cfg, err := rest.InClusterConfig()
	if err != nil {
		return nil, err
	}
	scheme := runtime.NewScheme()
	if err := corev1.AddToScheme(scheme); err != nil {
		return nil, err
	}
	if err := admissionregistrationv1.AddToScheme(scheme); err != nil {
		return nil, err
	}
	return client.New(cfg, client.Options{Scheme: scheme})
}

func loadAdmissionTLS(ctx context.Context, cl client.Client) (tls.Certificate, error) {
	var sec corev1.Secret
	if err := cl.Get(ctx, types.NamespacedName{
		Namespace: access.AdmissionTLSNamespace,
		Name:      access.AdmissionTLSSecretName,
	}, &sec); err != nil {
		return tls.Certificate{}, err
	}
	return tls.X509KeyPair(sec.Data["tls.crt"], sec.Data["tls.key"])
}

func waitForAdmissionEndpoint(ctx context.Context, url string, ca []byte, serverErr <-chan error) error {
	roots := x509.NewCertPool()
	if !roots.AppendCertsFromPEM(ca) {
		return fmt.Errorf("invalid webhook CA bundle")
	}
	transport := &http.Transport{TLSClientConfig: &tls.Config{RootCAs: roots, MinVersion: tls.VersionTLS12}}
	defer transport.CloseIdleConnections()
	client := &http.Client{Transport: transport, Timeout: 2 * time.Second,
		CheckRedirect: func(_ *http.Request, _ []*http.Request) error { return http.ErrUseLastResponse }}
	ticker := time.NewTicker(time.Second)
	defer ticker.Stop()
	var lastErr error
	for {
		select {
		case err := <-serverErr:
			return fmt.Errorf("admission server stopped: %w", err)
		case <-ctx.Done():
			return fmt.Errorf("waiting for admission service: %w (last probe: %v)", ctx.Err(), lastErr)
		case <-ticker.C:
			req, err := http.NewRequestWithContext(ctx, http.MethodGet, url, nil)
			if err != nil {
				return err
			}
			resp, err := client.Do(req)
			if err != nil {
				lastErr = err
				continue
			}
			resp.Body.Close()
			if resp.StatusCode == http.StatusOK {
				return nil
			}
			lastErr = fmt.Errorf("unexpected HTTP status %d", resp.StatusCode)
		}
	}
}

func getenv(k, def string) string {
	if v := os.Getenv(k); v != "" {
		return v
	}
	return def
}
