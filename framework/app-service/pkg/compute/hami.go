package compute

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"strings"

	"github.com/beclab/Olares/framework/app-service/pkg/constants"
	apierrors "k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/api/meta"
	"k8s.io/apimachinery/pkg/api/resource"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/klog/v2"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

func createHAMIBinding(ctx context.Context, c client.Client, allocation Allocation) error {
	if !IsHAMIMode(allocation.Mode) {
		return nil
	}
	name := hamiBindingName(allocation)
	obj := &unstructured.Unstructured{}
	obj.SetAPIVersion(hamiGPUBindingAPIVersion)
	obj.SetKind(hamiGPUBindingKind)
	obj.SetName(name)
	obj.SetLabels(map[string]string{
		managedByLabelKey:               managedByAppService,
		constants.ApplicationNameLabel:  allocation.AppName,
		constants.ApplicationOwnerLabel: allocation.Owner,
		allocationModeLabelKey:          allocation.Mode,
	})
	spec := map[string]any{
		"uuid":    allocation.DeviceID,
		"appName": allocation.AppName,
		"owner":   allocation.Owner,
		"podSelector": map[string]any{
			"matchLabels": map[string]any{
				constants.ApplicationNameLabel:  allocation.AppName,
				constants.ApplicationOwnerLabel: allocation.Owner,
			},
		},
	}
	// Omit spec.memory when the allocation does not partition memory
	// (Exclusive / TimeSlice — buildAllocation persists Memory=0 for
	// these). Leaving Spec.Memory nil keeps HAMI from emitting the
	// per-pod gpumem annotation, so the pod is not artificially capped
	// below the device's full capacity.
	if allocation.Memory > 0 {
		mem := resource.NewQuantity(allocation.Memory/mib, resource.DecimalSI)
		spec["memory"] = mem.String()
	}
	obj.Object["spec"] = spec

	if err := c.Create(ctx, obj); err != nil {
		if apierrors.IsAlreadyExists(err) {
			current := &unstructured.Unstructured{}
			current.SetAPIVersion(hamiGPUBindingAPIVersion)
			current.SetKind(hamiGPUBindingKind)
			if getErr := c.Get(ctx, client.ObjectKey{Name: name}, current); getErr != nil {
				return getErr
			}
			current.SetLabels(obj.GetLabels())
			current.Object["spec"] = obj.Object["spec"]
			return c.Update(ctx, current)
		}
		return err
	}
	return nil
}

func hamiBindingName(allocation Allocation) string {
	raw := fmt.Sprintf("%s-%s-%s", allocation.AppName, allocation.Owner, allocation.DeviceID)
	sum := sha256.Sum256([]byte(raw))
	suffix := "-" + hex.EncodeToString(sum[:])[:8]

	var b strings.Builder
	for _, r := range strings.ToLower(raw) {
		switch {
		case r >= 'a' && r <= 'z', r >= '0' && r <= '9', r == '-', r == '.':
			b.WriteRune(r)
		default:
			b.WriteRune('-')
		}
	}
	out := strings.Trim(b.String(), "-.")
	if maxOut := 63 - len(suffix); len(out) > maxOut {
		out = strings.Trim(out[:maxOut], "-.")
	}
	if out == "" {
		return "gpu-binding"
	}
	return out + suffix
}

func deleteHAMIBindingsForApp(ctx context.Context, c client.Client, appName, owner string) error {
	list := &unstructured.UnstructuredList{}
	list.SetAPIVersion(hamiGPUBindingAPIVersion)
	list.SetKind(hamiGPUBindingListKind)
	selector := client.MatchingLabels{
		managedByLabelKey:               managedByAppService,
		constants.ApplicationNameLabel:  appName,
		constants.ApplicationOwnerLabel: owner,
	}
	if err := c.List(ctx, list, selector); err != nil {
		if meta.IsNoMatchError(err) || apierrors.IsNotFound(err) {
			return nil
		}
		return err
	}
	for i := range list.Items {
		item := &list.Items[i]
		if err := c.Delete(ctx, item); err != nil && !apierrors.IsNotFound(err) {
			klog.Warningf("failed to delete GPUBinding %s for app %s owned by %s: %v", item.GetName(), appName, owner, err)
		}
	}
	return nil
}
