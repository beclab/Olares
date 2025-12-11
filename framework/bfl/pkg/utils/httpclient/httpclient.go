package httpclient

import (
	"time"

	"bytetrade.io/web3os/bfl/internal/log"

	"github.com/go-resty/resty/v2"
)

var (
	defaultTimeout = 5 * time.Second

	defaultRetryWaitTime = 2 * time.Second

	defaultRetry = 3
)

type Option struct {
	Debug         bool
	Timeout       time.Duration
	Retry         int
	RetryWaitTime time.Duration
}

func New(o *Option) *resty.Client {
	if o == nil {
		o = &Option{
			Timeout:       defaultTimeout,
			Retry:         defaultRetry,
			RetryWaitTime: defaultRetryWaitTime,
		}
	}

	c := resty.New()
	c.SetLogger(log.GetLogger())
	if o.Debug {
		c.SetDebug(true)
	}

	c.SetTimeout(o.Timeout).
		SetRetryCount(o.Retry).
		SetRetryWaitTime(o.RetryWaitTime).
		SetRetryMaxWaitTime(2 * o.RetryWaitTime)
	return c
}
