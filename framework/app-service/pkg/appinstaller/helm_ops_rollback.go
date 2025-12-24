package appinstaller

import (
	"errors"

	"github.com/beclab/Olares/framework/app-service/pkg/helm"
	"helm.sh/helm/v3/pkg/action"
)

// RollBack do a rollback for release if it can be rollback.
func (h *HelmOps) RollBack() error {
	can, err := h.canRollBack()
	if err != nil {
		return err
	}
	if can {
		return h.rollBack()
	}
	return errors.New("can not do rollback")
}

func (h *HelmOps) canRollBack() (bool, error) {
	client := action.NewGet(h.actionConfig)
	release, err := client.Run(h.app.AppName)
	if err != nil {
		return false, err
	}
	if release.Version > 1 {
		return true, nil
	}
	return false, nil
}

// rollBack to previous version
func (h *HelmOps) rollBack() error {
	err := helm.RollbackCharts(h.actionConfig, h.app.AppName)
	if err != nil {
		return err
	}
	return nil
}
