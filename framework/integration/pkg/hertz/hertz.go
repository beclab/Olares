package hertz

import (
	"fmt"
	router "integration/pkg/hertz/biz/router"
	"time"

	"github.com/cloudwego/hertz/pkg/app/server"
	"github.com/cloudwego/hertz/pkg/network/standard"
	"k8s.io/klog/v2"
)

const DefaultPort = "8090"

func HertzServer() {
	h := server.Default(
		server.WithHostPorts(fmt.Sprintf(":%s", DefaultPort)),
		server.WithMaxKeepBodySize(2<<20),
		server.WithTransport(standard.NewTransporter),
		server.WithReadTimeout(5*time.Minute),
		server.WithWriteTimeout(5*time.Minute),
		server.WithKeepAliveTimeout(3*time.Minute),
		server.WithStreamBody(false),
		server.WithReadBufferSize(64*1024),
		server.WithALPN(true),
		server.WithIdleTimeout(200*time.Second),
	)

	h.Use(router.TimingMiddleware())

	register(h)
	h.Spin()
	klog.Info("hertz server started")
}
