package apiserver

import (
	"bytetrade.io/web3os/terminusd/pkg/cluster/state"
	"fmt"
	"github.com/Masterminds/semver/v3"
	"net/http"

	"bytetrade.io/web3os/terminusd/pkg/commands"
	"github.com/gofiber/fiber/v2"
	"k8s.io/klog/v2"
)

type UpgradeReq struct {
	Version string `json:"version"`
}

func (r *UpgradeReq) Check() error {
	target, err := semver.NewVersion(r.Version)
	if err != nil {
		return fmt.Errorf("invalid target version %s: %v", r.Version, err)
	}
	if state.CurrentState.TerminusVersion != nil {
		current, err := semver.NewVersion(*state.CurrentState.TerminusVersion)
		if err != nil {
			return fmt.Errorf("invalid current version %s: %v", *state.CurrentState.TerminusVersion, err)
		}
		if !current.LessThan(target) {
			return fmt.Errorf("target version should be greater than current version: %s", *state.CurrentState.TerminusVersion)
		}
	}
	return nil
}

func (h *handlers) RequestOlaresUpgrade(ctx *fiber.Ctx, cmd commands.Interface) error {
	var req UpgradeReq
	if err := h.ParseBody(ctx, &req); err != nil {
		klog.Error("parse request error, ", err)
		return h.ErrJSON(ctx, http.StatusBadRequest, err.Error())
	}
	if err := req.Check(); err != nil {
		return h.ErrJSON(ctx, http.StatusBadRequest, err.Error())
	}

	if _, err := cmd.Execute(ctx.Context(), req.Version); err != nil {
		return h.ErrJSON(ctx, http.StatusBadRequest, err.Error())
	}

	return h.OkJSON(ctx, "successfully created upgrade target")
}

func (h *handlers) CancelOlaresUpgrade(ctx *fiber.Ctx, cmd commands.Interface) error {
	if _, err := cmd.Execute(ctx.Context(), nil); err != nil {
		return h.ErrJSON(ctx, http.StatusBadRequest, err.Error())
	}

	return h.OkJSON(ctx, "successfully removed upgrade target")
}
