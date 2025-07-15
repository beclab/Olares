package handlers

import (
	"net/http"

	"github.com/beclab/Olares/cli/pkg/common"
	"github.com/beclab/Olares/daemon/internel/client"
	"github.com/beclab/Olares/daemon/pkg/cluster/state"
	"github.com/beclab/Olares/daemon/pkg/commands"
	"github.com/gofiber/fiber/v2"
	"github.com/joho/godotenv"
)

const (
	SIGNATURE_HEADER = "X-Signature"
)

func (h *Handlers) WaitServerRunning(next func(ctx *fiber.Ctx) error) func(ctx *fiber.Ctx) error {
	return func(ctx *fiber.Ctx) error {
		if state.CurrentState.TerminusdState != state.Running {
			return h.ErrJSON(ctx, http.StatusForbidden, "server is not running, please wait and retry again later")
		}

		return next(ctx)
	}
}

func (h *Handlers) RequireSignature(next func(ctx *fiber.Ctx) error) func(ctx *fiber.Ctx) error {
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

func (h *Handlers) RequireLocal(next func(ctx *fiber.Ctx) error) func(ctx *fiber.Ctx) error {
	return func(ctx *fiber.Ctx) error {
		return next(ctx)
	}
}

func (h *Handlers) RequireOwner(next func(ctx *fiber.Ctx) error) func(ctx *fiber.Ctx) error {
	return func(ctx *fiber.Ctx) error {
		c, ok := ctx.Context().UserValue(client.ClIENT_CONTEXT).(client.Client)
		if !ok {
			return h.ErrJSON(ctx, http.StatusForbidden, "client not found")
		}

		// get owner from release file
		envs, err := godotenv.Read(common.OlaresReleaseFile)
		if err != nil {
			return h.ErrJSON(ctx, http.StatusForbidden, err.Error())
		}

		envOlaresID, ok := envs["OLARES_NAME"]
		if !ok {
			return h.ErrJSON(ctx, http.StatusForbidden, "OLARES_NAME not found")
		}

		if c.OlaresID() != envOlaresID {
			return h.ErrJSON(ctx, http.StatusForbidden, "not the owner of this terminus")
		}

		return next(ctx)
	}
}

func (h *Handlers) RunCommand(next func(ctx *fiber.Ctx, cmd commands.Interface) error,
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
