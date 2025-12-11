package app

import (
	"fmt"
	"os"
	"sync"

	"bytetrade.io/web3os/tapr/pkg/constants"
	"bytetrade.io/web3os/tapr/pkg/ws"
	"bytetrade.io/web3os/tapr/pkg/ws/middleware"
	"github.com/gofiber/fiber/v2"
	"k8s.io/client-go/rest"
	"k8s.io/klog/v2"
)

const (
	DEBUG = false
)

var appEnvPort = "40010"
var appEnvUrl = "/tapr/ws/conn/debug"

type Server struct {
	app        *fiber.App
	KubeConfig *rest.Config
	controller *appController

	appPath string

	webSocketServer ws.WebSocketServer

	sync.RWMutex
}

func (server *Server) Init() error {
	var app = fiber.New()
	server.app = app

	server.controller = NewController(server)
	server.webSocketServer = ws.NewWebSocketServer()
	server.webSocketServer.SetHandler(server.controller.handleWebSocketMessage)

	server.getEnvAppInfo()

	return nil
}

func (server *Server) ServerRun() {

	server.app.Get("/ws", middleware.RequireHeader(), server.webSocketServer.New())

	server.app.Get("/tapr/ws/conn/list", server.controller.ListConnection)
	server.app.Post("/tapr/ws/conn/send", server.controller.SendMessage)
	server.app.Post("/tapr/ws/conn/close", server.controller.CloseConnection)

	server.app.Post("/tapr/ws/conn/debug", server.controller.DebugFunc)

	klog.Info("websocket server listening on 40010")
	klog.Fatal(server.app.Listen(":40010"))
}

func (s *Server) getEnvAppInfo() {
	var appPortName, appUrlName string

	if DEBUG {
		appPortName = appEnvPort
		appUrlName = appEnvUrl
	} else {
		appPortName = os.Getenv(constants.WsEnvAppPort)
		if appPortName == "" {
			return
		}
		appUrlName = os.Getenv(constants.WsEnvAppUrl)
		if appUrlName == "" {
			return
		}
	}

	var appPath = fmt.Sprintf("http://127.0.0.1:%s%s", appPortName, appUrlName)
	klog.Infof("backend app path: %s", appPath)
	s.appPath = appPath
}
