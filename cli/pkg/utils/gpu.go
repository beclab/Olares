package utils

import (
	"encoding/xml"
	"errors"
	"os"
	"os/exec"
	"runtime"
	"strings"

	"bytetrade.io/web3os/installer/pkg/core/connector"
	"k8s.io/klog/v2"
)

type GPU struct {
	ID                  string `xml:"id,attr" json:"id"`
	ProductName         string `xml:"product_name" json:"product_name"`
	ProductBrand        string `xml:"product_brand" json:"product_brand"`
	ProductArchitecture string `xml:"product_architecture" json:"product_architecture"`
	DisplayMode         string `xml:"display_mode" json:"display_mode"`
	DisplayActive       string `xml:"display_active" json:"display_active"`
	// PersistenceMode     string `xml:"persistence_mode" json:"persistence_mode"`
	// mig_mode
	// mig_devices
	// AccountingMode           string `xml:"accounting_mode" json:"accounting_mode"`
	// AccountingModeBufferSize string `xml:"accounting_mode_buffer_size" json:"accounting_mode_buffer_size"`
	// CurrentDm                string `xml:"driver_model>current_dm" json:"current_dm"`
	// PendingDm                string `xml:"driver_model>pending_dm" json:"pending_dm"`
	// Serial        string `xml:"serial" json:"serial"`
	UUID string `xml:"uuid" json:"uuid"`
	// MinorNumber   string `xml:"minor_number" json:"minor_number"`
	VbiosVersion  string `xml:"vbios_version" json:"vbios_version"`
	MultigpuBoard string `xml:"multigpu_board" json:"multigpu_board"`
	BoardID       string `xml:"board_id" json:"board_id"`
	// gpu_part_number
	// gpu_module_id
	// ImgVersion string `xml:"inforom_version>img_version" json:"img_version"`
	// OemObject  string `xml:"inforom_version>oem_object" json:"oem_object"`
	// EccObject  string `xml:"inforom_version>ecc_object" json:"ecc_object"`
	// PwrObject  string `xml:"inforom_version>pwr_object" json:"pwr_object"`
	// CurrentGom string `xml:"gpu_operation_mode>current_gom" json:"current_gom"`
	// PendingGom string `xml:"gpu_operation_mode>pending_gom" json:"pending_gom"`
	// gsp_firmware_version
	// gpu_virtualization_mode
	// ibmnpu
	PciBus           string `xml:"pci>pci_bus" json:"pci_bus"`
	PciDevice        string `xml:"pci>pci_device" json:"pci_device"`
	PciDomain        string `xml:"pci>pci_domain" json:"pci_domain"`
	PciDeviceID      string `xml:"pci>pci_device_id" json:"pci_device_id"`
	PciBusID         string `xml:"pci>pci_bus_id" json:"pci_bus_id"`
	PciSubSystemID   string `xml:"pci>pci_sub_system_id" json:"pci_sub_system_id"`
	MaxLinkGen       string `xml:"pci>pci_gpu_link_info>pcie_gen>max_link_gen" json:"max_link_gen"`
	CurrentLinkGen   string `xml:"pci>pci_gpu_link_info>pcie_gen>current_link_gen" json:"current_link_gen"`
	MaxLinkWidth     string `xml:"pci>pci_gpu_link_info>link_widths>max_link_width" json:"max_link_width"`
	CurrentLinkWidth string `xml:"pci>pci_gpu_link_info>link_widths>current_link_width" json:"current_link_width"`
	// BridgeChipType   string `xml:"pci>pci_bridge_chip>bridge_chip_type" json:"bridge_chip_type"`
	// BridgeChipFw     string `xml:"pci>pci_bridge_chip>bridge_chip_fw" json:"bridge_chip_fw"`
	// ReplayCounter    string `xml:"pci>replay_counter" json:"replay_counter"`
	// replay_rollover_counter
	PciTxUtil                                     string `xml:"pci>tx_util" json:"pci_tx_util"`
	PciRxUtil                                     string `xml:"pci>rx_util" json:"pci_rx_util"`
	FanSpeed                                      string `xml:"fan_speed" json:"fan_speed"`
	PerformanceState                              string `xml:"performance_state" json:"performance_state"`
	ClocksThrottleReasonGpuIdle                   string `xml:"clocks_throttle_reasons>clocks_throttle_reason_gpu_idle" json:"clocks_throttle_reason_gpu_idle"`
	ClocksThrottleReasonApplicationsClocksSetting string `xml:"clocks_throttle_reasons>clocks_throttle_reason_applications_clocks_setting" json:"clocks_throttle_reason_applications_clocks_setting"`
	ClocksThrottleReasonSwPowerCap                string `xml:"clocks_throttle_reasons>clocks_throttle_reason_sw_power_cap" json:"clocks_throttle_reason_sw_power_cap"`
	ClocksThrottleReasonHwSlowdown                string `xml:"clocks_throttle_reasons>clocks_throttle_reason_hw_slowdown" json:"clocks_throttle_reason_hw_slowdown"`
	ClocksThrottleReasonHwThermalSlowdown         string `xml:"clocks_throttle_reasons>clocks_throttle_reason_hw_thermal_slowdown" json:"clocks_throttle_reason_hw_thermal_slowdown"`
	ClocksThrottleReasonHwPowerBrakeSlowdown      string `xml:"clocks_throttle_reasons>clocks_throttle_reason_hw_power_brake_slowdown" json:"clocks_throttle_reason_hw_power_brake_slowdown"`
	ClocksThrottleReasonSyncBoost                 string `xml:"clocks_throttle_reasons>clocks_throttle_reason_sync_boost" json:"clocks_throttle_reason_sync_boost"`
	ClocksThrottleReasonSwThermalSlowdown         string `xml:"clocks_throttle_reasons>clocks_throttle_reason_sw_thermal_slowdown" json:"clocks_throttle_reason_sw_thermal_slowdown"`
	ClocksThrottleReasonDisplayClocksSetting      string `xml:"clocks_throttle_reasons>clocks_throttle_reason_display_clocks_setting" json:"clocks_throttle_reason_display_clocks_setting"`
	// ClocksThrottleReasonUnknown                   string `xml:"clocks_throttle_reasons>clocks_throttle_reason_unknown" json:"clocks_throttle_reason_unknown"`
	FbMemoryUsageTotal    string `xml:"fb_memory_usage>total" json:"fb_memory_usage_total"`
	FbMemoryUsageReserved string `xml:"fb_memory_usage>reserved" json:"fb_memory_usage_reserved"`
	FbMemoryUsageUsed     string `xml:"fb_memory_usage>used" json:"fb_memory_usage_used"`
	FbMemoryUsageFree     string `xml:"fb_memory_usage>free" json:"fb_memory_usage_free"`
	Bar1Total             string `xml:"bar1_memory_usage>total" json:"bar1_total"`
	Bar1Used              string `xml:"bar1_memory_usage>used" json:"bar1_used"`
	Bar1Free              string `xml:"bar1_memory_usage>free" json:"bar1_free"`
	ComputeMode           string `xml:"compute_mode" json:"compute_mode"`
	GpuUtil               string `xml:"utilization>gpu_util" json:"gpu_util"`
	MemoryUtil            string `xml:"utilization>memory_util" json:"memory_util"`
	EncoderUtil           string `xml:"utilization>encoder_util" json:"encoder_util"`
	DecoderUtil           string `xml:"utilization>decoder_util" json:"decoder_util"`
	// encoder_stats
	// fbc_stats
	// CurrentEcc                                        string `xml:"ecc_mode>current_ecc" json:"current_ecc"`
	// PendingEcc                                        string `xml:"ecc_mode>pending_ecc" json:"pending_ecc"`
	// DeviceMemory                                      string `xml:"ecc_errors>volatile>single_bit>device_memory" json:"device_memory"`
	// L1Cache                                           string `xml:"ecc_errors>volatile>single_bit>l1_cache" json:"l_1_cache"`
	// TotalSingleBitVolatileEccErrorsGpu                string `xml:"ecc_errors>volatile>single_bit>total" json:"total_single_bit_volatile_ecc_errors_gpu"`
	// TextureMemory                                     string `xml:"ecc_errors>volatile>single_bit>texture_memory" json:"texture_memory"`
	// RegisterFile                                      string `xml:"ecc_errors>volatile>single_bit>register_file" json:"register_file"`
	// L2Cache                                           string `xml:"ecc_errors>volatile>single_bit>l2_cache" json:"l_2_cache"`
	// TextureMemoryDoubleBitVolatileEccErrorsGpu        string `xml:"ecc_errors>volatile>double_bit>texture_memory" json:"texture_memory_double_bit_volatile_ecc_errors_gpu"`
	// DeviceMemoryDoubleBitVolatileEccErrorsGpu         string `xml:"ecc_errors>volatile>double_bit>device_memory" json:"device_memory_double_bit_volatile_ecc_errors_gpu"`
	// RegisterFileDoubleBitVolatileEccErrorsGpu         string `xml:"ecc_errors>volatile>double_bit>register_file" json:"register_file_double_bit_volatile_ecc_errors_gpu"`
	// TotalDoubleBitVolatileEccErrorsGpu                string `xml:"ecc_errors>volatile>double_bit>total" json:"total_double_bit_volatile_ecc_errors_gpu"`
	// L2CacheDoubleBitVolatileEccErrorsGpu              string `xml:"ecc_errors>volatile>double_bit>l2_cache" json:"l_2_cache_double_bit_volatile_ecc_errors_gpu"`
	// L1CacheDoubleBitVolatileEccErrorsGpu              string `xml:"ecc_errors>volatile>double_bit>l1_cache" json:"l_1_cache_double_bit_volatile_ecc_errors_gpu"`
	// L2CacheSingleBitAggregateEccErrorsGpu             string `xml:"ecc_errors>aggregate>single_bit>l2_cache" json:"l_2_cache_single_bit_aggregate_ecc_errors_gpu"`
	// L1CacheSingleBitAggregateEccErrorsGpu             string `xml:"ecc_errors>aggregate>single_bit>l1_cache" json:"l_1_cache_single_bit_aggregate_ecc_errors_gpu"`
	// TextureMemorySingleBitAggregateEccErrorsGpu       string `xml:"ecc_errors>aggregate>single_bit>texture_memory" json:"texture_memory_single_bit_aggregate_ecc_errors_gpu"`
	// RegisterFileSingleBitAggregateEccErrorsGpu        string `xml:"ecc_errors>aggregate>single_bit>register_file" json:"register_file_single_bit_aggregate_ecc_errors_gpu"`
	// DeviceMemorySingleBitAggregateEccErrorsGpu        string `xml:"ecc_errors>aggregate>single_bit>device_memory" json:"device_memory_single_bit_aggregate_ecc_errors_gpu"`
	// TotalSingleBitAggregateEccErrorsGpu               string `xml:"ecc_errors>aggregate>single_bit>total" json:"total_single_bit_aggregate_ecc_errors_gpu"`
	// DeviceMemoryDoubleBitAggregateEccErrorsGpu        string `xml:"ecc_errors>aggregate>double_bit>device_memory" json:"device_memory_double_bit_aggregate_ecc_errors_gpu"`
	// TotalDoubleBitAggregateEccErrorsGpu               string `xml:"ecc_errors>aggregate>double_bit>total" json:"total_double_bit_aggregate_ecc_errors_gpu"`
	// RegisterFileDoubleBitAggregateEccErrorsGpu        string `xml:"ecc_errors>aggregate>double_bit>register_file" json:"register_file_double_bit_aggregate_ecc_errors_gpu"`
	// L2CacheDoubleBitAggregateEccErrorsGpu             string `xml:"ecc_errors>aggregate>double_bit>l2_cache" json:"l_2_cache_double_bit_aggregate_ecc_errors_gpu"`
	// L1CacheDoubleBitAggregateEccErrorsGpu             string `xml:"ecc_errors>aggregate>double_bit>l1_cache" json:"l_1_cache_double_bit_aggregate_ecc_errors_gpu"`
	// TextureMemoryDoubleBitAggregateEccErrorsGpu       string `xml:"ecc_errors>aggregate>double_bit>texture_memory" json:"texture_memory_double_bit_aggregate_ecc_errors_gpu"`
	// RetiredCount                                      string `xml:"retired_pages>multiple_single_bit_retirement>retired_count" json:"retired_count"`
	// RetiredPagelist                                   string `xml:"retired_pages>multiple_single_bit_retirement>retired_pagelist" json:"retired_pagelist"`
	// RetiredCountDoubleBitRetirementRetiredPagesGpu    string `xml:"retired_pages>double_bit_retirement>retired_count" json:"retired_count_double_bit_retirement_retired_pages_gpu"`
	// RetiredPagelistDoubleBitRetirementRetiredPagesGpu string `xml:"retired_pages>double_bit_retirement>retired_pagelist" json:"retired_pagelist_double_bit_retirement_retired_pages_gpu"`
	// pending_blacklist
	// PendingRetirement string `xml:"retired_pages>pending_retirement" json:"pending_retirement"`
	// remapped_rows
	GpuTemp                string `xml:"temperature>gpu_temp" json:"gpu_temp"`
	GpuTempMaxThreshold    string `xml:"temperature>gpu_temp_max_threshold" json:"gpu_temp_max_threshold"`
	GpuTempSlowThreshold   string `xml:"temperature>gpu_temp_slow_threshold" json:"gpu_temp_slow_threshold"`
	GpuTempMaxGpuThreshold string `xml:"temperature>gpu_temp_max_gpu_threshold" json:"gpu_temp_max_gpu_threshold"`
	GpuTargetTemp          string `xml:"temperature>gpu_target_temperature" json:"gpu_target_temperature"`
	MemoryTemp             string `xml:"temperature>memory_temp" json:"memory_temp"`
	GpuTempMaxMemThreshold string `xml:"temperature>gpu_temp_max_mem_threshold" json:"gpu_temp_max_mem_threshold"`
	GpuTargetTempMin       string `xml:"supported_gpu_target_temp>gpu_target_temp_min" json:"gpu_target_temp_min"`
	GpuTargetTempMax       string `xml:"supported_gpu_target_temp>gpu_target_temp_max" json:"gpu_target_temp_max"`
	PowerState             string `xml:"gpu_power_readings>power_state" json:"power_state"`
	PowerManagement        string `xml:"gpu_power_readings>power_management" json:"power_management"`
	PowerDraw              string `xml:"gpu_power_readings>power_draw" json:"power_draw"`
	PowerLimit             string `xml:"gpu_power_readings>current_power_limit" json:"power_limit"`
	DefaultPowerLimit      string `xml:"gpu_power_readings>default_power_limit" json:"default_power_limit"`
	EnforcedPowerLimit     string `xml:"gpu_power_readings>requested_power_limit" json:"enforced_power_limit"`
	MinPowerLimit          string `xml:"gpu_power_readings>min_power_limit" json:"min_power_limit"`
	MaxPowerLimit          string `xml:"gpu_power_readings>max_power_limit" json:"max_power_limit"`
	GraphicsClock          string `xml:"clocks>graphics_clock" json:"graphics_clock"`
	SmClock                string `xml:"clocks>sm_clock" json:"sm_clock"`
	MemClock               string `xml:"clocks>mem_clock" json:"mem_clock"`
	VideoClock             string `xml:"clocks>video_clock" json:"video_clock"`
	// GraphicsClockApplicationsClocksGpu        string `xml:"applications_clocks>graphics_clock" json:"graphics_clock_applications_clocks_gpu"`
	// MemClockApplicationsClocksGpu             string `xml:"applications_clocks>mem_clock" json:"mem_clock_applications_clocks_gpu"`
	// GraphicsClockDefaultApplicationsClocksGpu string `xml:"default_applications_clocks>graphics_clock" json:"graphics_clock_default_applications_clocks_gpu"`
	// MemClockDefaultApplicationsClocksGpu      string `xml:"default_applications_clocks>mem_clock" json:"mem_clock_default_applications_clocks_gpu"`
	GraphicsClockMax string `xml:"max_clocks>graphics_clock" json:"graphics_clock_max"`
	SmClockMax       string `xml:"max_clocks>sm_clock" json:"sm_clock_max"`
	MemClockMax      string `xml:"max_clocks>mem_clock" json:"mem_clock_max"`
	VideoClockMax    string `xml:"max_clocks>video_clock" json:"video_clock_max"`
	// max_customer_boost_clocks
	// AutoBoost        string `xml:"clock_policy>auto_boost" json:"auto_boost"`
	// AutoBoostDefault string `xml:"clock_policy>auto_boost_default" json:"auto_boost_default"`
	// voltage>graphics_volt
	// Value                  []string `xml:"supported_clocks>supported_mem_clock>value" json:"value"`
	// SupportedGraphicsClock []string `xml:"supported_clocks>supported_mem_clock>supported_graphics_clock" json:"supported_graphics_clock"`
	// Processes              string   `xml:"processes" json:"processes"`
	// AccountedProcesses     string   `xml:"accounted_processes" json:"accounted_processes"`
}

