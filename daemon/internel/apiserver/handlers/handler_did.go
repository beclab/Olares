package handlers

import (
	"net/http"

	"github.com/beclab/Olares/cli/pkg/web5/jws"
	"github.com/gofiber/fiber/v2"
	"k8s.io/klog/v2"
)

func (h *Handlers) ResolveOlaresName(c *fiber.Ctx) error {
	olaresName := c.Params("olaresName")
	if olaresName == "" {
		klog.Error("olaresName parameter is missing")
		return h.ErrJSON(c, fiber.StatusBadRequest, "olaresName parameter is required")
	}
	klog.Infof("Received olaresName: %s", olaresName)
	result, err := jws.ResolveOlaresName(olaresName)
	if err != nil {
		klog.Errorf("Failed to resolve DID for %s: %v", olaresName, err)
		return h.ErrJSON(c, fiber.StatusInternalServerError, "Failed to resolve DID")
	}
	// return DID protocol resolution result
	return c.Status(http.StatusOK).JSON(result)
}

func (h *Handlers) CheckJWS(c *fiber.Ctx) error {
	// Get JWS from request body
	// Parse request body
	var body struct {
		JWS      string `json:"jws"`
		Duration int64  `json:"duration"` // in milliseconds
	}

	if err := c.BodyParser(&body); err != nil {
		klog.Errorf("Failed to parse request body: %v", err)
		return h.ErrJSON(c, fiber.StatusBadRequest, "Invalid request body format")
	}

	if body.JWS == "" {
		klog.Error("JWS is missing in request body")
		return h.ErrJSON(c, fiber.StatusBadRequest, "JWS is required in request body")
	}

	if body.Duration == 0 {
		body.Duration = int64(3 * 60 * 1000) // 3 minutes in milliseconds
	}

	result, err := jws.CheckJWS(body.JWS, body.Duration)
	if err != nil {
		klog.Errorf("Failed to check JWS: %v", err)
		return h.ErrJSON(c, fiber.StatusBadRequest, "Invalid JWS")
	}

	return h.OkJSON(c, "success", result)
}
