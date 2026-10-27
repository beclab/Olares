package webhook

import (
	"context"
	"fmt"
	"os"
	"time"

	discoveryv1 "k8s.io/api/discovery/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/labels"
	"k8s.io/client-go/kubernetes"
	"k8s.io/klog/v2"
)

const (
	envWebhookServiceEndpointTimeout      = "WEBHOOK_SERVICE_ENDPOINT_TIMEOUT"
	envWebhookServiceEndpointPollInterval = "WEBHOOK_SERVICE_ENDPOINT_POLL_INTERVAL"
	envPodIP                              = "POD_IP"
)

// WaitForServiceEndpointReady polls EndpointSlices until podIP appears as a
// ready address for the given service port, or until timeout. On timeout it
// returns nil after logging so callers can still start controllers (Parts 2/3
// provide safety nets).
func WaitForServiceEndpointReady(
	ctx context.Context,
	kubeClient kubernetes.Interface,
	namespace, serviceName, podIP string,
	port int32,
	timeout, interval time.Duration,
) error {
	if podIP == "" {
		klog.Warning("POD_IP empty; skip waiting for webhook service endpoint")
		return nil
	}
	if timeout <= 0 {
		timeout = 90 * time.Second
	}
	if interval <= 0 {
		interval = time.Second
	}

	deadline := time.Now().Add(timeout)
	klog.Infof("Waiting for service endpoint namespace=%s service=%s podIP=%s port=%d timeout=%v",
		namespace, serviceName, podIP, port, timeout)

	for {
		ready, err := serviceEndpointHasPodIP(ctx, kubeClient, namespace, serviceName, podIP, port)
		if err != nil {
			klog.Warningf("list EndpointSlice for %s/%s failed: %v", namespace, serviceName, err)
		} else if ready {
			klog.Infof("Service webhook endpoint ready, podIP=%s service=%s/%s", podIP, namespace, serviceName)
			return nil
		}

		if time.Now().After(deadline) {
			klog.Errorf("timed out waiting for webhook service endpoint podIP=%s service=%s/%s after %v; starting controllers anyway",
				podIP, namespace, serviceName, timeout)
			return nil
		}

		timer := time.NewTimer(interval)
		select {
		case <-ctx.Done():
			timer.Stop()
			return ctx.Err()
		case <-timer.C:
		}
	}
}

func serviceEndpointHasPodIP(
	ctx context.Context,
	kubeClient kubernetes.Interface,
	namespace, serviceName, podIP string,
	port int32,
) (bool, error) {
	selector := labels.Set{"kubernetes.io/service-name": serviceName}.AsSelector().String()
	slices, err := kubeClient.DiscoveryV1().EndpointSlices(namespace).List(ctx, metav1.ListOptions{
		LabelSelector: selector,
	})
	if err != nil {
		return false, err
	}
	for i := range slices.Items {
		if endpointSliceHasReadyPodIP(&slices.Items[i], podIP, port) {
			return true, nil
		}
	}
	return false, nil
}

func endpointSliceHasReadyPodIP(slice *discoveryv1.EndpointSlice, podIP string, port int32) bool {
	portOK := false
	for _, p := range slice.Ports {
		if p.Port != nil && *p.Port == port {
			portOK = true
			break
		}
	}
	if !portOK && len(slice.Ports) > 0 {
		// Some slices list multiple ports; require an explicit match when ports are present.
		return false
	}
	for _, ep := range slice.Endpoints {
		if ep.Conditions.Ready != nil && !*ep.Conditions.Ready {
			continue
		}
		for _, addr := range ep.Addresses {
			if addr == podIP {
				return true
			}
		}
	}
	return false
}

// ResolvePodIP returns POD_IP from the environment (Downward API).
func ResolvePodIP() string {
	return os.Getenv(envPodIP)
}

// ParseServiceEndpointWaitConfig reads timeout/interval envs with defaults.
func ParseServiceEndpointWaitConfig() (timeout, interval time.Duration) {
	timeout = 90 * time.Second
	interval = time.Second
	if v := os.Getenv(envWebhookServiceEndpointTimeout); v != "" {
		if d, err := time.ParseDuration(v); err == nil {
			timeout = d
		} else {
			klog.Warningf("invalid %s=%q: %v", envWebhookServiceEndpointTimeout, v, err)
		}
	}
	if v := os.Getenv(envWebhookServiceEndpointPollInterval); v != "" {
		if d, err := time.ParseDuration(v); err == nil {
			interval = d
		} else {
			klog.Warningf("invalid %s=%q: %v", envWebhookServiceEndpointPollInterval, v, err)
		}
	}
	return timeout, interval
}

// FormatWaitSummary is used by tests.
func FormatWaitSummary(namespace, service, podIP string, port int32) string {
	return fmt.Sprintf("%s/%s podIP=%s port=%d", namespace, service, podIP, port)
}
