package handlers

import (
	"encoding/json"
	"fmt"
	"net/http"
	"strings"

	"github.com/beclab/Olares/daemon/internel/client"
	"github.com/beclab/Olares/daemon/pkg/cluster/clusterop"
	"github.com/beclab/Olares/daemon/pkg/cluster/inventory"
	"github.com/beclab/Olares/daemon/pkg/cluster/state"
	"github.com/beclab/Olares/daemon/pkg/commands"
	"github.com/beclab/Olares/daemon/pkg/utils"
	"github.com/gofiber/fiber/v2"
	"k8s.io/klog/v2"
)

const (
	SIGNATURE_HEADER = "X-Signature"
	AUTH_HEADER      = "X-Authorization"
)

// ownerClient is what a verified signature yields: the Olares ID that signed
// the request.
type ownerClient = client.Client

// These are indirected so a route's middleware chain can be exercised without
// an identity provider. The chain itself is never bypassed.
var (
	validateAccessToken = utils.AccessTokenValidate
	newTermipassClient  = client.NewTermipassClient
	olaresIDFromRelease = utils.GetOlaresNameFromReleaseFile
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

		if c, err := newTermipassClient(ctx.Context(), signature[0]); err != nil {
			return h.ErrJSON(ctx, http.StatusForbidden, err.Error())
		} else {
			// store client in the context, will be used in the next phase.
			ctx.Context().SetUserValue(client.ClIENT_CONTEXT, c)
		}

		return next(ctx)
	}
}

// RequireSignatureForRegisteredClusterOp admits POST /cluster/operations with
// an owner signature when the requested type has registered itself as needing
// one, and with an access token otherwise. An empty type fails closed to the
// signature path. The signature verification itself is unchanged.
func (h *Handlers) RequireSignatureForRegisteredClusterOp(next func(ctx *fiber.Ctx) error) func(ctx *fiber.Ctx) error {
	return func(ctx *fiber.Ctx) error {
		var peek struct {
			Type string `json:"type"`
		}
		_ = json.Unmarshal(ctx.Body(), &peek)
		typ := strings.TrimSpace(peek.Type)
		if typ == "" || clusterop.RequiresSignature(clusterop.Type(typ)) {
			return h.RequireSignature(next)(ctx)
		}
		return h.RequireAuthorization(next)(ctx)
	}
}

func (h *Handlers) RequireLocal(next func(ctx *fiber.Ctx) error) func(ctx *fiber.Ctx) error {
	return func(ctx *fiber.Ctx) error {
		return next(ctx)
	}
}

func (h *Handlers) RequireAuthorizationOrOwnerSignature(next fiber.Handler) fiber.Handler {
	return func(ctx *fiber.Ctx) error {
		if ctx.Get(AUTH_HEADER) != "" {
			return h.RequireAuthorization(next)(ctx)
		}
		return h.RequireSignature(h.RequireOwner(next))(ctx)
	}
}

