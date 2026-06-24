package compute

// Wire types for the compute-resources surface. Field shapes mirror
// TermiPass src/constant/compute.ts (ComputeResourceNode / Device / Binding,
// ComputeSupportTypeResult) so the CLI decodes exactly what the SPA reads.

// computeSpec is the per-binding requirement/capability spec.
type computeSpec struct {
	SupportMultiCards bool `json:"supportMultiCards"`
	SupportMultiNodes bool `json:"supportMultiNodes"`
}

// computeBinding is a single app bound to a device.
type computeBinding struct {
	AppID    string       `json:"appId,omitempty"`
	AppName  string       `json:"appName"`
	Owner    string       `json:"owner,omitempty"`
	Mode     string       `json:"mode,omitempty"`
	NodeName string       `json:"nodeName,omitempty"`
	DeviceID string       `json:"deviceId"`
	Memory   int64        `json:"memory,omitempty"`
	Spec     *computeSpec `json:"spec,omitempty"`
}

// computeDevice is a GPU/accelerator card on a node.
type computeDevice struct {
	ID                    string           `json:"id"`
	NodeName              string           `json:"nodeName"`
	Name                  string           `json:"name,omitempty"`
	Mode                  string           `json:"mode,omitempty"`
	CardModel             string           `json:"cardModel,omitempty"`
	Memory                int64            `json:"memory"`
	Health                string           `json:"health,omitempty"`
	SupportType           string           `json:"supportType"`
	AvailableSupportTypes []string         `json:"availableSupportTypes,omitempty"`
	Bindings              []computeBinding `json:"bindings,omitempty"`
}

// computeNode is a node and all of its accelerator devices.
type computeNode struct {
	NodeName string          `json:"nodeName"`
	GpuTypes []string        `json:"gpuTypes,omitempty"`
	Health   string          `json:"health,omitempty"`
	Devices  []computeDevice `json:"devices,omitempty"`
}

// deviceDisplayName mirrors the SPA deviceName(): name || cardModel ||
// computeModeTitle(mode) || id.
func (d computeDevice) deviceDisplayName() string {
	if d.Name != "" {
		return d.Name
	}
	if d.CardModel != "" {
		return d.CardModel
	}
	if d.Mode != "" {
		if t := computeModeTitle(d.Mode); t != "" {
			return t
		}
	}
	return d.ID
}

// usedMemory mirrors deviceUsedMemory: sum of allocated binding memory,
// capped at the device capacity so used never exceeds total.
func (d computeDevice) usedMemory() int64 {
	var used int64
	for _, b := range d.Bindings {
		used += b.Memory
	}
	if d.Memory > 0 && used > d.Memory {
		return d.Memory
	}
	return used
}

// effectiveUsedMemory mirrors effectiveDeviceUsedMemory: an Exclusive device
// with a bound app reads as fully used (the whole card is dedicated);
// otherwise it falls back to the summed binding usage.
func (d computeDevice) effectiveUsedMemory() int64 {
	if isExclusiveSupportType(d.SupportType) && len(d.Bindings) > 0 {
		return d.Memory
	}
	return d.usedMemory()
}

// computeAppRef is an app reference inside a support-type switch result.
type computeAppRef struct {
	AppName string `json:"appName"`
	Owner   string `json:"owner,omitempty"`
	State   string `json:"state,omitempty"`
	Reason  string `json:"reason,omitempty"`
}

// supportTypeResult mirrors ComputeSupportTypeResult. `status` discriminates
// the outcome regardless of the upstream business code (switched / unchanged /
// bound-apps-stop-blocked).
type supportTypeResult struct {
	Status      string          `json:"status"`
	Device      *computeDevice  `json:"device,omitempty"`
	StoppedApps []computeAppRef `json:"stoppedApps,omitempty"`
	BlockedApps []computeAppRef `json:"blockedApps,omitempty"`
}

// supportTypeEnvelope captures the discriminated wire shape used by the
// support-type switch. The inner payload may arrive either directly as a
// supportTypeResult, or wrapped as {type: 'computeDeviceSwitchBlocked', Data: {...}}.
type supportTypeEnvelope struct {
	Type string             `json:"type"`
	Data *supportTypeResult `json:"Data"`
}
