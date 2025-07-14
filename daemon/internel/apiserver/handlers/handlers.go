package handlers

import (
	"context"
	"fmt"
	"net/http"

	"github.com/asaskevich/govalidator"
	"github.com/beclab/Olares/daemon/internel/ble"
	"github.com/gofiber/fiber/v2"
)

type Handlers struct {
	mainCtx context.Context
	ApList  []ble.AccessPoint
}

var handlers *Handlers = &Handlers{}

func NewHandlers(ctx context.Context) *Handlers {
	handlers.mainCtx = ctx
	return handlers
}

func (h *Handlers) ParseBody(ctx *fiber.Ctx, value any) error {
	err := ctx.BodyParser(value)

	if err != nil {
		return fmt.Errorf("unable to parse body: %w", err)
	}

	valid, err := govalidator.ValidateStruct(value)

	if err != nil {
		return fmt.Errorf("unable to validate body: %w", err)
	}

	if !valid {
		return fmt.Errorf("body is not valid")
	}

	return nil
}

func (h *Handlers) ErrJSON(ctx *fiber.Ctx, code int, message string, data ...interface{}) error {
	switch len(data) {
	case 0:
		return ctx.Status(code).JSON(fiber.Map{
			"code":    code,
			"message": message,
		})
	case 1:
		return ctx.Status(code).JSON(fiber.Map{
			"code":    code,
			"message": message,
			"data":    data[0],
		})
	default:
		return ctx.Status(code).JSON(fiber.Map{
			"code":    code,
			"message": message,
			"data":    data,
		})
	}

}

func (h *Handlers) OkJSON(ctx *fiber.Ctx, message string, data ...interface{}) error {
	return h.ErrJSON(ctx, http.StatusOK, message, data...)
}

func (h *Handlers) NeedChoiceJSON(ctx *fiber.Ctx, message string, data ...interface{}) error {
	return h.ErrJSON(ctx, http.StatusMultipleChoices, message, data...)
}
