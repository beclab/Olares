package compute

import (
	"context"
	"encoding/json"
	"fmt"
	"strconv"
	"strings"
	"time"

	"github.com/beclab/Olares/framework/app-service/pkg/constants"
	"github.com/beclab/Olares/framework/app-service/pkg/utils"
	corev1 "k8s.io/api/core/v1"
	apierrors "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/types"
	"k8s.io/client-go/util/retry"
	"k8s.io/klog/v2"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

const mib = int64(1024 * 1024)

const (
	// hamiHandshake* mirror the values HAMi writes into its handshake
	// annotation and the window it lets a request go unanswered before
	// declaring the node's devices gone (HAMi pkg/util.CheckHealth). They have
	// to stay in step with HAMi for the two to agree about a node.
	hamiHandshakeRequesting = "Requesting"
	hamiHandshakeDeleted    = "Deleted"
	hamiHandshakeTimeout    = 60 * time.Second
)

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

func FetchSchedulableNodeComputeAllocations(ctx context.Context, c client.Client) ([]Node, error) {
	nodes, err := loadSchedulableNodeResources(ctx, c)
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

func loadSchedulableNodeResources(ctx context.Context, c client.Client) ([]Node, error) {
	var nodes corev1.NodeList
	if err := c.List(ctx, &nodes); err != nil {
		return nil, err
	}
	out := make([]Node, 0, len(nodes.Items))
	for i := range nodes.Items {
		if !utils.IsNodeReady(&nodes.Items[i]) || nodes.Items[i].Spec.Unschedulable {
			continue
		}
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
		register := CapabilityFor(mode).Register
		switch register.Decoder {
		case registerHAMI:
			n.Devices = append(n.Devices, decodeHAMIDevices(node, mode)...)
		case registerTuple:
			// Fall back to the single system-memory device when the register
			// carries no usable discrete entry, so a node that advertises the
			// mode still has a schedulable device.
			if devices := decodeTupleRegisterDevices(node, mode, register.Annotation, totalMemory); len(devices) > 0 {
				n.Devices = append(n.Devices, devices...)
			} else {
				n.Devices = append(n.Devices, nonHAMIDevice(node, mode, totalMemory))
			}
		default:
			n.Devices = append(n.Devices, nonHAMIDevice(node, mode, totalMemory))
		}
	}

	return n
}

// registeredCard is one card as a device register publishes it. SupportTypes
// narrows the mode's support types for this card (see cardSupportTypes); it
// stays empty, and the card offers its mode's full list, unless the mode's
// register names a Cards annotation listing the card.
type registeredCard struct {
	ID           string
	CardModel    string
	Memory       int64
	Health       string
	SupportTypes []string
}

func registeredDevice(node *corev1.Node, mode string, card registeredCard) Device {
	available := cardSupportTypes(mode, card.SupportTypes)
	return Device{
		ID:                    card.ID,
		NodeName:              node.Name,
		Mode:                  mode,
		CardModel:             card.CardModel,
		Memory:                card.Memory,
		Health:                card.Health,
		SupportType:           resolveSupportType(available, node.Annotations[shareModeAnnotationKey(card.ID)]),
		AvailableSupportTypes: available,
	}
}

// decodeTupleRegisterDevices builds one device per discrete GPU (dgpu) in a
// tuple register annotation (bytetrade.io/node-intel-register,
// bytetrade.io/node-amd-register). Each card's real VRAM (the entry's mem
// field, in bytes) is used as Memory instead of the node's system RAM.
// Integrated GPUs (igpu) are unified memory and stay on the nonHAMIDevice
// path, so they are skipped here. When a discrete entry's VRAM is unknown
// (mem=0) it falls back to the node's usable system memory so the card remains
// schedulable. Returns nil when the annotation is missing, malformed, or
// advertises no discrete card, letting the caller use the legacy device.
func decodeTupleRegisterDevices(node *corev1.Node, mode, annotation string, totalMemory int64) []Device {
	entries, err := utils.GPURegisterFromNode(node, annotation)
	if err != nil {
		klog.Warningf("compute: node %s has malformed %s annotation: %v", node.Name, annotation, err)
		return nil
	}

	var devices []Device
	for _, e := range entries {
		if e.Kind != utils.IntelGPUKindDiscrete {
			continue
		}
		memory := int64(e.Mem)
		if memory <= 0 {
			memory = totalMemory * 75 / 100
		}
		devices = append(devices, registeredDevice(node, mode, registeredCard{
			ID:        fmt.Sprintf("%s-%s-%s", node.Name, mode, e.Card),
			CardModel: e.Name,
			Memory:    memory,
			Health:    nodeHealth(node),
		}))
	}
	return devices
}

// nonHAMIDevice builds the single synthetic device used for unified-memory
// modes (cpu, apple-m, amd, intel, moore-soc, …): the whole node is one
// schedulable unit drawing from system memory. The support type starts at the
// mode's default but, for the modes that can switch,
// honors the per-device share-mode annotation written by SwitchDeviceMode so a
// switch to Exclusive survives a rebuild of this view. A mode with a single
// available support type stays pinned to it even if a stale annotation is left
// behind.
func nonHAMIDevice(node *corev1.Node, mode string, totalMemory int64) Device {
	deviceID := fmt.Sprintf("%s-%s-0", node.Name, mode)
	available := AvailableSupportTypes(mode)
	return Device{
		ID:                    deviceID,
		NodeName:              node.Name,
		Mode:                  mode,
		Memory:                totalMemory * 75 / 100,
		Health:                nodeHealth(node),
		SupportType:           resolveSupportType(available, node.Annotations[shareModeAnnotationKey(deviceID)]),
		AvailableSupportTypes: available,
	}
}

func decodeHAMIDevices(node *corev1.Node, mode string) []Device {
	register := CapabilityFor(mode).Register
	annotation := register.Annotation
	raw := node.Annotations[annotation]
	if !strings.Contains(raw, constants.OneContainerMultiDeviceSplitSymbol) {
		// The node carries the GPU label but the device plugin has not
		// registered anything usable, so it will appear with zero cards. Common
		// and transient while the plugin starts; persistent means the plugin is
		// not running there.
		klog.Warningf("compute: node %s advertises %s but its %s annotation is missing or unusable, so it exposes no card",
			node.Name, mode, annotation)
		return nil
	}
	// A card's own health bit is only as fresh as the annotation carrying it,
	// so it has to be read together with the node-level signals.
	nodeUsable := hamiNodeHealth(node, register.Handshake) == deviceHealthYes
	if !nodeUsable {
		klog.V(2).Infof("compute: every card on node %s is marked unhealthy by node-level signals: nodeReady=%s %s=%q",
			node.Name, nodeHealth(node), register.Handshake, node.Annotations[register.Handshake])
	}
	cardSupport := decodeRegisteredCardInfo(node, register.Cards)

	var devices []Device
	for _, encoded := range strings.Split(raw, constants.OneContainerMultiDeviceSplitSymbol) {
		if encoded == "" || !strings.Contains(encoded, ",") {
			continue
		}
		items := strings.Split(encoded, ",")
		if len(items) != 7 && len(items) != 9 && len(items) != 10 {
			klog.Warningf("compute: skipping a card on node %s: its %s entry has %d fields, expected 7, 9 or 10",
				node.Name, annotation, len(items))
			continue
		}
		devmem, _ := strconv.ParseInt(items[2], 10, 64)
		healthy, _ := strconv.ParseBool(items[6])
		devices = append(devices, registeredDevice(node, mode, registeredCard{
			ID:           items[0],
			CardModel:    items[4],
			Memory:       devmem * mib,
			Health:       boolHealth(healthy && nodeUsable),
			SupportTypes: cardSupport[items[0]],
		}))
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

// hamiNodeHealth reports whether the cards HAMi registered on a node can still
// be trusted. decodeHAMIDevices folds it into every card's own health
// bit, because that bit cannot express the node being gone: it is parsed out of
// the register annotation, which the device plugin writes while the node is up
// and which nothing ever clears — not even HAMi's own cleanup path, which marks
// the handshake Deleted and leaves the stale card list in place (NodeCleanUp ->
// MarkAnnotationsToDelete). Read on its own, a powered-off node therefore keeps
// offering healthy cards indefinitely, and an app bound to one is placed on a
// node HAMi has already dropped from its scheduler, leaving the pod Pending.
//
// The two signals combined here cover different failures and neither subsumes
// the other: NodeReady catches a node that is gone, the handshake catches a
// node that is up but whose GPU stack has died.
func hamiNodeHealth(node *corev1.Node, handshake string) string {
	if nodeHealth(node) != deviceHealthYes {
		return deviceHealthNo
	}
	return hamiHandshakeHealth(node.Annotations[handshake])
}

// registeredCardInfo is one entry of a register's Cards annotation.
type registeredCardInfo struct {
	UUID         string   `json:"uuid"`
	Card         string   `json:"card"`
	BDF          string   `json:"bdf"`
	SupportTypes []string `json:"supportTypes"`
}

// decodeRegisteredCardInfo maps card ID to the support types its Cards
// annotation lists. A missing or malformed annotation narrows nothing.
func decodeRegisteredCardInfo(node *corev1.Node, annotation string) map[string][]string {
	if annotation == "" || node.Annotations[annotation] == "" {
		return nil
	}
	var infos []registeredCardInfo
	if err := json.Unmarshal([]byte(node.Annotations[annotation]), &infos); err != nil {
		klog.Warningf("compute: node %s has malformed %s annotation: %v", node.Name, annotation, err)
		return nil
	}
	out := make(map[string][]string, len(infos))
	for _, info := range infos {
		out[info.UUID] = info.SupportTypes
	}
	return out
}

// hamiHandshakeHealth interprets HAMi's handshake annotation. The exchange it
// records: the device plugin republishes "Reported <time>" every 30s, HAMi's
// scheduler answers by stamping "Requesting_<time>" and then waits WITHOUT
// refreshing that stamp, so a request still unanswered after
// hamiHandshakeTimeout means no plugin is left to answer it. HAMi reacts by
// rewriting the value to "Deleted_<time>" and dropping the node.
//
// This mirrors HAMi's own util.CheckHealth so the two components cannot
// disagree about a node, with one deliberate difference: HAMi reports Deleted
// as healthy because there the branch exists only to stop it re-adding a node
// it already removed. Read as a statement about the cards, it means the exact
// opposite, so here it is unhealthy.
//
// An empty or unrecognized value is healthy, matching HAMi's own fallback, so a
// node registered by a plugin that predates the handshake is not taken out of
// service. Timestamps are compared in UTC, which is what both sides produce as
// long as neither container sets TZ.
func hamiHandshakeHealth(handshake string) string {
	switch {
	case strings.Contains(handshake, hamiHandshakeDeleted):
		return deviceHealthNo
	case strings.Contains(handshake, hamiHandshakeRequesting):
		_, stamp, ok := strings.Cut(handshake, "_")
		if !ok {
			return deviceHealthNo
		}
		requestedAt, err := time.Parse(time.DateTime, stamp)
		if err != nil {
			return deviceHealthNo
		}
		return boolHealth(time.Now().Before(requestedAt.Add(hamiHandshakeTimeout)))
	default:
		return deviceHealthYes
	}
}

func boolHealth(healthy bool) string {
	if healthy {
		return deviceHealthYes
	}
	return deviceHealthNo
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
// take, from the mode's capability row. The first entry is the default for a
// device whose share mode was never configured. A single card may offer fewer
// (cardSupportTypes).
func AvailableSupportTypes(mode string) []string {
	return CapabilityFor(mode).SupportTypes
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
	for ni := range nodes {
		for di := range nodes[ni].Devices {
			nodes[ni].Devices[di].Reserved = bindingsReserve(nodes[ni].Devices[di])
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
