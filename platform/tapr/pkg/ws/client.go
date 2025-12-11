package ws

import (
	"context"
	"crypto/md5"
	"encoding/json"
	"fmt"
	"time"

	"bytetrade.io/web3os/tapr/pkg/constants"
	"github.com/gofiber/contrib/websocket"
	"k8s.io/klog/v2"
)

const (
	ACTION_OPEN    = "open"
	ACTION_CLOSE   = "close"
	ACTION_MESSAGE = "message"
)

type Client struct {
	ctx          context.Context
	cancel       context.CancelFunc
	conn         *websocket.Conn
	closeHandler func(connId string)
	writeHandler func(connId string, msgType int, message interface{})
	readHandler  func(anonymous bool, token, connId, userName string, message interface{}, cookie string, action string)
}

func (client *Client) setLocals() *Client {
	client.conn.Locals(constants.WsLocalClientAddrKey, client.conn.NetConn().RemoteAddr().String())
	client.conn.Locals(constants.WsLocaExpiredKey, time.Now().Unix())

	return client
}

func (client *Client) noticeConnected(c *Client) *Client {
	var accessPublic = c.getAccessLevel()
	var userName = c.getUser() // userName or token(uuid)
	var connId = c.getConnId()
	var cookie = c.getCookie()
	var token = c.getTokenOriginal() // auth-token or token(uuid)
	client.readHandler(accessPublic, token, connId, userName, struct{}{}, cookie, ACTION_OPEN)
	return c
}

func (client *Client) onConnection() {
	var (
		mt  int
		msg []byte
		err error
	)

	var accessPublic = client.getAccessLevel()
	var connId = client.getConnId()
	var userName = client.getUser()
	var cookie = client.getCookie()
	var token = client.getTokenOriginal()

	for {
		select {
		case <-client.ctx.Done():
			if err := client.conn.Close(); err != nil {
				klog.Errorf("websocket connection close error, id: %s, accessPublic: %v, err: %v", connId, accessPublic, err)
			} else {
				klog.Infof("websocket connection closed, id: %s, accessPublic: %v", connId, accessPublic)
			}
			return
		default:
			if mt, msg, err = client.conn.ReadMessage(); err != nil || mt < 1 {
				klog.Infof("read message invalid, id: %s, accessPublic: %v, type: %d, closed: %v", connId, accessPublic, mt, err)
				client.readHandler(accessPublic, token, connId, userName, struct{}{}, cookie, ACTION_CLOSE)
				client.closeHandler(connId)
				return
			}

			klog.Infof("read message, type: %d, accessPublic: %v, connId: %s, user: %s, data: %s", mt, accessPublic, connId, userName, string(msg))

			if client.checkPingMessage(msg) {
				client.writeHandler(connId, websocket.TextMessage, map[string]interface{}{"event": "pong"})
				client.updateExpiration()
				continue
			}

			var data = map[string]interface{}{}
			if err = json.Unmarshal(msg, &data); err != nil {
				klog.Errorf("unmarshal message error %+v, data: %s", err, string(msg))
			}

			client.readHandler(accessPublic, token, connId, userName, data, cookie, ACTION_MESSAGE)
		}
	}
}

func (client *Client) close() {
	client.cancel()
}

func (client *Client) isExpired() bool {
	expirationTime, ok := client.conn.Locals(constants.WsLocaExpiredKey).(int64)
	if !ok || expirationTime == 0 {
		return true
	}

	currentTime := time.Now().Unix()
	return (currentTime - expirationTime) > expirationDuration
}

func (client *Client) getConnId() string {
	return client.conn.Locals(constants.WsLocalConnIdKey).(string)
}

func (client *Client) getToken() string {
	return client.conn.Locals(constants.WsLocalTokenKey).(string)
}

func (client *Client) getTokenOriginal() string {
	return client.conn.Locals(constants.WsLocalTokenKeyOriginal).(string)
}

func (client *Client) getAccessLevel() bool {
	return client.conn.Locals(constants.WsLocalAccessPublic).(bool)
}

func (client *Client) getUser() string {
	return client.conn.Locals(constants.WsLocalUserKey).(string)
}

func (client *Client) getUserAgent() string {
	return client.conn.Locals(constants.WsLocalUserAgentKey).(string)
}

func (client *Client) getCookie() string {
	return client.conn.Locals(constants.WsLocalCookie).(string)
}

func (client *Client) md5(b []byte) string {
	if b == nil || len(b) == 0 {
		return ""
	}
	h := md5.New()
	_, _ = h.Write(b)
	return fmt.Sprintf("%x", h.Sum(nil))
}

func (client *Client) checkPingMessage(msg []byte) bool {
	var err error
	var data = map[string]interface{}{}
	if err = json.Unmarshal(msg, &data); err != nil {
		return false
	}
	var m, ok = data["event"].(string)
	if !ok {
		return false
	}

	return m == "ping"
}

func (client *Client) updateExpiration() {
	client.conn.Locals(constants.WsLocaExpiredKey, time.Now().Unix())
}
