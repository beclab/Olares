package upgrade

import (
	"context"
	"encoding/json"
	"fmt"
	"net"
	"strings"
	"time"

	corev1 "k8s.io/api/core/v1"
	policyv1 "k8s.io/api/policy/v1"
	apierrors "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/apimachinery/pkg/labels"
	"k8s.io/apimachinery/pkg/types"
	"k8s.io/apimachinery/pkg/util/wait"
	"k8s.io/client-go/dynamic"
	"k8s.io/client-go/kubernetes"
)

// Network replacement can make every affected Pod unready before the first
// eviction. A successful dry-run does not reserve a disruption budget.
func checkFixedMACDisruptionBudget(ctx context.Context, kube kubernetes.Interface, p *fixedMACProgress) error {
	affected := map[types.UID]bool{}
	for _, item := range p.Instances {
		if item.OldUID != "" {
			affected[item.OldUID] = true
		}
	}
	if len(affected) == 0 {
		return nil
	}
	budgets, err := kube.PolicyV1().PodDisruptionBudgets("").List(ctx, metav1.ListOptions{})
	if err != nil {
		return err
	}
	for _, budget := range budgets.Items {
		if budget.Spec.Selector == nil {
			continue
		}
		selector, err := metav1.LabelSelectorAsSelector(budget.Spec.Selector)
		if err != nil {
			return err
		}
		pods, err := kube.CoreV1().Pods(budget.Namespace).List(ctx, metav1.ListOptions{LabelSelector: selector.String()})
		if err != nil {
			return err
		}
		touched, remaining := false, int32(0)
		for _, pod := range pods.Items {
			if !selector.Matches(labels.Set(pod.Labels)) {
				continue
			}
			if affected[pod.UID] {
				touched = true
				continue
			}
			if pod.DeletionTimestamp != nil {
				continue
			}
			if _, disrupted := budget.Status.DisruptedPods[pod.Name]; disrupted {
				continue
			}
			for _, c := range pod.Status.Conditions {
				if c.Type == corev1.PodReady && c.Status == corev1.ConditionTrue {
					remaining++
					break
				}
			}
		}
		if !touched {
			continue
		}
		if budget.Status.ObservedGeneration != budget.Generation {
			return fmt.Errorf("PDB %s/%s has stale status", budget.Namespace, budget.Name)
		}
		if budget.Spec.UnhealthyPodEvictionPolicy != nil && *budget.Spec.UnhealthyPodEvictionPolicy == policyv1.AlwaysAllow {
			continue
		}
		if remaining < budget.Status.DesiredHealthy {
			return fmt.Errorf("PDB %s/%s requires %d healthy Pods but only %d remain outside network migration; resolve maintenance policy before upgrading", budget.Namespace, budget.Name, budget.Status.DesiredHealthy, remaining)
		}
	}
	return nil
}

