package handlers

import (
	"errors"
	"net/http"

	"github.com/beclab/Olares/daemon/internel/client"
	"github.com/beclab/Olares/daemon/pkg/commands"
	collectlogs "github.com/beclab/Olares/daemon/pkg/commands/collect_logs"
	"github.com/beclab/Olares/daemon/pkg/utils"
	"github.com/gofiber/fiber/v2"
	"k8s.io/klog/v2"
)

// roleOf maps a verified access token to the role string authz expects.
func roleOf(t *utils.ValidToken) string {
	switch {
	case t.IsOwner():
		return utils.Owner
	case t.IsAdmin():
		return utils.Admin
	default:
		return utils.Normal
	}
}

func (h *Handlers) PostCollectLogs(ctx *fiber.Ctx, cmd commands.Interface) error {
	var param collectlogs.Param
	if len(ctx.Body()) > 0 {
		if err := h.ParseBody(ctx, &param); err != nil {
			return h.ErrJSON(ctx, http.StatusBadRequest, err.Error())
		}
	}

	userData, ok := ctx.Context().UserValue(client.USER_CONTEXT).(*utils.ValidToken)
	if !ok || userData == nil {
		return h.ErrJSON(ctx, http.StatusForbidden, "user data not found in context")
	}
	// Inject the verified caller identity; never trust a body-supplied value.
	param.CallerUsername = userData.Username
	param.CallerRole = roleOf(userData)
	// Forward the verified access token so the orchestrator can authenticate to
	// each node's node-local endpoint as the same caller.
	param.CallerToken = ctx.Get(AUTH_HEADER)

	res, err := cmd.Execute(ctx.Context(), &param)
	if err != nil {
		var denied *collectlogs.ScopeDeniedError
		switch {
		case errors.As(err, &denied):
			return h.ErrJSON(ctx, http.StatusForbidden, "requested scope exceeds caller permission", denied.Denied)
		case errors.Is(err, collectlogs.ErrNothingRequested):
			return h.ErrJSON(ctx, http.StatusBadRequest, err.Error())
		default:
			klog.Error("execute command error, ", err, ", ", cmd.OperationName().Stirng())
			return h.ErrJSON(ctx, http.StatusInternalServerError, err.Error())
		}
	}

	return h.OkJSON(ctx, "success to exec command", res)
}
