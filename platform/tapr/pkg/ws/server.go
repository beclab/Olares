package ws

import (
	"context"
	"encoding/json"
	"sync"
	"time"

	"github.com/gofiber/contrib/websocket"
	"github.com/gofiber/fiber/v2"
	"k8s.io/klog/v2"
)

type callback func(data *ReadMessage)

type WebSocketServer interface {
	New() func(c *fiber.Ctx) error
	SetHandler(cb callback)

	List() []map[string]interface{}
	Close(connIds []string, tokens []string, users []string, usersWithPublic int)
	Push(connId string, tokens []string, users []string, usersWithPublic int, message interface{})
}

type Server struct {
	queue struct {
		read  chan *ReadMessage  // from client
		write chan *WriteMessage // from app
		close chan *CloseMessage // from app
	}

	handler callback

	users map[string]*User

	sync.RWMutex
}

type Public struct {
	id    string // connId
	token string
	conn  *Client
}

type User struct {
	name  string             // userName
	conns map[string]*Client // connId / client
	sync.RWMutex
}

func NewWebSocketServer() WebSocketServer {
	var server = &Server{}
	server.users = map[string]*User{}
	server.queue.read = make(chan *ReadMessage, queueSize)
	server.queue.write = make(chan *WriteMessage, queueSize)
	server.queue.close = make(chan *CloseMessage, queueSize)

	go server.routineRead()
	go server.routineWrite()
	go server.routineClose()
	go server.checkExpired()

	return server
}

func (server *Server) New() func(c *fiber.Ctx) error {
	return websocket.New(func(c *websocket.Conn) {
		ctx, cancelFunc := context.WithCancel(context.Background())

		var client = &Client{
			conn:         c,
			ctx:          ctx,
			cancel:       cancelFunc,
			closeHandler: server.close,
			readHandler:  server.read,
			writeHandler: server.write,
		}

		client.setLocals()

		server.addClient(client).noticeConnected(client).onConnection()
	})
}

func (server *Server) SetHandler(cb callback) {
	server.handler = cb
}

func (server *Server) List() []map[string]interface{} {
	server.RLock()
	defer server.RUnlock()

	var users = []map[string]interface{}{}
	for _, z := range server.users {
		if z == nil {
			continue
		}

		var publics = []map[string]string{}
		var ccs = []map[string]string{}
		var r = map[string]interface{}{}
		r["name"] = z.name
		z.RLock()
		for _, c := range z.conns {
			accessPublic := c.getAccessLevel()
			tokenOriginal := c.getTokenOriginal()
			connId := c.getConnId()
			userAgent := c.getUserAgent()
			userAgentTag := c.md5([]byte(userAgent))

			var cs = map[string]string{}
			cs["id"] = connId
			cs["token"] = tokenOriginal
			cs["userAgent"] = userAgent
			cs["userAgentTag"] = userAgentTag
			if accessPublic {
				publics = append(publics, cs)
			} else {
				ccs = append(ccs, cs)
			}

		}
		r["conns"] = ccs
		r["conns_number"] = len(ccs)
		r["public_conns"] = publics
		r["public_conns_number"] = len(publics)
		users = append(users, r)
		z.RUnlock()
	}

	return users
}

func (server *Server) Close(connIds []string, tokens []string, users []string, usersAccessType int) {
	var m = &CloseMessage{
		ConnIds:         connIds,
		Tokens:          tokens,
		Users:           users,
		UsersAccessType: usersAccessType,
	}

	server.queue.close <- m
}

func (server *Server) Push(connId string, tokens []string, users []string, usersAccessType int, message interface{}) {
	var m = &WriteMessage{
		MessageType:     websocket.TextMessage,
		ConnId:          connId,
		Tokens:          tokens,
		Users:           users,
		UsersAccessType: usersAccessType,
		Message:         message,
	}

	server.queue.write <- m
}

func (server *Server) addClient(c *Client) *Client {
	server.Lock()
	defer server.Unlock()

	var userName = c.getUser()
	var connId = c.getConnId()

	user, ok := server.users[userName]
	if !ok {
		var newUser = &User{conns: map[string]*Client{}}
		newUser.Lock()
		newUser.name = userName
		newUser.conns[connId] = c
		server.users[userName] = newUser
		newUser.Unlock()
		return c
	}

	user.Lock()
	user.conns[connId] = c
	user.Unlock()

	return c
}

