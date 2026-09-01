package dashboard

// Per-vendor description of how a GPU's numbers are read and displayed.
// This is the Go counterpart of the SPA's utils/gpuVendor.ts; the titles
// below are that file's own en-US strings so the CLI and the page name the
// same panel the same way.
//
// NVIDIA is absent on purpose: HAMI only enumerates CUDA devices, so NVIDIA
// keeps the PromQL path in specs.go while Intel and AMD read KubeSphere.

const (
	gib = 1024 * 1024 * 1024
	mib = 1024 * 1024
)

// KsPanelSpec is one panel of the detail view, sourced from KubeSphere.
type KsPanelSpec struct {
	// Key is the stable identifier agents index on; titles are prose and may
	// be reworded, so nothing machine-readable should depend on them.
	Key   string
	Title string
	Value MetricRead
	// Total is the gauge denominator, and the ratio denominator when Ratio
	// is absent.
	Total *MetricRead
	// Ratio is a ready-made ratio metric, preferred over dividing Value by
	// Total because the exporter's own ratio accounts for detail the two
	// separate reads cannot (Intel's memory ratio, for one).
	Ratio *MetricRead
	// TotalConst is a fixed denominator for a value already expressed as a
	// percentage.
	TotalConst float64
	Unit       string
	// HideGauge marks a panel that contributes a trend line but no gauge.
	HideGauge bool
}

// reads lists every metric this panel touches, for batching.
func (p KsPanelSpec) reads() []MetricRead {
	out := []MetricRead{p.Value}
	if p.Total != nil {
		out = append(out, *p.Total)
	}
	if p.Ratio != nil {
		out = append(out, *p.Ratio)
	}
	return out
}

// KsDeviceColumn binds one device-table column to the metric that fills it.
type KsDeviceColumn struct {
	Field string
	Value MetricRead
}

// KsIdentity names the series whose labels carry a card's identity, one
// series per card, and how its labels map onto row fields.
type KsIdentity struct {
	Metric string
	Fields map[string]string
}

// KsVendorSpec is everything the CLI needs to render one KubeSphere-sourced
// vendor.
type KsVendorSpec struct {
	Vendor GPUVendor
	// DeviceLabels are the labels that pin one card; ID doubles as the
	// detail view's key.
	DeviceLabels struct {
		ID   string
		Node string
	}
	Identity KsIdentity
	// DeviceMetrics is one read per column, so every card comes back in the
	// same response.
	DeviceMetrics []KsDeviceColumn
	// Derived fills columns the exporter does not publish.
	Derived map[string]func(DeviceRow) any
	Gauges  []KsPanelSpec
	Lines   []KsPanelSpec
	// HiddenColumns are device-table columns this vendor cannot fill.
	HiddenColumns []string
	// SupportsTasks reports whether per-container attribution exists. Both
	// KubeSphere vendors publish no pod/namespace/container labels on any of
	// their metrics, so there is no workload dimension to build a task view
	// on.
	SupportsTasks bool
	// IsHealthy collapses this vendor's own health signal. The three vendors
	// publish health in three unrelated shapes, so each implements its own
	// rather than sharing a threshold template.
	IsHealthy func(DeviceRow) bool
	// HealthReason explains an unhealthy card when the vendor can say.
	HealthReason func(DeviceRow) string
}

// intelShaderUtilization is what "GPU utilization" means on Intel.
//
// xpumd publishes `hw_gpu_task="all"`, but that averages across all nine
// engine instances including the media and copy blocks, which sit idle
// during a compute workload and drag a saturated card down to single digits.
// The busiest of the compute and render pipelines is the reading that
// matches what NVIDIA and AMD call utilisation: how busy the shader array is.
var intelShaderUtilization = MetricRead{
	Metric: "intel_hw_gpu_utilization_ratio",
	Match:  map[string][]string{"hw_gpu_task": {"compute-all", "render-all"}},
	Agg:    AggMax,
	Scale:  100,
	// The exporter overshoots its own 0-1 range by a rounding error.
	Clamp: 100,
}

