package deviceinventory

// Annotation is the Node annotation populated by generic-device-plugin.
const Annotation = "bytetrade.io/devices"

// Item is one logical platform device. Node, Type and Reason are populated by
// app-service when it presents the inventory to callers; the remaining fields
// are published by generic-device-plugin.
type Item struct {
	UID          string   `json:"uid"`
	ResourceName string   `json:"resourceName"`
	DisplayName  string   `json:"displayName,omitempty"`
	Online       bool     `json:"online"`
	Selectable   bool     `json:"selectable"`
	KernelPaths  []string `json:"kernelPaths,omitempty"`
	PortUnstable bool     `json:"portUnstable,omitempty"`
	Adapter      string   `json:"adapter,omitempty"`
	Address      string   `json:"address,omitempty"`
	AccessMode   string   `json:"accessMode,omitempty"`
	Node         string   `json:"node"`
	Type         string   `json:"type"`
	Reason       string   `json:"reason,omitempty"`
}

// Inventory is schemaVersion 1 of the generic-device-plugin Node annotation.
type Inventory struct {
	SchemaVersion int    `json:"schemaVersion"`
	Serial        []Item `json:"serial,omitempty"`
	Video         []Item `json:"video,omitempty"`
	Audio         []Item `json:"audio,omitempty"`
	HID           []Item `json:"hid,omitempty"`
	Bluetooth     []Item `json:"bluetooth,omitempty"`
	Truncated     bool   `json:"truncated,omitempty"`
}
