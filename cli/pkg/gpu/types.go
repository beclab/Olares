package gpu

const (
	GpuLabelGroup = "gpu.bytetrade.io"
)

var (
	GpuDriverLabel        = GpuLabelGroup + "/driver"
	GpuCudaLabel          = GpuLabelGroup + "/cuda"
	GpuCudaSupportedLabel = GpuLabelGroup + "/cuda-supported"
	GpuType               = GpuLabelGroup + "/type"
)

const (
	NvidiaCardType = "nvidia"      // handling by HAMi
	AmdGpuCardType = "amd-gpu"     //
	AmdApuCardType = "amd-apu"     // AMD APU with integrated GPU , AI Max 395 etc.
	GB10ChipType   = "nvidia-gb10" // NVIDIA GB10 Superchip & unified system memory
)
