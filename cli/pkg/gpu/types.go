package gpu

const (
	GpuLabelGroup = "gpu.bytetrade.io"
)

var (
	GpuDriverLabel        = GpuLabelGroup + "/driver"
	GpuCudaLabel          = GpuLabelGroup + "/cuda"
	GpuCudaSupportedLabel = GpuLabelGroup + "/cuda-supported"
)
