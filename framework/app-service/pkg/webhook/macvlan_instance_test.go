package webhook

import (
	"github.com/beclab/Olares/framework/app-service/pkg/constants"
	"github.com/beclab/Olares/framework/app-service/pkg/utils/maclease"
	"github.com/stretchr/testify/require"
	appsv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/types"
	"strings"
	"testing"
)

func TestFixedMACTerminalPodBlocksUntilDeletionThenReusesMAC(t *testing.T) {
	for _, phase := range []corev1.PodPhase{corev1.PodSucceeded, corev1.PodUnknown} {
		t.Run(string(phase), func(t *testing.T) {
			wh := testMacvlanWebhook()
			old := macvlanPod()
			old.UID = "old-pod-uid"
			mac, err := wh.ensureOverlayMAC(t.Context(), old, false)
			require.NoError(t, err)
			app, err := wh.dynamicClient.AppV1alpha1().Applications().Get(t.Context(), "app-space-jellyfin", metav1.GetOptions{})
			require.NoError(t, err)
			require.NoError(t, maclease.Reserve(t.Context(), wh.kubeClient, old, app.Name, app.UID, mac))
			old.Status.Phase = phase
			_, err = wh.kubeClient.CoreV1().Pods(old.Namespace).Create(t.Context(), old, metav1.CreateOptions{})
			require.NoError(t, err)
			require.NoError(t, maclease.Reconcile(t.Context(), wh.kubeClient, app.UID))

			replacement := macvlanPod()
			replacement.Name = "replacement"
			_, err = wh.ensureOverlayMAC(t.Context(), replacement, false)
			require.ErrorContains(t, err, "still has pod")
			require.ErrorContains(t, maclease.Reserve(t.Context(), wh.kubeClient, replacement, app.Name, app.UID, mac), "still occupied")

			// Normal deletion may take time. Both admission guards must keep
			// protecting the MAC while the old Pod is still terminating.
			now := metav1.Now()
			old.DeletionTimestamp = &now
			_, err = wh.kubeClient.CoreV1().Pods(old.Namespace).Update(t.Context(), old, metav1.UpdateOptions{})
			require.NoError(t, err)
			_, err = wh.ensureOverlayMAC(t.Context(), replacement, false)
			require.ErrorContains(t, err, "still has pod")
			require.ErrorContains(t, maclease.Reserve(t.Context(), wh.kubeClient, replacement, app.Name, app.UID, mac), "still occupied")

			require.NoError(t, wh.kubeClient.CoreV1().Pods(old.Namespace).Delete(t.Context(), old.Name, metav1.DeleteOptions{}))
			reused, err := wh.ensureOverlayMAC(t.Context(), replacement, false)
			require.NoError(t, err)
			require.Equal(t, mac, reused)
			require.NoError(t, maclease.Reserve(t.Context(), wh.kubeClient, replacement, app.Name, app.UID, reused))
			require.NoError(t, maclease.Validate(t.Context(), wh.kubeClient, replacement, app.UID, reused))
			require.Error(t, maclease.Validate(t.Context(), wh.kubeClient, old, app.UID, mac))
		})
	}
}

