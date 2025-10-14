package intranet

import (
	"net/http"
)

type proxyServer struct {
	http.Server
}
