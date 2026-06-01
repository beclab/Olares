package utils

import (
	"time"

	"github.com/go-resty/resty/v2"
)

var RestClient *resty.Client

func init() {
	RestClient = resty.New().SetTimeout(15 * time.Second).SetDebug(false)
}
