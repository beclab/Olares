package access

import (
	"context"
	"errors"
	"fmt"
	"reflect"

	apierrors "k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/apimachinery/pkg/types"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/controller/controllerutil"
)

const bindingFinalizer = "access.olares.io/revoke-projection"

// ReconcileAccessPolicy resolves a persisted grant and withdraws existing permissions
// when any bound object is replaced, removed, or no longer matches approval.
func ReconcileAccessPolicy(ctx context.Context, c client.Client, u *unstructured.Unstructured, enabled bool) error {
	if !enabled {
		return nil
	}
	b, err := readBinding(u)
	if err != nil {
		return err
	}
	deleting := !u.GetDeletionTimestamp().IsZero()
	if deleting || b.policy.DesiredState == DesiredRevoked {
		if err := b.withdraw(ctx, c); err != nil {
			return err
		}
		if deleting && controllerutil.ContainsFinalizer(u, bindingFinalizer) {
			original := u.DeepCopy()
			controllerutil.RemoveFinalizer(u, bindingFinalizer)
			return c.Patch(ctx, u, client.MergeFromWithOptions(original, client.MergeFromWithOptimisticLock{}))
		}
		return nil
	}
	if err := b.resolve(ctx, c); err != nil {
		return errors.Join(err, b.withdraw(ctx, c))
	}
	if !controllerutil.ContainsFinalizer(u, bindingFinalizer) {
		original := u.DeepCopy()
		controllerutil.AddFinalizer(u, bindingFinalizer)
		return c.Patch(ctx, u, client.MergeFromWithOptions(original, client.MergeFromWithOptimisticLock{}))
	}
	if b.policy.RouteMode == RouteModeViaAppGateway {
		return b.reconcileGateway(ctx, c)
	}
	if err := b.checkTarget(ctx, c); err != nil {
		return errors.Join(err, b.withdrawAllows(ctx, c))
	}
	plan, err := CompilePolicy(b.policy, true)
	if err != nil {
		return err
	}
	objects, err := BuildLinkerdObjects(plan, b.policy)
	if err != nil {
		return err
	}
	if err := applyProjections(ctx, c, objects[:3]); err != nil {
		return errors.Join(err, b.withdrawAllows(ctx, c))
	}
	if ok, err := routeReady(ctx, c, objects[2], "policy.linkerd.io", "Server", b.policy.Namespace, objects[0].Name, ""); err != nil || !ok {
		return errors.Join(err, fmt.Errorf("target route is not accepted for its current generation"), b.withdrawAllows(ctx, c))
	}
	if err := b.stillActive(ctx, c); err != nil {
		return errors.Join(err, b.withdrawAllows(ctx, c))
	}
	if err := b.reconcileTransport(ctx, c, false); err != nil {
		return errors.Join(err, b.withdrawAllows(ctx, c))
	}
	if err := b.stillActive(ctx, c); err != nil {
		return errors.Join(err, b.withdrawAllows(ctx, c))
	}
	if err := applyProjections(ctx, c, objects[3:]); err != nil {
		return errors.Join(err, b.withdrawAllows(ctx, c))
	}
	return nil
}

func (b boundPolicy) withdraw(ctx context.Context, c client.Client) error {
	b.policy.DesiredState = DesiredRevoked
	if b.policy.RouteMode == RouteModeViaAppGateway {
		return b.withdrawGateway(ctx, c)
	}
	return errors.Join(reconcileProjectedPolicy(ctx, c, b.policy), b.reconcileTransport(ctx, c, true))
}

func reconcileProjectedPolicy(ctx context.Context, c client.Client, p PolicySpec) error {
	plan, err := CompilePolicy(p, true)
	if err != nil {
		return err
	}
	desired, err := BuildLinkerdObjects(plan, p)
	if err != nil {
		return err
	}
	return applyProjections(ctx, c, desired)
}

func applyProjections(ctx context.Context, c client.Client, desired []DesiredObject) error {
	for _, d := range desired {
		if err := ctx.Err(); err != nil {
			return err
		}
		existing := &unstructured.Unstructured{}
		existing.SetGroupVersionKind(d.GVK)
		err := c.Get(ctx, types.NamespacedName{Namespace: d.Namespace, Name: d.Name}, existing)
		switch {
		case apierrors.IsNotFound(err):
			if !d.Delete {
				if err := c.Create(ctx, d.Object); err != nil {
					return err
				}
			}
			continue
		case err != nil:
			return err
		}
		labels := existing.GetLabels()
		if labels[ManagedByLabel] != ManagedByAccessPolicy || labels[PolicyNameLabel] != d.Object.GetLabels()[PolicyNameLabel] {
			return fmt.Errorf("refusing to change %s %s/%s owned by another controller", d.GVK.Kind, d.Namespace, d.Name)
		}
		if d.Delete {
			uid, rv := existing.GetUID(), existing.GetResourceVersion()
			if err := c.Delete(ctx, existing, client.Preconditions{UID: &uid, ResourceVersion: &rv}); err != nil && !apierrors.IsNotFound(err) {
				return err
			}
			continue
		}
		if reflect.DeepEqual(existing.Object["spec"], d.Object.Object["spec"]) {
			continue
		}
		existing.Object["spec"] = d.Object.Object["spec"]
		if err := c.Update(ctx, existing); err != nil {
			return err
		}
	}
	return nil
}
