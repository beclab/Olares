package utils

import (
	"strings"

	"github.com/beclab/Olares/framework/app-service/pkg/constants"
	v1 "k8s.io/api/apps/v1"
	batchv1 "k8s.io/api/batch/v1"
	corev1 "k8s.io/api/core/v1"
	apiresource "k8s.io/apimachinery/pkg/api/resource"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes/scheme"
	"k8s.io/klog/v2"
)

// nvidiaGPUMem is the HAMi per-pod GPU-memory extended resource key.
const nvidiaGPUMem corev1.ResourceName = "nvidia.com/gpumem"

// mib is the unit HAMi defines nvidia.com/gpumem in: the resource carries a
// bare count of MiB, not a byte quantity.
const mib = 1024 * 1024

// WorkloadResourceTotals is the per-pod sum (across every rendered workload's
// pod template) of the container resource requests and limits relevant to
// install-time auto-resource resolution.
//
// cpu and memory are summed in their declared representation. GPU memory is
// normalized to bytes, because its two sources disagree on units: the
// nvidia.com/gpumem resource is a bare MiB count while the pod annotation is a
// plain quantity. Bytes is also what the manifest field these totals backfill
// (requiredGPUMemory) is parsed as, so the value needs no further conversion.
type WorkloadResourceTotals struct {
	RequestsCPU         apiresource.Quantity
	LimitsCPU           apiresource.Quantity
	RequestsMemory      apiresource.Quantity
	LimitsMemory        apiresource.Quantity
	RequestsGPUMemBytes apiresource.Quantity
	LimitsGPUMemBytes   apiresource.Quantity
}

// GetWorkloadResourcesFromChart renders the chart at chartPath with the given
// Helm values (a side-effect-free dry-run, reusing GetResourceListFromChart)
// and sums the per-pod container resource requests/limits across every
// workload's pod template.
//
// It is the install-time mechanism behind the auto-resource ("-1") sentinel:
// for a template app whose concrete resource demand only materializes once the
// user-selected appenv (model, gpu memory, ...) is injected into the chart, the
// caller renders once with the chosen mode + applied appenv and reads back the
// real requirement here.
//
// Per-pod effective request/limit for a resource follows the Kubernetes rule
// max(maxInitContainer, sum(regularContainers)).
func GetWorkloadResourcesFromChart(chartPath string, values map[string]interface{}) (WorkloadResourceTotals, error) {
	var totals WorkloadResourceTotals
	resources, err := GetResourceListFromChart(chartPath, values)
	if err != nil {
		klog.Infof("get resourcelist from chart err=%v", err)
		return totals, err
	}

	add := func(dst *apiresource.Quantity, perPod apiresource.Quantity) {
		if perPod.IsZero() {
			return
		}
		dst.Add(perPod)
	}

	// The resource mode requirement is a PER-POD value by convention: HAMI
	// binds GPU memory per pod, the sidecar webhook injects nvidia.com/gpumem
	// per pod from RequiredGPU, and node-pressure adds the requirement to a
	// single node. The compute package never multiplies by replica count, so
	// neither do we — replicas are intentionally ignored here.
	accumulate := func(meta metav1.ObjectMeta, podSpec corev1.PodSpec) {
		req, lim := effectivePodResources(podSpec)
		add(&totals.RequestsCPU, *req.Cpu())
		add(&totals.LimitsCPU, *lim.Cpu())
		add(&totals.RequestsMemory, *req.Memory())
		add(&totals.LimitsMemory, *lim.Memory())
		gpuReq, gpuLim := podGPUMemoryBytes(meta, req, lim)
		add(&totals.RequestsGPUMemBytes, gpuReq)
		add(&totals.LimitsGPUMemBytes, gpuLim)
	}

	for _, r := range resources {
		kind := r.Object.GetObjectKind().GroupVersionKind().Kind
		switch kind {
		case "Deployment":
			var d v1.Deployment
			if err := scheme.Scheme.Convert(r.Object, &d, nil); err != nil {
				return totals, err
			}
			accumulate(d.Spec.Template.ObjectMeta, d.Spec.Template.Spec)
		case "StatefulSet":
			var s v1.StatefulSet
			if err := scheme.Scheme.Convert(r.Object, &s, nil); err != nil {
				return totals, err
			}
			accumulate(s.Spec.Template.ObjectMeta, s.Spec.Template.Spec)
		case "DaemonSet":
			var ds v1.DaemonSet
			if err := scheme.Scheme.Convert(r.Object, &ds, nil); err != nil {
				return totals, err
			}
			accumulate(ds.Spec.Template.ObjectMeta, ds.Spec.Template.Spec)
		case "ReplicaSet":
			var rs v1.ReplicaSet
			if err := scheme.Scheme.Convert(r.Object, &rs, nil); err != nil {
				return totals, err
			}
			accumulate(rs.Spec.Template.ObjectMeta, rs.Spec.Template.Spec)
		case "Job":
			var j batchv1.Job
			if err := scheme.Scheme.Convert(r.Object, &j, nil); err != nil {
				return totals, err
			}
			accumulate(j.Spec.Template.ObjectMeta, j.Spec.Template.Spec)
		case "Pod":
			var p corev1.Pod
			if err := scheme.Scheme.Convert(r.Object, &p, nil); err != nil {
				return totals, err
			}
			accumulate(p.ObjectMeta, p.Spec)
		}
	}
	return totals, nil
}

