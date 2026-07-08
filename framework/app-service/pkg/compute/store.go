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
	// Retry on AlreadyExists in addition to Conflict: the first allocation on a
	// fresh cluster creates the config map, and two concurrent first-time
	// allocations both observe NotFound and both Create. The loser gets
	// AlreadyExists; retrying re-reads the now-existing config map and falls
	// through to the optimistic-locked Update path, so its allocation is merged
	// instead of being silently dropped.
	retriable := func(err error) bool {
		return apierrors.IsConflict(err) || apierrors.IsAlreadyExists(err)
	}
	err := retry.OnError(retry.DefaultRetry, retriable, func() error {
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

// buildNodeResource builds the compute view of a single physical node. A node
// may advertise several accelerator modes at once (e.g. an Olares One exposing
// both nvidia and intel): GPUTypes lists them all and Devices holds the devices
// for every mode, each tagged with the Mode it serves. The scheduler /
// availability code projects this down to a single mode via Node.viewForMode
// when it needs a per-mode view. A node with no GPU mode is represented with a
// single cpu memory-shared device, preserving the previous behavior for
// pure-CPU and unlabeled nodes.
func buildNodeResource(node *corev1.Node) Node {
	totalMemory := node.Status.Capacity.Memory().Value()
	modes := utils.NodeSupportedGPUTypes(node)

	n := Node{
		NodeName:       node.Name,
		GPUTypes:       modes,
		Health:         nodeHealth(node),
		memoryCapacity: totalMemory,
	}

	if len(modes) == 0 {
		n.Devices = []Device{nonHAMIDevice(node, utils.CPUType, totalMemory)}
		return n
	}

	for _, mode := range modes {
		if IsHAMIMode(mode) {
			n.Devices = append(n.Devices, decodeHAMINvidiaDevices(node, mode)...)
			continue
		}
		n.Devices = append(n.Devices, nonHAMIDevice(node, mode, totalMemory))
	}

	return n
}

// nonHAMIDevice builds the single synthetic device used for unified-memory
// modes (cpu, apple-m, amd, intel, moore-soc, …): the whole node is one
// schedulable unit drawing from system memory. The support type starts at the
// mode's default (defaultSupportType) but, for the modes that can switch (intel
// / amd / moore-soc), honors the per-device share-mode annotation written by
// SwitchDeviceMode so a switch to Exclusive survives a rebuild of this view.
// cpu / apple-m / discrete-GPU modes only have a single available support type,
// so the annotation (if any) is clamped back to their default.
func nonHAMIDevice(node *corev1.Node, mode string, totalMemory int64) Device {
	deviceID := fmt.Sprintf("%s-%s-0", node.Name, mode)
	return Device{
		ID:                    deviceID,
		NodeName:              node.Name,
		Mode:                  mode,
		Memory:                totalMemory * 75 / 100,
		Health:                nodeHealth(node),
		SupportType:           nonHAMISupportType(mode, node.Annotations[shareModeAnnotationKey(deviceID)]),
		AvailableSupportTypes: AvailableSupportTypes(mode),
	}
}

// nonHAMISupportType resolves a non-HAMI device's support type from its
// share-mode annotation: honor the annotation only when it names a support type
// the mode actually allows, otherwise fall back to the mode's default. This
// keeps device.SupportType always within AvailableSupportTypes(mode), so a mode
// with a single available type (cpu / apple-m / discrete GPUs) stays pinned to
// it even if a stale annotation is left behind.
func nonHAMISupportType(mode, shareMode string) string {
	for _, supportType := range AvailableSupportTypes(mode) {
		if code, _ := supportTypeToShareMode(supportType); code == shareMode {
			return supportType
		}
	}
	return defaultSupportType(mode)
}

func decodeHAMINvidiaDevices(node *corev1.Node, mode string) []Device {
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
		devices = append(devices, Device{
			ID:                    items[0],
			NodeName:              node.Name,
			Mode:                  mode,
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

// shareModeToSupportType maps a device's HAMI share-mode annotation to a
// support type. An unset / unrecognized annotation falls back to the mode's
// default (defaultSupportType): nvidia → TimeSlice, nvidia-gb10 → MemorySlice.
func shareModeToSupportType(gpuType, shareMode string) string {
	switch shareMode {
	case "0":
		return SupportTypeExclusive
	case "1":
		return SupportTypeMemorySlice
	default:
		return defaultSupportType(gpuType)
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

// AvailableSupportTypes lists the support types a device of the given mode may
// take. The first entry is the mode's default — the one assigned when no share
// mode is configured (see defaultSupportType). nvidia can switch among all
// three; nvidia-gb10 defaults to MemorySlice but may switch to Exclusive; the
// unified-memory integrated accelerators (intel / amd / moore-soc) default to
// MemorySlice but may also switch to Exclusive; cpu is MemorySlice-only (it has
// no dedicated device to monopolize); apple-m and any discrete-GPU (intel-gpu /
// amd-gpu) or future mode are Exclusive-only for now.
func AvailableSupportTypes(mode string) []string {
	switch mode {
	case utils.NvidiaCardType:
		return []string{SupportTypeTimeSlice, SupportTypeMemorySlice, SupportTypeExclusive}
	case utils.GB10ChipType, utils.IntelType, utils.AMDType, utils.MooreSocType:
		return []string{SupportTypeMemorySlice, SupportTypeExclusive}
	case utils.CPUType:
		return []string{SupportTypeMemorySlice}
	default:
		return []string{SupportTypeExclusive}
	}
}

// defaultSupportType is the support type a device of the given mode receives
// when its share mode has not been explicitly configured. It is, by contract,
// the first entry of AvailableSupportTypes(mode): nvidia → TimeSlice,
// nvidia-gb10 and the unified-memory modes (cpu / intel / amd / moore-soc) →
// MemorySlice, apple-m and any discrete-GPU / future mode → Exclusive.
func defaultSupportType(mode string) string {
	return AvailableSupportTypes(mode)[0]
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

// AttachBoundAppSpecs enriches every device binding with the bound app's
// currently-selected resource-mode requirement (Allocation.Spec), so callers
// listing compute resources can see each app's require/limit and multi-card /
// multi-node support without a second round of lookups. The requirement is
// resolved once per unique (appName, owner) pair from the app's
// ApplicationManager config and shared across that app's bindings. If any
// bound app's config can't be loaded or its selected mode can't be resolved,
// the whole listing fails with an error.
func AttachBoundAppSpecs(ctx context.Context, c client.Client, nodes []Node) error {
	specs := make(map[string]*Requirement)
	resolve := func(appName, owner string) (*Requirement, error) {
		key := owner + "/" + appName
		if spec, ok := specs[key]; ok {
			return spec, nil
		}
		cfg, err := loadAppConfigForOwner(ctx, c, appName, owner)
		if err != nil {
			return nil, fmt.Errorf("failed to load app config for %s/%s: %w", owner, appName, err)
		}
		req, ok := SelectedRequirement(cfg)
		if !ok {
			return nil, fmt.Errorf("failed to resolve selected resource mode for %s/%s", owner, appName)
		}
		specs[key] = &req
		return specs[key], nil
	}
	for ni := range nodes {
		for di := range nodes[ni].Devices {
			bindings := nodes[ni].Devices[di].Bindings
			for bi := range bindings {
				spec, err := resolve(bindings[bi].AppName, bindings[bi].Owner)
				if err != nil {
					return err
				}
				bindings[bi].Spec = spec
			}
		}
	}
	return nil
}
