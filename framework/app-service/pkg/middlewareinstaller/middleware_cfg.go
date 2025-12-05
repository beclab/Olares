package middlewareinstaller

import "bytetrade.io/web3os/app-service/pkg/appcfg"

// MiddlewareConfig contains details of a workflow.
type MiddlewareConfig struct {
	Namespace      string
	ChartsName     string
	RepoURL        string
	Title          string
	Version        string
	MiddlewareName string // name of application displayed on shortcut
	OwnerName      string // name of owner who installed application
	Cfg            *appcfg.AppConfiguration
}