// podGPUMemoryBytes reports one pod's declared GPU-memory request and limit in
// bytes, taken from whichever of the two channels the chart used: the
// nvidia.com/gpumem container resource, or the pod annotations that carry the
// same number for the modes which cannot express it as a resource (see
// constants.PodRequiredGPUMemory).
//
// The annotation wins when a pod carries both. A chart that sets it is
// declaring the quota for the mode it was rendered for, whereas the
// nvidia.com/gpumem alongside it belongs to the nvidia branch of the same
// template and would otherwise be double-counted.
func podGPUMemoryBytes(meta metav1.ObjectMeta, req, lim corev1.ResourceList) (apiresource.Quantity, apiresource.Quantity) {
	gpuReq := gpuMemResourceBytes(req)
	gpuLim := gpuMemResourceBytes(lim)
	if q, ok := annotatedGPUMemory(meta, constants.PodRequiredGPUMemory); ok {
		gpuReq = q
	}
	if q, ok := annotatedGPUMemory(meta, constants.PodLimitedGPUMemory); ok {
		gpuLim = q
	}
	return gpuReq, gpuLim
}

// gpuMemResourceBytes converts a pod's nvidia.com/gpumem entry into bytes. The
// resource is a bare MiB count by HAMi's convention, so it cannot be read as a
// byte quantity the way the annotation can.
func gpuMemResourceBytes(list corev1.ResourceList) apiresource.Quantity {
	q, ok := list[nvidiaGPUMem]
	if !ok || q.IsZero() {
		return apiresource.Quantity{}
	}
	return *apiresource.NewQuantity(q.Value()*mib, apiresource.BinarySI)
}

// annotatedGPUMemory reads one of the GPU-memory pod annotations. An
// unparseable or non-positive value is dropped with a log line instead of
// failing the install: the sentinel then stays unresolved and the app is
// treated as declaring no GPU-memory demand, which is exactly where it stood
// before the annotation existed.
func annotatedGPUMemory(meta metav1.ObjectMeta, key string) (apiresource.Quantity, bool) {
	raw := strings.TrimSpace(meta.Annotations[key])
	if raw == "" {
		return apiresource.Quantity{}, false
	}
	q, err := apiresource.ParseQuantity(raw)
	if err != nil {
		klog.Warningf("workload resources: ignoring pod annotation %s=%q, not a valid quantity: %v", key, raw, err)
		return apiresource.Quantity{}, false
	}
	if q.Sign() <= 0 {
		klog.Warningf("workload resources: ignoring pod annotation %s=%q, must be positive", key, raw)
		return apiresource.Quantity{}, false
	}
	return q, true
}

// effectivePodResources computes the pod-level effective resource requests and
// limits following the Kubernetes scheduling rule: for each resource the value
// is max(maxInitContainer, sum(regularContainers)).
func effectivePodResources(spec corev1.PodSpec) (corev1.ResourceList, corev1.ResourceList) {
	sumReq := corev1.ResourceList{}
	sumLim := corev1.ResourceList{}
	for _, c := range spec.Containers {
		addResourceList(sumReq, c.Resources.Requests)
		addResourceList(sumLim, c.Resources.Limits)
	}
	for _, c := range spec.InitContainers {
		maxResourceList(sumReq, c.Resources.Requests)
		maxResourceList(sumLim, c.Resources.Limits)
	}
	return sumReq, sumLim
}

func addResourceList(dst corev1.ResourceList, src corev1.ResourceList) {
	for name, q := range src {
		cur := dst[name]
		cur.Add(q)
		dst[name] = cur
	}
}

func maxResourceList(dst corev1.ResourceList, src corev1.ResourceList) {
	for name, q := range src {
		cur, ok := dst[name]
		if !ok || q.Cmp(cur) > 0 {
			dst[name] = q.DeepCopy()
		}
	}
}
