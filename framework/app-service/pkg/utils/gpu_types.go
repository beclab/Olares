package utils

const (
	NodeGPUTypeLabel = "gpu.bytetrade.io/type"
)

const (
	CPUType           = "cpu"         // force to use CPU, no GPU
	NvidiaCardType    = "nvidia"      // handling by HAMi
	AmdGpuCardType    = "amd-gpu"     //
	AmdApuCardType    = "amd-apu"     // AMD APU with integrated GPU , AI Max 395 etc.
	GB10ChipType      = "nvidia-gb10" // NVIDIA GB10 Superchip & unified system memory
	StrixHaloChipType = "strix-halo"  // AMD Strix Halo GPU & unified system memory
)
