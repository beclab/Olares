package router

import (
	"context"
	"time"

	commonutils "integration/pkg/utils"

	"github.com/cloudwego/hertz/pkg/app"
	"k8s.io/klog/v2"
)

var excludePaths = []string{"/ping"}

func TimingMiddleware() app.HandlerFunc {
	return func(ctx context.Context, c *app.RequestContext) {
		if !commonutils.ListContains(excludePaths, string(c.Request.Path())) {
			start := time.Now()

			path := c.Path()

			klog.Infof("%s %s starts at %v", string(c.Method()), path, start.Format("2006-01-02 15:04:05"))

			defer func() {
				elapsed := time.Since(start)
				klog.Infof("%s %s execution time: %v", string(c.Method()), path, elapsed)
			}()
		}
		c.Next(ctx)

	}
}
