package utils

import (
	"time"

	"github.com/go-resty/resty/v2"
)

var RestClient *resty.Client

func init() {
	var envDebug = GetenvOrDefault("DEBUG", "false")
	var debug = false
	if envDebug == "true" {
		debug = true
	}
	RestClient = resty.New().SetTimeout(15 * time.Second).SetDebug(debug)
}
