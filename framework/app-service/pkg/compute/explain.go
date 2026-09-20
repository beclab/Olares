package compute

import (
	"fmt"
	"strings"

	"github.com/beclab/Olares/framework/app-service/pkg/utils"
	"k8s.io/apimachinery/pkg/api/resource"
)

// Placement works by elimination: a node that stopped heartbeating never
// reaches the classifier, a card that does not fit is passed over by the
// picker, and what is left for the user to see is an app that came up without a
// GPU. None of the return values along that path carry which candidates existed
// or why each was dropped, so these helpers render that for the log — enough to
// settle "why did this app not get a card?" from the log alone, without
// reproducing the cluster state.

func describeRequirement(req Requirement) string {
	if req.Mode == utils.NvidiaCardType {
		return fmt.Sprintf("mode=%s requireVram=%s limitVram=%s multiCards=%t multiNodes=%t",
			req.Mode, humanBytes(req.RequiredGPU), humanBytes(req.LimitedGPU),
			req.SupportMultiCards, req.SupportMultiNodes)
	}
	return fmt.Sprintf("mode=%s requireMemory=%s limitMemory=%s",
		req.Mode, humanBytes(req.RequiredMemory), humanBytes(req.LimitedMemory))
}

// explainPlacement renders every candidate the picker had to choose from, each
// with the reason it was passed over. The pressure snapshot is the one the
// decision was made against; it is consulted only to name the resource
// dimension behind a pressure rejection, which the availability view does not
// carry on its own.
func explainPlacement(req Requirement, availability *AvailabilityResult, pressure PressureSnapshot) string {
	if availability == nil {
		return "no availability view was built"
	}
	head := fmt.Sprintf("scope=%s schedulable=%t reason=%s",
		availability.Scope, availability.Schedulable, orNone(availability.Reason))
	if len(availability.Nodes) == 0 {
		// Either the cluster has no node for this mode, or every node was
		// dropped before classification — the per-node logs upstream say which.
		return head + "; no candidate node was left to choose from"
	}
	parts := make([]string, 0, len(availability.Nodes)+1)
	parts = append(parts, head)
	for _, node := range availability.Nodes {
		parts = append(parts, describeNodeOption(req, node, pressure))
	}
	return strings.Join(parts, "; ")
}

func describeNodeOption(req Requirement, node NodeOption, pressure PressureSnapshot) string {
	pressureNote := nodePressureNote(node.NodeName, req, pressure)
	if len(node.Devices) == 0 {
		// No status note: with no card at all, "no card is large enough" would
		// misdescribe why the node is unusable.
		return fmt.Sprintf("node %s status=%s%s exposes no %s card",
			node.NodeName, node.Status, pressureNote, node.GPUType)
	}
	desc := fmt.Sprintf("node %s status=%s%s%s", node.NodeName, node.Status,
		nodeStatusNote(req, node.Status), pressureNote)
	cards := make([]string, 0, len(node.Devices))
	for _, device := range node.Devices {
		cards = append(cards, describeDeviceOption(req, node, device, pressureNote != ""))
	}
	return desc + " [" + strings.Join(cards, " | ") + "]"
}

// nodeStatusNote spells out what a node status means in terms of the request,
// since the status alone reads as a verdict without its reason. NotEnough and
// NotAvailable are distinct and easy to confuse: the first means the vram is not
// there at all, the second that it is there but taken. Which vram is compared
// depends on the request — a multi-card app is judged on the node's total, a
// single-card one on its largest card — so the note follows suit. When pressure
// is what turned the node down, the separate overPressure marker says so.
func nodeStatusNote(req Requirement, status string) string {
	aggregate := req.Mode == utils.NvidiaCardType && (req.SupportMultiCards || req.SupportMultiNodes)
	switch status {
	case NodeStatusNotEnough:
		if aggregate {
			return " (its cards do not add up to the request)"
		}
		return " (no card on it is large enough for the request)"
	case NodeStatusNotAvailable:
		if aggregate {
			return " (its cards add up to the request but not enough of it is free)"
		}
		return " (its cards are large enough but none has enough free)"
	}
	return ""
}

func describeDeviceOption(req Requirement, node NodeOption, device DeviceOption, nodePressured bool) string {
	desc := fmt.Sprintf("card %s type=%s vram=%s free=%s health=%s fit=%s",
		device.DeviceID, device.SupportType, humanBytes(device.Capacity),
		humanBytes(device.Available), device.Health, orNone(device.FitLevel))
	if len(device.Bindings) > 0 {
		holders := make([]string, 0, len(device.Bindings))
		for _, binding := range device.Bindings {
			holders = append(holders, binding.Owner+"/"+binding.AppName)
		}
		desc += " heldBy=" + strings.Join(holders, ",")
	}
	if reason := deviceSkipReason(req, node, device, nodePressured); reason != "" {
		desc += " SKIPPED(" + reason + ")"
	}
	return desc
}