type NvidiaGpuInfo struct {
	Timestamp     string `xml:"timestamp" json:"timestamp"`
	DriverVersion string `xml:"driver_version" json:"driver_version"`
	CudaVersion   string `xml:"cuda_version" json:"cuda_version"`
	AttachedGpus  string `xml:"attached_gpus" json:"attached_gpus"`
	GPUS          []GPU  `xml:"gpu" json:"gpus"`
}

func ExecNvidiaSmi(execRuntime connector.Runtime) (gpuInfo *NvidiaGpuInfo, installed bool, err error) {
	cmd := "nvidia-smi"
	if runtime.GOOS == "windows" {
		cmd += ".exe"
	}

	cmdPath, err := exec.LookPath(cmd)
	if err != nil {
		if os.IsNotExist(err) || errors.Is(err, exec.ErrNotFound) {
			// try to find wsl nvidia-smi
			var nvidiaSmiFile = "/usr/lib/wsl/lib/nvidia-smi"
			_, e := os.Stat(nvidiaSmiFile)
			if e != nil {
				if os.IsNotExist(e) {
					return nil, false, nil
				}
				return nil, false, err
			}

			cmdPath = nvidiaSmiFile
		} else {
			return nil, false, err
		}
	}

	out, err := execRuntime.GetRunner().SudoCmd(cmdPath+" -q -x", false, false)
	if err != nil {
		// when nvidia-smi command is installed but cuda is not installed
		if strings.Contains(out, "couldn't communicate with the NVIDIA driver") {
			return nil, false, nil
		}
		klog.Error("Error running nvidia-smi:", err)
		return nil, false, err
	}

	var data NvidiaGpuInfo
	if err := xml.Unmarshal([]byte(out), &data); nil != err {
		klog.Error("Error unmarshalling from XML:", err)
		return nil, false, err
	}

	return &data, true, nil
}
