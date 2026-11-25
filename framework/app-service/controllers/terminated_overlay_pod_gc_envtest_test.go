package controllers

import (
	"encoding/json"
	"strings"
	"time"

	"github.com/beclab/Olares/framework/app-service/pkg/constants"
	"github.com/beclab/Olares/framework/app-service/pkg/utils/maclease"
	"github.com/beclab/Olares/framework/app-service/pkg/webhook"
	appv1 "github.com/beclab/api/api/app.bytetrade.io/v1alpha1"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	admissionv1 "k8s.io/api/admission/v1"
	appsv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
	apierrors "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/client-go/kubernetes"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/controller/controllerutil"
)

var _ = Describe("Overlay terminal Pod recovery", func() {
	It("reclaims the Jellyfin migration failure and hands off its fixed MAC after deletion", func() {
		now := time.Now()
		d, _ := terminatedPodGCFixture(now, terminatedPodGCTestCase{replicas: 1})
		d.Name, d.Namespace, d.UID = "jellyfin", "default", ""
		labels := map[string]string{constants.ApplicationNameLabel: d.Name, constants.ApplicationOwnerLabel: "alice", constants.ApplicationMacvlanInitLabel: "true"}
		d.Spec.Selector = &metav1.LabelSelector{MatchLabels: labels}
		d.Spec.Strategy.Type = appsv1.RecreateDeploymentStrategyType
		fixtureContainers := func(names ...string) []corev1.Container {
			containers := make([]corev1.Container, 0, len(names))
			for _, name := range names {
				containers = append(containers, corev1.Container{Name: name, Image: "example.invalid/app:test"})
			}
			return containers
		}
		d.Spec.Template = corev1.PodTemplateSpec{ObjectMeta: metav1.ObjectMeta{Labels: labels}, Spec: corev1.PodSpec{
			InitContainers: fixtureContainers("init-chmod-data", "migrate-encoding", "linkerd-init", constants.MacvlanReplyInitContainerName, "olares-mesh-in-agent-iptables", "macvlan-bypass-iptables"),
			Containers:     fixtureContainers("linkerd-proxy", "jellyfin", "olares-mesh-in-agent"),
		}}
		create := func(obj client.Object) {
			Expect(k8sClient.Create(suiteCtx, obj)).To(Succeed())
			DeferCleanup(func() {
				// The API-only environment has no kubelet to finish Pod shutdown.
				err := k8sClient.Delete(suiteCtx, obj, client.GracePeriodSeconds(0))
				Expect(err == nil || apierrors.IsNotFound(err)).To(BeTrue())
				Eventually(func() bool {
					return apierrors.IsNotFound(k8sClient.Get(suiteCtx, client.ObjectKeyFromObject(obj), obj))
				}, 5*time.Second, 50*time.Millisecond).Should(BeTrue(), "cleanup of %T %s", obj, client.ObjectKeyFromObject(obj))
			})
		}
		create(d)
		node := &corev1.Node{ObjectMeta: metav1.ObjectMeta{
			Name: "overlay-gc-node", Labels: map[string]string{"node-role.kubernetes.io/control-plane": ""},
		}}
		create(node)
		node.Status.Conditions = []corev1.NodeCondition{{Type: corev1.NodeReady, Status: corev1.ConditionTrue}}
		Expect(k8sClient.Status().Update(suiteCtx, node)).To(Succeed())
		app := &appv1.Application{ObjectMeta: metav1.ObjectMeta{Name: d.Namespace + "-" + d.Name}, Spec: appv1.ApplicationSpec{
			Appid: "overlay-gc", Name: d.Name, Owner: "alice", Namespace: d.Namespace, DeploymentName: d.Name,
			Settings: map[string]string{"enableOverlayGateway": "true"},
		}}
		create(app)
		DeferCleanup(func() {
			var current appv1.Application
			Expect(k8sClient.Get(suiteCtx, client.ObjectKeyFromObject(app), &current)).To(Succeed())
			if controllerutil.RemoveFinalizer(&current, "app.bytetrade.io/overlay-mac-claim") {
				Expect(k8sClient.Update(suiteCtx, &current)).To(Succeed())
			}
		})
		rs := &appsv1.ReplicaSet{ObjectMeta: metav1.ObjectMeta{
			Name: d.Name + "-rs", Namespace: d.Namespace,
			OwnerReferences: []metav1.OwnerReference{*metav1.NewControllerRef(d, appsv1.SchemeGroupVersion.WithKind(deployment))},
		}, Spec: appsv1.ReplicaSetSpec{Selector: d.Spec.Selector.DeepCopy(), Template: *d.Spec.Template.DeepCopy()}}
		create(rs)
		leaseNamespace := &corev1.Namespace{ObjectMeta: metav1.ObjectMeta{Name: maclease.Namespace}}
		if err := k8sClient.Create(suiteCtx, leaseNamespace); err != nil {
			// An existing namespace can be shared with other specs.
			Expect(apierrors.IsAlreadyExists(err)).To(BeTrue())
		}
		admission, err := webhook.New(cfg)
		Expect(err).NotTo(HaveOccurred())
		kube, err := kubernetes.NewForConfig(cfg)
		Expect(err).NotTo(HaveOccurred())
		admit := func(pod *corev1.Pod) error {
			raw, err := json.Marshal(pod)
			if err != nil {
				return err
			}
			if _, err = admission.CreateMacvlanInitPatchWithContext(suiteCtx, &admissionv1.AdmissionRequest{
				Object: runtime.RawExtension{Raw: raw},
			}, pod); err != nil {
				return err
			}
			return admission.ValidateMacvlanAnnotation(suiteCtx, pod, pod.Namespace, false)
		}
		pod := &corev1.Pod{ObjectMeta: metav1.ObjectMeta{
			Name: d.Name + "-old", Namespace: d.Namespace, Labels: labels,
			Finalizers:      []string{"test.example/teardown"},
			OwnerReferences: []metav1.OwnerReference{*metav1.NewControllerRef(rs, appsv1.SchemeGroupVersion.WithKind(replicaSet))},
		}, Spec: *d.Spec.Template.Spec.DeepCopy()}
		pod.Spec.NodeName = node.Name
		replacement := pod.DeepCopy()
		replacement.Name = d.Name + "-replacement"
		replacement.Finalizers = nil
		Expect(admit(pod)).To(Succeed())
		create(pod)
		Expect(maclease.Reconcile(suiteCtx, kube, app.UID)).To(Succeed())
		const networks = "k8s.v1.cni.cncf.io/networks"
		oldSelection := pod.Annotations[networks]
		var selection []struct{ Mac string }
		Expect(json.Unmarshal([]byte(oldSelection), &selection)).To(Succeed())
		Expect(selection).To(HaveLen(1))
		mac := selection[0].Mac
		allocation := &unstructured.Unstructured{}
		allocation.SetAPIVersion("app.bytetrade.io/v1alpha1")
		allocation.SetKind("OverlayMACAllocation")
		allocationKey := client.ObjectKey{Name: strings.ReplaceAll(mac, ":", "")}
		Expect(k8sClient.Get(suiteCtx, allocationKey, allocation)).To(Succeed())
		allocationUID := allocation.GetUID()
		DeferCleanup(func() {
			Expect(k8sClient.Delete(suiteCtx, allocation)).To(Succeed())
			err := kube.CoordinationV1().Leases(maclease.Namespace).Delete(suiteCtx, maclease.Name(mac), metav1.DeleteOptions{})
			Expect(err == nil || apierrors.IsNotFound(err)).To(BeTrue())
		})
		DeferCleanup(func() {
			var remaining corev1.Pod
			if err := k8sClient.Get(suiteCtx, client.ObjectKeyFromObject(pod), &remaining); err == nil {
				remaining.Finalizers = nil
				Expect(k8sClient.Update(suiteCtx, &remaining)).To(Succeed())
			} else {
				Expect(apierrors.IsNotFound(err)).To(BeTrue())
			}
		})
		// Reproduce the captured Pod describe state with relative timestamps.
		// Unknown was a container exit reason; the Pod phase itself was Failed.
		finished := metav1.NewTime(now.Add(-2 * time.Minute))
		start := metav1.NewTime(finished.Add(-4*time.Minute - 44*time.Second))
		pod.Status = corev1.PodStatus{
			Phase: corev1.PodFailed, Reason: "UnexpectedAdmissionError", StartTime: &start,
			Message: "Pod was rejected: Allocate failed due to no healthy devices present; cannot allocate unhealthy devices gpu.intel.com/i915, which is unexpected",
			Conditions: []corev1.PodCondition{
				{Type: corev1.PodReadyToStartContainers, Status: corev1.ConditionFalse, LastTransitionTime: finished},
				{Type: corev1.PodInitialized, Status: corev1.ConditionTrue},
				{Type: corev1.PodReady, Status: corev1.ConditionFalse, LastTransitionTime: finished},
				{Type: corev1.ContainersReady, Status: corev1.ConditionFalse},
				{Type: corev1.PodScheduled, Status: corev1.ConditionTrue},
			},
		}
		for i, container := range pod.Spec.InitContainers {
			pod.Status.InitContainerStatuses = append(pod.Status.InitContainerStatuses, corev1.ContainerStatus{
				Name: container.Name, Ready: true, State: corev1.ContainerState{Terminated: &corev1.ContainerStateTerminated{
					Reason: "Completed", FinishedAt: metav1.NewTime(start.Add(time.Duration(i) * time.Second)),
				}},
			})
		}
		for _, container := range pod.Spec.Containers {
			pod.Status.ContainerStatuses = append(pod.Status.ContainerStatuses, corev1.ContainerStatus{
				Name: container.Name, State: corev1.ContainerState{Terminated: &corev1.ContainerStateTerminated{
					Reason: "Unknown", ExitCode: 255, FinishedAt: finished,
				}},
			})
		}
		Expect(k8sClient.Status().Update(suiteCtx, pod)).To(Succeed())
		Expect(admit(replacement.DeepCopy())).To(MatchError(ContainSubstring("still has pod")))
		r := &TerminatedPodGCController{Client: k8sClient, minPodAge: time.Minute, readyDelay: time.Minute, now: func() time.Time { return now }}
		_, err = r.Reconcile(suiteCtx, ctrl.Request{NamespacedName: client.ObjectKeyFromObject(d)})
		Expect(err).NotTo(HaveOccurred())
		var terminating corev1.Pod
		Expect(k8sClient.Get(suiteCtx, client.ObjectKeyFromObject(pod), &terminating)).To(Succeed())
		Expect(terminating.DeletionTimestamp).NotTo(BeNil())
		Expect(terminating.Finalizers).To(ConsistOf("test.example/teardown"))
		Expect(terminating.UID).To(Equal(pod.UID))
		Expect(admit(replacement.DeepCopy())).To(MatchError(ContainSubstring("still has pod")))
		Expect(maclease.Validate(suiteCtx, kube, pod, app.UID, mac)).To(Succeed())
		terminating.Finalizers = nil
		Expect(k8sClient.Update(suiteCtx, &terminating)).To(Succeed())
		Eventually(func() bool {
			return apierrors.IsNotFound(k8sClient.Get(suiteCtx, client.ObjectKeyFromObject(pod), &corev1.Pod{}))
		}, 5*time.Second, 50*time.Millisecond).Should(BeTrue())

		Expect(admit(replacement)).To(Succeed())
		Expect(replacement.Annotations[networks]).To(Equal(oldSelection))
		create(replacement)
		Expect(replacement.UID).NotTo(Equal(pod.UID))
		Expect(k8sClient.Get(suiteCtx, allocationKey, allocation)).To(Succeed())
		Expect(allocation.GetUID()).To(Equal(allocationUID))
		Expect(maclease.Validate(suiteCtx, kube, replacement, app.UID, mac)).To(Succeed())
		Expect(maclease.Validate(suiteCtx, kube, pod, app.UID, mac)).NotTo(Succeed())
	})
})