// deviceSkipReason names the first check the auto-picker would have failed on
// for this card. It reads the facts the view already recorded rather than
// re-running the fit logic, so it cannot drift out of step with the picker:
// Operable is exactly "healthy, has free vram, and sits on a node the current
// scope offers cards from", and FitLevel is empty exactly when the card fits
// neither the app's limit nor its request.
//
// The card's own shortfall is reported before the node verdict on purpose. A
// card too small for the request is also not Operable, since the node it sits
// on is NotEnough for the same reason, and blaming the node there just points
// back at the note the node already carries.
func deviceSkipReason(req Requirement, node NodeOption, device DeviceOption, nodePressured bool) string {
	switch {
	case device.Health != deviceHealthYes:
		return "the card reports itself unhealthy, or its node is down"
	case device.Available <= 0:
		return "the card has no free vram left"
	case device.FitLevel == "":
		return "the card does not fit, " + unfitReason(req, device, nodePressured)
	case !device.Operable:
		return "the card is not offered because its node is " + node.Status
	}
	return ""
}

// unfitReason explains an empty FitLevel, which is the one skip the card's own
// numbers do not account for: the card is healthy, has room, and sits on a node
// that offers cards, yet the fit check still turned it down. Multi-card requests
// are left out of the size comparison because there a card smaller than the
// target legitimately contributes only part of it, so its size is not what
// disqualified it.
func unfitReason(req Requirement, device DeviceOption, nodePressured bool) string {
	if nodePressured {
		return "its node is over the pressure threshold"
	}
	if req.SupportMultiCards || req.SupportMultiNodes {
		return "the aggregate fit check turned it down"
	}
	if required := requiredTargetForMode(req); device.Available < required {
		return fmt.Sprintf("%s free against a %s request", humanBytes(device.Available), humanBytes(required))
	}
	return "the fit check turned it down"
}

// nodePressureNote reports the dimensions that would exceed the pressure
// threshold once the app's cpu/memory/disk request lands on the node. This is
// the one rejection reason a card's own numbers cannot explain: a card with
// plenty of free vram is still refused when its node is too busy.
func nodePressureNote(nodeName string, req Requirement, pressure PressureSnapshot) string {
	dimensions := pressure.PressuredDimensions(Node{NodeName: nodeName}, AddedResources{
		CPU:    req.RequiredCPU,
		Memory: req.RequiredMemory,
		Disk:   req.RequiredDisk,
	})
	if len(dimensions) == 0 {
		return ""
	}
	over := make([]string, 0, len(dimensions))
	for _, dimension := range dimensions {
		over = append(over, fmt.Sprintf("%s at %d%% of capacity",
			dimension.Resource, percentOf(dimension.Used, dimension.Capacity)))
	}
	return fmt.Sprintf(" overPressure=[%s, threshold %d%%]",
		strings.Join(over, ","), percentOfFloat(pressure.Threshold))
}

func describeSelections(selections []BindingSelection) string {
	if len(selections) == 0 {
		return "none"
	}
	out := make([]string, 0, len(selections))
	for _, selection := range selections {
		out = append(out, fmt.Sprintf("%s/%s(%s)",
			selection.NodeName, selection.DeviceID, humanBytes(selection.Memory)))
	}
	return strings.Join(out, ",")
}

func describeAllocations(allocations []Allocation) string {
	if len(allocations) == 0 {
		return "none"
	}
	out := make([]string, 0, len(allocations))
	for _, allocation := range allocations {
		// buildAllocation deliberately persists Memory=0 for the whole-card
		// modes, so that the HAMI binding omits spec.memory and leaves the pod
		// unrestricted. Printing "0" would read as "no vram was reserved",
		// which is the opposite of what happened.
		amount := humanBytes(allocation.Memory)
		if allocation.Memory == 0 {
			amount = "whole card"
		}
		out = append(out, fmt.Sprintf("%s/%s(%s)",
			allocation.NodeName, allocation.DeviceID, amount))
	}
	return strings.Join(out, ",")
}

func humanBytes(value int64) string {
	return resource.NewQuantity(value, resource.BinarySI).String()
}

func percentOf(part, whole int64) int {
	if whole <= 0 {
		return 100
	}
	return int(float64(part) / float64(whole) * 100)
}

func percentOfFloat(ratio float64) int {
	if ratio <= 0 {
		ratio = defaultPressureThreshold
	}
	return int(ratio * 100)
}

func orNone(value string) string {
	if value == "" {
		return "none"
	}
	return value
}
