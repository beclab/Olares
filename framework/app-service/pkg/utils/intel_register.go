package utils

import (
	"fmt"
	"strconv"
	"strings"

	"github.com/beclab/Olares/framework/app-service/pkg/constants"
	corev1 "k8s.io/api/core/v1"
)

// Intel GPU kinds and drivers as published in the node register annotation.
const (
	IntelGPUKindIntegrated = "igpu"
	IntelGPUKindDiscrete   = "dgpu"

	IntelDriverI915 = "i915"
	IntelDriverXe   = "xe"
)

// IntelGPUEntry is a single Intel GPU card as advertised by the device plugin
// in the bytetrade.io/node-intel-register node annotation.
type IntelGPUEntry struct {
	Kind         string // igpu | dgpu
	Card         string // e.g. card0
	Driver       string // i915 | xe
	Name         string // product name, may be empty for unknown PCI ids
	Architecture string // e.g. Xe-HPG, may be empty
	Codename     string // e.g. Alchemist, may be empty
	Mem          uint64 // discrete VRAM in bytes; 0 for integrated
}

// ParseNodeIntelRegister parses the bytetrade.io/node-intel-register annotation
// value. The format is a ':'-separated list of entries, each a ','-separated
// tuple "<igpu|dgpu>,<cardN>,<i915|xe>,<name>,<architecture>,<codename>,<mem>".
// name/architecture/codename may be empty (unknown PCI id); mem is a byte count
// (0 for integrated). Empty input yields an empty slice with no error. A
// malformed entry (wrong field count, unknown kind/driver, non-numeric mem) is
// an error; callers may choose to log and fall back.
func ParseNodeIntelRegister(s string) ([]IntelGPUEntry, error) {
	s = strings.TrimSpace(s)
	if s == "" {
		return nil, nil
	}

	entries := make([]IntelGPUEntry, 0)
	for _, raw := range strings.Split(s, ":") {
		raw = strings.TrimSpace(raw)
		if raw == "" {
			continue
		}

		fields := strings.Split(raw, ",")
		if len(fields) != 7 {
			return nil, fmt.Errorf("invalid intel register entry %q: expected 7 fields, got %d", raw, len(fields))
		}

		mem, err := strconv.ParseUint(strings.TrimSpace(fields[6]), 10, 64)
		if err != nil {
			return nil, fmt.Errorf("invalid intel register entry %q: bad mem %q: %w", raw, fields[6], err)
		}

		entry := IntelGPUEntry{
			Kind:         strings.TrimSpace(fields[0]),
			Card:         strings.TrimSpace(fields[1]),
			Driver:       strings.TrimSpace(fields[2]),
			Name:         strings.TrimSpace(fields[3]),
			Architecture: strings.TrimSpace(fields[4]),
			Codename:     strings.TrimSpace(fields[5]),
			Mem:          mem,
		}

		switch entry.Kind {
		case IntelGPUKindIntegrated, IntelGPUKindDiscrete:
		default:
			return nil, fmt.Errorf("invalid intel register entry %q: unknown kind %q", raw, entry.Kind)
		}

		switch entry.Driver {
		case IntelDriverI915, IntelDriverXe:
		default:
			return nil, fmt.Errorf("invalid intel register entry %q: unknown driver %q", raw, entry.Driver)
		}

		entries = append(entries, entry)
	}

	return entries, nil
}

// IntelRegisterFromNode parses the register annotation off a node, returning an
// empty slice when the annotation is absent.
func IntelRegisterFromNode(node *corev1.Node) ([]IntelGPUEntry, error) {
	if node == nil {
		return nil, nil
	}

	return ParseNodeIntelRegister(node.Annotations[constants.NodeIntelRegisterKey])
}