func intelKsSpec() KsVendorSpec {
	spec := KsVendorSpec{
		Vendor: VendorIntel,
		Identity: KsIdentity{
			Metric: "intel_hw_gpu_info",
			Fields: map[string]string{
				"uuid":      "hw_id",
				"nodeUid":   "hw_id",
				"type":      "hw_model",
				"nodeName":  "node",
				"pciBdf":    "pci_bdf",
				"memoryEcc": "hw_memory_ecc",
			},
		},
		DeviceMetrics: []KsDeviceColumn{
			{
				Field: "memoryTotal",
				Value: MetricRead{Metric: "intel_hw_memory_size_bytes", Agg: AggSum, Scale: 1.0 / mib},
			},
			{
				Field: "memoryUtilizedPercent",
				Value: MetricRead{Metric: "intel_hw_memory_utilization_ratio", Agg: AggAvg, Scale: 100, Clamp: 100},
			},
			{Field: "coreUtilizedPercent", Value: intelShaderUtilization},
			{
				Field: "power",
				Value: MetricRead{
					Metric: "intel_hw_power_watts",
					Match:  map[string][]string{"hw_sensor_location": {"card"}},
					Agg:    AggAvg,
				},
			},
			{
				// 1 means the driver is asking for a reset.
				Field: "resetNeeded",
				Value: MetricRead{
					Metric: "intel_hw_status",
					Match:  map[string][]string{"hw_type": {"gpu"}, "hw_state": {"reset_needed"}},
					Agg:    AggMax,
				},
			},
		},
		HiddenColumns: []string{"mode", "temperature"},
		Gauges: []KsPanelSpec{
			{Key: "util_core", Title: "GPU utilization", Value: intelShaderUtilization, TotalConst: 100},
			{
				Key:   "util_mem",
				Title: "VRAM usage",
				Value: MetricRead{Metric: "intel_hw_memory_usage_bytes", Agg: AggSum, Scale: 1.0 / gib},
				Total: &MetricRead{Metric: "intel_hw_memory_size_bytes", Agg: AggSum, Scale: 1.0 / gib},
				Ratio: &MetricRead{Metric: "intel_hw_memory_utilization_ratio", Agg: AggAvg, Scale: 100, Clamp: 100},
				Unit:  "Gi",
			},
		},
		Lines: []KsPanelSpec{
			{
				Key:   "power",
				Title: "Power draw",
				Value: MetricRead{
					Metric: "intel_hw_power_watts",
					Match:  map[string][]string{"hw_sensor_location": {"card"}},
					Agg:    AggAvg,
				},
				Unit: "W",
			},
			{Key: "engine_compute", Title: "Compute", Value: intelEngine("compute-all"), Unit: "%", HideGauge: true},
			{Key: "engine_render", Title: "Render", Value: intelEngine("render-all"), Unit: "%", HideGauge: true},
			{Key: "engine_media", Title: "Media", Value: intelEngine("media-all"), Unit: "%", HideGauge: true},
			{Key: "engine_copy", Title: "Copy", Value: intelEngine("copy-all"), Unit: "%", HideGauge: true},
		},
		// xpumd publishes no pod/namespace/container labels on any of its
		// thirteen metrics.
		SupportsTasks: false,
		// A missing reset_needed series means the driver never asked for one.
		// `hw_state="ecc_disabled"` is a configuration, not a fault, so it is
		// deliberately not consulted here.
		IsHealthy: func(row DeviceRow) bool {
			v, ok := RowFloat(row, "resetNeeded")
			return !ok || v != 1
		},
	}
	// hw_id is derived from vendor + device + PCI slot and carries no serial
	// (hw_serial_number reads "unknown"), so two identical cards seated in
	// the same slot on different nodes share it. The node pins it down.
	spec.DeviceLabels.ID = "hw_id"
	spec.DeviceLabels.Node = "node"
	return spec
}

// intelEngine reads one engine pipeline's utilisation.
func intelEngine(task string) MetricRead {
	return MetricRead{
		Metric: "intel_hw_gpu_utilization_ratio",
		Match:  map[string][]string{"hw_gpu_task": {task}},
		Agg:    AggAvg,
		Scale:  100,
		Clamp:  100,
	}
}