func (h *Handlers) RequireOwner(next func(ctx *fiber.Ctx) error) func(ctx *fiber.Ctx) error {
	return func(ctx *fiber.Ctx) error {
		userData, ok := ctx.Context().UserValue(client.USER_CONTEXT).(*utils.ValidToken)
		if ok && userData != nil {
			if userData.IsOwner() {
				return next(ctx)
			} else {
				return h.ErrJSON(ctx, http.StatusForbidden, "not the owner of this Olares")
			}
		}

		c, ok := ctx.Context().UserValue(client.ClIENT_CONTEXT).(client.Client)
		if !ok {
			return h.ErrJSON(ctx, http.StatusForbidden, "client not found")
		}

		// get owner from release file
		envOlaresID, err := olaresIDFromRelease()
		if err != nil {
			return h.ErrJSON(ctx, http.StatusInternalServerError, fmt.Sprintf("failed to get Olares ID from release file: %v", err))
		}
		envOlaresID = strings.TrimSpace(envOlaresID)

		if envOlaresID == "" {
			// A node that joined an existing cluster has no OLARES_NAME in its
			// release file: that file is written where the activation happened,
			// and every later node is installed without one. The published
			// state snapshot carries the same Olares ID, resolved from the
			// cluster, which is the only place such a node can learn it.
			if snap, _ := currentStateSnapshot(); snap.TerminusName != nil {
				envOlaresID = strings.TrimSpace(*snap.TerminusName)
			}
		}

		if envOlaresID == "" {
			if isInstalled, err := state.IsTerminusInstalled(); err != nil {
				return h.ErrJSON(ctx, http.StatusInternalServerError, fmt.Sprintf("failed to check if Olares is installed: %v", err))
			} else if !isInstalled {
				// Nothing is installed, so there is no owner to be yet: this is
				// the install-time path, reached by whoever is setting the
				// machine up.
				return next(ctx)
			}
			// Olares is installed and neither source names an owner. Admitting
			// the request would make every signature the owner's on the one
			// node that cannot tell the difference.
			return h.ErrJSON(ctx, http.StatusForbidden, "cannot determine the owner of this Olares")
		}

		if c.OlaresID() != envOlaresID {
			return h.ErrJSON(ctx, http.StatusForbidden, "not the owner of this Olares")
		}

		return next(ctx)
	}
}

func (h *Handlers) RunCommand(next func(ctx *fiber.Ctx, cmd commands.Interface) error,
	cmdNew func() commands.Interface) func(ctx *fiber.Ctx) error {

	return func(ctx *fiber.Ctx) error {
		c := cmdNew()
		err := state.ValidateOp(state.CurrentState.TerminusState, c)
		if err != nil {
			return h.ErrJSON(ctx, http.StatusForbidden, err.Error())
		}

		return next(ctx, c)
	}
}

func (h *Handlers) RequireMaster(next func(ctx *fiber.Ctx) error) func(ctx *fiber.Ctx) error {
	return func(ctx *fiber.Ctx) error {
		switch state.CurrentState.TerminusState {
		case state.NotInstalled, state.Uninitialized, state.InitializeFailed, state.IPChanging:
			return h.ErrJSON(ctx, http.StatusForbidden, fmt.Sprintf("operation is not allowed in current state: %v", state.CurrentState.TerminusState))
		default:
			_, role, err := thisNodeInCluster(ctx.Context())
			if err != nil {
				klog.Error("get this node role error, ", err)
				return h.ErrJSON(ctx, http.StatusInternalServerError, "failed to get this node role")
			}

			if role != inventory.RoleMaster {
				return h.ErrJSON(ctx, http.StatusForbidden, "operation is only allowed on master node")
			}
		}
		return next(ctx)
	}
}

func (h *Handlers) RequireAuthorization(next func(ctx *fiber.Ctx) error) func(ctx *fiber.Ctx) error {
	return func(ctx *fiber.Ctx) error {
		authHeader := ctx.Get(AUTH_HEADER)
		if authHeader == "" {
			return h.ErrJSON(ctx, http.StatusForbidden, "authorization header is missing")
		}

		valid, tokenData, err := validateAccessToken(authHeader)
		if err != nil {
			return h.ErrJSON(ctx, http.StatusForbidden, fmt.Sprintf("invalid token: %v", err))
		}
		if !valid {
			return h.ErrJSON(ctx, http.StatusForbidden, "unauthorized token")
		}

		ctx.Context().SetUserValue(client.USER_CONTEXT, tokenData)
		return next(ctx)
	}
}

func (h *Handlers) RequireAdmin(next func(ctx *fiber.Ctx) error) func(ctx *fiber.Ctx) error {
	return func(ctx *fiber.Ctx) error {
		userData, ok := ctx.Context().UserValue(client.USER_CONTEXT).(*utils.ValidToken)
		if !ok || userData == nil {
			return h.ErrJSON(ctx, http.StatusForbidden, "user data not found in context")
		}

		if !userData.IsAdmin() {
			return h.ErrJSON(ctx, http.StatusForbidden, "operation is only allowed for admin users")
		}

		return next(ctx)
	}
}
