package amdgpu

import (
	"github.com/beclab/Olares/cli/pkg/bootstrap/precheck"
	"github.com/beclab/Olares/cli/pkg/common"
	"github.com/beclab/Olares/cli/pkg/core/connector"
	"github.com/beclab/Olares/cli/pkg/core/logger"
)

// RocmInstalled checks if AMD ROCm is installed on the system.
type RocmInstalled struct {
	common.KubePrepare
}

func (p *RocmInstalled) PreCheck(runtime connector.Runtime) (bool, error) {
	rocmV, err := connector.RocmVersion()
	if err != nil {
		logger.Debugf("ROCm version check error: %v", err)
		return false, nil
	}
	if rocmV == nil {
		return false, nil
	}

	logger.Infof("Detected ROCm version: %s", rocmV.Original())
	return true, nil
}

// RocmNotInstalled checks if AMD ROCm is NOT installed on the system.
type RocmNotInstalled struct {
	common.KubePrepare
	RocmInstalled
}

func (p *RocmNotInstalled) PreCheck(runtime connector.Runtime) (bool, error) {
	installed, err := p.RocmInstalled.PreCheck(runtime)
	if err != nil {
		return false, err
	}
	return !installed, nil
}

// ContainerdInstalled checks if containerd is installed on the system.
type ContainerdInstalled struct {
	common.KubePrepare
}

func (p *ContainerdInstalled) PreCheck(runtime connector.Runtime) (bool, error) {
	containerdCheck := precheck.ConflictingContainerdCheck{}
	if err := containerdCheck.Check(runtime); err != nil {
		return true, nil
	}

	logger.Info("containerd is not installed, ignore task")
	return false, nil
}
