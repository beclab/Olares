package compute

import (
	"context"
	"encoding/json"
	"fmt"
	"strconv"
	"strings"

	"github.com/beclab/Olares/framework/app-service/pkg/constants"
	"github.com/beclab/Olares/framework/app-service/pkg/utils"
	corev1 "k8s.io/api/core/v1"
	apierrors "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/types"
	"k8s.io/client-go/util/retry"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

const mib = int64(1024 * 1024)

type allocationMutation func(nodes []Node, allocations []Allocation) ([]Allocation, *Allocation, error)

func FetchNodeComputeAllocations(ctx context.Context, c client.Client) ([]Node, error) {
	nodes, err := loadNodeResources(ctx, c)
	if err != nil {
		return nil, err
	}
	allocations, err := loadAllocations(ctx, c)
	if err != nil {
		return nil, err
	}
	attachBindings(nodes, allocations)
	return nodes, nil
}

// FetchNodeComputeAllocationsExcludingApp returns the node view with the given
// app's own allocation rows excluded, so a re-allocation / re-binding of that
// app does not count its current claim against availability (an exclusive card
// the app already holds would otherwise report zero free memory). Callers that
// persist via replaceAppAllocations drop those rows anyway, so excluding them
// here keeps the availability view consistent with the post-write state.
func FetchNodeComputeAllocationsExcludingApp(ctx context.Context, c client.Client, appName, owner string) ([]Node, error) {
	nodes, err := loadNodeResources(ctx, c)
	if err != nil {
		return nil, err
	}
	allocations, err := loadAllocations(ctx, c)
	if err != nil {
		return nil, err
	}
	attachBindings(nodes, withoutAppAllocations(allocations, appName, owner))
	return nodes, nil
}

// withoutAppAllocations returns allocations excluding the rows owned by the
// given (appName, owner) pair.
func withoutAppAllocations(allocations []Allocation, appName, owner string) []Allocation {
	out := make([]Allocation, 0, len(allocations))
	for _, allocation := range allocations {
		if allocation.AppName == appName && allocation.Owner == owner {
			continue
		}
		out = append(out, allocation)
	}
	return out
}

func loadAllocations(ctx context.Context, c client.Client) ([]Allocation, error) {
	_, allocations, err := loadAllocationConfigMap(ctx, c)
	return allocations, err
}

func DeleteAllocationsForApp(ctx context.Context, c client.Client, appName, owner string) error {
	_, err := mutateAllocations(ctx, c, func(_ []Node, allocations []Allocation) ([]Allocation, *Allocation, error) {
		next := make([]Allocation, 0, len(allocations))
		for _, allocation := range allocations {
			if allocation.AppName == appName && allocation.Owner == owner {
				continue
			}
			next = append(next, allocation)
		}
		return next, nil, nil
	})
	if err != nil {
		return err
	}
	return deleteHAMIBindingsForApp(ctx, c, appName, owner)
}

func FindAllocationsForApp(ctx context.Context, c client.Client, appName, owner string) ([]Allocation, error) {
	allocations, err := loadAllocations(ctx, c)
	if err != nil {
		return nil, err
	}
	out := make([]Allocation, 0)
	for _, allocation := range allocations {
		if allocation.AppName == appName && allocation.Owner == owner {
			out = append(out, allocation)
		}
	}
	return out, nil
}

func mutateAllocations(ctx context.Context, c client.Client, mutation allocationMutation) (*Allocation, error) {
	var selected *Allocation
	err := retry.RetryOnConflict(retry.DefaultRetry, func() error {
		nodes, err := loadNodeResources(ctx, c)
		if err != nil {
			return err
		}
		cm, allocations, err := loadAllocationConfigMap(ctx, c)
		if err != nil {
			return err
		}
		// Bindings are intentionally not attached here: app-specific mutations
		// must decide whether to count an app's own existing allocation against
		// availability (re-allocation excludes it), so they attach the node view
		// themselves. Mutations that ignore nodes (e.g. delete) are unaffected.

		next, allocation, err := mutation(nodes, allocations)
		if err != nil {
			return err
		}
		data, err := json.Marshal(next)
		if err != nil {
			return err
		}
		if cm == nil {
			if len(next) == 0 {
				selected = allocation
				return nil
			}
			cm = &corev1.ConfigMap{
				ObjectMeta: metav1.ObjectMeta{
					Namespace: allocationConfigMapNamespace,
					Name:      allocationConfigMapName,
				},
				Data: map[string]string{allocationConfigMapKey: string(data)},
			}
			if err := c.Create(ctx, cm); err != nil {
				return err
			}
			selected = allocation
			return nil
		}
		if cm.Data == nil {
			cm.Data = make(map[string]string)
		}
		cm.Data[allocationConfigMapKey] = string(data)
		if err := c.Update(ctx, cm); err != nil {
			return err
		}
		selected = allocation
		return nil
	})
	return selected, err
}

func loadAllocationConfigMap(ctx context.Context, c client.Client) (*corev1.ConfigMap, []Allocation, error) {
	var cm corev1.ConfigMap
	err := c.Get(ctx, types.NamespacedName{
		Namespace: allocationConfigMapNamespace,
		Name:      allocationConfigMapName,
	}, &cm)
	if apierrors.IsNotFound(err) {
		return nil, nil, nil
	}
	if err != nil {
		return nil, nil, err
	}
	raw := cm.Data[allocationConfigMapKey]
	if raw == "" {
		return &cm, nil, nil
	}
	var allocations []Allocation
	if err := json.Unmarshal([]byte(raw), &allocations); err != nil {
		return nil, nil, err
	}
	return &cm, allocations, nil
}

func loadNodeResources(ctx context.Context, c client.Client) ([]Node, error) {
	var nodes corev1.NodeList
	if err := c.List(ctx, &nodes); err != nil {
		return nil, err
	}
	out := make([]Node, 0, len(nodes.Items))
	for i := range nodes.Items {
		out = append(out, buildNodeResource(&nodes.Items[i]))
	}
	return out, nil
}

func buildNodeResource(node *corev1.Node) Node {
	totalMemory := node.Status.Capacity.Memory().Value()
	mode := utils.NodeGPUType(node)

	n := Node{
		NodeName:       node.Name,
		GPUType:        mode,
		Health:         nodeHealth(node),
		memoryCapacity: totalMemory,
	}

	if IsHAMIMode(mode) {
		n.Devices = decodeHAMINvidiaDevices(node)
	} else {
		supportType := SupportTypeExclusive
		if mode == utils.CPUType {
			supportType = SupportTypeMemoryShared
		}
		n.Devices = []Device{{
			ID:                    fmt.Sprintf("%s-%s-0", node.Name, mode),
			NodeName:              node.Name,
			Memory:                totalMemory * 75 / 100,
			Health:                nodeHealth(node),
			SupportType:           supportType,
			AvailableSupportTypes: AvailableSupportTypes(mode),
		}}
	}

	return n
}

func decodeHAMINvidiaDevices(node *corev1.Node) []Device {
	raw := node.Annotations[constants.NodeNvidiaRegistryKey]
	if !strings.Contains(raw, constants.OneContainerMultiDeviceSplitSymbol) {
		return nil
	}

	var devices []Device
	for _, encoded := range strings.Split(raw, constants.OneContainerMultiDeviceSplitSymbol) {
		if encoded == "" || !strings.Contains(encoded, ",") {
			continue
		}
		items := strings.Split(encoded, ",")
		if len(items) != 7 && len(items) != 9 && len(items) != 10 {
			continue
		}
		devmem, _ := strconv.ParseInt(items[2], 10, 64)
		healthy, _ := strconv.ParseBool(items[6])
		mode := utils.NodeGPUType(node)
		devices = append(devices, Device{
			ID:                    items[0],
			NodeName:              node.Name,
			CardModel:             items[4],
			Memory:                devmem * mib,
			Health:                boolHealth(healthy),
			SupportType:           shareModeToSupportType(mode, node.Annotations[shareModeAnnotationKey(items[0])]),
			AvailableSupportTypes: AvailableSupportTypes(mode),
		})
	}
	return devices
}

func nodeHealth(node *corev1.Node) string {
	for _, condition := range node.Status.Conditions {
		if condition.Type == corev1.NodeReady && condition.Status == corev1.ConditionTrue {
			return deviceHealthYes
		}
	}
	return deviceHealthNo
}

func boolHealth(healthy bool) string {
	if healthy {
		return deviceHealthYes
	}
	return deviceHealthNo
}

func shareModeToSupportType(gpuType, shareMode string) string {
	switch shareMode {
	case "0":
		return SupportTypeExclusive
	case "1":
		return SupportTypeMemorySlice
	default:
		if gpuType == utils.GB10ChipType {
			return SupportTypeMemorySlice
		}
		return SupportTypeTimeSlice
	}
}

func supportTypeToShareMode(supportType string) (string, bool) {
	switch supportType {
	case SupportTypeExclusive:
		return "0", true
	case SupportTypeMemorySlice:
		return "1", true
	case SupportTypeTimeSlice:
		return "2", true
	default:
		return "", false
	}
}

func shareModeAnnotationKey(deviceID string) string {
	return fmt.Sprintf("sharemode.gpu.bytetrade.io/%s", deviceID)
}

func AvailableSupportTypes(mode string) []string {
	switch mode {
	case utils.NvidiaCardType:
		return []string{SupportTypeTimeSlice, SupportTypeMemorySlice, SupportTypeExclusive}
	case utils.GB10ChipType:
		return []string{SupportTypeMemorySlice, SupportTypeExclusive}
	case utils.CPUType:
		return []string{SupportTypeMemoryShared}
	default:
		return []string{SupportTypeExclusive}
	}
}

func attachBindings(nodes []Node, allocations []Allocation) {
	for ai := range allocations {
		for ni := range nodes {
			if nodes[ni].NodeName != allocations[ai].NodeName {
				continue
			}
			for di := range nodes[ni].Devices {
				if nodes[ni].Devices[di].ID == allocations[ai].DeviceID {
					nodes[ni].Devices[di].Bindings = append(nodes[ni].Devices[di].Bindings, allocations[ai])
				}
			}
		}
	}
}
