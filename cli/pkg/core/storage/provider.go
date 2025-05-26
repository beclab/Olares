package storage

import (
	"bytetrade.io/web3os/installer/pkg/model"
)

type Provider interface {
	StartupCheck() (err error)

	Ping() (err error)

	// Close the underlying storage provider.
	Close() (err error)

	SaveInstallConfig(config model.InstallModelReq) (err error)
	SaveInstallLog(msg string, state string, percent int64) (err error)
	QueryInstallState(tspan int64) (data []model.InstallState, err error)
}
