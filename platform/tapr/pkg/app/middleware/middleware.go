package middleware

import (
	"fmt"
	"net/http"

	"bytetrade.io/web3os/tapr/pkg/constants"
	"bytetrade.io/web3os/tapr/pkg/kubesphere"
	"bytetrade.io/web3os/tapr/pkg/vault/infisical"
	"github.com/gofiber/fiber/v2"
	"k8s.io/client-go/rest"
)

func RequireAuth(kubeconfig *rest.Config, next func(c *fiber.Ctx) error) func(c *fiber.Ctx) error {
	return func(c *fiber.Ctx) error {
		authToken := c.Get(constants.AuthorizationTokenKey, "")
		if authToken == "" {
			authToken = c.Cookies(constants.AuthorizationTokenCookieKey)

			if authToken == "" {
				return c.JSON(fiber.Map{"code": http.StatusUnauthorized, "message": "Auth token not found", "data": nil})
			}
		}

		username, err := kubesphere.ValidateToken(c.UserContext(), kubeconfig, authToken)
		if err != nil {
			return c.JSON(fiber.Map{
				"code":    http.StatusUnauthorized,
				"message": fmt.Sprintf("Auth token invalid, %s", err.Error()),
				"data":    nil,
			})
		}

		user, err := kubesphere.GetUser(c.UserContext(), kubeconfig, username)
		if err != nil {
			return c.JSON(fiber.Map{
				"code":    http.StatusUnauthorized,
				"message": fmt.Sprintf("User info error, %s", err.Error()),
				"data":    nil,
			})
		}

		c.Context().SetUserValueBytes(constants.UsernameCtxKey, username)
		c.Context().SetUserValueBytes(constants.UserEmailCtxKey, user.Spec.Email)
		c.Context().SetUserValueBytes([]byte(constants.AuthorizationTokenKey), authToken)

		// FIXME:
		// c.Context().SetUserValueBytes(constants.UserPwdCtxKey, user.Spec.EncryptedPassword)
		c.Context().SetUserValueBytes(constants.UserPwdCtxKey, infisical.Password)
		return next(c)
	}
}

func RequireAdmin(kubeconfig *rest.Config, next func(c *fiber.Ctx) error) func(c *fiber.Ctx) error {
	return func(c *fiber.Ctx) error {
		user := c.Context().UserValueBytes(constants.UsernameCtxKey)

		role, err := kubesphere.GetUserRole(c.UserContext(), kubeconfig, user.(string))
		if err != nil {
			return c.JSON(fiber.Map{
				"code":    http.StatusInternalServerError,
				"message": fmt.Sprintf("User info error, %s", err.Error()),
				"data":    nil,
			})
		}

		if role != "owner" && role != "admin" {
			return c.JSON(fiber.Map{
				"code":    http.StatusForbidden,
				"message": "Must be platform admin",
				"data":    nil,
			})
		}

		return next(c)
	}
}

// func GetOwnerInfo(kubeconfig *rest.Config, owner string, next func(c *fiber.Ctx) error) func(c *fiber.Ctx) error {
// 	return func(c *fiber.Ctx) error {
// 		user, err := kubesphere.GetUser(c.UserContext(), kubeconfig, owner)
// 		if err != nil {
// 			return c.JSON(fiber.Map{
// 				"code":    http.StatusUnauthorized,
// 				"message": fmt.Sprintf("User info error, %s", err.Error()),
// 				"data":    nil,
// 			})
// 		}

// 		c.Context().SetUserValueBytes(constants.UsernameCtxKey, owner)
// 		c.Context().SetUserValueBytes(constants.UserEmailCtxKey, user.Spec.Email)

// 		// FIXME:
// 		// c.Context().SetUserValueBytes(constants.UserPwdCtxKey, user.Spec.EncryptedPassword)
// 		c.Context().SetUserValueBytes(constants.UserPwdCtxKey, infisical.Password)
// 		return next(c)
// 	}
// }

func GetUserInfo(kubeconfig *rest.Config, next func(c *fiber.Ctx) error) func(c *fiber.Ctx) error {
	return func(c *fiber.Ctx) error {
		userName := c.Get(constants.WsHeaderBflUser, "")
		if userName == "" {
			return c.JSON(fiber.Map{
				"code":    http.StatusUnauthorized,
				"message": "User name not found in header",
				"data":    nil,
			})
		}

		user, err := kubesphere.GetUser(c.UserContext(), kubeconfig, userName)
		if err != nil {
			return c.JSON(fiber.Map{
				"code":    http.StatusUnauthorized,
				"message": fmt.Sprintf("User info error, %s", err.Error()),
				"data":    nil,
			})
		}

		c.Context().SetUserValueBytes(constants.UsernameCtxKey, userName)
		c.Context().SetUserValueBytes(constants.UserEmailCtxKey, user.Spec.Email)

		// FIXME:
		// c.Context().SetUserValueBytes(constants.UserPwdCtxKey, user.Spec.EncryptedPassword)
		c.Context().SetUserValueBytes(constants.UserPwdCtxKey, infisical.Password)
		return next(c)
	}
}
