package apiserver

import (
	"context"

	"github.com/beclab/Olares/daemon/internel/apiserver/handlers"
	"github.com/beclab/Olares/daemon/internel/apiserver/server"
	"github.com/beclab/Olares/daemon/internel/ble"
)

func NewServer(ctx context.Context, port int) *server.Server {
	server.API.Port = port
	h := handlers.NewHandlers(ctx)

	server.API.UpdateAps = func(aplist []ble.AccessPoint) {
		h.ApList = aplist
	}

	s := server.API

	return s
}
