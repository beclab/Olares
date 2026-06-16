package compute

import (
	"context"
	"fmt"

	corev1 "k8s.io/api/core/v1"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

func FindDevice(ctx context.Context, c client.Client, nodeName, deviceID string) (Node, Device, error) {
	nodes, err := FetchNodeComputeAllocations(ctx, c)
	if err != nil {
		return Node{}, Device{}, err
	}
	for _, node := range nodes {
		if node.NodeName != nodeName {
			continue
		}
		for _, device := range node.Devices {
			if device.ID == deviceID {
				return node, device, nil
			}
		}
		return Node{}, Device{}, fmt.Errorf("device %s not found on node %s", deviceID, nodeName)
	}
	return Node{}, Device{}, fmt.Errorf("node %s not found", nodeName)
}

// UpdateDeviceSupportType validates and switches a device's share-mode
// annotation. It refuses to switch when the device still has compute bindings;
// callers that have already torn the bindings down should use SwitchDeviceMode
// directly.
func UpdateDeviceSupportType(ctx context.Context, c client.Client, nodeName, deviceID, supportType string) (Device, error) {
	_, device, err := FindDevice(ctx, c, nodeName, deviceID)
	if err != nil {
		return Device{}, err
	}
	if !IsHAMIMode(device.Mode) {
		return Device{}, fmt.Errorf("device mode switching is not supported for gpu type %s", device.Mode)
	}
	if !SupportTypeAvailable(device.AvailableSupportTypes, supportType) {
		return Device{}, fmt.Errorf("support type %s is not available for gpu type %s", supportType, device.Mode)
	}
	if len(device.Bindings) > 0 {
		return Device{}, fmt.Errorf("device %s still has %d compute bindings", deviceID, len(device.Bindings))
	}
	if err := SwitchDeviceMode(ctx, c, nodeName, deviceID, supportType); err != nil {
		return Device{}, err
	}
	device.SupportType = supportType
	return device, nil
}

// SwitchDeviceMode flips the share-mode annotation on the underlying k8s node
// without re-validating bindings/support types. Callers are responsible for
// ensuring the device is no longer in use (typically by stopping bound apps
// and clearing their allocations first).
func SwitchDeviceMode(ctx context.Context, c client.Client, nodeName, deviceID, supportType string) error {
	shareMode, ok := supportTypeToShareMode(supportType)
	if !ok {
		return fmt.Errorf("unsupported support type %s", supportType)
	}
	var k8sNode corev1.Node
	if err := c.Get(ctx, client.ObjectKey{Name: nodeName}, &k8sNode); err != nil {
		return err
	}
	if k8sNode.Annotations == nil {
		k8sNode.Annotations = make(map[string]string)
	}
	k8sNode.Annotations[shareModeAnnotationKey(deviceID)] = shareMode
	return c.Update(ctx, &k8sNode)
}

func SupportTypeAvailable(items []string, target string) bool {
	for _, item := range items {
		if item == target {
			return true
		}
	}
	return false
}
