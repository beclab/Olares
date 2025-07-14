package apiserver

import (
	"context"

	"github.com/beclab/Olares/daemon/internel/apiserver/handlers"
	"github.com/beclab/Olares/daemon/internel/apiserver/server"
	"github.com/beclab/Olares/daemon/internel/ble"
	"github.com/gofiber/fiber/v2/middleware/cors"
	"github.com/gofiber/fiber/v2/middleware/logger"
)

func NewServer(ctx context.Context, port int) *server.Server {
	server.API.Port = port
	h := handlers.NewHandlers(ctx)

	server.API.UpdateAps = func(aplist []ble.AccessPoint) {
		h.ApList = aplist
	}

	s := server.API

	s.App.Use(cors.New())
	s.App.Use(logger.New())

	return s
}
