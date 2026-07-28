package handlers

import (
	"net/http"

	"github.com/beclab/Olares/daemon/internel/client"
	collectlogs "github.com/beclab/Olares/daemon/pkg/commands/collect_logs"
	"github.com/beclab/Olares/daemon/pkg/utils"
	"github.com/gofiber/fiber/v2"
)

// GetCollectLogsStatus returns the per-runID status of a collection. A normal
// user may only query runs they started; owner/admin may query any.
func (h *Handlers) GetCollectLogsStatus(ctx *fiber.Ctx) error {
	runID := ctx.Params("runID")
	if runID == "" {
		return h.ErrJSON(ctx, http.StatusBadRequest, "missing runID")
	}

	userData, ok := ctx.Context().UserValue(client.USER_CONTEXT).(*utils.ValidToken)
	if !ok || userData == nil {
		return h.ErrJSON(ctx, http.StatusForbidden, "user data not found in context")
	}

	task, found := collectlogs.GetTask(runID)
	if !found {
		// Not a transport error: report a normal state so the client handles a
		// lost/unknown runID uniformly with other states.
		return h.OkJSON(ctx, "success", collectlogs.TaskStatus{
			RunID: runID,
			State: collectlogs.StateNotFound,
		})
	}
	if task.Caller != userData.Username && !userData.IsAdmin() {
		// Likewise report permission as a state, carrying no run details.
		return h.OkJSON(ctx, "success", collectlogs.TaskStatus{
			RunID: runID,
			State: collectlogs.StateForbidden,
		})
	}

	return h.OkJSON(ctx, "success", task)
}
