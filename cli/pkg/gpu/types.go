package gpu

const (
	GpuLabelGroup = "gpu.bytetrade.io"
)

var (
	GpuDriverLabel        = GpuLabelGroup + "/driver"
	GpuCudaLabel          = GpuLabelGroup + "/cuda"
	GpuCudaSupportedLabel = GpuLabelGroup + "/cuda-supported"

	// GpuType is the legacy single-value node label (gpu.bytetrade.io/type).
	// olares-cli no longer writes it: a node's supported modes are now
	// advertised through the existence-based per-mode labels returned by
	// GpuModeLabel. It is still cleaned up on unlabel so stale values left by
	// older installs don't linger.
	GpuType = GpuLabelGroup + "/type"
)

const (
	CPUType        = "cpu"         // force to use CPU, no GPU
	NvidiaCardType = "nvidia"      // discrete NVIDIA card, handled by HAMi
	GB10ChipType   = "nvidia-gb10" // NVIDIA GB10 Superchip & unified system memory
	AppleMChipType = "apple-m"     // Apple M-series SoC
	IntelType      = "intel"       // Intel integrated GPU (unified memory)
	AMDType        = "amd"         // AMD integrated GPU (Ryzen AI Max)
	IntelGpuType   = "intel-gpu"   // Intel discrete GPU (not handled yet)
	AmdGpuType     = "amd-gpu"     // AMD discrete GPU (not handled yet)
	MooreSocType   = "moore-soc"   // Moore Threads SoC
)

// AllGpuModeTypes is the set of non-cpu modes olares-cli knows how to label a
// node with. It is used to strip every per-mode label on unlabel. cpu is never
// labeled — every node implicitly supports cpu workloads.
var AllGpuModeTypes = []string{
	NvidiaCardType,
	GB10ChipType,
	AppleMChipType,
	IntelType,
	AMDType,
	IntelGpuType,
	AmdGpuType,
	MooreSocType,
}

// GpuModeLabel returns the existence-based per-mode node label key for a mode,
// e.g. GpuModeLabel(NvidiaCardType) == "gpu.bytetrade.io/nvidia". A node
// supports the mode iff it carries this label (value is "true"); a node may
// carry several at once.
func GpuModeLabel(mode string) string {
	return GpuLabelGroup + "/" + mode
}
