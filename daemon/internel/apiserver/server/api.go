package server

import (
	"fmt"

	"github.com/beclab/Olares/daemon/internel/ble"
	"github.com/gofiber/fiber/v2"
	"k8s.io/klog/v2"
)

type Server struct {
	Port      int
	App       *fiber.App
	UpdateAps func(aplist []ble.AccessPoint)
}

var API *Server = &Server{
	App: fiber.New(),
}

func (s *Server) Start() error {
	return s.App.Listen(fmt.Sprintf(":%d", s.Port))
}

func (s *Server) Shutdown() error {
	klog.Info("shutdown api server")
	if s.App == nil {
		return nil
	}
	return s.App.Shutdown()
}
