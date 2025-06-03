package apiserver

import (
	"context"
	"fmt"
	"net/http"

	"bytetrade.io/web3os/terminusd/internel/ble"
	"github.com/asaskevich/govalidator"
	"github.com/gofiber/fiber/v2"
)

type handlers struct {
	mainCtx context.Context
	apList  []ble.AccessPoint
}

func (h *handlers) ParseBody(ctx *fiber.Ctx, value any) error {
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

func (h *handlers) ErrJSON(ctx *fiber.Ctx, code int, message string, data ...interface{}) error {
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

func (h *handlers) OkJSON(ctx *fiber.Ctx, message string, data ...interface{}) error {
	return h.ErrJSON(ctx, http.StatusOK, message, data...)
}

func (h *handlers) NeedChoiceJSON(ctx *fiber.Ctx, message string, data ...interface{}) error {
	return h.ErrJSON(ctx, http.StatusMultipleChoices, message, data...)
}
