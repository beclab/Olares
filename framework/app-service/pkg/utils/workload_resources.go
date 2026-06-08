package utils

import (
	v1 "k8s.io/api/apps/v1"
	batchv1 "k8s.io/api/batch/v1"
	corev1 "k8s.io/api/core/v1"
	apiresource "k8s.io/apimachinery/pkg/api/resource"
	"k8s.io/client-go/kubernetes/scheme"
	"k8s.io/klog/v2"
)

// nvidiaGPUMem is the HAMi per-pod GPU-memory extended resource key. Declared
// locally to keep this file free of an extra package dependency.
const nvidiaGPUMem corev1.ResourceName = "nvidia.com/gpumem"

// WorkloadResourceTotals is the per-pod sum (across every rendered workload's
// pod template) of the container resource requests and limits relevant to
// install-time auto-resource resolution.
//
// GPUMem is summed from the per-pod nvidia.com/gpumem extended resource. The
// remaining fields are standard k8s quantities (cpu, memory) and are summed in
// their declared representation.
type WorkloadResourceTotals struct {
	RequestsCPU    apiresource.Quantity
	LimitsCPU      apiresource.Quantity
	RequestsMemory apiresource.Quantity
	LimitsMemory   apiresource.Quantity
	RequestsGPUMem apiresource.Quantity
	LimitsGPUMem   apiresource.Quantity
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
	accumulate := func(podSpec corev1.PodSpec) {
		req, lim := effectivePodResources(podSpec)
		add(&totals.RequestsCPU, *req.Cpu())
		add(&totals.LimitsCPU, *lim.Cpu())
		add(&totals.RequestsMemory, *req.Memory())
		add(&totals.LimitsMemory, *lim.Memory())
		add(&totals.RequestsGPUMem, req[nvidiaGPUMem])
		add(&totals.LimitsGPUMem, lim[nvidiaGPUMem])
	}

	for _, r := range resources {
		kind := r.Object.GetObjectKind().GroupVersionKind().Kind
		switch kind {
		case "Deployment":
			var d v1.Deployment
			if err := scheme.Scheme.Convert(r.Object, &d, nil); err != nil {
				return totals, err
			}
			accumulate(d.Spec.Template.Spec)
		case "StatefulSet":
			var s v1.StatefulSet
			if err := scheme.Scheme.Convert(r.Object, &s, nil); err != nil {
				return totals, err
			}
			accumulate(s.Spec.Template.Spec)
		case "DaemonSet":
			var ds v1.DaemonSet
			if err := scheme.Scheme.Convert(r.Object, &ds, nil); err != nil {
				return totals, err
			}
			accumulate(ds.Spec.Template.Spec)
		case "ReplicaSet":
			var rs v1.ReplicaSet
			if err := scheme.Scheme.Convert(r.Object, &rs, nil); err != nil {
				return totals, err
			}
			accumulate(rs.Spec.Template.Spec)
		case "Job":
			var j batchv1.Job
			if err := scheme.Scheme.Convert(r.Object, &j, nil); err != nil {
				return totals, err
			}
			accumulate(j.Spec.Template.Spec)
		case "Pod":
			var p corev1.Pod
			if err := scheme.Scheme.Convert(r.Object, &p, nil); err != nil {
				return totals, err
			}
			accumulate(p.Spec)
		}
	}
	return totals, nil
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
