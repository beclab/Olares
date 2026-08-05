package mesh

import (
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/util/intstr"
	"k8s.io/klog/v2"

	"github.com/beclab/Olares/framework/app-service/pkg/constants"
)

const (
	// LinkerdProxyContainerName is the upstream injected proxy container name.
	LinkerdProxyContainerName = "linkerd-proxy"
	// LinkerdAwaitBinary is shipped in the proxy image (also used by postStart).
	LinkerdAwaitBinary = "/usr/lib/linkerd/linkerd-await"
	// LinkerdAdminPortName is the injected admin port name.
	LinkerdAdminPortName = "linkerd-admin"
)

// HardenLinkerdProxyAdminProbes converts linkerd-proxy HTTP probes against the
// admin port (4191) into in-container exec probes using linkerd-await.
//
// kubelet httpGet to PodIP:4191 traverses eth0 and can be DNATed by workloads
// that rewrite pod iptables (e.g. dockurr). Exec runs in the proxy netns and
// reaches 127.0.0.1:4191 without that path. Returns whether any probe changed.
func HardenLinkerdProxyAdminProbes(pod *corev1.Pod) bool {
	if pod == nil {
		return false
	}
	changed := false
	for i := range pod.Spec.Containers {
		c := &pod.Spec.Containers[i]
		if c.Name != LinkerdProxyContainerName {
			continue
		}
		if hardenLinkerdAdminProbe(&c.LivenessProbe) {
			changed = true
		}
		if hardenLinkerdAdminProbe(&c.ReadinessProbe) {
			changed = true
		}
		if hardenLinkerdAdminProbe(&c.StartupProbe) {
			changed = true
		}
	}
	if changed {
		klog.Infof("mesh-probe: hardened linkerd-proxy admin probes to exec await ns=%s pod=%s",
			pod.Namespace, pod.Name)
	}
	return changed
}

func hardenLinkerdAdminProbe(probe **corev1.Probe) bool {
	if probe == nil || *probe == nil {
		return false
	}
	p := *probe
	if isLinkerdAwaitExec(p) {
		return false
	}
	if p.HTTPGet == nil || !isLinkerdAdminPort(p.HTTPGet.Port) {
		return false
	}

	timeout := p.TimeoutSeconds
	if timeout < 2 {
		timeout = 2
	}
	p.HTTPGet = nil
	p.TCPSocket = nil
	p.GRPC = nil
	p.Exec = &corev1.ExecAction{
		Command: []string{
			LinkerdAwaitBinary,
			"--timeout=1s",
			"--port=4191",
		},
	}
	p.TimeoutSeconds = timeout
	return true
}

func isLinkerdAwaitExec(p *corev1.Probe) bool {
	if p == nil || p.Exec == nil || len(p.Exec.Command) == 0 {
		return false
	}
	return p.Exec.Command[0] == LinkerdAwaitBinary
}

func isLinkerdAdminPort(port intstr.IntOrString) bool {
	switch port.Type {
	case intstr.Int:
		return port.IntVal == int32(constants.LinkerdAdminPort)
	case intstr.String:
		return port.StrVal == LinkerdAdminPortName || port.StrVal == "4191"
	default:
		return false
	}
}
