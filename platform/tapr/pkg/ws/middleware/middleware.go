package middleware

import (
	"bytetrade.io/web3os/tapr/pkg/constants"
	"bytetrade.io/web3os/tapr/pkg/utils"
	"github.com/gofiber/contrib/websocket"
	"github.com/gofiber/fiber/v2"
	"github.com/google/uuid"
	"k8s.io/klog/v2"
)

func RequireHeader() func(c *fiber.Ctx) error {
	return func(c *fiber.Ctx) error {
		if !websocket.IsWebSocketUpgrade(c) {
			return fiber.ErrUpgradeRequired
		}

		var accessPublic bool = false
		var token = c.Cookies("auth_token")
		if token == "" {
			token = uuid.New().String()
			accessPublic = true
		}

		var headers = c.GetReqHeaders()
		if headers == nil {
			return fiber.ErrBadRequest
		}

		var connId = uuid.New().String()

		var userName = headers[constants.WsHeaderBflUser]
		var userAgent = headers[constants.WsHeaderUserAgent]
		var forwarded = headers[constants.WsHeaderForwardeFor]
		var cookie = headers[constants.WsHeaderCookie]

		klog.Infof("ws-client conn: %s, accessPublic: %v, token: %s, user: %s , header: %+v", connId, accessPublic, token, userName, headers)

		var secWebsocketProtocol, ok = headers[constants.WsHeaderSecWebsocketProtocol]
		if ok {
			c.Set(constants.WsHeaderSecWebsocketProtocol, secWebsocketProtocol)
		}

		c.Locals(constants.WsLocalAccessPublic, accessPublic)
		c.Locals(constants.WsLocalUserKey, userName)
		c.Locals(constants.WsLocalConnIdKey, connId)
		c.Locals(constants.WsLocalTokenKey, utils.MD5(token))
		c.Locals(constants.WsLocalTokenKeyOriginal, token)
		c.Locals(constants.WsLocalUserAgentKey, userAgent)
		c.Locals(constants.WsLocalClientIpKey, forwarded)
		c.Locals(constants.WsLocalCookie, cookie)

		return c.Next()
	}
}
