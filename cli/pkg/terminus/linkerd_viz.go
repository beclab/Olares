package terminus

import (
	"context"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"

	"github.com/beclab/Olares/cli/pkg/core/logger"
	"github.com/beclab/Olares/cli/pkg/utils"
	"github.com/pkg/errors"
	corev1 "k8s.io/api/core/v1"
	apierrors "k8s.io/apimachinery/pkg/api/errors"
	"helm.sh/helm/v3/pkg/cli"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

const (
	linkerdVizNamespace            = "linkerd-viz"
	defaultLinkerdPrometheusURL    = "http://prometheus-k8s.kubesphere-monitoring-system.svc.cluster.local:9090"
	linkerdVizValuesFileName       = "linkerd-viz-values.yaml"
	linkerdPrometheusProxyRBACFile = "deploy/linkerd/prometheus-pod-proxy-rbac.yaml"
)

// InstallLinkerdViz installs the optional linkerd-viz extension without bundled Prometheus.
func InstallLinkerdViz(ctx context.Context, c client.Client, settings *cli.EnvSettings, vendorDir, prometheusURL string) error {
	if prometheusURL == "" {
		prometheusURL = defaultLinkerdPrometheusURL
	}
	vals := linkerdVizValuesPath(vendorDir)
	if vals == "" {
		return errors.New("linkerd-viz-values.yaml not found in vendor (run framework/app-gateway/hack/sync-vendor-values.sh)")
	}
	if err := applyLinkerdVizManifest(ctx, settings, vals, prometheusURL); err != nil {
		return err
	}
	if err := removeBundledLinkerdVizPrometheus(ctx, settings); err != nil {
		return err
	}
	if err := ensureLinkerdVizNamespaceLabels(ctx, c); err != nil {
		return err
	}
	if err := applyLinkerdPrometheusRBAC(ctx, settings, vendorDir); err != nil {
		return err
	}
	logger.InfoInstallationProgress("linkerd-viz installed (platform Prometheus, no bundled Prometheus)")
	return nil
}

func applyLinkerdVizManifest(ctx context.Context, settings *cli.EnvSettings, valsPath, prometheusURL string) error {
	linkerd, err := exec.LookPath("linkerd")
	if err != nil {
		return errors.Wrap(err, "linkerd CLI not found in PATH")
	}
	logger.InfoInstallationProgress(fmt.Sprintf("Installing linkerd-viz (prometheusUrl=%s) ...", prometheusURL))
	cmd := exec.CommandContext(ctx, linkerd, "viz", "install",
		"-f", valsPath,
		"--set", "prometheusUrl="+prometheusURL,
	)
	if settings != nil && settings.KubeConfig != "" {
		cmd.Env = append(os.Environ(), "KUBECONFIG="+settings.KubeConfig)
	}
	out, err := cmd.Output()
	if err != nil {
		return errors.Wrap(err, "linkerd viz install")
	}
	tmp, err := os.CreateTemp("", "linkerd-viz-*.yaml")
	if err != nil {
		return err
	}
	defer os.Remove(tmp.Name())
	if _, err := tmp.Write(out); err != nil {
		return err
	}
	if err := tmp.Close(); err != nil {
		return err
	}
	return utils.KubectlApplyFile(ctx, settings, tmp.Name())
}

func removeBundledLinkerdVizPrometheus(ctx context.Context, settings *cli.EnvSettings) error {
	kubectl, err := exec.LookPath("kubectl")
	if err != nil {
		return errors.Wrap(err, "kubectl not found")
	}
	args := []string{"delete", "deploy/prometheus", "svc/prometheus", "cm/prometheus-config", "sa/prometheus",
		"-n", linkerdVizNamespace, "--ignore-not-found=true"}
	cmd := exec.CommandContext(ctx, kubectl, args...)
	if settings != nil && settings.KubeConfig != "" {
		cmd.Env = append(os.Environ(), "KUBECONFIG="+settings.KubeConfig)
	}
	return cmd.Run()
}

func ensureLinkerdVizNamespaceLabels(ctx context.Context, c client.Client) error {
	var ns corev1.Namespace
	if err := c.Get(ctx, client.ObjectKey{Name: linkerdVizNamespace}, &ns); err != nil {
		if apierrors.IsNotFound(err) {
			return nil
		}
		return err
	}
	patch := client.MergeFrom(ns.DeepCopy())
	if ns.Labels == nil {
		ns.Labels = map[string]string{}
	}
	ns.Labels["bytetrade.io/ns-type"] = "system"
	if err := c.Patch(ctx, &ns, patch); err != nil {
		return err
	}
	return nil
}

func applyLinkerdPrometheusRBAC(ctx context.Context, settings *cli.EnvSettings, vendorDir string) error {
	path := linkerdPrometheusRBACManifest(vendorDir)
	if path == "" {
		logger.Info("linkerd prometheus RBAC manifest not found; skip")
		return nil
	}
	return utils.KubectlApplyFile(ctx, settings, path)
}

func linkerdVizValuesPath(vendorDir string) string {
	if vendorDir != "" {
		p := filepath.Join(vendorDir, linkerdVizValuesFileName)
		if st, err := os.Stat(p); err == nil && !st.IsDir() {
			return p
		}
	}
	if root := os.Getenv("OLARES_SOURCE_ROOT"); root != "" {
		p := filepath.Join(root, "framework", "app-gateway", "vendor-charts-values", linkerdVizValuesFileName)
		if st, err := os.Stat(p); err == nil && !st.IsDir() {
			return p
		}
	}
	return ""
}

func linkerdPrometheusRBACManifest(vendorDir string) string {
	candidates := []string{}
	if vendorDir != "" {
		candidates = append(candidates, filepath.Join(vendorDir, "deploy", "linkerd", "prometheus-pod-proxy-rbac.yaml"))
	}
	if root := os.Getenv("OLARES_SOURCE_ROOT"); root != "" {
		candidates = append(candidates, filepath.Join(root, "framework", "app-gateway", linkerdPrometheusProxyRBACFile))
	}
	candidates = append(candidates, filepath.Join("framework", "app-gateway", linkerdPrometheusProxyRBACFile))
	for _, p := range candidates {
		if st, err := os.Stat(p); err == nil && !st.IsDir() {
			return p
		}
	}
	return ""
}
