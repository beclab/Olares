package controllers

import (
	"net/http"

	"bytetrade.io/web3os/tapr/pkg/constants"
	"bytetrade.io/web3os/tapr/pkg/vault/infisical"
	"github.com/gofiber/fiber/v2"
)

func FetchUserPrivateKey(clientset *Clientset, next func(c *fiber.Ctx) error) func(c *fiber.Ctx) error {
	return func(c *fiber.Ctx) error {
		user := c.Context().UserValueBytes(constants.UserCtxKey).(*infisical.UserEncryptionKeysPG)
		password := c.Context().UserValueBytes(constants.UserPwdCtxKey).(string)
		userPrivateKey, err := clientset.GetUserPrivateKey(user, password)
		if err != nil {
			return c.JSON(fiber.Map{
				"code":    http.StatusInternalServerError,
				"message": "get user private key error, " + err.Error(),
			})
		}

		c.Context().SetUserValueBytes(constants.UserPrivateKeyCtxKey, userPrivateKey)

		return next(c)
	}
}

func FetchUserOrganizationId(clientset *Clientset, next func(c *fiber.Ctx) error) func(c *fiber.Ctx) error {
	return func(c *fiber.Ctx) error {
		orgId := c.Context().UserValueBytes(constants.UserOrganizationIdCtxKey)
		if orgId == nil {
			token := c.Context().UserValueBytes(constants.UserAuthTokenCtxKey)

			orgId, err := clientset.GetUserOrganizationId(token.(string))
			if err != nil {
				return c.JSON(fiber.Map{
					"code":    http.StatusInternalServerError,
					"message": "get user organization error, " + err.Error(),
				})
			}

			c.Context().SetUserValueBytes(constants.UserOrganizationIdCtxKey, orgId)
		}
		return next(c)
	}
}
