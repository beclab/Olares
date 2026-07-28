package utils

import corev1 "k8s.io/api/core/v1"

const (
	// NodeGPUTypeLabel is the legacy single-value node label
	// (gpu.bytetrade.io/type=<mode>). Current olares-cli no longer writes it
	// — it uses the existence-based per-mode labels below so a node can
	// advertise several modes at once. It is still read so a historical
	// nvidia node labeled by an older olares-cli keeps working: nvidia is the
	// only pre-existing GPU type and its name is unchanged, so no alias
	// mapping is needed.
	NodeGPUTypeLabel = "gpu.bytetrade.io/type"

	// NodeGPUTypeLabelPrefix is the prefix for the existence-based per-mode
	// node labels. A node supports mode <m> iff it carries the label
	// gpu.bytetrade.io/<m> (the value is ignored; olares-cli writes "true").
	// cpu is never labeled — every node implicitly supports cpu workloads.
	NodeGPUTypeLabelPrefix = "gpu.bytetrade.io/"

	// noneGPUTypeLabel is the literal value some node-init paths write to
	// gpu.bytetrade.io/type to mark a node that doesn't have a GPU. It's
	// treated identically to a missing or empty label: the node runs cpu
	// workloads only.
	noneGPUTypeLabel = "none"
)

const (
	CPUType        = "cpu"         // force to use CPU, no GPU
	NvidiaCardType = "nvidia"      // discrete NVIDIA card, handled by HAMi
	GB10ChipType   = "nvidia-gb10" // NVIDIA GB10 Superchip & unified system memory
	AppleMChipType = "apple-m"     // Apple M-series SoC (unified memory)
	IntelType      = "intel"       // Intel integrated GPU (unified memory)
	AMDType        = "amd"         // AMD integrated GPU (Ryzen AI Max)
	IntelGPUType   = "intel-gpu"   // Intel discrete GPU
	AMDGPUType     = "amd-gpu"     // AMD discrete GPU
	MooreSocType   = "moore-soc"   // Moore Threads SoC
)

// canonicalGPUTypes is the ordered set of non-cpu modes the system recognizes
// as node-labelable. It is the only set scanned for the existence-based
// per-mode labels, so unrelated keys under the gpu.bytetrade.io/ prefix
// (driver, cuda, cuda-supported, mode, …) are never mistaken for a GPU mode.
var canonicalGPUTypes = []string{
	NvidiaCardType,
	GB10ChipType,
	AppleMChipType,
	IntelType,
	AMDType,
	IntelGPUType,
	AMDGPUType,
	MooreSocType,
}

// NodeSupportedGPUTypes returns the deduplicated set of non-cpu GPU modes a
// node supports, combining:
//
//   - the existence-based per-mode labels gpu.bytetrade.io/<mode> written by
//     current olares-cli (a node may carry several at once), and
//   - the legacy single-value gpu.bytetrade.io/type label, retained so a
//     historical nvidia node labeled by an older olares-cli is still honored.
//
// cpu is never included (every node implicitly supports cpu). Results are
// deduplicated. Declaration order follows canonicalGPUTypes with any extra
// legacy-label value appended last.
func NodeSupportedGPUTypes(node *corev1.Node) []string {
	if node == nil {
		return nil
	}
	seen := make(map[string]struct{})
	out := make([]string, 0)
	add := func(mode string) {
		if mode == "" || mode == CPUType {
			return
		}
		if _, ok := seen[mode]; ok {
			return
		}
		seen[mode] = struct{}{}
		out = append(out, mode)
	}
	for _, mode := range canonicalGPUTypes {
		if _, ok := node.Labels[NodeGPUTypeLabelPrefix+mode]; ok {
			add(mode)
		}
	}
	if t, ok := node.Labels[NodeGPUTypeLabel]; !IsCPUOnlyNodeLabel(t, ok) {
		add(t)
	}
	return out
}

// IsCPUOnlyNodeLabel reports whether the value read from a node's legacy
// gpu.bytetrade.io/type label denotes a cpu-only node. Use this at boundaries
// where you've already extracted the label string and need to decide whether
// to skip / fall back to CPUType, e.g. inside iteration over a NodeList.
func IsCPUOnlyNodeLabel(label string, present bool) bool {
	return !present || label == "" || label == noneGPUTypeLabel
}
