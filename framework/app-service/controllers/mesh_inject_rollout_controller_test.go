package controllers

import (
	"context"
	"testing"
	"time"

	"github.com/beclab/Olares/framework/app-service/pkg/gateway/meshinagent"
	"github.com/beclab/Olares/framework/app-service/pkg/mesh"
	appsv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/types"
	kubefake "k8s.io/client-go/kubernetes/fake"
	"sigs.k8s.io/controller-runtime/pkg/client/fake"
	"sigs.k8s.io/controller-runtime/pkg/reconcile"
)

func linkerdReadyObjs() []runtime.Object {
	ns := mesh.LinkerdNamespace
	mk := func(name string) *appsv1.Deployment {
		return &appsv1.Deployment{
			ObjectMeta: metav1.ObjectMeta{Name: name, Namespace: ns},
			Status:     appsv1.DeploymentStatus{ReadyReplicas: 1},
		}
	}
	return []runtime.Object{
		mk("linkerd-destination"),
		mk("linkerd-identity"),
		mk("linkerd-proxy-injector"),
		mk("linkerd-pki-guardian"),
	}
}

func TestMeshInjectRolloutReconcilerFirstReady(t *testing.T) {
	scheme := runtime.NewScheme()
	_ = corev1.AddToScheme(scheme)
	_ = appsv1.AddToScheme(scheme)

	meshinagent.SetMeshControlPlaneReadyCheck(mesh.IsControlPlaneReady)
	t.Cleanup(func() { meshinagent.SetMeshControlPlaneReadyCheck(nil) })

	c := fake.NewClientBuilder().WithScheme(scheme).Build()
	kube := kubefake.NewSimpleClientset(linkerdReadyObjs()...)
	w := meshinagent.NewRolloutWorker(c, meshinagent.NewRolloutQueue(2))
	r := &MeshInjectRolloutReconciler{
		Client:       c,
		Kube:         kube,
		Worker:       w,
		requeueAfter: time.Hour,
	}
	if _, err := r.Reconcile(context.Background(), reconcile.Request{}); err != nil {
		t.Fatal(err)
	}
	ready, epoch, err := meshinagent.LoadMeshInjectRolloutState(context.Background(), c)
	if err != nil || !ready || epoch != "1" {
		t.Fatalf("ready=%v epoch=%s err=%v", ready, epoch, err)
	}
	if _, err := r.Reconcile(context.Background(), reconcile.Request{}); err != nil {
		t.Fatal(err)
	}
	ready, epoch, err = meshinagent.LoadMeshInjectRolloutState(context.Background(), c)
	if err != nil || !ready || epoch != "1" {
		t.Fatalf("second ready=%v epoch=%s err=%v", ready, epoch, err)
	}
	var cm corev1.ConfigMap
	if err := c.Get(context.Background(), types.NamespacedName{
		Namespace: meshinagent.MeshInjectRolloutStateCMNamespace,
		Name:      meshinagent.MeshInjectRolloutStateCMName,
	}, &cm); err != nil {
		t.Fatal(err)
	}
}