func amdKsSpec() KsVendorSpec {
	spec := KsVendorSpec{
		Vendor: VendorAMD,
		Identity: KsIdentity{
			// The exporter publishes no info-style metric, but it stamps
			// every series with the card's full identity, so any per-card
			// metric can stand in for one.
			Metric: "amd_gpu_health",
			Fields: map[string]string{
				"uuid":           "gpu_uuid",
				"nodeUid":        "gpu_uuid",
				"type":           "card_model",
				"nodeName":       "node",
				"driver_version": "driver_version",
				"vbiosVersion":   "vbios_version",
			},
		},
		DeviceMetrics: []KsDeviceColumn{
			{Field: "memoryTotal", Value: MetricRead{Metric: "amd_gpu_total_vram", Agg: AggSum}},
			{Field: "memoryUsed", Value: MetricRead{Metric: "amd_gpu_used_vram", Agg: AggSum}},
			{Field: "coreUtilizedPercent", Value: MetricRead{Metric: "amd_gpu_gfx_activity", Agg: AggAvg}},
			{Field: "power", Value: MetricRead{Metric: "amd_gpu_average_package_power", Agg: AggAvg}},
			{Field: "temperature", Value: MetricRead{Metric: "amd_gpu_junction_temperature", Agg: AggAvg}},
			{Field: "gpuHealth", Value: MetricRead{Metric: "amd_gpu_health", Agg: AggMax}},
			{Field: "eccUncorrectTotal", Value: MetricRead{Metric: "amd_gpu_ecc_uncorrect_total", Agg: AggMax}},
		},
		// The exporter has no memory utilisation metric, unlike xpumd.
		Derived: map[string]func(DeviceRow) any{
			"memoryUtilizedPercent": func(row DeviceRow) any {
				total, okT := RowFloat(row, "memoryTotal")
				used, okU := RowFloat(row, "memoryUsed")
				if !okT || !okU || total == 0 {
					return float64(0)
				}
				return used / total * 100
			},
		},
		HiddenColumns: []string{"mode"},
		Gauges: []KsPanelSpec{
			{
				Key:        "util_core",
				Title:      "GPU utilization",
				Value:      MetricRead{Metric: "amd_gpu_gfx_activity", Agg: AggAvg},
				TotalConst: 100,
			},
			{
				Key:   "util_mem",
				Title: "VRAM usage",
				Value: MetricRead{Metric: "amd_gpu_used_vram", Agg: AggSum, Scale: 1.0 / 1024},
				Total: &MetricRead{Metric: "amd_gpu_total_vram", Agg: AggSum, Scale: 1.0 / 1024},
				Unit:  "Gi",
			},
		},
		Lines: []KsPanelSpec{
			{
				Key:   "power",
				Title: "Power draw",
				Value: MetricRead{Metric: "amd_gpu_average_package_power", Agg: AggAvg},
				Unit:  "W",
			},
			{
				Key:   "temperature",
				Title: "Temperature",
				Value: MetricRead{Metric: "amd_gpu_junction_temperature", Agg: AggAvg},
				Unit:  "℃",
			},
			{
				Key:       "engine_umc",
				Title:     "Memory controller",
				Value:     MetricRead{Metric: "amd_gpu_umc_activity", Agg: AggAvg},
				Unit:      "%",
				HideGauge: true,
			},
			{
				Key:       "engine_vcn",
				Title:     "Video engine",
				Value:     MetricRead{Metric: "amd_gpu_vcn_activity", Agg: AggAvg},
				Unit:      "%",
				HideGauge: true,
			},
		},
		// The exporter publishes no pod/namespace/container labels either.
		SupportsTasks: false,
		IsHealthy: func(row DeviceRow) bool {
			v, ok := RowFloat(row, "gpuHealth")
			return ok && v == 1
		},
		// The exporter's HealthService fails a card as soon as any
		// uncorrectable ECC counter leaves zero, so say which signal tripped
		// rather than showing a bare unhealthy label.
		HealthReason: func(row DeviceRow) string {
			health, okH := RowFloat(row, "gpuHealth")
			ecc, okE := RowFloat(row, "eccUncorrectTotal")
			if okH && health != 1 && okE && ecc > 0 {
				return "Uncorrectable ECC errors were detected on this GPU"
			}
			return ""
		},
	}
	spec.DeviceLabels.ID = "gpu_uuid"
	spec.DeviceLabels.Node = "node"
	return spec
}

// KsSpecFor returns the KubeSphere spec for a vendor, and ok=false for
// NVIDIA, which reads HAMI instead.
func KsSpecFor(v GPUVendor) (KsVendorSpec, bool) {
	switch v {
	case VendorIntel:
		return intelKsSpec(), true
	case VendorAMD:
		return amdKsSpec(), true
	default:
		return KsVendorSpec{}, false
	}
}
