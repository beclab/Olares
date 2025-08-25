package gpu

import (
	"fmt"
	"strings"

	"github.com/beclab/Olares/cli/pkg/common"
	"github.com/beclab/Olares/cli/pkg/core/logger"
	"github.com/beclab/Olares/cli/pkg/core/pipeline"

	"github.com/beclab/Olares/cli/pkg/gpu/amd"
	// import your existing NVIDIA package here; assumed path:
	nvidia "github.com/beclab/Olares/cli/pkg/gpu/nvidia"
)

type Vendor string

const (
	VendorAuto   Vendor = "auto"
	VendorAMD    Vendor = "amd"
	VendorNVIDIA Vendor = "nvidia"
)

type InstallOptions struct {
	Vendor Vendor // auto|amd|nvidia
}

func Install(runtime common.KubeRuntime, opt InstallOptions) error {
	v := strings.ToLower(string(opt.Vendor))
	if v == "" {
		v = string(VendorAuto)
	}
	logger.Infof("GPU install requested (vendor=%s)", v)

	switch Vendor(v) {
	case VendorAMD:
		return runAmd(runtime)
	case VendorNVIDIA:
		return runNvidia(runtime)
	default:
		// auto-detect on first control-plane
		first := runtime.GetFirstMaster()
		r := runtime.GetConnector().GetRuntime(first)

		amdFound := amd.HasAmdKernelBits(r) || amd.HasAmdPci(r)
		nvFound := hasNvidia(r)

		if !amdFound && !nvFound {
			return fmt.Errorf("no AMD or NVIDIA GPU detected on first control-plane node")
		}
		if nvFound {
			if err := runNvidia(runtime); err != nil {
				return err
			}
		}
		if amdFound {
			if err := runAmd(runtime); err != nil {
				return err
			}
		}
		return nil
	}
}

// ---- NVIDIA detection (mirror your existing helpers if you have them) ----
func hasNvidia(r common.Runtime) bool {
	out, _ := r.GetRunner().SudoCmd(`lsmod | grep -q '^nvidia' && echo yes || true`, false, false)
	if strings.Contains(out, "yes") {
		return true
	}
	out, _ = r.GetRunner().SudoCmd(`lspci -nn | egrep -i 'vga|3d|display' | grep -qi nvidia && echo yes || true`, false, false)
	return strings.Contains(out, "yes")
}

// ---- runners ----
func runAmd(runtime common.KubeRuntime) error {
	var m amd.Module
	m.Runtime = runtime
	m.Init()
	return pipeline.Run(&m)
}

func runNvidia(runtime common.KubeRuntime) error {
	var m nvidia.Module
	m.Runtime = runtime
	m.Init()
	return pipeline.Run(&m)
}