func (server *Server) close(connId string) {
	server.queue.close <- &CloseMessage{
		ConnIds: []string{connId},
	}
}

func (server *Server) routineClose() {
	for {
		select {
		case elem, ok := <-server.queue.close:
			if !ok {
				server.queue.close = make(chan *CloseMessage, queueSize)
				continue
			}
			server.closeConns(elem.ConnIds, elem.Tokens, elem.Users, elem.UsersAccessType)
		}
	}
}

func (server *Server) closeConns(connIds []string, tokens []string, users []string, usersAccessType int) {
	var filter = NewFilter(server)

	if users != nil && len(users) > 0 {
		filter.FilterByUsers(users)
		if usersAccessType == 1 {
			filter.FilterByUsersPrivatesOnly()
		} else if usersAccessType == 2 {
			filter.FilterByUsersPublicsOnly()
		}
	}
	if tokens != nil && len(tokens) > 0 {
		filter.FilterByTokens(tokens)
	}
	if connIds != nil && len(connIds) > 0 {
		filter.FilterByConnIds(connIds)
	}

	var result = filter.Result()
	if result == nil || len(result) == 0 {
		return
	}

	var removeusers []string

	server.Lock()
	for userName, userClients := range server.users {
		userClients.Lock()
		for _, connId := range result {
			client, ok := userClients.conns[connId]
			if ok && client.conn != nil {
				delete(userClients.conns, connId)
				client.close()
			}
		}
		if userClients.conns == nil || len(userClients.conns) == 0 {
			removeusers = append(removeusers, userName)
		}
		userClients.Unlock()
	}

	for _, removeuser := range removeusers {
		delete(server.users, removeuser)
	}

	server.Unlock()
}

func (server *Server) read(accessPublic bool, token, connId, userName string, message interface{}, cookie string, action string) {
	server.queue.read <- &ReadMessage{
		AccessPublic: accessPublic,
		Token:        token,
		ConnId:       connId,
		UserName:     userName,
		Data:         message,
		Action:       action,
		Cookie:       cookie,
	}
}

func (server *Server) routineRead() {
	for {
		select {
		case elem, ok := <-server.queue.read:
			if !ok {
				server.queue.read = make(chan *ReadMessage, queueSize)
				continue
			}
			server.handler(elem)
		}
	}
}

func (server *Server) write(connId string, msgType int, data interface{}) {
	var w = &WriteMessage{
		MessageType: msgType,
		Message:     data,
		ConnId:      connId,
	}
	server.queue.write <- w
}

func (server *Server) routineWrite() {
	for {
		select {
		case elem, ok := <-server.queue.write:
			if !ok {
				server.queue.write = make(chan *WriteMessage, queueSize)
				continue
			}
			msg, err := json.Marshal(elem.Message)
			if err != nil {
				klog.Errorf("send message marshal error %+v, data: %v", err, elem.Message)
				continue
			}

			klog.Infof("send message data: %s, connId: %s, token: %v, users: %v", string(msg), elem.ConnId, elem.Tokens, elem.Users)

			var filter = NewFilter(server)
			if elem.Users != nil && len(elem.Users) > 0 {
				filter.FilterByUsers(elem.Users)
				if elem.UsersAccessType == 1 {
					filter.FilterByUsersPrivatesOnly()
				} else if elem.UsersAccessType == 2 {
					filter.FilterByUsersPublicsOnly()
				}
			}
			if elem.Tokens != nil && len(elem.Tokens) > 0 {
				filter.FilterByTokens(elem.Tokens)
			}
			if elem.ConnId != "" {
				filter.FilterByConnIds([]string{elem.ConnId})
			}

			var result = filter.Result()

			if result != nil && len(result) > 0 {
				server.RLock()
				for _, userClients := range server.users {
					userClients.RLock()
					for _, connId := range result {
						conn, ok := userClients.conns[connId]
						if ok && conn != nil {
							conn.conn.WriteMessage(elem.MessageType, msg)
						}
					}
					userClients.RUnlock()
				}

				server.RUnlock()
			}
		}
	}
}

func (server *Server) checkExpired() {
	for range time.NewTicker(expirationTicker * time.Second).C {
		f := NewFilter(server)
		result := f.FilterByExpired().Result()
		if len(result) > 0 {
			server.queue.close <- &CloseMessage{
				ConnIds: result,
			}
		}
	}
}
