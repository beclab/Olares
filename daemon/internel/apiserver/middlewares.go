package apiserver

import (
	"net/http"

	"bytetrade.io/web3os/terminusd/internel/client"
	"bytetrade.io/web3os/terminusd/pkg/cluster/state"
	"bytetrade.io/web3os/terminusd/pkg/commands"
	"github.com/gofiber/fiber/v2"
)

const (
	SIGNATURE_HEADER = "X-Signature"
)

func (h *handlers) WaitServerRunning(next func(ctx *fiber.Ctx) error) func(ctx *fiber.Ctx) error {
	return func(ctx *fiber.Ctx) error {
		if state.CurrentState.TerminusdState != state.Running {
			return h.ErrJSON(ctx, http.StatusForbidden, "server is not running, please wait and retry again later")
		}

		return next(ctx)
	}
}

func (h *handlers) RequireSignature(next func(ctx *fiber.Ctx) error) func(ctx *fiber.Ctx) error {
	return func(ctx *fiber.Ctx) error {
		headers := ctx.GetReqHeaders()
		signature, ok := headers[SIGNATURE_HEADER]
		if !ok || len(signature) == 0 {
			return h.ErrJSON(ctx, http.StatusForbidden, "request is forbidden")
		}

		if c, err := client.NewTermipassClient(ctx.Context(), signature[0]); err != nil {
			return h.ErrJSON(ctx, http.StatusForbidden, err.Error())
		} else {
			// store client in the context, will be used in the next phase.
			ctx.Context().SetUserValue(client.ClIENT_CONTEXT, c)
		}

		return next(ctx)
	}
}

func (h *handlers) RunCommand(next func(ctx *fiber.Ctx, cmd commands.Interface) error,
	cmdNew func() commands.Interface) func(ctx *fiber.Ctx) error {

	return func(ctx *fiber.Ctx) error {
		c := cmdNew()
		err := state.CurrentState.TerminusState.ValidateOp(c)
		if err != nil {
			return h.ErrJSON(ctx, http.StatusForbidden, err.Error())
		}

		return next(ctx, c)
	}
}
