package access

import (
	"context"
	"testing"
	"time"

	admissionregv1 "k8s.io/api/admissionregistration/v1"
	coordinationv1 "k8s.io/api/coordination/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/client-go/kubernetes/fake"
	"k8s.io/client-go/tools/leaderelection/resourcelock"
	"k8s.io/utils/ptr"
	"sigs.k8s.io/controller-runtime/pkg/client"
	ctrlfake "sigs.k8s.io/controller-runtime/pkg/client/fake"
)

func TestLeaseContenderCannotStartWriter(t *testing.T) {
	now := metav1.NewMicroTime(time.Now())
	clientset := fake.NewClientset(&coordinationv1.Lease{
		ObjectMeta: metav1.ObjectMeta{Namespace: PolicyNamespace, Name: writerLease},
		Spec:       coordinationv1.LeaseSpec{HolderIdentity: ptr.To("current-pod"), LeaseDurationSeconds: ptr.To(int32(45)), AcquireTime: &now, RenewTime: &now},
	})
	lock := &resourcelock.LeaseLock{LeaseMeta: metav1.ObjectMeta{Namespace: PolicyNamespace, Name: writerLease},
		Client: clientset.CoordinationV1(), LockConfig: resourcelock.ResourceLockConfig{Identity: "contender"}}
	ctx, cancel := context.WithTimeout(context.Background(), 50*time.Millisecond)
	defer cancel()
	wrote := make(chan struct{}, 1)
	if err := runElected(ctx, lock, func(context.Context) { wrote <- struct{}{} }); err != nil {
		t.Fatal(err)
	}
	select {
	case <-wrote:
		t.Fatal("non-holder started a policy writer")
	default:
	}
}

func TestAdmittedGuardsAreRequiredForProjection(t *testing.T) {
	ctx := context.Background()
	scheme := runtime.NewScheme()
	_ = admissionregv1.AddToScheme(scheme)
	var objects []client.Object
	for _, name := range requiredGuards {
		policy := &admissionregv1.ValidatingAdmissionPolicy{
			ObjectMeta: metav1.ObjectMeta{Name: name, Generation: 2},
			Spec:       admissionregv1.ValidatingAdmissionPolicySpec{FailurePolicy: ptr.To(admissionregv1.Fail)},
			Status:     admissionregv1.ValidatingAdmissionPolicyStatus{ObservedGeneration: 2, TypeChecking: &admissionregv1.TypeChecking{}},
		}
		binding := &admissionregv1.ValidatingAdmissionPolicyBinding{ObjectMeta: metav1.ObjectMeta{Name: name},
			Spec: admissionregv1.ValidatingAdmissionPolicyBindingSpec{PolicyName: name, ValidationActions: []admissionregv1.ValidationAction{admissionregv1.Deny}}}
		if name == writerGuard {
			policy.Spec.ParamKind = &admissionregv1.ParamKind{APIVersion: "coordination.k8s.io/v1", Kind: "Lease"}
			binding.Spec.ParamRef = &admissionregv1.ParamRef{Name: writerLease, Namespace: PolicyNamespace,
				ParameterNotFoundAction: ptr.To(admissionregv1.DenyAction)}
		}
		objects = append(objects, policy, binding)
	}
	c := ctrlfake.NewClientBuilder().WithScheme(scheme).WithObjects(objects...).Build()
	if err := guardsReady(ctx, c); err != nil {
		t.Fatal(err)
	}
	binding := &admissionregv1.ValidatingAdmissionPolicyBinding{}
	if err := c.Get(ctx, client.ObjectKey{Name: requiredGuards[0]}, binding); err != nil {
		t.Fatal(err)
	}
	binding.Spec.ValidationActions = []admissionregv1.ValidationAction{admissionregv1.Audit}
	if err := c.Update(ctx, binding); err != nil {
		t.Fatal(err)
	}
	if err := guardsReady(ctx, c); err == nil {
		t.Fatal("a non-blocking admission policy approved projection")
	}
}

func TestLeaseHolderStopsItsWriterBeforeReturning(t *testing.T) {
	clientset := fake.NewClientset()
	lock := &resourcelock.LeaseLock{LeaseMeta: metav1.ObjectMeta{Namespace: PolicyNamespace, Name: writerLease},
		Client: clientset.CoordinationV1(), LockConfig: resourcelock.ResourceLockConfig{Identity: "holder-pod"}}
	ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	defer cancel()
	started := make(chan struct{})
	stopped := make(chan struct{})
	done := make(chan error, 1)
	go func() {
		done <- runElected(ctx, lock, func(writerCtx context.Context) {
			close(started)
			<-writerCtx.Done()
			close(stopped)
		})
	}()
	select {
	case <-started:
	case <-ctx.Done():
		t.Fatal("the lease holder did not start the writer")
	}
	cancel()
	if err := <-done; err != nil {
		t.Fatal(err)
	}
	select {
	case <-stopped:
	default:
		t.Fatal("the runnable returned without draining the writer")
	}
	lease, err := clientset.CoordinationV1().Leases(PolicyNamespace).Get(context.Background(), writerLease, metav1.GetOptions{})
	if err != nil || lease.Spec.HolderIdentity == nil || *lease.Spec.HolderIdentity != "holder-pod" {
		t.Fatalf("writer released its Lease before natural expiry: %v", err)
	}
}

func TestMissingPodIdentityBlocksCoordinatorStartup(t *testing.T) {
	if err := RunCoordinator(context.Background(), nil, nil, ""); err == nil {
		t.Fatal("a writer started without a Pod UID")
	}
}

func TestCancelledWriterCannotPublishProjection(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	c, u := bindingFixture(t, RouteModeDirect)
	b, err := readBinding(u)
	if err != nil {
		t.Fatal(err)
	}
	if err := b.resolve(ctx, c); err != nil {
		t.Fatal(err)
	}
	plan, err := CompilePolicy(b.policy, true)
	if err != nil {
		t.Fatal(err)
	}
	objects, err := BuildLinkerdObjects(plan, b.policy)
	if err != nil {
		t.Fatal(err)
	}
	cancel()
	if err := applyProjections(ctx, c, objects); err != context.Canceled {
		t.Fatalf("cancelled writer did not stop: %v", err)
	}
	mustAbsent(t, c, plan.AuthorizationPolicyName, b.policy.Namespace)
}
