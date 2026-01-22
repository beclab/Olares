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
	NvidiaCardType = "nvidia"           // handling by HAMi
	AmdCardType    = "amd"              //
	DgxSparkType   = "nvidia-dgx-spark" // NVIDIA GB10 Superchip & unified system memory
)
