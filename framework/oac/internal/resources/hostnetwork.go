package resources

import (
	"errors"
	"fmt"

	"helm.sh/helm/v3/pkg/kube"
	corev1 "k8s.io/api/core/v1"
)

// CheckHostNetwork rejects rendered workloads whose pod spec shares the
// host network or the host PID namespace. The check runs on the helm
// dry-run, so a template that sets hostNetwork from a value is judged by
// the boolean that would actually be installed, not by the source text.
func CheckHostNetwork(list kube.ResourceList) error {
	var errs []error
	walkPodSpecs(list, func(kind, name string, spec corev1.PodSpec) {
		if spec.HostNetwork {
			errs = append(errs, fmt.Errorf("%s %q must not set hostNetwork: true", kind, name))
		}
		if spec.HostPID {
			errs = append(errs, fmt.Errorf("%s %q must not set hostPID: true", kind, name))
		}
	})
	return errors.Join(errs...)
}
