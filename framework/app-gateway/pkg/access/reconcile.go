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
		if d.Delete {
			obj := d.Object.DeepCopy()
			obj.SetGroupVersionKind(d.GVK)
			obj.SetName(d.Name)
			obj.SetNamespace(d.Namespace)
			if err := c.Delete(ctx, obj); err != nil && !apierrors.IsNotFound(err) {
				return err
			}
			continue
		}
		existing := &unstructured.Unstructured{}
		existing.SetGroupVersionKind(d.GVK)
		err := c.Get(ctx, types.NamespacedName{Namespace: d.Namespace, Name: d.Name}, existing)
		if apierrors.IsNotFound(err) {
			if err := c.Create(ctx, d.Object); err != nil {
				return err
			}
			continue
		}
		if err != nil {
			return err
		}
		d.Object.SetResourceVersion(existing.GetResourceVersion())
		if err := c.Update(ctx, d.Object); err != nil {
			return err
		}
	}
	return nil
}