func testDeploymentOwner(name string) metav1.OwnerReference {
	yes := true
	return metav1.OwnerReference{APIVersion: "apps/v1", Kind: "ReplicaSet", Name: name + "-rs", UID: types.UID("rs-" + name), Controller: &yes}
}
func testDeploymentObjects(ns, name, owner string) []runtime.Object {
	yes := true
	one := int32(1)
	labels := map[string]string{constants.ApplicationNameLabel: name, constants.ApplicationOwnerLabel: owner}
	template := corev1.PodTemplateSpec{ObjectMeta: metav1.ObjectMeta{Labels: labels}}
	return []runtime.Object{
		&appsv1.Deployment{ObjectMeta: metav1.ObjectMeta{Name: name, Namespace: ns, UID: types.UID("deployment-" + name)}, Spec: appsv1.DeploymentSpec{Replicas: &one, Strategy: appsv1.DeploymentStrategy{Type: appsv1.RecreateDeploymentStrategyType}, Template: template}},
		&appsv1.ReplicaSet{ObjectMeta: metav1.ObjectMeta{Name: name + "-rs", Namespace: ns, UID: types.UID("rs-" + name), OwnerReferences: []metav1.OwnerReference{{APIVersion: "apps/v1", Kind: "Deployment", Name: name, UID: types.UID("deployment-" + name), Controller: &yes}}}, Spec: appsv1.ReplicaSetSpec{Template: template}},
	}
}
func tenantDeploymentObjects(ts []tenant) []runtime.Object {
	var out []runtime.Object
	for _, t := range ts {
		out = append(out, testDeploymentObjects(t.namespace, t.appName, t.owner)...)
	}
	return out
}
func TestFixedMACRejectsRollingOverlap(t *testing.T) {
	wh := testMacvlanWebhook()
	pod := macvlanPod()
	d, _ := wh.kubeClient.AppsV1().Deployments(pod.Namespace).Get(t.Context(), "jellyfin", metav1.GetOptions{})
	d.Spec.Strategy.Type = appsv1.RollingUpdateDeploymentStrategyType
	wh.kubeClient.AppsV1().Deployments(pod.Namespace).Update(t.Context(), d, metav1.UpdateOptions{})
	if _, err := wh.ensureOverlayMAC(t.Context(), pod, false); err == nil || !strings.Contains(err.Error(), "Recreate") {
		t.Fatalf("rolling strategy: %v", err)
	}
	d.Spec.Strategy.Type = appsv1.RecreateDeploymentStrategyType
	wh.kubeClient.AppsV1().Deployments(pod.Namespace).Update(t.Context(), d, metav1.UpdateOptions{})
	old := pod.DeepCopy()
	old.Name = "old-pod"
	now := metav1.Now()
	old.DeletionTimestamp = &now
	wh.kubeClient.CoreV1().Pods(pod.Namespace).Create(t.Context(), old, metav1.CreateOptions{})
	if _, err := wh.ensureOverlayMAC(t.Context(), pod, false); err == nil || !strings.Contains(err.Error(), "still has pod") {
		t.Fatalf("terminating overlap: %v", err)
	}
}
func TestFixedMACStatefulSetsDoNotShareOrdinal(t *testing.T) {
	wh := testMacvlanWebhook()
	yes := true
	var macs []string
	for _, name := range []string{"media", "database"} {
		sts := &appsv1.StatefulSet{ObjectMeta: metav1.ObjectMeta{Name: name, Namespace: "app-space", UID: types.UID(name)}, Spec: appsv1.StatefulSetSpec{Template: corev1.PodTemplateSpec{ObjectMeta: metav1.ObjectMeta{Labels: map[string]string{constants.ApplicationNameLabel: "jellyfin"}}}}}
		wh.kubeClient.AppsV1().StatefulSets("app-space").Create(t.Context(), sts, metav1.CreateOptions{})
		pod := macvlanPod()
		pod.Name = name + "-0"
		pod.OwnerReferences = []metav1.OwnerReference{{APIVersion: "apps/v1", Kind: "StatefulSet", Name: name, UID: types.UID(name), Controller: &yes}}
		mac, err := wh.ensureOverlayMAC(t.Context(), pod, false)
		if err != nil {
			t.Fatal(err)
		}
		macs = append(macs, mac)
	}
	if macs[0] == macs[1] {
		t.Fatal("distinct StatefulSets shared a MAC")
	}
}
func TestFixedMACRejectsForgedOwner(t *testing.T) {
	wh := testMacvlanWebhook()
	pod := macvlanPod()
	pod.OwnerReferences[0].UID = "forged"
	if _, err := wh.ensureOverlayMAC(t.Context(), pod, false); err == nil {
		t.Fatal("accepted forged owner UID")
	}
}
