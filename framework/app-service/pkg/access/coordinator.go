package access

import (
	"context"
	"fmt"
	"os"
	"slices"
	"sync"
	"time"

	admissionregv1 "k8s.io/api/admissionregistration/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/rest"
	"k8s.io/client-go/tools/leaderelection"
	"k8s.io/client-go/tools/leaderelection/resourcelock"
	"k8s.io/klog/v2"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

const (
	writerLease   = "app-service-access-policy-writer"
	catalogConfig = "olares-app-access-catalog"
	writerGuard   = "olares-access-projection-writer"
)

var requiredGuards = []string{"olares-access-posture", "olares-access-grant-lifecycle", writerGuard}

// RunCoordinator acquires its own Lease even when other app-service
// controllers run without leader election. Only the holder can project grants.
func RunCoordinator(ctx context.Context, cfg *rest.Config, scheme *runtime.Scheme, podUID string) error {
	if podUID == "" {
		return fmt.Errorf("APP_SERVICE_POD_UID is required for the access policy writer")
	}
	bounded := rest.CopyConfig(cfg)
	bounded.Timeout = 5 * time.Second
	kube, err := kubernetes.NewForConfig(bounded)
	if err != nil {
		return err
	}
	cl, err := client.New(bounded, client.Options{Scheme: scheme})
	if err != nil {
		return err
	}
	lock := &resourcelock.LeaseLock{
		LeaseMeta:  metav1.ObjectMeta{Namespace: PolicyNamespace, Name: writerLease},
		Client:     kube.CoordinationV1(),
		LockConfig: resourcelock.ResourceLockConfig{Identity: podUID},
	}
	return runElected(ctx, lock, func(leaderCtx context.Context) {
		runProjectionLoop(leaderCtx, cl)
	})
}

func runElected(ctx context.Context, lock resourcelock.Interface, writer func(context.Context)) error {
	// Drain the previous writer before returning from the runnable. The
	// lease remains held until expiry rather than being released on cancel.
	var mu sync.Mutex
	stopped := false
	elector, err := leaderelection.NewLeaderElector(leaderelection.LeaderElectionConfig{
		Lock: lock, LeaseDuration: 45 * time.Second, RenewDeadline: 15 * time.Second, RetryPeriod: 3 * time.Second,
		Callbacks: leaderelection.LeaderCallbacks{
			OnStartedLeading: func(leaderCtx context.Context) {
				mu.Lock()
				defer mu.Unlock()
				if !stopped {
					writer(leaderCtx)
				}
			},
			OnStoppedLeading: func() {},
		},
	})
	if err != nil {
		return err
	}
	elector.Run(ctx)
	mu.Lock()
	stopped = true
	mu.Unlock()
	if ctx.Err() == nil {
		return fmt.Errorf("access policy writer lost its lease")
	}
	return nil
}

func runProjectionLoop(ctx context.Context, c client.Client) {
	ticker := time.NewTicker(10 * time.Second)
	defer ticker.Stop()
	for ctx.Err() == nil {
		var gateErr error
		if err := guardsReady(ctx, c); err != nil {
			gateErr = err
		} else {
			ca, err := os.ReadFile("/etc/certs/ca.crt")
			if err == nil {
				err = ensureCatalogWebhook(ctx, c, ca, probeCatalogWebhook)
			}
			gateErr = err
		}
		if gateErr != nil {
			klog.Warningf("access policy writer is withholding grants: %v", gateErr)
		}
		if err := reconcilePolicyPass(ctx, c, gateErr == nil); err != nil && ctx.Err() == nil {
			klog.Warningf("access policy reconcile pass: %v", err)
		}
		select {
		case <-ctx.Done():
			return
		case <-ticker.C:
		}
	}
}

func guardsReady(ctx context.Context, c client.Client) error {
	for _, name := range requiredGuards {
		var policy admissionregv1.ValidatingAdmissionPolicy
		if err := c.Get(ctx, client.ObjectKey{Name: name}, &policy); err != nil {
			return err
		}
		if policy.Spec.FailurePolicy == nil || *policy.Spec.FailurePolicy != admissionregv1.Fail ||
			policy.Status.ObservedGeneration != policy.Generation || policy.Status.TypeChecking == nil ||
			len(policy.Status.TypeChecking.ExpressionWarnings) != 0 {
			return fmt.Errorf("admission policy %s is not fail closed and type checked", name)
		}
		var binding admissionregv1.ValidatingAdmissionPolicyBinding
		if err := c.Get(ctx, client.ObjectKey{Name: name}, &binding); err != nil {
			return err
		}
		if binding.Spec.PolicyName != name || binding.Spec.MatchResources != nil ||
			!slices.Contains(binding.Spec.ValidationActions, admissionregv1.Deny) {
			return fmt.Errorf("admission policy %s has no unrestricted deny binding", name)
		}
		if name == writerGuard {
			kind, ref := policy.Spec.ParamKind, binding.Spec.ParamRef
			if kind == nil || kind.APIVersion != "coordination.k8s.io/v1" || kind.Kind != "Lease" ||
				ref == nil || ref.Name != writerLease || ref.Namespace != PolicyNamespace || ref.Selector != nil ||
				ref.ParameterNotFoundAction == nil || *ref.ParameterNotFoundAction != admissionregv1.DenyAction {
				return fmt.Errorf("projection writer guard is not bound to the writer Lease")
			}
		} else if policy.Spec.ParamKind != nil || binding.Spec.ParamRef != nil {
			return fmt.Errorf("unexpected parameter for %s", name)
		}
	}
	return nil
}

func reconcilePolicyPass(ctx context.Context, c client.Client, ready bool) error {
	var list unstructured.UnstructuredList
	list.SetGroupVersionKind(accessGVK("AppAccessPolicyList"))
	if err := c.List(ctx, &list, client.InNamespace(PolicyNamespace)); err != nil {
		return err
	}
	// Withdrawal always precedes publication. A failing grant cannot keep
	// the remaining grants from being revoked during the same pass.
	for i := range list.Items {
		if ctx.Err() != nil {
			return ctx.Err()
		}
		u := &list.Items[i]
		state := value(u.Object, "spec", "desiredState")
		if ready && state == DesiredActive && u.GetDeletionTimestamp().IsZero() {
			continue
		}
		b, err := readBinding(u)
		if err == nil {
			err = b.withdraw(ctx, c)
		}
		if err != nil {
			klog.Warningf("withdraw access policy %s: %v", u.GetName(), err)
		}
	}
	if !ready {
		return nil
	}
	for i := range list.Items {
		if ctx.Err() != nil {
			return ctx.Err()
		}
		u := &list.Items[i]
		if value(u.Object, "spec", "desiredState") != DesiredActive || !u.GetDeletionTimestamp().IsZero() {
			continue
		}
		if err := ReconcileAccessPolicy(ctx, c, u, true); err != nil {
			klog.Warningf("reconcile access policy %s: %v", u.GetName(), err)
		}
	}
	return nil
}
