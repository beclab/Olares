package connectwifi

import (
	"context"
	"errors"

	"bytetrade.io/web3os/terminusd/pkg/commands"
	"bytetrade.io/web3os/terminusd/pkg/utils"
)

type connectWifi struct {
	commands.Operation
}

var _ commands.Interface = &connectWifi{}

func New() commands.Interface {
	return &connectWifi{
		Operation: commands.Operation{
			Name: commands.ConnectWifi,
		},
	}
}

func (s *connectWifi) Execute(ctx context.Context, p any) (res any, err error) {
	param, ok := p.(*Param)
	if !ok {
		err = errors.New("invalid param")
		return
	}

	// force enable wifi
	// err = utils.EnableWifi(ctx)
	// if err != nil {
	// 	return
	// }

	err = utils.ConnectWifi(ctx, param.SSID, param.Password)

	return
}
