package access

import (
	"context"
	"fmt"

	apierrors "k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/apimachinery/pkg/types"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

// ReconcileAccessPolicy applies or deletes Linkerd projections for one policy.
// When compilationEnabled is false, no API writes occur.
func ReconcileAccessPolicy(ctx context.Context, c client.Client, u *unstructured.Unstructured, compilationEnabled bool) error {
	if u == nil {
		return fmt.Errorf("nil policy")
	}
	p, err := PolicyFromUnstructured(u)
	if err != nil {
		return err
	}
	plan, err := CompilePolicy(p, compilationEnabled)
	if err != nil {
		return err
	}
	if !compilationEnabled {
		return nil
	}
	desired, err := BuildLinkerdObjects(plan, p)
	if err != nil {
		return err
	}
	for _, d := range desired {
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
		if labels[ManagedByLabel] != ManagedByAccessPolicy || labels[PolicyNameLabel] != p.Name {
			return fmt.Errorf("refusing to change %s %s/%s owned by another controller", d.GVK.Kind, d.Namespace, d.Name)
		}
		if d.Delete {
			uid, rv := existing.GetUID(), existing.GetResourceVersion()
			if err := c.Delete(ctx, existing, client.Preconditions{UID: &uid, ResourceVersion: &rv}); err != nil && !apierrors.IsNotFound(err) {
				return err
			}
			continue
		}
		d.Object.SetResourceVersion(existing.GetResourceVersion())
		if err := c.Update(ctx, d.Object); err != nil {
			return err
		}
	}
	return nil
}
