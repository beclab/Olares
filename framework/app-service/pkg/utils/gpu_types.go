package utils

import corev1 "k8s.io/api/core/v1"

const (
	NodeGPUTypeLabel = "gpu.bytetrade.io/type"

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
	IntelType      = "intel"       // Intel integrated GPU
	AMDType        = "amd"         // AMD integrated GPU
	IntelGPUType   = "intel-gpu"   // Intel discrete GPU
	AMDGPUType     = "amd-gpu"     // AMD discrete GPU
	MooreSocType   = "moore-soc"   // Moore Threads SoC
)

// NodeGPUType returns the canonical GPU type string for a node by reading the
// gpu.bytetrade.io/type label and folding the three "no GPU here" variants —
// label missing, label set to the empty string, or label set to "none" —
// into CPUType. Every other value is returned verbatim; the contract assumed
// across the codebase is that other label values, OlaresManifest mode names,
// and frontend GPU-type selections are already canonical (lowercased, no
// whitespace), so no further normalization is necessary.
//
// Use this helper at any boundary that ingests a node's label, instead of
// reading node.Labels[NodeGPUTypeLabel] directly: it keeps the cpu-fallback
// semantics consistent across the package and is the single place that has
// to be updated if more "no GPU" sentinel values ever get added.
func NodeGPUType(node *corev1.Node) string {
	if node == nil {
		return CPUType
	}
	t, ok := node.Labels[NodeGPUTypeLabel]
	if !ok || t == "" || t == noneGPUTypeLabel {
		return CPUType
	}
	return t
}

// IsCPUOnlyNodeLabel reports whether the value read from a node's
// gpu.bytetrade.io/type label denotes a cpu-only node. Use this at boundaries
// where you've already extracted the label string and need to decide whether
// to skip / fall back to CPUType, e.g. inside iteration over a NodeList.
func IsCPUOnlyNodeLabel(label string, present bool) bool {
	return !present || label == "" || label == noneGPUTypeLabel
}
