package resources

import (
	"errors"
	"fmt"

	"helm.sh/helm/v3/pkg/kube"
	corev1 "k8s.io/api/core/v1"
)

// CheckSubPathExpr rejects any volumeMount that uses subPathExpr.
//
// The platform prepares userspace hostPath directories from an admission
// webhook, which sees the mount as it is written in the pod spec: at that
// point subPathExpr is still the literal "$(VAR)" text, and references to
// the downward API (metadata.name, status.podIP, ...) have no value at
// all. Only kubelet expands it, well after admission. A platform that
// expanded the environment-variable half and gave up on the rest would
// silently prepare the wrong host directory, which is worse than not
// preparing it -- so the form is disallowed outright and charts use a
// static subPath instead.
//
// The check is typed on purpose: it runs over the helm-rendered
// kube.ResourceList rather than scanning templates/ line by line. Charts
// like ray and chaos-mesh vendor CRDs whose OpenAPI schema embeds the
// full PodSpec definition, so the literal string "subPathExpr" appears in
// their YAML without any workload actually using it.
func CheckSubPathExpr(list kube.ResourceList) error {
	var errs []error
	walkPodSpecs(list, func(kind, name string, spec corev1.PodSpec) {
		for _, c := range allContainers(spec) {
			for _, vm := range c.VolumeMounts {
				if vm.SubPathExpr == "" {
					continue
				}
				errs = append(errs, fmt.Errorf(
					"volumeMounts.subPathExpr is not supported: %s %s, container %s, mount %s (subPathExpr %q); use a static subPath instead",
					kind, name, c.Name, vm.Name, vm.SubPathExpr,
				))
			}
		}
	})
	return errors.Join(errs...)
}
