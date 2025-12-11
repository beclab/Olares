package filters

import (
	"bytes"
	"fmt"
	"net/http"
	"runtime"
	"time"

	"github.com/emicklei/go-restful/v3"
	"olares.com/backup-server/pkg/util"
	"olares.com/backup-server/pkg/util/log"
)

func LogStackOnRecover(panicReason interface{}, w http.ResponseWriter) {
	var buffer bytes.Buffer
	buffer.WriteString(fmt.Sprintf("recover from panic situation: - %v\r\n", panicReason))
	for i := 2; ; i += 1 {
		_, file, line, ok := runtime.Caller(i)
		if !ok {
			break
		}
		buffer.WriteString(fmt.Sprintf("    %s:%d\r\n", file, line))
	}
	log.Error(buffer.String())
}

func LogRequestAndResponse(req *restful.Request, resp *restful.Response, chain *restful.FilterChain) {
	start := time.Now()
	chain.ProcessFilter(req, resp)

	realIp := util.RealClientIP(req.Request)

	// ignore localhost request
	if realIp == "127.0.0.1" {
		return
	}

	log.Infof("%s - \"%s %s %s\" %d %d %dms",
		realIp,
		req.Request.Method,
		req.Request.URL,
		req.Request.Proto,
		resp.StatusCode(),
		resp.ContentLength(),
		time.Since(start)/time.Millisecond,
	)
}