// Validate legacy identity before stopping DHCP or evicting the old Pod. The
// admission webhook remains the authority that atomically adopts the claim.
func checkFixedMACLegacy(ctx context.Context, kube kubernetes.Interface, dc dynamic.Interface, p *fixedMACProgress) error {
	apps, err := dc.Resource(fixedMACApplications).List(ctx, metav1.ListOptions{})
	if err != nil {
		return err
	}
	for _, item := range p.Instances {
		var template corev1.PodTemplateSpec
		if item.Kind == "Deployment" {
			d, err := kube.AppsV1().Deployments(item.Namespace).Get(ctx, item.Workload, metav1.GetOptions{})
			if err != nil {
				return err
			}
			if d.UID != item.WorkloadUID {
				return fmt.Errorf("MAC preflight: Deployment identity changed")
			}
			template = d.Spec.Template
		} else {
			d, err := kube.AppsV1().StatefulSets(item.Namespace).Get(ctx, item.Workload, metav1.GetOptions{})
			if err != nil {
				return err
			}
			if d.UID != item.WorkloadUID {
				return fmt.Errorf("MAC preflight: StatefulSet identity changed")
			}
			template = d.Spec.Template
		}
		appName := template.Labels["applications.app.bytetrade.io/name"]
		var app *unstructured.Unstructured
		for i := range apps.Items {
			candidate := &apps.Items[i]
			ns, _, _ := unstructured.NestedString(candidate.Object, "spec", "namespace")
			name, _, _ := unstructured.NestedString(candidate.Object, "spec", "name")
			if ns == item.Namespace && name == appName {
				if app != nil {
					return fmt.Errorf("ambiguous Application identity")
				}
				app = candidate
			}
		}
		if app == nil {
			return fmt.Errorf("MAC preflight: no Application for %s/%s", item.Namespace, item.Workload)
		}
		settings, _, err := unstructured.NestedStringMap(app.Object, "spec", "settings")
		if err != nil {
			return err
		}
		ordinal := "singleton"
		if item.Kind == "StatefulSet" {
			ordinal = strings.TrimPrefix(item.Name, item.Workload+"-")
		}
		instance := item.Namespace + "/" + appName + "/" + item.Kind + "/" + item.Workload + "/" + ordinal
		values := map[string]string{}
		if raw := settings["overlayMacvlanMacByInstance"]; raw != "" {
			if err := json.Unmarshal([]byte(raw), &values); err != nil {
				return err
			}
		}
		mac, key := values[instance], instance
		legacy := mac == ""
		if legacy {
			key, mac = item.Namespace+"/"+appName, settings["overlayMacvlanMac"]
			if item.Kind == "StatefulSet" {
				values = map[string]string{}
				if raw := settings["overlayMacvlanMacByOrdinal"]; raw != "" {
					if err := json.Unmarshal([]byte(raw), &values); err != nil {
						return err
					}
				}
				key, mac = key+"/"+ordinal, values[ordinal]
			}
		}
		if mac == "" {
			continue
		}
		parsed, err := net.ParseMAC(mac)
		if err != nil || len(parsed) != 6 || parsed[0] != 0x02 || parsed.String() != mac {
			return fmt.Errorf("invalid persisted fixed MAC for %s", instance)
		}
		if legacy {
			count := 0
			matches := func(t corev1.PodTemplateSpec) bool {
				return t.Labels[fixedMACLabel] == "true" && t.Labels["applications.app.bytetrade.io/name"] == appName
			}
			if item.Kind == "Deployment" {
				list, err := kube.AppsV1().Deployments(item.Namespace).List(ctx, metav1.ListOptions{})
				if err != nil {
					return err
				}
				for _, d := range list.Items {
					if matches(d.Spec.Template) {
						count++
					}
				}
			} else {
				list, err := kube.AppsV1().StatefulSets(item.Namespace).List(ctx, metav1.ListOptions{})
				if err != nil {
					return err
				}
				for _, d := range list.Items {
					if matches(d.Spec.Template) {
						count++
					}
				}
			}
			if count != 1 {
				return fmt.Errorf("legacy MAC for %s has %d possible workloads; retain running Pods until ownership is resolved", instance, count)
			}
		}
		claim, err := dc.Resource(fixedMACAllocations).Get(ctx, strings.ReplaceAll(mac, ":", ""), metav1.GetOptions{})
		if err != nil {
			return fmt.Errorf("verify MAC claim for %s: %w", instance, err)
		}
		field := func(name string) string { v, _, _ := unstructured.NestedString(claim.Object, "spec", name); return v }
		owner := false
		for _, ref := range claim.GetOwnerReferences() {
			if ref.APIVersion == "app.bytetrade.io/v1alpha1" && ref.Kind == "Application" && ref.Name == app.GetName() && ref.UID == app.GetUID() && ref.BlockOwnerDeletion != nil && *ref.BlockOwnerDeletion {
				owner = true
			}
		}
		// A prior admission can have moved the claim before persisting the new index.
		if !owner || field("applicationUID") != string(app.GetUID()) || field("applicationRef") != app.GetName() || field("mac") != mac || (field("instanceKey") != key && field("instanceKey") != instance) || (field("phase") != "Bound" && field("phase") != "Pending") {
			return fmt.Errorf("MAC claim ownership is inconsistent for %s", instance)
		}
		if item.OldUID != "" {
			pod, err := kube.CoreV1().Pods(item.Namespace).Get(ctx, item.Name, metav1.GetOptions{})
			if err != nil {
				return err
			}
			if pod.UID != item.OldUID {
				return fmt.Errorf("MAC preflight: Pod identity changed")
			}
			var statuses []struct {
				Interface string `json:"interface"`
				MAC       string `json:"mac"`
			}
			if raw := pod.Annotations["k8s.v1.cni.cncf.io/network-status"]; raw != "" {
				if err := json.Unmarshal([]byte(raw), &statuses); err != nil {
					return err
				}
				for _, status := range statuses {
					if status.Interface == "net1" && status.MAC != mac {
						return fmt.Errorf("running MAC differs from persisted identity for %s", instance)
					}
				}
			}
		}
	}
	return nil
}

func evictFixedMACPod(ctx context.Context, kube kubernetes.Interface, item fixedMACInstance, interval time.Duration) error {
	deadline, cancel := context.WithTimeout(ctx, 5*time.Minute)
	defer cancel()
	var last error
	err := wait.PollUntilContextCancel(deadline, interval, true, func(ctx context.Context) (bool, error) {
		pod, err := kube.CoreV1().Pods(item.Namespace).Get(ctx, item.Name, metav1.GetOptions{})
		if apierrors.IsNotFound(err) {
			return true, nil
		}
		if err != nil {
			return false, err
		}
		if pod.UID != item.OldUID || pod.DeletionTimestamp != nil {
			return true, nil
		}
		uid := item.OldUID
		last = kube.PolicyV1().Evictions(item.Namespace).Evict(ctx, &policyv1.Eviction{ObjectMeta: metav1.ObjectMeta{Name: item.Name, Namespace: item.Namespace}, DeleteOptions: &metav1.DeleteOptions{Preconditions: &metav1.Preconditions{UID: &uid}}})
		if last == nil || apierrors.IsNotFound(last) {
			return true, nil
		}
		if apierrors.IsTooManyRequests(last) || apierrors.IsConflict(last) || apierrors.IsTimeout(last) || apierrors.IsServerTimeout(last) || apierrors.IsServiceUnavailable(last) {
			return false, nil
		}
		return false, last
	})
	if err != nil {
		return fmt.Errorf("evict %s/%s: %w (last response: %v)", item.Namespace, item.Name, err, last)
	}
	return nil
}
