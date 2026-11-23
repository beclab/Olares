// Command access-admission serves ValidatingAdmissionWebhook for access.olares.io
// and reconciles AppAccessPolicy projections. Product posture has no kill switch.
package main

import (
	"context"
	"io"
	"log/slog"
	"net/http"
	"os"
	"os/signal"
	"path/filepath"
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
	certFile := getenv("ACCESS_ADMISSION_CERT", "/tls/tls.crt")
	keyFile := getenv("ACCESS_ADMISSION_KEY", "/tls/tls.key")

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
	if err := access.PromoteAdmissionFailurePolicy(ctx, cl); err != nil {
		slog.Error("promote admission failurePolicy", "error", err)
		os.Exit(1)
	}
	if err := materializeAdmissionTLS(ctx, cl, certFile, keyFile); err != nil {
		slog.Error("materialize admission tls files", "error", err)
		os.Exit(1)
	}
	slog.Info("admission webhook policy reconciled to Fail")

	if uniauth.AccessPolicyReconcileDesired() {
		go runPolicyReconcileLoop(ctx)
	}

	go func() {
		slog.Info("access-admission listening tls", "addr", addr)
		if err := srv.ListenAndServeTLS(certFile, keyFile); err != nil && err != http.ErrServerClosed {
			slog.Error("server stopped", "error", err)
			stop()
		}
	}()

	<-ctx.Done()
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

func materializeAdmissionTLS(ctx context.Context, cl client.Client, certFile, keyFile string) error {
	var sec corev1.Secret
	if err := cl.Get(ctx, types.NamespacedName{
		Namespace: access.AdmissionTLSNamespace,
		Name:      access.AdmissionTLSSecretName,
	}, &sec); err != nil {
		return err
	}
	if err := os.MkdirAll(filepath.Dir(certFile), 0o755); err != nil {
		return err
	}
	if err := os.WriteFile(certFile, sec.Data["tls.crt"], 0o600); err != nil {
		return err
	}
	return os.WriteFile(keyFile, sec.Data["tls.key"], 0o600)
}

func getenv(k, def string) string {
	if v := os.Getenv(k); v != "" {
		return v
	}
	return def
}
