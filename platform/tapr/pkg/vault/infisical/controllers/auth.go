package controllers

import (
	"errors"

	"bytetrade.io/web3os/tapr/pkg/constants"
	"bytetrade.io/web3os/tapr/pkg/vault/infisical"
	"github.com/gofiber/fiber/v2"
)

type authController struct {
}

func (a *authController) AuthToken(c *fiber.Ctx) error {
	token := c.Context().UserValueBytes(constants.UserAuthTokenCtxKey)
	if token == nil {
		return errors.New("no auth token")
	}

	return c.JSON(fiber.Map{
		"token":     token,
		"secretKey": infisical.Password,
	})
}

func (a *authController) PrivateKey(c *fiber.Ctx) error {
	token := c.Context().UserValueBytes(constants.UserAuthTokenCtxKey)
	if token == nil {
		return errors.New("no auth token")
	}

	user := c.Context().UserValueBytes(constants.UserCtxKey)
	if user == nil {
		return errors.New("no user data")
	}

	return c.JSON(user)
}
