package utils

const (
	NodeGPUTypeLabel = "gpu.bytetrade.io/type"
)

const (
	NvidiaCardType = "nvidia"      // handling by HAMi
	AmdGpuCardType = "amd-gpu"     //
	AmdApuCardType = "amd-apu"     // AMD APU with integrated GPU , AI Max 395 etc.
	GB10ChipType   = "nvidia-gb10" // NVIDIA GB10 Superchip & unified system memory
)
